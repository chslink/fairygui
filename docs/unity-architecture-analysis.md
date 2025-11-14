# Unity 版本 FairyGUI 架构深度分析

## 概述

Unity 版本的 FairyGUI 采用了基于 Mesh 的渲染架构，提供高性能的 UI 渲染能力。本文档深入分析其核心设计思想和实现细节，为 Go + Ebiten 版本提供参考。

---

## 1. 整体架构概览

```
┌─────────────────────────────────────────────────────┐
│                   Stage (舞台)                        │
│                   StageEngine                        │
└──────────────────────┬────────────────────────────────┘
                       │
┌──────────────────────▼────────────────────────────────┐
│              UpdateContext                           │
│              (渲染上下文管理)                          │
└──────────────────────┬────────────────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        │                             │
┌───────▼────────┐          ┌────────▼────────┐
│   Container    │          │  DisplayObject  │
│   (容器)       │          │   (显示对象)     │
│                │          │                 │
│ • 层级管理     │          │ • 变换属性      │
│ • 剪裁处理     │          │ • 渲染控制      │
│ • 批处理优化   │          │ • 事件处理      │
└───────┬────────┘          └────────┬────────┘
        │                           │
        │                   ┌───────▼────────┐
        │                   │   NGraphics    │
        │                   │  (渲染核心)     │
        │                   │                │
        │                   │ • MeshFilter   │
        │                   │ • MeshRenderer │
        │                   │ • 材质管理     │
        │                   └───────┬────────┘
        │                           │
        │                   ┌───────▼────────┐
        │                   │ VertexBuffer   │
        │                   │ (顶点缓冲)     │
        │                   │                │
        │                   │ • 顶点数据     │
        │                   │ • 索引数据     │
        │                   │ • 颜色数据     │
        │                   └────────────────┘
        │
┌───────▼─────────────────────────────────────────────┐
│           IMeshFactory (网格工厂)                      │
│                                                         │
│  RectMesh      RoundedRectMesh   EllipseMesh           │
│  PolygonMesh   LineMesh           FillMesh             │
│  ...                                                   │
└─────────────────────────────────────────────────────────┘
```

---

## 2. 核心组件分析

### 2.1 StageEngine (舞台引擎)

**职责**: Unity MonoBehaviour，驱动整个渲染循环

```csharp
public class StageEngine : MonoBehaviour
{
    void LateUpdate() {
        Stage.inst.InternalUpdate();  // 更新舞台
        Stats.ObjectCount = ...;       // 统计信息
        Stats.GraphicsCount = ...;
    }

    void OnGUI() {
        Stage.inst.HandleGUIEvents(Event.current);  // 处理 GUI 事件
    }
}
```

**设计要点**:
- 继承 `MonoBehaviour`，生命周期由 Unity 管理
- `LateUpdate()` 确保在所有 Update 后执行渲染
- 统计信息用于性能监控
- `OnGUI()` 处理 Unity GUI 事件

### 2.2 UpdateContext (渲染上下文)

**职责**: 管理渲染状态、剪裁栈、批处理深度

```csharp
public class UpdateContext
{
    public struct ClipInfo {
        public Rect rect;              // 剪裁矩形
        public Vector4 clipBox;        // 剪裁盒 (用于 shader)
        public bool soft;              // 是否软边
        public uint clipId;            // 剪裁 ID
        public int rectMaskDepth;      // 矩形遮罩深度
        public int referenceValue;     // 模板值
        public bool reversed;          // 是否反向遮罩
    }

    Stack<ClipInfo> _clipStack;        // 剪裁栈

    public int renderingOrder;         // 渲染顺序
    public int batchingDepth;          // 批处理深度
    public int rectMaskDepth;          // 矩形遮罩深度
    public int stencilReferenceValue;  // 模板参考值
}
```

**核心方法**:
- `EnterClipping()`: 进入剪裁模式
- `LeaveClipping()`: 离开剪裁模式
- `ApplyClippingProperties()`: 应用剪裁属性到材质
- `ApplyAlphaMaskProperties()`: 应用 Alpha 遮罩属性

**设计亮点**:
- 使用栈结构管理嵌套剪裁
- 支持矩形剪裁和模板剪裁两种模式
- 通过 `clipBox` 优化 shader 计算 (`clipPos = xy * clipBox.zw + clipBox.xy`)

### 2.3 NGraphics (渲染核心)

**职责**: Unity 渲染组件的封装，管理 Mesh、材质、顶点数据

```csharp
public class NGraphics : IMeshFactory, IBatchable
{
    public MeshFilter meshFilter;      // Unity 网格过滤器
    public MeshRenderer meshRenderer;  // Unity 网格渲染器
    public Mesh mesh;                  // 网格数据
    public NTexture texture;           // 纹理

    public BlendMode blendMode;        // 混合模式
    public bool dontClip;              // 不参与剪裁

    MaterialManager _manager;          // 材质管理器
    IMeshFactory _meshFactory;         // 网格工厂
    VertexMatrix _vertexMatrix;        // 顶点变换矩阵

    public List<NGraphics> subInstances;  // 子实例列表
}
```

**关键方法**:
- `UpdateMesh()`: 更新网格数据
- `Update()`: 渲染更新（应用材质、颜色、Alpha）
- `OnPopulateMesh()`: 生成网格（实现 IMeshFactory 接口）

**性能优化**:
- 延迟更新：`_meshDirty` 标记，仅在需要时更新
- 顶点颜色备份：`_alphaBackup` 保存原始 Alpha
- 子实例支持：复杂对象的子部件渲染

### 2.4 MaterialManager (材质管理器)

**职责**: 智能复用材质，减少渲染状态切换

```csharp
public class MaterialManager
{
    NTexture _texture;                 // 关联纹理
    Shader _shader;                    // 关联着色器
    List<string> _addKeywords;         // 自定义关键词
    Dictionary<int, List<MaterialRef>> _materials;  // 材质缓存

    class MaterialRef {
        public Material material;      // 材质实例
        public int frame;              // 最后使用帧
        public BlendMode blendMode;    // 混合模式
        public uint group;             // 组 ID
    }

    const int internalKeywordsCount = 6;
    static string[] internalKeywords = new[] {
        "CLIPPED", "SOFT_CLIPPED", null,
        "ALPHA_MASK", "GRAYED", "COLOR_FILTER"
    };
}
```

**复用策略**:
1. **多维键值**: Shader + Texture + 关键词组合
2. **帧级缓存**: 记录每帧使用的材质
3. **自动清理**: 回收未使用的材质
4. **关键词管理**: 动态管理 shader 关键词

**设计精髓**:
```csharp
public Material GetMaterial(int flags, BlendMode blendMode, uint group)
{
    // 尝试复用相同参数组合的材质
    for (int i = 0; i < items.Count; i++) {
        if (item.group == group && item.blendMode == blendMode) {
            if (item.frame != frameId) {
                firstMaterialInFrame = true;
                item.frame = frameId;
            }
            return item.material;
        }
    }

    // 复用最近未使用的材质或创建新材质
    if (result == null) {
        result = new MaterialRef() { material = CreateMaterial(flags) };
        items.Add(result);
    }

    return result.material;
}
```

### 2.5 VertexBuffer (顶点缓冲)

**职责**: 管理顶点数据，使用对象池优化性能

```csharp
public sealed class VertexBuffer
{
    public readonly List<Vector3> vertices;    // 顶点位置
    public readonly List<Color32> colors;      // 顶点颜色
    public readonly List<Vector2> uvs;         // 纹理坐标
    public readonly List<Vector2> uvs2;        // 备用纹理坐标
    public readonly List<int> triangles;       // 三角形索引

    public Rect contentRect;                   // 内容矩形
    public Rect uvRect;                        // UV 矩形
    public Color32 vertexColor;                // 顶点颜色
    public Vector2 textureSize;                // 纹理大小

    static Stack<VertexBuffer> _pool = new Stack<VertexBuffer>();  // 对象池

    public static VertexBuffer Begin() {
        if (_pool.Count > 0) {
            return _pool.Pop();
        }
        return new VertexBuffer();
    }

    public void End() {
        _pool.Push(this);  // 返回对象池
    }
}
```

**优化策略**:
- **对象池**: 避免频繁分配/释放
- **复用语义**: `Begin()/End()` 模式
- **完整属性**: 位置、颜色、UV、索引全面管理
- **Alpha 支持**: `_alphaInVertexColor` 标记

### 2.6 Mesh Factory 系统

**接口设计**:
```csharp
public interface IMeshFactory
{
    void OnPopulateMesh(VertexBuffer vb);
}
```

**网格类型**:

#### RectMesh (矩形网格)
```csharp
public class RectMesh : IMeshFactory
{
    public void OnPopulateMesh(VertexBuffer vb) {
        // 添加 4 个顶点和 2 个三角形
        vb.AddQuad(rect, color, uvRect);
        vb.AddTriangles();
    }
}
```

#### RoundedRectMesh (圆角矩形)
- 复杂度: 8-32 个顶点（取决于圆角半径）
- 应用: 按钮、面板等 UI 元素

#### EllipseMesh (椭圆形)
- 应用: 进度条、圆形头像等

#### PolygonMesh (多边形)
- 应用: 不规则形状的 UI 元素

#### LineMesh (线段)
- 支持虚线、箭头等
- 应用: 连接线、流程图

---

## 3. 批处理系统 (Fairy Batching)

### 3.1 批处理原理

**目标**: 减少 DrawCall 数量，提升渲染性能

**条件**:
1. 相同的材质（Shader + Texture + Keywords）
2. 相邻的渲染顺序
3. 相同的混合模式
4. 未被标记为跳过批处理 (`SkipBatching`)

### 3.2 批处理实现

```csharp
public virtual BatchElement AddToBatch(List<BatchElement> batchElements, bool force)
{
    if (graphics != null || force)
    {
        if (_batchElement == null)
            _batchElement = new BatchElement(this, null);

        _batchElement.material = material;
        _batchElement.breakBatch = (_flags & Flags.SkipBatching) != 0;
        batchElements.Add(_batchElement);

        // 处理子实例
        if (graphics.subInstances != null)
        {
            foreach (var g in graphics.subInstances)
            {
                var m = g.material;
                if (m != null)
                {
                    var subBatchElement = g._batchElement;
                    if (subBatchElement == null)
                        subBatchElement = new BatchElement(g, _batchElement.bounds);
                    subBatchElement.material = m;
                    batchElements.Add(subBatchElement);
                }
            }
        }

        return _batchElement;
    }
    return null;
}
```

### 3.3 批处理元素

```csharp
public class BatchElement
{
    public DisplayObject displayObject;    // 关联显示对象
    public Material material;              // 材质
    public Rect bounds;                    // 边界
    public bool breakBatch;                // 是否中断批处理
}
```

---

## 4. 剪裁与遮罩系统

### 4.1 矩形剪裁

**实现方式**: Shader 关键词 + Uniform 参数

**Shader 关键词**:
- `CLIPPED`: 硬边剪裁
- `SOFT_CLIPPED`: 软边剪裁

**Uniform 参数**:
```hlsl
// 顶点着色器计算
float2 clipPos = input.vertex.xy * _ClipBox.zw + _ClipBox.xy;

// 片元着色器测试
if (abs(clipPos.x) > 1 || abs(clipPos.y) > 1) {
    discard;  // 丢弃片元
}
```

**软边效果**:
```hlsl
float2 softClip = input.vertex.xy * _ClipSoftness.zw + _ClipSoftness.xy;
clip(-max(0, abs(clipPos) - softClip));
```

### 4.2 模板剪裁 (Stencil Clipping)

**应用场景**: 复杂形状遮罩、多层遮罩

**原理**:
1. 第一遍：写入模板缓冲（Alpha Mask）
2. 第二遍：根据模板值渲染内容
3. 第三遍：擦除模板（如果需要）

**模板值管理**:
```csharp
public void EnterClipping(uint clipId, bool reversedMask)
{
    if (stencilReferenceValue == 0)
        stencilReferenceValue = 1;
    else
        stencilReferenceValue = stencilReferenceValue << 1;  // 左移 1 位

    if (reversedMask) {
        if (clipInfo.reversed)
            stencilCompareValue = (stencilReferenceValue >> 1) - 1;
        else
            stencilCompareValue = stencilReferenceValue - 1;
    } else {
        stencilCompareValue = (stencilReferenceValue << 1) - 1;
    }

    clipInfo.clipId = clipId;
    clipInfo.referenceValue = stencilReferenceValue;
    clipInfo.reversed = reversedMask;
    clipped = true;
}
```

**Unity Stencil 配置**:
```csharp
mat.SetInt(ShaderConfig.ID_StencilComp,
    (int)UnityEngine.Rendering.CompareFunction.Equal);
mat.SetInt(ShaderConfig.ID_Stencil, stencilCompareValue);
mat.SetInt(ShaderConfig.ID_StencilOp,
    (int)UnityEngine.Rendering.StencilOp.Keep);
```

---

## 5. 文本渲染系统

### 5.1 字体管理

```csharp
public abstract class BaseFont
{
    public string name;                    // 字体名称
    public string shader;                  // 关联着色器
    public NTexture mainTexture;           // 主纹理
    public int version;                    // 版本号（用于缓存失效）

    public abstract void GetGlyphInfo(char charCode, refGlyphInfo glyph);
    public abstract float GetLineHeight();
}
```

#### DynamicFont (动态字体)
- 基于 Unity 的 FreeType 字体渲染
- 实时生成字符纹理
- 支持系统字体

#### BitmapFont (位图字体)
- 预渲染字符集合
- 快速渲染，适合游戏字体
- 支持纹理图集

### 5.2 文本布局

```csharp
public class LineInfo
{
    public float width;                    // 行宽
    public float height;                   // 行高
    public int charCount;                  // 字符数
    public int charIndex;                  // 字符索引
}

List<LineInfo> _lines;                    // 行信息列表
List<CharPosition> _charPositions;        // 字符位置（用于输入）
```

**布局算法**:
1. 解析 UBB/HTML 标签
2. 计算每行字符宽度
3. 处理自动换行
4. 计算垂直对齐

### 5.3 富文本支持

**UBB 标签**:
- `[b]粗体[/b]`
- `[i]斜体[/i]`
- `[u]下划线[/u]`
- `[font color=#FF0000]红色[/font]`
- `[size=20]大号[/size]`
- `[img]image.png[/img]`

**解析流程**:
```csharp
List<HtmlElement> ParseHtml(string html)
{
    // 1. 标签识别
    // 2. 属性解析
    // 3. 嵌套处理
    // 4. 生成元素列表
}
```

---

## 6. 变换矩阵系统

### 6.1 透视变换

**应用**: 3D 效果、卡片翻转等

```csharp
public bool perspective {
    get { return _perspective; }
    set {
        if (_perspective != value) {
            _perspective = value;
            if (_perspective)
                cachedTransform.localEulerAngles = Vector3.zero;  // 屏蔽 Unity 变换
            else
                cachedTransform.localEulerAngles = _rotation;

            UpdateTransformMatrix();
        }
    }
}
```

### 6.2 顶点变换矩阵

```csharp
public class VertexMatrix
{
    public Vector3 cameraPos;              // 相机位置
    public Matrix4x4 matrix;               // 变换矩阵
}

void UpdateTransformMatrix()
{
    Matrix4x4 matrix = Matrix4x4.identity;

    // 应用斜切
    if (_skew.x != 0 || _skew.y != 0)
        ToolSet.SkewMatrix(ref matrix, _skew.x, _skew.y);

    // 应用旋转（如果启用透视）
    if (_perspective)
        matrix *= Matrix4x4.TRS(Vector3.zero, Quaternion.Euler(_rotation), Vector3.one);

    if (matrix.isIdentity)
        _vertexMatrix = null;
    else if (_vertexMatrix == null)
        _vertexMatrix = new VertexMatrix();

    if (_vertexMatrix != null) {
        _vertexMatrix.matrix = matrix;
        _vertexMatrix.cameraPos = new Vector3(
            _pivot.x * _contentRect.width,
            -_pivot.y * _contentRect.height,
            _focalLength
        );
    }
}
```

**顶点变换**:
```csharp
for (int i = 0; i < vertCount; i++) {
    Vector3 pt = vb.vertices[i];

    // 应用矩阵变换
    pt = _vertexMatrix.matrix.MultiplyPoint(pt);

    // 透视计算
    Vector3 vec = pt - camPos;
    float lambda = -camPos.z / vec.z;
    pt = camPos + lambda * vec;

    vb.vertices[i] = pt;
}
```

---

## 7. 绘画模式 (Painting Mode)

### 7.1 设计思想

将整个显示树（或部分）渲染到离屏纹理，然后对纹理进行后处理。

**应用场景**:
- 滤镜效果
- 颜色变换
- 截图功能
- 性能优化（cacheAsBitmap）

### 7.2 实现

```csharp
public void EnterPaintingMode(int requestorId, Margin? extend, float scale)
{
    bool first = _paintingMode == 0;
    _paintingMode |= requestorId;

    if (first) {
        // 创建绘画用 Graphics
        if (paintingGraphics == null) {
            paintingGraphics = new NGraphics(gameObject);
        }

        // 隐藏原内容
        if (graphics != null)
            _SetLayerDirect(CaptureCamera.hiddenLayer);

        // 通知容器更新
        if (this is Container)
            ((Container)this).UpdateBatchingFlags();
    }

    _paintingInfo.extend = (Margin)extend;
    _paintingInfo.scale = scale;
    _paintingInfo.flag = 1;  // 标记需要更新
}
```

### 7.3 离屏渲染

```csharp
void UpdatePainting()
{
    if (_paintingInfo.flag == 1) {
        _paintingInfo.flag = 0;

        // 计算纹理大小
        int textureWidth = Mathf.RoundToInt(
            paintingGraphics.contentRect.width * _paintingInfo.scale);
        int textureHeight = Mathf.RoundToInt(
            paintingGraphics.contentRect.height * _paintingInfo.scale);

        // 创建/调整 RenderTexture
        if (paintingTexture == null ||
            paintingTexture.width != textureWidth ||
            paintingTexture.height != textureHeight) {

            if (paintingTexture != null)
                paintingTexture.Dispose();

            paintingTexture = new NTexture(
                CaptureCamera.CreateRenderTexture(
                    textureWidth, textureHeight,
                    UIConfig.depthSupportForPaintingMode
                )
            );
            paintingGraphics.texture = paintingTexture;
        }
    }

    if (paintingTexture != null)
        paintingTexture.lastActive = Time.time;
}
```

---

## 8. 碰撞测试系统

### 8.1 IHitTest 接口

```csharp
public interface IHitTest
{
    bool HitTest(Vector2 worldPoint, Vector2 direction);
}
```

### 8.2 碰撞测试类型

#### RectHitTest (矩形碰撞)
```csharp
public class RectHitTest : IHitTest
{
    public Rect rect;                      // 矩形区域

    public bool HitTest(Vector2 worldPoint, Vector2 direction) {
        return rect.Contains(worldPoint);
    }
}
```

#### PixelHitTest (像素级碰撞)
```csharp
public class PixelHitTest : IHitTest
{
    public PixelHitTestData data;          // 像素数据
    public Vector2 scale;                  // 缩放比例
    public Vector2 offset;                 // 偏移

    public bool HitTest(Vector2 worldPoint, Vector2 direction) {
        // 转换为像素坐标
        int px = (int)((worldPoint.x - offset.x) * scale.x);
        int py = (int)((worldPoint.y - offset.y) * scale.y);

        // 检查 Alpha
        return data.GetAlpha(px, py) > 0;
    }
}
```

#### MeshColliderHitTest (3D 网格碰撞)
```csharp
public class MeshColliderHitTest : IHitTest
{
    public MeshCollider collider;          // Unity 网格碰撞器

    public bool HitTest(Vector2 worldPoint, Vector2 direction) {
        Camera cam = wsc.GetRenderCamera();
        Ray ray = cam.ScreenPointToRay(screenPoint);
        RaycastHit hit;
        return collider.Raycast(ray, out hit, 100);
    }
}
```

---

## 9. 性能优化总结

### 9.1 批处理优化
- **材质复用**: MaterialManager 智能复用
- **减少 DrawCall**: 合并相同状态的渲染
- **动态合批**: 运行时动态组合

### 9.2 内存优化
- **对象池**: VertexBuffer 对象池
- **延迟分配**: 顶点数据按需创建
- **及时释放**: 材质、纹理生命周期管理

### 9.3 渲染优化
- **动态 Mesh**: `mesh.MarkDynamic()` 优化频繁更新
- **分层渲染**: 渲染顺序优化
- **LOD 支持**: 根据距离调整细节

### 9.4 缓存优化
- **网格缓存**: 复用相同形状的网格
- **纹理缓存**: FontManager 管理字体纹理
- **材质缓存**: 帧级材质复用

---

## 10. 对 Go + Ebiten 版本的启示

### 10.1 可借鉴的设计

1. **MaterialManager 模式**: 智能复用渲染资源
2. **IMeshFactory 接口**: 灵活的网格生成
3. **批处理系统**: 减少渲染调用
4. **UpdateContext**: 统一渲染状态管理
5. **对象池**: 减少内存分配

### 10.2 Ebiten 适配建议

1. **使用 ebiten.Image 作为渲染目标**: 对应 Unity 的 RenderTexture
2. **自定义顶点格式**: 定义适合的顶点结构
3. **Shader 替换为 Filter**: Ebiten 的滤镜系统
4. **命令缓冲**: 实现类似 Unity Graphics.Draw 的命令系统

### 10.3 性能对比预期

| 特性 | Unity 版本 | Go + Ebiten 版本 |
|------|------------|------------------|
| 渲染性能 | 高 (GPU) | 中 (软件渲染) |
| 开发效率 | 高 | 高 |
| 跨平台性 | 中 | 高 |
| 内存占用 | 中 | 低 |
| 启动速度 | 中 | 高 |

---

## 11. 结论

Unity 版本的 FairyGUI 展示了工业级 UI 系统的设计深度：

**核心优势**:
1. **Mesh-Based 渲染**: 高效、可扩展
2. **智能批处理**: 显著提升性能
3. **材质管理**: 减少状态切换
4. **灵活剪裁**: 多种剪裁方式
5. **完备特性**: 文本、变换、动画全覆盖

**设计精髓**:
- 分层架构：业务逻辑与渲染解耦
- 接口抽象：IMeshFactory、IHitTest 等
- 状态管理：UpdateContext 统一管理
- 性能优先：对象池、批处理、缓存

这些设计思想和实现细节为 Go + Ebiten 版本提供了宝贵的参考，有助于构建同样高质量的 UI 系统。