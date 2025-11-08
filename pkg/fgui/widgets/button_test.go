package widgets

import (
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestButtonDefaults(t *testing.T) {
	btn := NewButton()
	if btn == nil || btn.GComponent == nil {
		t.Fatalf("expected GButton to wrap GComponent")
	}
	if btn.Mode() != ButtonModeCommon {
		t.Fatalf("unexpected default mode: %v", btn.Mode())
	}
	if !btn.ChangeStateOnClick() {
		t.Fatalf("expected changeStateOnClick to default to true")
	}
	if btn.DownEffectValue() != 0.8 {
		t.Fatalf("unexpected default down effect value: %v", btn.DownEffectValue())
	}
	if btn.SoundVolumeScale() != 1 {
		t.Fatalf("unexpected default sound volume scale: %v", btn.SoundVolumeScale())
	}
	if btn.Selected() {
		t.Fatalf("expected button to start unselected")
	}
}

func TestButtonSelectionAndMode(t *testing.T) {
	btn := NewButton()
	btn.SetMode(ButtonModeCheck)
	if btn.Mode() != ButtonModeCheck {
		t.Fatalf("expected mode to switch to check, got %v", btn.Mode())
	}
	btn.SetSelected(true)
	if !btn.Selected() {
		t.Fatalf("expected selection to stick in check mode")
	}
	btn.SetMode(ButtonModeCommon)
	if btn.Selected() {
		t.Fatalf("expected selection to reset when switching to common mode")
	}
	btn.SetSelected(true)
	if btn.Selected() {
		t.Fatalf("expected selection to remain false in common mode")
	}
}

func TestButtonMetadataAccessors(t *testing.T) {
	btn := NewButton()
	buttonCtrl := core.NewController("button")
	relatedCtrl := core.NewController("related")
	popup := core.NewGObject()

	btn.SetTitle("Title")
	btn.SetSelectedTitle("SelectedTitle")
	btn.SetIcon("ui://pkg/icon")
	btn.SetSelectedIcon("ui://pkg/icon_selected")
	btn.SetSound("sound://click")
	btn.SetSoundVolumeScale(0.5)
	btn.SetChangeStateOnClick(false)
	btn.SetButtonController(buttonCtrl)
	btn.SetRelatedController(relatedCtrl)
	btn.SetRelatedPageID("page")
	btn.SetLinkedPopup(popup)
	btn.SetDownEffect(1)
	btn.SetDownEffectValue(0.6)
	btn.SetDownScaled(true)

	if btn.Title() != "Title" {
		t.Fatalf("unexpected title: %s", btn.Title())
	}
	if btn.SelectedTitle() != "SelectedTitle" {
		t.Fatalf("unexpected selected title: %s", btn.SelectedTitle())
	}
	if btn.Icon() != "ui://pkg/icon" {
		t.Fatalf("unexpected icon: %s", btn.Icon())
	}
	if btn.SelectedIcon() != "ui://pkg/icon_selected" {
		t.Fatalf("unexpected selected icon: %s", btn.SelectedIcon())
	}
	if btn.Sound() != "sound://click" {
		t.Fatalf("unexpected sound: %s", btn.Sound())
	}
	if btn.SoundVolumeScale() != 0.5 {
		t.Fatalf("unexpected sound volume scale: %v", btn.SoundVolumeScale())
	}
	if btn.ChangeStateOnClick() {
		t.Fatalf("expected changeStateOnClick to be false after setter")
	}
	if btn.ButtonController() != buttonCtrl {
		t.Fatalf("expected button controller to persist")
	}
	if btn.RelatedController() != relatedCtrl {
		t.Fatalf("expected related controller to persist")
	}
	if btn.RelatedPageID() != "page" {
		t.Fatalf("unexpected related page id: %s", btn.RelatedPageID())
	}
	if btn.LinkedPopup() != popup {
		t.Fatalf("expected linked popup to persist")
	}
	if btn.DownEffect() != 1 {
		t.Fatalf("unexpected down effect: %d", btn.DownEffect())
	}
	if btn.DownEffectValue() != 0.6 {
		t.Fatalf("unexpected down effect value: %v", btn.DownEffectValue())
	}
	if !btn.DownScaled() {
		t.Fatalf("expected downScaled to be true after setter")
	}
}

func TestButtonTitleObjectSync(t *testing.T) {
	btn := NewButton()
	text := NewText()
	text.GObject.SetData(text)
	btn.SetTitleObject(text.GObject)
	btn.SetTitle("Hello")
	if got := text.Text(); got != "Hello" {
		t.Fatalf("expected title object to receive text, got %q", got)
	}
	btn.SetSelectedTitle("World")
	btn.SetMode(ButtonModeCheck)
	btn.SetSelected(true)
	if got := text.Text(); got != "World" {
		t.Fatalf("expected selected title to propagate, got %q", got)
	}
	btn.SetSelected(false)
	if got := text.Text(); got != "Hello" {
		t.Fatalf("expected title to revert after deselect, got %q", got)
	}
}

func TestButtonIconObjectSync(t *testing.T) {
	btn := NewButton()
	loader := NewLoader()
	loader.GObject.SetData(loader)
	btn.SetIconObject(loader.GObject)
	btn.SetIcon("ui://pkg/icon")
	if got := loader.URL(); got != "ui://pkg/icon" {
		t.Fatalf("expected icon url ui://pkg/icon, got %q", got)
	}
	btn.SetSelectedIcon("ui://pkg/icon_selected")
	btn.SetMode(ButtonModeCheck)
	btn.SetSelected(true)
	if got := loader.URL(); got != "ui://pkg/icon_selected" {
		t.Fatalf("expected selected icon to propagate, got %q", got)
	}
	btn.SetSelected(false)
	if got := loader.URL(); got != "ui://pkg/icon" {
		t.Fatalf("expected icon to revert after deselect, got %q", got)
	}
}

func TestButtonClickTogglesSelection(t *testing.T) {
	env := testutil.NewStageEnv(t, 200, 200)
	stage := env.Stage

	btn := NewButton()
	btn.SetMode(ButtonModeCheck)
	obj := btn.GComponent.GObject
	obj.SetSize(60, 30)
	obj.SetPosition(20, 20)
	stage.AddChild(obj.DisplayObject())

	env.Advance(16*time.Millisecond, laya.MouseState{X: 10, Y: 10, Primary: false})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 30, Y: 30, Primary: true})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 30, Y: 30, Primary: false})

	if !btn.Selected() {
		t.Fatalf("expected button to toggle selection on click")
	}

	btn.SetChangeStateOnClick(false)
	env.Advance(16*time.Millisecond, laya.MouseState{X: 30, Y: 30, Primary: true})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 30, Y: 30, Primary: false})

	if !btn.Selected() {
		t.Fatalf("expected selection to remain when changeStateOnClick disabled")
	}
}

// TestButtonCheckModeStateController 测试 Check 模式下按钮状态和 controller 的同步
func TestButtonCheckModeStateController(t *testing.T) {
	env := testutil.NewStageEnv(t, 200, 200)
	stage := env.Stage

	// 创建一个 Check 模式的按钮
	btn := NewButton()
	btn.SetMode(ButtonModeCheck)

	// 创建一个 button controller 模拟 Button4.xml 的配置
	// controller name="button" pages="0,up,1,down,2,over,3,selectedOver"
	buttonCtrl := core.NewController("button")
	buttonCtrl.SetPages(
		[]string{"0", "1", "2", "3"},
		[]string{"up", "down", "over", "selectedOver"},
	)
	btn.GComponent.AddController(buttonCtrl)
	btn.SetButtonController(buttonCtrl)

	obj := btn.GComponent.GObject
	obj.SetSize(131, 45)
	obj.SetPosition(50, 50)
	stage.AddChild(obj.DisplayObject())

	// 初始状态应该是 "up"
	if buttonCtrl.SelectedPageName() != "up" {
		t.Errorf("初始状态应该是 'up'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}
	if btn.Selected() {
		t.Error("初始状态按钮不应该被选中")
	}

	// 模拟第一次点击：应该切换到选中状态
	// 鼠标移到按钮上
	env.Advance(16*time.Millisecond, laya.MouseState{X: 100, Y: 70, Primary: false})
	// 鼠标按下
	env.Advance(16*time.Millisecond, laya.MouseState{X: 100, Y: 70, Primary: true})
	// 鼠标释放（点击完成）
	env.Advance(16*time.Millisecond, laya.MouseState{X: 100, Y: 70, Primary: false})

	if !btn.Selected() {
		t.Error("点击后按钮应该被选中")
	}
	// 点击后鼠标还在按钮上，所以应该是 "selectedOver" 状态
	if buttonCtrl.SelectedPageName() != "selectedOver" {
		t.Errorf("选中且鼠标在按钮上应该是 'selectedOver'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}

	// 模拟鼠标离开按钮：应该切换到 "down"
	env.Advance(16*time.Millisecond, laya.MouseState{X: 10, Y: 10, Primary: false})
	if buttonCtrl.SelectedPageName() != "down" {
		t.Errorf("选中且非悬停状态应该是 'down'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}

	// 模拟第二次点击：应该取消选中
	env.Advance(16*time.Millisecond, laya.MouseState{X: 100, Y: 70, Primary: false})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 100, Y: 70, Primary: true})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 100, Y: 70, Primary: false})

	if btn.Selected() {
		t.Error("第二次点击后按钮应该取消选中")
	}
	// 未选中且悬停状态应该是 "over"
	if buttonCtrl.SelectedPageName() != "over" {
		t.Errorf("未选中且悬停状态应该是 'over'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}

	// 模拟鼠标离开：应该回到 "up"
	env.Advance(16*time.Millisecond, laya.MouseState{X: 10, Y: 10, Primary: false})
	if buttonCtrl.SelectedPageName() != "up" {
		t.Errorf("未选中且非悬停状态应该是 'up'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}
}

// TestButtonCheckModeDirectSetSelected 测试直接调用 SetSelected 的效果
func TestButtonCheckModeDirectSetSelected(t *testing.T) {
	btn := NewButton()
	btn.SetMode(ButtonModeCheck)

	// 创建一个 button controller
	buttonCtrl := core.NewController("button")
	buttonCtrl.SetPages(
		[]string{"0", "1", "2", "3"},
		[]string{"up", "down", "over", "selectedOver"},
	)
	btn.GComponent.AddController(buttonCtrl)
	btn.SetButtonController(buttonCtrl)

	// 初始状态
	if btn.Selected() {
		t.Error("初始状态不应该被选中")
	}
	if buttonCtrl.SelectedPageName() != "up" {
		t.Errorf("初始状态应该是 'up'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}

	// 直接设置为选中
	btn.SetSelected(true)
	if !btn.Selected() {
		t.Error("SetSelected(true) 后应该被选中")
	}
	if buttonCtrl.SelectedPageName() != "down" {
		t.Errorf("SetSelected(true) 后状态应该是 'down'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}

	// 直接设置为未选中
	btn.SetSelected(false)
	if btn.Selected() {
		t.Error("SetSelected(false) 后不应该被选中")
	}
	if buttonCtrl.SelectedPageName() != "up" {
		t.Errorf("SetSelected(false) 后状态应该是 'up'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}
}

// TestButtonCheckModeWithGearXY 测试 Check 模式下 GearXY 是否正确应用
// 注意：这个测试只验证 controller 状态切换逻辑
// 实际的 GearXY 功能测试在 builder 包的集成测试中
func TestButtonCheckModeWithGearXY(t *testing.T) {
	btn := NewButton()
	btn.SetMode(ButtonModeCheck)
	btn.GComponent.GObject.SetSize(131, 45)

	// 创建一个 button controller 模拟 Button4.xml
	buttonCtrl := core.NewController("button")
	buttonCtrl.SetPages(
		[]string{"0", "1", "2", "3"},
		[]string{"up", "down", "over", "selectedOver"},
	)
	btn.GComponent.AddController(buttonCtrl)
	btn.SetButtonController(buttonCtrl)

	// 测试初始状态
	if buttonCtrl.SelectedPageName() != "up" {
		t.Errorf("初始状态应该是 'up'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}

	// 切换到选中状态
	btn.SetSelected(true)

	// 验证 controller 状态切换到 down
	if buttonCtrl.SelectedPageName() != "down" {
		t.Errorf("选中后状态应该是 'down'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}

	// 切换回未选中状态
	btn.SetSelected(false)

	// 验证 controller 状态切换回 up
	if buttonCtrl.SelectedPageName() != "up" {
		t.Errorf("取消选中后状态应该是 'up'，实际是 '%s'", buttonCtrl.SelectedPageName())
	}
}
