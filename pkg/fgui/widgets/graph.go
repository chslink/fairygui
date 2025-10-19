package widgets

import "github.com/chslink/fairygui/pkg/fgui/core"

// GGraph represents a simple graphic widget (placeholder).
type GGraph struct {
	*core.GObject
}

// NewGraph creates a new graphic widget.
func NewGraph() *GGraph {
	return &GGraph{GObject: core.NewGObject()}
}
