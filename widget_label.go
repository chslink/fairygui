package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Label - 简化的标签控件
// ============================================================================

// Label 是简化的标签控件，包装了 pkg/fgui/widgets.GLabel。
//
// Label 组合了文本和可选图标，是一个简单的显示控件。
type Label struct {
	label *widgets.GLabel
}

// NewLabel 创建一个新的标签控件。
//
// 示例：
//
//	label := fairygui.NewLabel()
//	label.SetTitle("Hello World")
//	label.SetIcon("ui://package/icon")
func NewLabel() *Label {
	return &Label{
		label: widgets.NewLabel(),
	}
}

// Title 返回标签文本。
func (l *Label) Title() string {
	return l.label.Title()
}

// SetTitle 设置标签文本。
//
// 示例：
//
//	label.SetTitle("Hello World")
func (l *Label) SetTitle(title string) {
	l.label.SetTitle(title)
}

// Icon 返回图标 URL。
func (l *Label) Icon() string {
	return l.label.Icon()
}

// SetIcon 设置图标 URL。
//
// 示例：
//
//	label.SetIcon("ui://package/icon")
func (l *Label) SetIcon(icon string) {
	l.label.SetIcon(icon)
}

// TitleColor 返回标题颜色（十六进制格式）。
func (l *Label) TitleColor() string {
	return l.label.TitleColor()
}

// SetTitleColor 设置标题颜色（十六进制格式）。
//
// 示例：
//
//	label.SetTitleColor("#FF0000")
func (l *Label) SetTitleColor(color string) {
	l.label.SetTitleColor(color)
}

// TitleOutlineColor 返回标题描边颜色（十六进制格式）。
func (l *Label) TitleOutlineColor() string {
	return l.label.TitleOutlineColor()
}

// SetTitleOutlineColor 设置标题描边颜色（十六进制格式）。
//
// 示例：
//
//	label.SetTitleOutlineColor("#000000")
func (l *Label) SetTitleOutlineColor(color string) {
	l.label.SetTitleOutlineColor(color)
}

// TitleFontSize 返回标题字号。
func (l *Label) TitleFontSize() int {
	return l.label.TitleFontSize()
}

// SetTitleFontSize 设置标题字号。
//
// 示例：
//
//	label.SetTitleFontSize(20)
func (l *Label) SetTitleFontSize(size int) {
	l.label.SetTitleFontSize(size)
}

// Position 返回标签位置。
func (l *Label) Position() (x, y float64) {
	return l.label.X(), l.label.Y()
}

// SetPosition 设置标签位置。
func (l *Label) SetPosition(x, y float64) {
	l.label.SetPosition(x, y)
}

// Size 返回标签大小。
func (l *Label) Size() (width, height float64) {
	return l.label.Width(), l.label.Height()
}

// SetSize 设置标签大小。
func (l *Label) SetSize(width, height float64) {
	l.label.SetSize(width, height)
}

// Visible 返回标签是否可见。
func (l *Label) Visible() bool {
	return l.label.Visible()
}

// SetVisible 设置标签可见性。
func (l *Label) SetVisible(visible bool) {
	l.label.SetVisible(visible)
}

// Name 返回标签名称。
func (l *Label) Name() string {
	return l.label.Name()
}

// SetName 设置标签名称。
func (l *Label) SetName(name string) {
	l.label.SetName(name)
}

// Alpha 返回标签透明度（0-1）。
func (l *Label) Alpha() float64 {
	return l.label.Alpha()
}

// SetAlpha 设置标签透明度（0-1）。
func (l *Label) SetAlpha(alpha float64) {
	l.label.SetAlpha(alpha)
}

// RawLabel 返回底层的 widgets.GLabel 对象。
//
// 仅在需要访问底层 API 时使用。
func (l *Label) RawLabel() *widgets.GLabel {
	return l.label
}
