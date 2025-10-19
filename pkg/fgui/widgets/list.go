package widgets

import "github.com/chslink/fairygui/pkg/fgui/assets"
import "github.com/chslink/fairygui/pkg/fgui/core"

// GList represents a minimal list widget backed by a component package item.
type GList struct {
	*core.GComponent
	packageItem *assets.PackageItem
	defaultItem string
	resource    string
}

// NewList constructs an empty list widget.
func NewList() *GList {
	return &GList{GComponent: core.NewGComponent()}
}

// SetPackageItem stores the component package item used by this list.
func (l *GList) SetPackageItem(item *assets.PackageItem) {
	l.packageItem = item
}

// PackageItem returns the associated package item, if any.
func (l *GList) PackageItem() *assets.PackageItem {
	return l.packageItem
}

// SetDefaultItem records the default item resource id for this list.
func (l *GList) SetDefaultItem(value string) {
	l.defaultItem = value
}

// DefaultItem returns the default item resource id.
func (l *GList) DefaultItem() string {
	return l.defaultItem
}

// SetResource stores the raw resource identifier declared on the component child.
func (l *GList) SetResource(res string) {
	l.resource = res
}

// Resource returns the stored raw resource identifier.
func (l *GList) Resource() string {
	return l.resource
}
