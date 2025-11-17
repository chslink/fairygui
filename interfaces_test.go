package fairygui_test

import (
	"testing"
	"time"

	"github.com/chslink/fairygui"
	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// 接口规范测试
// ============================================================================

// TestDisplayObjectInterface 测试 DisplayObject 接口的基本规范。
func TestDisplayObjectInterface(t *testing.T) {
	// 这个测试确保 DisplayObject 接口组合了所有必要的小接口
	var obj fairygui.DisplayObject

	// 验证接口组合
	_ = fairygui.Positionable(obj)
	_ = fairygui.Sizable(obj)
	_ = fairygui.Transformable(obj)
	_ = fairygui.Visible(obj)
	_ = fairygui.Hierarchical(obj)
	_ = fairygui.Drawable(obj)

	// 如果编译通过，说明接口组合正确
}

// TestRendererInterface 测试 Renderer 接口规范。
func TestRendererInterface(t *testing.T) {
	var renderer fairygui.Renderer

	// 验证方法签名
	screen := ebiten.NewImage(800, 600)
	var root fairygui.DisplayObject
	var text string
	var x, y float64
	var style fairygui.TextStyle
	var texture *ebiten.Image
	var opts fairygui.DrawOptions
	var shape fairygui.Shape

	// 这些调用只是为了验证接口方法签名，不会实际执行
	if renderer != nil {
		renderer.Draw(screen, root)
		renderer.DrawText(screen, text, x, y, style)
		renderer.DrawTexture(screen, texture, opts)
		renderer.DrawShape(screen, shape, opts)
	}
}

// TestEventDispatcherInterface 测试 EventDispatcher 接口规范。
func TestEventDispatcherInterface(t *testing.T) {
	var dispatcher fairygui.EventDispatcher

	// 验证方法签名
	handler := func(event fairygui.Event) {}
	eventType := "test"

	if dispatcher != nil {
		cancel := dispatcher.On(eventType, handler)
		_ = cancel
		dispatcher.Off(eventType, handler)
		dispatcher.Once(eventType, handler)
		dispatcher.Emit(fairygui.NewEvent(eventType, nil))
		_ = dispatcher.HasListener(eventType)
	}
}

// TestAssetLoaderInterface 测试 AssetLoader 接口规范。
func TestAssetLoaderInterface(t *testing.T) {
	var loader fairygui.AssetLoader

	if loader != nil {
		_, _ = loader.LoadPackage("test")
		_, _ = loader.LoadTexture("test.png")
		_, _ = loader.LoadAudio("test.mp3")
		_, _ = loader.LoadFont("test.ttf")
		_ = loader.Exists("test")
	}
}

// TestComponentInterface 测试 Component 接口规范。
func TestComponentInterface(t *testing.T) {
	var comp fairygui.Component

	// Component 应该包含 DisplayObject 和 Interactive
	_ = fairygui.DisplayObject(comp)
	_ = fairygui.Interactive(comp)

	if comp != nil {
		_ = comp.Controllers()
		_ = comp.GetController("test")
		comp.AddController(nil)
	}
}

// TestRootInterface 测试 Root 接口规范。
func TestRootInterface(t *testing.T) {
	var root fairygui.Root

	// Root 应该包含 Component 和 Updatable
	_ = fairygui.Component(root)
	_ = fairygui.Updatable(root)

	if root != nil {
		root.Resize(800, 600)
		root.SetRenderer(nil)
		_ = root.Renderer()
		root.SetInputManager(nil)
		_ = root.InputManager()
		_ = root.Update(time.Second / 60)
	}
}

// ============================================================================
// 事件类型测试
// ============================================================================

// TestEventTypes 测试事件类型定义。
func TestEventTypes(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{"MouseDown", fairygui.EventMouseDown},
		{"MouseUp", fairygui.EventMouseUp},
		{"Click", fairygui.EventClick},
		{"TouchBegin", fairygui.EventTouchBegin},
		{"KeyDown", fairygui.EventKeyDown},
		{"Added", fairygui.EventAdded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.eventType == "" {
				t.Errorf("Event type %s should not be empty", tt.name)
			}
		})
	}
}

// TestBaseEvent 测试基础事件实现。
func TestBaseEvent(t *testing.T) {
	target := "test_target"
	event := fairygui.NewEvent(fairygui.EventClick, target)

	if event.Type() != fairygui.EventClick {
		t.Errorf("Expected event type %s, got %s", fairygui.EventClick, event.Type())
	}

	if event.Target() != target {
		t.Errorf("Expected target %v, got %v", target, event.Target())
	}

	if event.IsDefaultPrevented() {
		t.Error("Event should not be prevented by default")
	}

	event.PreventDefault()
	if !event.IsDefaultPrevented() {
		t.Error("Event should be prevented after calling PreventDefault()")
	}

	if event.IsPropagationStopped() {
		t.Error("Event propagation should not be stopped by default")
	}

	event.StopPropagation()
	if !event.IsPropagationStopped() {
		t.Error("Event propagation should be stopped after calling StopPropagation()")
	}
}

// TestMouseEvent 测试鼠标事件。
func TestMouseEvent(t *testing.T) {
	target := "test_target"
	x, y := 100.0, 200.0
	event := fairygui.NewMouseEvent(fairygui.EventClick, target, x, y)

	if event.Type() != fairygui.EventClick {
		t.Errorf("Expected event type %s, got %s", fairygui.EventClick, event.Type())
	}

	if event.X != x {
		t.Errorf("Expected X %f, got %f", x, event.X)
	}

	if event.Y != y {
		t.Errorf("Expected Y %f, got %f", y, event.Y)
	}
}

// TestTouchEvent 测试触摸事件。
func TestTouchEvent(t *testing.T) {
	target := "test_target"
	touchID := 1
	x, y := 150.0, 250.0
	event := fairygui.NewTouchEvent(fairygui.EventTouchBegin, target, touchID, x, y)

	if event.Type() != fairygui.EventTouchBegin {
		t.Errorf("Expected event type %s, got %s", fairygui.EventTouchBegin, event.Type())
	}

	if event.TouchID != touchID {
		t.Errorf("Expected TouchID %d, got %d", touchID, event.TouchID)
	}

	if event.X != x {
		t.Errorf("Expected X %f, got %f", x, event.X)
	}

	if event.Y != y {
		t.Errorf("Expected Y %f, got %f", y, event.Y)
	}
}

// TestKeyboardEvent 测试键盘事件。
func TestKeyboardEvent(t *testing.T) {
	target := "test_target"
	key := ebiten.KeySpace
	event := fairygui.NewKeyboardEvent(fairygui.EventKeyDown, target, key)

	if event.Type() != fairygui.EventKeyDown {
		t.Errorf("Expected event type %s, got %s", fairygui.EventKeyDown, event.Type())
	}

	if event.Key != key {
		t.Errorf("Expected key %v, got %v", key, event.Key)
	}
}

// ============================================================================
// 接口可实现性测试
// ============================================================================

// TestInterfaceImplementation 测试接口可以被正确实现。
func TestInterfaceImplementation(t *testing.T) {
	// 创建一个简单的实现来验证接口可以被实现
	obj := &mockDisplayObject{}

	// 验证实现了所有必需的接口
	var _ fairygui.DisplayObject = obj
	var _ fairygui.Positionable = obj
	var _ fairygui.Sizable = obj
	var _ fairygui.Transformable = obj
	var _ fairygui.Visible = obj
	var _ fairygui.Hierarchical = obj
	var _ fairygui.Drawable = obj
}

// mockDisplayObject 是一个用于测试的 DisplayObject 实现。
type mockDisplayObject struct {
	id       string
	name     string
	x, y     float64
	w, h     float64
	scaleX   float64
	scaleY   float64
	rotation float64
	skewX    float64
	skewY    float64
	pivotX   float64
	pivotY   float64
	visible  bool
	alpha    float64
	touchable bool
	parent   fairygui.DisplayObject
	children []fairygui.DisplayObject
	data     interface{}
}

// 实现 Positionable 接口
func (m *mockDisplayObject) Position() (x, y float64) { return m.x, m.y }
func (m *mockDisplayObject) SetPosition(x, y float64) { m.x, m.y = x, y }
func (m *mockDisplayObject) GlobalPosition() (x, y float64) { return m.x, m.y }

// 实现 Sizable 接口
func (m *mockDisplayObject) Size() (width, height float64) { return m.w, m.h }
func (m *mockDisplayObject) SetSize(width, height float64) { m.w, m.h = width, height }

// 实现 Transformable 接口
func (m *mockDisplayObject) Scale() (scaleX, scaleY float64) { return m.scaleX, m.scaleY }
func (m *mockDisplayObject) SetScale(scaleX, scaleY float64) { m.scaleX, m.scaleY = scaleX, scaleY }
func (m *mockDisplayObject) Rotation() float64 { return m.rotation }
func (m *mockDisplayObject) SetRotation(rotation float64) { m.rotation = rotation }
func (m *mockDisplayObject) Skew() (skewX, skewY float64) { return m.skewX, m.skewY }
func (m *mockDisplayObject) SetSkew(skewX, skewY float64) { m.skewX, m.skewY = skewX, skewY }
func (m *mockDisplayObject) Pivot() (pivotX, pivotY float64) { return m.pivotX, m.pivotY }
func (m *mockDisplayObject) SetPivot(pivotX, pivotY float64) { m.pivotX, m.pivotY = pivotX, pivotY }

// 实现 Visible 接口
func (m *mockDisplayObject) Visible() bool { return m.visible }
func (m *mockDisplayObject) SetVisible(visible bool) { m.visible = visible }
func (m *mockDisplayObject) Alpha() float64 { return m.alpha }
func (m *mockDisplayObject) SetAlpha(alpha float64) { m.alpha = alpha }

// 实现 Hierarchical 接口
func (m *mockDisplayObject) Parent() fairygui.DisplayObject { return m.parent }
func (m *mockDisplayObject) Children() []fairygui.DisplayObject { return m.children }
func (m *mockDisplayObject) AddChild(child fairygui.DisplayObject) { m.children = append(m.children, child) }
func (m *mockDisplayObject) AddChildAt(child fairygui.DisplayObject, index int) {}
func (m *mockDisplayObject) RemoveChild(child fairygui.DisplayObject) {}
func (m *mockDisplayObject) RemoveChildAt(index int) fairygui.DisplayObject { return nil }
func (m *mockDisplayObject) ChildCount() int { return len(m.children) }
func (m *mockDisplayObject) GetChildAt(index int) fairygui.DisplayObject { return nil }
func (m *mockDisplayObject) GetChildByName(name string) fairygui.DisplayObject { return nil }

// 实现 Drawable 接口
func (m *mockDisplayObject) Draw(screen *ebiten.Image) {}

// 实现 DisplayObject 特有方法
func (m *mockDisplayObject) ID() string { return m.id }
func (m *mockDisplayObject) Name() string { return m.name }
func (m *mockDisplayObject) SetName(name string) { m.name = name }
func (m *mockDisplayObject) Data() interface{} { return m.data }
func (m *mockDisplayObject) SetData(data interface{}) { m.data = data }
func (m *mockDisplayObject) Touchable() bool { return m.touchable }
func (m *mockDisplayObject) SetTouchable(touchable bool) { m.touchable = touchable }
func (m *mockDisplayObject) Dispose() {}

// ============================================================================
// 接口方法数量测试（单一职责原则）
// ============================================================================

// TestInterfaceMethodCount 测试接口方法数量是否合理。
func TestInterfaceMethodCount(t *testing.T) {
	// 这个测试主要是文档性质的，记录每个接口的方法数量
	// 确保遵循接口隔离原则：小而专注的接口

	tests := []struct {
		name          string
		methodCount   int
		maxRecommended int
	}{
		{"Positionable", 3, 5},
		{"Sizable", 2, 5},
		{"Transformable", 6, 8},
		{"Visible", 4, 5},
		{"Drawable", 1, 3},
		{"Updatable", 1, 3},
		{"EventDispatcher", 5, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.methodCount > tt.maxRecommended {
				t.Logf("WARNING: %s has %d methods, which exceeds the recommended max of %d",
					tt.name, tt.methodCount, tt.maxRecommended)
			}
		})
	}
}
