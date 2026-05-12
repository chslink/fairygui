# FairyGUI TypeScript → Go 移植进度对比报告

> 审计日期: 2026-05-12

## 一、总体概览

| 指标 | TypeScript 原始版 | Go 移植版 |
|------|-------------------|-----------|
| 源文件总数 | **72 个 .ts** | **117 个 .go**（含测试） |
| 总代码行数 | **~16,400 行** | **~27,000 行**（含 ~10,000 行测试） |
| 完整实现率 | — | **77.8%**（56/72） |
| 部分实现率 | — | **2.8%**（2/72） |
| 未实现率 | — | **19.4%**（14/72） |

---

## 二、逐模块对比

### 2.1 核心运行时 ✅ Complete

| TS 模块 (行数) | Go 实现 (行数) | 状态 |
|---|---|---|
| `GObject.ts` (996) | `core/gobject.go` (1619) | ✅ 完整，含位置/大小/可见性/关系/齿轮/SetupBeforeAdd |
| `GComponent.ts` (1055) | `core/gcomponent.go` (913) + `component_access.go` (50) | ✅ 完整，含子管理/排序/Overflow/Mask/Scroll |
| `GRoot.ts` (347) | `core/groot.go` (458) | ✅ 完整，含 Popup 栈/Ticker/Stage 绑定 |
| `Controller.ts` (236) | `core/controller.go` (263) | ✅ 完整，含页面切换/选择回调/InversePage |
| `Relations.ts` (135) | `core/relations.go` (790) | ✅ 完整，RelationItem + Relations 合并实现 |
| `RelationItem.ts` (566) | 同上 | ✅ 完整，含 Ext/百分比/锚点逻辑 |
| `ScrollPane.ts` (1516) | `core/scrollpane.go` (1395) | ✅ 完整，含惯性/回弹/循环/分页/裁剪 |
| `Transition.ts` (1101) | `core/transition*.go` (4 文件, 1296) | ✅ 完整，拆分为解析/运行时/路径三个子模块 |

### 2.2 Widget 控件 ✅ Complete

| TS 模块 (行数) | Go 实现 (行数) | 状态 |
|---|---|---|
| `GButton.ts` (454) | `widgets/button.go` (867) | ✅ Common/Check/Radio 模式、hover/pressed/disabled 状态 |
| `GComboBox.ts` (396) | `widgets/combo.go` (789) | ✅ 下拉组合框，含嵌套组件构建 |
| `GGraph.ts` (259) | `widgets/graph.go` (409) | ✅ 绘图命令（矩形/椭圆/线条/多边形） |
| `GGroup.ts` (373) | `widgets/group.go` (163) | ✅ 组布局（Go 实现更精简） |
| `GImage.ts` (114) | `widgets/image.go` (223) + `render/image.go` (498) | ✅ 含九宫格/平铺/FillMethod |
| `GLabel.ts` (172) | `widgets/label.go` (267) | ✅ 图标+标题组合 |
| `GList.ts` (2099) | `widgets/list.go` (1841) + `list_virtual*.go` (910) | ✅ 含虚拟列表/循环列表/多选/分页 |
| `GLoader.ts` (453) | `widgets/loader.go` (773) + `render/draw_loader.go` (541) | ✅ 含 FillMethod/Scale9/组件加载 |
| `GMovieClip.ts` (115) | `widgets/movieclip.go` (448) | ✅ 帧动画含摆动/循环/回调 |
| `GProgressBar.ts` (173) | `widgets/progressbar.go` (481) | ✅ 含正/反向动画/标题格式化 |
| `GRichTextField.ts` (25) | `widgets/rich_text.go` (97) | ✅ 富文本 |
| `GScrollBar.ts` (135) | `widgets/scrollbar.go` (426) | ✅ 含拖拽/缓动 |
| `GSlider.ts` (224) | `widgets/slider.go` (469) | ✅ 滑块 |
| `GTextField.ts` (254) | `widgets/text.go` (601) | ✅ 含 UBB/字体/对齐/AutoSize |
| `GTextInput.ts` (91) | `widgets/text_input.go` (850) | ✅ 含 IME/光标（剪贴板/撤销未实现） |
| `GTree.ts` (332) | `widgets/tree.go` (531) | ✅ 含展开/折叠/层级 |
| `GTreeNode.ts` (214) | `widgets/tree_node.go` (346) | ✅ 树节点 |

### 2.3 Gears 齿轮系统 ✅ Complete

| TS 模块 (行数) | Go 实现 (行数) | 状态 |
|---|---|---|
| `GearBase.ts` (111) | `gears/base.go` (176) | ✅ |
| `GearDisplay.ts` (36) | `gears/display.go` (171) | ✅ |
| `GearDisplay2.ts` (28) | `gears/display2.go` (91) | ✅ |
| `GearXY.ts` (118) | `gears/xy.go` (315) | ✅ |
| `GearSize.ts` (107) | `gears/size.go` (218) | ✅ |
| `GearLook.ts` (100) | `gears/look.go` (206) | ✅ |
| `GearColor.ts` (45) | `gears/color.go` (149) | ✅ |
| `GearAnimation.ts` (46) | `gears/animation.go` (196) | ✅ |
| `GearText.ts` (31) | `gears/text.go` (117) | ✅ |
| `GearIcon.ts` (31) | `gears/icon.go` (117) | ✅ |
| `GearFontSize.ts` (31) | `gears/fontsize.go` (129) | ✅ |

> ⚠️ gears 包 11 个文件、0 个测试文件

### 2.4 Tween 补间引擎 ✅ Complete

| TS 模块 (行数) | Go 实现 (行数) | 状态 |
|---|---|---|
| `EaseManager.ts` (168) | `tween/ease.go` (255) | ✅ 含全部缓动函数 |
| `EaseType.ts` (36) | 同上（内嵌） | ✅ |
| `GTween.ts` (37) | `tween/tween.go` (909) | ✅ 合并为单文件 |
| `GTweener.ts` (438) | 同上 | ✅ |
| `TweenManager.ts` (106) | 同上 | ✅ |
| `TweenValue.ts` (55) | 同上 | ✅ |
| `GPath.ts` (250) | `core/transition_path.go` (218) | ✅ 路径插值移入过渡系统 |
| `GPathPoint.ts` (73) | `core/transition.go` (结构体) | ✅ |

### 2.5 Assets 资源系统 ✅ Complete

| TS 模块 (行数) | Go 实现 (行数) | 状态 |
|---|---|---|
| `UIPackage.ts` (698) | `assets/package.go` (609) + `registry.go` (114) | ✅ 含注册中心/全局查询 |
| `PackageItem.ts` (60) | `assets/types.go` (233) + `package.go` | ✅ |
| `display/Image.ts` (206) | `render/image.go` (498) | ✅ 逻辑融入渲染层 |
| `display/MovieClip.ts` (261) | `widgets/movieclip.go` (448) | ✅ |
| `display/FillUtils.ts` (322) | `render/draw_loader.go` + `texture_renderer.go` | ✅ |

### 2.6 Utils 工具 ✅ 大部分完整

| TS 模块 (行数) | Go 实现 (行数) | 状态 |
|---|---|---|
| `ByteBuffer.ts` (114) | `utils/bytebuffer.go` (242) | ✅ |
| `ColorMatrix.ts` (118) | `compat/laya/colorfilter.go` (95) | ✅ |
| `PixelHitTest.ts` (50) | `assets/types.go` + `render/pixel.go` | ✅ |
| `UBBParser.ts` (4) | `internal/text/ubb.go` | ✅ |
| `UIConfig.ts` (61) | `core/uiconfig.go` (70) | ✅ |
| `Events.ts` (31) | `compat/laya/event.go` (153) | ✅ |
| `Margin.ts` (14) | `core/overflow.go` | ✅ |

---

## 三、未实现的模块（14 个，19.4%）

### 3.1 完全缺失的核心功能

| TS 模块 | 行数 | 影响 | 说明 |
|---------|------|------|------|
| **Window.ts** | 235 | 🔴 高 | 窗口系统（模态/拖拽区域/关闭按钮/X/Y 约束）。Demo 中使用手动 `windowInstance` 替代 |
| **PopupMenu.ts** | 149 | 🔴 高 | 右键弹出菜单。仅 `uiconfig.go` 有 URL 配置，无实际实现 |
| **DragDropManager.ts** | 62 | 🟡 中 | 全局拖拽管理器。影响拖拽功能 |
| **TranslationHelper.ts** | 203 | 🟡 中 | 多语言翻译辅助。影响国际化 |
| **UIObjectFactory.ts** | 92 | 🟡 中 | 已实现基础 (`widgets/factory.go`) 但不支持 `setExtension()` 自定义类型注册 |
| **IUISource.ts** | 7 | 🟡 中 | UI 源接口定义，Window 依赖此接口 |
| **AsyncOperation.ts** | 171 | 🟢 低 | 异步操作抽象。Go 可用 goroutine/channel 替代 |
| **AssetProxy.ts** | 20 | 🟢 低 | 资源代理系统 |
| **GLoader3D.ts** | 356 | 🟢 低 | 3D 加载器，依赖 LayaAir 3D 引擎 |
| **FieldTypes.ts** | 185 | 🟢 低 | 枚举常量以 `const` 分散在各模块中 |

### 3.2 完全缺失的 action/ 目录

| TS 模块 | 行数 | 说明 |
|---------|------|------|
| `action/ControllerAction.ts` | 40 | 控制器动作基类 |
| `action/ChangePageAction.ts` | 39 | 切换页面动作 |
| `action/PlayTransitionAction.ts` | 35 | 播放过渡动画动作 |

> 这些定义了控制器页面切换时的关联行为（如自动播放过渡动画）。Go 的 Controller 只实现了基础页面选择与回调，不含动作链系统。

### 3.3 缺失的工具类

| TS 模块 | 行数 | 说明 |
|---------|------|------|
| `utils/ToolSet.ts` | 181 | 剪贴板/颜色转换等工具函数 |
| `utils/ChildHitArea.ts` | 26 | 子碰撞区域工具（`render/hit_area.go` 有替代方案） |

---

## 四、兼容层映射

`internal/compat/laya/` 不是 TS 源代码的直接翻译，而是 **LayaAir 引擎运行时的 Go 替代**：

| Go 文件 | 行数 | 替代的 LayaAir 概念 |
|---------|------|---------------------|
| `sprite.go` | 558 | Laya.Sprite - 显示对象/变换/层级 |
| `graphics.go` | 714 | Laya.Graphics - 绘图命令系统 |
| `stage.go` | 364 | Laya.Stage - 舞台/输入处理/调度 |
| `event.go` | 153 | Laya.Event - 事件分发/冒泡 |
| `timer.go` | 201 | Laya.Timer/Scheduler |
| `colorfilter.go` | 95 | 颜色矩阵/滤镜 |
| `matrix.go` | 52 | Laya.Matrix - 2D 仿射变换 |
| `input.go` | 66 | 输入状态管理（Mouse/Touch/Keyboard） |
| `geometry.go` | 35 | Point/Rect 几何类型 |

---

## 五、Go 新增功能（TS 版本无对应）

| Go 模块 | 功能 | 说明 |
|---------|------|------|
| `audio/audio.go` | 音频播放器 | 支持 WAV/MP3/Ogg，Ebiten 音频 |
| `debug/inspector.go` | 组件检视器 | 运行时组件树查看 |
| `debug/server.go` | 调试服务器 | HTTP 调试接口 |
| `debug/simulator.go` | 设备模拟器 | 不同分辨率模拟 |
| `render/atlas_ebiten.go` | Ebiten Atlas 管理 | 纹理图集加载与缓存 |
| `render/color_effects.go` | 颜色效果统一管线 | 灰度/颜色矩阵/BlendMode |
| `render/text_draw.go` | 文本渲染管线 | 系统字体/位图字体/UBB |
| `render/graphics_draw.go` | Graphics 命令渲染 | 消费 Sprite.Graphics 命令 |
| `assets/font.go` | 位图字体支持 | .fnt 解析与渲染 |
| `assets/fs_loader.go` | 文件系统加载器 | 从本地文件加载 .fui |

---

## 六、代码行数对比

| 分类 | TS 行数 | Go 行数 | 比例 |
|------|---------|---------|------|
| 核心运行时 | ~5,500 | ~7,100 | 1.29x |
| Widgets | ~5,000 | ~8,800 | 1.76x |
| Gears | ~650 | ~1,900 | 2.92x |
| Tween | ~1,160 | ~1,160 | 1.00x |
| Assets | ~1,080 | ~1,300 | 1.20x |
| Utils | ~490 | ~240 | 0.49x |
| 缺失模块 | ~3,520 | 0 | — |
| **总计** | **~16,400** | **~20,500** | **1.25x** |

> Go 代码行数更高的原因：(1) Go 不含类继承，需要用组合和接口；(2) 手写错误处理增加了行数；(3) 部分 Widget 实现比 TS 更详细（如 GTextInput 850 行 vs 91 行）。

---

## 七、总结

**已实现**: 56/72 模块（77.8%），所有核心 Widget、Gears、Tween、Relations、ScrollPane、Transition、资源加载均完整可用。

**未实现**: 14 个模块（19.4%），主要是：
1. **action/ 动作系统**（3 文件）- 控制器关联行为
2. **Window 系统及依赖**（Window, IUISource, PopupMenu, DragDropManager）
3. **辅助工具**（TranslationHelper, AssetProxy, AsyncOperation, ToolSet）
4. **3D 功能**（GLoader3D）
5. **UIObjectFactory 扩展机制**

**建议优先实现**:
- Window 窗口系统（P1 - Demo 中已有手动替代，正式化即可）
- PopupMenu 右键菜单（P1 - 影响交互体验）
- Controller Action 链（P2 - 完善控制器功能）
- UIObjectFactory.setExtension（P2 - 可扩展性）
