package core

import (
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/chslink/fairygui/pkg/fgui/gears"
	"github.com/chslink/fairygui/pkg/fgui/tween"
)

// Transition 代表运行时可播放的 Transition。
type Transition struct {
	mu           sync.Mutex
	owner        *GComponent
	info         TransitionInfo
	playing      bool
	timeScale    float64
	tasks        []*tween.GTweener
	targetCache  map[string]*GObject
	shakeTargets map[*GObject]struct{}
}

func newTransition(owner *GComponent, info TransitionInfo) *Transition {
	t := &Transition{
		owner:        owner,
		targetCache:  make(map[string]*GObject),
		shakeTargets: make(map[*GObject]struct{}),
		timeScale:    1,
	}
	t.reset(info)
	return t
}

func (t *Transition) reset(info TransitionInfo) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.info = info
	t.playing = false
	t.stopAllTweensLocked()
	t.targetCache = make(map[string]*GObject)
	t.shakeTargets = make(map[*GObject]struct{})
}

// Owner 返回所属组件。
func (t *Transition) Owner() *GComponent {
	return t.owner
}

// Info 返回当前 Transition 的元数据副本。
func (t *Transition) Info() TransitionInfo {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.info
}

// Name 返回 Transition 名称。
func (t *Transition) Name() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.info.Name
}

// Playing 指示是否处于播放状态。
func (t *Transition) Playing() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.playing
}

// TimeScale 返回当前播放时间缩放系数。
func (t *Transition) TimeScale() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.timeScale == 0 {
		return 0
	}
	return t.timeScale
}

// SetTimeScale 更新当前播放时间缩放系数（0 表示暂停推进）。
func (t *Transition) SetTimeScale(scale float64) {
	if math.IsNaN(scale) {
		return
	}
	if scale < 0 {
		scale = 0
	}
	t.mu.Lock()
	if t.timeScale == scale {
		t.mu.Unlock()
		return
	}
	t.timeScale = scale
	playing := t.playing
	tasks := append([]*tween.GTweener(nil), t.tasks...)
	items := append([]TransitionItem(nil), t.info.Items...)
	owner := t.owner
	t.mu.Unlock()

	for _, tw := range tasks {
		if tw != nil {
			tw.SetTimeScale(scale)
		}
	}
	if !playing {
		return
	}
	for _, item := range items {
		switch item.Type {
		case TransitionActionTransition:
			if owner != nil && item.Value.TransName != "" {
				if nested := owner.Transition(item.Value.TransName); nested != nil {
					nested.SetTimeScale(scale)
				}
			}
		case TransitionActionAnimation:
			target := t.resolveTarget(item.TargetID)
			if target != nil {
				target.SetProp(gears.ObjectPropIDTimeScale, scale)
			}
		}
	}
}

// Play 启动 Transition。times<=0 表示使用 AutoPlayTimes（≤0 则单次播放），delay<0 表示使用 AutoPlayDelay。
func (t *Transition) Play(times int, delay float64) {
	t.mu.Lock()
	info := t.info
	t.stopAllTweensLocked()
	t.playing = false
	t.tasks = nil
	if times == 0 {
		times = info.AutoPlayTimes
	}
	if times > 1 || times < 0 {
		log.Printf("transition: repeated playback (times=%d) not yet supported, falling back to single run", times)
		times = 1
	}
	baseDelay := delay
	if baseDelay < 0 {
		baseDelay = info.AutoPlayDelay
	}
	if baseDelay < 0 {
		baseDelay = 0
	}
	items := append([]TransitionItem(nil), info.Items...)
	totalDuration := info.TotalDuration
	t.mu.Unlock()

	if len(items) == 0 {
		t.finishPlayback()
		return
	}

	for i := range items {
		t.scheduleItem(&items[i], baseDelay)
	}

	t.mu.Lock()
	t.playing = true
	t.mu.Unlock()

	if totalDuration < 0 {
		totalDuration = 0
	}
	final := tween.DelayedCall(baseDelay + totalDuration).OnComplete(func(*tween.GTweener) {
		t.finishPlayback()
	})
	t.trackTweener(final)
}

// Stop 停止播放。complete==true 时代表直接跳至结束。
func (t *Transition) Stop(complete bool) {
	t.mu.Lock()
	info := t.info
	t.stopAllTweensLocked()
	t.playing = false
	t.mu.Unlock()

	if complete {
		animDurations, lastIndex := computeAnimationCompletion(info)
		for idx, item := range info.Items {
			target := t.resolveTarget(item.TargetID)
			if target == nil {
				continue
			}
			key := item.TargetID
			if key == "" {
				key = "_root"
			}
			if item.Tween != nil {
				val := item.Tween.End
				if item.Type == TransitionActionAnimation && idx == lastIndex[key] {
					if dt, ok := animDurations[key]; ok {
						val.DeltaTime = dt
						delete(animDurations, key)
					}
				}
				t.applyValue(target, item.Type, val)
			} else {
				val := item.Value
				if item.Type == TransitionActionAnimation && idx == lastIndex[key] {
					if dt, ok := animDurations[key]; ok {
						val.DeltaTime = dt
						delete(animDurations, key)
					}
				}
				t.applyValue(target, item.Type, val)
			}
		}
	}
}

func (t *Transition) finishPlayback() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.finishLocked()
}

func (t *Transition) finishLocked() {
	t.stopAllTweensLocked()
	t.playing = false
}

func (t *Transition) stopAllTweensLocked() {
	for _, tw := range t.tasks {
		if tw != nil {
			tw.Kill(false)
		}
	}
	t.tasks = nil
	t.resetShakeTargetsLocked()
}

func (t *Transition) resetShakeTargetsLocked() {
	for target := range t.shakeTargets {
		target.clearShake()
		delete(t.shakeTargets, target)
	}
}

func (t *Transition) scheduleItem(item *TransitionItem, baseDelay float64) {
	if item == nil {
		return
	}
	target := t.resolveTarget(item.TargetID)
	if target == nil {
		return
	}
	delay := baseDelay + item.Time
	if item.Tween != nil && item.Tween.Duration > 0 {
		if tw := t.createTweener(target, item, delay); tw != nil {
			t.trackTweener(tw)
		}
		return
	}
	cloned := *item
	task := tween.DelayedCall(delay).OnComplete(func(*tween.GTweener) {
		t.applyValue(target, cloned.Type, cloned.Value)
	})
	t.trackTweener(task)
}

func (t *Transition) trackTweener(tw *tween.GTweener) {
	if tw == nil {
		return
	}
	scale := t.TimeScale()
	tw.SetTarget(t, "transition")
	tw.SetTimeScale(scale)
	t.mu.Lock()
	t.tasks = append(t.tasks, tw)
	t.mu.Unlock()
}

func (t *Transition) createTweener(target *GObject, item *TransitionItem, delay float64) *tween.GTweener {
	cfg := item.Tween
	if cfg == nil || cfg.Duration <= 0 {
		return nil
	}
	var tw *tween.GTweener
	switch item.Type {
	case TransitionActionXY:
		startX, startY := t.resolvePair(item.Type, target, cfg.Start)
		endX, endY := t.resolvePair(item.Type, target, cfg.End)
		tw = tween.To2(startX, startY, endX, endY, cfg.Duration)
		var path tween.Path
		if len(cfg.Path) > 0 {
			if p := newTransitionPath(cfg.Path); p != nil {
				path = p
				tw.SetPath(p)
			} else {
				log.Printf("transition: failed to build path tween (transition=%s item=%v)", t.Name(), item.Type)
			}
		}
		usePath := path != nil
		tw.OnUpdate(func(tw *tween.GTweener) {
			val := tw.Value()
			if usePath {
				start := tw.StartValue()
				val.X += start.X
				val.Y += start.Y
			}
			t.applyValue(target, item.Type, TransitionValue{
				B1: true,
				B2: true,
				F1: val.X,
				F2: val.Y,
			})
		})
		tw.OnComplete(func(*tween.GTweener) {
			if usePath {
				val := tw.Value()
				start := tw.StartValue()
				t.applyValue(target, item.Type, TransitionValue{
					B1: true,
					B2: true,
					F1: val.X + start.X,
					F2: val.Y + start.Y,
				})
				return
			}
			t.applyValue(target, item.Type, TransitionValue{
				B1: true,
				B2: true,
				F1: endX,
				F2: endY,
			})
		})
	case TransitionActionSize, TransitionActionScale, TransitionActionSkew:
		startX, startY := t.resolvePair(item.Type, target, cfg.Start)
		endX, endY := t.resolvePair(item.Type, target, cfg.End)
		tw = tween.To2(startX, startY, endX, endY, cfg.Duration)
		tw.OnUpdate(func(tw *tween.GTweener) {
			val := tw.Value()
			t.applyValue(target, item.Type, TransitionValue{
				B1: true,
				B2: true,
				F1: val.X,
				F2: val.Y,
			})
		})
		tw.OnComplete(func(*tween.GTweener) {
			t.applyValue(target, item.Type, TransitionValue{
				B1: true,
				B2: true,
				F1: endX,
				F2: endY,
			})
		})
	case TransitionActionAlpha, TransitionActionRotation:
		start := t.resolveSingle(item.Type, target, cfg.Start)
		end := t.resolveSingle(item.Type, target, cfg.End)
		tw = tween.To(start, end, cfg.Duration)
		tw.OnUpdate(func(tw *tween.GTweener) {
			val := tw.Value()
			t.applyValue(target, item.Type, TransitionValue{
				B1: true,
				F1: val.X,
			})
		})
		tw.OnComplete(func(*tween.GTweener) {
			t.applyValue(target, item.Type, TransitionValue{
				B1: true,
				F1: end,
			})
		})
	case TransitionActionColor:
		start := t.resolveColor(target, cfg.Start)
		end := t.resolveColor(target, cfg.End)
		tw = tween.ToColor(start, end, cfg.Duration)
		tw.OnUpdate(func(tw *tween.GTweener) {
			val := tw.Value()
			t.applyValue(target, item.Type, TransitionValue{
				Color: val.Color(),
			})
		})
	case TransitionActionColorFilter:
		start := tween.Value{X: cfg.Start.F1, Y: cfg.Start.F2, Z: cfg.Start.F3, W: cfg.Start.F4}
		end := tween.Value{X: cfg.End.F1, Y: cfg.End.F2, Z: cfg.End.F3, W: cfg.End.F4}
		tw = tween.To4(start, end, cfg.Duration)
		tw.OnStart(func(*tween.GTweener) {
			t.applyValue(target, item.Type, cfg.Start)
		})
		tw.OnUpdate(func(tw *tween.GTweener) {
			val := tw.Value()
			t.applyValue(target, item.Type, TransitionValue{
				F1: val.X,
				F2: val.Y,
				F3: val.Z,
				F4: val.W,
			})
		})
		tw.OnComplete(func(*tween.GTweener) {
			t.applyValue(target, item.Type, cfg.End)
		})
	case TransitionActionShake:
		amp := cfg.Start.Amplitude
		if amp == 0 {
			amp = cfg.End.Amplitude
		}
		duration := cfg.Duration
		if duration <= 0 {
			duration = cfg.End.Duration
		}
		if duration <= 0 {
			duration = 0.3
		}
		tw = tween.Shake(target.X(), target.Y(), amp, duration)
		tw.OnUpdate(func(tw *tween.GTweener) {
			delta := tw.DeltaValue()
			t.applyValue(target, item.Type, TransitionValue{
				OffsetX: delta.X,
				OffsetY: delta.Y,
			})
		})
		tw.OnComplete(func(*tween.GTweener) {
			t.applyValue(target, item.Type, TransitionValue{
				OffsetX: 0,
				OffsetY: 0,
			})
			t.mu.Lock()
			delete(t.shakeTargets, target)
			t.mu.Unlock()
		})
	case TransitionActionAnimation, TransitionActionVisible, TransitionActionText, TransitionActionIcon, TransitionActionSound,
		TransitionActionTransition:
		// Currently not tweened; fall back to delayed application.
		cloned := *item
		task := tween.DelayedCall(delay + cfg.Duration).OnComplete(func(*tween.GTweener) {
			t.applyValue(target, cloned.Type, cloned.Tween.End)
		})
		return task
	default:
		return nil
	}

	if tw == nil {
		return nil
	}
	if delay > 0 {
		tw.SetDelay(delay)
	} else if item.Time > 0 {
		tw.SetDelay(item.Time)
	}
	if cfg.EaseType >= 0 {
		tw.SetEase(tween.EaseType(cfg.EaseType))
	}
	if cfg.Repeat > 0 {
		tw.SetRepeat(cfg.Repeat, cfg.Yoyo)
	}
	return tw
}

func (t *Transition) resolvePair(action TransitionAction, target *GObject, value TransitionValue) (float64, float64) {
	x, y := target.X(), target.Y()
	switch action {
	case TransitionActionSize:
		x, y = target.Width(), target.Height()
	case TransitionActionScale:
		x, y = target.Scale()
	case TransitionActionSkew:
		x, y = target.Skew()
	}
	if value.B1 {
		if value.B3 && t.owner != nil {
			x = value.F1 * t.owner.Width()
		} else {
			x = value.F1
		}
	}
	if value.B2 {
		if value.B3 && t.owner != nil {
			y = value.F2 * t.owner.Height()
		} else {
			y = value.F2
		}
	}
	return x, y
}

func (t *Transition) resolveSingle(action TransitionAction, target *GObject, value TransitionValue) float64 {
	switch action {
	case TransitionActionAlpha:
		if value.B1 {
			return value.F1
		}
		return target.Alpha()
	case TransitionActionRotation:
		if value.B1 {
			return value.F1
		}
		return target.Rotation() * 180 / math.Pi
	default:
		return value.F1
	}
}

func (t *Transition) resolveColor(target *GObject, value TransitionValue) uint32 {
	if value.Color != 0 {
		return value.Color
	}
	// fall back to stored prop (if any)
	prop := target.GetProp(gears.ObjectPropIDColor)
	if str, ok := prop.(string); ok {
		var r, g, b uint32
		if _, err := fmt.Sscanf(str, "#%02x%02x%02x", &r, &g, &b); err == nil {
			return (0xFF << 24) | (r << 16) | (g << 8) | b
		}
	}
	return 0xFFFFFFFF
}

func (t *Transition) applyValue(target *GObject, action TransitionAction, value TransitionValue) {
	if target == nil {
		return
	}
	switch action {
	case TransitionActionXY:
		x, y := target.X(), target.Y()
		if value.B1 {
			x = value.F1
		}
		if value.B2 {
			y = value.F2
		}
		target.SetPosition(x, y)
	case TransitionActionSize:
		w, h := target.Width(), target.Height()
		if value.B1 {
			w = value.F1
		}
		if value.B2 {
			h = value.F2
		}
		target.SetSize(w, h)
	case TransitionActionScale:
		sx, sy := target.Scale()
		if value.B1 {
			sx = value.F1
		}
		if value.B2 {
			sy = value.F2
		}
		target.SetScale(sx, sy)
	case TransitionActionPivot:
		px, py := target.Pivot()
		if value.B1 {
			px = value.F1
		}
		if value.B2 {
			py = value.F2
		}
		target.SetPivotWithAnchor(px, py, target.PivotAsAnchor())
	case TransitionActionAlpha:
		alpha := value.F1
		if !value.B1 {
			alpha = target.Alpha()
		}
		target.SetAlpha(alpha)
	case TransitionActionRotation:
		rotation := value.F1
		if !value.B1 {
			rotation = target.Rotation() * 180 / math.Pi
		}
		target.SetRotation(rotation * math.Pi / 180)
	case TransitionActionSkew:
		skewX, skewY := target.Skew()
		if value.B1 {
			skewX = value.F1 * math.Pi / 180
		}
		if value.B2 {
			skewY = value.F2 * math.Pi / 180
		}
		target.SetSkew(skewX, skewY)
	case TransitionActionVisible:
		target.SetVisible(value.Visible)
	case TransitionActionText:
		target.SetProp(gears.ObjectPropIDText, value.Text)
	case TransitionActionIcon:
		target.SetProp(gears.ObjectPropIDIcon, value.Text)
	case TransitionActionColor:
		color := value.Color
		if color == 0 {
			color = 0xFFFFFFFF
		}
		target.SetProp(gears.ObjectPropIDColor, argbToHex(color))
	case TransitionActionAnimation:
		if value.Frame >= 0 {
			target.SetProp(gears.ObjectPropIDFrame, value.Frame)
		}
		target.SetProp(gears.ObjectPropIDPlaying, value.Playing)
		target.SetProp(gears.ObjectPropIDTimeScale, t.TimeScale())
		if value.DeltaTime != 0 {
			target.SetProp(gears.ObjectPropIDDeltaTime, value.DeltaTime)
		} else {
			target.SetProp(gears.ObjectPropIDDeltaTime, 0.0)
		}
	case TransitionActionSound:
		playTransitionSound(value.Sound, value.Volume)
	case TransitionActionTransition:
		if value.TransName != "" {
			if nested := t.owner.Transition(value.TransName); nested != nil {
				nested.SetTimeScale(t.TimeScale())
				nested.Play(value.PlayTimes, -1)
			}
		}
	case TransitionActionShake:
		target.applyShake(value.OffsetX, value.OffsetY)
		t.mu.Lock()
		if value.OffsetX != 0 || value.OffsetY != 0 {
			t.shakeTargets[target] = struct{}{}
		} else {
			delete(t.shakeTargets, target)
		}
		t.mu.Unlock()
	case TransitionActionColorFilter:
		if value.F1 == 0 && value.F2 == 0 && value.F3 == 0 && value.F4 == 0 {
			target.ClearColorFilter()
		} else {
			target.SetColorFilter(value.F1, value.F2, value.F3, value.F4)
		}
	default:
		log.Printf("transition: action %v not yet supported", action)
	}
}

func (t *Transition) resolveTarget(id string) *GObject {
	if id == "" {
		if t.owner != nil {
			return t.owner.GObject
		}
		return nil
	}
	t.mu.Lock()
	if obj, ok := t.targetCache[id]; ok {
		t.mu.Unlock()
		return obj
	}
	t.mu.Unlock()
	if t.owner == nil {
		return nil
	}
	var match *GObject
	children := t.owner.Children()
	for _, child := range children {
		if child == nil {
			continue
		}
		if child.ResourceID() == id || child.Name() == id || child.ID() == id {
			match = child
			break
		}
	}
	if match != nil {
		t.mu.Lock()
		t.targetCache[id] = match
		t.mu.Unlock()
	}
	return match
}

func argbToHex(color uint32) string {
	r := (color >> 16) & 0xFF
	g := (color >> 8) & 0xFF
	b := color & 0xFF
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

type animationCompletionState struct {
	playing   bool
	playStart float64
	duration  float64
}

func computeAnimationCompletion(info TransitionInfo) (map[string]float64, map[string]int) {
	durations := make(map[string]float64)
	lastIndex := make(map[string]int)
	if len(info.Items) == 0 {
		return durations, lastIndex
	}
	states := make(map[string]*animationCompletionState)
	for idx, item := range info.Items {
		if item.Type != TransitionActionAnimation {
			continue
		}
		key := item.TargetID
		if key == "" {
			key = "_root"
		}
		lastIndex[key] = idx
		state := states[key]
		if state == nil {
			state = &animationCompletionState{playStart: -1}
			states[key] = state
		}
		if item.Tween != nil {
			continue
		}
		if item.Value.Playing {
			if !state.playing {
				state.playing = true
				state.playStart = item.Time
			}
		} else {
			if state.playing {
				if item.Time > state.playStart {
					state.duration += item.Time - state.playStart
				}
				state.playing = false
				state.playStart = -1
			}
		}
	}
	total := info.TotalDuration
	for key, state := range states {
		if state.playing {
			end := total
			if end < state.playStart {
				end = state.playStart
			}
			state.duration += end - state.playStart
			state.playing = false
		}
		durations[key] = state.duration * 1000
	}
	return durations, lastIndex
}
