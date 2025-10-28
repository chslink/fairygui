# FairyGUI Debug 工具包

完整的调试工具集，用于检查、分析和测试 FairyGUI 应用程序。

## 功能特性

### 1. 🔍 Inspector - 对象检查器

提供强大的对象查找和检查功能：

- **按名称查找**：支持部分匹配
- **按类型查找**：查找特定组件类型
- **按路径查找**：通过完整路径定位对象
- **按ID查找**：通过对象指针精确查找
- **复杂筛选**：支持位置、可见性等多条件筛选
- **统计功能**：统计对象数量和类型分布
- **属性获取**：获取对象的完整信息

### 2. 🖱️ EventSimulator - 事件模拟器

模拟用户交互事件，用于自动化测试：

- **点击模拟**：按对象或坐标模拟点击
- **触摸模拟**：支持多点触控
- **拖拽模拟**：模拟拖拽操作
- **自定义事件**：发送任意类型事件
- **批量操作**：支持连续事件序列

### 3. 🌐 Server - HTTP调试服务器

提供Web界面和REST API：

- **实时树形视图**：可视化查看UI结构
- **对象筛选**：按名称、类型、可见性筛选
- **RESTful API**：完整的HTTP接口
- **事件模拟接口**：通过HTTP触发事件
- **统计分析**：性能和状态统计
- **虚拟列表专项**：专门分析虚拟列表状态

## 快速开始

### 基础用法

```go
package main

import (
    "github.com/chslink/fairygui/pkg/fgui/core"
    "github.com/chslink/fairygui/pkg/fgui/debug"
    "github.com/chslink/fairygui/internal/compat/laya"
)

func main() {
    // 假设你已经有了 root 和 stage
    var root *core.GObject
    var stage *laya.Stage

    // 1. 创建Inspector
    inspector := debug.NewInspector(root)

    // 查找对象
    buttons := inspector.FindByType("GButton")
    obj := inspector.FindByPath("/Scene/Panel/Button")

    // 获取对象信息
    info := inspector.GetInfo(obj)
    println("对象:", info.Name, "类型:", info.Type)

    // 统计对象
    stats := inspector.CountObjects()
    println("总计:", stats["total"], "可见:", stats["visible"])

    // 2. 创建EventSimulator
    simulator := debug.NewEventSimulator(stage)

    // 模拟点击
    simulator.ClickByPath(inspector, "/Scene/Panel/Button")
    simulator.ClickByName(inspector, "SubmitButton")
    simulator.Click(100, 200)

    // 模拟拖拽
    simulator.DragObject(obj, 0, 0, 100, 100)

    // 3. 启动HTTP调试服务器
    server := debug.NewServer(root, stage, 8080)
    server.Start()

    println("调试服务器已启动: http://localhost:8080")
}
```

### 在Demo中集成

```go
package main

import (
    "github.com/chslink/fairygui/pkg/fgui/core"
    "github.com/chslink/fairygui/pkg/fgui/debug"
)

func main() {
    // ... 初始化 FairyGUI ...

    root := core.GRoot.Inst()

    // 启动调试服务器
    debugServer := debug.NewServer(root.GObject, stage, 8080)
    if err := debugServer.Start(); err != nil {
        log.Printf("调试服务器启动失败: %v", err)
    } else {
        log.Printf("🛠️  调试服务器: %s", debugServer.GetURL())
    }

    // ... 运行游戏循环 ...
}
```

## API 文档

### Inspector API

#### 查找方法

```go
// 按名称查找（部分匹配）
objects := inspector.FindByName("button")

// 按类型查找
lists := inspector.FindByType("GList")

// 按路径查找
panel := inspector.FindByPath("/Scene/Panel")

// 按ID查找
obj := inspector.FindByID("0x...")

// 复杂筛选
filter := debug.Filter{
    Name: "btn",
    Type: "GButton",
    Visible: &trueValue,
}
results := inspector.FindByFilter(filter)
```

#### 信息获取

```go
// 获取完整信息
info := inspector.GetInfo(obj)

// 获取路径
path := inspector.GetPath(obj)

// 获取子对象数量
count := inspector.GetChildrenCount(obj, true) // recursive

// 统计对象
stats := inspector.CountObjects()
```

### EventSimulator API

#### 点击操作

```go
// 按对象点击
simulator.ClickObject(obj)

// 按坐标点击
simulator.Click(100, 200)

// 按路径点击
simulator.ClickByPath(inspector, "/Scene/Button")

// 按名称点击
simulator.ClickByName(inspector, "SubmitButton")
```

#### 触摸和拖拽

```go
// 触摸（支持多点触控）
simulator.Touch(100, 200, touchID)
simulator.TouchObject(obj, touchID)

// 拖拽
simulator.DragObject(obj, fromX, fromY, toX, toY)

// 自定义事件
simulator.SendCustomEvent(obj, "CustomEvent", data)
```

### Server HTTP API

#### 端点列表

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/` | 首页 |
| GET | `/tree` | 树形视图（HTML） |
| GET | `/api/tree` | 获取对象树（JSON） |
| GET | `/api/object/{id}` | 获取对象信息 |
| GET | `/api/find?name=xxx` | 查找对象 |
| POST | `/api/click` | 模拟点击 |
| GET | `/api/stats` | 获取统计信息 |
| GET | `/api/virtual-lists` | 虚拟列表分析 |

#### 使用示例

**查找对象：**
```bash
curl "http://localhost:8080/api/find?name=button"
curl "http://localhost:8080/api/find?type=GList"
curl "http://localhost:8080/api/find?path=/Scene/Panel"
```

**模拟点击：**
```bash
# 按名称点击
curl -X POST http://localhost:8080/api/click \
  -H "Content-Type: application/json" \
  -d '{"target":"SubmitButton"}'

# 按路径点击
curl -X POST http://localhost:8080/api/click \
  -H "Content-Type: application/json" \
  -d '{"target":"/Scene/Panel/Button"}'

# 按坐标点击
curl -X POST http://localhost:8080/api/click \
  -H "Content-Type: application/json" \
  -d '{"x":100, "y":200}'
```

**获取统计：**
```bash
curl "http://localhost:8080/api/stats"
```

## 高级用法

### 自动化测试

```go
func TestButtonClick(t *testing.T) {
    inspector := debug.NewInspector(root)
    simulator := debug.NewEventSimulator(stage)

    // 查找按钮
    buttons := inspector.FindByType("GButton")
    if len(buttons) == 0 {
        t.Fatal("未找到按钮")
    }

    // 模拟点击
    err := simulator.ClickObject(buttons[0])
    if err != nil {
        t.Fatalf("点击失败: %v", err)
    }

    // 验证结果...
}
```

### 性能分析

```go
inspector := debug.NewInspector(root)

// 统计各类型对象数量
stats := inspector.CountObjects()
for objType, count := range stats {
    log.Printf("%s: %d", objType, count)
}

// 查找虚拟列表
vlists := inspector.FindByType("GList")
for _, list := range vlists {
    if glist, ok := list.Data().(*widgets.GList); ok && glist.IsVirtual() {
        log.Printf("虚拟列表 %s: %d项", list.Name(), glist.NumItems())
    }
}
```

### Web界面筛选

访问 `http://localhost:8080/tree` 并使用筛选面板：

- **对象类型**：输入 `GButton` 只显示按钮
- **对象名称**：输入 `btn` 显示名称包含btn的对象
- **可见性**：选择"仅可见"或"仅隐藏"
- **详细信息**：勾选显示对象的详细属性

## 注意事项

1. **性能开销**：调试服务器会有一定性能开销，建议仅在开发环境使用
2. **线程安全**：事件模拟需要在主线程或正确的协程中调用
3. **端口占用**：确保指定的端口未被占用
4. **生产环境**：不要在生产环境启用调试服务器

## 故障排除

### 调试服务器无法启动

- 检查端口是否被占用
- 确认防火墙设置
- 查看日志输出

### 找不到对象

- 确认对象路径正确（区分大小写）
- 检查对象是否已添加到场景
- 使用 `/api/tree` 查看完整树结构

### 事件模拟无效

- 确认对象可见且可交互
- 检查对象是否注册了事件监听器
- 验证坐标是否在对象范围内

## 示例项目

完整示例见 `demo/debug/server.go`

## 许可证

与 FairyGUI Ebiten 主项目相同
