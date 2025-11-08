# FairyGUI Ebiten Port Progress

## 2025-11-08
### ✅ 已完成
- 修复滚动条拖动时触发容器拖动的问题（通过实现stopPropagation机制）
- 修复Grid demo中星星不显示的问题（ProgressBar组件解析）
- 修复ObjectType枚举不匹配问题（Go:13 vs TypeScript:14）
- **修复ProgressBar渲染层面问题**：
  - 修复ProgressBar尺寸显示为0x0的问题
  - 实现模板组件尺寸继承机制（ensureSizeFromTemplate）
  - 在SetupAfterAdd中确保尺寸正确
  - 所有ProgressBar组件现在都有正确尺寸：145x12, 105x21, 277x32等
- **修复ProgressBar填充机制**：
  - 实现TypeScript风格的setFillAmount方法
  - 优先使用fillAmount（适用于GImage/GLoader），失败时回退到修改width/height
  - 修正GLoader的fillAmount范围为0-100（与TypeScript一致）
  - 在GImage和GLoader的SetFill/SetFillAmount中添加重绘触发

### 测试结果
- ✅ TestProgressBarComponentParsing - 所有ProgressBar组件解析正确
- ✅ TestProgressBarStarComponent - Grid中使用的ProgressBar组件正确
- ✅ TestProgressBarValueClamp - 值范围正确
- ✅ TestProgressBarFill - 填充渲染正确
- ✅ TestProgressBarTitleFormats - 标题格式化正确

### 关键代码修改
- `pkg/fgui/widgets/progressbar.go`:
  - 新增 `ensureSizeFromTemplate()` 方法
  - 在 `SetupAfterAdd()` 末尾调用尺寸检查
- `pkg/fgui/builder/progressbar_test.go`:
  - 添加尺寸验证测试

## 2025-10-18
- [x] Audited `laya_src/fairygui` TypeScript modules and catalogued LayaAir dependencies.
- [x] Authored `docs/architecture.md` describing Go/Ebiten layering, package mapping, and migration phases.
- [x] Bootstrapped Go scaffolding (`internal/compat/laya`, `pkg/fgui/core`) with geometry, events, scheduler, and base `GObject`/`GComponent` containers.
- [x] Expanded the compat sprite with transform state, affine matrix math, global bounds, and introduced a stage abstraction with scheduler/input integration.
- [x] Added foundational unit tests covering sprite coordinate transforms, stage mouse routing, and `GObject` size/position propagation.
- [x] Ported `fgui.utils.ByteBuffer` to Go with full string-table, colour, sub-buffer, and seek behaviours plus unit coverage.
- [x] Enhanced the stage/input layer with hit testing, pointer bubbling, click synthesis, and regression tests.
- [x] Introduced shared test utilities (stage env, event logs) and expanded coverage for scheduler and `GComponent` behaviours.
- [x] Bootstrapped Go asset pipeline scaffolding: resource loader abstraction, package header parsing, ByteBuffer enhancements, and parsing for package items, atlas sprites, and pixel hit-test metadata with unit tests.
- [x] Added raw DEFLATE decompression support, filesystem loader, verified parsing against demo `.fui` packages, and introduced Ebiten-tagged atlas manager plus pixel hit-test integration hooks.
- [x] Began component instantiation path: parsed component metadata now exposes structured child descriptors, builder scaffolding creates GObject trees from real `.fui` data, widgets bind sprite/text content, and component controllers are parsed and attached for runtime use.

## 2025-10-19
- [x] Enriched `core.GObject` with scale, rotation, and pivot state mirrored to compat sprites, enabling downstream systems to track transforms without poking display objects.
- [x] Applied component metadata transforms (scale, rotation, pivot, alpha) during factory builds so instantiated hierarchies better reflect original FairyGUI layouts.
- [x] Added focused unit tests covering the new geometry plumbing (`pkg/fgui/core`, `pkg/fgui/builder`) using `GOCACHE=$(pwd)/.gocache go test ./pkg/fgui/core ./pkg/fgui/builder`.
- [x] Introduced skew handling, pivot-anchor positioning, and cross-package asset resolution in the builder, alongside regression tests that exercise demo `.fui` dependencies.
- [x] Added compat sprite regression tests validating pivot/倾斜矩阵运算与锚点偏移，涵盖缩放、旋转、移动、尺寸变更场景。
- [x] Builder 现会将按钮、标签、列表等高级控件解析成对应 widget（携带包引用、默认项、图标资源），并在渲染阶段绘制文本及按钮图标。

### Upcoming Focus
- Wire parsed atlas sprites into texture loaders and expose hit-test data to rendering/input layers.
- Expand widget factories beyond image/text/button/loader and honor controller/gear transitions during instantiation.
- Wire atlas sprites and pixel masks into real rendering passes under Ebiten.
- Implement concrete loaders (filesystem/embedded) and integrate with Ebiten-friendly texture creation.
- Connect pointer events to higher-level UI abstractions (GRoot, drag/drop) leveraging the new compat stage.
- Profile pivot-aware transforms against upstream FairyGUI scenes and tune any drift discovered during animation playback.

## 2025-10-21
- [x] 完整移植 `core.Relations`/`RelationItem`，实现目标坐标、尺寸变动通知及百分比/Ext 逻辑，对齐 TS 行为并新增 `pkg/fgui/core/relations_test.go` 覆盖常见布局联动。
- [x] 扩展 gear 栈：`pkg/fgui/gears` 新增 Color/Animation/Text/Icon/Display2/FontSize 与 tween 配置，`GObject` 暴露 `GetProp/SetProp` 以串联 widgets 属性代理。
- [x] `GObject`/`GComponent` 补齐 gear 锁定、控制器切页回放与 relations → gear 反向刷新，新增 `pkg/fgui/core/gobject_gears_test.go` 验证多页同步。
- [x] Builder 解析 `.fui` 的 relations/gear 块，注册包依赖、推导资源目录，自动套用 controller 默认页并补写按钮、Loader 资源/图标解析。
- [x] Widgets 拓展按钮、文本、Loader 属性接口（标题、图标、颜色、字体尺寸、动画播放等），配合 gear/Builder 提供统一代理；补充 Loader 布局与填充单测。
- [x] Ebiten demo 接入 `core.GRoot` 与 compat Stage，`Game.Update` 推进 GTween/Scheduler 并同步鼠标状态，确保 tween 在帧循环中更新。
- [x] Builder 支持 `setup_afterAdd` 通用字段：解析 tooltip、group、controller 默认页及属性赋值路径，并在 `GObject` 暴露相应状态；补充 Transition Demo 分组测试。

### Upcoming Focus（2025-10-21）
1. 实现 `setup_afterAdd` 剩余属性赋值及 Loader 外链/嵌套组件加载逻辑，补齐控件默认态。
2. 将 GearColor/Icon/Text/FontSize 的变更反射到 `pkg/fgui/render` 与 demo 渲染路径，打通实际颜色与文本缓存。
3. 扩充 Controller/Gear 集成测试矩阵，并在 GUI 环境运行 `go test ./pkg/fgui/...`、`go run ./demo` 做回归。
4. 梳理 GRoot/拖拽/ScrollPane 依赖，规划下一阶段组件与交互移植。

## 2025-10-22
- [x] Builder 解析 `constructFromResource` Transition 区块，`GComponent` 缓存 Transition 元数据并对 Transition Demo 组件做验证覆盖。
- [x] 逐条解析 Transition item，收集 tween 配置、路径与目标引用，为后续动画回放提供完整元数据。
- [x] Transition 运行时执行涵盖 XY/尺寸/缩放外，新增抖动、颜色滤镜、声音与嵌套 Transition 播放支持，基于 GTween 推进并在 `GObject` 上落地效果。
- [x] 暴露 `core.Transition` 运行时封装，挂接 GTween 延迟调度并提供播放/停止接口，便于后续接入 Ebiten 帧循环。

### Upcoming Focus（2025-10-22）
1. 将运行时 Transition 绑定真实 item 执行路径，驱动 GObject 属性与 GTween 曲线同步。
2. 针对自动播放与嵌套 Transition 的行为添加单元测试，确保 demo 资产可回放。
3. 梳理 Transition 与 Controller/Gear 的交互路径，评估需要的事件/状态同步。
4. 设计 Transition 回放 API（播放/暂停/停止）与事件回调，串联未来的可视化调试工具。

## 2025-10-23
- [x] Transition 运行时现支持 ActionType.Animation：`GObject.SetProp` 代理至 widget 播放/帧接口，动画动作可切换播放状态与帧号并兼容 frame=-1 哨兵；新增单测覆盖暂停→恢复场景。
- [x] Pivot 动作在回放时保留 `pivotAsAnchor` 状态，避免 Transition 改动后锚点模式丢失。
- [x] Path 补间接入：解析 `TransitionTween.Path` 构造 Catmull-Rom/Bezier 路径并驱动 GTweener，XY 动作现可沿路径移动并通过单测验证中点与终点位置。
- [x] Transition timeScale/DeltaTime 状态与 GearAnimation 同步：暴露 `Transition.SetTimeScale`、回放时刷新 `ObjectPropIDTimeScale/DeltaTime`，并扩展 GearAnimation 存档与应用逻辑以保持跨页一致。
- [x] Component 命中区域接入：Builder 将 mask / pixel hit-test 元数据传递到 compat Sprite，自定义 hit tester 支持像素遮罩、子节点遮罩与反向遮罩；新增 `render` 单测验证命中/失效场景。
- [x] Touchable 标识下沉：`GObject.SetTouchable` 现同步 compat Sprite `mouseEnabled`，命中流程在非触摸组件上自动透传；补充 Sprite/GObject 单测确保父级禁用仍可命中子节点。
- [x] Button 事件流：`GButton` 绑定 compat 鼠标事件，支持 hover/down/click 状态切换、linked popup 切换与选择模式同步，新增 StageEnv 驱动单测。
- [x] List 基础交互：`GList` 提供 `AddItem`/`SetSelectedIndex` 并挂接子项点击事件，可驱动单选高亮；测试用 StageEnv 验证点击选择。
- [x] List ↔ Controller 联动：`core.Controller` 增加选择监听与 `-1` 支持，`GList.SetSelectionController` 同步 controller、列表状态并新增单测覆盖绑定与解绑流程。
- [x] List 多选模式：扩展 `ListSelectionMode` 枚举，新增多选/单击多选/禁用模式，支持编程式批量选中、索引查询与按钮选中态同步，补充多场景单测。
- [x] Controller 默认页修正：初始化/越界选择时自动回落至首页，避免 button 列表因 controller 为 -1 而整页隐藏。
- [x] Demo 调试输出：`demo/main.go` 每隔数秒打印当前场景、组件层级和按钮状态，便于排查渲染缺失与 controller 状态问题。
- [x] GImage 渲染落地：`GImage` 现保存包内资源并走 Ebiten draw 流程（含 Scale9Grid 支持），主菜单按钮背景可正确拉伸显示。
- [x] Atlas 精度防护：`AtlasManager.ResolveSprite` 校验裁剪区域，防止 0 尺寸或越界导致 Ebiten panic，便于后续兼容不同裁剪格式。
- [x] 新增 `cmd/nineslice-demo` 调试程序，可独立加载 `Bag/btnimage` 并用方向键实时调整尺寸验证九宫格渲染。
- [x] `FGUI_DEBUG_NINESLICE=1` 时输出九宫格渲染参数（目标尺寸、分段、九宫格切块），辅助定位 scale9 变形问题。
- [x] 新增 `FGUI_DEBUG_NINESLICE_OVERLAY=1` overlay，可在运行时绘制九宫格中心/边界线，肉眼观察拉伸区域。

### Upcoming Focus（2025-10-23）
1. 补齐 `GList` 多选场景下的 SHIFT/CTRL 输入、区间选择与虚拟列表适配，规划跨包滚动交互方案。
2. 将列表选中状态与 Button/Item 可视状态、gear 同步，补充渲染与交互测试矩阵。
3. 将 mask 与 pixel hit-test 贯穿到 Ebiten 渲染阶段，实现遮罩裁剪与反向遮罩绘制。

## 2025-10-24
- [x] `widgets.GTextField` 补齐字号、颜色、字距、行距、水平/垂直对齐、描边占位等样式字段，`drawTextImage` 现按样式生成缓存贴图并复用系统字体缓存。
- [x] 系统字体加载新增缓存索引，同字体多字号复用 `font.Face`，避免重复解析并方便按钮/标签共享字体。
- [x] Builder `applyTextFieldSettings`/`applyLabelTemplate` 统一处理按钮标题与模板文本，保证 Transition Demo 内按钮标题样式正确落地。
- [x] `widgets.GButton` 新增 `updateVisualState`，根据 hover/pressed/disabled/selected 切页以驱动控制器状态，恢复按钮悬浮与按下视觉反馈。
- [x] 建立 BitmapFont 渲染管线：解析 `.fnt` 包内字体，注册到渲染字体表，`drawBitmapFont` 负责字形拼接与对齐，Demo 中数字字体可正确显示。
- [x] 清理 Demo 层级的临时文本调试输出，确保只通过 FairyGUI 文本系统渲染。
- [x] ScrollPane 栈初步落地：`GComponent` 拆分容器并接入 `core.ScrollPane`，兼容 Laya scrollRect 裁剪与滚轮输入，Builder 解析 overflow=Scroll/Hidden 配置并同步视口、内容尺寸。
- [x] 新增 `widgets.GProgressBar/GSlider/GScrollBar`，Builder 解析模板与实例数据并绑定 ScrollPane；滚动条同步滚动监听，滑杆/进度条支持标题样式与拖拽交互。

### Upcoming Focus（2025-10-24）
1. 修复 Transition Demo「Play5」内 BitmapFont 与系统字体重复渲染问题，排查模板/控制器是否触发双重绘制。
2. 调整 Transition 动画起始帧逻辑，解决首次播放时特效贴图延迟出现的情况，并补充保护性测试。
3. 为 `drawTextImage` 与 `drawBitmapFont` 编写表驱动单测，覆盖对齐、字距/行距、多字号与中英文混排。
4. 扩展按钮状态映射，兼容 FairyGUI 项目中自定义命名页（如 `down2/over2`），并回归主要场景。
5. 补齐 ScrollPane 拖拽/分页/滚动条交互，并与 List、虚拟列表联动验证。

## 2025-10-25
- [x] 回归 Transition Demo，确认按钮点击/状态切换逻辑正常触发，定位 Play5 场景位图字体重复与首帧贴图延迟的剩余问题。
- [x] 回归 Basics 场景，验证新接入的 ProgressBar/Slider/ScrollBar，记录按钮模板尺寸异常与子 Demo 尚未实现的差距。
- [x] 梳理现有控件单测与 Demo 输出来匹配 TS 行为，为后续排期整理优先级。

### Upcoming Focus（2025-10-25）
1. 优先修复 Basics 场景按钮模板尺寸/布局，补齐各子 Demo（窗口、弹出、拖拽等）交互以对齐 Laya 版本。
2. 排查 Transition.Play5 位图字体重复绘制与初始贴图延迟，补充 BitmapFont 渲染表驱动单测覆盖字距/对齐。
3. 扩展按钮/控制器状态映射，支持自定义命名页（down2/over2 等），并在主 Demo 场景回归验证。
4. 强化 ScrollPane 拖拽、分页、滚动条联动测试，覆盖 List/VirtualList 等组合场景。
5. 跟进 `GTextField` 高级样式（描边、阴影、UBB）与富文本渲染计划，规划拆解与依赖。

## 2025-10-26
- [x] 扩展 `widgets.GTextField` 字段，新增阴影参数缓存与访问器，Builder 解析 shadow 配置并下发；补齐模板到渲染完整链路。
- [x] 新增 `internal/text.ParseUBB`，实现颜色/字体/字号/粗斜体/下划线/url 等基础标签解析，输出段落样式切片。
- [x] 重写 Ebiten 渲染侧文本管线，支持多段字体/字号混排、描边、阴影、加粗、下划线，系统/位图字体共存并按对齐、行距、字距布局。
- [x] 修订文本贴图尺寸计算，引入效果 padding，避免描边与阴影被裁剪；按钮/标签文本现按 FairyGUI 样式绘制。
- [x] 添加 UBB 解析表驱动单测及文字布局单测（`-tags ebiten`），保障跨段落 letterSpacing 计算与换行拆分。
- [x] 打通 `GTextField` AutoSize 流程：渲染测量结果回写 widget，Both/Height 模式自动刷新 `GObject` 尺寸并暴露 `TextWidth/TextHeight`。

### Upcoming Focus（2025-10-26）
1. 完善 Italic 倾斜与字体描边厚度映射，实现与 TS 版本一致的倾斜矩阵与 stroke 扩散模型。
2. 补齐 GTextField AutoSize 行为（尺寸回写 GObject）、UBB 高级标签（img/url callback）与 RichText 交互。
3. 结合 Demo 场景回归验证阴影/描边效果，排查 Basics/Transition 场景中文本重复或首帧延迟残留问题。
4. 将新文本渲染逻辑纳入性能/缓存评估，补写缓存命中单测与位图字体 Tint 适配方案。
5. 在具备 Ebiten/GLFW 依赖环境下回归 `go test -tags ebiten ./pkg/fgui/render`，确保渲染侧单测长期可执行。

## 2025-10-27
- [x] 对照 `laya_src/demo` 审视 `demo/scenes` 现状：`MainMenu` 导航与 `TransitionDemo` 基本对齐，`BasicsDemo` 仅覆盖按钮/文本/网格/进度条子示例，其余场景仍停留在静态组件加载。
- [x] 列出 Demo 行为缺口，涉及虚拟/循环列表渲染、下拉刷新回调、聊天与表情解析、背包窗口与物品随机化、摇杆模块、引导遮罩、冷却条 Tween 等，形成移植清单。
- [x] 移植 Joystick Demo：实现 `demo/scenes/joystick.go` 与舞台触控事件，模拟 TS 版摇杆半径约束、回弹 tween 与角度广播。
- [x] Basics Demo 恢复窗口/弹窗/拖拽三类子示例：补写简易窗口管理（含缩放动画、关闭按钮）、构建临时 Popup 菜单、实现局部拖拽/落点判定与 Stage 级监听。

### Upcoming Focus（2025-10-27）
1. 扩充 `demo/scenes/basics.go`，实现窗口、弹窗、拖拽、深度等子示例逻辑，并移植 `TestWin` 窗体与 Popup 菜单。
2. 为 `VirtualList`、`LoopList`、`ListEffect`、`ScrollPane` 场景接入虚拟列表、循环滚动、动效触发与渲染器（需移植 `MailItem`、`ScrollPaneHeader` 等扩展组件）。
3. 移植交互类场景：聊天消息/表情（`EmojiParser`）、背包窗口、引导遮罩定位与动画、冷却条 Tween 行为等。
4. 梳理触控/拖拽/全局事件所需的 compat 能力缺口，补齐 `pkg/fgui` 端事件派发、定时器与 Transition hook，确保 Demo 行为可复刻。

## 2025-10-28
- [x] 在 `internal/compat/laya` 新增 `Graphics` 与 `HitArea`，`Sprite` 现保存绘图命令并同步命中测试，完整对齐 Laya `DisplayObject.graphics` 行为。
- [x] `widgets.GGraph` 迁回 Laya 风格 API：恢复 `DrawRect/DrawEllipse/DrawPolygon/DrawRegularPolygon`，`updateGraph` 直接写入 compat `Graphics` 并刷新 `HitArea`。
- [x] 渲染层改为消费 Sprite 上的 Graphics 命令，统一处理圆角矩形、椭圆、多边形绘制，移除原集中式 `renderGraph` 临时图片构建。
- [x] `core.GObject` 在尺寸变化时回调宿主，实现 Graph 等组件在 `SetSize` 之后自动重绘。
- [x] `widgets.GImage` 与 Loader 贴图绘制迁移到 `Sprite.Graphics.DrawTexture`，支持颜色覆盖、九宫格与平铺命令，Ebiten 渲染层解析纹理指令并沿用 Laya 变换/轴心逻辑。

## 2025-10-29
- [x] 渲染路径统一向 `drawPackageItem`、`drawNineSlice`、`renderImageWithGeo` 传入 `*laya.Sprite`，由 `applyTintColor/applyColorEffects` 套用 Sprite 的颜色滤镜、灰度与 `BlendMode`。
- [x] `renderLoader`、`renderGraph` 等代码同步改造，Loader 九宫格/填充与矢量图形在最终绘制阶段都会尊重颜色矩阵和混合模式设置。
- [x] `applyTintColor` 聚合 Alpha 缩放与色彩覆盖逻辑，避免重复缩放，同时补充文档记录新的渲染数据流。
- [x] 新增 `GTree`/`GTreeNode`，扩展 `GList` 支持节点插入、删除和事件绑定；Builder 解析 `treeView="true"` 组件时构建层级、模板并恢复节点文本/图标/控制器，与 TS 行为对齐。

## 2025-10-30
- [x] `assets.PackageItem` 增加 MovieClip 元数据 (`Interval`/`RepeatDelay`/`Swing`/`Frames`)，`parseMovieClipData` 镜像 TS `loadMovieClip` 解析流程并附带单测覆盖真实 `Basics.fui`。
- [x] `core.GRoot.Advance` 引入 `tickAll`，对外暴露 `core.RegisterTicker(func(delta))`，让 MovieClip、Tweener 等模块能获取逐帧时间片而不依赖 compat Scheduler 任务堆。
- [x] 新增 `widgets.GMovieClip`：支持播放控制、`SetPlaySettings`、`SyncStatus`、时间缩放与颜色覆盖，挂接 `core.RegisterTicker` 自动推进帧序列并实现 `setup_beforeAdd` 行为，对应单测覆盖 Advance/TimeScale/EndHandler/SetupBeforeAdd 场景。
- [x] `render.DrawComponent` 处理 `GMovieClip`，`AtlasManager.ResolveMovieClipFrame` 根据帧级 atlas sprite 裁剪贴图并缓存；渲染阶段按帧偏移/原始尺寸缩放输出，保持 Trim 图对齐。
- [x] compat Stage 输入系统扩展：支持多指触控、键盘事件与 pointer capture/focus API，新增 `InputState`/`TouchInput`/`KeyboardEvent`，并在测试环境验证触控/键盘/捕获流程。
- [x] Graphics 命令渲染：实现 `drawLine`/`drawPie` 等命令并在 Ebiten 渲染层解析 `Sprite.Graphics`，为通用 Sprite 图形绘制提供 fallback 支持。
- [x] 同步 `docs/architecture.md` 与进度日志，记录新的 ticker 通道与 MovieClip 渲染架构。

### Upcoming Focus（2025-10-30）
1. 覆盖 MovieClip `flip`/`fill` 相关逻辑，确认 Loader/Transition 对 MovieClip 的 gear 与动画指令是否需要额外对齐。
2. 在 demo Basics/Basics 的 MovieClip 场景验证翻转、重复延迟与颜色覆盖是否与 Laya 一致（需 GUI 环境运行反馈）。
3. 为 `AtlasManager` 增补帧级缓存驱逐与九宫格/平铺路径共享逻辑，避免重复裁剪造成 GPU 纹理副本激增。

### Architectural Refactoring (2025-10-30 下午)
- [x] **完全重构 SetupBeforeAdd 架构**：将所有 widget 的属性设置统一为 `SetupBeforeAdd(buf, beginPos)` 单一入口，完全对齐 TypeScript 版本。
  - Phase 1: 实现 `GObject.SetupBeforeAdd` 逐行对应 TS，添加完整的继承链调用
  - Phase 2: 迁移所有 11 个 widget (GImage/GTextField/GButton/GList/GTree/GMovieClip/GGraph/GGroup/GLoader 等)
  - Phase 3: 重构 Builder，移除所有 `ApplyComponentChild` 调用和冗余属性设置器（~300 行代码简化）
  - Phase 4: 测试验证，所有使用真实 .fui 文件的测试通过
- [x] **Builder 代码大幅简化**：引入 `callSetupBeforeAdd`/`callSetupAfterAdd` 辅助函数消除重复，移除手动属性设置器（SetText/SetTitle/SetIcon/SetURL/SetResource/SetDefaultItem 等），保留核心逻辑（模板构建、对象创建器、滚动条设置）。
- [x] **API 提升到顶层包**：将 `builder.Factory` 提升为 `fgui.Factory`，作为主 API 入口。
  - 创建 `pkg/fgui/api.go` 统一 API 门面
  - 导出所有关键类型（Core/Assets/Builder/Constants）
  - 添加便捷函数（NewFactory/NewFactoryWithLoader/ParsePackage/NewFileLoader）
  - 完全向后兼容（使用 Go 类型别名，零开销）
  - 更新 demo 和文档使用新 API
- [x] **文档完善**：创建 API 迁移指南 (`docs/api-migration.md`)，更新 `CLAUDE.md` 架构说明，添加完整示例代码。

**架构收益**：
- ✅ 消除了旧架构的双重属性设置问题（ApplyComponentChild + 手动设置 + SetupBeforeAdd）
- ✅ 100% 对齐 TypeScript 版本的单一数据源架构
- ✅ Builder 代码减少 ~300 行，可维护性大幅提升
- ✅ 提供清晰的 API 分层（内部实现 vs 公开门面），避免循环依赖
- ✅ 简化导入，单一 `import "pkg/fgui"` 即可访问所有功能

### Upcoming Focus（2025-10-30 晚）
1. 修复 widgets 包剩余的 2 个测试失败（TestListSelectionOnClick/TestListSetVirtualPreservesEventListeners），排查鼠标事件处理逻辑。
2. 修复 render 包测试失败（TestRenderMovieClipWidgetUsesSourceSize），检查 MovieClip 源尺寸计算。
3. 在 demo/scenes 中验证所有场景使用新 API 后的行为一致性。
4. 考虑是否需要为 `fgui` 包添加更多便捷函数（如 CreateObject 等）以进一步简化 API。


## 2025-11-04
- [x] **完整实现 Overflow 功能**：对齐 TypeScript 版本的 overflow 行为，支持 Visible/Hidden/Scroll 三种模式。
  - 创建 `pkg/fgui/core/overflow.go`：定义 OverflowType 类型别名与 Margin 结构体（通过类型别名复用 assets.OverflowType 避免重复定义）
  - 扩展 GComponent：添加 margin/overflow 字段与访问器，实现 SetupOverflow/UpdateMask/SetMargin 方法
  - OverflowHidden：创建独立 container Sprite，设置 scrollRect 实现内容裁剪，支持 margin 偏移
  - OverflowVisible with Margin：创建独立 container 应用偏移，但不设置 scrollRect
  - OverflowScroll：委托给已有的 SetupScroll 方法（ScrollPane）
  - SetSize 集成：尺寸变化时自动调用 UpdateMask 更新裁剪区域
- [x] Builder 集成：`BuildComponent` 从 ComponentData 读取 margin 和 overflow，自动调用 SetupOverflow/SetupScroll
- [x] **完整测试覆盖**：
  - 单元测试 (`pkg/fgui/core/overflow_test.go`)：8 个测试全部通过
    - 验证 OverflowHidden 创建独立 container
    - 验证 scrollRect 尺寸与偏移正确性
    - 验证 margin 应用与容器偏移
    - 验证 OverflowVisible 行为
    - 验证 SetSize 触发 UpdateMask
    - 验证 Margin.IsZero 和访问器方法
  - 集成测试 (`pkg/fgui/builder/overflow_test.go`)：2 个测试通过
    - TestOverflowFromPackage：验证从真实 .fui 文件加载 overflow 配置（Component1/Component7/Component8）
    - TestOverflowBuildIntegration：手工构造测试数据验证完整构建流程
- [x] 文档更新：创建 `docs/overflow-investigation.md` 详细记录调研过程、TypeScript 参考实现、Go 实现方案与测试计划

**实现细节**：
- 使用 laya.Rect 替代 Rectangle（字段为 X/Y/W/H 而非 Width/Height）
- 通过 `display.SetScrollRect(rect)` 实现裁剪（laya 兼容层已支持）
- Container 创建时机：overflow=hidden 或 margin 非零时创建
- Margin 应用：container 偏移 (left, top)，scrollRect 起点也在 (left, top)，尺寸为 width-right, height-bottom

### Upcoming Focus（2025-11-04）
1. 在 GUI 环境运行 `go run ./demo`，验证 Demo_Clip&Scroll 场景的 overflow 视觉效果
2. 检查其他场景是否有 overflow 相关问题（如滚动容器、裁剪区域等）
3. 如需要，补充 overflow 与 ScrollPane 的集成测试

