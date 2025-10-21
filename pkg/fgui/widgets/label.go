package widgets

import "github.com/chslink/fairygui/pkg/fgui/assets"
import "github.com/chslink/fairygui/pkg/fgui/core"

// GLabel represents a simple label widget (text + optional icon).
type GLabel struct {
	*core.GComponent
	packageItem       *assets.PackageItem
	template          *core.GComponent
	titleObject       *core.GObject
	iconObject        *core.GObject
	title             string
	icon              string
	iconItem          *assets.PackageItem
	resource          string
	titleColor        string
	titleOutlineColor string
	titleFontSize     int
}

// NewLabel constructs an empty label widget.
func NewLabel() *GLabel {
	label := &GLabel{GComponent: core.NewGComponent()}
	label.titleColor = "#ffffff"
	label.titleFontSize = 12
	label.GComponent.SetData(label)
	return label
}

// SetTitle updates the label text currently stored on the widget.
func (l *GLabel) SetTitle(value string) {
	l.title = value
	l.applyTitleState()
}

// Title returns the label text previously assigned.
func (l *GLabel) Title() string {
	return l.title
}

// SetIcon updates the icon reference (resource URL) for the label.
func (l *GLabel) SetIcon(value string) {
	l.icon = value
	l.applyIconState()
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

// SetPackageItem stores the source package item backing this label template.
func (l *GLabel) SetPackageItem(item *assets.PackageItem) {
	l.packageItem = item
}

// PackageItem returns the associated package item, if any.
func (l *GLabel) PackageItem() *assets.PackageItem {
	return l.packageItem
}

// SetTemplateComponent installs the template component rendered by this label.
func (l *GLabel) SetTemplateComponent(comp *core.GComponent) {
	if l.template != nil && l.GComponent != nil {
		l.GComponent.RemoveChild(l.template.GObject)
	}
	l.template = comp
	if comp != nil && l.GComponent != nil {
		comp.GObject.SetPosition(0, 0)
		l.GComponent.AddChild(comp.GObject)
	}
	l.applyTitleState()
	l.applyIconState()
	l.applyTitleFormatting()
}

// TemplateComponent returns the template component associated with this label.
func (l *GLabel) TemplateComponent() *core.GComponent {
	return l.template
}

// SetTitleObject caches the underlying title object reference.
func (l *GLabel) SetTitleObject(obj *core.GObject) {
	l.titleObject = obj
	l.applyTitleState()
	l.applyTitleFormatting()
}

// TitleObject returns the cached title object.
func (l *GLabel) TitleObject() *core.GObject {
	return l.titleObject
}

// SetIconObject caches the icon display object.
func (l *GLabel) SetIconObject(obj *core.GObject) {
	l.iconObject = obj
	l.applyIconState()
}

// IconObject returns the cached icon object.
func (l *GLabel) IconObject() *core.GObject {
	return l.iconObject
}

// SetTitleColor stores the label text colour.
func (l *GLabel) SetTitleColor(value string) {
	l.titleColor = value
	l.applyTitleFormatting()
}

// TitleColor returns the stored label text colour.
func (l *GLabel) TitleColor() string {
	return l.titleColor
}

// SetTitleOutlineColor stores the outline colour for the label text.
func (l *GLabel) SetTitleOutlineColor(value string) {
	l.titleOutlineColor = value
	l.applyTitleFormatting()
}

// TitleOutlineColor returns the outline colour for the label text.
func (l *GLabel) TitleOutlineColor() string {
	return l.titleOutlineColor
}

// SetTitleFontSize records the font size associated with the label text.
func (l *GLabel) SetTitleFontSize(size int) {
	l.titleFontSize = size
	l.applyTitleFormatting()
}

// TitleFontSize returns the stored font size.
func (l *GLabel) TitleFontSize() int {
	return l.titleFontSize
}

func (l *GLabel) applyTitleState() {
	if l.titleObject == nil {
		return
	}
	text := l.title
	switch data := l.titleObject.Data().(type) {
	case *GTextField:
		data.SetText(text)
	case *GLabel:
		data.SetTitle(text)
	case *GButton:
		data.SetTitle(text)
	case string:
		if data != text {
			l.titleObject.SetData(text)
		}
	case nil:
		l.titleObject.SetData(text)
	default:
		l.titleObject.SetData(text)
	}
}

func (l *GLabel) applyIconState() {
	if l.iconObject == nil {
		return
	}
	icon := l.icon
	switch data := l.iconObject.Data().(type) {
	case *GLoader:
		data.SetURL(icon)
	case *GButton:
		data.SetIcon(icon)
	case *GLabel:
		data.SetIcon(icon)
	case string:
		if data != icon {
			l.iconObject.SetData(icon)
		}
	case nil:
		l.iconObject.SetData(icon)
	default:
		l.iconObject.SetData(icon)
	}
}

func (l *GLabel) applyTitleFormatting() {
	if l.titleObject == nil {
		return
	}
	switch data := l.titleObject.Data().(type) {
	case *GTextField:
		if l.titleColor != "" {
			data.SetColor(l.titleColor)
		}
		if l.titleOutlineColor != "" {
			data.SetOutlineColor(l.titleOutlineColor)
		}
		if l.titleFontSize != 0 {
			data.SetFontSize(l.titleFontSize)
		}
	case *GLabel:
		if l.titleColor != "" {
			data.SetTitleColor(l.titleColor)
		}
		if l.titleOutlineColor != "" {
			data.SetTitleOutlineColor(l.titleOutlineColor)
		}
		if l.titleFontSize != 0 {
			data.SetTitleFontSize(l.titleFontSize)
		}
	case *GButton:
		if l.titleColor != "" {
			data.SetTitleColor(l.titleColor)
		}
		if l.titleOutlineColor != "" {
			data.SetTitleOutlineColor(l.titleOutlineColor)
		}
		if l.titleFontSize != 0 {
			data.SetTitleFontSize(l.titleFontSize)
		}
	default:
		// no-op for unsupported payloads
	}
}
