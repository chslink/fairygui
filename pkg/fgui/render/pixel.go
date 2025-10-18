package render

import (
	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// ApplyPixelHitTest wires the pixel hit test data into the sprite's hit tester.
func ApplyPixelHitTest(sprite *laya.Sprite, data *assets.PixelHitTestData) {
	if sprite == nil {
		return
	}
	if data == nil {
		sprite.SetHitTester(nil)
		return
	}
	sprite.SetHitTester(func(x, y float64) bool {
		return data.Contains(x, y)
	})
}
