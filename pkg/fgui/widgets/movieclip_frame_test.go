package widgets

import (
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// TestGMovieClipFrameOffsetHandling tests that MovieClip properly handles
// frames with different offsets for correct alignment during rendering
func TestGMovieClipFrameOffsetHandling(t *testing.T) {
	clip := NewMovieClip()

	// Create a MovieClip with frames of different sizes and offsets
	item := &assets.PackageItem{
		Interval: 10,
		Width:    50,  // Overall size
		Height:   50,
		Frames: []*assets.MovieClipFrame{
			{
				Width:   40, // Smaller frame
				Height:  30,
				OffsetX: 5,  // Offset to center it
				OffsetY: 10,
			},
			{
				Width:   50, // Larger frame
				Height:  50,
				OffsetX: 0,
				OffsetY: 0,
			},
			{
				Width:   35, // Another different size
				Height:  40,
				OffsetX: 8,
				OffsetY: 5,
			},
		},
	}

	clip.SetPackageItem(item)

	// Test initial frame (frame 0)
	if clip.Frame() != 0 {
		t.Fatalf("expected initial frame 0, got %d", clip.Frame())
	}

	// Verify that MovieClip maintains the overall size from the package item
	expectedWidth := 50.0
	expectedHeight := 50.0
	if clip.Width() != expectedWidth || clip.Height() != expectedHeight {
		t.Fatalf("expected MovieClip size %.0fx%.0f, got %.0fx%.0f",
			expectedWidth, expectedHeight, clip.Width(), clip.Height())
	}

	// Test frame data is correctly set
	frame := clip.CurrentFrame()
	if frame == nil {
		t.Fatal("current frame should not be nil")
	}
	if frame.Width != 40 || frame.Height != 30 {
		t.Fatalf("frame 0 size mismatch: expected 40x30, got %dx%d", frame.Width, frame.Height)
	}
	if frame.OffsetX != 5 || frame.OffsetY != 10 {
		t.Fatalf("frame 0 offset mismatch: expected (5,10), got (%d,%d)", frame.OffsetX, frame.OffsetY)
	}

	// Advance to frame 1
	clip.SetFrame(1)

	// MovieClip size should remain the same
	if clip.Width() != 50.0 || clip.Height() != 50.0 {
		t.Fatalf("MovieClip size changed: expected 50x50, got %.0fx%.0f", clip.Width(), clip.Height())
	}

	// But frame data should update
	frame = clip.CurrentFrame()
	if frame == nil {
		t.Fatal("current frame should not be nil")
	}
	if frame.Width != 50 || frame.Height != 50 {
		t.Fatalf("frame 1 size mismatch: expected 50x50, got %dx%d", frame.Width, frame.Height)
	}
	if frame.OffsetX != 0 || frame.OffsetY != 0 {
		t.Fatalf("frame 1 offset mismatch: expected (0,0), got (%d,%d)", frame.OffsetX, frame.OffsetY)
	}

	// Advance to frame 2
	clip.SetFrame(2)

	// MovieClip size should still remain the same
	if clip.Width() != 50.0 || clip.Height() != 50.0 {
		t.Fatalf("MovieClip size changed: expected 50x50, got %.0fx%.0f", clip.Width(), clip.Height())
	}

	// Frame data should update
	frame = clip.CurrentFrame()
	if frame == nil {
		t.Fatal("current frame should not be nil")
	}
	if frame.Width != 35 || frame.Height != 40 {
		t.Fatalf("frame 2 size mismatch: expected 35x40, got %dx%d", frame.Width, frame.Height)
	}
	if frame.OffsetX != 8 || frame.OffsetY != 5 {
		t.Fatalf("frame 2 offset mismatch: expected (8,5), got (%d,%d)", frame.OffsetX, frame.OffsetY)
	}
}

// TestGMovieClipFrameAdvanceWithOffsets tests that advancing frames
// properly updates frame data including offsets
func TestGMovieClipFrameAdvanceWithOffsets(t *testing.T) {
	clip := NewMovieClip()

	item := &assets.PackageItem{
		Interval: 20,
		Frames: []*assets.MovieClipFrame{
			{Width: 20, Height: 20, OffsetX: 2, OffsetY: 3},
			{Width: 30, Height: 25, OffsetX: 0, OffsetY: 5},
			{Width: 25, Height: 30, OffsetX: 10, OffsetY: 0},
		},
	}

	clip.SetPackageItem(item)

	// Test initial frame data
	frame := clip.CurrentFrame()
	if frame.Width != 20 || frame.Height != 20 || frame.OffsetX != 2 || frame.OffsetY != 3 {
		t.Fatalf("initial frame data mismatch: expected (20,20,2,3), got (%d,%d,%d,%d)",
			frame.Width, frame.Height, frame.OffsetX, frame.OffsetY)
	}

	// Advance by interval to trigger frame change
	clip.Advance(20)

	// Should be on frame 1 now
	if clip.Frame() != 1 {
		t.Fatalf("expected frame 1, got %d", clip.Frame())
	}

	// Frame data should update
	frame = clip.CurrentFrame()
	if frame.Width != 30 || frame.Height != 25 || frame.OffsetX != 0 || frame.OffsetY != 5 {
		t.Fatalf("frame 1 data mismatch: expected (30,25,0,5), got (%d,%d,%d,%d)",
			frame.Width, frame.Height, frame.OffsetX, frame.OffsetY)
	}

	// Advance again
	clip.Advance(20)

	// Should be on frame 2 now
	if clip.Frame() != 2 {
		t.Fatalf("expected frame 2, got %d", clip.Frame())
	}

	// Frame data should update
	frame = clip.CurrentFrame()
	if frame.Width != 25 || frame.Height != 30 || frame.OffsetX != 10 || frame.OffsetY != 0 {
		t.Fatalf("frame 2 data mismatch: expected (25,30,10,0), got (%d,%d,%d,%d)",
			frame.Width, frame.Height, frame.OffsetX, frame.OffsetY)
	}
}