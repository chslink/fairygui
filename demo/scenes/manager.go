package scenes

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

const (
	mainMenuSceneName  = "MainMenu"
	closeButtonMargin  = 10.0
	defaultStageWidth  = 1136
	defaultStageHeight = 640
)

// Scene represents a runnable demo page.
type Scene interface {
	Name() string
	Load(ctx context.Context, mgr *Manager) (*core.GComponent, error)
	Dispose()
}

// FactoryFunc constructs a scene instance.
type FactoryFunc func() Scene

// Manager coordinates demo scene lifecycle, mirroring DemoEntry.ts from the TypeScript version.
type Manager struct {
	ctx      context.Context
	env      *Environment
	stage    *core.GComponent
	current  Scene
	root     *core.GComponent
	registry map[string]FactoryFunc
	width    int
	height   int
	closeBtn *core.GComponent
}

// NewManager instantiates the scene manager, registers built-in demos, and loads the main menu.
func NewManager(ctx context.Context, env *Environment) (*Manager, error) {
	mgr := &Manager{
		ctx:      ctx,
		env:      env,
		stage:    core.NewGComponent(),
		registry: make(map[string]FactoryFunc),
		width:    defaultStageWidth,
		height:   defaultStageHeight,
	}

	mgr.Register(mainMenuSceneName, func() Scene { return &MainMenu{} })
	mgr.Register("BasicsDemo", func() Scene { return NewBasicsDemo() })
	mgr.Register("TransitionDemo", func() Scene { return &TransitionDemo{} })
	mgr.Register("VirtualListDemo", func() Scene { return NewVirtualListDemo() })
	mgr.Register("LoopListDemo", func() Scene { return NewSimpleScene("LoopListDemo", "LoopList") })
	mgr.Register("HitTestDemo", func() Scene { return NewSimpleScene("HitTestDemo", "HitTest") })
	mgr.Register("PullToRefreshDemo", func() Scene { return NewSimpleScene("PullToRefreshDemo", "PullToRefresh") })
	mgr.Register("ModalWaitingDemo", func() Scene { return NewSimpleScene("ModalWaitingDemo", "ModalWaiting") })
	mgr.Register("JoystickDemo", func() Scene { return NewJoystickDemo() })
	mgr.Register("BagDemo", func() Scene { return NewSimpleScene("BagDemo", "Bag") })
	mgr.Register("ChatDemo", func() Scene { return NewSimpleScene("ChatDemo", "Chat") })
	mgr.Register("ListEffectDemo", func() Scene { return NewSimpleScene("ListEffectDemo", "ListEffect") })
	mgr.Register("ScrollPaneDemo", func() Scene { return NewSimpleScene("ScrollPaneDemo", "ScrollPane") })
	mgr.Register("TreeViewDemo", func() Scene { return NewSimpleScene("TreeViewDemo", "TreeView") })
	mgr.Register("GuideDemo", func() Scene { return NewSimpleScene("GuideDemo", "Guide") })
	mgr.Register("CooldownDemo", func() Scene { return NewSimpleScene("CooldownDemo", "Cooldown") })

	// 设置默认滚动条（在启动场景之前）
	// 从 Basics 包加载默认滚动条组件
	if err := mgr.setupDefaultScrollBars(ctx); err != nil {
		log.Printf("warning: failed to setup default scrollbars: %v", err)
	}

	if err := mgr.Start(mainMenuSceneName); err != nil {
		return nil, err
	}
	return mgr, nil
}

// setupDefaultScrollBars 设置全局默认滚动条
func (m *Manager) setupDefaultScrollBars(ctx context.Context) error {
	// 加载 Basics 包并设置默认滚动条
	// ScrollBar_VT 是垂直滚动条，ScrollBar_HZ 是水平滚动条
	return m.env.SetupDefaultScrollBars(ctx, "Basics", "ScrollBar_VT", "ScrollBar_HZ")
}

// Environment exposes shared loader/factory services to scene implementations.
func (m *Manager) Environment() *Environment {
	return m.env
}

// Stage returns the root component used for rendering (a container for the current scene).
func (m *Manager) Stage() *core.GComponent {
	return m.stage
}

// Width returns the stage width, defaulting to 800 when unspecified.
func (m *Manager) Width() int {
	if w := int(m.stage.Width()); w > 0 {
		return w
	}
	return m.width
}

// Height returns the stage height, defaulting to 600 when unspecified.
func (m *Manager) Height() int {
	if h := int(m.stage.Height()); h > 0 {
		return h
	}
	return m.height
}

// Register binds a scene factory to the given key (case-insensitive).
func (m *Manager) Register(name string, factory FactoryFunc) {
	if name == "" || factory == nil {
		return
	}
	m.registry[strings.ToLower(name)] = factory
}

// Start switches to the scene identified by name (case-insensitive).
func (m *Manager) Start(name string) error {
	factory, ok := m.registry[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("scene manager: scene %q not registered", name)
	}
	return m.switchTo(factory())
}

// Current returns the active scene (may be nil during transitions).
func (m *Manager) Current() Scene {
	return m.current
}

// CurrentComponent returns the root component of the active scene.
func (m *Manager) CurrentComponent() *core.GComponent {
	return m.root
}

func (m *Manager) switchTo(scene Scene) error {
	if scene == nil {
		return fmt.Errorf("scene manager: nil scene requested")
	}
	if m.current != nil {
		if obj := m.currentRoot(); obj != nil {
			obj.SetData(nil)
			m.stage.RemoveChild(obj)
		}
		m.current.Dispose()
		m.current = nil
		m.root = nil
	}
	component, err := scene.Load(m.ctx, m)
	if err != nil {
		return err
	}
	if component == nil {
		return fmt.Errorf("scene manager: scene %q returned nil component", scene.Name())
	}
	component.GObject.SetPosition(0, 0)
	component.GObject.SetData(component)
	m.stage.AddChild(component.GObject)
	width := component.Width()
	height := component.Height()
	if width <= 0 {
		width = float64(m.width)
	}
	if height <= 0 {
		height = float64(m.height)
	}
	m.stage.SetSize(width, height)
	m.width = int(width)
	m.height = int(height)
	if err := m.ensureCloseButton(); err != nil {
		log.Printf("[scene manager] ensure close button failed: %v", err)
	} else {
		m.repositionCloseButton()
		m.updateCloseButtonVisibility(scene)
	}
	m.current = scene
	m.root = component
	return nil
}

func (m *Manager) currentRoot() *core.GObject {
	if m.root == nil {
		return nil
	}
	return m.root.GObject
}

func (m *Manager) ensureCloseButton() error {
	if m == nil || m.stage == nil {
		return fmt.Errorf("scene manager: stage not initialized")
	}
	if m.closeBtn != nil {
		return nil
	}
	pkg, err := m.env.Package(m.ctx, "MainMenu")
	if err != nil {
		return err
	}
	item := chooseComponent(pkg, "CloseButton")
	if item == nil {
		return newMissingComponentError("MainMenu", "CloseButton")
	}
	component, err := m.env.Factory.BuildComponent(m.ctx, pkg, item)
	if err != nil {
		return err
	}
	obj := component.GObject
	obj.SetData(component)
	obj.SetTouchable(true)
	obj.SetVisible(false)
	obj.AddRelation(m.stage.GObject, core.RelationTypeRight_Right, false)
	obj.AddRelation(m.stage.GObject, core.RelationTypeBottom_Bottom, false)
	margin := closeButtonMargin
	obj.SetPosition(float64(m.width)-obj.Width()-margin, float64(m.height)-obj.Height()-margin)

	if sprite := obj.DisplayObject(); sprite != nil && sprite.Dispatcher() != nil {
		sprite.Dispatcher().On(laya.EventClick, func(laya.Event) {
			if err := m.Start(mainMenuSceneName); err != nil {
				log.Printf("[scene manager] return to %s failed: %v", mainMenuSceneName, err)
			}
		})
	} else {
		log.Printf("[scene manager] close button lacks dispatcher; click disabled")
	}

	m.stage.AddChild(obj)
	m.closeBtn = component
	return nil
}

func (m *Manager) repositionCloseButton() {
	if m.closeBtn == nil {
		return
	}
	obj := m.closeBtn.GObject
	margin := closeButtonMargin
	obj.SetPosition(float64(m.width)-obj.Width()-margin, float64(m.height)-obj.Height()-margin)
	m.stage.AddChild(obj)
}

func (m *Manager) updateCloseButtonVisibility(scene Scene) {
	if m.closeBtn == nil {
		return
	}
	obj := m.closeBtn.GObject
	isMain := scene == nil || strings.EqualFold(scene.Name(), mainMenuSceneName)
	obj.SetVisible(!isMain)
	obj.SetTouchable(!isMain)
}
