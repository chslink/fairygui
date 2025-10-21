package widgets

import (
	"math"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestLoaderDefaults(t *testing.T) {
	loader := NewLoader()
	if loader == nil || loader.GObject == nil {
		t.Fatalf("expected GLoader to wrap GObject")
	}
	if !loader.Playing() {
		t.Fatalf("expected loader to default to playing=true")
	}
	if loader.Color() != "#ffffff" {
		t.Fatalf("expected default color #ffffff, got %s", loader.Color())
	}
}

func TestLoaderAutoSizeFromSprite(t *testing.T) {
	loader := NewLoader()
	loader.SetAutoSize(true)
	item := &assets.PackageItem{
		Sprite: &assets.AtlasSprite{
			Rect: assets.Rect{Width: 40, Height: 20},
			OriginalSize: assets.Point{
				X: 80, Y: 50,
			},
		},
	}
	loader.SetPackageItem(item)

	if loader.Width() != 80 || loader.Height() != 50 {
		t.Fatalf("expected auto size (80,50), got (%v,%v)", loader.Width(), loader.Height())
	}
}

func TestLoaderComponentAttachment(t *testing.T) {
	loader := NewLoader()
	component := core.NewGComponent()
	component.SetSize(120, 60)

	loader.SetAutoSize(true)
	loader.SetComponent(component)

	if loader.Component() != component {
		t.Fatalf("component not stored on loader")
	}
	if loader.Width() != 120 || loader.Height() != 60 {
		t.Fatalf("loader should adopt component size, got (%v,%v)", loader.Width(), loader.Height())
	}
	if loader.DisplayObject() == nil || len(loader.DisplayObject().Children()) == 0 {
		t.Fatalf("expected loader display to contain component display")
	}
}

func TestLoaderFillAndAlign(t *testing.T) {
	item := &assets.PackageItem{
		Sprite: &assets.AtlasSprite{
			Rect: assets.Rect{Width: 50, Height: 25},
		},
	}

	loader := NewLoader()
	loader.SetPackageItem(item)
	loader.SetAlign(LoaderAlignCenter)
	loader.SetVerticalAlign(LoaderAlignMiddle)
	loader.SetFill(LoaderFillNone)

	loader.SetAutoSize(false)
	loader.GObject.SetSize(100, 100)
	loader.RefreshLayout()

	ox, oy := loader.ContentOffset()
	if ox != 25 || oy != 37.5 {
		t.Fatalf("unexpected offset (%v,%v)", ox, oy)
	}
	sx, sy := loader.ContentScale()
	if sx != 1 || sy != 1 {
		t.Fatalf("scale should remain 1 for default fill, got (%v,%v)", sx, sy)
	}

	loader.SetFill(LoaderFillScale)
	loader.RefreshLayout()
	sx, sy = loader.ContentScale()
	if sx != 2 || sy != 2 {
		t.Fatalf("expected scale factors 2 for scale fill, got (%v,%v)", sx, sy)
	}
	ox, oy = loader.ContentOffset()
	if ox != 0 || oy != 25 {
		t.Fatalf("unexpected offset after scale fill (%v,%v)", ox, oy)
	}
}

func TestLoaderComponentAlignAppliesTransform(t *testing.T) {
	comp := core.NewGComponent()
	comp.SetSize(40, 20)

	loader := NewLoader()
	loader.SetComponent(comp)
	loader.GObject.SetSize(80, 40)
	loader.SetAlign(LoaderAlignRight)
	loader.SetVerticalAlign(LoaderAlignBottom)
	loader.SetUseResize(false)
	loader.RefreshLayout()

	if comp.X() != 40 || comp.Y() != 20 {
		t.Fatalf("unexpected component position (%v,%v)", comp.X(), comp.Y())
	}
	sx, sy := comp.Scale()
	if sx != 1 || sy != 1 {
		t.Fatalf("component scale should remain 1 when no fill, got (%v,%v)", sx, sy)
	}
}

func TestLoaderPlaybackAndColor(t *testing.T) {
	loader := NewLoader()
	loader.SetPlaying(false)
	loader.SetFrame(5)
	loader.SetColor("#ff0000")
	loader.SetFillMethod(3)
	loader.SetFillOrigin(2)
	loader.SetFillClockwise(true)
	loader.SetFillAmount(0.42)

	if loader.Playing() {
		t.Fatalf("expected playing=false")
	}
	if loader.Frame() != 5 {
		t.Fatalf("unexpected frame %d", loader.Frame())
	}
	if loader.Color() != "#ff0000" {
		t.Fatalf("unexpected color %s", loader.Color())
	}
	if loader.FillMethod() != 3 || loader.FillOrigin() != 2 || !loader.FillClockwise() {
		t.Fatalf("unexpected fill settings")
	}
	if math.Abs(loader.FillAmount()-0.42) > 1e-6 {
		t.Fatalf("unexpected fill amount %v", loader.FillAmount())
	}
}
