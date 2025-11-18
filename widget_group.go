package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Group - 简化的分组控件
// ============================================================================

// Group 是简化的分组控件，包装了 pkg/fgui/widgets.GGroup。
//
// Group 是一个逻辑分组控件，不单独渲染，主要用于：
// - 布局管理（水平、垂直）
// - 成员可见性同步
// - 成员位置同步
type Group struct {
	group *widgets.GGroup
}

// GroupLayoutType 布局类型。
type GroupLayoutType int

const (
	// GroupLayoutTypeNone 无布局。
	GroupLayoutTypeNone GroupLayoutType = GroupLayoutType(widgets.GroupLayoutTypeNone)
	// GroupLayoutTypeHorizontal 水平布局。
	GroupLayoutTypeHorizontal GroupLayoutType = GroupLayoutType(widgets.GroupLayoutTypeHorizontal)
	// GroupLayoutTypeVertical 垂直布局。
	GroupLayoutTypeVertical GroupLayoutType = GroupLayoutType(widgets.GroupLayoutTypeVertical)
)

// NewGroup 创建一个新的分组控件。
//
// 示例：
//
//	group := fairygui.NewGroup()
func NewGroup() *Group {
	return &Group{
		group: widgets.NewGroup(),
	}
}

// Layout 返回布局类型。
func (g *Group) Layout() GroupLayoutType {
	return GroupLayoutType(g.group.Layout())
}

// LineGap 返回行间距。
func (g *Group) LineGap() int {
	return g.group.LineGap()
}

// ColumnGap 返回列间距。
func (g *Group) ColumnGap() int {
	return g.group.ColumnGap()
}

// ExcludeInvisibles 返回是否排除不可见子元素。
func (g *Group) ExcludeInvisibles() bool {
	return g.group.ExcludeInvisibles()
}

// AutoSizeDisabled 返回自动尺寸是否被禁用。
func (g *Group) AutoSizeDisabled() bool {
	return g.group.AutoSizeDisabled()
}

// MainGridIndex 返回主要网格索引。
func (g *Group) MainGridIndex() int {
	return g.group.MainGridIndex()
}

// MainGridMinSize 返回主要网格最小尺寸。
func (g *Group) MainGridMinSize() int {
	return g.group.MainGridMinSize()
}

// Position 返回分组位置。
func (g *Group) Position() (x, y float64) {
	return g.group.X(), g.group.Y()
}

// SetPosition 设置分组位置。
//
// 注意：分组位置改变时，所有成员会跟随移动。
func (g *Group) SetPosition(x, y float64) {
	g.group.SetPosition(x, y)
}

// Size 返回分组大小。
func (g *Group) Size() (width, height float64) {
	return g.group.Width(), g.group.Height()
}

// SetSize 设置分组大小。
func (g *Group) SetSize(width, height float64) {
	g.group.SetSize(width, height)
}

// Visible 返回分组是否可见。
func (g *Group) Visible() bool {
	return g.group.Visible()
}

// SetVisible 设置分组可见性。
//
// 注意：分组可见性改变时，所有成员的可见性也会同步改变。
func (g *Group) SetVisible(visible bool) {
	g.group.SetVisible(visible)
}

// Name 返回分组名称。
func (g *Group) Name() string {
	return g.group.Name()
}

// SetName 设置分组名称。
func (g *Group) SetName(name string) {
	g.group.SetName(name)
}

// Alpha 返回分组透明度（0-1）。
func (g *Group) Alpha() float64 {
	return g.group.Alpha()
}

// SetAlpha 设置分组透明度（0-1）。
func (g *Group) SetAlpha(alpha float64) {
	g.group.SetAlpha(alpha)
}

// RawGroup 返回底层的 widgets.GGroup 对象。
//
// 仅在需要访问底层 API 时使用。
func (g *Group) RawGroup() *widgets.GGroup {
	return g.group
}
