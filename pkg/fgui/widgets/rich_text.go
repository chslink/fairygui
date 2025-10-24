package widgets

import (
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
	"github.com/chslink/fairygui/internal/compat/laya"
)

// GRichTextField is a rich text widget that supports HTML/UBB formatting.
// It extends GTextField with rich text capabilities and follows LayaAir behavior.
type GRichTextField struct {
	*GTextField
	htmlEnabled bool
}

// NewRichText creates a new rich text field widget.
func NewRichText() *GRichTextField {
	base := NewText()
	rich := &GRichTextField{
		GTextField:  base,
		htmlEnabled: true, // LayaAir sets html = true by default for rich text
	}

	// 设置为富文本模式（等同于 LayaAir 的 this._displayObject.html = true）
	rich.SetUBBEnabled(true)

	// 关键：重新绑定 GObject.Data() 到 GRichTextField 实例
	// 这样渲染器才能正确识别类型
	base.GObject.SetData(rich)

	return rich
}

// SetHtmlEnabled toggles HTML rendering mode.
// This corresponds to LayaAir's displayObject.html property.
func (r *GRichTextField) SetHtmlEnabled(value bool) {
	r.htmlEnabled = value
	r.SetUBBEnabled(value)
}

// HtmlEnabled reports whether HTML mode is active.
func (r *GRichTextField) HtmlEnabled() bool {
	return r.htmlEnabled
}

// SetText implements GObject interface with rich text support.
func (r *GRichTextField) SetText(value string) {
	r.GTextField.SetText(value)
}

// Text returns the current text content.
func (r *GRichTextField) Text() string {
	return r.GTextField.Text()
}

// SetAutoSize overrides GTextField to ensure rich text behaves correctly.
func (r *GRichTextField) SetAutoSize(value TextAutoSize) {
	r.GTextField.SetAutoSize(value)
}

// AutoSize returns the current autosize mode.
func (r *GRichTextField) AutoSize() TextAutoSize {
	return r.GTextField.AutoSize()
}

// SetSingleLine overrides GTextField to handle rich text single line mode.
func (r *GRichTextField) SetSingleLine(value bool) {
	r.GTextField.SetSingleLine(value)
}

// SingleLine reports whether single-line mode is active.
func (r *GRichTextField) SingleLine() bool {
	return r.GTextField.SingleLine()
}

// SetUBBEnabled overrides to ensure rich text always supports UBB/HTML.
func (r *GRichTextField) SetUBBEnabled(value bool) {
	r.GTextField.SetUBBEnabled(value)
}

// UBBEnabled reports whether UBB formatting is enabled.
func (r *GRichTextField) UBBEnabled() bool {
	return r.GTextField.UBBEnabled()
}

// WidthAutoSize reports whether width auto-size mode is active.
func (r *GRichTextField) WidthAutoSize() bool {
	return r.GTextField.WidthAutoSize()
}

// SetupBeforeAdd implements PackageItem interface for rich text specific setup.
func (r *GRichTextField) SetupBeforeAdd(ctx *SetupContext, buf *utils.ByteBuffer) {
	// Call parent setup first
	r.GTextField.SetupBeforeAdd(ctx, buf)

	// Rich text specific setup
	r.htmlEnabled = true // Always enabled for rich text
	r.SetUBBEnabled(true)
}

// SetVisible implements GObject interface.
func (r *GRichTextField) SetVisible(value bool) {
	r.GTextField.SetVisible(value)
}

// Visible implements GObject interface.
func (r *GRichTextField) Visible() bool {
	return r.GTextField.Visible()
}

// GObject returns the underlying GObject for compatibility.
func (r *GRichTextField) GObject() *core.GObject {
	return r.GTextField.GObject
}

// DisplayObject returns the display object for rich text.
func (r *GRichTextField) DisplayObject() *laya.Sprite {
	return r.GTextField.DisplayObject()
}