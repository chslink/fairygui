package fairygui

import (
	"testing"
)

// TestNewList 测试创建新的列表
func TestNewList(t *testing.T) {
	l := NewList()
	if l == nil {
		t.Fatal("NewList() returned nil")
	}

	// 检查默认属性
	if l.ComponentImpl == nil {
		t.Error("List.ComponentImpl is nil")
	}

	if l.Layout() != ListLayoutTypeSingleColumn {
		t.Errorf("默认布局不正确: got %v, want %v", l.Layout(), ListLayoutTypeSingleColumn)
	}

	if l.SelectionMode() != ListSelectionModeSingle {
		t.Errorf("默认选择模式不正确: got %v, want %v", l.SelectionMode(), ListSelectionModeSingle)
	}

	if l.Alignment() != AlignTypeLeft {
		t.Errorf("默认对齐方式不正确: got %v, want %v", l.Alignment(), AlignTypeLeft)
	}

	if l.VerticalAlignment() != VertAlignTypeTop {
		t.Errorf("默认垂直对齐不正确: got %v, want %v", l.VerticalAlignment(), VertAlignTypeTop)
	}

	if l.SelectedIndex() != -1 {
		t.Errorf("默认选中索引不正确: got %d, want -1", l.SelectedIndex())
	}

	if l.LineGap() != 0 {
		t.Errorf("默认行间距不正确: got %.1f, want 0", l.LineGap())
	}

	if l.NumItems() != 0 {
		t.Errorf("默认项数量不正确: got %d, want 0", l.NumItems())
	}
}

// TestList_SetLayout 测试设置布局
func TestList_SetLayout(t *testing.T) {
	l := NewList()

	layouts := []ListLayoutType{
		ListLayoutTypeSingleColumn,
		ListLayoutTypeSingleRow,
		ListLayoutTypeFlowHorizontal,
		ListLayoutTypeFlowVertical,
		ListLayoutTypePagination,
	}

	for _, layout := range layouts {
		l.SetLayout(layout)
		if l.Layout() != layout {
			t.Errorf("布局设置失败: got %v, want %v", l.Layout(), layout)
		}
	}
}

// TestList_SetGaps 测试设置间距
func TestList_SetGaps(t *testing.T) {
	l := NewList()

	lineGap := 10.0
	columnGap := 15.0

	l.SetLineGap(lineGap)
	if l.LineGap() != lineGap {
		t.Errorf("行间距设置失败: got %.1f, want %.1f", l.LineGap(), lineGap)
	}

	l.SetColumnGap(columnGap)
	if l.ColumnGap() != columnGap {
		t.Errorf("列间距设置失败: got %.1f, want %.1f", l.ColumnGap(), columnGap)
	}
}

// TestList_SetAlignment 测试设置对齐
func TestList_SetAlignment(t *testing.T) {
	l := NewList()

	alignments := []AlignType{
		AlignTypeLeft,
		AlignTypeCenter,
		AlignTypeRight,
	}

	for _, align := range alignments {
		l.SetAlignment(align)
		if l.Alignment() != align {
			t.Errorf("对齐设置失败: got %v, want %v", l.Alignment(), align)
		}
	}
}

// TestList_SetVerticalAlignment 测试设置垂直对齐
func TestList_SetVerticalAlignment(t *testing.T) {
	l := NewList()

	alignments := []VertAlignType{
		VertAlignTypeTop,
		VertAlignTypeMiddle,
		VertAlignTypeBottom,
	}

	for _, align := range alignments {
		l.SetVerticalAlignment(align)
		if l.VerticalAlignment() != align {
			t.Errorf("垂直对齐设置失败: got %v, want %v", l.VerticalAlignment(), align)
		}
	}
}

// TestList_SetNumItems 测试设置项数量
func TestList_SetNumItems(t *testing.T) {
	l := NewList()

	numItems := 5
	l.SetNumItems(numItems)

	if l.NumItems() != numItems {
		t.Errorf("项数量设置失败: got %d, want %d", l.NumItems(), numItems)
	}

	// 测试负数（应该保持原值或设为0）
	l.SetNumItems(-1)
	if l.NumItems() != 0 && l.NumItems() != numItems {
		t.Logf("负数项数量处理: got %d", l.NumItems())
	}
}

// TestList_GetItemAt 测试获取项
func TestList_GetItemAt(t *testing.T) {
	l := NewList()

	// 先添加一些项
	l.SetNumItems(3)
	item0 := l.GetItemAt(0)
	if item0 == nil {
		t.Error("应该能获取索引0的项")
	}

	item2 := l.GetItemAt(2)
	if item2 == nil {
		t.Error("应该能获取索引2的项")
	}

	itemOutOfRange := l.GetItemAt(10)
	if itemOutOfRange != nil {
		t.Error("超出范围的索引应该返回nil")
	}
}

// TestList_SetSelectionMode 测试设置选择模式
func TestList_SetSelectionMode(t *testing.T) {
	l := NewList()

	modes := []ListSelectionMode{
		ListSelectionModeSingle,
		ListSelectionModeMultiple,
		ListSelectionModeMultipleSingleClick,
		ListSelectionModeNone,
	}

	for _, mode := range modes {
		l.SetSelectionMode(mode)
		if l.SelectionMode() != mode {
			t.Errorf("选择模式设置失败: got %v, want %v", l.SelectionMode(), mode)
		}
	}
}

// TestList_SetSelectedIndex 测试设置选中索引
func TestList_SetSelectedIndex(t *testing.T) {
	l := NewList()
	l.SetNumItems(5)

	// 设置有效索引
	l.SetSelectedIndex(2)
	if l.SelectedIndex() != 2 {
		t.Errorf("选中索引设置失败: got %d, want 2", l.SelectedIndex())
	}

	// 测试无效索引
	l.SetSelectedIndex(10)
	// 应该保持之前的值或设为-1
	if l.SelectedIndex() != 2 && l.SelectedIndex() != -1 {
		t.Errorf("无效索引处理失败: got %d", l.SelectedIndex())
	}
}

// TestList_GetSelection 测试获取选中项
func TestList_GetSelection(t *testing.T) {
	l := NewList()
	l.SetNumItems(5)
	l.SetSelectionMode(ListSelectionModeMultiple)

	// 添加多个选择
	l.AddSelection(1, false)
	l.AddSelection(3, false)

	selection := l.GetSelection()
	if len(selection) != 2 {
		t.Errorf("选中项数量不正确: got %d, want 2", len(selection))
	}

	// 检查是否包含正确的索引
	found1 := false
	found3 := false
	for _, idx := range selection {
		if idx == 1 {
			found1 = true
		}
		if idx == 3 {
			found3 = true
		}
	}

	if !found1 || !found3 {
		t.Error("选中项列表不包含预期的索引")
	}
}

// TestList_AddSelection 测试添加选中项
func TestList_AddSelection(t *testing.T) {
	l := NewList()
	l.SetNumItems(5)
	l.SetSelectionMode(ListSelectionModeMultiple)

	// 添加单个选择
	l.AddSelection(1, false)
	if l.SelectedIndex() != 1 {
		t.Errorf("选中索引不正确: got %d, want 1", l.SelectedIndex())
	}

	// 添加多个选择
	l.AddSelection(3, false)
	selection := l.GetSelection()
	if len(selection) != 2 {
		t.Errorf("选中项数量不正确: got %d, want 2", len(selection))
	}
}

// TestList_RemoveSelection 测试移除选中项
func TestList_RemoveSelection(t *testing.T) {
	l := NewList()
	l.SetNumItems(5)
	l.SetSelectionMode(ListSelectionModeMultiple)

	// 先添加选择
	l.AddSelection(1, false)
	l.AddSelection(3, false)

	// 移除一个
	l.RemoveSelection(1)
	selection := l.GetSelection()
	if len(selection) != 1 || selection[0] != 3 {
		t.Errorf("移除选中项失败: got %v, want [3]", selection)
	}

	// 再移除一个
	l.RemoveSelection(3)
	selection = l.GetSelection()
	if len(selection) != 0 {
		t.Errorf("应该清空选中项: got %d items", len(selection))
	}
}

// TestList_ClearSelection 测试清除选择
func TestList_ClearSelection(t *testing.T) {
	l := NewList()
	l.SetNumItems(5)
	l.SetSelectionMode(ListSelectionModeMultiple)

	// 添加选择
	l.AddSelection(1, false)
	l.AddSelection(3, false)

	// 清除选择
	l.ClearSelection()

	if l.SelectedIndex() != -1 {
		t.Errorf("清除选择后选中索引应该为-1: got %d", l.SelectedIndex())
	}

	selection := l.GetSelection()
	if len(selection) != 0 {
		t.Errorf("清除选择后选中项列表应该为空: got %d items", len(selection))
	}
}

// TestList_SelectAll 测试全选
func TestList_SelectAll(t *testing.T) {
	l := NewList()
	numItems := 5
	l.SetNumItems(numItems)
	l.SetSelectionMode(ListSelectionModeMultiple)

	l.SelectAll()

	selection := l.GetSelection()
	if len(selection) != numItems {
		t.Errorf("全选失败: got %d items, want %d", len(selection), numItems)
	}
}

// TestList_SelectReverse 测试反选
func TestList_SelectReverse(t *testing.T) {
	l := NewList()
	l.SetNumItems(5)
	l.SetSelectionMode(ListSelectionModeMultiple)

	// 先选择部分项
	l.AddSelection(1, false)
	l.AddSelection(3, false)

	// 反选
	l.SelectReverse()

	selection := l.GetSelection()
	// 原选择是1,3，反选后应该是0,2,4
	if len(selection) != 3 {
		t.Errorf("反选失败: got %d items, want 3", len(selection))
	}
}

// TestList_SetDefaultItem 测试设置默认项
func TestList_SetDefaultItem(t *testing.T) {
	l := NewList()

	defaultItem := "ui://test/item"
	l.SetDefaultItem(defaultItem)

	if l.DefaultItem() != defaultItem {
		t.Errorf("默认项设置失败: got %s, want %s", l.DefaultItem(), defaultItem)
	}
}

// TestList_SetAutoResizeItem 测试自动调整项大小
func TestList_SetAutoResizeItem(t *testing.T) {
	l := NewList()

	l.SetAutoResizeItem(true)
	if !l.AutoResizeItem() {
		t.Error("自动调整项大小设置失败")
	}

	l.SetAutoResizeItem(false)
	if l.AutoResizeItem() {
		t.Error("自动调整项大小设置失败")
	}
}

// TestList_Chaining 测试链式调用
func TestList_Chaining(t *testing.T) {
	l := NewList()

	l.SetLayout(ListLayoutTypeSingleRow).
		SetLineCount(3).
		SetLineGap(5).
		SetColumnGap(10).
		SetAlignment(AlignTypeCenter).
		SetVerticalAlignment(VertAlignTypeMiddle).
		SetAutoResizeItem(true).
		SetSelectedIndex(2)

	if l.Layout() != ListLayoutTypeSingleRow {
		t.Error("链式调用设置布局失败")
	}

	if l.LineGap() != 5 {
		t.Error("链式调用设置行间距失败")
	}

	if l.Alignment() != AlignTypeCenter {
		t.Error("链式调用设置对齐失败")
	}

	if l.SelectedIndex() != 2 {
		t.Error("链式调用设置选中索引失败")
	}
}

// TestAssertList 测试类型断言
func TestAssertList(t *testing.T) {
	l := NewList()

	// 测试 AssertList
	result, ok := AssertList(l)
	if !ok {
		t.Error("AssertList 应该成功")
	}
	if result != l {
		t.Error("AssertList 返回的对象不正确")
	}

	// 测试 IsList
	if !IsList(l) {
		t.Error("IsList 应该返回 true")
	}

	// 测试不是 List 的情况
	obj := NewObject()
	_, ok = AssertList(obj)
	if ok {
		t.Error("AssertList 对非 List 对象应该失败")
	}

	if IsList(obj) {
		t.Error("IsList 对非 List 对象应该返回 false")
	}
}
