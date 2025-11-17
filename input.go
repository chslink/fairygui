package fairygui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ============================================================================
// InputManager 实现
// ============================================================================

// BasicInputManager 是 InputManager 的基本实现。
type BasicInputManager struct {
	root *RootImpl // 根对象，用于 hit testing

	// 鼠标状态
	mouseX, mouseY   int
	lastMouseX       int
	lastMouseY       int
	mouseButtonState [3]bool // 左、右、中按钮状态

	// 触摸状态
	activeTouches map[ebiten.TouchID]struct {
		x, y int
	}

	// 鼠标悬停对象跟踪
	hoverTarget DisplayObject

	// 鼠标按下对象跟踪（用于点击合成）
	mouseDownTarget DisplayObject
}

// NewBasicInputManager 创建一个新的输入管理器。
func NewBasicInputManager(root *RootImpl) *BasicInputManager {
	return &BasicInputManager{
		root: root,
		activeTouches: make(map[ebiten.TouchID]struct {
			x, y int
		}),
	}
}

// Update 更新输入状态（每帧调用）。
func (im *BasicInputManager) Update() {
	im.updateMouse()
	im.updateTouch()
	im.updateKeyboard()
}

// MousePosition 返回鼠标位置。
func (im *BasicInputManager) MousePosition() (x, y int) {
	return im.mouseX, im.mouseY
}

// IsMouseButtonPressed 返回鼠标按钮是否按下。
func (im *BasicInputManager) IsMouseButtonPressed(button MouseButton) bool {
	if int(button) >= len(im.mouseButtonState) {
		return false
	}
	return im.mouseButtonState[button]
}

// IsKeyPressed 返回键盘按键是否按下。
func (im *BasicInputManager) IsKeyPressed(key Key) bool {
	return ebiten.IsKeyPressed(key)
}

// TouchIDs 返回当前所有触摸点的 ID。
func (im *BasicInputManager) TouchIDs() []int {
	ids := make([]int, 0, len(im.activeTouches))
	for id := range im.activeTouches {
		ids = append(ids, int(id))
	}
	return ids
}

// TouchPosition 返回指定触摸点的位置。
func (im *BasicInputManager) TouchPosition(id int) (x, y int) {
	if touch, ok := im.activeTouches[ebiten.TouchID(id)]; ok {
		return touch.x, touch.y
	}
	return 0, 0
}

// ============================================================================
// 内部更新方法
// ============================================================================

// updateMouse 更新鼠标输入状态。
func (im *BasicInputManager) updateMouse() {
	// 更新鼠标位置
	im.lastMouseX, im.lastMouseY = im.mouseX, im.mouseY
	im.mouseX, im.mouseY = ebiten.CursorPosition()

	// 检查鼠标是否移动
	mouseMoved := im.mouseX != im.lastMouseX || im.mouseY != im.lastMouseY

	// 执行 hit testing
	target := im.hitTest(float64(im.mouseX), float64(im.mouseY))

	// 处理 MouseOver/MouseOut 事件
	if mouseMoved {
		im.handleMouseMove(target)
	}

	// 处理鼠标按钮事件
	im.handleMouseButtons(target)
}

// updateTouch 更新触摸输入状态。
func (im *BasicInputManager) updateTouch() {
	// 获取所有触摸点
	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)

	// 处理新的触摸开始
	for _, id := range touchIDs {
		x, y := ebiten.TouchPosition(id)
		im.activeTouches[id] = struct{ x, y int }{x, y}

		// 执行 hit testing
		target := im.hitTest(float64(x), float64(y))
		if target != nil {
			event := NewTouchEvent(EventTouchBegin, target, int(id), float64(x), float64(y))
			im.calculateLocalPosition(event, target)
			target.DispatchEvent(event)
		}
	}

	// 处理触摸移动
	for id := range im.activeTouches {
		if inpututil.IsTouchJustReleased(id) {
			continue
		}

		x, y := ebiten.TouchPosition(id)
		oldTouch := im.activeTouches[id]

		// 检查是否移动
		if x != oldTouch.x || y != oldTouch.y {
			im.activeTouches[id] = struct{ x, y int }{x, y}

			target := im.hitTest(float64(x), float64(y))
			if target != nil {
				event := NewTouchEvent(EventTouchMove, target, int(id), float64(x), float64(y))
				im.calculateLocalPosition(event, target)
				target.DispatchEvent(event)
			}
		}
	}

	// 处理触摸结束
	releasedIDs := inpututil.AppendJustReleasedTouchIDs(nil)
	for _, id := range releasedIDs {
		if touch, ok := im.activeTouches[id]; ok {
			target := im.hitTest(float64(touch.x), float64(touch.y))
			if target != nil {
				event := NewTouchEvent(EventTouchEnd, target, int(id), float64(touch.x), float64(touch.y))
				im.calculateLocalPosition(event, target)
				target.DispatchEvent(event)
			}
			delete(im.activeTouches, id)
		}
	}
}

// updateKeyboard 更新键盘输入状态。
func (im *BasicInputManager) updateKeyboard() {
	// 获取刚按下的键
	keys := inpututil.AppendPressedKeys(nil)

	for _, key := range keys {
		if inpututil.IsKeyJustPressed(key) {
			// 发送 KeyDown 事件到焦点对象（当前简化实现发送到 root）
			if im.root != nil {
				event := NewKeyboardEvent(EventKeyDown, im.root, key)
				im.setModifierKeys(event)
				im.root.DispatchEvent(event)
			}
		}
	}

	// 处理按键释放
	// 注意：Ebiten 没有提供 JustReleasedKeys，需要自己跟踪
	// 这里简化实现，只处理按下事件
}

// handleMouseMove 处理鼠标移动事件。
func (im *BasicInputManager) handleMouseMove(target DisplayObject) {
	// 触发 MouseMove 事件
	if target != nil {
		event := NewMouseEvent(EventMouseMove, target, float64(im.mouseX), float64(im.mouseY))
		im.calculateLocalPosition(event, target)
		im.setMouseEventModifiers(event)
		target.DispatchEvent(event)
	}

	// 处理 MouseOver/MouseOut
	if target != im.hoverTarget {
		// MouseOut 事件
		if im.hoverTarget != nil {
			event := NewMouseEvent(EventMouseOut, im.hoverTarget, float64(im.mouseX), float64(im.mouseY))
			im.calculateLocalPosition(event, im.hoverTarget)
			im.setMouseEventModifiers(event)
			im.hoverTarget.DispatchEvent(event)
		}

		// MouseOver 事件
		if target != nil {
			event := NewMouseEvent(EventMouseOver, target, float64(im.mouseX), float64(im.mouseY))
			im.calculateLocalPosition(event, target)
			im.setMouseEventModifiers(event)
			target.DispatchEvent(event)
		}

		im.hoverTarget = target
	}
}

// handleMouseButtons 处理鼠标按钮事件。
func (im *BasicInputManager) handleMouseButtons(target DisplayObject) {
	// 左键
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		im.mouseButtonState[MouseButtonLeft] = true
		im.mouseDownTarget = target
		if target != nil {
			event := NewMouseEvent(EventMouseDown, target, float64(im.mouseX), float64(im.mouseY))
			event.Button = MouseButtonLeft
			im.calculateLocalPosition(event, target)
			im.setMouseEventModifiers(event)
			target.DispatchEvent(event)
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		im.mouseButtonState[MouseButtonLeft] = false
		if target != nil {
			event := NewMouseEvent(EventMouseUp, target, float64(im.mouseX), float64(im.mouseY))
			event.Button = MouseButtonLeft
			im.calculateLocalPosition(event, target)
			im.setMouseEventModifiers(event)
			target.DispatchEvent(event)

			// 合成 Click 事件（只有在同一个对象上按下和释放才触发）
			if target == im.mouseDownTarget {
				clickEvent := NewMouseEvent(EventClick, target, float64(im.mouseX), float64(im.mouseY))
				clickEvent.Button = MouseButtonLeft
				im.calculateLocalPosition(clickEvent, target)
				im.setMouseEventModifiers(clickEvent)
				target.DispatchEvent(clickEvent)
			}
		}
		im.mouseDownTarget = nil
	}

	// 右键
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		im.mouseButtonState[MouseButtonRight] = true
		if target != nil {
			event := NewMouseEvent(EventMouseDown, target, float64(im.mouseX), float64(im.mouseY))
			event.Button = MouseButtonRight
			im.calculateLocalPosition(event, target)
			im.setMouseEventModifiers(event)
			target.DispatchEvent(event)
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		im.mouseButtonState[MouseButtonRight] = false
		if target != nil {
			event := NewMouseEvent(EventMouseUp, target, float64(im.mouseX), float64(im.mouseY))
			event.Button = MouseButtonRight
			im.calculateLocalPosition(event, target)
			im.setMouseEventModifiers(event)
			target.DispatchEvent(event)

			// 右键点击事件
			clickEvent := NewMouseEvent(EventRightClick, target, float64(im.mouseX), float64(im.mouseY))
			clickEvent.Button = MouseButtonRight
			im.calculateLocalPosition(clickEvent, target)
			im.setMouseEventModifiers(clickEvent)
			target.DispatchEvent(clickEvent)
		}
	}

	// 中键
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		im.mouseButtonState[MouseButtonMiddle] = true
		if target != nil {
			event := NewMouseEvent(EventMouseDown, target, float64(im.mouseX), float64(im.mouseY))
			event.Button = MouseButtonMiddle
			im.calculateLocalPosition(event, target)
			im.setMouseEventModifiers(event)
			target.DispatchEvent(event)
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) {
		im.mouseButtonState[MouseButtonMiddle] = false
		if target != nil {
			event := NewMouseEvent(EventMouseUp, target, float64(im.mouseX), float64(im.mouseY))
			event.Button = MouseButtonMiddle
			im.calculateLocalPosition(event, target)
			im.setMouseEventModifiers(event)
			target.DispatchEvent(event)
		}
	}
}

// hitTest 执行碰撞检测，找到指定位置下最顶层的可触摸对象。
func (im *BasicInputManager) hitTest(x, y float64) DisplayObject {
	if im.root == nil {
		return nil
	}

	// 从根节点开始递归查找
	return im.hitTestRecursive(im.root, x, y)
}

// hitTestRecursive 递归执行碰撞检测。
func (im *BasicInputManager) hitTestRecursive(obj DisplayObject, x, y float64) DisplayObject {
	// 检查对象是否可见和可触摸
	if !obj.Visible() || !obj.Touchable() {
		return nil
	}

	// 先检查子对象（从后往前，因为后面的对象在上层）
	children := obj.Children()
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if result := im.hitTestRecursive(child, x, y); result != nil {
			return result
		}
	}

	// 检查当前对象
	if im.hitTestObject(obj, x, y) {
		return obj
	}

	return nil
}

// hitTestObject 检查点是否在对象内部。
func (im *BasicInputManager) hitTestObject(obj DisplayObject, x, y float64) bool {
	// 转换到对象的局部坐标
	localX, localY := im.globalToLocal(obj, x, y)

	// 简单矩形碰撞检测
	width, height := obj.Size()
	return localX >= 0 && localX <= width && localY >= 0 && localY <= height
}

// globalToLocal 将全局坐标转换为对象的局部坐标。
func (im *BasicInputManager) globalToLocal(obj DisplayObject, globalX, globalY float64) (localX, localY float64) {
	// 尝试使用对象自己的转换方法（如果是 Object 类型）
	if o, ok := obj.(*Object); ok {
		return o.GlobalToLocal(globalX, globalY)
	}
	if c, ok := obj.(*ComponentImpl); ok {
		return c.GlobalToLocal(globalX, globalY)
	}
	if r, ok := obj.(*RootImpl); ok {
		return r.GlobalToLocal(globalX, globalY)
	}

	// 备用方案：简单地减去对象的全局位置
	objX, objY := obj.GlobalPosition()
	return globalX - objX, globalY - objY
}

// calculateLocalPosition 计算事件的局部坐标。
func (im *BasicInputManager) calculateLocalPosition(event interface{}, target DisplayObject) {
	switch e := event.(type) {
	case *MouseEvent:
		e.LocalX, e.LocalY = im.globalToLocal(target, e.X, e.Y)
	case *TouchEvent:
		e.LocalX, e.LocalY = im.globalToLocal(target, e.X, e.Y)
	}
}

// setMouseEventModifiers 设置鼠标事件的修饰键状态。
func (im *BasicInputManager) setMouseEventModifiers(event *MouseEvent) {
	event.CtrlKey = ebiten.IsKeyPressed(ebiten.KeyControl)
	event.ShiftKey = ebiten.IsKeyPressed(ebiten.KeyShift)
	event.AltKey = ebiten.IsKeyPressed(ebiten.KeyAlt)
}

// setModifierKeys 设置键盘事件的修饰键状态。
func (im *BasicInputManager) setModifierKeys(event *KeyboardEvent) {
	event.CtrlKey = ebiten.IsKeyPressed(ebiten.KeyControl)
	event.ShiftKey = ebiten.IsKeyPressed(ebiten.KeyShift)
	event.AltKey = ebiten.IsKeyPressed(ebiten.KeyAlt)
}
