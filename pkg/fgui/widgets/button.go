package widgets

import (
	"strings"
	"sync"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
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
	// 关键：存储控制器监听器ID，以便在设置新控制器时能够取消注册
	controllerListenerID int
	titleColor         string
	titleOutlineColor  string
	titleFontSize      int
	eventOnce          sync.Once
	hovered            bool
	pressed            bool
	baseScaleX         float64
	baseScaleY         float64
}

// ComponentRoot exposes the underlying component for compatibility helpers.
func (b *GButton) ComponentRoot() *core.GComponent {
	if b == nil {
		return nil
	}
	return b.GComponent
}

// NewButton creates a button widget.
func NewButton() *GButton {
	btn := &GButton{
		GComponent:         core.NewGComponent(),
		mode:               ButtonModeCommon,
		changeStateOnClick: true,
		downEffectValue:    0.8,
		soundVolumeScale:   core.GetUIConfig().ButtonSoundVolumeScale,
		sound:              core.GetUIConfig().ButtonSound,
	}
	btn.titleColor = "#ffffff"
	btn.titleFontSize = 12
	btn.GComponent.SetData(btn)

	// 修复：按钮是交互式组件，需要拦截鼠标事件
	// 虽然TypeScript版本没有显式设置，但按钮需要能够接收点击
	btn.GComponent.GObject.SetTouchable(true)
	if sprite := btn.GComponent.GObject.DisplayObject(); sprite != nil {
		sprite.SetMouseThrough(false)
	}
	// 关键修改：不再立即绑定事件
	// 事件绑定将在 ConstructExtension 中完成
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

// SetupBeforeAdd parses button-specific properties from the component buffer.
// 对应 TypeScript 版本 GButton.setup_afterAdd 中不依赖其他对象的部分
func (b *GButton) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if b == nil || buf == nil {
		return
	}

	// 首先调用父类GComponent处理组件和基础属性
	b.GComponent.SetupBeforeAdd(buf, beginPos, nil)

	// 然后处理GButton特定属性
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(beginPos, 6) || buf.Remaining() <= 0 {
		return
	}

	// TypeScript: if (buffer.readByte() != this.packageItem.objectType) return;
	// 跳过 objectType 检查（在 builder 中已验证）
	_ = buf.ReadByte()

	// 读取不依赖其他对象的属性
	if title := buf.ReadS(); title != nil && *title != "" {
		b.SetTitle(*title)
	}
	if selectedTitle := buf.ReadS(); selectedTitle != nil && *selectedTitle != "" {
		b.SetSelectedTitle(*selectedTitle)
	}
	if icon := buf.ReadS(); icon != nil && *icon != "" {
		b.SetIcon(*icon)
	}
	if selectedIcon := buf.ReadS(); selectedIcon != nil && *selectedIcon != "" {
		b.SetSelectedIcon(*selectedIcon)
	}
	if buf.ReadBool() {
		if buf.Remaining() >= 4 {
			if color := buf.ReadColorString(true); color != "" {
				b.SetTitleColor(color)
			}
		}
	}
	if size := buf.ReadInt32(); size != 0 {
		b.SetTitleFontSize(int(size))
	}

	// relatedController 依赖父组件的 Controllers()，跳过这里的读取
	// 在 SetupAfterAdd 中处理
	_ = buf.ReadInt16() // skip controller index
	_ = buf.ReadS()     // skip relatedPageId

	// 关键修复：从 XML <Button sound=""> 属性读取音效并覆盖模板默认值
	// 对应 TypeScript 版本 GButton.setup_afterAdd 第 425-429 行
	// 如果 XML 中指定了 sound 属性（包括空字符串），则覆盖 ConstructExtension 的值
	if sound := buf.ReadS(); sound != nil {
		b.SetSound(*sound) // 包括空字符串，这会禁用音效
	}
	if buf.ReadBool() { // has soundVolumeScale
		b.SetSoundVolumeScale(float64(buf.ReadFloat32()))
	}
	_ = buf.ReadBool() // skip selected (在 ConstructExtension 中处理)
}

// SetupAfterAdd applies button configuration that depends on other objects.
// 对应 TypeScript 版本 GButton.setup_afterAdd 中依赖其他对象的部分
func (b *GButton) SetupAfterAdd(ctx *SetupContext, buf *utils.ByteBuffer) {
	if b == nil || buf == nil || ctx == nil || ctx.Child == nil {
		return
	}
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(0, 6) || buf.Remaining() <= 0 {
		return
	}

	// 跳过已在 SetupBeforeAdd 中读取的属性
	_ = buf.ReadByte()                          // objectType
	_ = buf.ReadS()                             // title
	_ = buf.ReadS()                             // selectedTitle
	_ = buf.ReadS()                             // icon
	_ = buf.ReadS()                             // selectedIcon
	if buf.ReadBool() && buf.Remaining() >= 4 { // hasColor
		_ = buf.ReadColorString(true) // titleColor
	}
	_ = buf.ReadInt32() // titleFontSize

	// 读取依赖父组件的属性
	if idx := buf.ReadInt16(); idx >= 0 && ctx.Parent != nil {
		controllers := ctx.Parent.Controllers()
		if int(idx) < len(controllers) {
			b.SetRelatedController(controllers[idx])
		}
	}
	if page := buf.ReadS(); page != nil {
		b.SetRelatedPageID(*page)
	}

	// 关键修复：禁用模板按钮的音效以避免重复播放
	// 当按钮有模板时，模板作为内部显示组件，不应该播放音效
	// 只有外层实例（n13）才播放音效
	if b.template != nil {
		if templateBtn, ok := b.template.GObject.Data().(*GButton); ok && templateBtn != nil {
			// 将模板的音效设置为空，这样模板就不会播放音效了
			templateBtn.SetSound("")
		}
	}

	// 关键修复：读取并设置 checked/selected 状态
	// 对应 TypeScript 版本 GButton.setup_afterAdd 第 431 行：this.selected = buffer.readBool();
	// 跳过 sound 相关的读取（它们在 ConstructExtension 中处理）
	_ = buf.ReadS()                             // sound (跳过，在 ConstructExtension 中处理)
	if buf.Remaining() >= 1 {
		if buf.ReadBool() { // has soundVolumeScale
			_ = buf.ReadFloat32() // 跳过 soundVolumeScale
		}
	}
	// 关键：读取并设置 selected 状态
	if buf.Remaining() >= 1 {
		selected := buf.ReadBool()
		if selected {
			b.SetSelected(true)
		}
	}

	// 修复：重新确保按钮的交互属性正确
	// SetupBeforeAdd 调用了 GComponent.SetupBeforeAdd，它会读取 opaque 并可能设置 MouseThrough=true
	// 这会覆盖按钮初始化时的 MouseThrough=false 设置，导致按钮不可点击
	// 参考 TypeScript 版本：GButton 始终可以接收点击事件
	b.GComponent.GObject.SetTouchable(true)
	if sprite := b.GComponent.GObject.DisplayObject(); sprite != nil {
		sprite.SetMouseThrough(false)
	}
}

// ConstructExtension 在组件完整构建后读取扩展属性并绑定事件
// 对应 TypeScript 版本 GButton.constructExtension()
func (b *GButton) ConstructExtension(buf *utils.ByteBuffer) error {
	if b == nil || buf == nil {
		return nil
	}

	// 读取 section 6 的按钮扩展属性
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()

	if !buf.Seek(0, 6) || buf.Remaining() <= 0 {
		return nil
	}

	// 读取按钮属性（之前在 SetupBeforeAdd 中读取）
	mode := ButtonMode(buf.ReadByte())
	b.SetMode(mode)

	// 音效相关设置
	// 关键修复：匹配 TypeScript 行为，只要 ReadS 不为 nil（包括空字符串""）就要设置
	// TS 版本（GButton.setup_afterAdd 第 425-429 行）：if (str != null) this._sound = str;
	// 包括空字符串的情况，空字符串表示"明确禁用音效"
	if sound := buf.ReadS(); sound != nil {
		b.SetSound(*sound) // 包括 sound = "" 的情况，这会覆盖全局音效设置
	}
	if buf.Remaining() >= 4 {
		b.SetSoundVolumeScale(float64(buf.ReadFloat32()))
	}

	// Down effect 设置
	if buf.Remaining() >= 1 {
		b.SetDownEffect(int(buf.ReadByte()))
	}
	if buf.Remaining() >= 4 {
		b.SetDownEffectValue(float64(buf.ReadFloat32()))
	}

	// 关键修复：初始化 downScaled 状态
	// 对应 TypeScript 版本：当 downEffect == 2 时，缩放效果会通过 setState 应用
	// 在Go版本中，我们通过 applyDownScale 函数手动应用缩放，需要正确初始化 downScaled
	if b.DownEffect() == 2 {
		b.SetPivotWithAnchor(0.5, 0.5, b.PivotAsAnchor())
		// 注意：downScaled 初始为 false，在 applyDownScale 中根据实际状态切换
	}

	// 查找 button controller（如果还没有设置）
	// 注意：controllers 已经在 BuildComponent 中添加到组件
	if b.ButtonController() == nil {
		for _, ctrl := range b.Controllers() {
			if strings.EqualFold(ctrl.Name, "button") {
				b.SetButtonController(ctrl)
				break
			}
		}
	}

	// 查找 titleObject 和 iconObject
	// 优先从 template 中查找
	if template := b.TemplateComponent(); template != nil {
		if titleObj := template.ChildByName("title"); titleObj != nil {
			b.SetTitleObject(titleObj)
		} else {
			// 如果 template 中没有，从 Button 的子组件中查找
			if titleObj := b.GComponent.ChildByName("title"); titleObj != nil {
				b.SetTitleObject(titleObj)
			}
		}
		if iconObj := template.ChildByName("icon"); iconObj != nil {
			b.SetIconObject(iconObj)
		} else {
			// 如果 template 中没有，从 Button 的子组件中查找
			if iconObj := b.GComponent.ChildByName("icon"); iconObj != nil {
				b.SetIconObject(iconObj)
			}
		}

		// 如果 buttons 没有 controller，尝试从 template 查找
		if b.ButtonController() == nil {
			if ctrl := template.ControllerByName("button"); ctrl != nil {
				b.SetButtonController(ctrl)
			} else if ctrl := template.ControllerByName("Button"); ctrl != nil {
				b.SetButtonController(ctrl)
			}
		}
	} else {
		// 没有 template，直接从 Button 的子组件中查找
		if titleObj := b.GComponent.ChildByName("title"); titleObj != nil {
			b.SetTitleObject(titleObj)
		}
		if iconObj := b.GComponent.ChildByName("icon"); iconObj != nil {
			b.SetIconObject(iconObj)
		}
	}

	// 关键：在构建完成后绑定事件
	b.bindEvents()

	return nil
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
	// 关键修复：取消注册之前控制器的监听器（如果有）
	if b.relatedController != nil && b.controllerListenerID != 0 {
		b.relatedController.RemoveSelectionListener(b.controllerListenerID)
		b.controllerListenerID = 0
	}

	b.relatedController = ctrl

	// 关键修复：注册新控制器的监听器以实现 Radio/Check 互斥
	// 当控制器状态变化时，HandleControllerChanged 会被调用，更新按钮状态
	if b.relatedController != nil {
		b.controllerListenerID = b.relatedController.AddSelectionListener(func(c *core.Controller) {
			b.HandleControllerChanged(c)
		})
	}
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

// OwnerSizeChanged 在 GButton 尺寸变化时同步更新模板组件的尺寸
// 这样模板组件内部的 Relations 系统会自动更新所有子对象（如背景图像）
func (b *GButton) OwnerSizeChanged(oldWidth, oldHeight float64) {
	if b.template != nil {
		newWidth := b.GComponent.GObject.Width()
		newHeight := b.GComponent.GObject.Height()
		b.template.GObject.SetSize(newWidth, newHeight)
	}
}

func (b *GButton) applyTitleState() {
	text := b.title
	if b.selected && b.selectedTitle != "" {
		text = b.selectedTitle
	}
	if b.titleObject == nil {
		// 修复：在没有titleObject时自动查找
		// 参考 TypeScript 版本中 titleObject 是通过 getChild("title") 查找的
		if titleChild := b.GComponent.ChildByName("title"); titleChild != nil {
			b.SetTitleObject(titleChild)
		} else if b.template != nil {
			// 如果Button有template，从template中查找
			if titleChild := b.template.ChildByName("title"); titleChild != nil {
				b.SetTitleObject(titleChild)
			}
		}
		// 如果仍然没有找到titleObject，直接返回
		if b.titleObject == nil {
			return
		}
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
		// 修复：在没有iconObject时自动查找
		// 参考 TypeScript 版本中 iconObject 是通过 getChild("icon") 查找的
		if iconChild := b.GComponent.ChildByName("icon"); iconChild != nil {
			b.SetIconObject(iconChild)
		} else if b.template != nil {
			// 如果Button有template，从template中查找
			if iconChild := b.template.ChildByName("icon"); iconChild != nil {
				b.SetIconObject(iconChild)
			}
		}
		// 如果仍然没有找到iconObject，直接返回
		if b.iconObject == nil {
			return
		}
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
		obj.On(laya.EventRollOver, func(evt *laya.Event) {
			b.onRollOver(evt)
		})
		obj.On(laya.EventRollOut, func(evt *laya.Event) {
			b.onRollOut(evt)
		})
		obj.On(laya.EventMouseDown, func(evt *laya.Event) {
			b.onMouseDown(evt)
		})
		obj.On(laya.EventMouseUp, func(evt *laya.Event) {
			b.onMouseUp(evt)
		})
		obj.On(laya.EventClick, func(evt *laya.Event) {
			b.onClick(evt)
		})
	})
}

func (b *GButton) onRollOver(evt *laya.Event) {
	b.hovered = true
	b.updateVisualState()
}

func (b *GButton) onRollOut(evt *laya.Event) {
	b.hovered = false
	if b.pressed {
		b.applyDownScale(false)
	}
	b.updateVisualState()
}

func (b *GButton) onMouseDown(evt *laya.Event) {
	if !b.GComponent.GObject.Touchable() {
		return
	}
	b.pressed = true
	// 关键修复：只有当 downEffect == 2（scale）时才应用缩放效果
	if b.downEffect == 2 {
		b.applyDownScale(true)
	}
	b.updateVisualState()
	if popup := b.linkedPopup; popup != nil {
		core.Root().TogglePopup(popup, b.GComponent.GObject, core.PopupDirectionAuto)
	}
}

func (b *GButton) onMouseUp(evt *laya.Event) {
	if !b.pressed {
		return
	}
	b.pressed = false
	// 关键修复：只有当 downEffect == 2（scale）时才应用缩放效果
	if b.downEffect == 2 {
		b.applyDownScale(false)
	}
	b.updateVisualState()
}

func (b *GButton) onClick(evt *laya.Event) {
	// 播放按钮点击音效
	if b.sound != "" {
		// 尝试从 UIPackage 获取音效文件
		if pi := assets.GetItemByURL(b.sound); pi != nil && pi.File != "" {
			core.Root().PlayOneShotSound(pi.File, b.soundVolumeScale)
		} else {
			// 如果没有找到 PackageItem，直接使用 sound 作为文件路径
			core.Root().PlayOneShotSound(b.sound, b.soundVolumeScale)
		}
	}

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
	// 关键修复：不要提前返回 downScaled
	// downScaled 只是用来记录当前是否已经缩放，不是执行条件
	// 对应 TypeScript 版本：down 时如果没缩放就缩放，up 时如果缩放就恢复
	obj := b.GComponent.GObject
	if obj == nil {
		return
	}
	factor := b.downEffectValue
	if factor <= 0 {
		factor = 1
	}
	if down {
		// 按下时：如果还没有缩放，就应用缩放
		if !b.downScaled {
			obj.SetScale(b.baseScaleX*factor, b.baseScaleY*factor)
			b.downScaled = true
		}
	} else {
		// 弹起时：如果已经缩放，就恢复
		if b.downScaled {
			obj.SetScale(b.baseScaleX, b.baseScaleY)
			b.downScaled = false
		}
	}
}

func (b *GButton) emitStateChanged(payload any) {
	b.GComponent.GObject.Emit(laya.EventStateChanged, payload)
}

// HandleControllerChanged 当相关控制器状态变化时调用
// 对应 TypeScript 版本 GButton.handleControllerChanged()
// 作用：实现 Radio/Check 按钮组的互斥逻辑
func (b *GButton) HandleControllerChanged(c *core.Controller) {
	// 调用父类的处理（如果需要）
	// GComponent 可能需要处理一些通用逻辑

	// 关键：如果变化的控制器是当前按钮的 relatedController
	if b.relatedController == c {
		// 检查控制器的当前页面是否匹配自己的页面
		// 如果匹配，则设置为选中；否则设置为未选中
		// 这实现了 Radio 按钮组的互斥：点击一个按钮会设置控制器页面，
		// 然后所有按钮都会收到通知，并更新自己的选中状态
		if c.SelectedPageID() == b.relatedPageID {
			b.SetSelected(true)
		} else {
			// 只有在非 Check 模式下才取消选中
			// Check 模式有自己的互斥逻辑（通过 SetOppositePageID）
			if b.mode != ButtonModeCheck {
				b.SetSelected(false)
			}
		}
	}
}

func (b *GButton) syncRelatedController() {
	if b.relatedController == nil || b.relatedPageID == "" {
		return
	}
	// 参考 TypeScript 原版：GButton.ts set selected()
	if b.selected {
		// 选中时：切换到 relatedPageID 指定的页面
		b.relatedController.SetSelectedPageID(b.relatedPageID)
	} else if b.mode == ButtonModeCheck && b.relatedController.SelectedPageID() == b.relatedPageID {
		// 取消选中时（Check 模式）：如果当前页面恰好是 relatedPageID，切换到相反的页面
		b.relatedController.SetOppositePageID(b.relatedPageID)
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
