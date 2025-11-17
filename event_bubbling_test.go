package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// 事件冒泡测试
// ============================================================================

func TestEventBubbling_Basic(t *testing.T) {
	// 创建三层级结构：grandparent -> parent -> child
	grandparent := fairygui.NewComponent()
	parent := fairygui.NewComponent()
	child := fairygui.NewObject()

	grandparent.AddChild(parent)
	parent.AddChild(child)

	// 记录事件触发顺序
	var eventLog []string

	grandparent.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "grandparent")
	})

	parent.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "parent")
	})

	child.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "child")
	})

	// 在子节点上分发事件
	event := fairygui.NewEvent(fairygui.EventClick, child)
	child.DispatchEvent(event)

	// 验证事件冒泡顺序：child -> parent -> grandparent
	if len(eventLog) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(eventLog))
	}
	if eventLog[0] != "child" {
		t.Errorf("Expected first event on child, got %s", eventLog[0])
	}
	if eventLog[1] != "parent" {
		t.Errorf("Expected second event on parent, got %s", eventLog[1])
	}
	if eventLog[2] != "grandparent" {
		t.Errorf("Expected third event on grandparent, got %s", eventLog[2])
	}
}

func TestEventBubbling_StopPropagation(t *testing.T) {
	// 创建三层级结构
	grandparent := fairygui.NewComponent()
	parent := fairygui.NewComponent()
	child := fairygui.NewObject()

	grandparent.AddChild(parent)
	parent.AddChild(child)

	// 记录事件触发
	var eventLog []string

	grandparent.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "grandparent")
	})

	parent.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "parent")
		// 在父节点停止传播
		event.StopPropagation()
	})

	child.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "child")
	})

	// 分发事件
	event := fairygui.NewEvent(fairygui.EventClick, child)
	child.DispatchEvent(event)

	// 验证事件在 parent 停止，没有到达 grandparent
	if len(eventLog) != 2 {
		t.Fatalf("Expected 2 events (stopped at parent), got %d", len(eventLog))
	}
	if eventLog[0] != "child" {
		t.Errorf("Expected first event on child, got %s", eventLog[0])
	}
	if eventLog[1] != "parent" {
		t.Errorf("Expected second event on parent, got %s", eventLog[1])
	}
}

func TestEventBubbling_CurrentTarget(t *testing.T) {
	// 创建层级结构
	parent := fairygui.NewComponent()
	parent.SetName("parent")
	child := fairygui.NewObject()
	child.SetName("child")

	parent.AddChild(child)

	// 验证 CurrentTarget 在冒泡过程中的变化
	var childCurrentTargetName string
	var parentCurrentTargetName string

	child.On(fairygui.EventClick, func(event fairygui.Event) {
		if target, ok := event.CurrentTarget().(fairygui.DisplayObject); ok {
			childCurrentTargetName = target.Name()
		}
	})

	parent.On(fairygui.EventClick, func(event fairygui.Event) {
		if target, ok := event.CurrentTarget().(fairygui.DisplayObject); ok {
			parentCurrentTargetName = target.Name()
		}
	})

	// 分发事件
	event := fairygui.NewEvent(fairygui.EventClick, child)
	child.DispatchEvent(event)

	// 验证 CurrentTarget（通过 Name 比较）
	if childCurrentTargetName != "child" {
		t.Errorf("Expected child's CurrentTarget name to be 'child', got %s", childCurrentTargetName)
	}
	if parentCurrentTargetName != "parent" {
		t.Errorf("Expected parent's CurrentTarget name to be 'parent', got %s", parentCurrentTargetName)
	}

	// 验证 Target 始终是原始目标
	if event.Target() != child {
		t.Error("Expected event Target to always be child")
	}
}

func TestEventBubbling_NoParent(t *testing.T) {
	// 创建单独的对象（没有父节点）
	obj := fairygui.NewObject()

	var called bool
	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		called = true
	})

	// 分发事件
	event := fairygui.NewEvent(fairygui.EventClick, obj)
	obj.DispatchEvent(event)

	// 验证事件被触发
	if !called {
		t.Error("Expected event to be triggered even without parent")
	}
}

func TestEventBubbling_MultipleHandlers(t *testing.T) {
	// 创建层级结构
	parent := fairygui.NewComponent()
	child := fairygui.NewObject()

	parent.AddChild(child)

	// 每个节点注册多个处理器
	var eventLog []string

	child.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "child-1")
	})
	child.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "child-2")
	})

	parent.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "parent-1")
	})
	parent.On(fairygui.EventClick, func(event fairygui.Event) {
		eventLog = append(eventLog, "parent-2")
	})

	// 分发事件
	event := fairygui.NewEvent(fairygui.EventClick, child)
	child.DispatchEvent(event)

	// 验证所有处理器都被调用
	if len(eventLog) != 4 {
		t.Fatalf("Expected 4 events, got %d", len(eventLog))
	}

	// 验证顺序：child 的两个处理器先执行，然后是 parent 的
	if eventLog[0] != "child-1" || eventLog[1] != "child-2" {
		t.Error("Expected child handlers to execute first")
	}
	if eventLog[2] != "parent-1" || eventLog[3] != "parent-2" {
		t.Error("Expected parent handlers to execute after child")
	}
}

func TestEventBubbling_MouseEvent(t *testing.T) {
	// 创建层级结构
	parent := fairygui.NewComponent()
	child := fairygui.NewObject()

	parent.AddChild(child)

	// 测试鼠标事件的冒泡
	var parentReceived bool
	var mouseX, mouseY float64

	parent.On(fairygui.EventMouseDown, func(event fairygui.Event) {
		parentReceived = true
		if me, ok := event.(*fairygui.MouseEvent); ok {
			mouseX = me.X
			mouseY = me.Y
		}
	})

	// 分发鼠标事件
	event := fairygui.NewMouseEvent(fairygui.EventMouseDown, child, 100, 200)
	child.DispatchEvent(event)

	// 验证父节点接收到事件和正确的坐标
	if !parentReceived {
		t.Error("Expected parent to receive mouse event")
	}
	if mouseX != 100 || mouseY != 200 {
		t.Errorf("Expected mouse position (100, 200), got (%.0f, %.0f)", mouseX, mouseY)
	}
}

// ============================================================================
// Emit 方法测试（不冒泡）
// ============================================================================

func TestEmit_NoBubbling(t *testing.T) {
	// 创建层级结构
	parent := fairygui.NewComponent()
	child := fairygui.NewObject()

	parent.AddChild(child)

	// 记录事件
	var parentCalled bool
	var childCalled bool

	parent.On(fairygui.EventClick, func(event fairygui.Event) {
		parentCalled = true
	})

	child.On(fairygui.EventClick, func(event fairygui.Event) {
		childCalled = true
	})

	// 使用 Emit（不冒泡）
	event := fairygui.NewEvent(fairygui.EventClick, child)
	child.Emit(event)

	// 验证只有子节点接收到事件
	if !childCalled {
		t.Error("Expected child to receive event")
	}
	if parentCalled {
		t.Error("Expected parent NOT to receive event when using Emit (no bubbling)")
	}
}
