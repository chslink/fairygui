package display

import (
	"fmt"
	"sync"

	"github.com/chslink/fairygui/internal/types"
)

// ============================================================================
// EventDispatcher 实现
// ============================================================================

// eventListener 表示单个事件监听器
type eventListener struct {
	handler  types.EventHandler
	once     bool
	callOnce bool
}

// EventDispatcher 是事件分发器的实现
type EventDispatcher struct {
	listeners map[string][]*eventListener
	mu        sync.RWMutex
}

// NewEventDispatcher 创建新的事件分发器
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		listeners: make(map[string][]*eventListener),
	}
}

// On 注册事件处理器，返回取消订阅函数
func (ed *EventDispatcher) On(eventType string, handler types.EventHandler) (cancel func()) {
	if handler == nil {
		return func() {}
	}

	ed.mu.Lock()
	defer ed.mu.Unlock()

	listener := &eventListener{
		handler: handler,
	}

	ed.listeners[eventType] = append(ed.listeners[eventType], listener)

	// 返回取消函数
	return func() {
		ed.Off(eventType, handler)
	}
}

// Off 移除事件处理器
func (ed *EventDispatcher) Off(eventType string, handler types.EventHandler) {
	if handler == nil {
		return
	}

	ed.mu.Lock()
	defer ed.mu.Unlock()

	listeners, exists := ed.listeners[eventType]
	if !exists {
		return
	}

	// 查找并移除匹配的监听器
	// 在 Go 中不能直接比较函数，所以这里移除最后一个监听器
	// TODO: 在 listener 中添加唯一 ID 以便精确移除
	if len(listeners) > 0 {
		i := len(listeners) - 1
		// 从切片中移除
		ed.listeners[eventType] = listeners[:i]
		return
	}
}

// Once 注册一次性事件处理器
func (ed *EventDispatcher) Once(eventType string, handler types.EventHandler) {
	if handler == nil {
		return
	}

	ed.mu.Lock()
	defer ed.mu.Unlock()

	listener := &eventListener{
		handler: handler,
		once:    true,
	}

	ed.listeners[eventType] = append(ed.listeners[eventType], listener)
}

// Emit 触发事件
func (ed *EventDispatcher) Emit(event types.Event) {
	if event == nil {
		return
	}

	eventType := event.Type()

	ed.mu.RLock()
	defer ed.mu.RUnlock()

	listeners, exists := ed.listeners[eventType]
	if !exists || len(listeners) == 0 {
		return
	}

	// 创建事件执行中的监听器切片副本，避免在回调中修改原列表
	currentListeners := make([]*eventListener, len(listeners))
	copy(currentListeners, listeners)

	// 解锁，允许在回调中修改监听器
	ed.mu.RUnlock()

	// 执行监听器
	for _, listener := range currentListeners {
		if listener.once && listener.callOnce {
			continue // 跳过已经调用过的一次性监听器
		}

		// 调用处理器
		listener.handler(event)

		// 标记一次性监听器已调用
		if listener.once {
			listener.callOnce = true
		}

		// 如果事件停止了传播，不再继续
		if event.IsPropagationStopped() {
			break
		}
	}

	// 清理一次性监听器
	ed.mu.Lock()
	defer ed.mu.Unlock()

	remaining := ed.listeners[eventType][:0]
	for _, listener := range ed.listeners[eventType] {
		if !(listener.once && listener.callOnce) {
			remaining = append(remaining, listener)
		}
	}
	ed.listeners[eventType] = remaining
}

// HasListener 检查是否有事件监听器
func (ed *EventDispatcher) HasListener(eventType string) bool {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	listeners, exists := ed.listeners[eventType]
	return exists && len(listeners) > 0
}

// RemoveAllListeners 移除所有事件监听器
func (ed *EventDispatcher) RemoveAllListeners() {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	ed.listeners = make(map[string][]*eventListener)
}

// RemoveListenersForType 移除指定类型的事件监听器
func (ed *EventDispatcher) RemoveListenersForType(eventType string) {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	delete(ed.listeners, eventType)
}

// ============================================================================
// Event 实现
// ============================================================================

// baseEvent 是基础事件类型
type baseEvent struct {
	eventType      string
	target         types.DisplayObject
	currentTarget  types.DisplayObject
	data           interface{}
	stopPropagation bool
	preventDefault  bool
	mu             sync.RWMutex
}

// NewEvent 创建新的事件
func NewEvent(eventType string, target types.DisplayObject) types.Event {
	return &baseEvent{
		eventType: eventType,
		target:    target,
		currentTarget: target,
	}
}

// NewEventWithData 创建带数据的事件
func NewEventWithData(eventType string, target types.DisplayObject, data interface{}) types.Event {
	return &baseEvent{
		eventType: eventType,
		target:    target,
		currentTarget: target,
		data:      data,
	}
}

// Type 返回事件类型
func (e *baseEvent) Type() string {
	return e.eventType
}

// Target 返回事件目标
func (e *baseEvent) Target() types.DisplayObject {
	return e.target
}

// CurrentTarget 返回当前事件目标
func (e *baseEvent) CurrentTarget() types.DisplayObject {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentTarget
}

// SetCurrentTarget 设置当前事件目标（用于事件冒泡）
func (e *baseEvent) SetCurrentTarget(target types.DisplayObject) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.currentTarget = target
}

// Data 返回事件附加数据
func (e *baseEvent) Data() interface{} {
	return e.data
}

// StopPropagation 停止事件冒泡
func (e *baseEvent) StopPropagation() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stopPropagation = true
}

// IsPropagationStopped 检查事件冒泡是否已停止
func (e *baseEvent) IsPropagationStopped() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stopPropagation
}

// PreventDefault 阻止默认行为
func (e *baseEvent) PreventDefault() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.preventDefault = true
}

// IsDefaultPrevented 检查默认行为是否已阻止
func (e *baseEvent) IsDefaultPrevented() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.preventDefault
}

// ============================================================================
// 事件类型包装器
// ============================================================================

// MouseEvent 表示鼠标事件
type MouseEvent struct {
	baseEvent

	x         float64
	y         float64
	button    int
	buttons   int
	modifiers int
}

// NewMouseEvent 创建鼠标事件
func NewMouseEvent(eventType string, target types.DisplayObject, x, y float64, button int) *MouseEvent {
	return &MouseEvent{
		baseEvent: baseEvent{
			eventType: eventType,
			target:    target,
			currentTarget: target,
		},
		x:      x,
		y:      y,
		button: button,
	}
}

// X 返回鼠标 X 坐标
func (e *MouseEvent) X() float64 {
	return e.x
}

// Y 返回鼠标 Y 坐标
func (e *MouseEvent) Y() float64 {
	return e.y
}

// Button 返回鼠标按钮
func (e *MouseEvent) Button() int {
	return e.button
}

// ============================================================================
// KeyboardEvent 表示键盘事件
type KeyboardEvent struct {
	baseEvent

	key       int
	keyCode   int
	char      string
	modifiers int
	repeat    bool
}

// NewKeyboardEvent 创建键盘事件
func NewKeyboardEvent(eventType string, target types.DisplayObject, key int) *KeyboardEvent {
	return &KeyboardEvent{
		baseEvent: baseEvent{
			eventType: eventType,
			target:    target,
			currentTarget: target,
		},
		key: key,
	}
}

// Key 返回按键
func (e *KeyboardEvent) Key() int {
	return e.key
}

// ============================================================================
// TouchEvent 表示触摸事件
type TouchEvent struct {
	baseEvent

	touchID int
	x       float64
	y       float64
	phase   TouchPhase
}

// TouchPhase 定义触摸阶段
type TouchPhase int

const (
	TouchPhaseBegan TouchPhase = iota
	TouchPhaseMoved
	TouchPhaseEnded
	TouchPhaseCancelled
)

// NewTouchEvent 创建触摸事件
func NewTouchEvent(eventType string, target types.DisplayObject, touchID int, x, y float64, phase TouchPhase) *TouchEvent {
	return &TouchEvent{
		baseEvent: baseEvent{
			eventType: eventType,
			target:    target,
			currentTarget: target,
		},
		touchID: touchID,
		x:       x,
		y:       y,
		phase:   phase,
	}
}

// TouchID 返回触摸点 ID
func (e *TouchEvent) TouchID() int {
	return e.touchID
}

// Phase 返回触摸阶段
func (e *TouchEvent) Phase() TouchPhase {
	return e.phase
}

// ============================================================================
// UIEvent 表示 UI 事件
type UIEvent struct {
	baseEvent
}

// NewUIEvent 创建 UI 事件
func NewUIEvent(eventType string, target types.DisplayObject) *UIEvent {
	return &UIEvent{
		baseEvent: baseEvent{
			eventType: eventType,
			target:    target,
			currentTarget: target,
		},
	}
}

// ============================================================================
// 事件类型常量
// ============================================================================

// 鼠标事件
const (
	EventTypeClick      = "click"
	EventTypeMouseDown  = "mousedown"
	EventTypeMouseUp    = "mouseup"
	EventTypeMouseMove  = "mousemove"
	EventTypeMouseEnter = "mouseenter"
	EventTypeMouseLeave = "mouseleave"
	EventTypeMouseOver  = "mouseover"
	EventTypeMouseOut   = "mouseout"
)

// 触摸事件
const (
	EventTypeTouchStart = "touchstart"
	EventTypeTouchEnd   = "touchend"
	EventTypeTouchMove  = "touchmove"
	EventTypeTouchCancel = "touchcancel"
)

// 键盘事件
const (
	EventTypeKeyDown = "keydown"
	EventTypeKeyUp   = "keyup"
	EventTypeKeyPress = "keypress"
)

// UI 事件
const (
	EventTypeChange     = "change"
	EventTypeAdded      = "added"
	EventTypeRemoved    = "removed"
	EventTypeResized    = "resized"
	EventTypeDisposed   = "disposed"
	EventTypeLoaded     = "loaded"
	EventTypeLoading     = "loading"
	EventTypeLoadFailed = "loadfailed"
)

// ============================================================================
// 事件验证
// ============================================================================

// ValidateEventHandler 验证事件处理器是否有效
func ValidateEventHandler(handler types.EventHandler) error {
	if handler == nil {
		return fmt.Errorf("event handler cannot be nil")
	}
	return nil
}

// ValidateEvent 验证事件是否有效
func ValidateEvent(event types.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}
	if event.Type() == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	return nil
}
