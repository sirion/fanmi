package ui

import (
	"fmt"
	"os"

	"github.com/sirion/fanmi/app/configuration"
)

type NoUI struct {
	done    chan bool
	running chan bool
}

func (ui *NoUI) Init(config *configuration.Configuration) chan bool {
	ui.done = make(chan bool)
	return ui.done
}

func (ui *NoUI) Run() {
	<-ui.running
}

func (ui *NoUI) Exit() {
	ui.done <- true
	ui.running <- true
}

func (*NoUI) Fatal(exitCode int, message string) {
	fmt.Fprint(os.Stderr, message)
	os.Exit(exitCode)
}
func (*NoUI) Temperature(float32)   {}
func (*NoUI) Speed(float32)         {}
func (*NoUI) Message(string)        {}
func (*NoUI) PowerMode(mode string) {}
