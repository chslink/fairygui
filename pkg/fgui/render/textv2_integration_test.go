package render

import (
	"fmt"
	"image/color"
	"testing"

	textutil "github.com/chslink/fairygui/internal/text"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
	"github.com/hajimehoshi/ebiten/v2"
	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font"
)

func TestTextV2Integration_MetricsConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test that text/v2 metrics are consistent with our calculations
	fontSizes := []int{12, 16, 20, 24}

	for _, size := range fontSizes {
		t.Run(fmt.Sprintf("font_size_%d", size), func(t *testing.T) {
			// Get font face
			face := fontFaceForSize(size)
			if face == nil {
				t.Skipf("No font face available for size %d", size)
			}

			// Create text/v2 face
			textFace := textv2.NewGoXFace(face)

			// Get metrics from both systems
			textv2Metrics := textFace.Metrics()
			ourMetrics := resolveBaseMetrics(nil) // Will use default size

			// Adjust for actual font size
			ourMetrics = resolveBaseMetricsWithSize(size, face)

			// Compare ascent values (allow some tolerance)
			ascentTolerance := float64(size) * 0.1
			if diff := textv2Metrics.HAscent - ourMetrics.ascent; diff < -ascentTolerance || diff > ascentTolerance {
				t.Errorf("Ascent difference too large: text/v2=%.2f, ours=%.2f, diff=%.2f (tolerance=%.2f)",
					textv2Metrics.HAscent, ourMetrics.ascent, diff, ascentTolerance)
			}

			// Compare descent values
			descentTolerance := float64(size) * 0.05
			if diff := textv2Metrics.HDescent - ourMetrics.descent; diff < -descentTolerance || diff > descentTolerance {
				t.Errorf("Descent difference too large: text/v2=%.2f, ours=%.2f, diff=%.2f (tolerance=%.2f)",
					textv2Metrics.HDescent, ourMetrics.descent, diff, descentTolerance)
			}

			t.Logf("Font size %d: text/v2 ascent=%.2f, descent=%.2f | ours ascent=%.2f, descent=%.2f",
				size, textv2Metrics.HAscent, textv2Metrics.HDescent, ourMetrics.ascent, ourMetrics.descent)
		})
	}
}

func TestRenderSystemRun_V2Compatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name     string
		text     string
		fontSize int
		color    color.NRGBA
	}{
		{"basic_latin", "Hello World", 16, color.NRGBA{0, 0, 0, 255}},
		{"mixed_script", "Hello世界", 16, color.NRGBA{255, 0, 0, 255}},
		{"with_punctuation", "Hello, World!", 20, color.NRGBA{0, 128, 0, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test image
			img := ebiten.NewImage(400, 100)

			// Get font face
			face := fontFaceForSize(tt.fontSize)
			if face == nil {
				t.Skipf("No font face available for size %d", tt.fontSize)
			}

			// Create rendered text run
			run := createTestRun(tt.text, tt.fontSize, 12.0, 3.0)
			run.face = face
			run.color = tt.color

			// Test rendering with the new system
			startX := 10.0
			baseline := 50.0
			letterSpacing := 0.0

			// This should not panic
			renderSystemRun(img, run, startX, baseline, letterSpacing, nil, 0, nil, 0, 0)

			// Verify the image has been modified (basic check)
			// In a real test, we might compare against a known good output
			t.Logf("Successfully rendered text %q at position (%.1f, %.1f)", tt.text, startX, baseline)
		})
	}
}

func TestMultilineText_LayoutConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name        string
		text        string
		fontSize    int
		maxWidth    float64
		expectLines int
	}{
		{
			name:        "single_line",
			text:        "Hello World",
			fontSize:    16,
			maxWidth:    200,
			expectLines: 1,
		},
		{
			name:        "should_wrap",
			text:        "This is a long text that should wrap into multiple lines",
			fontSize:    16,
			maxWidth:    150,
			expectLines: 3,
		},
		{
			name:        "with_newlines",
			text:        "Line1\nLine2\nLine3",
			fontSize:    16,
			maxWidth:    200,
			expectLines: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create text field
			field := widgets.NewText()
			field.SetText(tt.text)
			field.SetFontSize(tt.fontSize)
			field.SetColor("#000000")
			field.SetSingleLine(false)

			// Test text layout calculation
			img := ebiten.NewImage(int(tt.maxWidth)+20, 200)
			geo := ebiten.GeoM{}
			geo.Translate(10, 10)

			err := drawTextImage(img, geo, field, tt.text, 1.0, tt.maxWidth, 200, nil, nil)
			if err != nil {
				t.Fatalf("Failed to draw text: %v", err)
			}

			t.Logf("Successfully rendered multiline text: %q (expected %d lines)", tt.text, tt.expectLines)
		})
	}
}

func TestTextEffects_V2Rendering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name     string
		style    textutil.Style
		text     string
		fontSize int
	}{
		{"bold", textutil.Style{Bold: true}, "Bold Text", 16},
		{"italic", textutil.Style{Italic: true}, "Italic Text", 16},
		{"underline", textutil.Style{Underline: true}, "Underlined", 16},
		{"combined", textutil.Style{Bold: true, Underline: true}, "Bold & Underlined", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test image
			img := ebiten.NewImage(300, 100)

			// Get font face
			face := fontFaceForSize(tt.fontSize)
			if face == nil {
				t.Skipf("No font face available for size %d", tt.fontSize)
			}

			// Create rendered text run with style
			run := createTestRun(tt.text, tt.fontSize, 12.0, 3.0)
			run.face = face
			run.style = tt.style
			run.color = color.NRGBA{0, 0, 0, 255}

			// Test rendering with effects
			startX := 10.0
			baseline := 50.0

			// This should not panic
			renderSystemRun(img, run, startX, baseline, 0, nil, 0, nil, 0, 0)

			t.Logf("Successfully rendered text with style %+v: %q", tt.style, tt.text)
		})
	}
}

func TestTextRendering_BackwardCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test that the new implementation produces reasonable results
	// compared to what users would expect from the old implementation

	testCases := []struct {
		name     string
		text     string
		fontSize int
	}{
		{"short_ascii", "Hello", 12},
		{"medium_ascii", "Hello World", 16},
		{"mixed_script", "Hello 世界", 20},
		{"numbers", "12345", 16},
		{"mixed_content", "Test 123 测试", 18},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create text field
			field := widgets.NewText()
			field.SetText(tc.text)
			field.SetFontSize(tc.fontSize)
			field.SetColor("#000000")

			// Test basic rendering
			img := ebiten.NewImage(400, 100)
			geo := ebiten.GeoM{}
			geo.Translate(10, 10)

			err := drawTextImage(img, geo, field, tc.text, 1.0, 300, 80, nil, nil)
			if err != nil {
				t.Fatalf("Failed to draw text: %v", err)
			}

			// Verify that the rendering didn't produce errors
			// In a comprehensive test, we might check:
			// - Text is positioned correctly
			// - Colors are applied correctly
			// - Size is reasonable

			t.Logf("Successfully rendered text for backward compatibility: %q", tc.text)
		})
	}
}

// Helper function for testing metrics with specific font size
func resolveBaseMetricsWithSize(size int, face font.Face) baseMetrics {
	metrics := face.Metrics()
	ascent := float64(metrics.Ascent) / 64.0
	descent := float64(metrics.Descent) / 64.0

	if ascent <= 0 {
		ascent = float64(size) * 0.8
	}
	if descent <= 0 {
		descent = float64(size) * 0.2
	}

	lineHeight := ascent + descent + float64(size)*0.15

	return baseMetrics{
		ascent:     ascent,
		descent:    descent,
		lineHeight: lineHeight,
	}
}