package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Text - 简化的文本控件
// ============================================================================

// Text 是简化的文本控件，包装了 pkg/fgui/widgets.GTextField。
type Text struct {
	txt *widgets.GTextField
}

// NewText 创建一个新的文本控件。
//
// 示例：
//
//	txt := fairygui.NewText()
//	txt.SetText("Hello World")
//	txt.SetColor("#FF0000")
func NewText() *Text {
	return &Text{
		txt: widgets.NewText(),
	}
}

// Text 返回文本内容。
func (t *Text) Text() string {
	return t.txt.Text()
}

// SetText 设置文本内容。
//
// 示例：
//
//	txt.SetText("Hello World")
func (t *Text) SetText(text string) {
	t.txt.SetText(text)
}

// Color 返回文本颜色（十六进制格式）。
func (t *Text) Color() string {
	return t.txt.Color()
}

// SetColor 设置文本颜色（十六进制格式）。
//
// 示例：
//
//	txt.SetColor("#FF0000")  // 红色
//	txt.SetColor("#00FF00")  // 绿色
func (t *Text) SetColor(color string) {
	t.txt.SetColor(color)
}

// Font 返回字体名称。
func (t *Text) Font() string {
	return t.txt.Font()
}

// SetFont 设置字体名称。
func (t *Text) SetFont(font string) {
	t.txt.SetFont(font)
}

// FontSize 返回字体大小。
func (t *Text) FontSize() int {
	return t.txt.FontSize()
}

// SetFontSize 设置字体大小。
//
// 示例：
//
//	txt.SetFontSize(24)
func (t *Text) SetFontSize(size int) {
	t.txt.SetFontSize(size)
}

// Align 返回水平对齐方式。
func (t *Text) Align() widgets.TextAlign {
	return t.txt.Align()
}

// SetAlign 设置水平对齐方式。
//
// 示例：
//
//	txt.SetAlign("center")  // 居中对齐
func (t *Text) SetAlign(align widgets.TextAlign) {
	t.txt.SetAlign(align)
}

// VerticalAlign 返回垂直对齐方式。
func (t *Text) VerticalAlign() widgets.TextVerticalAlign {
	return t.txt.VerticalAlign()
}

// SetVerticalAlign 设置垂直对齐方式。
//
// 示例：
//
//	txt.SetVerticalAlign("middle")  // 垂直居中
func (t *Text) SetVerticalAlign(align widgets.TextVerticalAlign) {
	t.txt.SetVerticalAlign(align)
}

// AutoSize 返回自动调整大小模式。
func (t *Text) AutoSize() widgets.TextAutoSize {
	return t.txt.AutoSize()
}

// SetAutoSize 设置自动调整大小模式。
//
// 示例：
//
//	txt.SetAutoSize(widgets.TextAutoSizeBoth)
func (t *Text) SetAutoSize(autoSize widgets.TextAutoSize) {
	t.txt.SetAutoSize(autoSize)
}

// SingleLine 返回是否为单行模式。
func (t *Text) SingleLine() bool {
	return t.txt.SingleLine()
}

// SetSingleLine 设置单行模式。
func (t *Text) SetSingleLine(singleLine bool) {
	t.txt.SetSingleLine(singleLine)
}

// Bold 返回是否为粗体。
func (t *Text) Bold() bool {
	return t.txt.Bold()
}

// SetBold 设置粗体。
func (t *Text) SetBold(bold bool) {
	t.txt.SetBold(bold)
}

// Italic 返回是否为斜体。
func (t *Text) Italic() bool {
	return t.txt.Italic()
}

// SetItalic 设置斜体。
func (t *Text) SetItalic(italic bool) {
	t.txt.SetItalic(italic)
}

// Underline 返回是否显示下划线。
func (t *Text) Underline() bool {
	return t.txt.Underline()
}

// SetUnderline 设置下划线。
func (t *Text) SetUnderline(underline bool) {
	t.txt.SetUnderline(underline)
}

// LetterSpacing 返回字符间距（像素）。
func (t *Text) LetterSpacing() int {
	return t.txt.LetterSpacing()
}

// SetLetterSpacing 设置字符间距（像素）。
func (t *Text) SetLetterSpacing(spacing int) {
	t.txt.SetLetterSpacing(spacing)
}

// Leading 返回行间距（像素）。
func (t *Text) Leading() int {
	return t.txt.Leading()
}

// SetLeading 设置行间距（像素）。
func (t *Text) SetLeading(leading int) {
	t.txt.SetLeading(leading)
}

// StrokeSize 返回描边大小。
func (t *Text) StrokeSize() float64 {
	return t.txt.StrokeSize()
}

// SetStrokeSize 设置描边大小。
func (t *Text) SetStrokeSize(size float64) {
	t.txt.SetStrokeSize(size)
}

// StrokeColor 返回描边颜色。
func (t *Text) StrokeColor() string {
	return t.txt.StrokeColor()
}

// SetStrokeColor 设置描边颜色。
func (t *Text) SetStrokeColor(color string) {
	t.txt.SetStrokeColor(color)
}

// UBBEnabled 返回是否启用 UBB 格式。
func (t *Text) UBBEnabled() bool {
	return t.txt.UBBEnabled()
}

// SetUBBEnabled 设置是否启用 UBB 格式。
func (t *Text) SetUBBEnabled(enabled bool) {
	t.txt.SetUBBEnabled(enabled)
}

// SetShadow 设置阴影效果。
//
// 示例：
//
//	txt.SetShadow("#000000", 2, 2, 4)
func (t *Text) SetShadow(color string, offsetX, offsetY, blur float64) {
	t.txt.SetShadow(color, offsetX, offsetY, blur)
}

// ShadowColor 返回阴影颜色。
func (t *Text) ShadowColor() string {
	return t.txt.ShadowColor()
}

// ShadowOffset 返回阴影偏移（X, Y）。
func (t *Text) ShadowOffset() (float64, float64) {
	return t.txt.ShadowOffset()
}

// ShadowBlur 返回阴影模糊半径。
func (t *Text) ShadowBlur() float64 {
	return t.txt.ShadowBlur()
}

// Position 返回文本位置。
func (t *Text) Position() (x, y float64) {
	return t.txt.X(), t.txt.Y()
}

// SetPosition 设置文本位置。
func (t *Text) SetPosition(x, y float64) {
	t.txt.SetPosition(x, y)
}

// Size 返回文本大小。
func (t *Text) Size() (width, height float64) {
	return t.txt.Width(), t.txt.Height()
}

// SetSize 设置文本大小。
func (t *Text) SetSize(width, height float64) {
	t.txt.SetSize(width, height)
}

// Visible 返回文本是否可见。
func (t *Text) Visible() bool {
	return t.txt.Visible()
}

// SetVisible 设置文本可见性。
func (t *Text) SetVisible(visible bool) {
	t.txt.SetVisible(visible)
}

// Name 返回文本名称。
func (t *Text) Name() string {
	return t.txt.Name()
}

// SetName 设置文本名称。
func (t *Text) SetName(name string) {
	t.txt.SetName(name)
}

// Alpha 返回文本透明度（0-1）。
func (t *Text) Alpha() float64 {
	return t.txt.Alpha()
}

// SetAlpha 设置文本透明度（0-1）。
func (t *Text) SetAlpha(alpha float64) {
	t.txt.SetAlpha(alpha)
}

// RawText 返回底层的 widgets.GTextField 对象。
//
// 仅在需要访问底层 API 时使用。
func (t *Text) RawText() *widgets.GTextField {
	return t.txt
}
