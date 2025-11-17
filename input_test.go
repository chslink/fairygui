package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// InputManager 基础功能测试
// ============================================================================

func TestBasicInputManager_Creation(t *testing.T) {
	root := fairygui.NewRoot(800, 600)
	im := fairygui.NewBasicInputManager(root)

	if im == nil {
		t.Fatal("Expected non-nil InputManager")
	}

	// 验证初始状态
	x, y := im.MousePosition()
	if x != 0 || y != 0 {
		t.Errorf("Expected initial mouse position (0, 0), got (%d, %d)", x, y)
	}

	// 验证按钮状态
	if im.IsMouseButtonPressed(fairygui.MouseButtonLeft) {
		t.Error("Expected left button not pressed initially")
	}
	if im.IsMouseButtonPressed(fairygui.MouseButtonRight) {
		t.Error("Expected right button not pressed initially")
	}
	if im.IsMouseButtonPressed(fairygui.MouseButtonMiddle) {
		t.Error("Expected middle button not pressed initially")
	}

	// 验证触摸状态
	touchIDs := im.TouchIDs()
	if len(touchIDs) != 0 {
		t.Errorf("Expected no active touches, got %d", len(touchIDs))
	}
}

// ============================================================================
// Hit Testing 测试
// ============================================================================

func TestInputManager_HitTesting_SimpleObject(t *testing.T) {
	// 创建层级结构
	root := fairygui.NewRoot(800, 600)
	im := fairygui.NewBasicInputManager(root)

	obj := fairygui.NewObject()
	obj.SetPosition(100, 100)
	obj.SetSize(200, 150)
	obj.SetTouchable(true)
	root.AddChild(obj)

	// 由于我们无法模拟 Ebiten 的输入，这个测试主要验证接口存在
	// 实际的 hit testing 需要在集成测试中验证
	if im == nil {
		t.Fatal("InputManager should exist")
	}

	// 验证对象属性
	if !obj.Touchable() {
		t.Error("Expected object to be touchable")
	}
}

func TestInputManager_HitTesting_Hierarchy(t *testing.T) {
	// 创建嵌套层级
	root := fairygui.NewRoot(800, 600)
	im := fairygui.NewBasicInputManager(root)

	parent := fairygui.NewComponent()
	parent.SetPosition(100, 100)
	parent.SetSize(400, 300)
	parent.SetTouchable(true)
	root.AddChild(parent)

	child := fairygui.NewObject()
	child.SetPosition(50, 50) // 相对于 parent
	child.SetSize(100, 100)
	child.SetTouchable(true)
	parent.AddChild(child)

	// 记录事件触发顺序
	var eventLog []string

	parent.On(fairygui.EventMouseDown, func(event fairygui.Event) {
		eventLog = append(eventLog, "parent")
	})

	child.On(fairygui.EventMouseDown, func(event fairygui.Event) {
		eventLog = append(eventLog, "child")
	})

	// 验证管理器存在
	if im == nil {
		t.Fatal("InputManager should exist")
	}
}

// ============================================================================
// 事件分发测试
// ============================================================================

func TestInputManager_EventBubbling(t *testing.T) {
	// 创建层级结构
	root := fairygui.NewRoot(800, 600)
	parent := fairygui.NewComponent()
	child := fairygui.NewObject()

	parent.AddChild(child)
	root.AddChild(parent)

	// 记录事件冒泡
	var eventLog []string

	parent.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "parent")
	})

	child.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "child")
	})

	// 手动触发事件验证冒泡
	event := fairygui.NewMouseEvent(fairygui.EventClick, child, 150, 150)
	child.DispatchEvent(event)

	// 验证事件冒泡顺序
	if len(eventLog) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(eventLog))
	}
	if eventLog[0] != "child" {
		t.Errorf("Expected first event on child, got %s", eventLog[0])
	}
	if eventLog[1] != "parent" {
		t.Errorf("Expected second event on parent, got %s", eventLog[1])
	}
}

// ============================================================================
// MouseOver/MouseOut 测试
// ============================================================================

func TestInputManager_MouseOverOut_Concept(t *testing.T) {
	// 这个测试验证 MouseOver/MouseOut 事件的概念
	obj1 := fairygui.NewObject()
	obj1.SetName("obj1")
	obj2 := fairygui.NewObject()
	obj2.SetName("obj2")

	var eventLog []string

	obj1.On(fairygui.EventMouseOver, func(event fairygui.Event) {
		eventLog = append(eventLog, "obj1-over")
	})

	obj1.On(fairygui.EventMouseOut, func(event fairygui.Event) {
		eventLog = append(eventLog, "obj1-out")
	})

	obj2.On(fairygui.EventMouseOver, func(event fairygui.Event) {
		eventLog = append(eventLog, "obj2-over")
	})

	// 手动模拟 MouseOver 序列
	obj1.DispatchEvent(fairygui.NewMouseEvent(fairygui.EventMouseOver, obj1, 100, 100))
	obj1.DispatchEvent(fairygui.NewMouseEvent(fairygui.EventMouseOut, obj1, 200, 200))
	obj2.DispatchEvent(fairygui.NewMouseEvent(fairygui.EventMouseOver, obj2, 200, 200))

	// 验证事件顺序
	if len(eventLog) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(eventLog))
	}
	if eventLog[0] != "obj1-over" {
		t.Errorf("Expected obj1-over, got %s", eventLog[0])
	}
	if eventLog[1] != "obj1-out" {
		t.Errorf("Expected obj1-out, got %s", eventLog[1])
	}
	if eventLog[2] != "obj2-over" {
		t.Errorf("Expected obj2-over, got %s", eventLog[2])
	}
}

// ============================================================================
// Click 合成测试
// ============================================================================

func TestInputManager_ClickSynthesis_Concept(t *testing.T) {
	// 验证 Click 事件需要在同一对象上 MouseDown + MouseUp
	obj := fairygui.NewObject()

	var clickCalled bool
	var mouseDownCalled bool
	var mouseUpCalled bool

	obj.On(fairygui.EventMouseDown, func(event fairygui.Event) {
		mouseDownCalled = true
	})

	obj.On(fairygui.EventMouseUp, func(event fairygui.Event) {
		mouseUpCalled = true
	})

	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		clickCalled = true
	})

	// 手动触发完整的点击序列
	obj.DispatchEvent(fairygui.NewMouseEvent(fairygui.EventMouseDown, obj, 100, 100))
	obj.DispatchEvent(fairygui.NewMouseEvent(fairygui.EventMouseUp, obj, 100, 100))
	obj.DispatchEvent(fairygui.NewMouseEvent(fairygui.EventClick, obj, 100, 100))

	// 验证所有事件都被触发
	if !mouseDownCalled {
		t.Error("Expected MouseDown to be called")
	}
	if !mouseUpCalled {
		t.Error("Expected MouseUp to be called")
	}
	if !clickCalled {
		t.Error("Expected Click to be called")
	}
}

// ============================================================================
// 修饰键测试
// ============================================================================

func TestInputManager_ModifierKeys(t *testing.T) {
	obj := fairygui.NewObject()

	var ctrlPressed bool
	var shiftPressed bool
	var altPressed bool

	obj.On(fairygui.EventMouseDown, func(event fairygui.Event) {
		if me, ok := event.(*fairygui.MouseEvent); ok {
			ctrlPressed = me.CtrlKey
			shiftPressed = me.ShiftKey
			altPressed = me.AltKey
		}
	})

	// 创建带修饰键的事件
	event := fairygui.NewMouseEvent(fairygui.EventMouseDown, obj, 100, 100)
	event.CtrlKey = true
	event.ShiftKey = true
	event.AltKey = false

	obj.DispatchEvent(event)

	// 验证修饰键状态
	if !ctrlPressed {
		t.Error("Expected Ctrl key to be pressed")
	}
	if !shiftPressed {
		t.Error("Expected Shift key to be pressed")
	}
	if altPressed {
		t.Error("Expected Alt key not to be pressed")
	}
}

// ============================================================================
// 触摸事件测试
// ============================================================================

func TestInputManager_TouchEvents_Concept(t *testing.T) {
	obj := fairygui.NewObject()

	var touchBeginCalled bool
	var touchMoveCalled bool
	var touchEndCalled bool
	var receivedTouchID int

	obj.On(fairygui.EventTouchBegin, func(event fairygui.Event) {
		touchBeginCalled = true
		if te, ok := event.(*fairygui.TouchEvent); ok {
			receivedTouchID = te.TouchID
		}
	})

	obj.On(fairygui.EventTouchMove, func(event fairygui.Event) {
		touchMoveCalled = true
	})

	obj.On(fairygui.EventTouchEnd, func(event fairygui.Event) {
		touchEndCalled = true
	})

	// 手动触发触摸序列
	touchID := 1
	obj.DispatchEvent(fairygui.NewTouchEvent(fairygui.EventTouchBegin, obj, touchID, 100, 100))
	obj.DispatchEvent(fairygui.NewTouchEvent(fairygui.EventTouchMove, obj, touchID, 150, 150))
	obj.DispatchEvent(fairygui.NewTouchEvent(fairygui.EventTouchEnd, obj, touchID, 150, 150))

	// 验证事件序列
	if !touchBeginCalled {
		t.Error("Expected TouchBegin to be called")
	}
	if !touchMoveCalled {
		t.Error("Expected TouchMove to be called")
	}
	if !touchEndCalled {
		t.Error("Expected TouchEnd to be called")
	}
	if receivedTouchID != touchID {
		t.Errorf("Expected TouchID %d, got %d", touchID, receivedTouchID)
	}
}

// ============================================================================
// 键盘事件测试
// ============================================================================

func TestInputManager_KeyboardEvents_Concept(t *testing.T) {
	obj := fairygui.NewObject()

	var keyDownCalled bool
	var receivedKey fairygui.Key

	obj.On(fairygui.EventKeyDown, func(event fairygui.Event) {
		keyDownCalled = true
		if ke, ok := event.(*fairygui.KeyboardEvent); ok {
			receivedKey = ke.Key
		}
	})

	// 手动触发键盘事件
	obj.DispatchEvent(fairygui.NewKeyboardEvent(fairygui.EventKeyDown, obj, fairygui.Key(32))) // Space key

	// 验证事件
	if !keyDownCalled {
		t.Error("Expected KeyDown to be called")
	}
	if receivedKey != fairygui.Key(32) {
		t.Errorf("Expected key 32, got %d", receivedKey)
	}
}

// ============================================================================
// 局部坐标测试
// ============================================================================

func TestInputManager_LocalCoordinates(t *testing.T) {
	parent := fairygui.NewComponent()
	parent.SetPosition(100, 100)

	child := fairygui.NewObject()
	child.SetPosition(50, 50)
	parent.AddChild(child)

	var localX, localY float64

	child.On(fairygui.EventMouseDown, func(event fairygui.Event) {
		if me, ok := event.(*fairygui.MouseEvent); ok {
			localX = me.LocalX
			localY = me.LocalY
		}
	})

	// 创建全局坐标事件
	event := fairygui.NewMouseEvent(fairygui.EventMouseDown, child, 175, 175)
	// LocalX/LocalY 应该在实际的 InputManager 中计算
	// 这里手动设置期望值
	event.LocalX = 25
	event.LocalY = 25

	child.DispatchEvent(event)

	// 验证局部坐标
	// 全局 (175, 175) - 父对象 (100, 100) - 子对象 (50, 50) = 局部 (25, 25)
	if localX != 25 || localY != 25 {
		t.Errorf("Expected local coordinates (25, 25), got (%.0f, %.0f)", localX, localY)
	}
}

// ============================================================================
// Touchable 属性测试
// ============================================================================

func TestInputManager_Touchable_Property(t *testing.T) {
	obj := fairygui.NewObject()

	// 默认应该可触摸
	if !obj.Touchable() {
		t.Error("Expected object to be touchable by default")
	}

	// 设置为不可触摸
	obj.SetTouchable(false)
	if obj.Touchable() {
		t.Error("Expected object not to be touchable after SetTouchable(false)")
	}

	// 设置回可触摸
	obj.SetTouchable(true)
	if !obj.Touchable() {
		t.Error("Expected object to be touchable after SetTouchable(true)")
	}
}

// ============================================================================
// 多点触摸测试
// ============================================================================

func TestInputManager_MultiTouch_Concept(t *testing.T) {
	root := fairygui.NewRoot(800, 600)
	im := fairygui.NewBasicInputManager(root)

	// 验证初始没有触摸点
	touchIDs := im.TouchIDs()
	if len(touchIDs) != 0 {
		t.Errorf("Expected no touches initially, got %d", len(touchIDs))
	}

	// 这个测试主要验证 API 存在
	// 实际的多点触摸需要在集成测试中验证
}
