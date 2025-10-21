package widgets

import "github.com/chslink/fairygui/pkg/fgui/core"

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
	g.lineSize = size
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
	g.lineColor = value
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
	g.fillColor = value
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
		return
	}
	g.cornerRadius = make([]float64, len(values))
	copy(g.cornerRadius, values)
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
		return
	}
	g.polygonPoints = make([]float64, len(points))
	copy(g.polygonPoints, points)
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
}
