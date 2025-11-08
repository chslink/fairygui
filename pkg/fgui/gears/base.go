package gears

import (
	"github.com/chslink/fairygui/pkg/fgui/tween"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// Index constants match FairyGUI's gear ordering.
const (
	IndexDisplay = iota
	IndexXY
	IndexSize
	IndexLook
	IndexColor
	IndexAnimation
	IndexText
	IndexIcon
	IndexDisplay2
	IndexFontSize
	SlotCount
)

// Owner captures the minimal behaviour gears need from a GObject.
type Owner interface {
	ID() string
	X() float64
	Y() float64
	SetPosition(x, y float64)
	Width() float64
	Height() float64
	SetSize(width, height float64)
	Scale() (float64, float64)
	SetScale(scaleX, scaleY float64)
	ParentSize() (width, height float64)
	SetGearLocked(locked bool)
	GearLocked() bool
	Visible() bool
	SetVisible(visible bool)
	Alpha() float64
	SetAlpha(a float64)
	Rotation() float64
	SetRotation(radians float64)
	Grayed() bool
	SetGrayed(grayed bool)
	Touchable() bool
	SetTouchable(touchable bool)
	GetProp(id ObjectPropID) any
	SetProp(id ObjectPropID, value any)
}

// Controller is a soft dependency that exposes the selected page.
type Controller interface {
	SelectedPageID() string
	SelectedIndex() int
}

// ObjectPropID enumerates generic GObject property identifiers used by gears.
type ObjectPropID int

const (
	ObjectPropIDText ObjectPropID = iota
	ObjectPropIDIcon
	ObjectPropIDColor
	ObjectPropIDOutlineColor
	ObjectPropIDPlaying
	ObjectPropIDFrame
	ObjectPropIDDeltaTime
	ObjectPropIDTimeScale
	ObjectPropIDFontSize
	ObjectPropIDSelected
)

// ControllerResolver resolves controllers by index for gear setup.
type ControllerResolver interface {
	ControllerAt(index int) Controller
}

// Gear exposes the lifecycle hooks shared across concrete implementations.
type Gear interface {
	Owner() Owner
	Index() int
	Controller() Controller
	SetController(ctrl Controller)
	Setup(buffer *utils.ByteBuffer, resolver ControllerResolver)
	UpdateState()
	Apply()
	UpdateFromRelations(dx, dy float64)
}

// Base hosts shared state for concrete gears.
type Base struct {
	owner       Owner
	index       int
	controller  Controller
	tweenConfig *TweenConfig
}

// NewBase wires the owner/index pair for a new gear.
func NewBase(owner Owner, index int) Base {
	return Base{
		owner: owner,
		index: index,
	}
}

// Owner returns the gear owner.
func (b *Base) Owner() Owner {
	if b == nil {
		return nil
	}
	return b.owner
}

// Index returns the gear index.
func (b *Base) Index() int {
	if b == nil {
		return -1
	}
	return b.index
}

// Controller returns the active controller.
func (b *Base) Controller() Controller {
	if b == nil {
		return nil
	}
	return b.controller
}

// SetController swaps the bound controller reference.
func (b *Base) SetController(ctrl Controller) {
	if b == nil {
		return
	}
	b.controller = ctrl
}

// TweenConfig returns the tween configuration, creating it on first access.
func (b *Base) TweenConfig() *TweenConfig {
	if b == nil {
		return nil
	}
	if b.tweenConfig == nil {
		b.tweenConfig = &TweenConfig{
			Duration: 0.3,
			EaseType: tween.EaseTypeQuadOut,
		}
	}
	return b.tweenConfig
}

// TweenConfig holds tween parameters shared across gears.
type TweenConfig struct {
	Tween            bool
	EaseType         tween.EaseType
	Duration         float64
	Delay            float64
	DisplayLockToken uint32
	Tweener          *tween.GTweener
}

// DisableAllTweenEffect mirrors FairyGUI's static toggle.
var DisableAllTweenEffect bool

// Create instantiates a gear implementation for the specified slot.
func Create(owner Owner, index int) Gear {
	switch index {
	case IndexDisplay:
		return NewGearDisplay(owner)
	case IndexXY:
		return NewGearXY(owner)
	case IndexSize:
		return NewGearSize(owner)
	case IndexLook:
		return NewGearLook(owner)
	case IndexColor:
		return NewGearColor(owner)
	case IndexAnimation:
		return NewGearAnimation(owner)
	case IndexText:
		return NewGearText(owner)
	case IndexIcon:
		return NewGearIcon(owner)
	case IndexDisplay2:
		return NewGearDisplay2(owner)
	case IndexFontSize:
		return NewGearFontSize(owner)
	default:
		return nil
	}
}
