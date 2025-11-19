package fairygui

import (
	"sync"
)

// ============================================================================
// ButtonMode - 按钮模式定义
// ============================================================================

type ButtonMode int

const (
	ButtonModeCommon ButtonMode = iota
	ButtonModeCheck
	ButtonModeRadio
)

const (
	buttonStateUp               = "up"
	buttonStateDown             = "down"
	buttonStateOver             = "over"
	buttonStateDisabled         = "disabled"
	buttonStateSelectedOver     = "selectedOver"
	buttonStateSelectedDisabled = "selectedDisabled"
)

// ============================================================================
// Button - 按钮控件 V2 (基于新架构)
// ============================================================================

type Button struct {
	*ComponentImpl

	// 资源相关
	packageItem *PackageItemWrapper
	resource    string
	template    *ComponentImpl

	// 标题
	title         string
	selectedTitle string
	titleObject   DisplayObject

	// 图标
	icon         string
	selectedIcon string
	iconItem     *PackageItemWrapper
	iconObject   DisplayObject

	// 按钮模式
	mode               ButtonMode
	selected           bool
	changeStateOnClick bool

	// 交互状态
	hovered bool
	pressed bool

	// 音效
	sound            string
	soundVolumeScale float64

	// 控制器
	buttonController     Controller
	relatedController    Controller
	relatedPageID        string
	controllerListenerID int

	// 按下效果
	downEffect      int
	downEffectValue float64
	downScaled      bool
	baseScaleX      float64
	baseScaleY      float64

	// 样式
	titleColor        string
	titleOutlineColor string
	titleFontSize     int

	// 事件同步
	eventOnce sync.Once
}

// NewButton 创建一个新的按钮
func NewButton() *Button {
	btn := &Button{
		ComponentImpl:      NewComponent(),
		mode:               ButtonModeCommon,
		changeStateOnClick: true,
		downEffectValue:    0.8,
		soundVolumeScale:   1.0,
		baseScaleX:         1.0,
		baseScaleY:         1.0,
	}

	btn.titleColor = "#ffffff"
	btn.titleFontSize = 12

	// 按钮默认拦截事件
	btn.SetTouchable(true)

	// 查找 button controller
	btn.findButtonController()

	// 绑定事件
	btn.bindEvents()

	return btn
}

// ============================================================================
// 资源相关
// ============================================================================

// SetPackageItem 设置资源项
func (b *Button) SetPackageItem(item PackageItem) {
	if item == nil {
		b.packageItem = nil
		return
	}
	if wrapper, ok := item.(*PackageItemWrapper); ok {
		b.packageItem = wrapper
	}
}

// PackageItem 返回资源项
func (b *Button) PackageItem() PackageItem {
	return b.packageItem
}

// SetResource 设置资源标识符
func (b *Button) SetResource(res string) {
	b.resource = res
}

// Resource 返回资源标识符
func (b *Button) Resource() string {
	return b.resource
}

// SetTemplateComponent 设置模板组件
func (b *Button) SetTemplateComponent(comp *ComponentImpl) {
	if b.template != nil {
		b.RemoveChild(b.template)
	}
	b.template = comp
	if comp != nil {
		comp.SetPosition(0, 0)
		b.AddChild(comp)
	}
	b.updateVisualState()
}

// TemplateComponent 返回模板组件
func (b *Button) TemplateComponent() *ComponentImpl {
	return b.template
}

// ============================================================================
// 标题相关
// ============================================================================

// SetTitle 设置标题文本
func (b *Button) SetTitle(value string) {
	if b.title == value {
		return
	}
	b.title = value
	b.applyTitleState()
}

// Title 返回标题文本
func (b *Button) Title() string {
	return b.title
}

// SetSelectedTitle 设置选中状态的标题
func (b *Button) SetSelectedTitle(value string) {
	b.selectedTitle = value
	b.applyTitleState()
}

// SelectedTitle 返回选中状态的标题
func (b *Button) SelectedTitle() string {
	return b.selectedTitle
}

// SetTitleObject 设置标题对象
func (b *Button) SetTitleObject(obj DisplayObject) {
	b.titleObject = obj
	b.applyTitleState()
}

// TitleObject 返回标题对象
func (b *Button) TitleObject() DisplayObject {
	return b.titleObject
}

// SetTitleColor 设置标题颜色
func (b *Button) SetTitleColor(color string) {
	b.titleColor = color
	b.applyTitleState()
}

// TitleColor 返回标题颜色
func (b *Button) TitleColor() string {
	return b.titleColor
}

// SetTitleFontSize 设置标题字体大小
func (b *Button) SetTitleFontSize(size int) {
	b.titleFontSize = size
	b.applyTitleState()
}

// TitleFontSize 返回标题字体大小
func (b *Button) TitleFontSize() int {
	return b.titleFontSize
}

// ============================================================================
// 图标相关
// ============================================================================

// SetIcon 设置图标
func (b *Button) SetIcon(value string) {
	if b.icon == value {
		return
	}
	b.icon = value
	b.applyIconState()
}

// Icon 返回图标
func (b *Button) Icon() string {
	return b.icon
}

// SetSelectedIcon 设置选中状态的图标
func (b *Button) SetSelectedIcon(value string) {
	b.selectedIcon = value
	b.applyIconState()
}

// SelectedIcon 返回选中状态的图标
func (b *Button) SelectedIcon() string {
	return b.selectedIcon
}

// SetIconItem 设置图标资源项
func (b *Button) SetIconItem(item PackageItem) {
	if item == nil {
		b.iconItem = nil
		return
	}
	if wrapper, ok := item.(*PackageItemWrapper); ok {
		b.iconItem = wrapper
	}
}

// IconItem 返回图标资源项
func (b *Button) IconItem() PackageItem {
	return b.iconItem
}

// SetIconObject 设置图标对象
func (b *Button) SetIconObject(obj DisplayObject) {
	b.iconObject = obj
	b.applyIconState()
}

// IconObject 返回图标对象
func (b *Button) IconObject() DisplayObject {
	return b.iconObject
}

// ============================================================================
// 按钮模式
// ============================================================================

// SetMode 设置按钮模式
func (b *Button) SetMode(value ButtonMode) {
	if b.mode == value {
		return
	}

	if value == ButtonModeCommon {
		if b.selected {
			b.selected = false
			b.applyTitleState()
			b.applyIconState()
			b.updateVisualState()
		}
	}

	b.mode = value
}

// Mode 返回按钮模式
func (b *Button) Mode() ButtonMode {
	return b.mode
}

// SetSelected 设置选中状态
func (b *Button) SetSelected(value bool) {
	if b.mode == ButtonModeCommon {
		return
	}
	if b.selected == value {
		return
	}

	b.selected = value
	b.applyTitleState()
	b.applyIconState()
	b.syncRelatedController()
	b.updateVisualState()
}

// Selected 返回是否选中
func (b *Button) Selected() bool {
	return b.selected
}

// SetChangeStateOnClick 设置点击时是否切换状态
func (b *Button) SetChangeStateOnClick(value bool) {
	b.changeStateOnClick = value
}

// ChangeStateOnClick 返回点击时是否切换状态
func (b *Button) ChangeStateOnClick() bool {
	return b.changeStateOnClick
}

// ============================================================================
// 音效
// ============================================================================

// SetSound 设置音效
func (b *Button) SetSound(value string) {
	b.sound = value
}

// Sound 返音音效
func (b *Button) Sound() string {
	return b.sound
}

// SetSoundVolumeScale 设置音效音量缩放
func (b *Button) SetSoundVolumeScale(value float64) {
	b.soundVolumeScale = value
}

// SoundVolumeScale 返回音效音量缩放
func (b *Button) SoundVolumeScale() float64 {
	return b.soundVolumeScale
}

// ============================================================================
// 控制器
// ============================================================================

// SetButtonController 设置按钮控制器
func (b *Button) SetButtonController(ctrl Controller) {
	b.buttonController = ctrl
	b.updateVisualState()
}

// ButtonController 返回按钮控制器
func (b *Button) ButtonController() Controller {
	return b.buttonController
}

// SetRelatedController 设置关联控制器
func (b *Button) SetRelatedController(ctrl Controller) {
	if b.relatedController != nil && b.controllerListenerID != 0 {
		// 移除旧监听器
		if ctrl, ok := b.relatedController.(*ControllerImpl); ok {
			ctrl.RemoveSelectionListener(b.controllerListenerID)
		}
		b.controllerListenerID = 0
	}

	b.relatedController = ctrl

	if b.relatedController != nil {
		// 添加新监听器
		if ctrl, ok := b.relatedController.(*ControllerImpl); ok {
			b.controllerListenerID = ctrl.AddSelectionListener(func(c Controller) {
				b.handleControllerChanged(c)
			})
		}
	}
}

// RelatedController 返回关联控制器
func (b *Button) RelatedController() Controller {
	return b.relatedController
}

// SetRelatedPageID 设置关联页面ID
func (b *Button) SetRelatedPageID(value string) {
	b.relatedPageID = value
}

// RelatedPageID 返回关联页面ID
func (b *Button) RelatedPageID() string {
	return b.relatedPageID
}

// ============================================================================
// 按下效果
// ============================================================================

// SetDownEffect 设置按下效果
func (b *Button) SetDownEffect(value int) {
	b.downEffect = value
}

// DownEffect 返回按下效果
func (b *Button) DownEffect() int {
	return b.downEffect
}

// SetDownEffectValue 设置按下效果值
func (b *Button) SetDownEffectValue(value float64) {
	b.downEffectValue = value
}

// DownEffectValue 返回按下效果值
func (b *Button) DownEffectValue() float64 {
	return b.downEffectValue
}

// ============================================================================
// 内部方法
// ============================================================================

// findButtonController 查找按钮控制器
func (b *Button) findButtonController() {
	if b.buttonController != nil {
		return
	}

	// 优先从子控件中查找
	if child := b.GetChildByName("button"); child != nil {
		if comp, ok := child.(*ComponentImpl); ok {
			if ctrl := comp.GetController("button"); ctrl != nil {
				b.buttonController = ctrl
				return
			}
		}
	}

	// 从模板中查找
	if b.template != nil {
		if ctrl := b.template.GetController("button"); ctrl != nil {
			b.buttonController = ctrl
			return
		}
	}

	// 从当前组件查找
	controllers := b.Controllers()
	for _, ctrl := range controllers {
		if ctrl.Name() == "button" {
			b.buttonController = ctrl
			return
		}
	}
}

// bindEvents 绑定事件
func (b *Button) bindEvents() {
	b.eventOnce.Do(func() {
		b.OnMouseOver(func() {
			b.onRollOver()
		})
		b.OnMouseOut(func() {
			b.onRollOut()
		})
		b.OnMouseDown(func() {
			b.onMouseDown()
		})
		b.OnMouseUp(func() {
			b.onMouseUp()
		})
		b.OnClick(func() {
			b.onClick()
		})
	})
}

// 事件处理器
func (b *Button) onRollOver() {
	b.hovered = true
	b.updateVisualState()
}

func (b *Button) onRollOut() {
	b.hovered = false
	if b.pressed {
		b.applyDownScale(false)
	}
	b.updateVisualState()
}

func (b *Button) onMouseDown() {
	if !b.Touchable() {
		return
	}
	b.pressed = true

	if b.downEffect == 2 {
		b.applyDownScale(true)
	}

	b.updateVisualState()
}

func (b *Button) onMouseUp() {
	if !b.pressed {
		return
	}
	b.pressed = false

	if b.downEffect == 2 {
		b.applyDownScale(false)
	}

	b.updateVisualState()
}

func (b *Button) onClick() {
	if !b.Touchable() {
		return
	}

	// 播放音效
	if b.sound != "" {
		// TODO: 实现音效播放
		// PlayOneShotSound(b.sound, b.soundVolumeScale)
	}

	if !b.changeStateOnClick {
		return
	}

	prevSelected := b.selected

	switch b.mode {
	case ButtonModeCheck:
		b.SetSelected(!b.selected)
	case ButtonModeRadio:
		if !b.selected {
			b.SetSelected(true)
		}
	}

	if prevSelected != b.selected {
		b.emitStateChanged()
	}
}

// applyDownScale 应用按下缩放效果
func (b *Button) applyDownScale(down bool) {
	factor := b.downEffectValue
	if factor <= 0 {
		factor = 0.8
	}

	if down {
		if !b.downScaled {
			sx, sy := b.Scale()
			b.SetScale(sx*factor, sy*factor)
			b.downScaled = true
		}
	} else {
		if b.downScaled {
			// 恢复到基础缩放（从 Sprite 获取原始值）
			b.SetScale(b.baseScaleX, b.baseScaleY)
			b.downScaled = false
		}
	}
}

// emitStateChanged 触发状态改变事件
func (b *Button) emitStateChanged() {
	b.Emit(NewUIEvent("statechanged", b, b.selected))
}

// syncRelatedController 同步关联控制器
func (b *Button) syncRelatedController() {
	if b.relatedController == nil || b.relatedPageID == "" {
		return
	}

	if b.selected {
		if ctrl, ok := b.relatedController.(*ControllerImpl); ok {
			ctrl.SetSelectedPageID(b.relatedPageID)
		}
	}
}

// handleControllerChanged 处理控制器改变
func (b *Button) handleControllerChanged(ctrl Controller) {
	if b.relatedController != ctrl {
		return
	}

	var pageID string
	if controller, ok := b.relatedController.(*ControllerImpl); ok {
		pageID = controller.SelectedPageID()
	}

	if pageID == b.relatedPageID {
		b.SetSelected(true)
	} else if b.mode != ButtonModeCheck {
		b.SetSelected(false)
	}
}

// updateVisualState 更新视觉状态
func (b *Button) updateVisualState() {
	if b.buttonController == nil {
		return
	}

	state := b.determineState()
	b.applyState(state)
}

// determineState 确定当前状态
func (b *Button) determineState() string {
	if !b.Visible() || !b.Touchable() {
		return buttonStateDisabled
	}

	alpha := b.Alpha()
	if alpha <= 0.5 {
		return buttonStateDisabled
	}

	if b.mode == ButtonModeCommon {
		if b.pressed {
			return buttonStateDown
		}
		if b.hovered {
			return buttonStateOver
		}
		return buttonStateUp
	}

	if b.selected {
		if b.pressed {
			return buttonStateDown
		}
		if b.hovered {
			return buttonStateSelectedOver
		}
		return buttonStateUp
	}

	if b.pressed {
		return buttonStateDown
	}
	if b.hovered {
		return buttonStateOver
	}

	return buttonStateUp
}

// hasState 检查是否有指定状态
func (b *Button) hasState(name string) bool {
	if b.buttonController == nil {
		return false
	}

	ctrl := b.buttonController
	pages := ctrl.PageNames()
	if pages == nil {
		return false
	}

	for _, page := range pages {
		if page == name {
			return true
		}
	}

	return false
}

// applyState 应用状态
func (b *Button) applyState(name string) {
	if !b.hasState(name) {
		return
	}

	if ctrl := b.buttonController; ctrl != nil {
		ctrl.SetSelectedPageName(name)
	}
}

// ============================================================================
// 内部更新方法
// ============================================================================

// applyTitleState 应用标题状态
func (b *Button) applyTitleState() {
	text := b.title
	if b.selected && b.selectedTitle != "" {
		text = b.selectedTitle
	}

	if b.titleObject == nil {
		// 自动查找
		if child := b.GetChildByName("title"); child != nil {
			b.titleObject = child
		} else if b.template != nil {
			if child := b.template.GetChildByName("title"); child != nil {
				b.titleObject = child
			}
		}
	}

	if b.titleObject == nil {
		return
	}

	// 尝试设置文本（这里假设 titleObject 实现了文本设置接口）
	// 实际项目中可能会有 TextField 等专门的文本控件
	if tf, ok := b.titleObject.(*TextField); ok {
		tf.SetText(text)
		if b.titleColor != "" && tf.Color() == "" {
			tf.SetColor(b.titleColor)
		}
	} else if comp, ok := b.titleObject.(*ComponentImpl); ok {
		// 如果是组件，尝试设置数据
		comp.SetData(text)
	}
}

// applyIconState 应用图标状态
func (b *Button) applyIconState() {
	icon := b.icon
	if b.selected && b.selectedIcon != "" {
		icon = b.selectedIcon
	}

	if icon == "" {
		return
	}

	if b.iconObject == nil {
		// 自动查找
		if child := b.GetChildByName("icon"); child != nil {
			b.iconObject = child
		} else if b.template != nil {
			if child := b.template.GetChildByName("icon"); child != nil {
				b.iconObject = child
			}
		}
	}

	if b.iconObject == nil {
		return
	}

	// 尝试设置图标（这里假设实现了图标设置接口）
	if img, ok := b.iconObject.(*Image); ok {
		if item := b.iconItem; item != nil {
			img.SetPackageItem(item)
		}
	} else if loader, ok := b.iconObject.(*Loader); ok {
		loader.SetURL(icon)
	}
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertButton 类型断言
func AssertButton(obj DisplayObject) (*Button, bool) {
	btn, ok := obj.(*Button)
	return btn, ok
}

// IsButton 检查是否是 Button
func IsButton(obj DisplayObject) bool {
	_, ok := obj.(*Button)
	return ok
}
