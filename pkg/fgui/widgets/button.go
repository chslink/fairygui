package widgets

import "github.com/chslink/fairygui/pkg/fgui/assets"
import "github.com/chslink/fairygui/pkg/fgui/core"

// GButton represents a minimal button widget.
type GButton struct {
	*core.GComponent
	packageItem *assets.PackageItem
	resource    string
	title       string
	icon        string
	iconItem    *assets.PackageItem
}

// NewButton creates a button widget.
func NewButton() *GButton {
	return &GButton{GComponent: core.NewGComponent()}
}

// SetPackageItem stores the underlying package item representing this button template.
func (b *GButton) SetPackageItem(item *assets.PackageItem) {
	b.packageItem = item
}

// PackageItem returns the associated package item, if any.
func (b *GButton) PackageItem() *assets.PackageItem {
	return b.packageItem
}

// SetResource captures the raw resource identifier found on the component child.
func (b *GButton) SetResource(res string) {
	b.resource = res
}

// Resource returns the stored resource identifier (usually the package item id or name).
func (b *GButton) Resource() string {
	return b.resource
}

// SetTitle records the button title.
func (b *GButton) SetTitle(value string) {
	b.title = value
}

// Title returns the stored title text.
func (b *GButton) Title() string {
	return b.title
}

// SetIcon stores the icon resource identifier.
func (b *GButton) SetIcon(value string) {
	b.icon = value
}

// Icon returns the icon resource identifier.
func (b *GButton) Icon() string {
	return b.icon
}

// SetIconItem stores the resolved icon package item.
func (b *GButton) SetIconItem(item *assets.PackageItem) {
	b.iconItem = item
}

// IconItem returns the resolved icon package item.
func (b *GButton) IconItem() *assets.PackageItem {
	return b.iconItem
}
