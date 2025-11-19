package fairygui

import (
	"testing"
)

// TestNewScrollPaneV2 测试创建新的滚动面板
func TestNewScrollPaneV2(t *testing.T) {
	pane := NewScrollPaneV2()
	if pane == nil {
		t.Fatal("NewScrollPaneV2() returned nil")
	}

	if pane.ComponentImpl == nil {
		t.Error("ScrollPaneV2.ComponentImpl is nil")
	}

	// 验证默认值
	if pane.scrollStep != 10 {
		t.Errorf("默认滚动步长不正确: got %f, want 10", pane.scrollStep)
	}
}

// TestScrollPaneV2_AddScrollListener 测试添加滚动监听器
func TestScrollPaneV2_AddScrollListener(t *testing.T) {
	pane := NewScrollPaneV2()

	called := false
	var receivedInfo ScrollInfo

	listenerID := pane.AddScrollListener(func(info ScrollInfo) {
		called = true
		receivedInfo = info
	})

	if listenerID != 0 {
		t.Errorf("第一个监听器ID应该为0: got %d", listenerID)
	}

	// 触发滚动通知
	pane.notifyListeners()

	if !called {
		t.Error("监听器没有被调用")
	}

	if receivedInfo.PercX != 0 || receivedInfo.PercY != 0 {
		t.Error("初始滚动信息不正确")
	}

	// 添加第二个监听器
	listenerID2 := pane.AddScrollListener(func(info ScrollInfo) {})
	if listenerID2 != 1 {
		t.Errorf("第二个监听器ID应该为1: got %d", listenerID2)
	}
}

// TestScrollPaneV2_RemoveScrollListener 测试移除滚动监听器
func TestScrollPaneV2_RemoveScrollListener(t *testing.T) {
	pane := NewScrollPaneV2()

	// 目前只是桩实现，测试不报错即可
	listenerID := pane.AddScrollListener(func(info ScrollInfo) {})
	pane.RemoveScrollListener(listenerID)
}

// TestScrollPaneV2_SetPercX 测试设置水平滚动百分比
func TestScrollPaneV2_SetPercX(t *testing.T) {
	pane := NewScrollPaneV2()

	var receivedInfo ScrollInfo
	pane.AddScrollListener(func(info ScrollInfo) {
		receivedInfo = info
	})

	// 测试正常值
	pane.SetPercX(0.5, false)

	if pane.scrollPercX != 0.5 {
		t.Errorf("滚动百分比设置失败: got %f, want 0.5", pane.scrollPercX)
	}

	if receivedInfo.PercX != 0.5 {
		t.Errorf("监听器接收到的PercX不正确: got %f", receivedInfo.PercX)
	}

	// 测试边界值
	pane.SetPercX(0, false)
	if pane.scrollPercX != 0 {
		t.Errorf("滚动百分比设置失败: got %f", pane.scrollPercX)
	}

	pane.SetPercX(1, false)
	if pane.scrollPercX != 1 {
		t.Errorf("滚动百分比设置失败: got %f", pane.scrollPercX)
	}
}

// TestScrollPaneV2_SetPercY 测试设置垂直滚动百分比
func TestScrollPaneV2_SetPercY(t *testing.T) {
	pane := NewScrollPaneV2()

	var receivedInfo ScrollInfo
	pane.AddScrollListener(func(info ScrollInfo) {
		receivedInfo = info
	})

	// 测试正常值
	pane.SetPercY(0.7, false)

	if pane.scrollPercY != 0.7 {
		t.Errorf("滚动百分比设置失败: got %f, want 0.7", pane.scrollPercY)
	}

	if receivedInfo.PercY != 0.7 {
		t.Errorf("监听器接收到的PercY不正确: got %f", receivedInfo.PercY)
	}

	// 测试边界值
	pane.SetPercY(0, false)
	if pane.scrollPercY != 0 {
		t.Errorf("滚动百分比设置失败: got %f", pane.scrollPercY)
	}

	pane.SetPercY(1, false)
	if pane.scrollPercY != 1 {
		t.Errorf("滚动百分比设置失败: got %f", pane.scrollPercY)
	}
}

// TestScrollPaneV2_ScrollUp 测试向上滚动
func TestScrollPaneV2_ScrollUp(t *testing.T) {
	pane := NewScrollPaneV2()

	// 初始位置 0
	if pane.perY != 0 {
		t.Errorf("初始Y位置应该为0: got %f", pane.perY)
	}

	// 向下滚动一点
	pane.SetPercY(0.5, false)
	if pane.perY != 0.5 {
		t.Errorf("设置Y位置失败: got %f", pane.perY)
	}

	// 向上滚动
	pane.ScrollUp()

	// 应该减少 scrollStep / 100
	expected := 0.5 - 0.1 // scrollStep = 10, 所以 10/100 = 0.1
	if pane.perY != expected {
		t.Errorf("向上滚动失败: got %f, want %f", pane.perY, expected)
	}

	// 测试最小值限制
	pane.SetPercY(0, false)
	pane.ScrollUp()
	if pane.perY != 0 {
		t.Errorf("向上滚动应该限制在0: got %f", pane.perY)
	}
}

// TestScrollPaneV2_ScrollDown 测试向下滚动
func TestScrollPaneV2_ScrollDown(t *testing.T) {
	pane := NewScrollPaneV2()

	// 初始位置 0
	if pane.perY != 0 {
		t.Errorf("初始Y位置应该为0: got %f", pane.perY)
	}

	// 向下滚动
	pane.ScrollDown()

	// 应该增加 scrollStep / 100
	expected := 0.1 // scrollStep = 10, 所以 10/100 = 0.1
	if pane.perY != expected {
		t.Errorf("向下滚动失败: got %f, want %f", pane.perY, expected)
	}

	// 测试最大值限制
	pane.SetPercY(1, false)
	pane.ScrollDown()
	if pane.perY != 1 {
		t.Errorf("向下滚动应该限制在1: got %f", pane.perY)
	}
}

// TestScrollPaneV2_ScrollLeft 测试向左滚动
func TestScrollPaneV2_ScrollLeft(t *testing.T) {
	pane := NewScrollPaneV2()

	// 初始位置 0
	if pane.perX != 0 {
		t.Errorf("初始X位置应该为0: got %f", pane.perX)
	}

	// 向右滚动一点
	pane.SetPercX(0.5, false)
	if pane.perX != 0.5 {
		t.Errorf("设置X位置失败: got %f", pane.perX)
	}

	// 向左滚动
	pane.ScrollLeft()

	// 应该减少 scrollStep / 100
	expected := 0.5 - 0.1 // scrollStep = 10, 所以 10/100 = 0.1
	if pane.perX != expected {
		t.Errorf("向左滚动失败: got %f, want %f", pane.perX, expected)
	}

	// 测试最小值限制
	pane.SetPercX(0, false)
	pane.ScrollLeft()
	if pane.perX != 0 {
		t.Errorf("向左滚动应该限制在0: got %f", pane.perX)
	}
}

// TestScrollPaneV2_ScrollRight 测试向右滚动
func TestScrollPaneV2_ScrollRight(t *testing.T) {
	pane := NewScrollPaneV2()

	// 初始位置 0
	if pane.perX != 0 {
		t.Errorf("初始X位置应该为0: got %f", pane.perX)
	}

	// 向右滚动
	pane.ScrollRight()

	// 应该增加 scrollStep / 100
	expected := 0.1 // scrollStep = 10, 所以 10/100 = 0.1
	if pane.perX != expected {
		t.Errorf("向右滚动失败: got %f, want %f", pane.perX, expected)
	}

	// 测试最大值限制
	pane.SetPercX(1, false)
	pane.ScrollRight()
	if pane.perX != 1 {
		t.Errorf("向右滚动应该限制在1: got %f", pane.perX)
	}
}

// TestScrollPaneV2_MultipleListeners 测试多个监听器
func TestScrollPaneV2_MultipleListeners(t *testing.T) {
	pane := NewScrollPaneV2()

	callCount := 0
	listener1 := func(info ScrollInfo) {
		callCount++
	}
	listener2 := func(info ScrollInfo) {
		callCount++
	}

	pane.AddScrollListener(listener1)
	pane.AddScrollListener(listener2)

	pane.SetPercX(0.5, false)

	if callCount != 2 {
		t.Errorf("应该调用2个监听器: got %d", callCount)
	}
}

// TestScrollPaneV2_GetScrollPos 测试获取滚动位置
func TestScrollPaneV2_GetScrollPos(t *testing.T) {
	pane := NewScrollPaneV2()

	pane.SetPercX(0.5, false)
	pane.SetPercY(0.7, false)

	pos := pane.GetScrollPos()

	// GetScrollPos 返回 perX*100 和 perY*100
	if pos.X != 50 {
		t.Errorf("GetScrollPos X 不正确: got %f, want 50", pos.X)
	}

	if pos.Y != 70 {
		t.Errorf("GetScrollPos Y 不正确: got %f, want 70", pos.Y)
	}
}

// TestScrollPaneV2_GetViewSize 测试获取视口大小
func TestScrollPaneV2_GetViewSize(t *testing.T) {
	pane := NewScrollPaneV2()

	size := pane.GetViewSize()

	// 桩实现返回 {100, 100}
	if size.X != 100 || size.Y != 100 {
		t.Errorf("GetViewSize 返回值不正确: got %v", size)
	}
}

// TestScrollPaneV2_GetFirstChildInView 测试获取第一个可见子对象
func TestScrollPaneV2_GetFirstChildInView(t *testing.T) {
	pane := NewScrollPaneV2()

	// 桩实现返回 0
	index := pane.GetFirstChildInView()
	if index != 0 {
		t.Errorf("GetFirstChildInView 应该返回 0: got %d", index)
	}
}

// TestScrollPaneV2_Stubs 测试桩方法
func TestScrollPaneV2_Stubs(t *testing.T) {
	pane := NewScrollPaneV2()

	// 这些方法是桩实现，测试不panic即可
	pane.ScrollToView(5, true)
	pane.Clear()
	pane.SetBoundsChanged()
	pane.EnsureSizeCorrect()
}

// TestScrollPaneV2_Chaining 测试链式调用效果
func TestScrollPaneV2_Chaining(t *testing.T) {
	pane := NewScrollPaneV2()

	// 连续的滚动操作
	pane.ScrollDown()
	pane.ScrollDown()
	pane.ScrollRight()
	pane.ScrollLeft()

	// 最终位置应该是
	// Y: 0.1 + 0.1 = 0.2
	// X: 0.1 (ScrollRight) - 0.1 (ScrollLeft) = 0
	if pane.perY != 0.2 {
		t.Errorf("链式调用后Y位置错误: got %f, want 0.2", pane.perY)
	}

	if pane.perX != 0 {
		t.Errorf("链式调用后X位置错误: got %f, want 0", pane.perX)
	}
}
