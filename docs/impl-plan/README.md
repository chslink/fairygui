# FairyGUI Go 完全实现方案

> 创建日期: 2026-05-12
> 目标: 完全复刻 `laya_src/demo`，Go 原生 API，性能优化

---

## 一、目标概述

### 1.1 最终目标
- 完全复刻 `laya_src/demo` 全部 22 个场景的交互行为
- 修复已知致命 Bug
- 补齐缺失模块
- 提供 Go 原生的易用 API
- 性能优化

### 1.2 当前差距
| 维度 | 当前状态 | 目标状态 |
|------|---------|---------|
| 致命 Bug | 2 个 | 0 个 |
| TS 模块实现率 | 56/72 (77.8%) | 68/72 (94.4%) |
| Demo 场景完整度 | 4 自定义 + 10 静态 | 22 全交互 |
| API 易用性 | 底层 API 为主 | Go 原生习惯 API |
| 测试覆盖 | gears/debug 零测试 | 核心模块全覆盖 |

---

## 二、实施阶段总览

| 阶段 | 名称 | 预估工时 | 优先级 |
|------|------|---------|--------|
| **0** | 致命 Bug 修复 | 1 天 | P0 |
| **1** | 补齐核心缺失模块 | 5-7 天 | P0 |
| **2** | 补齐 Demo 场景交互 | 5-7 天 | P1 |
| **3** | Go 惯用 API 层 | 3-4 天 | P1 |
| **4** | 性能优化 | 3-4 天 | P2 |
| **5** | 收尾与文档 | 1-2 天 | P2 |

---

## 三、阶段依赖关系

```
Phase 0 (Bug修复)
    │
    ▼
Phase 1 (补全缺失模块)
    │
    ├──────────────────────────┐
    ▼                          ▼
Phase 2 (Demo场景)      Phase 3 (易用API)
    │                          │
    └──────────┬───────────────┘
               ▼
        Phase 4 (性能优化)
               │
               ▼
        Phase 5 (收尾)
```

---

## 四、各阶段详细文档索引

| 阶段 | 文档路径 |
|------|---------|
| 0 - Bug 修复 | `docs/impl-plan/phase0-bugfix.md` |
| 1 - 缺失模块 | `docs/impl-plan/phase1-missing-modules.md` |
| 2 - Demo 场景 | `docs/impl-plan/phase2-demo-scenes.md` |
| 3 - 易用 API | `docs/impl-plan/phase3-idiomatic-api.md` |
| 4 - 性能优化 | `docs/impl-plan/phase4-performance.md` |
| 5 - 收尾 | `docs/impl-plan/phase5-polish.md` |

---

## 五、模块实现优先级矩阵

### Phase 1 补全目标（共 14 个模块，选择 12 个实现）

| 模块 | Demo 使用 | 优先级 | 复杂度 |
|------|----------|--------|--------|
| **Window** | ✅ 是 | P0 | 中 |
| **PopupMenu** | ✅ 是 | P0 | 中 |
| **DragDropManager** | ✅ 是 | P0 | 低 |
| **UIObjectFactory (扩展)** | ✅ 是 | P0 | 低 |
| **ToolSet** | 间接 | P0 | 低 |
| **action/ControllerAction** | 间接 | P1 | 低 |
| **action/ChangePageAction** | 间接 | P1 | 低 |
| **action/PlayTransitionAction** | 间接 | P1 | 低 |
| **IUISource** | 间接 | P1 | 低 |
| **ChildHitArea** | 间接 | P2 | 中 |
| **AsyncOperation** | 否 | P2 | 中 |
| **TranslationHelper** | 否 | P3 | 中 |
| **AssetProxy** | 间接 | ❌ 跳过 | — |
| **GLoader3D** | 否 | ❌ 跳过 | — |

---

## 六、Demo 场景实现清单

| TS 场景 | Go 状态 | 需要实现 |
|---------|---------|---------|
| **MainMenu** | ✅ 自定义 | 无需变更 |
| **BasicsDemo** | ✅ 自定义 | 补充 Depth 子场景 |
| **TransitionDemo** | ✅ 自定义 | 无需变更 |
| **VirtualListDemo** | ✅ 自定义 | 无需变更 |
| **LoopListDemo** | ✅ 自定义 | 无需变更 |
| **JoystickDemo** | ✅ 自定义 | 无需变更 |
| **PullToRefreshDemo** | ⚠️ SimpleScene | 实现交互逻辑 |
| **ModalWaitingDemo** | ⚠️ SimpleScene | 实现交互逻辑 |
| **BagDemo** | ⚠️ SimpleScene | 实现交互逻辑 |
| **ChatDemo** | ⚠️ SimpleScene | 实现交互逻辑 + EmojiParser |
| **ListEffectDemo** | ⚠️ SimpleScene | 实现交互逻辑 + MailItem |
| **ScrollPaneDemo** | ⚠️ SimpleScene | 实现交互逻辑 |
| **TreeViewDemo** | ⚠️ SimpleScene | 实现交互逻辑 |
| **GuideDemo** | ⚠️ SimpleScene | 实现交互逻辑 |
| **CooldownDemo** | ⚠️ SimpleScene | 实现交互逻辑 |
| **HitTestDemo** | ✅ SimpleScene | 无需变更（静态） |
| **SceneTreeDemo** | ❌ 缺失 | 新增 |
| **ExtensionDemo** | ❌ 缺失 | 新增 |
| **RelationDemo** | ❌ 缺失 | 新增 |

---

## 七、非功能性目标

### 7.1 代码质量
- `gears` 包补全测试覆盖（至少 8 个测试文件）
- 消除 3 个被 Skip 的测试
- 所有新模块 100% 单元测试覆盖

### 7.2 性能目标
- 渲染帧率稳定在 60fps（复杂场景不低于 30fps）
- 虚拟列表创建 10000 项 < 100ms
- 对象池命中率 > 90%
- 纹理复用减少 GPU 内存占用

### 7.3 API 设计原则
- 使用 Go 的 Options 模式替代多参数构造函数
- 使用 `context.Context` 传递取消信号
- 使用 `error` 返回值而非 panic
- 提供链式调用 Builder 模式
- 避免 `interface{}` 返回类型，使用泛型或具体类型
