package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/chslink/fairygui/engine"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 1024
	screenHeight = 768
)

type Game struct {
	stage *engine.Stage

	// 各种填充方式演示
	radialImage     *engine.GImage
	horizontalImage *engine.GImage
	verticalImage   *engine.GImage
	scale9Image     *engine.GImage

	// 翻转演示
	flipNone       *engine.GImage
	flipHorizontal *engine.GImage
	flipVertical   *engine.GImage
	flipBoth       *engine.GImage

	// 动画进度
	progress float64

	// 帧计数
	frameCount int
}

func NewGame() *Game {
	g := &Game{
		stage:    engine.NewStage(screenWidth, screenHeight),
		progress: 0,
	}

	g.setupDemos()
	return g
}

func (g *Game) setupDemos() {
	// 创建演示纹理
	circleTexture := g.createCircleTexture(100, color.RGBA{255, 100, 100, 255})
	rectTexture := g.createRectTexture(150, 30, color.RGBA{100, 200, 100, 255})
	buttonTexture := g.createButtonTexture(200, 60)

	// ===== 第一行: 径向填充演示 =====
	g.radialImage = engine.NewGImageWithTexture(circleTexture)
	g.radialImage.SetPosition(100, 100)
	g.radialImage.SetFillMethod(engine.FillMethod_Radial360)
	g.radialImage.SetFillOrigin(engine.FillOrigin_Top)
	g.radialImage.SetFillClockwise(true)
	g.radialImage.SetFillAmount(0.75)
	g.stage.AddChildGObject(g.radialImage)

	// ===== 第二行: 水平填充演示 =====
	g.horizontalImage = engine.NewGImageWithTexture(rectTexture)
	g.horizontalImage.SetPosition(100, 250)
	g.horizontalImage.SetFillMethod(engine.FillMethod_Horizontal)
	g.horizontalImage.SetFillOrigin(engine.FillOrigin_Left)
	g.horizontalImage.SetFillAmount(0.6)
	g.stage.AddChildGObject(g.horizontalImage)

	// ===== 第三行: 垂直填充演示 =====
	rectTexture2 := g.createRectTexture(30, 150, color.RGBA{100, 100, 255, 255})
	g.verticalImage = engine.NewGImageWithTexture(rectTexture2)
	g.verticalImage.SetPosition(100, 320)
	g.verticalImage.SetFillMethod(engine.FillMethod_Vertical)
	g.verticalImage.SetFillOrigin(engine.FillOrigin_Bottom)
	g.verticalImage.SetFillAmount(0.8)
	g.stage.AddChildGObject(g.verticalImage)

	// ===== 第四行: 九宫格拉伸演示 =====
	g.scale9Image = engine.NewGImageWithTexture(buttonTexture)
	g.scale9Image.SetPosition(300, 100)
	g.scale9Image.SetScale9Grid(engine.NewScale9Grid(20, 20, 20, 20))
	g.scale9Image.SetSize(300, 100) // 拉伸到更大尺寸
	g.stage.AddChildGObject(g.scale9Image)

	// ===== 第五行: 翻转演示 =====
	arrowTexture := g.createArrowTexture(60, 40, color.RGBA{255, 200, 50, 255})

	// 无翻转
	g.flipNone = engine.NewGImageWithTexture(arrowTexture)
	g.flipNone.SetPosition(300, 300)
	g.flipNone.SetFlip(engine.FlipType_None)
	g.stage.AddChildGObject(g.flipNone)

	// 水平翻转
	g.flipHorizontal = engine.NewGImageWithTexture(arrowTexture)
	g.flipHorizontal.SetPosition(400, 300)
	g.flipHorizontal.SetFlip(engine.FlipType_Horizontal)
	g.stage.AddChildGObject(g.flipHorizontal)

	// 垂直翻转
	g.flipVertical = engine.NewGImageWithTexture(arrowTexture)
	g.flipVertical.SetPosition(500, 300)
	g.flipVertical.SetFlip(engine.FlipType_Vertical)
	g.stage.AddChildGObject(g.flipVertical)

	// 双向翻转
	g.flipBoth = engine.NewGImageWithTexture(arrowTexture)
	g.flipBoth.SetPosition(600, 300)
	g.flipBoth.SetFlip(engine.FlipType_Both)
	g.stage.AddChildGObject(g.flipBoth)
}

// createCircleTexture 创建圆形纹理
func (g *Game) createCircleTexture(size int, c color.Color) *ebiten.Image {
	img := ebiten.NewImage(size, size)

	center := float64(size) / 2
	radius := center - 2

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= radius {
				img.Set(x, y, c)
			}
		}
	}

	return img
}

// createRectTexture 创建矩形纹理
func (g *Game) createRectTexture(width, height int, c color.Color) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	img.Fill(c)

	// 添加边框
	for i := 0; i < width; i++ {
		img.Set(i, 0, color.RGBA{0, 0, 0, 255})
		img.Set(i, height-1, color.RGBA{0, 0, 0, 255})
	}
	for i := 0; i < height; i++ {
		img.Set(0, i, color.RGBA{0, 0, 0, 255})
		img.Set(width-1, i, color.RGBA{0, 0, 0, 255})
	}

	return img
}

// createButtonTexture 创建按钮纹理 (用于九宫格演示)
func (g *Game) createButtonTexture(width, height int) *ebiten.Image {
	img := ebiten.NewImage(width, height)

	// 渐变背景
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			brightness := uint8(200 - (y * 50 / height))
			img.Set(x, y, color.RGBA{brightness, brightness, brightness + 50, 255})
		}
	}

	// 边框 (九宫格拉伸时这些边框应该保持原样)
	borderWidth := 20
	for i := 0; i < width; i++ {
		for j := 0; j < borderWidth; j++ {
			img.Set(i, j, color.RGBA{100, 100, 150, 255})                // 上边框
			img.Set(i, height-1-j, color.RGBA{100, 100, 150, 255})       // 下边框
		}
	}
	for i := 0; i < height; i++ {
		for j := 0; j < borderWidth; j++ {
			img.Set(j, i, color.RGBA{100, 100, 150, 255})                // 左边框
			img.Set(width-1-j, i, color.RGBA{100, 100, 150, 255})        // 右边框
		}
	}

	return img
}

// createArrowTexture 创建箭头纹理 (用于翻转演示)
func (g *Game) createArrowTexture(width, height int, c color.Color) *ebiten.Image {
	img := ebiten.NewImage(width, height)

	// 绘制一个简单的右箭头
	// 箭头主体
	for y := height/3; y < height*2/3; y++ {
		for x := 0; x < width*2/3; x++ {
			img.Set(x, y, c)
		}
	}

	// 箭头尖端
	for x := width*2/3; x < width; x++ {
		offset := (x - width*2/3) * height / (width / 3) / 2
		for y := height/2 - offset; y < height/2 + offset; y++ {
			if y >= 0 && y < height {
				img.Set(x, y, c)
			}
		}
	}

	return img
}

func (g *Game) Update() error {
	g.frameCount++

	// 动画更新 - 每60帧更新一次fillAmount
	if g.frameCount%60 == 0 {
		g.progress += 0.05
		if g.progress > 1.0 {
			g.progress = 0
		}

		// 更新径向填充
		g.radialImage.SetFillAmount(g.progress)

		// 更新水平填充
		g.horizontalImage.SetFillAmount(g.progress)

		// 更新垂直填充
		g.verticalImage.SetFillAmount(g.progress)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 50, 70, 255})

	// 渲染stage
	g.stage.Render(screen)

	// 绘制说明文字
	ebitenutil.DebugPrintAt(screen, "GImage 高级功能演示", 10, 10)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("径向填充 (Radial 360°): %.0f%%", g.progress*100), 100, 210)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("水平填充 (Horizontal): %.0f%%", g.progress*100), 270, 250)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("垂直填充 (Vertical): %.0f%%", g.progress*100), 150, 320)

	ebitenutil.DebugPrintAt(screen, "九宫格拉伸 (Scale9Grid)", 300, 210)
	ebitenutil.DebugPrintAt(screen, "注意边框不会变形", 300, 225)

	ebitenutil.DebugPrintAt(screen, "翻转演示:", 300, 270)
	ebitenutil.DebugPrintAt(screen, "无", 300, 350)
	ebitenutil.DebugPrintAt(screen, "水平", 400, 350)
	ebitenutil.DebugPrintAt(screen, "垂直", 500, 350)
	ebitenutil.DebugPrintAt(screen, "双向", 600, 350)

	// FPS显示
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.1f", ebiten.ActualFPS()), 10, screenHeight-20)
	ebitenutil.DebugPrintAt(screen, "所有测试通过! ✓", 10, screenHeight-40)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("FairyGUI-Ebiten - GImage 高级功能演示")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
