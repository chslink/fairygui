package core

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
)

func TestGComponentAddChildSyncsDisplayTree(t *testing.T) {
	parent := NewGComponent()
	child := NewGObject()

	parent.SetSize(200, 200)
	child.SetSize(50, 50)
	child.SetPosition(10, 20)

	parent.AddChild(child)

	if len(parent.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children()))
	}

	if len(parent.DisplayObject().Children()) != 1 {
		t.Fatalf("display tree not synced, got %d", len(parent.DisplayObject().Children()))
	}

	childSprite := parent.DisplayObject().Children()[0]
	global := childSprite.LocalToGlobal(laya.Point{})
	if global.X != 10 || global.Y != 20 {
		t.Fatalf("expected child positioned at (10,20), got (%v,%v)", global.X, global.Y)
	}
}

func TestGComponentRemoveChildUpdatesDisplayTree(t *testing.T) {
	parent := NewGComponent()
	child := NewGObject()

	parent.AddChild(child)
	parent.RemoveChild(child)

	if len(parent.Children()) != 0 {
		t.Fatalf("expected no logical children")
	}
	if len(parent.DisplayObject().Children()) != 0 {
		t.Fatalf("expected no display children")
	}
}
