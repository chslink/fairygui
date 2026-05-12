# Phase 3: Go 惯用 API 层

> 阶段目标: 提供符合 Go 语言习惯的易用 API
> 预计工时: 3-4 天
> 前置依赖: Phase 1, Phase 2

---

## 3.0 问题分析

当前 API 的问题：
1. **类型转换不优雅**: `AsButton()` 返回 `interface{}`，需要手动断言
2. **混合中英文 API**: 一些方法是中文名（`EnsureSizeCorrect`），一些是英文
3. **构造模式不统一**: New + SetData 模式令人困惑
4. **缺少 Options 模式**: 大量 setter 调用冗长
5. **错误处理不统一**: 有些返回 error，有些静默失败
6. **缺少链式调用支持**: 常见于 UI 构建场景

---

## 3.1 API 层级设计

```
pkg/fgui/          ← 公开 API 门面 (已有 api.go)
├── api.go         ← 当前主入口，需要重命名类型别名
├── fgui.go        ← 新增: 顶层便捷函数
└── options.go     ← 新增: Options 模式配置
```

### 3.1.1 类型别名重命名

当前 `api.go` 的问题是类型别名让用户困惑。改为：

```go
// api.go - 清晰导出所有类型
package fgui

// Core types (从 pkg/fgui/core 重新导出)
type GObject = core.GObject
type GComponent = core.GComponent
type GRoot = core.GRoot
type Controller = core.Controller
// ... 等等

// 同时提供 Go 惯用别名
type Object = GObject        // 简短别名
type Component = GComponent  // 简短别名
type Root = GRoot            // 简短别名
```

### 3.1.2 便捷创建函数 (fgui.go)

```go
package fgui

// CreateComponent 创建普通组件
func CreateComponent(opts ...ComponentOption) *GComponent

// CreateButton 创建按钮
func CreateButton(opts ...ButtonOption) *GButton

// CreateLabel 创建标签
func CreateLabel(text string, opts ...LabelOption) *GLabel

// CreateList 创建列表
func CreateList(opts ...ListOption) *GList

// CreateWindow 创建窗口
func CreateWindow(opts ...WindowOption) *Window

// CreatePopupMenu 创建弹出菜单
func CreatePopupMenu(items ...MenuItem) *PopupMenu
```

---

## 3.2 Options 模式

### ComponentOption
```go
type componentConfig struct {
    id       string
    name     string
    x, y     float64
    width    float64
    height   float64
    scaleX   float64
    scaleY   float64
    alpha    float64
    rotation float64
    visible  bool
    touchable bool
    pivotX   float64
    pivotY   float64
}

type ComponentOption func(*componentConfig)

func WithID(id string) ComponentOption        { return func(c *componentConfig) { c.id = id } }
func WithPosition(x, y float64) ComponentOption { return func(c *componentConfig) { c.x, c.y = x, y } }
func WithSize(w, h float64) ComponentOption   { return func(c *componentConfig) { c.width, c.height = w, h } }
func WithScale(sx, sy float64) ComponentOption { return func(c *componentConfig) { c.scaleX, c.scaleY = sx, sy } }
func WithAlpha(a float64) ComponentOption     { return func(c *componentConfig) { c.alpha = a } }
func WithRotation(r float64) ComponentOption  { return func(c *componentConfig) { c.rotation = r } }
func WithPivot(px, py float64) ComponentOption { return func(c *componentConfig) { c.pivotX, c.pivotY = px, py } }
func Hidden() ComponentOption                 { return func(c *componentConfig) { c.visible = false } }
func Disabled() ComponentOption               { return func(c *componentConfig) { c.touchable = false } }
```

### 使用对比

**之前**:
```go
btn := widgets.NewButton()
btn.SetName("myButton")
btn.SetPosition(100, 200)
btn.SetSize(120, 40)
btn.SetAlpha(0.8)
btn.SetVisible(true)
```

**之后**:
```go
btn := fgui.CreateButton(
    fgui.WithPosition(100, 200),
    fgui.WithSize(120, 40),
    fgui.WithAlpha(0.8),
)
```

---

## 3.3 Builder 模式（链式调用）

```go
// 按钮构建器
type ButtonBuilder struct {
    btn *GButton
}

func NewButtonBuilder() *ButtonBuilder {
    return &ButtonBuilder{btn: NewButton()}
}

func (b *ButtonBuilder) Title(text string) *ButtonBuilder {
    b.btn.SetTitle(text)
    return b
}

func (b *ButtonBuilder) Icon(url string) *ButtonBuilder {
    b.btn.SetIcon(url)
    return b
}

func (b *ButtonBuilder) Position(x, y float64) *ButtonBuilder {
    b.btn.SetPosition(x, y)
    return b
}

func (b *ButtonBuilder) Size(w, h float64) *ButtonBuilder {
    b.btn.SetSize(w, h)
    return b
}

func (b *ButtonBuilder) OnClick(fn func()) *ButtonBuilder {
    b.btn.OnClick(fn)
    return b
}

func (b *ButtonBuilder) Selected(sel bool) *ButtonBuilder {
    b.btn.SetSelected(sel)
    return b
}

func (b *ButtonBuilder) Build() *GButton {
    return b.btn
}
```

**使用**:
```go
btn := fgui.NewButtonBuilder().
    Title("确认").
    Position(100, 200).
    Size(120, 40).
    OnClick(func() { fmt.Println("clicked") }).
    Build()
```

---

## 3.4 事件系统简化

```go
// 新事件 API - 返回取消函数
type CancelFunc func()

// GObject 新增方法
func (g *GObject) ListenClick(fn func()) CancelFunc {
    id := g.OnClick(fn)
    return func() { g.OffClick(id) }
}

func (g *GObject) ListenDrop(fn func(data any)) CancelFunc {
    g.On(laya.EventDrop, func(evt *laya.Event) { fn(evt.Data) })
    return func() { /* 对应的 Off */ }
}

// 使用
cancel := btn.ListenClick(func() { fmt.Println("clicked") })
defer cancel()
```

---

## 3.5 错误处理规范化

```go
// Option 校验
func CreateButton(opts ...ButtonOption) (*GButton, error) {
    cfg := defaultButtonConfig()
    for _, opt := range opts {
        opt(cfg)
    }
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("create button: %w", err)
    }
    return cfg.Build(), nil
}
```

---

## 3.6 Demo 使用新 API 重写

选择 3-5 个典型场景用新 API 重写，展示最佳实践：

| 场景 | 重写重点 |
|------|---------|
| MainMenu | Options 模式创建按钮 |
| BasicsDemo | Builder 模式构建复杂 UI |
| BagDemo | 链式调用 + Option |
| ChatDemo | 事件系统简化 |
| CooldownDemo | 简洁的 Tween API |

---

## 3.7 context.Context 集成

为异步操作提供取消支持：

```go
// UIPackage 加载支持 context
func LoadPackage(ctx context.Context, loader Loader, path string) (*Package, error)

// Window 显示支持 context（用于取消模态等待等）
func (w *Window) ShowWithContext(ctx context.Context)

// Transition 播放支持 context
func (t *Transition) PlayWithContext(ctx context.Context, times int, delay float64)
```

---

## 3.8 完成标准
- [ ] Options 模式覆盖所有主要组件
- [ ] Builder 模式覆盖 Button/List/Label/Window
- [ ] 事件系统提供 CancelFunc 模式
- [ ] 3+ 个 Demo 场景使用新 API 重写
- [ ] context.Context 集成
- [ ] 文档: `docs/api-guide.md` 使用指南
- [ ] 文档: `docs/migration-from-ts.md` TS 迁移指南
