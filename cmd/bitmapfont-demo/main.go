//go:build ebiten

package main

import (
	"context"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chslink/fairygui/pkg/fgui"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

func main() {
	ctx := context.Background()
	loader := assets.NewFileLoader("demo/assets")
	atlas := render.NewAtlasManager(loader)

	data, err := loader.LoadOne(ctx, "Transition.fui", assets.ResourceBinary)
	if err != nil {
		log.Fatalf("load fui: %v", err)
	}
	pkg, err := assets.ParsePackage(data, "Transition")
	if err != nil {
		log.Fatalf("parse package: %v", err)
	}
	render.RegisterBitmapFonts(pkg)
	if err := atlas.LoadPackage(ctx, pkg); err != nil {
		log.Fatalf("load atlas: %v", err)
	}

	stage := fgui.NewStage(600, 200)
	root := core.NewGRoot()
	root.AttachStage(stage)
	root.Resize(600, 200)

	textField := widgets.NewText()
	textField.SetFont("ui://c0hnre6olvxr1e")
	textField.SetAlign(widgets.TextAlignCenter)
	textField.SetFontSize(48)
	textField.SetText("1234567890")
	textField.GObject.SetPosition(120, 80)
	root.AddChild(textField.GObject)

	ebiten.SetWindowSize(600, 200)
	ebiten.SetWindowTitle("BitmapFont Demo")

	if err := ebiten.RunGame(&game{root: root, atlas: atlas}); err != nil {
		log.Fatal(err)
	}
}

type game struct {
	root  *core.GRoot
	atlas *render.AtlasManager
}

func (g *game) Update() error {
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 30, 255})
	if err := render.DrawComponent(screen, g.root.GComponent, g.atlas); err != nil {
		log.Printf("render error: %v", err)
	}
}

func (g *game) Layout(int, int) (int, int) {
	return 600, 200
}
