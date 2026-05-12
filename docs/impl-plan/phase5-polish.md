# Phase 5: 收尾与文档

> 阶段目标: 测试覆盖、文档完善、清理代码
> 预计工时: 1-2 天
> 前置依赖: Phase 0-4

---

## 5.1 测试覆盖补全

### 5.1.1 gears 包测试（当前 0 测试）

至少需要以下测试文件：

| 测试文件 | 覆盖内容 |
|---------|---------|
| `gears/display_test.go` | GearDisplay: 页面匹配、Connected/visibleCounter 切换 |
| `gears/display2_test.go` | GearDisplay2: Evaluate 组合逻辑 |
| `gears/xy_test.go` | GearXY: 位置 Tween/百分比/关系联动 |
| `gears/size_test.go` | GearSize: 尺寸变化百分比计算 |
| `gears/color_test.go` | GearColor: 颜色解析/覆盖/Tween |
| `gears/look_test.go` | GearLook: Alpha/Rotation/Grayed/Touchable |
| `gears/text_test.go` | GearText: SetProp 文本值切换 |
| `gears/icon_test.go` | GearIcon: SetProp 图标 URL 切换 |
| `gears/animation_test.go` | GearAnimation: Playing/Frame/TimeScale |
| `gears/fontsize_test.go` | GearFontSize: SetProp 字号切换 |

### 5.1.2 Phase 1 新模块测试

| 测试文件 | 覆盖内容 |
|---------|---------|
| `core/window_test.go` | Window: Show/Hide/CenterOn/Modal/Drag/动画 |
| `core/popupmenu_test.go` | PopupMenu: 添加/移除/选中/分隔线/Show |
| `core/dragdrop_test.go` | DragDropManager: StartDrag/Cancel/DROP 冒泡 |
| `widgets/object_factory_test.go` | UIObjectFactory: setExtension/newObject |

### 5.1.3 消除 Skip 测试

Phase 0 中已完成。

---

## 5.2 代码清理

### 5.2.1 删除注释掉的 Debug 代码

| 文件 | 行范围 | 操作 |
|------|--------|------|
| `render/draw_ebiten.go` | 多处 | 删除注释掉的 `fmt.Printf` |
| `core/scrollpane.go` | `debugLog` 函数 | 删除或替换为结构化日志 |

### 5.2.2 统一文件头注释风格

当前混杂中英文注释。建议：
- 公开 API 注释使用英文（用于 godoc）
- 内部实现注释可以使用中文
- 对齐 TypeScript 原版的注释标记

### 5.2.3 移除未使用的导入和变量

```bash
goimports -w ./...
go vet ./...
```

### 5.2.4 重命名不一致的导出符号

| 当前名 | 建议名 | 原因 |
|--------|--------|------|
| `gObjectCounter` | 不导出（小写） | 内部使用 |
| `buttonStateUp` 等 | 保持不变 | 内部常量 |

---

## 5.3 文档补全

### 5.3.1 必写文档

| 文档 | 路径 | 内容 |
|------|------|------|
| **API 使用指南** | `docs/api-guide.md` | Options 模式、Builder 模式、事件系统、快速开始 |
| **TS 迁移指南** | `docs/migration-from-ts.md` | TS API → Go API 映射表、常见模式翻译 |
| **Demo 说明** | `demo/README.md` | 运行方法、场景列表、按钮映射 |
| **更新架构文档** | `docs/architecture.md` | 补充新模块说明、更新 API 入口 |

### 5.3.2 更新现有文档

- `CLAUDE.md` — 更新 API 入口、cmd/ 工具状态
- `docs/refactor-progress.md` — 记录本次完整实施进度
- `AGENTS.md` — 更新测试命令、开发流程

---

## 5.4 最终验证 Checklist

### 编译与测试
- [ ] `go build ./...` 零错误
- [ ] `go test ./...` 零失败
- [ ] `go test -tags ebiten ./...` 零失败（需 GUI 环境）
- [ ] `go vet ./...` 零警告
- [ ] `goimports -l ./...` 无差异

### 功能验证（GUI 环境）
- [ ] `go run ./demo` 主菜单正常显示
- [ ] 22 个场景全部可导航
- [ ] BasicsDemo 所有子场景交互正常
- [ ] Window 弹出/关闭/拖动/模态正常
- [ ] PopupMenu 右键菜单正常
- [ ] DragDropManager 拖拽正常
- [ ] 文本输入焦点/输入/光标正常
- [ ] 过渡动画播放/停止正常
- [ ] 虚拟列表滚动流畅
- [ ] 循环列表无限滚动
- [ ] 摇杆触摸交互正常
- [ ] 音频播放正常

### 性能验证
- [ ] 所有场景 60fps 稳定（BasicsDemo 可接受 30fps）
- [ ] 虚拟列表 10000 项无卡顿
- [ ] 内存稳定，无持续增长

---

## 5.5 提交计划

```
phase0: fix(events): 修复 OffClick 无法注销 + 音频资源泄漏
phase1: feat(core): 实现 Window/PopupMenu/DragDropManager/UIObjectFactory/ControllerAction
phase1: feat(utils): 实现 ToolSet 工具函数集
phase2: feat(demo): 补全 13 个 Demo 场景交互逻辑
phase2: feat(demo): 新增 SceneTree/Extension/Relation Demo
phase3: feat(api): 添加 Options 模式、Builder 模式、Go 惯用 API
phase4: perf(render): 渲染缓存、可见性剔除、纹理 LRU、对象池
phase5: test(gears): 补全 gears 包测试覆盖
phase5: docs: 更新架构文档、API 指南、迁移指南
phase5: chore: 清理注释掉的 Debug 代码、统一命名
```

---

## 5.6 总体完成标准

| 指标 | 目标 |
|------|------|
| 致命 Bug | 0 |
| TS 模块实现率 | 68/72 (94.4%) |
| Demo 场景 | 22/22 全交互 |
| 测试覆盖 | gears 12 个测试文件，新模块全覆盖 |
| 代码风格 | gofmt + goimports + go vet 零警告 |
| 文档 | API 指南 + 迁移指南 + Demo README |
| 性能 | 60fps 虚拟列表 10000 项 |

---

## 5.7 项目结构最终形态

```
fairygui/
├── demo/
│   ├── main.go
│   ├── README.md                          ← 新增
│   ├── scenes/
│   │   ├── manager.go
│   │   ├── environment.go
│   │   ├── mainmenu.go
│   │   ├── basics.go           (修复 Depth)
│   │   ├── transition_demo.go
│   │   ├── virtual_list_demo.go
│   │   ├── loop_list_demo.go
│   │   ├── joystick.go
│   │   ├── pull_to_refresh.go   ← 新增
│   │   ├── modal_waiting.go     ← 新增
│   │   ├── bag.go               ← 新增
│   │   ├── chat.go              ← 新增
│   │   ├── emoji_parser.go      ← 新增
│   │   ├── list_effect.go       ← 新增
│   │   ├── scroll_pane.go       ← 新增
│   │   ├── tree_view.go         ← 新增
│   │   ├── guide.go             ← 新增
│   │   ├── cooldown.go          ← 新增
│   │   ├── scene_tree.go        ← 新增
│   │   ├── extension.go         ← 新增
│   │   ├── relation.go          ← 新增
│   │   ├── simple_scene.go
│   │   └── util.go
│   └── debug/server.go
├── pkg/fgui/
│   ├── api.go
│   ├── fgui.go                   ← 新增 (便捷函数)
│   ├── options.go                ← 新增 (Options 模式)
│   ├── builder.go                ← 新增 (Builder 模式)
│   ├── core/
│   │   ├── window.go             ← 新增
│   │   ├── popupmenu.go          ← 新增
│   │   ├── dragdrop.go           ← 新增
│   │   ├── controller_action.go  ← 新增
│   │   ├── change_page_action.go ← 新增
│   │   ├── play_transition_action.go ← 新增
│   │   ├── iuisource.go          ← 新增
│   │   ├── ... (现有文件, 部分修改)
│   ├── widgets/
│   │   ├── object_factory.go     ← 重命名自 factory.go
│   │   └── ... (现有文件)
│   ├── utils/
│   │   ├── toolset.go            ← 新增
│   │   └── bytebuffer.go
│   ├── gears/                    ← 新增测试文件
│   └── ...
├── docs/
│   ├── impl-plan/                ← 新增目录
│   │   ├── README.md
│   │   ├── phase0-bugfix.md
│   │   ├── phase1-missing-modules.md
│   │   ├── phase2-demo-scenes.md
│   │   ├── phase3-idiomatic-api.md
│   │   ├── phase4-performance.md
│   │   └── phase5-polish.md
│   ├── api-guide.md              ← 新增
│   ├── migration-from-ts.md      ← 新增
│   └── ... (现有文档, 更新)
└── internal/compat/laya/         ← 部分修改
```

---

## 5.8 风险与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| Window 模态层与现有 Popup 系统冲突 | 中 | 功能异常 | Phase 1 设计时充分理解 GRoot 现有 Popup 栈逻辑 |
| 虚拟列表优化改变现有 behavior | 中 | 回归 Bug | 保持现有测试通过，增量添加优化 |
| Options 模式 API 过于复杂 | 低 | 用户困惑 | 仅提供常用 Option，保留 Setter 方法 |
| Ebiten 版本升级导致 API 变更 | 低 | 编译失败 | 锁定 go.mod 版本，定期检查 |
