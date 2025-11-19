package fairygui

import (
	"testing"
)

// TestNewSlider 测试创建新的滑块
func TestNewSlider(t *testing.T) {
	s := NewSlider()
	if s == nil {
		t.Fatal("NewSlider() returned nil")
	}

	// 检查默认属性
	if s.ComponentImpl == nil {
		t.Error("Slider.ComponentImpl is nil")
	}

	if s.Max() != 100 {
		t.Errorf("默认值 max 不正确: got %.1f, want 100", s.Max())
	}

	if s.Min() != 0 {
		t.Errorf("默认值 min 不正确: got %.1f, want 0", s.Min())
	}

	if s.Value() != 0 {
		t.Errorf("默认值 value 不正确: got %.1f, want 0", s.Value())
	}

	if s.TitleType() != ProgressTitleTypePercent {
		t.Errorf("默认值 title type 不正确: got %v, want %v", s.TitleType(), ProgressTitleTypePercent)
	}

	if !s.ChangeOnClick() {
		t.Error("默认值 changeOnClick 应该是 true")
	}
}

// TestSlider_SetMinMaxValue 测试设置值
func TestSlider_SetMinMaxValue(t *testing.T) {
	s := NewSlider()

	// 设置正常值
	s.SetMin(10).SetMax(90).SetValue(50)

	if s.Min() != 10 {
		t.Errorf("Min 设置失败: got %.1f, want 10", s.Min())
	}

	if s.Max() != 90 {
		t.Errorf("Max 设置失败: got %.1f, want 90", s.Max())
	}

	if s.Value() != 50 {
		t.Errorf("Value 设置失败: got %.1f, want 50", s.Value())
	}

	// 测试超出范围的值
	s.SetValue(150)
	if s.Value() != 90 {
		t.Errorf("Value 超出最大值时应该被限制: got %.1f, want 90", s.Value())
	}

	s.SetValue(-10)
	if s.Value() != 10 {
		t.Errorf("Value 小于最小值时应该被限制: got %.1f, want 10", s.Value())
	}
}

// TestSlider_SetWholeNumbers 测试整数模式
func TestSlider_SetWholeNumbers(t *testing.T) {
	s := NewSlider()

	if s.WholeNumbers() {
		t.Error("默认应该是 false")
	}

	s.SetWholeNumbers(true)
	if !s.WholeNumbers() {
		t.Error("SetWholeNumbers(true) 失败")
	}

	// 测试整数模式下的值设置
	s.SetMin(0).SetMax(100).SetValue(50.7)
	if s.Value() != 51 {
		t.Errorf("整数模式下应该四舍五入: got %.1f, want 51", s.Value())
	}
}

// TestSlider_SetReverse 测试反向填充
func TestSlider_SetReverse(t *testing.T) {
	s := NewSlider()

	if s.Reverse() {
		t.Error("默认应该是正向")
	}

	s.SetReverse(true)
	if !s.Reverse() {
		t.Error("SetReverse(true) 失败")
	}
}

// TestSlider_SetChangeOnClick 测试点击改变
func TestSlider_SetChangeOnClick(t *testing.T) {
	s := NewSlider()

	if !s.ChangeOnClick() {
		t.Error("默认应该是 true")
	}

	s.SetChangeOnClick(false)
	if s.ChangeOnClick() {
		t.Error("SetChangeOnClick(false) 失败")
	}
}

// TestSlider_SetTitleType 测试标题类型
func TestSlider_SetTitleType(t *testing.T) {
	s := NewSlider()

	types := []ProgressTitleType{
		ProgressTitleTypePercent,
		ProgressTitleTypeValue,
		ProgressTitleTypeMax,
		ProgressTitleTypeValueAndMax,
	}

	for _, tp := range types {
		s.SetTitleType(tp)
		if s.TitleType() != tp {
			t.Errorf("TitleType 设置失败: got %v, want %v", s.TitleType(), tp)
		}
	}
}

// TestSlider_SetSize 测试尺寸设置
func TestSlider_SetSize(t *testing.T) {
	s := NewSlider()

	s.SetSize(200, 50)

	width, height := s.Size()
	if width != 200 || height != 50 {
		t.Errorf("尺寸设置失败: got (%.1f, %.1f), want (200, 50)", width, height)
	}
}

// TestAssertSlider 测试类型断言
func TestAssertSlider(t *testing.T) {
	s := NewSlider()

	result, ok := AssertSlider(s)
	if !ok {
		t.Error("AssertSlider 应该成功")
	}
	if result != s {
		t.Error("AssertSlider 返回的对象不正确")
	}

	if !IsSlider(s) {
		t.Error("IsSlider 应该返回 true")
	}

	obj := NewObject()
	_, ok = AssertSlider(obj)
	if ok {
		t.Error("AssertSlider 对非 Slider 对象应该失败")
	}

	if IsSlider(obj) {
		t.Error("IsSlider 对非 Slider 对象应该返回 false")
	}
}

// TestSlider_WithProgressBarSimilarity 测试与 ProgressBar 的相似功能
func TestSlider_WithProgressBarSimilarity(t *testing.T) {
	s := NewSlider()

	// Slider 应该支持类似 ProgressBar 的功能
	s.SetMin(0).SetMax(100).SetValue(75)

	// 计算百分比
	span := s.Max() - s.Min()
	if span <= 0 {
		span = 1
	}
	percent := (s.Value() - s.Min()) / span

	if percent != 0.75 {
		t.Errorf("百分比计算失败: got %.2f, want 0.75", percent)
	}

	// 反向测试
	s.SetReverse(true).SetValue(75)
	percent = (s.Value() - s.Min()) / span
	if percent != 0.75 {
		t.Errorf("反向模式下百分比计算失败: got %.2f, want 0.75", percent)
	}
}

// TestSlider_Chaining 测试链式调用
func TestSlider_Chaining(t *testing.T) {
	// 虽然是 Fluent API 风格，但 SetSize 返回 void，所以只测试那些返回 *Slider 的方法
	s := NewSlider()

	s.SetMin(10).SetMax(200).SetValue(100).SetTitleType(ProgressTitleTypeValue).SetReverse(true).SetWholeNumbers(true).SetChangeOnClick(false)

	if s.Min() != 10 || s.Max() != 200 || s.Value() != 100 ||
		s.TitleType() != ProgressTitleTypeValue || !s.Reverse() ||
		!s.WholeNumbers() || s.ChangeOnClick() {
		t.Error("链式调用失败")
	}
}
