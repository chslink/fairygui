package widgets

import "github.com/chslink/fairygui/pkg/fgui/core"

// GTextField is a minimal text widget.
type GTextField struct {
	*core.GObject
	text string
}

// NewText creates a new text field widget.
func NewText() *GTextField {
	return &GTextField{GObject: core.NewGObject()}
}

// SetText updates the displayed text.
func (t *GTextField) SetText(value string) {
	t.text = value
}

// Text returns the current text.
func (t *GTextField) Text() string {
	return t.text
}
