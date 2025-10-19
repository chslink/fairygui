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

func TestGObjectScaleAndRotation(t *testing.T) {
	obj := NewGObject()
	obj.SetScale(1.5, 0.75)
	obj.SetRotation(0.5)

	sx, sy := obj.Scale()
	if sx != 1.5 || sy != 0.75 {
		t.Fatalf("expected scale (1.5,0.75), got (%v,%v)", sx, sy)
	}
	if obj.Rotation() != 0.5 {
		t.Fatalf("expected rotation 0.5 radians, got %v", obj.Rotation())
	}

	sprite := obj.DisplayObject()
	if ax, ay := sprite.Scale(); ax != 1.5 || ay != 0.75 {
		t.Fatalf("sprite scale mismatch (%v,%v)", ax, ay)
	}
	if sprite.Rotation() != 0.5 {
		t.Fatalf("sprite rotation mismatch: %v", sprite.Rotation())
	}
}

func TestGObjectPivot(t *testing.T) {
	obj := NewGObject()
	obj.SetPivot(0.25, 0.75)

	px, py := obj.Pivot()
	if px != 0.25 || py != 0.75 {
		t.Fatalf("expected pivot (0.25,0.75), got (%v,%v)", px, py)
	}

	sprite := obj.DisplayObject()
	pivot := sprite.Pivot()
	if pivot.X != 0.25 || pivot.Y != 0.75 {
		t.Fatalf("sprite pivot mismatch (%v,%v)", pivot.X, pivot.Y)
	}
}

func TestGObjectPivotAnchorAdjustsPosition(t *testing.T) {
	obj := NewGObject()
	obj.SetSize(100, 50)
	obj.SetPivotWithAnchor(0.5, 0.5, true)
	obj.SetPosition(200, 120)

	pos := obj.DisplayObject().Position()
	expectedX := 200 - 0.5*100
	expectedY := 120 - 0.5*50
	if pos.X != expectedX || pos.Y != expectedY {
		t.Fatalf("expected anchored position (%v,%v), got (%v,%v)", expectedX, expectedY, pos.X, pos.Y)
	}
	if !obj.PivotAsAnchor() {
		t.Fatalf("expected pivot to be treated as anchor")
	}
}

func TestGObjectSkew(t *testing.T) {
	obj := NewGObject()
	obj.SetSkew(0.2, -0.1)

	sx, sy := obj.Skew()
	if sx != 0.2 || sy != -0.1 {
		t.Fatalf("expected skew (0.2,-0.1), got (%v,%v)", sx, sy)
	}
	ax, ay := obj.DisplayObject().Skew()
	if ax != 0.2 || ay != -0.1 {
		t.Fatalf("sprite skew mismatch (%v,%v)", ax, ay)
	}
}

func TestGObjectDisplayOwner(t *testing.T) {
	obj := NewGObject()
	owner := obj.DisplayObject().Owner()
	if owner != obj {
		t.Fatalf("expected sprite owner to be the gobject, got %#v", owner)
	}
}
