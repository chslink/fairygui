package debug

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// Server HTTP调试服务器
type Server struct {
	root      *core.GObject
	stage     *laya.Stage
	port      int
	enabled   bool
	inspector *Inspector
	simulator *EventSimulator
}

// NewServer 创建调试服务器
func NewServer(root *core.GObject, stage *laya.Stage, port int) *Server {
	return &Server{
		root:      root,
		stage:     stage,
		port:      port,
		inspector: NewInspector(root),
		simulator: NewEventSimulator(stage),
	}
}

// Start 启动调试服务器
func (s *Server) Start() error {
	if s.enabled {
		return fmt.Errorf("调试服务器已在运行")
	}

	go func() {
		log.Printf("🛠️  启动调试服务器在端口 %d", s.port)
		log.Printf("📊 访问 http://localhost:%d 查看调试界面", s.port)

		// 设置路由
		http.HandleFunc("/", s.handleIndex)
		http.HandleFunc("/tree", s.handleTreeView)
		http.HandleFunc("/api/tree", s.handleTreeAPI)
		http.HandleFunc("/api/object/", s.handleObjectAPI)
		http.HandleFunc("/api/find", s.handleFindAPI)
		http.HandleFunc("/api/click", s.handleClickAPI)
		http.HandleFunc("/api/stats", s.handleStatsAPI)
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
    <meta charset="utf-8">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Arial, sans-serif; background: #f5f7fa; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 40px; border-radius: 12px; margin-bottom: 30px; box-shadow: 0 10px 30px rgba(0,0,0,0.1); }
        .header h1 { font-size: 32px; margin-bottom: 10px; }
        .header p { font-size: 16px; opacity: 0.9; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .card { background: white; padding: 25px; border-radius: 12px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); transition: transform 0.2s, box-shadow 0.2s; }
        .card:hover { transform: translateY(-2px); box-shadow: 0 5px 20px rgba(0,0,0,0.1); }
        .card h3 { color: #333; margin-bottom: 15px; font-size: 18px; display: flex; align-items: center; }
        .card h3 .icon { margin-right: 10px; font-size: 24px; }
        .card p { color: #666; line-height: 1.6; margin-bottom: 15px; font-size: 14px; }
        .card a { display: inline-block; color: #667eea; text-decoration: none; font-weight: 500; padding: 8px 16px; border-radius: 6px; background: #f0f4ff; transition: background 0.2s; }
        .card a:hover { background: #e0e8ff; }
        .api-section { background: white; padding: 30px; border-radius: 12px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
        .api-section h2 { color: #333; margin-bottom: 20px; font-size: 24px; }
        .api-list { list-style: none; }
        .api-list li { padding: 12px 0; border-bottom: 1px solid #eee; }
        .api-list li:last-child { border-bottom: none; }
        .api-method { display: inline-block; padding: 4px 8px; border-radius: 4px; font-weight: 600; font-size: 11px; margin-right: 10px; }
        .api-method.get { background: #d1f4d1; color: #2d7d2d; }
        .api-method.post { background: #ffeaa7; color: #d63031; }
        .api-path { font-family: 'Courier New', monospace; color: #555; }
        .api-desc { color: #888; font-size: 13px; margin-left: 10px; }
        .feature-badge { display: inline-block; background: #667eea; color: white; padding: 4px 10px; border-radius: 12px; font-size: 12px; margin-right: 8px; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🛠️ FairyGUI Debug Server</h1>
            <p>实时调试和检查 FairyGUI 应用程序</p>
        </div>

        <div class="grid">
            <div class="card">
                <h3><span class="icon">🌳</span>渲染树视图</h3>
                <p>以树形结构实时查看所有UI对象，支持筛选和详细信息展示</p>
                <a href="/tree">查看渲染树 →</a>
            </div>

            <div class="card">
                <h3><span class="icon">📊</span>统计信息</h3>
                <p>查看对象统计、性能指标和虚拟列表专项分析</p>
                <a href="/api/stats">查看统计 →</a>
            </div>

            <div class="card">
                <h3><span class="icon">🔍</span>对象查找</h3>
                <p>按名称、类型、路径查找对象，支持复杂筛选条件</p>
                <a href="/api/find?name=">API文档 →</a>
            </div>

            <div class="card">
                <h3><span class="icon">🖱️</span>事件模拟</h3>
                <p>模拟点击、触摸、拖拽等事件，用于自动化测试</p>
                <a href="#api-section">查看API →</a>
            </div>
        </div>

        <div class="api-section" id="api-section">
            <h2>📡 API 接口</h2>
            <ul class="api-list">
                <li>
                    <span class="api-method get">GET</span>
                    <span class="api-path">/api/tree</span>
                    <span class="api-desc">获取完整对象树（JSON格式）</span>
                </li>
                <li>
                    <span class="api-method get">GET</span>
                    <span class="api-path">/api/object/{id}</span>
                    <span class="api-desc">获取指定对象的详细信息</span>
                </li>
                <li>
                    <span class="api-method get">GET</span>
                    <span class="api-path">/api/find?name=xxx&type=xxx</span>
                    <span class="api-desc">查找对象（支持name/type/path/visible参数）</span>
                </li>
                <li>
                    <span class="api-method post">POST</span>
                    <span class="api-path">/api/click</span>
                    <span class="api-desc">模拟点击事件（JSON: {"target":"name|path|id", "x":0, "y":0}）</span>
                </li>
                <li>
                    <span class="api-method get">GET</span>
                    <span class="api-path">/api/stats</span>
                    <span class="api-desc">获取统计信息</span>
                </li>
                <li>
                    <span class="api-method get">GET</span>
                    <span class="api-path">/api/virtual-lists</span>
                    <span class="api-desc">获取所有虚拟列表信息</span>
                </li>
            </ul>

            <h2 style="margin-top: 30px;">✨ 功能特性</h2>
            <div style="margin-top: 15px;">
                <span class="feature-badge">实时更新</span>
                <span class="feature-badge">对象检查</span>
                <span class="feature-badge">事件模拟</span>
                <span class="feature-badge">性能分析</span>
                <span class="feature-badge">筛选查找</span>
                <span class="feature-badge">RESTful API</span>
            </div>
        </div>
    </div>
</body>
</html>
	`)
}

// handleTreeView 处理树形视图（继续使用原有实现，但集成Inspector）
func (s *Server) handleTreeView(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	filterType := r.URL.Query().Get("type")
	filterName := r.URL.Query().Get("name")
	filterVisible := r.URL.Query().Get("visible")
	showDetails := r.URL.Query().Get("details") != "false"

	treeData := s.buildTreeHTML(filterType, filterName, filterVisible, showDetails)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, getTreeViewHTML(), filterType, filterName,
		getSelectedAttr(filterVisible == "true"), getSelectedAttr(filterVisible == "false"),
		getCheckedAttr(showDetails), treeData.Stats, treeData.HTML, filterVisible)
}

// handleTreeAPI 处理树形数据API
func (s *Server) handleTreeAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	tree := s.buildTreeJSON(s.root)
	response := map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"tree":      tree,
	}

	json.NewEncoder(w).Encode(response)
}

// handleObjectAPI 处理对象信息API
func (s *Server) handleObjectAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 从URL中提取对象ID
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "无效的对象ID", http.StatusBadRequest)
		return
	}
	objectID := parts[3]

	obj := s.inspector.FindByID(objectID)
	if obj == nil {
		http.Error(w, fmt.Sprintf("未找到对象: %s", objectID), http.StatusNotFound)
		return
	}

	info := s.inspector.GetInfo(obj)
	info.Properties["path"] = s.inspector.GetPath(obj)

	json.NewEncoder(w).Encode(info)
}

// handleFindAPI 处理查找API
func (s *Server) handleFindAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	query := r.URL.Query()
	var results []*core.GObject

	// 按不同方式查找
	if name := query.Get("name"); name != "" {
		results = s.inspector.FindByName(name)
	} else if typeName := query.Get("type"); typeName != "" {
		results = s.inspector.FindByType(typeName)
	} else if path := query.Get("path"); path != "" {
		if obj := s.inspector.FindByPath(path); obj != nil {
			results = []*core.GObject{obj}
		}
	} else {
		// 使用筛选器
		filter := Filter{
			Name: query.Get("filter_name"),
			Type: query.Get("filter_type"),
		}
		if visible := query.Get("visible"); visible != "" {
			v := visible == "true"
			filter.Visible = &v
		}
		results = s.inspector.FindByFilter(filter)
	}

	// 转换为ObjectInfo列表
	var infos []*ObjectInfo
	for _, obj := range results {
		infos = append(infos, s.inspector.GetInfo(obj))
	}

	response := map[string]interface{}{
		"count":     len(infos),
		"results":   infos,
		"timestamp": time.Now().Format("15:04:05"),
	}

	json.NewEncoder(w).Encode(response)
}

// handleClickAPI 处理点击模拟API
func (s *Server) handleClickAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != "POST" {
		http.Error(w, "仅支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Target string  `json:"target"` // 对象名称、路径或ID
		X      float64 `json:"x"`      // 坐标（可选）
		Y      float64 `json:"y"`      // 坐标（可选）
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("无效的JSON: %v", err), http.StatusBadRequest)
		return
	}

	var err error
	var targetName string

	if req.Target != "" {
		// 尝试按不同方式查找对象
		if strings.HasPrefix(req.Target, "/") {
			// 路径
			err = s.simulator.ClickByPath(s.inspector, req.Target)
			targetName = req.Target
		} else if strings.HasPrefix(req.Target, "0x") {
			// ID
			obj := s.inspector.FindByID(req.Target)
			if obj != nil {
				err = s.simulator.ClickObject(obj)
				targetName = obj.Name()
			} else {
				err = fmt.Errorf("未找到对象")
			}
		} else {
			// 名称
			err = s.simulator.ClickByName(s.inspector, req.Target)
			targetName = req.Target
		}
	} else if req.X != 0 || req.Y != 0 {
		// 坐标
		err = s.simulator.Click(req.X, req.Y)
		targetName = fmt.Sprintf("(%.0f, %.0f)", req.X, req.Y)
	} else {
		http.Error(w, "必须提供target或坐标", http.StatusBadRequest)
		return
	}

	result := SimulateClickResult{
		Success:   err == nil,
		Target:    targetName,
		Timestamp: time.Now().Format("15:04:05"),
	}

	if err != nil {
		result.Message = err.Error()
	} else {
		result.Message = "点击成功"
	}

	json.NewEncoder(w).Encode(result)
}

// handleStatsAPI 处理统计API
func (s *Server) handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	counts := s.inspector.CountObjects()
	response := map[string]interface{}{
		"counts":    counts,
		"timestamp": time.Now().Format("15:04:05"),
	}

	json.NewEncoder(w).Encode(response)
}

// handleVirtualListsAPI 处理虚拟列表专项分析API
func (s *Server) handleVirtualListsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	vlists := s.findVirtualLists()
	response := map[string]interface{}{
		"virtual_lists": vlists,
		"count":         len(vlists),
		"timestamp":     time.Now().Format("15:04:05"),
	}

	json.NewEncoder(w).Encode(response)
}

// TreeData 树形数据
type TreeData struct {
	Stats string
	HTML  string
}

// buildTreeHTML 构建树形HTML
func (s *Server) buildTreeHTML(filterType, filterName, filterVisible string, showDetails bool) TreeData {
	if s.root == nil {
		return TreeData{
			Stats: "错误：根组件为空",
			HTML:  "<div style='color:red;'>根组件未设置</div>",
		}
	}

	stats := struct {
		total       int
		visible     int
		virtual     int
		containers  int
		filtered    int
	}{}

	var html strings.Builder
	html.WriteString("<div class='tree'>")

	s.inspector.walkTree(s.root, func(obj *core.GObject) bool {
		stats.total++
		if obj.Visible() {
			stats.visible++
		}

		// 应用筛选
		if !s.shouldInclude(obj, filterType, filterName, filterVisible) {
			return true
		}
		stats.filtered++

		// 渲染节点
		html.WriteString(s.renderNode(obj, 0, showDetails))

		if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
			stats.virtual++
		}
		if _, ok := obj.Data().(*core.GComponent); ok {
			stats.containers++
		}

		return true
	})

	html.WriteString("</div>")

	filterInfo := ""
	if filterType != "" || filterName != "" || filterVisible != "" {
		filterInfo = fmt.Sprintf(", 筛选后: %d", stats.filtered)
	}

	return TreeData{
		Stats: fmt.Sprintf("📊 总计: %d 对象, %d 可见, %d 虚拟列表, %d 容器%s",
			stats.total, stats.visible, stats.virtual, stats.containers, filterInfo),
		HTML: html.String(),
	}
}

// buildTreeJSON 构建树形JSON
func (s *Server) buildTreeJSON(obj *core.GObject) map[string]interface{} {
	if obj == nil {
		return nil
	}

	data := map[string]interface{}{
		"id":       fmt.Sprintf("%p", obj),
		"name":     obj.Name(),
		"type":     GetObjectType(obj),
		"position": map[string]float64{"x": obj.X(), "y": obj.Y()},
		"size":     map[string]float64{"width": obj.Width(), "height": obj.Height()},
		"visible":  obj.Visible(),
	}

	if comp, ok := obj.Data().(*core.GComponent); ok {
		var children []map[string]interface{}
		for _, child := range comp.Children() {
			children = append(children, s.buildTreeJSON(child))
		}
		if len(children) > 0 {
			data["children"] = children
		}
	}

	return data
}

// renderNode 渲染节点
func (s *Server) renderNode(obj *core.GObject, depth int, showDetails bool) string {
	var result strings.Builder

	objType := GetObjectType(obj)
	result.WriteString(fmt.Sprintf("<div class='tree-item' style='margin-left:%dpx;'>", depth*20))
	result.WriteString(fmt.Sprintf("<span class='expand-btn'>•</span>"))
	result.WriteString(fmt.Sprintf("<span class='object-name'>%s</span> ", obj.Name()))
	result.WriteString(fmt.Sprintf("<span class='object-type'>(%s)</span>", objType))
	result.WriteString(fmt.Sprintf("<span class='object-props'>pos:%.0f,%.0f size:%.0fx%.0f",
		obj.X(), obj.Y(), obj.Width(), obj.Height()))

	if !obj.Visible() {
		result.WriteString(" [隐藏]")
	}
	result.WriteString("</span>")

	if showDetails {
		s.renderDetails(obj, &result)
	}

	result.WriteString("</div>")

	return result.String()
}

// renderDetails 渲染详细信息
func (s *Server) renderDetails(obj *core.GObject, result *strings.Builder) {
	props := s.inspector.getObjectProperties(obj)
	if len(props) > 0 {
		result.WriteString("<div class='detail-info'>")
		for key, val := range props {
			result.WriteString(fmt.Sprintf("%s: %v ", key, val))
		}
		result.WriteString("</div>")
	}
}

// shouldInclude 判断是否包含对象
func (s *Server) shouldInclude(obj *core.GObject, filterType, filterName, filterVisible string) bool {
	if filterType != "" {
		objType := GetObjectType(obj)
		if !strings.Contains(strings.ToLower(objType), strings.ToLower(filterType)) {
			return false
		}
	}

	if filterName != "" {
		if !strings.Contains(strings.ToLower(obj.Name()), strings.ToLower(filterName)) {
			return false
		}
	}

	if filterVisible != "" {
		visible := filterVisible == "true"
		if obj.Visible() != visible {
			return false
		}
	}

	return true
}

// findVirtualLists 查找所有虚拟列表
func (s *Server) findVirtualLists() []map[string]interface{} {
	var results []map[string]interface{}

	s.inspector.walkTree(s.root, func(obj *core.GObject) bool {
		if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
			info := map[string]interface{}{
				"name":          obj.Name(),
				"id":            fmt.Sprintf("%p", obj),
				"position":      map[string]float64{"x": obj.X(), "y": obj.Y()},
				"size":          map[string]float64{"width": obj.Width(), "height": obj.Height()},
				"numItems":      list.NumItems(),
				"childrenCount": list.ChildrenCount(),
			}

			if itemSize := list.VirtualItemSize(); itemSize != nil {
				info["itemSize"] = map[string]float64{"width": itemSize.X, "height": itemSize.Y}
			}

			results = append(results, info)
		}
		return true
	})

	return results
}

// 辅助函数
func getSelectedAttr(selected bool) string {
	if selected {
		return "selected"
	}
	return ""
}

func getCheckedAttr(checked bool) string {
	if checked {
		return "checked"
	}
	return ""
}

func parseFloat(s string, def float64) float64 {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return def
}

// getTreeViewHTML 返回树形视图HTML模板
func getTreeViewHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>FairyGUI 渲染树</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1400px; margin: 0 auto; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .tree-container { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .tree { font-family: 'Courier New', monospace; font-size: 12px; }
        .tree-item { margin: 2px 0; padding: 4px; }
        .tree-item:hover { background: #f8f9fa; }
        .object-name { color: #333; font-weight: bold; }
        .object-type { color: #666; font-size: 11px; }
        .object-props { color: #888; font-size: 11px; margin-left: 10px; }
        .detail-info { background: #e8f4fd; padding: 3px 6px; border-radius: 3px; margin: 2px 0; font-size: 11px; color: #2c5282; }
        .stats { background: #e8f5e8; padding: 10px; border-radius: 4px; margin: 10px 0; }
        .refresh-btn { position: fixed; top: 20px; right: 20px; padding: 10px 20px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
        .filter-panel { background: #f8f9fa; padding: 15px; border-radius: 8px; margin-bottom: 20px; }
        .filter-input { padding: 5px 8px; border: 1px solid #ccc; border-radius: 4px; margin: 5px; }
        .filter-btn { padding: 6px 12px; background: #28a745; color: white; border: none; border-radius: 4px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌳 FairyGUI 渲染树</h1>
            <button class="refresh-btn" onclick="location.reload()">🔄 刷新</button>
        </div>

        <div class="filter-panel">
            <h3>🔍 筛选选项</h3>
            <form method="get">
                <input type="text" name="type" class="filter-input" placeholder="对象类型" value="%s">
                <input type="text" name="name" class="filter-input" placeholder="对象名称" value="%s">
                <select name="visible" class="filter-input">
                    <option value="">全部</option>
                    <option value="true" %s>仅可见</option>
                    <option value="false" %s>仅隐藏</option>
                </select>
                <label>
                    <input type="checkbox" name="details" value="true" %s> 显示详细信息
                </label>
                <button type="submit" class="filter-btn">应用筛选</button>
            </form>
        </div>

        <div class="tree-container">
            <div class="stats">%s</div>
            <div class="tree">%s</div>
        </div>
    </div>

    <script>
        setTimeout(() => location.reload(), 5000);
    </script>
</body>
</html>`
}
