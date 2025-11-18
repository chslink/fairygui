package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Slider 基础测试
// ============================================================================

func TestSlider_Creation(t *testing.T) {
	slider := fairygui.NewSlider()
	if slider == nil {
		t.Fatal("Expected non-nil slider")
	}
}

func TestSlider_MinMax(t *testing.T) {
	slider := fairygui.NewSlider()

	// 设置范围
	slider.SetMin(0)
	slider.SetMax(100)

	if slider.Min() != 0 {
		t.Errorf("Expected min 0, got %.0f", slider.Min())
	}

	if slider.Max() != 100 {
		t.Errorf("Expected max 100, got %.0f", slider.Max())
	}
}

func TestSlider_Value(t *testing.T) {
	slider := fairygui.NewSlider()
	slider.SetMin(0)
	slider.SetMax(100)

	// 设置值
	slider.SetValue(50)
	if slider.Value() != 50 {
		t.Errorf("Expected value 50, got %.0f", slider.Value())
	}

	// 设置边界值
	slider.SetValue(0)
	if slider.Value() != 0 {
		t.Errorf("Expected value 0, got %.0f", slider.Value())
	}

	slider.SetValue(100)
	if slider.Value() != 100 {
		t.Errorf("Expected value 100, got %.0f", slider.Value())
	}
}

func TestSlider_ValueClamping(t *testing.T) {
	slider := fairygui.NewSlider()
	slider.SetMin(0)
	slider.SetMax(100)

	// 测试值会被限制在范围内
	slider.SetValue(150)
	if slider.Value() > 100 {
		t.Errorf("Expected value to be clamped to 100, got %.0f", slider.Value())
	}

	slider.SetValue(-10)
	if slider.Value() < 0 {
		t.Errorf("Expected value to be clamped to 0, got %.0f", slider.Value())
	}
}

func TestSlider_WholeNumbers(t *testing.T) {
	slider := fairygui.NewSlider()
	slider.SetMin(0)
	slider.SetMax(10)

	// 启用整数模式
	slider.SetWholeNumbers(true)

	// 设置小数值
	slider.SetValue(5.7)

	// 值应该被四舍五入
	val := slider.Value()
	if val != 6 && val != 5 {
		t.Errorf("Expected value to be rounded to whole number, got %.1f", val)
	}
}

func TestSlider_ChangeOnClick(t *testing.T) {
	slider := fairygui.NewSlider()

	// 设置点击改变值
	slider.SetChangeOnClick(true)

	// 注意：实际的点击行为需要在 GUI 环境中测试
	// 这里只验证设置不会出错
}

func TestSlider_Reverse(t *testing.T) {
	slider := fairygui.NewSlider()

	// 设置反向显示
	slider.SetReverse(true)

	// 注意：实际的显示效果需要在 GUI 环境中验证
	// 这里只验证设置不会出错
}

func TestSlider_TitleType(t *testing.T) {
	slider := fairygui.NewSlider()

	// 设置标题类型
	slider.SetTitleType(widgets.ProgressTitleTypePercent)
	slider.SetTitleType(widgets.ProgressTitleTypeValue)
	slider.SetTitleType(widgets.ProgressTitleTypeValueAndMax)

	// 注意：实际的标题显示需要在 GUI 环境中验证
	// 这里只验证设置不会出错
}

func TestSlider_Position(t *testing.T) {
	slider := fairygui.NewSlider()
	slider.SetPosition(100, 200)

	x, y := slider.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestSlider_Size(t *testing.T) {
	slider := fairygui.NewSlider()
	slider.SetSize(200, 30)

	w, h := slider.Size()
	if w != 200 || h != 30 {
		t.Errorf("Expected size (200, 30), got (%.0f, %.0f)", w, h)
	}
}

func TestSlider_Visible(t *testing.T) {
	slider := fairygui.NewSlider()

	// 默认可见
	if !slider.Visible() {
		t.Error("Expected slider to be visible by default")
	}

	// 隐藏
	slider.SetVisible(false)
	if slider.Visible() {
		t.Error("Expected slider to be hidden")
	}

	// 显示
	slider.SetVisible(true)
	if !slider.Visible() {
		t.Error("Expected slider to be visible")
	}
}

func TestSlider_Name(t *testing.T) {
	slider := fairygui.NewSlider()
	slider.SetName("MySlider")

	if slider.Name() != "MySlider" {
		t.Errorf("Expected name 'MySlider', got '%s'", slider.Name())
	}
}

func TestSlider_Alpha(t *testing.T) {
	slider := fairygui.NewSlider()

	// 默认透明度为 1
	if slider.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", slider.Alpha())
	}

	// 设置半透明
	slider.SetAlpha(0.5)
	if slider.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", slider.Alpha())
	}
}

// ============================================================================
// RawSlider 访问测试
// ============================================================================

func TestSlider_RawSlider(t *testing.T) {
	slider := fairygui.NewSlider()
	raw := slider.RawSlider()

	if raw == nil {
		t.Error("Expected non-nil raw slider")
	}
}
