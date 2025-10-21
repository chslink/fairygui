package render

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestConfigureComponentHitAreaPixel(t *testing.T) {
	comp := core.NewGComponent()
	comp.SetSize(4, 4)
	comp.SetSourceSize(4, 4)
	comp.SetHitTest(core.HitTest{Mode: core.HitTestModePixel})

	pixel := &assets.PixelHitTestData{
		Width:  4,
		Height: 4,
		Scale:  1,
		Data:   []byte{0xFF, 0x00},
	}

	ConfigureComponentHitArea(comp, comp.HitTest(), pixel)

	sprite := comp.DisplayObject()
	if sprite == nil {
		t.Fatalf("component display sprite missing")
	}

	if sprite.HitTest(laya.Point{X: 1, Y: 1}) == nil {
		t.Fatalf("expected hit inside pixel mask")
	}
	if sprite.HitTest(laya.Point{X: 1, Y: 3}) != nil {
		t.Fatalf("expected miss outside pixel mask")
	}
}

func TestConfigureComponentHitAreaMask(t *testing.T) {
	comp := core.NewGComponent()
	comp.SetSize(40, 40)

	mask := core.NewGObject()
	mask.SetSize(20, 20)
	comp.AddChild(mask)
	comp.SetMask(mask, false)

	if mask.DisplayObject().HitTest(laya.Point{X: 10, Y: 10}) == nil {
		t.Fatalf("expected mask display to report hit at 10,10")
	}

	maskFunc := maskHitTester(comp, mask)
	if maskFunc == nil {
		t.Fatalf("mask tester missing")
	}
	if !maskFunc(10, 10) {
		t.Fatalf("expected mask tester to report hit at 10,10")
	}

	ConfigureComponentHitArea(comp, comp.HitTest(), nil)

	sprite := comp.DisplayObject()
	if sprite.HitTest(laya.Point{X: 10, Y: 10}) == nil {
		t.Fatalf("expected hit inside mask")
	}
	if sprite.HitTest(laya.Point{X: 30, Y: 30}) != nil {
		t.Fatalf("expected miss outside mask")
	}

	comp.SetMask(mask, true)
	ConfigureComponentHitArea(comp, comp.HitTest(), nil)

	if sprite.HitTest(laya.Point{X: 10, Y: 10}) != nil {
		t.Fatalf("expected miss inside reversed mask")
	}
	if sprite.HitTest(laya.Point{X: 30, Y: 30}) == nil {
		t.Fatalf("expected hit outside reversed mask")
	}
}

func TestConfigureComponentHitAreaChild(t *testing.T) {
	comp := core.NewGComponent()
	comp.SetSize(40, 40)

	child := core.NewGObject()
	child.SetSize(12, 12)
	child.SetPosition(4, 4)
	comp.AddChild(child)

	hit := core.HitTest{Mode: core.HitTestModeChild, ChildIndex: 0}
	comp.SetHitTest(hit)
	ConfigureComponentHitArea(comp, hit, nil)

	sprite := comp.DisplayObject()
	if sprite.HitTest(laya.Point{X: 6, Y: 6}) == nil {
		t.Fatalf("expected hit inside child hit area")
	}
	if sprite.HitTest(laya.Point{X: 30, Y: 30}) != nil {
		t.Fatalf("expected miss outside child hit area")
	}
}

func TestConfigureComponentHitAreaPixelScalingOffset(t *testing.T) {
	comp := core.NewGComponent()
	comp.SetSourceSize(10, 10)
	comp.SetSize(20, 20) // scale factor 2
	comp.SetHitTest(core.HitTest{Mode: core.HitTestModePixel, OffsetX: 2, OffsetY: 1})

	pixel := &assets.PixelHitTestData{
		Width:  10,
		Height: 10,
		Scale:  1,
		Data:   []byte{0xFF, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	ConfigureComponentHitArea(comp, comp.HitTest(), pixel)

	sprite := comp.DisplayObject()
	if sprite == nil {
		t.Fatalf("component sprite missing")
	}

	// Account for offset: local space offset 2, so world coordinate 2*scale (4) should be first hit.
	if sprite.HitTest(laya.Point{X: 4, Y: 2}) == nil {
		t.Fatalf("expected hit after applying offset and scaling")
	}
	// Outside the mask band.
	if sprite.HitTest(laya.Point{X: 1, Y: 1}) != nil {
		t.Fatalf("expected miss before offset region")
	}
}
