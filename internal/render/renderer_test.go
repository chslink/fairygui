package render

import (
	"testing"

	"github.com/chslink/fairygui"
	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// 基础渲染器测试
// ============================================================================

func TestEbitenRenderer_Creation(t *testing.T) {
	renderer := NewEbitenRenderer()
	if renderer == nil {
		t.Fatal("NewEbitenRenderer() returned nil")
	}

	// 验证实现了 Renderer 接口
	var _ fairygui.Renderer = renderer
}

func TestEbitenRenderer_Draw(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	root := fairygui.NewObject()

	// 应该不会 panic
	renderer.Draw(screen, root)

	// 验证绘制调用次数
	if renderer.DrawCalls() < 0 {
		t.Error("DrawCalls should be non-negative")
	}
}

func TestEbitenRenderer_DrawNil(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)

	// nil root 应该不会 panic
	renderer.Draw(screen, nil)
}

func TestEbitenRenderer_DrawTexture(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	options := fairygui.DrawOptions{
		X:      10,
		Y:      20,
		Width:  100,
		Height: 100,
		Alpha:  1.0,
	}

	// 应该不会 panic
	renderer.DrawTexture(screen, texture, options)

	if renderer.DrawCalls() != 1 {
		t.Errorf("Expected 1 draw call, got %d", renderer.DrawCalls())
	}
}

func TestEbitenRenderer_DrawTextureWithTransform(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	options := fairygui.DrawOptions{
		X:        10,
		Y:        20,
		ScaleX:   2.0,
		ScaleY:   2.0,
		Rotation: 0.5,
		Alpha:    0.5,
	}

	// 应该不会 panic
	renderer.DrawTexture(screen, texture, options)

	if renderer.DrawCalls() != 1 {
		t.Errorf("Expected 1 draw call, got %d", renderer.DrawCalls())
	}
}

func TestEbitenRenderer_DrawTextureWithColor(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	color := uint32(0xFF0000FF) // 红色，完全不透明
	options := fairygui.DrawOptions{
		X:     10,
		Y:     20,
		Color: &color,
	}

	// 应该不会 panic
	renderer.DrawTexture(screen, texture, options)
}

func TestEbitenRenderer_DrawTextureWithBlendMode(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	testCases := []struct {
		name      string
		blendMode fairygui.BlendMode
	}{
		{"Normal", fairygui.BlendModeNormal},
		{"Add", fairygui.BlendModeAdd},
		{"Multiply", fairygui.BlendModeMultiply},
		{"Screen", fairygui.BlendModeScreen},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := fairygui.DrawOptions{
				BlendMode: tc.blendMode,
			}
			renderer.DrawTexture(screen, texture, options)
		})
	}
}

func TestEbitenRenderer_DrawTextureWithNineSlice(t *testing.T) {
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
		Width:     200,
		Height:    200,
		NineSlice: nineSlice,
	}

	// 应该不会 panic
	renderer.DrawTexture(screen, texture, options)
}

func TestEbitenRenderer_DrawTextureWithTiling(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(50, 50)

	options := fairygui.DrawOptions{
		Width:  200,
		Height: 200,
		Tiling: true,
	}

	// 应该不会 panic
	renderer.DrawTexture(screen, texture, options)
}

func TestEbitenRenderer_DrawText(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)

	style := fairygui.TextStyle{
		Size:  16,
		Color: 0x000000FF,
		Bold:  false,
	}

	// 应该不会 panic（即使是占位实现）
	renderer.DrawText(screen, "Hello, World!", 10, 20, style)
}

func TestEbitenRenderer_DrawShape(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)

	// 创建一个简单的矩形形状
	rect := &mockShape{shapeType: fairygui.ShapeTypeRect}

	options := fairygui.DrawOptions{
		X:      10,
		Y:      20,
		Width:  100,
		Height: 50,
	}

	// 应该不会 panic（即使是占位实现）
	renderer.DrawShape(screen, rect, options)
}

func TestEbitenRenderer_Statistics(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	texture := ebiten.NewImage(100, 100)

	// 初始状态
	if renderer.DrawCalls() != 0 {
		t.Errorf("Expected 0 initial draw calls, got %d", renderer.DrawCalls())
	}

	// 绘制一次
	options := fairygui.DrawOptions{}
	renderer.DrawTexture(screen, texture, options)

	if renderer.DrawCalls() != 1 {
		t.Errorf("Expected 1 draw call, got %d", renderer.DrawCalls())
	}
}

// ============================================================================
// Mock 类型
// ============================================================================

type mockShape struct {
	shapeType fairygui.ShapeType
}

func (m *mockShape) Type() fairygui.ShapeType {
	return m.shapeType
}
