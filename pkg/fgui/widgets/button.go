package widgets

import (
	"sync"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// ButtonMode mirrors FairyGUI's button selection modes.
type ButtonMode int

const (
	ButtonModeCommon ButtonMode = iota
	ButtonModeCheck
	ButtonModeRadio
)

const (
	buttonStateUp               = "up"
	buttonStateDown             = "down"
	buttonStateOver             = "over"
	buttonStateDisabled         = "disabled"
	buttonStateSelectedOver     = "selectedOver"
	buttonStateSelectedDisabled = "selectedDisabled"
)

// GButton represents a minimal button widget.
type GButton struct {
	*core.GComponent
	packageItem        *assets.PackageItem
	resource           string
	title              string
	selectedTitle      string
	icon               string
	selectedIcon       string
	iconItem           *assets.PackageItem
	template           *core.GComponent
	titleObject        *core.GObject
	iconObject         *core.GObject
	mode               ButtonMode
	selected           bool
	changeStateOnClick bool
	sound              string
	soundVolumeScale   float64
	buttonController   *core.Controller
	relatedController  *core.Controller
	relatedPageID      string
	linkedPopup        *core.GObject
	downEffect         int
	downEffectValue    float64
	downScaled         bool
	titleColor         string
	titleOutlineColor  string
	titleFontSize      int
	eventOnce          sync.Once
	hovered            bool
	pressed            bool
	baseScaleX         float64
	baseScaleY         float64
}

// NewButton creates a button widget.
func NewButton() *GButton {
	btn := &GButton{
		GComponent:         core.NewGComponent(),
		mode:               ButtonModeCommon,
		changeStateOnClick: true,
		downEffectValue:    0.8,
		soundVolumeScale:   1,
	}
	btn.titleColor = "#ffffff"
	btn.titleFontSize = 12
	btn.GComponent.SetData(btn)
	btn.bindEvents()
	return btn
}

// SetPackageItem stores the underlying package item representing this button template.
func (b *GButton) SetPackageItem(item *assets.PackageItem) {
	b.packageItem = item
}

// PackageItem returns the associated package item, if any.
func (b *GButton) PackageItem() *assets.PackageItem {
	return b.packageItem
}

// SetResource captures the raw resource identifier found on the component child.
func (b *GButton) SetResource(res string) {
	b.resource = res
}

// Resource returns the stored resource identifier (usually the package item id or name).
func (b *GButton) Resource() string {
	return b.resource
}

// SetTitle records the button title.
func (b *GButton) SetTitle(value string) {
	b.title = value
	b.applyTitleState()
}

// Title returns the stored title text.
func (b *GButton) Title() string {
	return b.title
}

// SetSelectedTitle captures the alternate title used when the button is selected.
func (b *GButton) SetSelectedTitle(value string) {
	b.selectedTitle = value
	b.applyTitleState()
}

// SelectedTitle returns the stored selected title text.
func (b *GButton) SelectedTitle() string {
	return b.selectedTitle
}

// SetIcon stores the icon resource identifier.
func (b *GButton) SetIcon(value string) {
	b.icon = value
	b.applyIconState()
}

// Icon returns the icon resource identifier.
func (b *GButton) Icon() string {
	return b.icon
}

// SetSelectedIcon stores the icon resource identifier used for the selected state.
func (b *GButton) SetSelectedIcon(value string) {
	b.selectedIcon = value
	b.applyIconState()
}

// SelectedIcon returns the icon resource identifier for the selected state.
func (b *GButton) SelectedIcon() string {
	return b.selectedIcon
}

// SetIconItem stores the resolved icon package item.
func (b *GButton) SetIconItem(item *assets.PackageItem) {
	b.iconItem = item
}

// IconItem returns the resolved icon package item.
func (b *GButton) IconItem() *assets.PackageItem {
	return b.iconItem
}

// Mode returns the button mode (common/check/radio).
func (b *GButton) Mode() ButtonMode {
	return b.mode
}

// SetMode updates the button selection mode.
func (b *GButton) SetMode(value ButtonMode) {
	if b.mode == value {
		return
	}
	if value == ButtonModeCommon {
		if b.selected {
			b.selected = false
			b.applyTitleState()
			b.applyIconState()
			b.syncRelatedController()
			b.updateVisualState()
		}
	}
	b.mode = value
}

// Selected returns whether the button is currently selected.
func (b *GButton) Selected() bool {
	return b.selected
}

// SetSelected updates the selection state respecting the current mode.
func (b *GButton) SetSelected(value bool) {
	if b.mode == ButtonModeCommon {
		return
	}
	if b.selected == value {
		return
	}
	b.selected = value
	b.applyTitleState()
	b.applyIconState()
	b.syncRelatedController()
	b.updateVisualState()
}

// ChangeStateOnClick indicates if the button toggles state on click.
func (b *GButton) ChangeStateOnClick() bool {
	return b.changeStateOnClick
}

// SetChangeStateOnClick updates whether the button should toggle on click.
func (b *GButton) SetChangeStateOnClick(value bool) {
	b.changeStateOnClick = value
}

// Sound returns the sound resource identifier used when the button is clicked.
func (b *GButton) Sound() string {
	return b.sound
}

// SetSound updates the sound resource identifier.
func (b *GButton) SetSound(value string) {
	b.sound = value
}

// SoundVolumeScale returns the playback volume scale.
func (b *GButton) SoundVolumeScale() float64 {
	return b.soundVolumeScale
}

// SetSoundVolumeScale updates the playback volume scale.
func (b *GButton) SetSoundVolumeScale(value float64) {
	b.soundVolumeScale = value
}

// ButtonController returns the controller responsible for button states.
func (b *GButton) ButtonController() *core.Controller {
	return b.buttonController
}

// SetButtonController assigns the controller responsible for button states.
func (b *GButton) SetButtonController(ctrl *core.Controller) {
	b.buttonController = ctrl
	b.updateVisualState()
}

// RelatedController returns the associated controller used for radio/check behaviour.
func (b *GButton) RelatedController() *core.Controller {
	return b.relatedController
}

// SetRelatedController assigns the controller used for radio/check behaviour.
func (b *GButton) SetRelatedController(ctrl *core.Controller) {
	b.relatedController = ctrl
}

// RelatedPageID returns the controller page that this button toggles.
func (b *GButton) RelatedPageID() string {
	return b.relatedPageID
}

// SetRelatedPageID updates the controller page that this button toggles.
func (b *GButton) SetRelatedPageID(value string) {
	b.relatedPageID = value
}

// LinkedPopup returns the popup associated with the button.
func (b *GButton) LinkedPopup() *core.GObject {
	return b.linkedPopup
}

// SetLinkedPopup assigns the popup associated with the button.
func (b *GButton) SetLinkedPopup(obj *core.GObject) {
	b.linkedPopup = obj
}

// DownEffect returns the configured down effect.
func (b *GButton) DownEffect() int {
	return b.downEffect
}

// SetDownEffect updates the down effect.
func (b *GButton) SetDownEffect(value int) {
	b.downEffect = value
}

// DownEffectValue returns the down effect intensity.
func (b *GButton) DownEffectValue() float64 {
	return b.downEffectValue
}

// SetDownEffectValue updates the down effect intensity.
func (b *GButton) SetDownEffectValue(value float64) {
	b.downEffectValue = value
}

// DownScaled returns whether the button should apply down scaling.
func (b *GButton) DownScaled() bool {
	return b.downScaled
}

// SetDownScaled toggles the down scaling effect.
func (b *GButton) SetDownScaled(value bool) {
	b.downScaled = value
}

// TitleColor returns the stored title colour.
func (b *GButton) TitleColor() string {
	return b.titleColor
}

// SetTitleColor updates the stored title colour.
func (b *GButton) SetTitleColor(value string) {
	b.titleColor = value
}

// TitleOutlineColor returns the outline colour used for the title text.
func (b *GButton) TitleOutlineColor() string {
	return b.titleOutlineColor
}

// SetTitleOutlineColor updates the stored outline colour.
func (b *GButton) SetTitleOutlineColor(value string) {
	b.titleOutlineColor = value
}

// TitleFontSize returns the stored font size for the title text.
func (b *GButton) TitleFontSize() int {
	return b.titleFontSize
}

// SetTitleFontSize updates the stored font size for the title text.
func (b *GButton) SetTitleFontSize(size int) {
	b.titleFontSize = size
}

// TemplateComponent returns the instantiated template component (if any).
func (b *GButton) TemplateComponent() *core.GComponent {
	return b.template
}

// SetTemplateComponent stores the template component instance used by this button.
func (b *GButton) SetTemplateComponent(comp *core.GComponent) {
	if b.template != nil && b.GComponent != nil {
		b.GComponent.RemoveChild(b.template.GObject)
	}
	b.template = comp
	if comp != nil && b.GComponent != nil {
		comp.GObject.SetPosition(0, 0)
		b.GComponent.AddChild(comp.GObject)
	}
	b.updateVisualState()
}

// SetTitleObject caches the underlying title display object reference.
func (b *GButton) SetTitleObject(obj *core.GObject) {
	b.titleObject = obj
	b.applyTitleState()
}

// TitleObject returns the cached title display object.
func (b *GButton) TitleObject() *core.GObject {
	return b.titleObject
}

// SetIconObject caches the underlying icon display object reference.
func (b *GButton) SetIconObject(obj *core.GObject) {
	b.iconObject = obj
	b.applyIconState()
}

// IconObject returns the cached icon display object.
func (b *GButton) IconObject() *core.GObject {
	return b.iconObject
}

// UpdateTemplateBounds resizes the template component to match the button.
func (b *GButton) UpdateTemplateBounds(width, height float64) {
	if b.template == nil {
		return
	}
	b.template.GObject.SetSize(width, height)
}

func (b *GButton) applyTitleState() {
	text := b.title
	if b.selected && b.selectedTitle != "" {
		text = b.selectedTitle
	}
	if b.titleObject == nil {
		return
	}
	switch data := b.titleObject.Data().(type) {
	case *GTextField:
		data.SetText(text)
	case *GLabel:
		data.SetTitle(text)
	case *GButton:
		data.SetTitle(text)
	case string:
		if data != text {
			b.titleObject.SetData(text)
		}
	case nil:
		b.titleObject.SetData(text)
	default:
		// best-effort: try to set data string for unknown types
		b.titleObject.SetData(text)
	}
}

func (b *GButton) applyIconState() {
	icon := b.icon
	if b.selected && b.selectedIcon != "" {
		icon = b.selectedIcon
	}
	if b.iconObject == nil {
		return
	}
	switch data := b.iconObject.Data().(type) {
	case *GLoader:
		data.SetURL(icon)
	case *GButton:
		data.SetIcon(icon)
	case string:
		if data != icon {
			b.iconObject.SetData(icon)
		}
	case nil:
		b.iconObject.SetData(icon)
	default:
		b.iconObject.SetData(icon)
	}
}

func (b *GButton) bindEvents() {
	b.eventOnce.Do(func() {
		obj := b.GComponent.GObject
		if obj == nil {
			return
		}
		b.baseScaleX, b.baseScaleY = obj.Scale()
		obj.On(laya.EventRollOver, func(evt laya.Event) {
			b.onRollOver(evt)
		})
		obj.On(laya.EventRollOut, func(evt laya.Event) {
			b.onRollOut(evt)
		})
		obj.On(laya.EventMouseDown, func(evt laya.Event) {
			b.onMouseDown(evt)
		})
		obj.On(laya.EventMouseUp, func(evt laya.Event) {
			b.onMouseUp(evt)
		})
		obj.On(laya.EventClick, func(evt laya.Event) {
			b.onClick(evt)
		})
	})
}

func (b *GButton) onRollOver(evt laya.Event) {
	b.hovered = true
	b.updateVisualState()
}

func (b *GButton) onRollOut(evt laya.Event) {
	b.hovered = false
	if b.pressed {
		b.applyDownScale(false)
	}
	b.updateVisualState()
}

func (b *GButton) onMouseDown(evt laya.Event) {
	if !b.GComponent.GObject.Touchable() {
		return
	}
	b.pressed = true
	b.applyDownScale(true)
	b.updateVisualState()
	if popup := b.linkedPopup; popup != nil {
		core.Root().TogglePopup(popup, b.GComponent.GObject, core.PopupDirectionAuto)
	}
}

func (b *GButton) onMouseUp(evt laya.Event) {
	if !b.pressed {
		return
	}
	b.pressed = false
	b.applyDownScale(false)
	b.updateVisualState()
}

func (b *GButton) onClick(evt laya.Event) {
	if !b.changeStateOnClick {
		return
	}
	prev := b.selected
	switch b.mode {
	case ButtonModeCheck:
		b.SetSelected(!b.selected)
	case ButtonModeRadio:
		if !b.selected {
			b.SetSelected(true)
		}
	}
	if prev != b.selected {
		b.emitStateChanged(evt.Data)
	}
}

func (b *GButton) applyDownScale(down bool) {
	if !b.downScaled {
		return
	}
	obj := b.GComponent.GObject
	if obj == nil {
		return
	}
	factor := b.downEffectValue
	if factor <= 0 {
		factor = 1
	}
	if down {
		obj.SetScale(b.baseScaleX*factor, b.baseScaleY*factor)
	} else {
		obj.SetScale(b.baseScaleX, b.baseScaleY)
	}
}

func (b *GButton) emitStateChanged(payload any) {
	b.GComponent.GObject.Emit(laya.EventStateChanged, payload)
}

func (b *GButton) syncRelatedController() {
	if b.relatedController == nil || b.relatedPageID == "" {
		return
	}
	if b.selected {
		b.relatedController.SetSelectedPageID(b.relatedPageID)
	}
}

func (b *GButton) updateVisualState() {
	state := b.determineState()
	b.applyState(state)
}

func (b *GButton) determineState() string {
	obj := b.GComponent.GObject
	if obj == nil {
		return buttonStateUp
	}
	disabled := obj.Grayed() || !obj.Touchable()
	if b.mode == ButtonModeCommon {
		if disabled && b.hasState(buttonStateDisabled) {
			return buttonStateDisabled
		}
		if b.pressed && b.hasState(buttonStateDown) {
			return buttonStateDown
		}
		if b.hovered && b.hasState(buttonStateOver) {
			return buttonStateOver
		}
		return buttonStateUp
	}

	if disabled {
		if b.selected {
			if b.hasState(buttonStateSelectedDisabled) {
				return buttonStateSelectedDisabled
			}
		}
		if b.hasState(buttonStateDisabled) {
			return buttonStateDisabled
		}
	}

	if b.selected {
		if b.pressed && b.hasState(buttonStateDown) {
			return buttonStateDown
		}
		if b.hovered && b.hasState(buttonStateSelectedOver) {
			return buttonStateSelectedOver
		}
		if b.hasState(buttonStateDown) {
			return buttonStateDown
		}
		return buttonStateUp
	}
	if b.pressed && b.hasState(buttonStateDown) {
		return buttonStateDown
	}
	if b.hovered && b.hasState(buttonStateOver) {
		return buttonStateOver
	}
	return buttonStateUp
}

func (b *GButton) hasState(name string) bool {
	if name == "" {
		return false
	}
	if ctrl := b.buttonController; ctrl != nil {
		for _, page := range ctrl.PageNames {
			if page == name {
				return true
			}
		}
	}
	return false
}

func (b *GButton) applyState(name string) {
	if ctrl := b.buttonController; ctrl != nil {
		if b.hasState(name) {
			ctrl.SetSelectedPageName(name)
		}
	}
}
