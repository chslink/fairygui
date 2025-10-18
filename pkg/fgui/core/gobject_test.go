package core

import "testing"

func TestGObjectSetPositionUpdatesSprite(t *testing.T) {
	obj := NewGObject()
	obj.SetPosition(42, 84)

	pos := obj.DisplayObject().Position()
	if pos.X != 42 || pos.Y != 84 {
		t.Fatalf("expected sprite position (42,84), got (%v,%v)", pos.X, pos.Y)
	}
}

func TestGObjectSetSizeUpdatesSprite(t *testing.T) {
	obj := NewGObject()
	obj.SetSize(128, 64)

	w, h := obj.DisplayObject().Size()
	if w != 128 || h != 64 {
		t.Fatalf("expected sprite size (128,64), got (%v,%v)", w, h)
	}
}
