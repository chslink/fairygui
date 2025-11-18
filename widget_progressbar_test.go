package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// ProgressBar 基础测试
// ============================================================================

func TestProgressBar_Creation(t *testing.T) {
	bar := fairygui.NewProgressBar()
	if bar == nil {
		t.Fatal("Expected non-nil progress bar")
	}
}

func TestProgressBar_MinMax(t *testing.T) {
	bar := fairygui.NewProgressBar()

	// 设置范围
	bar.SetMin(0)
	bar.SetMax(100)

	if bar.Min() != 0 {
		t.Errorf("Expected min 0, got %.0f", bar.Min())
	}

	if bar.Max() != 100 {
		t.Errorf("Expected max 100, got %.0f", bar.Max())
	}
}

func TestProgressBar_Value(t *testing.T) {
	bar := fairygui.NewProgressBar()
	bar.SetMin(0)
	bar.SetMax(100)

	// 设置值
	bar.SetValue(50)
	if bar.Value() != 50 {
		t.Errorf("Expected value 50, got %.0f", bar.Value())
	}

	// 设置边界值
	bar.SetValue(0)
	if bar.Value() != 0 {
		t.Errorf("Expected value 0, got %.0f", bar.Value())
	}

	bar.SetValue(100)
	if bar.Value() != 100 {
		t.Errorf("Expected value 100, got %.0f", bar.Value())
	}
}

func TestProgressBar_ValueClamping(t *testing.T) {
	bar := fairygui.NewProgressBar()
	bar.SetMin(0)
	bar.SetMax(100)

	// 测试值会被限制在范围内
	bar.SetValue(150)
	if bar.Value() > 100 {
		t.Errorf("Expected value to be clamped to 100, got %.0f", bar.Value())
	}

	bar.SetValue(-10)
	if bar.Value() < 0 {
		t.Errorf("Expected value to be clamped to 0, got %.0f", bar.Value())
	}
}

func TestProgressBar_TitleType(t *testing.T) {
	bar := fairygui.NewProgressBar()

	// 测试不同标题类型
	titleTypes := []widgets.ProgressTitleType{
		widgets.ProgressTitleTypePercent,
		widgets.ProgressTitleTypeValueAndMax,
		widgets.ProgressTitleTypeValue,
		widgets.ProgressTitleTypeMax,
	}

	for _, titleType := range titleTypes {
		bar.SetTitleType(titleType)
		if bar.TitleType() != titleType {
			t.Errorf("Expected title type %d, got %d", titleType, bar.TitleType())
		}
	}
}

func TestProgressBar_Reverse(t *testing.T) {
	bar := fairygui.NewProgressBar()

	// 设置反向显示
	bar.SetReverse(true)
	// 注意：没有 Reverse() getter，只验证设置不会出错

	bar.SetReverse(false)
}

func TestProgressBar_Position(t *testing.T) {
	bar := fairygui.NewProgressBar()
	bar.SetPosition(100, 200)

	x, y := bar.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestProgressBar_Size(t *testing.T) {
	bar := fairygui.NewProgressBar()
	bar.SetSize(200, 30)

	w, h := bar.Size()
	if w != 200 || h != 30 {
		t.Errorf("Expected size (200, 30), got (%.0f, %.0f)", w, h)
	}
}

func TestProgressBar_Visible(t *testing.T) {
	bar := fairygui.NewProgressBar()

	// 默认可见
	if !bar.Visible() {
		t.Error("Expected progress bar to be visible by default")
	}

	// 隐藏
	bar.SetVisible(false)
	if bar.Visible() {
		t.Error("Expected progress bar to be hidden")
	}

	// 显示
	bar.SetVisible(true)
	if !bar.Visible() {
		t.Error("Expected progress bar to be visible")
	}
}

func TestProgressBar_Name(t *testing.T) {
	bar := fairygui.NewProgressBar()
	bar.SetName("MyProgressBar")

	if bar.Name() != "MyProgressBar" {
		t.Errorf("Expected name 'MyProgressBar', got '%s'", bar.Name())
	}
}

func TestProgressBar_Alpha(t *testing.T) {
	bar := fairygui.NewProgressBar()

	// 默认透明度为 1
	if bar.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", bar.Alpha())
	}

	// 设置半透明
	bar.SetAlpha(0.5)
	if bar.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", bar.Alpha())
	}
}

// ============================================================================
// RawProgressBar 访问测试
// ============================================================================

func TestProgressBar_RawProgressBar(t *testing.T) {
	bar := fairygui.NewProgressBar()
	raw := bar.RawProgressBar()

	if raw == nil {
		t.Error("Expected non-nil raw progress bar")
	}
}
