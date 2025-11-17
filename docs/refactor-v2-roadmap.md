# FairyGUI Go V2 重构路线图

## 项目目标

将 FairyGUI Go 版本从基于 LayaAir 兼容层的架构，重构为充分利用 Go 语言特性和 Ebiten 引擎的现代化架构。

## 核心改进

1. **简化 API** - 从 `github.com/chslink/fairygui/pkg/fgui/core` 简化为 `github.com/chslink/fairygui`
2. **接口驱动** - 引入 Go 风格的接口设计，提升可测试性和扩展性
3. **性能优化** - 移除 LayaAir 兼容层，预期性能提升 30-40%
4. **向后兼容** - 保持兼容层，现有代码无需修改

## 阶段规划

### Phase 1: 接口设计与验证（2 周）

**目标**：定义核心接口并验证可行性

**任务**：
- [ ] 创建 `interfaces.go` 定义核心接口
  - `DisplayObject` - 显示对象接口
  - `Renderer` - 渲染器接口
  - `EventDispatcher` - 事件分发接口
  - `AssetLoader` - 资源加载接口
- [ ] 编写接口规范测试
- [ ] 实现 Mock 类型用于测试
- [ ] 验证接口设计合理性

**交付物**：
- `interfaces.go` - 核心接口定义
- `interfaces_test.go` - 接口规范测试
- `mock_test.go` - Mock 实现

**验收标准**：
- 所有核心功能都有对应接口
- 接口可以被 Mock
- 通过评审

### Phase 2: 显示对象重写（3 周）

**目标**：实现基于 Ebiten 的新显示对象系统

**任务**：
- [ ] 创建 `internal/display` 包
- [ ] 实现 `Sprite` 类型（不依赖 LayaAir）
- [ ] 实现基础 `Object` 类型
- [ ] 实现 `Component` 容器类型
- [ ] 实现 `Root` 根对象
- [ ] 迁移坐标变换逻辑
- [ ] 迁移层级管理逻辑

**交付物**：
- `internal/display/sprite.go`
- `ui.go` - Object, Component, Root
- 单元测试覆盖率 >80%

**验收标准**：
- Object 实现 DisplayObject 接口
- 支持基本的层级操作
- 支持位置、缩放、旋转等变换
- 通过所有单元测试

### Phase 3: 渲染系统重构（2 周）

**目标**：实现高性能的 Ebiten 渲染器

**任务**：
- [ ] 创建 `internal/render` 包
- [ ] 实现 `EbitenRenderer` 类型
- [ ] 实现纹理渲染
- [ ] 实现文本渲染
- [ ] 实现图形渲染（矩形、圆形等）
- [ ] 实现颜色效果（Alpha、灰度、颜色矩阵）
- [ ] 实现批处理优化
- [ ] 性能基准测试

**交付物**：
- `internal/render/renderer.go`
- `internal/render/text.go`
- `internal/render/texture.go`
- `internal/render/batch.go`
- 性能基准测试报告

**验收标准**：
- 渲染器实现 Renderer 接口
- 渲染性能优于 V1
- 支持批处理
- 通过渲染测试

### Phase 4: 事件系统重构（1 周）

**目标**：实现 Go 风格的事件系统

**任务**：
- [ ] 创建 `event.go`
- [ ] 实现 `EventDispatcher` 类型
- [ ] 实现事件冒泡和捕获
- [ ] 实现输入处理（鼠标、触摸、键盘）
- [ ] 集成到 Object 和 Component

**交付物**：
- `event.go` - 事件系统
- `input.go` - 输入处理
- 事件系统测试

**验收标准**：
- 支持类型安全的事件处理
- 支持事件冒泡和捕获
- 支持输入处理
- 通过事件测试

### Phase 5: 资源系统简化（1 周）

**目标**：简化资源加载流程

**任务**：
- [ ] 创建 `loader.go`
- [ ] 实现 `FileLoader` 类型
- [ ] 实现 `Package` 类型
- [ ] 实现自动依赖管理
- [ ] 支持 URL 方式创建对象

**交付物**：
- `loader.go` - 资源加载
- `package.go` - 包管理
- `factory.go` - 工厂函数

**验收标准**：
- 一行代码加载包
- 支持 URL 方式创建对象
- 自动管理依赖
- 通过资源加载测试

### Phase 6: 控件迁移（3 周）

**目标**：迁移所有核心控件

**任务**：
- [ ] 迁移 Button
- [ ] 迁移 Image
- [ ] 迁移 Text
- [ ] 迁移 List
- [ ] 迁移 ScrollBar
- [ ] 迁移 Slider
- [ ] 迁移 ComboBox
- [ ] 迁移其他控件

**交付物**：
- `widget_button.go`
- `widget_image.go`
- `widget_text.go`
- `widget_list.go`
- ... 其他控件
- 控件测试

**验收标准**：
- 所有控件功能完整
- 通过控件测试
- API 简洁易用

### Phase 7: 高级功能迁移（2 周）

**目标**：迁移 Gears、Relations、Transitions 等高级功能

**任务**：
- [ ] 创建 `advanced/` 包
- [ ] 迁移 Gears 系统
- [ ] 迁移 Relations 系统
- [ ] 迁移 Transitions 系统
- [ ] 迁移 Tween 动画

**交付物**：
- `advanced/gears/` 包
- `advanced/relations/` 包
- `tween.go` - 补间动画
- 高级功能测试

**验收标准**：
- 高级功能可选导入
- 功能完整
- 通过测试

### Phase 8: 兼容层与迁移工具（1 周）

**目标**：确保向后兼容

**任务**：
- [ ] 创建 `pkg/fgui/compat.go`
- [ ] 提供类型别名（GObject → Object）
- [ ] 提供函数包装
- [ ] 编写迁移指南
- [ ] 编写迁移示例

**交付物**：
- `pkg/fgui/compat.go` - 兼容层
- `docs/migration-guide.md` - 迁移指南
- 迁移示例代码

**验收标准**：
- 现有代码无需修改即可运行
- 迁移指南清晰
- 提供迁移示例

### Phase 9: 测试与文档（2 周）

**目标**：完善测试和文档

**任务**：
- [ ] 完善单元测试（覆盖率 >85%）
- [ ] 添加集成测试
- [ ] 性能基准测试
- [ ] 编写 API 文档
- [ ] 编写使用示例
- [ ] 编写最佳实践指南

**交付物**：
- 完整的测试套件
- API 文档
- 示例代码
- 最佳实践指南

**验收标准**：
- 测试覆盖率 >85%
- 文档完整
- 示例可运行

### Phase 10: 发布与推广（1 周）

**目标**：发布 V2 版本

**任务**：
- [ ] 发布 v2.0.0
- [ ] 更新 README
- [ ] 发布博客文章
- [ ] 通知用户

**交付物**：
- v2.0.0 release
- 发布说明
- 博客文章

**验收标准**：
- 通过所有测试
- 文档完整
- 发布成功

## 时间线

```
Week 1-2:   Phase 1 - 接口设计
Week 3-5:   Phase 2 - 显示对象
Week 6-7:   Phase 3 - 渲染系统
Week 8:     Phase 4 - 事件系统
Week 9:     Phase 5 - 资源系统
Week 10-12: Phase 6 - 控件迁移
Week 13-14: Phase 7 - 高级功能
Week 15:    Phase 8 - 兼容层
Week 16-17: Phase 9 - 测试文档
Week 18:    Phase 10 - 发布

总计：18 周（约 4.5 个月）
```

## 快速开始

### 第一周任务清单

**Day 1-2: 环境准备**
```bash
# 1. 创建新分支
git checkout -b refactor-v2

# 2. 创建新目录结构
mkdir -p internal/display internal/render internal/builder

# 3. 创建核心文件
touch interfaces.go ui.go event.go loader.go
```

**Day 3-4: 定义接口**
```go
// interfaces.go
package fairygui

// DisplayObject 显示对象接口
type DisplayObject interface {
    // TODO: 定义方法
}

// Renderer 渲染器接口
type Renderer interface {
    // TODO: 定义方法
}
```

**Day 5: 编写测试**
```go
// interfaces_test.go
package fairygui_test

func TestDisplayObject(t *testing.T) {
    // TODO: 测试接口
}
```

## 关键决策点

### 决策 1: 是否保留 LayaAir 兼容层？

**选项 A**: 完全移除
- ✅ 代码更简洁
- ✅ 性能更好
- ❌ 需要完全重写

**选项 B**: 保留作为内部实现
- ✅ 迁移工作量小
- ❌ 仍有性能损失
- ❌ 维护负担重

**推荐**: **选项 A** - 完全移除，重新基于 Ebiten 设计

### 决策 2: 新旧 API 如何共存？

**选项 A**: 兼容层重定向到新 API
```go
// pkg/fgui/compat.go
type GObject = fairygui.Object
```

**选项 B**: 维护两套独立 API
- ❌ 维护成本高
- ❌ 容易混淆

**推荐**: **选项 A** - 兼容层重定向

### 决策 3: 何时发布 v2.0？

**选项 A**: 功能完整后一次性发布
- ✅ 用户体验好
- ❌ 等待时间长

**选项 B**: 分阶段发布 (v2.0-alpha, v2.0-beta, v2.0)
- ✅ 早期反馈
- ✅ 降低风险
- ❌ 可能有 API 变更

**推荐**: **选项 B** - 分阶段发布

## 风险管理

### 高风险项

1. **接口设计不合理**
   - **缓解**: 先定义接口，评审后实现
   - **应急**: 保留修改接口的权利（v2.0-beta 前）

2. **性能不达预期**
   - **缓解**: 提前性能测试，每个 Phase 都测
   - **应急**: 保留 V1 作为回退方案

3. **工作量超预期**
   - **缓解**: 分阶段发布，优先核心功能
   - **应急**: 调整时间线，延后非核心功能

## 成功指标

- [ ] 导入路径简化为 `import "github.com/chslink/fairygui"`
- [ ] 性能提升 >30%
- [ ] 测试覆盖率 >85%
- [ ] 代码量减少 >40%（用户侧）
- [ ] 100% 向后兼容
- [ ] 文档完整
- [ ] 社区反馈积极

## 相关文档

- [详细设计文档](./refactor-v2-design.md)
- [架构对比](./refactor-v2-comparison.md)
- [当前架构](./architecture.md)

## 联系方式

如有问题，请在以下平台讨论：
- GitHub Issues
- GitHub Discussions
- 项目邮件列表
