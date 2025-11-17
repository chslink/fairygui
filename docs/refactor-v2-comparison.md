# FairyGUI Go V2 架构对比与实施指南

## 快速对比

### 导入路径对比

#### 当前（V1）
```go
import (
    "github.com/chslink/fairygui/pkg/fgui"
    "github.com/chslink/fairygui/pkg/fgui/core"
    "github.com/chslink/fairygui/pkg/fgui/assets"
    "github.com/chslink/fairygui/pkg/fgui/widgets"
    "github.com/chslink/fairygui/pkg/fgui/render"
    "github.com/chslink/fairygui/internal/compat/laya"
)
```

#### 新架构（V2）
```go
import (
    "github.com/chslink/fairygui"           // 核心 API
    "github.com/chslink/fairygui/advanced"  // 高级功能（可选）
)
```

### 使用对比

#### 当前（V1）- 创建简单 UI

```go
// 1. 初始化多个子系统
atlas := render.NewAtlasManager(loader)
factory := fgui.NewFactoryWithLoader(atlas, loader)

// 2. 创建 Root
root := core.Root()
stage := fgui.NewStage(800, 600)
root.AttachStage(stage)

// 3. 加载包
loader := assets.NewFileLoader("./assets")
data, _ := os.ReadFile("./assets/Main.fui")
pkg, _ := assets.ParsePackage(data, "Main")

// 4. 构建组件
comp, _ := factory.BuildComponent(ctx, pkg, item)
root.AddChild(comp.GObject)

// 5. 在游戏循环中
func (g *Game) Update() error {
    mouse := laya.MouseState{...}
    root.Advance(delta, mouse)
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    render.DrawComponent(screen, root, ...)
}
```

#### 新架构（V2）- 创建简单 UI

```go
// 1. 创建 Root（内置一切）
ui := fairygui.NewRoot(800, 600)

// 2. 加载包（一行搞定）
loader := fairygui.NewFileLoader("./assets")
pkg, _ := loader.LoadPackage("Main")

// 3. 创建组件（直接）
window := fairygui.CreateComponent(pkg, "MainWindow")
ui.AddChild(window)

// 4. 在游戏循环中（自动处理一切）
func (g *Game) Update() error {
    return g.ui.Update()  // 自动处理输入、动画等
}

func (g *Game) Draw(screen *ebiten.Image) {
    g.ui.Draw(screen)  // 自动渲染整个树
}
```

**代码量减少：~60%**

## 详细架构对比

### 1. 显示对象系统

#### 当前（V1）- 三层架构

```
用户代码
    ↓
GObject (pkg/fgui/core)
    ↓
laya.Sprite (internal/compat/laya) ← 兼容层，增加复杂度
    ↓
Graphics Commands
    ↓
Ebiten 渲染 (pkg/fgui/render)
```

**问题**：
- 3 层抽象，性能损失
- laya.Sprite 模拟 LayaAir，但不符合 Go 习惯
- 紧耦合，难以替换渲染实现

#### 新架构（V2）- 接口驱动

```
用户代码
    ↓
Object (实现 DisplayObject 接口)
    ↓
Ebiten 渲染 (内部实现)
```

**改进**：
- 2 层抽象，性能提升 ~30%
- 接口驱动，易于测试和扩展
- 直接对接 Ebiten，充分利用引擎特性

### 2. 接口设计

#### 当前（V1）- 无接口抽象

```go
// 直接使用具体类型，难以解耦
type GObject struct {
    display *laya.Sprite  // 紧耦合
}

// 渲染函数也是具体的
func DrawComponent(screen *ebiten.Image, obj *core.GObject) {
    // 直接访问内部字段
    sprite := obj.DisplayObject()
    // ...
}
```

**问题**：
- 无法 Mock，测试困难
- 无法替换实现
- 违反依赖倒置原则

#### 新架构（V2）- 小而专注的接口

```go
// 接口在使用处定义
type DisplayObject interface {
    Positionable
    Sizable
    Visible
    Drawable
}

// 小接口组合
type Positionable interface {
    Position() (x, y float64)
    SetPosition(x, y float64)
}

// 渲染器接口
type Renderer interface {
    Draw(screen *ebiten.Image, root DisplayObject)
}

// 使用接口，易于测试
func TestRendering(t *testing.T) {
    mock := &MockRenderer{}
    ui := NewRootWithRenderer(800, 600, mock)
    // 验证渲染调用
}
```

**改进**：
- 符合接口隔离原则
- 易于 Mock 和测试
- 支持依赖注入

### 3. 包结构

#### 当前（V1）

```
pkg/fgui/
  ├── api.go         ← 统一入口，但仍需导入子包
  ├── core/          ← 核心类型
  ├── assets/        ← 资源管理
  ├── widgets/       ← 控件
  ├── render/        ← 渲染
  ├── gears/         ← Gears 系统
  ├── tween/         ← 动画
  └── utils/         ← 工具（反模式）

internal/compat/laya/  ← 兼容层（不应存在）
```

**问题**：
- 过度嵌套，导入路径长
- `utils` 包是反模式
- `internal/compat` 增加维护负担

#### 新架构（V2）

```
github.com/chslink/fairygui/  ← 所有公开 API
  ├── ui.go              # Object, Component, Root
  ├── interfaces.go      # 核心接口
  ├── package.go         # Package, PackageItem
  ├── loader.go          # FileLoader
  ├── factory.go         # Factory
  ├── event.go           # 事件系统
  ├── tween.go           # 补间动画
  ├── widget_button.go   # Button 控件
  ├── widget_image.go    # Image 控件
  ├── widget_text.go     # Text 控件
  └── widget_list.go     # List 控件

internal/               ← 实现细节（外部不可见）
  ├── display/          # 显示对象实现
  ├── render/           # Ebiten 渲染
  ├── builder/          # 组件构建
  ├── assets/           # 资源管理
  ├── input/            # 输入处理
  └── animation/        # 动画系统

advanced/               ← 高级功能（可选）
  ├── gears/            # Gears 系统
  ├── relations/        # Relations 系统
  └── effects/          # 特效
```

**改进**：
- 扁平化，导入简单
- 核心功能直接暴露
- 高级功能分离
- 无 `utils` 反模式

### 4. 事件系统

#### 当前（V1）- 模拟 LayaAir

```go
// 基于 LayaAir 的事件系统
sprite := laya.NewSprite()
sprite.On("click", func(evt *laya.Event) {
    // 处理点击
})

// 需要手动冒泡
evt := laya.NewEvent("click")
sprite.DispatchEvent(evt)
```

**问题**：
- 字符串类型事件，易出错
- 手动管理冒泡
- 不符合 Go 习惯

#### 新架构（V2）- Go 风格

```go
// 类型安全的事件
button := fairygui.NewButton()
button.OnClick(func(e *fairygui.ClickEvent) {
    log.Println("clicked!")
})

// 自动冒泡和捕获
button.OnEvent(fairygui.EventClick, func(e Event) {
    // 通用事件处理
})
```

**改进**：
- 类型安全
- 自动冒泡
- 符合 Go 习惯（回调函数）

### 5. 资源加载

#### 当前（V1）

```go
// 步骤繁琐
loader := assets.NewFileLoader("./assets")
data, _ := loader.Load("Main.fui")
pkg, _ := assets.ParsePackage(data, "Main")

atlas := render.NewAtlasManager(loader)
factory := builder.NewFactoryWithLoader(atlas, loader)
factory.RegisterPackage(pkg)

comp, _ := factory.BuildComponent(ctx, pkg, item)
```

**问题**：
- 多个步骤
- 需要手动管理 AtlasManager
- 容易出错

#### 新架构（V2）

```go
// 一行搞定
loader := fairygui.NewFileLoader("./assets")
pkg, _ := loader.LoadPackage("Main")

// 直接创建组件
window := fairygui.CreateComponent(pkg, "MainWindow")

// 或者使用 URL 方式（更简单）
window := fairygui.CreateObjectFromURL("ui://Main/MainWindow")
```

**改进**：
- 简化流程
- 自动管理依赖
- 支持 URL 方式

### 6. 渲染系统

#### 当前（V1）- 命令模式

```go
// GObject → laya.Sprite → Graphics Commands
sprite.Graphics().DrawRect(0, 0, 100, 50, color)

// 渲染时消费命令
func DrawComponent(screen *ebiten.Image, obj *core.GObject) {
    sprite := obj.DisplayObject()
    graphics := sprite.Graphics()
    for _, cmd := range graphics.Commands() {
        // 执行命令
    }
}
```

**问题**：
- 间接层过多
- 命令累积，内存占用高
- 不易批处理

#### 新架构（V2）- 直接渲染

```go
// 实现 Drawable 接口
func (obj *Object) Draw(screen *ebiten.Image) {
    if obj.texture != nil {
        screen.DrawImage(obj.texture, &obj.drawOpts)
    }

    // 自定义渲染
    if obj.customDraw != nil {
        obj.customDraw(screen)
    }
}

// 支持批处理
func (r *Renderer) DrawBatch(screen *ebiten.Image, objects []DisplayObject) {
    // 合并相同纹理，一次绘制
}
```

**改进**：
- 减少中间层
- 支持批处理
- 内存占用低

## 性能对比（预期）

| 指标 | V1 | V2 | 改进 |
|------|----|----|------|
| 对象创建开销 | 3 次分配（GObject + Sprite + Graphics） | 1 次分配（Object） | **-67%** |
| 方法调用层级 | 3 层（GObject → Sprite → 状态） | 1 层（Object → 状态） | **-67%** |
| 渲染调用次数 | 每个对象多次（命令模式） | 每个对象一次 | **-50%** |
| 内存占用 | 高（命令累积） | 低（直接渲染） | **-40%** |
| 接口虚表开销 | 无（具体类型） | 有（接口调用） | **+5%** |
| **总体性能** | 基线 | **提升 ~40%** | ✅ |

## 测试对比

### 当前（V1）- 难以测试

```go
func TestButton(t *testing.T) {
    // 需要完整的环境
    root := core.Root()
    stage := laya.NewStage(800, 600)
    root.AttachStage(stage)

    // 需要真实的 Atlas
    atlas := render.NewAtlasManager(loader)

    // 难以 Mock
    button := widgets.NewButton()
    // ...
}
```

**问题**：
- 需要完整环境
- 难以隔离测试
- 无法 Mock 依赖

### 新架构（V2）- 易于测试

```go
func TestButton(t *testing.T) {
    // Mock 渲染器
    mock := &MockRenderer{}

    // 创建按钮（不需要完整环境）
    button := fairygui.NewButton()
    button.SetSize(100, 40)

    // 测试点击
    button.OnClick(func(e *fairygui.ClickEvent) {
        called = true
    })
    button.SimulateClick()

    assert.True(t, called)
}

// Mock 渲染器
type MockRenderer struct {
    DrawCalls []DrawCall
}

func (m *MockRenderer) Draw(screen *ebiten.Image, root fairygui.DisplayObject) {
    m.DrawCalls = append(m.DrawCalls, DrawCall{Root: root})
}
```

**改进**：
- 不需要完整环境
- 易于 Mock
- 快速测试

## 迁移路径

### 阶段 1: 准备（1-2 周）

1. **创建新包结构**
   ```bash
   mkdir -p internal/display internal/render internal/builder
   ```

2. **定义核心接口**
   ```go
   // interfaces.go
   package fairygui

   type DisplayObject interface { ... }
   type Renderer interface { ... }
   ```

3. **编写适配器**
   ```go
   // 让现有代码兼容新接口
   func WrapGObject(obj *core.GObject) DisplayObject { ... }
   ```

### 阶段 2: 重写核心（2-3 周）

1. **实现新的 Object**
   ```go
   // ui.go
   package fairygui

   type Object struct {
       // 基于 Ebiten，不依赖 LayaAir
   }
   ```

2. **实现新的 Renderer**
   ```go
   // internal/render/renderer.go
   package render

   type EbitenRenderer struct { ... }
   ```

3. **重写事件系统**
   ```go
   // event.go
   package fairygui

   type EventDispatcher struct { ... }
   ```

### 阶段 3: 迁移功能（3-4 周）

1. **迁移控件**
   - Button
   - Image
   - Text
   - List
   - ...

2. **迁移高级功能**
   - Gears
   - Relations
   - Transitions

### 阶段 4: 兼容层（1 周）

```go
// pkg/fgui/compat.go
package fgui

import "github.com/chslink/fairygui"

// 保持向后兼容
type GObject = fairygui.Object
type GComponent = fairygui.Component
```

### 阶段 5: 测试与文档（1-2 周）

1. 编写测试
2. 更新文档
3. 创建示例
4. 性能基准测试

**总计：8-12 周**

## 风险评估

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 破坏现有代码 | 高 | 中 | 保持兼容层，提供迁移工具 |
| 性能不如预期 | 中 | 低 | 提前性能测试，保留优化空间 |
| 接口设计不合理 | 高 | 中 | 先定义接口，评审后实现 |
| 工作量超预期 | 中 | 高 | 分阶段发布，持续迭代 |
| 学习曲线陡峭 | 低 | 中 | 详细文档，多示例 |

## 成功标准

1. ✅ 导入路径简化到 `import "github.com/chslink/fairygui"`
2. ✅ 核心功能可通过接口测试（Mock 覆盖率 >80%）
3. ✅ 性能提升 >30%（对象创建、渲染）
4. ✅ 代码量减少 >40%（用户侧）
5. ✅ 测试覆盖率 >85%
6. ✅ 100% 向后兼容（通过兼容层）
7. ✅ 文档完整（API 文档、示例、迁移指南）

## 下一步行动

1. **立即**：
   - [ ] 评审此设计文档
   - [ ] 确定迁移优先级
   - [ ] 创建 GitHub Project 跟踪进度

2. **本周**：
   - [ ] 定义核心接口（interfaces.go）
   - [ ] 设计新的 Object/Component 结构
   - [ ] 编写接口规范测试

3. **下周**：
   - [ ] 实现基础 Object 类型
   - [ ] 实现 EbitenRenderer
   - [ ] 迁移第一个控件（Button）

4. **本月**：
   - [ ] 完成核心功能迁移
   - [ ] 建立兼容层
   - [ ] 发布 v2.0-alpha

## 参考资源

- [Go 接口设计最佳实践](https://go.dev/doc/effective_go#interfaces)
- [Ebiten 官方文档](https://ebitengine.org/)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [ebitenui 项目结构参考](https://github.com/ebitenui/ebitenui)
