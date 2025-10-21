package core

import (
	"math"
	"testing"
)

func TestRelationsAddAndRemove(t *testing.T) {
	owner := NewGObject()
	target := NewGObject()

	owner.AddRelation(target, RelationTypeLeft_Left, false)
	owner.AddRelation(target, RelationTypeLeft_Left, false) // duplicate ignored
	owner.AddRelation(target, RelationTypeSize, true)       // expands to width/height

	rels := owner.Relations()
	if rels == nil {
		t.Fatalf("expected relations to be created")
	}
	items := rels.Items()
	if len(items) != 1 {
		t.Fatalf("expected one relation item, got %d", len(items))
	}
	item := items[0]
	if item.Target() != target {
		t.Fatalf("relation target mismatch")
	}
	if item.IsEmpty() {
		t.Fatalf("relation item should contain definitions")
	}

	owner.RemoveRelation(target, RelationTypeWidth)
	owner.RemoveRelation(target, RelationTypeHeight)
	owner.RemoveRelation(target, RelationTypeLeft_Left)
	if !item.IsEmpty() {
		t.Fatalf("expected relation item to be empty after removals")
	}

	owner.RemoveRelations(target)
	if len(owner.Relations().Items()) != 0 {
		t.Fatalf("expected relations slice to be cleared")
	}
}

func TestRelationsClearFor(t *testing.T) {
	owner := NewGObject()
	targetA := NewGObject()
	targetB := NewGObject()

	owner.AddRelation(targetA, RelationTypeLeft_Left, false)
	owner.AddRelation(targetB, RelationTypeTop_Top, false)

	rels := owner.Relations()
	if len(rels.Items()) != 2 {
		t.Fatalf("expected two relation items")
	}

	rels.ClearFor(targetA)
	items := rels.Items()
	if len(items) != 1 {
		t.Fatalf("expected single relation item after clearFor")
	}
	if items[0].Target() != targetB {
		t.Fatalf("remaining item should point to targetB")
	}
}

func TestRelationItemTargetXYAdjustsOwner(t *testing.T) {
	target := NewGObject()
	owner := NewGObject()
	owner.AddRelation(target, RelationTypeLeft_Left, false)
	owner.AddRelation(target, RelationTypeTop_Top, false)

	owner.SetPosition(5, 5)
	target.SetPosition(10, 20)

	if owner.X() != 15 {
		t.Fatalf("expected owner x to shift with target, got %.1f", owner.X())
	}
	if owner.Y() != 25 {
		t.Fatalf("expected owner y to shift with target, got %.1f", owner.Y())
	}
}

func TestRelationItemTargetSizePercentAdjustsOwner(t *testing.T) {
	target := NewGObject()
	target.SetSize(100, 100)
	owner := NewGObject()
	owner.SetSize(50, 50)
	owner.AddRelation(target, RelationTypeLeft_Left, true)
	owner.SetPosition(10, 0)
	start := owner.xMin()
	target.SetSize(200, 100)
	if owner.xMin() != start*2 {
		t.Fatalf("expected owner xMin scaled to %.1f, got %.1f", start*2, owner.xMin())
	}
}

func TestRelationWidthExtAdjustsOwner(t *testing.T) {
	target := NewGObject()
	target.SetSize(100, 100)
	owner := NewGObject()
	owner.SetSize(50, 50)
	owner.AddRelation(target, RelationTypeRightExt_Right, false)
	target.SetSize(120, 100)
	if owner.Width() != 70 {
		t.Fatalf("expected owner width to grow to 70, got %.1f", owner.Width())
	}
}

func TestRelationWidthPercentAgainstParent(t *testing.T) {
	parent := NewGComponent()
	parent.SetSize(200, 100)
	parent.SetSourceSize(200, 100)
	parent.SetInitSize(200, 100)
	child := NewGObject()
	child.SetSize(100, 40)
	child.SetSourceSize(100, 40)
	child.SetInitSize(100, 40)
	parent.AddChild(child)
	child.AddRelation(parent.GObject, RelationTypeWidth, true)
	parent.SetSize(300, 100)
	if math.Abs(child.Width()-150) > 0.001 {
		t.Fatalf("expected child width to scale to 150, got %.3f", child.Width())
	}
}

func TestRelationRightExtPercentParent(t *testing.T) {
	parent := NewGComponent()
	parent.SetSize(200, 100)
	parent.SetSourceSize(200, 100)
	parent.SetInitSize(200, 100)
	child := NewGObject()
	child.SetPosition(50, 0)
	child.SetSize(100, 40)
	child.SetSourceSize(100, 40)
	child.SetInitSize(100, 40)
	parent.AddChild(child)
	parent.GObject.AddRelation(child, RelationTypeRightExt_Right, true)
	child.SetSize(150, 40)
	if math.Abs(parent.Width()-275) > 0.001 {
		t.Fatalf("expected parent width to adjust to 275, got %.3f", parent.Width())
	}
}
