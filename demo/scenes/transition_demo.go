package scenes

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/gears"
	"github.com/chslink/fairygui/pkg/fgui/tween"
)

const (
	transitionPackageName = "Transition"
	defaultTransitionTime = 0.8
)

// TransitionDemo 复现 FairyGUI Transition 示例的核心交互。
type TransitionDemo struct {
	component     *core.GComponent
	stage         *core.GComponent
	targets       map[string]*core.GComponent
	buttons       []*core.GObject
	current       *core.GComponent
	finishTweener *tween.GTweener
	valueTweener  *tween.GTweener
	random        *rand.Rand
}

func (d *TransitionDemo) Name() string {
	return "TransitionDemo"
}

func (d *TransitionDemo) Load(ctx context.Context, mgr *Manager) (*core.GComponent, error) {
	env := mgr.Environment()
	pkg, err := env.Package(ctx, transitionPackageName)
	if err != nil {
		return nil, err
	}
	mainItem := chooseComponent(pkg, "Main")
	if mainItem == nil {
		return nil, newMissingComponentError(transitionPackageName, "Main")
	}
	component, err := env.Factory.BuildComponent(ctx, pkg, mainItem)
	if err != nil {
		return nil, err
	}

	d.random = rand.New(rand.NewSource(time.Now().UnixNano()))
	d.component = component
	d.stage = mgr.Stage()
	d.targets = make(map[string]*core.GComponent)

	loadTarget := func(name string, required bool) (*core.GComponent, error) {
		item := chooseComponent(pkg, name)
		if item == nil {
			if required {
				return nil, newMissingComponentError(transitionPackageName, name)
			}
			log.Printf("[transition-demo] component %s missing; skip optional target", name)
			return nil, nil
		}
		target, buildErr := env.Factory.BuildComponent(ctx, pkg, item)
		if buildErr != nil {
			return nil, buildErr
		}
		target.GObject.SetData(target)
		target.GObject.SetVisible(true)
		target.GObject.SetTouchable(true)
		d.targets[name] = target
		return target, nil
	}

	// 预加载各个转场组件，便于快速切换。
	targetSpecs := []struct {
		name     string
		required bool
	}{
		{name: "BOSS", required: true},
		{name: "BOSS_SKILL", required: true},
		{name: "TRAP", required: true},
		{name: "GoodHit", required: true},
		{name: "PowerUp", required: true},
		{name: "PathDemo", required: false},
	}
	for _, spec := range targetSpecs {
		if _, err := loadTarget(spec.name, spec.required); err != nil {
			return nil, err
		}
	}

	buttonNames := []string{"btn0", "btn1", "btn2", "btn3", "btn4", "btn5"}
	d.buttons = make([]*core.GObject, 0, len(buttonNames))
	for _, name := range buttonNames {
		if child := component.ChildByName(name); child != nil {
			d.buttons = append(d.buttons, child)
		}
	}

	bindSimple := func(buttonName, targetName string) {
		button := component.ChildByName(buttonName)
		target := d.targets[targetName]
		if button == nil || target == nil {
			return
		}
		button.SetTouchable(true)
		sprite := button.DisplayObject()
		if sprite == nil || sprite.Dispatcher() == nil {
			return
		}
		sprite.Dispatcher().On(laya.EventClick, func(evt laya.Event) {
			d.playTarget(target, nil)
		})
	}

	bindSimple("btn0", "BOSS")
	bindSimple("btn1", "BOSS_SKILL")
	bindSimple("btn2", "TRAP")
	bindSimple("btn5", "PathDemo")

	if target := d.targets["PathDemo"]; target == nil {
		if btn := component.ChildByName("btn5"); btn != nil {
			btn.SetVisible(false)
			btn.SetTouchable(false)
		}
	}

	if button := component.ChildByName("btn3"); button != nil {
		button.On(laya.EventClick, func(laya.Event) {
			d.playGoodHit()
		})
	}
	if button := component.ChildByName("btn4"); button != nil {
		button.On(laya.EventClick, func(laya.Event) {
			d.playPowerUp()
		})
	}

	return component, nil
}

func (d *TransitionDemo) Dispose() {
	d.stopCurrent(false)
	d.component = nil
	d.stage = nil
	d.targets = nil
	d.buttons = nil
	d.random = nil
}

func (d *TransitionDemo) playGoodHit() {
	target := d.targets["GoodHit"]
	if target == nil {
		return
	}
	setup := func(comp *core.GComponent) {
		if d.stage == nil {
			return
		}
		width := d.stage.Width()
		if width <= 0 {
			width = comp.Width() + 40
		}
		comp.GObject.SetPosition(width-comp.Width()-20, 100)
	}
	d.playTarget(target, setup)
}

func (d *TransitionDemo) playPowerUp() {
	target := d.targets["PowerUp"]
	if target == nil {
		return
	}
	setup := func(comp *core.GComponent) {
		if d.stage != nil {
			height := d.stage.Height()
			if height <= 0 {
				height = comp.Height() + 200
			}
			comp.GObject.SetPosition(20, height-comp.Height()-100)
		} else {
			comp.GObject.SetPosition(20, 200)
		}
		startValue := 10000
		add := 1000
		if d.random != nil {
			add = d.random.Intn(2000) + 1000
		}
		endValue := startValue + add
		setText(comp.ChildByName("value"), strconv.Itoa(startValue))
		setText(comp.ChildByName("add_value"), fmt.Sprintf("+%d", add))
		if d.valueTweener != nil {
			d.valueTweener.Kill(false)
			d.valueTweener = nil
		}
		d.valueTweener = tween.To(float64(startValue), float64(endValue), 0.3).
			SetEase(tween.EaseTypeLinear).
			OnUpdate(func(tw *tween.GTweener) {
				current := int(tw.Value().X + 0.5)
				setText(comp.ChildByName("value"), strconv.Itoa(current))
			})
		d.valueTweener.OnComplete(func(*tween.GTweener) {
			log.Printf("[transition-demo] value tween completed %d -> %d", startValue, endValue)
		})
	}
	d.playTarget(target, setup)
}

func (d *TransitionDemo) playTarget(target *core.GComponent, setup func(*core.GComponent)) {
	if target == nil || d.stage == nil {
		return
	}
	log.Printf("[transition-demo] play target %s data=%T", target.Name(), target.GObject.Data())
	d.stopCurrent(true)
	d.current = target
	if setup != nil {
		setup(target)
	}
	d.logStageChildren("before-add")
	d.stage.AddChild(target.GObject)
	d.logStageChildren("after-add")
	d.setButtonsEnabled(false)
	duration := d.playTransition(target, "t0")
	if d.finishTweener != nil {
		d.finishTweener.Kill(false)
	}
	d.finishTweener = tween.DelayedCall(duration).OnComplete(func(*tween.GTweener) {
		d.finishCurrent(target)
	})
}

func (d *TransitionDemo) playTransition(comp *core.GComponent, name string) float64 {
	if comp == nil {
		return defaultTransitionTime
	}
	tx := comp.Transition(name)
	if tx == nil {
		log.Printf("[transition-demo] transition %s missing on %s", name, comp.Name())
		return defaultTransitionTime
	}
	info := tx.Info()
	tx.Stop(false)
	tx.Play(1, -1)
	if info.TotalDuration > 0 {
		return info.TotalDuration
	}
	return defaultTransitionTime
}

func (d *TransitionDemo) finishCurrent(target *core.GComponent) {
	if target != nil && d.stage != nil {
		d.stage.RemoveChild(target.GObject)
	}
	if d.finishTweener != nil {
		d.finishTweener = nil
	}
	if d.valueTweener != nil {
		d.valueTweener = nil
	}
	d.current = nil
	d.setButtonsEnabled(true)
	log.Printf("[transition-demo] finished target")
}

func (d *TransitionDemo) stopCurrent(resetButtons bool) {
	if d.finishTweener != nil {
		d.finishTweener.Kill(false)
		d.finishTweener = nil
	}
	if d.valueTweener != nil {
		d.valueTweener.Kill(false)
		d.valueTweener = nil
	}
	if d.current != nil && d.stage != nil {
		d.stage.RemoveChild(d.current.GObject)
	}
	d.current = nil
	if resetButtons {
		d.setButtonsEnabled(true)
	}
}

func (d *TransitionDemo) setButtonsEnabled(enabled bool) {
	for _, btn := range d.buttons {
		if btn == nil {
			continue
		}
		btn.SetTouchable(enabled)
		btn.SetVisible(enabled)
	}
}

func (d *TransitionDemo) logStageChildren(tag string) {
	if d.stage == nil {
		return
	}
	children := d.stage.Children()
	log.Printf("[transition-demo] stage %s children=%d", tag, len(children))
	for i, child := range children {
		if child == nil {
			continue
		}
		log.Printf("[transition-demo] stage %s child[%d] name=%s data=%T visible=%t pos=(%.1f,%.1f) size=(%.1f,%.1f)", tag, i, child.Name(), child.Data(), child.Visible(), child.X(), child.Y(), child.Width(), child.Height())
	}
}

func setText(obj *core.GObject, text string) {
	if obj == nil {
		return
	}
	obj.SetProp(gears.ObjectPropIDText, text)
}
