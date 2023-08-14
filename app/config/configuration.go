package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
)

type Configuration struct {
	Mode            string  `json:"powerMode"`
	CheckIntervalMs uint    `json:"checkIntervalMs"`
	MinChange       float32 `json:"minChange"`
	Values          Values  `json:"values"`

	ModeChanged bool   `json:"-"`
	Running     bool   `json:"-"`
	Active      bool   `json:"-"`
	UI          string `json:"-"`
}

func ReadConfig() *Configuration {
	// Read CLI options
	var ui string
	var configPath string

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error locating config directory for current user: %s", err.Error())
		os.Exit(ExitCodeUserConfigDir)
	}
	configDir := path.Join(userConfigDir, "fanmi")
	configPath = path.Join(configDir, "config.json")

	flag.StringVar(&ui, "ui", "graphic", `Which UI to use, either "graphic", "console" or "none"`)
	flag.StringVar(&configPath, "config", configPath, `Path to the (optional) configuration file`)
	help := flag.Bool("help", false, "Show this help")
	flag.Parse()

	if *help {
		showHelp()
		os.Exit(0)
	}

	config := &defaultConfig

	if len(configPath) > 0 {
		// Load configuration
		data, err := os.ReadFile(configPath)
		noConfigFile := os.IsNotExist(err)
		if err != nil && !noConfigFile {
			fmt.Fprintf(os.Stderr, "Error reading config file: %s", err.Error())
			os.Exit(ExitCodeUserConfigFile)
		}

		if !noConfigFile {
			// Parse configuration
			config = &Configuration{}
			err = json.Unmarshal(data, config)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing config file: %s", err.Error())
				os.Exit(ExitCodeUserParseConfig)
			}
		}
	}

	config.Active = true
	config.Running = true
	config.UI = ui

	sort.Sort(config.Values)

	// TODO: Sort Values
	return config
}

func showHelp() {
	defaultConfigJSON, _ := json.MarshalIndent(defaultConfig, "", "\t")

	fmt.Println(`CLI Options:`)
	flag.PrintDefaults()

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
	os.Exit(0)
}

func (c *Configuration) SetPowerMode(mode string) {
	c.ModeChanged = mode != c.Mode
	c.Mode = mode
	// fmt.Printf("Power mode changed to %s (%t)\n", c.Mode, c.ModeChanged)
}
