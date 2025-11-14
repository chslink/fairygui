# Overflow 功能调研报告

## 问题场景

`demo/UIProject/assets/Basics/Demo_Clip&Scroll.xml` 场景展示了多种 Overflow 行为，但 Go 版本完全未实现该功能。

## 场景中的 Overflow 示例

### 1. Component1.xml (Overflow: hidden)
```xml
<component size="224,310" overflow="hidden">
  <displayList>
    <image id="n1" name="n1" src="es4130" xy="0,0"/>
  </displayList>
</component>
```

### 2. Component7.xml (Overflow: Free Scroll)
```xml
<component size="225,310" overflow="scroll" scroll="both">
  <displayList>
    <image id="n1" name="n1" src="es4130" xy="0,0"/>
  </displayList>
</component>
```

### 3. Component8.xml (Overflow: Vertical scroll and margins)
```xml
<component size="225,343" overflow="scroll" scroll="both" margin="30,30,30,30">
  <displayList>
    <image id="n1" name="n1" src="es4130" xy="0,0"/>
  </displayList>
</component>
```

## TypeScript 参考实现

### OverflowType 枚举 (FieldTypes.ts:45-49)
```typescript
export enum OverflowType {
    Visible,  // 0 - 默认，内容可见且不裁剪
    Hidden,   // 1 - 裁剪超出边界的内容
    Scroll    // 2 - 创建滚动区域
}
```

### setupOverflow 方法 (GComponent.ts:746-762)
```typescript
protected setupOverflow(overflow: number): void {
    if (overflow == OverflowType.Hidden) {
        if (this._displayObject == this._container) {
            this._container = new Laya.Sprite();
            this._displayObject.addChild(this._container);
        }
        this.updateMask();
        this._container.pos(this._margin.left, this._margin.top);
    }
    else if (this._margin.left != 0 || this._margin.top != 0) {
        if (this._displayObject == this._container) {
            this._container = new Laya.Sprite();
            this._displayObject.addChild(this._container);
        }
        this._container.pos(this._margin.left, this._margin.top);
    }
}
```

### updateMask 方法 (GComponent.ts:724-734)
```typescript
protected updateMask(): void {
    var rect: Laya.Rectangle = this._displayObject.scrollRect;
    if (!rect)
        rect = new Laya.Rectangle();

    rect.x = this._margin.left;
    rect.y = this._margin.top;
    rect.width = this._width - this._margin.right;
    rect.height = this._height - this._margin.bottom;

    this._displayObject.scrollRect = rect;
}
```

### 从 buffer 读取 overflow (GComponent.ts:1040-1054)
```typescript
// 读取 margin
if (buffer.version >= 2) {
    this._margin.top = buffer.getInt32();
    this._margin.bottom = buffer.getInt32();
    this._margin.left = buffer.getInt32();
    this._margin.right = buffer.getInt32();
}

// 读取 overflow
var overflow: number = buffer.readByte();
if (overflow == OverflowType.Scroll) {
    var savedPos: number = buffer.pos;
    buffer.seek(0, 7);
    this.setupScroll(buffer);
    buffer.pos = savedPos;
}
else
    this.setupOverflow(overflow);
```

## Go 版本当前状态

### 缺失的功能

1. **OverflowType 枚举**
   - 完全缺失，需要在 `pkg/fgui/core` 中定义

2. **Margin 结构体**
   - TypeScript 有 `Margin` 类 (top, bottom, left, right)
   - Go 版本完全缺失

3. **setupOverflow 方法**
   - `pkg/fgui/core/gcomponent.go` 中不存在
   - 需要实现创建独立 container 和 mask 的逻辑

4. **updateMask 方法**
   - Go 版本缺失
   - 需要使用兼容层的 scrollRect 功能

5. **buffer 读取逻辑**
   - `pkg/fgui/builder/component.go` 的 `BuildComponent` 中未读取 margin 和 overflow
   - ComponentData 结构体可能也缺少 margin 字段

### 已存在的功能

1. **SetupScroll 方法**
   - `pkg/fgui/core/gcomponent.go:118` 已经实现
   - 处理 `overflow="scroll"` 的情况

## 实现计划

### 第一步：定义数据结构

在 `pkg/fgui/core/types.go` 或新建 `pkg/fgui/core/overflow.go`：

```go
// OverflowType 定义组件溢出内容的处理方式
type OverflowType int

const (
	OverflowVisible OverflowType = 0 // 默认，内容可见且不裁剪
	OverflowHidden  OverflowType = 1 // 裁剪超出边界的内容
	OverflowScroll  OverflowType = 2 // 创建滚动区域
)

// Margin 定义组件的边距
type Margin struct {
	Top    int
	Bottom int
	Left   int
	Right  int
}
```

### 第二步：扩展 GComponent

在 `pkg/fgui/core/gcomponent.go` 中：

```go
type GComponent struct {
    // ... 现有字段 ...
    margin       Margin
    overflow     OverflowType
    container    *laya.Sprite  // 可能与 display 不同，用于应用 margin offset
}

// SetupOverflow 配置组件的 overflow 行为
func (c *GComponent) SetupOverflow(overflow OverflowType) {
    if overflow == OverflowHidden {
        // 如果 display 和 container 是同一个，创建新的 container
        if c.display == c.container {
            c.container = laya.NewSprite()
            c.display.AddChild(c.container)
        }
        c.updateMask()
        c.container.SetPosition(float64(c.margin.Left), float64(c.margin.Top))
    } else if c.margin.Left != 0 || c.margin.Top != 0 {
        // 即使不是 Hidden，如果有 margin 也需要独立的 container
        if c.display == c.container {
            c.container = laya.NewSprite()
            c.display.AddChild(c.container)
        }
        c.container.SetPosition(float64(c.margin.Left), float64(c.margin.Top))
    }
}

// UpdateMask 更新裁剪矩形（用于 overflow=hidden）
func (c *GComponent) UpdateMask() {
    // 使用 laya.Sprite 的 SetScrollRect 方法
    rect := &laya.Rectangle{
        X:      float64(c.margin.Left),
        Y:      float64(c.margin.Top),
        Width:  c.width - float64(c.margin.Right),
        Height: c.height - float64(c.margin.Bottom),
    }
    c.display.SetScrollRect(rect)
}
```

### 第三步：更新 ComponentData

在 `pkg/fgui/assets/component.go` 中：

```go
type ComponentData struct {
    // ... 现有字段 ...
    Margin   Margin       // 边距
    Overflow OverflowType // 溢出处理方式
}
```

### 第四步：Builder 读取 overflow

在 `pkg/fgui/builder/component.go` 的 `BuildComponent` 方法中：

```go
// 在构建根组件后，读取 RawData 的 section 6（与 TypeScript 的 setup 方法对应）
if buf := item.RawData; buf != nil {
    saved := buf.Pos()
    defer buf.SetPos(saved)

    if buf.Seek(0, 6) && buf.Remaining() > 0 {
        // 读取 margin (version >= 2)
        if buf.Version() >= 2 {
            root.SetMargin(core.Margin{
                Top:    int(buf.ReadInt32()),
                Bottom: int(buf.ReadInt32()),
                Left:   int(buf.ReadInt32()),
                Right:  int(buf.ReadInt32()),
            })
        }

        // 读取 overflow
        overflow := core.OverflowType(buf.ReadByte())
        if overflow == core.OverflowScroll {
            // 已经有 SetupScroll，需要确保在正确位置调用
            savedPos := buf.Pos()
            if buf.Seek(0, 7) {
                root.SetupScroll(buf)
            }
            buf.SetPos(savedPos)
        } else {
            root.SetupOverflow(overflow)
        }
    }
}
```

### 第五步：HandleSizeChanged

在 `pkg/fgui/core/gcomponent.go` 的 `HandleSizeChanged` 中：

```go
func (c *GComponent) HandleSizeChanged() {
    // ... 现有代码 ...

    if c.scrollPane != nil {
        c.scrollPane.OnOwnerSizeChanged()
    } else if c.display.ScrollRect() != nil {
        c.UpdateMask()
    }
}
```

## 兼容层需求

需要确认 `internal/compat/laya/sprite.go` 中已实现：

1. `SetScrollRect(rect *Rectangle)` - 设置裁剪矩形
2. `ScrollRect() *Rectangle` - 获取当前裁剪矩形

如果未实现，需要添加这些方法。

## 测试计划

1. **单元测试**：
   - 测试 SetupOverflow(OverflowHidden) 创建独立 container
   - 测试 UpdateMask 正确设置 scrollRect
   - 测试 margin 偏移正确应用

2. **集成测试**：
   - 加载 Component1.xml (overflow=hidden)
   - 验证内容被正确裁剪
   - 验证 margin 正确应用

3. **Demo 验证**：
   - 运行 `go run ./demo`
   - 切换到 Demo_Clip&Scroll 场景
   - 视觉验证各种 overflow 行为

## 优先级

1. **高优先级**：OverflowHidden - 最常用，影响布局
2. **中优先级**：Margin 支持 - 与 overflow 配合使用
3. **低优先级**：OverflowScroll 改进 - 已有基础实现，可能需要调整

## 注意事项

1. **Container vs Display**：
   - 通常 `container == display`
   - 当需要 margin offset 或 mask 时，创建独立的 container
   - 所有子对象应该添加到 container 而不是 display

2. **渲染顺序**：
   - UpdateMask 应该在 HandleSizeChanged 时调用
   - SetupOverflow 应该在组件构建完成后、添加子对象前调用

3. **兼容性**：
   - scrollRect 是 Laya 的标准功能
   - Ebiten 可能需要用其他方式实现（如裁剪区域）

4. **性能**：
   - 避免每帧更新 mask
   - 只在尺寸改变时更新
