package widgets

import (
	"math"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestScrollBarSyncsWithScrollPane(t *testing.T) {
	owner := core.NewGComponent()
	owner.SetSize(100, 100)
	pane := owner.EnsureScrollPane(core.ScrollTypeVertical)
	pane.SetContentSize(100, 300)

	sb := NewScrollBar()
	sb.SetSize(20, 100)
	sb.SetTemplateComponent(makeScrollBarTemplate(true))
	sb.SetScrollPane(pane, true)

	pane.SetPos(0, 100, false)
	if diff := math.Abs(sb.scrollPerc - 0.5); diff > 0.01 {
		t.Fatalf("expected scrollPerc around 0.5, got %.2f", sb.scrollPerc)
	}
	length := sb.bar.Height() - sb.extraMargin()
	gripLen := sb.grip.Height()
	expected := length * (pane.ViewHeight() / 300)
	if diff := math.Abs(gripLen - expected); diff > 1 {
		t.Fatalf("expected grip length %.2f got %.2f", expected, gripLen)
	}
}

func makeScrollBarTemplate(vertical bool) *core.GComponent {
	tmpl := core.NewGComponent()
	tmpl.SetSize(20, 100)

	bar := core.NewGObject()
	bar.SetName("bar")
	if vertical {
		bar.SetSize(10, 90)
	} else {
		bar.SetSize(90, 10)
	}
	tmpl.AddChild(bar)

	grip := core.NewGObject()
	grip.SetName("grip")
	if vertical {
		grip.SetSize(10, 30)
	} else {
		grip.SetSize(30, 10)
	}
	tmpl.AddChild(grip)

	return tmpl
}
