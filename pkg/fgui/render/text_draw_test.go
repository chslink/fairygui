package render

import (
	"testing"

	textutil "github.com/chslink/fairygui/internal/text"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

func TestBuildRenderedLineLetterSpacingAcrossSegments(t *testing.T) {
	field := widgets.NewText()
	field.SetLetterSpacing(2)
	field.SetFontSize(14)

	baseStyle, baseColor := deriveBaseStyle(field)
	base := resolveBaseMetrics(field)
	runA := buildRenderedRun(textutil.Segment{Text: "A", Style: baseStyle}, field, baseColor, base, float64(field.LetterSpacing()))
	runB := buildRenderedRun(textutil.Segment{Text: "B", Style: baseStyle}, field, baseColor, base, float64(field.LetterSpacing()))
	line := buildRenderedLineFromRuns([]*renderedTextRun{runA, runB}, base, float64(field.LetterSpacing()))
	if len(line.runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(line.runs))
	}
	expectedWidth := line.runs[0].width + line.runs[1].width + float64(field.LetterSpacing())
	if diff := line.width - expectedWidth; diff < -0.5 || diff > 0.5 {
		t.Fatalf("line width mismatch: got %.2f expected %.2f", line.width, expectedWidth)
	}
}

func TestWrapRenderedRunsBreaksLongText(t *testing.T) {
	field := widgets.NewText()
	field.SetFontSize(16)
	baseStyle, baseColor := deriveBaseStyle(field)
	base := resolveBaseMetrics(field)
	run := buildRenderedRun(textutil.Segment{Text: "ABCD", Style: baseStyle}, field, baseColor, base, 0)
	if run == nil {
		t.Fatalf("failed to build run")
	}
	firstWidth := run.advanceAt(0)
	parts := []textPart{{run: run}}
	wrapped := wrapRenderedRuns(parts, firstWidth+0.1, 0, true)
	if len(wrapped) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(wrapped))
	}
	sample := []rune("ABCD")
	for i, line := range wrapped {
		if len(line) != 1 {
			t.Fatalf("line %d expected 1 segment, got %d", i, len(line))
		}
		if got := line[0].text; got != string(sample[i]) {
			t.Fatalf("line %d text mismatch: %q", i, got)
		}
	}
}
