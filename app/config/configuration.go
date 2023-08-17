package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/sirion/fanmi/app/debug"
)

type Configuration struct {
	Mode            string            `json:"powerMode"`
	CheckIntervalMs uint              `json:"checkIntervalMs"`
	MinChange       float32           `json:"minChange"`
	Curves          map[string]Values `json:"curves"`
	CurrentCurve    string            `json:"curve"`

	ModeChanged bool     `json:"-"`
	Running     bool     `json:"-"`
	Active      bool     `json:"-"`
	UI          string   `json:"-"`
	CurveNames  []string `json:"-"`
	Curve       Values   `json:"-"`
}

func ReadConfig() *Configuration {
	// Read CLI options
	var ui string

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error locating config directory for current user: %s", err.Error())
		os.Exit(ExitCodeUserConfigDir)
	}
	configDir := path.Join(userConfigDir, "fanmi")
	defaultConfigPath := path.Join(configDir, "config.json")
	configPath := ""

	flag.StringVar(&ui, "ui", "graphic", `Which UI to use, either "graphic", "console" or "none"`)
	flag.StringVar(&configPath, "config", defaultConfigPath, `Path to the (optional) configuration file`)
	flag.BoolVar(&debug.DebugOutput, "v", debug.DebugOutput, `Print debug information to stdout`)
	help := flag.Bool("help", false, "Show this help (add -v for more information)")
	flag.Parse()

	if *help {
		showHelp(debug.DebugOutput)
		os.Exit(0)
	}

	config := &defaultConfig
	config.Active = true
	config.Running = true
	config.UI = ui
	config.loadFromFile(configPath, defaultConfigPath)
	config.prepareCurves()
	// Make sure the configured mode is set before reading current mode
	config.ModeChanged = config.Mode != ""

	debug.LogJSON("Configuration:\n", config, "\n\n")

	return config
}

func showHelp(verbose bool) {
	defaultConfigJSON, _ := json.MarshalIndent(defaultConfig, "", "\t")

	fmt.Println(`CLI Options:`)
	flag.PrintDefaults()

	if verbose {
		fmt.Println(``)
		fmt.Println(`Exit Codes:`)
		fmt.Printf("+-----+-----------------------------------------------+\n")
		fmt.Printf("| %3d | Could not open device                         |\n", ExitCodeOpenDevice)
		fmt.Printf("| %3d | Could not find decive                         |\n", ExitCodeFindDevice)
		fmt.Printf("| %3d | Could not find at least one compatible device |\n", ExitCodeFindCompatibleDevice)
		fmt.Printf("| %3d | Could not determine current user              |\n", ExitCodeGetUser)
		fmt.Printf("| %3d | You do not have (effective) root permissions  |\n", ExitCodeRoot)
		fmt.Printf("| %3d | Could not read temperature                    |\n", ExitCodeReadTemperature)
		fmt.Printf("| %3d | Could not write to file                       |\n", ExitCodeWriteFile)
		fmt.Printf("| %3d | Could not write fan speed                     |\n", ExitCodeSpeedWrite)
		fmt.Printf("| %3d | Could not read fan speed                      |\n", ExitCodeReadSpeed)
		fmt.Printf("| %3d | Could not find user config directory          |\n", ExitCodeUserConfigDir)
		fmt.Printf("| %3d | Could not read configuration file             |\n", ExitCodeUserConfigFile)
		fmt.Printf("| %3d | Could not parse configuration file            |\n", ExitCodeUserParseConfig)
		fmt.Printf("+-----+-----------------------------------------------+\n")

		fmt.Println(``)
		fmt.Println(`Default configuration:`)
		fmt.Println(string(defaultConfigJSON))
		fmt.Println(``)
	}
	os.Exit(0)
}

func (c *Configuration) SetPowerMode(mode string) {
	c.ModeChanged = mode != c.Mode
	c.Mode = mode
	debug.Log("Power mode changed to %s (%t)\n", c.Mode, c.ModeChanged)
}

func (c *Configuration) SetCurve(curveName string) {
	ok := false
	c.Curve, ok = c.Curves[curveName]

	if !ok {
		fmt.Fprintf(os.Stderr, "Selected fan curve '%s' not found\n", curveName)
		for name, curve := range c.Curves {
			fmt.Fprintf(os.Stderr, "Using fan curve '%s'\n", name)
			c.CurrentCurve = name
			c.Curve = curve
			break
		}
	} else {
		c.CurrentCurve = curveName
	}

	debug.Log("Curve changed to %s\n", curveName)
	debug.LogJSON("Curve: ", c.Curve, "")
}

func (c *Configuration) NextCurve() {
	i := -1
	name := ""
	for i, name = range c.CurveNames {
		if name == c.CurrentCurve {
			break
		}
	}

	if i == len(c.CurveNames)-1 {
		i = 0
	} else {
		i = i + 1
	}

	c.SetCurve(c.CurveNames[i])
}

func (config *Configuration) loadFromFile(configPath, defaultConfigPath string) {
	configLoaded := false
	var data []byte
	var err error

	if configPath != "" {
		// Load user provided configuration
		debug.Log("Loading configuration file %s\n", configPath)
		data, err = os.ReadFile(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config file: %s, trying default config path\n", err.Error())
		} else {
			configLoaded = true
		}
	}
	if !configLoaded {
		debug.Log("Loading default configuration file %s\n", defaultConfigPath)
		data, err = os.ReadFile(defaultConfigPath)
		noConfigFile := os.IsNotExist(err)
		if err != nil && !noConfigFile {
			fmt.Fprintf(os.Stderr, "Error reading default config file %s: %s, fallback to defaults\n", configPath, err.Error())
		} else if !noConfigFile {
			configLoaded = true
		}
	}

	if configLoaded {
		// Parse configuration
		err = json.Unmarshal(data, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing config file: %s\n", err.Error())
			os.Exit(ExitCodeUserParseConfig)
		}
	}
}

func (config *Configuration) prepareCurves() {
	for i := range config.Curves {
		sort.Sort(config.Curves[i])
	}

	if len(config.Curves) == 0 {
		fmt.Fprint(os.Stderr, "No available fan curves")
		os.Exit(ExitCodeNoCurves)
	}

	if config.CurrentCurve == "" && len(config.Curves) == 1 {
		for _, curve := range config.Curves {
			config.Curve = curve
			break
		}
	} else {
		curve, ok := config.Curves[config.CurrentCurve]
		if !ok {
			fmt.Fprintf(os.Stderr, "Selected fan curve '%s' not found", config.CurrentCurve)
			os.Exit(ExitCodeNoCurves)
		}
		config.Curve = curve
	}

	// Sorted list of available curveNames
	config.CurveNames = make([]string, 0, len(config.Curves))
	for key := range config.Curves {
		config.CurveNames = append(config.CurveNames, key)
	}
	sort.Slice(config.CurveNames, func(a, b int) bool {
		return config.CurveNames[a] < config.CurveNames[b]
	})

}
