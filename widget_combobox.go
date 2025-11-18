package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// ComboBox - 简化的下拉框控件
// ============================================================================

// ComboBox 是简化的下拉框控件，包装了 pkg/fgui/widgets.GComboBox。
type ComboBox struct {
	combo *widgets.GComboBox
}

// PopupDirection 下拉方向。
type PopupDirection int

const (
	// PopupDirectionAuto 自动方向。
	PopupDirectionAuto PopupDirection = PopupDirection(widgets.PopupDirectionAuto)
	// PopupDirectionUp 向上。
	PopupDirectionUp PopupDirection = PopupDirection(widgets.PopupDirectionUp)
	// PopupDirectionDown 向下。
	PopupDirectionDown PopupDirection = PopupDirection(widgets.PopupDirectionDown)
)

// NewComboBox 创建一个新的下拉框控件。
//
// 示例：
//
//	combo := fairygui.NewComboBox()
//	combo.SetItems([]string{"选项1", "选项2", "选项3"}, nil, nil)
//	combo.SetSelectedIndex(0)
func NewComboBox() *ComboBox {
	return &ComboBox{
		combo: widgets.NewComboBox(),
	}
}

// Items 返回选项列表。
func (c *ComboBox) Items() []string {
	return c.combo.Items()
}

// NumItems 返回选项数量。
func (c *ComboBox) NumItems() int {
	return c.combo.NumItems()
}

// Values 返回选项值列表。
func (c *ComboBox) Values() []string {
	return c.combo.Values()
}

// Icons 返回选项图标列表。
func (c *ComboBox) Icons() []string {
	return c.combo.Icons()
}

// SetItems 设置选项列表。
//
// 参数:
//   - items: 显示文本列表
//   - values: 值列表（可选，传 nil 则使用 items）
//   - icons: 图标列表（可选，传 nil 则无图标）
//
// 示例：
//
//	combo.SetItems([]string{"选项1", "选项2"}, nil, nil)
//	combo.SetItems(
//	    []string{"苹果", "香蕉"},
//	    []string{"apple", "banana"},
//	    []string{"ui://pkg/icon1", "ui://pkg/icon2"},
//	)
func (c *ComboBox) SetItems(items, values, icons []string) {
	c.combo.SetItems(items, values, icons)
}

// SelectedIndex 返回当前选中的索引。
func (c *ComboBox) SelectedIndex() int {
	return c.combo.SelectedIndex()
}

// SetSelectedIndex 设置选中的索引。
//
// 示例：
//
//	combo.SetSelectedIndex(0)  // 选中第一项
func (c *ComboBox) SetSelectedIndex(index int) {
	c.combo.SetSelectedIndex(index)
}

// Value 返回当前选中项的值。
//
// 如果未设置值列表，返回空字符串。
func (c *ComboBox) Value() string {
	return c.combo.Value()
}

// Text 返回当前显示的文本。
func (c *ComboBox) Text() string {
	return c.combo.Text()
}

// SetText 设置显示的文本。
func (c *ComboBox) SetText(text string) {
	c.combo.SetText(text)
}

// Icon 返回当前显示的图标。
func (c *ComboBox) Icon() string {
	return c.combo.Icon()
}

// SetIcon 设置显示的图标。
func (c *ComboBox) SetIcon(icon string) {
	c.combo.SetIcon(icon)
}

// VisibleItemCount 返回下拉框显示的最大项数。
func (c *ComboBox) VisibleItemCount() int {
	return c.combo.VisibleItemCount()
}

// SetVisibleItemCount 设置下拉框显示的最大项数。
//
// 示例：
//
//	combo.SetVisibleItemCount(5)  // 最多显示5个项目
func (c *ComboBox) SetVisibleItemCount(count int) {
	c.combo.SetVisibleItemCount(count)
}

// PopupDirection 返回下拉方向。
func (c *ComboBox) PopupDirection() PopupDirection {
	return PopupDirection(c.combo.PopupDirection())
}

// SetPopupDirection 设置下拉方向。
//
// 示例：
//
//	combo.SetPopupDirection(fairygui.PopupDirectionUp)
func (c *ComboBox) SetPopupDirection(dir PopupDirection) {
	c.combo.SetPopupDirection(widgets.PopupDirection(dir))
}

// TitleColor 返回标题颜色。
func (c *ComboBox) TitleColor() string {
	return c.combo.TitleColor()
}

// SetTitleColor 设置标题颜色。
//
// 示例：
//
//	combo.SetTitleColor("#FF0000")
func (c *ComboBox) SetTitleColor(color string) {
	c.combo.SetTitleColor(color)
}

// TitleFontSize 返回标题字号。
func (c *ComboBox) TitleFontSize() int {
	return c.combo.TitleFontSize()
}

// SetTitleFontSize 设置标题字号。
//
// 示例：
//
//	combo.SetTitleFontSize(20)
func (c *ComboBox) SetTitleFontSize(size int) {
	c.combo.SetTitleFontSize(size)
}

// TitleOutlineColor 返回标题描边颜色。
func (c *ComboBox) TitleOutlineColor() string {
	return c.combo.TitleOutlineColor()
}

// SetTitleOutlineColor 设置标题描边颜色。
//
// 示例：
//
//	combo.SetTitleOutlineColor("#000000")
func (c *ComboBox) SetTitleOutlineColor(color string) {
	c.combo.SetTitleOutlineColor(color)
}

// Position 返回下拉框位置。
func (c *ComboBox) Position() (x, y float64) {
	return c.combo.X(), c.combo.Y()
}

// SetPosition 设置下拉框位置。
func (c *ComboBox) SetPosition(x, y float64) {
	c.combo.SetPosition(x, y)
}

// Size 返回下拉框大小。
func (c *ComboBox) Size() (width, height float64) {
	return c.combo.Width(), c.combo.Height()
}

// SetSize 设置下拉框大小。
func (c *ComboBox) SetSize(width, height float64) {
	c.combo.SetSize(width, height)
}

// Visible 返回下拉框是否可见。
func (c *ComboBox) Visible() bool {
	return c.combo.Visible()
}

// SetVisible 设置下拉框可见性。
func (c *ComboBox) SetVisible(visible bool) {
	c.combo.SetVisible(visible)
}

// Name 返回下拉框名称。
func (c *ComboBox) Name() string {
	return c.combo.Name()
}

// SetName 设置下拉框名称。
func (c *ComboBox) SetName(name string) {
	c.combo.SetName(name)
}

// Alpha 返回下拉框透明度（0-1）。
func (c *ComboBox) Alpha() float64 {
	return c.combo.Alpha()
}

// SetAlpha 设置下拉框透明度（0-1）。
func (c *ComboBox) SetAlpha(alpha float64) {
	c.combo.SetAlpha(alpha)
}

// RawComboBox 返回底层的 widgets.GComboBox 对象。
//
// 仅在需要访问底层 API 时使用。
func (c *ComboBox) RawComboBox() *widgets.GComboBox {
	return c.combo
}
