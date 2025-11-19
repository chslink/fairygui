package fairygui

import (
	"testing"
)

// TestNewScrollBar 测试创建新的滚动条
func TestNewScrollBar(t *testing.T) {
	sb := NewScrollBar()
	if sb == nil {
		t.Fatal("NewScrollBar() returned nil")
	}

	if sb.ComponentImpl == nil {
		t.Error("ScrollBar.ComponentImpl is nil")
	}

	if sb.ScrollPerc() != 0 {
		t.Errorf("默认滚动百分比不正确: got %f, want 0", sb.ScrollPerc())
	}

	if sb.DisplayPerc() != 1 {
		t.Errorf("默认显示百分比不正确: got %f, want 1", sb.DisplayPerc())
	}

	if sb.IsVertical() {
		t.Error("新创建的 ScrollBar 默认应该为水平方向")
	}

	if sb.IsDragging() {
		t.Error("新创建的 ScrollBar 不应该处于拖拽状态")
	}
}

// TestScrollBar_SetScrollPane 测试绑定滚动目标
func TestScrollBar_SetScrollPane(t *testing.T) {
	sb := NewScrollBar()
	pane := NewScrollPaneV2()

	sb.SetScrollPane(pane, true)

	if !sb.IsVertical() {
		t.Error("设置后应该为垂直方向")
	}
}

// TestScrollBar_SyncFromPane 测试从 ScrollPane 同步状态
func TestScrollBar_SyncFromPane(t *testing.T) {
	sb := NewScrollBar()
	pane := NewScrollPaneV2()

	sb.SetScrollPane(pane, false)

	// 模拟滚动信息
	info := ScrollInfo{
		PercX:        0.5,
		PercY:        0.3,
		DisplayPercX: 0.8,
		DisplayPercY: 0.6,
	}

	sb.SyncFromPane(info)

	if sb.ScrollPerc() != 0.5 {
		t.Errorf("滚动百分比同步失败: got %f, want 0.5", sb.ScrollPerc())
	}

	if sb.DisplayPerc() != 0.8 {
		t.Errorf("显示百分比同步失败: got %f, want 0.8", sb.DisplayPerc())
	}
}

// TestScrollBar_VerticalHorizontal 测试垂直和水平模式
func TestScrollBar_VerticalHorizontal(t *testing.T) {
	sb := NewScrollBar()
	pane := NewScrollPaneV2()

	// 测试垂直模式
	sb.SetScrollPane(pane, true)
	if !sb.IsVertical() {
		t.Error("应该为垂直方向")
	}

	info := ScrollInfo{
		PercY:        0.4,
		DisplayPercY: 0.7,
	}
	sb.SyncFromPane(info)

	if sb.ScrollPerc() != 0.4 {
		t.Errorf("垂直模式滚动百分比错误: got %f", sb.ScrollPerc())
	}

	// 测试水平模式
	sb.SetScrollPane(pane, false)
	if sb.IsVertical() {
		t.Error("应该为水平方向")
	}

	info2 := ScrollInfo{
		PercX:        0.6,
		DisplayPercX: 0.5,
	}
	sb.SyncFromPane(info2)

	if sb.ScrollPerc() != 0.6 {
		t.Errorf("水平模式滚动百分比错误: got %f", sb.ScrollPerc())
	}
}

// TestScrollBar_SetFixedGrip 测试固定尺寸滑块
func TestScrollBar_SetFixedGrip(t *testing.T) {
	sb := NewScrollBar()
	sb.SetFixedGrip(true)

	pane := NewScrollPaneV2()
	sb.SetScrollPane(pane, true)

	info := ScrollInfo{
		PercY:        0.5,
		DisplayPercY: 0.3, // displayPerc 不应该影响滑块大小
	}
	sb.SyncFromPane(info)

	if sb.ScrollPerc() != 0.5 {
		t.Errorf("滚动百分比错误: got %f", sb.ScrollPerc())
	}

	if sb.DisplayPerc() != 0.3 {
		t.Errorf("显示百分比错误: got %f", sb.DisplayPerc())
	}
}

// TestScrollBar_ScrollPercRange 测试滚动百分比范围限制
func TestScrollBar_ScrollPercRange(t *testing.T) {
	sb := NewScrollBar()
	pane := NewScrollPaneV2()
	sb.SetScrollPane(pane, true)

	// 测试小于 0
	info := ScrollInfo{
		PercY:        -0.5,
		DisplayPercY: 0.5,
	}
	sb.SyncFromPane(info)

	if sb.ScrollPerc() != 0 {
		t.Errorf("滚动百分比应该限制为 0: got %f", sb.ScrollPerc())
	}

	// 测试大于 1
	info2 := ScrollInfo{
		PercY:        1.5,
		DisplayPercY: 0.5,
	}
	sb.SyncFromPane(info2)

	if sb.ScrollPerc() != 1 {
		t.Errorf("滚动百分比应该限制为 1: got %f", sb.ScrollPerc())
	}
}

// TestScrollBar_DisplayPercRange 测试显示百分比范围限制
func TestScrollBar_DisplayPercRange(t *testing.T) {
	sb := NewScrollBar()
	pane := NewScrollPaneV2()
	sb.SetScrollPane(pane, true)

	// 测试小于 0
	info := ScrollInfo{
		PercY:        0.5,
		DisplayPercY: -0.5,
	}
	sb.SyncFromPane(info)

	if sb.DisplayPerc() != 0 {
		t.Errorf("显示百分比应该限制为 0: got %f", sb.DisplayPerc())
	}

	// 测试大于 1
	info2 := ScrollInfo{
		PercY:        0.5,
		DisplayPercY: 1.5,
	}
	sb.SyncFromPane(info2)

	if sb.DisplayPerc() != 1 {
		t.Errorf("显示百分比应该限制为 1: got %f", sb.DisplayPerc())
	}
}

// TestScrollBar_PackageItem 测试资源项操作
func TestScrollBar_PackageItem(t *testing.T) {
	sb := NewScrollBar()

	// 测试空值 - 应该接受nil而不panic
	sb.SetPackageItem(nil)
	// PackageItem() 返回 interface{}，即使内部是nil，也不直接等于nil
	// 所以这里只测试不panic即可
}

// TestScrollBar_Chaining 测试链式调用
func TestScrollBar_Chaining(t *testing.T) {
	sb := NewScrollBar()
	pane := NewScrollPaneV2()

	// 注意：这里的链式调用测试主要是为了验证不报错
	// ScrollBar 的方法大多不返回自身，所以链式调用有限
	sb.SetFixedGrip(true)
	sb.SetScrollPane(pane, true)

	info := ScrollInfo{
		PercY:        0.5,
		DisplayPercY: 0.5,
	}
	sb.SyncFromPane(info)

	if sb.ScrollPerc() != 0.5 {
		t.Error("链式调用后状态错误")
	}
}

// TestAssertScrollBar 测试类型断言
func TestAssertScrollBar(t *testing.T) {
	sb := NewScrollBar()

	result, ok := AssertScrollBar(sb)
	if !ok {
		t.Error("AssertScrollBar 应该成功")
	}
	if result != sb {
		t.Error("AssertScrollBar 返回的对象不正确")
	}

	if !IsScrollBar(sb) {
		t.Error("IsScrollBar 应该返回 true")
	}

	obj := NewObject()
	_, ok = AssertScrollBar(obj)
	if ok {
		t.Error("AssertScrollBar 对非 ScrollBar 对象应该失败")
	}

	if IsScrollBar(obj) {
		t.Error("IsScrollBar 对非 ScrollBar 对象应该返回 false")
	}
}
