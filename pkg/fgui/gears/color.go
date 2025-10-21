package gears

import "github.com/chslink/fairygui/pkg/fgui/utils"

type colorValue struct {
	Color   string
	Outline string
}

// GearColor synchronises color/outline colour across controller pages.
type GearColor struct {
	Base

	storage      map[string]colorValue
	defaultValue colorValue
	initialized  bool
}

// NewGearColor constructs a GearColor instance bound to the provided owner.
func NewGearColor(owner Owner) *GearColor {
	g := &GearColor{
		Base:    NewBase(owner, IndexColor),
		storage: make(map[string]colorValue),
	}
	g.ensureInit()
	return g
}

// SetController wires the controller reference and refreshes cached state.
func (g *GearColor) SetController(ctrl Controller) {
	if g == nil {
		return
	}
	g.Base.SetController(ctrl)
	g.ensureInit()
	g.UpdateState()
}

// Setup initialises the gear from serialized data.
func (g *GearColor) Setup(buffer *utils.ByteBuffer, resolver ControllerResolver) {
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

// Apply restores the cached colour information for the active page.
func (g *GearColor) Apply() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	val := g.valueForPage(g.pageID())
	g.Owner().SetProp(ObjectPropIDColor, val.Color)
	g.Owner().SetProp(ObjectPropIDOutlineColor, val.Outline)
}

// UpdateState snapshots the owner's current colour state for the active page.
func (g *GearColor) UpdateState() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	val := colorValue{}
	if c, ok := g.Owner().GetProp(ObjectPropIDColor).(string); ok {
		val.Color = c
	}
	if oc, ok := g.Owner().GetProp(ObjectPropIDOutlineColor).(string); ok {
		val.Outline = oc
	}
	page := g.pageID()
	if page == "" {
		g.defaultValue = val
	} else {
		g.storage[page] = val
	}
}

// UpdateFromRelations is unused for GearColor but required by the interface.
func (g *GearColor) UpdateFromRelations(dx, dy float64) {
	// Color properties are independent of relations.
}

func (g *GearColor) ensureInit() {
	if g == nil || g.Owner() == nil || g.initialized {
		return
	}
	g.initialized = true
	if g.storage == nil {
		g.storage = make(map[string]colorValue)
	}
	g.defaultValue = colorValue{}
	if c, ok := g.Owner().GetProp(ObjectPropIDColor).(string); ok {
		g.defaultValue.Color = c
	}
	if oc, ok := g.Owner().GetProp(ObjectPropIDOutlineColor).(string); ok {
		g.defaultValue.Outline = oc
	}
}

func (g *GearColor) pageID() string {
	if g == nil || g.Controller() == nil {
		return ""
	}
	return g.Controller().SelectedPageID()
}

func (g *GearColor) valueForPage(page string) colorValue {
	if g == nil {
		return colorValue{}
	}
	if page == "" {
		return g.defaultValue
	}
	if val, ok := g.storage[page]; ok {
		return val
	}
	return g.defaultValue
}

func (g *GearColor) readStatus(page string, buffer *utils.ByteBuffer) {
	if g == nil || buffer == nil {
		return
	}
	val := colorValue{
		Color:   buffer.ReadColorString(true),
		Outline: buffer.ReadColorString(true),
	}
	if page == "" {
		g.defaultValue = val
	} else {
		g.storage[page] = val
	}
}
