package render

import (
	"image/color"
	"math"

	"github.com/chslink/fairygui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	// emptyImage 是用于 DrawTriangles 的 1x1 白色图像
	emptyImage *ebiten.Image
)

func init() {
	emptyImage = ebiten.NewImage(1, 1)
	emptyImage.Fill(color.White)
}

// ============================================================================
// 形状类型实现
// ============================================================================

// RectShape 矩形形状。
type RectShape struct {
	X, Y, Width, Height float64
}

func (r *RectShape) Type() fairygui.ShapeType {
	return fairygui.ShapeTypeRect
}

// NewRectShape 创建矩形形状。
func NewRectShape(x, y, width, height float64) *RectShape {
	return &RectShape{X: x, Y: y, Width: width, Height: height}
}

// CircleShape 圆形形状。
type CircleShape struct {
	X, Y, Radius float64
}

func (c *CircleShape) Type() fairygui.ShapeType {
	return fairygui.ShapeTypeCircle
}

// NewCircleShape 创建圆形形状。
func NewCircleShape(x, y, radius float64) *CircleShape {
	return &CircleShape{X: x, Y: y, Radius: radius}
}

// EllipseShape 椭圆形状。
type EllipseShape struct {
	X, Y, RadiusX, RadiusY float64
}

func (e *EllipseShape) Type() fairygui.ShapeType {
	return fairygui.ShapeTypeEllipse
}

// NewEllipseShape 创建椭圆形状。
func NewEllipseShape(x, y, radiusX, radiusY float64) *EllipseShape {
	return &EllipseShape{X: x, Y: y, RadiusX: radiusX, RadiusY: radiusY}
}

// PolygonShape 多边形形状。
type PolygonShape struct {
	Points []float64 // [x1, y1, x2, y2, ...]
}

func (p *PolygonShape) Type() fairygui.ShapeType {
	return fairygui.ShapeTypePolygon
}

// NewPolygonShape 创建多边形形状。
func NewPolygonShape(points []float64) *PolygonShape {
	return &PolygonShape{Points: points}
}

// ============================================================================
// 形状绘制实现
// ============================================================================

// DrawShapeRect 绘制矩形。
func (r *EbitenRenderer) DrawShapeRect(
	screen *ebiten.Image,
	rect *RectShape,
	options fairygui.DrawOptions,
	fillColor color.Color,
	strokeColor color.Color,
	strokeWidth float64,
) {
	x := float32(rect.X + options.X)
	y := float32(rect.Y + options.Y)
	width := float32(rect.Width)
	height := float32(rect.Height)

	// 绘制填充
	if fillColor != nil {
		vector.DrawFilledRect(screen, x, y, width, height, fillColor, false)
		r.drawCalls++
	}

	// 绘制描边
	if strokeColor != nil && strokeWidth > 0 {
		vector.StrokeRect(screen, x, y, width, height, float32(strokeWidth), strokeColor, false)
		r.drawCalls++
	}
}

// DrawShapeCircle 绘制圆形。
func (r *EbitenRenderer) DrawShapeCircle(
	screen *ebiten.Image,
	circle *CircleShape,
	options fairygui.DrawOptions,
	fillColor color.Color,
	strokeColor color.Color,
	strokeWidth float64,
) {
	x := float32(circle.X + options.X)
	y := float32(circle.Y + options.Y)
	radius := float32(circle.Radius)

	// 绘制填充
	if fillColor != nil {
		vector.DrawFilledCircle(screen, x, y, radius, fillColor, false)
		r.drawCalls++
	}

	// 绘制描边
	if strokeColor != nil && strokeWidth > 0 {
		vector.StrokeCircle(screen, x, y, radius, float32(strokeWidth), strokeColor, false)
		r.drawCalls++
	}
}

// DrawShapeEllipse 绘制椭圆。
func (r *EbitenRenderer) DrawShapeEllipse(
	screen *ebiten.Image,
	ellipse *EllipseShape,
	options fairygui.DrawOptions,
	fillColor color.Color,
	strokeColor color.Color,
	strokeWidth float64,
) {
	// Ebiten 没有直接的椭圆绘制 API，使用多边形近似
	segments := 64
	points := make([]float64, 0, segments*2)

	cx := ellipse.X + options.X
	cy := ellipse.Y + options.Y
	rx := ellipse.RadiusX
	ry := ellipse.RadiusY

	for i := 0; i < segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		x := cx + rx*math.Cos(angle)
		y := cy + ry*math.Sin(angle)
		points = append(points, x, y)
	}

	// 绘制多边形
	poly := &PolygonShape{Points: points}
	r.DrawShapePolygon(screen, poly, options, fillColor, strokeColor, strokeWidth)
}

// DrawShapePolygon 绘制多边形。
func (r *EbitenRenderer) DrawShapePolygon(
	screen *ebiten.Image,
	polygon *PolygonShape,
	options fairygui.DrawOptions,
	fillColor color.Color,
	strokeColor color.Color,
	strokeWidth float64,
) {
	if len(polygon.Points) < 6 {
		// 至少需要 3 个点（6 个坐标）
		return
	}

	// 应用偏移
	offsetX := float32(options.X)
	offsetY := float32(options.Y)

	// 转换为 float32 数组并应用偏移
	path := make([]float32, len(polygon.Points))
	for i := 0; i < len(polygon.Points); i += 2 {
		path[i] = float32(polygon.Points[i]) + offsetX
		path[i+1] = float32(polygon.Points[i+1]) + offsetY
	}

	// 绘制填充
	if fillColor != nil {
		// 创建路径
		var vp vector.Path
		vp.MoveTo(path[0], path[1])
		for i := 2; i < len(path); i += 2 {
			vp.LineTo(path[i], path[i+1])
		}
		vp.Close()

		// 填充路径
		vertices, indices := vp.AppendVerticesAndIndicesForFilling(nil, nil)

		// 应用颜色到顶点
		fr, fg, fb, fa := fillColor.RGBA()
		for i := range vertices {
			vertices[i].ColorR = float32(fr) / 0xFFFF
			vertices[i].ColorG = float32(fg) / 0xFFFF
			vertices[i].ColorB = float32(fb) / 0xFFFF
			vertices[i].ColorA = float32(fa) / 0xFFFF
		}

		// 绘制三角形
		screen.DrawTriangles(vertices, indices, emptyImage, &ebiten.DrawTrianglesOptions{})
		r.drawCalls++
	}

	// 绘制描边
	if strokeColor != nil && strokeWidth > 0 {
		var sp vector.Path
		sp.MoveTo(path[0], path[1])
		for i := 2; i < len(path); i += 2 {
			sp.LineTo(path[i], path[i+1])
		}
		sp.Close()

		opts := &vector.StrokeOptions{
			Width:      float32(strokeWidth),
			LineJoin:   vector.LineJoinMiter,
			LineCap:    vector.LineCapButt,
			MiterLimit: 10,
		}

		vertices, indices := sp.AppendVerticesAndIndicesForStroke(nil, nil, opts)

		// 应用颜色到顶点
		sr, sg, sb, sa := strokeColor.RGBA()
		for i := range vertices {
			vertices[i].ColorR = float32(sr) / 0xFFFF
			vertices[i].ColorG = float32(sg) / 0xFFFF
			vertices[i].ColorB = float32(sb) / 0xFFFF
			vertices[i].ColorA = float32(sa) / 0xFFFF
		}

		// 绘制三角形
		screen.DrawTriangles(vertices, indices, emptyImage, &ebiten.DrawTrianglesOptions{})
		r.drawCalls++
	}
}

// DrawShapeAuto 根据形状类型自动选择绘制方法。
func (r *EbitenRenderer) DrawShapeAuto(
	screen *ebiten.Image,
	shape fairygui.Shape,
	options fairygui.DrawOptions,
	fillColor color.Color,
	strokeColor color.Color,
	strokeWidth float64,
) {
	switch shape.Type() {
	case fairygui.ShapeTypeRect:
		if rect, ok := shape.(*RectShape); ok {
			r.DrawShapeRect(screen, rect, options, fillColor, strokeColor, strokeWidth)
		}
	case fairygui.ShapeTypeCircle:
		if circle, ok := shape.(*CircleShape); ok {
			r.DrawShapeCircle(screen, circle, options, fillColor, strokeColor, strokeWidth)
		}
	case fairygui.ShapeTypeEllipse:
		if ellipse, ok := shape.(*EllipseShape); ok {
			r.DrawShapeEllipse(screen, ellipse, options, fillColor, strokeColor, strokeWidth)
		}
	case fairygui.ShapeTypePolygon:
		if polygon, ok := shape.(*PolygonShape); ok {
			r.DrawShapePolygon(screen, polygon, options, fillColor, strokeColor, strokeWidth)
		}
	}
}
