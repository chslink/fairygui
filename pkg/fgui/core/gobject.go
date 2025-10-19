package core

import (
	"fmt"
	"sync/atomic"

	"github.com/chslink/fairygui/internal/compat/laya"
)

var gObjectCounter uint64

// GObject is the base building block for all FairyGUI entities.
type GObject struct {
	id      string
	name    string
	display *laya.Sprite

	x      float64
	y      float64
	width  float64
	height float64
	scaleX float64
	scaleY float64

	parent        *GComponent
	alpha         float64
	visible       bool
	rotation      float64
	skewX         float64
	skewY         float64
	pivotX        float64
	pivotY        float64
	pivotAsAnchor bool
	data          any
}

// NewGObject creates a base object with a backing sprite.
func NewGObject() *GObject {
	counter := atomic.AddUint64(&gObjectCounter, 1)
	display := laya.NewSprite()
	obj := &GObject{
		id:      fmt.Sprintf("gobj-%d", counter),
		display: display,
		alpha:   1.0,
		visible: true,
		scaleX:  1.0,
		scaleY:  1.0,
	}
	display.SetOwner(obj)
	return obj
}

// ID returns the unique identifier.
func (g *GObject) ID() string {
	return g.id
}

// Name returns the display name.
func (g *GObject) Name() string {
	return g.name
}

// SetName updates the display name.
func (g *GObject) SetName(name string) {
	g.name = name
	if g.display != nil {
		g.display.SetName(name)
	}
}

// DisplayObject exposes the underlying compat sprite.
func (g *GObject) DisplayObject() *laya.Sprite {
	return g.display
}

// SetPosition moves the object within its parent coordinate space.
func (g *GObject) SetPosition(x, y float64) {
	g.x = x
	g.y = y
	g.refreshTransform()
}

// SetSize updates width and height.
func (g *GObject) SetSize(width, height float64) {
	g.width = width
	g.height = height
	if g.display != nil {
		g.display.SetSize(width, height)
	}
	g.refreshTransform()
}

// SetScale updates the scaling factors on both axes.
func (g *GObject) SetScale(scaleX, scaleY float64) {
	g.scaleX = scaleX
	g.scaleY = scaleY
	if g.display != nil {
		g.display.SetScale(scaleX, scaleY)
	}
	g.refreshTransform()
}

// Scale returns the scaling factors.
func (g *GObject) Scale() (float64, float64) {
	return g.scaleX, g.scaleY
}

// SetRotation stores the rotation in radians and mirrors it to the sprite.
func (g *GObject) SetRotation(radians float64) {
	g.rotation = radians
	if g.display != nil {
		g.display.SetRotation(radians)
	}
	g.refreshTransform()
}

// Rotation returns the current rotation in radians.
func (g *GObject) Rotation() float64 {
	return g.rotation
}

// SetSkew updates skew factors (in radians).
func (g *GObject) SetSkew(skewX, skewY float64) {
	g.skewX = skewX
	g.skewY = skewY
	if g.display != nil {
		g.display.SetSkew(skewX, skewY)
	}
	g.refreshTransform()
}

// Skew returns the skew factors.
func (g *GObject) Skew() (float64, float64) {
	return g.skewX, g.skewY
}

// SetPivot configures the normalized pivot point.
func (g *GObject) SetPivot(px, py float64) {
	g.SetPivotWithAnchor(px, py, false)
}

// SetPivotWithAnchor configures the pivot and whether it acts as an anchor.
func (g *GObject) SetPivotWithAnchor(px, py float64, asAnchor bool) {
	g.pivotX = px
	g.pivotY = py
	g.pivotAsAnchor = asAnchor
	if g.display != nil {
		g.display.SetPivotWithAnchor(px, py, asAnchor)
	}
	g.refreshTransform()
}

// Pivot returns the normalized pivot point.
func (g *GObject) Pivot() (float64, float64) {
	return g.pivotX, g.pivotY
}

// PivotAsAnchor reports whether the pivot acts as an anchor.
func (g *GObject) PivotAsAnchor() bool {
	return g.pivotAsAnchor
}

// SetAlpha adjusts transparency.
func (g *GObject) SetAlpha(alpha float64) {
	g.alpha = alpha
	if g.display != nil {
		g.display.SetAlpha(alpha)
	}
}

// Alpha returns the current alpha value.
func (g *GObject) Alpha() float64 {
	return g.alpha
}

// SetVisible toggles visibility.
func (g *GObject) SetVisible(visible bool) {
	g.visible = visible
	if g.display != nil {
		g.display.SetVisible(visible)
	}
}

// X returns the local X position.
func (g *GObject) X() float64 {
	return g.x
}

// Y returns the local Y position.
func (g *GObject) Y() float64 {
	return g.y
}

// Width returns the current width.
func (g *GObject) Width() float64 {
	return g.width
}

// Height returns the current height.
func (g *GObject) Height() float64 {
	return g.height
}

// SetData assigns arbitrary user data to the object.
func (g *GObject) SetData(value any) {
	g.data = value
}

// Data returns the user data associated with the object.
func (g *GObject) Data() any {
	return g.data
}

// Visible reports whether the object is visible.
func (g *GObject) Visible() bool {
	return g.visible
}

func (g *GObject) refreshTransform() {
	if g.display == nil {
		return
	}
	x := g.x
	y := g.y
	if g.pivotAsAnchor {
		x -= g.pivotX * g.width
		y -= g.pivotY * g.height
	}
	g.display.SetPosition(x, y)
}
