//go:build ebiten

package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

const (
	defaultWidth  = 320
	defaultHeight = 180
)

func main() {
	ctx := context.Background()
	loader := assets.NewFileLoader("demo/assets")

	bagData, err := loader.LoadOne(ctx, "MainMenu.fui", assets.ResourceBinary)
	if err != nil {
		log.Fatalf("failed to load MainMenu.fui: %v", err)
	}
	bagPkg, err := assets.ParsePackage(bagData, "MainMenu")
	if err != nil {
		log.Fatalf("failed to parse MainMenu.fui: %v", err)
	}

	atlasMgr := render.NewAtlasManager(loader)
	if err := atlasMgr.LoadPackage(ctx, bagPkg); err != nil {
		log.Fatalf("failed to load Bag atlas: %v", err)
	}

	spriteID := "rftu5"
	btnImage := bagPkg.ItemByID(spriteID)
	if btnImage == nil {
		log.Fatalf("package Bag missing sprite %s", spriteID)
	}
	fmt.Printf("btnImage=%v\n", btnImage.Sprite.Offset)
	imgWidget := widgets.NewImage()
	imgWidget.SetPackageItem(btnImage)
	imgWidget.GObject.SetPosition(160, 120)
	imgWidget.GObject.SetPivotWithAnchor(0.5, 0.5, true)
	imgWidget.GObject.SetSize(defaultWidth, defaultHeight)

	root := core.NewGComponent()
	root.AddChild(imgWidget.GObject)

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Nine-Slice Debug")

	game := &Game{
		atlas:   atlasMgr,
		root:    root,
		image:   imgWidget,
		pkgItem: btnImage,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game drives a minimal Ebiten loop that stretches the nine-slice image.
type Game struct {
	atlas   *render.AtlasManager
	root    *core.GComponent
	image   *widgets.GImage
	pkgItem *assets.PackageItem
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {
	if g.image == nil {
		return nil
	}
	w := g.image.GObject.Width()
	h := g.image.GObject.Height()

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		w -= 4
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		w += 4
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		h -= 4
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		h += 4
	}
	w = math.Max(20, w)
	h = math.Max(20, h)
	g.image.GObject.SetSize(w, h)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x22, 0x22, 0x22, 0xff})
	if g.root != nil {
		if err := render.DrawComponent(screen, g.root, g.atlas); err != nil {
			ebitenutil.DebugPrint(screen, fmt.Sprintf("draw error: %v", err))
		}
	}
	if g.image != nil && g.pkgItem != nil {
		info := fmt.Sprintf("sprite: %s\nsize: %.0fx%.0f\nscale9: %v\narrow keys resize",
			g.pkgItem.Name, g.image.GObject.Width(), g.image.GObject.Height(), g.pkgItem.Scale9Grid != nil)
		ebitenutil.DebugPrintAt(screen, info, 16, 16)
	}
}
