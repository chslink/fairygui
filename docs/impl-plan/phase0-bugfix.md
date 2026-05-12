# Phase 0: 致命 Bug 修复

> 阶段目标: 修复 2 个致命 Bug，消除 3 个 Skip 测试
> 预计工时: 1 天

---

## Bug 1: OffClick 永远无法工作

### 问题描述
`GObject.OffClick()` 创建新的匿名闭包传给 `Off()`，事件系统的 `Off()` 用 `reflect.ValueOf(fn).Pointer()` 做函数指针比对，新闭包指针永远不匹配原始 `OnClick` 注册的回调。

### 影响范围
- `pkg/fgui/core/gobject.go:480-487` — OffClick
- `pkg/fgui/core/gobject.go:484-486` — OffClick 内部匿名函数
- 所有使用 `OffClick` 的地方（目前可能没有调用方，但 API 是公开的）

### 根本原因
当前事件系统架构限制：
```go
// event.go - 用函数指针做 key
func (d *BasicEventDispatcher) Off(evt EventType, fn Listener) {
    key := reflect.ValueOf(fn).Pointer()
    // 在 listeners 中查找 key 匹配的条目删除
}

// gobject.go - OffClick 创建新闭包，指针每次不同
func (g *GObject) OffClick(fn func()) {
    g.Off(laya.EventClick, func(evt *laya.Event) { // 新闭包!
        fn()
    })
}
```

### 修复方案

**方案 A（推荐）: 基于 ID 的监听器注册系统**

```go
// internal/compat/laya/event.go - 新增

type ListenerID uint64

// OnWithID 注册监听器并返回唯一 ID
func (d *BasicEventDispatcher) OnWithID(evt EventType, fn Listener) ListenerID {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.nextID++
    id := d.nextID
    d.entries = append(d.entries, listenerEntry{
        id:   id,
        fn:   fn,
        key:  reflect.ValueOf(fn).Pointer(),
    })
    return ListenerID(id)
}

// OffByID 通过 ID 移除监听器
func (d *BasicEventDispatcher) OffByID(evt EventType, id ListenerID) {
    // 按 ID 查找并删除
}

// 修改 GObject 的 Click 系列方法
func (g *GObject) OnClick(fn func()) ListenerID {
    return g.OnWithID(laya.EventClick, func(evt *laya.Event) { fn() })
}

func (g *GObject) OffClick(id ListenerID) {
    g.OffByID(laya.EventClick, id)
}
```

**方案 B: 存储原始函数引用**

```go
type GObject struct {
    // ... 现有字段 ...
    clickListeners map[func()]listenerEntry
}
```

选方案 A，因为：
1. 不需要在 GObject 上增加额外存储
2. ID 模式更通用，可以扩展到所有事件类型
3. 与 TypeScript 原版的 `Laya.EventDispatcher` 行为更接近

### 实施步骤

1. 在 `internal/compat/laya/event.go` 中：
   - 给 `listenerEntry` 添加 `id uint64` 字段
   - 给 `BasicEventDispatcher` 添加 `nextID uint64` 和 `mu sync.Mutex`
   - 新增 `OnWithID(evt, fn) ListenerID` 方法
   - 新增 `OffByID(evt, id)` 方法
   - 保持原有 `On/Off` 接口不变

2. 在 `pkg/fgui/core/gobject.go` 中：
   - 修改 `OnClick` 返回 `ListenerID`
   - 修改 `OffClick` 接收 `ListenerID`
   - 同样处理 `OnLink`、`OnStateChanged` 等便捷方法

3. 在 `demo/scenes/basics.go` 中：
   - 更新使用 Click 方法的代码

4. 编写单元测试验证 OnClick/OffClick 配对工作

### 验证方法
```go
func TestOffClick(t *testing.T) {
    obj := NewGObject()
    count := 0
    id := obj.OnClick(func() { count++ })
    obj.Emit(laya.EventClick, nil)
    assert.Equal(t, 1, count)
    
    obj.OffClick(id)
    obj.Emit(laya.EventClick, nil)
    assert.Equal(t, 1, count) // 不应再增加
}
```

---

## Bug 2: PlayStream 音频资源泄漏

### 问题描述
`audio.go` 中 `playBytes()` 在 goroutine 中调用 `playStream()`，后者创建 Ebiten `Player` 后立即返回。当 goroutine 退出时 Player 可能还在播放：
- 音频流被提前 GC 导致截断
- Player 未被显式关闭

### 影响范围
- `pkg/fgui/audio/audio.go:104-124` — Play 方法
- `pkg/fgui/audio/audio.go:156-203` — playBytes/playStream

### 修复方案

**方案 A（推荐）: 使用 sync.WaitGroup + 回调等待**

```go
type AudioPlayer struct {
    audioContext *audio.Context
    wg           sync.WaitGroup  // 跟踪活跃的播放
}

func (p *AudioPlayer) playStream(stream io.ReadSeeker, volume float64) {
    player, err := p.audioContext.NewPlayer(stream)
    if err != nil {
        return
    }
    
    p.wg.Add(1)
    player.SetVolume(volume)
    player.Play()
    
    // 在后台等待播放完成
    go func() {
        defer p.wg.Add(-1) // 注意：这里不对，应该先 defer wg.Done()
        for player.IsPlaying() {
            time.Sleep(50 * time.Millisecond)
        }
        player.Close()
    }()
}
```

**方案 B: 播放器池管理**

维护一个活跃播放器列表，定期清理已完成播放的。

选方案 A，实现简单且满足需求。

### 实施步骤

1. 在 `AudioPlayer` 结构体中：
   - 添加 `activePlayers map[*audio.Player]struct{}` 和 `playerMu sync.Mutex`
   - 添加清理 goroutine

2. 修改 `playStream`:
   - 保存 Player 引用到 map
   - 在 goroutine 中等待 `IsPlaying()` 返回 false
   - 调用 `player.Close()`
   - 从 map 移除

3. 在 `playBytes` 中：
   - 移除不必要的 goroutine 嵌套（解码可以在当前 goroutine 中完成）
   - 只在播放阶段异步

### 验证方法
```go
func TestAudioPlayerLifecycle(t *testing.T) {
    p := GetInstance()
    p.RegisterAudioData("test", wavData)
    p.Play("test", 1.0)
    time.Sleep(200 * time.Millisecond)
    // 验证播放后资源被正确释放
    p.playerMu.Lock()
    assert.Equal(t, 0, len(p.activePlayers))
    p.playerMu.Unlock()
}
```

---

## 修复 3 个被 Skip 的测试

### Test 3.1: builder/button_interaction_test.go
**Skip 原因**: "修复Button模板Opaque测试"

需要先调查具体失败原因。可能原因：
- Opaque 属性默认值与 TS 不一致
- Button 模板的 mouseThrough 设置不正确

### Test 3.2: widgets/list_test.go (点击事件)
**Skip 原因**: "修复List点击事件测试 - 与SetMouseThrough(true)或事件模拟相关"

可能原因：List 的 mouseThrough 设置导致事件穿透。

### Test 3.3: widgets/list_test.go (虚拟列表)
**Skip 原因**: "修复虚拟列表测试 - 与defaultItem或creator相关"

可能原因：defaultItem 模板解析或 creator 回调未正确设置。

### 实施步骤
1. 逐个运行被 Skip 的测试，观察实际输出
2. 对比 TypeScript 原版行为
3. 修复根因
4. 移除 `t.Skip()` 调用
5. 验证所有测试通过

---

## 完成标准 (DoD)
- [ ] OffClick 可以正确注销之前注册的回调
- [ ] 音频播放完成后资源正确释放
- [ ] 3 个被 Skip 的测试能正常运行并通过
- [ ] `go test ./...` 零失败
- [ ] 新增事件 ID 系统的单元测试
- [ ] 新增音频生命周期的单元测试
