package widgets

import (
	"math"
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestSliderClampAndTitle(t *testing.T) {
	slider := NewSlider()
	slider.SetSize(200, 20)
	slider.SetTemplateComponent(makeSliderTemplate(true))
	slider.SetMin(10)
	slider.SetMax(30)
	slider.SetWholeNumbers(true)
	slider.SetValue(9.3)
	if slider.Value() != 10 {
		t.Fatalf("expected clamp to min, got %.2f", slider.Value())
	}
	slider.SetValue(25.7)
	if slider.Value() != 26 {
		t.Fatalf("expected rounding to 26, got %.2f", slider.Value())
	}
	slider.SetTitleType(ProgressTitleTypeValueAndMax)
	if got := slider.titleObject.Data().(string); got != "26/30" {
		t.Fatalf("unexpected title %q", got)
	}
}

func TestSliderBarUpdate(t *testing.T) {
	tests := []struct {
		name        string
		reverse     bool
		value       float64
		expectWidth float64
		expectPos   float64
	}{
		{"forward-50", false, 50, 100, 0},
		{"reverse-50", true, 50, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slider := NewSlider()
			slider.SetSize(200, 20)
			slider.SetTemplateComponent(makeSliderTemplate(true))
			slider.SetMin(0)
			slider.SetMax(100)
			slider.SetReverse(tt.reverse)
			slider.SetValue(tt.value)

			if slider.barObjectH == nil {
				t.Fatalf("horizontal bar missing")
			}
			width := slider.barObjectH.Width()
			if !floatEquals(width, tt.expectWidth) {
				t.Fatalf("expected bar width %.1f got %.1f", tt.expectWidth, width)
			}
			pos := slider.barObjectH.X()
			if !floatEquals(pos, tt.expectPos) {
				t.Fatalf("expected bar x %.1f got %.1f", tt.expectPos, pos)
			}
		})
	}
}

func TestSliderChangeOnClick(t *testing.T) {
	slider := NewSlider()
	slider.SetSize(200, 20)
	slider.SetTemplateComponent(makeSliderTemplate(true))
	slider.SetMin(0)
	slider.SetMax(100)
	slider.SetChangeOnClick(true)

	prev := core.Root().Stage()
	stage := laya.NewStage(400, 300)
	core.Root().AttachStage(stage)
	defer core.Root().AttachStage(prev)

	// simulate click at 150px
	event := laya.PointerEvent{Position: laya.Point{X: 150, Y: 0}}
	slider.onBarMouseDown(laya.Event{Data: event})
	if math.Abs(slider.Value()-75) > 0.01 {
		t.Fatalf("expected value around 75, got %.2f", slider.Value())
	}
}

func makeSliderTemplate(horizontal bool) *core.GComponent {
	tmpl := core.NewGComponent()
	tmpl.SetSize(200, 20)

	title := core.NewGObject()
	title.SetName("title")
	tmpl.AddChild(title)

	if horizontal {
		bar := core.NewGObject()
		bar.SetName("bar")
		bar.SetSize(200, 10)
		tmpl.AddChild(bar)
	}

	grip := core.NewGObject()
	grip.SetName("grip")
	grip.SetSize(10, 20)
	tmpl.AddChild(grip)
	return tmpl
}
