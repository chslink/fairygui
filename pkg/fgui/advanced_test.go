package fgui_test

import (
	"testing"
	"time"

	"github.com/chslink/fairygui/pkg/fgui"
)

// ============================================================================
// Tween Animation API 测试
// ============================================================================

func TestTween_To(t *testing.T) {
	// 测试一维补间
	tw := fgui.TweenTo(0, 100, 1.0)
	if tw == nil {
		t.Fatal("Expected non-nil tweener")
	}

	// 测试链式调用
	tw.SetDelay(0.5).SetEase(fgui.EaseType(0))
}

func TestTween_To2(t *testing.T) {
	// 测试二维补间
	tw := fgui.TweenTo2(0, 0, 100, 200, 1.0)
	if tw == nil {
		t.Fatal("Expected non-nil tweener")
	}
}

func TestTween_To3(t *testing.T) {
	// 测试三维补间
	tw := fgui.TweenTo3(0, 0, 0, 100, 200, 300, 1.0)
	if tw == nil {
		t.Fatal("Expected non-nil tweener")
	}
}

func TestTween_ToColor(t *testing.T) {
	// 测试颜色补间
	tw := fgui.TweenToColor(0xFFFF0000, 0xFF00FF00, 1.0)
	if tw == nil {
		t.Fatal("Expected non-nil tweener")
	}
}

func TestTween_Shake(t *testing.T) {
	// 测试抖动效果
	tw := fgui.TweenShake(100, 100, 10, 0.5)
	if tw == nil {
		t.Fatal("Expected non-nil tweener")
	}
}

func TestTween_DelayedCall(t *testing.T) {
	// 测试延迟调用
	called := false
	tw := fgui.TweenDelayedCall(0.1).OnComplete(func(t *fgui.GTweener) {
		called = true
	})
	if tw == nil {
		t.Fatal("Expected non-nil tweener")
	}

	// 推进时间
	fgui.TweenAdvance(200 * time.Millisecond)

	// 验证回调被调用
	if !called {
		t.Error("Expected callback to be called after delay")
	}
}

func TestTween_IsTweening(t *testing.T) {
	// 创建测试对象
	obj := &struct{ x float64 }{x: 0}

	// 创建补间
	tw := fgui.TweenTo(0, 100, 1.0).SetTarget(obj, "x")
	if tw == nil {
		t.Fatal("Expected non-nil tweener")
	}

	// 检查是否在补间中
	if !fgui.IsTweening(obj, "x") {
		t.Error("Expected object to be tweening")
	}

	// 终止补间
	fgui.KillTween(obj, false, "x")

	// 再次检查
	if fgui.IsTweening(obj, "x") {
		t.Error("Expected object to not be tweening after kill")
	}
}

func TestTween_GetTween(t *testing.T) {
	// 创建测试对象
	obj := &struct{ x float64 }{x: 0}

	// 创建补间
	tw := fgui.TweenTo(0, 100, 1.0).SetTarget(obj, "x")
	if tw == nil {
		t.Fatal("Expected non-nil tweener")
	}

	// 获取补间
	retrieved := fgui.GetTween(obj, "x")
	if retrieved == nil {
		t.Error("Expected to retrieve tweener")
	}

	// 清理
	fgui.KillTween(obj, false, "x")
}

// ============================================================================
// Relations API 测试
// ============================================================================

func TestRelations_Types(t *testing.T) {
	// 测试 Relations 类型可以使用
	var _ fgui.RelationType

	// 测试 Relations 对象（从 GObject 获取）
	obj := fgui.NewGObject()
	relations := obj.Relations()
	if relations == nil {
		t.Error("Expected non-nil relations")
	}
}

// ============================================================================
// Transitions API 测试
// ============================================================================

func TestTransitions_Types(t *testing.T) {
	// 测试 Transition 类型可以使用
	var _ *fgui.Transition
	var _ *fgui.TransitionInfo
}

func TestTransitions_FromComponent(t *testing.T) {
	// 测试从 GComponent 获取 Transition
	comp := fgui.NewGComponent()
	trans := comp.Transition("test")
	// Transition 不存在时返回 nil 是正常的
	_ = trans
}

// ============================================================================
// Gears API 测试
// ============================================================================

func TestGears_Types(t *testing.T) {
	// 测试 Gear 接口可以使用
	var _ fgui.Gear

	// 测试 Controller 类型可以使用
	var _ *fgui.Controller
}

func TestGears_FromComponent(t *testing.T) {
	// 测试从 GComponent 获取 Controller
	comp := fgui.NewGComponent()
	ctrl := comp.GetController("test")
	// Controller 不存在时返回 nil 是正常的
	_ = ctrl
}

// ============================================================================
// 集成测试：高级功能组合使用
// ============================================================================

func TestAdvanced_Integration(t *testing.T) {
	// 创建组件
	comp := fgui.NewGComponent()
	comp.SetSize(100, 100)

	// 使用 Tween 动画改变位置
	obj := &struct {
		x, y float64
	}{x: 0, y: 0}

	tw := fgui.TweenTo2(0, 0, 100, 100, 1.0).
		SetTarget(obj).
		OnUpdate(func(t *fgui.GTweener) {
			val := t.Value()
			x, y := val.XY()
			comp.SetPosition(x, y)
		})

	if tw == nil {
		t.Fatal("Expected non-nil tweener")
	}

	// 推进时间
	fgui.TweenAdvance(500 * time.Millisecond)

	// 验证位置改变
	x, y := comp.X(), comp.Y()
	if x == 0 && y == 0 {
		t.Error("Expected position to change after tween advance")
	}

	// 清理
	fgui.KillTween(obj, false)
}
