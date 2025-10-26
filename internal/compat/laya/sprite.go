package laya

import (
	"math"
)

// BlendMode enumerates available blending operations.
type BlendMode int

const (
	BlendModeNormal BlendMode = iota
	BlendModeAdd
)

// Sprite emulates the subset of Laya.Sprite behaviour required by FairyGUI.
type Sprite struct {
	dispatcher    *BasicEventDispatcher
	parent        *Sprite
	children      []*Sprite
	visible       bool
	owner         any
	alpha         float64
	name          string
	mouseEnabled  bool
	mouseThrough  bool // 允许鼠标事件穿透到子对象
	position      Point
	rawPosition   Point
	scaleX        float64
	scaleY        float64
	rotation      float64
	width         float64
	height        float64
	pivot         Point // normalized (0-1) pivot factors
	scrollRect    *Rect
	pivotAsAnchor bool
	pivotOffset   Point
	skewX         float64
	skewY         float64
	hitTester     func(x, y float64) bool
	graphics      *Graphics
	hitArea       *HitArea
	repaintDirty  bool
	colorFilter   [4]float64
	filterEnabled bool
	colorMatrix        [20]float64
	colorMatrixEnabled bool
	grayEnabled        bool
	blendMode          BlendMode
}

// NewSprite constructs a sprite with sensible defaults.
func NewSprite() *Sprite {
	return &Sprite{
		dispatcher:   NewEventDispatcher(),
		visible:      true,
		alpha:        1.0,
		mouseEnabled: true,
		scaleX:       1,
		scaleY:       1,
		colorMatrix:  identityColorMatrix,
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

// SetOwner associates arbitrary metadata with the sprite (e.g., owning UI element).
func (s *Sprite) SetOwner(owner any) {
	s.owner = owner
}

// Owner returns the metadata previously set via SetOwner.
func (s *Sprite) Owner() any {
	return s.owner
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

// MouseEnabled reports whether the sprite should respond to pointer hit tests.
func (s *Sprite) MouseEnabled() bool {
	return s.mouseEnabled
}

// SetMouseEnabled toggles participation in hit testing for this sprite.
func (s *Sprite) SetMouseEnabled(enabled bool) {
	s.mouseEnabled = enabled
}

// MouseThrough returns whether mouse events pass through to children.
// When true, this sprite won't intercept mouse events, allowing children to receive them.
func (s *Sprite) MouseThrough() bool {
	return s.mouseThrough
}

// SetMouseThrough sets whether mouse events pass through to children.
// When true, this sprite won't intercept mouse events, allowing children to receive them.
func (s *Sprite) SetMouseThrough(through bool) {
	s.mouseThrough = through
}

// SetScrollRect defines the clipping rectangle applied to this sprite.
func (s *Sprite) SetScrollRect(rect *Rect) {
	if rect == nil {
		s.scrollRect = nil
		return
	}
	copy := *rect
	s.scrollRect = &copy
}

// ScrollRect returns a copy of the sprite's clipping rectangle, if any.
func (s *Sprite) ScrollRect() *Rect {
	if s.scrollRect == nil {
		return nil
	}
	copy := *s.scrollRect
	return &copy
}

// Dispatcher exposes the internal event dispatcher.
func (s *Sprite) Dispatcher() EventDispatcher {
	return s.dispatcher
}

// SetHitTester registers a custom hit testing function evaluated in local coordinates.
func (s *Sprite) SetHitTester(fn func(x, y float64) bool) {
	s.hitTester = fn
}

// SetColorFilter stores colour filter coefficients (applied by renderer).
func (s *Sprite) SetColorFilter(brightness, contrast, saturation, hue float64) {
	s.colorFilter = [4]float64{brightness, contrast, saturation, hue}
	matrix, enabled := computeColorMatrix(brightness, contrast, saturation, hue)
	s.colorMatrix = matrix
	s.colorMatrixEnabled = enabled
	s.filterEnabled = enabled
	if enabled {
		s.grayEnabled = false
	}
}

func (s *Sprite) ClearColorFilter() {
	s.colorFilter = [4]float64{}
	s.filterEnabled = false
	s.colorMatrix = identityColorMatrix
	s.colorMatrixEnabled = false
	s.grayEnabled = false
}

// ColorFilter returns current colour filter and enabled flag.
func (s *Sprite) ColorFilter() (enabled bool, values [4]float64) {
	return s.filterEnabled, s.colorFilter
}

// ColorEffects reports grayscale usage and the current colour matrix.
func (s *Sprite) ColorEffects() (gray bool, enabled bool, matrix [20]float64) {
	return s.grayEnabled, s.colorMatrixEnabled, s.colorMatrix
}

// SetGray toggles grayscale rendering for the sprite.
func (s *Sprite) SetGray(enabled bool) {
	s.grayEnabled = enabled
}

// SetBlendMode updates the blending mode.
func (s *Sprite) SetBlendMode(mode BlendMode) {
	s.blendMode = mode
}

// BlendMode reports the current blending mode.
func (s *Sprite) BlendMode() BlendMode {
	return s.blendMode
}

// Graphics returns the sprite's drawing command collection.
func (s *Sprite) Graphics() *Graphics {
	if s.graphics == nil {
		s.graphics = NewGraphics()
	}
	return s.graphics
}

// HitArea returns the sprite's hit area.
func (s *Sprite) HitArea() *HitArea {
	return s.hitArea
}

// SetHitArea assigns a hit area used during hit tests.
func (s *Sprite) SetHitArea(area *HitArea) {
	s.hitArea = area
}

// Repaint marks the sprite as needing redraw.
func (s *Sprite) Repaint() {
	s.repaintDirty = true
}

// ConsumeRepaint resets the repaint flag and reports the previous state.
func (s *Sprite) ConsumeRepaint() bool {
	dirty := s.repaintDirty
	s.repaintDirty = false
	return dirty
}

// Emit dispatches an event without bubbling.
func (s *Sprite) Emit(evt EventType, data any) {
	s.dispatcher.Emit(evt, data)
}

// EmitWithBubble dispatches an event on the sprite and parents up to the root.
func (s *Sprite) EmitWithBubble(evt EventType, data any) {
	for current := s; current != nil; current = current.parent {
		current.dispatcher.Emit(evt, data)
	}
}

// SetPosition sets the local position relative to the parent.
func (s *Sprite) SetPosition(x, y float64) {
	s.rawPosition = Point{X: x, Y: y}
	s.applyPivotOffset(true)
}

// Position returns the actual local position (including pivot offsets).
func (s *Sprite) Position() Point {
	return s.position
}

// Move offsets the sprite by the given delta.
func (s *Sprite) Move(dx, dy float64) {
	s.SetPosition(s.rawPosition.X+dx, s.rawPosition.Y+dy)
}

// SetScale updates the local scale factors.
func (s *Sprite) SetScale(sx, sy float64) {
	if s.scaleX == sx && s.scaleY == sy {
		return
	}
	s.scaleX = sx
	s.scaleY = sy
	s.updatePivotOffset()
	s.applyPivotOffset(true)
}

// Scale returns the local scale factors.
func (s *Sprite) Scale() (float64, float64) {
	return s.scaleX, s.scaleY
}

// SetRotation sets the rotation in radians.
func (s *Sprite) SetRotation(radians float64) {
	if s.rotation == radians {
		return
	}
	s.rotation = radians
	s.updatePivotOffset()
	s.applyPivotOffset(true)
}

// Rotation returns the rotation in radians.
func (s *Sprite) Rotation() float64 {
	return s.rotation
}

// SetSkew updates the skew factors (in radians) on both axes.
func (s *Sprite) SetSkew(sx, sy float64) {
	if s.skewX == sx && s.skewY == sy {
		return
	}
	s.skewX = sx
	s.skewY = sy
	s.updatePivotOffset()
	s.applyPivotOffset(true)
}

// Skew returns the skew factors.
func (s *Sprite) Skew() (float64, float64) {
	return s.skewX, s.skewY
}

// SetSize updates the sprite's logical bounds.
func (s *Sprite) SetSize(width, height float64) {
	if s.width == width && s.height == height {
		return
	}
	s.width = width
	s.height = height
	s.updatePivotOffset()
	s.applyPivotOffset(true)
}

// Size returns the logical bounds.
func (s *Sprite) Size() (float64, float64) {
	return s.width, s.height
}

// SetPivot defines the normalized pivot factors (0..1).
func (s *Sprite) SetPivot(px, py float64) {
	s.SetPivotWithAnchor(px, py, s.pivotAsAnchor)
}

// SetPivotWithAnchor defines the pivot and whether it should act as an anchor.
func (s *Sprite) SetPivotWithAnchor(px, py float64, asAnchor bool) {
	if s.pivot.X == px && s.pivot.Y == py && s.pivotAsAnchor == asAnchor {
		return
	}
	s.pivot = Point{X: px, Y: py}
	s.pivotAsAnchor = asAnchor
	s.updatePivotOffset()
	s.applyPivotOffset(true)
}

// Pivot returns the normalized pivot factors.
func (s *Sprite) Pivot() Point {
	return s.pivot
}

func (s *Sprite) applyPivotOffset(emit bool) {
	actualX := s.rawPosition.X + s.pivotOffset.X
	actualY := s.rawPosition.Y + s.pivotOffset.Y
	changed := s.position.X != actualX || s.position.Y != actualY
	s.position = Point{X: actualX, Y: actualY}
	if changed && emit {
		s.dispatcher.Emit(EventXYChanged, s)
	}
}

func (s *Sprite) updatePivotOffset() {
	// 只有当 pivotAsAnchor = true 时才需要计算偏移
	// 否则轴心点仅作为旋转中心，不影响位置
	if !s.pivotAsAnchor {
		s.pivotOffset = Point{}
		return
	}
	if s.pivot.X == 0 && s.pivot.Y == 0 {
		s.pivotOffset = Point{}
		return
	}
	px := s.pivot.X * s.width
	py := s.pivot.Y * s.height

	cosY := math.Cos(s.rotation + s.skewY)
	sinY := math.Sin(s.rotation + s.skewY)
	cosX := math.Cos(s.rotation - s.skewX)  // 恢复减法
	sinX := math.Sin(s.rotation - s.skewX)  // 恢复减法

	a := cosY * s.scaleX
	b := sinY * s.scaleX
	c := -sinX * s.scaleY  // 恢复负号
	d := cosX * s.scaleY

	transformedX := a*px + c*py
	transformedY := b*px + d*py
	s.pivotOffset = Point{
		X: px - transformedX,
		Y: py - transformedY,
	}
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
	rot := s.rotation
	skewX := s.skewX
	skewY := s.skewY

	cosY := math.Cos(rot + skewY)
	sinY := math.Sin(rot + skewY)
	cosX := math.Cos(rot - skewX)  // 恢复减法
	sinX := math.Sin(rot - skewX)  // 恢复减法

	a := cosY * s.scaleX
	b := sinY * s.scaleX
	c := -sinX * s.scaleY  // 恢复负号
	d := cosX * s.scaleY

	var tx, ty float64
	if s.pivotAsAnchor {
		// 当pivotAsAnchor=true时，rawPosition表示pivot点的全局位置
		// 公式：rawPosition = (pivot在局部) 经过变换后的位置
		// rawPosition.X = a*pivotX + c*pivotY + tx
		// 解出：tx = rawPosition.X - a*pivotX - c*pivotY
		tx = s.rawPosition.X - pivotX*a - pivotY*c
		ty = s.rawPosition.Y - pivotX*b - pivotY*d
	} else {
		// 当pivotAsAnchor=false时，rawPosition表示左上角(0,0)的位置
		// 但仍需围绕pivot点进行旋转缩放
		// 标准pivot变换：先移到pivot，变换，再移回，最后平移到目标
		// tx = position.X - pivotX*a - pivotY*c + pivotX
		tx = s.rawPosition.X - pivotX*a - pivotY*c + pivotX
		ty = s.rawPosition.Y - pivotX*b - pivotY*d + pivotY
	}

	return Matrix{
		A:  a,
		B:  b,
		C:  c,
		D:  d,
		Tx: tx,
		Ty: ty,
	}
}

// LocalMatrix returns a copy of the local transform matrix.
func (s *Sprite) LocalMatrix() Matrix {
	return s.localMatrix()
}

// PivotOffset returns the cached pivot offset in local coordinates.
func (s *Sprite) PivotOffset() Point {
	return s.pivotOffset
}

func (s *Sprite) worldMatrix() Matrix {
	local := s.localMatrix()
	if s.parent == nil {
		return local
	}
	return s.parent.worldMatrix().Multiply(local)
}

// HitTest returns the deepest child sprite that contains the given global point.
func (s *Sprite) HitTest(global Point) *Sprite {
	if !s.visible {
		return nil
	}
	local := s.GlobalToLocal(global)
	if rect := s.scrollRect; rect != nil {
		right := rect.X + rect.W
		bottom := rect.Y + rect.H
		if rect.W <= 0 || rect.H <= 0 ||
			local.X < rect.X || local.X > right ||
			local.Y < rect.Y || local.Y > bottom {
			return nil
		}
	}
	for i := len(s.children) - 1; i >= 0; i-- {
		if hit := s.children[i].HitTest(global); hit != nil {
			return hit
		}
	}
	if s.scrollRect == nil {
		if s.width == 0 && s.height == 0 {
			return nil
		}
		if local.X < 0 || local.Y < 0 || local.X > s.width || local.Y > s.height {
			return nil
		}
	}
	if s.hitTester != nil && !s.hitTester(local.X, local.Y) {
		return nil
	}
	if s.hitArea != nil && !s.hitArea.Contains(local.X, local.Y) {
		return nil
	}
	if !s.mouseEnabled {
		return nil
	}
	// 如果设置了 mouseThrough，则不拦截事件，允许穿透
	if s.mouseThrough {
		return nil
	}
	return s
}
