package core

import (
	"math"
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

func TestScrollPaneSetPos(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(120, 80)

	pane := comp.EnsureScrollPane(ScrollTypeBoth)
	if pane == nil {
		t.Fatalf("expected scroll pane to be created")
	}
	pane.SetContentSize(320, 240)
	pane.SetPos(60, 40, false)

	if diff := math.Abs(pane.PosX() - 60); diff > 1e-6 {
		t.Fatalf("unexpected x pos: %.2f", pane.PosX())
	}
	if diff := math.Abs(pane.PosY() - 40); diff > 1e-6 {
		t.Fatalf("unexpected y pos: %.2f", pane.PosY())
	}

	if pane.container == nil {
		t.Fatalf("expected container sprite")
	}
	pos := pane.container.Position()
	if diff := math.Abs(pos.X + 60); diff > 1e-6 {
		t.Fatalf("expected container x offset -60, got %.2f", pos.X)
	}
	if diff := math.Abs(pos.Y + 40); diff > 1e-6 {
		t.Fatalf("expected container y offset -40, got %.2f", pos.Y)
	}

	rect := pane.maskContainer.ScrollRect()
	if rect == nil || rect.W != 120 || rect.H != 80 {
		t.Fatalf("unexpected scroll rect: %+v", rect)
	}
}

func TestScrollPaneMouseWheel(t *testing.T) {
	stage := laya.NewStage(400, 300)
	root := NewGComponent()
	root.SetSize(400, 300)
	stage.AddChild(root.DisplayObject())

	comp := NewGComponent()
	comp.SetPosition(20, 20)
	comp.SetSize(100, 100)
	pane := comp.EnsureScrollPane(ScrollTypeVertical)
	pane.SetContentSize(100, 260)
	root.AddChild(comp.GObject)

	state := laya.MouseState{X: 40, Y: 40, WheelY: -1}
	stage.Update(16*time.Millisecond, state)

	if pane.PosY() <= 0 {
		t.Fatalf("expected vertical position to increase after wheel, got %.2f", pane.PosY())
	}
	if pane.PosX() != 0 {
		t.Fatalf("expected horizontal position unchanged, got %.2f", pane.PosX())
	}
}

func TestGComponentSetupScroll(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(120, 90)
	data := []byte{
		byte(ScrollTypeVertical),
		byte(ScrollBarDisplayHidden),
		0, 0, 0, 0,
		0,
		0xFF, 0xFE,
		0xFF, 0xFE,
		0xFF, 0xFE,
		0xFF, 0xFE,
	}
	buf := utils.NewByteBuffer(data)
	comp.SetupScroll(buf)
	pane := comp.ScrollPane()
	if pane == nil {
		t.Fatalf("expected scroll pane after setup")
	}
	if pane.scrollType != ScrollTypeVertical {
		t.Fatalf("expected vertical scroll type, got %d", pane.scrollType)
	}
	if pane.MouseWheelEnabled() {
		t.Fatalf("expected mouse wheel disabled when scrollbar hidden")
	}
	if pane.ViewWidth() != comp.Width() || pane.ViewHeight() != comp.Height() {
		t.Fatalf("view size mismatch, got %.1fx%.1f want %.1fx%.1f", pane.ViewWidth(), pane.ViewHeight(), comp.Width(), comp.Height())
	}
}

func TestScrollPaneDrag(t *testing.T) {
	env := testutil.NewStageEnv(t, 400, 300)
	stage := env.Stage
	prevStage := Root().Stage()
	Root().AttachStage(stage)
	defer Root().AttachStage(prevStage)

	root := NewGComponent()
	root.SetSize(400, 300)
	stage.AddChild(root.DisplayObject())

	comp := NewGComponent()
	comp.SetPosition(10, 10)
	comp.SetSize(120, 120)
	pane := comp.EnsureScrollPane(ScrollTypeBoth)
	pane.SetContentSize(320, 260)
	root.AddChild(comp.GObject)

	env.Advance(16*time.Millisecond, laya.MouseState{X: 60, Y: 60, Primary: false})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 60, Y: 60, Primary: true})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 20, Y: 20, Primary: true})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 20, Y: 20, Primary: false})

	if pane.PosX() <= 0 {
		t.Fatalf("expected horizontal scroll after drag, got %.2f", pane.PosX())
	}
	if pane.PosY() <= 0 {
		t.Fatalf("expected vertical scroll after drag, got %.2f", pane.PosY())
	}
}

func TestScrollPaneSnapToPage(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(100, 100)
	pane := comp.EnsureScrollPane(ScrollTypeHorizontal)
	pane.SetContentSize(320, 100)
	pane.pageMode = true
	pane.pageSize = laya.Point{X: 100, Y: 100}
	pane.SetPos(130, 0, false)
	pane.snapToNearestPage()

	if diff := math.Abs(pane.PosX() - 100); diff > 1e-6 {
		t.Fatalf("expected snap to 100, got %.2f", pane.PosX())
	}
}
