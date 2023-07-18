package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sirion/fanmi/app/config"
	"github.com/sirion/fanmi/app/ui"
)

const (
	FANMODE_MANUAL = "1"
	FANMODE_AUTO   = "2"
)

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
		ui.Fatal(config.ExitCodeReadTemperature, fmt.Sprintf("Error reading temperature from %s: %s\n", filePath, err.Error()))
	}

	str := strings.TrimSpace(string(data))
	temp, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		ui.Fatal(config.ExitCodeReadTemperature, fmt.Sprintf("Error reading temperature from %s: %s\n", filePath, err.Error()))
	}

	fTemp := float32(temp) / 1000
	return fTemp
}

func readSpeed(ui ui.UI, filePath string) float32 {
	data, err := os.ReadFile(filePath)
	if err != nil {
		ui.Fatal(config.ExitCodeReadSpeed, fmt.Sprintf("Error reading temperature from %s: %s\n", filePath, err.Error()))
	}

	str := strings.TrimSpace(string(data))
	temp, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		ui.Fatal(config.ExitCodeReadSpeed, fmt.Sprintf("Error reading temperature from %s: %s\n", filePath, err.Error()))
	}

	fSpeed := float32(temp) / 255
	return fSpeed
}

func writeFile(ui ui.UI, filePath string, value string) {
	file, err := os.OpenFile(filePath, os.O_WRONLY, os.ModePerm)
	if err != nil {
		ui.Fatal(config.ExitCodeWriteFile, fmt.Sprintf("Error writing to %s: %s\n", filePath, err.Error()))
	}

	_, err = file.Write([]byte(value))
	if err != nil {
		ui.Fatal(config.ExitCodeWriteFile, fmt.Sprintf("Error writing to %s: %s\n", filePath, err.Error()))
	}
}

func setSpeed(ui ui.UI, filePath string, factor float32) {
	if factor > 1 {
		factor = 1
	} else if factor < 0 {
		factor = 0
	}
	value := strconv.FormatInt(int64(factor*255), 10) + "\n"

	file, err := os.OpenFile(filePath, os.O_WRONLY, os.ModePerm)
	if err != nil {
		ui.Fatal(config.ExitCodeSpeedWrite, fmt.Sprintf("Error writing to %s: %s\n", filePath, err.Error()))
	}

	_, err = file.Write([]byte(value))
	if err != nil {
		ui.Fatal(config.ExitCodeSpeedWrite, fmt.Sprintf("Error writing to %s: %s\n", filePath, err.Error()))
	}

	ui.Speed(factor)
}

func calculateStep(temp float32, lowEntry, highEntry config.Entry) float32 {
	// Interpolate between steps
	smallerTemp := lowEntry.Temp
	smallerSpeed := lowEntry.Speed

	relTemp := (temp - smallerTemp) / (highEntry.Temp - smallerTemp)
	relSpeed := relTemp*(highEntry.Speed-smallerSpeed) + smallerSpeed

	return relSpeed
}

func fanControl(ui ui.UI, dirPath string, config *config.Configuration) chan bool {
	pwmPath := path.Join(dirPath, "pwm1")
	fanModePath := path.Join(dirPath, "pwm1_enable")
	tempInputPath := path.Join(dirPath, "temp1_input")

	done := make(chan bool)

	go (func() {
		var lastTemp float32 = -500

		for config.Running {
			temp := readTemp(ui, tempInputPath)
			ui.Temperature(temp)
			if !config.Active {
				speed := readSpeed(ui, pwmPath)
				ui.Speed(speed)

				if lastTemp != -500 {
					lastTemp = -500
					writeFile(ui, fanModePath, FANMODE_AUTO)
				}
				time.Sleep(time.Duration(config.CheckIntervalMs) * time.Millisecond)
				continue
			}

			deltaTemp := float64(lastTemp - temp)
			if math.Abs(deltaTemp) > float64(config.MinChange) {
				lastTemp = temp

				min := config.Values[0]
				max := config.Values[len(config.Values)-1]

				writeFile(ui, fanModePath, FANMODE_MANUAL)

				if temp < min.Temp {
					setSpeed(ui, pwmPath, min.Speed)
				} else if temp >= max.Temp {
					setSpeed(ui, pwmPath, max.Speed)
				} else {
					// between min and max
					var factor float32
					for i, en := range config.Values {
						if temp < en.Temp {
							factor = calculateStep(temp, config.Values[i-1], en)
							break
						}
					}
					setSpeed(ui, pwmPath, factor)
				}
			}

			time.Sleep(time.Duration(config.CheckIntervalMs) * time.Millisecond)

		}
		done <- true

		ui.Message("Resetting FanMode to Auto\n")
		writeFile(ui, fanModePath, FANMODE_AUTO)
	})()

	return done
}
