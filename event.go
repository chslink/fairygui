package fairygui

// 事件类型常量
const (
	// 鼠标事件
	EventMouseDown  = "mousedown"
	EventMouseUp    = "mouseup"
	EventMouseMove  = "mousemove"
	EventMouseOver  = "mouseover"
	EventMouseOut   = "mouseout"
	EventClick      = "click"
	EventDoubleClick = "doubleclick"
	EventRightClick = "rightclick"

	// 触摸事件
	EventTouchBegin = "touchbegin"
	EventTouchMove  = "touchmove"
	EventTouchEnd   = "touchend"
	EventTouchCancel = "touchcancel"

	// 键盘事件
	EventKeyDown = "keydown"
	EventKeyUp   = "keyup"

	// UI 事件
	EventAdded        = "added"         // 添加到舞台
	EventRemoved      = "removed"       // 从舞台移除
	EventSizeChanged  = "sizechanged"   // 尺寸改变
	EventPositionChanged = "positionchanged" // 位置改变
	EventVisibleChanged = "visiblechanged"   // 可见性改变

	// 焦点事件
	EventFocusIn  = "focusin"
	EventFocusOut = "focusout"

	// 拖拽事件
	EventDragStart = "dragstart"
	EventDragMove  = "dragmove"
	EventDragEnd   = "dragend"
	EventDrop      = "drop"
)

// BaseEvent 是事件的基础实现。
type BaseEvent struct {
	typ              string
	target           interface{}
	currentTarget    interface{}
	propagationStopped bool
	defaultPrevented bool
}

// NewEvent 创建一个新事件。
func NewEvent(eventType string, target interface{}) *BaseEvent {
	return &BaseEvent{
		typ:    eventType,
		target: target,
	}
}

// Type 返回事件类型。
func (e *BaseEvent) Type() string {
	return e.typ
}

// Target 返回事件目标对象。
func (e *BaseEvent) Target() interface{} {
	return e.target
}

// CurrentTarget 返回当前处理事件的对象。
func (e *BaseEvent) CurrentTarget() interface{} {
	if e.currentTarget != nil {
		return e.currentTarget
	}
	return e.target
}

// SetCurrentTarget 设置当前处理事件的对象（内部使用）。
func (e *BaseEvent) SetCurrentTarget(target interface{}) {
	e.currentTarget = target
}

// StopPropagation 停止事件传播。
func (e *BaseEvent) StopPropagation() {
	e.propagationStopped = true
}

// IsPropagationStopped 返回事件传播是否已停止。
func (e *BaseEvent) IsPropagationStopped() bool {
	return e.propagationStopped
}

// PreventDefault 阻止事件的默认行为。
func (e *BaseEvent) PreventDefault() {
	e.defaultPrevented = true
}

// IsDefaultPrevented 返回是否已阻止默认行为。
func (e *BaseEvent) IsDefaultPrevented() bool {
	return e.defaultPrevented
}

// MouseEvent 表示鼠标事件。
type MouseEvent struct {
	BaseEvent
	X         float64     // 鼠标 X 坐标（相对于舞台）
	Y         float64     // 鼠标 Y 坐标（相对于舞台）
	LocalX    float64     // 鼠标 X 坐标（相对于目标对象）
	LocalY    float64     // 鼠标 Y 坐标（相对于目标对象）
	Button    MouseButton // 按下的鼠标按钮
	CtrlKey   bool        // 是否按下 Ctrl 键
	ShiftKey  bool        // 是否按下 Shift 键
	AltKey    bool        // 是否按下 Alt 键
}

// NewMouseEvent 创建一个新的鼠标事件。
func NewMouseEvent(eventType string, target interface{}, x, y float64) *MouseEvent {
	return &MouseEvent{
		BaseEvent: BaseEvent{
			typ:    eventType,
			target: target,
		},
		X: x,
		Y: y,
	}
}

// TouchEvent 表示触摸事件。
type TouchEvent struct {
	BaseEvent
	TouchID int     // 触摸点 ID
	X       float64 // 触摸 X 坐标（相对于舞台）
	Y       float64 // 触摸 Y 坐标（相对于舞台）
	LocalX  float64 // 触摸 X 坐标（相对于目标对象）
	LocalY  float64 // 触摸 Y 坐标（相对于目标对象）
}

// NewTouchEvent 创建一个新的触摸事件。
func NewTouchEvent(eventType string, target interface{}, touchID int, x, y float64) *TouchEvent {
	return &TouchEvent{
		BaseEvent: BaseEvent{
			typ:    eventType,
			target: target,
		},
		TouchID: touchID,
		X:       x,
		Y:       y,
	}
}

// KeyboardEvent 表示键盘事件。
type KeyboardEvent struct {
	BaseEvent
	Key      Key  // 按下的键
	CtrlKey  bool // 是否按下 Ctrl 键
	ShiftKey bool // 是否按下 Shift 键
	AltKey   bool // 是否按下 Alt 键
}

// NewKeyboardEvent 创建一个新的键盘事件。
func NewKeyboardEvent(eventType string, target interface{}, key Key) *KeyboardEvent {
	return &KeyboardEvent{
		BaseEvent: BaseEvent{
			typ:    eventType,
			target: target,
		},
		Key: key,
	}
}

// UIEvent 表示 UI 事件（尺寸、位置等变化）。
type UIEvent struct {
	BaseEvent
	Data interface{} // 事件相关数据
}

// NewUIEvent 创建一个新的 UI 事件。
func NewUIEvent(eventType string, target interface{}, data interface{}) *UIEvent {
	return &UIEvent{
		BaseEvent: BaseEvent{
			typ:    eventType,
			target: target,
		},
		Data: data,
	}
}
