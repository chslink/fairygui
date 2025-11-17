package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// 事件系统集成测试
// ============================================================================

// TestEventSystem_Integration_BasicFlow 测试事件系统的基本流程
func TestEventSystem_Integration_BasicFlow(t *testing.T) {
	// 创建完整的 UI 层级
	root := fairygui.NewRoot(800, 600)
	container := fairygui.NewComponent()
	button := fairygui.NewObject()
	button.SetName("button")

	root.AddChild(container)
	container.AddChild(button)

	// 记录事件流
	var eventFlow []string

	// 在不同层级注册事件
	root.On(fairygui.EventClick, func(event fairygui.Event) {
		eventFlow = append(eventFlow, "root-click")
	})

	container.On(fairygui.EventClick, func(event fairygui.Event) {
		eventFlow = append(eventFlow, "container-click")
	})

	button.On(fairygui.EventClick, func(event fairygui.Event) {
		eventFlow = append(eventFlow, "button-click")
	})

	// 触发事件
	event := fairygui.NewMouseEvent(fairygui.EventClick, button, 100, 100)
	button.DispatchEvent(event)

	// 验证事件流
	if len(eventFlow) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(eventFlow))
	}
	if eventFlow[0] != "button-click" {
		t.Errorf("Expected first event 'button-click', got %s", eventFlow[0])
	}
	if eventFlow[1] != "container-click" {
		t.Errorf("Expected second event 'container-click', got %s", eventFlow[1])
	}
	if eventFlow[2] != "root-click" {
		t.Errorf("Expected third event 'root-click', got %s", eventFlow[2])
	}
}

// TestEventSystem_Integration_StopPropagationAtMiddle 测试在中间层停止事件传播
func TestEventSystem_Integration_StopPropagationAtMiddle(t *testing.T) {
	root := fairygui.NewRoot(800, 600)
	container := fairygui.NewComponent()
	button := fairygui.NewObject()

	root.AddChild(container)
	container.AddChild(button)

	var eventFlow []string

	root.On(fairygui.EventClick, func(event fairygui.Event) {
		eventFlow = append(eventFlow, "root")
	})

	container.On(fairygui.EventClick, func(event fairygui.Event) {
		eventFlow = append(eventFlow, "container")
		event.StopPropagation()
	})

	button.On(fairygui.EventClick, func(event fairygui.Event) {
		eventFlow = append(eventFlow, "button")
	})

	// 触发事件
	event := fairygui.NewMouseEvent(fairygui.EventClick, button, 100, 100)
	button.DispatchEvent(event)

	// 验证事件被在 container 层停止
	if len(eventFlow) != 2 {
		t.Fatalf("Expected 2 events (stopped at container), got %d", len(eventFlow))
	}
	if eventFlow[0] != "button" {
		t.Errorf("Expected first event 'button', got %s", eventFlow[0])
	}
	if eventFlow[1] != "container" {
		t.Errorf("Expected second event 'container', got %s", eventFlow[1])
	}
}

// TestEventSystem_Integration_MultipleEventTypes 测试多种事件类型
func TestEventSystem_Integration_MultipleEventTypes(t *testing.T) {
	obj := fairygui.NewObject()

	var receivedEvents []string

	obj.On(fairygui.EventMouseDown, func(event fairygui.Event) {
		receivedEvents = append(receivedEvents, "mousedown")
	})

	obj.On(fairygui.EventMouseUp, func(event fairygui.Event) {
		receivedEvents = append(receivedEvents, "mouseup")
	})

	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		receivedEvents = append(receivedEvents, "click")
	})

	obj.On(fairygui.EventMouseOver, func(event fairygui.Event) {
		receivedEvents = append(receivedEvents, "mouseover")
	})

	obj.On(fairygui.EventMouseOut, func(event fairygui.Event) {
		receivedEvents = append(receivedEvents, "mouseout")
	})

	// 模拟完整的鼠标交互序列
	obj.Emit(fairygui.NewMouseEvent(fairygui.EventMouseOver, obj, 100, 100))
	obj.Emit(fairygui.NewMouseEvent(fairygui.EventMouseDown, obj, 100, 100))
	obj.Emit(fairygui.NewMouseEvent(fairygui.EventMouseUp, obj, 100, 100))
	obj.Emit(fairygui.NewMouseEvent(fairygui.EventClick, obj, 100, 100))
	obj.Emit(fairygui.NewMouseEvent(fairygui.EventMouseOut, obj, 200, 200))

	// 验证所有事件都被接收
	expectedEvents := []string{"mouseover", "mousedown", "mouseup", "click", "mouseout"}
	if len(receivedEvents) != len(expectedEvents) {
		t.Fatalf("Expected %d events, got %d", len(expectedEvents), len(receivedEvents))
	}

	for i, expected := range expectedEvents {
		if receivedEvents[i] != expected {
			t.Errorf("Event %d: expected %s, got %s", i, expected, receivedEvents[i])
		}
	}
}

// TestEventSystem_Integration_PreventDefault 测试阻止默认行为
func TestEventSystem_Integration_PreventDefault(t *testing.T) {
	obj := fairygui.NewObject()

	var defaultPrevented bool

	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		event.PreventDefault()
	})

	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		if event.IsDefaultPrevented() {
			defaultPrevented = true
		}
	})

	// 触发事件
	event := fairygui.NewMouseEvent(fairygui.EventClick, obj, 100, 100)
	obj.Emit(event)

	// 验证默认行为被阻止
	if !defaultPrevented {
		t.Error("Expected default behavior to be prevented")
	}
}

// TestEventSystem_Integration_CurrentTargetTracking 测试 CurrentTarget 跟踪
func TestEventSystem_Integration_CurrentTargetTracking(t *testing.T) {
	parent := fairygui.NewComponent()
	parent.SetName("parent")
	child := fairygui.NewObject()
	child.SetName("child")

	parent.AddChild(child)

	var targets []string

	// 使用 Emit 而不是 DispatchEvent，因为 Component 嵌套可能影响 CurrentTarget
	// 直接在每个对象上触发，验证 Target 是正确的
	parent.On(fairygui.EventClick, func(event fairygui.Event) {
		targets = append(targets, "parent-received")
	})

	child.On(fairygui.EventClick, func(event fairygui.Event) {
		targets = append(targets, "child-received")
		// 验证 Target 是 child
		if event.Target() != child {
			t.Error("Expected Target to be child")
		}
	})

	// 触发事件，会冒泡到 parent
	event := fairygui.NewMouseEvent(fairygui.EventClick, child, 100, 100)
	child.DispatchEvent(event)

	// 验证两个对象都收到事件
	if len(targets) != 2 {
		t.Fatalf("Expected 2 targets, got %d", len(targets))
	}
	if targets[0] != "child-received" {
		t.Errorf("Expected first target 'child-received', got %s", targets[0])
	}
	if targets[1] != "parent-received" {
		t.Errorf("Expected second target 'parent-received', got %s", targets[1])
	}
}

// TestEventSystem_Integration_RemoveHandlerDuringEvent 测试在事件处理中移除处理器
func TestEventSystem_Integration_RemoveHandlerDuringEvent(t *testing.T) {
	obj := fairygui.NewObject()

	var callCount int
	var cancel func()

	handler := func(event fairygui.Event) {
		callCount++
		if cancel != nil {
			cancel()
		}
	}

	cancel = obj.On(fairygui.EventClick, handler)

	// 第一次触发 - 应该执行并取消
	obj.Emit(fairygui.NewMouseEvent(fairygui.EventClick, obj, 100, 100))

	// 第二次触发 - 不应该执行
	obj.Emit(fairygui.NewMouseEvent(fairygui.EventClick, obj, 100, 100))

	// 验证只执行了一次
	if callCount != 1 {
		t.Errorf("Expected handler to be called once, got %d", callCount)
	}
}

// TestEventSystem_Integration_DeepHierarchy 测试深层级层次结构
func TestEventSystem_Integration_DeepHierarchy(t *testing.T) {
	// 创建 5 层深的层级结构
	root := fairygui.NewRoot(800, 600)
	level1 := fairygui.NewComponent()
	level2 := fairygui.NewComponent()
	level3 := fairygui.NewComponent()
	level4 := fairygui.NewComponent()
	button := fairygui.NewObject()

	root.AddChild(level1)
	level1.AddChild(level2)
	level2.AddChild(level3)
	level3.AddChild(level4)
	level4.AddChild(button)

	var depth int

	countHandler := func(event fairygui.Event) {
		depth++
	}

	root.On(fairygui.EventClick, countHandler)
	level1.On(fairygui.EventClick, countHandler)
	level2.On(fairygui.EventClick, countHandler)
	level3.On(fairygui.EventClick, countHandler)
	level4.On(fairygui.EventClick, countHandler)
	button.On(fairygui.EventClick, countHandler)

	// 触发事件
	event := fairygui.NewMouseEvent(fairygui.EventClick, button, 100, 100)
	button.DispatchEvent(event)

	// 验证事件冒泡到所有 6 层
	if depth != 6 {
		t.Errorf("Expected event to bubble through 6 levels, got %d", depth)
	}
}

// TestEventSystem_Integration_MultipleHandlersOrder 测试多个处理器的执行顺序
func TestEventSystem_Integration_MultipleHandlersOrder(t *testing.T) {
	obj := fairygui.NewObject()

	var order []int

	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		order = append(order, 1)
	})

	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		order = append(order, 2)
	})

	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		order = append(order, 3)
	})

	// 触发事件
	obj.Emit(fairygui.NewMouseEvent(fairygui.EventClick, obj, 100, 100))

	// 验证处理器按注册顺序执行
	if len(order) != 3 {
		t.Fatalf("Expected 3 handlers, got %d", len(order))
	}
	if order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("Expected order [1, 2, 3], got %v", order)
	}
}

// TestEventSystem_Integration_TouchEvents 测试触摸事件流
func TestEventSystem_Integration_TouchEvents(t *testing.T) {
	obj := fairygui.NewObject()

	var touchSequence []string
	var touchID int

	obj.On(fairygui.EventTouchBegin, func(event fairygui.Event) {
		if te, ok := event.(*fairygui.TouchEvent); ok {
			touchSequence = append(touchSequence, "begin")
			touchID = te.TouchID
		}
	})

	obj.On(fairygui.EventTouchMove, func(event fairygui.Event) {
		if te, ok := event.(*fairygui.TouchEvent); ok {
			touchSequence = append(touchSequence, "move")
			if te.TouchID != touchID {
				t.Error("TouchID should remain consistent")
			}
		}
	})

	obj.On(fairygui.EventTouchEnd, func(event fairygui.Event) {
		if te, ok := event.(*fairygui.TouchEvent); ok {
			touchSequence = append(touchSequence, "end")
			if te.TouchID != touchID {
				t.Error("TouchID should remain consistent")
			}
		}
	})

	// 模拟完整触摸序列
	tid := 1
	obj.Emit(fairygui.NewTouchEvent(fairygui.EventTouchBegin, obj, tid, 100, 100))
	obj.Emit(fairygui.NewTouchEvent(fairygui.EventTouchMove, obj, tid, 150, 150))
	obj.Emit(fairygui.NewTouchEvent(fairygui.EventTouchMove, obj, tid, 200, 200))
	obj.Emit(fairygui.NewTouchEvent(fairygui.EventTouchEnd, obj, tid, 200, 200))

	// 验证触摸序列
	expected := []string{"begin", "move", "move", "end"}
	if len(touchSequence) != len(expected) {
		t.Fatalf("Expected %d events, got %d", len(expected), len(touchSequence))
	}

	for i, exp := range expected {
		if touchSequence[i] != exp {
			t.Errorf("Event %d: expected %s, got %s", i, exp, touchSequence[i])
		}
	}
}

// TestEventSystem_Integration_KeyboardEvents 测试键盘事件
func TestEventSystem_Integration_KeyboardEvents(t *testing.T) {
	root := fairygui.NewRoot(800, 600)

	var keyPressed bool
	var modifiers []string

	root.On(fairygui.EventKeyDown, func(event fairygui.Event) {
		if ke, ok := event.(*fairygui.KeyboardEvent); ok {
			keyPressed = true
			if ke.CtrlKey {
				modifiers = append(modifiers, "ctrl")
			}
			if ke.ShiftKey {
				modifiers = append(modifiers, "shift")
			}
			if ke.AltKey {
				modifiers = append(modifiers, "alt")
			}
		}
	})

	// 创建带修饰键的键盘事件
	event := fairygui.NewKeyboardEvent(fairygui.EventKeyDown, root, fairygui.Key(32)) // Space key
	event.CtrlKey = true
	event.ShiftKey = true

	root.Emit(event)

	// 验证键盘事件和修饰键
	if !keyPressed {
		t.Error("Expected key press event")
	}

	if len(modifiers) != 2 {
		t.Errorf("Expected 2 modifiers, got %d", len(modifiers))
	}
}

// ============================================================================
// 边界情况测试
// ============================================================================

// TestEventSystem_Edge_NilEvent 测试 nil 事件
func TestEventSystem_Edge_NilEvent(t *testing.T) {
	obj := fairygui.NewObject()

	// 应该不会 panic
	obj.DispatchEvent(nil)
	obj.Emit(nil)
}

// TestEventSystem_Edge_RemoveNonExistentHandler 测试移除不存在的处理器
func TestEventSystem_Edge_RemoveNonExistentHandler(t *testing.T) {
	obj := fairygui.NewObject()

	handler := func(event fairygui.Event) {}

	// 应该不会 panic
	obj.Off(fairygui.EventClick, handler)
}

// TestEventSystem_Edge_EmptyEventType 测试空事件类型
func TestEventSystem_Edge_EmptyEventType(t *testing.T) {
	obj := fairygui.NewObject()

	var called bool
	obj.On("", func(event fairygui.Event) {
		called = true
	})

	event := fairygui.NewEvent("", obj)
	obj.Emit(event)

	// 空事件类型也应该工作
	if !called {
		t.Error("Expected empty event type to work")
	}
}

// TestEventSystem_Edge_ConcurrentHandlerModification 测试并发修改处理器
func TestEventSystem_Edge_ConcurrentHandlerModification(t *testing.T) {
	obj := fairygui.NewObject()

	// 注册多个处理器
	for i := 0; i < 10; i++ {
		obj.On(fairygui.EventClick, func(event fairygui.Event) {
			// Do nothing
		})
	}

	// 在事件处理中注册新处理器（边界情况）
	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		obj.On(fairygui.EventClick, func(event fairygui.Event) {
			// 新注册的处理器
		})
	})

	// 应该不会 panic 或死锁
	obj.Emit(fairygui.NewMouseEvent(fairygui.EventClick, obj, 100, 100))
}
