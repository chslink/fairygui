package widgets

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

func TestGMovieClipAdvanceSimple(t *testing.T) {
	clip := NewMovieClip()
	item := &assets.PackageItem{
		Interval: 20,
		Frames: []*assets.MovieClipFrame{
			{Width: 20, Height: 20},
			{Width: 20, Height: 20},
		},
	}
	clip.SetPackageItem(item)
	if clip.Frame() != 0 {
		t.Fatalf("expected initial frame 0, got %d", clip.Frame())
	}
	clip.Advance(20)
	if clip.Frame() != 1 {
		t.Fatalf("expected frame 1 after advance, got %d", clip.Frame())
	}
	clip.Advance(20)
	if clip.Frame() != 0 {
		t.Fatalf("expected frame loop to 0, got %d", clip.Frame())
	}
}

func TestGMovieClipTimeScale(t *testing.T) {
	clip := NewMovieClip()
	item := &assets.PackageItem{
		Interval: 10,
		Frames: []*assets.MovieClipFrame{
			{Width: 10, Height: 10},
			{Width: 10, Height: 10},
		},
	}
	clip.SetPackageItem(item)
	clip.Rewind()
	clip.SetTimeScale(2)
	clip.Advance(5)
	if clip.Frame() != 1 {
		t.Fatalf("expected time-scaled advance to reach frame 1, got %d", clip.Frame())
	}
}

func TestGMovieClipPlaySettingsEndHandler(t *testing.T) {
	clip := NewMovieClip()
	item := &assets.PackageItem{
		Interval: 15,
		Frames: []*assets.MovieClipFrame{
			{Width: 16, Height: 16},
			{Width: 16, Height: 16},
		},
	}
	clip.SetPackageItem(item)
	done := false
	clip.SetPlaySettings(0, 1, 1, 1, func() { done = true })
	clip.Advance(15)
	if clip.Frame() != 1 {
		t.Fatalf("expected to land on frame 1, got %d", clip.Frame())
	}
	clip.Advance(15)
	if !done {
		t.Fatalf("expected end handler to run after configured loops")
	}
	if clip.Frame() != 1 {
		t.Fatalf("expected frame to remain at endAt, got %d", clip.Frame())
	}
	clip.Advance(15)
	if clip.Frame() != 1 {
		t.Fatalf("expected frame to stay ended, got %d", clip.Frame())
	}
}

func TestGMovieClipSetupBeforeAdd(t *testing.T) {
	clip := NewMovieClip()
	item := &assets.PackageItem{
		Interval: 10,
		Frames: []*assets.MovieClipFrame{
			{Width: 12, Height: 12},
			{Width: 12, Height: 12},
			{Width: 12, Height: 12},
		},
	}
	clip.SetPackageItem(item)

	data := make([]byte, 2+6*2+11)
	data[0] = 6
	data[1] = 1
	blockOffset := uint16(2 + 6*2)
	binary.BigEndian.PutUint16(data[2+5*2:], blockOffset)
	pos := int(blockOffset)
	data[pos] = 1
	pos++
	data[pos+0] = 0x12
	data[pos+1] = 0x34
	data[pos+2] = 0x56
	data[pos+3] = 0xFF
	pos += 4
	data[pos] = 0
	pos++
	binary.BigEndian.PutUint32(data[pos:], uint32(2))
	pos += 4
	data[pos] = 0

	buf := utils.NewByteBuffer(data)
	clip.SetupBeforeAdd(buf, 0)
	if clip.Color() != "#123456" {
		t.Fatalf("expected color #123456, got %q", clip.Color())
	}
	if clip.Frame() != 2 {
		t.Fatalf("expected frame 2 from setup, got %d", clip.Frame())
	}
	if clip.Flip() != FlipTypeNone {
		t.Fatalf("expected default flip none, got %v", clip.Flip())
	}
	if clip.Playing() {
		t.Fatalf("expected playing flag to be disabled by buffer data")
	}
	clip.SetFlip(FlipTypeHorizontal)
	if clip.Flip() != FlipTypeHorizontal {
		t.Fatalf("expected horizontal flip after setter")
	}
	clip.SetFill(1, 2, true, 1.5)
	method, origin, clockwise, amount := clip.Fill()
	if method != 1 || origin != 2 || !clockwise {
		t.Fatalf("unexpected fill metadata: method=%d origin=%d clockwise=%v", method, origin, clockwise)
	}
	if math.Abs(amount-1) > 1e-6 {
		t.Fatalf("expected fill amount clamped to 1, got %.3f", amount)
	}
}
