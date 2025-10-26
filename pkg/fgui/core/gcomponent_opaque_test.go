package core

import (
	"testing"
)

// TestSetOpaque 验证 SetOpaque 正确设置 mouseThrough
func TestSetOpaque(t *testing.T) {
	comp := NewGComponent()
	if comp == nil || comp.DisplayObject() == nil {
		t.Fatal("NewGComponent 返回 nil")
	}

	sprite := comp.DisplayObject()

	// 1. 初始状态：opaque=false, mouseEnabled=true, mouseThrough=true
	// 参考 TypeScript 版本：GComponent 默认 opaque=false
	if !sprite.MouseEnabled() {
		t.Error("初始 mouseEnabled 应该为 true")
	}
	if !sprite.MouseThrough() {
		t.Error("初始 mouseThrough 应该为 true（因为默认 opaque=false）")
	}
	if comp.Opaque() {
		t.Error("初始 Opaque 应该为 false")
	}

	// 2. SetOpaque(true) → mouseThrough 应该变为 false
	comp.SetOpaque(true)
	if comp.Opaque() != true {
		t.Error("SetOpaque(true) 后 Opaque() 应该返回 true")
	}
	if sprite.MouseThrough() {
		t.Error("SetOpaque(true) 后 mouseThrough 应该为 false（拦截事件）")
	}

	// 3. SetOpaque(false) → mouseThrough 应该变回 true
	comp.SetOpaque(false)
	if comp.Opaque() != false {
		t.Error("SetOpaque(false) 后 Opaque() 应该返回 false")
	}
	if !sprite.MouseThrough() {
		t.Error("SetOpaque(false) 后 mouseThrough 应该为 true（穿透事件）")
	}

	// 4. 重复设置相同值不应该有副作用
	comp.SetOpaque(false)
	if !sprite.MouseThrough() {
		t.Error("重复 SetOpaque(false) 后 mouseThrough 应该仍为 true")
	}
}
