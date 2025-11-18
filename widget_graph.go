package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Graph - 简化的图形绘制控件
// ============================================================================

// Graph 是简化的图形绘制控件，包装了 pkg/fgui/widgets.GGraph。
//
// 可以绘制矩形、椭圆、多边形等基本图形。
type Graph struct {
	graph *widgets.GGraph
}

// GraphType 图形类型。
type GraphType int

const (
	// GraphTypeEmpty 空图形。
	GraphTypeEmpty GraphType = GraphType(widgets.GraphTypeEmpty)
	// GraphTypeRect 矩形。
	GraphTypeRect GraphType = GraphType(widgets.GraphTypeRect)
	// GraphTypeEllipse 椭圆。
	GraphTypeEllipse GraphType = GraphType(widgets.GraphTypeEllipse)
	// GraphTypePolygon 多边形。
	GraphTypePolygon GraphType = GraphType(widgets.GraphTypePolygon)
	// GraphTypeRegularPolygon 正多边形。
	GraphTypeRegularPolygon GraphType = GraphType(widgets.GraphTypeRegularPolygon)
)

// NewGraph 创建一个新的图形控件。
//
// 示例：
//
//	graph := fairygui.NewGraph()
//	graph.DrawRect(1, "#000000", "#FF0000", nil)  // 绘制红色矩形
func NewGraph() *Graph {
	return &Graph{
		graph: widgets.NewGraph(),
	}
}

// Type 返回图形类型。
func (g *Graph) Type() GraphType {
	return GraphType(g.graph.Type())
}

// SetType 设置图形类型。
func (g *Graph) SetType(t GraphType) {
	g.graph.SetType(widgets.GraphType(t))
}

// LineSize 返回线条宽度。
func (g *Graph) LineSize() float64 {
	return g.graph.LineSize()
}

// SetLineSize 设置线条宽度。
func (g *Graph) SetLineSize(size float64) {
	g.graph.SetLineSize(size)
}

// LineColor 返回线条颜色（十六进制格式）。
func (g *Graph) LineColor() string {
	return g.graph.LineColor()
}

// SetLineColor 设置线条颜色（十六进制格式）。
//
// 示例：
//
//	graph.SetLineColor("#000000")
func (g *Graph) SetLineColor(color string) {
	g.graph.SetLineColor(color)
}

// FillColor 返回填充颜色（十六进制格式）。
func (g *Graph) FillColor() string {
	return g.graph.FillColor()
}

// SetFillColor 设置填充颜色（十六进制格式）。
//
// 示例：
//
//	graph.SetFillColor("#FF0000")
func (g *Graph) SetFillColor(color string) {
	g.graph.SetFillColor(color)
}

// Color 返回填充颜色（FillColor 的别名）。
func (g *Graph) Color() string {
	return g.graph.Color()
}

// SetColor 设置填充颜色（SetFillColor 的别名）。
func (g *Graph) SetColor(color string) {
	g.graph.SetColor(color)
}

// CornerRadius 返回圆角半径列表（最多4个值，对应左上、右上、右下、左下）。
func (g *Graph) CornerRadius() []float64 {
	return g.graph.CornerRadius()
}

// SetCornerRadius 设置圆角半径。
//
// 示例：
//
//	graph.SetCornerRadius([]float64{10, 10, 10, 10})  // 四个角都是10px
func (g *Graph) SetCornerRadius(values []float64) {
	g.graph.SetCornerRadius(values)
}

// PolygonPoints 返回多边形顶点坐标列表（x1, y1, x2, y2, ...）。
func (g *Graph) PolygonPoints() []float64 {
	return g.graph.PolygonPoints()
}

// SetPolygonPoints 设置多边形顶点坐标。
//
// 示例：
//
//	graph.SetPolygonPoints([]float64{0, 0, 100, 0, 50, 100})  // 三角形
func (g *Graph) SetPolygonPoints(points []float64) {
	g.graph.SetPolygonPoints(points)
}

// RegularPolygon 返回正多边形参数（边数、起始角度、距离倍数列表）。
func (g *Graph) RegularPolygon() (sides int, startAngle float64, distances []float64) {
	return g.graph.RegularPolygon()
}

// SetRegularPolygon 设置正多边形参数。
//
// 参数:
//   - sides: 边数
//   - startAngle: 起始角度（度数）
//   - distances: 每个顶点的距离倍数（可选）
//
// 示例：
//
//	graph.SetRegularPolygon(6, 0, nil)  // 六边形
func (g *Graph) SetRegularPolygon(sides int, startAngle float64, distances []float64) {
	g.graph.SetRegularPolygon(sides, startAngle, distances)
}

// DrawRect 绘制矩形。
//
// 参数:
//   - lineSize: 线条宽度
//   - lineColor: 线条颜色
//   - fillColor: 填充颜色
//   - cornerRadius: 圆角半径（可选）
//
// 示例：
//
//	graph.DrawRect(1, "#000000", "#FF0000", nil)             // 矩形
//	graph.DrawRect(1, "#000000", "#FF0000", []float64{10})  // 圆角矩形
func (g *Graph) DrawRect(lineSize float64, lineColor, fillColor string, cornerRadius []float64) {
	g.graph.DrawRect(lineSize, lineColor, fillColor, cornerRadius)
}

// DrawEllipse 绘制椭圆。
//
// 参数:
//   - lineSize: 线条宽度
//   - lineColor: 线条颜色
//   - fillColor: 填充颜色
//
// 示例：
//
//	graph.DrawEllipse(1, "#000000", "#00FF00")
func (g *Graph) DrawEllipse(lineSize float64, lineColor, fillColor string) {
	g.graph.DrawEllipse(lineSize, lineColor, fillColor)
}

// DrawPolygon 绘制多边形。
//
// 参数:
//   - lineSize: 线条宽度
//   - lineColor: 线条颜色
//   - fillColor: 填充颜色
//   - points: 顶点坐标列表（x1, y1, x2, y2, ...）
//
// 示例：
//
//	graph.DrawPolygon(1, "#000000", "#0000FF", []float64{0, 0, 100, 0, 50, 100})
func (g *Graph) DrawPolygon(lineSize float64, lineColor, fillColor string, points []float64) {
	g.graph.DrawPolygon(lineSize, lineColor, fillColor, points)
}

// DrawRegularPolygon 绘制正多边形。
//
// 参数:
//   - lineSize: 线条宽度
//   - lineColor: 线条颜色
//   - fillColor: 填充颜色
//   - sides: 边数
//   - startAngle: 起始角度（度数）
//   - distances: 每个顶点的距离倍数（可选）
//
// 示例：
//
//	graph.DrawRegularPolygon(1, "#000000", "#FFFF00", 6, 0, nil)  // 六边形
func (g *Graph) DrawRegularPolygon(lineSize float64, lineColor, fillColor string, sides int, startAngle float64, distances []float64) {
	g.graph.DrawRegularPolygon(lineSize, lineColor, fillColor, sides, startAngle, distances)
}

// Position 返回图形位置。
func (g *Graph) Position() (x, y float64) {
	return g.graph.X(), g.graph.Y()
}

// SetPosition 设置图形位置。
func (g *Graph) SetPosition(x, y float64) {
	g.graph.SetPosition(x, y)
}

// Size 返回图形大小。
func (g *Graph) Size() (width, height float64) {
	return g.graph.Width(), g.graph.Height()
}

// SetSize 设置图形大小。
func (g *Graph) SetSize(width, height float64) {
	g.graph.SetSize(width, height)
}

// Visible 返回图形是否可见。
func (g *Graph) Visible() bool {
	return g.graph.Visible()
}

// SetVisible 设置图形可见性。
func (g *Graph) SetVisible(visible bool) {
	g.graph.SetVisible(visible)
}

// Name 返回图形名称。
func (g *Graph) Name() string {
	return g.graph.Name()
}

// SetName 设置图形名称。
func (g *Graph) SetName(name string) {
	g.graph.SetName(name)
}

// Alpha 返回图形透明度（0-1）。
func (g *Graph) Alpha() float64 {
	return g.graph.Alpha()
}

// SetAlpha 设置图形透明度（0-1）。
func (g *Graph) SetAlpha(alpha float64) {
	g.graph.SetAlpha(alpha)
}

// RawGraph 返回底层的 widgets.GGraph 对象。
//
// 仅在需要访问底层 API 时使用。
func (g *Graph) RawGraph() *widgets.GGraph {
	return g.graph
}
