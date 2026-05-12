# Phase 1: 补齐核心缺失模块

> 阶段目标: 实现 12 个缺失模块，使 TS 模块实现率达到 94.4%
> 预计工时: 5-7 天
> 前置依赖: Phase 0 (Bug 修复)

---

## 1.1 Window 窗口系统 (P0, 中复杂度)

### TypeScript 参考
`laya_src/fairygui/Window.ts` (235 行)

### 功能概述
Window 是带模态/拖拽/关闭按钮的弹出窗口容器。TS 原版核心能力：
- `contentPane` getter/setter — 自动解析 frame/closeButton/dragArea/contentArea
- `show()` / `hide()` / `hideImmediately()` — 显示/隐藏
- `modal` — 背景模态遮罩
- `bringToFontOnClick` — 点击置顶
- `centerOn(root)` — 居中（带关系约束）
- `showModalWait()` / `closeModalWait()` — 窗口级加载等待
- `doShowAnimation()` / `doHideAnimation()` — 动画钩子
- `onInit()` / `onShown()` / `onHide()` — 生命周期回调

### Go 实现设计

**新建文件**: `pkg/fgui/core/window.go`

```go
package core

type Window struct {
    *GComponent
    
    contentPane          *GComponent
    closeButton          *GObject
    dragArea             *GObject
    contentArea          *GObject
    frame                *GObject
    modalWaitingPane     *GObject
    modal                bool
    bringToFrontOnClick  bool
    isShowing            bool
    isTop                bool
    
    // 动画相关
    showAnimation        func()
    hideAnimation        func()
    
    // 拖拽状态
    dragging             bool
    dragStartPos         Point
    
    // 用户回调
    onInitHandler        func()
    onShownHandler       func()
    onHideHandler        func()
    
    // UI 资源源（延迟加载）
    uiSource             IUISource
    initDone             bool
}

func NewWindow() *Window
func (w *Window) SetContentPane(pane *GComponent)
func (w *Window) ContentPane() *GComponent
func (w *Window) Show()
func (w *Window) Hide()
func (w *Window) HideImmediately()
func (w *Window) ToggleStatus()
func (w *Window) BringToFront()
func (w *Window) CenterOn(root *GComponent, restraint bool)
func (w *Window) SetModal(modal bool)
func (w *Window) IsShowing() bool
func (w *Window) ShowModalWait() 
func (w *Window) CloseModalWait()
func (w *Window) SetOnInit(fn func())
func (w *Window) SetOnShown(fn func())
func (w *Window) SetOnHide(fn func())
```

**关键实现细节**:

1. **contentPane 设置时自动提取子组件**:
   ```go
   func (w *Window) SetContentPane(pane *GComponent) {
       w.contentPane = pane
       w.AddChild(pane)
       // 从 contentPane 中查找标准命名的子组件
       w.frame = pane.ChildByName("frame")
       w.closeButton = pane.ChildByName("closeButton")
       w.dragArea = pane.ChildByName("dragArea")
       w.contentArea = pane.ChildByName("contentArea")
       // 绑定关闭按钮事件
       if w.closeButton != nil {
           w.closeButton.OnClick(w.Hide)
       }
       // 绑定拖拽区域事件
       // ...
   }
   ```

2. **Show() 需要调用 GRoot 的 ShowWindow**:
   需要在 GRoot 中新增 `ShowWindow`/`HideWindow`/`HideWindowImmediately` 方法。

3. **拖拽实现**:
   监听 dragArea 的 MOUSE_DOWN → 记录偏移 → stage MOUSE_MOVE → 移动窗口 → stage MOUSE_UP

4. **模态实现**:
   在 GRoot 中管理模态遮罩层（半透明 GGraph）。

**需要修改的现有文件**:
- `pkg/fgui/core/groot.go` — 新增 ShowWindow/HideWindow/HideWindowImmediately/BringToFront/modalLayer

**依赖模块**: IUISource（见 1.8）

---

## 1.2 PopupMenu 右键菜单 (P0, 中复杂度)

### TypeScript 参考
`laya_src/fairygui/PopupMenu.ts` (149 行)

### 功能概述
- 从 `UIConfig.popupMenu` URL 加载模板
- `addItem(caption, handler)` / `addSeperator()`
- `setItemVisible/setItemGrayed/setItemCheckable/setItemChecked`
- `removeItem/clearItems`
- `show(target, dir)` — 委托给 `GRoot.showPopup()`

### Go 实现设计

**新建文件**: `pkg/fgui/core/popupmenu.go`

```go
package core

type PopupMenu struct {
    contentPane *GComponent
    list        *GList
    
    items       []*popupMenuItem
}

type popupMenuItem struct {
    caption   string
    handler   func()
    checkable bool
    checked   bool
    visible   bool
    grayed    bool
    separator bool
}

func NewPopupMenu() *PopupMenu
func (m *PopupMenu) AddItem(caption string, handler func()) *popupMenuItem
func (m *PopupMenu) AddItemAt(index int, caption string, handler func()) *popupMenuItem
func (m *PopupMenu) AddSeparator()
func (m *PopupMenu) SetItemText(index int, text string)
func (m *PopupMenu) SetItemVisible(index int, visible bool)
func (m *PopupMenu) SetItemGrayed(index int, grayed bool)
func (m *PopupMenu) SetItemCheckable(index int, checkable bool)
func (m *PopupMenu) SetItemChecked(index int, checked bool)
func (m *PopupMenu) RemoveItem(index int)
func (m *PopupMenu) ClearItems()
func (m *PopupMenu) ItemCount() int
func (m *PopupMenu) Show(target *GObject, dir PopupDirection)
func (m *PopupMenu) Hide()
```

**实现要点**:
1. 内容面板从 `UIConfig.popupMenu` 资源创建
2. 每个 item 是一个 GButton，通过 `contentPane.ChildByName("list")` 的 GList 管理
3. 分隔线用 `UIConfig.popupMenu_seperator` 资源
4. 点击 item 后延迟 100ms 关闭（防抖），调用 handler
5. 选中状态通过 item 的 `selected` + 图标切换

**需要修改的现有文件**:
- `pkg/fgui/core/uiconfig.go` — 添加 `popupMenu_seperator` 字段

---

## 1.3 DragDropManager 拖拽管理器 (P0, 低复杂度)

### TypeScript 参考
`laya_src/fairygui/DragDropManager.ts` (62 行)

### 功能概述
- 单例模式
- 创建 GLoader 作为拖拽代理（100x100, pivot 0.5, sortingOrder=1000000）
- `startDrag(source, icon, sourceData, touchID)` — 显示代理，添加到 GRoot，开始拖拽
- `cancel()` — 取消拖拽
- 拖拽结束时冒泡查找 `Events.DROP` 监听器，传递 `sourceData`

### Go 实现设计

**新建文件**: `pkg/fgui/core/dragdrop.go`

```go
package core

var dragDropInst *DragDropManager

type DragDropManager struct {
    agent       *GLoader  // 拖拽视觉代理
    sourceData  any       // 传递的数据
    dragging    bool
    touchID     int
}

func DragDrop() *DragDropManager  // 单例访问
func (d *DragDropManager) StartDrag(source *GObject, icon string, sourceData any, touchID int)
func (d *DragDropManager) Cancel()
func (d *DragDropManager) IsDragging() bool
func (d *DragDropManager) Agent() *GLoader
```

**实现要点**:
1. `StartDrag` 创建/复用 GLoader，设置图标，添加到 GRoot
2. 监听 stage MOUSE_MOVE/MOUSE_UP 事件
3. MOUSE_MOVE → 更新代理位置
4. MOUSE_UP → 遍历 hit test 结果查找 DROP 监听器 → 触发回调 → 移除代理
5. 需要在 `internal/compat/laya/event.go` 添加 `EventDrop` 事件类型

**核心算法 — DROP 事件冒泡**:
```go
func (d *DragDropManager) onDragEnd(hitTarget *laya.Sprite) {
    // 从 hitTarget 向上遍历显示树
    for current := hitTarget; current != nil; current = current.Parent() {
        owner := ownerAsGObject(current)
        if owner != nil {
            // 检查 owner 是否注册了 DROP 监听器
            // 如果注册了，触发回调并传递 sourceData
            owner.Emit(laya.EventDrop, d.sourceData)
        }
    }
    d.Cancel()
}
```

**需要修改的现有文件**:
- `internal/compat/laya/event.go` — 添加 `EventDrop`
- `pkg/fgui/core/gobject.go` — 添加 `OnDrop` 便捷方法

---

## 1.4 UIObjectFactory 扩展机制 (P0, 低复杂度)

### TypeScript 参考
`laya_src/fairygui/UIObjectFactory.ts` (92 行)

### 功能概述
- `setExtension(url, type)` — 注册 URL 到自定义 GComponent 子类的映射
- `setPackageItemExtension(url, type)` — 同上（别名）
- `setLoaderExtension(type)` — 设置自定义加载器类
- `resolvePackageItemExtension(pi)` — 延迟解析
- `newObject(type | PackageItem)` — 核心工厂方法

### Go 实现设计

**修改文件**: `pkg/fgui/widgets/factory.go` → 重命名为 `pkg/fgui/widgets/object_factory.go`

```go
package widgets

import "sync"

var (
    objFactoryOnce sync.Once
    objectFactory  *UIObjectFactory
)

type UIObjectFactory struct {
    extensions       map[string]func() *core.GComponent  // URL → 构造函数
    loaderType       func() *GLoader                      // 自定义加载器类型
}

func GetObjectFactory() *UIObjectFactory

// SetExtension 注册自定义组件类型
// url 格式: "ui://PackageName/ItemName" 或 "PackageName/ItemName"
func (f *UIObjectFactory) SetExtension(url string, ctor func() *core.GComponent)

// SetLoaderExtension 设置自定义 GLoader 子类
func (f *UIObjectFactory) SetLoaderExtension(ctor func() *GLoader)

// NewObject 根据类型创建对象
// 支持通过 ObjectTypeID 创建（GImage=1, GMovieClip=2, GGraph=5, GGroup=6, GTextField=7, 
//   GRichTextField=8, GTextInput=9, GLoader=10, GList=11, GLabel=12, GButton=13,
//   GComboBox=14, GProgressBar=15, GSlider=16, GScrollBar=17, GTree=18, GLoader3D=22）
func (f *UIObjectFactory) NewObject(pi *assets.PackageItem) *core.GObject

// NewObjectWithClass 使用用户指定的类创建对象
func (f *UIObjectFactory) NewObjectWithClass(pi *assets.PackageItem, userClass func() *core.GComponent) *core.GObject
```

**实现细节**:
```go
func (f *UIObjectFactory) NewObject(pi *assets.PackageItem) *core.GObject {
    // 1. 优先使用扩展注册
    if ctor, ok := f.extensions[pi.ID]; ok {
        comp := ctor()
        return comp.GObject
    }
    // 2. 根据 ObjectType 创建
    switch pi.ObjectType {
    case 1: return NewImage().GObject
    case 2: return NewMovieClip().GObject
    case 5: return NewGraph().GObject
    // ... 等等
    default: return core.NewGObject()
    }
}
```

---

## 1.5 ToolSet 工具函数集 (P0, 低复杂度)

### TypeScript 参考
`laya_src/fairygui/utils/ToolSet.ts` (181 行)

### Go 实现设计

**新建文件**: `pkg/fgui/utils/toolset.go`

```go
package utils

import (
    "fmt"
    "math"
    "strconv"
    "strings"
)

// ARGBToHTMLColor 将 ARGB uint32 转换为 #AARRGGBB 格式
func ARGBToHTMLColor(argb uint32, withAlpha bool) string

// HTMLColorToUint32 将 #RRGGBB 或 #AARRGGBB 字符串转换为 uint32
func HTMLColorToUint32(str string) (uint32, error)

// Clamp 数值约束
func Clamp(value, min, max float64) float64

// Clamp01 约束到 [0,1]
func Clamp01(value float64) float64

// Lerp 线性插值
func Lerp(start, end, percent float64) float64

// Distance 欧几里得距离
func Distance(x1, y1, x2, y2 float64) float64

// StartsWith 不区分大小写前缀检查
func StartsWith(s, prefix string) bool

// EndsWith 不区分大小写后缀检查
func EndsWith(s, suffix string) bool

// EncodeHTML 转义 HTML 实体
func EncodeHTML(s string) string

// DisplayObjectToGObject 通过 owner 链找到 GObject
func DisplayObjectToGObject(sprite *laya.Sprite) *GObject
```

**`setColorFilter` 的 Go 实现**:
```go
// SetColorFilter 对 sprite 应用颜色滤镜
// color 可以是:
//   - nil: 清除滤镜
//   - string: 颜色值如 "#RRGGBB" 或 "#AARRGGBB"
//   - []float64: 20 元素的颜色矩阵
func SetColorFilter(sprite *laya.Sprite, color any)
```

---

## 1.6 Controller Action 系统 (P1, 低复杂度)

### 背景
TS 原版的 Controller 可以关联 Action，当页面切换时自动触发（如播放过渡动画、切换其他控制器页面）。Go 版 Controller 目前只有选择和回调。

### 文件清单

**新建文件**:
1. `pkg/fgui/core/controller_action.go` — ControllerAction 基类
2. `pkg/fgui/core/change_page_action.go` — 切换页面动作
3. `pkg/fgui/core/play_transition_action.go` — 播放过渡动作

### Go 实现设计

#### controller_action.go
```go
package core

type ControllerAction struct {
    fromPages []string  // 触发页面条件
    toPages   []string  // 目标页面条件
}

func (a *ControllerAction) Run(ctrl *Controller, prevPage, curPage string) {
    // 检查 fromPages/toPages 是否匹配
    // 匹配 → enter(ctrl) / leave(ctrl)
}

func (a *ControllerAction) Enter(ctrl *Controller) {}  // 子类覆盖
func (a *ControllerAction) Leave(ctrl *Controller) {}  // 子类覆盖
```

#### change_page_action.go
```go
type ChangePageAction struct {
    ControllerAction
    objectID       string  // 目标组件 ID（空 = 控制器父组件）
    controllerName string
    targetPage     string  // "~1"=按索引, "~2"=按页面名
}

func (a *ChangePageAction) Enter(ctrl *Controller) {
    // 解析目标组件 → 获取控制器 → 设置页面
}
```

#### play_transition_action.go
```go
type PlayTransitionAction struct {
    ControllerAction
    transitionName string
    playTimes      int
    delay          float64
    stopOnExit     bool
}

func (a *PlayTransitionAction) Enter(ctrl *Controller) {
    // 获取 Transition → Play()
}
func (a *PlayTransitionAction) Leave(ctrl *Controller) {
    // 如果 stopOnExit → Stop()
}
```

**需要修改的现有文件**:
- `pkg/fgui/core/controller.go` — 添加 `actions []*ControllerAction` 字段，在 `applySelection()` 中调用 actions

**解析集成**:
- 在 Builder 的 `parseControllerActions()` 中解析二进制数据创建 Action

---

## 1.7 IUISource 接口 (P1, 低复杂度)

### TypeScript 参考
`laya_src/fairygui/IUISource.ts` (7 行，仅接口定义)

### Go 实现

```go
// 位置: pkg/fgui/core/iuisource.go
package core

type IUISource interface {
    FileName() string
    IsLoaded() bool
    Load(callback func())
}
```

Window 使用 IUISource 进行延迟加载：如果设置了 `uiSource`，初始化时调用 `Load()`，回调中设置 `contentPane`。

---

## 1.8 ChildHitArea 子碰撞区域 (P2, 中复杂度)

### 功能
将命中测试委托给子精灵。Go 版本可在 `render/hit_area.go` 中实现。

---

## 1.9 AsyncOperation 异步构建 (P2, 中复杂度)

### 功能
跨多帧异步构建 UI，防止大组件导致帧卡顿。Go 版本可用 goroutine + channel 实现。

---

## 1.10 需要修改的现有文件汇总

| 文件 | 修改内容 |
|------|---------|
| `internal/compat/laya/event.go` | 添加 ListenerID、EventDrop |
| `pkg/fgui/core/gobject.go` | OnClick/OffClick 改用 ID、添加 OnDrop |
| `pkg/fgui/core/groot.go` | 添加 ShowWindow/HideWindow/modalLayer |
| `pkg/fgui/core/uiconfig.go` | 添加 popupMenu_seperator、modalWaiting URL |
| `pkg/fgui/core/controller.go` | 添加 actions 支持 |
| `pkg/fgui/widgets/factory.go` | 重写为 UIObjectFactory |
| `pkg/fgui/builder/component.go` | 添加 ControllerAction 解析 |
| `pkg/fgui/api.go` | 导出新类型 |

---

## 1.11 完成标准
- [ ] Window 系统可用：Show/Hide/CenterOn/Modal/动画
- [ ] PopupMenu 可用：添加/移除/选中/分隔线
- [ ] DragDropManager 可用：StartDrag/Cancel/DROP 事件
- [ ] UIObjectFactory.setExtension 可用
- [ ] ToolSet 工具函数完整
- [ ] Controller Action 系统完整
- [ ] 所有新模块 100% 测试覆盖
- [ ] `go build ./...` 通过
- [ ] `go test ./...` 通过
