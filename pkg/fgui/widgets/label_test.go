package widgets

import (
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

func TestLabelAccessors(t *testing.T) {
	label := NewLabel()
	if label.Title() != "" || label.Icon() != "" {
		t.Fatalf("expected empty initial label state")
	}
	label.SetTitle("Hello")
	label.SetIcon("ui://package/icon")
	item := &assets.PackageItem{ID: "icon"}
	label.SetIconItem(item)
	label.SetResource("label-resource")
	if label.Title() != "Hello" {
		t.Fatalf("unexpected title: %s", label.Title())
	}
	if label.Icon() != "ui://package/icon" {
		t.Fatalf("unexpected icon: %s", label.Icon())
	}
	if label.IconItem() != item {
		t.Fatalf("expected icon item to be stored")
	}
	if label.Resource() != "label-resource" {
		t.Fatalf("unexpected resource: %s", label.Resource())
	}
}
