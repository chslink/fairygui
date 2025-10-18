package laya

import "math"

// Sprite emulates the subset of Laya.Sprite behaviour required by FairyGUI.
type Sprite struct {
	dispatcher *BasicEventDispatcher
	parent     *Sprite
	children   []*Sprite
	visible    bool
	alpha      float64
	name       string
	position   Point
	scaleX     float64
	scaleY     float64
	rotation   float64
	width      float64
	height     float64
	pivot      Point // normalized (0-1) pivot factors
}

// NewSprite constructs a sprite with sensible defaults.
func NewSprite() *Sprite {
	return &Sprite{
		dispatcher: NewEventDispatcher(),
		visible:    true,
		alpha:      1.0,
		scaleX:     1,
		scaleY:     1,
	}
}

// Name returns the sprite's identifier.
func (s *Sprite) Name() string {
	return s.name
}

// SetName updates the sprite's identifier.
func (s *Sprite) SetName(name string) {
	s.name = name
}

// Parent returns the sprite's parent.
func (s *Sprite) Parent() *Sprite {
	return s.parent
}

// Children exposes the sprite's direct children.
func (s *Sprite) Children() []*Sprite {
	return append([]*Sprite(nil), s.children...)
}

// AddChild attaches a child sprite at the end of the list.
func (s *Sprite) AddChild(child *Sprite) {
	s.AddChildAt(child, len(s.children))
}

// AddChildAt inserts a child sprite at the specified index.
func (s *Sprite) AddChildAt(child *Sprite, index int) {
	if child == nil || child == s {
		return
	}
	if child.parent != nil {
		child.parent.RemoveChild(child)
	}
	if index < 0 {
		index = 0
	} else if index > len(s.children) {
		index = len(s.children)
	}
	child.parent = s
	s.children = append(s.children, nil)
	copy(s.children[index+1:], s.children[index:])
	s.children[index] = child
	s.dispatcher.Emit(EventAdded, child)
}

// RemoveChild detaches a child sprite.
func (s *Sprite) RemoveChild(child *Sprite) {
	for i, current := range s.children {
		if current == child {
			s.RemoveChildAt(i)
			return
		}
	}
}

// RemoveChildAt removes a child by index.
func (s *Sprite) RemoveChildAt(index int) {
	if index < 0 || index >= len(s.children) {
		return
	}
	child := s.children[index]
	child.parent = nil
	copy(s.children[index:], s.children[index+1:])
	s.children = s.children[:len(s.children)-1]
	s.dispatcher.Emit(EventRemoved, child)
}

// RemoveChildren clears all children.
func (s *Sprite) RemoveChildren() {
	for len(s.children) > 0 {
		s.RemoveChildAt(len(s.children) - 1)
	}
}

// Visible reports whether the sprite should participate in rendering.
func (s *Sprite) Visible() bool {
	return s.visible
}

// SetVisible toggles sprite visibility.
func (s *Sprite) SetVisible(v bool) {
	if s.visible == v {
		return
	}
	s.visible = v
	if v {
		s.dispatcher.Emit(EventDisplay, s)
	} else {
		s.dispatcher.Emit(EventUndisplay, s)
	}
}

// Alpha returns transparency in [0,1].
func (s *Sprite) Alpha() float64 {
	return s.alpha
}

// SetAlpha updates transparency.
func (s *Sprite) SetAlpha(v float64) {
	if v < 0 {
		v = 0
	} else if v > 1 {
		v = 1
	}
	s.alpha = v
}

// Dispatcher exposes the internal event dispatcher.
func (s *Sprite) Dispatcher() EventDispatcher {
	return s.dispatcher
}

// SetPosition sets the local position relative to the parent.
func (s *Sprite) SetPosition(x, y float64) {
	if s.position.X == x && s.position.Y == y {
		return
	}
	s.position = Point{X: x, Y: y}
	s.dispatcher.Emit(EventXYChanged, s)
}

// Position returns the local position.
func (s *Sprite) Position() Point {
	return s.position
}

// Move offsets the sprite by the given delta.
func (s *Sprite) Move(dx, dy float64) {
	s.SetPosition(s.position.X+dx, s.position.Y+dy)
}

// SetScale updates the local scale factors.
func (s *Sprite) SetScale(sx, sy float64) {
	if s.scaleX == sx && s.scaleY == sy {
		return
	}
	s.scaleX = sx
	s.scaleY = sy
}

// Scale returns the local scale factors.
func (s *Sprite) Scale() (float64, float64) {
	return s.scaleX, s.scaleY
}

// SetRotation sets the rotation in radians.
func (s *Sprite) SetRotation(radians float64) {
	s.rotation = radians
}

// Rotation returns the rotation in radians.
func (s *Sprite) Rotation() float64 {
	return s.rotation
}

// SetSize updates the sprite's logical bounds.
func (s *Sprite) SetSize(width, height float64) {
	s.width = width
	s.height = height
}

// Size returns the logical bounds.
func (s *Sprite) Size() (float64, float64) {
	return s.width, s.height
}

// SetPivot defines the normalized pivot factors (0..1).
func (s *Sprite) SetPivot(px, py float64) {
	s.pivot = Point{X: px, Y: py}
}

// Pivot returns the normalized pivot factors.
func (s *Sprite) Pivot() Point {
	return s.pivot
}

// LocalToGlobal converts a local point to global coordinates.
func (s *Sprite) LocalToGlobal(pt Point) Point {
	world := s.worldMatrix()
	return world.Apply(pt)
}

// GlobalToLocal converts a global point to local coordinates.
func (s *Sprite) GlobalToLocal(pt Point) Point {
	world := s.worldMatrix()
	if inv, ok := world.Invert(); ok {
		return inv.Apply(pt)
	}
	return Point{}
}

// Bounds returns the axis-aligned bounding box in global coordinates.
func (s *Sprite) Bounds() Rect {
	w, h := s.width, s.height
	matrix := s.worldMatrix()
	points := []Point{
		matrix.Apply(Point{X: 0, Y: 0}),
		matrix.Apply(Point{X: w, Y: 0}),
		matrix.Apply(Point{X: w, Y: h}),
		matrix.Apply(Point{X: 0, Y: h}),
	}
	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, p := range points[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

func (s *Sprite) localMatrix() Matrix {
	pivotX := s.pivot.X * s.width
	pivotY := s.pivot.Y * s.height
	cos := math.Cos(s.rotation)
	sin := math.Sin(s.rotation)

	a := cos * s.scaleX
	b := sin * s.scaleX
	c := -sin * s.scaleY
	d := cos * s.scaleY
	tx := s.position.X - pivotX*a - pivotY*c
	ty := s.position.Y - pivotX*b - pivotY*d

	return Matrix{
		A:  a,
		B:  b,
		C:  c,
		D:  d,
		Tx: tx,
		Ty: ty,
	}
}

func (s *Sprite) worldMatrix() Matrix {
	local := s.localMatrix()
	if s.parent == nil {
		return local
	}
	return s.parent.worldMatrix().Multiply(local)
}
