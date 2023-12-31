package configuration

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
	ExitCodeNoCurves             = 13
	ExitCodeReadStdIn            = 14
)

// const (
// 	ModeCurve = "curve"
// )

var defaultConfig = Configuration{
	Running:         true,
	Active:          true,
	UI:              "graphic",
	CheckIntervalMs: 3000,
	PowerMode:       "auto",
	MinChange:       2.0,
	MaxStepUp:       4,
	MaxStepDown:     2,
	CurrentCurve:    "",
	Curves: map[string]Values{
		"default": {
			{40, 0},
			{60, 0.2},
			{80, 0.5},
			{85, 0.7},
			{90, 1},
		},
	},
}
