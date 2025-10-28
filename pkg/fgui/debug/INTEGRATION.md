# Debug工具集成指南

本文档说明如何将debug工具集成到你的FairyGUI应用中。

## 概述

Debug工具包已经从demo中提取并优化为框架级功能，位于 `pkg/fgui/debug`。

### 主要改进

与demo中的原始版本相比，新版本提供了：

1. ✅ **模块化设计**：分离Inspector、EventSimulator和Server
2. ✅ **事件模拟**：支持点击、触摸、拖拽等操作
3. ✅ **强大的查找**：支持名称、类型、路径、筛选器等多种方式
4. ✅ **RESTful API**：完整的HTTP接口
5. ✅ **统计功能**：对象计数、类型分布等
6. ✅ **容器支持**：显示子对象数量
7. ✅ **参数筛选**：按类型、名称、可见性筛选
8. ✅ **性能优化**：减少重复遍历，优化数据结构

## 快速集成

### 步骤1：导入包

```go
import (
    "github.com/chslink/fairygui/pkg/fgui/debug"
)
```

### 步骤2：启动调试服务器

在你的`main.go`中添加：

```go
func main() {
    // ... 初始化 FairyGUI ...

    root := core.GRoot.Inst()

    // 启动调试服务器（仅在开发环境）
    if isDevelopment() {
        debugServer := debug.NewServer(root.GObject, stage, 8080)
        if err := debugServer.Start(); err != nil {
            log.Printf("❌ 调试服务器启动失败: %v", err)
        } else {
            log.Printf("🛠️  调试服务器: %s", debugServer.GetURL())
        }
    }

    // ... 运行游戏循环 ...
}

func isDevelopment() bool {
    // 根据你的需求判断是否为开发环境
    return os.Getenv("ENV") != "production"
}
```

### 步骤3：访问调试界面

启动应用后，在浏览器中访问：

```
http://localhost:8080
```

## 详细功能说明

### 1. 对象检查（Inspector）

```go
inspector := debug.NewInspector(root)

// 按名称查找
buttons := inspector.FindByName("button")

// 按类型查找
lists := inspector.FindByType("GList")

// 按路径查找
panel := inspector.FindByPath("/Scene/Panel")

// 获取对象信息
info := inspector.GetInfo(obj)
fmt.Printf("对象: %s, 类型: %s, 子对象: %d\n",
    info.Name, info.Type, info.Children)

// 统计对象
stats := inspector.CountObjects()
fmt.Printf("总计: %d, 可见: %d, 容器: %d\n",
    stats["total"], stats["visible"], stats["containers"])
```

### 2. 事件模拟（EventSimulator）

```go
simulator := debug.NewEventSimulator(stage)

// 模拟点击
simulator.ClickByPath(inspector, "/Scene/Button")
simulator.ClickByName(inspector, "SubmitButton")
simulator.Click(100, 200)

// 模拟拖拽
simulator.DragObject(obj, 0, 0, 100, 100)

// 自定义事件
simulator.SendCustomEvent(obj, "CustomEvent", data)
```

### 3. HTTP API

所有API端点都支持CORS，可以从任何客户端调用：

**查找对象：**
```bash
curl "http://localhost:8080/api/find?name=button"
curl "http://localhost:8080/api/find?type=GList&visible=true"
```

**模拟点击：**
```bash
curl -X POST http://localhost:8080/api/click \
  -H "Content-Type: application/json" \
  -d '{"target":"SubmitButton"}'
```

**获取统计：**
```bash
curl "http://localhost:8080/api/stats"
```

**虚拟列表分析：**
```bash
curl "http://localhost:8080/api/virtual-lists"
```

## 与原demo版本的对比

| 功能 | 原demo版本 | 新框架版本 |
|------|-----------|-----------|
| **模块化** | 单文件 | 3个独立模块 |
| **对象查找** | 仅遍历 | 多种查找方式 |
| **事件模拟** | ❌ 无 | ✅ 完整支持 |
| **筛选功能** | 基础 | 高级筛选器 |
| **统计功能** | 基础 | 详细统计 |
| **子对象计数** | ❌ 无 | ✅ 支持 |
| **API完整性** | 部分 | 完整RESTful |
| **性能** | 一般 | 优化 |

## 使用场景

### 开发调试

- 查看UI结构
- 检查对象属性
- 定位布局问题
- 分析虚拟列表状态

### 自动化测试

```go
func TestUI(t *testing.T) {
    inspector := debug.NewInspector(root)
    simulator := debug.NewEventSimulator(stage)

    // 查找并点击按钮
    buttons := inspector.FindByType("GButton")
    for _, btn := range buttons {
        simulator.ClickObject(btn)
        // 验证结果...
    }
}
```

### 性能分析

```go
inspector := debug.NewInspector(root)
stats := inspector.CountObjects()

log.Printf("对象统计:")
for objType, count := range stats {
    log.Printf("  %s: %d", objType, count)
}

// 检查虚拟列表
vlists := inspector.FindByType("GList")
for _, list := range vlists {
    info := inspector.GetInfo(list)
    if props, ok := info.Properties["virtual"].(bool); ok && props {
        log.Printf("虚拟列表 %s: %v项",
            info.Name, info.Properties["numItems"])
    }
}
```

## 配置选项

### 端口配置

```go
// 默认端口8080
server := debug.NewServer(root, stage, 8080)

// 自定义端口
server := debug.NewServer(root, stage, 9000)
```

### 性能考虑

调试服务器会有一定性能开销：

- **轻量级**: Inspector和EventSimulator几乎无开销
- **HTTP服务器**: 每次请求会遍历对象树
- **自动刷新**: 网页每5秒自动刷新

建议：
- ✅ 开发环境：启用所有功能
- ⚠️ 测试环境：仅启用需要的功能
- ❌ 生产环境：完全禁用

## 迁移指南

如果你正在使用demo中的debug功能，迁移步骤：

### 1. 替换导入

```go
// 旧代码
import "your-project/demo/debug"

// 新代码
import "github.com/chslink/fairygui/pkg/fgui/debug"
```

### 2. 更新API调用

大部分API保持兼容，但有一些改进：

```go
// 旧代码
server := debug.NewServer(root.GComponent, 8080)

// 新代码（需要stage参数）
server := debug.NewServer(root.GObject, stage, 8080)
```

### 3. 利用新功能

```go
// 新增：对象查找
inspector := debug.NewInspector(root)
objs := inspector.FindByName("button")

// 新增：事件模拟
simulator := debug.NewEventSimulator(stage)
simulator.ClickByPath(inspector, "/Scene/Button")

// 新增：高级筛选
filter := debug.Filter{
    Type: "GButton",
    Visible: &trueValue,
}
results := inspector.FindByFilter(filter)
```

## 故障排除

### 问题：调试服务器无法启动

**可能原因**：
- 端口被占用
- 权限不足
- Stage未初始化

**解决方案**：
```go
// 检查错误
if err := server.Start(); err != nil {
    log.Printf("启动失败: %v", err)
    // 尝试其他端口
}
```

### 问题：找不到对象

**可能原因**：
- 对象未添加到场景
- 路径不正确
- 对象名称区分大小写

**解决方案**：
```go
// 1. 检查对象树
curl "http://localhost:8080/api/tree"

// 2. 使用模糊查找
objs := inspector.FindByName("button") // 部分匹配

// 3. 使用类型查找
objs := inspector.FindByType("GButton")
```

### 问题：事件模拟无效

**可能原因**：
- 对象不可见
- 对象未注册事件监听器
- Stage未正确设置

**解决方案**：
```go
// 检查对象状态
info := inspector.GetInfo(obj)
if !info.Visible {
    log.Println("对象不可见")
}

// 确保对象在正确位置
pt := laya.Point{X: 100, Y: 200}
target := stage.HitTest(pt)
if target == nil {
    log.Println("坐标处无对象")
}
```

## 扩展开发

如果需要添加自定义功能：

### 1. 自定义属性提取

```go
// 在inspector.go中扩展getObjectProperties
func (i *Inspector) getObjectProperties(obj *core.GObject) map[string]interface{} {
    props := make(map[string]interface{})

    // 添加你的自定义属性
    switch widget := obj.Data().(type) {
    case *YourCustomWidget:
        props["customProp"] = widget.CustomProp()
    }

    return props
}
```

### 2. 自定义API端点

```go
// 在server.go中添加
http.HandleFunc("/api/custom", s.handleCustomAPI)

func (s *Server) handleCustomAPI(w http.ResponseWriter, r *http.Request) {
    // 你的逻辑
}
```

## 最佳实践

1. **仅在开发环境启用**：使用环境变量控制
2. **合理选择端口**：避免与其他服务冲突
3. **定期查看统计**：了解对象数量和性能
4. **使用筛选功能**：快速定位问题对象
5. **自动化测试集成**：利用EventSimulator进行UI测试

## 下一步

- 查看 [README.md](README.md) 了解完整API文档
- 查看 `example_test.go` 了解更多使用示例
- 参考原demo中的集成方式：`demo/debug/server.go`

## 反馈和贡献

如有问题或建议，欢迎提Issue或PR。
