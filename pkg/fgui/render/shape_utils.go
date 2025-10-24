package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func computeStrokePadding(color *color.NRGBA, lineSize float64) float64 {
	if color == nil || lineSize <= 0 {
		return 0
	}
	return math.Ceil(lineSize*0.5 + 1)
}

func ensureGraphCanvasSize(width, height float64, strokePad float64) (int, int) {
	canvasW := int(math.Ceil(width + strokePad*2))
	canvasH := int(math.Ceil(height + strokePad*2))
	if canvasW <= 0 {
		canvasW = 1
	}
	if canvasH <= 0 {
		canvasH = 1
	}
	return canvasW, canvasH
}

func applyLocalOffset(geo ebiten.GeoM, dx, dy float64) ebiten.GeoM {
	local := ebiten.GeoM{}
	local.Translate(dx, dy)
	local.Concat(geo)
	return local
}

func buildRoundedRectPath(path *vector.Path, width, height float64, radii []float64, offsetX, offsetY float64) bool {
	if path == nil || width <= 0 || height <= 0 {
		return false
	}
	corners := normalizeCornerRadii(width, height, radii)
	w := float32(width)
	h := float32(height)
	ox := float32(offsetX)
	oy := float32(offsetY)
	tl, tr, br, bl := corners[0], corners[1], corners[2], corners[3]

	path.MoveTo(ox+tl, oy)
	path.LineTo(ox+w-tr, oy)
	if tr > 0 {
		path.Arc(ox+w-tr, oy+tr, tr, -math.Pi/2, 0, vector.Clockwise)
	} else {
		path.LineTo(ox+w, oy)
	}
	path.LineTo(ox+w, oy+h-br)
	if br > 0 {
		path.Arc(ox+w-br, oy+h-br, br, 0, math.Pi/2, vector.Clockwise)
	} else {
		path.LineTo(ox+w, oy+h)
	}
	path.LineTo(ox+bl, oy+h)
	if bl > 0 {
		path.Arc(ox+bl, oy+h-bl, bl, math.Pi/2, math.Pi, vector.Clockwise)
	} else {
		path.LineTo(ox, oy+h)
	}
	path.LineTo(ox, oy+tl)
	if tl > 0 {
		path.Arc(ox+tl, oy+tl, tl, math.Pi, 3*math.Pi/2, vector.Clockwise)
	} else {
		path.LineTo(ox, oy)
	}
	path.Close()
	return true
}

func buildEllipsePath(path *vector.Path, width, height float64, offsetX, offsetY float64) bool {
	if path == nil || width <= 0 || height <= 0 {
		return false
	}
	rx := float32(width * 0.5)
	ry := float32(height * 0.5)
	cx := float32(offsetX) + rx
	cy := float32(offsetY) + ry
	segments := 32
	step := 2 * math.Pi / float64(segments)
	path.MoveTo(cx+rx, cy)
	for i := 1; i <= segments; i++ {
		a := float64(i) * step
		path.LineTo(cx+rx*float32(math.Cos(a)), cy+ry*float32(math.Sin(a)))
	}
	path.Close()
	return true
}

func buildPolygonPath(path *vector.Path, points []float64, offsetX, offsetY float64) bool {
	if path == nil || len(points) < 6 {
		return false
	}
	path.MoveTo(float32(points[0]+offsetX), float32(points[1]+offsetY))
	for i := 2; i < len(points); i += 2 {
		path.LineTo(float32(points[i]+offsetX), float32(points[i+1]+offsetY))
	}
	path.Close()
	return true
}

func buildRegularPolygonPath(path *vector.Path, width, height float64, sides int, startAngle float64, distances []float64, offsetX, offsetY float64) bool {
	if path == nil || sides < 3 {
		return false
	}
	radius := math.Min(width, height) / 2
	angle := startAngle * math.Pi / 180
	delta := 2 * math.Pi / float64(sides)
	cx := offsetX + radius
	cy := offsetY + radius
	for i := 0; i < sides; i++ {
		dist := 1.0
		if i < len(distances) && !math.IsNaN(distances[i]) {
			dist = distances[i]
		}
		x := cx + radius*dist*math.Cos(angle)
		y := cy + radius*dist*math.Sin(angle)
		if i == 0 {
			path.MoveTo(float32(x), float32(y))
		} else {
			path.LineTo(float32(x), float32(y))
		}
		angle += delta
	}
	path.Close()
	return true
}

func buildPiePath(path *vector.Path, cx, cy, radius, startAngle, endAngle float64, offsetX, offsetY float64) bool {
	if path == nil || radius <= 0 {
		return false
	}
	start := startAngle * math.Pi / 180
	end := endAngle * math.Pi / 180
	if end < start {
		end += 2 * math.Pi
	}
	segments := int(math.Ceil((end - start) / (math.Pi / 12)))
	if segments < 2 {
		segments = 2
	}
	step := (end - start) / float64(segments)
	cx += offsetX
	cy += offsetY
	path.MoveTo(float32(cx), float32(cy))
	for i := 0; i <= segments; i++ {
		a := start + float64(i)*step
		x := cx + radius*math.Cos(a)
		y := cy + radius*math.Sin(a)
		path.LineTo(float32(x), float32(y))
	}
	path.Close()
	return true
}

func normalizeCornerRadii(width, height float64, input []float64) [4]float32 {
	var radii [4]float32
	if len(input) == 0 {
		return radii
	}
	if len(input) == 1 {
		v := float32(input[0])
		for i := range radii {
			radii[i] = v
		}
		return radii
	}
	for i := 0; i < 4; i++ {
		val := input[i%len(input)]
		if val < 0 {
			val = 0
		}
		radii[i] = float32(val)
	}
	maxH := float32(width / 2)
	maxV := float32(height / 2)
	for i := range radii {
		if radii[i] > maxH {
			radii[i] = maxH
		}
		if radii[i] > maxV {
			radii[i] = maxV
		}
	}
	return radii
}
