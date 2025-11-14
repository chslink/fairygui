# 变换顺序修复说明

## 问题描述

之前的实现中，翻转和旋转操作在**全局坐标系**中进行，导致图像在错误的位置翻转。

### 错误行为
- 图像先移动到目标位置
- 然后在全局坐标系中翻转
- 结果：翻转中心点不对，图像位置错误

### 预期行为
- 图像先在本地坐标系（原点）翻转/旋转
- 然后整体移动到目标位置
- 结果：图像在自己的中心翻转，位置正确

## 修复方案

### 变换矩阵组合顺序

**正确的顺序**（从右到左应用）:
```
最终矩阵 = 父变换 × 本地变换

其中本地变换 = Sprite Offset × 缩放 × 翻转 × 命令偏移
```

### 代码实现

**修复前** (`texture_renderer.go` 旧版本):
```go
// ❌ 错误：直接修改父变换矩阵
geo.Translate(spriteOffset.X, spriteOffset.Y)  // 污染父变换
geo.Translate(cmd.OffsetX, cmd.OffsetY)         // 继续污染

// 然后组合翻转
localGeo.Scale(flipX, flipY)
localGeo.Concat(geo)  // 翻转在错误的坐标系
```

**修复后** (`texture_renderer.go` 新版本):
```go
// ✅ 正确：构建完整的本地变换，再应用父变换
localGeo := ebiten.GeoM{}

// 1. Sprite offset（图像裁剪偏移）
localGeo.Translate(spriteOffsetX, spriteOffsetY)

// 2. 缩放到目标尺寸
localGeo.Scale(sx, sy)

// 3. 翻转（在本地坐标系）
localGeo.Scale(flipX, flipY)
if flipX < 0 {
    localGeo.Translate(dstW, 0)  // 翻转后平移回来
}

// 4. 命令偏移
localGeo.Translate(cmd.OffsetX, cmd.OffsetY)

// 5. 最后应用父变换（位置、旋转等）
localGeo.Concat(parentGeo)  // ✅ 不污染 parentGeo
```

### 关键改进

1. **不修改传入的 `parentGeo`**
   - 旧代码：`geo.Translate(...)` 直接修改
   - 新代码：创建新的 `localGeo`，保持 `parentGeo` 不变

2. **变换顺序正确**
   - 旧代码：位置 → 翻转（错误）
   - 新代码：翻转 → 位置（正确）

3. **适用所有渲染模式**
   - `renderSimple`: 简单缩放渲染
   - `renderScale9`: 九宫格渲染
   - `renderTiled`: 平铺渲染

## 变换数学原理

### 2D 仿射变换矩阵

```
[a  c  tx]   [ScaleX  0      OffsetX]
[b  d  ty] = [0       ScaleY OffsetY]
[0  0  1 ]   [0       0      1      ]
```

### 矩阵乘法顺序

对于点 `P`，应用多个变换：
```
P' = M_parent × M_local × P
```

其中：
- `M_local` = Offset × Scale × Flip
- `M_parent` = 父对象的位置、旋转、缩放

**Ebiten 中的 Concat 顺序**:
```go
m1.Concat(m2)  // 结果: m2 × m1 (注意顺序反转！)
```

所以代码中：
```go
localGeo.Concat(parentGeo)  // 实际计算: parentGeo × localGeo ✅
```

## 测试验证

### 测试用例

```go
// 测试水平翻转
img.SetFlip(widgets.FlipTypeHorizontal)
// 预期：图像在自己的中心水平翻转，然后移动到(x, y)

// 测试旋转 + 翻转
obj.SetRotation(45)  // 父对象旋转
img.SetFlip(widgets.FlipTypeVertical)
// 预期：图像先垂直翻转，然后随父对象一起旋转45度
```

### 验证方法

1. **单元测试**: `TestGImageGeneratesTextureCommand` 验证命令生成
2. **集成测试**: 在 Demo 中观察实际渲染效果
3. **对比测试**: 与原版 LayaAir 行为对比

## 影响范围

### 修改的文件
- `pkg/fgui/render/texture_renderer.go` - 核心修复

### 受益的组件
- `GImage` - 图像翻转
- `GLoader` - 外部资源加载（未来迁移）
- `GMovieClip` - 动画帧（未来迁移）

### 兼容性
- ✅ 向后兼容：API 未变化
- ✅ 行为修复：现在与 LayaAir 一致
- ✅ 测试通过：所有现有测试仍然通过

## 相关文档

- **变换矩阵教程**: https://learnopengl.com/Getting-started/Transformations
- **Ebiten GeoM 文档**: https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2#GeoM
- **LayaAir Transform**: `laya_src/fairygui` 参考实现

## 后续工作

1. 将其他纹理 Widget (GLoader, GMovieClip) 迁移到命令模式时应用相同修复
2. 添加更多集成测试覆盖复杂变换场景
3. 文档化变换顺序的最佳实践

---

**修复日期**: 2025-10-26
**影响版本**: GImage 命令模式重构
**状态**: ✅ 已修复并测试
