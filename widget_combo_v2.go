package fairygui

import (
	"strings"
	"sync"
)

// ============================================================================
// ComboBox - 组合框控件 V2
// ============================================================================

type ComboBox struct {
	*ComponentImpl

	// 资源相关
	packageItem      *PackageItemWrapper
	template         *ComponentImpl
	dropDownResource string

	// 子对象
	titleObject  DisplayObject
	iconObject   DisplayObject
	dropDownButton DisplayObject
	popupList    *List

	// 数据
	items           []string
	values          []string
	icons           []string
	selectedIndex   int
	visibleItemCount int
	popupDirection  string
	itemRenderer    ComboBoxItemRenderer

	// 下拉状态
	popupPane       *ComponentImpl
	popup           *ComponentImpl
	dropDownVisible bool
	disabled        bool

	// 事件监听器
	onDropDown  EventHandler
	onSelection EventHandler

	// 内部状态
	dropDownMutex sync.Mutex
}

// ComboBoxItemRenderer - 项渲染器函数类型
type ComboBoxItemRenderer func(index int, item *ComponentImpl, text string)

// NewComboBox 创建新的组合框
func NewComboBox() *ComboBox {
	cb := &ComboBox{
		ComponentImpl:    NewComponent(),
		selectedIndex:    -1,
		visibleItemCount: 10,
		popupDirection:   "down",
		disabled:         false,
		items:            make([]string, 0),
		values:           make([]string, 0),
		icons:            make([]string, 0),
	}

	return cb
}

// ============================================================================
// 资源相关
// ============================================================================

// SetPackageItem 设置资源项
func (cb *ComboBox) SetPackageItem(item PackageItem) {
	if item == nil {
		cb.packageItem = nil
		return
	}
	if wrapper, ok := item.(*PackageItemWrapper); ok {
		cb.packageItem = wrapper
	}
}

// PackageItem 返回资源项
func (cb *ComboBox) PackageItem() PackageItem {
	return cb.packageItem
}

// SetTemplateComponent 设置模板组件
func (cb *ComboBox) SetTemplateComponent(comp *ComponentImpl) {
	if cb.template != nil {
		cb.RemoveChild(cb.template)
	}
	cb.template = comp
	if comp != nil {
		comp.SetPosition(0, 0)
		cb.AddChild(comp)
	}
	cb.resolveTemplate()
}

// TemplateComponent 返回模板组件
func (cb *ComboBox) TemplateComponent() *ComponentImpl {
	return cb.template
}

// ============================================================================
// 数据管理
// ============================================================================

// AddItem 添加项
func (cb *ComboBox) AddItem(item string) {
	cb.items = append(cb.items, item)
}

// AddItems 添加多个项
func (cb *ComboBox) AddItems(items []string) {
	cb.items = append(cb.items, items...)
}

// SetItems 设置所有项
func (cb *ComboBox) SetItems(items []string) {
	cb.items = make([]string, len(items))
	copy(cb.items, items)

	if cb.selectedIndex >= len(cb.items) {
		cb.selectedIndex = -1
	}
}

// Items 返回所有项
func (cb *ComboBox) Items() []string {
	result := make([]string, len(cb.items))
	copy(result, cb.items)
	return result
}

// SetValues 设置项值（可用于数据绑定）
func (cb *ComboBox) SetValues(values []string) {
	cb.values = make([]string, len(values))
	copy(cb.values, values)
}

// Values 返回所有值
func (cb *ComboBox) Values() []string {
	result := make([]string, len(cb.values))
	copy(result, cb.values)
	return result
}

// SetIcons 设置图标
func (cb *ComboBox) SetIcons(icons []string) {
	cb.icons = make([]string, len(icons))
	copy(cb.icons, icons)
}

// Icons 返回所有图标
func (cb *ComboBox) Icons() []string {
	result := make([]string, len(cb.icons))
	copy(result, cb.icons)
	return result
}

// SetSelectedIndex 设置选中索引
func (cb *ComboBox) SetSelectedIndex(index int) {
	if index < 0 || index >= len(cb.items) {
		cb.selectedIndex = -1
	} else {
		cb.selectedIndex = index
	}

	cb.updateSelection()
}

// SelectedIndex 返回选中索引
func (cb *ComboBox) SelectedIndex() int {
	return cb.selectedIndex
}

// GetSelection 返回当前选中项的文本
func (cb *ComboBox) GetSelection() string {
	if cb.selectedIndex >= 0 && cb.selectedIndex < len(cb.items) {
		return cb.items[cb.selectedIndex]
	}
	return ""
}

// GetValue 返回当前选中项的值
func (cb *ComboBox) GetValue() string {
	if cb.selectedIndex >= 0 && cb.selectedIndex < len(cb.values) {
		return cb.values[cb.selectedIndex]
	}

	if cb.selectedIndex >= 0 && cb.selectedIndex < len(cb.items) {
		return cb.items[cb.selectedIndex]
	}
	return ""
}

// Clear 清空所有项
func (cb *ComboBox) Clear() {
	cb.items = make([]string, 0)
	cb.values = make([]string, 0)
	cb.icons = make([]string, 0)
	cb.selectedIndex = -1

	if cb.popupList != nil {
		cb.popupList.Clear()
	}

	cb.updateSelection()
}

// SetSelectedItem 根据项文本设置选中
func (cb *ComboBox) SetSelectedItem(text string) bool {
	for i, item := range cb.items {
		if item == text {
			cb.SetSelectedIndex(i)
			return true
		}
	}
	return false
}

// SetSelectedValue 根据值设置选中
func (cb *ComboBox) SetSelectedValue(value string) bool {
	for i, v := range cb.values {
		if v == value {
			cb.SetSelectedIndex(i)
			return true
		}
	}
	return false
}

// GetItemAt 返回指定索引的项
func (cb *ComboBox) GetItemAt(index int) string {
	if index >= 0 && index < len(cb.items) {
		return cb.items[index]
	}
	return ""
}

// GetValueAt 返回指定索引的值
func (cb *ComboBox) GetValueAt(index int) string {
	if index >= 0 && index < len(cb.values) {
		return cb.values[index]
	}
	return ""
}

// GetIconAt 返回指定索引的图标
func (cb *ComboBox) GetIconAt(index int) string {
	if index >= 0 && index < len(cb.icons) {
		return cb.icons[index]
	}
	return ""
}

// NumItems 返回项数量
func (cb *ComboBox) NumItems() int {
	return len(cb.items)
}

// ============================================================================
// 下拉列表管理
// ============================================================================

// SetVisibleItemCount 设置下拉列表可见项数量
func (cb *ComboBox) SetVisibleItemCount(count int) {
	cb.visibleItemCount = count
	if cb.popupList != nil {
		cb.popupList.SetLineCount(count)
	}
}

// VisibleItemCount 返回下拉列表可见项数量
func (cb *ComboBox) VisibleItemCount() int {
	return cb.visibleItemCount
}

// SetDisabled 设置禁用状态
func (cb *ComboBox) SetDisabled(value bool) {
	cb.disabled = value
}

// IsDisabled 返回是否禁用
func (cb *ComboBox) IsDisabled() bool {
	return cb.disabled
}

// IsDropdownVisible 返回下拉是否可见
func (cb *ComboBox) IsDropdownVisible() bool {
	return cb.dropDownVisible
}

// ============================================================================
// 下拉显示/隐藏
// ============================================================================

// ShowDropdown 显示下拉列表
func (cb *ComboBox) ShowDropdown() {
	if cb.dropDownVisible {
		return
	}

	if cb.disabled {
		return
	}

	cb.dropDownMutex.Lock()
	defer cb.dropDownMutex.Unlock()

	if cb.popupList == nil {
		cb.createPopupList()
	}

	// 更新列表数据
	cb.popupList.SetNumItems(len(cb.items))

	// 设置选中项
	cb.popupList.SetSelectedIndex(cb.selectedIndex)

	// 设置下拉列表的尺寸
	cb.updateDropdownSize()

	// 显示下拉列表（简化实现）
	cb.dropDownVisible = true

	// 触发事件
	cb.emitDropDown()
}

// HideDropdown 隐藏下拉列表
func (cb *ComboBox) HideDropdown() {
	if !cb.dropDownVisible {
		return
	}

	cb.dropDownMutex.Lock()
	defer cb.dropDownMutex.Unlock()

	cb.dropDownVisible = false
}

// ToggleDropdown 切换下拉列表
func (cb *ComboBox) ToggleDropdown() {
	if cb.dropDownVisible {
		cb.HideDropdown()
	} else {
		cb.ShowDropdown()
	}
}

// createPopupList 创建下拉列表（内部）
func (cb *ComboBox) createPopupList() {
	cb.popupList = NewList()
	cb.popupList.SetLayout(ListLayoutTypeSingleColumn)
	cb.popupList.SetSelectionMode(ListSelectionModeSingle)
	cb.popupList.SetTouchEnabled(true)

	// 设置项渲染器
	cb.popupList.itemRenderer = func(index int, item *ComponentImpl) {
		if index >= 0 && index < len(cb.items) {
			text := cb.items[index]
			icon := ""
			if index < len(cb.icons) {
				icon = cb.icons[index]
			}

			// 自定义渲染
			if cb.itemRenderer != nil {
				cb.itemRenderer(index, item, text)
			} else {
				// 默认渲染
				if tf, ok := item.GetChildByName("title").(*TextField); ok {
					tf.SetText(text)
				}

				if icon != "" {
					if img, ok := item.GetChildByName("icon").(*Image); ok {
						// 设置图标
						_ = img
					}
				}
			}
		}
	}

	// 绑定选择事件
	cb.popupList.OnClickItem(func(event Event) {
		// 处理选择
		selectedIndex := cb.popupList.SelectedIndex()
		if selectedIndex >= 0 {
			cb.SetSelectedIndex(selectedIndex)
		}

		// 隐藏下拉列表
		cb.HideDropdown()

		// 触发自定义选择事件
		cb.emitSelection()
	})

	// 设置可见项数量
	cb.popupList.SetLineCount(cb.visibleItemCount)
}

// updateDropdownSize 更新下拉列表尺寸（内部）
func (cb *ComboBox) updateDropdownSize() {
	if cb.popupList == nil {
		return
	}

	// 计算下拉列表的高度
	itemHeight := 20.0 // 默认项高度
	if cb.popupList.NumItems() > 0 && cb.popupList.NumItems() < cb.visibleItemCount {
		cb.popupList.SetHeight(float64(cb.popupList.NumItems()) * itemHeight)
	} else {
		cb.popupList.SetHeight(float64(cb.visibleItemCount) * itemHeight)
	}

	// 设置下拉列表的宽度
	if cb.Width() > 0 {
		cb.popupList.SetWidth(cb.Width())
	}
}

// ============================================================================
// 选择更新
// ============================================================================

// updateSelection 更新选中显示
func (cb *ComboBox) updateSelection() {
	if cb.selectedIndex >= 0 && cb.selectedIndex < len(cb.items) {
		selectedText := cb.items[cb.selectedIndex]

		// 更新标题文本
		if cb.titleObject != nil {
			if tf, ok := cb.titleObject.(*TextField); ok {
				tf.SetText(selectedText)
			} else if comp, ok := cb.titleObject.(*ComponentImpl); ok {
				comp.SetData(selectedText)
			}
		}

		// 更新图标
		if cb.iconObject != nil && cb.selectedIndex < len(cb.icons) {
			if icon := cb.icons[cb.selectedIndex]; icon != "" {
				if img, ok := cb.iconObject.(*Image); ok {
					// 设置图标资源
					_ = img
				}
			}
		}
	} else {
		// 没有选中项
		if cb.titleObject != nil {
			if tf, ok := cb.titleObject.(*TextField); ok {
				tf.SetText("")
			}
		}
	}
}

// ============================================================================
// 事件处理
// ============================================================================

// OnDropDown 设置下拉事件
func (cb *ComboBox) OnDropDown(handler EventHandler) {
	cb.dropDownMutex.Lock()
	defer cb.dropDownMutex.Unlock()
	cb.onDropDown = handler
}

// OnSelection 设置选择事件
func (cb *ComboBox) OnSelection(handler EventHandler) {
	cb.dropDownMutex.Lock()
	defer cb.dropDownMutex.Unlock()
	cb.onSelection = handler
}

// emitDropDown 触发展开下拉事件
func (cb *ComboBox) emitDropDown() {
	cb.dropDownMutex.Lock()
	handler := cb.onDropDown
	cb.dropDownMutex.Unlock()

	if handler != nil {
		handler(NewUIEvent("dropdown", cb, nil))
	}
}

// emitSelection 触发选择事件
func (cb *ComboBox) emitSelection() {
	cb.dropDownMutex.Lock()
	handler := cb.onSelection
	cb.dropDownMutex.Unlock()

	if handler != nil {
		handler(NewUIEvent("selection", cb, cb.selectedIndex))
	}

	// 同时触发状态改变事件
	cb.Emit(NewUIEvent("statechanged", cb, cb.selectedIndex))
}

// ============================================================================
// 内部方法
// ============================================================================

// resolveTemplate 解析模板
func (cb *ComboBox) resolveTemplate() {
	if cb.template == nil {
		return
	}

	// 查找 title
	if child := cb.template.GetChildByName("title"); child != nil {
		cb.titleObject = child
	}

	// 查找 icon
	if child := cb.template.GetChildByName("icon"); child != nil {
		cb.iconObject = child
	}

	// 查找 dropDownButton
	if child := cb.template.GetChildByName("dropDownButton"); child != nil {
		cb.dropDownButton = child

		// 绑定点击事件
		if btn, ok := child.(*Button); ok {
			btn.OnClick(func() {
				cb.ToggleDropdown()
			})
		} else if comp, ok := child.(*ComponentImpl); ok {
			comp.On("mousedown", func(event Event) {
				if baseEvent, ok := event.(*BaseEvent); ok {
					baseEvent.StopPropagation()
				}
				cb.ToggleDropdown()
			})
		}
	}

	cb.updateSelection()
}

// ParseList 解析列表字符串（用于初始化）
func (cb *ComboBox) ParseList(text string) {
	items := strings.Split(text, ",")
	cb.SetItems(items)
	cb.updateSelection()
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertComboBox 类型断言
func AssertComboBox(obj DisplayObject) (*ComboBox, bool) {
	combo, ok := obj.(*ComboBox)
	return combo, ok
}

// IsComboBox 检查是否是 ComboBox
func IsComboBox(obj DisplayObject) bool {
	_, ok := obj.(*ComboBox)
	return ok
}
