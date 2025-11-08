package widgets

import (
	"math"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

const (
	movieClipStatusNone = iota
	movieClipStatusNextLoop
	movieClipStatusEnding
	movieClipStatusEnded
)

// PlayEndHandler mirrors FairyGUI's SimpleHandler callback invoked after looping completes.
type PlayEndHandler func()

// GMovieClip renders animated frame sequences stored in PackageItem resources.
type GMovieClip struct {
	*core.GObject

	packageItem *assets.PackageItem
	color       string

	interval    int
	repeatDelay int
	swing       bool
	frames      []*assets.MovieClipFrame
	fillMethod  int
	fillOrigin  int
	fillClock   bool
	fillAmount  float64

	playing       bool
	frame         int
	timeScale     float64
	frameElapsed  float64
	repeatedCount int
	reversed      bool
	start         int
	end           int
	times         int
	endAt         int
	status        int
	endHandler    PlayEndHandler

	tickerCancel func()
	flip         FlipType
}

// NewMovieClip constructs an animated widget with default playback settings.
func NewMovieClip() *GMovieClip {
	obj := core.NewGObject()
	clip := &GMovieClip{
		GObject:   obj,
		color:     "#ffffff",
		playing:   true,
		timeScale: 1,
		start:     0,
		end:       -1,
		endAt:     -1,
	}
	obj.SetData(clip)
	if sprite := obj.DisplayObject(); sprite != nil {
		sprite.SetMouseEnabled(false)
		dispatcher := sprite.Dispatcher()
		dispatcher.On(laya.EventDisplay, func(*laya.Event) { clip.ensureTicker() })
		dispatcher.On(laya.EventUndisplay, func(*laya.Event) { clip.stopTicker() })
	}
	return clip
}

// SetPackageItem binds the package movie clip data to the widget.
func (m *GMovieClip) SetPackageItem(item *assets.PackageItem) {
	if m.packageItem == item {
		return
	}
	m.packageItem = item
	if item == nil {
		m.frames = nil
		m.interval = 0
		m.repeatDelay = 0
		m.swing = false
		m.SetSize(0, 0)
		m.stopTicker()
		m.frame = 0
		m.drawFrame()
		return
	}
	m.interval = item.Interval
	m.repeatDelay = item.RepeatDelay
	m.swing = item.Swing
	m.frames = item.Frames
	w := float64(item.Width)
	h := float64(item.Height)
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	if w > 0 || h > 0 {
		m.SetSize(w, h)
		m.SetSourceSize(w, h)
		m.SetInitSize(w, h)
	}
	if m.end < 0 || m.end >= m.frameCount() {
		m.end = m.frameCount() - 1
	}
	if m.endAt < 0 || m.endAt >= m.frameCount() {
		m.endAt = m.end
	}
	m.frame = clampInt(m.frame, 0, m.frameCount()-1)
	m.frameElapsed = 0
	m.repeatedCount = 0
	m.reversed = false
	m.status = movieClipStatusNone
	m.drawFrame()
	m.ensureTicker()
}

// PackageItem returns the bound movie clip package item.
func (m *GMovieClip) PackageItem() *assets.PackageItem {
	return m.packageItem
}

// SetColor updates the tint colour applied during rendering.
func (m *GMovieClip) SetColor(value string) {
	if value == "" {
		value = "#ffffff"
	}
	if m.color == value {
		return
	}
	m.color = value
	m.updateGraphics()
}

// Color reports the current tint colour string.
func (m *GMovieClip) Color() string {
	return m.color
}

// SetFill configures partial rendering using FairyGUI fill semantics.
func (m *GMovieClip) SetFill(method int, origin int, clockwise bool, amount float64) {
	if amount < 0 {
		amount = 0
	} else if amount > 1 {
		amount = 1
	}
	if m.fillMethod == method && m.fillOrigin == origin && m.fillClock == clockwise && math.Abs(m.fillAmount-amount) < 1e-6 {
		return
	}
	m.fillMethod = method
	m.fillOrigin = origin
	m.fillClock = clockwise
	m.fillAmount = amount
	m.updateGraphics()
}

// Fill returns the stored fill configuration.
func (m *GMovieClip) Fill() (method int, origin int, clockwise bool, amount float64) {
	return m.fillMethod, m.fillOrigin, m.fillClock, m.fillAmount
}

// Flip returns the current flip mode.
func (m *GMovieClip) Flip() FlipType {
	return m.flip
}

// SetFlip updates the flip mode and requests a redraw.
func (m *GMovieClip) SetFlip(value FlipType) {
	if m.flip == value {
		return
	}
	m.flip = value
	m.updateGraphics()
}

// Playing reports whether automatic playback is active.
func (m *GMovieClip) Playing() bool {
	return m.playing
}

// SetPlaying toggles automatic playback and ticker registration.
func (m *GMovieClip) SetPlaying(v bool) {
	if m.playing == v {
		return
	}
	m.playing = v
	if v {
		m.ensureTicker()
	} else {
		m.stopTicker()
	}
}

// Frame returns the current frame index.
func (m *GMovieClip) Frame() int {
	return m.frame
}

// SetFrame jumps to the specified frame.
func (m *GMovieClip) SetFrame(index int) {
	if m.frameCount() == 0 {
		m.frame = index
		m.frameElapsed = 0
		m.updateGraphics()
		return
	}
	clamped := clampInt(index, 0, m.frameCount()-1)
	if m.frame != clamped {
		m.frame = clamped
		m.frameElapsed = 0
		m.drawFrame()
	}
}

// TimeScale returns the playback time scale factor.
func (m *GMovieClip) TimeScale() float64 {
	return m.timeScale
}

// SetTimeScale adjusts the playback speed multiplier.
func (m *GMovieClip) SetTimeScale(scale float64) {
	if scale == 0 {
		scale = 1
	}
	m.timeScale = scale
}

// DeltaTime exposes accumulated delta for gear compatibility (always zero in this port).
func (m *GMovieClip) DeltaTime() float64 {
	return 0
}

// SetDeltaTime advances the clip by the specified milliseconds.
func (m *GMovieClip) SetDeltaTime(ms float64) {
	if ms <= 0 {
		return
	}
	m.Advance(ms)
}

// Rewind resets playback to the first frame.
func (m *GMovieClip) Rewind() {
	m.frame = 0
	m.frameElapsed = 0
	m.reversed = false
	m.repeatedCount = 0
	m.status = movieClipStatusNone
	m.drawFrame()
}

// Advance steps the clip forward by the provided delta in milliseconds.
func (m *GMovieClip) Advance(ms float64) {
	if ms <= 0 {
		return
	}
	m.advance(ms)
}

// SyncStatus copies playback progress from another movie clip.
func (m *GMovieClip) SyncStatus(other *GMovieClip) {
	if other == nil {
		return
	}
	m.frame = other.frame
	m.frameElapsed = other.frameElapsed
	m.reversed = other.reversed
	m.repeatedCount = other.repeatedCount
	m.status = other.status
	m.drawFrame()
}

// SetPlaySettings configures loop range and completion callback.
func (m *GMovieClip) SetPlaySettings(start, end, times, endAt int, handler PlayEndHandler) {
	if start < 0 {
		start = 0
	}
	m.start = start
	m.end = end
	if m.end == -1 || m.end >= m.frameCount() {
		m.end = m.frameCount() - 1
	}
	m.times = times
	m.endAt = endAt
	if m.endAt == -1 {
		m.endAt = m.end
	}
	m.status = movieClipStatusNone
	m.endHandler = handler
	m.repeatedCount = 0
	m.reversed = false
	m.frameElapsed = 0
	m.SetFrame(start)
}

// SetupBeforeAdd applies serialized configuration from the component buffer.
func (m *GMovieClip) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if m == nil || buf == nil {
		return
	}

	// 首先调用父类GObject处理基础属性
	m.GObject.SetupBeforeAdd(buf, beginPos)

	// 然后处理GMovieClip特定属性（block 5）
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(beginPos, 5) {
		return
	}
	if buf.ReadBool() {
		m.SetColor(buf.ReadColorString(true))
	}
	m.SetFlip(FlipType(buf.ReadByte()))
	m.SetFrame(int(buf.ReadInt32()))
	m.SetPlaying(buf.ReadBool())
}

// OwnerSizeChanged requests a redraw when the parent component resizes the widget.
func (m *GMovieClip) OwnerSizeChanged(_, _ float64) {
	m.updateGraphics()
}

// CurrentFrame returns the metadata for the active frame.
func (m *GMovieClip) CurrentFrame() *assets.MovieClipFrame {
	if m == nil || len(m.frames) == 0 {
		return nil
	}
	idx := clampInt(m.frame, 0, len(m.frames)-1)
	return m.frames[idx]
}

func (m *GMovieClip) frameCount() int {
	return len(m.frames)
}

func (m *GMovieClip) ensureTicker() {
	if m == nil {
		return
	}
	if !m.playing || m.frameCount() == 0 {
		return
	}
	if m.tickerCancel != nil {
		return
	}
	m.tickerCancel = core.RegisterTicker(func(delta time.Duration) {
		if delta <= 0 {
			return
		}
		m.advance(float64(delta) / float64(time.Millisecond))
	})
}

func (m *GMovieClip) stopTicker() {
	if m == nil || m.tickerCancel == nil {
		return
	}
	m.tickerCancel()
	m.tickerCancel = nil
}

func (m *GMovieClip) advance(deltaMS float64) {
	if m == nil || !m.playing || m.frameCount() == 0 || m.status == movieClipStatusEnded {
		return
	}
	if deltaMS > 100 {
		deltaMS = 100
	}
	deltaMS *= m.timeScale
	if deltaMS <= 0 {
		return
	}
	m.frameElapsed += deltaMS
	frame := m.CurrentFrame()
	threshold := float64(m.interval)
	if frame != nil {
		threshold += float64(frame.AddDelay)
	}
	if m.frame == 0 && m.repeatedCount > 0 {
		threshold += float64(m.repeatDelay)
	}
	if m.frameElapsed < threshold {
		return
	}
	m.frameElapsed -= threshold
	if m.frameElapsed > float64(m.interval) {
		m.frameElapsed = float64(m.interval)
	}
	count := m.frameCount()
	if count <= 0 {
		return
	}
	if m.swing {
		if m.reversed {
			m.frame--
			if m.frame <= 0 {
				m.frame = 0
				m.repeatedCount++
				m.reversed = !m.reversed
			}
		} else {
			m.frame++
			if m.frame > count-1 {
				if count > 1 {
					m.frame = int(math.Max(0, float64(count-2)))
				} else {
					m.frame = 0
				}
				m.repeatedCount++
				m.reversed = !m.reversed
			}
		}
	} else {
		m.frame++
		if m.frame > count-1 {
			m.frame = 0
			m.repeatedCount++
		}
	}

	switch m.status {
	case movieClipStatusNextLoop:
		m.frame = m.start
		m.frameElapsed = 0
		m.status = movieClipStatusNone
	case movieClipStatusEnding:
		m.frame = m.endAt
		m.frameElapsed = 0
		m.status = movieClipStatusEnded
		if m.endHandler != nil {
			handler := m.endHandler
			m.endHandler = nil
			handler()
		}
	default:
		if m.frame == m.end {
			if m.times > 0 {
				m.times--
				if m.times == 0 {
					m.status = movieClipStatusEnding
					m.frame = m.end
				} else {
					m.status = movieClipStatusNextLoop
				}
			} else {
				m.status = movieClipStatusNextLoop
			}
		}
	}

	m.drawFrame()
}

func (m *GMovieClip) drawFrame() {
	// IMPORTANT: In Laya, MovieClip inherits from Image, and Image automatically
	// resizes when texture changes. But in our Go implementation, we maintain
	// the MovieClip's overall size and use frame offsets for alignment.
	// This matches the FairyGUI design where MovieClip has a fixed size
	// and frames are positioned within that area using offsets.
	m.updateGraphics()
}

func (m *GMovieClip) updateGraphics() {
	if m == nil || m.GObject == nil {
		return
	}
	if sprite := m.DisplayObject(); sprite != nil {
		sprite.Repaint()
	}
}

func clampInt(value, min, max int) int {
	if max < min {
		return min
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
