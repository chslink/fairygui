//go:build ebiten

package render

import (
	"bytes"
	"context"
	"errors"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// AtlasManager loads and caches atlas textures and sprite images.
type AtlasManager struct {
	loader      assets.Loader
	atlasImages map[string]*ebiten.Image
	spriteCache map[string]*ebiten.Image
}

// NewAtlasManager creates a manager using the provided Loader.
func NewAtlasManager(loader assets.Loader) *AtlasManager {
	return &AtlasManager{
		loader:      loader,
		atlasImages: make(map[string]*ebiten.Image),
		spriteCache: make(map[string]*ebiten.Image),
	}
}

// LoadPackage ensures all atlas textures referenced by the package are loaded.
func (m *AtlasManager) LoadPackage(ctx context.Context, pkg *assets.Package) error {
	for _, item := range pkg.Items {
		if item.Type != assets.PackageItemTypeAtlas {
			continue
		}
		if _, ok := m.atlasImages[item.ID]; ok {
			continue
		}
		data, err := m.loader.LoadOne(ctx, item.File, assets.ResourceImage)
		if err != nil {
			return err
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return err
		}
		m.atlasImages[item.ID] = ebiten.NewImageFromImage(img)
	}
	return nil
}

// ResolveSprite returns an Ebiten image representing the sprite for the given item.
func (m *AtlasManager) ResolveSprite(item *assets.PackageItem) (any, error) {
	if item == nil || item.Sprite == nil || item.Sprite.Atlas == nil {
		return nil, errors.New("render: package item has no sprite data")
	}
	if sprite, ok := m.spriteCache[item.ID]; ok {
		return sprite, nil
	}
	atlasID := item.Sprite.Atlas.ID
	atlasImg, ok := m.atlasImages[atlasID]
	if !ok {
		return nil, errors.New("render: atlas texture not loaded")
	}
	rect := image.Rect(
		item.Sprite.Rect.X,
		item.Sprite.Rect.Y,
		item.Sprite.Rect.X+item.Sprite.Rect.Width,
		item.Sprite.Rect.Y+item.Sprite.Rect.Height,
	)
	sub := atlasImg.SubImage(rect)
	spriteImg := ebiten.NewImageFromImage(sub)
	m.spriteCache[item.ID] = spriteImg
	return spriteImg, nil
}
