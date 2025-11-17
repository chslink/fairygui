package fairygui

import (
	"math"
	"time"

	"github.com/chslink/fairygui/internal/display"
	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// Object - 基础 UI 对象
// ============================================================================

// Object 是所有 UI 元素的基类。
//
// 它包装了 internal/display.Sprite，提供用户友好的 API。
type Object struct {
	sprite *display.Sprite

	// 事件分发器
	dispatcher *basicEventDispatcher

	// 自定义绘制回调
	customDraw func(screen *ebiten.Image)
}

// NewObject 创建一个新的 Object。
func NewObject() *Object {
	obj := &Object{
		sprite:     display.NewSprite(),
		dispatcher: newBasicEventDispatcher(),
	}
	obj.sprite.SetOwner(obj)
	return obj
}

// ============================================================================
// 实现 DisplayObject 接口
// ============================================================================

// ID 返回对象的唯一标识符。
func (o *Object) ID() string {
	return o.sprite.ID()
}

// Name 返回对象的名称。
func (o *Object) Name() string {
	return o.sprite.Name()
}

// SetName 设置对象的名称。
func (o *Object) SetName(name string) {
	o.sprite.SetName(name)
}

// Data 返回对象的自定义数据。
func (o *Object) Data() interface{} {
	return o.sprite.Data()
}

// SetData 设置对象的自定义数据。
func (o *Object) SetData(data interface{}) {
	o.sprite.SetData(data)
}

// ============================================================================
// 实现 Positionable 接口
// ============================================================================

// Position 返回对象的位置（相对于父对象）。
func (o *Object) Position() (x, y float64) {
	return o.sprite.Position()
}

// SetPosition 设置对象的位置（相对于父对象）。
func (o *Object) SetPosition(x, y float64) {
	oldX, oldY := o.sprite.Position()
	o.sprite.SetPosition(x, y)

	// 触发位置改变事件
	if x != oldX || y != oldY {
		o.dispatcher.Emit(NewUIEvent(EventPositionChanged, o, nil))
	}
}

// GlobalPosition 返回对象的全局位置（相对于舞台）。
func (o *Object) GlobalPosition() (x, y float64) {
	return o.sprite.GlobalPosition()
}

// ============================================================================
// 实现 Sizable 接口
// ============================================================================

// Size 返回对象的尺寸。
func (o *Object) Size() (width, height float64) {
	return o.sprite.Size()
}

// SetSize 设置对象的尺寸。
func (o *Object) SetSize(width, height float64) {
	oldWidth, oldHeight := o.sprite.Size()
	o.sprite.SetSize(width, height)

	// 触发尺寸改变事件
	if width != oldWidth || height != oldHeight {
		o.dispatcher.Emit(NewUIEvent(EventSizeChanged, o, nil))
	}
}

// ============================================================================
// 实现 Transformable 接口
// ============================================================================

// Scale 返回对象的缩放比例。
func (o *Object) Scale() (scaleX, scaleY float64) {
	return o.sprite.Scale()
}

// SetScale 设置对象的缩放比例。
func (o *Object) SetScale(scaleX, scaleY float64) {
	o.sprite.SetScale(scaleX, scaleY)
}

// Rotation 返回对象的旋转角度（度）。
func (o *Object) Rotation() float64 {
	// 转换弧度到度
	return o.sprite.Rotation() * 180 / math.Pi
}

// SetRotation 设置对象的旋转角度（度）。
func (o *Object) SetRotation(rotation float64) {
	// 转换度到弧度
	o.sprite.SetRotation(rotation * math.Pi / 180)
}

// Skew 返回对象的倾斜角度（度）。
func (o *Object) Skew() (skewX, skewY float64) {
	radX, radY := o.sprite.Skew()
	// 转换弧度到度
	return radX * 180 / math.Pi, radY * 180 / math.Pi
}

// SetSkew 设置对象的倾斜角度（度）。
func (o *Object) SetSkew(skewX, skewY float64) {
	// 转换度到弧度
	o.sprite.SetSkew(skewX*math.Pi/180, skewY*math.Pi/180)
}

// Pivot 返回对象的锚点（0-1 归一化坐标）。
func (o *Object) Pivot() (pivotX, pivotY float64) {
	return o.sprite.Pivot()
}

// SetPivot 设置对象的锚点（0-1 归一化坐标）。
func (o *Object) SetPivot(pivotX, pivotY float64) {
	o.sprite.SetPivot(pivotX, pivotY)
}

// ============================================================================
// 实现 Visible 接口
// ============================================================================

// Visible 返回对象是否可见。
func (o *Object) Visible() bool {
	return o.sprite.Visible()
}

// SetVisible 设置对象是否可见。
func (o *Object) SetVisible(visible bool) {
	oldVisible := o.sprite.Visible()
	o.sprite.SetVisible(visible)

	// 触发可见性改变事件
	if visible != oldVisible {
		o.dispatcher.Emit(NewUIEvent(EventVisibleChanged, o, visible))
	}
}

// Alpha 返回对象的透明度（0-1）。
func (o *Object) Alpha() float64 {
	return o.sprite.Alpha()
}

// SetAlpha 设置对象的透明度（0-1）。
func (o *Object) SetAlpha(alpha float64) {
	o.sprite.SetAlpha(alpha)
}

// ============================================================================
// 实现 Hierarchical 接口
// ============================================================================

// Parent 返回父对象。
func (o *Object) Parent() DisplayObject {
	sprite := o.sprite.Parent()
	if sprite == nil {
		return nil
	}
	owner := sprite.Owner()
	if obj, ok := owner.(DisplayObject); ok {
		return obj
	}
	return nil
}

// Children 返回子对象列表（只读）。
func (o *Object) Children() []DisplayObject {
	sprites := o.sprite.Children()
	result := make([]DisplayObject, 0, len(sprites))
	for _, sprite := range sprites {
		if owner := sprite.Owner(); owner != nil {
			if obj, ok := owner.(DisplayObject); ok {
				result = append(result, obj)
			}
		}
	}
	return result
}

// AddChild 添加子对象。
func (o *Object) AddChild(child DisplayObject) {
	o.AddChildAt(child, o.ChildCount())
}

// AddChildAt 在指定索引位置添加子对象。
func (o *Object) AddChildAt(child DisplayObject, index int) {
	if child == nil {
		return
	}

	// 获取子对象的 sprite
	childSprite := getSprite(child)
	if childSprite == nil {
		return
	}

	o.sprite.AddChildAt(childSprite, index)

	// 触发添加到舞台事件
	if childObj, ok := child.(*Object); ok {
		childObj.dispatcher.Emit(NewUIEvent(EventAdded, child, o))
	}
}

// RemoveChild 移除子对象。
func (o *Object) RemoveChild(child DisplayObject) {
	if child == nil {
		return
	}

	childSprite := getSprite(child)
	if childSprite == nil {
		return
	}

	if o.sprite.RemoveChild(childSprite) {
		// 触发从舞台移除事件
		if childObj, ok := child.(*Object); ok {
			childObj.dispatcher.Emit(NewUIEvent(EventRemoved, child, o))
		}
	}
}

// RemoveChildAt 移除指定索引位置的子对象。
func (o *Object) RemoveChildAt(index int) DisplayObject {
	sprite := o.sprite.RemoveChildAt(index)
	if sprite == nil {
		return nil
	}

	owner := sprite.Owner()
	if obj, ok := owner.(DisplayObject); ok {
		// 触发从舞台移除事件
		if childObj, ok := obj.(*Object); ok {
			childObj.dispatcher.Emit(NewUIEvent(EventRemoved, obj, o))
		}
		return obj
	}

	return nil
}

// ChildCount 返回子对象数量。
func (o *Object) ChildCount() int {
	return o.sprite.ChildCount()
}

// GetChildAt 返回指定索引位置的子对象。
func (o *Object) GetChildAt(index int) DisplayObject {
	sprite := o.sprite.GetChildAt(index)
	if sprite == nil {
		return nil
	}

	owner := sprite.Owner()
	if obj, ok := owner.(DisplayObject); ok {
		return obj
	}

	return nil
}

// GetChildByName 返回指定名称的子对象。
func (o *Object) GetChildByName(name string) DisplayObject {
	sprite := o.sprite.GetChildByName(name)
	if sprite == nil {
		return nil
	}

	owner := sprite.Owner()
	if obj, ok := owner.(DisplayObject); ok {
		return obj
	}

	return nil
}

// ============================================================================
// 实现 Drawable 接口
// ============================================================================

// Draw 在屏幕上绘制对象。
func (o *Object) Draw(screen *ebiten.Image) {
	if !o.sprite.Visible() {
		return
	}

	// 如果有自定义绘制回调，使用它
	if o.customDraw != nil {
		o.customDraw(screen)
		return
	}

	// 否则使用 sprite 的默认绘制
	o.sprite.Draw(screen)
}

// SetCustomDraw 设置自定义绘制回调。
func (o *Object) SetCustomDraw(draw func(screen *ebiten.Image)) {
	o.customDraw = draw
}

// ============================================================================
// 实现 Interactive 接口
// ============================================================================

// Touchable 返回对象是否可以接收触摸事件。
func (o *Object) Touchable() bool {
	return o.sprite.Touchable()
}

// SetTouchable 设置对象是否可以接收触摸事件。
func (o *Object) SetTouchable(touchable bool) {
	o.sprite.SetTouchable(touchable)
}

// On 注册事件处理器。
func (o *Object) On(eventType string, handler EventHandler) func() {
	return o.dispatcher.On(eventType, handler)
}

// Off 移除事件处理器。
func (o *Object) Off(eventType string, handler EventHandler) {
	o.dispatcher.Off(eventType, handler)
}

// Once 注册一次性事件处理器。
func (o *Object) Once(eventType string, handler EventHandler) {
	o.dispatcher.Once(eventType, handler)
}

// Emit 触发事件（不冒泡）。
func (o *Object) Emit(event Event) {
	o.dispatcher.Emit(event)
}

// DispatchEvent 分发事件（支持冒泡）。
//
// 事件会经历三个阶段：
// 1. 捕获阶段：从根节点向下到目标节点
// 2. 目标阶段：在目标节点上触发
// 3. 冒泡阶段：从目标节点向上到根节点
//
// 当前实现只支持冒泡阶段。
func (o *Object) DispatchEvent(event Event) {
	if event == nil {
		return
	}

	// 收集父节点链（用于冒泡）
	var chain []DisplayObject
	current := DisplayObject(o)
	for current != nil {
		chain = append(chain, current)
		current = current.Parent()
	}

	// 目标阶段 + 冒泡阶段
	for i := 0; i < len(chain); i++ {
		target := chain[i]

		// 设置当前目标
		if baseEvent, ok := event.(*BaseEvent); ok {
			baseEvent.SetCurrentTarget(target)
		}

		// 触发事件
		target.Emit(event)

		// 检查传播是否已停止
		if event.IsPropagationStopped() {
			break
		}
	}
}

// HasListener 返回是否有指定类型的事件监听器。
func (o *Object) HasListener(eventType string) bool {
	return o.dispatcher.HasListener(eventType)
}

// OnClick 注册点击事件处理器（便捷方法）。
func (o *Object) OnClick(handler func()) {
	o.On(EventClick, func(event Event) {
		handler()
	})
}

// OnMouseOver 注册鼠标移入事件处理器。
func (o *Object) OnMouseOver(handler func()) {
	o.On(EventMouseOver, func(event Event) {
		handler()
	})
}

// OnMouseOut 注册鼠标移出事件处理器。
func (o *Object) OnMouseOut(handler func()) {
	o.On(EventMouseOut, func(event Event) {
		handler()
	})
}

// OnMouseDown 注册鼠标按下事件处理器。
func (o *Object) OnMouseDown(handler func()) {
	o.On(EventMouseDown, func(event Event) {
		handler()
	})
}

// OnMouseUp 注册鼠标释放事件处理器。
func (o *Object) OnMouseUp(handler func()) {
	o.On(EventMouseUp, func(event Event) {
		handler()
	})
}

// ============================================================================
// Dispose
// ============================================================================

// Dispose 释放对象资源。
func (o *Object) Dispose() {
	o.sprite.Dispose()
	o.dispatcher = nil
	o.customDraw = nil
}

// ============================================================================
// 内部辅助方法
// ============================================================================

// getSprite 从 DisplayObject 获取内部 sprite（辅助函数）。
func getSprite(obj DisplayObject) *display.Sprite {
	if o, ok := obj.(*Object); ok {
		return o.sprite
	}
	if c, ok := obj.(*ComponentImpl); ok {
		return c.Object.sprite
	}
	if r, ok := obj.(*RootImpl); ok {
		return r.ComponentImpl.Object.sprite
	}
	return nil
}

// Sprite 返回内部 sprite（用于渲染系统）。
func (o *Object) Sprite() *display.Sprite {
	return o.sprite
}

// SetTexture 设置纹理（便捷方法）。
func (o *Object) SetTexture(texture *ebiten.Image) {
	o.sprite.SetTexture(texture)
}

// Texture 返回纹理。
func (o *Object) Texture() *ebiten.Image {
	return o.sprite.Texture()
}

// HitTest 检测点是否在对象内部（局部坐标）。
func (o *Object) HitTest(localX, localY float64) bool {
	return o.sprite.HitTest(localX, localY)
}

// HitTestGlobal 检测点是否在对象内部（全局坐标）。
func (o *Object) HitTestGlobal(globalX, globalY float64) bool {
	return o.sprite.HitTestGlobal(globalX, globalY)
}

// LocalToGlobal 将局部坐标转换为全局坐标。
func (o *Object) LocalToGlobal(localX, localY float64) (globalX, globalY float64) {
	return o.sprite.LocalToGlobal(localX, localY)
}

// GlobalToLocal 将全局坐标转换为局部坐标。
func (o *Object) GlobalToLocal(globalX, globalY float64) (localX, localY float64) {
	return o.sprite.GlobalToLocal(globalX, globalY)
}

// ============================================================================
// Component - 容器组件
// ============================================================================

// ComponentImpl 是容器组件的实现。
type ComponentImpl struct {
	*Object

	controllers []Controller
}

// NewComponent 创建一个新的 Component。
func NewComponent() *ComponentImpl {
	return &ComponentImpl{
		Object:      NewObject(),
		controllers: make([]Controller, 0),
	}
}

// Controllers 返回组件的所有控制器。
func (c *ComponentImpl) Controllers() []Controller {
	result := make([]Controller, len(c.controllers))
	copy(result, c.controllers)
	return result
}

// GetController 根据名称获取控制器。
func (c *ComponentImpl) GetController(name string) Controller {
	for _, ctrl := range c.controllers {
		if ctrl.Name() == name {
			return ctrl
		}
	}
	return nil
}

// AddController 添加控制器。
func (c *ComponentImpl) AddController(controller Controller) {
	if controller == nil {
		return
	}
	c.controllers = append(c.controllers, controller)
}

// ============================================================================
// Root - 根对象
// ============================================================================

// RootImpl 是根对象的实现。
type RootImpl struct {
	*ComponentImpl

	width  int
	height int

	renderer     Renderer
	inputManager InputManager

	// 更新时间
	lastUpdate time.Time
}

// NewRoot 创建一个新的 Root。
func NewRoot(width, height int) *RootImpl {
	root := &RootImpl{
		ComponentImpl: NewComponent(),
		width:         width,
		height:        height,
		lastUpdate:    time.Now(),
	}

	root.SetSize(float64(width), float64(height))
	return root
}

// Update 更新根对象（每帧调用）。
func (r *RootImpl) Update(delta time.Duration) error {
	// 更新输入管理器
	if r.inputManager != nil {
		r.inputManager.Update()
	}

	// TODO: 更新动画、tween 等

	return nil
}

// Draw 绘制根对象。
func (r *RootImpl) Draw(screen *ebiten.Image) {
	if r.renderer != nil {
		r.renderer.Draw(screen, r)
	} else {
		// 使用默认绘制
		r.Object.Draw(screen)
	}
}

// Resize 调整舞台尺寸。
func (r *RootImpl) Resize(width, height int) {
	r.width = width
	r.height = height
	r.SetSize(float64(width), float64(height))
}

// SetRenderer 设置渲染器。
func (r *RootImpl) SetRenderer(renderer Renderer) {
	r.renderer = renderer
}

// Renderer 返回当前渲染器。
func (r *RootImpl) Renderer() Renderer {
	return r.renderer
}

// SetInputManager 设置输入管理器。
func (r *RootImpl) SetInputManager(input InputManager) {
	r.inputManager = input
}

// InputManager 返回输入管理器。
func (r *RootImpl) InputManager() InputManager {
	return r.inputManager
}

// Width 返回舞台宽度。
func (r *RootImpl) Width() int {
	return r.width
}

// Height 返回舞台高度。
func (r *RootImpl) Height() int {
	return r.height
}
