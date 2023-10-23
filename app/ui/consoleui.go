package ui

import (
	"fmt"
	"os"

	"github.com/sirion/fanmi/app/configuration"
	"golang.org/x/term"
)

type ConsoleUI struct {
	temp    float32
	speed   float32
	done    chan bool
	running chan bool
	config  *configuration.Configuration
}

func (ui *ConsoleUI) Init(conf *configuration.Configuration) chan bool {
	ui.config = conf
	ui.done = make(chan bool, 2)
	ui.running = make(chan bool, 2)

	go (func() {
		// switch stdin into 'raw' mode
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			ui.Message(fmt.Sprintf("Error setting stdin to raw cannot read keys: %s\n", err.Error()))
			return
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		bt := make([]byte, 4)
		for ui.config.Running {
			n, err := os.Stdin.Read(bt)
			if err != nil {
				ui.Message(fmt.Sprintf("Error reading from standard input: %s\n", err.Error()))
				os.Exit(configuration.ExitCodeReadStdIn)
			}
			fmt.Printf("\n\r%x\n", bt[0:n])
			if n == 0 {
				ui.Message("End of input from console.\n")
				os.Exit(configuration.ExitCodeReadStdIn)

			} else if bt[0] == ' ' {
				// Space toggles active/inactive
				ui.config.Active = !ui.config.Active
				ui.update()
			} else if bt[0] == 'a' {
				ui.config.SetPowerMode("auto")
			} else if bt[0] == 'l' {
				ui.config.SetPowerMode("low")
			} else if bt[0] == 'h' {
				ui.config.SetPowerMode("high")
			} else if bt[0] == 'c' {
				ui.config.NextCurve()
				// } else if bt[0] == 't' {
				// 	// Switch between target and Curve mode
				// 	if ui.config.Mode == configuration.ModeCurve {
				// 		ui.config.Mode = configuration.ModeTemp
				// 	} else if len(ui.config.Curves) > 0 {
				// 		ui.config.Mode = configuration.ModeCurve
				// 		ui.config.SetCurve(ui.config.CurveNames[0])
				// 	}
				// } else if n == 3 && bt[0] == 0x1b && bt[1] == 0x5b && bt[2] == 0x41 {
				// 	// Up
				// 	ui.config.TargetTemperature += 1

				// } else if n == 3 && bt[0] == 0x1b && bt[1] == 0x5b && bt[2] == 0x42 {
				// 	// Down
				// 	ui.config.TargetTemperature -= 1

				// } else if n == 3 && bt[0] == 0x1b && bt[1] == 0x5b && bt[2] == 0x43 {
				// 	// Left
				// 	ui.config.MaximumTemperatureDelta += 1

				// } else if n == 3 && bt[0] == 0x1b && bt[1] == 0x5b && bt[2] == 0x44 {
				// 	// Left
				// 	if ui.config.MaximumTemperatureDelta >= 2 {
				// 		ui.config.MaximumTemperatureDelta -= 1
				// 	}

			} else if bt[0] == 'q' {
				ui.Message("Exiting\n")
				ui.config.Running = false
				ui.config.Active = false
				ui.Exit()
			} else if bt[0] == 3 {
				// Ctrl-C
				ui.Message("Ctrl-c caught - Exiting\n")
				ui.config.Running = false
				ui.config.Active = false
				ui.Exit()
			}
			ui.update()
		}
	})()

	return ui.done
}

func (ui *ConsoleUI) Run() {
	<-ui.running
}

func (ui *ConsoleUI) Exit() {
	ui.done <- true
	ui.running <- true
	os.Stderr.Sync()
	os.Stdout.Sync()
}

func (*ConsoleUI) Fatal(exitCode int, message string) {
	fmt.Fprint(os.Stderr, message)
	os.Exit(exitCode)
}

func (ui *ConsoleUI) Temperature(temp float32) {
	ui.temp = temp
	ui.update()
}

func (ui *ConsoleUI) Speed(speed float32) {
	ui.speed = speed
	ui.update()
}

func (*ConsoleUI) PowerMode(mode string) {
	// Ignored for now
}

func (*ConsoleUI) Message(message string) {
	fmt.Printf("\x1b[0K\r\n")
	fmt.Print(message)
}

func (ui *ConsoleUI) update() {
	speedPercent := float32(int(ui.speed*10000)) / 100
	prefix := ""
	if !ui.config.Active {
		prefix = "[INACTIVE] "
	}
	powerMode := ""
	if ui.config.PowerMode != "" {
		powerMode = fmt.Sprintf("\t(Profile: %s)", ui.config.PowerMode)
	}

	mode := ""
	// if ui.config.Mode == configuration.ModeCurve {
	if len(ui.config.Curves) > 1 {
		mode = fmt.Sprintf("\t(Curve: %s)", ui.config.CurrentCurve)
	}
	// } else {
	// 	mode = fmt.Sprintf("\t(Target: %2.1f° (+%2.1f))", ui.config.TargetTemperature, ui.config.MaximumTemperatureDelta)
	// }

	fmt.Printf("\r\x1b[0K%sTemperature: %2.0f°\tSpeed: %3.2f%%%s%s\r", prefix, ui.temp, speedPercent, powerMode, mode)
}
