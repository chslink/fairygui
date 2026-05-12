# Phase 4: 性能优化

> 阶段目标: 渲染帧率 60fps，万级列表可滚动，纹理复用
> 预计工时: 3-4 天
> 前置依赖: Phase 2, Phase 3

---

## 4.0 当前性能瓶颈分析

### 已知问题

| 问题 | 位置 | 影响 |
|------|------|------|
| 每帧遍历全部子对象渲染 | `render/draw_ebiten.go` | 大量不可见对象仍参与渲染计算 |
| 图形渲染无缓存 | `render/draw_ebiten.go` | GGraph 每帧重建 Ebiten Image |
| 纹理无 LRU 驱逐 | `render/atlas_ebiten.go` | 内存持续增长 |
| 虚拟列表 MeasureItem 重复计算 | `widgets/list_virtual_impl.go` | 滚动时卡顿 |
| 颜色矩阵每帧重建 | `render/color_effects.go` | 不必要的对象分配 |
| 文本贴图无有效缓存 | `render/text_draw.go` | 同文本重复渲染 |
| drawComponent 递归调用栈深 | `render/draw_ebiten.go` | 深层嵌套组件栈溢出风险 |
| 事件系统每次 Emit 都复制 listeners | `internal/compat/laya/event.go` | 高频事件产生大量分配 |

---

## 4.1 渲染管线优化

### 4.1.1 可见性剔除

```go
// draw_ebiten.go
func drawComponent(target *ebiten.Image, comp *GComponent, atlas *AtlasManager, 
                   parentGeo ebiten.GeoM, parentAlpha float64, clipRect *laya.Rect) error {
    
    for _, child := range comp.Children() {
        // 1. 快速跳过不可见对象
        if !child.Visible() {
            continue
        }
        
        // 2. alpha=0 跳过
        if child.Alpha() <= 0 {
            continue
        }
        
        // 3. 视口裁剪（计算子对象边界，检查是否在裁剪区域内）
        if clipRect != nil {
            if !intersects(childBounds(child), clipRect) {
                continue
            }
        }
        
        // 4. 渲染
        drawObject(target, child, atlas, geo, alpha)
    }
}

// 预计算子对象世界空间边界
func childBounds(obj *GObject) laya.Rect {
    // 缓存边界，仅在 transform 变化时重新计算
}
```

### 4.1.2 GGraph 渲染缓存优化

```go
// 当前: 每帧读取 Graphics 命令 → 生成 Ebiten Image → 绘制
// 优化: 缓存 Ebiten Image，仅在 Graphics Dirty 时重建

type GraphRenderCache struct {
    image    *ebiten.Image
    hash     string  // Graphics 命令的 hash
    lastSize Point
}

var graphCaches sync.Map  // key: objectID → *GraphRenderCache

func renderGraph(target *ebiten.Image, obj *GObject, ggraph *GImage, ...) {
    sprite := obj.DisplayObject()
    graphics := sprite.Graphics()
    
    // 检查是否需要重建
    hash := graphics.CommandHash()
    cache, ok := graphCaches.Load(obj.ID())
    if ok {
        c := cache.(*GraphRenderCache)
        if c.hash == hash && c.lastSize == obj.Size() {
            // 命中缓存，直接绘制
            drawCachedImage(target, c.image, ...)
            return
        }
    }
    
    // 重建缓存
    img := buildGraphImage(graphics, obj.Size())
    graphCaches.Store(obj.ID(), &GraphRenderCache{
        image: img,
        hash:  hash,
        lastSize: obj.Size(),
    })
    drawCachedImage(target, img, ...)
}
```

### 4.1.3 纹理图集 LRU 驱逐

```go
type AtlasManager struct {
    mu       sync.RWMutex
    atlases  map[string]*PackageAtlas
    lru      *lruCache  // 最近最少使用驱逐
    maxSize  int64       // 最大 GPU 内存限制
    curSize  int64
}

type lruCache struct {
    items    map[string]*lruEntry
    head     *lruEntry
    tail     *lruEntry
    maxItems int
}

// 每次访问时更新 LRU
func (a *AtlasManager) ResolveSprite(item *PackageItem) (*ebiten.Image, ...) {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    // LRU 访问
    a.lru.touch(item.ID)
    
    if atlas, ok := a.atlases[item.Owner.ID]; ok {
        return atlas.getSprite(item)
    }
    return nil, ErrNotLoaded
}

// 加载新 Atlas 时检查内存限制
func (a *AtlasManager) LoadPackage(ctx context.Context, pkg *Package) error {
    // 如果超过限制，驱逐最少使用的 Atlas
    for a.curSize+estimatedSize > a.maxSize {
        victim := a.lru.evict()
        a.unloadAtlas(victim)
    }
    // ...
}
```

---

## 4.2 虚拟列表优化

### 4.2.1 Item 测量缓存

```go
// list_virtual_impl.go
type itemMeasureCache struct {
    mu     sync.RWMutex
    sizes  map[string]laya.Point  // item 模板 ID → 尺寸
}

// 首次测量后缓存
func (l *GList) measureItemSize() laya.Point {
    if l.itemTemplateSize != nil {
        return *l.itemTemplateSize
    }
    // 创建测试 item，测量尺寸
    testItem := l.createItem(0)
    size := laya.Point{X: testItem.Width(), Y: testItem.Height()}
    l.itemTemplateSize = &size
    l.pool.ReturnObject(testItem)
    return size
}
```

### 4.2.2 滚动时增量更新

```go
// 当前: scroll 事件触发时重新计算所有可见 item
// 优化: 仅更新进入/离开视口的 item

func (l *GList) onScroll() {
    newFirst := l.calcFirstIndex()
    if newFirst == l.firstIndex {
        return  // 没有新 item 进入视口
    }
    
    // 增量更新
    if newFirst > l.firstIndex {
        l.removeTopItems(newFirst - l.firstIndex)
        l.addBottomItems(newFirst - l.firstIndex)
    } else {
        l.addTopItems(l.firstIndex - newFirst)
        l.removeBottomItems(l.firstIndex - newFirst)
    }
    l.firstIndex = newFirst
}
```

---

## 4.3 内存优化

### 4.3.1 对象池扩展

```go
// 全局对象池按类型分组，提升命中率
var typedPools = map[string]*sync.Pool{
    "GButton":     {},
    "GTextField":  {},
    "GImage":      {},
    "GLoader":     {},
}

func GetFromPool(typeName string) *GObject {
    pool := typedPools[typeName]
    obj := pool.Get()
    if obj != nil {
        return obj.(*GObject)
    }
    return createByType(typeName)
}

func ReturnToPool(typeName string, obj *GObject) {
    obj.RemoveFromParent()
    obj.Reset()  // 重置所有属性
    typedPools[typeName].Put(obj)
}
```

### 4.3.2 文本贴图缓存

```go
type textCache struct {
    mu      sync.RWMutex
    entries map[string]*textCacheEntry  // hash(text+font+color+size) → image
    maxSize int
}

type textCacheEntry struct {
    image    *ebiten.Image
    lastUsed time.Time
}

// 每 10 秒清理超过 60 秒未使用的缓存
func (c *textCache) startCleanup(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Second)
    go func() {
        for {
            select {
            case <-ticker.C:
                c.cleanup()
            case <-ctx.Done():
                return
            }
        }
    }()
}
```

---

## 4.4 帧循环优化

### 4.4.1 增量脏标记

```go
// 仅重绘标记为 dirty 的区域
type DirtyRect struct {
    X, Y, W, H float64
}

// GComponent 新增
func (c *GComponent) MarkDirty(rect DirtyRect) {
    c.dirtyRects = append(c.dirtyRects, rect)
    if c.parent != nil {
        // 将脏区域转换到父空间并向上传播
        parentRect := transformRect(rect, c.Matrix())
        c.parent.MarkDirty(parentRect)
    }
}

// 渲染时仅重绘脏区域
func drawComponent(target *ebiten.Image, comp *GComponent, ...) {
    if len(comp.dirtyRects) > 0 {
        for _, rect := range comp.dirtyRects {
            subImg := target.SubImage(rect.toImageRect())
            drawChildren(subImg, comp, ...)
        }
        comp.dirtyRects = nil
    }
}
```

### 4.4.2 减少分配

```go
// 事件系统 - 预分配 Event 对象池
var eventPool = sync.Pool{
    New: func() any { return &laya.Event{} },
}

func (d *BasicEventDispatcher) Emit(evt EventType, data any) {
    e := eventPool.Get().(*laya.Event)
    e.Type = evt
    e.Data = data
    e.stopped = false
    defer eventPool.Put(e)
    // ...
}
```

---

## 4.5 性能基准

### 需要建立的基准测试

| 测试 | 当前值 | 目标值 |
|------|--------|--------|
| 虚拟列表 10000 项创建 | ? | < 50ms |
| 虚拟列表滚动帧时间 | ? | < 1ms |
| 100 个 GButton 创建 | ? | < 10ms |
| 文本渲染缓存命中率 | ? | > 80% |
| 图形渲染缓存命中率 | ? | > 90% |
| 每帧渲染 GC 暂停 | ? | < 1ms |
| GPU 纹理内存峰值 | ? | < 100MB |

### 基准测试命令

```bash
# 创建基准
go test -bench=BenchmarkVirtualList -benchmem ./pkg/fgui/widgets
go test -bench=BenchmarkRender -benchmem -tags ebiten ./pkg/fgui/render

# CPU 分析
go test -bench=BenchmarkRender -cpuprofile=cpu.prof -tags ebiten ./pkg/fgui/render

# 内存分析
go test -bench=BenchmarkRender -memprofile=mem.prof -tags ebiten ./pkg/fgui/render
```

---

## 4.6 完成标准
- [ ] 虚拟列表 10000 项创建 < 50ms
- [ ] 滚动帧时间 < 1ms (在 GUI 环境中验证)
- [ ] GGraph 渲染缓存命中率 > 90%
- [ ] 文本贴图缓存有效
- [ ] 纹理图集 LRU 驱逐正常
- [ ] 基准测试就位并可运行
- [ ] `go test -bench . ./pkg/fgui/...` 全部通过
