package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// ScrollBar 基础测试
// ============================================================================

func TestScrollBar_Creation(t *testing.T) {
	scrollbar := fairygui.NewScrollBar()
	if scrollbar == nil {
		t.Fatal("Expected non-nil scrollbar")
	}
}

func TestScrollBar_Position(t *testing.T) {
	scrollbar := fairygui.NewScrollBar()
	scrollbar.SetPosition(100, 200)

	x, y := scrollbar.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestScrollBar_Size(t *testing.T) {
	scrollbar := fairygui.NewScrollBar()
	scrollbar.SetSize(20, 300)

	w, h := scrollbar.Size()
	if w != 20 || h != 300 {
		t.Errorf("Expected size (20, 300), got (%.0f, %.0f)", w, h)
	}
}

func TestScrollBar_Visible(t *testing.T) {
	scrollbar := fairygui.NewScrollBar()

	// 默认可见
	if !scrollbar.Visible() {
		t.Error("Expected scrollbar to be visible by default")
	}

	// 隐藏
	scrollbar.SetVisible(false)
	if scrollbar.Visible() {
		t.Error("Expected scrollbar to be hidden")
	}

	// 显示
	scrollbar.SetVisible(true)
	if !scrollbar.Visible() {
		t.Error("Expected scrollbar to be visible")
	}
}

func TestScrollBar_Name(t *testing.T) {
	scrollbar := fairygui.NewScrollBar()
	scrollbar.SetName("MyScrollBar")

	if scrollbar.Name() != "MyScrollBar" {
		t.Errorf("Expected name 'MyScrollBar', got '%s'", scrollbar.Name())
	}
}

func TestScrollBar_Alpha(t *testing.T) {
	scrollbar := fairygui.NewScrollBar()

	// 默认透明度为 1
	if scrollbar.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", scrollbar.Alpha())
	}

	// 设置半透明
	scrollbar.SetAlpha(0.5)
	if scrollbar.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", scrollbar.Alpha())
	}
}

// ============================================================================
// RawScrollBar 访问测试
// ============================================================================

func TestScrollBar_RawScrollBar(t *testing.T) {
	scrollbar := fairygui.NewScrollBar()
	raw := scrollbar.RawScrollBar()

	if raw == nil {
		t.Error("Expected non-nil raw scrollbar")
	}
}
