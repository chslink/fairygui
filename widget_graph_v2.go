package fairygui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// GraphType - 图形类型
// ============================================================================

type GraphType int

const (
	GraphTypeRectangle GraphType = iota
	GraphTypeRectangleCorner
	GraphTypeCircle
	GraphTypeEllipse
	GraphTypePolygon
	GraphTypeRegularPolygon
)

// ============================================================================
// Graph - 图形控件 V2
// ============================================================================

type Graph struct {
	*Object

	// 图形类型
	graphType GraphType

	// 线条样式
	lineSize   float64
	lineColor  color.RGBA
	strokeSize float64
	strokeColor color.RGBA

	// 填充
	fillColor color.RGBA
	fillShape bool

	// 矩形圆角（用于 GraphTypeRectangleCorner）
	cornerRadius float64

	// 圆形/椭圆属性
	radiusX float64
	radiusY float64

	// 多边形顶点数
	sides int

	// 旋转角度（度）
	rotation float64

	// 内切/外接（用于正多边形）
	polygonCut bool
}

// NewGraph 创建新的图形控件
func NewGraph() *Graph {
	g := &Graph{
		Object:       NewObject(),
		graphType:    GraphTypeRectangle,
		lineSize:     1,
		lineColor:    color.RGBA{R: 0, G: 0, B: 0, A: 255},
		strokeSize:   0,
		strokeColor:  color.RGBA{R: 0, G: 0, B: 0, A: 0},
		fillColor:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
		fillShape:    true,
		cornerRadius: 0,
		radiusX:      50,
		radiusY:      50,
		sides:        5,
		rotation:     0,
		polygonCut:   false,
	}

	// 默认不拦截事件
	g.SetTouchable(false)

	return g
}

// ============================================================================
// 图形类型设置
// ============================================================================

// SetGraphType 设置图形类型
func (g *Graph) SetGraphType(graphType GraphType) *Graph {
	if g.graphType == graphType {
		return g
	}
	g.graphType = graphType
	g.updateGraphics()
	return g
}

// GraphType 返回图形类型
func (g *Graph) GraphType() GraphType {
	return g.graphType
}

// SetLineStyle 设置线条样式
func (g *Graph) SetLineStyle(size float64, color color.RGBA) *Graph {
	if g.lineSize == size && g.lineColor == color {
		return g
	}
	g.lineSize = size
	g.lineColor = color
	g.updateGraphics()
	return g
}

// SetLineSize 设置线宽（1=1px, 2=2px, -1=无边框）
func (g *Graph) SetLineSize(size float64) *Graph {
	if g.lineSize == size {
		return g
	}
	g.lineSize = size
	g.updateGraphics()
	return g
}

// LineSize 返回线宽
func (g *Graph) LineSize() float64 {
	return g.lineSize
}

// SetLineColor 设置线条颜色
func (g *Graph) SetLineColor(color color.RGBA) *Graph {
	if g.lineColor == color {
		return g
	}
	g.lineColor = color
	g.updateGraphics()
	return g
}

// LineColor 返回线条颜色
func (g *Graph) LineColor() color.RGBA {
	return g.lineColor
}

// SetStroke 设置描边
func (g *Graph) SetStroke(size float64, color color.RGBA) {
	if g.strokeSize == size && g.strokeColor == color {
		return
	}
	g.strokeSize = size
	g.strokeColor = color
	g.updateGraphics()
}

// StrokeSize 返回描边大小
func (g *Graph) StrokeSize() float64 {
	return g.strokeSize
}

// StrokeColor 返回描边颜色
func (g *Graph) StrokeColor() color.RGBA {
	return g.strokeColor
}

// SetFillColor 设置填充颜色
func (g *Graph) SetFillColor(color color.RGBA, fillShape bool) {
	if g.fillColor == color && g.fillShape == fillShape {
		return
	}
	g.fillColor = color
	g.fillShape = fillShape
	g.updateGraphics()
}

// FillColor 返回填充颜色
func (g *Graph) FillColor() color.RGBA {
	return g.fillColor
}

// FillShape 返回是否填充
func (g *Graph) FillShape() bool {
	return g.fillShape
}

// ============================================================================
// 圆角矩形
// ============================================================================

// SetCornerRadius 设置圆角半径（仅用于 GraphTypeRectangleCorner）
func (g *Graph) SetCornerRadius(radius float64) {
	if g.cornerRadius == radius {
		return
	}
	g.cornerRadius = radius
	g.updateGraphics()
}

// CornerRadius 返回圆角半径
func (g *Graph) CornerRadius() float64 {
	return g.cornerRadius
}

// ============================================================================
// 圆形/椭圆
// ============================================================================

// SetRadius 设置半径（用于圆形）
func (g *Graph) SetRadius(radius float64) {
	if g.radiusX == radius && g.radiusY == radius {
		return
	}
	g.radiusX = radius
	g.radiusY = radius
	g.updateGraphics()
}

// SetRadiusXY 设置XY方向半径（用于椭圆）
func (g *Graph) SetRadiusXY(radiusX, radiusY float64) {
	if g.radiusX == radiusX && g.radiusY == radiusY {
		return
	}
	g.radiusX = radiusX
	g.radiusY = radiusY
	g.updateGraphics()
}

// RadiusX 返回X方向半径
func (g *Graph) RadiusX() float64 {
	return g.radiusX
}

// RadiusY 返回Y方向半径
func (g *Graph) RadiusY() float64 {
	return g.radiusY
}

// SetSides 设置多边形边数
func (g *Graph) SetSides(sides int) {
	if g.sides == sides {
		return
	}
	g.sides = sides
	if g.sides < 3 {
		g.sides = 3
	}
	g.updateGraphics()
}

// Sides 返回边数
func (g *Graph) Sides() int {
	return g.sides
}

// ============================================================================
// 旋转
// ============================================================================

// SetRotation 设置旋转角度（度）
func (g *Graph) SetRotation(rotation float64) {
	if g.rotation == rotation {
		return
	}
	g.rotation = rotation
	g.updateGraphics()
}

// Rotation 返回旋转角度（度）
func (g *Graph) Rotation() float64 {
	return g.rotation
}

// ============================================================================
// 多边形属性
// ============================================================================

// SetPolygonCut 设置内切/外接（仅用于 GraphTypeRegularPolygon）
func (g *Graph) SetPolygonCut(cut bool) {
	if g.polygonCut == cut {
		return
	}
	g.polygonCut = cut
	g.updateGraphics()
}

// PolygonCut 返回内切/外接
func (g *Graph) PolygonCut() bool {
	return g.polygonCut
}

// ============================================================================
// 便捷绘制方法
// ============================================================================

// DrawRect 绘制矩形
func (g *Graph) DrawRect(lineSize float64, lineColor color.RGBA, fillColor color.RGBA) {
	g.graphType = GraphTypeRectangle
	g.lineSize = lineSize
	g.lineColor = lineColor
	g.fillColor = fillColor
	g.fillShape = true
	g.updateGraphics()
}

// DrawRoundedRect 绘制圆角矩形
func (g *Graph) DrawRoundedRect(lineSize float64, lineColor color.RGBA, fillColor color.RGBA, cornerRadius float64) {
	g.graphType = GraphTypeRectangleCorner
	g.lineSize = lineSize
	g.lineColor = lineColor
	g.fillColor = fillColor
	g.fillShape = true
	g.cornerRadius = cornerRadius
	g.updateGraphics()
}

// DrawCircle 绘制圆形
func (g *Graph) DrawCircle(lineSize float64, lineColor color.RGBA, fillColor color.RGBA) {
	g.graphType = GraphTypeCircle
	g.lineSize = lineSize
	g.lineColor = lineColor
	g.fillColor = fillColor
	g.fillShape = true
	g.updateGraphics()
}

// DrawEllipse 绘制椭圆
func (g *Graph) DrawEllipse(lineSize float64, lineColor color.RGBA, fillColor color.RGBA) {
	g.graphType = GraphTypeEllipse
	g.lineSize = lineSize
	g.lineColor = lineColor
	g.fillColor = fillColor
	g.fillShape = true
	g.updateGraphics()
}

// DrawPolygon 绘制多边形
func (g *Graph) DrawPolygon(lineSize float64, lineColor color.RGBA, fillColor color.RGBA, sides int) {
	g.graphType = GraphTypePolygon
	g.lineSize = lineSize
	g.lineColor = lineColor
	g.fillColor = fillColor
	g.fillShape = true
	g.sides = sides
	g.updateGraphics()
}

// DrawRegularPolygon 绘制正多边形
func (g *Graph) DrawRegularPolygon(lineSize float64, lineColor color.RGBA, fillColor color.RGBA, sides int, cut bool) {
	g.graphType = GraphTypeRegularPolygon
	g.lineSize = lineSize
	g.lineColor = lineColor
	g.fillColor = fillColor
	g.fillShape = true
	g.sides = sides
	g.polygonCut = cut
	g.updateGraphics()
}

// Clear 清除图形
func (g *Graph) Clear() {
	g.fillShape = false
	g.lineSize = 0
	g.updateGraphics()
}

// ============================================================================
// 内部方法
// ============================================================================

// updateGraphics 更新图形
func (g *Graph) updateGraphics() {
	// 根据自动大小设置尺寸
	if g.graphType == GraphTypeCircle {
		radius := g.radiusX
		if g.radiusY > radius {
			radius = g.radiusY
		}
		size := radius * 2
		g.SetSize(size, size)
	} else {
		// 其他形状使用当前尺寸
		// 实际绘制在渲染层处理
	}
}

// Draw 绘制图形
func (g *Graph) Draw(screen *ebiten.Image) {
	// 先调用父类绘制
	g.Object.Draw(screen)

	// TODO: 实现实际绘制
	// 这些应该在渲染层根据 g.graphType 和属性绘制
}

// ============================================================================
// 几何计算辅助函数
// ============================================================================

// calculatePolygonVertex 计算多边形顶点
func calculatePolygonVertex(centerX, centerY, radius float64, sides int, rotation float64, index int) (x, y float64) {
	angle := (2*math.Pi * float64(index)) / float64(sides)
	angle += rotation * math.Pi / 180

	x = centerX + radius*math.Cos(angle)
	y = centerY + radius*math.Sin(angle)

	return
}

// calculateStarVertex 计算星形顶点（用于 cut=true）
func calculateStarVertex(centerX, centerY, outerRadius, innerRadius float64, sides int, rotation float64, index int) (x, y float64) {
	isOuter := index%2 == 0
	angle := (math.Pi * float64(index)) / float64(sides)
	angle += rotation * math.Pi / 180

	var radius float64
	if isOuter {
		radius = outerRadius
	} else {
		radius = innerRadius
	}

	x = centerX + radius*math.Cos(angle)
	y = centerY + radius*math.Sin(angle)

	return
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertGraph 类型断言
func AssertGraph(obj DisplayObject) (*Graph, bool) {
	g, ok := obj.(*Graph)
	return g, ok
}

// IsGraph 检查是否是 Graph
func IsGraph(obj DisplayObject) bool {
	_, ok := obj.(*Graph)
	return ok
}
