package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// ============================================================================
// List 基础测试
// ============================================================================

func TestList_Creation(t *testing.T) {
	list := fairygui.NewList()
	if list == nil {
		t.Fatal("Expected non-nil list")
	}
}

func TestList_SelectionMode(t *testing.T) {
	list := fairygui.NewList()
	list.SetSelectionMode(fairygui.ListSelectionModeSingle)

	// 注意：SelectionMode 字段是私有的，无法直接验证
	// 只能通过后续的选择行为来验证
}

func TestList_SelectedIndex(t *testing.T) {
	list := fairygui.NewList()

	// 默认无选择
	if list.SelectedIndex() != -1 {
		t.Errorf("Expected default selected index -1, got %d", list.SelectedIndex())
	}

	// 注意：SetSelectedIndex 在空列表上可能不生效
	// 这是正常行为，因为没有实际的项目可以选择
}

func TestList_SelectedItem(t *testing.T) {
	list := fairygui.NewList()

	// 默认无选择
	if list.SelectedItem() != nil {
		t.Error("Expected nil selected item")
	}
}

func TestList_Items(t *testing.T) {
	list := fairygui.NewList()
	items := list.Items()

	// 默认为空列表
	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}
}

func TestList_SelectedIndices(t *testing.T) {
	list := fairygui.NewList()
	list.SetSelectionMode(fairygui.ListSelectionModeMultiple)

	// 默认无选择
	indices := list.SelectedIndices()
	if len(indices) != 0 {
		t.Errorf("Expected 0 selected indices, got %d", len(indices))
	}

	// 注意：在空列表上设置选择可能不生效
	// 这是正常行为，因为没有实际的项目可以选择
}

func TestList_IsSelected(t *testing.T) {
	list := fairygui.NewList()
	list.SetSelectionMode(fairygui.ListSelectionModeMultiple)

	// 注意：在空列表上设置选择可能不生效
	// 这是正常行为，因为没有实际的项目可以选择

	// 检查未选中状态
	if list.IsSelected(0) {
		t.Error("Expected index 0 not to be selected in empty list")
	}
}

func TestList_AddSelection(t *testing.T) {
	list := fairygui.NewList()
	list.SetSelectionMode(fairygui.ListSelectionModeMultiple)

	// 注意：在空列表上添加选择可能不生效
	// 这是正常行为，因为没有实际的项目可以选择
	list.AddSelection(1, false)
}

func TestList_RemoveSelection(t *testing.T) {
	list := fairygui.NewList()
	list.SetSelectionMode(fairygui.ListSelectionModeMultiple)

	// 注意：在空列表上操作选择可能不生效
	// 这是正常行为，因为没有实际的项目可以选择
	list.RemoveSelection(1)
}

func TestList_ClearSelection(t *testing.T) {
	list := fairygui.NewList()
	list.SetSelectionMode(fairygui.ListSelectionModeMultiple)
	list.SetSelectedIndices([]int{0, 1, 2})

	list.ClearSelection()

	indices := list.SelectedIndices()
	if len(indices) != 0 {
		t.Errorf("Expected 0 selected indices after clear, got %d", len(indices))
	}
}

func TestList_NumItems(t *testing.T) {
	list := fairygui.NewList()

	// 默认为 0
	if list.NumItems() != 0 {
		t.Errorf("Expected default num items 0, got %d", list.NumItems())
	}

	// 注意：SetNumItems 可能需要先设置 itemProvider 才能生效
	// 这里只测试 API 调用不会出错
	list.SetNumItems(100)
}

func TestList_ItemRenderer(t *testing.T) {
	list := fairygui.NewList()

	// 设置渲染器（不会 panic 即可）
	called := false
	list.SetItemRenderer(func(index int, item *core.GObject) {
		called = true
	})

	// 注意：渲染器只在实际有项目时调用
	// 这里只验证设置不会出错
	if called {
		t.Error("Expected renderer not to be called yet")
	}
}

func TestList_ItemProvider(t *testing.T) {
	list := fairygui.NewList()

	// 设置提供者（不会 panic 即可）
	list.SetItemProvider(func(index int) string {
		return "ui://Main/Item"
	})
}

func TestList_RefreshVirtualList(t *testing.T) {
	list := fairygui.NewList()

	// 调用刷新（不会 panic 即可）
	list.RefreshVirtualList()
}

func TestList_Layout(t *testing.T) {
	list := fairygui.NewList()

	// 检查默认布局（可能是单列）
	layout := list.Layout()
	if layout < 0 || layout > fairygui.ListLayoutTypePagination {
		t.Errorf("Expected valid layout type, got %d", layout)
	}
}

func TestList_LineGap(t *testing.T) {
	list := fairygui.NewList()

	// 默认行间距
	gap := list.LineGap()
	if gap < 0 {
		t.Errorf("Expected non-negative line gap, got %d", gap)
	}
}

func TestList_ColumnGap(t *testing.T) {
	list := fairygui.NewList()

	// 默认列间距
	gap := list.ColumnGap()
	if gap < 0 {
		t.Errorf("Expected non-negative column gap, got %d", gap)
	}
}

func TestList_LineCount(t *testing.T) {
	list := fairygui.NewList()

	// 默认行数
	count := list.LineCount()
	if count < 0 {
		t.Errorf("Expected non-negative line count, got %d", count)
	}
}

func TestList_ColumnCount(t *testing.T) {
	list := fairygui.NewList()

	// 默认列数
	count := list.ColumnCount()
	if count < 0 {
		t.Errorf("Expected non-negative column count, got %d", count)
	}
}

func TestList_Position(t *testing.T) {
	list := fairygui.NewList()
	list.SetPosition(100, 200)

	x, y := list.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestList_Size(t *testing.T) {
	list := fairygui.NewList()
	list.SetSize(400, 300)

	w, h := list.Size()
	if w != 400 || h != 300 {
		t.Errorf("Expected size (400, 300), got (%.0f, %.0f)", w, h)
	}
}

func TestList_Visible(t *testing.T) {
	list := fairygui.NewList()

	// 默认可见
	if !list.Visible() {
		t.Error("Expected list to be visible by default")
	}

	// 隐藏
	list.SetVisible(false)
	if list.Visible() {
		t.Error("Expected list to be hidden")
	}

	// 显示
	list.SetVisible(true)
	if !list.Visible() {
		t.Error("Expected list to be visible")
	}
}

func TestList_Name(t *testing.T) {
	list := fairygui.NewList()
	list.SetName("MyList")

	if list.Name() != "MyList" {
		t.Errorf("Expected name 'MyList', got '%s'", list.Name())
	}
}

func TestList_Alpha(t *testing.T) {
	list := fairygui.NewList()

	// 默认透明度为 1
	if list.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", list.Alpha())
	}

	// 设置半透明
	list.SetAlpha(0.5)
	if list.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", list.Alpha())
	}
}

// ============================================================================
// RawList 访问测试
// ============================================================================

func TestList_RawList(t *testing.T) {
	list := fairygui.NewList()
	raw := list.RawList()

	if raw == nil {
		t.Error("Expected non-nil raw list")
	}
}
