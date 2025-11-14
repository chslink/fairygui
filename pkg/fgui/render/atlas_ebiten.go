package render

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// AtlasManager loads and caches atlas textures and sprite images.
// 借鉴 Unity 版本的 MaterialManager 设计思想，扩展支持 DrawParams 缓存
type AtlasManager struct {
	loader      assets.Loader
	atlasImages map[string]*ebiten.Image
	spriteCache map[string]*ebiten.Image
	movieCache  map[string]*ebiten.Image

	// drawParamsCache 缓存 DrawImageOptions，减少对象分配
	// Key: image + color + blend + filter 的组合
	drawParamsCache map[string]*DrawParams
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

// AddAtlasImage manually adds an atlas image to the manager.
func (m *AtlasManager) AddAtlasImage(item *assets.PackageItem, img *ebiten.Image) error {
	if item == nil {
		return errors.New("render: nil package item")
	}
	if img == nil {
		return errors.New("render: nil image")
	}
	key := atlasKey(item)
	m.atlasImages[key] = img
	return nil
}

// GetAtlasImage returns the loaded atlas texture for the given PackageItem.
// This is used for BMFont atlas mode where glyphs are extracted from a single atlas texture.
func (m *AtlasManager) GetAtlasImage(item *assets.PackageItem) (*ebiten.Image, error) {
	if item == nil {
		return nil, errors.New("render: nil package item")
	}
	key := atlasKey(item)
	img, ok := m.atlasImages[key]
	if !ok {
		return nil, fmt.Errorf("render: atlas texture not loaded for %s", key)
	}
	return img, nil
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

// ResolveMovieClipFrameAligned returns an Ebiten image for the movie clip frame with alignment correction.
// This creates a corrected image where the frame is positioned according to its offset within the MovieClip bounds.
// The targetWidth and targetHeight parameters allow for scaled rendering while maintaining alignment.
func (m *AtlasManager) ResolveMovieClipFrameAligned(item *assets.PackageItem, frame *assets.MovieClipFrame, movieClipWidth, movieClipHeight int) (*ebiten.Image, error) {
	if frame == nil || frame.Sprite == nil || frame.Sprite.Atlas == nil {
		return nil, errors.New("render: movie clip frame missing sprite data")
	}

	// Create a unique key that includes the MovieClip dimensions for alignment
	key := fmt.Sprintf("%s:aligned:%dx%d", movieClipFrameKey(item, frame), movieClipWidth, movieClipHeight)
	if img, ok := m.movieCache[key]; ok {
		return img, nil
	}

	// Get the base frame image
	baseImg, err := m.ResolveMovieClipFrame(item, frame)
	if err != nil {
		return nil, err
	}

	// Create a new image with MovieClip dimensions
	alignedImg := ebiten.NewImage(movieClipWidth, movieClipHeight)

	// Calculate the position to draw the frame within the MovieClip bounds
	// The offset defines how the frame is positioned relative to the MovieClip origin
	drawX := frame.OffsetX
	drawY := frame.OffsetY

	// Apply sprite offset if available (additional fine-tuning)
	if frame.Sprite != nil {
		off := frame.Sprite.Offset
		drawX += int(off.X)
		drawY += int(off.Y)
	}

	// Draw the frame at the calculated position
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(drawX), float64(drawY))

	alignedImg.DrawImage(baseImg, opts)

	m.movieCache[key] = alignedImg
	return alignedImg, nil
}

// ResolveMovieClipFrameScaled returns an Ebiten image for the movie clip frame with scaling support.
// This handles cases where the MovieClip is displayed at a different size than its original dimensions.
func (m *AtlasManager) ResolveMovieClipFrameScaled(item *assets.PackageItem, frame *assets.MovieClipFrame, displayWidth, displayHeight int) (*ebiten.Image, error) {
	if frame == nil || frame.Sprite == nil || frame.Sprite.Atlas == nil {
		return nil, errors.New("render: movie clip frame missing sprite data")
	}

	// Create a unique key that includes the display dimensions for scaling
	key := fmt.Sprintf("%s:scaled:%dx%d", movieClipFrameKey(item, frame), displayWidth, displayHeight)
	if img, ok := m.movieCache[key]; ok {
		return img, nil
	}

	// Get the base frame image
	baseImg, err := m.ResolveMovieClipFrame(item, frame)
	if err != nil {
		return nil, err
	}
	if baseImg == nil {
		return nil, errors.New("render: base frame image is nil")
	}

	// Create a new image with the display dimensions
	scaledImg := ebiten.NewImage(displayWidth, displayHeight)
	if scaledImg == nil {
		return nil, errors.New("render: failed to create scaled image")
	}

	// Calculate scale factors
	scaleX := float64(displayWidth) / float64(item.Width)
	scaleY := float64(displayHeight) / float64(item.Height)

	// Calculate the scaled position for the frame within the display bounds
	// The offset is scaled along with the frame positioning
	scaledOffsetX := float64(frame.OffsetX) * scaleX
	scaledOffsetY := float64(frame.OffsetY) * scaleY

	// Apply sprite offset if available (additional fine-tuning)
	if frame.Sprite != nil {
		off := frame.Sprite.Offset
		scaledOffsetX += float64(off.X) * scaleX
		scaledOffsetY += float64(off.Y) * scaleY
	}

	// Debug info for n15
	if item != nil && item.ID == "hixt1v" {
		log.Printf("[n15-scaled] Creating scaled frame: display=%dx%d original=%dx%d scale=%.2fx%.2f offset=%.1f,%.1f baseBounds=%v",
			displayWidth, displayHeight, item.Width, item.Height, scaleX, scaleY,
			scaledOffsetX, scaledOffsetY, baseImg.Bounds())
	}

	// Draw the frame at the calculated scaled position with scaling applied
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scaleX, scaleY)
	opts.GeoM.Translate(scaledOffsetX, scaledOffsetY)

	scaledImg.DrawImage(baseImg, opts)

	m.movieCache[key] = scaledImg
	return scaledImg, nil
}

func movieClipFrameKey(item *assets.PackageItem, frame *assets.MovieClipFrame) string {
	ownerID := ""
	if item != nil && item.Owner != nil {
		ownerID = item.Owner.ID
	}
	return fmt.Sprintf("mc:%s:%s:%d:%d:%d:%d", ownerID, frame.SpriteID, frame.Sprite.Rect.X, frame.Sprite.Rect.Y, frame.Sprite.Rect.Width, frame.Sprite.Rect.Height)
}

// DrawParams 封装绘制参数，借鉴 Unity MaterialManager 设计
// 用于减少 DrawImageOptions 对象分配
type DrawParams struct {
	Image      *ebiten.Image
	ColorScale ebiten.ColorScale
	Blend      ebiten.Blend
	Filter     ebiten.Filter
}

// GetDrawParams 从缓存获取或创建绘制参数
// 借鉴 Unity 版本的 MaterialManager.GetMaterial()
func (m *AtlasManager) GetDrawParams(img *ebiten.Image, colorScale ebiten.ColorScale, blend ebiten.Blend, filter ebiten.Filter) *DrawParams {
	if m.drawParamsCache == nil {
		m.drawParamsCache = make(map[string]*DrawParams)
	}

	// 生成缓存键（类似 Unity 的多维键值）
	key := m.generateDrawParamsKey(img, colorScale, blend, filter)

	// 尝试从缓存获取
	if params, ok := m.drawParamsCache[key]; ok {
		return params
	}

	// 缓存未命中，创建新参数
	params := &DrawParams{
		Image:      img,
		ColorScale: colorScale,
		Blend:      blend,
		Filter:     filter,
	}

	m.drawParamsCache[key] = params
	return params
}

// generateDrawParamsKey 生成 DrawParams 的缓存键
func (m *AtlasManager) generateDrawParamsKey(img *ebiten.Image, colorScale ebiten.ColorScale, blend ebiten.Blend, filter ebiten.Filter) string {
	// 使用图像指针和参数组合作为键
	return fmt.Sprintf("%p_%v_%v_%v", img, colorScale, blend, filter)
}
