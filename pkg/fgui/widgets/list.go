package widgets

import (
	"log"
	"math"
	"sort"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/gears"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// globalObjectPool 全局对象池，供所有GList实例共享
// 这样可以最大化对象重用效率，避免每个列表都创建自己的池
var globalObjectPool *GObjectPool

// init 初始化全局对象池
func init() {
	globalObjectPool = NewGObjectPool()
}

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

// GList represents a list widget with virtual list support.
type GList struct {
	*core.GComponent
	packageItem    *assets.PackageItem
	defaultItem    string
	resource       string
	items          []*core.GObject
	itemHandlers   map[*core.GObject]laya.Listener
	hoverHandlers  map[*core.GObject]laya.ListenerID
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

	// 虚拟化支持 - 对应 TypeScript 版本的核心字段
	virtual            bool        // 是否启用虚拟化
	loop               bool        // 是否循环
	numItems           int         // 数据项总数
	realNumItems       int         // 实际项数（循环模式）
	firstIndex         int         // 左上角索引
	curLineItemCount   int         // 每行项目数
	curLineItemCount2  int         // 只用在页面模式，表示垂直方向的项目数
	itemSize           *laya.Point // 项目尺寸
	virtualListChanged int         // 1-内容改变, 2-尺寸改变
	virtualItems       []*ItemInfo // 虚拟项数组
	itemInfoVer        int         // 项信息版本
	eventLocked        bool        // 事件锁定
	scrollListenerID   int         // 滚动监听器 ID

	// 渲染回调
	itemRenderer func(index int, item *core.GObject) // 项目渲染器
	itemProvider func(index int) string              // 项目提供者
	pool         *GObjectPool                        // 对象池

	// 对象创建器，用于动态创建对象
	creator ObjectCreator // 对象创建器

	// 批量操作标志，用于避免重复计算布局
	batchAdding bool
	// 首次布局标志
	boundsInitialized bool
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

// ChildIndexToItemIndex converts a child index to a data item index.
// This is necessary for virtual lists where child indices don't directly map to data indices.
func (l *GList) ChildIndexToItemIndex(index int) int {
	if l == nil {
		return -1
	}
	
	// For non-virtual lists, child index equals item index
	if !l.virtual {
		return index
	}

	// Handle pagination layout
	if l.layout == ListLayoutTypePagination {
		for i := l.firstIndex; i < l.realNumItems; i++ {
			if l.virtualItems[i].obj != nil {
				index--
				if index < 0 {
					return i
				}
			}
		}
		return index
	}
	
	// Handle other layouts
	index += l.firstIndex
	// Apply loop behavior if enabled
	if l.loop && l.numItems > 0 {
		// Ensure index is positive and within bounds
		if index < 0 {
			index = (index % l.numItems) + l.numItems
		} else {
			index = index % l.numItems
		}
	}

	return index
}

// GetFirstChildInView returns the index of the first data item that is visible in the viewport.
// This delegates to the GComponent implementation and converts the result.
func (l *GList) GetFirstChildInView() int {
	if l == nil || l.GComponent == nil {
		return -1
	}
	childIndex := l.GComponent.GetFirstChildInView()
	if childIndex == -1 {
		return -1
	}
	return l.ChildIndexToItemIndex(childIndex)
}

// NewList constructs an empty list widget with virtual list support.
func NewList() *GList {
	list := &GList{
		GComponent:    core.NewGComponent(),
		selected:      -1,
		lastSelected:  -1,
		selectionMode: ListSelectionModeSingle,
		layout:        ListLayoutTypeSingleColumn,
		align:         LoaderAlignLeft,
		verticalAlign: LoaderAlignTop,
		// 关键修复：初始化items数组
		items:        make([]*core.GObject, 0),
		itemHandlers:  make(map[*core.GObject]laya.Listener),
		hoverHandlers: make(map[*core.GObject]laya.ListenerID),
		selectedSet:   make(map[int]struct{}),
		// 虚拟化相关初始化
		itemSize:     &laya.Point{},
		virtualItems: make([]*ItemInfo, 0),
		// 使用全局对象池，而不是每个列表创建独立的池
		// 这样可以在多个列表间共享相同类型的对象，最大化重用效率
		pool: globalObjectPool,
	}
	// 参考 TypeScript 原版：GList.ts 构造函数中设置 opaque=true
	list.GComponent.SetOpaque(true)
	return list
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

// SetSize 覆盖 GComponent 的 SetSize 方法，在尺寸变化时触发布局更新
// 对应 TypeScript 版本中通过 setBoundsChangedFlag() 触发延迟布局的机制
func (l *GList) SetSize(width, height float64) {
	if l == nil || l.GComponent == nil {
		return
	}

	oldWidth := l.GComponent.Width()
	oldHeight := l.GComponent.Height()

	// 调用父类 SetSize
	l.GComponent.SetSize(width, height)

	// 如果不是虚拟列表且有子项，在以下情况触发布局更新：
	// 1. 尺寸发生变化
	// 2. 首次调用（boundsInitialized=false）
	sizeChanged := width != oldWidth || height != oldHeight
	shouldUpdate := !l.virtual && len(l.items) > 0 && (sizeChanged || !l.boundsInitialized)

	if shouldUpdate {
		l.updateBounds()
		l.boundsInitialized = true
	}
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
	l.attachItemHover(obj)
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

	// 对应 TypeScript 版本 GComponent.addChildAt 中的 setBoundsChangedFlag() 调用
	// 添加子项后需要重新计算布局（除非在批量添加中或视口尺寸未设置）
	if !l.virtual && !l.batchAdding && l.GComponent.ViewWidth() > 0 {
		l.updateBounds()
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
	if id, ok := l.hoverHandlers[obj]; ok {
		obj.OffByID(laya.EventRollOver, id)
		delete(l.hoverHandlers, obj)
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
	handler := func(evt *laya.Event) {
		index := l.indexOf(obj)
		if index >= 0 {
			l.handleItemClick(index)
		}
	}
	l.itemHandlers[obj] = handler
	obj.On(laya.EventClick, handler)
}

// attachItemHover 注册 hover 事件监听，确保列表项的 hover 高亮
// 不受子元素（如 checkbox）拦截影响
func (l *GList) attachItemHover(obj *core.GObject) {
	if l.hoverHandlers == nil {
		l.hoverHandlers = make(map[*core.GObject]laya.ListenerID)
	}
	if id, ok := l.hoverHandlers[obj]; ok {
		obj.OffByID(laya.EventRollOver, id)
	}
	// 直接在 item 上监听 rollOver/rollOut，绕过子元素拦截
	rollOverID := obj.OnWithID(laya.EventRollOver, func(evt *laya.Event) {
		l.updateItemHover(obj, true)
	})
	// rollOut 也需要监听（虽然 rollOver 已经更新 hovered=true，但 rollOut 需要清除）
	_ = obj.OnWithID(laya.EventRollOut, func(evt *laya.Event) {
		l.updateItemHover(obj, false)
	})
	l.hoverHandlers[obj] = rollOverID
}

func (l *GList) updateItemHover(obj *core.GObject, hovered bool) {
	if obj == nil {
		return
	}
	if btn, ok := obj.Data().(*GButton); ok {
		btn.SetHovered(hovered)
	} else if hoverable, ok := obj.Data().(interface{ SetHovered(bool) }); ok {
		hoverable.SetHovered(hovered)
	}
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
	if len(l.items) == 0 && !l.virtual {
		l.clearSelection(true)
		return
	}

	// 检查索引范围：对虚拟列表使用 numItems，对非虚拟列表使用 len(items)
	maxIdx := len(l.items)
	if l.virtual {
		maxIdx = l.numItems
	}
	if index < 0 || index >= maxIdx {
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
// scrollItToView: whether to scroll the item into view
// 对应 TypeScript 版本的 addSelection(index: number, scrollItToView?: boolean)
func (l *GList) AddSelection(index int, scrollItToView bool) {
	if l == nil {
		return
	}

	// 虚拟列表：检查数据索引范围
	// 非虚拟列表：检查子对象数组范围
	if l.virtual {
		if index < 0 || index >= l.numItems {
			return
		}
	} else {
		if index < 0 || index >= len(l.items) {
			return
		}
	}

	if l.selectionMode == ListSelectionModeNone {
		return
	}

	// 虚拟列表需要先刷新以确保 virtualItems 数组有效
	// 对应 TypeScript: this.checkVirtualList()
	if l.virtual {
		l.CheckVirtualList()
	}

	if l.selectionMode == ListSelectionModeSingle {
		l.SetSelectedIndex(index)
		// 单选模式下，SetSelectedIndex 内部不会触发滚动，需要这里处理
		if scrollItToView {
			l.ScrollToView(index)
		}
		return
	}

	// 对应 TypeScript: if (scrollItToView) this.scrollToView(index)
	if scrollItToView {
		l.ScrollToView(index)
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

// ScrollToView scrolls the list to make the specified item visible.
// 对应 TypeScript 版本的 scrollToView(index: number, ani?: boolean, setFirst?: boolean)
// 当前简化版本不支持 ani 和 setFirst 参数
func (l *GList) ScrollToView(index int) {
	if l == nil {
		return
	}

	if l.virtual {
		// 虚拟列表：根据索引计算位置
		if l.numItems == 0 {
			return
		}

		// 确保虚拟列表已刷新
		l.CheckVirtualList()

		// 关键修复：确保 virtualItems 数组足够大，可以访问任意索引
		// virtualItems 应该包含所有数据项，而不仅仅是可见项
		l.EnsureVirtualItems(l.numItems)

		if index < 0 || index >= l.numItems {
			log.Printf("❌ ScrollToView: invalid index %d, numItems=%d", index, l.numItems)
			return
		}

		// 处理循环模式（对应 TypeScript GList.ts:869-870）
		// if (this._loop)
		//     index = Math.floor(this._firstIndex / this._numItems) * this._numItems + index;
		if l.loop {
			index = (l.firstIndex/l.numItems)*l.numItems + index
		}

		var x, y, width, height float64

		// 根据布局类型计算项目位置
		// 对应 TypeScript GList.ts:872-890
		if l.layout == ListLayoutTypeSingleColumn || l.layout == ListLayoutTypeFlowHorizontal {
			// 垂直滚动：累加之前所有项的高度
			// TypeScript: for (var i: number = 0; i < index; i++)
			//     pos += this._virtualItems[i].height + this._lineGap;
			var pos float64 = 0
			for i := 0; i < index; i++ {
				itemHeight := l.virtualItems[i].height
				if itemHeight == 0 && l.itemSize.Y > 0 {
					itemHeight = int(l.itemSize.Y)
				}
				pos += float64(itemHeight) + float64(l.lineGap)
			}
			x = 0
			y = pos
			itemWidth := l.virtualItems[index].width
			if itemWidth == 0 && l.itemSize.X > 0 {
				itemWidth = int(l.itemSize.X)
			}
			width = float64(itemWidth)
			height = l.itemSize.Y
		} else if l.layout == ListLayoutTypeSingleRow || l.layout == ListLayoutTypeFlowVertical {
			// 水平滚动：累加之前所有项的宽度
			// TypeScript: for (var i: number = 0; i < index; i++)
			//     pos += this._virtualItems[i].width + this._columnGap;
			var pos float64 = 0
			for i := 0; i < index; i++ {
				itemWidth := l.virtualItems[i].width
				if itemWidth == 0 && l.itemSize.X > 0 {
					itemWidth = int(l.itemSize.X)
				}
				pos += float64(itemWidth) + float64(l.columnGap)
			}
			x = pos
			y = 0
			width = l.itemSize.X
			itemHeight := l.virtualItems[index].height
			if itemHeight == 0 && l.itemSize.Y > 0 {
				itemHeight = int(l.itemSize.Y)
			}
			height = float64(itemHeight)
		} else {
			// 分页布局（Pagination）
			// TypeScript GList.ts:892-895
			// var page: number = index / (this._curLineItemCount * this._curLineItemCount2);
			// rect = new Laya.Rectangle(page * this.viewWidth + (index % this._curLineItemCount) * (ii.width + this._columnGap),
			//     (index / this._curLineItemCount) % this._curLineItemCount2 * (ii.height + this._lineGap),
			//     ii.width, ii.height);
			if l.curLineItemCount > 0 && l.curLineItemCount2 > 0 {
				page := index / (l.curLineItemCount * l.curLineItemCount2)
				pane := l.GComponent.ScrollPane()
				viewWidth := float64(0)
				if pane != nil {
					viewWidth = pane.ViewWidth()
				}
				itemWidth := l.virtualItems[index].width
				if itemWidth == 0 && l.itemSize.X > 0 {
					itemWidth = int(l.itemSize.X)
				}
				itemHeight := l.virtualItems[index].height
				if itemHeight == 0 && l.itemSize.Y > 0 {
					itemHeight = int(l.itemSize.Y)
				}
				x = float64(page)*viewWidth + float64(index%l.curLineItemCount)*(float64(itemWidth)+float64(l.columnGap))
				y = float64((index/l.curLineItemCount)%l.curLineItemCount2) * (float64(itemHeight) + float64(l.lineGap))
				width = float64(itemWidth)
				height = float64(itemHeight)
			} else {
				log.Printf("⚠️  ScrollToView: pagination layout not initialized")
				return
			}
		}

		// 调用 ScrollPane.ScrollToRect
		// 对应 TypeScript GList.ts:898
		// if (this._scrollPane)
		//     this._scrollPane.scrollToView(rect, ani, setFirst);
		pane := l.GComponent.ScrollPane()
		if pane != nil {
			log.Printf("📍 ScrollToView: scrolling to index=%d, rect=(%.0f,%.0f,%.0f,%.0f)", index, x, y, width, height)
			pane.ScrollToRect(x, y, width, height, false)
		}
	} else {
		// 非虚拟列表：使用现有的 scrollItemToView 方法
		// 对应 TypeScript GList.ts:901-906
		l.scrollItemToView(index)
	}
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
	if l.scrollToView {
		l.scrollItemToView(index)
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
			l.AddSelection(index, false)
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

func (l *GList) scrollItemToView(index int) {
	if l == nil || index < 0 || index >= len(l.items) {
		return
	}
	item := l.items[index]
	if item == nil {
		return
	}
	pane := l.GComponent.ScrollPane()
	if pane == nil {
		return
	}
	width := item.Width()
	height := item.Height()
	if comp, ok := item.Data().(*core.GComponent); ok && comp != nil {
		if width <= 0 {
			width = comp.Width()
		}
		if height <= 0 {
			height = comp.Height()
		}
	}
	if width <= 0 {
		width = l.GComponent.Width()
	}
	if height <= 0 {
		height = l.GComponent.Height()
	}
	pane.ScrollToRect(item.X(), item.Y(), width, height, true)
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
		// 关键修复：对于虚拟列表，使用 numItems（数据项总数）而不是 len(items)（渲染项数量）
		// 对应 TypeScript 版本中虚拟列表的选择范围应该是 [0, numItems)
		maxIdx := len(l.items)
		if l.virtual {
			maxIdx = l.numItems
		}
		for idx := range newSet {
			if idx >= 0 && idx < maxIdx {
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

	// 处理渲染项目的选中状态
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

	// 关键修复：对于虚拟列表，更新 ItemInfo.selected 字段
	// 这样当项目被滚动到视野内时，可以正确应用选中状态
	if l.virtual && l.virtualItems != nil {
		for idx := range clean {
			if idx >= 0 && idx < len(l.virtualItems) {
				if l.virtualItems[idx] != nil {
					l.virtualItems[idx].selected = true
				}
			}
		}
		// 清除不再选中的项
		for i, ii := range l.virtualItems {
			if ii != nil {
				if _, inClean := clean[i]; !inClean {
					ii.selected = false
				}
			}
		}
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

// SetScrollItemToViewOnClick toggles automatic scrolling to selected item.
func (l *GList) SetScrollItemToViewOnClick(value bool) {
	if l == nil {
		return
	}
	l.scrollToView = value
}

// FoldInvisibleItems 指示是否折叠不可见元素。
func (l *GList) FoldInvisibleItems() bool {
	return l.foldInvisible
}

// SetupBeforeAdd 解析列表加入父节点前的配置。
// 对应 TypeScript 版本 GList.setup_beforeAdd (GList.ts:2241-2282)
func (l *GList) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if l == nil || buf == nil {
		return
	}

	// 首先调用父类GComponent处理组件和基础属性
	// TypeScript: super.setup_beforeAdd(buffer, beginPos);
	l.GComponent.SetupBeforeAdd(buf, beginPos, nil)

	// 然后处理GList特定属性（block 5）
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(beginPos, 5) || buf.Remaining() <= 0 {
		return
	}

	// TypeScript: this._layout = buffer.readByte();
	l.layout = clampListLayout(ListLayoutType(buf.ReadByte()))

	// TypeScript: this._selectionMode = buffer.readByte();
	mode := ListSelectionMode(buf.ReadByte())
	if mode < ListSelectionModeSingle || mode > ListSelectionModeNone {
		mode = ListSelectionModeSingle
	}
	l.SetSelectionMode(mode)

	// TypeScript: i1 = buffer.readByte(); this._align = ...
	l.align = mapListAlign(buf.ReadByte())

	// TypeScript: i1 = buffer.readByte(); this._verticalAlign = ...
	l.verticalAlign = mapListVerticalAlign(buf.ReadByte())

	// TypeScript: this._lineGap = buffer.getInt16();
	l.lineGap = int(buf.ReadInt16())

	// TypeScript: this._columnGap = buffer.getInt16();
	l.columnGap = int(buf.ReadInt16())

	// TypeScript: this._lineCount = buffer.getInt16();
	l.lineCount = int(buf.ReadInt16())

	// TypeScript: this._columnCount = buffer.getInt16();
	l.columnCount = int(buf.ReadInt16())

	// TypeScript: this._autoResizeItem = buffer.readBool();
	l.autoResizeItem = buf.ReadBool()

	// TypeScript: this._childrenRenderOrder = buffer.readByte();
	l.childrenOrder = mapChildrenRenderOrder(buf.ReadByte())

	// TypeScript: this._apexIndex = buffer.getInt16();
	l.apexIndex = int(buf.ReadInt16())

	// TypeScript: if (buffer.readBool()) { ... }
	if buf.ReadBool() {
		l.margin.Top = int(buf.ReadInt32())
		l.margin.Bottom = int(buf.ReadInt32())
		l.margin.Left = int(buf.ReadInt32())
		l.margin.Right = int(buf.ReadInt32())
	}

	// TypeScript: var overflow: number = buffer.readByte();
	overflow := assets.OverflowType(buf.ReadByte())
	l.overflow = overflow

	// TypeScript: if (overflow == OverflowType.Scroll) { setupScroll(buffer); }
	if overflow == assets.OverflowTypeScroll {
		savedPos := buf.Pos()
		if buf.Seek(beginPos, 7) {
			l.GComponent.SetupScroll(buf)
		}
		_ = buf.SetPos(savedPos)
	}

	// TypeScript: if (buffer.readBool()) buffer.skip(8); //clipSoftness
	if buf.ReadBool() && buf.Remaining() >= 8 {
		_ = buf.Skip(8)
	}

	// TypeScript: if (buffer.version >= 2) { ... }
	if buf.Version >= 2 {
		l.scrollToView = buf.ReadBool()
		l.foldInvisible = buf.ReadBool()
	}

	// 读取 defaultItem (block 8)
	if !buf.Seek(beginPos, 8) {
		return
	}
	if def := buf.ReadS(); def != nil && *def != "" {
		l.SetDefaultItem(*def)
	}

	// 读取并创建列表项
	if buf.Remaining() >= 2 {
		l.readItems(buf)
	}
}

// SetupAfterAdd 在组件加入父对象后应用控制器索引等设置。
// 对应TypeScript版本的setup_afterAdd方法(GList.ts:2368-2376)
func (l *GList) SetupAfterAdd(ctx *SetupContext, buf *utils.ByteBuffer) {
	if l == nil || buf == nil {
		return
	}
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()

	// 注意：buf 是 SubBuffer，起始位置为 0，所以 Seek(0, 6) 是正确的
	// 这与 TypeScript 中的 buffer.seek(beginPos, 6) 等价
	if !buf.Seek(0, 6) || buf.Remaining() < 2 {
		return
	}

	index := int(buf.ReadInt16())
	// TypeScript 原版: if (i != -1)
	// 即：只有当索引为 -1 时才跳过设置控制器
	if index != -1 {
		// 安全检查：确保父组件和控制器有效
		if ctx != nil && ctx.Parent != nil {
			controllers := ctx.Parent.Controllers()
			if index >= 0 && index < len(controllers) {
				// TypeScript 原版: this._selectionController = this._parent.getControllerAt(i)
				l.SetSelectionController(controllers[index])
			}
		}
	}

	// 备用布局触发：如果 SetSize 没有被调用（例如 FUI 中尺寸是负数/自动尺寸）
	// 在这里作为最后的备份触发一次布局
	if !l.virtual && len(l.items) > 0 && !l.boundsInitialized {
		// 对于不依赖容器尺寸的布局类型（SingleColumn/SingleRow），即使尺寸是 0 也触发
		// 对于 Flow 布局，如果尺寸是 0，尝试从父组件获取参考尺寸
		canLayout := false
		switch l.layout {
		case ListLayoutTypeSingleColumn, ListLayoutTypeSingleRow:
			canLayout = true // 这些布局不依赖容器尺寸
		case ListLayoutTypeFlowHorizontal, ListLayoutTypeFlowVertical, ListLayoutTypePagination:
			// Flow 布局依赖容器尺寸
			if l.GComponent.Width() > 0 || l.GComponent.Height() > 0 {
				canLayout = true
			} else if ctx != nil && ctx.Parent != nil {
				// 尝试使用父组件的尺寸作为参考
				parentWidth := ctx.Parent.Width()
				parentHeight := ctx.Parent.Height()
				if parentWidth > 0 && parentHeight > 0 {
					// 设置 GList 尺寸为父组件尺寸，触发布局计算
					l.GComponent.SetSize(parentWidth, parentHeight)
					canLayout = true
				}
			}
		}

		if canLayout {
			l.updateBounds()
			l.boundsInitialized = true
		}
	}

	// 关键修复：对于虚拟列表，在 SetupAfterAdd 完成后触发首次刷新
	// 这确保在滚动条创建前，contentSize 已经被正确计算
	// 对应 TypeScript 版本中 virtual list 在初始化后立即刷新的行为
	if l.virtual && l.numItems > 0 {
		// 设置标志，表示需要完整的布局刷新
		l.SetVirtualListChangedFlag(true)
		// 注意：不直接调用 refreshVirtualList，而是设置标志
		// 因为此时 ScrollPane 可能还没有完全初始化
		// 标志会在下一次 CheckVirtualList 时触发刷新
	}
}

// readItems 读取并创建列表项 - 对应TypeScript版本的readItems方法
func (l *GList) readItems(buf *utils.ByteBuffer) {
	if l == nil || buf == nil {
		return
	}

	// 批量添加模式：避免每次 AddItem 都重新计算布局
	l.batchAdding = true
	defer func() {
		l.batchAdding = false
	}()

	cnt := int(buf.ReadInt16())
	for i := 0; i < cnt; i++ {
		if buf.Remaining() < 2 {
			break
		}
		nextPos := int(buf.ReadInt16()) + buf.Pos()

		// 读取项目资源URL
		str := buf.ReadS()
		if str == nil || *str == "" {
			// 如果没有指定资源URL，使用默认项
			str = &l.defaultItem
			if str == nil || *str == "" {
				// 跳过这个项目
				_ = buf.SetPos(nextPos)
				continue
			}
		}

		// 从对象池获取对象
		obj := l.getFromPool(*str)
		if obj != nil {
			// 关键修复：使用AddItem而不是AddChild，确保items数组被正确更新
			l.AddItem(obj)
			// 设置项目属性
			l.setupItem(buf, obj)
		}

		_ = buf.SetPos(nextPos)
	}

	// 注意：不在这里调用 updateBounds()，因为此时组件尺寸还未设置
	// 将在 SetupAfterAdd 中调用
}

// setupItem 设置项目属性 - 对应TypeScript版本的setupItem方法
func (l *GList) setupItem(buf *utils.ByteBuffer, obj *core.GObject) {
	if l == nil || buf == nil || obj == nil {
		return
	}

	// 读取并设置文本
	str := buf.ReadS()
	if str != nil && *str != "" {
		// 设置对象的文本属性
		if textField, ok := obj.Data().(*GTextField); ok {
			textField.SetText(*str)
		} else if button, ok := obj.Data().(*GButton); ok {
			button.SetTitle(*str)
		}
	}

	// 读取并设置选中标题（仅对按钮有效）
	str = buf.ReadS()
	if str != nil && *str != "" {
		if button, ok := obj.Data().(*GButton); ok {
			button.SetSelectedTitle(*str)
		}
	}

	// 读取并设置图标
	str = buf.ReadS()
	if str != nil && *str != "" {
		if button, ok := obj.Data().(*GButton); ok {
			button.SetIcon(*str)
		}
		// 注意：GImage没有SetURL方法，需要通过PackageItem设置
		// 这里暂时只处理按钮的图标设置
	}

	// 读取并设置选中图标（仅对按钮有效）
	str = buf.ReadS()
	if str != nil && *str != "" {
		if button, ok := obj.Data().(*GButton); ok {
			button.SetSelectedIcon(*str)
		}
	}

	// 读取并设置名称
	str = buf.ReadS()
	if str != nil && *str != "" {
		obj.SetName(*str)
	}

	// 设置控制器状态（仅对组件有效）
	if comp, ok := obj.Data().(*core.GComponent); ok {
		// 读取控制器设置
		cnt := int(buf.ReadInt16())
		for i := 0; i < cnt; i++ {
			if buf.Remaining() < 2 {
				break
			}
			ctrlName := buf.ReadS()
			ctrlPage := buf.ReadS()
			if ctrlName != nil && ctrlPage != nil {
				if controller := comp.ControllerByName(*ctrlName); controller != nil {
					controller.SetSelectedPageID(*ctrlPage)
				}
			}
		}

		// 版本2+：设置子对象属性
		if buf.Version >= 2 {
			if buf.Remaining() >= 2 {
				cnt = int(buf.ReadInt16())
				for i := 0; i < cnt; i++ {
					if buf.Remaining() < 6 {
						break
					}
					target := buf.ReadS()
					propertyId := int(buf.ReadInt16())
					value := buf.ReadS()
					if target != nil && value != nil {
						if obj2 := core.FindChildByPath(comp, *target); obj2 != nil {
							// 将int转换为ObjectPropID
							propID := gears.ObjectPropID(propertyId)
							obj2.SetProp(propID, *value)
						}
					}
				}
			}
		}
	}
}

// getFromPool 从对象池获取对象 - 对应TypeScript版本的getFromPool
func (l *GList) getFromPool(url string) *core.GObject {
	if l == nil {
		return nil
	}

	// 优先使用对象池
	if l.pool != nil {
		return l.pool.GetObject(url)
	}

	// 如果没有对象池，使用对象创建器
	if l.creator != nil {
		return l.creator.CreateObject(url)
	}

	return nil
}

// returnToPool 将对象返回到对象池
// 对应 TypeScript 版本的 returnToPool 方法 (GList.ts:220-223)
func (l *GList) returnToPool(obj *core.GObject) {
	if l == nil || obj == nil {
		return
	}

	if l.pool != nil {
		// TypeScript: obj.displayObject.cacheAs = "none";
		// 在Go版本中，我们不需要设置cacheAs
		// TypeScript: this._pool.returnObject(obj);
		l.pool.ReturnObject(obj)
	}
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

// =============== 虚拟化相关方法 ===============

// SetVirtual 设置是否启用虚拟化
// 对应 TypeScript 版本的 _setVirtual 方法 (GList.ts:964-1008)
func (l *GList) SetVirtual(value bool) {
	if l.virtual == value {
		return
	}

	l.virtual = value
	if value {
		// 启用虚拟化
		log.Printf("🔧 SetVirtual(true) 开始 - 列表名称:%s", l.GComponent.GObject.Name())
		log.Printf("   状态检查: creator=%v, pool=%v, defaultItem=%s",
			l.creator != nil, l.pool != nil, l.defaultItem)

		// TypeScript: if (this._scrollPane == null) throw new Error("Virtual list must be scrollable!");
		scrollPane := l.GComponent.ScrollPane()
		if scrollPane == nil {
			log.Printf("⚠️  警告: 虚拟列表必须可滚动! 将创建ScrollPane")
			log.Printf("   列表尺寸: %.0fx%.0f", l.GComponent.Width(), l.GComponent.Height())
			// 创建默认的垂直滚动面板
			scrollPane = l.GComponent.EnsureScrollPane(core.ScrollTypeVertical)
			log.Printf("   ScrollPane创建后视图尺寸: %.0fx%.0f", scrollPane.ViewWidth(), scrollPane.ViewHeight())
		} else {
			// ScrollPane已存在，但需要刷新视图尺寸
			// 这是关键修复：FUI加载的ScrollPane的viewSize可能未正确初始化
			log.Printf("   ScrollPane已存在，当前视图尺寸: %.0fx%.0f", scrollPane.ViewWidth(), scrollPane.ViewHeight())
			scrollPane.SetViewSize(l.GComponent.Width(), l.GComponent.Height())
			log.Printf("   刷新后视图尺寸: %.0fx%.0f", scrollPane.ViewWidth(), scrollPane.ViewHeight())
		}

		// 移除所有子对象到对象池
		// TypeScript: this.removeChildrenToPool();
		children := l.GComponent.Children()
		log.Printf("   移除 %d 个现有子对象", len(children))
		for _, child := range children {
			if child != nil {
				l.GComponent.RemoveChild(child)
			}
		}

		// 初始化虚拟项数组
		if l.virtualItems == nil {
			l.virtualItems = make([]*ItemInfo, 0)
		}

		// 自动初始化 itemSize
		// TypeScript: if (this._itemSize == null) { ... }
		log.Printf("   itemSize检查: %v (X=%.0f, Y=%.0f)",
			l.itemSize,
			func() float64 {
				if l.itemSize != nil {
					return l.itemSize.X
				}
				return 0
			}(),
			func() float64 {
				if l.itemSize != nil {
					return l.itemSize.Y
				}
				return 0
			}())

		if l.itemSize == nil || (l.itemSize.X == 0 && l.itemSize.Y == 0) {
			if l.itemSize == nil {
				l.itemSize = &laya.Point{}
			}

			// 从对象池获取默认对象来测量尺寸
			// TypeScript: var obj: GObject = this.getFromPool(null);
			log.Printf("   尝试从池中获取对象: url=%s", l.defaultItem)
			obj := l.getFromPool(l.defaultItem)
			if obj == nil {
				log.Printf("❌ 错误: 无法获取默认列表项! defaultItem=%s, creator=%v, pool=%v",
					l.defaultItem, l.creator != nil, l.pool != nil)
				// 使用默认尺寸避免崩溃
				l.itemSize.X = 100
				l.itemSize.Y = 30
			} else {
				// TypeScript: this._itemSize.x = obj.width; this._itemSize.y = obj.height;
				l.itemSize.X = obj.Width()
				l.itemSize.Y = obj.Height()
				log.Printf("✅ 成功测量itemSize: %.0fx%.0f", l.itemSize.X, l.itemSize.Y)
				// TypeScript: this.returnToPool(obj);
				l.returnToPool(obj)
			}
		} else {
			log.Printf("   itemSize已存在，跳过初始化")
		}

		// 设置滚动步长
		// TypeScript: if (this._layout == ListLayoutType.SingleColumn || this._layout == ListLayoutType.FlowHorizontal)
		if scrollPane != nil {
			if l.layout == ListLayoutTypeSingleColumn || l.layout == ListLayoutTypeFlowHorizontal {
				// TypeScript: this._scrollPane.scrollStep = this._itemSize.y;
				scrollPane.SetScrollStep(l.itemSize.Y)
			} else {
				// TypeScript: this._scrollPane.scrollStep = this._itemSize.x;
				scrollPane.SetScrollStep(l.itemSize.X)
			}

			// 关键修复：注册滚动事件监听器
			// 对应 TypeScript 版本 GList.ts:1005 - this.on(Events.SCROLL, this, this.__scrolled);
			// 当滚动时调用 handleScroll(false) 来更新可见的虚拟项
			if l.scrollListenerID != 0 {
				scrollPane.RemoveScrollListener(l.scrollListenerID)
			}
			l.scrollListenerID = scrollPane.AddScrollListener(func(info core.ScrollInfo) {
				// 对应 TypeScript 的 __scrolled 回调 (GList.ts:1207-1209)
				// private __scrolled(evt: Laya.Event): void {
				//     this.handleScroll(false);
				// }
				if !l.eventLocked {
					l.handleScroll(false)
				}
			})
			log.Printf("   ✅ 已注册滚动监听器 ID=%d", l.scrollListenerID)
		}

		// 设置虚拟列表改变标记
		l.SetVirtualListChangedFlag(true)
	} else {
		// 禁用虚拟化
		// 取消注册滚动监听器
		if l.scrollListenerID != 0 {
			if scrollPane := l.GComponent.ScrollPane(); scrollPane != nil {
				scrollPane.RemoveScrollListener(l.scrollListenerID)
				log.Printf("   ✅ 已取消滚动监听器 ID=%d", l.scrollListenerID)
			}
			l.scrollListenerID = 0
		}

		l.ClearVirtualItems()
		children := l.GComponent.Children()
		for _, child := range children {
			if child != nil {
				l.GComponent.RemoveChild(child)
			}
		}
		// 重新添加现有的items并恢复事件监听
		for _, item := range l.items {
			if item != nil {
				l.GComponent.AddChild(item)
				l.attachItemClick(item)
			}
		}
	}
}

// IsVirtual 返回是否启用虚拟化
func (l *GList) IsVirtual() bool {
	return l.virtual
}

// SetLoop 设置是否循环（仅在虚拟模式下有效）
func (l *GList) SetLoop(value bool) {
	if !l.virtual {
		return
	}
	if l.loop != value {
		l.loop = value
		l.SetVirtualListChangedFlag(true)
		// 设置 ScrollPane 的循环模式
		// 1=水平循环, 2=垂直循环
		scrollPane := l.GComponent.ScrollPane()
		if scrollPane != nil {
			if l.layout == ListLayoutTypeSingleColumn || l.layout == ListLayoutTypeFlowHorizontal {
				// 垂直循环
				if value {
					scrollPane.SetLoop(2)
				} else {
					scrollPane.SetLoop(0)
				}
			} else if l.layout == ListLayoutTypeSingleRow || l.layout == ListLayoutTypeFlowVertical {
				// 水平循环
				if value {
					scrollPane.SetLoop(1)
				} else {
					scrollPane.SetLoop(0)
				}
			}
		}
	}
}

// IsLoop 返回是否循环
func (l *GList) IsLoop() bool {
	return l.loop
}

// ChildrenCount returns the number of child objects in the list.
func (l *GList) ChildrenCount() int {
	if l == nil || l.GComponent == nil {
		return 0
	}
	return len(l.GComponent.Children())
}

// NumItems 返回数据项总数（虚拟模式）
func (l *GList) NumItems() int {
	if l.virtual {
		return l.numItems
	}
	return len(l.items)
}

// SetNumItems 设置数据项总数（虚拟模式）
func (l *GList) SetNumItems(value int) {
	if !l.virtual {
		return
	}

	if l.numItems == value {
		return
	}

	l.numItems = value
	if l.loop {
		l.realNumItems = l.numItems * 6 // 设置6倍数量，用于循环滚动
	} else {
		l.realNumItems = l.numItems
	}

	// 重置选择状态
	l.ClearSelection()

	// 确保虚拟项数组大小
	l.EnsureVirtualItems(l.realNumItems)

	// 设置虚拟列表改变标记
	l.SetVirtualListChangedFlag(true)
}

// SetItemRenderer 设置项目渲染器
func (l *GList) SetItemRenderer(renderer func(index int, item *core.GObject)) {
	l.itemRenderer = renderer
}

// SetItemProvider 设置项目提供者
func (l *GList) SetItemProvider(provider func(index int) string) {
	l.itemProvider = provider
}

// VirtualItemSize 返回虚拟项目尺寸
func (l *GList) VirtualItemSize() *laya.Point {
	return l.itemSize
}

// SetVirtualItemSize 设置虚拟项目尺寸
func (l *GList) SetVirtualItemSize(value *laya.Point) {
	if !l.virtual {
		return
	}
	if value != nil {
		l.itemSize.X = value.X
		l.itemSize.Y = value.Y
		l.SetVirtualListChangedFlag(true)
	}
}

// SetObjectCreator 设置对象创建器，用于动态创建对象
func (l *GList) SetObjectCreator(creator ObjectCreator) {
	l.creator = creator
	if l.pool != nil {
		l.pool.creator = creator
	}
}

// GetObjectCreator 返回对象创建器
func (l *GList) GetObjectCreator() ObjectCreator {
	if l == nil {
		return nil
	}
	return l.creator
}

// RefreshVirtualList 刷新虚拟列表
func (l *GList) RefreshVirtualList() {
	if !l.virtual {
		// 如果不是虚拟列表，不做任何操作
		return
	}

	// 这里是关键问题！原始代码只是设置了一个标志，但没有实际触发刷新
	// 现在修改为直接调用刷新方法
	l.SetVirtualListChangedFlag(false) // 传入false表示内容改变而非布局改变
	l.refreshVirtualList()             // 直接调用刷新方法
}

// RefreshVirtualListPublic 刷新虚拟列表（公共方法，兼容旧代码）
func (l *GList) RefreshVirtualListPublic() {
	l.RefreshVirtualList()
}

// =============== 内部辅助方法 ===============

// SetVirtualListChangedFlag 设置虚拟列表改变标记
func (l *GList) SetVirtualListChangedFlag(layoutChanged bool) {
	if !l.virtual {
		return
	}

	if layoutChanged {
		l.virtualListChanged = 2
	} else if l.virtualListChanged == 0 {
		l.virtualListChanged = 1
	}

	// 延迟刷新，避免频繁更新
	// 这里简化处理，直接刷新
	l.refreshVirtualList()
}

// CheckVirtualList 检查虚拟列表是否需要刷新
func (l *GList) CheckVirtualList() bool {
	if !l.virtual {
		return false
	}

	if l.virtualListChanged != 0 {
		l.refreshVirtualList()
		return true
	}
	return false
}

// EnsureVirtualItems 确保虚拟项数组大小
func (l *GList) EnsureVirtualItems(count int) {
	currentCount := len(l.virtualItems)
	if currentCount < count {
		// 扩展数组
		for i := currentCount; i < count; i++ {
			l.virtualItems = append(l.virtualItems, &ItemInfo{})
		}
	} else if currentCount > count {
		// 缩容数组
		l.virtualItems = l.virtualItems[:count]
	}
}

// ClearVirtualItems 清空虚拟项
func (l *GList) ClearVirtualItems() {
	if l.pool != nil {
		// 将所有对象返回池中
		for _, ii := range l.virtualItems {
			if ii != nil && ii.obj != nil {
				l.pool.ReturnObject(ii.obj)
				ii.obj = nil
			}
		}
	}
	l.virtualItems = make([]*ItemInfo, 0)
}

// ResetAllUpdateFlags 重置所有更新标记
func (l *GList) ResetAllUpdateFlags() {
	l.itemInfoVer++
}

// IsItemUpdated 检查项是否已更新
func (l *GList) IsItemUpdated(item *ItemInfo) bool {
	return item.updateFlag == l.itemInfoVer
}

// MarkItemUpdated 标记项为已更新
func (l *GList) MarkItemUpdated(item *ItemInfo) {
	item.updateFlag = l.itemInfoVer
}

// OwnerSizeChanged 处理所有者尺寸变化 - 用于虚拟列表刷新
func (l *GList) OwnerSizeChanged(oldWidth, oldHeight float64) {
	if l.virtual {
		// 当列表尺寸变化时，刷新虚拟列表
		l.SetVirtualListChangedFlag(true)
	}
}

// updateBounds 计算并更新非虚拟列表中所有子元素的位置
// 对应 TypeScript 版本的 GList.updateBounds() 方法 (GList.ts:1919-2100)
func (l *GList) updateBounds() {
	if l == nil || l.virtual {
		return
	}

	var curX, curY float64 = 0, 0
	var maxWidth, maxHeight float64 = 0, 0
	var cw, ch float64
	var j int = 0 // 当前行/列的项目计数

	cnt := len(l.items)
	if cnt == 0 {
		l.setBounds(0, 0, 0, 0)
		return
	}

	// 使用视口尺寸，而不是组件总尺寸
	// 对于有滚动的列表，这是滚动面板的可视区域
	viewWidth := l.GComponent.ViewWidth()
	viewHeight := l.GComponent.ViewHeight()

	// 关键修复：如果 ViewWidth/ViewHeight 返回 0，使用组件尺寸作为 fallback
	// 这处理了 ScrollPane 的 ViewSize 还未初始化的情况
	if viewWidth == 0 {
		viewWidth = l.GComponent.Width()
	}
	if viewHeight == 0 {
		viewHeight = l.GComponent.Height()
	}

	// 处理容器尺寸为 0 的情况
	// - SingleColumn/SingleRow: 不依赖容器尺寸，可以直接布局
	// - Flow 布局: 需要容器尺寸来判断换行/换列，如果是 0 则暂时跳过
	if viewWidth == 0 || viewHeight == 0 {
		switch l.layout {
		case ListLayoutTypeSingleColumn, ListLayoutTypeSingleRow:
			// 这些布局不依赖容器尺寸，允许继续
			if viewWidth == 0 {
				viewWidth = 1 // 避免除零错误
			}
			if viewHeight == 0 {
				viewHeight = 1
			}
		case ListLayoutTypeFlowHorizontal, ListLayoutTypeFlowVertical, ListLayoutTypePagination:
			// Flow 布局依赖容器尺寸，如果尺寸无效则跳过
			log.Printf("⚠️  updateBounds: Flow布局需要有效的容器尺寸，当前 viewSize=%.0fx%.0f，跳过布局",
				viewWidth, viewHeight)
			return
		}
	}

	switch l.layout {
	case ListLayoutTypeSingleColumn:
		// 单列垂直布局
		for i := 0; i < cnt; i++ {
			child := l.items[i]
			if child == nil {
				continue
			}
			if l.foldInvisible && !child.Visible() {
				continue
			}

			if curY != 0 {
				curY += float64(l.lineGap)
			}
			child.SetPosition(child.X(), curY)
			if l.autoResizeItem {
				child.SetSize(viewWidth, child.Height())
			}
			curY += math.Ceil(child.Height())
			if child.Width() > maxWidth {
				maxWidth = child.Width()
			}
		}
		ch = curY
		cw = math.Ceil(maxWidth)

	case ListLayoutTypeSingleRow:
		// 单行水平布局
		for i := 0; i < cnt; i++ {
			child := l.items[i]
			if child == nil {
				continue
			}
			if l.foldInvisible && !child.Visible() {
				continue
			}

			if curX != 0 {
				curX += float64(l.columnGap)
			}
			child.SetPosition(curX, child.Y())
			if l.autoResizeItem {
				child.SetSize(child.Width(), viewHeight)
			}
			curX += math.Ceil(child.Width())
			if child.Height() > maxHeight {
				maxHeight = child.Height()
			}
		}
		cw = curX
		ch = math.Ceil(maxHeight)

	case ListLayoutTypeFlowHorizontal:
		// 多列流动布局（水平方向）
		// 对应 TypeScript: else 分支 (GList.ts:2044-2071)
		for i := 0; i < cnt; i++ {
			child := l.items[i]
			if child == nil {
				continue
			}
			if l.foldInvisible && !child.Visible() {
				continue
			}

			if curX != 0 {
				curX += float64(l.columnGap)
			}

			// 检查是否需要换行
			// TypeScript: if (this._columnCount != 0 && j >= this._columnCount
			//             || this._columnCount == 0 && curX + child.width > viewWidth && maxHeight != 0)
			willWrap := false
			if l.columnCount != 0 && j >= l.columnCount {
				willWrap = true
			} else if l.columnCount == 0 && curX+child.Width() > viewWidth && maxHeight != 0 {
				willWrap = true
			}

			if willWrap {
				// 换行
				curX = 0
				curY += math.Ceil(maxHeight) + float64(l.lineGap)
				maxHeight = 0
				j = 0
			}

			child.SetPosition(curX, curY)
			curX += math.Ceil(child.Width())
			if curX > maxWidth {
				maxWidth = curX
			}
			if child.Height() > maxHeight {
				maxHeight = child.Height()
			}
			j++
		}
		ch = curY + math.Ceil(maxHeight)
		cw = math.Ceil(maxWidth)

	case ListLayoutTypeFlowVertical:
		// 多行流动布局（垂直方向）
		// 对应 TypeScript: else 分支 (GList.ts:2113-2139)
		for i := 0; i < cnt; i++ {
			child := l.items[i]
			if child == nil {
				continue
			}
			if l.foldInvisible && !child.Visible() {
				continue
			}

			if curY != 0 {
				curY += float64(l.lineGap)
			}

			// 检查是否需要换列
			// TypeScript: if (this._lineCount != 0 && j >= this._lineCount
			//             || this._lineCount == 0 && curY + child.height > viewHeight && maxWidth != 0)
			if l.lineCount != 0 && j >= l.lineCount ||
				l.lineCount == 0 && curY+child.Height() > viewHeight && maxWidth != 0 {
				// 换列
				curY = 0
				curX += math.Ceil(maxWidth) + float64(l.columnGap)
				maxWidth = 0
				j = 0
			}

			child.SetPosition(curX, curY)
			curY += math.Ceil(child.Height())
			if curY > maxHeight {
				maxHeight = curY
			}
			if child.Width() > maxWidth {
				maxWidth = child.Width()
			}
			j++
		}
		cw = curX + math.Ceil(maxWidth)
		ch = math.Ceil(maxHeight)

	case ListLayoutTypePagination:
		// 分页布局
		// 简化处理，暂时按照流动布局处理
		for i := 0; i < cnt; i++ {
			child := l.items[i]
			if child == nil {
				continue
			}
			if l.foldInvisible && !child.Visible() {
				continue
			}

			if curX != 0 {
				curX += float64(l.columnGap)
			}

			if l.columnCount != 0 && j >= l.columnCount {
				curX = 0
				curY += math.Ceil(maxHeight) + float64(l.lineGap)
				maxHeight = 0
				j = 0
			}

			child.SetPosition(curX, curY)
			curX += math.Ceil(child.Width())
			if curX > maxWidth {
				maxWidth = curX
			}
			if child.Height() > maxHeight {
				maxHeight = child.Height()
			}
			j++
		}
		ch = curY + math.Ceil(maxHeight)
		cw = math.Ceil(maxWidth)
	}

	l.setBounds(0, 0, cw, ch)
}

// setBounds 设置列表内容边界
// 对应 TypeScript 版本的 GComponent.setBounds() 方法
func (l *GList) setBounds(ax, ay, aw, ah float64) {
	if l == nil {
		return
	}

	// 如果有滚动面板，更新内容尺寸
	if pane := l.GComponent.ScrollPane(); pane != nil {
		pane.SetContentSize(ax+aw, ay+ah)
	}
}
