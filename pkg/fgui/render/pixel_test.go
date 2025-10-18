package render

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
)

func TestApplyPixelHitTest(t *testing.T) {
	sprite := laya.NewSprite()
	sprite.SetSize(10, 10)

	mask := make([]byte, 13)
	for i := 0; i < 50; i++ {
		mask[i>>3] |= 1 << (uint(i) & 7)
	}
	data := &assets.PixelHitTestData{
		Width:  10,
		Height: 10,
		Scale:  1,
		Data:   mask,
	}
	ApplyPixelHitTest(sprite, data)

	if sprite.HitTest(laya.Point{X: 1, Y: 1}) != sprite {
		t.Fatalf("expected hit inside pixel data")
	}
	if sprite.HitTest(laya.Point{X: 9, Y: 9}) != nil {
		t.Fatalf("expected miss outside pixel data")
	}
}
