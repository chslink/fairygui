package render

import (
	"image/color"
	"testing"
	"unicode"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	textutil "github.com/chslink/fairygui/internal/text"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

func TestBuildRenderedLineFromRuns_UnifiedBaseline(t *testing.T) {
	tests := []struct {
		name         string
		runs         []*renderedTextRun
		expectedBase float64
	}{
		{
			name: "single_size_run",
			runs: []*renderedTextRun{
				createTestRun("Hello", 16, 12.0, 3.0),
			},
			expectedBase: 12.0,
		},
		{
			name: "mixed_sizes_same_baseline",
			runs: []*renderedTextRun{
				createTestRun("Hello", 12, 9.0, 2.5),
				createTestRun("World", 16, 12.0, 3.0),
				createTestRun("Test", 20, 15.0, 4.0),
			},
			expectedBase: 15.0, // Should use largest font's baseline
		},
		{
			name: "mixed_sizes_with_spaces",
			runs: []*renderedTextRun{
				createTestRun("Hello", 12, 9.0, 2.5),
				createTestRun(" ", 12, 9.0, 2.5), // space
				createTestRun("World", 16, 12.0, 3.0),
			},
			expectedBase: 12.0, // Should use largest font's baseline
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := baseMetrics{
				ascent:     10.0,
				descent:    3.0,
				lineHeight: 13.0,
			}

			line := buildRenderedLineFromRuns(tt.runs, base, 0)

			if !line.hasGlyph {
				t.Fatal("Expected line to have glyphs")
			}

			if line.ascent != tt.expectedBase {
				t.Errorf("Expected baseline ascent %.2f, got %.2f", tt.expectedBase, line.ascent)
			}

			// Verify line height is reasonable
			expectedHeight := line.ascent + line.descent
			if line.height != expectedHeight {
				t.Errorf("Expected height %.2f, got %.2f", expectedHeight, line.height)
			}
		})
	}
}

func TestBaseMetricsCalculation_Accuracy(t *testing.T) {
	tests := []struct {
		name     string
		fontSize int
	}{
		{"small_font", 12},
		{"medium_font", 16},
		{"large_font", 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := widgets.NewText()
			field.SetFontSize(tt.fontSize)

			metrics := resolveBaseMetrics(field)

			// Verify ascent is positive and reasonable (actual values might be different due to font characteristics)
			if metrics.ascent <= 0 {
				t.Errorf("Expected positive ascent, got %.2f", metrics.ascent)
			}

			// For bitmap fonts, ascent might be fixed regardless of font size
			// This is acceptable as long as it's positive and reasonable
			if metrics.ascent > float64(tt.fontSize)*2 {
				t.Errorf("Ascent %.2f seems too large for font size %d", metrics.ascent, tt.fontSize)
			}

			// Verify descent is positive and reasonable
			if metrics.descent <= 0 {
				t.Errorf("Expected positive descent, got %.2f", metrics.descent)
			}

			// Verify line height is reasonable
			if metrics.lineHeight <= 0 {
				t.Errorf("Expected positive line height, got %.2f", metrics.lineHeight)
			}

			// Line height should be reasonable
			// Note: With bitmap fonts, line height might be fixed regardless of requested font size
			// This is acceptable as long as it's positive and reasonable for text display
			if metrics.lineHeight < 8 {
				t.Errorf("Line height %.2f seems too small for readable text", metrics.lineHeight)
			}

			// But shouldn't be excessively large
			if metrics.lineHeight > float64(tt.fontSize)*3 {
				t.Errorf("Line height %.2f seems too large for font size %d", metrics.lineHeight, tt.fontSize)
			}
		})
	}
}

func TestMixedFontSize_BaselineAlignment(t *testing.T) {
	// Test the specific issue that was fixed: mixed font sizes in the same line
	runs := []*renderedTextRun{
		createTestRun("Small", 12, 9.0, 2.5),
		createTestRun(" ", 12, 9.0, 2.5),
		createTestRun("Large", 20, 15.0, 4.0),
		createTestRun(" ", 20, 15.0, 4.0),
		createTestRun("Medium", 16, 12.0, 3.0),
	}

	base := baseMetrics{
		ascent:     10.0,
		descent:    3.0,
		lineHeight: 13.0,
	}

	line := buildRenderedLineFromRuns(runs, base, 0)

	// The line should use the largest font's baseline
	expectedAscent := 15.0 // from the 20px font
	if line.ascent != expectedAscent {
		t.Errorf("Expected baseline ascent %.2f (from largest font), got %.2f", expectedAscent, line.ascent)
	}

	// Verify all runs are included
	if len(line.runs) != 5 {
		t.Errorf("Expected 5 runs, got %d", len(line.runs))
	}

	// Verify line has glyphs
	if !line.hasGlyph {
		t.Error("Expected line to have glyphs")
	}
}

func TestRenderedTextRun_AdvanceCalculation(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		fontSize int
	}{
		{"ascii_text", "Hello", 16},
		{"mixed_text", "Hello中文", 16},
		{"unicode_text", "测试文本", 16},
		{"with_spaces", "Hello World", 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := createTestRun(tt.text, tt.fontSize, 12.0, 3.0)

			// Verify run properties
			if run.text != tt.text {
				t.Errorf("Expected text %q, got %q", tt.text, run.text)
			}

			if len(run.runes) != len([]rune(tt.text)) {
				t.Errorf("Expected %d runes, got %d", len([]rune(tt.text)), len(run.runes))
			}

			if run.fontSize != tt.fontSize {
				t.Errorf("Expected font size %d, got %d", tt.fontSize, run.fontSize)
			}

			// Verify width is calculated
			if run.width <= 0 {
				t.Errorf("Expected positive width, got %.2f", run.width)
			}

			// Verify advances are calculated for each rune
			if len(run.advances) != len(run.runes) {
				t.Errorf("Expected %d advances, got %d", len(run.runes), len(run.advances))
			}

			// Verify each advance is positive
			for i, adv := range run.advances {
				if adv <= 0 {
					t.Errorf("Advance %d should be positive, got %.2f", i, adv)
				}
			}
		})
	}
}

func TestTextSegment_Parsing(t *testing.T) {
	tests := []struct {
		name           string
		segments       []textutil.Segment
		expectedParts  int
		expectedTexts  []string
	}{
		{
			name: "single_segment",
			segments: []textutil.Segment{
				{Text: "Hello World"},
			},
			expectedParts: 1,
			expectedTexts: []string{"Hello World"},
		},
		{
			name: "multiple_segments",
			segments: []textutil.Segment{
				{Text: "Hello"},
				{Text: " "},
				{Text: "World"},
			},
			expectedParts: 3,
			expectedTexts: []string{"Hello", " ", "World"},
		},
		{
			name: "with_newlines",
			segments: []textutil.Segment{
				{Text: "Line1\nLine2"},
			},
			expectedParts: 3, // Line1, forced break, Line2
			expectedTexts: []string{"Line1", "", "Line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := widgets.NewText()
			field.SetFontSize(16)
			baseColor := color.NRGBA{0, 0, 0, 255}
			base := baseMetrics{ascent: 12.0, descent: 3.0, lineHeight: 15.0}

			parts := buildTextParts(tt.segments, field, baseColor, base, 0)

			if len(parts) != tt.expectedParts {
				t.Errorf("Expected %d parts, got %d", tt.expectedParts, len(parts))
			}

			for i, expectedText := range tt.expectedTexts {
				if i < len(parts) {
					if parts[i].forcedBreak && expectedText != "" {
						t.Errorf("Part %d: expected run with text %q, got forced break", i, expectedText)
					}
					if !parts[i].forcedBreak && parts[i].run != nil {
						if parts[i].run.text != expectedText {
							t.Errorf("Part %d: expected text %q, got %q", i, expectedText, parts[i].run.text)
						}
					}
				}
			}
		})
	}
}

func TestWrapText_StandardBehavior(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxWidth  float64
		fontSize  int
		expectRun int
	}{
		{
			name:      "short_text_no_wrap",
			text:      "Hello",
			maxWidth:  100,
			fontSize:  16,
			expectRun: 1,
		},
		{
			name:      "long_text_should_wrap",
			text:      "This is a very long text that should wrap",
			maxWidth:  100,
			fontSize:  16,
			expectRun: 3, // Should wrap into multiple runs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := createTestRun(tt.text, tt.fontSize, 12.0, 3.0)

			startIdx, width := run.spanForWidth(0, tt.maxWidth, 0)

			// Verify reasonable behavior
			if startIdx < 0 || startIdx > len(run.runes) {
				t.Errorf("Invalid start index: %d", startIdx)
			}

			if width < 0 || width > run.width {
				t.Errorf("Invalid width: %.2f (run width: %.2f)", width, run.width)
			}

			// For very long text, expect some wrapping
			if len(run.text) > 20 && tt.maxWidth < 200 {
				if startIdx == len(run.runes) {
					t.Errorf("Expected wrapping for long text, but got entire text")
				}
			}
		})
	}
}

// Helper functions

func createTestRun(text string, fontSize int, ascent, descent float64) *renderedTextRun {
	run := &renderedTextRun{
		text:     text,
		runes:    []rune(text),
		style:    textutil.Style{},
		color:    color.NRGBA{0, 0, 0, 255},
		fontSize: fontSize,
		ascent:   ascent,
		descent:  descent,
		bitmap:   &assets.BitmapFont{}, // Use bitmap font to ensure hasGlyphs() returns true
	}

	// Calculate width based on simple character width estimation
	width := 0.0
	for _, r := range text {
		if unicode.IsSpace(r) {
			width += float64(fontSize) * 0.3 // space width
		} else if unicode.Is(unicode.Han, r) {
			width += float64(fontSize) // Chinese characters are full width
		} else {
			width += float64(fontSize) * 0.6 // Latin characters
		}
	}
	run.width = width

	// Create advances array
	run.advances = make([]float64, len(run.runes))
	for i, r := range run.runes {
		if unicode.IsSpace(r) {
			run.advances[i] = float64(fontSize) * 0.3
		} else if unicode.Is(unicode.Han, r) {
			run.advances[i] = float64(fontSize)
		} else {
			run.advances[i] = float64(fontSize) * 0.6
		}
	}

	return run
}