# FairyGUI Go 版本 V2 架构重构设计

## 设计目标

1. **简化导入路径** - 用户级 API 直接暴露在 `github.com/chslink/fairygui`
2. **Go 语言特色** - 采用接口驱动设计，实现解耦和可测试性
3. **Ebiten 深度集成** - 充分利用 Ebiten 引擎特性，而非模拟 LayaAir
4. **渐进式迁移** - 保持向后兼容，支持平滑过渡

## 核心设计原则

### 1. 接口驱动设计

**问题**：当前架构紧耦合到 LayaAir 兼容层（`laya.Sprite`），难以测试和扩展。

**解决方案**：定义小而专注的接口，在使用处而非实现处定义接口。

```go
// 在根目录定义核心接口
package fairygui

// DisplayObject 表示可显示的对象
type DisplayObject interface {
    // 位置和变换
    Position() (x, y float64)
    SetPosition(x, y float64)
    Size() (width, height float64)
    SetSize(width, height float64)

    // 显示属性
    Visible() bool
    SetVisible(visible bool)
    Alpha() float64
    SetAlpha(alpha float64)

    // 层级关系
    Parent() DisplayObject
    Children() []DisplayObject
    AddChild(child DisplayObject)
    RemoveChild(child DisplayObject)
}

// Renderer 表示渲染器接口
type Renderer interface {
    // 绘制显示对象树
    Draw(screen *ebiten.Image, root DisplayObject)

    // 绘制文本
    DrawText(screen *ebiten.Image, text string, x, y float64, style TextStyle)

    // 绘制纹理
    DrawTexture(screen *ebiten.Image, texture *ebiten.Image, dst Rect, opts DrawOptions)
}

// EventDispatcher 表示事件分发器
type EventDispatcher interface {
    On(eventType string, handler EventHandler) func()
    Off(eventType string, handler EventHandler)
    Emit(event Event)
}

// AssetLoader 表示资源加载器
type AssetLoader interface {
    LoadPackage(name string) (*Package, error)
    LoadTexture(url string) (*ebiten.Image, error)
    LoadAudio(url string) ([]byte, error)
}
```

### 2. 扁平化包结构

**当前结构问题**：
```
import "github.com/chslink/fairygui/pkg/fgui/assets"    // 太长
import "github.com/chslink/fairygui/pkg/fgui/widgets"   // 太长
import "github.com/chslink/fairygui/pkg/fgui/core"      // 太长
```

**新结构**：
```
github.com/chslink/fairygui/              # 核心 API 和类型
├── ui.go                                  # UI 对象：Object, Component, Root
├── interfaces.go                          # 核心接口定义
├── package.go                             # 包管理：Package, PackageItem
├── loader.go                              # 资源加载：FileLoader
├── factory.go                             # 组件工厂
├── event.go                               # 事件系统
├── tween.go                               # 补间动画
├── widget_*.go                            # 常用控件（Button, Image, Text, List等）
│
├── internal/                              # 内部实现（外部不可导入）
│   ├── display/                          # 显示对象实现
│   │   ├── sprite.go                     # Sprite 实现
│   │   └── graphics.go                   # Graphics 绘制命令
│   │
│   ├── render/                           # Ebiten 渲染实现
│   │   ├── renderer.go                   # Renderer 接口实现
│   │   ├── text.go                       # 文本渲染
│   │   ├── texture.go                    # 纹理渲染
│   │   └── effects.go                    # 颜色效果和滤镜
│   │
│   ├── builder/                          # 组件构建器
│   │   ├── component.go                  # 组件构建
│   │   └── parser.go                     # XML/二进制解析
│   │
│   ├── assets/                           # 资源管理内部实现
│   │   ├── package.go                    # 包解析
│   │   ├── atlas.go                      # 图集管理
│   │   └── font.go                       # 字体管理
│   │
│   ├── input/                            # 输入处理
│   │   ├── mouse.go                      # 鼠标输入
│   │   ├── touch.go                      # 触摸输入
│   │   └── keyboard.go                   # 键盘输入
│   │
│   └── animation/                        # 动画系统实现
│       ├── tween.go                      # Tween 引擎
│       ├── transition.go                 # 过渡动画
│       └── movieclip.go                  # 帧动画
│
└── advanced/                              # 高级特性（可选导入）
    ├── gears/                            # Gears 系统
    ├── relations/                        # Relations 系统
    └── effects/                          # 高级特效
```

**使用示例**：
```go
// 简单！直接从根包导入
import "github.com/chslink/fairygui"

// 创建 UI
root := fairygui.NewRoot()
loader := fairygui.NewFileLoader("./assets")
pkg, _ := loader.LoadPackage("Main")
comp := fairygui.CreateComponent(pkg, "MainWindow")
root.AddChild(comp)

// 高级功能按需导入
import "github.com/chslink/fairygui/advanced/gears"
```

### 3. 移除 LayaAir 兼容层依赖

**问题**：当前架构模拟 LayaAir，增加了不必要的抽象层。

**解决方案**：直接基于 Ebiten 设计显示对象系统。

**当前设计**（3层）：
```
GObject → laya.Sprite → Ebiten 渲染
```

**新设计**（2层）：
```
Object → Ebiten 渲染
```

**核心类型重新设计**：

```go
// Object 是所有 UI 元素的基类
type Object struct {
    // 标识
    id         string
    name       string

    // 显示属性（直接存储，不通过中间层）
    x, y       float64
    width, height float64
    scaleX, scaleY float64
    rotation   float64
    alpha      float64
    visible    bool

    // 层级
    parent     *Component

    // 渲染数据（直接对接 Ebiten）
    texture    *ebiten.Image  // 纹理（如果有）
    drawOpts   ebiten.DrawImageOptions  // 绘制选项

    // 事件
    events     eventDispatcher

    // 自定义渲染回调（可选）
    customDraw func(screen *ebiten.Image)
}

// Component 是容器对象
type Component struct {
    Object
    children   []DisplayObject
    controllers []Controller
}
```

### 4. Ebiten 特性深度集成

**利用 Ebiten 的优势**：

#### 4.1 图像缓存和批处理

```go
// 使用 Ebiten 的离屏渲染优化
type Component struct {
    Object
    children      []DisplayObject

    // 缓存渲染结果
    cachedImage   *ebiten.Image
    cacheDirty    bool

    // 是否启用批处理
    batchable     bool
}

func (c *Component) Draw(screen *ebiten.Image) {
    if c.batchable && !c.cacheDirty && c.cachedImage != nil {
        // 使用缓存，避免重复渲染
        screen.DrawImage(c.cachedImage, &c.drawOpts)
        return
    }

    // 渲染子对象
    for _, child := range c.children {
        child.Draw(screen)
    }

    // 更新缓存
    if c.batchable {
        c.updateCache()
    }
}
```

#### 4.2 顶点缓冲和自定义着色器

```go
// 为复杂效果提供 Ebiten 着色器支持
type EffectOptions struct {
    Shader    *ebiten.Shader
    Uniforms  map[string]interface{}
}

func (obj *Object) SetEffect(opts EffectOptions) {
    obj.shader = opts.Shader
    obj.uniforms = opts.Uniforms
}
```

#### 4.3 输入处理直接对接

```go
// 不通过兼容层，直接使用 Ebiten 输入
func (r *Root) Update() error {
    // 鼠标输入
    x, y := ebiten.CursorPosition()
    pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

    // 分发到对象树
    r.dispatchMouseEvent(float64(x), float64(y), pressed)

    // 键盘输入
    for key := ebiten.Key(0); key <= ebiten.KeyMax; key++ {
        if ebiten.IsKeyPressed(key) {
            r.dispatchKeyEvent(key)
        }
    }

    return nil
}
```

### 5. 接口分离原则

**小而专注的接口**：

```go
// 位置接口
type Positionable interface {
    Position() (x, y float64)
    SetPosition(x, y float64)
}

// 尺寸接口
type Sizable interface {
    Size() (width, height float64)
    SetSize(width, height float64)
}

// 可见性接口
type Visible interface {
    Visible() bool
    SetVisible(visible bool)
}

// 可交互接口
type Interactive interface {
    Touchable() bool
    SetTouchable(touchable bool)
    OnClick(handler func())
}

// 渲染接口
type Drawable interface {
    Draw(screen *ebiten.Image)
}

// 更新接口
type Updatable interface {
    Update(delta time.Duration) error
}
```

**组合接口**：

```go
// DisplayObject 组合多个小接口
type DisplayObject interface {
    Positionable
    Sizable
    Visible
    Drawable
}

// UIElement 扩展交互能力
type UIElement interface {
    DisplayObject
    Interactive
    Updatable
}
```

## API 设计示例

### 简单场景

```go
package main

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/chslink/fairygui"
)

type Game struct {
    ui *fairygui.Root
}

func (g *Game) Update() error {
    // Root 自动处理输入和更新
    return g.ui.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
    // 直接渲染
    g.ui.Draw(screen)
}

func (g *Game) Layout(w, h int) (int, int) {
    return 800, 600
}

func main() {
    // 1. 创建 Root
    ui := fairygui.NewRoot(800, 600)

    // 2. 加载资源
    loader := fairygui.NewFileLoader("./assets")
    pkg, _ := loader.LoadPackage("Main")

    // 3. 创建组件
    window := fairygui.CreateComponent(pkg, "MainWindow")
    ui.AddChild(window)

    // 4. 运行游戏
    game := &Game{ui: ui}
    ebiten.RunGame(game)
}
```

### 高级场景（使用接口解耦）

```go
// 自定义渲染器
type CustomRenderer struct{}

func (r *CustomRenderer) Draw(screen *ebiten.Image, root fairygui.DisplayObject) {
    // 自定义渲染逻辑
}

func (r *CustomRenderer) DrawText(screen *ebiten.Image, text string, x, y float64, style fairygui.TextStyle) {
    // 自定义文本渲染
}

// 使用自定义渲染器
ui := fairygui.NewRootWithRenderer(800, 600, &CustomRenderer{})
```

## 迁移策略

### 阶段 1: 接口定义（不破坏现有代码）

在根目录创建接口定义，但保持现有 `pkg/fgui` 包不变。

```go
// interfaces.go
package fairygui

type DisplayObject interface { ... }
type Renderer interface { ... }

// 提供适配器，让现有代码实现新接口
func WrapObject(obj *core.GObject) DisplayObject { ... }
```

### 阶段 2: 重构内部实现

将 `internal/compat/laya` 逐步替换为基于 Ebiten 的实现。

```go
// internal/display/sprite.go
package display

import "github.com/hajimehoshi/ebiten/v2"

type Sprite struct {
    // 直接使用 Ebiten 类型，不再模拟 LayaAir
    image *ebiten.Image
    opts  ebiten.DrawImageOptions
}
```

### 阶段 3: 暴露新 API

在根目录创建新的用户级 API。

```go
// ui.go
package fairygui

// Object 是新的 UI 对象类型
type Object struct {
    impl *display.Sprite
}

// NewObject 创建新对象
func NewObject() *Object {
    return &Object{
        impl: display.NewSprite(),
    }
}
```

### 阶段 4: 兼容性层

保持 `pkg/fgui` 作为兼容层，桥接到新 API。

```go
// pkg/fgui/api_compat.go
package fgui

import "github.com/chslink/fairygui"

// 兼容旧 API
type GObject = fairygui.Object
type GComponent = fairygui.Component

func NewGObject() *GObject {
    return fairygui.NewObject()
}
```

### 阶段 5: 文档和示例更新

提供迁移指南和新 API 文档。

## 性能优化机会

### 1. 减少中间层开销

移除 LayaAir 兼容层后，调用链更短：

**Before**: `GObject.SetPosition` → `Sprite.SetPos` → `内部状态更新`
**After**: `Object.SetPosition` → `直接更新状态`

### 2. Ebiten 批处理

```go
// 批量绘制相同纹理的对象
type BatchRenderer struct {
    vertices []ebiten.Vertex
    indices  []uint16
}

func (r *BatchRenderer) DrawBatch(screen *ebiten.Image, objects []*Object) {
    // 合并顶点数据，一次 DrawTriangles 调用
    screen.DrawTriangles(r.vertices, r.indices, texture, nil)
}
```

### 3. 对象池

```go
// 复用对象，减少 GC 压力
var objectPool = sync.Pool{
    New: func() interface{} {
        return &Object{}
    },
}

func NewObject() *Object {
    obj := objectPool.Get().(*Object)
    obj.Reset()
    return obj
}

func (obj *Object) Dispose() {
    objectPool.Put(obj)
}
```

## 测试改进

### 接口Mock便于测试

```go
// 测试时可以 mock 渲染器
type MockRenderer struct {
    DrawCalls []DrawCall
}

func (m *MockRenderer) Draw(screen *ebiten.Image, root DisplayObject) {
    m.DrawCalls = append(m.DrawCalls, DrawCall{...})
}

func TestUIRendering(t *testing.T) {
    mock := &MockRenderer{}
    ui := NewRootWithRenderer(800, 600, mock)

    // ... 测试逻辑 ...

    assert.Equal(t, 5, len(mock.DrawCalls))
}
```

## 问题与风险

### 1. 破坏性变更

**风险**：新架构与现有代码不兼容。

**缓解**：
- 保持 `pkg/fgui` 作为兼容层
- 提供迁移工具和文档
- 分阶段发布（v2.0, v2.1, ...）

### 2. 学习曲线

**风险**：用户需要学习新 API。

**缓解**：
- 详细的文档和示例
- 迁移指南
- 保持核心概念不变（Object, Component, Package）

### 3. 工作量

**风险**：重构工作量大。

**缓解**：
- 渐进式迁移
- 保持现有功能可用
- 优先迁移核心模块

## 总结

新架构设计的核心改进：

1. ✅ **简化导入** - `import "github.com/chslink/fairygui"`
2. ✅ **接口驱动** - 小而专注的接口，易于测试和扩展
3. ✅ **Go 风格** - 符合 Go 社区最佳实践
4. ✅ **Ebiten 优化** - 充分利用引擎特性，而非模拟其他引擎
5. ✅ **性能提升** - 减少中间层，支持批处理和对象池
6. ✅ **易于测试** - 接口 Mock，依赖注入
7. ✅ **向后兼容** - 保持兼容层，平滑迁移

下一步：
1. 定义核心接口
2. 实现基于 Ebiten 的 Sprite 系统
3. 创建新的 Object/Component 类型
4. 迁移渲染系统
5. 更新文档和示例
