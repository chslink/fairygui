# 渲染管线统一方案

## 问题诊断

### 当前架构问题

1. **渲染路径不统一**
   - Graphics 命令系统(矢量图形) → `renderGraphicsSprite`
   - 纹理类 Widget → 独立的 `renderImageWidget/renderLoader`
   - 文本类 Widget → `drawTextImage`
   - 12 个类型分支在 `drawObject` 中处理

2. **重复的变换和颜色效果**
   - `applyColorEffects` 在多处调用
   - `parentGeo` 在每层手动传递
   - Alpha 值在多个地方累积

3. **缓存策略不一致**
   - `graphRenderCache` (GGraph)
   - `graphicsCache` (通用 Graphics)
   - 文本渲染有独立缓存
   - MovieClip 帧缓存

### 原版 LayaAir 为何简单

LayaAir 自带完整渲染引擎:
- DisplayObject 直接映射到渲染树
- Graphics 命令由引擎自动处理
- 变换矩阵自动计算
- 不需要手动类型分发

## 统一方案设计

### 核心思路

**让所有 Widget 都通过 Sprite.Graphics 命令渲染**

```
┌─────────────┐
│   Widget    │ 生成命令
└──────┬──────┘
       │ DrawTexture/DrawRect/DrawText...
       ▼
┌─────────────┐
│  Graphics   │ 记录命令
└──────┬──────┘
       │
       ▼
┌─────────────┐
│RenderEngine │ 统一执行(类型无关)
└─────────────┘
```

### 分步实现

#### 阶段 1: 纹理渲染统一

**现状**: GImage/GLoader 走独立路径
**改进**: 通过 Graphics.DrawTexture 记录

```go
// widgets/gimage.go
func (g *GImage) updateDisplayObject() {
    sprite := g.DisplayObject()
    gfx := sprite.Graphics()
    gfx.Clear()

    // 不直接渲染,而是记录命令
    gfx.DrawTexture(g.packageItem, TextureCommand{
        Mode: TextureModeScale9,
        Dest: Rect{W: g.Width(), H: g.Height()},
        Scale9Grid: g.packageItem.Scale9Grid,
        Color: g.color,
    })
}
```

**渲染层改为**:
```go
// render/draw_ebiten.go
func drawObject(...) {
    sprite := obj.DisplayObject()

    // 统一入口:检查 Graphics 命令
    if sprite.Graphics() != nil && !sprite.Graphics().IsEmpty() {
        return renderGraphicsCommands(target, sprite, parentGeo, alpha)
    }

    // 只处理复合组件
    if comp, ok := obj.Data().(*core.GComponent); ok {
        return drawComponent(target, comp, ...)
    }
}
```

#### 阶段 2: 文本渲染统一

**新增命令类型**:
```go
// laya/graphics.go
const (
    GraphicsCommandText  // 新增
)

type TextCommand struct {
    Text      string
    Style     *TextStyle
    Bounds    Rect
}

func (g *Graphics) DrawText(text string, style *TextStyle, bounds Rect) {
    // 记录文本命令
}
```

**Widget 层改为**:
```go
// widgets/gtextfield.go
func (g *GTextField) updateDisplayObject() {
    gfx := g.DisplayObject().Graphics()
    gfx.Clear()
    gfx.DrawText(g.text, &TextStyle{
        Font: g.font,
        Size: g.fontSize,
        Color: g.color,
        // ...
    }, Rect{W: g.Width(), H: g.Height()})
}
```

#### 阶段 3: 统一渲染器

```go
// render/command_renderer.go (新文件)
type CommandRenderer struct {
    atlas     *AtlasManager
    textCache *TextCache
    gfxCache  *GraphicsCache
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

    for _, cmd := range gfx.Commands() {
        switch cmd.Type {
        case GraphicsCommandTexture:
            r.renderTexture(target, cmd.Texture, geo, alpha, sprite)
        case GraphicsCommandText:
            r.renderText(target, cmd.Text, geo, alpha, sprite)
        case GraphicsCommandRect:
            r.renderRect(target, cmd.Rect, geo, alpha, sprite)
        // ... 其他命令
        }
    }
    return nil
}

// 统一应用颜色效果的地方
func (r *CommandRenderer) prepareDrawOptions(
    geo ebiten.GeoM,
    alpha float64,
    sprite *laya.Sprite,
) *ebiten.DrawImageOptions {
    opts := &ebiten.DrawImageOptions{GeoM: geo}
    opts.ColorM.Scale(1, 1, 1, alpha)
    applyColorEffects(opts, sprite)
    return opts
}
```

**主渲染循环简化为**:
```go
// render/draw_ebiten.go
func drawObject(target *ebiten.Image, obj *core.GObject, ...) error {
    if !obj.Visible() || alpha <= 0 {
        return nil
    }

    sprite := obj.DisplayObject()
    localMatrix := sprite.LocalMatrix()
    combined := buildGeoM(localMatrix).Concat(parentGeo)

    // 统一命令渲染
    renderer.Render(target, sprite, combined, parentAlpha * obj.Alpha())

    // 递归渲染子对象
    if comp, ok := obj.Data().(*core.GComponent); ok {
        for _, child := range comp.Children() {
            drawObject(target, child, atlas, combined, alpha)
        }
    }
    return nil
}
```

### 预期效果

#### 代码减少
- ❌ 删除: 12 个 case 分支
- ❌ 删除: `renderImageWidget`, `renderLoader`, `drawTextImage` 等独立函数
- ✅ 保留: 统一的 `CommandRenderer.Render`

#### 复杂度降低
```
之前: Widget类型(12种) × 渲染逻辑 = 高复杂度
之后: 命令类型(7种) × 渲染逻辑 = 低复杂度
```

#### 维护性提升
- 新增 Widget → 只需生成命令
- 修改渲染 → 只改命令执行器
- 调试简单 → 可记录/回放命令

## 迁移路线图

### Week 1: 基础设施
- [ ] 扩展 Graphics 命令类型(添加 Text/Texture)
- [ ] 实现 CommandRenderer 骨架
- [ ] 编写单元测试

### Week 2: 渲染统一
- [ ] GImage/GLoader 改为命令模式
- [ ] GTextField/GLabel 改为命令模式
- [ ] GGraph 已支持,保持不变

### Week 3: 清理与优化
- [ ] 删除旧的类型分发代码
- [ ] 统一缓存策略
- [ ] 性能基准测试

### Week 4: 验证
- [ ] 运行所有 Demo 场景
- [ ] 像素对比测试
- [ ] 文档更新

## 风险与缓解

### 风险 1: 性能下降
- **缓解**: 命令缓存 + 脏标记机制
- **验证**: 对比前后 FPS

### 风险 2: 兼容性问题
- **缓解**: 保留旧代码作为 fallback
- **验证**: 逐个 Widget 迁移

### 风险 3: 文本渲染复杂
- **缓解**: TextCommand 只记录参数,实际渲染复用现有逻辑
- **验证**: 富文本/UBB 测试

## 参考实现

### Laya 原版行为
```typescript
// TypeScript
class GImage extends GObject {
    updateDisplayObject() {
        this._displayObject.graphics.clear();
        this._displayObject.graphics.drawTexture(
            this._texture, 0, 0, this.width, this.height
        );
    }
}
```

### Go 等价实现
```go
// Go
func (g *GImage) UpdateDisplayObject() {
    sprite := g.DisplayObject()
    sprite.Graphics().Clear()
    sprite.Graphics().DrawTexture(
        g.packageItem,
        0, 0,
        g.Width(), g.Height(),
    )
}
```

**关键**: Go 版本应该和 TS 版本一样简洁!

## 总结

### 当前问题
- ❌ 渲染路径分裂(命令 vs 直接渲染)
- ❌ 类型分发过多(12个case)
- ❌ 重复的效果应用

### 统一后
- ✅ 单一渲染路径(命令驱动)
- ✅ 类型无关的渲染器
- ✅ 集中的效果管理
- ✅ 与 Laya 行为对齐

### 核心原则
**Widget 不关心如何渲染,只关心生成什么命令**
