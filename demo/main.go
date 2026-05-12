package main

import (
	"context"
	"image/color"
	"log"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/chslink/fairygui/demo/debug"
	"github.com/chslink/fairygui/demo/scenes"
	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

var fallbackFont font.Face = basicfont.Face7x13

func main() {
	ctx := context.Background()

	render.SetTextFont(fallbackFont)
	if face, err := loadPreferredFont(18); err == nil {
		render.SetTextFont(face)
	} else {
		log.Printf("warning: fallback basic font used, Chinese glyphs may not render: %v", err)
	}

	game, err := newGame(ctx)
	if err != nil {
		log.Fatalf("failed to initialise demo: %v", err)
	}

	ebiten.SetWindowSize(game.width, game.height)
	ebiten.SetWindowTitle("FairyGUI Ebiten Demo")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

type game struct {
	root       *core.GRoot
	manager    *scenes.Manager
	atlas      *render.AtlasManager
	width      int
	height     int
	lastUpdate time.Time
	keysDown   map[ebiten.Key]bool
	debug      *debug.Server
}

func newGame(ctx context.Context) (*game, error) {
	// 初始化音频系统（必须首先调用，创建音频上下文）
	fgui.InitAudio(48000) // 使用48000采样率

	// 注册按钮音效播放器
	fgui.RegisterButtonSoundPlayer()

	// 设置音频系统的资源加载器（用于自动从包中加载音效）
	assetsDir := filepath.Join("demo", "assets")
	loader := fgui.NewFileLoader(assetsDir)
	fgui.SetAudioLoader(loader)

	atlas := render.NewAtlasManager(loader)
	factory := fgui.NewFactoryWithLoader(atlas, loader)
	env := scenes.NewEnvironment(loader, factory, atlas)
	manager, err := scenes.NewManager(ctx, env)
	if err != nil {
		return nil, err
	}

	root := core.Root()
	stage := fgui.NewStage(manager.Width(), manager.Height())
	root.AttachStage(stage)
	root.Resize(manager.Width(), manager.Height())
	root.SetData(root)

	for _, child := range root.Children() {
		root.RemoveChild(child)
	}
	stageRoot := manager.Stage()
	stageRoot.GObject.SetPosition(0, 0)
	stageRoot.GObject.SetData(stageRoot)
	root.AddChild(stageRoot.GObject)

	// 启动调试服务器 - 监控manager的stage而不是root
	debugServer := debug.NewServer(stageRoot, 8090)
	if err := debugServer.Start(); err != nil {
		log.Printf("warning: failed to start debug server: %v", err)
	} else {
		log.Printf("debug server started at %s", debugServer.GetURL())
	}

	return &game{
		root:     root,
		manager:  manager,
		atlas:    atlas,
		width:    manager.Width(),
		height:   manager.Height(),
		keysDown: make(map[ebiten.Key]bool),
		debug:    debugServer,
	}, nil
}

func (g *game) Update() error {
	if g.root == nil {
		return nil
	}

	now := time.Now()
	delta := frameDelta(g.lastUpdate, now)
	g.lastUpdate = now

	g.syncStageSize()

	// 收集完整的输入状态(鼠标 + 键盘)
	input := laya.InputState{
		Mouse: g.mouseState(),
		Keys:  g.keyboardEvents(),
	}
	g.root.AdvanceInput(delta, input)

	// Deliver typed characters (including IME) to the focused text input.
	// ebiten.AppendInputChars handles IME composition natively on Windows.
	for _, r := range ebiten.AppendInputChars(nil) {
		widgets.InputChar(string(r))
	}

	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x1e, 0x1e, 0x1e, 0xff})
	if g.root == nil {
		return
	}
	if err := render.DrawComponent(screen, g.root.GComponent, g.atlas); err != nil {
		log.Printf("render component failed: %v", err)
	}
}

func (g *game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.syncStageSize()
	if g.width == 0 || g.height == 0 {
		return 800, 600
	}
	return g.width, g.height
}

func (g *game) syncStageSize() {
	if g.root == nil || g.manager == nil {
		return
	}
	width := g.manager.Width()
	height := g.manager.Height()
	if width <= 0 || height <= 0 {
		return
	}
	if width != g.width || height != g.height {
		g.root.Resize(width, height)
		g.width = width
		g.height = height
	}
}

func (g *game) mouseState() fgui.MouseState {
	x, y := ebiten.CursorPosition()
	wheelX, wheelY := ebiten.Wheel()
	buttons := fgui.MouseButtons{
		Left:   ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		Right:  ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight),
		Middle: ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle),
	}
	modifiers := fgui.KeyModifiers{
		Shift: ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight),
		Ctrl:  ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight),
		Alt:   ebiten.IsKeyPressed(ebiten.KeyAltLeft) || ebiten.IsKeyPressed(ebiten.KeyAltRight),
		Meta:  ebiten.IsKeyPressed(ebiten.KeyMetaLeft) || ebiten.IsKeyPressed(ebiten.KeyMetaRight),
	}
	state := fgui.MouseState{
		X:       float64(x),
		Y:       float64(y),
		Primary: buttons.Left,
		Buttons: buttons,
		Modifiers: modifiers,
		WheelX:  wheelX,
		WheelY:  wheelY,
	}
	if g.root == nil {
		return state
	}
	if stage := g.root.Stage(); stage != nil {
		if sprite := stage.Root(); sprite != nil {
			local := sprite.GlobalToLocal(laya.Point{X: state.X, Y: state.Y})
			state.X = local.X
			state.Y = local.Y
		}
	}
	return state
}

// keyboardEvents 收集当前帧的键盘事件
func (g *game) keyboardEvents() []laya.KeyboardEvent {
	var events []laya.KeyboardEvent

	// 获取修饰键状态
	modifiers := laya.KeyModifiers{
		Shift: ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight),
		Ctrl:  ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight),
		Alt:   ebiten.IsKeyPressed(ebiten.KeyAltLeft) || ebiten.IsKeyPressed(ebiten.KeyAltRight),
		Meta:  ebiten.IsKeyPressed(ebiten.KeyMetaLeft) || ebiten.IsKeyPressed(ebiten.KeyMetaRight),
	}

	// 收集字符输入（由 Update 中的 widget.InputChar 统一处理）
	// 这里只生成控制键事件，不再消费 AppendInputChars

	// 检查特殊按键(只在刚按下时触发)
	specialKeys := map[ebiten.Key]laya.KeyCode{
		ebiten.KeyBackspace: laya.KeyCodeBackspace,
		ebiten.KeyTab:       laya.KeyCodeTab,
		ebiten.KeyEnter:     laya.KeyCodeEnter,
		ebiten.KeyEscape:    laya.KeyCodeEscape,
		ebiten.KeySpace:     laya.KeyCodeSpace,
		ebiten.KeyLeft:      laya.KeyCodeLeft,
		ebiten.KeyUp:        laya.KeyCodeUp,
		ebiten.KeyRight:     laya.KeyCodeRight,
		ebiten.KeyDown:      laya.KeyCodeDown,
		ebiten.KeyDelete:    laya.KeyCodeDelete,
		ebiten.KeyHome:      laya.KeyCodeHome,
		ebiten.KeyEnd:       laya.KeyCodeEnd,
		ebiten.KeyA:         laya.KeyCodeA,
		ebiten.KeyC:         laya.KeyCodeC,
		ebiten.KeyV:         laya.KeyCodeV,
		ebiten.KeyX:         laya.KeyCodeX,
		ebiten.KeyZ:         laya.KeyCodeZ,
		ebiten.KeyF1:        laya.KeyCodeF1, // 添加F1键用于显示调试信息
	}

	// 遍历所有按键,检测状态变化
	for ebitenKey, layaCode := range specialKeys {
		isPressed := ebiten.IsKeyPressed(ebitenKey)
		wasPressed := g.keysDown[ebitenKey]

		// 按键刚按下(按下事件)
		if isPressed && !wasPressed {
			// 特殊处理F1键 - 显示调试信息
			if ebitenKey == ebiten.KeyF1 {
				g.showDebugInfo()
			}
			events = append(events, laya.KeyboardEvent{
				Code:      layaCode,
				Down:      true,
				Modifiers: modifiers,
			})
			g.keysDown[ebitenKey] = true
		}

		// 按键刚松开(抬起事件)
		if !isPressed && wasPressed {
			events = append(events, laya.KeyboardEvent{
				Code:      layaCode,
				Down:      false,
				Modifiers: modifiers,
			})
			g.keysDown[ebitenKey] = false
		}
	}

	return events
}

// showDebugInfo 显示调试信息
func (g *game) showDebugInfo() {
	if g.debug == nil || !g.debug.IsEnabled() {
		log.Println("调试服务器未启用")
		return
	}

	url := g.debug.GetURL()
	log.Printf("=== FairyGUI Debug Information ===")
	log.Printf("🌐 Debug Server: %s", url)
	log.Printf("📊 Tree View: %s/tree", url)
	log.Printf("📋 JSON API: %s/api/tree", url)
	log.Printf("🔄 Virtual Lists: %s/api/virtual-lists", url)
	log.Printf("=====================================")

	// 显示当前场景信息
	if g.manager != nil {
		currentScene := g.manager.Current()
		if currentScene != nil {
			log.Printf("🎮 Current Scene: %s", currentScene.Name())
		}

		// 显示根组件的子对象数量
		if g.root != nil {
			childrenCount := len(g.root.Children())
			log.Printf("🏗️  Root Children: %d", childrenCount)
		}
	}
}

func frameDelta(previous time.Time, now time.Time) time.Duration {
	if previous.IsZero() {
		return time.Second / 60
	}
	delta := now.Sub(previous)
	if delta <= 0 {
		return time.Second / 60
	}
	if delta > time.Second {
		return time.Second
	}
	return delta
}

func loadPreferredFont(px float64) (font.Face, error) {
	face, path, err := render.LoadSystemFont(px)
	if err != nil {
		return nil, err
	}
	log.Printf("loaded system font: %s", path)
	return face, nil
}
