package gears

import (
	"github.com/chslink/fairygui/pkg/fgui/tween"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// SizeValue captures cached dimension/scale state.
type SizeValue struct {
	Width  float64
	Height float64
	ScaleX float64
	ScaleY float64
}

// GearSize synchronises size/scale across controller pages.
type GearSize struct {
	Base

	storage      map[string]SizeValue
	defaultValue SizeValue
	initialized  bool
}

// NewGearSize constructs a GearSize bound to the provided owner.
func NewGearSize(owner Owner) *GearSize {
	g := &GearSize{
		Base:    NewBase(owner, IndexSize),
		storage: make(map[string]SizeValue),
	}
	g.ensureInit()
	return g
}

// SetController wires the controller reference.
// 参考 TypeScript 原版：GearBase.ts set controller()
// 只调用 init()，不调用 updateState()
func (g *GearSize) SetController(ctrl Controller) {
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
func (g *GearSize) Setup(buffer *utils.ByteBuffer, resolver ControllerResolver) {
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

// UpdateState snapshots the owner's current size for the active page.
func (g *GearSize) UpdateState() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	sx, sy := g.Owner().Scale()
	val := SizeValue{
		Width:  g.Owner().Width(),
		Height: g.Owner().Height(),
		ScaleX: sx,
		ScaleY: sy,
	}
	pageID := g.pageID()
	if pageID == "" {
		g.defaultValue = val
	} else {
		g.storage[pageID] = val
	}
}

// Apply restores the cached size/scale for the active page.
func (g *GearSize) Apply() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	val := g.valueForPage(g.pageID())
	owner := g.Owner()
	cfg := g.TweenConfig()
	if cfg != nil && cfg.Tween && !DisableAllTweenEffect {
		if cfg.Tweener != nil {
			end := cfg.Tweener.EndValue()
			if end.X != val.Width || end.Y != val.Height || end.Z != val.ScaleX || end.W != val.ScaleY {
				cfg.Tweener.Kill(false)
				cfg.Tweener = nil
			} else {
				return
			}
		}
		sx, sy := owner.Scale()
		current := tween.Value{X: owner.Width(), Y: owner.Height(), Z: sx, W: sy}
		tw := tween.To4(current, tween.Value{X: val.Width, Y: val.Height, Z: val.ScaleX, W: val.ScaleY}, cfg.Duration)
		tw.SetDelay(cfg.Delay)
		tw.SetEase(cfg.EaseType)
		tw.SetTarget(owner)
		tw.OnUpdate(func(tw *tween.GTweener) {
			v := tw.Value()
			owner.SetGearLocked(true)
			owner.SetSize(v.X, v.Y)
			owner.SetScale(v.Z, v.W)
			owner.SetGearLocked(false)
		})
		tw.OnComplete(func(*tween.GTweener) {
			owner.SetGearLocked(true)
			owner.SetSize(val.Width, val.Height)
			owner.SetScale(val.ScaleX, val.ScaleY)
			owner.SetGearLocked(false)
			cfg.Tweener = nil
		})
		cfg.Tweener = tw
		return
	}
	if owner.GearLocked() {
		owner.SetSize(val.Width, val.Height)
		owner.SetScale(val.ScaleX, val.ScaleY)
		return
	}
	owner.SetGearLocked(true)
	owner.SetSize(val.Width, val.Height)
	owner.SetScale(val.ScaleX, val.ScaleY)
	owner.SetGearLocked(false)
}

// UpdateFromRelations shifts cached sizes when relations adjust the owner.
func (g *GearSize) UpdateFromRelations(dw, dh float64) {
	if g == nil || (dw == 0 && dh == 0) {
		return
	}
	g.ensureInit()
	for key, val := range g.storage {
		val.Width += dw
		val.Height += dh
		g.storage[key] = val
	}
	g.defaultValue.Width += dw
	g.defaultValue.Height += dh
	g.UpdateState()
}

// Value exposes the cached value for the supplied page (empty string for default).
func (g *GearSize) Value(pageID string) SizeValue {
	if g == nil {
		return SizeValue{}
	}
	g.ensureInit()
	return g.valueForPage(pageID)
}

func (g *GearSize) ensureInit() {
	if g == nil || g.Owner() == nil || g.initialized {
		return
	}
	g.initialized = true
	g.storage = make(map[string]SizeValue)
	sx, sy := g.Owner().Scale()
	g.defaultValue = SizeValue{
		Width:  g.Owner().Width(),
		Height: g.Owner().Height(),
		ScaleX: sx,
		ScaleY: sy,
	}
}

func (g *GearSize) pageID() string {
	if g == nil || g.Controller() == nil {
		return ""
	}
	return g.Controller().SelectedPageID()
}

func (g *GearSize) valueForPage(pageID string) SizeValue {
	if g == nil {
		return SizeValue{}
	}
	if pageID == "" {
		return g.defaultValue
	}
	if val, ok := g.storage[pageID]; ok {
		return val
	}
	return g.defaultValue
}

func (g *GearSize) readStatus(page string, buffer *utils.ByteBuffer) {
	if g == nil || buffer == nil {
		return
	}
	val := SizeValue{
		Width:  float64(buffer.ReadInt32()),
		Height: float64(buffer.ReadInt32()),
		ScaleX: float64(buffer.ReadFloat32()),
		ScaleY: float64(buffer.ReadFloat32()),
	}
	if page == "" {
		g.defaultValue = val
	} else {
		g.storage[page] = val
	}
}
