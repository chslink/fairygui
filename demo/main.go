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
	"golang.org/x/image/font/basicfont"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/builder"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
)

func main() {
	ctx := context.Background()

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

	pkg, err := assets.ParsePackage(data, filepath.Join(assetsDir, "MainMenu"))
	if err != nil {
		return nil, err
	}

	componentItem := findFirstComponent(pkg)
	if componentItem == nil {
		return nil, fmt.Errorf("no component items found in package")
	}

	atlas := render.NewAtlasManager(loader)
	factory := builder.NewFactory(atlas)

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
	s.drawComponent(screen, s.root, 0, 0)
}

func (s *Scene) drawComponent(screen *ebiten.Image, comp *core.GComponent, offsetX, offsetY float64) {
	if comp == nil {
		return
	}
	for _, child := range comp.Children() {
		s.drawObject(screen, child, offsetX, offsetY)
	}
}

func (s *Scene) drawObject(screen *ebiten.Image, obj *core.GObject, offsetX, offsetY float64) {
	if obj == nil {
		return
	}

	x := offsetX + obj.X()
	y := offsetY + obj.Y()

	switch data := obj.Data().(type) {
	case *assets.PackageItem:
		if spriteAny, err := s.atlas.ResolveSprite(data); err == nil {
			if img, ok := spriteAny.(*ebiten.Image); ok {
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(x, y)
				screen.DrawImage(img, opts)
			}
		}
	case *core.GComponent:
		s.drawComponent(screen, data, x, y)
	case string:
		if data != "" {
			ascent := basicfont.Face7x13.Metrics().Ascent.Ceil()
			text.Draw(screen, data, basicfont.Face7x13, int(x), int(y)+ascent, color.White)
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
