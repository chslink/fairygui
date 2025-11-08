package gears

import "github.com/chslink/fairygui/pkg/fgui/utils"

// GearText synchronises textual content across controller pages.
type GearText struct {
	Base

	storage      map[string]string
	defaultValue string
	initialized  bool
}

// NewGearText constructs a GearText bound to the provided owner.
func NewGearText(owner Owner) *GearText {
	g := &GearText{
		Base:    NewBase(owner, IndexText),
		storage: make(map[string]string),
	}
	g.ensureInit()
	return g
}

// SetController wires the controller reference.
// 参考 TypeScript 原版：GearBase.ts set controller()
// 只调用 init()，不调用 updateState()
func (g *GearText) SetController(ctrl Controller) {
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
func (g *GearText) Setup(buffer *utils.ByteBuffer, resolver ControllerResolver) {
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
		if text := buffer.ReadS(); text != nil {
			g.storage[*pagePtr] = *text
		}
	}
	if buffer.ReadBool() {
		if text := buffer.ReadS(); text != nil {
			g.defaultValue = *text
		}
	}
	if buffer.ReadBool() {
		_ = buffer.ReadByte()
		_ = buffer.ReadFloat32()
		_ = buffer.ReadFloat32()
	}
	g.Apply()
}

// Apply restores the cached text for the active page.
func (g *GearText) Apply() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	value, ok := g.storage[g.pageID()]
	if !ok {
		value = g.defaultValue
	}
	g.Owner().SetProp(ObjectPropIDText, value)
}

// UpdateState snapshots the owner's text for the active page.
func (g *GearText) UpdateState() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	value, _ := g.Owner().GetProp(ObjectPropIDText).(string)
	page := g.pageID()
	if page == "" {
		g.defaultValue = value
	} else {
		g.storage[page] = value
	}
}

// UpdateFromRelations is unused for GearText but required by the interface.
func (g *GearText) UpdateFromRelations(dx, dy float64) {
	// Text content is independent of relations.
}

func (g *GearText) ensureInit() {
	if g == nil || g.initialized {
		return
	}
	g.initialized = true
	if g.storage == nil {
		g.storage = make(map[string]string)
	}
	if g.defaultValue == "" {
		if txt, ok := g.Owner().GetProp(ObjectPropIDText).(string); ok {
			g.defaultValue = txt
		}
	}
}

func (g *GearText) pageID() string {
	if g == nil || g.Controller() == nil {
		return ""
	}
	return g.Controller().SelectedPageID()
}
