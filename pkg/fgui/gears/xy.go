package gears

import (
	"github.com/chslink/fairygui/pkg/fgui/tween"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// XYValue captures a stored positional state.
type XYValue struct {
	X  float64
	Y  float64
	PX float64
	PY float64
}

// GearXY synchronises position against controller pages.
type GearXY struct {
	Base

	positionsInPercent bool
	storage            map[string]XYValue
	defaultValue       XYValue
	initialized        bool
}

// NewGearXY constructs a GearXY instance bound to the provided owner.
func NewGearXY(owner Owner) *GearXY {
	g := &GearXY{
		Base:    NewBase(owner, IndexXY),
		storage: make(map[string]XYValue),
	}
	g.ensureInit()
	return g
}

// PositionsInPercent reports whether stored values are percentage based.
func (g *GearXY) PositionsInPercent() bool {
	return g != nil && g.positionsInPercent
}

// SetPositionsInPercent toggles percentage based storage.
func (g *GearXY) SetPositionsInPercent(enabled bool) {
	if g == nil {
		return
	}
	if g.positionsInPercent == enabled {
		return
	}
	g.positionsInPercent = enabled
	// Re-sync cached data to honour the new mode.
	g.UpdateState()
}

// SetController wires the controller reference and refreshes cached state.
func (g *GearXY) SetController(ctrl Controller) {
	if g == nil {
		return
	}
	g.Base.SetController(ctrl)
	g.ensureInit()
	g.UpdateState()
}

// Setup initialises the gear from serialized data.
func (g *GearXY) Setup(buffer *utils.ByteBuffer, resolver ControllerResolver) {
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

	if buffer.Version >= 2 {
		if buffer.ReadBool() {
			g.positionsInPercent = true
			for i := 0; i < count; i++ {
				page := buffer.ReadS()
				if page == nil {
					continue
				}
				g.readPercent(*page, buffer)
			}
			if buffer.ReadBool() {
				g.readPercent("", buffer)
			}
		}
	}
	g.Apply()
}

// UpdateState snapshots the owner's current position for the active page.
func (g *GearXY) UpdateState() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	val := XYValue{
		X: g.Owner().X(),
		Y: g.Owner().Y(),
	}
	g.updatePercent(&val)
	pageID := g.pageID()
	if pageID == "" {
		g.defaultValue = val
	} else {
		g.storage[pageID] = val
	}
}

// Apply restores the cached position for the active page.
func (g *GearXY) Apply() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	val := g.valueForPage(g.pageID())
	x := val.X
	y := val.Y
	if g.positionsInPercent {
		if pw, ph := g.Owner().ParentSize(); pw != 0 || ph != 0 {
			if pw != 0 {
				x = val.PX * pw
			} else {
				x = 0
			}
			if ph != 0 {
				y = val.PY * ph
			} else {
				y = 0
			}
		} else {
			x = 0
			y = 0
		}
	}
	cfg := g.TweenConfig()
	if cfg != nil && cfg.Tween && !DisableAllTweenEffect {
		if cfg.Tweener != nil {
			end := cfg.Tweener.EndValue()
			if end.X != x || end.Y != y {
				cfg.Tweener.Kill(false)
				cfg.Tweener = nil
			} else {
				return
			}
		}
		ox, oy := g.Owner().X(), g.Owner().Y()
		if ox == x && oy == y {
			return
		}
		tw := tween.To2(ox, oy, x, y, cfg.Duration)
		tw.SetDelay(cfg.Delay)
		tw.SetEase(cfg.EaseType)
		tw.SetTarget(g.Owner())
		tw.OnUpdate(func(tw *tween.GTweener) {
			vx, vy := tw.Value().XY()
			g.Owner().SetGearLocked(true)
			g.Owner().SetPosition(vx, vy)
			g.Owner().SetGearLocked(false)
		})
		tw.OnComplete(func(*tween.GTweener) {
			g.Owner().SetGearLocked(true)
			g.Owner().SetPosition(x, y)
			g.Owner().SetGearLocked(false)
			cfg.Tweener = nil
		})
		cfg.Tweener = tw
		return
	}
	if g.Owner().GearLocked() {
		g.Owner().SetPosition(x, y)
		return
	}
	g.Owner().SetGearLocked(true)
	g.Owner().SetPosition(x, y)
	g.Owner().SetGearLocked(false)
}

// UpdateFromRelations shifts cached positions when relations adjust the owner.
func (g *GearXY) UpdateFromRelations(dx, dy float64) {
	if g == nil || (dx == 0 && dy == 0) {
		return
	}
	g.ensureInit()
	if g.positionsInPercent {
		// Percent-based storage is recomputed from scratch on UpdateState below.
		return
	}
	for key, val := range g.storage {
		val.X += dx
		val.Y += dy
		g.updatePercent(&val)
		g.storage[key] = val
	}
	g.defaultValue.X += dx
	g.defaultValue.Y += dy
	g.updatePercent(&g.defaultValue)
	g.UpdateState()
}

// Value exposes the cached value for the supplied page (empty string for default).
func (g *GearXY) Value(pageID string) XYValue {
	if g == nil {
		return XYValue{}
	}
	g.ensureInit()
	return g.valueForPage(pageID)
}

func (g *GearXY) ensureInit() {
	if g == nil || g.Owner() == nil || g.initialized {
		return
	}
	g.initialized = true
	g.storage = make(map[string]XYValue)
	g.defaultValue = XYValue{
		X: g.Owner().X(),
		Y: g.Owner().Y(),
	}
	g.updatePercent(&g.defaultValue)
}

func (g *GearXY) pageID() string {
	if g == nil || g.Controller() == nil {
		return ""
	}
	return g.Controller().SelectedPageID()
}

func (g *GearXY) valueForPage(pageID string) XYValue {
	if g == nil {
		return XYValue{}
	}
	if pageID == "" {
		return g.defaultValue
	}
	if val, ok := g.storage[pageID]; ok {
		return val
	}
	return g.defaultValue
}

func (g *GearXY) updatePercent(val *XYValue) {
	if g == nil || val == nil || g.Owner() == nil {
		return
	}
	pw, ph := g.Owner().ParentSize()
	if pw != 0 {
		val.PX = val.X / pw
	} else {
		val.PX = 0
	}
	if ph != 0 {
		val.PY = val.Y / ph
	} else {
		val.PY = 0
	}
}

func (g *GearXY) readStatus(page string, buffer *utils.ByteBuffer) {
	if g == nil || buffer == nil {
		return
	}
	x := float64(buffer.ReadInt32())
	y := float64(buffer.ReadInt32())
	if page == "" {
		g.defaultValue.X = x
		g.defaultValue.Y = y
		g.updatePercent(&g.defaultValue)
	} else {
		val := g.storage[page]
		val.X = x
		val.Y = y
		g.updatePercent(&val)
		g.storage[page] = val
	}
}

func (g *GearXY) readPercent(page string, buffer *utils.ByteBuffer) {
	if g == nil || buffer == nil {
		return
	}
	px := float64(buffer.ReadFloat32())
	py := float64(buffer.ReadFloat32())
	if page == "" {
		g.defaultValue.PX = px
		g.defaultValue.PY = py
	} else {
		val := g.storage[page]
		val.PX = px
		val.PY = py
		g.storage[page] = val
	}
}
