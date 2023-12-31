package ui

import "github.com/sirion/fanmi/app/configuration"

type UI interface {
	Init(config *configuration.Configuration) chan bool
	Run()
	Exit()
	Message(string)
	Fatal(exitCode int, message string)
	Temperature(temp float32)
	Speed(speed float32)
	PowerMode(mode string)
}

func CreateUI(uiType string) UI {
	if uiType == "console" {
		return &ConsoleUI{}
	} else if uiType == "none" {
		return &NoUI{}
	} else {
		return &FyneUI{}
	}
}
