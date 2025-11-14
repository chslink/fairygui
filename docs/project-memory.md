# 项目记忆 - Unity 版本分析总结

## 概述
添加了 Unity 版本的 FairyGUI SDK 代码分析 (`unity_src/Scripts/`)，共 158 个 C# 文件，深入研究了其 Mesh-based 渲染架构。

## 关键发现

### 1. 渲染架构
- **Mesh-Based 系统**: 基于 Unity Mesh 的高性能渲染
- **NGraphics 核心**: 封装 MeshFilter/MeshRenderer，统一渲染接口
- **VertexBuffer**: 顶点缓冲管理，使用对象池优化

### 2. 性能优化
- **MaterialManager**: 智能材质复用，减少渲染状态切换
  - 基于 Shader + Texture + Keywords 组合
  - 帧级缓存，自动清理未使用材质
  - 支持 6 种内部关键词：CLIPPED, SOFT_CLIPPED, ALPHA_MASK, GRAYED, COLOR_FILTER
- **Fairy Batching**: 批处理系统，减少 DrawCall
  - 条件：相同材质、相邻顺序、相同混合模式
  - BatchElement 记录批处理信息

### 3. 渲染特性
- **剪裁系统**:
  - 矩形剪裁：shader 关键词 + clipBox 计算
  - 模板剪裁：Unity Stencil Buffer，支持复杂遮罩
  - 软边效果：clipSoftness 支持边缘软化
- **变换矩阵**: perspective 模式，模拟 3D 效果
- **绘画模式**: cacheAsBitmap，离屏渲染 + 后处理

### 4. 文本系统
- **字体管理**: BaseFont 接口，DynamicFont + BitmapFont 实现
- **富文本**: UBB 标签解析，HTML 支持
- **布局算法**: 自动换行、垂直对齐、字符位置跟踪

### 5. 碰撞测试
- **IHitTest 接口**: 统一碰撞测试接口
- **多种测试**:
  - RectHitTest: 矩形碰撞
  - PixelHitTest: 像素级精确碰撞
  - MeshColliderHitTest: 3D 网格碰撞
  - ShapeHitTest: 任意形状碰撞

### 6. 核心组件
```
StageEngine (MonoBehaviour)
    ↓
Stage (Container)
    ↓
UpdateContext (渲染上下文)
    ↓
DisplayObject → NGraphics (渲染核心)
    ↓
VertexBuffer (顶点数据)
    ↓
IMeshFactory → RectMesh/RoundedRectMesh/EllipseMesh/...
```

## 对 Go + Ebiten 的启示

### 可借鉴设计
1. **MaterialManager 模式**: 智能资源复用
2. **IMeshFactory 接口**: 灵活的网格生成
3. **UpdateContext**: 统一渲染状态管理
4. **对象池**: VertexBuffer 复用模式
5. **批处理优化**: Fairy Batching 思想

### 适配建议
1. 使用 ebiten.Image 作为 RenderTexture 等价
2. 自定义顶点结构替代 Unity Mesh
3. Shader → Ebiten Filter 适配
4. 命令缓冲系统替代 Graphics.Draw

## 创建文档

### 1. `docs/unity-architecture-analysis.md` (新建)
- 完整的 Unity 版本架构分析
- 11 个章节，详细解析核心组件
- 包含代码示例和设计思路
- 对 Go + Ebiten 的建议

### 2. `docs/project-memory.md` (新建)
- 本文档，项目记忆总结
- 关键发现和设计要点
- 快速参考指南

### 3. `docs/architecture.md` (更新建议)
- 已建议在主架构文档中添加 Unity 版本特性分析
- 重点关注可借鉴的设计模式

## 可行性评估结果 ⭐⭐⭐⭐

完成 Unity 版本设计在 Go + Ebiten 环境下的详细可行性评估，创建了 `docs/feasibility-assessment.md`。

### 高优先级（立即实现）⭐⭐⭐⭐⭐

1. **材质管理系统**: 扩展 AtlasManager，支持 DrawParams 缓存
   - **收益**: 减少对象分配，提升性能
   - **复杂度**: 低
   - **风险**: 低

2. **顶点缓冲对象池**: 添加 VertexBufferPool
   - **收益**: 减少 GC，显著提升性能
   - **复杂度**: 低
   - **风险**: 低

### 中优先级（评估后实现）⭐⭐⭐⭐

3. **批处理系统**: 部分实现命令缓冲
   - **收益**: 中等（软件渲染）
   - **复杂度**: 中
   - **风险**: 中

4. **统一渲染状态**: 参考 UpdateContext
   - **收益**: 代码清晰度
   - **复杂度**: 中
   - **风险**: 中

### 评估结论

| 特性 | Unity 实现 | Ebiten 可行性 | 收益 | 推荐 |
|------|-----------|--------------|------|------|
| 材质管理 | MaterialManager | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ✅ 立即实现 |
| 对象池 | VertexBuffer Pool | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ✅ 立即实现 |
| 批处理 | Fairy Batching | ⭐⭐⭐ | ⭐⭐ | ⭐ 评估实现 |
| 状态管理 | UpdateContext | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐ 评估实现 |
| 剪裁系统 | Shader + Stencil | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ✅ 已实现 |
| 变换矩阵 | VertexMatrix | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ✅ 已实现 |
| 文本系统 | BaseFont + UBB | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐ 长期优化 |

## 实施路线图

### Phase 1: 快速收益（1-2 天）
1. 实现 VertexBufferPool
   - 在 `drawMovieClipWidget()` 中试点
   - 验证性能提升

2. 扩展 AtlasManager
   - 添加 DrawParams 缓存
   - 减少 DrawImageOptions 分配

### Phase 2: 架构优化（1 周）
3. 实现命令缓冲系统
   - 添加 BatchRenderer
   - 重构渲染流程

### Phase 3: 长期优化（持续）
4. 性能基准测试
5. 持续优化迭代

## 技术要点

**Unity 版本代码统计**:
- 总文件数: 158 个 C# 文件
- 核心模块: Core (显示、渲染、文本)、UI (控件)、Event (事件)、Tween (动画)
- 关键文件: NGraphics.cs (879 行)、DisplayObject.cs (1921 行)、UpdateContext.cs (319 行)

**设计精髓**:
- 分层架构：业务与渲染解耦
- 接口抽象：IMeshFactory、IHitTest
- 状态管理：UpdateContext 统一
- 性能优先：对象池、批处理、缓存

---

## 行动项

### 立即行动
- [x] 分析 Unity 版本架构（已完成）
- [x] 评估在 Go + Ebiten 中的可行性（已完成）
- [ ] 实现 VertexBufferPool（高优先级）
- [ ] 扩展 AtlasManager 支持 DrawParams 缓存（高优先级）

### 后续评估
- [ ] 评估批处理系统在 Ebiten 中的实际收益
- [ ] 研究 Ebiten 的 Filter 系统替代 Shader
- [ ] 建立性能基准测试场景

---

**创建时间**: 2025-11-14
**分析范围**: `unity_src/Scripts/` 下所有 C# 文件
**参考价值**: 高 - 工业级 UI 系统设计典范
**可行性评估**: ⭐⭐⭐⭐ (4/5) - 高可行性，重点实现对象池和材质管理