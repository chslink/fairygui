//go:build ebiten

package capture

import (
	"errors"
	"fmt"
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chslink/fairygui/pkg/fgui"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
)

// DebugFunc receives formatted debug logs.
type DebugFunc func(string, ...any)

// CaptureComponent renders the provided component into an RGBA image of the requested size.
func CaptureComponent(width, height int, component *core.GComponent, atlas *render.AtlasManager, debug DebugFunc) (*image.RGBA, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("capture: invalid size %dx%d", width, height)
	}
	game := newCaptureGame(width, height, component, atlas, debug)

	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowResizable(false)
	ebiten.SetWindowTitle("fgui-capture")

	if err := ebiten.RunGame(game); err != nil && !errors.Is(err, ebiten.Termination) {
		return nil, err
	}
	if game.err != nil {
		return nil, game.err
	}
	if !game.rendered {
		return nil, errors.New("capture: render incomplete")
	}
	return game.image(), nil
}

type captureGame struct {
	width     int
	height    int
	component *core.GComponent
	atlas     *render.AtlasManager
	root      *core.GRoot
	stage     *fgui.Stage
	debug     DebugFunc

	rendered bool
	pixels   []byte
	err      error
}

func newCaptureGame(width, height int, component *core.GComponent, atlas *render.AtlasManager, debug DebugFunc) *captureGame {
	stage := fgui.NewStage(width, height)
	root := core.NewGRoot()
	root.AttachStage(stage)
	root.Resize(width, height)
	root.SetData(root)
	if component != nil {
		component.SetData(component)
		component.SetPosition(0, 0)
		root.AddChild(component.GObject)
	}
	return &captureGame{
		width:     width,
		height:    height,
		component: component,
		atlas:     atlas,
		root:      root,
		stage:     stage,
		debug:     debug,
	}
}

func (g *captureGame) Update() error {
	if g.err != nil {
		return g.err
	}
	if g.rendered {
		return ebiten.Termination
	}
	g.root.Advance(time.Second/60, fgui.MouseState{})
	return nil
}

func (g *captureGame) Draw(screen *ebiten.Image) {
	if g.err != nil || g.rendered {
		return
	}
	screen.Clear()
	if err := render.DrawComponent(screen, g.root.GComponent, g.atlas); err != nil {
		g.err = err
		return
	}
	w, h := screen.Size()
	if g.debug != nil {
		g.debug("ReadPixels request size: %dx%d", w, h)
	}
	pixels := make([]byte, 4*w*h)
	screen.ReadPixels(pixels)
	g.pixels = pixels
	g.rendered = true
}

func (g *captureGame) Layout(_, _ int) (int, int) {
	return g.width, g.height
}

func (g *captureGame) image() *image.RGBA {
	return &image.RGBA{
		Pix:    g.pixels,
		Stride: 4 * g.width,
		Rect:   image.Rect(0, 0, g.width, g.height),
	}
}
