package render

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	movieCache  map[string]*ebiten.Image
}

// NewAtlasManager creates a manager using the provided Loader.
func NewAtlasManager(loader assets.Loader) *AtlasManager {
	return &AtlasManager{
		loader:      loader,
		atlasImages: make(map[string]*ebiten.Image),
		spriteCache: make(map[string]*ebiten.Image),
		movieCache:  make(map[string]*ebiten.Image),
	}
}

// LoadPackage ensures all atlas textures referenced by the package are loaded.
func (m *AtlasManager) LoadPackage(ctx context.Context, pkg *assets.Package) error {
	for _, item := range pkg.Items {
		if item.Type != assets.PackageItemTypeAtlas {
			continue
		}
		key := atlasKey(item)
		if _, ok := m.atlasImages[key]; ok {
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
		m.atlasImages[key] = ebiten.NewImageFromImage(img)
	}
	return nil
}

// ResolveSprite returns an Ebiten image representing the sprite for the given item.
func (m *AtlasManager) ResolveSprite(item *assets.PackageItem) (any, error) {
	if item == nil || item.Sprite == nil || item.Sprite.Atlas == nil {
		return nil, errors.New("render: package item has no sprite data")
	}
	if sprite, ok := m.spriteCache[spriteKey(item)]; ok {
		return sprite, nil
	}
	atlasImg, ok := m.atlasImages[atlasKey(item.Sprite.Atlas)]
	if !ok {
		return nil, errors.New("render: atlas texture not loaded")
	}
	rect := image.Rect(
		item.Sprite.Rect.X,
		item.Sprite.Rect.Y,
		item.Sprite.Rect.X+item.Sprite.Rect.Width,
		item.Sprite.Rect.Y+item.Sprite.Rect.Height,
	)
	atlasBounds := atlasImg.Bounds()
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		return nil, fmt.Errorf("render: sprite %s has invalid rect %v", item.ID, rect)
	}
	if !rect.In(atlasBounds) {
		rect = rect.Intersect(atlasBounds)
		if rect.Dx() <= 0 || rect.Dy() <= 0 {
			return nil, fmt.Errorf("render: sprite %s rect out of atlas bounds %v", item.ID, atlasBounds)
		}
	}
	sub := atlasImg.SubImage(rect)
	spriteImg := ebiten.NewImageFromImage(sub)
	m.spriteCache[spriteKey(item)] = spriteImg
	return spriteImg, nil
}

func atlasKey(item *assets.PackageItem) string {
	if item == nil {
		return ""
	}
	ownerID := ""
	if item.Owner != nil {
		ownerID = item.Owner.ID
	}
	return ownerID + ":" + item.ID
}

func spriteKey(item *assets.PackageItem) string {
	if item == nil {
		return ""
	}
	ownerID := ""
	if item.Owner != nil {
		ownerID = item.Owner.ID
	}
	return ownerID + ":" + item.ID
}

// ResolveMovieClipFrame returns an Ebiten image for the supplied movie clip frame.
func (m *AtlasManager) ResolveMovieClipFrame(item *assets.PackageItem, frame *assets.MovieClipFrame) (*ebiten.Image, error) {
	if frame == nil || frame.Sprite == nil || frame.Sprite.Atlas == nil {
		return nil, errors.New("render: movie clip frame missing sprite data")
	}
	key := movieClipFrameKey(item, frame)
	if img, ok := m.movieCache[key]; ok {
		return img, nil
	}
	atlasImg, ok := m.atlasImages[atlasKey(frame.Sprite.Atlas)]
	if !ok {
		return nil, errors.New("render: atlas texture not loaded")
	}
	rect := image.Rect(
		frame.Sprite.Rect.X,
		frame.Sprite.Rect.Y,
		frame.Sprite.Rect.X+frame.Sprite.Rect.Width,
		frame.Sprite.Rect.Y+frame.Sprite.Rect.Height,
	)
	atlasBounds := atlasImg.Bounds()
	if !rect.In(atlasBounds) {
		rect = rect.Intersect(atlasBounds)
		if rect.Empty() {
			return nil, fmt.Errorf("render: movie clip frame rect out of bounds %v", atlasBounds)
		}
	}
	sub, ok := atlasImg.SubImage(rect).(*ebiten.Image)
	if !ok {
		return nil, errors.New("render: atlas sub-image type mismatch")
	}
	m.movieCache[key] = sub
	return sub, nil
}

func movieClipFrameKey(item *assets.PackageItem, frame *assets.MovieClipFrame) string {
	ownerID := ""
	if item != nil && item.Owner != nil {
		ownerID = item.Owner.ID
	}
	return fmt.Sprintf("mc:%s:%s:%d:%d:%d:%d", ownerID, frame.SpriteID, frame.Sprite.Rect.X, frame.Sprite.Rect.Y, frame.Sprite.Rect.Width, frame.Sprite.Rect.Height)
}
