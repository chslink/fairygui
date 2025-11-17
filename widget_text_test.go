package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Text 基础测试
// ============================================================================

func TestText_Creation(t *testing.T) {
	txt := fairygui.NewText()
	if txt == nil {
		t.Fatal("Expected non-nil text")
	}
}

func TestText_Text(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetText("Hello World")

	if txt.Text() != "Hello World" {
		t.Errorf("Expected text 'Hello World', got '%s'", txt.Text())
	}
}

func TestText_Color(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetColor("#FF0000")

	if txt.Color() != "#FF0000" {
		t.Errorf("Expected color '#FF0000', got '%s'", txt.Color())
	}
}

func TestText_Font(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetFont("Arial")

	if txt.Font() != "Arial" {
		t.Errorf("Expected font 'Arial', got '%s'", txt.Font())
	}
}

func TestText_FontSize(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetFontSize(24)

	if txt.FontSize() != 24 {
		t.Errorf("Expected font size 24, got %d", txt.FontSize())
	}
}

func TestText_Align(t *testing.T) {
	txt := fairygui.NewText()

	// 默认左对齐
	if txt.Align() != "left" {
		t.Errorf("Expected default align 'left', got '%s'", txt.Align())
	}

	// 居中对齐
	txt.SetAlign("center")
	if txt.Align() != "center" {
		t.Errorf("Expected align 'center', got '%s'", txt.Align())
	}

	// 右对齐
	txt.SetAlign("right")
	if txt.Align() != "right" {
		t.Errorf("Expected align 'right', got '%s'", txt.Align())
	}
}

func TestText_VerticalAlign(t *testing.T) {
	txt := fairygui.NewText()

	// 默认顶部对齐
	if txt.VerticalAlign() != "top" {
		t.Errorf("Expected default vertical align 'top', got '%s'", txt.VerticalAlign())
	}

	// 垂直居中
	txt.SetVerticalAlign("middle")
	if txt.VerticalAlign() != "middle" {
		t.Errorf("Expected vertical align 'middle', got '%s'", txt.VerticalAlign())
	}

	// 底部对齐
	txt.SetVerticalAlign("bottom")
	if txt.VerticalAlign() != "bottom" {
		t.Errorf("Expected vertical align 'bottom', got '%s'", txt.VerticalAlign())
	}
}

func TestText_AutoSize(t *testing.T) {
	txt := fairygui.NewText()

	// 默认自动调整
	if txt.AutoSize() != widgets.TextAutoSizeBoth {
		t.Errorf("Expected default autosize Both, got %d", txt.AutoSize())
	}

	// 只调整高度
	txt.SetAutoSize(widgets.TextAutoSizeHeight)
	if txt.AutoSize() != widgets.TextAutoSizeHeight {
		t.Errorf("Expected autosize Height, got %d", txt.AutoSize())
	}

	// 不自动调整
	txt.SetAutoSize(widgets.TextAutoSizeNone)
	if txt.AutoSize() != widgets.TextAutoSizeNone {
		t.Errorf("Expected autosize None, got %d", txt.AutoSize())
	}
}

func TestText_SingleLine(t *testing.T) {
	txt := fairygui.NewText()

	// 默认非单行
	if txt.SingleLine() {
		t.Error("Expected default single line to be false")
	}

	// 设置单行
	txt.SetSingleLine(true)
	if !txt.SingleLine() {
		t.Error("Expected single line to be true")
	}
}

func TestText_Bold(t *testing.T) {
	txt := fairygui.NewText()

	// 默认非粗体
	if txt.Bold() {
		t.Error("Expected default bold to be false")
	}

	// 设置粗体
	txt.SetBold(true)
	if !txt.Bold() {
		t.Error("Expected bold to be true")
	}
}

func TestText_Italic(t *testing.T) {
	txt := fairygui.NewText()

	// 默认非斜体
	if txt.Italic() {
		t.Error("Expected default italic to be false")
	}

	// 设置斜体
	txt.SetItalic(true)
	if !txt.Italic() {
		t.Error("Expected italic to be true")
	}
}

func TestText_Underline(t *testing.T) {
	txt := fairygui.NewText()

	// 默认无下划线
	if txt.Underline() {
		t.Error("Expected default underline to be false")
	}

	// 设置下划线
	txt.SetUnderline(true)
	if !txt.Underline() {
		t.Error("Expected underline to be true")
	}
}

func TestText_LetterSpacing(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetLetterSpacing(5)

	if txt.LetterSpacing() != 5 {
		t.Errorf("Expected letter spacing 5, got %d", txt.LetterSpacing())
	}
}

func TestText_Leading(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetLeading(10)

	if txt.Leading() != 10 {
		t.Errorf("Expected leading 10, got %d", txt.Leading())
	}
}

func TestText_Stroke(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetStrokeSize(2.0)
	txt.SetStrokeColor("#000000")

	if txt.StrokeSize() != 2.0 {
		t.Errorf("Expected stroke size 2.0, got %.1f", txt.StrokeSize())
	}

	if txt.StrokeColor() != "#000000" {
		t.Errorf("Expected stroke color '#000000', got '%s'", txt.StrokeColor())
	}
}

func TestText_UBBEnabled(t *testing.T) {
	txt := fairygui.NewText()

	// 默认禁用 UBB
	if txt.UBBEnabled() {
		t.Error("Expected default UBB enabled to be false")
	}

	// 启用 UBB
	txt.SetUBBEnabled(true)
	if !txt.UBBEnabled() {
		t.Error("Expected UBB enabled to be true")
	}
}

func TestText_Shadow(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetShadow("#000000", 2, 2, 4)

	if txt.ShadowColor() != "#000000" {
		t.Errorf("Expected shadow color '#000000', got '%s'", txt.ShadowColor())
	}

	offsetX, offsetY := txt.ShadowOffset()
	if offsetX != 2 || offsetY != 2 {
		t.Errorf("Expected shadow offset (2, 2), got (%.0f, %.0f)", offsetX, offsetY)
	}

	if txt.ShadowBlur() != 4 {
		t.Errorf("Expected shadow blur 4, got %.0f", txt.ShadowBlur())
	}
}

func TestText_Position(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetPosition(100, 200)

	x, y := txt.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestText_Size(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetSize(300, 150)

	w, h := txt.Size()
	if w != 300 || h != 150 {
		t.Errorf("Expected size (300, 150), got (%.0f, %.0f)", w, h)
	}
}

func TestText_Visible(t *testing.T) {
	txt := fairygui.NewText()

	// 默认可见
	if !txt.Visible() {
		t.Error("Expected text to be visible by default")
	}

	// 隐藏
	txt.SetVisible(false)
	if txt.Visible() {
		t.Error("Expected text to be hidden")
	}

	// 显示
	txt.SetVisible(true)
	if !txt.Visible() {
		t.Error("Expected text to be visible")
	}
}

func TestText_Name(t *testing.T) {
	txt := fairygui.NewText()
	txt.SetName("MyText")

	if txt.Name() != "MyText" {
		t.Errorf("Expected name 'MyText', got '%s'", txt.Name())
	}
}

func TestText_Alpha(t *testing.T) {
	txt := fairygui.NewText()

	// 默认透明度为 1
	if txt.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", txt.Alpha())
	}

	// 设置半透明
	txt.SetAlpha(0.5)
	if txt.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", txt.Alpha())
	}
}

// ============================================================================
// RawText 访问测试
// ============================================================================

func TestText_RawText(t *testing.T) {
	txt := fairygui.NewText()
	raw := txt.RawText()

	if raw == nil {
		t.Error("Expected non-nil raw text")
	}
}
