// +build ignore

// 这个文件展示了如何在实际的FairyGUI应用中集成调试服务器
// 使用方式：
// 1. 复制这个文件到你的应用中
// 2. 修改main函数中的根组件引用
// 3. 在应用启动时调用StartDebugServer

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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
	label := widgets.NewLabel()
	label.SetSize(200, 30)
	label.SetPosition(50, 120)
	if textField, ok := label.Data().(*widgets.GTextField); ok {
		textField.SetText("调试服务器测试")
	}
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

// DebugServer 简化的调试服务器
type DebugServer struct {
	root *core.GComponent
}

// NewDebugServer 创建调试服务器
func NewDebugServer(root *core.GComponent) *DebugServer {
	return &DebugServer{root: root}
}

// Start 启动调试服务器
func (ds *DebugServer) Start(port int) error {
	// 简单的路由
	http.HandleFunc("/", ds.handleIndex)
	http.HandleFunc("/tree", ds.handleTreeView)
	http.HandleFunc("/api/tree", ds.handleTreeAPI)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("调试服务器启动在 http://localhost%s", addr)
	log.Printf("访问 /tree 查看渲染树")
	log.Printf("访问 /api/tree 获取JSON数据")

	return http.ListenAndServe(addr, nil)
}

func (ds *DebugServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>FairyGUI Debug Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .link { margin: 10px 0; }
        a { color: #007bff; text-decoration: none; font-size: 16px; }
        a:hover { text-decoration: underline; }
        .header { background: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .info { background: #e8f5e8; padding: 15px; border-radius: 8px; margin: 10px 0; }
        code { background: #f1f1f1; padding: 2px 4px; border-radius: 3px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🛠️ FairyGUI Debug Server</h1>
        <p>实时查看和分析 FairyGUI 渲染树结构</p>
    </div>

    <div class="link">📊 <a href="/tree">查看渲染树</a></div>
    <div class="link">📋 <a href="/api/tree">获取JSON数据</a></div>

    <div class="info">
        <h3>使用说明：</h3>
        <ul>
            <li><strong>渲染树视图</strong>：以树形结构显示所有UI对象</li>
            <li><strong>JSON数据</strong>：获取完整的对象数据用于分析</li>
            <li><strong>实时更新</strong>：数据每5秒自动刷新</li>
        </ul>
    </div>

    <div class="info">
        <h3>集成到应用：</h3>
        <p>在你的应用中添加以下代码：</p>
        <pre><code>// 启动调试服务器（在goroutine中运行）
go func() {
    debugServer := NewDebugServer(yourRootComponent)
    if err := debugServer.Start(8080); err != nil {
        log.Printf("Debug server error: %v", err)
    }
}()</code></pre>
    </div>
</body>
</html>
	`)
}

func (ds *DebugServer) handleTreeView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 获取渲染树数据
	treeData := ds.collectTreeData()

	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>FairyGUI 渲染树</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .tree-container { background: white; padding: 20px; border-radius: 8px; }
        .tree { font-family: 'Courier New', monospace; font-size: 12px; }
        .tree-item { margin: 2px 0; padding: 4px; border-left: 2px solid transparent; }
        .tree-item:hover { background: #f8f9fa; }
        .object-name { color: #333; font-weight: bold; }
        .object-type { color: #666; font-size: 11px; }
        .object-props { color: #888; font-size: 11px; margin-left: 10px; }
        .virtual-info { background: #fff3cd; padding: 4px 8px; border-radius: 4px; margin: 4px 0; font-size: 11px; }
        .children { margin-left: 20px; }
        .expand-btn { cursor: pointer; user-select: none; margin-right: 5px; }
        .stats { background: #e8f5e8; padding: 10px; border-radius: 4px; margin: 10px 0; }
        .refresh-btn { position: fixed; top: 20px; right: 20px; padding: 10px 20px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #0056b3; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌳 FairyGUI 渲染树</h1>
            <button class="refresh-btn" onclick="location.reload()">🔄 刷新</button>
        </div>

        <div class="tree-container">
            <div id="stats" class="stats">%s</div>
            <div id="tree" class="tree">%s</div>
        </div>
    </div>

    <script>
        // 每5秒自动刷新
        setTimeout(function() {
            location.reload();
        }, 5000);
    </script>
</body>
</html>
	`, treeData.Stats, treeData.HTML)
}

func (ds *DebugServer) handleTreeAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data := ds.collectJSONData()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}

// TreeData 树形数据
type TreeData struct {
	Stats string
	HTML  string
}

// collectTreeData 收集树形数据用于HTML显示
func (ds *DebugServer) collectTreeData() TreeData {
	if ds.root == nil {
		return TreeData{
			Stats: "错误：根组件为空",
			HTML:  "<div style='color:red;'>根组件未设置</div>",
		}
	}

	var stats struct {
		total   int
		visible int
		virtual int
	}

	var html strings.Builder
	html.WriteString("<div class='tree'>")

	// 遍历根对象的子对象
	if rootObj := ds.root.GObject; rootObj != nil {
		if comp, ok := rootObj.Data().(*core.GComponent); ok {
			children := comp.Children()
			for _, child := range children {
				if child != nil {
					html.WriteString(ds.renderTreeItem(child, 0, &stats))
				}
			}
		}
	}

	html.WriteString("</div>")

	return TreeData{
		Stats: fmt.Sprintf("📊 总计: %d 对象, %d 可见, %d 虚拟列表", stats.total, stats.visible, stats.virtual),
		HTML:  html.String(),
	}
}

// renderTreeItem 渲染树形项目
func (ds *DebugServer) renderTreeItem(obj *core.GObject, depth int, stats *struct {
	total   int
	visible int
	virtual int
}) string {
	if obj == nil {
		return ""
	}

	stats.total++
	if obj.Visible() {
		stats.visible++
	}

	var result strings.Builder

	// 基本信息
	objType := ds.getObjectType(obj)
	result.WriteString(fmt.Sprintf("<div class='tree-item' style='margin-left:%dpx;'>", depth*20))
	result.WriteString(fmt.Sprintf("<span class='expand-btn'>•</span>"))
	result.WriteString(fmt.Sprintf("<span class='object-name'>%s</span> ", obj.Name()))
	result.WriteString(fmt.Sprintf("<span class='object-type'>(%s)</span>", objType))
	result.WriteString(fmt.Sprintf("<span class='object-props'>pos:%.0f,%.0f size:%.0fx%.0f",
		obj.X(), obj.Y(), obj.Width(), obj.Height()))

	if obj.Rotation() != 0 {
		result.WriteString(fmt.Sprintf(" rot:%.1f", obj.Rotation()))
	}
	if obj.Alpha() != 1 {
		result.WriteString(fmt.Sprintf(" alpha:%.2f", obj.Alpha()))
	}
	if !obj.Visible() {
		result.WriteString(" [隐藏]")
	}
	result.WriteString("</span>")

	// 虚拟列表特殊信息
	if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
		stats.virtual++
		info := ds.getVirtualListInfo(list)
		if info != nil {
			result.WriteString(fmt.Sprintf("<div class='virtual-info'>🔄 虚拟列表: %d项, 视图:%dx%d, 子对象:%d</div>",
				info.NumItems, info.ViewWidth, info.ViewHeight, info.ChildrenCount))
		}
	}

	result.WriteString("</div>")

	// 递归处理子对象
	if comp, ok := obj.Data().(*core.GComponent); ok {
		children := comp.Children()
		for _, child := range children {
			if child != nil {
				result.WriteString(ds.renderTreeItem(child, depth+1, stats))
			}
		}
	}

	return result.String()
}

// collectJSONData 收集JSON数据
func (ds *DebugServer) collectJSONData() map[string]interface{} {
	if ds.root == nil {
		return map[string]interface{}{
			"error": "根组件为空",
		}
	}

	var objects []map[string]interface{}

	// 遍历根对象的子对象
	if rootObj := ds.root.GObject; rootObj != nil {
		if comp, ok := rootObj.Data().(*core.GComponent); ok {
			children := comp.Children()
			for _, child := range children {
				if child != nil {
					objects = append(objects, ds.collectObjectJSON(child))
				}
			}
		}
	}

	return map[string]interface{}{
		"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		"root":      "GRoot",
		"objects":   objects,
		"count":     len(objects),
	}
}

// collectObjectJSON 收集对象JSON数据
func (ds *DebugServer) collectObjectJSON(obj *core.GObject) map[string]interface{} {
	if obj == nil {
		return nil
	}

	data := map[string]interface{}{
		"id":       fmt.Sprintf("%p", obj),
		"name":     obj.Name(),
		"type":     ds.getObjectType(obj),
		"position": map[string]float64{"x": obj.X(), "y": obj.Y()},
		"size":     map[string]float64{"width": obj.Width(), "height": obj.Height()},
		"rotation": obj.Rotation(),
		"alpha":    obj.Alpha(),
		"visible":  obj.Visible(),
	}

	// 虚拟列表特殊处理
	if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
		info := ds.getVirtualListInfo(list)
		if info != nil {
			data["virtualInfo"] = info
		}
	}

	// 子对象
	if comp, ok := obj.Data().(*core.GComponent); ok {
		children := comp.Children()
		var childData []map[string]interface{}
		for _, child := range children {
			if child != nil {
				childData = append(childData, ds.collectObjectJSON(child))
			}
		}
		if len(childData) > 0 {
			data["children"] = childData
		}
	}

	return data
}

// getObjectType 获取对象类型
func (ds *DebugServer) getObjectType(obj *core.GObject) string {
	if obj == nil {
		return "Unknown"
	}

	switch data := obj.Data().(type) {
	case *core.GComponent:
		return "GComponent"
	case *widgets.GList:
		return "GList"
	case *widgets.GButton:
		return "GButton"
	case *widgets.GTextField:
		return "GTextField"
	case *widgets.GImage:
		return "GImage"
	case *widgets.GGraph:
		return "GGraph"
	case *widgets.GLoader:
		return "GLoader"
	case *widgets.GComboBox:
		return "GComboBox"
	case *widgets.GTree:
		return "GTree"
	case *widgets.GMovieClip:
		return "GMovieClip"
	case *widgets.GProgressBar:
		return "GProgressBar"
	case *widgets.GSlider:
		return "GSlider"
	case *widgets.GScrollBar:
		return "GScrollBar"
	case *widgets.GRichTextField:
		return "GRichTextField"
	case *widgets.GTextInput:
		return "GTextInput"
	case *widgets.GGroup:
		return "GGroup"
	default:
		return fmt.Sprintf("%T", data)
	}
}

// getVirtualListInfo 获取虚拟列表信息
func (ds *DebugServer) getVirtualListInfo(list *widgets.GList) *VirtualListInfo {
	if list == nil || !list.IsVirtual() {
		return nil
	}

	info := &VirtualListInfo{
		IsVirtual:     list.IsVirtual(),
		NumItems:      list.NumItems(),
		ChildrenCount: len(list.GComponent.Children()),
	}

	// 获取视图尺寸
	if scrollPane := list.GComponent.ScrollPane(); scrollPane != nil {
		info.ViewWidth = int(scrollPane.ViewWidth())
		info.ViewHeight = int(scrollPane.ViewHeight())
	} else {
		info.ViewWidth = int(list.Width())
		info.ViewHeight = int(list.Height())
	}

	// 获取项目尺寸
	if itemSize := list.VirtualItemSize(); itemSize != nil {
		info.ItemSize = fmt.Sprintf("%.0fx%.0f", itemSize.X, itemSize.Y)
	}

	return info
}

// VirtualListInfo 虚拟列表信息
type VirtualListInfo struct {
	IsVirtual     bool   `json:"is_virtual"`
	NumItems      int    `json:"num_items"`
	ChildrenCount int    `json:"children_count"`
	ViewWidth     int    `json:"view_width"`
	ViewHeight    int    `json:"view_height"`
	ItemSize      string `json:"item_size"`
}

// 主函数 - 运行演示
func main() {
	fmt.Println("=== FairyGUI 调试服务器集成演示 ===")
	IntegrationWithDemo()
}