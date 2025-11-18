package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// ScrollBar - 简化的滚动条控件
// ============================================================================

// ScrollBar 是简化的滚动条控件，包装了 pkg/fgui/widgets.GScrollBar。
//
// 注意：ScrollBar 通常由 ScrollPane 自动创建和管理，用户很少需要手动操作。
type ScrollBar struct {
	scrollbar *widgets.GScrollBar
}

// NewScrollBar 创建一个新的滚动条控件。
//
// 注意：通常不需要手动创建 ScrollBar，它会由 ScrollPane 自动管理。
//
// 示例：
//
//	scrollbar := fairygui.NewScrollBar()
func NewScrollBar() *ScrollBar {
	return &ScrollBar{
		scrollbar: widgets.NewScrollBar(),
	}
}

// Position 返回滚动条位置。
func (s *ScrollBar) Position() (x, y float64) {
	return s.scrollbar.X(), s.scrollbar.Y()
}

// SetPosition 设置滚动条位置。
func (s *ScrollBar) SetPosition(x, y float64) {
	s.scrollbar.SetPosition(x, y)
}

// Size 返回滚动条大小。
func (s *ScrollBar) Size() (width, height float64) {
	return s.scrollbar.Width(), s.scrollbar.Height()
}

// SetSize 设置滚动条大小。
func (s *ScrollBar) SetSize(width, height float64) {
	s.scrollbar.SetSize(width, height)
}

// Visible 返回滚动条是否可见。
func (s *ScrollBar) Visible() bool {
	return s.scrollbar.Visible()
}

// SetVisible 设置滚动条可见性。
func (s *ScrollBar) SetVisible(visible bool) {
	s.scrollbar.SetVisible(visible)
}

// Name 返回滚动条名称。
func (s *ScrollBar) Name() string {
	return s.scrollbar.Name()
}

// SetName 设置滚动条名称。
func (s *ScrollBar) SetName(name string) {
	s.scrollbar.SetName(name)
}

// Alpha 返回滚动条透明度（0-1）。
func (s *ScrollBar) Alpha() float64 {
	return s.scrollbar.Alpha()
}

// SetAlpha 设置滚动条透明度（0-1）。
func (s *ScrollBar) SetAlpha(alpha float64) {
	s.scrollbar.SetAlpha(alpha)
}

// RawScrollBar 返回底层的 widgets.GScrollBar 对象。
//
// 仅在需要访问底层 API 时使用。
func (s *ScrollBar) RawScrollBar() *widgets.GScrollBar {
	return s.scrollbar
}
