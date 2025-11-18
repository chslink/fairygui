package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// RichText 基础测试
// ============================================================================

func TestRichText_Creation(t *testing.T) {
	richText := fairygui.NewRichText()
	if richText == nil {
		t.Fatal("Expected non-nil rich text")
	}
}

func TestRichText_Text(t *testing.T) {
	richText := fairygui.NewRichText()

	// 设置普通文本
	richText.SetText("Hello World")
	if richText.Text() != "Hello World" {
		t.Errorf("Expected text 'Hello World', got '%s'", richText.Text())
	}
}

func TestRichText_RichText(t *testing.T) {
	richText := fairygui.NewRichText()

	// 设置富文本（UBB 格式）
	richText.SetText("[color=#FF0000]红色[/color][b]粗体[/b]")
	// 注意：这里只测试设置，不测试渲染结果
	if richText.Text() != "[color=#FF0000]红色[/color][b]粗体[/b]" {
		t.Errorf("Expected UBB text, got '%s'", richText.Text())
	}
}

func TestRichText_HtmlEnabled(t *testing.T) {
	richText := fairygui.NewRichText()

	// 默认启用 HTML 模式
	if !richText.HtmlEnabled() {
		t.Error("Expected HTML enabled by default")
	}

	// 禁用
	richText.SetHtmlEnabled(false)
	if richText.HtmlEnabled() {
		t.Error("Expected HTML disabled")
	}

	// 启用
	richText.SetHtmlEnabled(true)
	if !richText.HtmlEnabled() {
		t.Error("Expected HTML enabled")
	}
}

func TestRichText_UBBEnabled(t *testing.T) {
	richText := fairygui.NewRichText()

	// 默认启用 UBB
	if !richText.UBBEnabled() {
		t.Error("Expected UBB enabled by default")
	}

	// 禁用
	richText.SetUBBEnabled(false)
	if richText.UBBEnabled() {
		t.Error("Expected UBB disabled")
	}

	// 启用
	richText.SetUBBEnabled(true)
	if !richText.UBBEnabled() {
		t.Error("Expected UBB enabled")
	}
}

func TestRichText_Color(t *testing.T) {
	richText := fairygui.NewRichText()

	// 设置颜色
	richText.SetColor("#FF0000")
	if richText.Color() != "#FF0000" {
		t.Errorf("Expected color '#FF0000', got '%s'", richText.Color())
	}
}

func TestRichText_FontSize(t *testing.T) {
	richText := fairygui.NewRichText()

	// 设置字号
	richText.SetFontSize(20)
	if richText.FontSize() != 20 {
		t.Errorf("Expected font size 20, got %d", richText.FontSize())
	}
}

func TestRichText_StrokeSize(t *testing.T) {
	richText := fairygui.NewRichText()

	// 设置描边宽度
	richText.SetStrokeSize(2.0)
	if richText.StrokeSize() != 2.0 {
		t.Errorf("Expected stroke size 2.0, got %.1f", richText.StrokeSize())
	}
}

func TestRichText_StrokeColor(t *testing.T) {
	richText := fairygui.NewRichText()

	// 设置描边颜色
	richText.SetStrokeColor("#000000")
	if richText.StrokeColor() != "#000000" {
		t.Errorf("Expected stroke color '#000000', got '%s'", richText.StrokeColor())
	}
}

func TestRichText_Position(t *testing.T) {
	richText := fairygui.NewRichText()
	richText.SetPosition(100, 200)

	x, y := richText.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestRichText_Size(t *testing.T) {
	richText := fairygui.NewRichText()
	richText.SetSize(200, 100)

	w, h := richText.Size()
	if w != 200 || h != 100 {
		t.Errorf("Expected size (200, 100), got (%.0f, %.0f)", w, h)
	}
}

func TestRichText_Visible(t *testing.T) {
	richText := fairygui.NewRichText()

	// 默认可见
	if !richText.Visible() {
		t.Error("Expected rich text to be visible by default")
	}

	// 隐藏
	richText.SetVisible(false)
	if richText.Visible() {
		t.Error("Expected rich text to be hidden")
	}

	// 显示
	richText.SetVisible(true)
	if !richText.Visible() {
		t.Error("Expected rich text to be visible")
	}
}

func TestRichText_Name(t *testing.T) {
	richText := fairygui.NewRichText()
	richText.SetName("MyRichText")

	if richText.Name() != "MyRichText" {
		t.Errorf("Expected name 'MyRichText', got '%s'", richText.Name())
	}
}

func TestRichText_Alpha(t *testing.T) {
	richText := fairygui.NewRichText()

	// 默认透明度为 1
	if richText.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", richText.Alpha())
	}

	// 设置半透明
	richText.SetAlpha(0.5)
	if richText.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", richText.Alpha())
	}
}

// ============================================================================
// RawRichText 访问测试
// ============================================================================

func TestRichText_RawRichText(t *testing.T) {
	richText := fairygui.NewRichText()
	raw := richText.RawRichText()

	if raw == nil {
		t.Error("Expected non-nil raw rich text")
	}
}
