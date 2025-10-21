package gears

import "github.com/chslink/fairygui/pkg/fgui/utils"

// GearFontSize synchronises font size across controller pages.
type GearFontSize struct {
	Base

	storage      map[string]int
	defaultValue int
	initialized  bool
}

// NewGearFontSize constructs a GearFontSize for the provided owner.
func NewGearFontSize(owner Owner) *GearFontSize {
	g := &GearFontSize{
		Base:    NewBase(owner, IndexFontSize),
		storage: make(map[string]int),
	}
	g.ensureInit()
	return g
}

// SetController wires the controller reference and refreshes cached state.
func (g *GearFontSize) SetController(ctrl Controller) {
	if g == nil {
		return
	}
	g.Base.SetController(ctrl)
	g.ensureInit()
	g.UpdateState()
}

// Setup initialises the gear from serialized data.
func (g *GearFontSize) Setup(buffer *utils.ByteBuffer, resolver ControllerResolver) {
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
		g.storage[*pagePtr] = int(buffer.ReadInt32())
	}
	if buffer.ReadBool() {
		g.defaultValue = int(buffer.ReadInt32())
	}
	if buffer.ReadBool() {
		_ = buffer.ReadByte()
		_ = buffer.ReadFloat32()
		_ = buffer.ReadFloat32()
	}
	g.Apply()
}

// Apply restores the cached font size for the active page.
func (g *GearFontSize) Apply() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	value, ok := g.storage[g.pageID()]
	if !ok {
		value = g.defaultValue
	}
	g.Owner().SetProp(ObjectPropIDFontSize, value)
}

// UpdateState snapshots the owner's font size for the active page.
func (g *GearFontSize) UpdateState() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	if size, ok := toInt(g.Owner().GetProp(ObjectPropIDFontSize)); ok {
		page := g.pageID()
		if page == "" {
			g.defaultValue = size
		} else {
			g.storage[page] = size
		}
	}
}

// UpdateFromRelations is unused for GearFontSize but required by the interface.
func (g *GearFontSize) UpdateFromRelations(dx, dy float64) {
}

func (g *GearFontSize) ensureInit() {
	if g == nil || g.Owner() == nil || g.initialized {
		return
	}
	g.initialized = true
	if g.storage == nil {
		g.storage = make(map[string]int)
	}
	if size, ok := toInt(g.Owner().GetProp(ObjectPropIDFontSize)); ok {
		g.defaultValue = size
	}
}

func (g *GearFontSize) pageID() string {
	if g == nil || g.Controller() == nil {
		return ""
	}
	return g.Controller().SelectedPageID()
}

func toInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}
