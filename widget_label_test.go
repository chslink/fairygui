package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// Label 基础测试
// ============================================================================

func TestLabel_Creation(t *testing.T) {
	label := fairygui.NewLabel()
	if label == nil {
		t.Fatal("Expected non-nil label")
	}
}

func TestLabel_Title(t *testing.T) {
	label := fairygui.NewLabel()

	// 设置标题
	label.SetTitle("Hello World")
	if label.Title() != "Hello World" {
		t.Errorf("Expected title 'Hello World', got '%s'", label.Title())
	}
}

func TestLabel_Icon(t *testing.T) {
	label := fairygui.NewLabel()

	// 设置图标
	label.SetIcon("ui://package/icon")
	if label.Icon() != "ui://package/icon" {
		t.Errorf("Expected icon 'ui://package/icon', got '%s'", label.Icon())
	}
}

func TestLabel_TitleColor(t *testing.T) {
	label := fairygui.NewLabel()

	// 设置颜色
	label.SetTitleColor("#FF0000")
	if label.TitleColor() != "#FF0000" {
		t.Errorf("Expected title color '#FF0000', got '%s'", label.TitleColor())
	}
}

func TestLabel_TitleOutlineColor(t *testing.T) {
	label := fairygui.NewLabel()

	// 设置描边颜色
	label.SetTitleOutlineColor("#000000")
	if label.TitleOutlineColor() != "#000000" {
		t.Errorf("Expected outline color '#000000', got '%s'", label.TitleOutlineColor())
	}
}

func TestLabel_TitleFontSize(t *testing.T) {
	label := fairygui.NewLabel()

	// 设置字号
	label.SetTitleFontSize(20)
	if label.TitleFontSize() != 20 {
		t.Errorf("Expected font size 20, got %d", label.TitleFontSize())
	}
}

func TestLabel_Position(t *testing.T) {
	label := fairygui.NewLabel()
	label.SetPosition(100, 200)

	x, y := label.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestLabel_Size(t *testing.T) {
	label := fairygui.NewLabel()
	label.SetSize(150, 30)

	w, h := label.Size()
	if w != 150 || h != 30 {
		t.Errorf("Expected size (150, 30), got (%.0f, %.0f)", w, h)
	}
}

func TestLabel_Visible(t *testing.T) {
	label := fairygui.NewLabel()

	// 默认可见
	if !label.Visible() {
		t.Error("Expected label to be visible by default")
	}

	// 隐藏
	label.SetVisible(false)
	if label.Visible() {
		t.Error("Expected label to be hidden")
	}

	// 显示
	label.SetVisible(true)
	if !label.Visible() {
		t.Error("Expected label to be visible")
	}
}

func TestLabel_Name(t *testing.T) {
	label := fairygui.NewLabel()
	label.SetName("MyLabel")

	if label.Name() != "MyLabel" {
		t.Errorf("Expected name 'MyLabel', got '%s'", label.Name())
	}
}

func TestLabel_Alpha(t *testing.T) {
	label := fairygui.NewLabel()

	// 默认透明度为 1
	if label.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", label.Alpha())
	}

	// 设置半透明
	label.SetAlpha(0.5)
	if label.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", label.Alpha())
	}
}

// ============================================================================
// RawLabel 访问测试
// ============================================================================

func TestLabel_RawLabel(t *testing.T) {
	label := fairygui.NewLabel()
	raw := label.RawLabel()

	if raw == nil {
		t.Error("Expected non-nil raw label")
	}
}
