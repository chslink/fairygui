# GButton 重构方案：对齐 TS 版本的构建流程

## 问题背景

在 Go 版本中，按钮音效会触发两次，原因是事件绑定时机过早且不一致。

## TS vs Go 版本差异

### TypeScript 版本的关键流程（GComponent.ts）

```typescript
1. constructor()           - 仅初始化字段，不绑定事件
2. constructFromResource() - 构建完整的组件树
   ├─ setup_beforeAdd()    - 读取基础属性
   ├─ setup_afterAdd()     - 读取依赖属性
3. constructExtension()    - 最后绑定事件
```

### Go 版本当前的问题

```go
1. NewButton()             - 立即调用 bindEvents() ⚠️
2. BuildComponent()        - 构建子组件树（事件已绑定）
3. applyButtonTemplate()   - 设置模板（事件已绑定）
```

## 重构方案

### 核心改动

1. **延迟事件绑定**：在 `NewButton()` 中不再立即绑定事件
2. **添加 ConstructExtension 接口**：模仿 TS 版本的 constructExtension
3. **统一调用时机**：在组件完整构建完成后才绑定事件

### 文件修改清单

#### 1. pkg/fgui/widgets/button.go

- 修改 `NewButton()`：移除立即的 `bindEvents()` 调用
- 添加 `ConstructExtension()` 方法：读取 section 6 并绑定事件
- 修改 `SetupBeforeAdd()`：移除 section 6 的重复读取

#### 2. pkg/fgui/widgets/interfaces.go（新建）

- 添加 `ExtensionConstructor` 接口：定义 ConstructExtension 方法

#### 3. pkg/fgui/builder/component.go

- 修改 `BuildComponent()`：在构建完成后调用 ConstructExtension
- 删除重复的按钮属性设置代码
- 修改 `applyButtonTemplate()`：移除 section 6 读取

## 详细修改步骤

### 步骤 1：修改 GButton 构造函数

在 `NewButton()` 中移除立即绑定事件的代码：

```go
// 修改前
btn.bindEvents()  // 立即绑定事件 ⚠️
return btn

// 修改后
// 不再立即绑定事件
// 事件绑定将在 ConstructExtension 中完成
return btn
```

### 步骤 2：添加 ConstructExtension 方法

```go
func (b *GButton) ConstructExtension(buf *utils.ByteBuffer) error {
    // 读取 section 6 的扩展属性
    if !buf.Seek(0, 6) || buf.Remaining() <= 0 {
        return nil
    }

    mode := ButtonMode(buf.ReadByte())
    b.SetMode(mode)

    // 音效设置
    if sound := buf.ReadS(); sound != nil {
        b.SetSound(*sound)
    }
    b.SetSoundVolumeScale(float64(buf.ReadFloat32()))

    // Down effect 设置
    b.SetDownEffect(int(buf.ReadByte()))
    b.SetDownEffectValue(float64(buf.ReadFloat32()))

    // 查找 button controller
    for _, ctrl := range b.Controllers() {
        if strings.EqualFold(ctrl.Name, "button") {
            b.SetButtonController(ctrl)
            break
        }
    }

    // 查找 titleObject 和 iconObject
    if template := b.TemplateComponent(); template != nil {
        if titleObj := template.ChildByName("title"); titleObj != nil {
            b.SetTitleObject(titleObj)
        }
        if iconObj := template.ChildByName("icon"); iconObj != nil {
            b.SetIconObject(iconObj)
        }
    }

    // 关键：在构建完成后绑定事件
    b.bindEvents()

    return nil
}
```

### 步骤 3：添加 ExtensionConstructor 接口

新建文件 `pkg/fgui/widgets/interfaces.go`：

```go
package widgets

import "github.com/chslink/fairygui/pkg/fgui/utils"

type ExtensionConstructor interface {
    ConstructExtension(buf *utils.ByteBuffer) error
}
```

### 步骤 4：修改 builder 调用时机

在 `BuildComponent()` 的适当位置添加：

```go
// 在 setupRelations 和 setupGears 之后
f.setupRelations(item, root)
f.setupGears(item, root)

// 调用 ConstructExtension（如果实现了该接口）
if widget, ok := root.GObject.Data().(ExtensionConstructor); ok {
    if buf := item.RawData; buf != nil {
        if err := widget.ConstructExtension(buf); err != nil {
            fmt.Printf("builder: ConstructExtension failed: %v\n", err)
        }
    }
}
```

### 步骤 5：清理冗余代码

1. 从 `BuildComponent()` 中移除重复的按钮属性设置
2. 从 `applyButtonTemplate()` 中移除 section 6 读取
3. 修改 `SetupBeforeAdd()`，跳过在 ConstructExtension 中已处理的属性

## 重构后的流程对比

### 重构后的 Go 版本流程

```go
1. NewButton()
   - 初始化字段：mode, title, icon, sound 等
   - 不绑定事件 ✓

2. BuildComponent() -> buildChild()
   - 构建子对象树
   - SetupBeforeAdd() - 读取基础属性（跳过 section 6）
   - SetupAfterAdd() - 读取依赖属性

3. widget.ConstructExtension(buf)
   - 读取 section 6 的扩展属性
   - 调用 bindEvents() 绑定事件 ✓
```

## 验证要点

重构后，按钮应该只会触发一次音效：

1. **事件只绑定一次**：在 `ConstructExtension` 中调用 `bindEvents()`
2. **正确的调用时机**：在组件完整构建后绑定事件
3. **统一的入口**：所有 GButton 实例都通过 `ConstructExtension` 绑定事件

## 风险控制

- 保留 `sync.Once` 作为第二层防护
- 保持现有 `SetupBeforeAdd`/`SetupAfterAdd` 的接口不变
- 不影响其他组件类型（只有 GButton 实现 ExtensionConstructor）
