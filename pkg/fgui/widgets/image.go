package widgets

import "github.com/chslink/fairygui/pkg/fgui/core"

// GImage is a minimal image widget built atop GObject.
type GImage struct {
	*core.GObject
}

// NewImage constructs a GImage.
func NewImage() *GImage {
	return &GImage{GObject: core.NewGObject()}
}
