//go:build ebiten

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
	segments := []textutil.Segment{
		{Text: "A", Style: baseStyle},
		{Text: "B", Style: baseStyle},
	}
	line := buildRenderedLine(segments, field, baseColor, base, float64(field.LetterSpacing()))
	if len(line.runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(line.runs))
	}
	expectedWidth := line.runs[0].width + line.runs[1].width + float64(field.LetterSpacing())
	if diff := line.width - expectedWidth; diff < -0.5 || diff > 0.5 {
		t.Fatalf("line width mismatch: got %.2f expected %.2f", line.width, expectedWidth)
	}
}

func TestSplitSegmentsIntoLines(t *testing.T) {
	base := textutil.Style{Color: "#ffffff", FontSize: 12}
	segments := []textutil.Segment{
		{Text: "Hello\nWorld", Style: base},
	}
	lines := splitSegmentsIntoLines(segments)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if len(lines[0]) != 1 || lines[0][0].Text != "Hello" {
		t.Fatalf("unexpected first line: %#v", lines[0])
	}
	if len(lines[1]) != 1 || lines[1][0].Text != "World" {
		t.Fatalf("unexpected second line: %#v", lines[1])
	}
}
