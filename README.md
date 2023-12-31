# fanmi - Linux AMDGPU Fan Control

A small utility that supports defining your own fan-curve and changing the power-profile.

## Why?

Inpired by [amdgpu-fancontrol](https://github.com/grmat/amdgpu-fancontrol), but I wanted a UI.

My Vega56 automatic fan control under linux runs way to hot and radeon-profile for some reason segfaults on my machine, so I used [amdgpu-fancontrol](https://github.com/grmat/amdgpu-fancontrol), which worked like a charm, but I wanted the following features:

- GUI
- SUID root for the executable so I can start it without using sudo

Also, I wanted to learn a new GUI-Framework with a simple project, that's why this one uses [fyne](https://fyne.io/).

| GUI | Terminal |
|-|-|
| ![GUI Interface](doc/main_window01.png) | ![](doc/main_console01.png) <hr>![](doc/main_console02.png) |

All relevant information about the used files comes from [AMDGPU Documentation](https://docs.kernel.org/6.1/gpu/amdgpu/thermal.html).

## Getting started

- Download the latest release binary or compile it yourself (see [build fanmi section](#build-fanmi)).
- (optional) Make the binary launch as root without sudo (`sudo chown root:root ./fanmi && sudo chmod u+s ./fanmi`)

## Configuration file

The configuration file is in JSON.
You can provide a path to a configuration file with `-config [path/to/file]`.
The default path will be displayed when usnig the `-help` argument. It should be something like `/home/[USERNAME]/.config/fanmi/config.json`.

Here are some configuration file examples:

- [Default configuration](doc/fanmi_default_config.json)
- [Start with low power](doc/fanmi_low_config.json)
- [Multiple curves](doc/fanmi_multicurve_config.json)

The configuration file may contain the following properties:

| Property | default value | |
| :-       | :- | :- |
| checkIntervalMs | 3000 |  How often to measure (and update) fan speed (in milliseconds) |
| minChange | 2.0 | The minimum change that needs to bemeasured (in °C) before a different speed is set |
| powerMode | "" | The powermode ("auto", "high", "low") of the graphics card |
| curves | [default*](#default-fan-curve) | Map of named fan curves |
| curve | "default" | The curve active at start-up |
| maxStepUp | 4.0 | The maximum upwards % change of the fan per `checkIntervalMs` |
| maxStepDown | 2.0 | The maximum downwards % change of the fan per `checkIntervalMs` |

## CLI Use

You can have three UI-options:

1. `--ui graphic` - GUI - (Default) Shows a window with temperature, fan-speed and option to switch on/off and a power profile dropdown
2. `--ui console` - Console - Prints out temperature and fan-speed on the console - press space to switch on/off, 'a', 'l', 'h' to switch power profile, 'c' to change the curve and 'q' or ctrl-c to exit
3. `--ui none` - No output

### Curves

If you want to change the fan-curve, you can create a configuration file in JSON.
Either provide the path to the file as CLI-argument (`--config "path/to/config.json"`), or put it at the default location, which you can find by invoking fanmi with `--help`. (Should be `/home/[USERNAME]/.config/fanmi/config.json` on most systems.)

You can define multiple curves in a configuration file.

#### Default fan curve

The default fan curve looks like this:

| Temp | Speed |
|   -: |    -: |
|  40° |    0% |
|  60° |   20% |
|  80° |   50% |
|  85° |   70% |
|  90° |  100% |

Which corresponds to the following map in the configuration file:

```json
"curves": {
    "default": [
        {
            "Temp": 40,
            "Speed": 0
        },
        {
            "Temp": 60,
            "Speed": 0.2
        },
        {
            "Temp": 80,
            "Speed": 0.5
        },
        {
            "Temp": 85,
            "Speed": 0.7
        },
        {
            "Temp": 90,
            "Speed": 1
        }
    ]
}
```

### Console UI

When using the console-UI, you can use the following keys:

|     Key | Action                                     |
|      -: | :-                                         |
| [SPACE] | Activate/deactivate                        |
|       a | Set power-profile to "auto"                |
|       h | Set power-profile to "high"                |
|       l | Set power-profile to "low"                 |
|       c | Switch to next curve                       |
|       q | Quit                                       |

### Exit codes

In addition to printing the error message to stderr, the application exits with an exitcode describing the problem:

| Code | Description |
|-|-|
| 1 | Could not open device |
| 2 | Could not find decive |
| 3 | Could not find at least one compatible device |
| 4 | Could not determine current user |
| 5 | You do not have (effective) root permissions |
| 6 | Could not read temperature |
| 7 | Could not write to file |
| 8 | Could not write fan speed |
| 9 | Could not read fan speed |
| 10 | Could not find user config directory |
| 11 | Could not read configuration file |
| 12 | Could not parse configuration file |
| 13 | No fan curves found |

## Build fanmi

There are three ways to build fanmi:

1. The normal way: `go build -o bin/fanmi app/*.go`
1. Build and make the executable file SUID root: `util/build.sh`
1. Build, compress and make the executable file SUID root: `util/build.sh release`

See [util/build.sh](util/build.sh) for the steps.

## Future

### Known Bugs

- The app sometimes hangs after pressing ctrl-c when using the GUI
- Debug output (`-v`) and console-UI are not compatible

### ToDos & Planned Features

- The Configuration should be editable in the GUI
- A graphic representation (chart) of the fan curve
- Minimize to systray (this is supported by fyne, but does not work on my system)
- Autostart feature (maybe this should just be part of the documentation)
 