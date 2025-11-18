package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// Group 基础测试
// ============================================================================

func TestGroup_Creation(t *testing.T) {
	group := fairygui.NewGroup()
	if group == nil {
		t.Fatal("Expected non-nil group")
	}
}

func TestGroup_Layout(t *testing.T) {
	group := fairygui.NewGroup()

	// 默认为 None
	if group.Layout() != fairygui.GroupLayoutTypeNone {
		t.Errorf("Expected default layout to be None, got %d", group.Layout())
	}
}

func TestGroup_LineGap(t *testing.T) {
	group := fairygui.NewGroup()

	// 默认值
	lineGap := group.LineGap()
	if lineGap < 0 {
		t.Errorf("Expected non-negative line gap, got %d", lineGap)
	}
}

func TestGroup_ColumnGap(t *testing.T) {
	group := fairygui.NewGroup()

	// 默认值
	columnGap := group.ColumnGap()
	if columnGap < 0 {
		t.Errorf("Expected non-negative column gap, got %d", columnGap)
	}
}

func TestGroup_ExcludeInvisibles(t *testing.T) {
	group := fairygui.NewGroup()

	// 默认值（布尔值，任意值都合法）
	_ = group.ExcludeInvisibles()
}

func TestGroup_AutoSizeDisabled(t *testing.T) {
	group := fairygui.NewGroup()

	// 默认值（布尔值，任意值都合法）
	_ = group.AutoSizeDisabled()
}

func TestGroup_MainGridIndex(t *testing.T) {
	group := fairygui.NewGroup()

	// 默认为 -1
	if group.MainGridIndex() != -1 {
		t.Errorf("Expected default main grid index to be -1, got %d", group.MainGridIndex())
	}
}

func TestGroup_MainGridMinSize(t *testing.T) {
	group := fairygui.NewGroup()

	// 默认为 50
	if group.MainGridMinSize() != 50 {
		t.Errorf("Expected default main grid min size to be 50, got %d", group.MainGridMinSize())
	}
}

func TestGroup_Position(t *testing.T) {
	group := fairygui.NewGroup()
	group.SetPosition(100, 200)

	x, y := group.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestGroup_Size(t *testing.T) {
	group := fairygui.NewGroup()
	group.SetSize(150, 120)

	w, h := group.Size()
	if w != 150 || h != 120 {
		t.Errorf("Expected size (150, 120), got (%.0f, %.0f)", w, h)
	}
}

func TestGroup_Visible(t *testing.T) {
	group := fairygui.NewGroup()

	// 默认可见
	if !group.Visible() {
		t.Error("Expected group to be visible by default")
	}

	// 隐藏
	group.SetVisible(false)
	if group.Visible() {
		t.Error("Expected group to be hidden")
	}

	// 显示
	group.SetVisible(true)
	if !group.Visible() {
		t.Error("Expected group to be visible")
	}
}

func TestGroup_Name(t *testing.T) {
	group := fairygui.NewGroup()
	group.SetName("MyGroup")

	if group.Name() != "MyGroup" {
		t.Errorf("Expected name 'MyGroup', got '%s'", group.Name())
	}
}

func TestGroup_Alpha(t *testing.T) {
	group := fairygui.NewGroup()

	// 默认透明度为 1
	if group.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", group.Alpha())
	}

	// 设置半透明
	group.SetAlpha(0.5)
	if group.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", group.Alpha())
	}
}

// ============================================================================
// RawGroup 访问测试
// ============================================================================

func TestGroup_RawGroup(t *testing.T) {
	group := fairygui.NewGroup()
	raw := group.RawGroup()

	if raw == nil {
		t.Error("Expected non-nil raw group")
	}
}
