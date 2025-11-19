package fairygui

import (
	"testing"
)

// TestNewProgressBar 测试创建新的进度条
func TestNewProgressBar(t *testing.T) {
	pb := NewProgressBar()
	if pb == nil {
		t.Fatal("NewProgressBar() returned nil")
	}

	// 检查默认属性
	if pb.ComponentImpl == nil {
		t.Error("ProgressBar.ComponentImpl is nil")
	}

	if pb.Max() != 100 {
		t.Errorf("默认值 max 不正确: got %.1f, want 100", pb.Max())
	}

	if pb.Min() != 0 {
		t.Errorf("默认值 min 不正确: got %.1f, want 0", pb.Min())
	}

	if pb.Value() != 0 {
		t.Errorf("默认值 value 不正确: got %.1f, want 0", pb.Value())
	}

	if pb.TitleType() != ProgressTitleTypePercent {
		t.Errorf("默认值 title type 不正确: got %v, want %v", pb.TitleType(), ProgressTitleTypePercent)
	}
}

// TestProgressBar_SetMinMaxValue 测试设置值
func TestProgressBar_SetMinMaxValue(t *testing.T) {
	pb := NewProgressBar()

	// 设置正常值
	pb.SetValue(50)
	if pb.Value() != 50 {
		t.Errorf("Value 设置失败: got %.1f, want 50", pb.Value())
	}

	// 测试超出范围的值
	pb.SetValue(150)
	if pb.Value() != 100 {
		t.Errorf("Value 超出最大值时应该被限制: got %.1f, want 100", pb.Value())
	}

	pb.SetValue(-10)
	if pb.Value() != 0 {
		t.Errorf("Value 小于最小值时应该被限制: got %.1f, want 0", pb.Value())
	}
}

// TestProgressBar_SetTitleType 测试标题类型
func TestProgressBar_SetTitleType(t *testing.T) {
	pb := NewProgressBar()

	types := []ProgressTitleType{
		ProgressTitleTypePercent,
		ProgressTitleTypeValue,
		ProgressTitleTypeMax,
		ProgressTitleTypeValueAndMax,
	}

	for _, tp := range types {
		pb.SetTitleType(tp)
		if pb.TitleType() != tp {
			t.Errorf("TitleType 设置失败: got %v, want %v", pb.TitleType(), tp)
		}
	}
}

// TestProgressBar_CalculatePercent 测试百分比计算
func TestProgressBar_CalculatePercent(t *testing.T) {
	pb := NewProgressBar()

	pb.SetMin(10)
	pb.SetMax(90)
	pb.SetValue(50)

	// 计算百分比
	span := pb.Max() - pb.Min()
	if span == 0 {
		span = 1
	}
	percent := (pb.Value() - pb.Min()) / span

	if percent != 0.5 {
		t.Errorf("百分比计算失败: got %.2f, want 0.5", percent)
	}
}

// TestAssertProgressBar 测试类型断言
func TestAssertProgressBar(t *testing.T) {
	pb := NewProgressBar()

	result, ok := AssertProgressBar(pb)
	if !ok {
		t.Error("AssertProgressBar 应该成功")
	}
	if result != pb {
		t.Error("AssertProgressBar 返回的对象不正确")
	}

	if !IsProgressBar(pb) {
		t.Error("IsProgressBar 应该返回 true")
	}

	obj := NewObject()
	_, ok = AssertProgressBar(obj)
	if ok {
		t.Error("AssertProgressBar 对非 ProgressBar 对象应该失败")
	}

	if IsProgressBar(obj) {
		t.Error("IsProgressBar 对非 ProgressBar 对象应该返回 false")
	}
}
