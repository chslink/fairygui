package debug

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

// Server 调试服务器
type Server struct {
	root    *core.GComponent
	port    int
	enabled bool
}

// NewServer 创建调试服务器
func NewServer(root *core.GComponent, port int) *Server {
	return &Server{
		root: root,
		port: port,
	}
}

// Start 启动调试服务器
func (s *Server) Start() error {
	if s.enabled {
		return fmt.Errorf("调试服务器已在运行")
	}

	// 在goroutine中启动，避免阻塞主应用
	go func() {
		log.Printf("🛠️ 启动调试服务器在端口 %d", s.port)
		log.Printf("📊 访问 http://localhost:%d 查看调试界面", s.port)

		// 设置路由
		http.HandleFunc("/", s.handleIndex)
		http.HandleFunc("/tree", s.handleTreeView)
		http.HandleFunc("/api/tree", s.handleTreeAPI)
		http.HandleFunc("/api/virtual-lists", s.handleVirtualListsAPI)

		addr := fmt.Sprintf(":%d", s.port)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("❌ 调试服务器启动失败: %v", err)
		}
	}()

	s.enabled = true
	return nil
}

// IsEnabled 返回调试服务器是否启用
func (s *Server) IsEnabled() bool {
	return s.enabled
}

// GetURL 返回调试服务器的访问URL
func (s *Server) GetURL() string {
	return fmt.Sprintf("http://localhost:%d", s.port)
}

// handleIndex 处理首页
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
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
        .feature-list { margin: 10px 0; }
        .feature-list li { margin: 5px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🛠️ FairyGUI Debug Server</h1>
        <p>实时查看和分析 FairyGUI 渲染树结构</p>
    </div>

    <div class="link">📊 <a href="/tree">查看渲染树</a></div>
    <div class="link">📋 <a href="/api/tree">获取JSON数据</a></div>
    <div class="link">🔄 <a href="/api/virtual-lists">虚拟列表专项分析</a></div>

    <div class="info">
        <h3>✨ 功能特性：</h3>
        <ul class="feature-list">
            <li><strong>渲染树视图</strong>：以树形结构显示所有UI对象</li>
            <li><strong>虚拟列表专项</strong>：专门分析虚拟列表状态</li>
            <li><strong>实时更新</strong>：数据每5秒自动刷新</li>
            <li><strong>轻量级</strong>：最小性能开销</li>
        </ul>
    </div>

    <div class="info">
        <h3>🔧 集成说明：</h3>
        <p>此调试服务器已集成到 FairyGUI Ebiten Demo 中，可以实时查看当前场景的UI结构。</p>
        <p>特别适用于调试虚拟列表等复杂组件。</p>
    </div>
</body>
</html>
	`)
}

// handleTreeView 处理树形视图
func (s *Server) handleTreeView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 解析查询参数
	filterType := r.URL.Query().Get("type")
	filterName := r.URL.Query().Get("name")
	filterVisible := r.URL.Query().Get("visible")
	showDetails := r.URL.Query().Get("details") != "false" // 默认显示详细信息

	treeData := s.collectTreeDataWithFilter(filterType, filterName, filterVisible, showDetails)

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
        .detail-info { background: #e8f4fd; padding: 3px 6px; border-radius: 3px; margin: 2px 0; font-size: 11px; color: #2c5282; border-left: 3px solid #3182ce; }
        .children { margin-left: 20px; }
        .expand-btn { cursor: pointer; user-select: none; margin-right: 5px; }
        .stats { background: #e8f5e8; padding: 10px; border-radius: 4px; margin: 10px 0; }
        .refresh-btn { position: fixed; top: 20px; right: 20px; padding: 10px 20px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #0056b3; }
        .debug-info { background: #d1ecf1; padding: 10px; border-radius: 4px; margin: 10px 0; font-size: 12px; }
        .filter-panel { background: #f8f9fa; padding: 15px; border-radius: 8px; margin-bottom: 20px; border: 1px solid #dee2e6; }
        .filter-row { margin: 8px 0; }
        .filter-label { display: inline-block; width: 80px; font-weight: bold; }
        .filter-input { padding: 5px 8px; border: 1px solid #ccc; border-radius: 4px; width: 200px; }
        .filter-checkbox { margin-left: 10px; }
        .filter-btn { padding: 6px 12px; background: #28a745; color: white; border: none; border-radius: 4px; cursor: pointer; margin-left: 10px; }
        .filter-btn:hover { background: #218838; }
        .query-params { background: #fff3cd; padding: 8px; border-radius: 4px; margin: 10px 0; font-size: 11px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌳 FairyGUI 渲染树</h1>
            <button class="refresh-btn" onclick="location.reload()">🔄 刷新</button>
        </div>

        <div class="debug-info">
            <strong>💡 提示：</strong> 此视图显示当前demo场景中的所有UI对象。虚拟列表会显示专项信息。
            <br>数据每5秒自动更新，也可以手动刷新页面查看最新状态。
        </div>

        <!-- 筛选面板 -->
        <div class="filter-panel">
            <h3>🔍 筛选选项</h3>
            <form id="filterForm" onsubmit="applyFilter(event)">
                <div class="filter-row">
                    <label class="filter-label">对象类型:</label>
                    <input type="text" id="typeFilter" class="filter-input" placeholder="如: GButton, GList"
                           value="%s" onchange="updateQueryString()">
                </div>
                <div class="filter-row">
                    <label class="filter-label">对象名称:</label>
                    <input type="text" id="nameFilter" class="filter-input" placeholder="包含的名称"
                           value="%s" onchange="updateQueryString()">
                </div>
                <div class="filter-row">
                    <label class="filter-label">可见性:</label>
                    <select id="visibleFilter" class="filter-input" onchange="updateQueryString()">
                        <option value="">全部</option>
                        <option value="true" %s>仅可见</option>
                        <option value="false" %s>仅隐藏</option>
                    </select>
                    <label class="filter-checkbox">
                        <input type="checkbox" id="showDetails" %s onchange="updateQueryString()"> 显示详细信息
                    </label>
                    <button type="submit" class="filter-btn">应用筛选</button>
                </div>
            </form>
            <div class="query-params" id="queryParams"></div>
        </div>

        <div class="tree-container">
            <div id="stats" class="stats">%s</div>
            <div id="tree" class="tree">%s</div>
        </div>
    </div>

    <script>
        // 设置初始值
        document.getElementById('visibleFilter').value = '%s';

        // 更新查询字符串
        function updateQueryString() {
            const params = new URLSearchParams();
            const typeFilter = document.getElementById('typeFilter').value;
            const nameFilter = document.getElementById('nameFilter').value;
            const visibleFilter = document.getElementById('visibleFilter').value;
            const showDetails = document.getElementById('showDetails').checked;

            if (typeFilter) params.set('type', typeFilter);
            if (nameFilter) params.set('name', nameFilter);
            if (visibleFilter) params.set('visible', visibleFilter);
            if (!showDetails) params.set('details', 'false');

            const queryString = params.toString();
            document.getElementById('queryParams').textContent = queryString ? '当前查询: ' + queryString : '';

            // 更新URL但不重新加载
            const newUrl = queryString ? '?' + queryString : window.location.pathname;
            window.history.replaceState({}, '', newUrl);
        }

        // 应用筛选
        function applyFilter(event) {
            event.preventDefault();
            updateQueryString();
            location.reload();
        }

        // 每5秒自动刷新
        setTimeout(function() {
            location.reload();
        }, 5000);

        // 初始化查询参数显示
        updateQueryString();
    </script>
</body>
</html>
	`, filterType, filterName,
		getSelectedAttr(filterVisible == "true"), getSelectedAttr(filterVisible == "false"),
		getCheckedAttr(showDetails), treeData.Stats, treeData.HTML, filterVisible)
}

// handleTreeAPI 处理树形数据API
func (s *Server) handleTreeAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data := s.collectJSONData()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}

// handleVirtualListsAPI 处理虚拟列表专项分析API
func (s *Server) handleVirtualListsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	vlists := s.collectVirtualLists()
	response := map[string]interface{}{
		"virtual_lists": vlists,
		"count":         len(vlists),
		"timestamp":     time.Now().Format("15:04:05"),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response)
}

// TreeData 树形数据
type TreeData struct {
	Stats string
	HTML  string
}

// collectTreeData 收集树形数据用于HTML显示
func (s *Server) collectTreeData() TreeData {
	if s.root == nil {
		return TreeData{
			Stats: "错误：根组件为空",
			HTML:  "<div style='color:red;'>根组件未设置</div>",
		}
	}

	var stats struct {
		total       int
		visible     int
		virtual     int
		gcomponents int
	}

	var html strings.Builder
	html.WriteString("<div class='tree'>")

	// 遍历根对象的子对象
	if rootObj := s.root.GObject; rootObj != nil {
		if comp, ok := rootObj.Data().(*core.GComponent); ok {
			children := comp.Children()
			for _, child := range children {
				if child != nil {
					html.WriteString(s.renderTreeItem(child, 0, &stats))
				}
			}
		}
	}

	html.WriteString("</div>")

	return TreeData{
		Stats: fmt.Sprintf("📊 总计: %d 对象, %d 可见, %d 虚拟列表, %d 容器",
			stats.total, stats.visible, stats.virtual, stats.gcomponents),
		HTML:  html.String(),
	}
}

// renderTreeItem 渲染树形项目
func (s *Server) renderTreeItem(obj *core.GObject, depth int, stats *struct {
	total       int
	visible     int
	virtual     int
	gcomponents int
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
	objType := s.getObjectType(obj)
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

	// 添加对象特定属性的详细显示
	s.renderObjectDetails(obj, &result)

	// 虚拟列表特殊信息
	if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
		stats.virtual++
		info := s.getVirtualListInfo(list)
		if info != nil {
			result.WriteString(fmt.Sprintf("<div class='virtual-info'>🔄 虚拟列表: %d项, 视图:%dx%d, 子对象:%d</div>",
				info.NumItems, info.ViewWidth, info.ViewHeight, info.ChildrenCount))
		}
	}

	result.WriteString("</div>")

	// 递归处理子对象
	if comp, ok := obj.Data().(*core.GComponent); ok {
		stats.gcomponents++
		children := comp.Children()
		for _, child := range children {
			if child != nil {
				result.WriteString(s.renderTreeItem(child, depth+1, stats))
			}
		}
	}

	return result.String()
}

// collectJSONData 收集JSON数据
func (s *Server) collectJSONData() map[string]interface{} {
	if s.root == nil {
		return map[string]interface{}{
			"error": "根组件为空",
		}
	}

	var objects []map[string]interface{}

	// 遍历根对象的子对象
	if rootObj := s.root.GObject; rootObj != nil {
		if comp, ok := rootObj.Data().(*core.GComponent); ok {
			children := comp.Children()
			for _, child := range children {
				if child != nil {
					objects = append(objects, s.collectObjectJSON(child))
				}
			}
		}
	}

	return map[string]interface{}{
		"timestamp": time.Now().Format("15:04:05"),
		"root":      "GRoot",
		"objects":   objects,
		"count":     len(objects),
	}
}

// collectObjectJSON 收集对象JSON数据
func (s *Server) collectObjectJSON(obj *core.GObject) map[string]interface{} {
	if obj == nil {
		return nil
	}

	data := map[string]interface{}{
		"id":       fmt.Sprintf("%p", obj),
		"name":     obj.Name(),
		"type":     s.getObjectType(obj),
		"position": map[string]float64{"x": obj.X(), "y": obj.Y()},
		"size":     map[string]float64{"width": obj.Width(), "height": obj.Height()},
		"rotation": obj.Rotation(),
		"alpha":    obj.Alpha(),
		"visible":  obj.Visible(),
	}

	// 添加更多详细信息
	if obj.Parent() != nil {
		data["parent"] = obj.Parent().Name()
	}

	// 添加对象特定属性
	s.addObjectSpecificData(obj, data)

	// 虚拟列表特殊处理
	if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
		info := s.getVirtualListInfo(list)
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
				childData = append(childData, s.collectObjectJSON(child))
			}
		}
		if len(childData) > 0 {
			data["children"] = childData
		}

		// 添加组件特定信息
		if scrollPane := comp.ScrollPane(); scrollPane != nil {
			data["scrollPane"] = map[string]interface{}{
				"viewWidth":  scrollPane.ViewWidth(),
				"viewHeight": scrollPane.ViewHeight(),
			}
		}

		// 添加控制器信息
		if controllers := comp.Controllers(); len(controllers) > 0 {
			var ctrlData []map[string]interface{}
			for _, ctrl := range controllers {
				if ctrl != nil {
					ctrlData = append(ctrlData, map[string]interface{}{
						"name":          ctrl.Name,
						"selectedIndex": ctrl.SelectedIndex(),
						"selectedPage":  ctrl.SelectedPageID(),
						"pageCount":     ctrl.PageCount(),
					})
				}
			}
			data["controllers"] = ctrlData
		}
	}

	return data
}

// collectTreeDataWithFilter 收集带筛选条件的树形数据
func (s *Server) collectTreeDataWithFilter(filterType, filterName, filterVisible string, showDetails bool) TreeData {
	if s.root == nil {
		return TreeData{
			Stats: "错误：根组件为空",
			HTML:  "<div style='color:red;'>根组件未设置</div>",
		}
	}

	var stats struct {
		total       int
		visible     int
		virtual     int
		gcomponents int
		filtered    int
	}

	var html strings.Builder
	html.WriteString("<div class='tree'>")

	// 遍历根对象的子对象
	if rootObj := s.root.GObject; rootObj != nil {
		if comp, ok := rootObj.Data().(*core.GComponent); ok {
			children := comp.Children()
			for _, child := range children {
				if child != nil {
					html.WriteString(s.renderTreeItemWithFilter(child, 0, &stats, filterType, filterName, filterVisible, showDetails))
				}
			}
		}
	}

	html.WriteString("</div>")

	filterInfo := ""
	if filterType != "" || filterName != "" || filterVisible != "" {
		filterInfo = fmt.Sprintf(", 筛选后: %d 对象", stats.filtered)
	}

	return TreeData{
		Stats: fmt.Sprintf("📊 总计: %d 对象, %d 可见, %d 虚拟列表, %d 容器%s",
			stats.total, stats.visible, stats.virtual, stats.gcomponents, filterInfo),
		HTML:  html.String(),
	}
}

// renderTreeItemWithFilter 渲染带筛选条件的树形项目
func (s *Server) renderTreeItemWithFilter(obj *core.GObject, depth int, stats *struct {
	total       int
	visible     int
	virtual     int
	gcomponents int
	filtered    int
}, filterType, filterName, filterVisible string, showDetails bool) string {
	if obj == nil {
		return ""
	}

	stats.total++
	if obj.Visible() {
		stats.visible++
	}

	// 应用筛选条件
	if !s.shouldIncludeObject(obj, filterType, filterName, filterVisible) {
		return ""
	}

	stats.filtered++

	var result strings.Builder

	// 基本信息
	objType := s.getObjectType(obj)
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

	// 添加对象特定属性的详细显示
	if showDetails {
		s.renderObjectDetails(obj, &result)
	}

	// 虚拟列表特殊信息
	if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
		stats.virtual++
		info := s.getVirtualListInfo(list)
		if info != nil {
			result.WriteString(fmt.Sprintf("<div class='virtual-info'>🔄 虚拟列表: %d项, 视图:%dx%d, 子对象:%d</div>",
				info.NumItems, info.ViewWidth, info.ViewHeight, info.ChildrenCount))
		}
	}

	result.WriteString("</div>")

	// 递归处理子对象
	if comp, ok := obj.Data().(*core.GComponent); ok {
		stats.gcomponents++
		children := comp.Children()
		for _, child := range children {
			if child != nil {
				result.WriteString(s.renderTreeItemWithFilter(child, depth+1, stats, filterType, filterName, filterVisible, showDetails))
			}
		}
	}

	return result.String()
}

// shouldIncludeObject 判断对象是否应该包含在结果中
func (s *Server) shouldIncludeObject(obj *core.GObject, filterType, filterName, filterVisible string) bool {
	// 类型筛选
	if filterType != "" {
		objType := s.getObjectType(obj)
		if !strings.Contains(strings.ToLower(objType), strings.ToLower(filterType)) {
			return false
		}
	}

	// 名称筛选
	if filterName != "" {
		if !strings.Contains(strings.ToLower(obj.Name()), strings.ToLower(filterName)) {
			return false
		}
	}

	// 可见性筛选
	if filterVisible != "" {
		visible := filterVisible == "true"
		if obj.Visible() != visible {
			return false
		}
	}

	return true
}

// getSelectedAttr 获取选中属性
func getSelectedAttr(selected bool) string {
	if selected {
		return "selected"
	}
	return ""
}

// getCheckedAttr 获取勾选属性
func getCheckedAttr(checked bool) string {
	if checked {
		return "checked"
	}
	return ""
}

// collectVirtualLists 收集所有虚拟列表信息
func (s *Server) collectVirtualLists() []map[string]interface{} {
	var vlists []map[string]interface{}

	// 递归查找所有虚拟列表
	s.findVirtualLists(s.root.GObject, &vlists)
	return vlists
}

// findVirtualLists 递归查找虚拟列表
func (s *Server) findVirtualLists(obj *core.GObject, result *[]map[string]interface{}) {
	if obj == nil {
		return
	}

	// 检查当前对象是否是虚拟列表
	if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
		info := s.getVirtualListInfo(list)
		if info != nil {
			*result = append(*result, map[string]interface{}{
				"name":          obj.Name(),
				"id":            fmt.Sprintf("%p", obj),
				"position":      map[string]float64{"x": obj.X(), "y": obj.Y()},
				"size":          map[string]float64{"width": obj.Width(), "height": obj.Height()},
				"virtual_info":  info,
				"children_count": list.ChildrenCount(), // 使用新的ChildrenCount方法
			})
		}
	}

	// 递归检查子对象
	if comp, ok := obj.Data().(*core.GComponent); ok {
		children := comp.Children()
		for _, child := range children {
			if child != nil {
				s.findVirtualLists(child, result)
			}
		}
	}
}

// getObjectType 获取对象类型
func (s *Server) getObjectType(obj *core.GObject) string {
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

// addObjectSpecificData 添加对象特定属性数据
func (s *Server) addObjectSpecificData(obj *core.GObject, data map[string]interface{}) {
	switch widget := obj.Data().(type) {
	case *widgets.GButton:
		data["title"] = widget.Title()
		data["selected"] = widget.Selected()
		if icon := widget.Icon(); icon != "" {
			data["icon"] = icon
		}
	case *widgets.GTextField:
		data["text"] = widget.Text()
		data["fontSize"] = widget.FontSize()
		data["color"] = fmt.Sprintf("#%06X", widget.Color())
		data["align"] = widget.Align()
		data["valign"] = widget.VerticalAlign()
		data["singleLine"] = widget.SingleLine()
		data["autoSize"] = widget.AutoSize()
	case *widgets.GImage:
		data["flip"] = widget.Flip()
	case *widgets.GList:
		data["layout"] = widget.Layout()
		data["lineCount"] = widget.LineCount()
		data["columnCount"] = widget.ColumnCount()
		data["lineGap"] = widget.LineGap()
		data["columnGap"] = widget.ColumnGap()
		data["autoResizeItem"] = widget.AutoResizeItem()
		data["childrenCount"] = widget.ChildrenCount() // 添加children数量属性
		if widget.IsVirtual() {
			data["virtual"] = true
			data["numItems"] = widget.NumItems()
			if itemSize := widget.VirtualItemSize(); itemSize != nil {
				data["virtualItemSize"] = map[string]float64{
					"width":  itemSize.X,
					"height": itemSize.Y,
				}
			}
		}
	case *widgets.GComboBox:
		data["items"] = widget.Items()
		data["values"] = widget.Values()
		data["icons"] = widget.Icons()
		data["selectedIndex"] = widget.SelectedIndex()
		data["visibleItemCount"] = widget.VisibleItemCount()
	case *widgets.GProgressBar:
		data["value"] = widget.Value()
		data["max"] = widget.Max()
		data["min"] = widget.Min()
		data["titleType"] = widget.TitleType()
	case *widgets.GSlider:
		data["value"] = widget.Value()
		data["max"] = widget.Max()
		data["min"] = widget.Min()
	case *widgets.GMovieClip:
		data["playing"] = widget.Playing()
		data["frame"] = widget.Frame()
		data["timeScale"] = widget.TimeScale()
	case *widgets.GTree:
		data["indent"] = widget.Indent()
		data["clickToExpand"] = widget.ClickToExpand()
	case *widgets.GTextInput:
		data["text"] = widget.Text()
		data["maxLength"] = widget.MaxLength()
		data["restrict"] = widget.Restrict()
		data["editable"] = widget.Editable()
		data["promptText"] = widget.PromptText()
	case *widgets.GGroup:
		data["layout"] = widget.Layout()
		data["lineGap"] = widget.LineGap()
		data["columnGap"] = widget.ColumnGap()
		data["excludeInvisibles"] = widget.ExcludeInvisibles()
		data["autoSizeDisabled"] = widget.AutoSizeDisabled()
	}

	// 为所有 GComponent 添加控制器信息
	if comp, ok := obj.Data().(*core.GComponent); ok {
		if controllers := comp.Controllers(); len(controllers) > 0 {
			var ctrlData []map[string]interface{}
			for _, ctrl := range controllers {
				if ctrl != nil {
					ctrlData = append(ctrlData, map[string]interface{}{
						"name":          ctrl.Name,
						"selectedIndex": ctrl.SelectedIndex(),
						"selectedPage":  ctrl.SelectedPageID(),
						"pageCount":     ctrl.PageCount(),
					})
				}
			}
			data["controllers"] = ctrlData
		}
	}
}

// getVirtualListInfo 获取虚拟列表信息
func (s *Server) getVirtualListInfo(list *widgets.GList) *VirtualListInfo {
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

// renderObjectDetails 渲染对象详细信息
func (s *Server) renderObjectDetails(obj *core.GObject, result *strings.Builder) {
	switch widget := obj.Data().(type) {
	case *widgets.GButton:
		if title := widget.Title(); title != "" {
			result.WriteString(fmt.Sprintf("<div class='detail-info'>🔘 按钮: %s", title))
			if widget.Selected() {
				result.WriteString(" [已选择]")
			}
			result.WriteString("</div>")
		}
	case *widgets.GTextField:
		if text := widget.Text(); text != "" {
			result.WriteString(fmt.Sprintf("<div class='detail-info'>📝 文本: \"%s\" (字体:%dpx, 颜色:#%06X)",
				text, widget.FontSize(), widget.Color()))
			if widget.SingleLine() {
				result.WriteString(" [单行]")
			}
			if autoSize := widget.AutoSize(); autoSize != 0 {
				result.WriteString(fmt.Sprintf(" [自动大小:%d]", autoSize))
			}
			result.WriteString("</div>")
		}
	case *widgets.GImage:
		if flip := widget.Flip(); flip != 0 {
			result.WriteString(fmt.Sprintf("<div class='detail-info'>🖼️ 图片: 翻转=%d</div>", flip))
		}
	case *widgets.GList:
		result.WriteString(fmt.Sprintf("<div class='detail-info'>📋 列表: 布局=%d", widget.Layout()))
		if widget.LineCount() > 0 {
			result.WriteString(fmt.Sprintf(", 行数=%d", widget.LineCount()))
		}
		if widget.ColumnCount() > 0 {
			result.WriteString(fmt.Sprintf(", 列数=%d", widget.ColumnCount()))
		}
		if widget.LineGap() != 0 {
			result.WriteString(fmt.Sprintf(", 行间距=%d", widget.LineGap()))
		}
		if widget.ColumnGap() != 0 {
			result.WriteString(fmt.Sprintf(", 列间距=%d", widget.ColumnGap()))
		}
		if widget.IsVirtual() {
			result.WriteString(fmt.Sprintf(" [虚拟列表: %d项]", widget.NumItems()))
			if itemSize := widget.VirtualItemSize(); itemSize != nil {
				result.WriteString(fmt.Sprintf(" [项目尺寸:%.0fx%.0f]", itemSize.X, itemSize.Y))
			}
		}
		result.WriteString(fmt.Sprintf(" [子对象:%d]", widget.ChildrenCount())) // 添加children数量显示
		result.WriteString("</div>")
	case *widgets.GComboBox:
		items := widget.Items()
		if len(items) > 0 {
			result.WriteString(fmt.Sprintf("<div class='detail-info'>🗂️ 下拉框: %d个选项, 选中=%d",
				len(items), widget.SelectedIndex()))
			if visibleCount := widget.VisibleItemCount(); visibleCount > 0 {
				result.WriteString(fmt.Sprintf(", 可见项=%d", visibleCount))
			}
			result.WriteString("</div>")
		}
	case *widgets.GProgressBar:
		result.WriteString(fmt.Sprintf("<div class='detail-info'>📊 进度条: %.1f/%.1f",
			widget.Value(), widget.Max()))
		if titleType := widget.TitleType(); titleType != 0 {
			result.WriteString(fmt.Sprintf(", 标题类型=%d", titleType))
		}
		result.WriteString("</div>")
	case *widgets.GSlider:
		result.WriteString(fmt.Sprintf("<div class='detail-info'>🎚️ 滑块: %.1f/%.1f",
			widget.Value(), widget.Max()))
		result.WriteString("</div>")
	case *widgets.GMovieClip:
		if playing := widget.Playing(); playing {
			result.WriteString(fmt.Sprintf("<div class='detail-info'>🎬 动画: 播放中 (帧%d, 速度%.1f)",
				widget.Frame(), widget.TimeScale()))
		} else {
			result.WriteString(fmt.Sprintf("<div class='detail-info'>🎬 动画: 暂停 (帧%d, 速度%.1f)",
				widget.Frame(), widget.TimeScale()))
		}
		result.WriteString("</div>")
	case *widgets.GTree:
		result.WriteString("<div class='detail-info'>🌳 树形控件</div>")
	case *widgets.GTextInput:
		if text := widget.Text(); text != "" {
			result.WriteString(fmt.Sprintf("<div class='detail-info'>📝 输入框: \"%s\"", text))
		} else if prompt := widget.PromptText(); prompt != "" {
			result.WriteString(fmt.Sprintf("<div class='detail-info'>📝 输入框: 提示\"%s\"", prompt))
		} else {
			result.WriteString("<div class='detail-info'>📝 输入框: 空")
		}
		if !widget.Editable() {
			result.WriteString(" [只读]")
		}
		if maxLen := widget.MaxLength(); maxLen > 0 {
			result.WriteString(fmt.Sprintf(" [最大长度:%d]", maxLen))
		}
		result.WriteString("</div>")
	case *widgets.GGroup:
		result.WriteString(fmt.Sprintf("<div class='detail-info'>📦 组: 布局=%d", widget.Layout()))
		if widget.LineGap() != 0 {
			result.WriteString(fmt.Sprintf(", 行间距=%d", widget.LineGap()))
		}
		if widget.ColumnGap() != 0 {
			result.WriteString(fmt.Sprintf(", 列间距=%d", widget.ColumnGap()))
		}
		if widget.ExcludeInvisibles() {
			result.WriteString(" [排除隐藏项]")
		}
		result.WriteString("</div>")
	}

	// 添加父对象信息
	if parent := obj.Parent(); parent != nil {
		result.WriteString(fmt.Sprintf("<div class='detail-info'>👪 父对象: %s (%s)</div>",
			parent.Name(), s.getObjectType(parent.GObject)))
	}

	// 添加控制器信息
	if comp, ok := obj.Data().(*core.GComponent); ok {
		if controllers := comp.Controllers(); len(controllers) > 0 {
			result.WriteString(fmt.Sprintf("<div class='detail-info'>🎮 控制器: %d个", len(controllers)))
			for i, ctrl := range controllers {
				if ctrl != nil {
					result.WriteString(fmt.Sprintf("<br>  %d. %s: page=%d/%d (%s)",
						i+1, ctrl.Name, ctrl.SelectedIndex(), ctrl.PageCount(), ctrl.SelectedPageID()))
				}
			}
			result.WriteString("</div>")
		}
	}
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