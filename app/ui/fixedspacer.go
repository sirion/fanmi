package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type FixedSpacer struct {
	widget.BaseWidget
	SpacerSize fyne.Size
}

func (sp *FixedSpacer) CreateRenderer() fyne.WidgetRenderer {
	// rect := canvas.NewRectangle(color.White)
	rect := canvas.NewRectangle(color.Transparent)
	rect.Resize(sp.SpacerSize)
	return widget.NewSimpleRenderer(rect)
}

// MinSize returns the minimum size this object needs to be drawn.
func (sp *FixedSpacer) MinSize() fyne.Size {
	return sp.SpacerSize
}

// Size returns the current size of this object.
func (sp *FixedSpacer) Size() fyne.Size {
	return sp.SpacerSize
}
