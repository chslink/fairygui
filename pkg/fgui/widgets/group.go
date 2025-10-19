package widgets

import "github.com/chslink/fairygui/pkg/fgui/core"

// GGroup is a simple container widget representing a group.
type GGroup struct {
	*core.GComponent
}

// NewGroup creates a new group widget.
func NewGroup() *GGroup {
	return &GGroup{GComponent: core.NewGComponent()}
}
