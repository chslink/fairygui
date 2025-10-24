# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 语言规范

**全程使用中文进行问答和交流**。代码注释和文档可以使用英文或中文，但与 AI 助手的交互必须使用中文。

## 项目概述

这是一个将 FairyGUI（原基于 LayaAir/TypeScript）移植到 Go + Ebiten 游戏引擎的项目。目标是在保持公开 `fgui` API 兼容性的同时，用 Go 重新实现完整的 FairyGUI 运行时。

**核心约束**：
- Ebiten 提供帧驱动游戏循环和软件渲染原语
- 需要兼容层模拟 LayaAir 的子集服务（sprite 层级、事件、定时器、资源加载）
- 参考实现位于 `laya_src/fairygui`（上游 TypeScript 版本）

## 关键文档

- **架构设计**：`docs/architecture.md` - 分层架构、模块迁移计划、兼容层蓝图
- **进度跟踪**：`docs/refactor-progress.md` - 每日更新的迁移进度和待办事项
- **开发指南**：`AGENTS.md` - 项目结构、编码规范、测试指南

**重要**：添加或移动模块时，必须同步更新这些文档。

## 开发环境限制

⚠️ **沙盒环境限制**：当前开发环境无法运行需要实际 GUI 渲染的测试和 demo。

**工作流程**：
1. Claude 输出需要执行的测试/demo 命令
2. 开发者在 GUI 环境中运行并反馈结果
3. 基于反馈结果进行调整

**可在沙盒运行**：不依赖 Ebiten 渲染的纯逻辑测试（如 ByteBuffer、Relations、事件系统）

**需 GUI 环境**：带 `-tags ebiten` 的测试、`go run ./demo`、`cmd/` 下的调试工具

## 构建与测试命令

### 基础命令
```bash
# 编译检查（防止回归）
go build ./...

# 运行完整测试套件
go test ./...

# 运行特定包测试
go test ./pkg/fgui/core
go test ./pkg/fgui/core -run TestGComponent

# 运行需要 Ebiten 的测试（需 GUI 环境）
go test -tags ebiten ./pkg/fgui/render
go test -tags ebiten ./...

# 性能基准测试
go test -bench . ./pkg/fgui/...

# 使用自定义缓存目录运行测试
GOCACHE=$(pwd)/.gocache go test ./pkg/fgui/core ./pkg/fgui/builder
```

### Demo 运行
```bash
# 主 demo（需 GUI 环境）
go run ./demo

# 调试工具（需 GUI 环境）
go run ./cmd/inspect              # 资源包检视器
go run ./cmd/pixeldiff           # 像素差异对比
go run ./cmd/nineslice-demo      # 九宫格渲染调试
go run ./cmd/bitmapfont-demo     # 位图字体渲染测试
go run ./cmd/text-demo           # 文本渲染测试

# 启用调试输出
FGUI_DEBUG_NINESLICE=1 go run ./cmd/nineslice-demo
FGUI_DEBUG_NINESLICE_OVERLAY=1 go run ./cmd/nineslice-demo
```

## 代码架构

### 分层结构

**应用层** (`cmd/*`, `demo/`)
- 拥有 `ebiten.Game`，驱动 update/draw 循环
- 集成 FGUI 运行时和渲染器
- 示例场景位于 `demo/scenes/`

**FGUI 运行时** (`pkg/fgui/...`)
- **公开 API**：Go 等价的 TypeScript 类
- 包映射：
  - `pkg/fgui/core` - GObject, GComponent, GRoot, Relations, Controllers, Transitions
  - `pkg/fgui/widgets` - GButton, GImage, GTextField, GList, GTree, GMovieClip, etc.
  - `pkg/fgui/assets` - UIPackage, PackageItem, 资源加载与解析
  - `pkg/fgui/builder` - 从 .fui 包构建组件树
  - `pkg/fgui/gears` - 状态齿轮系统（Size, Position, Animation, Color, etc.）
  - `pkg/fgui/tween` - 补间动画引擎
  - `pkg/fgui/utils` - ByteBuffer, 碰撞测试, 颜色工具
  - `pkg/fgui/render` - Ebiten 渲染实现（文本、图形、纹理、色彩效果）

**兼容层** (`internal/compat/laya`)
- 模拟 LayaAir 类型：Sprite, DisplayObject, Graphics, HitArea
- 事件系统：EventDispatcher, Event, 事件冒泡
- 定时器/调度器：Timer, Scheduler, frame loop
- 数学类型：Point, Rect, Matrix
- 输入系统：触控、键盘、focus/capture 管理

**基础设施** (`internal/`)
- `internal/text` - UBB 解析、字体管理、文本布局
- `internal/compat/laya/testutil` - 测试工具（StageEnv, 事件日志）

### 关键渲染流程

1. **显示树**：`GObject` → `laya.Sprite` → `Graphics` 命令
2. **渲染**：`GRoot.Draw` 遍历树 → `render.DrawComponent` 消费 Graphics 命令
3. **颜色效果**：`applyColorEffects` 统一处理颜色矩阵、灰度、BlendMode
4. **纹理绘制**：九宫格/平铺通过 `Graphics.DrawTexture` 命令，渲染层解析
5. **文本**：支持 UBB、多段样式、描边、阴影、系统/位图字体混排

### 帧循环集成

- `GRoot.Advance(delta)` 驱动 ticker、tween、input
- `core.RegisterTicker(func(delta))` 注册帧回调（MovieClip、Tweener）
- compat Stage 处理输入事件并转换为 FairyGUI 事件

## 编码规范

### 格式化
- 使用 `gofmt`（制表符缩进，Go 标准风格）
- 使用 `goimports` 整理导入

### 命名约定
- 导出标识符：`CamelCase`
- 内部帮助函数：`lowerCamelCase`
- 包名：简短、小写、与 FairyGUI 概念对齐
- 避免隐藏全局变量，优先使用显式构造函数/工厂函数

### 注释
- 为非显而易见的行为添加注释
- 标注移植注意事项和与 TypeScript 版本的差异
- 中英文注释均可接受

### 开发原则

**⚠️ 避免过度设计和兼容代码**
- **如果没有明确要求，不要编写兼容代码**
- 专注于当前需求，避免"可能需要"的功能
- YAGNI 原则（You Aren't Gonna Need It）：只实现当前确实需要的功能
- 如果未来需要兼容性，可以在有具体需求时重构

**原因**：
- 兼容代码增加维护负担和复杂度
- 过早的抽象会导致不必要的间接层
- 实际需求出现时再重构更高效

**例外**：
- 有明确的上游 TypeScript 行为需要保持兼容
- 文档或需求明确要求的向后兼容性

## 测试策略

### 测试类型

**单元测试** - 放在包旁边的 `*_test.go`
- 表驱动测试覆盖核心逻辑
- 布局数学、资源解析、定时器、事件传播
- 使用注入的假时钟而非 sleep
- 重点覆盖 `pkg/fgui` 和 `internal/compat`（公开 API 锚点）

**集成测试**
- 使用 `internal/compat/laya/testutil.StageEnv` 模拟舞台环境
- 测试输入路由、事件冒泡、组件交互
- 快照式布局测试：对比预期边界/位置

**渲染测试** (需 `-tags ebiten`)
- 文本渲染、图形绘制、颜色效果
- 使用离屏 `ebiten.Image` 验证像素输出
- 基准测试布局/tween 性能热点

**测试资产**
- 运行时资产：`demo/assets/`
- 测试固定数据：`internal/assets/testdata/`（避免污染 demo）

### 运行测试

```bash
# 快速检查（跳过 Ebiten 依赖）
go test ./pkg/fgui/core ./pkg/fgui/assets

# 完整套件（需 GUI 环境）
go test -tags ebiten ./...

# 单个测试
go test ./pkg/fgui/core -run TestRelations

# 带覆盖率
go test -cover ./pkg/fgui/...
```

## 常见开发任务

### 添加新 Widget

1. 在 `pkg/fgui/widgets/` 创建文件（例如 `newcomponent.go`）
2. 实现与 TypeScript 对应的接口
3. 在 `widgets/factory.go` 注册工厂函数
4. 在 `builder/component.go` 添加解析逻辑
5. 在 `render/draw_*.go` 添加渲染逻辑（如需要）
6. 编写 `*_test.go` 覆盖行为
7. 更新 `docs/refactor-progress.md`

### 移植 TypeScript 模块

1. 参考 `laya_src/fairygui/` 原始实现
2. 识别 LayaAir 依赖 → 映射到兼容层
3. 翻译到 Go，保持方法签名对齐
4. 编写对照测试验证行为一致性
5. 集成到 demo 场景验证
6. 更新架构文档和进度日志

### 调试渲染问题

```bash
# 使用专用调试工具
go run ./cmd/inspect              # 检查资源包内容
go run ./cmd/pixeldiff           # 对比渲染输出
go run ./cmd/nineslice-demo      # 调试九宫格

# 启用调试输出
FGUI_DEBUG_NINESLICE=1 go run ./demo
FGUI_DEBUG_NINESLICE_OVERLAY=1 go run ./demo

# 检查 demo 输出日志（每隔数秒打印层级/状态）
go run ./demo 2>&1 | grep "Scene:"
```

### 更新资产

- 编辑 `demo/UIProject/` 中的 FairyGUI 项目
- 导出到 `demo/assets/*.fui`
- 在 demo 中测试加载
- 如有新格式，更新 `assets/package.go` 解析器

## 提交与 PR 规范

### Commit 格式

遵循约定式提交：
```
feat(scope): 简短描述

详细说明（可选）

破坏性变更（如有）
```

**常见 scope**：
- `assets` - 资源加载/解析
- `core` - GObject/GComponent/GRoot
- `widgets` - 具体控件
- `render` - 渲染实现
- `compat` - LayaAir 兼容层
- `demo` - 示例应用
- `docs` - 文档更新

**示例**：
```
feat(widgets): 实现 GMovieClip 播放控制

- 支持 SetPlaySettings、SyncStatus、时间缩放
- 挂接 core.RegisterTicker 自动推进帧序列
- 新增单测覆盖 Advance/TimeScale/EndHandler 场景
```

### Pull Request 要求

1. **说明目的**：解释为什么需要这个改动
2. **关键代码路径**：指出重要的实现部分
3. **测试证据**：
   - 粘贴 `go test` 输出
   - 对于 UI 改动，提供截图或描述 GUI 环境运行结果
4. **关联 Issue**：链接相关问题
5. **破坏性变更**：标注 API 变更，便于下游消费者规划

## 资源与配置

### 资源位置
- **demo 资源**：`demo/assets/*.fui`
- **测试资源**：`internal/assets/testdata/`（不要放在 demo 中）
- **上游参考**：`laya_src/fairygui/`（TypeScript 原始实现）

### 配置
- Ebiten 配置通过 `demo/main.go` 设置
- 避免在 `pkg/fgui` 导出表面添加配置，保持稳定
- 新增大文件或资源工作流需在 `docs/refactor-progress.md` 记录

## 关键技术细节

### Ticker 系统
- `core.GRoot.Advance(delta)` 调用 `tickAll`
- 使用 `core.RegisterTicker(func(delta))` 注册帧回调
- MovieClip、GTweener 等依赖此机制推进状态

### Graphics 命令系统
- `GGraph`/`GImage`/`GLoader` → `Sprite.Graphics.DrawXXX`
- 渲染层消费 `Graphics` 命令（DrawRect, DrawEllipse, DrawTexture）
- 保持与 Laya 行为一致，支持九宫格/平铺/颜色覆盖

### 颜色效果管线
- `applyColorEffects(sprite, img, blendMode)` 统一处理
- 支持：Alpha、颜色覆盖、灰度、颜色矩阵、BlendMode
- 避免重复缩放 Alpha

### 文本渲染
- UBB 解析：`internal/text.ParseUBB`
- 支持多段样式、描边、阴影、加粗、下划线
- 系统字体与位图字体混排
- AutoSize 模式：测量结果回写 GObject 尺寸

### 输入系统
- compat Stage 处理 Ebiten 输入 → FairyGUI 事件
- 支持多指触控、键盘、focus/capture
- 使用 `internal/compat/laya/testutil.StageEnv` 测试

## 迁移阶段（当前进度见 refactor-progress.md）

1. ✅ **Bootstrap**：兼容层骨架（math, sprite, timer, events）
2. ✅ **Core Port**：GObject, GComponent, relations, controllers
3. 🔄 **渲染组件**：display objects, text, loaders, atlas
4. 🔄 **高级功能**：gears, transitions, tweens, drag-drop
5. 🔄 **资源流程**：UIPackage, fonts, sounds
6. ⏳ **验证**：TypeScript 示例数据 → Go 运行时对比
7. ⏳ **优化**：性能分析、批处理、缓存

## 常见陷阱

- **不要**在没有明确要求的情况下编写兼容代码或过度抽象
- **不要**实现"可能需要"的功能（遵循 YAGNI 原则）
- **不要**在没有读取文件的情况下创建 `CLAUDE.md` 内容
- **不要**重复 `AGENTS.md` 或 `docs/architecture.md` 中已有的内容
- **不要**添加通用开发实践（如"编写单元测试"）
- **记得**同步更新文档（`docs/architecture.md`, `docs/refactor-progress.md`）
- **记得**在 GUI 环境限制下输出命令供人工运行
- **记得**保持与 TypeScript 版本的行为一致性
