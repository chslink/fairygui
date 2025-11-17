package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// Button 基础测试
// ============================================================================

func TestButton_Creation(t *testing.T) {
	btn := fairygui.NewButton()
	if btn == nil {
		t.Fatal("Expected non-nil button")
	}
}

func TestButton_Title(t *testing.T) {
	btn := fairygui.NewButton()
	btn.SetTitle("Test Button")

	if btn.Title() != "Test Button" {
		t.Errorf("Expected title 'Test Button', got '%s'", btn.Title())
	}
}

func TestButton_Icon(t *testing.T) {
	btn := fairygui.NewButton()
	btn.SetIcon("ui://Main/icon")

	if btn.Icon() != "ui://Main/icon" {
		t.Errorf("Expected icon 'ui://Main/icon', got '%s'", btn.Icon())
	}
}

func TestButton_Mode(t *testing.T) {
	btn := fairygui.NewButton()

	// 默认模式
	if btn.Mode() != fairygui.ButtonModeCommon {
		t.Errorf("Expected default mode ButtonModeCommon, got %d", btn.Mode())
	}

	// 设置为 Check 模式
	btn.SetMode(fairygui.ButtonModeCheck)
	if btn.Mode() != fairygui.ButtonModeCheck {
		t.Errorf("Expected mode ButtonModeCheck, got %d", btn.Mode())
	}

	// 设置为 Radio 模式
	btn.SetMode(fairygui.ButtonModeRadio)
	if btn.Mode() != fairygui.ButtonModeRadio {
		t.Errorf("Expected mode ButtonModeRadio, got %d", btn.Mode())
	}
}

func TestButton_Selected(t *testing.T) {
	btn := fairygui.NewButton()
	btn.SetMode(fairygui.ButtonModeCheck)

	// 默认未选中
	if btn.Selected() {
		t.Error("Expected button to be unselected by default")
	}

	// 设置为选中
	btn.SetSelected(true)
	if !btn.Selected() {
		t.Error("Expected button to be selected")
	}

	// 取消选中
	btn.SetSelected(false)
	if btn.Selected() {
		t.Error("Expected button to be unselected")
	}
}

func TestButton_Enabled(t *testing.T) {
	btn := fairygui.NewButton()

	// 默认启用
	if !btn.Enabled() {
		t.Error("Expected button to be enabled by default")
	}

	// 禁用
	btn.SetEnabled(false)
	if btn.Enabled() {
		t.Error("Expected button to be disabled")
	}

	// 启用
	btn.SetEnabled(true)
	if !btn.Enabled() {
		t.Error("Expected button to be enabled")
	}
}

func TestButton_Sound(t *testing.T) {
	btn := fairygui.NewButton()
	btn.SetSound("ui://Main/click_sound")

	if btn.Sound() != "ui://Main/click_sound" {
		t.Errorf("Expected sound 'ui://Main/click_sound', got '%s'", btn.Sound())
	}
}

func TestButton_Position(t *testing.T) {
	btn := fairygui.NewButton()
	btn.SetPosition(100, 200)

	x, y := btn.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestButton_Size(t *testing.T) {
	btn := fairygui.NewButton()
	btn.SetSize(120, 40)

	w, h := btn.Size()
	if w != 120 || h != 40 {
		t.Errorf("Expected size (120, 40), got (%.0f, %.0f)", w, h)
	}
}

func TestButton_Visible(t *testing.T) {
	btn := fairygui.NewButton()

	// 默认可见
	if !btn.Visible() {
		t.Error("Expected button to be visible by default")
	}

	// 隐藏
	btn.SetVisible(false)
	if btn.Visible() {
		t.Error("Expected button to be hidden")
	}

	// 显示
	btn.SetVisible(true)
	if !btn.Visible() {
		t.Error("Expected button to be visible")
	}
}

func TestButton_Name(t *testing.T) {
	btn := fairygui.NewButton()
	btn.SetName("MyButton")

	if btn.Name() != "MyButton" {
		t.Errorf("Expected name 'MyButton', got '%s'", btn.Name())
	}
}

func TestButton_OnClick(t *testing.T) {
	btn := fairygui.NewButton()

	clicked := false
	btn.OnClick(func() {
		clicked = true
	})

	// 注意：这里只验证 OnClick 不会 panic
	// 实际点击测试需要完整的 UI 环境
	if clicked {
		t.Error("Expected click handler not to be called yet")
	}
}

// ============================================================================
// 便捷构造函数测试
// ============================================================================

func TestNewCheckButton(t *testing.T) {
	btn := fairygui.NewCheckButton("Enable")

	if btn.Title() != "Enable" {
		t.Errorf("Expected title 'Enable', got '%s'", btn.Title())
	}

	if btn.Mode() != fairygui.ButtonModeCheck {
		t.Errorf("Expected mode ButtonModeCheck, got %d", btn.Mode())
	}
}

func TestNewRadioButton(t *testing.T) {
	btn := fairygui.NewRadioButton("Option 1")

	if btn.Title() != "Option 1" {
		t.Errorf("Expected title 'Option 1', got '%s'", btn.Title())
	}

	if btn.Mode() != fairygui.ButtonModeRadio {
		t.Errorf("Expected mode ButtonModeRadio, got %d", btn.Mode())
	}
}

// ============================================================================
// RawButton 访问测试
// ============================================================================

func TestButton_RawButton(t *testing.T) {
	btn := fairygui.NewButton()
	raw := btn.RawButton()

	if raw == nil {
		t.Error("Expected non-nil raw button")
	}
}
