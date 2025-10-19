package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

func TestBuildComponentFromRealFUI(t *testing.T) {
	root := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(root, "Bag.fui"))
	if err != nil {
		t.Skipf("demo assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(root, "Bag"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	var component *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Component != nil {
			component = item
			break
		}
	}
	if component == nil {
		t.Fatalf("no component items found")
	}

	factory := NewFactory(nil)
	built, err := factory.BuildComponent(context.Background(), pkg, component)
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	if len(built.Children()) != len(component.Component.Children) {
		t.Fatalf("expected %d children, got %d", len(component.Component.Children), len(built.Children()))
	}

	var imageIndex int = -1
	for idx, meta := range component.Component.Children {
		if meta.Type == assets.ObjectTypeImage && meta.Src != "" {
			imageIndex = idx
			break
		}
	}

	if imageIndex >= 0 {
		builtChild := built.ChildAt(imageIndex)
		if builtChild == nil {
			t.Fatalf("image child missing at index %d", imageIndex)
		}
		meta := component.Component.Children[imageIndex]
		if builtChild.X() != float64(meta.X) || builtChild.Y() != float64(meta.Y) {
			t.Fatalf("image child position mismatch: got (%v,%v) expected (%d,%d)", builtChild.X(), builtChild.Y(), meta.X, meta.Y)
		}
		data, ok := builtChild.Data().(*assets.PackageItem)
		if !ok || data == nil {
			t.Fatalf("expected image child to reference package item")
		}
		expectedWidth := meta.Width
		expectedHeight := meta.Height
		if expectedWidth < 0 && data.Sprite != nil {
			expectedWidth = data.Sprite.Rect.Width
		}
		if expectedHeight < 0 && data.Sprite != nil {
			expectedHeight = data.Sprite.Rect.Height
		}
		if expectedWidth >= 0 && builtChild.Width() != float64(expectedWidth) {
			t.Fatalf("expected width %d, got %v", expectedWidth, builtChild.Width())
		}
		if expectedHeight >= 0 && builtChild.Height() != float64(expectedHeight) {
			t.Fatalf("expected height %d, got %v", expectedHeight, builtChild.Height())
		}
	}

	var textIndex int = -1
	for idx, meta := range component.Component.Children {
		if (meta.Type == assets.ObjectTypeText || meta.Type == assets.ObjectTypeRichText) && meta.Text != "" {
			textIndex = idx
			break
		}
	}
	if textIndex >= 0 {
		builtChild := built.ChildAt(textIndex)
		if builtChild == nil {
			t.Fatalf("text child missing at index %d", textIndex)
		}
		meta := component.Component.Children[textIndex]
		data, ok := builtChild.Data().(string)
		if !ok {
			t.Fatalf("expected text child to store string data")
		}
		if data != meta.Text {
			t.Fatalf("expected text %q, got %q", meta.Text, data)
		}
	}

	var buttonIndex int = -1
	for idx, meta := range component.Component.Children {
		if meta.Type == assets.ObjectTypeButton {
			buttonIndex = idx
			break
		}
	}
	if buttonIndex >= 0 {
		builtChild := built.ChildAt(buttonIndex)
		if builtChild == nil {
			t.Fatalf("button child missing at index %d", buttonIndex)
		}
		if builtChild.Data() == nil {
			t.Fatalf("expected button child to store package item reference")
		}
	}

	var loaderIndex int = -1
	for idx, meta := range component.Component.Children {
		if meta.Type == assets.ObjectTypeLoader {
			loaderIndex = idx
			break
		}
	}
	if loaderIndex >= 0 {
		builtChild := built.ChildAt(loaderIndex)
		if builtChild == nil {
			t.Fatalf("loader child missing at index %d", loaderIndex)
		}
		if builtChild.Data() == nil {
			t.Fatalf("expected loader child to store resource reference")
		}
	}

	if len(component.Component.Controllers) > 0 {
		controllers := built.Controllers()
		if len(controllers) != len(component.Component.Controllers) {
			t.Fatalf("expected %d controllers, got %d", len(component.Component.Controllers), len(controllers))
		}
		if controllers[0].Name != component.Component.Controllers[0].Name {
			t.Fatalf("controller name mismatch")
		}
	}
}
