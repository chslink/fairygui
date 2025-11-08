package scenes

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/gears"
	"github.com/chslink/fairygui/pkg/fgui/tween"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// BasicsDemo mirrors the logic-driven samples from BasicsDemo.ts.
type BasicsDemo struct {
	view          *core.GComponent
	backBtn       *core.GObject
	container     *core.GComponent
	controller    *core.Controller
	env           *Environment
	demos         map[string]*demoInfo
	progressTick  func()
	progressAlive bool
	winA          *windowInstance
	winB          *windowInstance
	popupMenu     *core.GComponent
	popupItems    []string
	popupOverlay  *core.GComponent
	dragCtx       *dragContext
	stageMove     laya.Listener
	stageUp       laya.Listener
	mainButtons   []*core.GObject
}

type demoInfo struct {
	component   *core.GComponent
	initialized bool
}

// NewBasicsDemo constructs the Basics scene.
func NewBasicsDemo() Scene {
	return &BasicsDemo{
		demos: make(map[string]*demoInfo),
	}
}

func (d *BasicsDemo) Name() string {
	return "BasicsDemo"
}

func (d *BasicsDemo) Load(ctx context.Context, mgr *Manager) (*core.GComponent, error) {
	env := mgr.Environment()
	d.env = env

	if _, err := env.Package(ctx, "Basics"); err != nil {
		return nil, err
	}
	pkg, err := env.Package(ctx, "Basics")
	if err != nil {
		return nil, err
	}
	item := chooseComponent(pkg, "Main")
	if item == nil {
		return nil, newMissingComponentError("Basics", "Main")
	}
	view, err := env.Factory.BuildComponent(ctx, pkg, item)
	if err != nil {
		return nil, err
	}
	d.view = view
	d.container = childAsComponent(view, "container")
	d.controller = view.ControllerByName("c1")
	d.backBtn = view.ChildByName("btn_Back")
	if d.backBtn != nil {
		d.backBtn.SetVisible(false)
		if sprite := d.backBtn.DisplayObject(); sprite != nil {
			sprite.Dispatcher().On(laya.EventClick, func(*laya.Event) {
				d.showMainMenu()
			})
		}
	}

	// Wire main buttons.
	for _, child := range view.Children() {
		if child == nil {
			continue
		}
		if grp := child.Group(); grp != nil {
			if grp.Name() != "" && strings.EqualFold(grp.Name(), "btns") {
				d.attachDemoButton(child)
			}
		}
	}

	return view, nil
}

func (d *BasicsDemo) Dispose() {
	d.progressAlive = false
	d.progressTick = nil
	if d.winA != nil {
		d.winA.Dispose()
		d.winA = nil
	}
	if d.winB != nil {
		d.winB.Dispose()
		d.winB = nil
	}
	d.popupMenu = nil
	d.popupOverlay = nil
	d.popupItems = nil
	d.dragCtx = nil
	d.removeStageDragHandlers()
	d.view = nil
	d.container = nil
	d.controller = nil
	d.demos = make(map[string]*demoInfo)
	d.mainButtons = nil
}

func (d *BasicsDemo) attachDemoButton(obj *core.GObject) {
	if obj == nil {
		return
	}
	name := obj.Name()
	if !strings.HasPrefix(name, "btn_") {
		return
	}
	btnName := strings.TrimPrefix(name, "btn_")
	if btnName == "" {
		return
	}
	log.Printf("[basics] attach demo button %s", btnName)
	if sprite := obj.DisplayObject(); sprite != nil {
		d.mainButtons = append(d.mainButtons, obj)
		scene := btnName
		sprite.Dispatcher().On(laya.EventClick, func(*laya.Event) {
			log.Printf("[basics] click demo button %s", btnName)
			d.runDemo(scene)
		})
	}
}

func (d *BasicsDemo) runDemo(kind string) {
	if d.controller == nil || d.container == nil {
		return
	}

	d.stopProgressLoop()

	info := d.ensureDemo(kind)
	if info == nil || info.component == nil {
		return
	}

	if !info.initialized {
		d.initializeDemo(kind, info)
		info.initialized = true
	}

	removeAllChildren(d.container)
	d.container.AddChild(info.component.GObject)
	d.setControllerIndex(1)
	d.setMainButtonsVisible(false)

	if d.backBtn != nil {
		d.backBtn.SetVisible(true)
	}
}

func (d *BasicsDemo) ensureDemo(kind string) *demoInfo {
	if info, ok := d.demos[kind]; ok {
		return info
	}
	pkg, err := d.env.Package(context.Background(), "Basics")
	if err != nil {
		log.Printf("[basics] load package failed: %v", err)
		return nil
	}
	item := chooseComponent(pkg, "Demo_"+kind)
	if item == nil {
		log.Printf("[basics] component Demo_%s missing", kind)
		return nil
	}
	component, err := d.env.Factory.BuildComponent(context.Background(), pkg, item)
	if err != nil {
		log.Printf("[basics] build Demo_%s failed: %v", kind, err)
		return nil
	}
	info := &demoInfo{component: component}
	d.demos[kind] = info
	return info
}

func (d *BasicsDemo) initializeDemo(kind string, info *demoInfo) {
	switch kind {
	case "Button":
		d.initButton(info.component)
	case "Text":
		d.initText(info.component)
	case "Grid":
		d.initGrid(info.component)
	case "ProgressBar":
		d.initProgress(info.component)
	case "List":
		d.initList(info.component)
	case "Window":
		d.initWindow(info.component)
	case "Popup":
		d.initPopup(info.component)
	case "Drag&Drop":
		d.initDragDrop(info.component)
	// 注：Depth 场景需要 sortingOrder 和 draggable 功能，待实现后添加
	// 注：TypeScript 原版没有实现 Loader、MovieClip、Image、Graph 等场景的交互
	// 这些场景只展示静态内容，不需要额外初始化
	default:
		log.Printf("[basics] demo %s has no explicit init", kind)
	}
}

func (d *BasicsDemo) initButton(component *core.GComponent) {
	if component == nil {
		return
	}
	// 参考 TypeScript 原版：BasicsDemo.ts playButton()
	// 只给 n34 添加简单的点击事件
	target := component.ChildByName("btn_Button")
	if target != nil {
		if sprite := target.DisplayObject(); sprite != nil {
			sprite.Dispatcher().On(laya.EventClick, func(*laya.Event) {
				log.Printf("[basics] ✅ click button - EVENT RECEIVED!")
			})
		}
	}
}

func (d *BasicsDemo) initText(component *core.GComponent) {
	if component == nil {
		return
	}
	const linkTemplate = "[img]ui://9leh0eyft9fj5f[/img][color=#FF0000]你点击了链接[/color]：%s"
	rich := component.ChildByName("n12")
	if rich != nil {
		if sprite := rich.DisplayObject(); sprite != nil {
			sprite.Dispatcher().On(laya.EventLink, func(evt *laya.Event) {
				link := ""
				if str, ok := evt.Data.(string); ok {
					link = str
				} else if evt.Data != nil {
					link = fmt.Sprint(evt.Data)
				}
				if setter, ok := rich.Data().(interface{ SetText(string) }); ok && setter != nil {
					setter.SetText(fmt.Sprintf(linkTemplate, link))
				}
			})
		}
	}
	copyBtn := component.ChildByName("n25")
	dest := component.ChildByName("n24")
	src := component.ChildByName("n22")
	if copyBtn == nil || dest == nil || src == nil {
		return
	}
	if sprite := copyBtn.DisplayObject(); sprite != nil {
		sprite.Dispatcher().On(laya.EventClick, func(*laya.Event) {
			var content string
			if reader, ok := src.Data().(interface{ Text() string }); ok && reader != nil {
				content = reader.Text()
			}
			if writer, ok := dest.Data().(interface{ SetText(string) }); ok && writer != nil {
				writer.SetText(content)
			}
		})
	}

	// 为输入框添加焦点处理
	if inputBox := component.ChildByName("n22"); inputBox != nil {
		if sprite := inputBox.DisplayObject(); sprite != nil {
			// 点击时请求焦点以显示光标
			sprite.Dispatcher().On(laya.EventClick, func(*laya.Event) {
				if input, ok := inputBox.Data().(*widgets.GTextInput); ok && input != nil {
					input.RequestFocus()
				}
			})
		}
		// 启动时自动请求焦点，直接显示光标
		if input, ok := inputBox.Data().(*widgets.GTextInput); ok && input != nil {
			input.RequestFocus()
		}
	}
}

func (d *BasicsDemo) initGrid(component *core.GComponent) {
	if component == nil {
		return
	}
	names := []string{"苹果手机操作系统", "安卓手机操作系统", "微软手机操作系统", "微软桌面操作系统", "苹果桌面操作系统", "未知操作系统"}
	colors := []string{"#FFFF00", "#FF0000", "#FFFFFF", "#0000FF"}

	fillList := func(obj *core.GObject, updater func(item *core.GComponent, idx int)) {
		if obj == nil {
			return
		}
		if list, ok := obj.Data().(*widgets.GList); ok && list != nil {
			items := list.GComponent.Children()
			for i := 0; i < len(items) && i < len(names); i++ {
				if child := items[i]; child != nil {
					if childComp := core.ComponentFrom(child); childComp != nil {
						updater(childComp, i)
					}
				}
			}
		}
	}

	fillList(component.ChildByName("list1"), func(item *core.GComponent, idx int) {
		setComponentText(item, "t0", strconv.Itoa(idx+1))
		setComponentText(item, "t1", names[idx])
		setComponentColor(item, "t2", colors[rand.Intn(len(colors))])
		if star := item.ChildByName("star"); star != nil {
			if bar := getProgressBar(star); bar != nil {
				bar.SetMin(0)
				bar.SetMax(100)
				bar.SetValue(float64((rand.Intn(3) + 1) * 33))
			}
		}
	})

	fillList(component.ChildByName("list2"), func(item *core.GComponent, idx int) {
		if cb := item.ChildByName("cb"); cb != nil {
			if btn, ok := cb.Data().(*widgets.GButton); ok {
				btn.SetSelected(false)
			}
		}
		// 参考 TypeScript 原版：BasicsDemo.ts playGrid()
		// 设置 MovieClip 的播放状态：偶数索引播放，奇数索引停止
		if mc := item.ChildByName("mc"); mc != nil {
			if clip, ok := mc.Data().(*widgets.GMovieClip); ok && clip != nil {
				clip.SetPlaying(idx%2 == 0)
			}
		}
		setComponentText(item, "t1", names[idx])
		setComponentText(item, "t3", strconv.Itoa(rand.Intn(10000)))
	})
}

func (d *BasicsDemo) initList(component *core.GComponent) {
	if component == nil {
		return
	}
	entries := []listEntry{
		{title: "100", icon: "ui://9leh0eyfkpev64"},
		{title: "1", icon: "ui://9leh0eyfkpev64"},
		{title: "2", icon: "ui://9leh0eyfkpev64"},
		{title: "99", icon: "ui://9leh0eyfkpev64"},
		{title: "4", icon: "ui://9leh0eyfkpev64"},
		{title: "5", icon: "ui://9leh0eyfkpev64"},
	}

	status := widgets.NewText()
	status.SetFontSize(20)
	status.SetColor("#333333")
	status.SetText("选择任意条目以查看选中结果")
	status.GObject.SetPivot(0, 0)
	status.GObject.SetSize(420, 32)
	status.GObject.SetPosition(34, component.Height()-48)
	component.AddChild(status.GObject)

	type listConfig struct {
		name          string
		label         string
		mode          widgets.ListSelectionMode
		scrollToView  bool
		defaultSelect int
	}
	configs := []listConfig{
		{name: "n0", label: "n2", mode: widgets.ListSelectionModeSingle, scrollToView: false, defaultSelect: 0},
		{name: "n4", label: "n5", mode: widgets.ListSelectionModeSingle, scrollToView: true, defaultSelect: 1},
		{name: "n7", label: "n8", mode: widgets.ListSelectionModeMultipleSingleClick, scrollToView: true, defaultSelect: -1},
		{name: "n9", label: "n10", mode: widgets.ListSelectionModeSingle, scrollToView: true, defaultSelect: 2},
	}

	for _, cfg := range configs {
		list := childAsList(component, cfg.name)
		if list == nil {
			continue
		}
		list.SetSelectionMode(cfg.mode)
		list.SetScrollItemToViewOnClick(cfg.scrollToView)
		applyListEntries(list, entries)
		if cfg.defaultSelect >= 0 {
			list.SetSelectedIndex(cfg.defaultSelect)
		}

		var baseLabel string
		var labelField *widgets.GTextField
		if labelObj := component.ChildByName(cfg.label); labelObj != nil {
			if txt, ok := labelObj.Data().(*widgets.GTextField); ok {
				labelField = txt
				baseLabel = txt.Text()
			}
		}

		updateLabel := func() {
			text := formatListLabel(baseLabel, list)
			if labelField != nil {
				labelField.SetText(text)
			} else {
				status.SetText(text)
			}
		}
		updateLabel()

		list.GComponent.GObject.On(laya.EventStateChanged, func(*laya.Event) {
			updateLabel()
		})
	}
}

func (d *BasicsDemo) initProgress(component *core.GComponent) {
	if component == nil {
		return
	}
	d.progressAlive = true
	if d.progressTick == nil {
		d.progressTick = func() {
			if !d.progressAlive {
				return
			}
			children := component.Children()
			for _, child := range children {
				if bar, ok := child.Data().(*widgets.GProgressBar); ok {
					next := bar.Value() + 1
					if next > bar.Max() {
						next = bar.Min()
					}
					bar.SetValue(next)
				}
			}
		}
		if scheduler := core.Root().Scheduler(); scheduler != nil {
			scheduler.Every(33*time.Millisecond, d.progressTick)
		}
	}
}

func (d *BasicsDemo) initWindow(component *core.GComponent) {
	if component == nil {
		return
	}
	if btn := component.ChildByName("n0"); btn != nil {
		if disp := btn.DisplayObject(); disp != nil {
			disp.Dispatcher().On(laya.EventClick, func(*laya.Event) {
				d.showWindowA()
			})
		}
	}
	if btn := component.ChildByName("n1"); btn != nil {
		if disp := btn.DisplayObject(); disp != nil {
			disp.Dispatcher().On(laya.EventClick, func(*laya.Event) {
				d.showWindowB()
			})
		}
	}
}

func (d *BasicsDemo) initPopup(component *core.GComponent) {
	if component == nil {
		return
	}
	if d.popupItems == nil {
		d.popupItems = []string{"Item 1", "Item 2", "Item 3", "Item 4"}
	}
	if btn := component.ChildByName("n0"); btn != nil {
		if disp := btn.DisplayObject(); disp != nil {
			anchor := btn
			disp.Dispatcher().On(laya.EventClick, func(*laya.Event) {
				d.showPopupMenu(anchor)
			})
		}
	}
	if btn := component.ChildByName("n1"); btn != nil {
		if disp := btn.DisplayObject(); disp != nil {
			anchor := btn
			disp.Dispatcher().On(laya.EventClick, func(*laya.Event) {
				d.showPopupOverlay(anchor)
			})
		}
	}
}

func (d *BasicsDemo) initDragDrop(component *core.GComponent) {
	if component == nil {
		return
	}
	d.ensureStageDragHandlers()

	btnA := component.ChildByName("a")
	d.makeDraggable(btnA, dragOptions{})

	btnBObj := component.ChildByName("b")
	btnCObj := component.ChildByName("c")
	if btnBObj != nil && btnCObj != nil {
		d.makeDraggable(btnBObj, dragOptions{
			Payload: func() any {
				if btn, ok := btnBObj.Data().(*widgets.GButton); ok {
					return btn.Icon()
				}
				return nil
			},
			DropTargets: []dragTarget{
				{
					Object: btnCObj,
					Handler: func(payload any) {
						if payload == nil {
							return
						}
						icon, ok := payload.(string)
						if !ok {
							return
						}
						if btn, ok := btnCObj.Data().(*widgets.GButton); ok {
							btn.SetIcon(icon)
						}
					},
				},
			},
		})
	}

	btnD := component.ChildByName("d")
	if btnD != nil {
		boundsObj := component.ChildByName("bounds")
		parent := btnD.Parent()
		d.makeDraggable(btnD, dragOptions{
			Bounds: func() *laya.Rect {
				if boundsObj == nil || parent == nil {
					return nil
				}
				rect := projectBounds(boundsObj, parent)
				return &rect
			},
		})
	}
}

func (d *BasicsDemo) showWindowA() {
	if d.winA == nil {
		comp, err := d.buildBasicsComponent(context.Background(), "WindowA")
		if err != nil {
			log.Printf("[basics] load WindowA failed: %v", err)
			return
		}
		instance := newWindowInstance(comp)
		d.attachWindowClose(comp, instance.Hide)
		instance.onShow = func(w *windowInstance) {
			d.populateWindowA()
		}
		d.winA = instance
	}
	d.winA.Show()
}

func (d *BasicsDemo) showWindowB() {
	if d.winB == nil {
		comp, err := d.buildBasicsComponent(context.Background(), "WindowB")
		if err != nil {
			log.Printf("[basics] load WindowB failed: %v", err)
			return
		}
		comp.GObject.SetPivotWithAnchor(0.5, 0.5, true)
		instance := newWindowInstance(comp)
		instance.onShow = func(w *windowInstance) {
			if w == nil || w.component == nil {
				return
			}
			w.animateScale(0.1, 0.1, 1, 1, 0.3, func() {
				if tx := w.component.Transition("t1"); tx != nil {
					tx.Play(1, 0)
				}
			})
		}
		instance.onHide = func(w *windowInstance) {
			if w == nil || w.component == nil {
				return
			}
			if tx := w.component.Transition("t1"); tx != nil {
				tx.Stop(false)
			}
			w.animateScale(1, 1, 0.1, 0.1, 0.3, func() {
				w.finishHide()
			})
		}
		d.attachWindowClose(comp, instance.Hide)
		d.winB = instance
	}
	d.winB.Show()
}

func (d *BasicsDemo) populateWindowA() {
	if d.winA == nil || d.winA.component == nil {
		return
	}
	listObj := d.winA.component.ChildByName("n6")
	if listObj == nil {
		return
	}
	listComp := core.ComponentFrom(listObj)
	if listComp == nil {
		return
	}
	children := listComp.Children()
	for i, child := range children {
		if child == nil {
			continue
		}
		switch data := child.Data().(type) {
		case *widgets.GButton:
			data.SetTitle(strconv.Itoa(i))
		case *widgets.GLabel:
			data.SetTitle(strconv.Itoa(i))
		default:
			child.SetData(strconv.Itoa(i))
		}
	}
}

func (d *BasicsDemo) stopProgressLoop() {
	d.progressAlive = false
}

func (d *BasicsDemo) showMainMenu() {
	d.stopProgressLoop()
	d.setControllerIndex(0)
	d.setMainButtonsVisible(true)
	if d.backBtn != nil {
		d.backBtn.SetVisible(false)
	}
	if d.container != nil {
		removeAllChildren(d.container)
	}
}

func (d *BasicsDemo) setMainButtonsVisible(visible bool) {
	for _, btn := range d.mainButtons {
		if btn == nil {
			continue
		}
		btn.SetVisible(visible)
	}
	if d.view != nil {
		if group := d.view.ChildByName("btns"); group != nil {
			group.SetVisible(visible)
		}
	}
}

func (d *BasicsDemo) setControllerIndex(index int) {
	if d.controller == nil {
		return
	}
	prev := gears.DisableAllTweenEffect
	gears.DisableAllTweenEffect = true
	d.controller.SetSelectedIndex(index)
	gears.DisableAllTweenEffect = prev
}

func (d *BasicsDemo) buildBasicsComponent(ctx context.Context, name string) (*core.GComponent, error) {
	if d.env == nil {
		return nil, newMissingComponentError("Basics", name)
	}
	pkg, err := d.env.Package(ctx, "Basics")
	if err != nil {
		return nil, err
	}
	item := chooseComponent(pkg, name)
	if item == nil {
		return nil, newMissingComponentError("Basics", name)
	}
	comp, err := d.env.Factory.BuildComponent(ctx, pkg, item)
	if err != nil {
		return nil, err
	}
	return comp, nil
}

func (d *BasicsDemo) attachWindowClose(comp *core.GComponent, hide func()) {
	if comp == nil || hide == nil {
		return
	}
	frame := childAsComponent(comp, "frame")
	if frame == nil {
		return
	}
	if closeBtn := frame.ChildByName("closeButton"); closeBtn != nil {
		if disp := closeBtn.DisplayObject(); disp != nil {
			disp.Dispatcher().On(laya.EventClick, func(*laya.Event) {
				hide()
			})
		}
	}
}

func (d *BasicsDemo) showPopupMenu(anchor *core.GObject) {
	menu := d.ensurePopupMenu()
	if menu == nil {
		return
	}
	root := core.Root()
	if root == nil {
		return
	}
	root.ShowPopup(menu.GObject, anchor, core.PopupDirectionDown)
}

func (d *BasicsDemo) ensurePopupMenu() *core.GComponent {
	if d.popupMenu != nil {
		return d.popupMenu
	}
	items := d.popupItems
	if len(items) == 0 {
		items = []string{"Item 1", "Item 2", "Item 3", "Item 4"}
	}
	menu := d.buildSimplePopupMenu(items)
	d.popupMenu = menu
	return menu
}

func (d *BasicsDemo) buildSimplePopupMenu(items []string) *core.GComponent {
	menu := core.NewGComponent()
	menu.GObject.SetTouchable(true)
	padding := 6.0
	itemHeight := 28.0
	width := 160.0
	height := padding*2 + itemHeight*float64(len(items))
	if height < 1 {
		height = padding * 2
	}
	menu.GObject.SetSize(width, height)

	bg := widgets.NewGraph()
	bg.SetType(widgets.GraphTypeRect)
	bg.SetFillColor("#2d2d2d")
	bg.SetLineSize(1)
	bg.SetLineColor("#555555")
	bg.GObject.SetTouchable(false)
	bg.GObject.SetSize(width, height)
	menu.AddChild(bg.GObject)

	for i, text := range items {
		label := text
		item := core.NewGObject()
		item.SetSize(width-2*padding, itemHeight)
		item.SetPosition(padding, padding+itemHeight*float64(i))
		item.SetTouchable(true)
		item.SetData(label)
		menu.AddChild(item)
		if disp := item.DisplayObject(); disp != nil {
			disp.Dispatcher().On(laya.EventClick, func(*laya.Event) {
				log.Printf("[basics] popup item selected: %s", label)
				core.Root().HidePopup(menu.GObject)
			})
		}
	}
	return menu
}

func (d *BasicsDemo) showPopupOverlay(anchor *core.GObject) {
	overlay := d.ensurePopupOverlay()
	if overlay == nil {
		return
	}
	root := core.Root()
	if root == nil {
		return
	}
	centerComponent(overlay)
	root.ShowPopup(overlay.GObject, anchor, core.PopupDirectionAuto)
}

func (d *BasicsDemo) ensurePopupOverlay() *core.GComponent {
	if d.popupOverlay != nil {
		return d.popupOverlay
	}
	comp, err := d.buildBasicsComponent(context.Background(), "Component12")
	if err != nil {
		log.Printf("[basics] load popup overlay failed: %v", err)
		return nil
	}
	comp.GObject.SetTouchable(true)
	d.popupOverlay = comp
	return comp
}

func (d *BasicsDemo) ensureStageDragHandlers() {
	if d.stageMove != nil || d.stageUp != nil {
		return
	}
	stage := core.Root().Stage()
	if stage == nil || stage.Root() == nil {
		return
	}
	dispatcher := stage.Root().Dispatcher()
	if dispatcher == nil {
		return
	}
	d.stageMove = func(evt *laya.Event) {
		d.onStageDragMove(evt)
	}
	d.stageUp = func(evt *laya.Event) {
		d.onStageDragUp(evt)
	}
	dispatcher.On(laya.EventMouseMove, d.stageMove)
	dispatcher.On(laya.EventStageMouseUp, d.stageUp)
}

func (d *BasicsDemo) removeStageDragHandlers() {
	stage := core.Root().Stage()
	if stage != nil && stage.Root() != nil {
		if dispatcher := stage.Root().Dispatcher(); dispatcher != nil {
			if d.stageMove != nil {
				dispatcher.Off(laya.EventMouseMove, d.stageMove)
			}
			if d.stageUp != nil {
				dispatcher.Off(laya.EventStageMouseUp, d.stageUp)
			}
		}
	}
	d.stageMove = nil
	d.stageUp = nil
	d.dragCtx = nil
}

func (d *BasicsDemo) makeDraggable(obj *core.GObject, opts dragOptions) {
	if obj == nil {
		return
	}
	obj.SetTouchable(true)
	if disp := obj.DisplayObject(); disp != nil {
		disp.Dispatcher().On(laya.EventMouseDown, func(evt *laya.Event) {
			if d.dragCtx != nil && d.dragCtx.active {
				return
			}
			pe, ok := evt.Data.(laya.PointerEvent)
			if !ok {
				return
			}
			d.startDrag(obj, pe, opts)
		})
	}
}

func (d *BasicsDemo) startDrag(obj *core.GObject, pe laya.PointerEvent, opts dragOptions) {
	if obj == nil || obj.Parent() == nil || obj.Parent().DisplayObject() == nil {
		return
	}
	d.ensureStageDragHandlers()
	parent := obj.Parent()
	if pe.TouchID <= 0 {
		pe.TouchID = 1
	}
	parentLocal := parent.DisplayObject().GlobalToLocal(pe.Position)
	offset := laya.Point{
		X: parentLocal.X - obj.X(),
		Y: parentLocal.Y - obj.Y(),
	}

	var bounds *laya.Rect
	if opts.Bounds != nil {
		if rect := opts.Bounds(); rect != nil {
			copyRect := *rect
			bounds = &copyRect
		}
	}

	var payload any
	if opts.Payload != nil {
		payload = opts.Payload()
	}

	targets := make([]dragTarget, len(opts.DropTargets))
	copy(targets, opts.DropTargets)

	d.dragCtx = &dragContext{
		active:      true,
		obj:         obj,
		pointerID:   pe.TouchID,
		offset:      offset,
		bounds:      bounds,
		dropTargets: targets,
		payload:     payload,
	}
}

func (d *BasicsDemo) onStageDragMove(evt *laya.Event) {
	if d.dragCtx == nil || !d.dragCtx.active {
		return
	}
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok || pe.TouchID != d.dragCtx.pointerID {
		return
	}
	obj := d.dragCtx.obj
	if obj == nil || obj.Parent() == nil || obj.Parent().DisplayObject() == nil {
		return
	}
	parent := obj.Parent()
	parentLocal := parent.DisplayObject().GlobalToLocal(pe.Position)
	newX := parentLocal.X - d.dragCtx.offset.X
	newY := parentLocal.Y - d.dragCtx.offset.Y
	if bounds := d.dragCtx.bounds; bounds != nil {
		maxX := bounds.X + bounds.W - obj.Width()
		maxY := bounds.Y + bounds.H - obj.Height()
		if newX < bounds.X {
			newX = bounds.X
		}
		if newY < bounds.Y {
			newY = bounds.Y
		}
		if obj.Width() > 0 {
			if newX > maxX {
				newX = maxX
			}
		}
		if obj.Height() > 0 {
			if newY > maxY {
				newY = maxY
			}
		}
	}
	obj.SetPosition(newX, newY)
}

func (d *BasicsDemo) onStageDragUp(evt *laya.Event) {
	if d.dragCtx == nil || !d.dragCtx.active {
		return
	}
	pe, ok := evt.Data.(laya.PointerEvent)
	if !ok || pe.TouchID != d.dragCtx.pointerID {
		return
	}
	ctx := d.dragCtx
	ctx.active = false

	if len(ctx.dropTargets) > 0 {
		point := pe.Position
		for _, target := range ctx.dropTargets {
			if target.Object == nil {
				continue
			}
			rect := globalRect(target.Object)
			if rect.W == 0 && rect.H == 0 {
				continue
			}
			if pointInRect(point, rect) {
				if target.Handler != nil {
					target.Handler(ctx.payload)
				}
				break
			}
		}
	}

	d.dragCtx = nil
}

func projectBounds(source *core.GObject, parent *core.GComponent) laya.Rect {
	if source == nil || parent == nil || source.DisplayObject() == nil || parent.DisplayObject() == nil {
		return laya.Rect{}
	}
	bounds := source.DisplayObject().Bounds()
	topLeft := parent.DisplayObject().GlobalToLocal(laya.Point{X: bounds.X, Y: bounds.Y})
	bottomRight := parent.DisplayObject().GlobalToLocal(laya.Point{X: bounds.X + bounds.W, Y: bounds.Y + bounds.H})
	if bottomRight.X < topLeft.X {
		topLeft.X, bottomRight.X = bottomRight.X, topLeft.X
	}
	if bottomRight.Y < topLeft.Y {
		topLeft.Y, bottomRight.Y = bottomRight.Y, topLeft.Y
	}
	return laya.Rect{
		X: topLeft.X,
		Y: topLeft.Y,
		W: bottomRight.X - topLeft.X,
		H: bottomRight.Y - topLeft.Y,
	}
}

func globalRect(obj *core.GObject) laya.Rect {
	if obj == nil || obj.DisplayObject() == nil {
		return laya.Rect{}
	}
	return obj.DisplayObject().Bounds()
}

func pointInRect(pt laya.Point, rect laya.Rect) bool {
	if rect.W < 0 {
		rect.X = rect.X + rect.W
		rect.W = -rect.W
	}
	if rect.H < 0 {
		rect.Y = rect.Y + rect.H
		rect.H = -rect.H
	}
	if rect.W == 0 || rect.H == 0 {
		return false
	}
	return pt.X >= rect.X && pt.X <= rect.X+rect.W &&
		pt.Y >= rect.Y && pt.Y <= rect.Y+rect.H
}

func centerComponent(comp *core.GComponent) {
	if comp == nil {
		return
	}
	root := core.Root()
	if root == nil {
		return
	}
	width := comp.Width()
	if width <= 0 {
		width = comp.GObject.Width()
	}
	height := comp.Height()
	if height <= 0 {
		height = comp.GObject.Height()
	}
	rootWidth := float64(root.Width())
	rootHeight := float64(root.Height())
	x := (rootWidth - width) / 2
	y := (rootHeight - height) / 2
	comp.GObject.SetPosition(x, y)
}

type windowInstance struct {
	component *core.GComponent
	visible   bool
	tweener   *tween.GTweener
	onShow    func(*windowInstance)
	onHide    func(*windowInstance)
}

func newWindowInstance(comp *core.GComponent) *windowInstance {
	return &windowInstance{component: comp}
}

func (w *windowInstance) Show() {
	if w == nil || w.component == nil {
		return
	}
	root := core.Root()
	if root == nil {
		return
	}
	obj := w.component.GObject
	if obj == nil {
		return
	}
	if obj.Parent() != nil {
		obj.Parent().RemoveChild(obj)
	}
	root.AddChild(obj)
	centerComponent(w.component)
	w.visible = true
	if w.onShow != nil {
		w.onShow(w)
	}
}

func (w *windowInstance) Hide() {
	if w == nil || !w.visible {
		return
	}
	if w.onHide != nil {
		w.onHide(w)
		return
	}
	w.finishHide()
}

func (w *windowInstance) finishHide() {
	if w == nil || w.component == nil {
		return
	}
	obj := w.component.GObject
	if obj != nil && obj.Parent() != nil {
		obj.Parent().RemoveChild(obj)
	}
	w.visible = false
}

func (w *windowInstance) Dispose() {
	if w == nil {
		return
	}
	if w.tweener != nil {
		w.tweener.Kill(false)
		w.tweener = nil
	}
	w.finishHide()
	w.component = nil
}

func (w *windowInstance) animateScale(fromX, fromY, toX, toY float64, duration float64, onComplete func()) {
	if w == nil || w.component == nil {
		return
	}
	if w.tweener != nil {
		w.tweener.Kill(false)
		w.tweener = nil
	}
	w.component.GObject.SetScale(fromX, fromY)
	w.tweener = tween.To2(fromX, fromY, toX, toY, duration).
		SetEase(tween.EaseTypeQuadOut).
		OnUpdate(func(tw *tween.GTweener) {
			val := tw.Value()
			w.component.GObject.SetScale(val.X, val.Y)
		}).
		OnComplete(func(*tween.GTweener) {
			w.component.GObject.SetScale(toX, toY)
			w.tweener = nil
			if onComplete != nil {
				onComplete()
			}
		})
}

type dragContext struct {
	active      bool
	obj         *core.GObject
	pointerID   int
	offset      laya.Point
	bounds      *laya.Rect
	dropTargets []dragTarget
	payload     any
}

type dragTarget struct {
	Object  *core.GObject
	Handler func(payload any)
}

type dragOptions struct {
	Bounds      func() *laya.Rect
	DropTargets []dragTarget
	Payload     func() any
}

func childAsComponent(parent *core.GComponent, name string) *core.GComponent {
	if parent == nil {
		return nil
	}
	child := parent.ChildByName(name)
	if child == nil {
		return nil
	}
	if comp, ok := child.Data().(*core.GComponent); ok {
		return comp
	}
	return nil
}

func removeAllChildren(comp *core.GComponent) {
	if comp == nil {
		return
	}
	children := comp.Children()
	for i := len(children) - 1; i >= 0; i-- {
		if children[i] != nil {
			comp.RemoveChild(children[i])
		}
	}
}

func setComponentText(comp *core.GComponent, childName, value string) {
	if comp == nil {
		return
	}
	child := comp.ChildByName(childName)
	if child == nil {
		return
	}
	switch data := child.Data().(type) {
	case *widgets.GTextField:
		data.SetText(value)
	case *widgets.GLabel:
		data.SetTitle(value)
	case *widgets.GButton:
		data.SetTitle(value)
	case nil:
		child.SetData(value)
	default:
		child.SetData(value)
	}
}

func setComponentColor(comp *core.GComponent, childName, color string) {
	child := comp.ChildByName(childName)
	if child == nil {
		return
	}
	switch data := child.Data().(type) {
	case *widgets.GTextField:
		data.SetColor(color)
	case *widgets.GLabel:
		data.SetTitleColor(color)
	case *widgets.GButton:
		data.SetTitleColor(color)
	}
}

func getProgressBar(obj *core.GObject) *widgets.GProgressBar {
	if obj == nil {
		return nil
	}
	if bar, ok := obj.Data().(*widgets.GProgressBar); ok {
		return bar
	}
	if comp, ok := obj.Data().(*core.GComponent); ok && comp != nil {
		if inner := comp.ChildByName("bar"); inner != nil {
			if b, ok := inner.Data().(*widgets.GProgressBar); ok {
				return b
			}
		}
	}
	return nil
}

func childAsList(parent *core.GComponent, name string) *widgets.GList {
	if parent == nil {
		return nil
	}
	child := parent.ChildByName(name)
	if child == nil {
		return nil
	}
	if list, ok := child.Data().(*widgets.GList); ok {
		return list
	}
	return nil
}

type listEntry struct {
	title string
	icon  string
}

func applyListEntries(list *widgets.GList, entries []listEntry) {
	if list == nil || len(entries) == 0 {
		return
	}
	shuffled := append([]listEntry(nil), entries...)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	apply := func(obj *core.GObject, idx int) {
		if obj == nil {
			return
		}
		entry := shuffled[idx%len(shuffled)]
		switch data := obj.Data().(type) {
		case *widgets.GButton:
			data.SetTitle(entry.title)
			data.SetIcon(entry.icon)
		case *widgets.GLabel:
			data.SetTitle(entry.title)
		case *core.GComponent:
			setComponentText(data, "title", entry.title)
			if iconChild := data.ChildByName("icon"); iconChild != nil {
				switch iconData := iconChild.Data().(type) {
				case *widgets.GLoader:
					iconData.SetURL(entry.icon)
				case *widgets.GButton:
					iconData.SetIcon(entry.icon)
				}
			}
		default:
			// no-op
		}
	}

	items := list.Items()
	if len(items) > 0 {
		for i, obj := range items {
			apply(obj, i)
		}
		return
	}
	children := list.GComponent.Children()
	for i, child := range children {
		apply(child, i)
	}
}

func formatListLabel(base string, list *widgets.GList) string {
	if list == nil {
		return base
	}
	idx := list.SelectedIndex()
	if idx < 0 {
		if base == "" {
			return "当前未选中条目"
		}
		return base
	}
	title := listItemTitle(list, idx)
	if base == "" {
		return fmt.Sprintf("选中：%s", title)
	}
	return fmt.Sprintf("%s（当前：%s）", base, title)
}

func listItemTitle(list *widgets.GList, index int) string {
	if list == nil || index < 0 {
		return ""
	}
	items := list.Items()
	var item *core.GObject
	if index < len(items) {
		item = items[index]
	} else {
		children := list.GComponent.Children()
		if index < len(children) {
			item = children[index]
		}
	}
	if item == nil {
		return fmt.Sprintf("#%d", index+1)
	}
	switch data := item.Data().(type) {
	case *widgets.GButton:
		if text := data.Title(); text != "" {
			return text
		}
	case *widgets.GLabel:
		if text := data.Title(); text != "" {
			return text
		}
	case *core.GComponent:
		if titleChild := data.ChildByName("title"); titleChild != nil {
			switch titleData := titleChild.Data().(type) {
			case *widgets.GTextField:
				if text := titleData.Text(); text != "" {
					return text
				}
			case *widgets.GLabel:
				if text := titleData.Title(); text != "" {
					return text
				}
			case *widgets.GButton:
				if text := titleData.Title(); text != "" {
					return text
				}
			}
		}
	}
	return fmt.Sprintf("#%d", index+1)
}
