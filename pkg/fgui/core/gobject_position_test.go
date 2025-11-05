package core

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
)

// TestGObjectSetPosition 验证 GObject.SetPosition() 是否正确更新 DisplayObject 的位置
func TestGObjectSetPosition(t *testing.T) {
	obj := NewGObject()
	if obj.DisplayObject() == nil {
		t.Fatal("DisplayObject is nil")
	}

	// 设置位置
	obj.SetPosition(100, 200)

	// 验证 GObject 的位置
	if obj.X() != 100 || obj.Y() != 200 {
		t.Errorf("GObject position mismatch: got (%.1f, %.1f), want (100, 200)", obj.X(), obj.Y())
	}

	// 验证 DisplayObject 的位置
	pos := obj.DisplayObject().Position()
	if pos.X != 100 || pos.Y != 200 {
		t.Errorf("DisplayObject position mismatch: got (%.1f, %.1f), want (100, 200)", pos.X, pos.Y)
	}
}

// TestGObjectSetPositionWithParent 验证在父容器中 SetPosition 是否正确工作
func TestGObjectSetPositionWithParent(t *testing.T) {
	parent := laya.NewSprite()
	child := NewGObject()

	parent.AddChild(child.DisplayObject())

	// 设置子对象位置
	child.SetPosition(50, 75)

	// 验证位置
	pos := child.DisplayObject().Position()
	if pos.X != 50 || pos.Y != 75 {
		t.Errorf("Child DisplayObject position mismatch: got (%.1f, %.1f), want (50, 75)", pos.X, pos.Y)
	}
}

// TestScrollBarPositionUpdate 验证滚动条位置更新场景
func TestScrollBarPositionUpdate(t *testing.T) {
	// 模拟 owner displayObject
	owner := laya.NewSprite()

	// 创建滚动条
	scrollBar := NewGObject()
	scrollBar.SetSize(17, 100) // 典型的滚动条尺寸

	// 添加到 owner
	owner.AddChild(scrollBar.DisplayObject())

	// 初始位置应该是 (0, 0)
	pos := scrollBar.DisplayObject().Position()
	if pos.X != 0 || pos.Y != 0 {
		t.Errorf("Initial position mismatch: got (%.1f, %.1f), want (0, 0)", pos.X, pos.Y)
	}

	// 模拟 SetViewSize 更新位置
	width := 425.0
	scrollBarWidth := scrollBar.Width()
	vtX := width - scrollBarWidth // 应该是 425 - 17 = 408

	scrollBar.SetPosition(vtX, 0)

	// 验证位置是否更新
	pos = scrollBar.DisplayObject().Position()
	expectedX := 408.0
	if pos.X != expectedX {
		t.Errorf("Updated position mismatch: got (%.1f, %.1f), want (%.1f, 0)", pos.X, pos.Y, expectedX)
	}

	// 验证 GObject 的位置也是正确的
	if scrollBar.X() != expectedX {
		t.Errorf("GObject X mismatch: got %.1f, want %.1f", scrollBar.X(), expectedX)
	}
}
