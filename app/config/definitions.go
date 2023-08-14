package config

const (
	ExitCodeOpenDevice           = 1
	ExitCodeFindDevice           = 2
	ExitCodeFindCompatibleDevice = 3
	ExitCodeGetUser              = 4
	ExitCodeRoot                 = 5
	ExitCodeReadTemperature      = 6
	ExitCodeWriteFile            = 7
	ExitCodeSpeedWrite           = 8
	ExitCodeReadSpeed            = 9
	ExitCodeUserConfigDir        = 10
	ExitCodeUserConfigFile       = 11
	ExitCodeUserParseConfig      = 12
)

var defaultConfig = Configuration{
	Running:         true,
	Active:          true,
	UI:              "graphic",
	CheckIntervalMs: 3000,
	Mode:            "",
	MinChange:       2,
	Values: Values{
		{40, 0},
		{60, 0.2},
		{80, 0.5},
		{85, 0.7},
		{90, 1},
	},
}
