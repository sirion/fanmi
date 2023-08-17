package ui

import (
	"fmt"
	"os"

	"github.com/sirion/fanmi/app/config"
	"golang.org/x/term"
)

type ConsoleUI struct {
	temp    float32
	speed   float32
	done    chan bool
	running chan bool
	config  *config.Configuration
}

func (ui *ConsoleUI) Init(conf *config.Configuration) chan bool {
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

		bt := make([]byte, 1)
		for ui.config.Running {
			n, err := os.Stdin.Read(bt)
			if err != nil {
				ui.Message(fmt.Sprintf("Error reading from standard input: %s\n", err.Error()))
				os.Exit(config.ExitCodeReadStdIn)
			}
			if n == 0 {
				ui.Message("End of input from console.\n")
				os.Exit(config.ExitCodeReadStdIn)

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
	if ui.config.Mode != "" {
		powerMode = fmt.Sprintf("\t(Profile: %s)", ui.config.Mode)
	}
	curve := ""
	if len(ui.config.Curves) > 1 {
		curve = fmt.Sprintf("\t(Curve: %s)", ui.config.CurrentCurve)
	}

	fmt.Printf("\r\x1b[0K%sTemperature: %2.0f°\tSpeed: %3.2f%%%s%s\r", prefix, ui.temp, speedPercent, powerMode, curve)
}
