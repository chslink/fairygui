package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// Image 基础测试
// ============================================================================

func TestImage_Creation(t *testing.T) {
	img := fairygui.NewImage()
	if img == nil {
		t.Fatal("Expected non-nil image")
	}
}

func TestImage_Color(t *testing.T) {
	img := fairygui.NewImage()
	img.SetColor("#FF0000")

	if img.Color() != "#FF0000" {
		t.Errorf("Expected color '#FF0000', got '%s'", img.Color())
	}
}

func TestImage_Flip(t *testing.T) {
	img := fairygui.NewImage()

	// 默认不翻转
	if img.Flip() != fairygui.FlipTypeNone {
		t.Errorf("Expected default flip FlipTypeNone, got %d", img.Flip())
	}

	// 水平翻转
	img.SetFlip(fairygui.FlipTypeHorizontal)
	if img.Flip() != fairygui.FlipTypeHorizontal {
		t.Errorf("Expected flip FlipTypeHorizontal, got %d", img.Flip())
	}

	// 垂直翻转
	img.SetFlip(fairygui.FlipTypeVertical)
	if img.Flip() != fairygui.FlipTypeVertical {
		t.Errorf("Expected flip FlipTypeVertical, got %d", img.Flip())
	}

	// 双向翻转
	img.SetFlip(fairygui.FlipTypeBoth)
	if img.Flip() != fairygui.FlipTypeBoth {
		t.Errorf("Expected flip FlipTypeBoth, got %d", img.Flip())
	}
}

func TestImage_Position(t *testing.T) {
	img := fairygui.NewImage()
	img.SetPosition(100, 200)

	x, y := img.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestImage_Size(t *testing.T) {
	img := fairygui.NewImage()
	img.SetSize(120, 80)

	w, h := img.Size()
	if w != 120 || h != 80 {
		t.Errorf("Expected size (120, 80), got (%.0f, %.0f)", w, h)
	}
}

func TestImage_Visible(t *testing.T) {
	img := fairygui.NewImage()

	// 默认可见
	if !img.Visible() {
		t.Error("Expected image to be visible by default")
	}

	// 隐藏
	img.SetVisible(false)
	if img.Visible() {
		t.Error("Expected image to be hidden")
	}

	// 显示
	img.SetVisible(true)
	if !img.Visible() {
		t.Error("Expected image to be visible")
	}
}

func TestImage_Name(t *testing.T) {
	img := fairygui.NewImage()
	img.SetName("MyImage")

	if img.Name() != "MyImage" {
		t.Errorf("Expected name 'MyImage', got '%s'", img.Name())
	}
}

func TestImage_Alpha(t *testing.T) {
	img := fairygui.NewImage()

	// 默认透明度为 1
	if img.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", img.Alpha())
	}

	// 设置半透明
	img.SetAlpha(0.5)
	if img.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", img.Alpha())
	}

	// 设置完全透明
	img.SetAlpha(0.0)
	if img.Alpha() != 0.0 {
		t.Errorf("Expected alpha 0.0, got %.2f", img.Alpha())
	}
}

// ============================================================================
// RawImage 访问测试
// ============================================================================

func TestImage_RawImage(t *testing.T) {
	img := fairygui.NewImage()
	raw := img.RawImage()

	if raw == nil {
		t.Error("Expected non-nil raw image")
	}
}
