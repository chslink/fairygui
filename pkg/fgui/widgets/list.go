package widgets

import (
	"sort"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// GList represents a minimal list widget backed by a component package item.
type GList struct {
	*core.GComponent
	packageItem    *assets.PackageItem
	defaultItem    string
	resource       string
	items          []*core.GObject
	selected       int
	selectionMode  ListSelectionMode
	selectionCtrl  *core.Controller
	ctrlListener   func(*core.Controller)
	ctrlListenerID int
	selectedSet    map[int]struct{}
	lastSelected   int
	updatingCtrl   bool
	updatingList   bool
}

// ListSelectionMode mirrors FairyGUI's list selection options.
type ListSelectionMode int

const (
	ListSelectionModeSingle ListSelectionMode = iota
	ListSelectionModeMultiple
	ListSelectionModeMultipleSingleClick
	ListSelectionModeNone
)

// NewList constructs an empty list widget.
func NewList() *GList {
	return &GList{
		GComponent:    core.NewGComponent(),
		selected:      -1,
		lastSelected:  -1,
		selectionMode: ListSelectionModeSingle,
	}
}

// SetPackageItem stores the component package item used by this list.
func (l *GList) SetPackageItem(item *assets.PackageItem) {
	l.packageItem = item
}

// PackageItem returns the associated package item, if any.
func (l *GList) PackageItem() *assets.PackageItem {
	return l.packageItem
}

// SetDefaultItem records the default item resource id for this list.
func (l *GList) SetDefaultItem(value string) {
	l.defaultItem = value
}

// DefaultItem returns the default item resource id.
func (l *GList) DefaultItem() string {
	return l.defaultItem
}

// SetResource stores the raw resource identifier declared on the component child.
func (l *GList) SetResource(res string) {
	l.resource = res
}

// Resource returns the stored raw resource identifier.
func (l *GList) Resource() string {
	return l.resource
}

// SetSelectionMode updates the list selection strategy.
func (l *GList) SetSelectionMode(mode ListSelectionMode) {
	if l == nil {
		return
	}
	if l.selectionMode == mode {
		return
	}
	l.selectionMode = mode
	switch mode {
	case ListSelectionModeNone:
		l.clearSelection(true)
	case ListSelectionModeSingle:
		if len(l.selectedSet) > 1 && l.selected >= 0 {
			l.SetSelectedIndex(l.selected)
		} else if len(l.selectedSet) == 0 {
			l.selected = -1
		}
	}
}

// SetSelectionController binds a controller that mirrors list selection.
func (l *GList) SetSelectionController(ctrl *core.Controller) {
	if l == nil {
		return
	}
	if l.selectionCtrl != nil && l.ctrlListenerID != 0 {
		l.selectionCtrl.RemoveSelectionListener(l.ctrlListenerID)
	}
	l.selectionCtrl = ctrl
	l.ctrlListener = nil
	l.ctrlListenerID = 0
	if ctrl != nil {
		listener := func(c *core.Controller) {
			if l.selectionCtrl != c {
				return
			}
			if l.updatingCtrl {
				return
			}
			l.updatingList = true
			l.SetSelectedIndex(c.SelectedIndex())
			l.updatingList = false
		}
		l.ctrlListener = listener
		l.ctrlListenerID = ctrl.AddSelectionListener(listener)
		l.updatingList = true
		l.SetSelectedIndex(ctrl.SelectedIndex())
		l.updatingList = false
	}
}

// AddItem appends a child object to the list and wires basic click selection.
func (l *GList) AddItem(obj *core.GObject) {
	if l == nil || obj == nil {
		return
	}
	l.GComponent.AddChild(obj)
	index := len(l.items)
	l.items = append(l.items, obj)
	obj.On(laya.EventClick, func(evt laya.Event) {
		l.handleItemClick(index)
	})
}

// Items returns the child objects tracked by the list.
func (l *GList) Items() []*core.GObject {
	return append([]*core.GObject(nil), l.items...)
}

// SelectedIndex returns the currently selected item index (or -1 when none).
func (l *GList) SelectedIndex() int {
	if l == nil {
		return -1
	}
	if l.selected < 0 || l.selected >= len(l.items) {
		return -1
	}
	return l.selected
}

// SelectedItem returns the currently selected child (if any).
func (l *GList) SelectedItem() *core.GObject {
	if idx := l.SelectedIndex(); idx >= 0 {
		return l.items[idx]
	}
	return nil
}

// SelectedIndices returns all selected indices in ascending order.
func (l *GList) SelectedIndices() []int {
	if l == nil || len(l.selectedSet) == 0 {
		return nil
	}
	out := make([]int, 0, len(l.selectedSet))
	for idx := range l.selectedSet {
		out = append(out, idx)
	}
	sort.Ints(out)
	return out
}

// IsSelected reports whether the item at index is selected.
func (l *GList) IsSelected(index int) bool {
	if l == nil || index < 0 {
		return false
	}
	_, ok := l.selectedSet[index]
	return ok
}

// SetSelectedIndex programmatically updates the selected item.
func (l *GList) SetSelectedIndex(index int) {
	if l == nil {
		return
	}
	if len(l.items) == 0 {
		l.clearSelection(true)
		return
	}
	if index < 0 || index >= len(l.items) {
		l.clearSelection(true)
		return
	}
	if l.selectionMode == ListSelectionModeNone {
		l.clearSelection(true)
		return
	}
	set := map[int]struct{}{index: {}}
	l.updateSelection(set, index, true)
}

// SetSelectedIndices replaces the current selection with the provided indices.
func (l *GList) SetSelectedIndices(indices []int) {
	if l == nil {
		return
	}
	if len(indices) == 0 {
		l.clearSelection(true)
		return
	}
	if l.selectionMode == ListSelectionModeNone {
		l.clearSelection(true)
		return
	}
	if l.selectionMode == ListSelectionModeSingle {
		l.SetSelectedIndex(indices[0])
		return
	}
	set := make(map[int]struct{}, len(indices))
	for _, idx := range indices {
		if idx >= 0 && idx < len(l.items) {
			set[idx] = struct{}{}
		}
	}
	if len(set) == 0 {
		l.clearSelection(true)
		return
	}
	primary := -1
	for idx := range set {
		if primary == -1 || idx < primary {
			primary = idx
		}
	}
	l.updateSelection(set, primary, true)
}

// AddSelection adds the specified index to the current selection.
func (l *GList) AddSelection(index int) {
	if l == nil || index < 0 || index >= len(l.items) {
		return
	}
	if l.selectionMode == ListSelectionModeNone {
		return
	}
	if l.selectionMode == ListSelectionModeSingle {
		l.SetSelectedIndex(index)
		return
	}
	if l.selectedSet == nil {
		l.selectedSet = make(map[int]struct{})
	}
	if _, exists := l.selectedSet[index]; exists && l.selected == index {
		return
	}
	set := l.copySelectionSet()
	set[index] = struct{}{}
	l.updateSelection(set, index, true)
}

// RemoveSelection removes the specified index from the current selection.
func (l *GList) RemoveSelection(index int) {
	if l == nil || index < 0 {
		return
	}
	if len(l.selectedSet) == 0 {
		return
	}
	if _, exists := l.selectedSet[index]; !exists {
		return
	}
	set := l.copySelectionSet()
	delete(set, index)
	primary := l.selected
	if primary == index {
		primary = -1
	}
	l.updateSelection(set, primary, true)
}

// ClearSelection clears all selections.
func (l *GList) ClearSelection() {
	l.clearSelection(true)
}

func (l *GList) handleItemClick(index int) {
	if l == nil {
		return
	}
	switch l.selectionMode {
	case ListSelectionModeNone:
		return
	case ListSelectionModeSingle:
		l.SetSelectedIndex(index)
	case ListSelectionModeMultiple:
		if l.IsSelected(index) {
			l.SetSelectedIndex(index)
		} else if len(l.selectedSet) == 0 {
			l.SetSelectedIndex(index)
		} else {
			set := l.copySelectionSet()
			set[index] = struct{}{}
			l.updateSelection(set, index, true)
		}
	case ListSelectionModeMultipleSingleClick:
		if l.IsSelected(index) {
			l.RemoveSelection(index)
		} else {
			l.AddSelection(index)
		}
	default:
		l.SetSelectedIndex(index)
	}
}

func (l *GList) clearSelection(notify bool) {
	if l == nil {
		return
	}
	if len(l.selectedSet) == 0 && l.selected == -1 {
		if notify {
			l.GComponent.GObject.Emit(laya.EventStateChanged, nil)
		}
		l.lastSelected = -1
		return
	}
	l.updateSelection(nil, -1, notify)
}

func (l *GList) copySelectionSet() map[int]struct{} {
	set := make(map[int]struct{}, len(l.selectedSet))
	for idx := range l.selectedSet {
		set[idx] = struct{}{}
	}
	return set
}

func (l *GList) updateSelection(newSet map[int]struct{}, primary int, notify bool) {
	if l == nil {
		return
	}
	var clean map[int]struct{}
	if len(newSet) > 0 {
		clean = make(map[int]struct{}, len(newSet))
		for idx := range newSet {
			if idx >= 0 && idx < len(l.items) {
				clean[idx] = struct{}{}
			}
		}
	} else {
		clean = make(map[int]struct{})
	}

	changed := false
	if len(clean) != len(l.selectedSet) {
		changed = true
	} else {
		for idx := range clean {
			if _, ok := l.selectedSet[idx]; !ok {
				changed = true
				break
			}
		}
		if !changed {
			for idx := range l.selectedSet {
				if _, ok := clean[idx]; !ok {
					changed = true
					break
				}
			}
		}
	}

	newPrimary := -1
	if primary >= 0 {
		if _, ok := clean[primary]; ok {
			newPrimary = primary
		}
	}
	if newPrimary == -1 && len(clean) > 0 {
		for idx := range clean {
			if newPrimary == -1 || idx < newPrimary {
				newPrimary = idx
			}
		}
	}

	if !changed && newPrimary == l.selected {
		return
	}

	for idx, child := range l.items {
		if child == nil {
			continue
		}
		_, inNew := clean[idx]
		_, inOld := l.selectedSet[idx]
		if inNew == inOld {
			continue
		}
		l.applyItemSelection(idx, inNew)
	}

	if len(clean) == 0 {
		l.selectedSet = nil
	} else {
		l.selectedSet = clean
	}
	l.selected = newPrimary
	if newPrimary >= 0 {
		l.lastSelected = newPrimary
	} else {
		l.lastSelected = -1
	}

	var selectedObj *core.GObject
	if newPrimary >= 0 && newPrimary < len(l.items) {
		selectedObj = l.items[newPrimary]
	}

	if notify {
		if selectedObj != nil {
			l.GComponent.GObject.Emit(laya.EventStateChanged, selectedObj)
		} else {
			l.GComponent.GObject.Emit(laya.EventStateChanged, nil)
		}
	}

	if l.selectionCtrl != nil && !l.updatingList {
		if newPrimary >= 0 {
			l.updatingCtrl = true
			l.selectionCtrl.SetSelectedIndex(newPrimary)
			l.updatingCtrl = false
		}
	}
}

func (l *GList) applyItemSelection(index int, selected bool) {
	if l == nil || index < 0 || index >= len(l.items) {
		return
	}
	item := l.items[index]
	if item == nil {
		return
	}
	switch data := item.Data().(type) {
	case *GButton:
		data.SetSelected(selected)
	case interface{ SetSelected(bool) }:
		data.SetSelected(selected)
	}
}
