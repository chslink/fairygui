package gears

import "github.com/chslink/fairygui/pkg/fgui/utils"

// GearIcon synchronises icon URLs across controller pages.
type GearIcon struct {
	Base

	storage      map[string]string
	defaultValue string
	initialized  bool
}

// NewGearIcon constructs a GearIcon bound to the provided owner.
func NewGearIcon(owner Owner) *GearIcon {
	g := &GearIcon{
		Base:    NewBase(owner, IndexIcon),
		storage: make(map[string]string),
	}
	g.ensureInit()
	return g
}

// SetController wires the controller reference.
// 参考 TypeScript 原版：GearBase.ts set controller()
// 只调用 init()，不调用 updateState()
func (g *GearIcon) SetController(ctrl Controller) {
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
func (g *GearIcon) Setup(buffer *utils.ByteBuffer, resolver ControllerResolver) {
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
		if icon := buffer.ReadS(); icon != nil {
			g.storage[*pagePtr] = *icon
		}
	}
	if buffer.ReadBool() {
		if icon := buffer.ReadS(); icon != nil {
			g.defaultValue = *icon
		}
	}
	if buffer.ReadBool() {
		_ = buffer.ReadByte()
		_ = buffer.ReadFloat32()
		_ = buffer.ReadFloat32()
	}
	g.Apply()
}

// Apply restores the cached icon for the active page.
func (g *GearIcon) Apply() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	value, ok := g.storage[g.pageID()]
	if !ok {
		value = g.defaultValue
	}
	g.Owner().SetProp(ObjectPropIDIcon, value)
}

// UpdateState snapshots the owner's icon for the active page.
func (g *GearIcon) UpdateState() {
	if g == nil || g.Owner() == nil {
		return
	}
	g.ensureInit()
	value, _ := g.Owner().GetProp(ObjectPropIDIcon).(string)
	page := g.pageID()
	if page == "" {
		g.defaultValue = value
	} else {
		g.storage[page] = value
	}
}

// UpdateFromRelations is unused for GearIcon but required by the interface.
func (g *GearIcon) UpdateFromRelations(dx, dy float64) {
	// Icon resource is independent of relations.
}

func (g *GearIcon) ensureInit() {
	if g == nil || g.initialized {
		return
	}
	g.initialized = true
	if g.storage == nil {
		g.storage = make(map[string]string)
	}
	if g.defaultValue == "" {
		if icon, ok := g.Owner().GetProp(ObjectPropIDIcon).(string); ok {
			g.defaultValue = icon
		}
	}
}

func (g *GearIcon) pageID() string {
	if g == nil || g.Controller() == nil {
		return ""
	}
	return g.Controller().SelectedPageID()
}
