package widgets

import "testing"

func TestTreeBasicInsertionAndExpansion(t *testing.T) {
	tree := NewTree()

	folder := NewTreeNode(true, "")
	folder.SetText("Folder")
	leaf := NewTreeNode(false, "")
	leaf.SetText("Leaf")

	tree.RootNode().AddChild(folder)
	folder.AddChild(leaf)
	folder.SetExpanded(true)

	if got := len(tree.items); got != 2 {
		t.Fatalf("expected 2 visible items, got %d", got)
	}

	if tree.indexOfNode(folder) != 0 {
		t.Fatalf("folder index mismatch")
	}
	if tree.indexOfNode(leaf) != 1 {
		t.Fatalf("leaf index mismatch")
	}

	folder.SetExpanded(false)
	if folder.Expanded() {
		t.Fatalf("folder should be collapsed")
	}
	if got := len(tree.items); got != 1 {
		t.Fatalf("expected 1 visible item after collapse, got %d", got)
	}

	folder.SetExpanded(true)
	if got := len(tree.items); got != 2 {
		t.Fatalf("expected 2 visible items after expand, got %d", got)
	}
}

func TestTreeSelection(t *testing.T) {
	tree := NewTree()
	folder := NewTreeNode(true, "")
	leaf := NewTreeNode(false, "")
	tree.RootNode().AddChild(folder)
	folder.AddChild(leaf)

	folder.SetExpanded(false)
	tree.SelectNode(leaf, false)

	if !folder.Expanded() {
		t.Fatalf("parent should be expanded after selecting child")
	}
	if sel := tree.GetSelectedNode(); sel != leaf {
		t.Fatalf("expected selected leaf, got %v", sel)
	}

	tree.UnselectNode(leaf)
	if sel := tree.GetSelectedNode(); sel != nil {
		t.Fatalf("expected no selection, got %v", sel)
	}
}
