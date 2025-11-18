package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// ComboBox 基础测试
// ============================================================================

func TestComboBox_Creation(t *testing.T) {
	combo := fairygui.NewComboBox()
	if combo == nil {
		t.Fatal("Expected non-nil combobox")
	}
}

func TestComboBox_Items(t *testing.T) {
	combo := fairygui.NewComboBox()

	// 设置选项
	items := []string{"选项1", "选项2", "选项3"}
	combo.SetItems(items, nil, nil)

	// 验证选项
	gotItems := combo.Items()
	if len(gotItems) != 3 {
		t.Errorf("Expected 3 items, got %d", len(gotItems))
	}

	if combo.NumItems() != 3 {
		t.Errorf("Expected NumItems() to return 3, got %d", combo.NumItems())
	}

	for i, item := range items {
		if gotItems[i] != item {
			t.Errorf("Expected item[%d] = '%s', got '%s'", i, item, gotItems[i])
		}
	}
}

func TestComboBox_ItemsWithValues(t *testing.T) {
	combo := fairygui.NewComboBox()

	items := []string{"苹果", "香蕉", "橙子"}
	values := []string{"apple", "banana", "orange"}
	combo.SetItems(items, values, nil)

	// 验证 values
	gotValues := combo.Values()
	if len(gotValues) != 3 {
		t.Errorf("Expected 3 values, got %d", len(gotValues))
	}

	for i, val := range values {
		if gotValues[i] != val {
			t.Errorf("Expected value[%d] = '%s', got '%s'", i, val, gotValues[i])
		}
	}
}

func TestComboBox_ItemsWithIcons(t *testing.T) {
	combo := fairygui.NewComboBox()

	items := []string{"项目1", "项目2"}
	icons := []string{"ui://pkg/icon1", "ui://pkg/icon2"}
	combo.SetItems(items, nil, icons)

	// 验证 icons
	gotIcons := combo.Icons()
	if len(gotIcons) != 2 {
		t.Errorf("Expected 2 icons, got %d", len(gotIcons))
	}

	for i, icon := range icons {
		if gotIcons[i] != icon {
			t.Errorf("Expected icon[%d] = '%s', got '%s'", i, icon, gotIcons[i])
		}
	}
}

func TestComboBox_SelectedIndex(t *testing.T) {
	combo := fairygui.NewComboBox()
	combo.SetItems([]string{"选项1", "选项2", "选项3"}, nil, nil)

	// 默认未选中
	if combo.SelectedIndex() != -1 {
		t.Errorf("Expected default SelectedIndex to be -1, got %d", combo.SelectedIndex())
	}

	// 选中第一项
	combo.SetSelectedIndex(0)
	if combo.SelectedIndex() != 0 {
		t.Errorf("Expected SelectedIndex to be 0, got %d", combo.SelectedIndex())
	}

	// 选中第三项
	combo.SetSelectedIndex(2)
	if combo.SelectedIndex() != 2 {
		t.Errorf("Expected SelectedIndex to be 2, got %d", combo.SelectedIndex())
	}
}

func TestComboBox_Value(t *testing.T) {
	combo := fairygui.NewComboBox()

	items := []string{"苹果", "香蕉"}
	values := []string{"apple", "banana"}
	combo.SetItems(items, values, nil)

	// 选中第一项
	combo.SetSelectedIndex(0)
	if combo.Value() != "apple" {
		t.Errorf("Expected Value() to be 'apple', got '%s'", combo.Value())
	}

	// 选中第二项
	combo.SetSelectedIndex(1)
	if combo.Value() != "banana" {
		t.Errorf("Expected Value() to be 'banana', got '%s'", combo.Value())
	}
}

func TestComboBox_Text(t *testing.T) {
	combo := fairygui.NewComboBox()

	// 设置文本
	combo.SetText("测试文本")
	if combo.Text() != "测试文本" {
		t.Errorf("Expected Text() to be '测试文本', got '%s'", combo.Text())
	}
}

func TestComboBox_Icon(t *testing.T) {
	combo := fairygui.NewComboBox()

	// 设置图标
	combo.SetIcon("ui://pkg/icon")
	if combo.Icon() != "ui://pkg/icon" {
		t.Errorf("Expected Icon() to be 'ui://pkg/icon', got '%s'", combo.Icon())
	}
}

func TestComboBox_VisibleItemCount(t *testing.T) {
	combo := fairygui.NewComboBox()

	// 默认值
	if combo.VisibleItemCount() <= 0 {
		t.Error("Expected default VisibleItemCount to be positive")
	}

	// 设置可见项数
	combo.SetVisibleItemCount(5)
	if combo.VisibleItemCount() != 5 {
		t.Errorf("Expected VisibleItemCount to be 5, got %d", combo.VisibleItemCount())
	}
}

func TestComboBox_PopupDirection(t *testing.T) {
	combo := fairygui.NewComboBox()

	// 默认为 Auto
	if combo.PopupDirection() != fairygui.PopupDirectionAuto {
		t.Errorf("Expected default PopupDirection to be Auto, got %d", combo.PopupDirection())
	}

	// 设置为向上
	combo.SetPopupDirection(fairygui.PopupDirectionUp)
	if combo.PopupDirection() != fairygui.PopupDirectionUp {
		t.Errorf("Expected PopupDirection to be Up, got %d", combo.PopupDirection())
	}

	// 设置为向下
	combo.SetPopupDirection(fairygui.PopupDirectionDown)
	if combo.PopupDirection() != fairygui.PopupDirectionDown {
		t.Errorf("Expected PopupDirection to be Down, got %d", combo.PopupDirection())
	}
}

func TestComboBox_TitleColor(t *testing.T) {
	combo := fairygui.NewComboBox()

	// 设置颜色
	combo.SetTitleColor("#FF0000")
	if combo.TitleColor() != "#FF0000" {
		t.Errorf("Expected TitleColor to be '#FF0000', got '%s'", combo.TitleColor())
	}
}

func TestComboBox_TitleFontSize(t *testing.T) {
	combo := fairygui.NewComboBox()

	// 设置字号
	combo.SetTitleFontSize(20)
	if combo.TitleFontSize() != 20 {
		t.Errorf("Expected TitleFontSize to be 20, got %d", combo.TitleFontSize())
	}
}

func TestComboBox_TitleOutlineColor(t *testing.T) {
	combo := fairygui.NewComboBox()

	// 设置描边颜色
	combo.SetTitleOutlineColor("#000000")
	if combo.TitleOutlineColor() != "#000000" {
		t.Errorf("Expected TitleOutlineColor to be '#000000', got '%s'", combo.TitleOutlineColor())
	}
}

func TestComboBox_Position(t *testing.T) {
	combo := fairygui.NewComboBox()
	combo.SetPosition(100, 200)

	x, y := combo.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestComboBox_Size(t *testing.T) {
	combo := fairygui.NewComboBox()
	combo.SetSize(200, 30)

	w, h := combo.Size()
	if w != 200 || h != 30 {
		t.Errorf("Expected size (200, 30), got (%.0f, %.0f)", w, h)
	}
}

func TestComboBox_Visible(t *testing.T) {
	combo := fairygui.NewComboBox()

	// 默认可见
	if !combo.Visible() {
		t.Error("Expected combobox to be visible by default")
	}

	// 隐藏
	combo.SetVisible(false)
	if combo.Visible() {
		t.Error("Expected combobox to be hidden")
	}

	// 显示
	combo.SetVisible(true)
	if !combo.Visible() {
		t.Error("Expected combobox to be visible")
	}
}

func TestComboBox_Name(t *testing.T) {
	combo := fairygui.NewComboBox()
	combo.SetName("MyComboBox")

	if combo.Name() != "MyComboBox" {
		t.Errorf("Expected name 'MyComboBox', got '%s'", combo.Name())
	}
}

func TestComboBox_Alpha(t *testing.T) {
	combo := fairygui.NewComboBox()

	// 默认透明度为 1
	if combo.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", combo.Alpha())
	}

	// 设置半透明
	combo.SetAlpha(0.5)
	if combo.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", combo.Alpha())
	}
}

// ============================================================================
// RawComboBox 访问测试
// ============================================================================

func TestComboBox_RawComboBox(t *testing.T) {
	combo := fairygui.NewComboBox()
	raw := combo.RawComboBox()

	if raw == nil {
		t.Error("Expected non-nil raw combobox")
	}
}
