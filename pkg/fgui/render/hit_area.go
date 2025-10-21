package render

import (
	"math"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// ConfigureComponentHitArea wires mask and hit-test metadata to the component's display object.
// Pixel hit-tests honour component scaling, child hit-tests defer to the referenced child,
// and masks can further constrain (or invert) the accepted region. When no custom tester is
// required the sprite falls back to its default width/height bounds.
func ConfigureComponentHitArea(comp *core.GComponent, hit core.HitTest, pixel *assets.PixelHitTestData) {
	if comp == nil {
		return
	}
	sprite := comp.DisplayObject()
	if sprite == nil {
		return
	}

	// Reset any previous tester so the sprite falls back to its default behaviour when needed.
	sprite.SetHitTester(nil)

	var tester func(x, y float64) bool

	switch hit.Mode {
	case core.HitTestModePixel:
		tester = pixelHitTester(comp, pixel, hit.OffsetX, hit.OffsetY)
	case core.HitTestModeChild:
		tester = childHitTester(comp, hit.ChildIndex)
	default:
		// leave tester nil so default bounds are used unless a mask overrides it.
	}

	mask, reversed := comp.Mask()
	if maskTester := maskHitTester(comp, mask); maskTester != nil {
		if maskSprite := mask.DisplayObject(); maskSprite != nil {
			maskSprite.SetHitTester(func(x, y float64) bool { return false })
			maskSprite.SetMouseEnabled(false)
		}
		tester = combineMaskTester(tester, maskTester, reversed)
	} else if mask != nil && mask.DisplayObject() != nil {
		mask.DisplayObject().SetHitTester(nil)
		mask.DisplayObject().SetMouseEnabled(true)
	}

	if tester != nil {
		sprite.SetHitTester(tester)
		return
	}

	if comp.Opaque() {
		sprite.SetHitTester(func(x, y float64) bool {
			return x >= 0 && y >= 0 && x <= comp.Width() && y <= comp.Height()
		})
	}
}

func pixelHitTester(comp *core.GComponent, data *assets.PixelHitTestData, offsetX, offsetY int) func(x, y float64) bool {
	if data == nil {
		return nil
	}
	scaleX := componentAxisScale(comp.Width(), comp.SourceWidth())
	scaleY := componentAxisScale(comp.Height(), comp.SourceHeight())
	if scaleX == 0 {
		scaleX = 1
	}
	if scaleY == 0 {
		scaleY = 1
	}
	ox := float64(offsetX)
	oy := float64(offsetY)
	return func(x, y float64) bool {
		localX := x/scaleX - ox
		localY := y/scaleY - oy
		return data.Contains(localX, localY)
	}
}

func childHitTester(comp *core.GComponent, index int) func(x, y float64) bool {
	child := comp.ChildAt(index)
	if child == nil || child.DisplayObject() == nil {
		return nil
	}
	childSprite := child.DisplayObject()
	parentSprite := comp.DisplayObject()
	return func(x, y float64) bool {
		global := parentSprite.LocalToGlobal(laya.Point{X: x, Y: y})
		return childSprite.HitTest(global) != nil
	}
}

func maskHitTester(comp *core.GComponent, mask *core.GObject) func(x, y float64) bool {
	if mask == nil || mask.DisplayObject() == nil {
		return nil
	}
	maskSprite := mask.DisplayObject()
	parentSprite := comp.DisplayObject()
	width := mask.Width()
	height := mask.Height()
	if width <= 0 || height <= 0 {
		return nil
	}
	return func(x, y float64) bool {
		global := parentSprite.LocalToGlobal(laya.Point{X: x, Y: y})
		local := maskSprite.GlobalToLocal(global)
		if local.X < 0 || local.Y < 0 || local.X > width || local.Y > height {
			return false
		}
		return true
	}
}

func combineMaskTester(base, mask func(x, y float64) bool, reversed bool) func(x, y float64) bool {
	if mask == nil {
		return base
	}
	if base == nil {
		if reversed {
			return func(x, y float64) bool { return !mask(x, y) }
		}
		return mask
	}
	if reversed {
		return func(x, y float64) bool {
			if mask(x, y) {
				return false
			}
			return base(x, y)
		}
	}
	return func(x, y float64) bool {
		if !mask(x, y) {
			return false
		}
		return base(x, y)
	}
}

func componentAxisScale(current, source float64) float64 {
	if source <= 0 {
		return 1
	}
	if current <= 0 {
		current = source
	}
	scale := current / source
	if math.Abs(scale) < 1e-9 {
		return 1
	}
	return scale
}
