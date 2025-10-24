package widgets

import (
	"math"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// GraphType mirrors FairyGUI's graph shape enumeration.
type GraphType int

const (
	GraphTypeEmpty GraphType = iota
	GraphTypeRect
	GraphTypeEllipse
	GraphTypePolygon
	GraphTypeRegularPolygon
)

// GGraph represents a graphic widget (rectangles, ellipses, polygons).
type GGraph struct {
	*core.GObject
	graphType     GraphType
	lineSize      float64
	lineColor     string
	fillColor     string
	cornerRadius  []float64
	polygonPoints []float64
	sides         int
	startAngle    float64
	distances     []float64
}

// NewGraph creates a new graphic widget with default styling.
func NewGraph() *GGraph {
	obj := core.NewGObject()
	graph := &GGraph{
		GObject:   obj,
		lineColor: "#000000",
		fillColor: "#ffffff",
	}
	obj.SetData(graph)
	return graph
}

// Type returns the shape type currently stored on the graph.
func (g *GGraph) Type() GraphType {
	if g == nil {
		return GraphTypeEmpty
	}
	return g.graphType
}

// SetType updates the shape type.
func (g *GGraph) SetType(t GraphType) {
	if g == nil {
		return
	}
	g.graphType = t
	g.updateGraph()
}

// LineSize returns the stroke width.
func (g *GGraph) LineSize() float64 {
	if g == nil {
		return 0
	}
	return g.lineSize
}

// SetLineSize stores the stroke width.
func (g *GGraph) SetLineSize(size float64) {
	if g == nil {
		return
	}
	if size < 0 {
		size = 0
	}
	if g.lineSize == size {
		return
	}
	g.lineSize = size
	g.updateGraph()
}

// LineColor returns the stroke colour.
func (g *GGraph) LineColor() string {
	if g == nil {
		return ""
	}
	return g.lineColor
}

// SetLineColor updates the stroke colour.
func (g *GGraph) SetLineColor(value string) {
	if g == nil {
		return
	}
	if g.lineColor == value {
		return
	}
	g.lineColor = value
	g.updateGraph()
}

// FillColor returns the fill colour.
func (g *GGraph) FillColor() string {
	if g == nil {
		return ""
	}
	return g.fillColor
}

// SetFillColor updates the fill colour.
func (g *GGraph) SetFillColor(value string) {
	if g == nil {
		return
	}
	if g.fillColor == value {
		return
	}
	g.fillColor = value
	g.updateGraph()
}

// Color returns the fill colour (alias used by gears).
func (g *GGraph) Color() string {
	return g.FillColor()
}

// SetColor updates the fill colour (alias used by gears).
func (g *GGraph) SetColor(value string) {
	g.SetFillColor(value)
}

// CornerRadius returns a copy of the rounded corner radii.
func (g *GGraph) CornerRadius() []float64 {
	if g == nil || len(g.cornerRadius) == 0 {
		return nil
	}
	out := make([]float64, len(g.cornerRadius))
	copy(out, g.cornerRadius)
	return out
}

// SetCornerRadius stores the rounded corner radii (expects up to four values).
func (g *GGraph) SetCornerRadius(values []float64) {
	if g == nil {
		return
	}
	if len(values) == 0 {
		g.cornerRadius = nil
		g.updateGraph()
		return
	}
	g.cornerRadius = make([]float64, len(values))
	copy(g.cornerRadius, values)
	g.updateGraph()
}

// PolygonPoints returns a copy of the polygon point list.
func (g *GGraph) PolygonPoints() []float64 {
	if g == nil || len(g.polygonPoints) == 0 {
		return nil
	}
	out := make([]float64, len(g.polygonPoints))
	copy(out, g.polygonPoints)
	return out
}

// SetPolygonPoints records raw polygon point coordinates.
func (g *GGraph) SetPolygonPoints(points []float64) {
	if g == nil {
		return
	}
	if len(points) == 0 {
		g.polygonPoints = nil
		g.updateGraph()
		return
	}
	g.polygonPoints = make([]float64, len(points))
	copy(g.polygonPoints, points)
	g.updateGraph()
}

// RegularPolygon returns sides, start angle (degrees), and distance multipliers.
func (g *GGraph) RegularPolygon() (int, float64, []float64) {
	if g == nil {
		return 0, 0, nil
	}
	dist := make([]float64, len(g.distances))
	copy(dist, g.distances)
	return g.sides, g.startAngle, dist
}

// SetRegularPolygon configures a regular polygon definition.
func (g *GGraph) SetRegularPolygon(sides int, startAngle float64, distances []float64) {
	if g == nil {
		return
	}
	if sides < 0 {
		sides = 0
	}
	g.sides = sides
	g.startAngle = startAngle
	if len(distances) == 0 {
		g.distances = nil
	} else {
		g.distances = make([]float64, len(distances))
		copy(g.distances, distances)
	}
	g.updateGraph()
}

// DrawRect configures并绘制矩形。
func (g *GGraph) DrawRect(lineSize float64, lineColor, fillColor string, cornerRadius []float64) {
	if g == nil {
		return
	}
	g.SetType(GraphTypeRect)
	g.SetLineSize(lineSize)
	g.SetLineColor(lineColor)
	g.SetFillColor(fillColor)
	g.SetCornerRadius(cornerRadius)
}

// DrawEllipse configures并绘制椭圆。
func (g *GGraph) DrawEllipse(lineSize float64, lineColor, fillColor string) {
	if g == nil {
		return
	}
	g.SetType(GraphTypeEllipse)
	g.SetLineSize(lineSize)
	g.SetLineColor(lineColor)
	g.SetFillColor(fillColor)
}

// DrawPolygon 配置并绘制多边形。
func (g *GGraph) DrawPolygon(lineSize float64, lineColor, fillColor string, points []float64) {
	if g == nil {
		return
	}
	g.SetType(GraphTypePolygon)
	g.SetLineSize(lineSize)
	g.SetLineColor(lineColor)
	g.SetFillColor(fillColor)
	g.SetPolygonPoints(points)
}

// DrawRegularPolygon 配置并绘制正多边形。
func (g *GGraph) DrawRegularPolygon(lineSize float64, lineColor, fillColor string, sides int, startAngle float64, distances []float64) {
	if g == nil {
		return
	}
	g.SetType(GraphTypeRegularPolygon)
	g.SetLineSize(lineSize)
	g.SetLineColor(lineColor)
	g.SetFillColor(fillColor)
	g.SetRegularPolygon(sides, startAngle, distances)
}

// OwnerSizeChanged 在宿主尺寸变化时刷新图形。
func (g *GGraph) OwnerSizeChanged(_, _ float64) {
	g.updateGraph()
}

func (g *GGraph) updateGraph() {
	if g == nil || g.GObject == nil {
		return
	}
	obj := g.GObject
	sprite := obj.DisplayObject()
	if sprite == nil {
		return
	}
	sprite.SetMouseEnabled(obj.Touchable())
	graphics := sprite.Graphics()
	if graphics == nil {
		return
	}
	graphics.Clear()

	hitArea := sprite.HitArea()
	if hitArea == nil {
		hitArea = laya.NewHitArea()
		sprite.SetHitArea(hitArea)
	}

	if g.graphType == GraphTypeEmpty {
		hitArea.SetGraphics(graphics)
		sprite.Repaint()
		return
	}

	w := obj.Width()
	h := obj.Height()
	if w <= 0 || h <= 0 {
		hitArea.SetGraphics(graphics)
		sprite.Repaint()
		return
	}

	fillStyle := colorToFill(g.fillColor)
	strokeStyle := colorToStroke(g.lineColor, g.lineSize)

	switch g.graphType {
	case GraphTypeRect:
		if drew := g.drawRect(graphics, w, h, fillStyle, strokeStyle); !drew {
			graphics.DrawRect(0, 0, w, h, fillStyle, strokeStyle)
		}
	case GraphTypeEllipse:
		rx := w / 2
		ry := h / 2
		graphics.DrawEllipse(rx, ry, rx, ry, fillStyle, strokeStyle)
	case GraphTypePolygon:
		points := g.polygonPoints
		if len(points) >= 6 {
			graphics.DrawPolygon(points, fillStyle, strokeStyle)
		}
	case GraphTypeRegularPolygon:
		points := g.ensureRegularPolygonPoints(w, h)
		if len(points) >= 6 {
			graphics.DrawPolygon(points, fillStyle, strokeStyle)
		}
	default:
		graphics.DrawRect(0, 0, w, h, fillStyle, strokeStyle)
	}

	hitArea.SetGraphics(graphics)
	sprite.Repaint()
}

// SetupBeforeAdd loads graph definition from the component buffer.
func (g *GGraph) SetupBeforeAdd(_ *SetupContext, buf *utils.ByteBuffer) {
	if g == nil || buf == nil {
		return
	}
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(0, 5) || buf.Remaining() <= 0 {
		return
	}
	g.SetType(GraphType(buf.ReadByte()))
	if g.Type() == GraphTypeEmpty {
		return
	}
	if buf.Remaining() >= 4 {
		g.SetLineSize(float64(buf.ReadInt32()))
	}
	if buf.Remaining() >= 4 {
		g.SetLineColor(buf.ReadColorString(true))
	}
	if buf.Remaining() >= 4 {
		g.SetFillColor(buf.ReadColorString(true))
	}
	if buf.Remaining() > 0 && buf.ReadBool() {
		radii := make([]float64, 0, 4)
		for i := 0; i < 4 && buf.Remaining() >= 4; i++ {
			radii = append(radii, float64(buf.ReadFloat32()))
		}
		g.SetCornerRadius(radii)
	}
	switch g.Type() {
	case GraphTypePolygon:
		if buf.Remaining() >= 2 {
			cnt := int(buf.ReadInt16())
			points := make([]float64, 0, cnt)
			for i := 0; i < cnt && buf.Remaining() >= 4; i++ {
				points = append(points, float64(buf.ReadFloat32()))
			}
			g.SetPolygonPoints(points)
		}
	case GraphTypeRegularPolygon:
		sides := 0
		if buf.Remaining() >= 2 {
			sides = int(buf.ReadInt16())
		}
		angle := 0.0
		if buf.Remaining() >= 4 {
			angle = float64(buf.ReadFloat32())
		}
		var distances []float64
		if buf.Remaining() >= 2 {
			cnt := int(buf.ReadInt16())
			distances = make([]float64, 0, cnt)
			for i := 0; i < cnt && buf.Remaining() >= 4; i++ {
				distances = append(distances, float64(buf.ReadFloat32()))
			}
		}
		g.SetRegularPolygon(sides, angle, distances)
	}
}

func (g *GGraph) drawRect(graphics *laya.Graphics, width, height float64, fill *laya.FillStyle, stroke *laya.StrokeStyle) bool {
	radii := g.cornerRadius
	if len(radii) == 0 {
		return false
	}
	radiiCopy := make([]float64, len(radii))
	copy(radiiCopy, radii)
	graphics.DrawRoundRect(0, 0, width, height, radiiCopy, fill, stroke)
	return true
}

func (g *GGraph) ensureRegularPolygonPoints(width, height float64) []float64 {
	if g.sides <= 0 {
		return nil
	}
	radius := math.Min(width, height) / 2
	points := make([]float64, 0, g.sides*2)
	angle := g.startAngle * math.Pi / 180
	delta := 2 * math.Pi / float64(g.sides)
	for i := 0; i < g.sides; i++ {
		dist := 1.0
		if i < len(g.distances) && !math.IsNaN(g.distances[i]) {
			dist = g.distances[i]
		}
		cx := radius + radius*dist*math.Cos(angle)
		cy := radius + radius*dist*math.Sin(angle)
		points = append(points, cx, cy)
		angle += delta
	}
	g.polygonPoints = points
	return points
}

func colorToFill(color string) *laya.FillStyle {
	if color == "" {
		return nil
	}
	return &laya.FillStyle{Color: color}
}

func colorToStroke(color string, width float64) *laya.StrokeStyle {
	if color == "" || width <= 0 {
		return nil
	}
	return &laya.StrokeStyle{Color: color, Width: width}
}
