# FairyGUI Ebiten Port Progress

## 2026-05-12 — Phase 0-5: 大规模修复与增强

### Phase 0: 致命 Bug 修复
- [x] **修复 OffClick 永远无法注销 Bug**: 实现 `ListenerID` 事件系统，`OnWithID`/`OffByID` 正确处理匿名闭包
- [x] **修复音频资源泄漏**: `AudioPlayer` 追踪活跃播放器 goroutine，等待播放完成后 `Close()`
- [x] 排查 3 个被 Skip 的测试（已有深度问题，记录待后续修复）

### Phase 1: 补齐核心缺失模块
- [x] `core/iuisource.go` — IUISource 延迟加载接口
- [x] `core/dragdrop.go` — DragDropManager 全局拖拽管理器（含 DROP 事件冒泡）
- [x] `core/window.go` — Window 模态窗口（拖拽/关闭按钮/动画钩子/模态遮罩）
- [x] `core/popupmenu.go` — PopupMenu 右键菜单（含分隔线/选中状态）
- [x] `core/controller_action.go` — ControllerAction 基类
- [x] `core/change_page_action.go` — ChangePageAction（切换控制器页面）
- [x] `core/play_transition_action.go` — PlayTransitionAction（播放过渡动画）
- [x] `utils/toolset.go` — ToolSet 工具函数（颜色转换/Clamp/Lerp/HTML 转义）
- [x] `uiconfig.go` — 扩展 BringWindowToFrontOnClick/GlobalModalWaiting/PopupMenuSeperator
- [x] `groot.go` — 添加 Window 栈管理/ShowWindow/HideWindow/ModalWait/模态遮罩
- [x] `controller.go` — 集成 Actions 链，`RunActions()` 在页面切换时触发

### Phase 2: Demo 场景 + ScrollPane 增强
- [x] `scrollpane.go` — 添加 Header/Footer/LockHeader/LockFooter/OnPullDownRelease/OnPullUpRelease/IsBottomMost
- [x] `simple_scene.go` — 添加 `WithSetup()` 回调支持，可附加交互逻辑
- [x] `manager.go` — ModalWaiting/Bag/Cooldown/Guide/PullToRefresh 场景的 Setup 回调
- [x] `api.go` — 导出所有 Widget 类型 (GList/GButton/GLabel 等 22 个)

### Phase 3: Go 惯用 API 层
- [x] `options.go` — Options 模式（WithPosition/WithSize/WithAlpha/Hidden/Disabled 等 12 个 Option）
- [x] `options.go` — CreateButton/CreateLabel 工厂方法
- [x] `builder_api.go` — Builder 链式模式（ButtonBuilder 17 方法 + ListBuilder 13 方法）
- [x] `events.go` — 简化事件 API（ListenClick/ListenLink/ListenDrop 返回 CancelFunc）
- [x] `context.go` — Context 集成（LoadPackage/WaitForTransition/AdvanceWithContext）

### Phase 4: 性能优化 + 测试补全
- [x] `gears/display_test.go` — GearDisplay 4 个测试（Connected/IndexMatch/Lock/VisibleCounter）
- [x] `gears/display2_test.go` — GearDisplay2 3 个测试（Evaluate/NoController/OrCondition）
- [x] `gears/gears_test.go` — 8 个测试（Size/Color/Look/Text/Icon/Animation/FontSize）
- [x] `gears/xy_test.go` — GearXY 4 个测试（Apply/PercentMode/DefaultValue/UpdateFromRelations）
- [x] `core/benchmark_test.go` — 5 项基准测试（Object/Component/AddChild/GearApply/TickAll）

### Phase 5: 收尾
- [x] 清理 `draw_ebiten.go` 和 `scrollpane.go` 中注释掉的 Debug 代码
- [x] 全量编译通过
- [x] 6 个包测试全绿（core/gears/utils/audio/compat/text）

### 关键代码统计
- **新增文件**: 19 个（8 个 core + 1 个 utils + 4 个 fgui 根 + 3 个 gears 测试 + 1 个 benchmark + 2 个 docs）
- **修改文件**: 11 个（event/gobject/groot/scrollpane/uiconfig/controller/api/simple_scene/manager/audio/draw_ebiten/refactor-progress）
- **TS 模块实现率**: 56/72 → 64/72 (88.9%)

## 2025-11-09
### ✅ 已完成
- **彻底修复 ComboBox 下拉组件显示问题**：
  - **问题1 - 实例不匹配**：factory 设置和 ConstructExtension 调用发生在不同实例上
    - 添加 factory 字段到 GComboBox struct，使用 public SetFactoryInternal 方法
    - 修复 buildNestedComponent 模式，确保 dropdown 在同一实例上创建
  - **问题2 - 尺寸计算缺失**：dropdown 尺寸为 0x0，导致弹窗定位错误
    - 在 showDropdown 方法中添加正确的尺寸计算逻辑
    - 触发 list.GComponent.EnsureBoundsCorrect() 确保尺寸正确
    - 使用 ComboBox 宽度和 list 实际高度设置 dropdown 尺寸
  - **架构对齐 TypeScript**：Go 构造流程与 TypeScript constructExtension 模式完全一致
    - dropdown 创建和事件绑定在同一方法中完成
    - 使用 buildNestedComponent 替代手动创建 widget 实例
    - 确保 SetupAfterAdd 在正确实例上调用

### 测试结果
- ✅ TestComboBoxComponentAccess - 所有 ComboBox 组件访问正常
- ✅ n1 ComboBox - dropdown/list 正确设置，items 数量 8
- ✅ n4 ComboBox - 可见项目数 5
- ✅ n5 ComboBox - 可见项目数 10
- ✅ n6 ComboBox - 可见项目数 5
- ✅ TestShowDropdown - 下拉显示功能正常，列表项正确加载
- ✅ TestCheckPackageItems - 资源项查找和 FactoryObjectCreator 创建正常

### 关键代码修改
- `pkg/fgui/widgets/combo.go`:
  - 添加 factory 字段用于 ConstructExtension 中创建 dropdown
  - 添加 public SetFactoryInternal 方法（避免反射）
  - 修复 showDropdown 方法中的尺寸计算逻辑
- `pkg/fgui/builder/component.go`:
  - 修改 GComboBox case，使用 buildNestedComponent 模式
  - 确保 SetupAfterAdd 在正确实例上调用
- `pkg/fgui/builder/combobox_test.go`:
  - 完整测试覆盖 ComboBox 组件的构建和获取

### 技术成就
- **架构对齐**: 成功使 Go 构造流程与 TypeScript 模式一致
- **实例管理**: 解决了多实例环境下的工厂模式问题
- **尺寸计算**: 确保 UI 组件正确定位和显示
- **测试完全通过**: 所有 ComboBox 相关功能验证成功

