# FairyGUI Go 代码质量审计报告

> 审计日期: 2026-05-12

## 1. 严重 Bug

### 1.1 `OffClick` 永远无法取消注册

**位置**: `pkg/fgui/core/gobject.go:480-487`

```go
func (g *GObject) OffClick(fn func()) {
    g.Off(laya.EventClick, func(evt *laya.Event) {  // 创建新闭包
        fn()
    })
}
```

事件系统用 `reflect.ValueOf(fn).Pointer()` 做函数指针比对 (`internal/compat/laya/event.go:127`)，每次调用 `OffClick` 都创建一个**新闭包**，指针永远不匹配原始 `OnClick` 注册的回调。

**修复建议**: 需要存储原始函数引用，或改用基于 ID 的注册/注销系统。

### 1.2 `PlayStream` 音频资源泄漏

**位置**: `pkg/fgui/audio/audio.go:189-203`

`playBytes` 在 goroutine 中调用 `playStream`，后者创建 `Player` 后立即返回。goroutine 退出时 `Player` 可能正在播放，导致：
- 音频流被提前 GC
- 播放被意外截断
- 潜在的内存泄漏

**修复建议**: 保持 Player 引用直到播放完成，或使用 channel 通知播放结束。

---

## 2. 中等问题

### 2.1 `gears` 包零测试覆盖

11 个源文件（~2000 行），0 个测试文件。齿轮系统是控制器驱动的核心机制，缺乏测试属于高风险盲区。

### 2.2 `debug` 包零测试覆盖

3 个源文件（`inspector.go`, `server.go`, `simulator.go`），0 个测试。

### 2.3 多个测试被 Skip

| 测试文件 | 行号 | Skip 原因 |
|---------|------|-----------|
| `builder/button_interaction_test.go` | 152 | "修复Button模板Opaque测试" |
| `widgets/list_test.go` | 27 | "修复List点击事件测试 - 与SetMouseThrough(true)或事件模拟相关" |
| `widgets/list_test.go` | 276 | "修复虚拟列表测试 - 与defaultItem或creator相关" |

### 2.4 `cmd/` 目录缺失

CLAUDE.md 引用了 `cmd/inspect`、`cmd/pixeldiff`、`cmd/nineslice-demo`、`cmd/bitmapfont-demo`、`cmd/text-demo`，但 `cmd/` 目录不存在。

### 2.5 未实现的功能 (TODO 标记)

| 文件 | 缺失功能 |
|------|---------|
| `widgets/text_input.go:698-717` | 剪贴板复制/粘贴、撤销/重做 |
| `audio/audio.go:87` | `LoadFile` 文件加载实现 |
| `render/draw_loader.go:586` | Loader 中 MovieClip 填充渲染 |

---

## 3. 代码质量问题

### 3.1 `SetBounds` 中的魔数阈值

**位置**: `core/gcomponent.go:981-983`

```go
if newHeight < currentSize.Y*0.8 {  // 20%阈值缺乏说明
    newHeight = currentSize.Y
}
```

20% 阈值缺乏注释解释，可能在边界情况下导致布局不稳定。需要验证 TypeScript 原版是否有类似逻辑。

### 3.2 `blendModeFromByte` 静默降级

**位置**: `core/gobject.go:23-30`

遇到未知 blend 模式值时静默降级为 `BlendModeNormal`，没有日志或警告。

### 3.3 `Off` 方法泛用性问题

事件系统依赖 `reflect.ValueOf(fn).Pointer()` 做函数匹配。任何使用匿名函数的 `Off` 调用都无法生效，这是架构层面的限制。

### 3.4 `AudioPlayer` 缓存竞态

**位置**: `pkg/fgui/audio/audio.go`

`tryAutoLoadAndPlay` 在 goroutine 中异步加载。`hasAudioDataInCache` 和 `RegisterAudioData` 之间存在 TOCTOU 窗口。影响较小（重复写入相同数据），但不够严谨。

### 3.5 大量注释掉的 Debug 代码

- `render/draw_ebiten.go`: 8+ 行被注释掉的 `fmt.Printf`
- `core/scrollpane.go`: `debugLog` 空实现函数

应使用结构化日志或删除。

### 3.6 类型转换方法返回 `interface{}`

**位置**: `core/gobject.go:950-1004`

```go
func (g *GObject) AsButton() interface{} { return g.data }
func (g *GObject) AsList()  interface{} { return g.data }
```

失去了 Go 的类型安全，调用方需要手动类型断言。增加了出错风险。

---

## 4. 架构设计观察

### 4.1 `data` 字段模式

GObject 使用 `data any` 存储具体 widget 类型（`*GButton`、`*GComponent` 等）。这模拟了 TypeScript 的类继承，但绕过了 Go 的类型系统。

**风险**:
- 运行时类型断言失败难以在编译时发现
- 重构时容易遗漏接口实现
- 测试覆盖率不足时更难发现问题

### 4.2 `EnsureSizeCorrect` 的虚方法模拟

```go
func (g *GObject) EnsureSizeCorrect() {
    // Base implementation is empty - subclasses can override
}
```

Go 没有虚方法，实际行为依赖 `data` 字段的类型断言和接口实现。GTextField 等需要 AutoSize 的组件依赖此模式，易在重构中遗漏。

### 4.3 依赖中心辐射

`builder` 包导入几乎所有 `pkg/fgui` 子包，形成了中心辐射依赖图。如果 widget 需要 builder 功能，会形成循环依赖。

---

## 5. 改进建议

| 优先级 | 问题 | 建议 |
|--------|------|------|
| **P0** | OffClick 无法工作 | 存储原始闭包引用或改用基于 ID 的注册系统 |
| **P0** | PlayStream 资源泄漏 | 保持 Player 引用直到播放完成 |
| **P1** | gears 包无测试 | 至少为 GearDisplay、GearXY、GearSize 添加测试 |
| **P1** | cmd/ 工具缺失 | 创建或更新 CLAUDE.md 移除错误引用 |
| **P2** | 3 个被 Skip 的测试 | 分析并修复根因 |
| **P2** | 魔数 0.8 阈值 | 添加注释解释或提取为常量 |
| **P2** | blend 模式未知值 | 添加日志警告 |
| **P3** | 注释掉的 printf | 删除或替换为结构化日志 |
| **P3** | AsButton/AsList 返回 interface{} | 考虑泛型重写（Go 1.18+） |

---

## 6. 测试统计

| 包 | 源文件 | 测试文件 | 测试数 | 状态 |
|----|--------|---------|--------|------|
| `pkg/fgui` (根) | 1 | 2 | 若干 | ✅ |
| `pkg/fgui/assets` | 7 | 6 | 若干 | ✅ |
| `pkg/fgui/audio` | 2 | 1 | 若干 | ✅ |
| `pkg/fgui/builder` | 1 | 25 | 若干 | ⚠️ 3 skip |
| `pkg/fgui/core` | 15 | 14 | 若干 | ✅ |
| `pkg/fgui/debug` | 3 | 0 | 0 | ❌ 无测试 |
| `pkg/fgui/gears` | 11 | 0 | 0 | ❌ 无测试 |
| `pkg/fgui/render` | 18 | 13 | 若干 | ✅ |
| `pkg/fgui/tween` | 2 | 1 | 若干 | ✅ |
| `pkg/fgui/utils` | 1 | 1 | 若干 | ✅ |
| `pkg/fgui/widgets` | 23 | 25 | 若干 | ⚠️ 2 skip |
| `internal/compat/laya` | 11 | 4 | 若干 | ✅ |
| `internal/text` | 1 | 1 | 若干 | ✅ |

**总计**: 100 个测试文件，3 个被 Skip，2 个包零测试覆盖。
