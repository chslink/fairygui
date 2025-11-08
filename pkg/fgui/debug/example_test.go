package debug_test

import (
	"fmt"
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/debug"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ExampleInspector 演示如何使用Inspector
func ExampleInspector() {
	// 创建测试场景
	root := core.NewGObject()
	root.SetName("Root")

	comp := core.NewGComponent()
	comp.GObject.SetName("Scene")
	root.AddChild(comp.GObject)

	button := widgets.NewButton()
	button.GComponent.GObject.SetName("SubmitButton")
	comp.AddChild(button.GComponent.GObject)

	// 创建Inspector
	inspector := debug.NewInspector(root)

	// 查找对象
	buttons := inspector.FindByType("GButton")
	fmt.Printf("找到 %d 个按钮\n", len(buttons))

	// 按名称查找
	objs := inspector.FindByName("Submit")
	if len(objs) > 0 {
		fmt.Printf("找到对象: %s\n", objs[0].Name())
	}

	// 获取对象信息
	info := inspector.GetInfo(button.GComponent.GObject)
	fmt.Printf("对象类型: %s\n", info.Type)

	// 统计对象
	stats := inspector.CountObjects()
	fmt.Printf("总对象数: %d\n", stats["total"])

	// Output:
	// 找到 1 个按钮
	// 找到对象: SubmitButton
	// 对象类型: GButton
	// 总对象数: 3
}

// TestInspectorFindByPath 测试路径查找
func TestInspectorFindByPath(t *testing.T) {
	// 创建测试场景
	root := core.NewGObject()
	root.SetName("Root")

	scene := core.NewGComponent()
	scene.GObject.SetName("Scene")
	root.AddChild(scene.GObject)

	panel := core.NewGComponent()
	panel.GObject.SetName("Panel")
	scene.AddChild(panel.GObject)

	button := widgets.NewButton()
	button.GComponent.GObject.SetName("Button")
	panel.AddChild(button.GComponent.GObject)

	// 创建Inspector
	inspector := debug.NewInspector(root)

	// 测试路径查找
	tests := []struct {
		path string
		want string
	}{
		{"/Root/Scene", "Scene"},
		{"/Root/Scene/Panel", "Panel"},
		{"/Root/Scene/Panel/Button", "Button"},
		{"/Root/Invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			obj := inspector.FindByPath(tt.path)
			if tt.want == "" {
				if obj != nil {
					t.Errorf("期望未找到对象，实际找到: %s", obj.Name())
				}
			} else {
				if obj == nil {
					t.Errorf("未找到对象: %s", tt.path)
					return
				}
				if obj.Name() != tt.want {
					t.Errorf("期望对象名: %s, 实际: %s", tt.want, obj.Name())
				}
			}
		})
	}
}

// TestInspectorFilter 测试复杂筛选
func TestInspectorFilter(t *testing.T) {
	// 创建测试场景
	root := core.NewGObject()
	root.SetName("Root")

	comp := core.NewGComponent()
	comp.GObject.SetName("Scene")
	root.AddChild(comp.GObject)

	// 添加多个按钮
	for i := 1; i <= 3; i++ {
		btn := widgets.NewButton()
		btn.GComponent.GObject.SetName(fmt.Sprintf("Button%d", i))
		if i == 2 {
			btn.GComponent.GObject.SetVisible(false)
		}
		comp.AddChild(btn.GComponent.GObject)
	}

	inspector := debug.NewInspector(root)

	// 测试可见性筛选
	visible := true
	filter := debug.Filter{
		Type:    "GButton",
		Visible: &visible,
	}

	results := inspector.FindByFilter(filter)
	if len(results) != 2 {
		t.Errorf("期望找到2个可见按钮，实际: %d", len(results))
	}

	// 测试名称筛选
	filter2 := debug.Filter{
		Name: "Button2",
	}
	results2 := inspector.FindByFilter(filter2)
	if len(results2) != 1 {
		t.Errorf("期望找到1个对象，实际: %d", len(results2))
	}
}

// TestEventSimulator 测试事件模拟
func TestEventSimulator(t *testing.T) {
	// 创建测试环境
	env := testutil.NewStageEnv(t, 800, 600)

	// 创建测试对象
	button := widgets.NewButton()
	button.GComponent.GObject.SetName("TestButton")
	button.GComponent.GObject.SetSize(100, 50)
	button.GComponent.GObject.SetPosition(100, 100)

	env.Stage.Root().AddChild(button.GComponent.GObject.DisplayObject())

	// 创建Inspector和Simulator
	inspector := debug.NewInspector(env.Stage.Root().ToGObject())
	simulator := debug.NewEventSimulator(env.Stage)

	// 记录点击事件
	clicked := false
	button.GComponent.GObject.On(laya.EventClick, func(evt *laya.Event) {
		clicked = true
	})

	// 模拟点击
	err := simulator.ClickByName(inspector, "TestButton")
	if err != nil {
		t.Fatalf("点击失败: %v", err)
	}

	if !clicked {
		t.Error("按钮未被点击")
	}
}

// TestInspectorGetInfo 测试对象信息获取
func TestInspectorGetInfo(t *testing.T) {
	// 创建测试对象
	button := widgets.NewButton()
	button.GComponent.GObject.SetName("TestButton")
	button.GComponent.GObject.SetSize(120, 40)
	button.GComponent.GObject.SetPosition(10, 20)
	button.GComponent.GObject.SetVisible(true)
	button.SetTitle("Click Me")

	root := core.NewGObject()
	root.SetName("Root")
	root.AddChild(button.GComponent.GObject)

	// 创建Inspector
	inspector := debug.NewInspector(root)

	// 获取对象信息
	info := inspector.GetInfo(button.GComponent.GObject)

	// 验证基本信息
	if info.Name != "TestButton" {
		t.Errorf("期望名称: TestButton, 实际: %s", info.Name)
	}
	if info.Type != "GButton" {
		t.Errorf("期望类型: GButton, 实际: %s", info.Type)
	}
	if info.Size.Width != 120 {
		t.Errorf("期望宽度: 120, 实际: %.0f", info.Size.Width)
	}
	if info.Size.Height != 40 {
		t.Errorf("期望高度: 40, 实际: %.0f", info.Size.Height)
	}
	if !info.Visible {
		t.Error("期望对象可见")
	}

	// 验证属性
	if title, ok := info.Properties["title"]; !ok || title != "Click Me" {
		t.Errorf("期望标题: Click Me, 实际: %v", title)
	}
}

// TestInspectorCountObjects 测试对象统计
func TestInspectorCountObjects(t *testing.T) {
	// 创建复杂场景
	root := core.NewGObject()
	root.SetName("Root")

	comp := core.NewGComponent()
	comp.GObject.SetName("Scene")
	root.AddChild(comp.GObject)

	// 添加5个按钮（3个可见，2个隐藏）
	for i := 1; i <= 5; i++ {
		btn := widgets.NewButton()
		btn.GComponent.GObject.SetName(fmt.Sprintf("Button%d", i))
		btn.GComponent.GObject.SetVisible(i <= 3)
		comp.AddChild(btn.GComponent.GObject)
	}

	// 添加2个列表
	for i := 1; i <= 2; i++ {
		list := widgets.NewList()
		list.GComponent.GObject.SetName(fmt.Sprintf("List%d", i))
		comp.AddChild(list.GComponent.GObject)
	}

	inspector := debug.NewInspector(root)
	stats := inspector.CountObjects()

	// 验证统计结果
	// Root + Scene + 5 Buttons + 2 Lists = 9
	if stats["total"] != 9 {
		t.Errorf("期望总对象数: 9, 实际: %d", stats["total"])
	}

	// Root + Scene + 3 可见Buttons + 2 Lists = 7
	if stats["visible"] != 7 {
		t.Errorf("期望可见对象数: 7, 实际: %d", stats["visible"])
	}

	// Scene + 5 Button GComponents + 2 List GComponents = 8
	// （每个Widget都有一个GComponent）
	if stats["containers"] < 1 {
		t.Errorf("期望至少1个容器, 实际: %d", stats["containers"])
	}
}

// BenchmarkInspectorFindByName 基准测试
func BenchmarkInspectorFindByName(b *testing.B) {
	// 创建大型场景
	root := core.NewGObject()
	root.SetName("Root")

	comp := core.NewGComponent()
	comp.GObject.SetName("Scene")
	root.AddChild(comp.GObject)

	// 添加100个对象
	for i := 0; i < 100; i++ {
		btn := widgets.NewButton()
		btn.GComponent.GObject.SetName(fmt.Sprintf("Button%d", i))
		comp.AddChild(btn.GComponent.GObject)
	}

	inspector := debug.NewInspector(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inspector.FindByName("Button50")
	}
}
