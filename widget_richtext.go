package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// RichText - 简化的富文本控件
// ============================================================================

// RichText 是简化的富文本控件，包装了 pkg/fgui/widgets.GRichTextField。
//
// RichText 支持 HTML/UBB 格式的富文本显示，继承了 Text 控件的所有功能。
type RichText struct {
	richText *widgets.GRichTextField
}

// NewRichText 创建一个新的富文本控件。
//
// 示例：
//
//	richText := fairygui.NewRichText()
//	richText.SetText("[color=#FF0000]红色文本[/color]")
func NewRichText() *RichText {
	return &RichText{
		richText: widgets.NewRichText(),
	}
}

// Text 返回文本内容。
func (r *RichText) Text() string {
	return r.richText.Text()
}

// SetText 设置文本内容。
//
// 支持 UBB 标签格式化，如：
//   - [color=#FF0000]红色[/color]
//   - [size=20]大字[/size]
//   - [b]粗体[/b], [i]斜体[/i], [u]下划线[/u]
//
// 示例：
//
//	richText.SetText("[color=#FF0000]红色[/color][b]粗体[/b]")
func (r *RichText) SetText(text string) {
	r.richText.SetText(text)
}

// HtmlEnabled 返回是否启用 HTML 模式。
func (r *RichText) HtmlEnabled() bool {
	return r.richText.HtmlEnabled()
}

// SetHtmlEnabled 设置是否启用 HTML 模式。
//
// 示例：
//
//	richText.SetHtmlEnabled(true)
func (r *RichText) SetHtmlEnabled(enabled bool) {
	r.richText.SetHtmlEnabled(enabled)
}

// UBBEnabled 返回是否启用 UBB 格式。
func (r *RichText) UBBEnabled() bool {
	return r.richText.UBBEnabled()
}

// SetUBBEnabled 设置是否启用 UBB 格式。
func (r *RichText) SetUBBEnabled(enabled bool) {
	r.richText.SetUBBEnabled(enabled)
}

// Color 返回文本颜色（十六进制格式）。
func (r *RichText) Color() string {
	return r.richText.Color()
}

// SetColor 设置文本颜色（十六进制格式）。
func (r *RichText) SetColor(color string) {
	r.richText.SetColor(color)
}

// FontSize 返回字号。
func (r *RichText) FontSize() int {
	return r.richText.FontSize()
}

// SetFontSize 设置字号。
func (r *RichText) SetFontSize(size int) {
	r.richText.SetFontSize(size)
}

// StrokeSize 返回描边宽度。
func (r *RichText) StrokeSize() float64 {
	return r.richText.StrokeSize()
}

// SetStrokeSize 设置描边宽度。
func (r *RichText) SetStrokeSize(stroke float64) {
	r.richText.SetStrokeSize(stroke)
}

// StrokeColor 返回描边颜色。
func (r *RichText) StrokeColor() string {
	return r.richText.StrokeColor()
}

// SetStrokeColor 设置描边颜色。
func (r *RichText) SetStrokeColor(color string) {
	r.richText.SetStrokeColor(color)
}

// Position 返回富文本位置。
func (r *RichText) Position() (x, y float64) {
	return r.richText.X(), r.richText.Y()
}

// SetPosition 设置富文本位置。
func (r *RichText) SetPosition(x, y float64) {
	r.richText.SetPosition(x, y)
}

// Size 返回富文本大小。
func (r *RichText) Size() (width, height float64) {
	return r.richText.Width(), r.richText.Height()
}

// SetSize 设置富文本大小。
func (r *RichText) SetSize(width, height float64) {
	r.richText.SetSize(width, height)
}

// Visible 返回富文本是否可见。
func (r *RichText) Visible() bool {
	return r.richText.Visible()
}

// SetVisible 设置富文本可见性。
func (r *RichText) SetVisible(visible bool) {
	r.richText.SetVisible(visible)
}

// Name 返回富文本名称。
func (r *RichText) Name() string {
	return r.richText.Name()
}

// SetName 设置富文本名称。
func (r *RichText) SetName(name string) {
	r.richText.SetName(name)
}

// Alpha 返回富文本透明度（0-1）。
func (r *RichText) Alpha() float64 {
	return r.richText.Alpha()
}

// SetAlpha 设置富文本透明度（0-1）。
func (r *RichText) SetAlpha(alpha float64) {
	r.richText.SetAlpha(alpha)
}

// RawRichText 返回底层的 widgets.GRichTextField 对象。
//
// 仅在需要访问底层 API 时使用。
func (r *RichText) RawRichText() *widgets.GRichTextField {
	return r.richText
}
