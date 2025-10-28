// +build ignore

// 这个文件展示了如何在实际的FairyGUI应用中集成调试服务器
// 使用方式：
// 1. 复制这个文件到你的应用中
// 2. 修改main函数中的根组件引用
// 3. 在应用启动时调用StartDebugServer

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// DebugIntegration 调试集成工具
type DebugIntegration struct {
	root     *core.GComponent
	server   *DebugServer
	port     int
	enabled  bool
}

// NewDebugIntegration 创建调试集成
func NewDebugIntegration(root *core.GComponent, port int) *DebugIntegration {
	return &DebugIntegration{
		root:    root,
		port:    port,
		enabled: false,
	}
}

// Start 启动调试服务器
func (di *DebugIntegration) Start() error {
	if di.enabled {
		return fmt.Errorf("调试服务器已在运行")
	}

	di.server = NewDebugServer(di.root)

	// 在goroutine中启动服务器，避免阻塞主应用
	go func() {
		log.Printf("🛠️ 启动调试服务器在端口 %d", di.port)
		log.Printf("📊 访问 http://localhost:%d 查看调试界面", di.port)

		if err := di.server.Start(di.port); err != nil {
			log.Printf("❌ 调试服务器启动失败: %v", err)
		}
	}()

	di.enabled = true
	return nil
}

// Stop 停止调试服务器
func (di *DebugIntegration) Stop() {
	if !di.enabled {
		return
	}

	// 注意：http.Server没有简单的停止方法，这里只是标记状态
	di.enabled = false
	log.Println("调试服务器已停止")
}

// IsEnabled 返回调试服务器是否启用
func (di *DebugIntegration) IsEnabled() bool {
	return di.enabled
}

// GetURL 返回调试服务器的访问URL
func (di *DebugIntegration) GetURL() string {
	return fmt.Sprintf("http://localhost:%d", di.port)
}

// ExampleUsage 使用示例
func ExampleUsage() {
	// 假设这是你的应用中的某个地方
	var rootComponent *core.GComponent // 你的根组件

	// 创建调试集成
	debug := NewDebugIntegration(rootComponent, 8080)

	// 启动调试服务器
	if err := debug.Start(); err != nil {
		log.Printf("调试服务器启动失败: %v", err)
	} else {
		log.Printf("调试服务器已启动，访问: %s", debug.GetURL())
	}

	// 在你的应用的其他地方，你可以检查调试状态
	if debug.IsEnabled() {
		log.Println("调试功能已启用")
	}
}

// IntegrationWithDemo 与演示集成的完整示例
func IntegrationWithDemo() {
	fmt.Println("=== FairyGUI 调试服务器集成演示 ===")

	// 1. 创建根组件（通常在应用初始化时）
	root := core.NewGComponent()
	root.SetSize(800, 600)

	// 2. 创建一些测试对象
	createTestUI(root)

	// 3. 创建并启动调试服务器
	debug := NewDebugIntegration(root, 8080)

	fmt.Println("启动调试服务器...")
	if err := debug.Start(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}

	fmt.Printf("✅ 调试服务器已启动！\n")
	fmt.Printf("🌐 访问地址: %s\n", debug.GetURL())
	fmt.Println("📋 可以查看：")
	fmt.Println("   • 所有UI对象的层次结构")
	fmt.Println("   • 每个对象的位置、尺寸、可见性等属性")
	fmt.Println("   • 虚拟列表的详细信息")
	fmt.Println("   • 实时更新的对象状态")
	fmt.Println("")
	fmt.Println("按 Ctrl+C 停止服务器...")

	// 4. 保持程序运行（在实际应用中，这会是你的主循环）
	select {}
}

// createTestUI 创建测试UI
func createTestUI(root *core.GComponent) {
	// 创建一些测试对象来演示调试功能

	// 普通按钮
	button1 := widgets.NewButton()
	button1.SetSize(100, 40)
	button1.SetPosition(50, 50)
	button1.SetTitle("测试按钮1")
	root.AddChild(button1.GObject)

	button2 := widgets.NewButton()
	button2.SetSize(100, 40)
	button2.SetPosition(200, 50)
	button2.SetTitle("测试按钮2")
	button2.SetVisible(false) // 隐藏按钮，用于测试可见性筛选
	root.AddChild(button2.GObject)

	// 文本标签
	label := widgets.NewTextField()
	label.SetSize(200, 30)
	label.SetPosition(50, 120)
	label.SetText("调试服务器测试")
	root.AddChild(label.GObject)

	// 虚拟列表（重点测试对象）
	vlist := widgets.NewList()
	vlist.SetSize(200, 200)
	vlist.SetPosition(350, 50)
	vlist.SetVirtual(true)
	vlist.SetNumItems(50)
	vlist.SetDefaultItem("ui://test/item")

	// 设置对象创建器
	vlist.SetObjectCreator(&TestObjectCreator{})

	// 设置项目渲染器
	vlist.SetItemRenderer(func(index int, item *core.GObject) {
		if item != nil {
			item.SetName(fmt.Sprintf("项目_%d", index))
		}
	})

	root.AddChild(vlist.GObject)

	// 刷新虚拟列表
	vlist.RefreshVirtualList()

	fmt.Printf("创建了测试UI：%d 个对象\n", len(root.Children()))
}

// TestObjectCreator 测试对象创建器
type TestObjectCreator struct{}

func (t *TestObjectCreator) CreateObject(url string) *core.GObject {
	obj := core.NewGObject()
	obj.SetName(url)
	obj.SetSize(100, 30)
	return obj
}

// 主函数 - 运行演示
func main() {
	IntegrationWithDemo()
}