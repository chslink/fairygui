package widgets

import "github.com/chslink/fairygui/pkg/fgui/core"

// GLoader represents a resource loader widget.
type GLoader struct {
	*core.GObject
}

// NewLoader creates a loader widget.
func NewLoader() *GLoader {
	return &GLoader{GObject: core.NewGObject()}
}
