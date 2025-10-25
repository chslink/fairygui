package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

func TestRenderMovieClipWidgetUsesSourceSize(t *testing.T) {
	pkg := &assets.Package{ID: "pkg"}
	atlasItem := &assets.PackageItem{
		ID:    "atlas",
		Type:  assets.PackageItemTypeAtlas,
		Owner: pkg,
	}
	item := &assets.PackageItem{
		ID:     "movie",
		Type:   assets.PackageItemTypeMovieClip,
		Owner:  pkg,
		Width:  20,
		Height: 18,
	}

	frame0 := &assets.MovieClipFrame{
		SpriteID: "frame0",
		Sprite: &assets.AtlasSprite{
			Atlas: atlasItem,
			Rect: assets.Rect{
				X:      0,
				Y:      0,
				Width:  10,
				Height: 10,
			},
		},
		OffsetX: 5,
		OffsetY: 3,
		Width:   10,
		Height:  10,
	}
	frame1 := &assets.MovieClipFrame{
		SpriteID: "frame1",
		Sprite: &assets.AtlasSprite{
			Atlas: atlasItem,
			Rect: assets.Rect{
				X:      10,
				Y:      0,
				Width:  8,
				Height: 10,
			},
		},
		OffsetX: 6,
		OffsetY: 4,
		Width:   8,
		Height:  10,
	}
	item.Frames = []*assets.MovieClipFrame{frame0, frame1}

	manager := NewAtlasManager(nil)
	atlasImg := ebiten.NewImage(32, 32)
	atlasImg.Clear()
	fillRect(atlasImg, frame0.Sprite.Rect, color.NRGBA{R: 255, A: 255})
	fillRect(atlasImg, frame1.Sprite.Rect, color.NRGBA{G: 255, A: 255})
	if err := manager.AddAtlasImage(atlasItem, atlasImg); err != nil {
		t.Fatalf("failed to add atlas image: %v", err)
	}

	clip := widgets.NewMovieClip()
	clip.SetPackageItem(item)
	clip.SetSize(20, 18)

	testCases := []struct {
		index int
		frame *assets.MovieClipFrame
	}{
		{index: 0, frame: frame0},
		{index: 1, frame: frame1},
	}

	for _, tc := range testCases {
		clip.SetFrame(tc.index)
		dst := ebiten.NewImage(64, 64)
		if err := renderMovieClipWidget(dst, clip, manager, ebiten.GeoM{}, 1, nil); err != nil {
			t.Fatalf("renderMovieClipWidget frame %d failed: %v", tc.index, err)
		}
		bounds, ok := alphaBounds(dst)
		if !ok {
			t.Fatalf("no pixels rendered for frame %d", tc.index)
		}
		if got := bounds.Dx(); got != tc.frame.Width {
			t.Fatalf("frame %d width mismatch: expected %d, got %d", tc.index, tc.frame.Width, got)
		}
		if got := bounds.Dy(); got != tc.frame.Height {
			t.Fatalf("frame %d height mismatch: expected %d, got %d", tc.index, tc.frame.Height, got)
		}
		if got := bounds.Min.X; got != tc.frame.OffsetX {
			t.Fatalf("frame %d offsetX mismatch: expected %d, got %d", tc.index, tc.frame.OffsetX, got)
		}
		if got := bounds.Min.Y; got != tc.frame.OffsetY {
			t.Fatalf("frame %d offsetY mismatch: expected %d, got %d", tc.index, tc.frame.OffsetY, got)
		}
	}
}

func fillRect(img *ebiten.Image, rect assets.Rect, c color.NRGBA) {
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			img.Set(x, y, c)
		}
	}
}

func alphaBounds(img *ebiten.Image) (image.Rectangle, bool) {
	if img == nil {
		return image.Rectangle{}, false
	}
	bounds := img.Bounds()
	if bounds.Empty() {
		return image.Rectangle{}, false
	}
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X-1, bounds.Min.Y-1
	found := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
			found = true
		}
	}
	if !found {
		return image.Rectangle{}, false
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), true
}
