//go:build ebiten

package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

const (
	windowWidth  = 640
	windowHeight = 360
)

func main() {
	const fontSize = 24
	face, path, err := render.LoadSystemFont(fontSize)
	if err != nil {
		log.Fatalf("无法加载系统字体（可通过设置环境变量 FGUI_FONT_PATH 指定路径）: %v", err)
	}
	log.Printf("使用字体: %s", path)
	render.SetTextFont(face)

	root := core.NewGComponent()
	root.GObject.SetSize(windowWidth, windowHeight)

	text := widgets.NewText()
	text.SetText("你好，FairyGUI！这是一个文本渲染测试。")
	text.GObject.SetSize(windowWidth, fontSize*2)
	text.GObject.SetPosition(40, 120)
	root.AddChild(text.GObject)

	game := &Game{
		root:  root,
		atlas: render.NewAtlasManager(nil),
	}

	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("FGUI Text Demo")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	root  *core.GComponent
	atlas *render.AtlasManager
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x20, 0x20, 0x20, 0xff})
	if g.root != nil && g.atlas != nil {
		if err := render.DrawComponent(screen, g.root, g.atlas); err != nil {
			log.Printf("绘制失败: %v", err)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowWidth, windowHeight
}
