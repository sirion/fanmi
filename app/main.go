package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"os/user"
	"path"
	"syscall"

	"github.com/sirion/fanmi/app/configuration"
	"github.com/sirion/fanmi/app/ui"
)

func main() {
	// Task: Read Configuration
	config := configuration.ReadConfig()

	ui := ui.CreateUI(config.UI)

	uiClosed := ui.Init(config)
	go (func() {
		<-uiClosed
		config.Running = false
	})()

	u, err := user.Current()
	if err != nil {
		ui.Fatal(configuration.ExitCodeGetUser, fmt.Sprintf("You are not root. Error: %s\n", err.Error()))
	}
	uid := os.Geteuid()
	if uid != 0 {
		ui.Fatal(configuration.ExitCodeRoot, fmt.Sprintf("You are not root: %s\n", u.Username))
	}

	fsDirDrm := os.DirFS("/sys/class/drm/")

	// Task: Find card /sys/class/drm/card?/device/hwmon/hwmon?
	pwmMatches, err := fs.Glob(fsDirDrm, "card?/device/hwmon/hwmon?/pwm1")
	if err != nil {
		ui.Fatal(configuration.ExitCodeOpenDevice, fmt.Sprintf("Error opening device: %s\n", err.Error()))
	}
	if len(pwmMatches) == 0 {
		ui.Fatal(configuration.ExitCodeFindDevice, "No device found at /sys/class/drm/card?\n")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT, syscall.SIGSTKFLT, syscall.SIGSYS)

	go (func() {
		<-c
		config.Running = false
		config.Active = false
		ui.Message("Signal caught. Exiting.\n")
	})()

	workers := make([]chan bool, 0)
	for _, matchPath := range pwmMatches {
		// Task: Start Monitor Routine per card
		hwmonDirPath := path.Dir("/sys/class/drm/" + matchPath)
		deviceDirPath := path.Dir(path.Dir(hwmonDirPath))

		pwmPath := path.Join(hwmonDirPath, "pwm1")
		enablePath := path.Join(hwmonDirPath, "pwm1_enable")
		tempInputPath := path.Join(hwmonDirPath, "temp1_input")
		if !fileExists(pwmPath) || !fileExists(enablePath) || !fileExists(tempInputPath) {
			continue
		}

		worker := NewFanControl(ui, deviceDirPath, hwmonDirPath, config)
		workers = append(workers, worker.Run())
	}

	// Task: Check Compatibility
	if len(workers) == 0 {
		ui.Fatal(configuration.ExitCodeFindCompatibleDevice, "No compatible devices found\n")
	}

	ui.Run()
	for _, worker := range workers {
		<-worker
	}
	ui.Exit()

	fmt.Println("")
}
