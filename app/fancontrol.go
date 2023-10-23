package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sirion/fanmi/app/configuration"
	"github.com/sirion/fanmi/app/debug"
	"github.com/sirion/fanmi/app/ui"
)

const (
	FANMODE_MANUAL = "1"
	FANMODE_AUTO   = "2"
)

type FanControl struct {
	done          chan bool
	ui            ui.UI
	config        *configuration.Configuration
	deviceDirPath string
	hwmonDirPath  string
	powerModePath string
	pwmPath       string
	fanModePath   string
	tempInputPath string
	byTempData    byTempData
}

type byTempData struct {
	currentFactor float32
}

func NewFanControl(ui ui.UI, deviceDirPath, hwmonDirPath string, config *configuration.Configuration) *FanControl {
	return &FanControl{
		done:          make(chan bool),
		ui:            ui,
		deviceDirPath: deviceDirPath,
		hwmonDirPath:  hwmonDirPath,
		config:        config,
		powerModePath: path.Join(deviceDirPath, "power_dpm_force_performance_level"),
		pwmPath:       path.Join(hwmonDirPath, "pwm1"),
		fanModePath:   path.Join(hwmonDirPath, "pwm1_enable"),
		tempInputPath: path.Join(hwmonDirPath, "temp1_input"),
		byTempData: byTempData{
			currentFactor: -1,
		},
	}
}

func (f *FanControl) Run() chan bool {
	go (func() {
		powerModeAvailable := true
		var lastTemp float32 = -500
		lastCurve := f.config.Curve

		for f.config.Running {
			// Power Mode
			if powerModeAvailable {
				var err error

				if f.config.PowerModeChanged && f.config.PowerMode != "" {
					err = writePowerMode(f.powerModePath, f.config.PowerMode)
					f.config.PowerModeChanged = false
					if err != nil {
						f.ui.Message(err.Error())
					}
				}
				f.config.PowerMode, err = readPowerMode(f.powerModePath)
				if err != nil {
					if powerModeAvailable {
						fmt.Fprintf(os.Stderr, "Cannot read power mode: %s", err.Error())
					}
					powerModeAvailable = false
					f.config.PowerMode = ""
				}
			}
			f.ui.PowerMode(f.config.PowerMode)

			temp := readTemp(f.ui, f.tempInputPath)
			f.ui.Temperature(temp)
			if !f.config.Active {
				speed := readSpeed(f.ui, f.pwmPath)
				f.ui.Speed(speed)

				if lastTemp != -500 {
					lastTemp = -500
					writeFile(f.ui, f.fanModePath, FANMODE_AUTO)
				}
				time.Sleep(time.Duration(f.config.CheckIntervalMs) * time.Millisecond)
				continue
			}

			deltaTemp := float32(lastTemp - temp)

			if /* f.config.Mode == configuration.ModeCurve && */ &f.config.Curve != &lastCurve {
				deltaTemp = f.config.MinChange + 1
			}
			if math.Abs(float64(deltaTemp)) > float64(f.config.MinChange) {
				// switch f.config.Mode {
				// case configuration.ModeCurve:
				// 	fallthrough
				// default:
				f.byCurve(temp, f.ui, f.pwmPath, f.fanModePath, f.config)
				// }

				lastTemp = temp
			}
			time.Sleep(time.Duration(f.config.CheckIntervalMs) * time.Millisecond)
		}

		f.done <- true

		f.ui.Message("Resetting FanMode to Auto\n")
		writeFile(f.ui, f.fanModePath, FANMODE_AUTO)
	})()

	return f.done
}

func (f *FanControl) byCurve(temp float32, ui ui.UI, pwmPath, fanModePath string, config *configuration.Configuration) {
	writeFile(ui, fanModePath, FANMODE_MANUAL)

	min := config.Curve[0]
	max := config.Curve[len(config.Curve)-1]

	if temp < min.Temp {
		setSpeed(ui, pwmPath, min.Speed)
	} else if temp >= max.Temp {
		setSpeed(ui, pwmPath, max.Speed)
	} else {
		// between min and max
		var factor float32
		for i, en := range config.Curve {
			if temp < en.Temp {
				factor = calculateStep(temp, config.Curve[i-1], en)
				break
			}
		}
		setSpeed(ui, pwmPath, factor)
	}
}

func fileExists(filePath string) bool {
	stat, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	if stat.IsDir() {
		return false
	}

	return true
}

func readTemp(ui ui.UI, filePath string) float32 {
	data, err := os.ReadFile(filePath)
	if err != nil {
		ui.Fatal(configuration.ExitCodeReadTemperature, fmt.Sprintf("Error reading temperature from %s: %s\n", filePath, err.Error()))
	}

	str := strings.TrimSpace(string(data))
	temp, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		ui.Fatal(configuration.ExitCodeReadTemperature, fmt.Sprintf("Error reading temperature from %s: %s\n", filePath, err.Error()))
	}

	fTemp := float32(temp) / 1000
	debug.Log("Read temperature %f to %s\n", fTemp, filePath)
	return fTemp
}

func readSpeed(ui ui.UI, filePath string) float32 {
	data, err := os.ReadFile(filePath)
	if err != nil {
		ui.Fatal(configuration.ExitCodeReadSpeed, fmt.Sprintf("Error reading temperature from %s: %s\n", filePath, err.Error()))
	}

	str := strings.TrimSpace(string(data))
	temp, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		ui.Fatal(configuration.ExitCodeReadSpeed, fmt.Sprintf("Error reading temperature from %s: %s\n", filePath, err.Error()))
	}

	fSpeed := float32(temp) / 255
	debug.Log("Read fan speed %f to %s\n", fSpeed, filePath)

	return fSpeed
}

func writeFile(ui ui.UI, filePath string, value string) {
	file, err := os.OpenFile(filePath, os.O_WRONLY, os.ModePerm)
	if err != nil {
		ui.Fatal(configuration.ExitCodeWriteFile, fmt.Sprintf("Error writing to %s: %s\n", filePath, err.Error()))
	}

	_, err = file.Write([]byte(value))
	if err != nil {
		ui.Fatal(configuration.ExitCodeWriteFile, fmt.Sprintf("Error writing to %s: %s\n", filePath, err.Error()))
	}
}

func setSpeed(ui ui.UI, filePath string, factor float32) {
	if factor > 1 {
		factor = 1
	} else if factor < 0 {
		factor = 0
	}
	value := strconv.FormatInt(int64(factor*255), 10) + "\n"
	debug.Log("Writing %s to %s\n", strings.TrimSpace(value), filePath)

	file, err := os.OpenFile(filePath, os.O_WRONLY, os.ModePerm)
	if err != nil {
		ui.Fatal(configuration.ExitCodeSpeedWrite, fmt.Sprintf("Error writing to %s: %s\n", filePath, err.Error()))
	}

	_, err = file.Write([]byte(value))
	if err != nil {
		ui.Fatal(configuration.ExitCodeSpeedWrite, fmt.Sprintf("Error writing to %s: %s\n", filePath, err.Error()))
	}

	debug.Log("Wrote fan speed %f to %s\n", factor, filePath)

	ui.Speed(factor)
}

func calculateStep(temp float32, lowEntry, highEntry configuration.Entry) float32 {
	// Interpolate between steps
	smallerTemp := lowEntry.Temp
	smallerSpeed := lowEntry.Speed

	relTemp := (temp - smallerTemp) / (highEntry.Temp - smallerTemp)
	relSpeed := relTemp*(highEntry.Speed-smallerSpeed) + smallerSpeed

	return relSpeed
}

func writePowerMode(powerModePath, mode string) error {
	file, err := os.OpenFile(powerModePath, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing to %s: %s", powerModePath, err.Error())
	}

	_, err = file.Write([]byte(mode))
	if err != nil {
		return fmt.Errorf("error writing to %s: %s", powerModePath, err.Error())
	}

	debug.Log("Written %s to %s\n", mode, powerModePath)
	return nil
}

func readPowerMode(powerModePath string) (string, error) {
	data, err := os.ReadFile(powerModePath)
	if err != nil {
		return "", fmt.Errorf("error reading temperature from %s: %s", powerModePath, err.Error())
	}

	debug.Log("Read power mode: %s", data)
	return strings.TrimSpace(string(data)), nil
}
