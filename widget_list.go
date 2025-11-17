package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// List - 简化的列表控件
// ============================================================================

// List 是简化的列表控件，包装了 pkg/fgui/widgets.GList。
type List struct {
	list *widgets.GList
}

// ListSelectionMode 定义列表选择模式。
type ListSelectionMode int

const (
	// ListSelectionModeSingle 单选模式
	ListSelectionModeSingle ListSelectionMode = iota
	// ListSelectionModeMultiple 多选模式（Ctrl/Shift 辅助）
	ListSelectionModeMultiple
	// ListSelectionModeMultipleSingleClick 多选模式（单击切换）
	ListSelectionModeMultipleSingleClick
	// ListSelectionModeNone 不可选择
	ListSelectionModeNone
)

// ListLayoutType 定义列表布局类型。
type ListLayoutType int

const (
	// ListLayoutTypeSingleColumn 单列布局
	ListLayoutTypeSingleColumn ListLayoutType = iota
	// ListLayoutTypeSingleRow 单行布局
	ListLayoutTypeSingleRow
	// ListLayoutTypeFlowHorizontal 水平流式布局
	ListLayoutTypeFlowHorizontal
	// ListLayoutTypeFlowVertical 垂直流式布局
	ListLayoutTypeFlowVertical
	// ListLayoutTypePagination 分页布局
	ListLayoutTypePagination
)

// NewList 创建一个新的列表控件。
//
// 示例：
//
//	list := fairygui.NewList()
//	list.SetSelectionMode(fairygui.ListSelectionModeSingle)
func NewList() *List {
	return &List{
		list: widgets.NewList(),
	}
}

// Items 返回列表中的所有项。
func (l *List) Items() []*core.GObject {
	return l.list.Items()
}

// SelectedIndex 返回选中项的索引（-1 表示无选择）。
func (l *List) SelectedIndex() int {
	return l.list.SelectedIndex()
}

// SetSelectedIndex 设置选中项的索引。
//
// 示例：
//
//	list.SetSelectedIndex(0)  // 选中第一项
func (l *List) SetSelectedIndex(index int) {
	l.list.SetSelectedIndex(index)
}

// SelectedItem 返回选中的项对象。
func (l *List) SelectedItem() *core.GObject {
	return l.list.SelectedItem()
}

// SelectedIndices 返回所有选中项的索引（多选模式）。
func (l *List) SelectedIndices() []int {
	return l.list.SelectedIndices()
}

// SetSelectedIndices 设置多个选中项的索引（多选模式）。
func (l *List) SetSelectedIndices(indices []int) {
	l.list.SetSelectedIndices(indices)
}

// IsSelected 返回指定索引的项是否被选中。
func (l *List) IsSelected(index int) bool {
	return l.list.IsSelected(index)
}

// AddSelection 添加选中项（多选模式）。
//
// scrollItToView: 是否滚动到视图中。
func (l *List) AddSelection(index int, scrollItToView bool) {
	l.list.AddSelection(index, scrollItToView)
}

// RemoveSelection 移除选中项（多选模式）。
func (l *List) RemoveSelection(index int) {
	l.list.RemoveSelection(index)
}

// ClearSelection 清除所有选择。
func (l *List) ClearSelection() {
	l.list.ClearSelection()
}

// ScrollToView 滚动列表以显示指定索引的项。
//
// 示例：
//
//	list.ScrollToView(10)  // 滚动到第 10 项
func (l *List) ScrollToView(index int) {
	l.list.ScrollToView(index)
}

// SetSelectionMode 设置选择模式。
//
// 示例：
//
//	list.SetSelectionMode(fairygui.ListSelectionModeMultiple)
func (l *List) SetSelectionMode(mode ListSelectionMode) {
	l.list.SetSelectionMode(widgets.ListSelectionMode(mode))
}

// NumItems 返回虚拟列表的数据项总数。
func (l *List) NumItems() int {
	return l.list.NumItems()
}

// SetNumItems 设置虚拟列表的数据项总数。
//
// 示例：
//
//	list.SetNumItems(100)  // 设置 100 个数据项
func (l *List) SetNumItems(count int) {
	l.list.SetNumItems(count)
}

// SetItemRenderer 设置项目渲染器（虚拟列表）。
//
// 示例：
//
//	list.SetItemRenderer(func(index int, item *core.GObject) {
//	    // 更新 item 显示 index 对应的数据
//	})
func (l *List) SetItemRenderer(renderer func(index int, item *core.GObject)) {
	l.list.SetItemRenderer(renderer)
}

// SetItemProvider 设置项目提供者（虚拟列表）。
//
// 示例：
//
//	list.SetItemProvider(func(index int) string {
//	    return "ui://Main/ItemA"  // 返回项目资源 URL
//	})
func (l *List) SetItemProvider(provider func(index int) string) {
	l.list.SetItemProvider(provider)
}

// RefreshVirtualList 刷新虚拟列表。
func (l *List) RefreshVirtualList() {
	l.list.RefreshVirtualList()
}

// Layout 返回列表布局类型。
func (l *List) Layout() ListLayoutType {
	return ListLayoutType(l.list.Layout())
}

// LineGap 返回行间距（像素）。
func (l *List) LineGap() int {
	return l.list.LineGap()
}

// ColumnGap 返回列间距（像素）。
func (l *List) ColumnGap() int {
	return l.list.ColumnGap()
}

// LineCount 返回列表行数。
func (l *List) LineCount() int {
	return l.list.LineCount()
}

// ColumnCount 返回列表列数。
func (l *List) ColumnCount() int {
	return l.list.ColumnCount()
}

// Position 返回列表位置。
func (l *List) Position() (x, y float64) {
	return l.list.X(), l.list.Y()
}

// SetPosition 设置列表位置。
func (l *List) SetPosition(x, y float64) {
	l.list.SetPosition(x, y)
}

// Size 返回列表大小。
func (l *List) Size() (width, height float64) {
	return l.list.Width(), l.list.Height()
}

// SetSize 设置列表大小。
func (l *List) SetSize(width, height float64) {
	l.list.SetSize(width, height)
}

// Visible 返回列表是否可见。
func (l *List) Visible() bool {
	return l.list.Visible()
}

// SetVisible 设置列表可见性。
func (l *List) SetVisible(visible bool) {
	l.list.SetVisible(visible)
}

// Name 返回列表名称。
func (l *List) Name() string {
	return l.list.Name()
}

// SetName 设置列表名称。
func (l *List) SetName(name string) {
	l.list.SetName(name)
}

// Alpha 返回列表透明度（0-1）。
func (l *List) Alpha() float64 {
	return l.list.Alpha()
}

// SetAlpha 设置列表透明度（0-1）。
func (l *List) SetAlpha(alpha float64) {
	l.list.SetAlpha(alpha)
}

// RawList 返回底层的 widgets.GList 对象。
//
// 仅在需要访问底层 API 时使用。
func (l *List) RawList() *widgets.GList {
	return l.list
}
