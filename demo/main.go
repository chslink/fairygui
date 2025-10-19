//go:build ebiten

package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/builder"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
)

var (
	textFont font.Face = basicfont.Face7x13
)

func main() {
	ctx := context.Background()

	if face, err := loadPreferredFont(18); err == nil {
		textFont = face
		render.SetTextFont(face)
	} else {
		log.Printf("warning: fallback basic font used, Chinese glyphs may not render: %v", err)
		render.SetTextFont(textFont)
	}

	scene, err := NewScene(ctx)
	if err != nil {
		log.Fatalf("failed to build scene: %v", err)
	}

	ebiten.SetWindowSize(scene.Width(), scene.Height())
	ebiten.SetWindowTitle("FairyGUI Ebiten Demo")

	if err := ebiten.RunGame(&Game{scene: scene}); err != nil {
		log.Fatal(err)
	}
}

// Game wraps the scene for Ebiten's game loop.
type Game struct {
	scene *Scene
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x1e, 0x1e, 0x1e, 0xff})
	g.scene.Draw(screen)
	drawHUD(screen, g.scene)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.scene.Width(), g.scene.Height()
}

// Scene bundles the loaded package, atlas manager, and component hierarchy.
type Scene struct {
	atlas  *render.AtlasManager
	root   *core.GComponent
	width  int
	height int
}

func NewScene(ctx context.Context) (*Scene, error) {
	assetsDir := filepath.Join("demo", "assets")
	loader := assets.NewFileLoader(assetsDir)

	data, err := loader.LoadOne(ctx, "MainMenu.fui", assets.ResourceBinary)
	if err != nil {
		return nil, err
	}
	pkg, err := assets.ParsePackage(data, filepath.Join("MainMenu"))
	if err != nil {
		return nil, err
	}

	componentItem := findFirstComponent(pkg)
	if componentItem == nil {
		return nil, fmt.Errorf("no component items found in package")
	}

	atlas := render.NewAtlasManager(loader)
	factory := builder.NewFactoryWithLoader(atlas, loader)

	root, err := factory.BuildComponent(ctx, pkg, componentItem)
	if err != nil {
		return nil, err
	}

	width := int(root.Width())
	if width <= 0 {
		width = 800
	}
	height := int(root.Height())
	if height <= 0 {
		height = 600
	}

	return &Scene{atlas: atlas, root: root, width: width, height: height}, nil
}

func (s *Scene) Width() int  { return s.width }
func (s *Scene) Height() int { return s.height }

func (s *Scene) Draw(screen *ebiten.Image) {
	if err := render.DrawComponent(screen, s.root, s.atlas); err != nil {
		log.Printf("render component failed: %v", err)
	}
	s.drawText(screen, s.root, 0, 0)
}

func (s *Scene) drawText(target *ebiten.Image, comp *core.GComponent, offsetX, offsetY float64) {
	if comp == nil {
		return
	}
	for _, child := range comp.Children() {
		if child == nil || !child.Visible() {
			continue
		}
		x := offsetX + child.X()
		y := offsetY + child.Y()
		switch data := child.Data().(type) {
		case string:
			if data != "" {
				ascent := textFont.Metrics().Ascent.Ceil()
				text.Draw(target, data, textFont, int(x), int(y)+ascent, color.White)
			}
		case *core.GComponent:
			s.drawText(target, data, x, y)
		}
	}
}

func findFirstComponent(pkg *assets.Package) *assets.PackageItem {
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Component != nil {
			return item
		}
	}
	return nil
}

func drawHUD(screen *ebiten.Image, scene *Scene) {
	lines := []string{
		fmt.Sprintf("Root size: %.0fx%.0f", scene.root.Width(), scene.root.Height()),
		fmt.Sprintf("Total controllers: %d", len(scene.root.Controllers())),
	}
	y := 16
	for _, line := range lines {
		text.Draw(screen, line, basicfont.Face7x13, 16, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
		y += 16
	}
}

func loadPreferredFont(px float64) (font.Face, error) {
	return nil, fmt.Errorf("preferred font loading not implemented")
}
