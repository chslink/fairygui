package fairygui_test

import (
	"testing"
	"time"

	"github.com/chslink/fairygui"
	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// Object 测试
// ============================================================================

func TestObject_Creation(t *testing.T) {
	obj := fairygui.NewObject()
	if obj == nil {
		t.Fatal("NewObject() returned nil")
	}

	if obj.ID() == "" {
		t.Error("Object ID should not be empty")
	}

	// 默认值测试
	if !obj.Visible() {
		t.Error("Object should be visible by default")
	}

	if obj.Alpha() != 1.0 {
		t.Errorf("Expected alpha 1.0, got %f", obj.Alpha())
	}

	scaleX, scaleY := obj.Scale()
	if scaleX != 1.0 || scaleY != 1.0 {
		t.Errorf("Expected scale (1.0, 1.0), got (%f, %f)", scaleX, scaleY)
	}
}

func TestObject_Position(t *testing.T) {
	obj := fairygui.NewObject()

	// 设置位置
	obj.SetPosition(100, 200)

	x, y := obj.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", x, y)
	}
}

func TestObject_Size(t *testing.T) {
	obj := fairygui.NewObject()

	// 设置尺寸
	obj.SetSize(300, 400)

	width, height := obj.Size()
	if width != 300 || height != 400 {
		t.Errorf("Expected size (300, 400), got (%f, %f)", width, height)
	}
}

func TestObject_Transform(t *testing.T) {
	obj := fairygui.NewObject()

	// 测试缩放
	obj.SetScale(2.0, 3.0)
	scaleX, scaleY := obj.Scale()
	if scaleX != 2.0 || scaleY != 3.0 {
		t.Errorf("Expected scale (2.0, 3.0), got (%f, %f)", scaleX, scaleY)
	}

	// 测试旋转（度）
	obj.SetRotation(90)
	rotation := obj.Rotation()
	if rotation != 90 {
		t.Errorf("Expected rotation 90, got %f", rotation)
	}

	// 测试锚点
	obj.SetPivot(0.5, 0.5)
	pivotX, pivotY := obj.Pivot()
	if pivotX != 0.5 || pivotY != 0.5 {
		t.Errorf("Expected pivot (0.5, 0.5), got (%f, %f)", pivotX, pivotY)
	}
}

func TestObject_Visibility(t *testing.T) {
	obj := fairygui.NewObject()

	// 测试可见性
	obj.SetVisible(false)
	if obj.Visible() {
		t.Error("Object should be invisible")
	}

	obj.SetVisible(true)
	if !obj.Visible() {
		t.Error("Object should be visible")
	}

	// 测试透明度
	obj.SetAlpha(0.5)
	if obj.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %f", obj.Alpha())
	}

	// 测试透明度边界
	obj.SetAlpha(-1)
	if obj.Alpha() != 0 {
		t.Errorf("Alpha should be clamped to 0, got %f", obj.Alpha())
	}

	obj.SetAlpha(2)
	if obj.Alpha() != 1 {
		t.Errorf("Alpha should be clamped to 1, got %f", obj.Alpha())
	}
}

func TestObject_Touchable(t *testing.T) {
	obj := fairygui.NewObject()

	// 默认应该可触摸
	if !obj.Touchable() {
		t.Error("Object should be touchable by default")
	}

	obj.SetTouchable(false)
	if obj.Touchable() {
		t.Error("Object should not be touchable")
	}
}

func TestObject_NameAndData(t *testing.T) {
	obj := fairygui.NewObject()

	// 测试名称
	obj.SetName("TestObject")
	if obj.Name() != "TestObject" {
		t.Errorf("Expected name 'TestObject', got '%s'", obj.Name())
	}

	// 测试自定义数据
	data := map[string]interface{}{"key": "value"}
	obj.SetData(data)
	retrievedData := obj.Data()
	if retrievedData == nil {
		t.Error("Data should not be nil")
	}
	// 验证数据内容
	if dataMap, ok := retrievedData.(map[string]interface{}); ok {
		if dataMap["key"] != "value" {
			t.Error("Custom data content mismatch")
		}
	} else {
		t.Error("Data should be a map[string]interface{}")
	}
}

// ============================================================================
// 层级关系测试
// ============================================================================

func TestObject_Hierarchy(t *testing.T) {
	parent := fairygui.NewObject()
	child1 := fairygui.NewObject()
	child2 := fairygui.NewObject()

	// 添加子对象
	parent.AddChild(child1)
	parent.AddChild(child2)

	// 验证子对象数量
	if parent.ChildCount() != 2 {
		t.Errorf("Expected 2 children, got %d", parent.ChildCount())
	}

	// 验证父对象
	if child1.Parent() != parent {
		t.Error("child1 parent mismatch")
	}

	// 验证获取子对象
	if parent.GetChildAt(0) != child1 {
		t.Error("GetChildAt(0) should return child1")
	}

	if parent.GetChildAt(1) != child2 {
		t.Error("GetChildAt(1) should return child2")
	}

	// 移除子对象
	parent.RemoveChild(child1)
	if parent.ChildCount() != 1 {
		t.Errorf("Expected 1 child after removal, got %d", parent.ChildCount())
	}

	if child1.Parent() != nil {
		t.Error("child1 should have no parent after removal")
	}
}

func TestObject_AddChildAt(t *testing.T) {
	parent := fairygui.NewObject()
	child1 := fairygui.NewObject()
	child2 := fairygui.NewObject()
	child3 := fairygui.NewObject()

	parent.AddChild(child1)
	parent.AddChild(child3)

	// 在中间插入
	parent.AddChildAt(child2, 1)

	if parent.GetChildAt(0) != child1 {
		t.Error("Child at index 0 should be child1")
	}
	if parent.GetChildAt(1) != child2 {
		t.Error("Child at index 1 should be child2")
	}
	if parent.GetChildAt(2) != child3 {
		t.Error("Child at index 2 should be child3")
	}
}

func TestObject_GetChildByName(t *testing.T) {
	parent := fairygui.NewObject()
	child := fairygui.NewObject()
	child.SetName("TestChild")

	parent.AddChild(child)

	found := parent.GetChildByName("TestChild")
	if found != child {
		t.Error("GetChildByName should return the correct child")
	}

	notFound := parent.GetChildByName("NonExistent")
	if notFound != nil {
		t.Error("GetChildByName should return nil for non-existent name")
	}
}

func TestObject_RemoveChildAt(t *testing.T) {
	parent := fairygui.NewObject()
	child1 := fairygui.NewObject()
	child2 := fairygui.NewObject()

	parent.AddChild(child1)
	parent.AddChild(child2)

	removed := parent.RemoveChildAt(0)
	if removed != child1 {
		t.Error("RemoveChildAt(0) should return child1")
	}

	if parent.ChildCount() != 1 {
		t.Errorf("Expected 1 child, got %d", parent.ChildCount())
	}

	if parent.GetChildAt(0) != child2 {
		t.Error("Remaining child should be child2")
	}
}

// ============================================================================
// 事件系统测试
// ============================================================================

func TestObject_Events(t *testing.T) {
	obj := fairygui.NewObject()

	called := false
	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		called = true
	})

	// 触发事件
	obj.Emit(fairygui.NewEvent(fairygui.EventClick, obj))

	if !called {
		t.Error("Event handler should have been called")
	}
}

func TestObject_EventOnce(t *testing.T) {
	obj := fairygui.NewObject()

	callCount := 0
	obj.Once(fairygui.EventClick, func(event fairygui.Event) {
		callCount++
	})

	// 触发两次
	obj.Emit(fairygui.NewEvent(fairygui.EventClick, obj))
	obj.Emit(fairygui.NewEvent(fairygui.EventClick, obj))

	if callCount != 1 {
		t.Errorf("Once handler should be called only once, got %d calls", callCount)
	}
}

func TestObject_OnClick(t *testing.T) {
	obj := fairygui.NewObject()

	clicked := false
	obj.OnClick(func() {
		clicked = true
	})

	obj.Emit(fairygui.NewEvent(fairygui.EventClick, obj))

	if !clicked {
		t.Error("OnClick handler should have been called")
	}
}

func TestObject_EventStopPropagation(t *testing.T) {
	obj := fairygui.NewObject()

	firstCalled := false
	secondCalled := false

	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		firstCalled = true
		event.StopPropagation()
	})

	obj.On(fairygui.EventClick, func(event fairygui.Event) {
		secondCalled = true
	})

	event := fairygui.NewEvent(fairygui.EventClick, obj)
	obj.Emit(event)

	if !firstCalled {
		t.Error("First handler should have been called")
	}

	if secondCalled {
		t.Error("Second handler should not have been called after StopPropagation")
	}
}

func TestObject_HasListener(t *testing.T) {
	obj := fairygui.NewObject()

	if obj.HasListener(fairygui.EventClick) {
		t.Error("Should not have listener initially")
	}

	obj.On(fairygui.EventClick, func(event fairygui.Event) {})

	if !obj.HasListener(fairygui.EventClick) {
		t.Error("Should have listener after On()")
	}
}

// ============================================================================
// Component 测试
// ============================================================================

func TestComponent_Creation(t *testing.T) {
	comp := fairygui.NewComponent()
	if comp == nil {
		t.Fatal("NewComponent() returned nil")
	}

	// Component 应该实现 Component 接口
	var _ fairygui.Component = comp

	// 默认应该没有控制器
	if len(comp.Controllers()) != 0 {
		t.Error("New component should have no controllers")
	}
}

func TestComponent_Controllers(t *testing.T) {
	comp := fairygui.NewComponent()

	// 创建 Mock Controller
	ctrl := &mockController{name: "TestController"}
	comp.AddController(ctrl)

	if len(comp.Controllers()) != 1 {
		t.Errorf("Expected 1 controller, got %d", len(comp.Controllers()))
	}

	found := comp.GetController("TestController")
	if found != ctrl {
		t.Error("GetController should return the correct controller")
	}
}

// ============================================================================
// Root 测试
// ============================================================================

func TestRoot_Creation(t *testing.T) {
	root := fairygui.NewRoot(800, 600)
	if root == nil {
		t.Fatal("NewRoot() returned nil")
	}

	// Root 应该实现 Root 接口
	var _ fairygui.Root = root

	if root.Width() != 800 {
		t.Errorf("Expected width 800, got %d", root.Width())
	}

	if root.Height() != 600 {
		t.Errorf("Expected height 600, got %d", root.Height())
	}
}

func TestRoot_Resize(t *testing.T) {
	root := fairygui.NewRoot(800, 600)

	root.Resize(1024, 768)

	if root.Width() != 1024 {
		t.Errorf("Expected width 1024, got %d", root.Width())
	}

	if root.Height() != 768 {
		t.Errorf("Expected height 768, got %d", root.Height())
	}

	// 尺寸也应该更新
	width, height := root.Size()
	if width != 1024 || height != 768 {
		t.Errorf("Expected size (1024, 768), got (%f, %f)", width, height)
	}
}

func TestRoot_Update(t *testing.T) {
	root := fairygui.NewRoot(800, 600)

	// Update 应该不报错
	err := root.Update(time.Second / 60)
	if err != nil {
		t.Errorf("Update should not return error, got %v", err)
	}
}

func TestRoot_Renderer(t *testing.T) {
	root := fairygui.NewRoot(800, 600)

	// 默认没有渲染器
	if root.Renderer() != nil {
		t.Error("Root should have no renderer by default")
	}

	// 设置渲染器
	mockRenderer := &mockRenderer{}
	root.SetRenderer(mockRenderer)

	if root.Renderer() != mockRenderer {
		t.Error("Renderer mismatch")
	}
}

func TestRoot_InputManager(t *testing.T) {
	root := fairygui.NewRoot(800, 600)

	// 默认没有输入管理器
	if root.InputManager() != nil {
		t.Error("Root should have no input manager by default")
	}

	// 设置输入管理器
	mockInput := &mockInputManager{}
	root.SetInputManager(mockInput)

	if root.InputManager() != mockInput {
		t.Error("InputManager mismatch")
	}
}

func TestRoot_Draw(t *testing.T) {
	root := fairygui.NewRoot(800, 600)
	screen := ebiten.NewImage(800, 600)

	// 没有渲染器时应该使用默认绘制
	root.Draw(screen)

	// 使用自定义渲染器
	mockRenderer := &mockRenderer{}
	root.SetRenderer(mockRenderer)

	root.Draw(screen)

	if !mockRenderer.drawCalled {
		t.Error("Custom renderer Draw should have been called")
	}
}

// ============================================================================
// Mock 类型
// ============================================================================

type mockController struct {
	name string
}

func (m *mockController) Name() string                        { return m.name }
func (m *mockController) SelectedIndex() int                  { return 0 }
func (m *mockController) SetSelectedIndex(index int)          {}
func (m *mockController) SelectedPage() string                { return "" }
func (m *mockController) SetSelectedPage(name string)         {}
func (m *mockController) PageCount() int                      { return 0 }
func (m *mockController) OnChanged(handler func())            {}

type mockRenderer struct {
	drawCalled bool
}

func (m *mockRenderer) Draw(screen *ebiten.Image, root fairygui.DisplayObject) {
	m.drawCalled = true
}

func (m *mockRenderer) DrawText(screen *ebiten.Image, text string, x, y float64, style fairygui.TextStyle) {
}

func (m *mockRenderer) DrawTexture(screen *ebiten.Image, texture *ebiten.Image, options fairygui.DrawOptions) {
}

func (m *mockRenderer) DrawShape(screen *ebiten.Image, shape fairygui.Shape, options fairygui.DrawOptions) {
}

type mockInputManager struct {
	updateCalled bool
}

func (m *mockInputManager) Update() {
	m.updateCalled = true
}

func (m *mockInputManager) MousePosition() (x, y int) {
	return 0, 0
}

func (m *mockInputManager) IsMouseButtonPressed(button fairygui.MouseButton) bool {
	return false
}

func (m *mockInputManager) IsKeyPressed(key fairygui.Key) bool {
	return false
}

func (m *mockInputManager) TouchIDs() []int {
	return nil
}

func (m *mockInputManager) TouchPosition(id int) (x, y int) {
	return 0, 0
}
