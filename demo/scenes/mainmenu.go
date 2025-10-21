package scenes

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// MainMenu mirrors the FairyGUI main menu demo scene.
type MainMenu struct {
	component *core.GComponent
}

func (s *MainMenu) Name() string {
	return "MainMenu"
}

func (s *MainMenu) Load(ctx context.Context, mgr *Manager) (*core.GComponent, error) {
	env := mgr.Environment()
	pkg, err := env.Package(ctx, "MainMenu")
	if err != nil {
		return nil, err
	}
	item := chooseComponent(pkg, "Main")
	if item == nil {
		return nil, newMissingComponentError("MainMenu", "Main")
	}
	component, err := env.Factory.BuildComponent(ctx, pkg, item)
	if err != nil {
		return nil, err
	}
	s.component = component
	s.attachButtons(component, mgr)
	return component, nil
}

func (s *MainMenu) Dispose() {
	s.component = nil
}

func (s *MainMenu) attachButtons(component *core.GComponent, mgr *Manager) {
	if component == nil {
		return
	}
	buttonToScene := map[string]string{
		"n1":  "BasicsDemo",
		"n2":  "TransitionDemo",
		"n4":  "VirtualListDemo",
		"n5":  "LoopListDemo",
		"n6":  "HitTestDemo",
		"n7":  "PullToRefreshDemo",
		"n8":  "ModalWaitingDemo",
		"n9":  "JoystickDemo",
		"n10": "BagDemo",
		"n11": "ChatDemo",
		"n12": "ListEffectDemo",
		"n13": "ScrollPaneDemo",
		"n14": "TreeViewDemo",
		"n15": "GuideDemo",
		"n16": "CooldownDemo",
	}
	for id, sceneName := range buttonToScene {
		child := component.ChildByName(id)
		if child == nil {
			continue
		}
		if _, ok := child.Data().(*widgets.GButton); !ok {
			continue
		}
		sprite := child.DisplayObject()
		if sprite == nil || sprite.Dispatcher() == nil {
			continue
		}
		if _, registered := mgr.registry[strings.ToLower(sceneName)]; !registered {
			log.Printf("[mainmenu] scene %s not registered; button %s click ignored", sceneName, id)
			continue
		}
		log.Printf("[mainmenu] wiring button %s -> %s", id, sceneName)
		scene := sceneName
		sprite.Dispatcher().On(laya.EventClick, func(laya.Event) {
			if err := mgr.Start(scene); err != nil {
				log.Printf("[mainmenu] start scene %s failed: %v", scene, err)
			}
		})
	}
}

// ErrMissingComponent is returned when a requested component is absent from a package.
type ErrMissingComponent struct {
	Package string
	Target  string
}

func newMissingComponentError(pkg, target string) error {
	return ErrMissingComponent{Package: pkg, Target: target}
}

func (e ErrMissingComponent) Error() string {
	return fmt.Sprintf("scene: package %s missing component %s", e.Package, e.Target)
}

func chooseComponent(pkg *assets.Package, candidates ...string) *assets.PackageItem {
	if pkg == nil {
		return nil
	}
	for _, name := range candidates {
		if name == "" {
			continue
		}
		if item := pkg.ItemByName(name); item != nil && item.Type == assets.PackageItemTypeComponent && item.Component != nil {
			return item
		}
	}
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Component != nil {
			return item
		}
	}
	return nil
}
