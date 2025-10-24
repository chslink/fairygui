package widgets

import (
	"sort"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// ListLayoutType mirrors FairyGUI 的列表布局枚举。
type ListLayoutType int

const (
	ListLayoutTypeSingleColumn ListLayoutType = iota
	ListLayoutTypeSingleRow
	ListLayoutTypeFlowHorizontal
	ListLayoutTypeFlowVertical
	ListLayoutTypePagination
)

// ListChildrenRenderOrder mirrors FairyGUI 的子对象渲染顺序。
type ListChildrenRenderOrder int

const (
	ListChildrenRenderOrderAscent ListChildrenRenderOrder = iota
	ListChildrenRenderOrderDescent
	ListChildrenRenderOrderArch
)

// ListMargin 记录列表内容边距。
type ListMargin struct {
	Top    int
	Bottom int
	Left   int
	Right  int
}

// GList represents a minimal list widget backed by a component package item.
type GList struct {
	*core.GComponent
	packageItem    *assets.PackageItem
	defaultItem    string
	resource       string
	items          []*core.GObject
	itemHandlers   map[*core.GObject]laya.Listener
	selected       int
	selectionMode  ListSelectionMode
	selectionCtrl  *core.Controller
	ctrlListener   func(*core.Controller)
	ctrlListenerID int
	selectedSet    map[int]struct{}
	lastSelected   int
	updatingCtrl   bool
	updatingList   bool
	layout         ListLayoutType
	align          LoaderAlign
	verticalAlign  LoaderAlign
	lineGap        int
	columnGap      int
	lineCount      int
	columnCount    int
	autoResizeItem bool
	childrenOrder  ListChildrenRenderOrder
	apexIndex      int
	margin         ListMargin
	overflow       assets.OverflowType
	scrollToView   bool
	foldInvisible  bool
}

// ComponentRoot exposes the embedded component for helpers.
func (l *GList) ComponentRoot() *core.GComponent {
	if l == nil {
		return nil
	}
	return l.GComponent
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
		layout:        ListLayoutTypeSingleColumn,
		align:         LoaderAlignLeft,
		verticalAlign: LoaderAlignTop,
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
	l.InsertItemAt(obj, len(l.items))
}

// InsertItemAt inserts an object at the specified index.
func (l *GList) InsertItemAt(obj *core.GObject, index int) {
	if l == nil || obj == nil {
		return
	}
	if index < 0 {
		index = 0
	}
	if index > len(l.items) {
		index = len(l.items)
	}
	l.GComponent.AddChildAt(obj, index)
	l.items = append(l.items, nil)
	copy(l.items[index+1:], l.items[index:])
	l.items[index] = obj
	l.attachItemClick(obj)
	if l.selected >= index && l.selected != -1 {
		l.selected++
	}
	if len(l.selectedSet) > 0 {
		updated := make(map[int]struct{}, len(l.selectedSet))
		for idx := range l.selectedSet {
			if idx >= index {
				updated[idx+1] = struct{}{}
			} else {
				updated[idx] = struct{}{}
			}
		}
		l.selectedSet = updated
	}
}

// RemoveItemAt removes the object at the given index.
func (l *GList) RemoveItemAt(index int) {
	if l == nil || index < 0 || index >= len(l.items) {
		return
	}
	obj := l.items[index]
	if handler, ok := l.itemHandlers[obj]; ok {
		obj.Off(laya.EventClick, handler)
		delete(l.itemHandlers, obj)
	}
	l.GComponent.RemoveChild(obj)
	copy(l.items[index:], l.items[index+1:])
	l.items = l.items[:len(l.items)-1]

	if l.selected == index {
		l.selected = -1
	} else if l.selected > index {
		l.selected--
	}

	if len(l.selectedSet) > 0 {
		updated := make(map[int]struct{}, len(l.selectedSet))
		for idx := range l.selectedSet {
			switch {
			case idx == index:
				// drop
			case idx > index:
				updated[idx-1] = struct{}{}
			default:
				updated[idx] = struct{}{}
			}
		}
		if len(updated) == 0 {
			l.selectedSet = nil
		} else {
			l.selectedSet = updated
		}
	}
}

func (l *GList) attachItemClick(obj *core.GObject) {
	if l.itemHandlers == nil {
		l.itemHandlers = make(map[*core.GObject]laya.Listener)
	}
	if handler, ok := l.itemHandlers[obj]; ok && handler != nil {
		obj.Off(laya.EventClick, handler)
	}
	handler := func(evt laya.Event) {
		index := l.indexOf(obj)
		if index >= 0 {
			l.handleItemClick(index)
		}
	}
	l.itemHandlers[obj] = handler
	obj.On(laya.EventClick, handler)
}

func (l *GList) indexOf(obj *core.GObject) int {
	for idx, item := range l.items {
		if item == obj {
			return idx
		}
	}
	return -1
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

// Layout returns当前列表布局类型。
func (l *GList) Layout() ListLayoutType {
	if l == nil {
		return ListLayoutTypeSingleColumn
	}
	return l.layout
}

// Align 返回水平对齐方式。
func (l *GList) Align() LoaderAlign {
	if l == nil {
		return LoaderAlignLeft
	}
	return l.align
}

// VerticalAlign 返回垂直对齐方式。
func (l *GList) VerticalAlign() LoaderAlign {
	if l == nil {
		return LoaderAlignTop
	}
	return l.verticalAlign
}

// LineGap 返回行间距。
func (l *GList) LineGap() int {
	return l.lineGap
}

// ColumnGap 返回列间距。
func (l *GList) ColumnGap() int {
	return l.columnGap
}

// LineCount 返回行数限制。
func (l *GList) LineCount() int {
	return l.lineCount
}

// ColumnCount 返回列数限制。
func (l *GList) ColumnCount() int {
	return l.columnCount
}

// AutoResizeItem 表示是否自动调整子项尺寸。
func (l *GList) AutoResizeItem() bool {
	return l.autoResizeItem
}

// ChildrenRenderOrder 返回子对象渲染顺序。
func (l *GList) ChildrenRenderOrder() ListChildrenRenderOrder {
	return l.childrenOrder
}

// ApexIndex 返回拱形顺序起点索引。
func (l *GList) ApexIndex() int {
	return l.apexIndex
}

// Margin 返回内容边距。
func (l *GList) Margin() ListMargin {
	return l.margin
}

// Overflow 返回溢出处理策略。
func (l *GList) Overflow() assets.OverflowType {
	return l.overflow
}

// ScrollItemToViewOnClick 指示点击条目时是否滚动至可视区域。
func (l *GList) ScrollItemToViewOnClick() bool {
	return l.scrollToView
}

// FoldInvisibleItems 指示是否折叠不可见元素。
func (l *GList) FoldInvisibleItems() bool {
	return l.foldInvisible
}

// SetupBeforeAdd 解析列表加入父节点前的配置。
func (l *GList) SetupBeforeAdd(ctx *SetupContext, buf *utils.ByteBuffer) {
	if l == nil || buf == nil {
		return
	}
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(0, 5) || buf.Remaining() <= 0 {
		return
	}
	l.layout = clampListLayout(ListLayoutType(buf.ReadByte()))
	if buf.Remaining() > 0 {
		mode := ListSelectionMode(buf.ReadByte())
		if mode < ListSelectionModeSingle || mode > ListSelectionModeNone {
			mode = ListSelectionModeSingle
		}
		l.SetSelectionMode(mode)
	}
	if buf.Remaining() > 0 {
		l.align = mapListAlign(buf.ReadByte())
	}
	if buf.Remaining() > 0 {
		l.verticalAlign = mapListVerticalAlign(buf.ReadByte())
	}
	if buf.Remaining() >= 2 {
		l.lineGap = int(buf.ReadInt16())
	}
	if buf.Remaining() >= 2 {
		l.columnGap = int(buf.ReadInt16())
	}
	if buf.Remaining() >= 2 {
		l.lineCount = int(buf.ReadInt16())
	}
	if buf.Remaining() >= 2 {
		l.columnCount = int(buf.ReadInt16())
	}
	if buf.Remaining() > 0 {
		l.autoResizeItem = buf.ReadBool()
	}
	if buf.Remaining() > 0 {
		l.childrenOrder = mapChildrenRenderOrder(buf.ReadByte())
	}
	if buf.Remaining() >= 2 {
		l.apexIndex = int(buf.ReadInt16())
	}
	if buf.Remaining() > 0 {
		if buf.ReadBool() {
			if buf.Remaining() >= 16 {
				l.margin.Top = int(buf.ReadInt32())
				l.margin.Bottom = int(buf.ReadInt32())
				l.margin.Left = int(buf.ReadInt32())
				l.margin.Right = int(buf.ReadInt32())
			}
		} else {
			l.margin = ListMargin{}
		}
	}
	if buf.Remaining() > 0 {
		l.overflow = assets.OverflowType(buf.ReadByte())
	} else {
		l.overflow = assets.OverflowTypeVisible
	}
	if l.overflow == assets.OverflowTypeScroll {
		savedPos := buf.Pos()
		if buf.Seek(0, 7) {
			l.GComponent.SetupScroll(buf)
		}
		_ = buf.SetPos(savedPos)
	}
	if buf.Remaining() > 0 && buf.ReadBool() {
		if buf.Remaining() >= 8 {
			_ = buf.Skip(8)
		}
	}
	if buf.Version >= 2 {
		if buf.Remaining() > 0 {
			l.scrollToView = buf.ReadBool()
		}
		if buf.Remaining() > 0 {
			l.foldInvisible = buf.ReadBool()
		}
	}
	if !buf.Seek(0, 8) {
		return
	}
	if def := buf.ReadS(); def != nil && *def != "" {
		l.SetDefaultItem(*def)
	}
	if buf.Remaining() < 2 {
		return
	}
	cnt := int(buf.ReadInt16())
	for i := 0; i < cnt; i++ {
		if buf.Remaining() < 2 {
			break
		}
		nextPos := int(buf.ReadInt16()) + buf.Pos()
		if buf.Remaining() >= 2 {
			_ = buf.ReadS()
		}
		if nextPos < 0 || nextPos > buf.Len() {
			break
		}
		_ = buf.SetPos(nextPos)
	}
}

// SetupAfterAdd 在组件加入父对象后应用控制器索引等设置。
func (l *GList) SetupAfterAdd(ctx *SetupContext, buf *utils.ByteBuffer) {
	if l == nil || buf == nil {
		return
	}
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(0, 6) || buf.Remaining() < 2 {
		return
	}
	index := int(buf.ReadInt16())
	if index < 0 || ctx == nil || ctx.Parent == nil {
		return
	}
	controllers := ctx.Parent.Controllers()
	if index >= len(controllers) {
		return
	}
	l.SetSelectionController(controllers[index])
}

func mapListAlign(code int8) LoaderAlign {
	switch code {
	case 1:
		return LoaderAlignCenter
	case 2:
		return LoaderAlignRight
	default:
		return LoaderAlignLeft
	}
}

func mapListVerticalAlign(code int8) LoaderAlign {
	switch code {
	case 1:
		return LoaderAlignMiddle
	case 2:
		return LoaderAlignBottom
	default:
		return LoaderAlignTop
	}
}

func mapChildrenRenderOrder(code int8) ListChildrenRenderOrder {
	switch code {
	case 1:
		return ListChildrenRenderOrderDescent
	case 2:
		return ListChildrenRenderOrderArch
	default:
		return ListChildrenRenderOrderAscent
	}
}

func clampListLayout(value ListLayoutType) ListLayoutType {
	if value < ListLayoutTypeSingleColumn || value > ListLayoutTypePagination {
		return ListLayoutTypeSingleColumn
	}
	return value
}
