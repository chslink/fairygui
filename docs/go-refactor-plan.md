# FairyGUI Go版本重构方案

## 重构目标
1. 提升渲染性能，支持批处理渲染
2. 采用清晰的分层架构，分离UI逻辑与渲染层
3. 利用Go语言的并发和类型安全优势
4. 保持与现有API的兼容性
5. 增强扩展性和可维护性

## 核心重构点

### 1. 渲染系统重构

#### 1.1 Mesh系统设计
```go
// 定义Mesh接口
type Mesh interface {
    // 获取顶点数据
    Vertices() []float32
    // 获取索引数据
    Indices() []uint16
    // 获取UV数据
    UVs() []float32
    // 获取颜色数据
    Colors() []byte
    // 释放资源
    Release()
}

// 定义Mesh工厂接口
type MeshFactory interface {
    CreateMesh(params interface{}) Mesh
    // 其他工厂方法...
}
```

#### 1.2 批处理渲染
```go
// BatchRenderer 负责合并多个Mesh并渲染
type BatchRenderer struct {
    meshes     []Mesh
    batchMesh  Mesh
    isDirty    bool
    // 其他批处理相关字段...
}

func (r *BatchRenderer) AddMesh(m Mesh) {
    r.meshes = append(r.meshes, m)
    r.isDirty = true
}

func (r *BatchRenderer) Render() {
    if r.isDirty {
        r.mergeMeshes()
        r.isDirty = false
    }
    // 渲染合并后的batchMesh
    renderBatchMesh(r.batchMesh)
}
```

### 2. 分层架构重构

#### 2.1 分离GObject与渲染层
```go
// 定义DisplayObject接口
type DisplayObject interface {
    // 获取Mesh
    Mesh() Mesh
    // 获取变换信息
    Transform() Transform
    // 获取颜色效果
    ColorEffects() ColorEffects
    // 其他渲染相关方法...
}

// 修改GObject结构
type GObject struct {
    // UI逻辑相关字段...
    displayObject DisplayObject // 替换原来的laya.Sprite
}
```

#### 2.2 引入DisplayObjectManager
```go
// DisplayObjectManager 负责管理DisplayObject的生命周期
type DisplayObjectManager struct {
    displayObjects map[uint32]DisplayObject
    batchRenderer  BatchRenderer
    // 其他管理相关字段...
}

func (m *DisplayObjectManager) CreateDisplayObject(t DisplayObjectType) DisplayObject {
    // 创建DisplayObject实例
    // 加入批处理渲染器
    m.batchRenderer.AddMesh(d.Mesh())
}

func (m *DisplayObjectManager) Update() {
    // 更新所有DisplayObject
    // 渲染批处理
    m.batchRenderer.Render()
}
```

### 3. 并发优化

#### 3.1 异步资源加载
```go
// 异步资源加载器
type AsyncResourceLoader struct {
    workChan   chan ResourceRequest
    resultChan chan ResourceResult
    wg         sync.WaitGroup
    // 其他加载器相关字段...
}

func (l *AsyncResourceLoader) LoadResource(req ResourceRequest) {
    l.workChan <- req
}

func (l *AsyncResourceLoader) worker() {
    for req := range l.workChan {
        // 异步加载资源
        result := loadResource(req)
        l.resultChan <- result
        l.wg.Done()
    }
}
```

#### 3.2 并行Mesh生成
```go
// 并行生成Mesh
func GenerateMeshesAsync(objs []*GObject) []Mesh {
    var wg sync.WaitGroup
    meshChan := make(chan Mesh, len(objs))

    for _, obj := range objs {
        wg.Add(1)
        go func(o *GObject) {
            defer wg.Done()
            // 生成Mesh
            mesh := generateMesh(o)
            meshChan <- mesh
        }(obj)
    }

    // 等待所有Mesh生成完成
    go func() {
        wg.Wait()
        close(meshChan)
    }()

    // 收集Mesh
    var meshes []Mesh
    for mesh := range meshChan {
        meshes = append(meshes, mesh)
    }

    return meshes
}
```

### 4. 内存优化

#### 4.1 对象池管理
```go
// Mesh对象池
type MeshPool struct {
    pools map[string][]*Mesh
    mu    sync.Mutex
}

func (p *MeshPool) GetMesh(type string) *Mesh {
    p.mu.Lock()
    defer p.mu.Unlock()

    if pool, exists := p.pools[type]; exists && len(pool) > 0 {
        mesh := pool[len(pool)-1]
        p.pools[type] = pool[:len(pool)-1]
        return mesh
    }
    // 创建新Mesh
    return createNewMesh(type)
}

func (p *MeshPool) ReleaseMesh(mesh *Mesh) {
    p.mu.Lock()
    defer p.mu.Unlock()

    type := mesh.Type()
    if pool, exists := p.pools[type]; exists {
        p.pools[type] = append(p.pools[type], mesh)
    } else {
        p.pools[type] = []*Mesh{mesh}
    }
}
```

#### 4.2 减少接口使用
- 对于性能敏感的路径，使用具体类型而非接口
- 避免不必要的类型断言

### 5. 扩展性增强

#### 5.1 插件机制
```go
// 定义插件接口
type Plugin interface {
    Name() string
    Init()
    Update()
    Render()
    // 其他插件相关方法...
}

// 插件管理器
type PluginManager struct {
    plugins map[string]Plugin
    // 其他管理相关字段...
}

func (m *PluginManager) RegisterPlugin(plugin Plugin) {
    m.plugins[plugin.Name()] = plugin
}

func (m *PluginManager) InitPlugins() {
    for _, plugin := range m.plugins {
        plugin.Init()
    }
}

func (m *PluginManager) UpdatePlugins() {
    for _, plugin := range m.plugins {
        plugin.Update()
    }
}
```

#### 5.2 自定义Mesh支持
```go
// 提供自定义Mesh的接口
type CustomMeshFactory interface {
    MeshFactory
    // 自定义Mesh相关方法...
}

// 注册自定义Mesh工厂
func RegisterCustomMeshFactory(factory CustomMeshFactory) {
    meshFactoryRegistry[factory.Type()] = factory
}
```

## 实施计划

### 阶段1：基础架构搭建 (4-6周)
1. 实现Mesh系统和批处理渲染
2. 完成DisplayObject接口和基础实现
3. 修改GObject结构，分离UI逻辑与渲染层

### 阶段2：核心功能移植 (6-8周)
1. 移植现有UI组件到新架构
2. 实现颜色效果和纹理绘制
3. 集成Ebiten渲染

### 阶段3：性能优化 (3-4周)
1. 实现Mesh对象池
2. 优化批处理算法
3. 并行化Mesh生成

### 阶段4：兼容性和测试 (3-4周)
1. 保持与现有API的兼容性
2. 编写单元测试和性能测试
3. 修复bug和优化

### 阶段5：文档和扩展 (2-3周)
1. 编写重构后的架构文档
2. 实现插件机制和自定义Mesh支持
3. 提供迁移指南

## 预期收益

1. **性能提升**：批处理渲染减少Draw Call数量，提升渲染性能
2. **可维护性增强**：清晰的分层架构降低耦合度
3. **扩展性提升**：支持自定义Mesh和插件扩展
4. **并发优势**：利用Go的并发能力提升资源加载和渲染效率
5. **内存效率**：对象池减少GC压力

该重构方案既借鉴了Unity版本的优秀设计，又充分利用了Go语言的优势，为FairyGUI Go版本的长期发展奠定了坚实基础。