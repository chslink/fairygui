package core

import "testing"

func TestControllerSetPagesDefaultsToFirstPage(t *testing.T) {
	ctrl := NewController("demo")
	ctrl.SetPages([]string{"id-0", "id-1"}, []string{"page0", "page1"})

	if idx := ctrl.SelectedIndex(); idx != 0 {
		t.Fatalf("expected controller to select first page, got %d", idx)
	}
}

func TestControllerClampNegativeSelection(t *testing.T) {
	ctrl := NewController("clamp")
	ctrl.SetPages([]string{"id-0", "id-1"}, []string{"page0", "page1"})

	ctrl.SetSelectedIndex(-1)
	if idx := ctrl.SelectedIndex(); idx != 0 {
		t.Fatalf("expected negative selection to clamp to 0, got %d", idx)
	}

	ctrl.SetSelectedIndex(5)
	if idx := ctrl.SelectedIndex(); idx != 1 {
		t.Fatalf("expected selection to clamp to last index, got %d", idx)
	}
}

func TestControllerAllowsNegativeWhenNoPages(t *testing.T) {
	ctrl := NewController("empty")
	ctrl.SetSelectedIndex(0)
	if idx := ctrl.SelectedIndex(); idx != -1 {
		t.Fatalf("expected selection to remain -1 when no pages, got %d", idx)
	}
}
