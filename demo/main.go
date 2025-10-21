//go:build ebiten

package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/chslink/fairygui/demo/scenes"
	"github.com/chslink/fairygui/pkg/fgui"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/builder"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
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

	debugEnabled := os.Getenv("FGUI_DEBUG_COMPONENTS") != ""

	scene, err := NewScene(ctx, debugEnabled)
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
	scene      *Scene
	lastUpdate time.Time
}

func (g *Game) Update() error {
	if g.scene == nil {
		return nil
	}

	now := time.Now()
	var delta time.Duration
	if g.lastUpdate.IsZero() {
		delta = time.Second / 60
	} else {
		delta = now.Sub(g.lastUpdate)
		if delta <= 0 {
			delta = time.Second / 60
		} else if delta > time.Second {
			delta = time.Second
		}
	}
	g.lastUpdate = now

	cursorX, cursorY := ebiten.CursorPosition()
	mouse := g.scene.MouseState(cursorX, cursorY, ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft))
	g.scene.Advance(delta, mouse)
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
	atlas        *render.AtlasManager
	manager      *scenes.Manager
	root         *core.GRoot
	width        int
	height       int
	debugEnabled bool
	debugLastLog time.Time
}

func NewScene(ctx context.Context, debug bool) (*Scene, error) {
	assetsDir := filepath.Join("demo", "assets")
	loader := assets.NewFileLoader(assetsDir)

	atlas := render.NewAtlasManager(loader)
	factory := builder.NewFactoryWithLoader(atlas, loader)
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

	// 清理旧的显示树后将场景根节点挂到 GRoot。
	for _, child := range root.Children() {
		root.RemoveChild(child)
	}
	stageRoot := manager.Stage()
	stageRoot.GObject.SetPosition(0, 0)
	stageRoot.GObject.SetData(stageRoot)
	root.AddChild(stageRoot.GObject)

	scene := &Scene{
		atlas:        atlas,
		manager:      manager,
		root:         root,
		width:        manager.Width(),
		height:       manager.Height(),
		debugEnabled: debug,
	}

	if debug {
		logComponentTree(manager.Stage())
	}

	return scene, nil
}

func (s *Scene) Width() int  { return s.width }
func (s *Scene) Height() int { return s.height }

func (s *Scene) Advance(delta time.Duration, mouse fgui.MouseState) {
	if s.root == nil {
		return
	}
	s.syncRootSize()
	s.root.Advance(delta, mouse)
}

func (s *Scene) MouseState(x, y int, pressed bool) fgui.MouseState {
	if s.root == nil {
		return fgui.MouseState{}
	}
	px := float64(x)
	py := float64(y)
	if stage := s.root.Stage(); stage != nil {
		rootSprite := stage.Root()
		if rootSprite != nil {
			pos := rootSprite.Position()
			px -= pos.X
			py -= pos.Y
			if sx, sy := rootSprite.Scale(); sx != 0 || sy != 0 {
				if sx != 0 {
					px /= sx
				}
				if sy != 0 {
					py /= sy
				}
			}
		}
	}
	return fgui.MouseState{
		X:       px,
		Y:       py,
		Primary: pressed,
	}
}

func (s *Scene) syncRootSize() {
	if s.root == nil {
		return
	}
	width := s.manager.Width()
	height := s.manager.Height()
	if width == s.width && height == s.height {
		return
	}
	s.root.Resize(width, height)
	s.width = width
	s.height = height
}

func (s *Scene) Draw(screen *ebiten.Image) {
	if s.root == nil {
		return
	}
	s.syncRootSize()
	if s.debugEnabled && time.Since(s.debugLastLog) > 2*time.Second {
		if active := s.manager.Current(); active != nil {
			comp := s.manager.CurrentComponent()
			log.Printf("[debug] scene=%s size=%dx%d root=%p atlas=%t", active.Name(), s.width, s.height, comp, s.atlas != nil)
			s.dumpComponent(comp, 0)
		} else {
			log.Printf("[debug] no active scene; size=%dx%d", s.width, s.height)
		}
		s.debugLastLog = time.Now()
	}
	if err := render.DrawComponent(screen, s.root.GComponent, s.atlas); err != nil {
		log.Printf("render component failed: %v", err)
	}
	s.drawText(screen, s.root.GComponent, 0, 0)
}

func (s *Scene) dumpComponent(comp *core.GComponent, depth int) {
	if comp == nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	log.Printf("%s[component] name=%s id=%s visible=%t children=%d controllers=%d", indent, comp.Name(), comp.ID(), comp.Visible(), len(comp.Children()), len(comp.Controllers()))
	for _, ctrl := range comp.Controllers() {
		if ctrl == nil {
			continue
		}
		log.Printf("%s  [controller] name=%s selected=%d pages=%v", indent, ctrl.Name, ctrl.SelectedIndex(), ctrl.PageNames)
	}
	for _, child := range comp.Children() {
		if child == nil {
			continue
		}
		data := child.Data()
		var dataType string
		if data != nil {
			dataType = fmt.Sprintf("%T", data)
		} else {
			dataType = "<nil>"
		}
		log.Printf("%s  [child] name=%q id=%s rid=%s type=%s pos=(%.1f,%.1f) size=(%.1f,%.1f) visible=%t touchable=%t", indent, child.Name(), child.ID(), child.ResourceID(), dataType, child.X(), child.Y(), child.Width(), child.Height(), child.Visible(), child.Touchable())
		if compChild, ok := data.(*core.GComponent); ok {
			s.dumpComponent(compChild, depth+2)
			continue
		}
		switch widget := data.(type) {
		case *widgets.GButton:
			log.Printf("%s    [button] mode=%d title=%q selected=%t controller=%v related=%v template=%p", indent, widget.Mode(), widget.Title(), widget.Selected(), widget.ButtonController(), widget.RelatedController(), widget.TemplateComponent())
			if tpl := widget.TemplateComponent(); tpl != nil {
				log.Printf("%s      [template] name=%s id=%s visible=%t children=%d controllers=%d", indent, tpl.Name(), tpl.ID(), tpl.Visible(), len(tpl.Children()), len(tpl.Controllers()))
				for _, tplChild := range tpl.Children() {
					if tplChild == nil {
						continue
					}
					var tplType string
					childData := tplChild.Data()
					if childData != nil {
						tplType = fmt.Sprintf("%T", childData)
					} else {
						tplType = "<nil>"
					}
					if pkgItem, ok := childData.(*assets.PackageItem); ok && pkgItem != nil {
						log.Printf("%s        [template child] name=%q id=%s rid=%s type=%s pos=(%.1f,%.1f) size=(%.1f,%.1f) visible=%t sprite=%s scale9=%v",
							indent, tplChild.Name(), tplChild.ID(), tplChild.ResourceID(), tplType, tplChild.X(), tplChild.Y(), tplChild.Width(), tplChild.Height(), tplChild.Visible(),
							describePackageSprite(pkgItem), pkgItem.Scale9Grid != nil)
					} else {
						log.Printf("%s        [template child] name=%q id=%s rid=%s type=%s pos=(%.1f,%.1f) size=(%.1f,%.1f) visible=%t",
							indent, tplChild.Name(), tplChild.ID(), tplChild.ResourceID(), tplType, tplChild.X(), tplChild.Y(), tplChild.Width(), tplChild.Height(), tplChild.Visible())
					}
				}
			}
		case *assets.PackageItem:
			log.Printf("%s    [pkg-item] id=%s name=%s type=%d sprite=%s", indent, widget.ID, widget.Name, widget.Type, describePackageSprite(widget))
		case *widgets.GLabel:
			log.Printf("%s    [label] title=%q icon=%q", indent, widget.Title(), widget.Icon())
		case *widgets.GLoader:
			log.Printf("%s    [loader] url=%q playing=%t frame=%d", indent, widget.URL(), widget.Playing(), widget.Frame())
		}
	}
}

func describePackageSprite(item *assets.PackageItem) string {
	if item == nil || item.Sprite == nil || item.Sprite.Atlas == nil {
		return "<nil>"
	}
	rect := item.Sprite.Rect
	offset := item.Sprite.Offset
	atlas := item.Sprite.Atlas
	name := atlas.ID
	if name == "" {
		name = atlas.Name
	}
	return fmt.Sprintf("%s rect=%dx%d+%d+%d offset=(%d,%d)", name, rect.Width, rect.Height, rect.X, rect.Y, offset.X, offset.Y)
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
		case *widgets.GTextField:
			text.Draw(target, data.Text(), textFont, int(x), int(y)+textFont.Metrics().Ascent.Ceil(), color.White)
		}
	}
}

func drawHUD(screen *ebiten.Image, scene *Scene) {
	current := "none"
	if active := scene.manager.Current(); active != nil {
		current = active.Name()
	}
	lines := []string{
		fmt.Sprintf("Stage size: %dx%d", scene.Width(), scene.Height()),
		fmt.Sprintf("Active scene: %s", current),
	}
	y := 16
	for _, line := range lines {
		text.Draw(screen, line, basicfont.Face7x13, 16, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
		y += 16
	}
}

func loadPreferredFont(px float64) (font.Face, error) {
	face, path, err := render.LoadSystemFont(px)
	if err != nil {
		return nil, err
	}
	log.Printf("loaded system font: %s", path)
	return face, nil
}

func logComponentTree(root *core.GComponent) {
	if root == nil {
		log.Println("[scene] component tree: <nil>")
		return
	}
	log.Println("[scene] component tree:")
	visited := make(map[*core.GComponent]bool)
	logComponentNode(root, 0, "root", visited)
}

func logComponentNode(comp *core.GComponent, depth int, label string, visited map[*core.GComponent]bool) {
	if comp == nil {
		return
	}
	logComponentLine(comp.GObject, depth, label)
	if visited[comp] {
		log.Printf("%s  (component already logged)", strings.Repeat("  ", depth))
		return
	}
	visited[comp] = true
	for _, child := range comp.Children() {
		logComponentBranch(child, depth+1, visited)
	}
}

func logComponentBranch(obj *core.GObject, depth int, visited map[*core.GComponent]bool) {
	if obj == nil {
		return
	}
	logComponentLine(obj, depth, "")
	if nested, ok := obj.Data().(*core.GComponent); ok && nested != nil {
		logComponentNode(nested, depth+1, "nested", visited)
	}
}

func logComponentLine(obj *core.GObject, depth int, label string) {
	if obj == nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	builder := strings.Builder{}
	builder.WriteString(indent)
	builder.WriteString("- ")
	if label != "" {
		builder.WriteString(label)
		builder.WriteByte(' ')
	}
	builder.WriteString(fmt.Sprintf("id=%s", obj.ID()))
	if rid := obj.ResourceID(); rid != "" {
		builder.WriteString(fmt.Sprintf(" rid=%s", rid))
	}
	if name := obj.Name(); name != "" {
		builder.WriteString(fmt.Sprintf(" name=%q", name))
	}
	builder.WriteString(fmt.Sprintf(" pos=(%.1f,%.1f) size=(%.1f,%.1f)", obj.X(), obj.Y(), obj.Width(), obj.Height()))
	data := obj.Data()
	if data == nil {
		builder.WriteString(" data=<nil>")
	} else {
		builder.WriteString(fmt.Sprintf(" data=%T", data))
	}
	switch v := data.(type) {
	case *assets.PackageItem:
		builder.WriteString(fmt.Sprintf(" package=%s/%s type=%v", safePackageName(v.Owner), v.ID, v.Type))
	case *widgets.GLoader:
		builder.WriteString(" loader")
		if url := v.URL(); url != "" {
			builder.WriteString(fmt.Sprintf(" url=%s", url))
		}
		if item := v.PackageItem(); item != nil {
			builder.WriteString(fmt.Sprintf(" packageItem=%s", item.ID))
		}
		if nested := v.Component(); nested != nil {
			builder.WriteString(fmt.Sprintf(" component=%s", nested.ID()))
		}
	case *core.GComponent:
		builder.WriteString(fmt.Sprintf(" nestedComponent=%s", v.ID()))
	}
	log.Println(builder.String())
}

func safePackageName(pkg *assets.Package) string {
	if pkg == nil {
		return "<nil>"
	}
	if pkg.Name != "" {
		return pkg.Name
	}
	if pkg.ID != "" {
		return pkg.ID
	}
	return "<unnamed>"
}
