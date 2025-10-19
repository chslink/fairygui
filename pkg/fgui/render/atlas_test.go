//go:build ebiten

package render

import (
	"context"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

func TestAtlasManagerWithRealFUI(t *testing.T) {
	root := filepath.Join("..", "..", "demo", "assets")
	fuiPath := filepath.Join(root, "Bag.fui")
	data, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("demo assets not available: %v", err)
	}

	pkg, err := assets.ParsePackage(data, filepath.Join(root, "Bag"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	manager := NewAtlasManager(assets.NewFileLoader(root))
	if err := manager.LoadPackage(context.Background(), pkg); err != nil {
		t.Fatalf("LoadPackage failed: %v", err)
	}

	var imageItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeImage && item.Sprite != nil && item.Sprite.Atlas != nil {
			imageItem = item
			break
		}
	}
	if imageItem == nil {
		t.Skip("no image items with sprite data in package")
	}

	spriteImg, err := manager.ResolveSprite(imageItem)
	if err != nil {
		t.Fatalf("ResolveSprite failed: %v", err)
	}
	if spriteImg == nil {
		t.Fatalf("expected sprite image")
	}
	rect := spriteImg.(image.Image).Bounds()
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		t.Fatalf("invalid sprite bounds: %v", rect)
	}

}
