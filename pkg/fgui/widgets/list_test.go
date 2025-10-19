package widgets

import "testing"

func TestListAccessors(t *testing.T) {
	list := NewList()
	list.SetDefaultItem("ui://pkg/item")
	list.SetResource("resourceID")
	if list.DefaultItem() != "ui://pkg/item" {
		t.Fatalf("unexpected default item: %s", list.DefaultItem())
	}
	if list.Resource() != "resourceID" {
		t.Fatalf("unexpected resource: %s", list.Resource())
	}
}
