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

	parent  *GComponent
	alpha   float64
	visible bool
	data    any
}

// NewGObject creates a base object with a backing sprite.
func NewGObject() *GObject {
	counter := atomic.AddUint64(&gObjectCounter, 1)
	return &GObject{
		id:      fmt.Sprintf("gobj-%d", counter),
		display: laya.NewSprite(),
		alpha:   1.0,
		visible: true,
	}
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
	if g.display != nil {
		g.display.SetPosition(x, y)
	}
}

// SetSize updates width and height.
func (g *GObject) SetSize(width, height float64) {
	g.width = width
	g.height = height
	if g.display != nil {
		g.display.SetSize(width, height)
	}
}

// SetAlpha adjusts transparency.
func (g *GObject) SetAlpha(alpha float64) {
	g.alpha = alpha
	if g.display != nil {
		g.display.SetAlpha(alpha)
	}
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
