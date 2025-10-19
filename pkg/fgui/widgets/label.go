package widgets

import "github.com/chslink/fairygui/pkg/fgui/assets"
import "github.com/chslink/fairygui/pkg/fgui/core"

// GLabel represents a simple label widget (text + optional icon).
type GLabel struct {
	*core.GComponent
	title    string
	icon     string
	iconItem *assets.PackageItem
	resource string
}

// NewLabel constructs an empty label widget.
func NewLabel() *GLabel {
	return &GLabel{GComponent: core.NewGComponent()}
}

// SetTitle updates the label text currently stored on the widget.
func (l *GLabel) SetTitle(value string) {
	l.title = value
}

// Title returns the label text previously assigned.
func (l *GLabel) Title() string {
	return l.title
}

// SetIcon updates the icon reference (resource URL) for the label.
func (l *GLabel) SetIcon(value string) {
	l.icon = value
}

// Icon returns the icon resource reference.
func (l *GLabel) Icon() string {
	return l.icon
}

// SetIconItem stores the resolved package item representing the icon sprite.
func (l *GLabel) SetIconItem(item *assets.PackageItem) {
	l.iconItem = item
}

// IconItem returns the resolved icon package item, if any.
func (l *GLabel) IconItem() *assets.PackageItem {
	return l.iconItem
}

// SetResource stores the raw resource identifier associated with the label.
func (l *GLabel) SetResource(res string) {
	l.resource = res
}

// Resource returns the stored resource identifier.
func (l *GLabel) Resource() string {
	return l.resource
}
