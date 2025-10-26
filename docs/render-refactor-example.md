# 渲染管线重构示例

## 重构前后对比

### 示例 1: GImage 渲染

#### 当前实现 (分散,复杂)

```go
// widgets/gimage.go - Widget层直接暴露渲染数据
type GImage struct {
    *GObject
    packageItem *assets.PackageItem
    color       string
    flip        FlipType
    scaleByTile bool
    tileGridIndice int
}

func (g *GImage) PackageItem() *assets.PackageItem { return g.packageItem }
func (g *GImage) Color() string { return g.color }
func (g *GImage) Flip() FlipType { return g.flip }
func (g *GImage) ScaleSettings() (bool, int) {
    return g.scaleByTile, g.tileGridIndice
}

// render/draw_ebiten.go - 渲染层做大量判断
func renderImageWidget(target, widget, atlas, parentGeo, alpha, sprite) error {
    item := widget.PackageItem()  // ❌ 耦合到具体类型
    tint := parseColor(widget.Color())

    // 处理翻转
    if flip := widget.Flip(); flip != widgets.FlipTypeNone {
        // 30+ 行翻转逻辑
    }

    // 处理九宫格
    if grid := item.Scale9Grid; grid != nil {
        // 50+ 行九宫格逻辑
    }

    // 处理平铺
    scaleByTile, tileGrid := widget.ScaleSettings()
    if scaleByTile {
        // 40+ 行平铺逻辑
    }

    // 普通渲染
    renderImageWithGeo(...)
}
```

**问题**:
- Widget 和 Render 紧耦合
- 渲染逻辑分散在多个函数
- 每种效果都是独立代码路径
- 难以测试和维护

#### 重构后 (命令驱动,简洁)

```go
// widgets/gimage.go - Widget只生成命令
type GImage struct {
    *GObject
    packageItem *assets.PackageItem
    color       string
    flip        FlipType
    scaleByTile bool
    tileGridIndice int
}

// ✅ 核心方法:生成Graphics命令
func (g *GImage) updateDisplayObject() {
    sprite := g.DisplayObject()
    gfx := sprite.Graphics()
    gfx.Clear()

    // 一行代码描述渲染需求
    gfx.DrawTexture(&laya.TextureCommand{
        Texture:     g.packageItem,
        Mode:        g.determineMode(),
        Dest:        laya.Rect{W: g.Width(), H: g.Height()},
        Scale9Grid:  g.packageItem.Scale9Grid,
        ScaleByTile: g.scaleByTile,
        TileGridIndice: g.tileGridIndice,
        Color:       g.color,
        ScaleX:      g.flipScaleX(),
        ScaleY:      g.flipScaleY(),
        OffsetX:     g.flipOffsetX(),
        OffsetY:     g.flipOffsetY(),
    })
}

func (g *GImage) determineMode() laya.TextureCommandMode {
    if g.packageItem.Scale9Grid != nil {
        return laya.TextureModeScale9
    }
    if g.scaleByTile {
        return laya.TextureModeTile
    }
    return laya.TextureModeSimple
}

func (g *GImage) flipScaleX() float64 {
    if g.flip == FlipTypeHorizontal || g.flip == FlipTypeBoth {
        return -1
    }
    return 1
}
// ... 其他辅助方法

// render/texture_renderer.go - 统一的纹理渲染器
func (r *TextureRenderer) Render(
    target *ebiten.Image,
    cmd *laya.TextureCommand,
    geo ebiten.GeoM,
    alpha float64,
    sprite *laya.Sprite,
) error {
    // ✅ 所有纹理渲染统一处理
    switch cmd.Mode {
    case laya.TextureModeSimple:
        return r.renderSimple(target, cmd, geo, alpha, sprite)
    case laya.TextureModeScale9:
        return r.renderScale9(target, cmd, geo, alpha, sprite)
    case laya.TextureModeTile:
        return r.renderTiled(target, cmd, geo, alpha, sprite)
    }
}

// render/draw_ebiten.go - 主循环简化
func drawObject(target, obj, atlas, parentGeo, parentAlpha) error {
    sprite := obj.DisplayObject()
    gfx := sprite.Graphics()

    // ✅ 统一入口
    if gfx != nil && !gfx.IsEmpty() {
        return renderer.RenderCommands(target, gfx, combined, alpha, sprite)
    }

    // 递归处理容器
    if comp, ok := obj.Data().(*core.GComponent); ok {
        return drawComponent(target, comp, ...)
    }
}
```

**优势**:
- Widget 只关心"画什么",不关心"怎么画"
- 渲染逻辑集中在命令执行器
- 新增效果只需扩展命令,不改Widget
- 易于单元测试(命令可序列化)

---

### 示例 2: GTextField 渲染

#### 当前实现

```go
// render/draw_ebiten.go - 在主循环中特殊处理
func drawObject(...) {
    // ...
    case *widgets.GTextField:
        if textValue := data.Text(); textValue != "" {
            // ❌ 直接调用渲染函数
            drawTextImage(target, combined, data, textValue,
                          alpha, obj.Width(), obj.Height(), atlas, sprite)
        }
    case *widgets.GLabel:
        // ❌ 另一套逻辑处理图标+文本
        iconItem := data.IconItem()
        if iconItem != nil {
            drawPackageItem(...)
            // 手动计算文本偏移
            shift := ebiten.GeoM{}
            shift.Translate(float64(iconItem.Sprite.Rect.Width)+4, 0)
            textMatrix = shift.Concat(combined)
        }
        drawTextImage(target, textMatrix, ...)
}

// render/text_draw.go - 300+ 行文本渲染代码
func drawTextImage(...) error {
    // 样式提取
    style := extractTextStyle(field)

    // UBB 解析
    segments := parseUBB(text)

    // 布局计算
    layout := computeLayout(segments, bounds, style)

    // 描边/阴影
    if style.Stroke != nil { ... }
    if style.Shadow != nil { ... }

    // 绘制
    for _, line := range layout.Lines {
        for _, segment := range line.Segments {
            // 系统字体 vs 位图字体
            if isBitmapFont(segment.Font) {
                drawBitmapFont(...)
            } else {
                drawSystemFont(...)
            }
        }
    }
}
```

**问题**:
- 文本渲染逻辑无法复用
- Label/TextField/RichTextField 各自处理
- 渲染和布局混在一起

#### 重构后

```go
// widgets/gtextfield.go
func (g *GTextField) updateDisplayObject() {
    sprite := g.DisplayObject()
    gfx := sprite.Graphics()
    gfx.Clear()

    // ✅ 命令化
    gfx.DrawText(&laya.TextCommand{
        Text:   g.text,
        Style: &laya.TextStyle{
            Font:          g.font,
            Size:          g.fontSize,
            Color:         g.color,
            LetterSpacing: g.letterSpacing,
            LineSpacing:   g.lineSpacing,
            Align:         g.align,
            VAlign:        g.valign,
            Bold:          g.bold,
            Italic:        g.italic,
            Underline:     g.underline,
            Stroke:        g.stroke,
            StrokeColor:   g.strokeColor,
            Shadow:        g.shadowOffset,
            ShadowColor:   g.shadowColor,
        },
        Bounds: laya.Rect{W: g.Width(), H: g.Height()},
        AutoSize: g.autoSize,
    })
}

// widgets/glabel.go
func (g *GLabel) updateDisplayObject() {
    sprite := g.DisplayObject()
    gfx := sprite.Graphics()
    gfx.Clear()

    // ✅ 图标也是纹理命令
    if g.iconItem != nil {
        gfx.DrawTexture(&laya.TextureCommand{
            Texture: g.iconItem,
            Dest: laya.Rect{
                W: float64(g.iconItem.Width),
                H: float64(g.iconItem.Height),
            },
        })
    }

    // ✅ 文本命令自动处理偏移
    bounds := laya.Rect{W: g.Width(), H: g.Height()}
    if g.iconItem != nil {
        bounds.X = float64(g.iconItem.Width + 4)
        bounds.W -= bounds.X
    }

    gfx.DrawText(&laya.TextCommand{
        Text: g.title,
        Style: g.textStyle(),
        Bounds: bounds,
    })
}

// render/text_renderer.go - 独立的文本渲染器
type TextRenderer struct {
    cache      *TextCache
    ubbParser  *UBBParser
    layouter   *TextLayouter
}

func (r *TextRenderer) Render(
    target *ebiten.Image,
    cmd *laya.TextCommand,
    geo ebiten.GeoM,
    alpha float64,
    sprite *laya.Sprite,
) error {
    // ✅ 统一流程
    segments := r.ubbParser.Parse(cmd.Text, cmd.Style)
    layout := r.layouter.Layout(segments, cmd.Bounds, cmd.Style)

    // 缓存检查
    cacheKey := r.buildCacheKey(cmd, layout)
    if cached := r.cache.Get(cacheKey); cached != nil {
        return r.drawCached(target, cached, geo, alpha, sprite)
    }

    // 渲染到临时图像
    tmp := r.renderToImage(layout, cmd.Style)
    r.cache.Set(cacheKey, tmp)

    // AutoSize 回写
    if cmd.AutoSize != AutoSizeNone {
        r.reportSize(layout.Width, layout.Height)
    }

    return r.drawCached(target, tmp, geo, alpha, sprite)
}
```

**优势**:
- 所有文本 Widget 共享渲染逻辑
- 布局和渲染分离
- 缓存策略统一
- 易于扩展(如 Emoji 支持)

---

## 统一命令系统架构

### 命令类型定义

```go
// internal/compat/laya/graphics.go
type GraphicsCommandType int

const (
    GraphicsCommandTexture  // GImage, GLoader, GMovieClip
    GraphicsCommandText     // GTextField, GLabel
    GraphicsCommandRect     // GGraph 矩形
    GraphicsCommandEllipse  // GGraph 椭圆
    GraphicsCommandPolygon  // GGraph 多边形
    GraphicsCommandLine     // GGraph 线段
    GraphicsCommandPie      // GGraph 扇形
)

type TextureCommand struct {
    Texture        *assets.PackageItem
    Mode           TextureCommandMode  // Simple/Scale9/Tile
    Dest           Rect
    OffsetX, OffsetY float64
    Scale9Grid     *Rect
    ScaleByTile    bool
    TileGridIndice int
    Color          string
    ScaleX, ScaleY float64
}

type TextCommand struct {
    Text     string
    Style    *TextStyle
    Bounds   Rect
    AutoSize AutoSizeType
}

type TextStyle struct {
    Font          string
    Size          int
    Color         string
    LetterSpacing float64
    LineSpacing   float64
    Align         AlignType
    VAlign        VAlignType
    Bold          bool
    Italic        bool
    Underline     bool
    Stroke        float64
    StrokeColor   string
    Shadow        Point
    ShadowColor   string
}
```

### 渲染器架构

```go
// render/command_renderer.go
type CommandRenderer struct {
    atlas    *AtlasManager
    texture  *TextureRenderer
    text     *TextRenderer
    vector   *VectorRenderer
}

func (r *CommandRenderer) Render(
    target *ebiten.Image,
    sprite *laya.Sprite,
    geo ebiten.GeoM,
    alpha float64,
) error {
    gfx := sprite.Graphics()
    if gfx == nil || gfx.IsEmpty() {
        return nil
    }

    // ✅ 遍历命令,分发到专门的渲染器
    for _, cmd := range gfx.Commands() {
        var err error
        switch cmd.Type {
        case laya.GraphicsCommandTexture:
            err = r.texture.Render(target, cmd.Texture, geo, alpha, sprite)
        case laya.GraphicsCommandText:
            err = r.text.Render(target, cmd.Text, geo, alpha, sprite)
        case laya.GraphicsCommandRect:
            err = r.vector.RenderRect(target, cmd.Rect, geo, alpha, sprite)
        case laya.GraphicsCommandEllipse:
            err = r.vector.RenderEllipse(target, cmd.Ellipse, geo, alpha, sprite)
        // ... 其他命令
        }
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 主渲染循环

```go
// render/draw_ebiten.go - 大幅简化!
func DrawComponent(target *ebiten.Image, root *core.GComponent, atlas *AtlasManager) error {
    renderer := NewCommandRenderer(atlas)
    var geo ebiten.GeoM
    return drawObject(target, root.GObject, renderer, geo, 1.0)
}

func drawObject(
    target *ebiten.Image,
    obj *core.GObject,
    renderer *CommandRenderer,
    parentGeo ebiten.GeoM,
    parentAlpha float64,
) error {
    if !obj.Visible() {
        return nil
    }

    alpha := parentAlpha * obj.Alpha()
    if alpha <= 0 {
        return nil
    }

    sprite := obj.DisplayObject()
    combined := buildCombinedGeoM(sprite.LocalMatrix(), parentGeo)

    // ✅ 统一渲染入口
    if err := renderer.Render(target, sprite, combined, alpha); err != nil {
        return err
    }

    // ✅ 递归处理子对象
    if comp, ok := obj.Data().(*core.GComponent); ok {
        for _, child := range comp.Children() {
            if err := drawObject(target, child, renderer, combined, alpha); err != nil {
                return err
            }
        }
    }

    return nil
}
```

**从 150+ 行缩减到 30 行!**

---

## 迁移检查清单

### Phase 1: 基础设施
- [ ] 扩展 `laya.Graphics` 添加 `DrawTexture`/`DrawText` 方法
- [ ] 定义 `TextureCommand` 和 `TextCommand` 结构
- [ ] 实现 `CommandRenderer` 框架

### Phase 2: 纹理渲染
- [ ] 重构 `GImage.updateDisplayObject()` 使用命令
- [ ] 重构 `GLoader.updateDisplayObject()` 使用命令
- [ ] 重构 `GMovieClip.updateDisplayObject()` 使用命令
- [ ] 实现 `TextureRenderer` (复用现有九宫格/平铺逻辑)

### Phase 3: 文本渲染
- [ ] 重构 `GTextField.updateDisplayObject()` 使用命令
- [ ] 重构 `GLabel.updateDisplayObject()` 使用命令
- [ ] 重构 `GRichTextField.updateDisplayObject()` 使用命令
- [ ] 实现 `TextRenderer` (复用现有 UBB/布局逻辑)

### Phase 4: 清理
- [ ] 删除 `renderImageWidget`
- [ ] 删除 `renderLoader`
- [ ] 删除 `drawTextImage`
- [ ] 删除 `drawObject` 中的类型分发逻辑

### Phase 5: 验证
- [ ] 运行所有 Demo 场景
- [ ] 像素回归测试
- [ ] 性能基准对比

---

## 预期收益

### 代码量
| 模块 | 重构前 | 重构后 | 减少 |
|------|--------|--------|------|
| draw_ebiten.go | ~1200 行 | ~200 行 | **-83%** |
| Widget 类 | 混合逻辑 | 纯数据 | **-40%** |
| 渲染器 | 分散 | 集中 | **+30%** |
| **总计** | ~3000 行 | ~1500 行 | **-50%** |

### 复杂度
- 类型分支: **12 → 0**
- 渲染路径: **多条 → 单条**
- 缓存策略: **3 种 → 1 种**

### 可维护性
- ✅ 新增 Widget 无需改渲染层
- ✅ 渲染优化只改命令执行器
- ✅ 单元测试覆盖率 +50%
- ✅ 与 LayaAir 行为完全对齐

---

## 风险评估

### 低风险
- ✅ Graphics 命令系统已存在且稳定
- ✅ 现有渲染逻辑可直接迁移到渲染器
- ✅ 可逐个 Widget 迁移,渐进式重构

### 中风险
- ⚠️ 性能可能有变化(需基准测试)
- ⚠️ 缓存键设计需仔细考虑
- **缓解**: 保留旧代码,对比验证

### 零风险
- ✅ 不改变对外 API
- ✅ Demo 行为保持一致
- ✅ 可随时回滚

---

## 总结

### 当前问题本质
原版 LayaAir 简单是因为:**渲染引擎自动处理 Graphics 命令**

重构版复杂是因为:**手动解析命令 + 类型分发**

### 解决方案
**模拟 LayaAir 的命令驱动渲染引擎**

```
Widget 生成命令 → Graphics 记录 → Renderer 执行
(和 Laya 一样)    (已有)        (需改进)
```

### 核心原则
**Widget 层不应该知道 Ebiten 的存在**

这样就能和原版一样简洁!
