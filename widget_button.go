package fairygui

import (
	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Button - 简化的按钮控件
// ============================================================================

// Button 是简化的按钮控件，包装了 pkg/fgui/widgets.GButton。
type Button struct {
	btn *widgets.GButton
}

// ButtonMode 定义按钮的行为模式。
type ButtonMode int

const (
	// ButtonModeCommon 普通按钮模式（点击触发事件）
	ButtonModeCommon ButtonMode = iota
	// ButtonModeCheck 复选框模式（可切换选中状态）
	ButtonModeCheck
	// ButtonModeRadio 单选框模式（单选组）
	ButtonModeRadio
)

// NewButton 创建一个新的按钮。
//
// 示例：
//
//	btn := fairygui.NewButton()
//	btn.SetTitle("Click Me")
//	btn.OnClick(func() {
//	    fmt.Println("Button clicked!")
//	})
func NewButton() *Button {
	return &Button{
		btn: widgets.NewButton(),
	}
}

// Title 返回按钮标题。
func (b *Button) Title() string {
	return b.btn.Title()
}

// SetTitle 设置按钮标题。
func (b *Button) SetTitle(title string) {
	b.btn.SetTitle(title)
}

// Icon 返回按钮图标资源 URL。
func (b *Button) Icon() string {
	return b.btn.Icon()
}

// SetIcon 设置按钮图标资源 URL。
//
// 图标 URL 格式: ui://packageName/iconName
func (b *Button) SetIcon(icon string) {
	b.btn.SetIcon(icon)
}

// Selected 返回按钮是否被选中（仅在 Check 或 Radio 模式有效）。
func (b *Button) Selected() bool {
	return b.btn.Selected()
}

// SetSelected 设置按钮选中状态（仅在 Check 或 Radio 模式有效）。
func (b *Button) SetSelected(selected bool) {
	b.btn.SetSelected(selected)
}

// Mode 返回按钮模式。
func (b *Button) Mode() ButtonMode {
	return ButtonMode(b.btn.Mode())
}

// SetMode 设置按钮模式。
//
// 示例：
//
//	btn.SetMode(fairygui.ButtonModeCheck) // 设置为复选框模式
func (b *Button) SetMode(mode ButtonMode) {
	b.btn.SetMode(widgets.ButtonMode(mode))
}

// OnClick 注册点击事件处理器。
//
// 示例：
//
//	btn.OnClick(func() {
//	    fmt.Println("Clicked!")
//	})
func (b *Button) OnClick(handler func()) {
	b.btn.GObject.On(laya.EventType("Click"), func(evt *laya.Event) {
		handler()
	})
}

// Enabled 返回按钮是否启用。
func (b *Button) Enabled() bool {
	return !b.btn.Grayed()
}

// SetEnabled 设置按钮启用状态。
//
// 禁用的按钮不响应交互事件。
func (b *Button) SetEnabled(enabled bool) {
	b.btn.SetGrayed(!enabled)
}

// Sound 返回按钮点击音效资源 URL。
func (b *Button) Sound() string {
	return b.btn.Sound()
}

// SetSound 设置按钮点击音效资源 URL。
func (b *Button) SetSound(sound string) {
	b.btn.SetSound(sound)
}

// Position 返回按钮位置。
func (b *Button) Position() (x, y float64) {
	return b.btn.X(), b.btn.Y()
}

// SetPosition 设置按钮位置。
func (b *Button) SetPosition(x, y float64) {
	b.btn.SetPosition(x, y)
}

// Size 返回按钮大小。
func (b *Button) Size() (width, height float64) {
	return b.btn.Width(), b.btn.Height()
}

// SetSize 设置按钮大小。
func (b *Button) SetSize(width, height float64) {
	b.btn.SetSize(width, height)
}

// Visible 返回按钮是否可见。
func (b *Button) Visible() bool {
	return b.btn.Visible()
}

// SetVisible 设置按钮可见性。
func (b *Button) SetVisible(visible bool) {
	b.btn.SetVisible(visible)
}

// Name 返回按钮名称。
func (b *Button) Name() string {
	return b.btn.Name()
}

// SetName 设置按钮名称。
func (b *Button) SetName(name string) {
	b.btn.SetName(name)
}

// RawButton 返回底层的 widgets.GButton 对象。
//
// 仅在需要访问底层 API 时使用。
func (b *Button) RawButton() *widgets.GButton {
	return b.btn
}

// ============================================================================
// 便捷构造函数
// ============================================================================

// NewCheckButton 创建一个复选框按钮。
//
// 示例：
//
//	checkBox := fairygui.NewCheckButton("Enable feature")
func NewCheckButton(title string) *Button {
	btn := NewButton()
	btn.SetTitle(title)
	btn.SetMode(ButtonModeCheck)
	return btn
}

// NewRadioButton 创建一个单选框按钮。
//
// 示例：
//
//	radio1 := fairygui.NewRadioButton("Option 1")
//	radio2 := fairygui.NewRadioButton("Option 2")
func NewRadioButton(title string) *Button {
	btn := NewButton()
	btn.SetTitle(title)
	btn.SetMode(ButtonModeRadio)
	return btn
}
