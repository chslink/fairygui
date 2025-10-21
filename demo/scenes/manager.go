package scenes

import (
	"context"
	"fmt"
	"strings"

	"github.com/chslink/fairygui/pkg/fgui/core"
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
}

// NewManager instantiates the scene manager, registers built-in demos, and loads the main menu.
func NewManager(ctx context.Context, env *Environment) (*Manager, error) {
	mgr := &Manager{
		ctx:      ctx,
		env:      env,
		stage:    core.NewGComponent(),
		registry: make(map[string]FactoryFunc),
		width:    1136,
		height:   640,
	}

	mgr.Register("MainMenu", func() Scene { return &MainMenu{} })
	mgr.Register("BagDemo", func() Scene { return &BagDemo{} })

	if err := mgr.Start("MainMenu"); err != nil {
		return nil, err
	}
	return mgr, nil
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
