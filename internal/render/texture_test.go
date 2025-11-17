package render

import (
	"testing"

	"github.com/chslink/fairygui"
	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// 九宫格渲染测试
// ============================================================================

func TestDrawNineSliceTexture_Normal(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	nineSlice := &fairygui.NineSlice{
		Left:   10,
		Right:  10,
		Top:    10,
		Bottom: 10,
	}

	options := fairygui.DrawOptions{
		X:         0,
		Y:         0,
		Width:     200,
		Height:    200,
		NineSlice: nineSlice,
	}

	baseOpts := &ebiten.DrawImageOptions{}

	// 记录初始绘制调用次数
	initialCalls := renderer.DrawCalls()

	renderer.DrawNineSliceTexture(screen, texture, options, baseOpts)

	// 九宫格应该产生 9 次绘制调用
	expectedCalls := 9
	actualCalls := renderer.DrawCalls() - initialCalls

	if actualCalls != expectedCalls {
		t.Errorf("Expected %d draw calls for 9-slice, got %d", expectedCalls, actualCalls)
	}
}

func TestDrawNineSliceTexture_SmallTarget(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	nineSlice := &fairygui.NineSlice{
		Left:   10,
		Right:  10,
		Top:    10,
		Bottom: 10,
	}

	// 目标尺寸小于九宫格边距
	options := fairygui.DrawOptions{
		X:         0,
		Y:         0,
		Width:     15,
		Height:    15,
		NineSlice: nineSlice,
	}

	baseOpts := &ebiten.DrawImageOptions{}
	initialCalls := renderer.DrawCalls()

	renderer.DrawNineSliceTexture(screen, texture, options, baseOpts)

	// 应该退化为单次缩放绘制
	actualCalls := renderer.DrawCalls() - initialCalls
	if actualCalls != 1 {
		t.Errorf("Expected 1 draw call for small target, got %d", actualCalls)
	}
}

func TestDrawNineSliceTexture_NoNineSlice(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	// 没有九宫格参数
	options := fairygui.DrawOptions{
		X:      0,
		Y:      0,
		Width:  200,
		Height: 200,
	}

	baseOpts := &ebiten.DrawImageOptions{}
	initialCalls := renderer.DrawCalls()

	renderer.DrawNineSliceTexture(screen, texture, options, baseOpts)

	// 应该是单次绘制
	actualCalls := renderer.DrawCalls() - initialCalls
	if actualCalls != 1 {
		t.Errorf("Expected 1 draw call without 9-slice, got %d", actualCalls)
	}
}

func TestDrawNineSliceTexture_DefaultSize(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	nineSlice := &fairygui.NineSlice{
		Left:   10,
		Right:  10,
		Top:    10,
		Bottom: 10,
	}

	// 不指定宽高，应该使用纹理原始尺寸
	options := fairygui.DrawOptions{
		X:         0,
		Y:         0,
		NineSlice: nineSlice,
	}

	baseOpts := &ebiten.DrawImageOptions{}

	// 应该不会 panic
	renderer.DrawNineSliceTexture(screen, texture, options, baseOpts)
}

// ============================================================================
// 平铺渲染测试
// ============================================================================

func TestDrawTilingTexture_Normal(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(50, 50)

	options := fairygui.DrawOptions{
		X:      0,
		Y:      0,
		Width:  200,
		Height: 200,
		Tiling: true,
	}

	baseOpts := &ebiten.DrawImageOptions{}
	initialCalls := renderer.DrawCalls()

	renderer.DrawTilingTexture(screen, texture, options, baseOpts)

	// 计算预期的平铺次数:
	// x 方向: 0, 50, 100, 150 (4个，因为 200 >= targetWidth)
	// y 方向: 0, 50, 100, 150 (4个，因为 200 >= targetHeight)
	// 总共: 4 * 4 = 16
	expectedCalls := 16
	actualCalls := renderer.DrawCalls() - initialCalls

	if actualCalls != expectedCalls {
		t.Errorf("Expected %d draw calls for tiling, got %d", expectedCalls, actualCalls)
	}
}

func TestDrawTilingTexture_PartialTiles(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(50, 50)

	// 目标尺寸不是纹理尺寸的整数倍
	options := fairygui.DrawOptions{
		X:      0,
		Y:      0,
		Width:  125,
		Height: 125,
		Tiling: true,
	}

	baseOpts := &ebiten.DrawImageOptions{}

	// 应该不会 panic，即使需要裁剪
	renderer.DrawTilingTexture(screen, texture, options, baseOpts)
}

func TestDrawTilingTexture_DefaultSize(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(50, 50)

	// 不指定宽高，应该使用纹理原始尺寸
	options := fairygui.DrawOptions{
		X:      0,
		Y:      0,
		Tiling: true,
	}

	baseOpts := &ebiten.DrawImageOptions{}
	initialCalls := renderer.DrawCalls()

	renderer.DrawTilingTexture(screen, texture, options, baseOpts)

	// 单个 tile
	actualCalls := renderer.DrawCalls() - initialCalls
	if actualCalls != 1 {
		t.Errorf("Expected 1 draw call for default size tiling, got %d", actualCalls)
	}
}

func TestDrawTilingTexture_SingleTile(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	// 目标尺寸小于纹理尺寸
	options := fairygui.DrawOptions{
		X:      0,
		Y:      0,
		Width:  50,
		Height: 50,
		Tiling: true,
	}

	baseOpts := &ebiten.DrawImageOptions{}
	initialCalls := renderer.DrawCalls()

	renderer.DrawTilingTexture(screen, texture, options, baseOpts)

	// 只需要一个 tile
	actualCalls := renderer.DrawCalls() - initialCalls
	if actualCalls != 1 {
		t.Errorf("Expected 1 draw call for single tile, got %d", actualCalls)
	}
}

// ============================================================================
// 缩放纹理测试
// ============================================================================

func TestDrawScaledTexture_WithSize(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	options := fairygui.DrawOptions{
		X:      10,
		Y:      20,
		Width:  200,
		Height: 150,
	}

	baseOpts := &ebiten.DrawImageOptions{}
	initialCalls := renderer.DrawCalls()

	renderer.DrawScaledTexture(screen, texture, options, baseOpts)

	actualCalls := renderer.DrawCalls() - initialCalls
	if actualCalls != 1 {
		t.Errorf("Expected 1 draw call, got %d", actualCalls)
	}
}

func TestDrawScaledTexture_DefaultSize(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	options := fairygui.DrawOptions{
		X: 10,
		Y: 20,
	}

	baseOpts := &ebiten.DrawImageOptions{}

	// 应该不会 panic
	renderer.DrawScaledTexture(screen, texture, options, baseOpts)
}

func TestDrawScaledTexture_OnlyWidth(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	options := fairygui.DrawOptions{
		X:     10,
		Y:     20,
		Width: 200,
	}

	baseOpts := &ebiten.DrawImageOptions{}

	// 应该不会 panic
	renderer.DrawScaledTexture(screen, texture, options, baseOpts)
}

func TestDrawScaledTexture_OnlyHeight(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	options := fairygui.DrawOptions{
		X:      10,
		Y:      20,
		Height: 200,
	}

	baseOpts := &ebiten.DrawImageOptions{}

	// 应该不会 panic
	renderer.DrawScaledTexture(screen, texture, options, baseOpts)
}
