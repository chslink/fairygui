package laya

// Point mirrors the subset of Laya's Point needed by FairyGUI.
type Point struct {
	X float64
	Y float64
}

// Clone returns a shallow copy of the point.
func (p Point) Clone() Point {
	return Point{X: p.X, Y: p.Y}
}

// Offset moves the point by the given delta.
func (p *Point) Offset(dx, dy float64) {
	p.X += dx
	p.Y += dy
}

// Rect represents an axis-aligned rectangle.
type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

// Right returns the x coordinate of the right edge.
func (r Rect) Right() float64 {
	return r.X + r.W
}

// Bottom returns the y coordinate of the bottom edge.
func (r Rect) Bottom() float64 {
	return r.Y + r.H
}

// Contains tests whether the point lies within the rectangle.
func (r Rect) Contains(pt Point) bool {
	return pt.X >= r.X && pt.X <= r.Right() &&
		pt.Y >= r.Y && pt.Y <= r.Bottom()
}
