package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// DebugServer HTTP调试服务器
type DebugServer struct {
	root *core.GComponent
}

// ObjectInfo 对象信息
type ObjectInfo struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Position    string      `json:"position"`
	Size        string      `json:"size"`
	Scale       string      `json:"scale"`
	Rotation    float64     `json:"rotation"`
	Alpha       float64     `json:"alpha"`
	Visible     bool        `json:"visible"`
	Children    []ObjectInfo `json:"children,omitempty"`
	Data        interface{} `json:"data,omitempty"`
	VirtualInfo *VirtualListInfo `json:"virtual_info,omitempty"`
}

// VirtualListInfo 虚拟列表信息
type VirtualListInfo struct {
	IsVirtual      bool   `json:"is_virtual"`
	NumItems       int    `json:"num_items"`
	RealNumItems   int    `json:"real_num_items"`
	FirstIndex     int    `json:"first_index"`
	ViewWidth      int    `json:"view_width"`
	ViewHeight     int    `json:"view_height"`
	ItemSize       string `json:"item_size"`
	ChildrenCount  int    `json:"children_count"`
}

// NewDebugServer 创建调试服务器
func NewDebugServer(root *core.GComponent) *DebugServer {
	return &DebugServer{root: root}
}

// Start 启动调试服务器
func (ds *DebugServer) Start(port int) error {
	http.HandleFunc("/", ds.handleIndex)
	http.HandleFunc("/api/tree", ds.handleTreeAPI)
	http.HandleFunc("/api/object/", ds.handleObjectAPI)
	http.HandleFunc("/tree", ds.handleTreeView)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("调试服务器启动在 http://localhost%s\n", addr)
	fmt.Printf("访问 /tree 查看渲染树\n")
	fmt.Printf("访问 /api/tree 获取JSON数据\n")

	return http.ListenAndServe(addr, nil)
}

// handleIndex 处理首页
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
        a { color: #007bff; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>FairyGUI Debug Server</h1>
    <div class="link"><a href="/tree">🌳 查看渲染树</a></div>
    <div class="link"><a href="/api/tree">📊 获取JSON数据</a></div>
    <div class="link"><a href="/api/tree?type=GList">📋 只查看列表对象</a></div>
    <div class="link"><a href="/api/tree?visible=true">👁️ 只查看可见对象</a></div>
</body>
</html>
	`)
}

// handleTreeAPI 处理树形数据API
func (ds *DebugServer) handleTreeAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 解析查询参数
	filterType := r.URL.Query().Get("type")
	visibleOnly := r.URL.Query().Get("visible") == "true"
	nameFilter := r.URL.Query().Get("name")

	// 收集对象信息
	var tree []ObjectInfo
	if ds.root != nil {
		tree = ds.collectObjectInfo(ds.root.GObject(), filterType, visibleOnly, nameFilter, 0)
	}

	// 返回JSON数据
	response := map[string]interface{}{
		"root":     "GRoot",
		"children": tree,
		"filter": map[string]string{
			"type":    filterType,
			"visible": strconv.FormatBool(visibleOnly),
			"name":    nameFilter,
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response)
}

// handleObjectAPI 处理单个对象API
func (ds *DebugServer) handleObjectAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 从URL路径中提取对象ID
	path := strings.TrimPrefix(r.URL.Path, "/api/object/")
	if path == "" {
		http.Error(w, "需要对象ID", http.StatusBadRequest)
		return
	}

	// 查找对象
	obj := ds.findObjectByID(ds.root.GObject(), path)
	if obj == nil {
		http.Error(w, "对象未找到", http.StatusNotFound)
		return
	}

	// 返回对象详细信息
	info := ds.getObjectDetail(obj)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(info)
}

// handleTreeView 处理树形视图页面
func (ds *DebugServer) handleTreeView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>FairyGUI 渲染树</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .controls { margin: 10px 0; }
        .controls input, .controls select { margin: 5px; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        .controls button { margin: 5px; padding: 8px 16px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
        .controls button:hover { background: #0056b3; }
        .tree-container { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .tree { font-family: 'Courier New', monospace; font-size: 12px; }
        .tree-item { margin: 2px 0; padding: 4px; border-left: 2px solid transparent; }
        .tree-item:hover { background: #f8f9fa; }
        .tree-item.selected { background: #e3f2fd; border-left-color: #2196f3; }
        .object-name { color: #333; font-weight: bold; }
        .object-type { color: #666; font-size: 11px; }
        .object-props { color: #888; font-size: 11px; margin-left: 10px; }
        .virtual-info { background: #fff3cd; padding: 4px 8px; border-radius: 4px; margin: 4px 0; font-size: 11px; }
        .children { margin-left: 20px; }
        .expand-btn { cursor: pointer; user-select: none; margin-right: 5px; }
        .expand-btn:hover { color: #007bff; }
        .stats { background: #e8f5e8; padding: 10px; border-radius: 4px; margin: 10px 0; font-size: 12px; }
        .error { background: #f8d7da; color: #721c24; padding: 10px; border-radius: 4px; margin: 10px 0; }
        .loading { text-align: center; padding: 20px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌳 FairyGUI 渲染树查看器</h1>
            <div class="controls">
                <input type="text" id="nameFilter" placeholder="按名称筛选" />
                <select id="typeFilter">
                    <option value="">所有类型</option>
                    <option value="GList">列表</option>
                    <option value="GButton">按钮</option>
                    <option value="GTextField">文本</option>
                    <option value="GImage">图片</option>
                    <option value="GGraph">图形</option>
                </select>
                <label><input type="checkbox" id="visibleOnly" /> 仅可见对象</label>
                <button onclick="refreshTree()">🔄 刷新</button>
                <button onclick="expandAll()">📂 展开全部</button>
                <button onclick="collapseAll()">📁 收起全部</button>
            </div>
        </div>

        <div class="tree-container">
            <div id="stats" class="stats" style="display:none;"></div>
            <div id="error" class="error" style="display:none;"></div>
            <div id="loading" class="loading">加载中...</div>
            <div id="tree" class="tree"></div>
        </div>
    </div>

    <script>
        let treeData = null;
        let expandedItems = new Set();

        function refreshTree() {
            const nameFilter = document.getElementById('nameFilter').value;
            const typeFilter = document.getElementById('typeFilter').value;
            const visibleOnly = document.getElementById('visibleOnly').checked;

            const params = new URLSearchParams();
            if (nameFilter) params.set('name', nameFilter);
            if (typeFilter) params.set('type', typeFilter);
            if (visibleOnly) params.set('visible', 'true');

            document.getElementById('loading').style.display = 'block';
            document.getElementById('tree').innerHTML = '';
            document.getElementById('error').style.display = 'none';

            fetch('/api/tree?' + params.toString())
                .then(response => response.json())
                .then(data => {
                    treeData = data;
                    renderTree();
                    updateStats();
                    document.getElementById('loading').style.display = 'none';
                })
                .catch(error => {
                    document.getElementById('error').textContent = '加载失败: ' + error.message;
                    document.getElementById('error').style.display = 'block';
                    document.getElementById('loading').style.display = 'none';
                });
        }

        function renderTree() {
            const container = document.getElementById('tree');
            container.innerHTML = '';

            if (treeData.children && treeData.children.length > 0) {
                treeData.children.forEach(item => {
                    renderTreeItem(item, container, 0);
                });
            } else {
                container.innerHTML = '<div style="text-align:center; color:#666; padding:20px;">没有找到对象</div>';
            }
        }

        function renderTreeItem(item, container, depth) {
            const div = document.createElement('div');
            div.className = 'tree-item';
            div.style.marginLeft = (depth * 20) + 'px';
            div.dataset.id = item.ID;

            let hasChildren = item.children && item.children.length > 0;
            let isExpanded = expandedItems.has(item.ID);

            let expandBtn = '';
            if (hasChildren) {
                expandBtn = '<span class="expand-btn" onclick="toggleExpand(\'' + item.ID + '\')">' + (isExpanded ? '▼' : '▶') + '</span>';
            } else {
                expandBtn = '<span style="margin-right: 5px;">•</span>';
            }

            let virtualInfo = '';
            if (item.VirtualInfo && item.VirtualInfo.IsVirtual) {
                const vi = item.VirtualInfo;
                virtualInfo = '<div class="virtual-info">' +
                    '🔄 虚拟列表: ' + vi.NumItems + '项/' + vi.RealNumItems + '实项, ' +
                    '视图: ' + vi.ViewWidth + 'x' + vi.ViewHeight + ', ' +
                    '项目: ' + vi.ItemSize + ', ' +
                    '起始: ' + vi.FirstIndex + ', ' +
                    '子对象: ' + vi.ChildrenCount +
                    '</div>';
            }

            div.innerHTML = expandBtn +
                '<span class="object-name">' + escapeHtml(item.Name) + '</span> ' +
                '<span class="object-type">(' + item.Type + ')</span>' +
                '<span class="object-props">' +
                    ' pos:' + item.Position +
                    ' size:' + item.Size +
                    (item.Rotation != 0 ? ' rot:' + item.Rotation.toFixed(1) : '') +
                    (item.Alpha != 1 ? ' alpha:' + item.Alpha.toFixed(2) : '') +
                    (!item.Visible ? ' [隐藏]' : '') +
                '</span>' +
                virtualInfo;

            container.appendChild(div);

            if (hasChildren && isExpanded) {
                const childrenDiv = document.createElement('div');
                childrenDiv.className = 'children';
                childrenDiv.id = 'children-' + item.ID;
                container.appendChild(childrenDiv);

                item.children.forEach(child => {
                    renderTreeItem(child, childrenDiv, depth + 1);
                });
            }
        }

        function toggleExpand(id) {
            if (expandedItems.has(id)) {
                expandedItems.delete(id);
            } else {
                expandedItems.add(id);
            }
            renderTree();
        }

        function expandAll() {
            function addAllIds(items) {
                items.forEach(item => {
                    if (item.children && item.children.length > 0) {
                        expandedItems.add(item.ID);
                        addAllIds(item.children);
                    }
                });
            }
            addAllIds(treeData.children || []);
            renderTree();
        }

        function collapseAll() {
            expandedItems.clear();
            renderTree();
        }

        function updateStats() {
            let totalCount = 0;
            let visibleCount = 0;
            let virtualLists = 0;

            function countItems(items) {
                items.forEach(item => {
                    totalCount++;
                    if (item.Visible) visibleCount++;
                    if (item.VirtualInfo && item.VirtualInfo.IsVirtual) virtualLists++;
                    if (item.children) countItems(item.children);
                });
            }

            if (treeData.children) {
                countItems(treeData.children);
            }

            const stats = document.getElementById('stats');
            stats.innerHTML = '📊 总计: ' + totalCount + ' 对象, ' + visibleCount + ' 可见, ' + virtualLists + ' 虚拟列表';
            stats.style.display = 'block';
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        // 初始加载
        refreshTree();

        // 自动刷新（每5秒）
        setInterval(refreshTree, 5000);
    </script>
</body>
</html>
	`;

	w.Write([]byte(tmpl))
}

// collectObjectInfo 收集对象信息
func (ds *DebugServer) collectObjectInfo(obj *core.GObject, filterType string, visibleOnly bool, nameFilter string, depth int) []ObjectInfo {
	if obj == nil {
		return nil
	}

	// 应用筛选条件
	if visibleOnly && !obj.Visible() {
		return nil
	}

	objName := obj.Name()
	if nameFilter != "" && !strings.Contains(strings.ToLower(objName), strings.ToLower(nameFilter)) {
		return nil
	}

	objType := getObjectType(obj)
	if filterType != "" && objType != filterType {
		return nil
	}

	info := ObjectInfo{
		ID:       fmt.Sprintf("%p", obj),
		Name:     objName,
		Type:     objType,
		Position: fmt.Sprintf("%.0f,%.0f", obj.X(), obj.Y()),
		Size:     fmt.Sprintf("%.0fx%.0f", obj.Width(), obj.Height()),
		Scale:    fmt.Sprintf("%.2fx%.2f", obj.ScaleX(), obj.ScaleY()),
		Rotation: obj.Rotation(),
		Alpha:    obj.Alpha(),
		Visible:  obj.Visible(),
	}

	// 特殊处理虚拟列表
	if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
		info.VirtualInfo = ds.getVirtualListInfo(list)
	}

	// 收集子对象
	if comp, ok := obj.Data().(*core.GComponent); ok {
		children := comp.Children()
		for _, child := range children {
			if child != nil {
				childInfo := ds.collectObjectInfo(child, filterType, visibleOnly, nameFilter, depth+1)
				info.Children = append(info.Children, childInfo...)
			}
		}
	}

	return []ObjectInfo{info}
}

// getObjectType 获取对象类型
func getObjectType(obj *core.GObject) string {
	if obj == nil {
		return "Unknown"
	}

	// 检查内嵌的数据类型
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

	// 获取实际项目数量（循环模式）
	if comp := list.ComponentRoot(); comp != nil {
		// 这里需要根据实际实现获取realNumItems
		// 暂时使用NumItems作为替代
		info.RealNumItems = list.NumItems()
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

// findObjectByID 根据ID查找对象
func (ds *DebugServer) findObjectByID(obj *core.GObject, id string) *core.GObject {
	if obj == nil {
		return nil
	}

	// 检查当前对象
	if fmt.Sprintf("%p", obj) == id {
		return obj
	}

	// 递归检查子对象
	if comp, ok := obj.Data().(*core.GComponent); ok {
		children := comp.Children()
		for _, child := range children {
			if child != nil {
				if found := ds.findObjectByID(child, id); found != nil {
					return found
				}
			}
		}
	}

	return nil
}

// getObjectDetail 获取对象详细信息
func (ds *DebugServer) getObjectDetail(obj *core.GObject) map[string]interface{} {
	if obj == nil {
		return map[string]interface{}{"error": "对象为空"}
	}

	detail := map[string]interface{}{
		"id":          fmt.Sprintf("%p", obj),
		"name":        obj.Name(),
		"type":        getObjectType(obj),
		"position":    map[string]float64{"x": obj.X(), "y": obj.Y()},
		"size":        map[string]float64{"width": obj.Width(), "height": obj.Height()},
		"scale":       map[string]float64{"x": obj.ScaleX(), "y": obj.ScaleY()},
		"rotation":    obj.Rotation(),
		"alpha":       obj.Alpha(),
		"visible":     obj.Visible(),
		"pivot":       map[string]float64{"x": obj.PivotX(), "y": obj.PivotY()},
		"parent":      obj.Parent() != nil,
		"displayObject": obj.DisplayObject() != nil,
	}

	// 特殊处理不同类型的对象
	switch data := obj.Data().(type) {
	case *widgets.GList:
		detail["virtual"] = data.IsVirtual()
		detail["numItems"] = data.NumItems()
		detail["childrenCount"] = len(data.GComponent.Children())
		if data.IsVirtual() {
			detail["virtualInfo"] = ds.getVirtualListInfo(data)
		}
	case *widgets.GButton:
		detail["selected"] = data.Selected()
		detail["title"] = data.Title()
		detail["icon"] = data.Icon()
	case *widgets.GTextField:
		detail["text"] = data.Text()
		detail["fontSize"] = data.FontSize()
		detail["color"] = data.Color()
	case *core.GComponent:
		detail["childrenCount"] = len(data.Children())
		detail["hasScrollPane"] = data.ScrollPane() != nil
		if scrollPane := data.ScrollPane(); scrollPane != nil {
			detail["scrollPos"] = map[string]float64{
				"x": scrollPane.PosX(),
				"y": scrollPane.PosY(),
			}
			detail["viewSize"] = map[string]float64{
				"width":  scrollPane.ViewWidth(),
				"height": scrollPane.ViewHeight(),
			}
		}
	}

	return detail
}

func main() {
	// 这里需要一个实际的根对象，可以从demo中获取
	fmt.Println("FairyGUI Debug Server")
	fmt.Println("===================")
	fmt.Println("这个工具需要集成到实际的FairyGUI应用中")
	fmt.Println("使用方法:")
	fmt.Println("1. 在你的应用中创建DebugServer实例")
	fmt.Println("2. 传入你的GRoot或其他根组件")
	fmt.Println("3. 调用Start(port)启动服务器")
	fmt.Println("")
	fmt.Println("示例代码:")
	fmt.Println("  debugServer := NewDebugServer(yourRootComponent)")
	fmt.Println("  go debugServer.Start(8080) // 在goroutine中运行")
	fmt.Println("")
	fmt.Println("然后访问 http://localhost:8080 查看调试界面")
}