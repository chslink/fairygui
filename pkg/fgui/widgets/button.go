package widgets

import "github.com/chslink/fairygui/pkg/fgui/core"

// GButton represents a minimal button widget.
type GButton struct {
	*core.GComponent
}

// NewButton creates a button widget.
func NewButton() *GButton {
	return &GButton{GComponent: core.NewGComponent()}
}
