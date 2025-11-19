package fairygui

import (
	"math"
	"sync"
)

// ============================================================================
// 基础类型定义
// ============================================================================

// AlignType - 对齐类型
type AlignType int

const (
	AlignTypeLeft AlignType = iota
	AlignTypeCenter
	AlignTypeRight
)

// VertAlignType - 垂直对齐类型
type VertAlignType int

const (
	VertAlignTypeTop VertAlignType = iota
	VertAlignTypeMiddle
	VertAlignTypeBottom
)

// ChildrenRenderOrder - 子对象渲染顺序
type ChildrenRenderOrder int

const (
	ChildrenRenderOrderAscent ChildrenRenderOrder = iota
	ChildrenRenderOrderDescent
	ChildrenRenderOrderArch
)

// ============================================================================
// ListLayoutType - 列表布局类型
// ============================================================================

type ListLayoutType int

const (
	ListLayoutTypeSingleColumn ListLayoutType = iota
	ListLayoutTypeSingleRow
	ListLayoutTypeFlowHorizontal
	ListLayoutTypeFlowVertical
	ListLayoutTypePagination
)

// ============================================================================
// ListSelectionMode - 列表选择模式
// ============================================================================

type ListSelectionMode int

const (
	ListSelectionModeSingle ListSelectionMode = iota
	ListSelectionModeMultiple
	ListSelectionModeMultipleSingleClick
	ListSelectionModeNone
)

// ============================================================================
// List - 列表控件 V2 (基于新架构)
// ============================================================================

type List struct {
	*ComponentImpl

	// 资源相关
	packageItem *PackageItemWrapper
	template    *ComponentImpl
	defaultItem string

	// 子对象
	scrollBar *ScrollBar

	// 布局参数
	layout               ListLayoutType
	lineCount            int
	columnCount          int
	lineGap              float64
	columnGap            float64
	autoResizeItem       bool
	selectionMode        ListSelectionMode
	align                AlignType
	verticalAlign        VertAlignType

	// 虚拟列表
	virtual              bool
	realNumItems         int
	loop                 bool

	// 项管理
	itemInfoVer          uint
	numItems             int
	refreshAllItems      bool
	itemPool             *ListItemPool
	itemProvider         ListItemProvider
	itemRenderer         ListItemRenderer
	childrenRenderOrder  ChildrenRenderOrder
	apexIndex            int

	// 选择状态
	selectedIndex        int
	selection            []int
	lastSelectedIndex    int
	selectionHandled     bool

	// 滚动状态
	scrollPane           *ScrollPaneV2
	scrollItemToViewOnClick bool

	// 事件
	onClickItem          EventHandler
	onRightClickItem     EventHandler

	// 内部状态
	itemInfoMutex        sync.RWMutex
	eventMutex           sync.Mutex
}

// ListItemProvider - 项提供者函数类型
type ListItemProvider func(index int) string

// ListItemRenderer - 项渲染器函数类型
type ListItemRenderer func(index int, item *ComponentImpl)

// NewList 创建新的列表
func NewList() *List {
	list := &List{
		ComponentImpl:          NewComponent(),
		layout:                 ListLayoutTypeSingleColumn,
		selectionMode:          ListSelectionModeSingle,
		align:                  AlignTypeLeft,
		verticalAlign:          VertAlignTypeTop,
		lineGap:                0,
		columnGap:              0,
		autoResizeItem:         true,
		scrollItemToViewOnClick: true,
		selectedIndex:          -1,
		lastSelectedIndex:      -1,
		childrenRenderOrder:    ChildrenRenderOrderAscent,
		itemPool:               NewListItemPool(),
	}

	list.scrollPane = NewScrollPaneV2()
	list.AddChild(list.scrollPane)

	return list
}

// ============================================================================
// 资源相关
// ============================================================================

// SetPackageItem 设置资源项
func (list *List) SetPackageItem(item PackageItem) {
	if item == nil {
		list.packageItem = nil
		return
	}
	if wrapper, ok := item.(*PackageItemWrapper); ok {
		list.packageItem = wrapper
	}
}

// PackageItem 返回资源项
func (list *List) PackageItem() PackageItem {
	return list.packageItem
}

// SetTemplateComponent 设置模板组件
func (list *List) SetTemplateComponent(comp *ComponentImpl) {
	if list.template != nil {
		list.RemoveChild(list.template)
	}
	list.template = comp
	if comp != nil {
		comp.SetPosition(0, 0)
		list.AddChild(comp)
	}
	list.resolveTemplate()
}

// TemplateComponent 返回模板组件
func (list *List) TemplateComponent() *ComponentImpl {
	return list.template
}

// SetDefaultItem 设置默认项资源
func (list *List) SetDefaultItem(val string) {
	list.defaultItem = val
}

// DefaultItem 返回默认项资源
func (list *List) DefaultItem() string {
	return list.defaultItem
}

// ============================================================================
// 布局设置
// ============================================================================

// SetLayout 设置布局类型
func (list *List) SetLayout(layout ListLayoutType) *List {
	if list.layout == layout {
		return list
	}
	list.layout = layout
	list.setBoundsChangedFlag()
	return list
}

// Layout 返回布局类型
func (list *List) Layout() ListLayoutType {
	return list.layout
}

// SetLineCount 设置行数
func (list *List) SetLineCount(value int) *List {
	if list.lineCount == value {
		return list
	}
	list.lineCount = value
	list.setBoundsChangedFlag()
	return list
}

// LineCount 返回行数
func (list *List) LineCount() int {
	return list.lineCount
}

// SetColumnCount 设置列数
func (list *List) SetColumnCount(value int) *List {
	if list.columnCount == value {
		return list
	}
	list.columnCount = value
	list.setBoundsChangedFlag()
	return list
}

// ColumnCount 返回列数
func (list *List) ColumnCount() int {
	return list.columnCount
}

// SetLineGap 设置行间距
func (list *List) SetLineGap(value float64) *List {
	if list.lineGap == value {
		return list
	}
	list.lineGap = value
	list.setBoundsChangedFlag()
	return list
}

// LineGap 返回行间距
func (list *List) LineGap() float64 {
	return list.lineGap
}

// SetColumnGap 设置列间距
func (list *List) SetColumnGap(value float64) *List {
	if list.columnGap == value {
		return list
	}
	list.columnGap = value
	list.setBoundsChangedFlag()
	return list
}

// ColumnGap 返回列间距
func (list *List) ColumnGap() float64 {
	return list.columnGap
}

// SetAlignment 设置对齐方式
func (list *List) SetAlignment(value AlignType) *List {
	if list.align == value {
		return list
	}
	list.align = value
	list.setBoundsChangedFlag()
	return list
}

// Alignment 返回对齐方式
func (list *List) Alignment() AlignType {
	return list.align
}

// SetVerticalAlignment 设置垂直对齐方式
func (list *List) SetVerticalAlignment(value VertAlignType) *List {
	if list.verticalAlign == value {
		return list
	}
	list.verticalAlign = value
	list.setBoundsChangedFlag()
	return list
}

// VerticalAlignment 返回垂直对齐方式
func (list *List) VerticalAlignment() VertAlignType {
	return list.verticalAlign
}

// SetAutoResizeItem 设置是否自动调整项大小
func (list *List) SetAutoResizeItem(value bool) *List {
	list.autoResizeItem = value
	list.setBoundsChangedFlag()
	return list
}

// AutoResizeItem 返回是否自动调整项大小
func (list *List) AutoResizeItem() bool {
	return list.autoResizeItem
}

// ============================================================================
// 项管理
// ============================================================================

// SetNumItems 设置项数量
func (list *List) SetNumItems(value int) *List {
	if list.virtual && list.itemRenderer == nil {
		list.numItems = value
		return list
	}

	if list.itemInfoVer != 0 {
		list.itemInfoVer++
	}

	if list.loop && value > 0 {
		value = int(math.Floor(float64(value+list.lineCount-1)/float64(list.lineCount)) * float64(list.lineCount))
	}

	list.numItems = value
	list.itemInfoMutex.Lock()
	list.refreshAllItems = true
	list.itemInfoMutex.Unlock()

	list.removeChildrenToPool(0, -1)

	if list.virtual && !list.scrollPane.isNone {
		list.scrollPane.Clear()
	}

	list.setBoundsChangedFlag()
	return list
}

// NumItems 返回项数量
func (list *List) NumItems() int {
	return list.numItems
}

// RefreshVirtualList 刷新虚拟列表
func (list *List) RefreshVirtualList() *List {
	if list.itemInfoVer != 0 {
		list.itemInfoVer++
	}
	list.itemInfoMutex.Lock()
	list.refreshAllItems = true
	list.itemInfoMutex.Unlock()

	list.setBoundsChangedFlag()
	return list
}

// GetItemAt 获取指定索引的项
func (list *List) GetItemAt(index int) *ComponentImpl {
	if index >= 0 && index < list.numChildren() {
		child := list.GetChildAt(index)
		if comp, ok := child.(*ComponentImpl); ok {
			return comp
		}
	}
	return nil
}

// AddItemFromPool 从对象池添加项
func (list *List) AddItemFromPool(url string) *ComponentImpl {
	if url == "" {
		url = list.defaultItem
	}

	obj := list.itemPool.Get(url)
	list.addChild(obj)
	return obj
}

// RemoveChildToPoolAt 移除子对象到对象池
func (list *List) RemoveChildToPoolAt(index int) {
	child := list.GetChildAt(index)
	list.removeChildToPool(child)
}

// RemoveChildToPool 移除子对象到对象池
func (list *List) RemoveChildToPool(child DisplayObject) {
	list.removeChildToPool(child)
}

// removeChildToPool 内部移除子对象到对象池
func (list *List) removeChildToPool(child DisplayObject) {
	if child == nil {
		return
	}

	list.RemoveChild(child)

	if child.Parent() != nil {
		return
	}

	list.itemPool.Put(child)
}

// RemoveChildrenToPool 移除所有子对象到对象池
func (list *List) RemoveChildrenToPool() {
	list.removeChildrenToPool(0, -1)
}

// removeChildrenToPool 内部移除子对象到对象池
func (list *List) removeChildrenToPool(beginIndex int, endIndex int) {
	if endIndex < 0 || endIndex >= list.numChildren() {
		endIndex = list.numChildren() - 1
	}

	for i := beginIndex; i <= endIndex; i++ {
		list.RemoveChildToPoolAt(beginIndex)
	}
}

// ============================================================================
// 选择管理
// ============================================================================

// SetSelectionMode 设置选择模式
func (list *List) SetSelectionMode(value ListSelectionMode) *List {
	list.selectionMode = value
	return list
}

// SelectionMode 返回选择模式
func (list *List) SelectionMode() ListSelectionMode {
	return list.selectionMode
}

// SetSelectedIndex 设置选中索引
func (list *List) SetSelectedIndex(value int) *List {
	if list.selectionMode == ListSelectionModeNone {
		return list
	}

	if list.selectedIndex == value {
		return list
	}

	list.clearSelection()
	if value >= 0 && value < list.numItems {
		list.AddSelection(value, false)
	}

	return list
}

// SelectedIndex 返回选中索引
func (list *List) SelectedIndex() int {
	return list.selectedIndex
}

// GetSelection 获取选中项列表
func (list *List) GetSelection() []int {
	return list.selection
}

// AddSelection 添加选中项
func (list *List) AddSelection(index int, scrollToView bool) *List {
	if list.selectionMode == ListSelectionModeNone {
		return list
	}

	if list.selectionMode == ListSelectionModeSingle {
		list.clearSelection()
		list.selectedIndex = index
		list.selection = []int{index}
	} else {
		if !list.isChildInView(index) {
			list.selection = append(list.selection, index)
		}
		list.selectedIndex = index
	}

	if scrollToView && list.scrollItemToViewOnClick {
		list.ScrollToView(index)
	}

	return list
}

// RemoveSelection 移除选中项
func (list *List) RemoveSelection(index int) *List {
	if list.selectionMode == ListSelectionModeNone {
		return list
	}

	for i, sel := range list.selection {
		if sel == index {
			list.selection = append(list.selection[:i], list.selection[i+1:]...)
			break
		}
	}

	if list.selectedIndex == index {
		if len(list.selection) > 0 {
			list.selectedIndex = list.selection[0]
		} else {
			list.selectedIndex = -1
		}
	}

	return list
}

// ClearSelection 清除选择
func (list *List) ClearSelection() *List {
	list.clearSelection()
	return list
}

// clearSelection 内部清除选择
func (list *List) clearSelection() {
	list.selectedIndex = -1
	list.selection = []int{}
}

// SelectAll 全选
func (list *List) SelectAll() *List {
	if list.selectionMode == ListSelectionModeNone || list.selectionMode == ListSelectionModeSingle {
		return list
	}

	list.clearSelection()
	for i := 0; i < list.numItems; i++ {
		list.selection = append(list.selection, i)
	}

	if len(list.selection) > 0 {
		list.selectedIndex = list.selection[0]
	}

	return list
}

// SelectReverse 反选
func (list *List) SelectReverse() *List {
	if list.selectionMode == ListSelectionModeNone || list.selectionMode == ListSelectionModeSingle {
		return list
	}

	var temp []int
	for i := 0; i < list.numItems; i++ {
		found := false
		for _, sel := range list.selection {
			if sel == i {
				found = true
				break
			}
		}
		if !found {
			temp = append(temp, i)
		}
	}

	list.selection = temp
	if len(list.selection) > 0 {
		list.selectedIndex = list.selection[0]
	} else {
		list.selectedIndex = -1
	}

	return list
}

// HandleControllerChanged 处理控制器改变
func (list *List) HandleControllerChanged(c Controller) {
	list.selectedIndex = c.SelectedIndex()
	list.selection = []int{list.selectedIndex}
}

// ============================================================================
// 滚动相关
// ============================================================================

// EnsureBoundsCorrect 确保边界正确
func (list *List) EnsureBoundsCorrect() {
	list.ensureBoundsCorrect()
}

// ScrollToView 滚动到指定项
func (list *List) ScrollToView(index int) {
	list.scrollToView(index, false)
}

// scrollToView 内部滚动到指定项
func (list *List) scrollToView(index int, animated bool) {
	if list.scrollPane == nil || !list.scrollPane.isDragged {
		return
	}

	// 确保在范围内
	if index < 0 || index >= list.numItems {
		return
	}

	list.scrollPane.ScrollToView(index, animated)
}

// GetFirstChildInView 获取视图中第一个子对象
func (list *List) GetFirstChildInView() int {
	if list.scrollPane != nil {
		return list.scrollPane.GetFirstChildInView()
	}
	return 0
}

// isChildInView 判断子对象是否在视图中
func (list *List) isChildInView(index int) bool {
	for _, sel := range list.selection {
		if sel == index {
			return true
		}
	}
	return false
}

// ============================================================================
// 内部方法
// ============================================================================

// resolveTemplate 解析模板
func (list *List) resolveTemplate() {
	if list.template == nil {
		return
	}

	// 解析滚动条
	if child := list.template.GetChildByName("scrollBar"); child != nil {
		if sb, ok := child.(*ScrollBar); ok {
			list.scrollBar = sb
		}
	}

	// 绑定滚动条事件
	if list.scrollBar != nil {
		list.scrollBar.On("scroll", func(event Event) {
			list.onScroll(event)
		})
	}
}

// setBoundsChangedFlag 设置边界改变标志
func (list *List) setBoundsChangedFlag() {
	if list.scrollPane != nil {
		list.scrollPane.SetBoundsChanged()
	}
}

// ensureBoundsCorrect 确保边界正确
func (list *List) ensureBoundsCorrect() {
	if list.scrollPane != nil {
		list.scrollPane.EnsureSizeCorrect()
	}
}

// numChildren 返回子对象数量
func (list *List) numChildren() int {
	return list.template.NumChildren()
}

// addChild 添加子对象
func (list *List) addChild(child DisplayObject) {
	if list.template != nil {
		list.template.AddChild(child)
	}
}

// GetChildAt 获取指定索引的子对象
func (list *List) GetChildAt(index int) DisplayObject {
	if list.template != nil {
		return list.template.GetChildAt(index)
	}
	return nil
}

// RemoveChild 移除子对象
func (list *List) RemoveChild(child DisplayObject) {
	if list.template != nil {
		list.template.RemoveChild(child)
	}
}

// onScroll 滚动事件处理
func (list *List) onScroll(event Event) {
	if list.itemRenderer == nil || list.virtual {
		return
	}

	list.itemInfoMutex.Lock()
	refreshAll := list.refreshAllItems
	list.itemInfoMutex.Unlock()

	if refreshAll {
		list.updateVirtualList()
	}
}

// updateVirtualList 更新虚拟列表
func (list *List) updateVirtualList() {
	if !list.virtual || list.itemRenderer == nil {
		return
	}

	list.itemInfoMutex.Lock()
	list.refreshAllItems = false
	list.itemInfoMutex.Unlock()

	// 获取可见范围
	scrollPos := list.scrollPane.GetScrollPos()
	viewSize := list.scrollPane.GetViewSize()

	// 计算需要显示的项
	startIndex := 0
	endIndex := list.numItems

	if list.layout == ListLayoutTypeSingleColumn || list.layout == ListLayoutTypeFlowVertical {
		// 垂直滚动
		itemHeight := list.estimateItemHeight()
		startIndex = int(scrollPos.Y / (itemHeight + list.lineGap))
		endIndex = int((scrollPos.Y + viewSize.Y) / (itemHeight + list.lineGap)) + 1
	} else if list.layout == ListLayoutTypeSingleRow || list.layout == ListLayoutTypeFlowHorizontal {
		// 水平滚动
		itemWidth := list.estimateItemWidth()
		startIndex = int(scrollPos.X / (itemWidth + list.columnGap))
		endIndex = int((scrollPos.X + viewSize.X) / (itemWidth + list.columnGap)) + 1
	}

	// 限制范围
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > list.numItems {
		endIndex = list.numItems
	}

	// 移除不需要的项
	for i := 0; i < list.numChildren(); i++ {
		child := list.GetChildAt(i)
		if comp, ok := child.(*ComponentImpl); ok {
			idx := comp.GetDataInt()
			if idx < startIndex || idx >= endIndex {
				list.removeChildToPool(child)
				i--
			}
		}
	}

	// 添加需要的项
	for i := startIndex; i < endIndex; i++ {
		found := false
		for j := 0; j < list.numChildren(); j++ {
			child := list.GetChildAt(j)
			if comp, ok := child.(*ComponentImpl); ok {
				if comp.GetDataInt() == i {
					found = true
					break
				}
			}
		}

		if !found {
			item := list.AddItemFromPool("")
			item.SetData(i)
			list.itemRenderer(i, item)
		}
	}

	list.updateItemPositions()
}

// updateItemPositions 更新项位置
func (list *List) updateItemPositions() {
	// 根据布局更新所有项的位置
	if list.layout == ListLayoutTypeSingleColumn {
		list.updateSingleColumnLayout()
	} else if list.layout == ListLayoutTypeSingleRow {
		list.updateSingleRowLayout()
	} else if list.layout == ListLayoutTypeFlowHorizontal {
		list.updateFlowHorizontalLayout()
	} else if list.layout == ListLayoutTypeFlowVertical {
		list.updateFlowVerticalLayout()
	}
}

// updateSingleColumnLayout 单列布局
func (list *List) updateSingleColumnLayout() {
	y := 0.0
	for i := 0; i < list.numChildren(); i++ {
		child := list.GetChildAt(i)
		if comp, ok := child.(*ComponentImpl); ok {
			idx := comp.GetDataInt()
			if idx >= 0 && idx < list.numItems {
				child.SetPosition(0, y)
				_, h := child.Size()
				y += h + list.lineGap
			}
		}
	}
}

// updateSingleRowLayout 单行布局
func (list *List) updateSingleRowLayout() {
	x := 0.0
	for i := 0; i < list.numChildren(); i++ {
		child := list.GetChildAt(i)
		if comp, ok := child.(*ComponentImpl); ok {
			idx := comp.GetDataInt()
			if idx >= 0 && idx < list.numItems {
				child.SetPosition(x, 0)
				w, _ := child.Size()
				x += w + list.columnGap
			}
		}
	}
}

// updateFlowHorizontalLayout 水平流布局
func (list *List) updateFlowHorizontalLayout() {
	if list.columnCount <= 0 {
		list.columnCount = 1
	}

	x := 0.0
	y := 0.0
	col := 0
	maxHeight := 0.0

	for i := 0; i < list.numChildren(); i++ {
		child := list.GetChildAt(i)
		w, h := child.Size()

		if col > 0 && x + w > list.Width() {
			x = 0
			y += maxHeight + list.lineGap
			col = 0
			maxHeight = 0
		}

		child.SetPosition(x, y)
		x += w + list.columnGap
		col++

		if h > maxHeight {
			maxHeight = h
		}
	}
}

// updateFlowVerticalLayout 垂直流布局
func (list *List) updateFlowVerticalLayout() {
	if list.lineCount <= 0 {
		list.lineCount = 1
	}

	x := 0.0
	y := 0.0
	row := 0
	maxWidth := 0.0

	for i := 0; i < list.numChildren(); i++ {
		child := list.GetChildAt(i)
		w, h := child.Size()

		if row > 0 && y + h > list.Height() {
			y = 0
			x += maxWidth + list.columnGap
			row = 0
			maxWidth = 0
		}

		child.SetPosition(x, y)
		y += h + list.lineGap
		row++

		if w > maxWidth {
			maxWidth = w
		}
	}
}

// estimateItemHeight 估算项高度
func (list *List) estimateItemHeight() float64 {
	if list.numChildren() > 0 {
		child := list.GetChildAt(0)
		_, h := child.Size()
		return h
	}
	return 50 // 默认值
}

// estimateItemWidth 估算项宽度
func (list *List) estimateItemWidth() float64 {
	if list.numChildren() > 0 {
		child := list.GetChildAt(0)
		w, _ := child.Size()
		return w
	}
	return 100 // 默认值
}

// ============================================================================
// 事件处理
// ============================================================================

// OnClickItem 设置点击项事件
func (list *List) OnClickItem(handler EventHandler) {
	list.eventMutex.Lock()
	defer list.eventMutex.Unlock()
	list.onClickItem = handler
}

// OnRightClickItem 设置右键点击项事件
func (list *List) OnRightClickItem(handler EventHandler) {
	list.eventMutex.Lock()
	defer list.eventMutex.Unlock()
	list.onRightClickItem = handler
}

// emitClickItem 触发点击项事件
func (list *List) emitClickItem(item *ComponentImpl, evt Event) {
	list.eventMutex.Lock()
	handler := list.onClickItem
	list.eventMutex.Unlock()

	if handler != nil {
		handler(evt)
	}

	list.Emit(NewUIEvent("click_item", list, item))
}

// emitRightClickItem 触发右键点击项事件
func (list *List) emitRightClickItem(item *ComponentImpl, evt Event) {
	list.eventMutex.Lock()
	defer list.eventMutex.Unlock()

	if list.onRightClickItem != nil {
		list.onRightClickItem(evt)
	}

	list.Emit(NewUIEvent("right_click_item", list, item))
}

// ListItemPool - 列表项对象池
type ListItemPool struct {
	pool map[string][]DisplayObject
}

func NewListItemPool() *ListItemPool {
	return &ListItemPool{
		pool: make(map[string][]DisplayObject),
	}
}

func (p *ListItemPool) Get(url string) *ComponentImpl {
	if items, ok := p.pool[url]; ok && len(items) > 0 {
		item := items[0]
		p.pool[url] = items[1:]
		if comp, ok := item.(*ComponentImpl); ok {
			return comp
		}
	}

	// 如果没有可用项，创建新项
	comp := NewComponent()
	return comp
}

func (p *ListItemPool) Put(item DisplayObject) {
	// 简化实现
}

// ============================================================================
// 辅助方法
// ============================================================================

// Clear 清空所有项（简化实现）
func (list *List) Clear() {
	list.itemInfoMutex.Lock()
	list.refreshAllItems = true
	list.itemInfoMutex.Unlock()

	list.removeChildrenToPool(0, -1)
	list.numItems = 0
	list.selectedIndex = -1
	list.selection = []int{}
}

// SetTouchEnabled 设置是否可触摸（简化实现）
func (list *List) SetTouchEnabled(enabled bool) {
	// 这里应该设置触摸事件处理
	// 简化实现：记录状态
	_ = enabled
}

// SetHeight 设置高度（兼容方法）
func (list *List) SetHeight(height float64) {
	w, _ := list.Size()
	list.SetSize(w, height)
}

// SetWidth 设置宽度（兼容方法）
func (list *List) SetWidth(width float64) {
	_, h := list.Size()
	list.SetSize(width, h)
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertList 类型断言
func AssertList(obj DisplayObject) (*List, bool) {
	list, ok := obj.(*List)
	return list, ok
}

// IsList 检查是否是 List
func IsList(obj DisplayObject) bool {
	_, ok := obj.(*List)
	return ok
}
