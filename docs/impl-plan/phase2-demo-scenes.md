# Phase 2: 补齐 Demo 场景交互

> 阶段目标: 将 10 个 SimpleScene 转为完整交互场景，补充 3 个缺失场景
> 预计工时: 5-7 天
> 前置依赖: Phase 1 (缺失模块)

---

## 2.0 架构背景

### 现有 Go Demo 架构
```
demo/
├── main.go            ← Ebiten 游戏循环入口
├── scenes/
│   ├── manager.go     ← 场景生命周期管理器
│   ├── environment.go ← 共享服务 (Loader, Factory, Atlas)
│   ├── mainmenu.go    ← 主菜单（14 按钮 → 场景路由）
│   ├── basics.go      ← 综合演示（自定义场景）
│   ├── transition_demo.go ← 过渡动画演示
│   ├── virtual_list_demo.go ← 虚拟列表
│   ├── loop_list_demo.go  ← 循环列表
│   ├── joystick.go    ← 摇杆演示
│   ├── simple_scene.go ← 通用静态场景包装器
│   └── util.go        ← 辅助函数
└── debug/server.go     ← HTTP 调试服务器
```

### SimpleScene 机制
`SimpleScene` 是一个通用包装器，仅调用 `Factory.BuildComponent()` 加载组件并显示。10 个场景用 SimpleScene 加载了 UI 但缺乏交互逻辑。

### 改造策略
为每个需要交互的场景创建独立文件，删除 `manager.go` 中对应的 `SimpleScene` 注册。

---

## 2.1 场景逐个实现

### 2.1.1 PullToRefreshDemo（下拉刷新）

**TS 参考**: `PullToRefreshDemo.ts` (82 行) + `ScrollPaneHeader.ts` (24 行)

**需要的能力**:
- GList.SetVirtual()
- ScrollPane 的 PULL_DOWN_RELEASE / PULL_UP_RELEASE 事件
- ScrollPane.lockHeader(高度) 锁定刷新头部
- Laya.timer.once() 模拟异步请求延迟
- UIObjectFactory.setExtension() 注册自定义 ScrollPaneHeader

**Go 实现文件**: `demo/scenes/pull_to_refresh.go`

**核心交互代码**:
```go
type PullToRefreshScene struct {
    comp     *fgui.GComponent
    list1    *fgui.GList
    list2    *fgui.GList
}

func (s *PullToRefreshScene) setup() {
    // 1. 获取两个 GList
    s.list1 = s.comp.ChildByName("list1").(*fgui.GList)
    s.list2 = s.comp.ChildByName("list2").(*fgui.GList)
    
    // 2. 设置虚拟列表
    s.list1.SetVirtual(true)
    s.list1.SetNumItems(20)
    s.list1.SetItemRenderer(func(index int, obj *fgui.GObject) {
        // 设置标题
    })
    
    // 3. 监听下拉释放事件
    s.list1.ScrollPane().OnPullDownRelease(func() {
        // 更新 header 文本
        header := s.list1.ScrollPane().Header()
        header.ChildByName("title").SetProp("text", "正在刷新...")
        // 锁定 header
        s.list1.ScrollPane().LockHeader(50)
        // 模拟异步
        laya.CallLater(2*time.Second, func() {
            s.list1.SetNumItems(30)
            s.list1.ScrollPane().LockHeader(0)
        })
    })
}
```

**需要新增/修改的库文件**:
- `pkg/fgui/core/scrollpane.go` — 添加 PullDownRelease/PullUpRelease 事件、LockHeader 方法、Header/Footer 访问器
- `pkg/fgui/core/gobject.go` — 支持 `SetProp("text", value)` 文本属性

---

### 2.1.2 ModalWaitingDemo（模态等待）

**TS 参考**: `ModalWaitingDemo.ts` (28 行)

**需要的能力**:
- UIConfig.globalModalWaiting / windowModalWaiting URL 配置
- GRoot.ShowModalWait() / CloseModalWait()
- Window.ShowModalWait() / CloseModalWait()

**Go 实现文件**: `demo/scenes/modal_waiting.go`

```go
func (s *ModalWaitingScene) onShowGlobal() {
    root := fgui.Root()
    root.ShowModalWait("正在处理中...")
    laya.CallLater(3*time.Second, func() {
        root.CloseModalWait()
    })
}
```

**需要修改的库文件**:
- `pkg/fgui/core/groot.go` — 添加 ShowModalWait/CloseModalWait
- `pkg/fgui/core/uiconfig.go` — 添加 WindowModalWaiting/GlobalModalWaiting

---

### 2.1.3 BagDemo（背包窗口）

**TS 参考**: `BagDemo.ts` (50 行)

**需要的能力**:
- Window 系统（contentPane, onShown 生命周期）
- GList.SetVirtual()
- 随机物品数据生成

**Go 实现文件**: `demo/scenes/bag.go`

---

### 2.1.4 ChatDemo（聊天演示）

**TS 参考**: `ChatDemo.ts` (116 行) + `EmojiParser.ts` (20 行)

**需要的能力**:
- UBB 解析器扩展（EmojiParser）
- GList.SetVirtual() + itemProvider（多类型 item）
- GRichTextField.EnsureSizeCorrect() 动态调整
- PopupMenu 用于表情选择
- GTextInput 键盘事件（ENTER 发送）

**Go 实现文件**: 
- `demo/scenes/chat.go`
- `demo/scenes/emoji_parser.go`

**EmojiParser 实现**:
```go
type EmojiParser struct {
    // 持有对内部 UBB 解析器或 render 的引用
}

func (p *EmojiParser) Parse(text string) []TextSegment {
    // 解析 [:emoji_name] 标签
    // 替换为对应的表情图标
}
```

---

### 2.1.5 ListEffectDemo（列表特效）

**TS 参考**: `ListEffectDemo.ts` (39 行) + `MailItem.ts` (24 行)

**需要的能力**:
- UIObjectFactory.setExtension() 注册 MailItem 扩展
- MailItem 是 GButton 子类，带控制器和 Transition
- 为列表项设置递进延迟播放动画

**Go 实现文件**: 
- `demo/scenes/list_effect.go`
- 需要实现 `MailItem` 扩展类

---

### 2.1.6 ScrollPaneDemo（滚动面板演示）

**TS 参考**: `ScrollPaneDemo.ts` (58 行)

**需要的能力**:
- 嵌套 ScrollPane（每个 item 内有滑动面板）
- ScrollPane 位置检测与重置

**Go 实现文件**: `demo/scenes/scroll_pane.go`

---

### 2.1.7 TreeViewDemo（树形视图）

**TS 参考**: `TreeViewDemo.ts` (71 行)

**需要的能力**:
- GTree.treeNodeRender 自定义渲染
- GTreeNode 数据绑定

**Go 实现文件**: `demo/scenes/tree_view.go`

---

### 2.1.8 GuideDemo（引导演示）

**TS 参考**: `GuideDemo.ts` (33 行)

**需要的能力**:
- localToGlobalRect / globalToLocalRect 坐标转换
- GTween.to2() 平滑移动

**Go 实现文件**: `demo/scenes/guide.go`

---

### 2.1.9 CooldownDemo（冷却演示）

**TS 参考**: `CooldownDemo.ts` (23 行)

**需要的能力**:
- GProgressBar 循环动画
- GTween 无限重复

**Go 实现文件**: `demo/scenes/cooldown.go`

---

### 2.1.10 BasicsDemo Depth 子场景（修复）

**当前状态**: 代码中有注释 "需要 sortingOrder 和 draggable 功能，待实现后添加"

**需要的能力**: sortingOrder 已实现，draggable 在 Phase 1 中实现

---

## 2.2 新增 3 个场景

### 2.2.1 SceneTreeDemo（场景树演示）
展示 GComponent 层级嵌套、查找、遍历功能。

### 2.2.2 ExtensionDemo（扩展演示）
展示自定义 GComponent 子类注册和使用。

### 2.2.3 RelationDemo（关联布局演示）
展示 Relations 系统所有关联类型的可视化效果。

---

## 2.3 主菜单更新

修改 `demo/scenes/mainmenu.go`，添加新场景按钮：
- 补充缺失的 `n3` (JoystickDemo) 在 TS 中的位置
- 添加 SceneTree/Extension/Relation Demo 按钮
- 确保 22 个场景全部可导航

---

## 2.4 需要修改/新增的库文件

| 文件 | 修改 |
|------|------|
| `pkg/fgui/core/scrollpane.go` | PullDownRelease/PullUpRelease 事件、LockHeader、Header/Footer |
| `pkg/fgui/core/groot.go` | ShowModalWait/CloseModalWait |
| `pkg/fgui/core/uiconfig.go` | WindowModalWaiting/GlobalModalWaiting URL |
| `pkg/fgui/core/gobject.go` | SetProp 文本属性路径 |
| `pkg/fgui/widgets/text.go` | 支持 EmojiParser 自定义解析器 |
| `internal/text/ubb.go` | 开放 UBB 解析器扩展接口 |

---

## 2.5 完成标准
- [ ] 10 个 SimpleScene 全部转为独立交互场景
- [ ] 3 个新场景实现
- [ ] 主菜单支持所有 22 个场景导航
- [ ] Basics Demo Depth 子场景修复
- [ ] 所有场景可在 GUI 环境中正确渲染和交互
- [ ] `go build ./...` 通过
