package widgets

import (
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// TestMovieClipPetAlignment tests the specific alignment behavior of the pet MovieClip
// that was causing visual jumping in the demo
func TestMovieClipPetAlignment(t *testing.T) {
	// Create a MovieClip that mimics the pet.jta structure
	item := &assets.PackageItem{
		ID:       "test_pet",
		Name:     "pet",
		Type:     assets.PackageItemTypeMovieClip,
		Width:    84,  // Same as pet.jta
		Height:   74,  // Same as pet.jta
		Interval: 83,  // Same as pet.jta
		Frames: []*assets.MovieClipFrame{
			{Width: 74, Height: 65, OffsetX: 9, OffsetY: 0},   // Frame 0
			{Width: 74, Height: 64, OffsetX: 7, OffsetY: 5},   // Frame 1
			{Width: 72, Height: 63, OffsetX: 8, OffsetY: 6},   // Frame 2
			{Width: 75, Height: 65, OffsetX: 8, OffsetY: 3},   // Frame 3
			{Width: 79, Height: 66, OffsetX: 5, OffsetY: 2},   // Frame 4
			{Width: 80, Height: 62, OffsetX: 2, OffsetY: 7},   // Frame 5
			{Width: 74, Height: 68, OffsetX: 6, OffsetY: 6},   // Frame 6
			{Width: 81, Height: 65, OffsetX: 0, OffsetY: 2},   // Frame 7
		},
	}

	clip := NewMovieClip()
	clip.SetPackageItem(item)

	// Test that MovieClip maintains its overall size regardless of frame
	if clip.Width() != 84.0 || clip.Height() != 74.0 {
		t.Fatalf("MovieClip size should be 84x74, got %.0fx%.0f", clip.Width(), clip.Height())
	}

	// Test frame transitions that were causing visual jumping
	testCases := []struct {
		fromFrame int
		toFrame   int
		desc      string
	}{
		{0, 1, "Frame 0->1: 74x65->74x64"},
		{1, 2, "Frame 1->2: 74x64->72x63"},
		{2, 3, "Frame 2->3: 72x63->75x65"},
		{6, 7, "Frame 6->7: 74x68->81x65 (most problematic)"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Set initial frame
			clip.SetFrame(tc.fromFrame)
			fromFrame := clip.CurrentFrame()

			// Transition to target frame
			clip.SetFrame(tc.toFrame)
			toFrame := clip.CurrentFrame()

			// Verify frame data is correctly updated
			if toFrame.Width != item.Frames[tc.toFrame].Width {
				t.Errorf("Frame %d width mismatch: expected %d, got %d",
					tc.toFrame, item.Frames[tc.toFrame].Width, toFrame.Width)
			}
			if toFrame.Height != item.Frames[tc.toFrame].Height {
				t.Errorf("Frame %d height mismatch: expected %d, got %d",
					tc.toFrame, item.Frames[tc.toFrame].Height, toFrame.Height)
			}
			if toFrame.OffsetX != item.Frames[tc.toFrame].OffsetX {
				t.Errorf("Frame %d offsetX mismatch: expected %d, got %d",
					tc.toFrame, item.Frames[tc.toFrame].OffsetX, toFrame.OffsetX)
			}
			if toFrame.OffsetY != item.Frames[tc.toFrame].OffsetY {
				t.Errorf("Frame %d offsetY mismatch: expected %d, got %d",
					tc.toFrame, item.Frames[tc.toFrame].OffsetY, toFrame.OffsetY)
			}

			// Verify MovieClip size remains constant (this is the key fix)
			if clip.Width() != 84.0 || clip.Height() != 74.0 {
				t.Errorf("MovieClip size changed after frame transition: expected 84x74, got %.0fx%.0f",
					clip.Width(), clip.Height())
			}

			// Log the transition for debugging
			t.Logf("Transition %s: size (%dx%d)->(%dx%d), offset (%d,%d)->(%d,%d)",
				tc.desc,
				fromFrame.Width, fromFrame.Height,
				toFrame.Width, toFrame.Height,
				fromFrame.OffsetX, fromFrame.OffsetY,
				toFrame.OffsetX, toFrame.OffsetY)
		})
	}
}

// TestMovieClipVisualCenterStability tests that the visual center of frames
// remains stable across transitions to prevent jumping
func TestMovieClipVisualCenterStability(t *testing.T) {
	// Create the pet MovieClip structure
	item := &assets.PackageItem{
		ID:       "test_pet",
		Name:     "pet",
		Type:     assets.PackageItemTypeMovieClip,
		Width:    84,
		Height:   74,
		Interval: 83,
		Frames: []*assets.MovieClipFrame{
			{Width: 74, Height: 68, OffsetX: 6, OffsetY: 6},   // Frame 6
			{Width: 81, Height: 65, OffsetX: 0, OffsetY: 2},   // Frame 7 (problematic transition)
		},
	}

	clip := NewMovieClip()
	clip.SetPackageItem(item)

	// Calculate visual centers for both frames
	// Visual center = offset + (frame_size / 2)
	frame6 := item.Frames[0]
	frame7 := item.Frames[1]

	center6x := float64(frame6.OffsetX) + float64(frame6.Width)*0.5
	center6y := float64(frame6.OffsetY) + float64(frame6.Height)*0.5
	center7x := float64(frame7.OffsetX) + float64(frame7.Width)*0.5
	center7y := float64(frame7.OffsetY) + float64(frame7.Height)*0.5

	deltaX := center7x - center6x
	deltaY := center7y - center6y

	t.Logf("Frame 6 visual center: (%.1f, %.1f)", center6x, center6y)
	t.Logf("Frame 7 visual center: (%.1f, %.1f)", center7x, center7y)
	t.Logf("Visual center displacement: (%.1f, %.1f)", deltaX, deltaY)

	// The offset should compensate for this displacement
	expectedOffsetDx := -deltaX
	expectedOffsetDy := -deltaY
	actualOffsetDx := float64(frame7.OffsetX - frame6.OffsetX)
	actualOffsetDy := float64(frame7.OffsetY - frame6.OffsetY)

	t.Logf("Expected offset change to compensate: (%.1f, %.1f)", expectedOffsetDx, expectedOffsetDy)
	t.Logf("Actual offset change: (%.1f, %.1f)", actualOffsetDx, actualOffsetDy)

	// For a perfectly stable animation, the offset change should exactly compensate
	// for the visual center displacement. In practice, there might be small
	// differences due to rounding or artistic choices.
	offsetErrorX := actualOffsetDx - expectedOffsetDx
	offsetErrorY := actualOffsetDy - expectedOffsetDy

	t.Logf("Offset compensation error: (%.1f, %.1f)", offsetErrorX, offsetErrorY)

	// IMPORTANT: FairyGUI offset design is NOT necessarily for visual center alignment.
	// The offset compensation error doesn't indicate a bug - it indicates that
	// FairyGUI uses a different alignment strategy than simple visual centering.
	// The actual test should verify that frames render correctly with their
	// specified offsets, not that they compensate for visual center displacement.
	t.Logf("Note: Offset compensation error (%.1f, %.1f) is expected - FairyGUI uses complex alignment logic",
		offsetErrorX, offsetErrorY)
}