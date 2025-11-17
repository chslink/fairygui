package render

import (
	"image/color"
	"testing"

	"github.com/chslink/fairygui"
	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// 形状类型测试
// ============================================================================

func TestRectShape(t *testing.T) {
	rect := NewRectShape(10, 20, 100, 50)

	if rect.Type() != fairygui.ShapeTypeRect {
		t.Errorf("Expected ShapeTypeRect, got %v", rect.Type())
	}

	if rect.X != 10 || rect.Y != 20 {
		t.Errorf("Expected position (10, 20), got (%f, %f)", rect.X, rect.Y)
	}

	if rect.Width != 100 || rect.Height != 50 {
		t.Errorf("Expected size (100, 50), got (%f, %f)", rect.Width, rect.Height)
	}
}

func TestCircleShape(t *testing.T) {
	circle := NewCircleShape(50, 50, 25)

	if circle.Type() != fairygui.ShapeTypeCircle {
		t.Errorf("Expected ShapeTypeCircle, got %v", circle.Type())
	}

	if circle.X != 50 || circle.Y != 50 {
		t.Errorf("Expected center (50, 50), got (%f, %f)", circle.X, circle.Y)
	}

	if circle.Radius != 25 {
		t.Errorf("Expected radius 25, got %f", circle.Radius)
	}
}

func TestEllipseShape(t *testing.T) {
	ellipse := NewEllipseShape(100, 100, 50, 30)

	if ellipse.Type() != fairygui.ShapeTypeEllipse {
		t.Errorf("Expected ShapeTypeEllipse, got %v", ellipse.Type())
	}

	if ellipse.X != 100 || ellipse.Y != 100 {
		t.Errorf("Expected center (100, 100), got (%f, %f)", ellipse.X, ellipse.Y)
	}

	if ellipse.RadiusX != 50 || ellipse.RadiusY != 30 {
		t.Errorf("Expected radii (50, 30), got (%f, %f)", ellipse.RadiusX, ellipse.RadiusY)
	}
}

func TestPolygonShape(t *testing.T) {
	points := []float64{0, 0, 100, 0, 100, 100, 0, 100}
	polygon := NewPolygonShape(points)

	if polygon.Type() != fairygui.ShapeTypePolygon {
		t.Errorf("Expected ShapeTypePolygon, got %v", polygon.Type())
	}

	if len(polygon.Points) != 8 {
		t.Errorf("Expected 8 points, got %d", len(polygon.Points))
	}
}

// ============================================================================
// 形状绘制测试
// ============================================================================

func TestDrawShapeRect(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	rect := NewRectShape(10, 20, 100, 50)

	options := fairygui.DrawOptions{
		X: 0,
		Y: 0,
	}

	fillColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	strokeColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	initialCalls := renderer.DrawCalls()

	// 绘制带填充和描边的矩形
	renderer.DrawShapeRect(screen, rect, options, fillColor, strokeColor, 2.0)

	// 应该产生 2 次绘制调用（填充 + 描边）
	actualCalls := renderer.DrawCalls() - initialCalls
	if actualCalls != 2 {
		t.Errorf("Expected 2 draw calls (fill + stroke), got %d", actualCalls)
	}
}

func TestDrawShapeRect_FillOnly(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	rect := NewRectShape(10, 20, 100, 50)

	options := fairygui.DrawOptions{}
	fillColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	initialCalls := renderer.DrawCalls()

	// 只绘制填充
	renderer.DrawShapeRect(screen, rect, options, fillColor, nil, 0)

	// 应该产生 1 次绘制调用（只填充）
	actualCalls := renderer.DrawCalls() - initialCalls
	if actualCalls != 1 {
		t.Errorf("Expected 1 draw call (fill only), got %d", actualCalls)
	}
}

func TestDrawShapeCircle(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	circle := NewCircleShape(50, 50, 25)

	options := fairygui.DrawOptions{
		X: 100,
		Y: 100,
	}

	fillColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	initialCalls := renderer.DrawCalls()

	renderer.DrawShapeCircle(screen, circle, options, fillColor, nil, 0)

	actualCalls := renderer.DrawCalls() - initialCalls
	if actualCalls != 1 {
		t.Errorf("Expected 1 draw call, got %d", actualCalls)
	}
}

func TestDrawShapeEllipse(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)
	ellipse := NewEllipseShape(100, 100, 50, 30)

	options := fairygui.DrawOptions{}
	fillColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	// 应该不会 panic
	renderer.DrawShapeEllipse(screen, ellipse, options, fillColor, nil, 0)
}

func TestDrawShapePolygon(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)

	// 创建三角形
	points := []float64{0, 0, 100, 0, 50, 100}
	polygon := NewPolygonShape(points)

	options := fairygui.DrawOptions{
		X: 50,
		Y: 50,
	}

	fillColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}

	initialCalls := renderer.DrawCalls()

	renderer.DrawShapePolygon(screen, polygon, options, fillColor, nil, 0)

	actualCalls := renderer.DrawCalls() - initialCalls
	if actualCalls != 1 {
		t.Errorf("Expected 1 draw call, got %d", actualCalls)
	}
}

func TestDrawShapePolygon_InvalidPoints(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)

	// 点数不足（少于 3 个点）
	points := []float64{0, 0, 100, 0}
	polygon := NewPolygonShape(points)

	options := fairygui.DrawOptions{}
	fillColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}

	initialCalls := renderer.DrawCalls()

	// 应该不会 panic，但也不会绘制
	renderer.DrawShapePolygon(screen, polygon, options, fillColor, nil, 0)

	actualCalls := renderer.DrawCalls() - initialCalls
	if actualCalls != 0 {
		t.Errorf("Expected 0 draw calls for invalid polygon, got %d", actualCalls)
	}
}

func TestDrawShapeAuto(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)

	options := fairygui.DrawOptions{}
	fillColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	testCases := []struct {
		name  string
		shape fairygui.Shape
	}{
		{"Rect", NewRectShape(0, 0, 100, 50)},
		{"Circle", NewCircleShape(50, 50, 25)},
		{"Ellipse", NewEllipseShape(100, 100, 50, 30)},
		{"Polygon", NewPolygonShape([]float64{0, 0, 100, 0, 50, 100})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 应该不会 panic
			renderer.DrawShapeAuto(screen, tc.shape, options, fillColor, nil, 0)
		})
	}
}

func TestRendererDrawShape(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)

	rect := NewRectShape(10, 20, 100, 50)
	color := uint32(0xFF0000FF) // 红色

	options := fairygui.DrawOptions{
		X:     50,
		Y:     50,
		Color: &color,
	}

	// 应该不会 panic
	renderer.DrawShape(screen, rect, options)
}

func TestRendererDrawShape_NilShape(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)

	options := fairygui.DrawOptions{}

	// nil shape 应该不会 panic
	renderer.DrawShape(screen, nil, options)
}

func TestRendererDrawShape_DefaultColor(t *testing.T) {
	renderer := NewEbitenRenderer()
	screen := ebiten.NewImage(800, 600)

	rect := NewRectShape(10, 20, 100, 50)
	options := fairygui.DrawOptions{}

	// 没有指定颜色，应该使用默认黑色
	renderer.DrawShape(screen, rect, options)
}

// ============================================================================
// 颜色转换测试
// ============================================================================

// 注意：colorToColorScale 函数已被移除
// 在 DrawShapePolygon 中，我们现在直接在顶点数据中设置颜色
// 而不是使用 ColorScale，因此不需要单独的颜色转换测试
