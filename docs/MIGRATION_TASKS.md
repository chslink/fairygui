# FairyGUI Go V2 迁移任务清单

> **分支**: `refactor-v2`
> **开始日期**: 2025-11-17
> **预计完成**: 2025-04 (18周)

## 当前进度

- [x] Phase 0: 准备工作
  - [x] 创建 refactor-v2 分支
  - [x] 编写设计文档
  - [x] 创建任务清单
- [ ] Phase 1: 接口设计与验证 (Week 1-2)
- [ ] Phase 2: 显示对象重写 (Week 3-5)
- [ ] Phase 3: 渲染系统重构 (Week 6-7)
- [ ] Phase 4: 事件系统重构 (Week 8)
- [ ] Phase 5: 资源系统简化 (Week 9)
- [ ] Phase 6: 控件迁移 (Week 10-12)
- [ ] Phase 7: 高级功能迁移 (Week 13-14)
- [ ] Phase 8: 兼容层 (Week 15)
- [ ] Phase 9: 测试与文档 (Week 16-17)
- [ ] Phase 10: 发布 (Week 18)

---

## Phase 1: 接口设计与验证 (当前阶段)

### 目标
定义核心接口并验证可行性，为后续实现奠定基础。

### 任务清单

#### 1.1 创建目录结构 ✅
```bash
mkdir -p internal/display
mkdir -p internal/render
mkdir -p internal/builder
mkdir -p internal/input
mkdir -p internal/animation
```

- [x] 创建 internal/display
- [x] 创建 internal/render
- [x] 创建 internal/builder
- [x] 创建 internal/input
- [x] 创建 internal/animation

#### 1.2 定义核心接口
**文件**: `interfaces.go`

- [ ] DisplayObject 接口
  - [ ] 位置相关方法 (Position, SetPosition)
  - [ ] 尺寸相关方法 (Size, SetSize)
  - [ ] 变换相关方法 (Scale, Rotation, Pivot)
  - [ ] 显示属性 (Visible, Alpha)
  - [ ] 层级关系 (Parent, Children, AddChild, RemoveChild)
  - [ ] 渲染方法 (Draw)

- [ ] Renderer 接口
  - [ ] Draw 方法 (渲染显示对象树)
  - [ ] DrawText 方法 (渲染文本)
  - [ ] DrawTexture 方法 (渲染纹理)
  - [ ] DrawShape 方法 (渲染形状)

- [ ] EventDispatcher 接口
  - [ ] On 方法 (注册事件处理器)
  - [ ] Off 方法 (移除事件处理器)
  - [ ] Emit 方法 (触发事件)
  - [ ] Once 方法 (一次性事件)

- [ ] AssetLoader 接口
  - [ ] LoadPackage 方法
  - [ ] LoadTexture 方法
  - [ ] LoadAudio 方法
  - [ ] LoadFont 方法

- [ ] Updatable 接口
  - [ ] Update 方法 (帧更新)

#### 1.3 定义辅助接口
**文件**: `interfaces.go`

- [ ] Positionable 接口
- [ ] Sizable 接口
- [ ] Transformable 接口
- [ ] Visible 接口
- [ ] Interactive 接口
- [ ] Drawable 接口

#### 1.4 定义事件类型
**文件**: `event.go`

- [ ] Event 基础结构
- [ ] MouseEvent 类型
- [ ] TouchEvent 类型
- [ ] KeyboardEvent 类型
- [ ] EventType 常量定义

#### 1.5 编写接口测试
**文件**: `interfaces_test.go`

- [ ] TestDisplayObjectInterface - 测试 DisplayObject 接口规范
- [ ] TestRendererInterface - 测试 Renderer 接口规范
- [ ] TestEventDispatcherInterface - 测试事件接口规范
- [ ] TestAssetLoaderInterface - 测试资源加载接口

#### 1.6 实现 Mock 类型
**文件**: `mock_test.go`

- [ ] MockDisplayObject - DisplayObject 的 Mock 实现
- [ ] MockRenderer - Renderer 的 Mock 实现
- [ ] MockEventDispatcher - EventDispatcher 的 Mock 实现
- [ ] MockAssetLoader - AssetLoader 的 Mock 实现

#### 1.7 接口验证测试
**文件**: `interface_validation_test.go`

- [ ] 验证接口可以被正确实现
- [ ] 验证接口可以被 Mock
- [ ] 验证接口组合符合预期
- [ ] 验证接口不包含过多方法（单一职责）

### 验收标准

- [ ] 所有接口定义完成
- [ ] 所有 Mock 实现完成
- [ ] 所有接口测试通过
- [ ] 代码通过 `go vet` 和 `gofmt` 检查
- [ ] 代码通过 review

### 预计完成时间
2 周 (2025-12-01)

---

## Phase 2: 显示对象重写

### 目标
实现基于 Ebiten 的新显示对象系统，移除 LayaAir 依赖。

### 任务清单

#### 2.1 实现 Sprite 基础类型
**文件**: `internal/display/sprite.go`

- [ ] Sprite 结构定义
- [ ] 位置和变换字段
- [ ] 层级关系管理
- [ ] 实现 DisplayObject 接口
- [ ] 单元测试

#### 2.2 实现 Object 类型
**文件**: `ui.go`

- [ ] Object 结构定义
- [ ] 包装 internal/display.Sprite
- [ ] 实现公开 API
- [ ] 属性 getter/setter
- [ ] 单元测试

#### 2.3 实现 Component 容器类型
**文件**: `ui.go`

- [ ] Component 结构定义
- [ ] 继承 Object
- [ ] 子对象管理 (Children, AddChild, RemoveChild)
- [ ] Controller 管理
- [ ] 单元测试

#### 2.4 实现 Root 根对象
**文件**: `ui.go`

- [ ] Root 结构定义
- [ ] 继承 Component
- [ ] Update 循环实现
- [ ] Draw 循环实现
- [ ] 输入处理集成
- [ ] 单元测试

#### 2.5 实现坐标变换
**文件**: `internal/display/transform.go`

- [ ] LocalToGlobal 方法
- [ ] GlobalToLocal 方法
- [ ] 矩阵计算
- [ ] 单元测试

#### 2.6 迁移测试
**文件**: `ui_test.go`

- [ ] Object 创建测试
- [ ] Component 层级测试
- [ ] Root 更新测试
- [ ] 坐标变换测试

### 验收标准

- [ ] 所有类型实现完成
- [ ] 单元测试覆盖率 >80%
- [ ] 性能测试通过
- [ ] 代码 review 通过

### 预计完成时间
3 周 (2025-12-22)

---

## Phase 3: 渲染系统重构

### 目标
实现高性能的 Ebiten 渲染器，支持批处理。

### 任务清单

#### 3.1 实现基础渲染器
**文件**: `internal/render/renderer.go`

- [ ] EbitenRenderer 结构
- [ ] 实现 Renderer 接口
- [ ] Draw 方法实现
- [ ] 单元测试

#### 3.2 实现纹理渲染
**文件**: `internal/render/texture.go`

- [ ] DrawTexture 实现
- [ ] 九宫格支持
- [ ] 平铺支持
- [ ] 裁剪支持
- [ ] 单元测试

#### 3.3 实现文本渲染
**文件**: `internal/render/text.go`

- [ ] DrawText 实现
- [ ] 系统字体支持
- [ ] 位图字体支持
- [ ] UBB 解析集成
- [ ] 单元测试

#### 3.4 实现图形渲染
**文件**: `internal/render/shape.go`

- [ ] 矩形渲染
- [ ] 圆形渲染
- [ ] 多边形渲染
- [ ] 单元测试

#### 3.5 实现颜色效果
**文件**: `internal/render/effects.go`

- [ ] Alpha 混合
- [ ] 颜色叠加
- [ ] 灰度效果
- [ ] 颜色矩阵
- [ ] BlendMode 支持
- [ ] 单元测试

#### 3.6 实现批处理
**文件**: `internal/render/batch.go`

- [ ] 批处理检测
- [ ] 顶点合并
- [ ] DrawTriangles 优化
- [ ] 性能基准测试

### 验收标准

- [ ] 所有渲染功能完成
- [ ] 渲染效果与 V1 一致
- [ ] 性能优于 V1（基准测试）
- [ ] 单元测试通过

### 预计完成时间
2 周 (2026-01-05)

---

## Phase 4: 事件系统重构

### 目标
实现 Go 风格的事件系统，支持类型安全。

### 任务清单

#### 4.1 实现 EventDispatcher
**文件**: `event.go`

- [ ] EventDispatcher 结构
- [ ] On/Off/Emit 实现
- [ ] Once 实现
- [ ] 事件冒泡
- [ ] 事件捕获
- [ ] 单元测试

#### 4.2 实现输入处理
**文件**: `internal/input/input.go`

- [ ] 鼠标输入处理
- [ ] 触摸输入处理
- [ ] 键盘输入处理
- [ ] 输入状态管理
- [ ] 单元测试

#### 4.3 集成到 Object
**文件**: `ui.go`

- [ ] Object 添加事件字段
- [ ] OnClick 便捷方法
- [ ] OnMouseOver/Out 方法
- [ ] 单元测试

### 验收标准

- [ ] 事件系统完整
- [ ] 类型安全
- [ ] 单元测试通过

### 预计完成时间
1 周 (2026-01-12)

---

## Phase 5: 资源系统简化

### 目标
简化资源加载流程，支持一行代码加载。

### 任务清单

#### 5.1 实现 FileLoader
**文件**: `loader.go`

- [ ] FileLoader 结构
- [ ] LoadPackage 实现
- [ ] LoadTexture 实现
- [ ] LoadAudio 实现
- [ ] 依赖自动加载
- [ ] 单元测试

#### 5.2 实现 Package
**文件**: `package.go`

- [ ] Package 结构
- [ ] PackageItem 结构
- [ ] 资源查找方法
- [ ] URL 解析
- [ ] 单元测试

#### 5.3 实现 Factory
**文件**: `factory.go`

- [ ] Factory 结构
- [ ] CreateComponent 方法
- [ ] CreateObject 方法
- [ ] CreateObjectFromURL 方法
- [ ] 单元测试

### 验收标准

- [ ] 资源加载简化
- [ ] 支持 URL 方式
- [ ] 自动依赖管理
- [ ] 单元测试通过

### 预计完成时间
1 周 (2026-01-19)

---

## Phase 6: 控件迁移

### 目标
迁移所有核心控件到新架构。

### 任务清单

#### 6.1 迁移 Button
**文件**: `widget_button.go`

- [ ] Button 结构定义
- [ ] 继承 Component
- [ ] 状态管理（up, over, down, disabled）
- [ ] OnClick 事件
- [ ] 单元测试

#### 6.2 迁移 Image
**文件**: `widget_image.go`

- [ ] Image 结构定义
- [ ] 纹理显示
- [ ] 九宫格支持
- [ ] 平铺支持
- [ ] 单元测试

#### 6.3 迁移 Text
**文件**: `widget_text.go`

- [ ] Text 结构定义
- [ ] 文本显示
- [ ] UBB 支持
- [ ] AutoSize 支持
- [ ] 单元测试

#### 6.4 迁移 List
**文件**: `widget_list.go`

- [ ] List 结构定义
- [ ] 虚拟列表支持
- [ ] 循环列表支持
- [ ] 滚动集成
- [ ] 单元测试

#### 6.5 迁移其他控件
**文件**: `widget_*.go`

- [ ] ScrollBar
- [ ] Slider
- [ ] ComboBox
- [ ] ProgressBar
- [ ] TextField (输入框)
- [ ] 单元测试

### 验收标准

- [ ] 所有控件功能完整
- [ ] API 简洁易用
- [ ] 单元测试通过

### 预计完成时间
3 周 (2026-02-09)

---

## Phase 7: 高级功能迁移

### 目标
迁移 Gears、Relations、Transitions 等高级功能。

### 任务清单

#### 7.1 迁移 Gears 系统
**文件**: `advanced/gears/*.go`

- [ ] 创建 advanced/gears 包
- [ ] GearBase 基类
- [ ] GearXY - 位置齿轮
- [ ] GearSize - 尺寸齿轮
- [ ] GearLook - 外观齿轮
- [ ] GearColor - 颜色齿轮
- [ ] GearAnimation - 动画齿轮
- [ ] 单元测试

#### 7.2 迁移 Relations 系统
**文件**: `advanced/relations/*.go`

- [ ] 创建 advanced/relations 包
- [ ] Relations 结构
- [ ] RelationItem 结构
- [ ] 各种关系类型
- [ ] 单元测试

#### 7.3 迁移 Tween 动画
**文件**: `tween.go`

- [ ] Tween 结构
- [ ] TweenManager
- [ ] Ease 函数
- [ ] 补间动画
- [ ] 单元测试

#### 7.4 迁移 Transition
**文件**: `advanced/transition/*.go`

- [ ] Transition 结构
- [ ] 过渡动画解析
- [ ] 过渡动画播放
- [ ] 单元测试

### 验收标准

- [ ] 高级功能完整
- [ ] 可选导入
- [ ] 单元测试通过

### 预计完成时间
2 周 (2026-02-23)

---

## Phase 8: 兼容层与迁移工具

### 目标
确保向后兼容，提供平滑迁移路径。

### 任务清单

#### 8.1 创建兼容层
**文件**: `pkg/fgui/compat.go`

- [ ] 类型别名 (GObject → Object)
- [ ] 函数包装
- [ ] 保持原有 API
- [ ] 测试兼容性

#### 8.2 编写迁移指南
**文件**: `docs/migration-guide-v2.md`

- [ ] V1 到 V2 对照表
- [ ] 迁移步骤
- [ ] 常见问题
- [ ] 示例代码

#### 8.3 创建迁移示例
**文件**: `examples/migration/*.go`

- [ ] V1 示例代码
- [ ] V2 迁移后代码
- [ ] 对比说明

### 验收标准

- [ ] 现有代码无需修改
- [ ] 迁移指南清晰
- [ ] 示例完整

### 预计完成时间
1 周 (2026-03-02)

---

## Phase 9: 测试与文档

### 目标
完善测试覆盖和文档。

### 任务清单

#### 9.1 完善单元测试
- [ ] 所有包测试覆盖率 >85%
- [ ] 边界条件测试
- [ ] 错误处理测试

#### 9.2 集成测试
**文件**: `integration_test.go`

- [ ] 完整流程测试
- [ ] 资源加载测试
- [ ] UI 交互测试

#### 9.3 性能基准测试
**文件**: `benchmark_test.go`

- [ ] 对象创建基准
- [ ] 渲染性能基准
- [ ] 内存占用基准
- [ ] 与 V1 对比

#### 9.4 编写 API 文档
**文件**: `docs/api-reference.md`

- [ ] 核心类型文档
- [ ] 接口文档
- [ ] 控件文档
- [ ] 示例代码

#### 9.5 编写使用指南
**文件**: `docs/user-guide.md`

- [ ] 快速开始
- [ ] 核心概念
- [ ] 常用场景
- [ ] 最佳实践

### 验收标准

- [ ] 测试覆盖率 >85%
- [ ] 性能优于 V1
- [ ] 文档完整
- [ ] 示例可运行

### 预计完成时间
2 周 (2026-03-16)

---

## Phase 10: 发布

### 目标
发布 V2.0 版本。

### 任务清单

#### 10.1 准备发布
- [ ] 更新 CHANGELOG
- [ ] 更新 README
- [ ] 更新版本号
- [ ] 创建 release notes

#### 10.2 发布流程
- [ ] 合并到 main 分支
- [ ] 创建 v2.0.0 tag
- [ ] 发布 GitHub Release
- [ ] 更新文档站点

#### 10.3 推广
- [ ] 发布博客文章
- [ ] 社交媒体宣传
- [ ] 通知现有用户

### 验收标准

- [ ] 所有测试通过
- [ ] 文档完整
- [ ] 发布成功

### 预计完成时间
1 周 (2026-03-23)

---

## 日常开发流程

### 开发新功能
1. 从 `refactor-v2` 创建特性分支
2. 实现功能并编写测试
3. 运行测试: `go test ./...`
4. 运行检查: `go vet ./...`
5. 格式化: `gofmt -w .`
6. 提交代码
7. 合并回 `refactor-v2`

### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/display

# 带覆盖率
go test -cover ./...

# 性能基准测试
go test -bench=. ./...
```

### 代码检查
```bash
# 静态检查
go vet ./...

# 格式化
gofmt -w .

# 导入整理
goimports -w .
```

---

## 注意事项

### ⚠️ 重要提醒

1. **不要破坏现有功能**
   - 兼容层必须保持 100% 兼容
   - 所有现有测试必须通过

2. **优先测试**
   - 接口先行，实现跟随
   - 每个功能都要有测试

3. **性能优先**
   - 每个 Phase 都要做性能测试
   - 确保不低于 V1

4. **文档同步**
   - 代码和文档同步更新
   - 及时记录重要决策

### 📝 进度跟踪

- 每周更新此文档
- 记录遇到的问题
- 记录重要决策
- 更新预计时间

### 🔗 相关资源

- [设计文档](./refactor-v2-design.md)
- [架构对比](./refactor-v2-comparison.md)
- [实施路线图](./refactor-v2-roadmap.md)
- [当前架构](./architecture.md)

---

## 问题与决策记录

### 2025-11-17
- **决策**: 创建 refactor-v2 分支开始重构
- **决策**: 采用渐进式迁移策略
- **决策**: 保持 100% 向后兼容

---

**最后更新**: 2025-11-17
**当前阶段**: Phase 1 - 接口设计与验证
**下一里程碑**: 完成核心接口定义 (2025-12-01)
