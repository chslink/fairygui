package core

// RelationType mirrors FairyGUI's RelationType enum.
type RelationType int

const (
	RelationTypeLeft_Left RelationType = iota
	RelationTypeLeft_Center
	RelationTypeLeft_Right
	RelationTypeCenter_Center
	RelationTypeRight_Left
	RelationTypeRight_Center
	RelationTypeRight_Right
	RelationTypeTop_Top
	RelationTypeTop_Middle
	RelationTypeTop_Bottom
	RelationTypeMiddle_Middle
	RelationTypeBottom_Top
	RelationTypeBottom_Middle
	RelationTypeBottom_Bottom
	RelationTypeWidth
	RelationTypeHeight
	RelationTypeLeftExt_Left
	RelationTypeLeftExt_Right
	RelationTypeRightExt_Left
	RelationTypeRightExt_Right
	RelationTypeTopExt_Top
	RelationTypeTopExt_Bottom
	RelationTypeBottomExt_Top
	RelationTypeBottomExt_Bottom
	RelationTypeSize
)

// Relations tracks how a GObject reacts to other objects' size/position changes.
type Relations struct {
	owner     *GObject
	items     []*RelationItem
	handling  *GObject
	sizeDirty bool
}

// NewRelations constructs an empty relation container for the specified owner.
func NewRelations(owner *GObject) *Relations {
	return &Relations{
		owner: owner,
		items: make([]*RelationItem, 0),
	}
}

// Add registers a relation between the owner and target.
func (r *Relations) Add(target *GObject, relation RelationType, usePercent bool) {
	if r == nil || target == nil || r.owner == target {
		return
	}
	item := r.ensureItem(target)
	item.Add(relation, usePercent)
}

// Remove removes a relation from the owner to target. When relation is RelationTypeSize
// it will remove both width/height constraints, matching FairyGUI semantics.
func (r *Relations) Remove(target *GObject, relation RelationType) {
	if r == nil || target == nil {
		return
	}
	for i := 0; i < len(r.items); {
		item := r.items[i]
		if item.Target() != target {
			i++
			continue
		}
		item.Remove(relation)
		if item.IsEmpty() {
			item.detach()
			r.items = append(r.items[:i], r.items[i+1:]...)
		} else {
			i++
		}
		return
	}
}

// ClearFor removes all relations that target the specified object.
func (r *Relations) ClearFor(target *GObject) {
	if r == nil || target == nil {
		return
	}
	for i := 0; i < len(r.items); {
		item := r.items[i]
		if item.Target() == target {
			item.detach()
			r.items = append(r.items[:i], r.items[i+1:]...)
		} else {
			i++
		}
	}
}

// ClearAll removes every relation entry stored on the owner.
func (r *Relations) ClearAll() {
	if r == nil {
		return
	}
	for _, item := range r.items {
		item.detach()
	}
	r.items = r.items[:0]
}

// Contains reports whether the owner references the target through a relation.
func (r *Relations) Contains(target *GObject) bool {
	if r == nil || target == nil {
		return false
	}
	for _, item := range r.items {
		if item.Target() == target {
			return true
		}
	}
	return false
}

// Items exposes a copy of the current relation items.
func (r *Relations) Items() []*RelationItem {
	if r == nil {
		return nil
	}
	out := make([]*RelationItem, len(r.items))
	copy(out, r.items)
	return out
}

// OnOwnerSizeChanged propagates owner size deltas to relation constraints.
func (r *Relations) OnOwnerSizeChanged(dWidth, dHeight float64, applyPivot bool) {
	if r == nil || len(r.items) == 0 {
		return
	}
	if dWidth == 0 && dHeight == 0 {
		return
	}
	for _, item := range r.items {
		item.applyOnSelfResized(dWidth, dHeight, applyPivot)
	}
}

// EnsureRelationsSizeCorrect ensures all relation targets have correct sizes.
// 对应 TypeScript: public ensureRelationsSizeCorrect(): void (Relations.ts:112-122)
func (r *Relations) EnsureRelationsSizeCorrect() {
	if r == nil || len(r.items) == 0 {
		return
	}

	r.sizeDirty = false
	for _, item := range r.items {
		if item != nil && item.target != nil {
			item.target.EnsureSizeCorrect()
		}
	}
}

func (r *Relations) ensureItem(target *GObject) *RelationItem {
	for _, item := range r.items {
		if item.Target() == target {
			return item
		}
	}
	item := newRelationItem(r.owner, target)
	item.attach()
	r.items = append(r.items, item)
	return item
}

// RelationItem represents the set of constraints between owner and target object.
type RelationItem struct {
	owner            *GObject
	target           *GObject
	defs             []RelationDef
	targetX          float64
	targetY          float64
	targetWidth      float64
	targetHeight     float64
	targetInitX      float64
	targetInitY      float64
	targetInitWidth  float64
	targetInitHeight float64
	attached         bool
}

// RelationDef describes a single relation constraint.
type RelationDef struct {
	Type    RelationType
	Percent bool
	Axis    int
}

func newRelationItem(owner, target *GObject) *RelationItem {
	return &RelationItem{
		owner:  owner,
		target: target,
		defs:   make([]RelationDef, 0),
	}
}

// Target returns the relation target.
func (r *RelationItem) Target() *GObject {
	if r == nil {
		return nil
	}
	return r.target
}

// Add records the given relation definition.
func (r *RelationItem) Add(relation RelationType, usePercent bool) {
	if r == nil {
		return
	}
	if relation == RelationTypeSize {
		r.Add(RelationTypeWidth, usePercent)
		r.Add(RelationTypeHeight, usePercent)
		return
	}
	for _, def := range r.defs {
		if def.Type == relation {
			return
		}
	}
	r.defs = append(r.defs, RelationDef{
		Type:    relation,
		Percent: usePercent,
		Axis:    relationAxis(relation),
	})
}

// Remove clears a relation definition; if RelationTypeSize is provided both width
// and height relations will be removed.
func (r *RelationItem) Remove(relation RelationType) {
	if r == nil {
		return
	}
	if relation == RelationTypeSize {
		r.Remove(RelationTypeWidth)
		r.Remove(RelationTypeHeight)
		return
	}
	for i := 0; i < len(r.defs); i++ {
		if r.defs[i].Type == relation {
			r.defs = append(r.defs[:i], r.defs[i+1:]...)
			return
		}
	}
}

// IsEmpty reports whether no relations remain.
func (r *RelationItem) IsEmpty() bool {
	return r == nil || len(r.defs) == 0
}

func (r *RelationItem) attach() {
	if r == nil || r.attached || r.target == nil {
		return
	}
	r.target.addRelationDependent(r)
	r.targetX = r.target.X()
	r.targetY = r.target.Y()
	r.targetWidth = r.target.Width()
	r.targetHeight = r.target.Height()
	r.targetInitX = r.target.X()
	r.targetInitY = r.target.Y()
	r.targetInitWidth = r.target.InitWidth()
	if r.targetInitWidth == 0 {
		r.targetInitWidth = r.targetWidth
	}
	r.targetInitHeight = r.target.InitHeight()
	if r.targetInitHeight == 0 {
		r.targetInitHeight = r.targetHeight
	}
	r.attached = true
}

func (r *RelationItem) detach() {
	if r == nil || !r.attached {
		return
	}
	if r.target != nil {
		r.target.removeRelationDependent(r)
	}
	r.attached = false
}

func (r *RelationItem) onTargetXYChanged(dx, dy float64) {
	if r == nil || r.IsEmpty() || (dx == 0 && dy == 0) {
		return
	}
	if rel := r.owner.Relations(); rel != nil {
		rel.handling = r.target
		defer func() { rel.handling = nil }()
	}
	for _, def := range r.defs {
		r.applyOnXYChanged(def, dx, dy)
	}
	r.targetX += dx
	r.targetY += dy
}

func (r *RelationItem) onTargetSizeChanged(dw, dh float64) {
	if r == nil || r.IsEmpty() || (dw == 0 && dh == 0) {
		return
	}
	if rel := r.owner.Relations(); rel != nil {
		rel.handling = r.target
		defer func() { rel.handling = nil }()
	}
	oldX := r.owner.X()
	oldY := r.owner.Y()
	oldW := r.owner.Width()
	oldH := r.owner.Height()

	for _, def := range r.defs {
		r.applyOnSizeChanged(def)
	}

	r.targetWidth = r.target.Width()
	r.targetHeight = r.target.Height()

	if oldX != r.owner.X() || oldY != r.owner.Y() {
		r.owner.updateGearFromRelationsSafe(1, r.owner.X()-oldX, r.owner.Y()-oldY)
	}
	if oldW != r.owner.Width() || oldH != r.owner.Height() {
		r.owner.updateGearFromRelationsSafe(2, r.owner.Width()-oldW, r.owner.Height()-oldH)
	}
}

func relationAxis(t RelationType) int {
	switch t {
	case RelationTypeLeft_Left,
		RelationTypeLeft_Center,
		RelationTypeLeft_Right,
		RelationTypeCenter_Center,
		RelationTypeRight_Left,
		RelationTypeRight_Center,
		RelationTypeRight_Right,
		RelationTypeWidth,
		RelationTypeLeftExt_Left,
		RelationTypeLeftExt_Right,
		RelationTypeRightExt_Left,
		RelationTypeRightExt_Right:
		return 0
	default:
		return 1
	}
}

func (r *RelationItem) applyOnSizeChanged(def RelationDef) {
	if r == nil || r.target == nil {
		return
	}
	owner := r.owner
	target := r.target

	ownerParentObj := (*GObject)(nil)
	if parent := owner.Parent(); parent != nil {
		ownerParentObj = parent.GObject
	}
	targetParentObj := (*GObject)(nil)
	if parent := target.Parent(); parent != nil {
		targetParentObj = parent.GObject
	}
	ownerIsTargetParent := targetParentObj == owner
	targetIsOwnerParent := ownerParentObj == target

	ownerXMin := owner.xMin()
	ownerYMin := owner.yMin()
	ownerRawWidth := owner.RawWidth()
	if ownerRawWidth == 0 {
		ownerRawWidth = owner.Width()
	}
	ownerRawHeight := owner.RawHeight()
	if ownerRawHeight == 0 {
		ownerRawHeight = owner.Height()
	}
	targetWidth := target.Width()
	targetHeight := target.Height()

	targetPivotX, targetPivotY := 0.0, 0.0
	if target.PivotAsAnchor() {
		targetPivotX, targetPivotY = target.Pivot()
	}

	var pos, pivot, delta float64
	if def.Axis == 0 {
		if !targetIsOwnerParent {
			pos = target.X()
		}
		pivot = targetPivotX
		if def.Percent {
			if r.targetWidth != 0 {
				delta = targetWidth / r.targetWidth
			} else {
				delta = 1
			}
		} else {
			delta = targetWidth - r.targetWidth
		}
	} else {
		if !targetIsOwnerParent {
			pos = target.Y()
		}
		pivot = targetPivotY
		if def.Percent {
			if r.targetHeight != 0 {
				delta = targetHeight / r.targetHeight
			} else {
				delta = 1
			}
		} else {
			delta = targetHeight - r.targetHeight
		}
	}

	sourceWidth := r.resolveSourceWidth(owner)
	sourceHeight := r.resolveSourceHeight(owner)
	targetInitWidth := r.targetInitWidth
	if targetInitWidth == 0 {
		targetInitWidth = r.targetWidth
	}
	targetInitHeight := r.targetInitHeight
	if targetInitHeight == 0 {
		targetInitHeight = r.targetHeight
	}

	switch def.Type {
	case RelationTypeLeft_Left:
		if def.Percent {
			owner.setXMin(pos + (ownerXMin-pos)*delta)
		} else if pivot != 0 {
			owner.SetPosition(owner.X()+delta*(-pivot), owner.Y())
		}
	case RelationTypeLeft_Center:
		if def.Percent {
			owner.setXMin(pos + (ownerXMin-pos)*delta)
		} else {
			owner.SetPosition(owner.X()+delta*(0.5-pivot), owner.Y())
		}
	case RelationTypeLeft_Right:
		if def.Percent {
			owner.setXMin(pos + (ownerXMin-pos)*delta)
		} else {
			owner.SetPosition(owner.X()+delta*(1-pivot), owner.Y())
		}
	case RelationTypeCenter_Center:
		if def.Percent {
			owner.setXMin(pos + (ownerXMin+ownerRawWidth*0.5-pos)*delta - ownerRawWidth*0.5)
		} else {
			owner.SetPosition(owner.X()+delta*(0.5-pivot), owner.Y())
		}
	case RelationTypeRight_Left:
		if def.Percent {
			owner.setXMin(pos + (ownerXMin+ownerRawWidth-pos)*delta - ownerRawWidth)
		} else if pivot != 0 {
			owner.SetPosition(owner.X()+delta*(-pivot), owner.Y())
		}
	case RelationTypeRight_Center:
		if def.Percent {
			owner.setXMin(pos + (ownerXMin+ownerRawWidth-pos)*delta - ownerRawWidth)
		} else {
			owner.SetPosition(owner.X()+delta*(0.5-pivot), owner.Y())
		}
	case RelationTypeRight_Right:
		if def.Percent {
			owner.setXMin(pos + (ownerXMin+ownerRawWidth-pos)*delta - ownerRawWidth)
		} else {
			owner.SetPosition(owner.X()+delta*(1-pivot), owner.Y())
		}
	case RelationTypeTop_Top:
		if def.Percent {
			owner.setYMin(pos + (ownerYMin-pos)*delta)
		} else if pivot != 0 {
			owner.SetPosition(owner.X(), owner.Y()+delta*(-pivot))
		}
	case RelationTypeTop_Middle:
		if def.Percent {
			owner.setYMin(pos + (ownerYMin-pos)*delta)
		} else {
			owner.SetPosition(owner.X(), owner.Y()+delta*(0.5-pivot))
		}
	case RelationTypeTop_Bottom:
		if def.Percent {
			owner.setYMin(pos + (ownerYMin-pos)*delta)
		} else {
			owner.SetPosition(owner.X(), owner.Y()+delta*(1-pivot))
		}
	case RelationTypeMiddle_Middle:
		if def.Percent {
			owner.setYMin(pos + (ownerYMin+ownerRawHeight*0.5-pos)*delta - ownerRawHeight*0.5)
		} else {
			owner.SetPosition(owner.X(), owner.Y()+delta*(0.5-pivot))
		}
	case RelationTypeBottom_Top:
		if def.Percent {
			owner.setYMin(pos + (ownerYMin+ownerRawHeight-pos)*delta - ownerRawHeight)
		} else if pivot != 0 {
			owner.SetPosition(owner.X(), owner.Y()+delta*(-pivot))
		}
	case RelationTypeBottom_Middle:
		if def.Percent {
			owner.setYMin(pos + (ownerYMin+ownerRawHeight-pos)*delta - ownerRawHeight)
		} else {
			owner.SetPosition(owner.X(), owner.Y()+delta*(0.5-pivot))
		}
	case RelationTypeBottom_Bottom:
		if def.Percent {
			owner.setYMin(pos + (ownerYMin+ownerRawHeight-pos)*delta - ownerRawHeight)
		} else {
			owner.SetPosition(owner.X(), owner.Y()+delta*(1-pivot))
		}
	case RelationTypeWidth:
		if def.Percent {
			if ownerIsTargetParent {
				width := pos + targetWidth - targetWidth*pivot + (sourceWidth-r.targetInitX-targetInitWidth+targetInitWidth*pivot)*delta
				tmp := ownerXMin
				owner.SetSize(width, ownerRawHeight)
				if owner.PivotAsAnchor() {
					owner.setXMin(tmp)
				}
			} else {
				v := pos + (ownerXMin+ownerRawWidth-pos)*delta - (ownerXMin + ownerRawWidth)
				owner.SetSize(ownerRawWidth+v, ownerRawHeight)
				owner.setXMin(ownerXMin)
			}
		} else {
			if ownerIsTargetParent {
				width := sourceWidth + pos - r.targetInitX + (targetWidth-targetInitWidth)*(1-pivot)
				tmp := ownerXMin
				owner.SetSize(width, ownerRawHeight)
				if owner.PivotAsAnchor() {
					owner.setXMin(tmp)
				}
			} else {
				v := delta * (1 - pivot)
				owner.SetSize(ownerRawWidth+v, ownerRawHeight)
				owner.setXMin(ownerXMin)
			}
		}
	case RelationTypeHeight:
		if def.Percent {
			if ownerIsTargetParent {
				height := pos + targetHeight - targetHeight*pivot + (sourceHeight-r.targetInitY-targetInitHeight+targetInitHeight*pivot)*delta
				tmp := ownerYMin
				owner.SetSize(ownerRawWidth, height)
				if owner.PivotAsAnchor() {
					owner.setYMin(tmp)
				}
			} else {
				v := pos + (ownerYMin+ownerRawHeight-pos)*delta - (ownerYMin + ownerRawHeight)
				owner.SetSize(ownerRawWidth, ownerRawHeight+v)
				owner.setYMin(ownerYMin)
			}
		} else {
			if ownerIsTargetParent {
				height := sourceHeight + pos - r.targetInitY + (targetHeight-targetInitHeight)*(1-pivot)
				tmp := ownerYMin
				owner.SetSize(ownerRawWidth, height)
				if owner.PivotAsAnchor() {
					owner.setYMin(tmp)
				}
			} else {
				v := delta * (1 - pivot)
				owner.SetSize(ownerRawWidth, ownerRawHeight+v)
				owner.setYMin(ownerYMin)
			}
		}
	case RelationTypeLeftExt_Left:
		tmp := ownerXMin
		var v float64
		if def.Percent {
			v = pos + (tmp-pos)*delta - tmp
		} else {
			v = delta * (-pivot)
		}
		owner.SetSize(ownerRawWidth-v, ownerRawHeight)
		owner.setXMin(tmp + v)
	case RelationTypeLeftExt_Right:
		tmp := ownerXMin
		var v float64
		if def.Percent {
			v = pos + (tmp-pos)*delta - tmp
		} else {
			v = delta * (1 - pivot)
		}
		owner.SetSize(ownerRawWidth-v, ownerRawHeight)
		owner.setXMin(tmp + v)
	case RelationTypeRightExt_Left:
		tmp := ownerXMin
		var v float64
		if def.Percent {
			v = pos + (tmp+ownerRawWidth-pos)*delta - (tmp + ownerRawWidth)
		} else {
			v = delta * (-pivot)
		}
		owner.SetSize(ownerRawWidth+v, ownerRawHeight)
		owner.setXMin(tmp)
	case RelationTypeRightExt_Right:
		tmp := ownerXMin
		if def.Percent {
			if ownerIsTargetParent {
				width := pos + targetWidth - targetWidth*pivot + (sourceWidth-r.targetInitX-targetInitWidth+targetInitWidth*pivot)*delta
				owner.SetSize(width, ownerRawHeight)
				if owner.PivotAsAnchor() {
					owner.setXMin(tmp)
				}
			} else {
				v := pos + (tmp+ownerRawWidth-pos)*delta - (tmp + ownerRawWidth)
				owner.SetSize(ownerRawWidth+v, ownerRawHeight)
				owner.setXMin(tmp)
			}
		} else {
			if ownerIsTargetParent {
				width := sourceWidth + pos - r.targetInitX + (targetWidth-targetInitWidth)*(1-pivot)
				owner.SetSize(width, ownerRawHeight)
				if owner.PivotAsAnchor() {
					owner.setXMin(tmp)
				}
			} else {
				v := delta * (1 - pivot)
				owner.SetSize(ownerRawWidth+v, ownerRawHeight)
				owner.setXMin(tmp)
			}
		}
	case RelationTypeTopExt_Top:
		tmp := ownerYMin
		var v float64
		if def.Percent {
			v = pos + (tmp-pos)*delta - tmp
		} else {
			v = delta * (-pivot)
		}
		owner.SetSize(ownerRawWidth, ownerRawHeight-v)
		owner.setYMin(tmp + v)
	case RelationTypeTopExt_Bottom:
		tmp := ownerYMin
		var v float64
		if def.Percent {
			v = pos + (tmp-pos)*delta - tmp
		} else {
			v = delta * (1 - pivot)
		}
		owner.SetSize(ownerRawWidth, ownerRawHeight-v)
		owner.setYMin(tmp + v)
	case RelationTypeBottomExt_Top:
		tmp := ownerYMin
		var v float64
		if def.Percent {
			v = pos + (tmp+ownerRawHeight-pos)*delta - (tmp + ownerRawHeight)
		} else {
			v = delta * (-pivot)
		}
		owner.SetSize(ownerRawWidth, ownerRawHeight+v)
		owner.setYMin(tmp)
	case RelationTypeBottomExt_Bottom:
		tmp := ownerYMin
		if def.Percent {
			if ownerIsTargetParent {
				height := pos + targetHeight - targetHeight*pivot + (sourceHeight-r.targetInitY-targetInitHeight+targetInitHeight*pivot)*delta
				owner.SetSize(ownerRawWidth, height)
				if owner.PivotAsAnchor() {
					owner.setYMin(tmp)
				}
			} else {
				v := pos + (tmp+ownerRawHeight-pos)*delta - (tmp + ownerRawHeight)
				owner.SetSize(ownerRawWidth, ownerRawHeight+v)
				owner.setYMin(tmp)
			}
		} else {
			if ownerIsTargetParent {
				height := sourceHeight + pos - r.targetInitY + (targetHeight-targetInitHeight)*(1-pivot)
				owner.SetSize(ownerRawWidth, height)
				if owner.PivotAsAnchor() {
					owner.setYMin(tmp)
				}
			} else {
				v := delta * (1 - pivot)
				owner.SetSize(ownerRawWidth, ownerRawHeight+v)
				owner.setYMin(tmp)
			}
		}
	}
}

func (r *RelationItem) resolveSourceWidth(owner *GObject) float64 {
	if owner == nil {
		return 0
	}
	if w := owner.SourceWidth(); w > 0 {
		return w
	}
	if w := owner.InitWidth(); w > 0 {
		return w
	}
	return owner.RawWidth()
}

func (r *RelationItem) resolveSourceHeight(owner *GObject) float64 {
	if owner == nil {
		return 0
	}
	if h := owner.SourceHeight(); h > 0 {
		return h
	}
	if h := owner.InitHeight(); h > 0 {
		return h
	}
	return owner.RawHeight()
}

func pivotAdjust(t RelationType, pivot float64, axis int) float64 {
	if axis == 0 {
		switch t {
		case RelationTypeLeft_Left:
			return -pivot
		case RelationTypeLeft_Center:
			return 0.5 - pivot
		case RelationTypeLeft_Right:
			return 1 - pivot
		case RelationTypeCenter_Center:
			return 0.5 - pivot
		case RelationTypeRight_Left:
			return -pivot
		case RelationTypeRight_Center:
			return 0.5 - pivot
		case RelationTypeRight_Right:
			return 1 - pivot
		}
	} else {
		switch t {
		case RelationTypeTop_Top:
			return -pivot
		case RelationTypeTop_Middle:
			return 0.5 - pivot
		case RelationTypeTop_Bottom:
			return 1 - pivot
		case RelationTypeMiddle_Middle:
			return 0.5 - pivot
		case RelationTypeBottom_Top:
			return -pivot
		case RelationTypeBottom_Middle:
			return 0.5 - pivot
		case RelationTypeBottom_Bottom:
			return 1 - pivot
		}
	}
	return 0
}
func (r *RelationItem) applyOnXYChanged(def RelationDef, dx, dy float64) {
	if r == nil {
		return
	}
	owner := r.owner
	switch def.Type {
	case RelationTypeLeft_Left, RelationTypeLeft_Center, RelationTypeLeft_Right,
		RelationTypeCenter_Center, RelationTypeRight_Left, RelationTypeRight_Center, RelationTypeRight_Right:
		owner.SetPosition(owner.X()+dx, owner.Y())
	case RelationTypeTop_Top, RelationTypeTop_Middle, RelationTypeTop_Bottom,
		RelationTypeMiddle_Middle, RelationTypeBottom_Top, RelationTypeBottom_Middle, RelationTypeBottom_Bottom:
		owner.SetPosition(owner.X(), owner.Y()+dy)
	case RelationTypeLeftExt_Left, RelationTypeLeftExt_Right:
		newWidth := owner.Width() - dx
		if newWidth < 0 {
			newWidth = 0
		}
		owner.SetPosition(owner.X()+dx, owner.Y())
		owner.SetSize(newWidth, owner.Height())
	case RelationTypeRightExt_Left, RelationTypeRightExt_Right:
		owner.SetSize(owner.Width()+dx, owner.Height())
	case RelationTypeTopExt_Top, RelationTypeTopExt_Bottom:
		newHeight := owner.Height() - dy
		if newHeight < 0 {
			newHeight = 0
		}
		owner.SetPosition(owner.X(), owner.Y()+dy)
		owner.SetSize(owner.Width(), newHeight)
	case RelationTypeBottomExt_Top, RelationTypeBottomExt_Bottom:
		owner.SetSize(owner.Width(), owner.Height()+dy)
	}
}

func (r *RelationItem) applyOnSelfResized(dWidth, dHeight float64, applyPivot bool) {
	if r == nil || len(r.defs) == 0 {
		return
	}
	if rel := r.owner.Relations(); rel != nil {
		rel.handling = r.target
		defer func() { rel.handling = nil }()
	}
	x := r.owner.X()
	y := r.owner.Y()
	px, py := r.owner.Pivot()
	if !applyPivot {
		px = 0
		py = 0
	}
	for _, def := range r.defs {
		switch def.Type {
		case RelationTypeCenter_Center:
			x -= (0.5 - px) * dWidth
		case RelationTypeRight_Left, RelationTypeRight_Center, RelationTypeRight_Right:
			x -= (1 - px) * dWidth
		case RelationTypeMiddle_Middle:
			y -= (0.5 - py) * dHeight
		case RelationTypeBottom_Top, RelationTypeBottom_Middle, RelationTypeBottom_Bottom:
			y -= (1 - py) * dHeight
		}
	}
	if x != r.owner.X() || y != r.owner.Y() {
		r.owner.SetPosition(x, y)
	}
}
