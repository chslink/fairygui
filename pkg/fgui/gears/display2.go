package gears

import "github.com/chslink/fairygui/pkg/fgui/utils"

// GearDisplay2 mirrors FairyGUI's GearDisplay2 behaviour, combining visibility conditions.
type GearDisplay2 struct {
	Base

	pages          []string
	condition      int8
	visibleCounter int
}

// NewGearDisplay2 constructs a GearDisplay2 for the provided owner.
func NewGearDisplay2(owner Owner) *GearDisplay2 {
	return &GearDisplay2{
		Base: NewBase(owner, IndexDisplay2),
	}
}

// SetController assigns the driving controller reference.
func (g *GearDisplay2) SetController(ctrl Controller) {
	if g == nil {
		return
	}
	g.Base.SetController(ctrl)
	g.Apply()
}

// Setup initialises the gear from serialized data.
func (g *GearDisplay2) Setup(buffer *utils.ByteBuffer, resolver ControllerResolver) {
	if g == nil || buffer == nil {
		return
	}
	ctrlIndex := int(buffer.ReadInt16())
	var ctrl Controller
	if resolver != nil && ctrlIndex >= 0 {
		ctrl = resolver.ControllerAt(ctrlIndex)
	}
	g.SetController(ctrl)

	count := int(buffer.ReadInt16())
	g.pages = g.pages[:0]
	for i := 0; i < count; i++ {
		page := buffer.ReadS()
		if page != nil {
			g.pages = append(g.pages, *page)
		}
	}
	// No default status block for GearDisplay2; skip tween flag.
	if buffer.ReadBool() {
		_ = buffer.ReadByte()
		_ = buffer.ReadFloat32()
		_ = buffer.ReadFloat32()
	}
	if buffer.Version >= 2 {
		g.condition = int8(buffer.ReadByte())
	}
	g.Apply()
}

// Apply evaluates controller visibility for the current page.
func (g *GearDisplay2) Apply() {
	if g == nil {
		return
	}
	g.visibleCounter = 0
	if g.Controller() == nil {
		return
	}
	if len(g.pages) == 0 {
		g.visibleCounter = 1
		return
	}
	pageID := g.Controller().SelectedPageID()
	for _, candidate := range g.pages {
		if candidate == pageID {
			g.visibleCounter = 1
			return
		}
	}
}

// Evaluate combines this gear's visibility with the upstream connected flag.
func (g *GearDisplay2) Evaluate(connected bool) bool {
	if g == nil {
		return connected
	}
	visible := g.Controller() == nil || g.visibleCounter > 0
	if g.condition == 0 {
		return visible && connected
	}
	return visible || connected
}

// UpdateState is unused for GearDisplay2.
func (g *GearDisplay2) UpdateState() {
}

// UpdateFromRelations is unused for GearDisplay2.
func (g *GearDisplay2) UpdateFromRelations(dx, dy float64) {
}
