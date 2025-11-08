package gears

import "github.com/chslink/fairygui/pkg/fgui/utils"

type animationValue struct {
	Playing   bool
	Frame     int
	TimeScale float64
	DeltaTime float64
}

// GearAnimation synchronises playing/frame properties across controller pages.
type GearAnimation struct {
	Base

	storage      map[string]animationValue
	defaultValue animationValue
	initialized  bool
}

// NewGearAnimation constructs a GearAnimation instance bound to the provided owner.
func NewGearAnimation(owner Owner) *GearAnimation {
	g := &GearAnimation{
		Base:    NewBase(owner, IndexAnimation),
		storage: make(map[string]animationValue),
	}
	g.ensureInit()
	return g
}

// SetController wires the controller reference.
// 参考 TypeScript 原版：GearBase.ts set controller()
// 只调用 init()，不调用 updateState()
func (g *GearAnimation) SetController(ctrl Controller) {
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
func (g *GearAnimation) Setup(buffer *utils.ByteBuffer, resolver ControllerResolver) {
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
		pagePtr := buffer.ReadS()
		if pagePtr == nil {
			continue
		}
		g.readStatus(*pagePtr, buffer)
	}
	if buffer.ReadBool() {
		g.readStatus("", buffer)
	}
	if buffer.ReadBool() {
		_ = buffer.ReadByte()
		_ = buffer.ReadFloat32()
		_ = buffer.ReadFloat32()
	}
	g.Apply()
}

// Apply restores the cached playing/frame state for the active page.
func (g *GearAnimation) Apply() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	val := g.valueForPage(g.pageID())
	g.Owner().SetProp(ObjectPropIDPlaying, val.Playing)
	g.Owner().SetProp(ObjectPropIDFrame, val.Frame)
	if val.TimeScale == 0 {
		g.Owner().SetProp(ObjectPropIDTimeScale, 0.0)
	} else {
		g.Owner().SetProp(ObjectPropIDTimeScale, val.TimeScale)
	}
	g.Owner().SetProp(ObjectPropIDDeltaTime, val.DeltaTime)
}

// UpdateState snapshots the owner's animation state for the active page.
func (g *GearAnimation) UpdateState() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	val := animationValue{TimeScale: 1}
	if playing, ok := g.Owner().GetProp(ObjectPropIDPlaying).(bool); ok {
		val.Playing = playing
	}
	switch frame := g.Owner().GetProp(ObjectPropIDFrame).(type) {
	case int:
		val.Frame = frame
	case int32:
		val.Frame = int(frame)
	case int64:
		val.Frame = int(frame)
	}
	val.TimeScale = floatFromProp(g.Owner().GetProp(ObjectPropIDTimeScale), 1)
	val.DeltaTime = floatFromProp(g.Owner().GetProp(ObjectPropIDDeltaTime), 0)
	page := g.pageID()
	if page == "" {
		g.defaultValue = val
	} else {
		g.storage[page] = val
	}
}

// UpdateFromRelations is unused for GearAnimation but required by the interface.
func (g *GearAnimation) UpdateFromRelations(dx, dy float64) {
	// Animation state is independent of relations.
}

func (g *GearAnimation) ensureInit() {
	if g == nil || g.Owner() == nil || g.initialized {
		return
	}
	g.initialized = true
	if g.storage == nil {
		g.storage = make(map[string]animationValue)
	}
	g.defaultValue = animationValue{TimeScale: 1}
	if playing, ok := g.Owner().GetProp(ObjectPropIDPlaying).(bool); ok {
		g.defaultValue.Playing = playing
	}
	switch frame := g.Owner().GetProp(ObjectPropIDFrame).(type) {
	case int:
		g.defaultValue.Frame = frame
	case int32:
		g.defaultValue.Frame = int(frame)
	case int64:
		g.defaultValue.Frame = int(frame)
	}
	g.defaultValue.TimeScale = floatFromProp(g.Owner().GetProp(ObjectPropIDTimeScale), 1)
	g.defaultValue.DeltaTime = floatFromProp(g.Owner().GetProp(ObjectPropIDDeltaTime), 0)
}

func (g *GearAnimation) pageID() string {
	if g == nil || g.Controller() == nil {
		return ""
	}
	return g.Controller().SelectedPageID()
}

func (g *GearAnimation) valueForPage(page string) animationValue {
	if g == nil {
		return animationValue{}
	}
	if page == "" {
		return g.defaultValue
	}
	if val, ok := g.storage[page]; ok {
		return val
	}
	return g.defaultValue
}

func (g *GearAnimation) readStatus(page string, buffer *utils.ByteBuffer) {
	if g == nil || buffer == nil {
		return
	}
	val := animationValue{
		Playing:   buffer.ReadBool(),
		Frame:     int(buffer.ReadInt32()),
		TimeScale: 1,
	}
	if page == "" {
		g.defaultValue = val
	} else {
		g.storage[page] = val
	}
}

func floatFromProp(v any, fallback float64) float64 {
	switch value := v.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case uint:
		return float64(value)
	case uint32:
		return float64(value)
	case uint64:
		return float64(value)
	case nil:
		return fallback
	default:
		return fallback
	}
}
