package widgets

import (
	"math"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestProgressBarValueClamp(t *testing.T) {
	bar := NewProgressBar()
	bar.SetSize(200, 20)
	bar.SetTemplateComponent(makeProgressTemplate(false))
	bar.SetMin(10)
	bar.SetMax(20)

	bar.SetValue(5)
	if bar.Value() != 10 {
		t.Fatalf("expected clamp to min, got %.2f", bar.Value())
	}
	bar.SetValue(25)
	if bar.Value() != 20 {
		t.Fatalf("expected clamp to max, got %.2f", bar.Value())
	}
}

func TestProgressBarFill(t *testing.T) {
	tests := []struct {
		name        string
		reverse     bool
		value       float64
		expectWidth float64
		expectPos   float64
	}{
		{"forward-25", false, 25, 50, 0},
		{"forward-75", false, 75, 150, 0},
		{"reverse-25", true, 25, 50, 150},
		{"reverse-75", true, 75, 150, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := NewProgressBar()
			bar.SetSize(200, 20)
			bar.SetTemplateComponent(makeProgressTemplate(false))
			bar.SetMin(0)
			bar.SetMax(100)
			bar.SetReverse(tt.reverse)
			bar.SetValue(tt.value)

			if bar.barObjectH == nil {
				t.Fatalf("horizontal bar missing")
			}
			width := bar.barObjectH.Width()
			if !floatEquals(width, tt.expectWidth) {
				t.Fatalf("expected width %.1f got %.1f", tt.expectWidth, width)
			}
			pos := bar.barObjectH.X()
			if !floatEquals(pos, tt.expectPos) {
				t.Fatalf("expected pos %.1f got %.1f", tt.expectPos, pos)
			}
		})
	}
}

func TestProgressBarTitleFormats(t *testing.T) {
	bar := NewProgressBar()
	bar.SetSize(200, 20)
	bar.SetTemplateComponent(makeProgressTemplate(false))
	bar.SetMin(0)
	bar.SetMax(100)
	bar.SetValue(25)

	check := func(tp ProgressTitleType, want string) {
		bar.SetTitleType(tp)
		text := bar.titleObject.Data().(string)
		if text != want {
			t.Fatalf("want %q got %q", want, text)
		}
	}

	check(ProgressTitleTypePercent, "25%")
	check(ProgressTitleTypeValue, "25")
	check(ProgressTitleTypeMax, "100")
	check(ProgressTitleTypeValueAndMax, "25/100")
}

func makeProgressTemplate(withVertical bool) *core.GComponent {
	template := core.NewGComponent()
	template.SetSize(200, 200)

	title := core.NewGObject()
	title.SetName("title")
	template.AddChild(title)

	bar := core.NewGObject()
	bar.SetName("bar")
	bar.SetSize(200, 10)
	template.AddChild(bar)

	if withVertical {
		barV := core.NewGObject()
		barV.SetName("bar_v")
		barV.SetSize(10, 200)
		template.AddChild(barV)
	}

	return template
}

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) <= 0.0001
}
