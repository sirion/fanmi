package ui

import "github.com/sirion/fanmi/app/config"

type UI interface {
	Init(config *config.Configuration) chan bool
	Run()
	Exit()
	Message(string)
	Fatal(exitCode int, message string)
	Temperature(temp float32)
	Speed(speed float32)
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
