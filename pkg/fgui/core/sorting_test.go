package core

import (
	"testing"
)

// TestSortingOrder verifies that children with sortingOrder are positioned correctly.
func TestSortingOrder(t *testing.T) {
	comp := NewGComponent()

	// Create children with different sorting orders
	child1 := NewGObject()
	child1.SetName("child1")
	child1.SetSortingOrder(0) // Normal order

	child2 := NewGObject()
	child2.SetName("child2")
	child2.SetSortingOrder(0) // Normal order

	child3 := NewGObject()
	child3.SetName("child3")
	child3.SetSortingOrder(100) // High priority

	child4 := NewGObject()
	child4.SetName("child4")
	child4.SetSortingOrder(50) // Medium priority

	// Add children
	comp.AddChild(child1)
	comp.AddChild(child2)
	comp.AddChild(child3)
	comp.AddChild(child4)

	// Expected order: child1, child2, child4 (sort=50), child3 (sort=100)
	children := comp.Children()
	if len(children) != 4 {
		t.Fatalf("expected 4 children, got %d", len(children))
	}

	if children[0].Name() != "child1" {
		t.Errorf("children[0] should be child1, got %s", children[0].Name())
	}
	if children[1].Name() != "child2" {
		t.Errorf("children[1] should be child2, got %s", children[1].Name())
	}
	if children[2].Name() != "child4" {
		t.Errorf("children[2] should be child4 (sort=50), got %s", children[2].Name())
	}
	if children[3].Name() != "child3" {
		t.Errorf("children[3] should be child3 (sort=100), got %s", children[3].Name())
	}

	// Verify sortingChildCount
	if comp.sortingChildCount != 2 {
		t.Errorf("expected sortingChildCount=2, got %d", comp.sortingChildCount)
	}
}

// TestSortingOrderChange verifies that changing sortingOrder repositions the child.
func TestSortingOrderChange(t *testing.T) {
	comp := NewGComponent()

	child1 := NewGObject()
	child1.SetName("child1")

	child2 := NewGObject()
	child2.SetName("child2")
	child2.SetSortingOrder(100)

	child3 := NewGObject()
	child3.SetName("child3")
	child3.SetSortingOrder(200)

	comp.AddChild(child1)
	comp.AddChild(child2)
	comp.AddChild(child3)

	// Initial order: child1, child2 (100), child3 (200)
	children := comp.Children()
	if children[0].Name() != "child1" || children[1].Name() != "child2" || children[2].Name() != "child3" {
		t.Fatalf("initial order wrong: %s, %s, %s", children[0].Name(), children[1].Name(), children[2].Name())
	}

	// Change child3's sortingOrder to 50 (should move before child2)
	child3.SetSortingOrder(50)

	// New order: child1, child3 (50), child2 (100)
	children = comp.Children()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}
	if children[0].Name() != "child1" {
		t.Errorf("children[0] should be child1, got %s", children[0].Name())
	}
	if children[1].Name() != "child3" {
		t.Errorf("children[1] should be child3 (sort changed to 50), got %s", children[1].Name())
	}
	if children[2].Name() != "child2" {
		t.Errorf("children[2] should be child2 (sort=100), got %s", children[2].Name())
	}

	// Change child2's sortingOrder to 0 (should move to front)
	child2.SetSortingOrder(0)

	// New order: child1, child2 (0), child3 (50)
	children = comp.Children()
	if children[0].Name() != "child1" {
		t.Errorf("children[0] should be child1, got %s", children[0].Name())
	}
	if children[1].Name() != "child2" {
		t.Errorf("children[1] should be child2 (sort changed to 0), got %s", children[1].Name())
	}
	if children[2].Name() != "child3" {
		t.Errorf("children[2] should be child3 (sort=50), got %s", children[2].Name())
	}

	// Verify sortingChildCount updated correctly
	if comp.sortingChildCount != 1 {
		t.Errorf("expected sortingChildCount=1 after child2 changed to 0, got %d", comp.sortingChildCount)
	}
}

// TestRemoveChildUpdatesSortingCount verifies that removing a child updates sortingChildCount.
func TestRemoveChildUpdatesSortingCount(t *testing.T) {
	comp := NewGComponent()

	child1 := NewGObject()
	child1.SetSortingOrder(100)

	child2 := NewGObject()
	child2.SetSortingOrder(200)

	child3 := NewGObject()
	child3.SetSortingOrder(0)

	comp.AddChild(child1)
	comp.AddChild(child2)
	comp.AddChild(child3)

	if comp.sortingChildCount != 2 {
		t.Fatalf("expected sortingChildCount=2, got %d", comp.sortingChildCount)
	}

	// Remove a sorting child
	comp.RemoveChild(child1)

	if comp.sortingChildCount != 1 {
		t.Errorf("expected sortingChildCount=1 after removing sorting child, got %d", comp.sortingChildCount)
	}

	// Remove a non-sorting child
	comp.RemoveChild(child3)

	if comp.sortingChildCount != 1 {
		t.Errorf("expected sortingChildCount=1 after removing non-sorting child, got %d", comp.sortingChildCount)
	}

	// Remove last sorting child
	comp.RemoveChild(child2)

	if comp.sortingChildCount != 0 {
		t.Errorf("expected sortingChildCount=0 after removing all sorting children, got %d", comp.sortingChildCount)
	}
}
