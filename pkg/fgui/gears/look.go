package gears

import (
	"github.com/chslink/fairygui/pkg/fgui/tween"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// LookValue captures appearance-related state.
type LookValue struct {
	Alpha     float64
	Rotation  float64
	Grayed    bool
	Touchable bool
}

// GearLook synchronises alpha/rotation/grayed/touchable against controller pages.
type GearLook struct {
	Base

	storage      map[string]LookValue
	defaultValue LookValue
	initialized  bool
}

// NewGearLook constructs a GearLook bound to the provided owner.
func NewGearLook(owner Owner) *GearLook {
	g := &GearLook{
		Base:    NewBase(owner, IndexLook),
		storage: make(map[string]LookValue),
	}
	g.ensureInit()
	return g
}

// SetController wires the controller reference.
// 参考 TypeScript 原版：GearBase.ts set controller()
// 只调用 init()，不调用 updateState()
func (g *GearLook) SetController(ctrl Controller) {
	if g == nil {
		return
	}
	if g.Controller() == ctrl {
		return
	}
	g.Base.SetController(ctrl)
	if ctrl != nil {
		g.ensureInit()
	}
}

// Setup initialises the gear from serialized data.
func (g *GearLook) Setup(buffer *utils.ByteBuffer, resolver ControllerResolver) {
	if g == nil || buffer == nil {
		return
	}
	ctrlIndex := int(buffer.ReadInt16())
	var ctrl Controller
	if resolver != nil && ctrlIndex >= 0 {
		ctrl = resolver.ControllerAt(ctrlIndex)
	}
	g.SetController(ctrl)
	g.ensureInit()

	count := int(buffer.ReadInt16())
	for i := 0; i < count; i++ {
		page := buffer.ReadS()
		if page == nil {
			continue
		}
		g.readStatus(*page, buffer)
	}
	if buffer.ReadBool() {
		g.readStatus("", buffer)
	}
	if buffer.ReadBool() {
		cfg := g.TweenConfig()
		cfg.Tween = true
		cfg.EaseType = tween.EaseType(buffer.ReadByte())
		cfg.Duration = float64(buffer.ReadFloat32())
		cfg.Delay = float64(buffer.ReadFloat32())
	}
	g.Apply()
}

// UpdateState snapshots the owner's appearance for the active page.
func (g *GearLook) UpdateState() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	val := LookValue{
		Alpha:     g.Owner().Alpha(),
		Rotation:  g.Owner().Rotation(),
		Grayed:    g.Owner().Grayed(),
		Touchable: g.Owner().Touchable(),
	}
	pageID := g.pageID()
	if pageID == "" {
		g.defaultValue = val
	} else {
		g.storage[pageID] = val
	}
}

// Apply restores the cached appearance for the active page.
func (g *GearLook) Apply() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	val := g.valueForPage(g.pageID())
	owner := g.Owner()
	owner.SetGrayed(val.Grayed)
	owner.SetTouchable(val.Touchable)
	cfg := g.TweenConfig()
	if cfg != nil && cfg.Tween && !DisableAllTweenEffect {
		if cfg.Tweener != nil {
			end := cfg.Tweener.EndValue()
			if end.X != val.Alpha || end.Y != val.Rotation {
				cfg.Tweener.Kill(false)
				cfg.Tweener = nil
			} else {
				return
			}
		}
		current := tween.Value{X: owner.Alpha(), Y: owner.Rotation()}
		tw := tween.To2(current.X, current.Y, val.Alpha, val.Rotation, cfg.Duration)
		tw.SetDelay(cfg.Delay)
		tw.SetEase(cfg.EaseType)
		tw.SetTarget(owner)
		tw.OnUpdate(func(tw *tween.GTweener) {
			v := tw.Value()
			owner.SetGearLocked(true)
			owner.SetAlpha(v.X)
			owner.SetRotation(v.Y)
			owner.SetGearLocked(false)
		})
		tw.OnComplete(func(*tween.GTweener) {
			owner.SetGearLocked(true)
			owner.SetAlpha(val.Alpha)
			owner.SetRotation(val.Rotation)
			owner.SetGearLocked(false)
			cfg.Tweener = nil
		})
		cfg.Tweener = tw
		return
	}
	if owner.GearLocked() {
		owner.SetAlpha(val.Alpha)
		owner.SetRotation(val.Rotation)
		return
	}
	owner.SetGearLocked(true)
	owner.SetAlpha(val.Alpha)
	owner.SetRotation(val.Rotation)
	owner.SetGearLocked(false)
}

// UpdateFromRelations is unused for GearLook but required by the interface.
func (g *GearLook) UpdateFromRelations(dx, dy float64) {
	// Look attributes are independent of relations.
}

// Value exposes the cached value for the provided page ID.
func (g *GearLook) Value(pageID string) LookValue {
	if g == nil {
		return LookValue{}
	}
	g.ensureInit()
	return g.valueForPage(pageID)
}

func (g *GearLook) ensureInit() {
	if g == nil || g.Owner() == nil || g.initialized {
		return
	}
	g.initialized = true
	g.storage = make(map[string]LookValue)
	g.defaultValue = LookValue{
		Alpha:     g.Owner().Alpha(),
		Rotation:  g.Owner().Rotation(),
		Grayed:    g.Owner().Grayed(),
		Touchable: g.Owner().Touchable(),
	}
}

func (g *GearLook) pageID() string {
	if g == nil || g.Controller() == nil {
		return ""
	}
	return g.Controller().SelectedPageID()
}

func (g *GearLook) valueForPage(pageID string) LookValue {
	if g == nil {
		return LookValue{}
	}
	if pageID == "" {
		return g.defaultValue
	}
	if val, ok := g.storage[pageID]; ok {
		return val
	}
	return g.defaultValue
}

func (g *GearLook) readStatus(page string, buffer *utils.ByteBuffer) {
	if g == nil || buffer == nil {
		return
	}
	val := LookValue{
		Alpha:     float64(buffer.ReadFloat32()),
		Rotation:  float64(buffer.ReadFloat32()),
		Grayed:    buffer.ReadBool(),
		Touchable: buffer.ReadBool(),
	}
	if page == "" {
		g.defaultValue = val
	} else {
		g.storage[page] = val
	}
}
