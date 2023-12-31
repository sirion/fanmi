package ui

import (
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sirion/fanmi/app/configuration"
)

/// Helper Functions

func NewBigText(text string) *canvas.Text {
	app := fyne.CurrentApp()
	th := app.Settings().Theme()
	thVar := app.Settings().ThemeVariant()

	cTxt := canvas.NewText(text, th.Color(theme.ColorNameForeground, thVar))
	cTxt.TextSize = theme.TextSize() * 2

	return cTxt
}

/// FyneUI

type FyneUI struct {
	done   chan bool
	config *configuration.Configuration

	app   fyne.App
	win   fyne.Window
	temp  *canvas.Text
	speed *canvas.Text
}

func (ui *FyneUI) Init(config *configuration.Configuration) chan bool {
	ui.config = config

	ui.app = app.New()

	ui.app.SetIcon(iconApp)

	ui.win = ui.app.NewWindow("FanMi")
	ui.win.SetMaster()
	ui.win.SetFixedSize(true)
	ui.win.SetOnClosed(ui.win.Close)
	ui.win.Resize(fyne.NewSize(300, 100))

	ui.temp = NewBigText("0")
	ui.temp.Alignment = fyne.TextAlignTrailing

	ui.speed = NewBigText("0")
	ui.speed.Alignment = fyne.TextAlignTrailing

	chkActive := widget.NewCheck("active", func(b bool) {
		config.Active = b
		if b {
			ui.temp.Color = theme.ForegroundColor()
			ui.speed.Color = theme.ForegroundColor()
		} else {
			ui.temp.Color = theme.DisabledColor()
			ui.speed.Color = theme.DisabledColor()
		}
		ui.temp.Refresh()
		ui.speed.Refresh()
	})
	chkActive.SetChecked(ui.config.Active)

	content := container.NewVBox(
		container.NewHBox(
			container.NewVBox(
				NewBigText("Temperature:"),
				layout.NewSpacer(),
				NewBigText("Fan Speed:"),
			),
			layout.NewSpacer(),
			container.NewVBox(
				ui.temp,
				layout.NewSpacer(),
				ui.speed,
			),
			container.NewVBox(
				NewBigText("°C"),
				layout.NewSpacer(),
				NewBigText("%"),
			),
		),
		container.NewHBox(
			chkActive,
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", iconSettings, ui.showSettingsWindow),
		),
	)

	ui.win.SetContent(content)

	ui.win.Show()

	ui.done = make(chan bool)
	return ui.done
}

func (ui *FyneUI) showSettingsWindow() {
	var curveText *canvas.Text
	var curve *widget.Select

	content := container.NewVBox()
	form := container.New(layout.NewFormLayout())
	content.Add(form)

	// Switch Power profile
	modeLabel := canvas.NewText("Power:", theme.ForegroundColor())
	mode := widget.NewSelect([]string{
		"auto",
		"low",
		"high",
		// "manual",
		// "profile_standard",
		// "profile_min_sclk",
		// "profile_min_mclk",
		// "profile_peak",
	}, func(mode string) {
		ui.config.SetPowerMode(mode)
	})
	mode.Selected = ui.config.PowerMode
	form.Add(modeLabel)
	form.Add(mode)
	AddSpacer(form)

	// Set Change Interval
	AddIntegerField(form, &ui.config.CheckIntervalMs, "Interval (ms):")

	// Set Minimal Temperature Change
	AddDecimalField(form, &ui.config.MinChange, "Min Change (°):")
	AddSpacer(form)

	// Set Minimal Up/Down Steps
	AddDecimalField(form, &ui.config.MaxStepUp, "Max Step Up (%):")
	AddDecimalField(form, &ui.config.MaxStepDown, "Max Step Down (%):")
	AddSpacer(form)

	// Switch Curve
	curve = widget.NewSelect(ui.config.CurveNames, func(string) {})
	curve.OnChanged = func(name string) {
		ui.config.SetCurve(name)
		curve.Selected = name
	}
	curve.Selected = ui.config.CurrentCurve
	curveText = canvas.NewText("Curve:", theme.ForegroundColor())

	form.Add(curveText)
	form.Add(curve) // TODO: Move to settings, remove from ui

	if len(ui.config.Curves) <= 1 {
		curve.Disable()
		curveText.Color = theme.DisabledColor()
	}

	win := ui.app.NewWindow("Settings")
	win.SetContent(content)
	// win.Resize(fyne.NewSize(300, 100))
	// win.SetFixedSize(true)

	win.Show()
}

func (ui *FyneUI) Run() {
	ui.app.Run()
	ui.done <- true
}

func (ui *FyneUI) Exit() {
	ui.app.Quit()
	go (func() {
		ui.done <- true
	})()
}

func (*FyneUI) Fatal(exitCode int, message string) {
	fmt.Fprint(os.Stderr, message)
	os.Exit(exitCode)
}
func (ui *FyneUI) Temperature(temp float32) {
	ui.temp.Text = fmt.Sprintf("%2.0f", temp)
	ui.temp.Refresh()
}

func (ui *FyneUI) Speed(speed float32) {
	ui.speed.Text = fmt.Sprintf("%2.1f", speed*100)
	ui.speed.Refresh()
}

func (ui *FyneUI) PowerMode(mode string) {} // Ignored, only shown in settings

func (*FyneUI) Message(message string) {
	fmt.Print(message)
}

/// Helper

func AddIntegerField(form *fyne.Container, configValue *uint32, label string) {
	input := widget.NewEntry()
	input.Text = fmt.Sprintf("%03d", *configValue)
	input.OnChanged = func(value string) {
		input.TextStyle.Bold = true
		input.TextStyle.Italic = true
		input.Refresh()
	}
	input.OnSubmitted = func(value string) {
		temp, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			input.Text = fmt.Sprintf("%03d", *configValue)
		}
		*configValue = uint32(temp)

		input.TextStyle.Bold = false
		input.TextStyle.Italic = false
		input.Refresh()
	}
	form.Add(canvas.NewText(label, theme.ForegroundColor()))
	form.Add(input)
}

func AddDecimalField(form *fyne.Container, configValue *float32, label string) {
	input := widget.NewEntry()
	input.Text = fmt.Sprintf("%2.1f", *configValue)
	input.OnChanged = func(value string) {
		input.TextStyle.Bold = true
		input.TextStyle.Italic = true
		input.Refresh()
	}
	input.OnSubmitted = func(value string) {
		temp, err := strconv.ParseFloat(value, 32)
		if err != nil {
			input.Text = fmt.Sprintf("%2.1f", *configValue)
		}
		*configValue = float32(temp)

		input.TextStyle.Bold = false
		input.TextStyle.Italic = false
		input.Refresh()
	}
	form.Add(canvas.NewText(label, theme.ForegroundColor()))
	form.Add(input)
}

func AddSpacer(form *fyne.Container) {
	form.Add(canvas.NewText(" ", theme.ForegroundColor()))
	form.Add(canvas.NewText(" ", theme.ForegroundColor()))
}

var iconApp = fyne.NewStaticResource("appIcon.png", []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x20, 0x08, 0x03, 0x00, 0x00, 0x00, 0x44, 0xA4, 0x8A, 0xC6, 0x00, 0x00, 0x00, 0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x00, 0xF1, 0x00, 0x00, 0x00, 0xF1, 0x01, 0x39, 0x0C, 0x0C, 0xE4, 0x00, 0x00, 0x00, 0x19, 0x74, 0x45, 0x58, 0x74, 0x53, 0x6F, 0x66, 0x74, 0x77, 0x61, 0x72, 0x65, 0x00, 0x77, 0x77, 0x77, 0x2E, 0x69, 0x6E, 0x6B, 0x73, 0x63, 0x61, 0x70, 0x65, 0x2E, 0x6F, 0x72, 0x67, 0x9B, 0xEE, 0x3C, 0x1A, 0x00, 0x00, 0x01, 0x14, 0x50, 0x4C, 0x54, 0x45, 0x47, 0x70, 0x4C, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFC, 0xFC, 0xFC, 0xFE, 0xFE, 0xFE, 0x36, 0x36, 0x36, 0x20, 0x20, 0x20, 0x2D, 0x2D, 0x2D, 0xFF, 0xFF, 0xFF, 0x1C, 0x1C, 0x1C, 0xD0, 0xD0, 0xD0, 0xFF, 0xFF, 0xFF, 0x4E, 0x4E, 0x4E, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x23, 0x23, 0x23, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFC, 0xFC, 0xFC, 0x26, 0x26, 0x26, 0x22, 0x22, 0x22, 0xF4, 0xF4, 0xF4, 0x3D, 0x3D, 0x3D, 0xFF, 0xFF, 0xFF, 0xD7, 0xD7, 0xD7, 0x2F, 0x2F, 0x2F, 0xBE, 0xBE, 0xBE, 0x96, 0x96, 0x96, 0xC6, 0xC6, 0xC6, 0x42, 0x42, 0x42, 0x32, 0x32, 0x32, 0x5F, 0x5F, 0x5F, 0xF2, 0xF2, 0xF2, 0x31, 0x31, 0x31, 0xDF, 0xDF, 0xDF, 0xC9, 0xC9, 0xC9, 0x14, 0x14, 0x14, 0x5C, 0x5C, 0x5C, 0x2C, 0x2C, 0x2C, 0xA1, 0xA1, 0xA1, 0x3C, 0x3C, 0x3C, 0x7C, 0x7C, 0x7C, 0xEB, 0xEB, 0xEB, 0x16, 0x16, 0x16, 0x72, 0x72, 0x72, 0x88, 0x88, 0x88, 0xF6, 0xF6, 0xF6, 0x6A, 0x6A, 0x6A, 0x46, 0x46, 0x46, 0x84, 0x84, 0x84, 0x60, 0x60, 0x60, 0x74, 0x74, 0x74, 0x1E, 0x1E, 0x1E, 0x09, 0x09, 0x09, 0x10, 0x10, 0x10, 0xE2, 0xE2, 0xE2, 0xCC, 0xCC, 0xCC, 0x19, 0x19, 0x19, 0xE0, 0xE0, 0xE0, 0x3A, 0x3A, 0x3A, 0x6B, 0x6B, 0x6B, 0x92, 0x92, 0x92, 0xB8, 0xB8, 0xB8, 0xA0, 0xA0, 0xA0, 0xA5, 0xA5, 0xA5, 0x87, 0x87, 0x87, 0x28, 0x28, 0x28, 0x49, 0x49, 0x49, 0xB6, 0xB6, 0xB6, 0xF9, 0xF9, 0xF9, 0x60, 0x60, 0x60, 0x87, 0x87, 0x87, 0x9D, 0x9D, 0x9D, 0x73, 0x73, 0x73, 0xAA, 0xAA, 0xAA, 0x38, 0x38, 0x38, 0xA0, 0xA0, 0xA0, 0xA2, 0xA2, 0xA2, 0x3A, 0x3A, 0x3A, 0x00, 0x00, 0x00, 0x04, 0x04, 0x04, 0x02, 0x02, 0x02, 0x0D, 0x0D, 0x0D, 0x1A, 0x1A, 0x1A, 0x15, 0x15, 0x15, 0x1F, 0x1F, 0x1F, 0x09, 0x09, 0x09, 0x25, 0x25, 0x25, 0x2B, 0x2B, 0x2B, 0x1F, 0x54, 0xFE, 0xF4, 0x00, 0x00, 0x00, 0x52, 0x74, 0x52, 0x4E, 0x53, 0x00, 0x02, 0x01, 0x0F, 0x04, 0x15, 0x30, 0x35, 0xD6, 0xEE, 0xE5, 0x11, 0xF1, 0x57, 0x0C, 0xC2, 0x25, 0x21, 0xB1, 0x07, 0x29, 0x3C, 0xEB, 0xE3, 0x47, 0xFE, 0x1B, 0x4F, 0xDF, 0x60, 0x79, 0x63, 0xCC, 0xFE, 0x9B, 0x59, 0xF4, 0x53, 0x4D, 0xF1, 0xB2, 0x97, 0x81, 0xB1, 0x96, 0x3C, 0xF8, 0xA8, 0x84, 0x8E, 0xB5, 0xDD, 0xB8, 0xDF, 0xBB, 0xD9, 0x9E, 0xF6, 0x47, 0x3E, 0xF2, 0x73, 0xC7, 0x76, 0xC5, 0xA2, 0x58, 0xAD, 0xC5, 0xC9, 0xF5, 0x77, 0x64, 0xCC, 0x6F, 0x70, 0xDF, 0xC1, 0x85, 0x4B, 0x92, 0xE1, 0x8B, 0x82, 0x42, 0x7F, 0x00, 0x00, 0x02, 0x7D, 0x49, 0x44, 0x41, 0x54, 0x38, 0xCB, 0x6D, 0x53, 0xE5, 0x9A, 0xE3, 0x30, 0x0C, 0x8C, 0xC3, 0xD0, 0xA6, 0x6D, 0xA0, 0xBC, 0x65, 0xEE, 0x32, 0x33, 0xF3, 0x1E, 0xDA, 0xB1, 0x43, 0xEF, 0xFF, 0x1E, 0xA7, 0x40, 0x7B, 0xBB, 0xF7, 0x9D, 0x7E, 0xC9, 0x9E, 0xB1, 0x6C, 0xCF, 0x48, 0x1C, 0x97, 0x07, 0x42, 0x22, 0xE2, 0xFF, 0x93, 0xE7, 0xC1, 0xDB, 0x46, 0x59, 0x71, 0x04, 0x94, 0xC2, 0x46, 0x45, 0x96, 0x9B, 0x36, 0xFA, 0x82, 0x1B, 0x4E, 0x47, 0x3D, 0xAB, 0x6D, 0x48, 0xB0, 0x8B, 0x9A, 0x13, 0xD5, 0x2B, 0x0C, 0x5C, 0x01, 0x7D, 0xC6, 0xB7, 0x4A, 0x04, 0x63, 0x5C, 0xD8, 0x90, 0x78, 0x4E, 0x98, 0x14, 0x20, 0xC5, 0x7A, 0xCA, 0xCE, 0xAF, 0x94, 0x5A, 0x6D, 0x9C, 0x86, 0x56, 0x94, 0xC4, 0x72, 0x2F, 0xCB, 0x53, 0x76, 0x8E, 0x77, 0x35, 0xD8, 0xF0, 0x08, 0x44, 0x5B, 0x91, 0x94, 0x29, 0xCE, 0x19, 0x2D, 0x23, 0xC3, 0x85, 0x62, 0x82, 0x13, 0xEA, 0xB3, 0x20, 0x08, 0x36, 0xDD, 0x2B, 0x4A, 0x72, 0x86, 0xBE, 0x25, 0x24, 0x35, 0x6C, 0x57, 0x4D, 0xCE, 0x87, 0x0C, 0x70, 0x36, 0x5A, 0xCE, 0xAF, 0x59, 0x98, 0x13, 0x70, 0xC9, 0xB1, 0xE1, 0x81, 0xCD, 0x4D, 0x9C, 0x11, 0xA0, 0x8A, 0xBA, 0x9C, 0x5B, 0xB5, 0x60, 0x4D, 0x20, 0x1D, 0x83, 0xE7, 0x90, 0xAC, 0xA7, 0x8B, 0xD0, 0x0F, 0xE2, 0xA3, 0xD7, 0x7A, 0xC7, 0xEA, 0xB3, 0xD5, 0x15, 0x18, 0xAB, 0x15, 0x9E, 0x13, 0x9D, 0xB3, 0x34, 0xF7, 0x48, 0x18, 0x1E, 0xD4, 0xDD, 0x79, 0x2D, 0xA6, 0x1E, 0x2C, 0xBC, 0x74, 0xF3, 0x52, 0x46, 0x1C, 0x2A, 0x97, 0xD6, 0x07, 0xE8, 0x6C, 0xD7, 0x3A, 0x18, 0x91, 0xE4, 0xC1, 0xD9, 0x4B, 0xD5, 0x0A, 0x68, 0x61, 0x9E, 0xEB, 0xAB, 0x12, 0xFE, 0xE1, 0xEC, 0xE9, 0x61, 0xEF, 0x30, 0x0C, 0x8E, 0xAA, 0x11, 0xD4, 0xC1, 0x7A, 0xCB, 0x84, 0x6F, 0xF0, 0xD2, 0xF9, 0x4C, 0x23, 0x1E, 0xA1, 0x2C, 0xB8, 0xFD, 0x4E, 0xA3, 0x46, 0xE3, 0x6E, 0x67, 0x67, 0x61, 0x1D, 0xFB, 0x1E, 0x6E, 0x8F, 0x05, 0x31, 0x15, 0xDA, 0x2C, 0x2B, 0xF5, 0x97, 0xC3, 0xE7, 0x60, 0xBB, 0x47, 0xE1, 0x1A, 0x36, 0xBA, 0xEA, 0xBE, 0x9F, 0x2E, 0x7E, 0x7B, 0xBD, 0xFA, 0xCA, 0x0E, 0x1E, 0xFC, 0x53, 0x6E, 0xAC, 0x13, 0xFD, 0x0D, 0xFB, 0x11, 0x9C, 0x7C, 0x1A, 0x0C, 0x4E, 0xEF, 0x37, 0xDE, 0xE5, 0xB5, 0x5D, 0xBC, 0xA1, 0xEC, 0x0E, 0x4E, 0x7F, 0x14, 0x5F, 0x8F, 0x71, 0x3F, 0x8E, 0xE2, 0xB8, 0x34, 0x3D, 0xB1, 0x5E, 0x94, 0x8A, 0x29, 0xAE, 0xBD, 0xAA, 0x1F, 0xC4, 0xD3, 0x81, 0x22, 0x77, 0x97, 0xDB, 0x31, 0x0D, 0xE9, 0x68, 0xEF, 0xD9, 0xA3, 0xB5, 0xFD, 0xA1, 0x5C, 0x31, 0x33, 0xDC, 0x28, 0xEA, 0x2C, 0xC0, 0x0F, 0xC3, 0xE1, 0xDB, 0xC9, 0x23, 0xFC, 0xB1, 0x70, 0xF7, 0xD8, 0xC7, 0x20, 0x16, 0xAB, 0x96, 0x2E, 0x12, 0xC7, 0x79, 0xA1, 0x58, 0x25, 0xCC, 0x27, 0x8D, 0xA3, 0x6F, 0x94, 0xDE, 0xEE, 0xFD, 0xEA, 0x0C, 0xF7, 0x63, 0x1F, 0x37, 0xFA, 0x61, 0x40, 0xBD, 0xEA, 0x04, 0xFC, 0x44, 0xE5, 0x36, 0x26, 0x8C, 0xE1, 0x46, 0x83, 0xE1, 0x88, 0x5D, 0x5E, 0x6F, 0xFD, 0x3C, 0x8E, 0x31, 0x8B, 0xFC, 0x54, 0xAA, 0x5E, 0x13, 0x94, 0x94, 0xB7, 0xB1, 0x47, 0x23, 0x10, 0x0A, 0xBE, 0x10, 0x7A, 0x33, 0x47, 0x5E, 0x14, 0xC2, 0x98, 0x92, 0x54, 0x6C, 0x9A, 0x48, 0x5D, 0x01, 0xB7, 0x09, 0x1C, 0xA1, 0x41, 0x44, 0x89, 0x36, 0x94, 0x6F, 0x2C, 0x2D, 0xF6, 0x73, 0xBF, 0x6A, 0x0E, 0x02, 0x21, 0x3B, 0xC0, 0x0D, 0x41, 0x47, 0x3F, 0x24, 0xDE, 0xFE, 0xFD, 0xDC, 0x52, 0xFB, 0xEB, 0x8E, 0xD8, 0x4C, 0x5E, 0x69, 0xCB, 0xA5, 0xBC, 0xE1, 0x80, 0xB8, 0xB8, 0x58, 0xEA, 0xB1, 0xBF, 0xC2, 0x55, 0xC5, 0x4E, 0xFE, 0x69, 0xD4, 0xF5, 0x75, 0x87, 0xB4, 0x9C, 0x5D, 0x7F, 0xDD, 0x72, 0x5A, 0x31, 0xD3, 0x92, 0x37, 0xC6, 0xDA, 0xAA, 0x4F, 0x15, 0xC1, 0x55, 0xBD, 0x7C, 0x51, 0xED, 0x1A, 0xB9, 0x96, 0xA2, 0xD4, 0x2D, 0xE4, 0x4D, 0x58, 0x46, 0xC2, 0xB8, 0x9A, 0xE5, 0xED, 0xB1, 0xB4, 0xD2, 0x3A, 0x61, 0xA4, 0x83, 0xA1, 0x4D, 0x4C, 0x30, 0x2E, 0x6B, 0xF2, 0x92, 0x6B, 0x7C, 0x1A, 0x2D, 0x51, 0x70, 0x77, 0x34, 0xD2, 0x6B, 0x25, 0xA3, 0x62, 0x1B, 0xC5, 0xF6, 0x54, 0xFD, 0x90, 0xCD, 0x2F, 0xC3, 0x8B, 0xCC, 0xA6, 0x22, 0x3B, 0x46, 0x3A, 0x4A, 0x22, 0xCC, 0x71, 0x45, 0xB2, 0xB9, 0x7F, 0x02, 0x46, 0x7E, 0x75, 0x25, 0x8F, 0xD0, 0xDF, 0xE9, 0xFF, 0x03, 0x82, 0x0B, 0xA8, 0x86, 0x86, 0x51, 0x26, 0xD1, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
})

var iconSettings = fyne.NewStaticResource("settingsIcon.png", []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0x40, 0x08, 0x03, 0x00, 0x00, 0x00, 0x9D, 0xB7, 0x81, 0xEC, 0x00, 0x00, 0x00, 0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x0E, 0xC3, 0x00, 0x00, 0x0E, 0xC3, 0x01, 0xC7, 0x6F, 0xA8, 0x64, 0x00, 0x00, 0x00, 0x19, 0x74, 0x45, 0x58, 0x74, 0x53, 0x6F, 0x66, 0x74, 0x77, 0x61, 0x72, 0x65, 0x00, 0x77, 0x77, 0x77, 0x2E, 0x69, 0x6E, 0x6B, 0x73, 0x63, 0x61, 0x70, 0x65, 0x2E, 0x6F, 0x72, 0x67, 0x9B, 0xEE, 0x3C, 0x1A, 0x00, 0x00, 0x00, 0x69, 0x50, 0x4C, 0x54, 0x45, 0x47, 0x70, 0x4C, 0x02, 0x02, 0x02, 0x40, 0x54, 0x5D, 0x0C, 0x0F, 0x0F, 0x25, 0x30, 0x36, 0x17, 0x1F, 0x23, 0x55, 0x69, 0x71, 0x52, 0x65, 0x6E, 0x56, 0x69, 0x72, 0x56, 0x69, 0x72, 0x44, 0x59, 0x64, 0x16, 0x1C, 0x1C, 0x43, 0x57, 0x61, 0x45, 0x59, 0x63, 0x54, 0x67, 0x72, 0x3D, 0x4F, 0x58, 0x36, 0x46, 0x4E, 0x57, 0x6B, 0x74, 0x3F, 0x3F, 0x49, 0x3F, 0x53, 0x5C, 0x55, 0x68, 0x71, 0x52, 0x66, 0x6F, 0x56, 0x69, 0x73, 0x38, 0x4A, 0x51, 0x45, 0x5A, 0x64, 0xFF, 0xFF, 0xFF, 0x4C, 0x60, 0x69, 0xF3, 0xF5, 0xF5, 0xDE, 0xE1, 0xE2, 0x3D, 0x4F, 0x57, 0x49, 0x5D, 0x66, 0x8B, 0x96, 0x9C, 0x9D, 0xA7, 0xAC, 0xC0, 0xC6, 0xC9, 0xC5, 0xCB, 0xCE, 0x71, 0x87, 0x44, 0x70, 0x00, 0x00, 0x00, 0x18, 0x74, 0x52, 0x4E, 0x53, 0x00, 0x21, 0xBE, 0x08, 0x52, 0x34, 0x8D, 0xFD, 0xFC, 0xAF, 0xF9, 0x12, 0xD6, 0xE9, 0x23, 0xA3, 0x77, 0x38, 0x03, 0xB1, 0xC7, 0x5F, 0xE4, 0x8E, 0xCD, 0xDC, 0x99, 0xA6, 0x00, 0x00, 0x02, 0x3E, 0x49, 0x44, 0x41, 0x54, 0x58, 0xC3, 0xED, 0x57, 0xD9, 0x82, 0xA3, 0x20, 0x10, 0x0C, 0x87, 0x02, 0xA2, 0xF1, 0x88, 0x26, 0x88, 0xA8, 0x49, 0xFE, 0xFF, 0x23, 0x37, 0x3B, 0x93, 0x89, 0x5C, 0x01, 0xE2, 0x3C, 0xEE, 0xF6, 0x2B, 0x56, 0xD9, 0x07, 0x5D, 0xDD, 0x1C, 0x0E, 0xFF, 0x2D, 0x6E, 0x47, 0x76, 0xA6, 0x67, 0x36, 0x1C, 0xF7, 0xE2, 0x07, 0x2A, 0x17, 0xB1, 0x48, 0x42, 0xFB, 0x76, 0x17, 0xBE, 0x25, 0xE2, 0x69, 0x92, 0x0E, 0x7B, 0x08, 0x7A, 0xB1, 0x19, 0xD9, 0x11, 0xC6, 0x51, 0x6A, 0x04, 0x22, 0x4F, 0xF0, 0x38, 0x67, 0x2C, 0xDF, 0x7E, 0xD4, 0x31, 0x1D, 0x2F, 0xCA, 0x18, 0xBC, 0xCB, 0x89, 0x94, 0x8B, 0xA4, 0x3F, 0x7F, 0x6A, 0x7B, 0xC3, 0x01, 0x51, 0xC5, 0xF0, 0xEC, 0xF9, 0x3D, 0x61, 0x6D, 0xC7, 0xDB, 0xBC, 0x27, 0x8B, 0xF8, 0xC4, 0x83, 0x17, 0xFE, 0x91, 0x71, 0x42, 0x29, 0x91, 0x16, 0x5C, 0x88, 0x53, 0xC4, 0x7F, 0x29, 0x22, 0x06, 0xC3, 0xF9, 0x8B, 0xE2, 0x85, 0x0C, 0xD6, 0x31, 0x8F, 0xE2, 0x45, 0xD1, 0x77, 0x01, 0x02, 0x18, 0x27, 0x10, 0x32, 0xFF, 0x25, 0x41, 0x90, 0xE1, 0x22, 0x92, 0x18, 0x18, 0x7A, 0x57, 0x84, 0xAC, 0x48, 0x63, 0xA0, 0x8C, 0xF5, 0xE7, 0x73, 0xEE, 0xE6, 0x13, 0xA5, 0xB9, 0xF0, 0xA0, 0xF8, 0xEA, 0x6F, 0x4F, 0x6F, 0xF2, 0xD2, 0xFA, 0x92, 0xAC, 0x77, 0x35, 0x4D, 0xEA, 0xBE, 0x12, 0x1F, 0x0F, 0x71, 0xF5, 0xA1, 0x31, 0x7F, 0x74, 0x9D, 0xC6, 0xA7, 0x4D, 0x57, 0xDF, 0x1D, 0xE9, 0x1D, 0x02, 0xAC, 0x1F, 0x53, 0x35, 0x6A, 0xA6, 0xA8, 0x4B, 0xB0, 0x38, 0x69, 0x18, 0x74, 0xFC, 0x34, 0x1A, 0x36, 0x79, 0x18, 0x86, 0x40, 0x08, 0x52, 0x8D, 0x96, 0x29, 0x37, 0x0A, 0xA7, 0xB7, 0xEA, 0xED, 0xEC, 0x3A, 0x3A, 0x76, 0x75, 0x08, 0x1C, 0x79, 0xD8, 0xAA, 0x40, 0x26, 0x97, 0x60, 0x72, 0x6A, 0x51, 0xD8, 0xFA, 0xB7, 0xF5, 0xFF, 0x3A, 0x7A, 0x6C, 0x8D, 0x79, 0xA0, 0xB5, 0xE3, 0xDD, 0x47, 0x70, 0x0B, 0xE7, 0xE0, 0xA8, 0xEB, 0x89, 0xF2, 0x11, 0x28, 0x3B, 0x82, 0x4C, 0x57, 0x13, 0x46, 0xF5, 0x2C, 0x4F, 0x3E, 0x82, 0xC9, 0xC2, 0x37, 0xC8, 0x98, 0x5F, 0xC6, 0x61, 0x02, 0x01, 0xC4, 0x7C, 0x6B, 0xC4, 0xA3, 0x9D, 0xE0, 0x68, 0x08, 0x25, 0x06, 0x7A, 0x5B, 0x33, 0x3B, 0x3D, 0xB7, 0x48, 0x12, 0xAB, 0xCC, 0x54, 0x85, 0xCA, 0x26, 0x88, 0x95, 0xB1, 0xB1, 0xB4, 0xD1, 0x91, 0x12, 0xE9, 0xBB, 0x48, 0xDB, 0x35, 0xA9, 0xB8, 0x75, 0x03, 0x1C, 0x0F, 0x7C, 0x57, 0x79, 0x0D, 0x34, 0x81, 0x47, 0x4F, 0x55, 0x28, 0x85, 0x38, 0xAC, 0x24, 0xDF, 0xCD, 0x6E, 0x05, 0xA1, 0xB4, 0x39, 0x57, 0xD8, 0x11, 0x1C, 0x32, 0x9F, 0x64, 0xA9, 0xB7, 0xCD, 0xEC, 0x0E, 0x69, 0x54, 0x7A, 0x08, 0xE6, 0x75, 0x93, 0xB4, 0x75, 0x36, 0xBC, 0xE3, 0x61, 0x31, 0x7B, 0x31, 0x88, 0xF5, 0xF6, 0x57, 0x54, 0x6F, 0xEB, 0x3C, 0x9B, 0x27, 0xCC, 0x75, 0xA1, 0xF1, 0x0E, 0x85, 0xF9, 0x69, 0x09, 0xF3, 0x89, 0xE3, 0xBA, 0xAC, 0x4A, 0x08, 0xD3, 0x86, 0x8B, 0x6F, 0xDD, 0x42, 0x20, 0xCB, 0x32, 0x00, 0x70, 0xD2, 0x84, 0x14, 0xC5, 0xDB, 0x11, 0xD9, 0x81, 0x53, 0x12, 0x43, 0x60, 0xD5, 0x48, 0x62, 0x58, 0x42, 0xBB, 0x0A, 0x48, 0x89, 0xA2, 0x0E, 0xAD, 0x2A, 0x29, 0x63, 0x3A, 0x0B, 0x6E, 0x5B, 0x71, 0x17, 0x6A, 0x14, 0x5C, 0xB7, 0xA2, 0x59, 0x80, 0x20, 0xBC, 0x30, 0x3A, 0x17, 0xBB, 0xAC, 0x1B, 0x9C, 0x6D, 0x15, 0x8E, 0xE1, 0x6D, 0x79, 0x80, 0x38, 0x03, 0x1C, 0x75, 0x08, 0x5C, 0xBE, 0x93, 0x53, 0xC7, 0xF0, 0x96, 0x07, 0xA7, 0x97, 0x7C, 0x3E, 0xAE, 0x2B, 0x84, 0x35, 0x8E, 0xE2, 0xF5, 0x21, 0x6B, 0x26, 0xBC, 0x03, 0x0F, 0x43, 0xF1, 0x9D, 0xDF, 0xA8, 0x23, 0x44, 0x1F, 0x3F, 0x32, 0x3A, 0xA4, 0x29, 0x54, 0x81, 0xF7, 0xBC, 0x73, 0x38, 0x2E, 0x7D, 0xF3, 0xEB, 0x13, 0x06, 0x80, 0x2F, 0xB0, 0x2A, 0xAA, 0x13, 0xE6, 0x7B, 0x5F, 0x7B, 0x88, 0x7F, 0xF5, 0x37, 0x3A, 0xFC, 0x03, 0xF6, 0x07, 0x9F, 0x23, 0xBD, 0x62, 0x36, 0x45, 0x9D, 0xA8, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
})
