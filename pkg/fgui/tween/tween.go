package tween

import (
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	defaultEaseOvershoot = 1.70158
	valueSizeColor       = 5
	valueSizeShake       = 6
)

// CatchCallbackPanics mirrors FairyGUI 的全局开关，缺省捕获回调 panic 以防止动画线程崩溃。
var CatchCallbackPanics = true

// Value 表示最多四维的补间数值，兼容 FairyGUI TweenValue。
type Value struct {
	X float64
	Y float64
	Z float64
	W float64
}

// XY 以 tuple 形式返回前两个分量。
func (v Value) XY() (float64, float64) {
	return v.X, v.Y
}

// Component 返回指定索引的分量值，越界时返回 0。
func (v Value) Component(index int) float64 {
	switch index {
	case 0:
		return v.X
	case 1:
		return v.Y
	case 2:
		return v.Z
	case 3:
		return v.W
	default:
		return 0
	}
}

// SetComponent 设置指定分量的值。
func (v *Value) SetComponent(index int, value float64) {
	switch index {
	case 0:
		v.X = value
	case 1:
		v.Y = value
	case 2:
		v.Z = value
	case 3:
		v.W = value
	}
}

// SetZero 将所有分量清零。
func (v *Value) SetZero() {
	v.X, v.Y, v.Z, v.W = 0, 0, 0, 0
}

// SetColor 以 0xAARRGGBB 写入颜色分量。
func (v *Value) SetColor(color uint32) {
	v.X = float64((color >> 16) & 0xFF)
	v.Y = float64((color >> 8) & 0xFF)
	v.Z = float64(color & 0xFF)
	v.W = float64((color >> 24) & 0xFF)
}

// Color 以 0xAARRGGBB 读出颜色分量。
func (v Value) Color() uint32 {
	r := clampColorComponent(v.X)
	g := clampColorComponent(v.Y)
	b := clampColorComponent(v.Z)
	a := clampColorComponent(v.W)
	return (a << 24) | (r << 16) | (g << 8) | b
}

func clampColorComponent(x float64) uint32 {
	if x < 0 {
		return 0
	}
	if x > 255 {
		return 255
	}
	return uint32(math.Round(x))
}

// Path 描述补间曲线上的取样接口，对应 FairyGUI 的 GPath。
type Path interface {
	PointAt(t float64) (x, y float64)
}

type GTweener struct {
	start Value
	end   Value
	value Value
	delta Value

	duration   float64
	delay      float64
	breakpoint float64
	ease       EaseType
	easePeriod float64
	easeAmount float64
	repeat     int
	yoyo       bool
	timeScale  float64
	snapping   bool
	userData   any
	target     any
	prop       string
	path       Path
	onUpdate   func(*GTweener)
	onStart    func(*GTweener)
	onComplete func(*GTweener)
	valueSize  int
	elapsed    float64
	normalized float64
	started    bool
	paused     bool
	killed     bool
	endedState int
}

const (
	endedNone = iota
	endedComplete
	endedBreakpoint
)

type manager struct {
	mu       sync.Mutex
	tweeners []*GTweener
}

var globalManager manager

// To 补间单个分量。
func To(start, end float64, duration float64) *GTweener {
	tw := newTweener(1, duration)
	tw.start.X = start
	tw.end.X = end
	tw.value = tw.start
	return tw
}

// To2 补间两个分量（通常为 X、Y）。
func To2(startX, startY, endX, endY float64, duration float64) *GTweener {
	tw := newTweener(2, duration)
	tw.start.X = startX
	tw.start.Y = startY
	tw.end.X = endX
	tw.end.Y = endY
	tw.value = tw.start
	return tw
}

// To3 补间三个分量。
func To3(startX, startY, startZ, endX, endY, endZ float64, duration float64) *GTweener {
	tw := newTweener(3, duration)
	tw.start.X = startX
	tw.start.Y = startY
	tw.start.Z = startZ
	tw.end.X = endX
	tw.end.Y = endY
	tw.end.Z = endZ
	tw.value = tw.start
	return tw
}

// To4 补间四个分量。
func To4(start Value, end Value, duration float64) *GTweener {
	tw := newTweener(4, duration)
	tw.start = start
	tw.end = end
	tw.value = start
	return tw
}

// ToColor 补间颜色，参数为 0xAARRGGBB。
func ToColor(start, end uint32, duration float64) *GTweener {
	tw := newTweener(valueSizeColor, duration)
	tw.start.SetColor(start)
	tw.end.SetColor(end)
	tw.value = tw.start
	return tw
}

// Shake 实现 FairyGUI 的抖动 tween。
func Shake(startX, startY, amplitude, duration float64) *GTweener {
	tw := newTweener(valueSizeShake, duration)
	tw.start.X = startX
	tw.start.Y = startY
	tw.start.W = amplitude
	tw.value = tw.start
	return tw
}

// DelayedCall 创建一个仅用于延迟执行回调的 tween。
func DelayedCall(delay float64) *GTweener {
	tw := newTweener(0, 0)
	tw.delay = delay
	return tw
}

func newTweener(valueSize int, duration float64) *GTweener {
	tw := &GTweener{}
	tw.reset(valueSize, duration)
	globalManager.add(tw)
	return tw
}

func (t *GTweener) reset(valueSize int, duration float64) {
	t.start.SetZero()
	t.end.SetZero()
	t.value.SetZero()
	t.delta.SetZero()
	t.valueSize = valueSize
	t.duration = duration
	t.delay = 0
	t.breakpoint = -1
	t.ease = EaseTypeQuadOut
	t.easePeriod = 0
	t.easeAmount = defaultEaseOvershoot
	t.repeat = 0
	t.yoyo = false
	t.timeScale = 1
	t.snapping = false
	t.userData = nil
	t.target = nil
	t.prop = ""
	t.path = nil
	t.onUpdate = nil
	t.onStart = nil
	t.onComplete = nil
	t.elapsed = 0
	t.normalized = 0
	t.started = false
	t.paused = false
	t.killed = false
	t.endedState = endedNone
}

// SetDelay 设置延迟。
func (t *GTweener) SetDelay(seconds float64) *GTweener {
	if t != nil {
		t.delay = seconds
	}
	return t
}

// Delay 返回当前延迟。
func (t *GTweener) Delay() float64 {
	if t == nil {
		return 0
	}
	return t.delay
}

// SetDuration 设置时长。
func (t *GTweener) SetDuration(seconds float64) *GTweener {
	if t != nil {
		t.duration = seconds
	}
	return t
}

// Duration 返回时长。
func (t *GTweener) Duration() float64 {
	if t == nil {
		return 0
	}
	return t.duration
}

// SetBreakpoint 设置断点（秒）。
func (t *GTweener) SetBreakpoint(seconds float64) *GTweener {
	if t != nil {
		t.breakpoint = seconds
	}
	return t
}

// SetEase 指定缓动类型。
func (t *GTweener) SetEase(ease EaseType) *GTweener {
	if t != nil {
		t.ease = ease
	}
	return t
}

// SetEasePeriod 设置 Elastic 缓动的周期参数。
func (t *GTweener) SetEasePeriod(period float64) *GTweener {
	if t != nil {
		t.easePeriod = period
	}
	return t
}

// SetEaseOvershootOrAmplitude 设置 Back/Elastic 缓动的幅度参数。
func (t *GTweener) SetEaseOvershootOrAmplitude(amount float64) *GTweener {
	if t != nil {
		t.easeAmount = amount
	}
	return t
}

// SetRepeat 设置重复次数（0 为不重复，负数表示无限），yoyo 为 true 时往返播放。
func (t *GTweener) SetRepeat(repeat int, yoyo bool) *GTweener {
	if t != nil {
		t.repeat = repeat
		t.yoyo = yoyo
	}
	return t
}

// Repeat 返回重复次数设置。
func (t *GTweener) Repeat() int {
	if t == nil {
		return 0
	}
	return t.repeat
}

// SetTimeScale 调整时间缩放（0 会暂停推进）。
func (t *GTweener) SetTimeScale(scale float64) *GTweener {
	if t != nil {
		t.timeScale = scale
	}
	return t
}

// SetSnapping 打开/关闭整数对齐。
func (t *GTweener) SetSnapping(enabled bool) *GTweener {
	if t != nil {
		t.snapping = enabled
	}
	return t
}

// SetTarget 记录 tween 关联对象，可选 prop 作为标识（字符串）。
func (t *GTweener) SetTarget(target any, prop ...string) *GTweener {
	if t == nil {
		return t
	}
	t.target = target
	if len(prop) > 0 {
		t.prop = prop[0]
	} else {
		t.prop = ""
	}
	return t
}

// Target 返回关联对象。
func (t *GTweener) Target() any {
	if t == nil {
		return nil
	}
	return t.target
}

// TargetProp 返回关联的属性标识。
func (t *GTweener) TargetProp() string {
	if t == nil {
		return ""
	}
	return t.prop
}

// SetPath 绑定路径补间。
func (t *GTweener) SetPath(path Path) *GTweener {
	if t != nil {
		t.path = path
	}
	return t
}

// SetUserData 装载用户数据。
func (t *GTweener) SetUserData(data any) *GTweener {
	if t != nil {
		t.userData = data
	}
	return t
}

// UserData 读取用户数据。
func (t *GTweener) UserData() any {
	if t == nil {
		return nil
	}
	return t.userData
}

// OnStart 注册开始回调。
func (t *GTweener) OnStart(fn func(*GTweener)) *GTweener {
	if t != nil {
		t.onStart = fn
	}
	return t
}

// OnUpdate 注册更新回调。
func (t *GTweener) OnUpdate(fn func(*GTweener)) *GTweener {
	if t != nil {
		t.onUpdate = fn
	}
	return t
}

// OnComplete 注册完成回调。
func (t *GTweener) OnComplete(fn func(*GTweener)) *GTweener {
	if t != nil {
		t.onComplete = fn
	}
	return t
}

// StartValueRef 返回起始值引用，便于二次修改。
func (t *GTweener) StartValueRef() *Value {
	if t == nil {
		return nil
	}
	return &t.start
}

// EndValueRef 返回目标值引用。
func (t *GTweener) EndValueRef() *Value {
	if t == nil {
		return nil
	}
	return &t.end
}

// ValueRef 返回当前值引用。
func (t *GTweener) ValueRef() *Value {
	if t == nil {
		return nil
	}
	return &t.value
}

// DeltaValueRef 返回当前帧的增量引用。
func (t *GTweener) DeltaValueRef() *Value {
	if t == nil {
		return nil
	}
	return &t.delta
}

// StartValue 返回起始值副本。
func (t *GTweener) StartValue() Value {
	if t == nil {
		return Value{}
	}
	return t.start
}

// EndValue 返回目标值副本。
func (t *GTweener) EndValue() Value {
	if t == nil {
		return Value{}
	}
	return t.end
}

// Value 返回当前值副本。
func (t *GTweener) Value() Value {
	if t == nil {
		return Value{}
	}
	return t.value
}

// DeltaValue 返回当前增量副本。
func (t *GTweener) DeltaValue() Value {
	if t == nil {
		return Value{}
	}
	return t.delta
}

// NormalizedTime 返回 [0,1] 的归一化进度。
func (t *GTweener) NormalizedTime() float64 {
	if t == nil {
		return 0
	}
	return t.normalized
}

// Completed 表示 tween 是否已进入完成阶段。
func (t *GTweener) Completed() bool {
	if t == nil {
		return false
	}
	return t.endedState != endedNone
}

// AllCompleted 区分完整播放完成（排除断点）。
func (t *GTweener) AllCompleted() bool {
	if t == nil {
		return false
	}
	return t.endedState == endedComplete
}

// SetPaused 切换暂停状态。
func (t *GTweener) SetPaused(paused bool) *GTweener {
	if t != nil {
		t.paused = paused
	}
	return t
}

// Seek 调整时间进度（秒）。
func (t *GTweener) Seek(timePosition float64) {
	if t == nil || t.killed {
		return
	}

	t.elapsed = timePosition
	if t.elapsed < t.delay {
		if t.started {
			t.elapsed = t.delay
		} else {
			return
		}
	}
	t.update()
	if t.endedState != endedNone && !t.killed {
		t.callComplete()
		t.killed = true
		globalManager.remove(t)
	}
}

// Kill 停止 tween；complete 为 true 时会跳至终点并执行完成回调。
func (t *GTweener) Kill(complete bool) {
	if t == nil || t.killed {
		return
	}
	if complete {
		if t.endedState == endedNone {
			switch {
			case t.breakpoint >= 0:
				t.elapsed = t.delay + t.breakpoint
			case t.repeat >= 0:
				t.elapsed = t.delay + t.duration*float64(t.repeat+1)
			default:
				t.elapsed = t.delay + t.duration*2
			}
			t.update()
		}
		t.callComplete()
	}
	t.killed = true
	globalManager.remove(t)
}

func (t *GTweener) advance(delta float64) bool {
	if t.killed {
		return true
	}
	if t.paused {
		return false
	}
	if t.timeScale != 1 && t.timeScale != 0 {
		delta *= t.timeScale
	}
	if delta == 0 {
		return t.killed
	}

	t.elapsed += delta
	if t.elapsed < 0 {
		t.elapsed = 0
	}
	t.update()
	if t.killed {
		return true
	}
	if t.endedState != endedNone {
		t.callComplete()
		t.killed = true
		return true
	}
	return false
}

func (t *GTweener) update() {
	t.endedState = endedNone

	if t.valueSize == 0 {
		if t.elapsed >= t.delay+t.duration {
			t.endedState = endedComplete
		}
		return
	}

	if !t.started {
		if t.elapsed < t.delay {
			return
		}
		t.started = true
		t.callStart()
		if t.killed {
			return
		}
	}

	tt := t.elapsed - t.delay
	if tt < 0 {
		tt = 0
	}

	reversed := false
	if t.breakpoint >= 0 && tt >= t.breakpoint {
		tt = t.breakpoint
		t.endedState = endedBreakpoint
	}

	if t.repeat != 0 && t.duration > 0 {
		rounds := math.Floor(tt / t.duration)
		tt -= t.duration * rounds
		if t.yoyo {
			reversed = int(rounds)%2 == 1
		}
		if t.repeat > 0 && float64(t.repeat)-rounds < 0 {
			if t.yoyo {
				reversed = t.repeat%2 == 1
			}
			tt = t.duration
			t.endedState = endedComplete
		}
	} else if t.duration > 0 && tt >= t.duration {
		tt = t.duration
		t.endedState = endedComplete
	} else if t.duration <= 0 {
		tt = 0
		t.endedState = endedComplete
	}

	duration := t.duration
	if duration <= 0 {
		duration = 1
	}
	current := tt
	if t.yoyo && reversed {
		current = t.duration - tt
		if current < 0 {
			current = 0
		}
	}
	if t.duration <= 0 {
		t.normalized = 1
	} else {
		t.normalized = easeValue(t.ease, current, t.duration, t.easeAmount, t.easePeriod)
	}

	prev := t.value
	t.value.SetZero()
	t.delta.SetZero()

	switch t.valueSize {
	case valueSizeShake:
		if t.endedState == endedNone {
			radius := t.start.W * (1 - t.normalized)
			rx := (rand.Float64()*2 - 1) * radius
			ry := (rand.Float64()*2 - 1) * radius
			t.delta.X = rx
			t.delta.Y = ry
			t.value.X = t.start.X + rx
			t.value.Y = t.start.Y + ry
		} else {
			t.value.X = t.start.X
			t.value.Y = t.start.Y
		}
	default:
		if t.path != nil {
			x, y := t.path.PointAt(t.normalized)
			if t.snapping {
				x = math.Round(x)
				y = math.Round(y)
			}
			t.delta.X = x - prev.X
			t.delta.Y = y - prev.Y
			t.value.X = x
			t.value.Y = y
		} else {
			cnt := t.valueSize
			if cnt > 4 {
				cnt = 4
			}
			for i := 0; i < cnt; i++ {
				n1 := t.start.Component(i)
				n2 := t.end.Component(i)
				f := n1 + (n2-n1)*t.normalized
				if t.snapping {
					f = math.Round(f)
				}
				t.value.SetComponent(i, f)
				t.delta.SetComponent(i, f-prev.Component(i))
			}
		}
	}

	t.callUpdate()
}

func (t *GTweener) callStart() {
	safeInvoke(t.onStart, t)
}

func (t *GTweener) callUpdate() {
	safeInvoke(t.onUpdate, t)
}

func (t *GTweener) callComplete() {
	safeInvoke(t.onComplete, t)
}

func safeInvoke(fn func(*GTweener), t *GTweener) {
	if fn == nil {
		return
	}
	if CatchCallbackPanics {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("tween: callback panic: %v", r)
			}
		}()
	}
	fn(t)
}

// Advance 推进全局补间（通常由 GRoot.Advance 调用）。
func Advance(delta time.Duration) {
	seconds := delta.Seconds()
	if seconds <= 0 {
		return
	}
	list := globalManager.snapshot()
	for _, tw := range list {
		if tw == nil {
			continue
		}
		if tw.advance(seconds) {
			globalManager.remove(tw)
		}
	}
}

// IsTweening 判断 target 是否存在匹配补间；prop 为空表示任意。
func IsTweening(target any, prop ...string) bool {
	propKey := ""
	if len(prop) > 0 {
		propKey = prop[0]
	}
	return globalManager.isTweening(target, propKey)
}

// Kill 移除 target 上的补间；prop 为空代表全部。
func Kill(target any, complete bool, prop ...string) bool {
	propKey := ""
	if len(prop) > 0 {
		propKey = prop[0]
	}
	matches := globalManager.collect(target, propKey)
	for _, tw := range matches {
		tw.Kill(complete)
	}
	return len(matches) > 0
}

// GetTween 返回首个匹配的补间。
func GetTween(target any, prop ...string) *GTweener {
	propKey := ""
	if len(prop) > 0 {
		propKey = prop[0]
	}
	return globalManager.get(target, propKey)
}

func (m *manager) add(t *GTweener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tweeners = append(m.tweeners, t)
}

func (m *manager) remove(t *GTweener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, tw := range m.tweeners {
		if tw == t {
			m.tweeners = append(m.tweeners[:i], m.tweeners[i+1:]...)
			break
		}
	}
}

func (m *manager) snapshot() []*GTweener {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.tweeners) == 0 {
		return nil
	}
	list := make([]*GTweener, len(m.tweeners))
	copy(list, m.tweeners)
	return list
}

func (m *manager) isTweening(target any, prop string) bool {
	if target == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, tw := range m.tweeners {
		if tw != nil && !tw.killed && tw.target == target && (prop == "" || tw.prop == prop) {
			return true
		}
	}
	return false
}

func (m *manager) collect(target any, prop string) []*GTweener {
	if target == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var matches []*GTweener
	for _, tw := range m.tweeners {
		if tw != nil && !tw.killed && tw.target == target && (prop == "" || tw.prop == prop) {
			matches = append(matches, tw)
		}
	}
	return matches
}

func (m *manager) get(target any, prop string) *GTweener {
	if target == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, tw := range m.tweeners {
		if tw != nil && !tw.killed && tw.target == target && (prop == "" || tw.prop == prop) {
			return tw
		}
	}
	return nil
}
