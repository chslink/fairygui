package widgets

import (
	"context"
	"strings"
	"sync"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// PopupDirection mirrors FairyGUI 下拉方向枚举。
type PopupDirection int

const (
	PopupDirectionAuto PopupDirection = iota
	PopupDirectionUp
	PopupDirectionDown
)

const defaultComboBoxVisibleItemCount = 10

// GComboBox 实现 FairyGUI ComboBox 控件的核心字段与初始化流程。
type GComboBox struct {
	*core.GComponent
	packageItem         *assets.PackageItem
	resource            string
	template            *core.GComponent
	titleObject         *core.GObject
	iconObject          *core.GObject
	buttonController    *core.Controller
	dropdown            *core.GComponent
	dropdownURL         string
	dropdownItem        *assets.PackageItem
	list                *GList
	items               []string
	values              []string
	icons               []string
	selectedIndex       int
	visibleItemCount    int
	popupDirection      PopupDirection
	text                string
	icon                string
	titleColor          string
	titleFontSize       int
	titleOutlineColor   string
	selectionController *core.Controller
	down               bool
	over               bool
	itemsUpdated       bool
	eventOnce          sync.Once

	// factory 用于在 ConstructExtension 中创建 dropdown（模仿 TypeScript 模式）
	factory             interface {
		BuildComponent(ctx context.Context, pkg *assets.Package, item *assets.PackageItem) (*core.GComponent, error)
	}
}

// ComponentRoot 返回组合控件根节点。
func (c *GComboBox) ComponentRoot() *core.GComponent {
	if c == nil {
		return nil
	}
	return c.GComponent
}

// SetFactoryInternal 设置内部factory（仅供builder使用）
// 采用public方法避免反射，符合Go最佳实践
func (c *GComboBox) SetFactoryInternal(factory interface {
	BuildComponent(ctx context.Context, pkg *assets.Package, item *assets.PackageItem) (*core.GComponent, error)
}) {
	if c != nil {
		c.factory = factory
	}
}

// setFactoryInternal 内部方法保留（兼容旧代码）
func (c *GComboBox) setFactoryInternal(factory interface {
	BuildComponent(ctx context.Context, pkg *assets.Package, item *assets.PackageItem) (*core.GComponent, error)
}) {
	if c != nil {
		c.factory = factory
	}
}

// NewComboBox 创建一个空的 ComboBox。
func NewComboBox() *GComboBox {
	comp := core.NewGComponent()
	cb := &GComboBox{
		GComponent:       comp,
		selectedIndex:    -1,
		visibleItemCount: defaultComboBoxVisibleItemCount,
		popupDirection:   PopupDirectionAuto,
		itemsUpdated:     true,
	}
	comp.SetData(cb)
	return cb
}

// SetPackageItem 记录模板包内资源。
func (c *GComboBox) SetPackageItem(item *assets.PackageItem) {
	c.packageItem = item
}

// PackageItem 返回模板资源。
func (c *GComboBox) PackageItem() *assets.PackageItem {
	return c.packageItem
}

// SetResource 记录 child.Data/Src。
func (c *GComboBox) SetResource(res string) {
	c.resource = res
}

// Resource 返回 child.Data/Src。
func (c *GComboBox) Resource() string {
	return c.resource
}

// SetTemplateComponent 安装模板组件。
func (c *GComboBox) SetTemplateComponent(comp *core.GComponent) {
	if c.template != nil && c.GComponent != nil {
		c.GComponent.RemoveChild(c.template.GObject)
	}
	c.template = comp
	if comp != nil && c.GComponent != nil {
		comp.GObject.SetPosition(0, 0)
		c.GComponent.AddChild(comp.GObject)
		if ctrl := comp.ControllerByName("button"); ctrl != nil {
			c.SetButtonController(ctrl)
		} else if ctrl := comp.ControllerByName("Button"); ctrl != nil {
			c.SetButtonController(ctrl)
		}
		if title := comp.ChildByName("title"); title != nil {
			c.SetTitleObject(title)
		}
		if icon := comp.ChildByName("icon"); icon != nil {
			c.SetIconObject(icon)
		}
	}
}

// TemplateComponent 返回模板组件。
func (c *GComboBox) TemplateComponent() *core.GComponent {
	return c.template
}

// SetTitleObject 缓存标题对象引用。
func (c *GComboBox) SetTitleObject(obj *core.GObject) {
	c.titleObject = obj
	c.applyTitleState()
	c.applyTitleFormatting()
}

// TitleObject 返回标题对象。
func (c *GComboBox) TitleObject() *core.GObject {
	return c.titleObject
}

// SetIconObject 缓存图标对象引用。
func (c *GComboBox) SetIconObject(obj *core.GObject) {
	c.iconObject = obj
	c.applyIconState()
}

// IconObject 返回图标对象。
func (c *GComboBox) IconObject() *core.GObject {
	return c.iconObject
}

// SetButtonController 记录按钮状态控制器。
func (c *GComboBox) SetButtonController(ctrl *core.Controller) {
	c.buttonController = ctrl
}

// ButtonController 返回按钮控制器。
func (c *GComboBox) ButtonController() *core.Controller {
	return c.buttonController
}

// SetDropdownComponent 设置下拉组件。
func (c *GComboBox) SetDropdownComponent(comp *core.GComponent) {
	c.dropdown = comp
	if comp != nil && comp.GObject != nil && comp.GObject.Name() == "" {
		comp.GObject.SetName("dropdown")
	}
}

// Dropdown 返回下拉组件。
func (c *GComboBox) Dropdown() *core.GComponent {
	return c.dropdown
}

// SetDropdownURL 记录下拉资源 URL。
func (c *GComboBox) SetDropdownURL(url string) {
	c.dropdownURL = url
}

// DropdownURL 返回下拉 URL。
func (c *GComboBox) DropdownURL() string {
	return c.dropdownURL
}

// SetDropdownItem 缓存下拉包资源。
func (c *GComboBox) SetDropdownItem(item *assets.PackageItem) {
	c.dropdownItem = item
}

// DropdownItem 返回下拉包资源。
func (c *GComboBox) DropdownItem() *assets.PackageItem {
	return c.dropdownItem
}

// SetList 缓存弹出列表。
func (c *GComboBox) SetList(list *GList) {
	c.list = list
}

// List 返回弹出列表。
func (c *GComboBox) List() *GList {
	return c.list
}

// getObjectCreator 返回对象创建器
func (c *GComboBox) getObjectCreator() ObjectCreator {
	if c == nil || c.list != nil {
		if c.list != nil {
			return c.list.GetObjectCreator()
		}
	}
	// 如果没有list，尝试从GComponent中查找
	if c.GComponent != nil {
		for _, child := range c.GComponent.Children() {
			if child != nil {
				if list, ok := child.Data().(*GList); ok && list != nil {
					return list.GetObjectCreator()
				}
			}
		}
	}
	return nil
}

// Items 返回当前条目副本。
func (c *GComboBox) Items() []string {
	return append([]string(nil), c.items...)
}

// NumItems 返回条目数量。
func (c *GComboBox) NumItems() int {
	return len(c.items)
}

// Values 返回条目 value 副本。
func (c *GComboBox) Values() []string {
	return append([]string(nil), c.values...)
}

// Icons 返回条目图标副本。
func (c *GComboBox) Icons() []string {
	return append([]string(nil), c.icons...)
}

// SetItems 替换条目内容。
func (c *GComboBox) SetItems(items, values, icons []string) {
	c.items = append([]string(nil), items...)
	c.values = append([]string(nil), values...)
	if len(icons) > 0 {
		c.icons = append([]string(nil), icons...)
	} else {
		c.icons = nil
	}
	c.itemsUpdated = true
	if c.selectedIndex >= len(c.items) {
		c.SetSelectedIndex(-1)
	} else {
		c.applySelectionState()
	}
}

// SetVisibleItemCount 设置可视条目数。
func (c *GComboBox) SetVisibleItemCount(count int) {
	if count <= 0 {
		count = defaultComboBoxVisibleItemCount
	}
	c.visibleItemCount = count
}

// VisibleItemCount 返回可视条目数量。
func (c *GComboBox) VisibleItemCount() int {
	return c.visibleItemCount
}

// SetPopupDirection 更新下拉方向。
func (c *GComboBox) SetPopupDirection(dir PopupDirection) {
	if dir < PopupDirectionAuto || dir > PopupDirectionDown {
		dir = PopupDirectionAuto
	}
	c.popupDirection = dir
}

// PopupDirection 返回下拉方向。
func (c *GComboBox) PopupDirection() PopupDirection {
	return c.popupDirection
}

// SetSelectionController 绑定同步控制器。
func (c *GComboBox) SetSelectionController(ctrl *core.Controller) {
	c.selectionController = ctrl
}

// SelectionController 返回绑定控制器。
func (c *GComboBox) SelectionController() *core.Controller {
	return c.selectionController
}

// SetTitleColor 设置标题颜色。
func (c *GComboBox) SetTitleColor(color string) {
	color = strings.TrimSpace(color)
	if color == "" {
		c.titleColor = ""
	} else {
		c.titleColor = color
	}
	c.applyTitleFormatting()
}

// TitleColor 返回标题颜色。
func (c *GComboBox) TitleColor() string {
	return c.titleColor
}

// SetTitleOutlineColor 设置标题描边颜色。
func (c *GComboBox) SetTitleOutlineColor(color string) {
	c.titleOutlineColor = strings.TrimSpace(color)
	c.applyTitleFormatting()
}

// TitleOutlineColor 返回标题描边颜色。
func (c *GComboBox) TitleOutlineColor() string {
	return c.titleOutlineColor
}

// SetTitleFontSize 设置标题字号。
func (c *GComboBox) SetTitleFontSize(size int) {
	c.titleFontSize = size
	c.applyTitleFormatting()
}

// TitleFontSize 返回标题字号。
func (c *GComboBox) TitleFontSize() int {
	return c.titleFontSize
}

// SetText 更新文本。
func (c *GComboBox) SetText(text string) {
	c.text = text
	c.applyTitleState()
}

// Text 返回当前文本。
func (c *GComboBox) Text() string {
	return c.text
}

// SetIcon 更新图标 URL。
func (c *GComboBox) SetIcon(icon string) {
	c.icon = icon
	c.applyIconState()
}

// Icon 返回当前图标。
func (c *GComboBox) Icon() string {
	return c.icon
}

// SelectedIndex 返回当前选中索引。
func (c *GComboBox) SelectedIndex() int {
	return c.selectedIndex
}

// SetSelectedIndex 更新选中索引并刷新显示。
func (c *GComboBox) SetSelectedIndex(index int) {
	if len(c.items) == 0 {
		index = -1
	}
	if index < 0 || index >= len(c.items) {
		c.selectedIndex = -1
		c.applySelectionState()
		return
	}
	if c.selectedIndex == index && c.text == c.items[index] {
		return
	}
	c.selectedIndex = index
	c.applySelectionState()
}

// Value 返回当前选中 value。
func (c *GComboBox) Value() string {
	if c.selectedIndex >= 0 && c.selectedIndex < len(c.values) {
		return c.values[c.selectedIndex]
	}
	return ""
}

// SetupAfterAdd 解析 ComboBox setup_afterAdd。
func (c *GComboBox) SetupAfterAdd(ctx *SetupContext, buf *utils.ByteBuffer) {
	if c == nil || buf == nil || ctx == nil || ctx.Child == nil {
		return
	}
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(0, 6) || buf.Remaining() <= 0 {
		return
	}
	objType := assets.ObjectType(buf.ReadByte())
	childType := ctx.Child.Type
	if objType != childType {
		if ctx.ResolvedItem != nil && objType == ctx.ResolvedItem.ObjectType {
			// allow: component extension referencing specialised template
		} else if childType != assets.ObjectTypeComponent {
			return
		}
	}

	itemCount := int(buf.ReadInt16())
	if itemCount < 0 {
		itemCount = 0
	}
	items := make([]string, itemCount)
	values := make([]string, itemCount)
	icons := make([]string, itemCount)
	haveIcons := false
	for i := 0; i < itemCount; i++ {
		if buf.Remaining() < 2 {
			break
		}
		nextPos := int(buf.ReadInt16())
		nextPos += buf.Pos()

		if s := buf.ReadS(); s != nil {
			items[i] = *s
		}
		if v := buf.ReadS(); v != nil {
			values[i] = *v
		}
		if icon := buf.ReadS(); icon != nil && *icon != "" {
			icons[i] = *icon
			haveIcons = true
		}

		if nextPos >= 0 && nextPos <= buf.Len() {
			_ = buf.SetPos(nextPos)
		} else {
			break
		}
	}
	if !haveIcons {
		icons = nil
	}
	c.items = items
	c.values = values
	c.icons = icons
	c.itemsUpdated = true

	if text := buf.ReadS(); text != nil {
		c.SetText(*text)
		c.selectedIndex = indexOfString(items, *text)
	} else if len(items) > 0 {
		c.selectedIndex = 0
		c.applySelectionState()
	} else {
		c.selectedIndex = -1
		c.applySelectionState()
	}

	if icon := buf.ReadS(); icon != nil {
		c.SetIcon(*icon)
	} else if c.selectedIndex >= 0 {
		c.applySelectionState()
	}

	if buf.Remaining() > 0 && buf.ReadBool() && buf.Remaining() >= 4 {
		color := buf.ReadColorString(true)
		if color != "" {
			c.SetTitleColor(color)
		}
	}
	if buf.Remaining() >= 4 {
		count := int(buf.ReadInt32())
		if count > 0 {
			c.SetVisibleItemCount(count)
		}
	}
	if buf.Remaining() > 0 {
		dir := PopupDirection(buf.ReadByte())
		c.SetPopupDirection(dir)
	}
	if buf.Remaining() >= 2 {
		index := int(buf.ReadInt16())
		if index >= 0 && ctx.Parent != nil {
			ctrl := ctx.Parent.ControllerAt(index)
			if ctrl != nil {
				c.SetSelectionController(ctrl)
			}
		}
	}

	c.applySelectionState()
}

func (c *GComboBox) applySelectionState() {
	if c == nil {
		return
	}
	if c.selectedIndex >= 0 && c.selectedIndex < len(c.items) {
		c.text = c.items[c.selectedIndex]
		if c.selectedIndex < len(c.icons) && c.icons != nil {
			c.icon = c.icons[c.selectedIndex]
		}
	} else {
		c.text = ""
		if len(c.icons) > 0 {
			c.icon = ""
		}
	}
	c.applyTitleState()
	c.applyIconState()
}

func (c *GComboBox) applyTitleState() {
	if c == nil || c.titleObject == nil {
		return
	}
	switch data := c.titleObject.Data().(type) {
	case *GTextField:
		data.SetText(c.text)
	case *GLabel:
		data.SetTitle(c.text)
	case *GButton:
		data.SetTitle(c.text)
	case string:
		if data != c.text {
			c.titleObject.SetData(c.text)
		}
	case nil:
		c.titleObject.SetData(c.text)
	default:
		c.titleObject.SetData(c.text)
	}
}

func (c *GComboBox) applyIconState() {
	if c == nil || c.iconObject == nil {
		return
	}
	icon := c.icon
	switch data := c.iconObject.Data().(type) {
	case *GLoader:
		data.SetURL(icon)
	case *GLabel:
		data.SetIcon(icon)
	case *GButton:
		data.SetIcon(icon)
	case string:
		if data != icon {
			c.iconObject.SetData(icon)
		}
	case nil:
		c.iconObject.SetData(icon)
	default:
		c.iconObject.SetData(icon)
	}
}

func (c *GComboBox) applyTitleFormatting() {
	if c == nil || c.titleObject == nil {
		return
	}
	switch data := c.titleObject.Data().(type) {
	case *GTextField:
		if c.titleColor != "" {
			data.SetColor(c.titleColor)
		}
		if c.titleFontSize > 0 {
			data.SetFontSize(c.titleFontSize)
		}
	case *GLabel:
		if c.titleColor != "" {
			data.SetTitleColor(c.titleColor)
		}
		if c.titleFontSize > 0 {
			data.SetTitleFontSize(c.titleFontSize)
		}
	case *GButton:
		if c.titleColor != "" {
			data.SetTitleColor(c.titleColor)
		}
		if c.titleFontSize > 0 {
			data.SetTitleFontSize(c.titleFontSize)
		}
	}
}

func indexOfString(entries []string, target string) int {
	for i, entry := range entries {
		if entry == target {
			return i
		}
	}
	return -1
}

// ConstructExtension 在组件完整构建后绑定事件监听器
// 对应 TypeScript 版本 GComboBox.constructExtension()
func (c *GComboBox) ConstructExtension(buf *utils.ByteBuffer) error {
	if c == nil || buf == nil {
		return nil
	}

	// 保存当前位置，函数结束时恢复
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()

	// 查找button controller和title/icon对象
	c.buttonController = c.GComponent.ControllerByName("button")
	if c.titleObject == nil {
		c.titleObject = c.GComponent.ChildByName("title")
	}
	if c.iconObject == nil {
		c.iconObject = c.GComponent.ChildByName("icon")
	}

	// 关键修复：像TypeScript版本一样，在constructExtension中创建dropdown
	// TypeScript版本在第295-317行从buffer读取dropdown URL并创建组件
	if !buf.Seek(0, 6) {
	} else {
		dropdownURL := buf.ReadS()
		if dropdownURL != nil && *dropdownURL != "" {
			// 优先使用factory创建（TypeScript模式）
			if c.factory != nil {
				if item := assets.GetItemByURL(*dropdownURL); item != nil {
					targetPkg := item.Owner
					if dropdownComp, err := c.factory.BuildComponent(context.Background(), targetPkg, item); err == nil && dropdownComp != nil {
						c.SetDropdownComponent(dropdownComp)

						// 从dropdown中查找list
						if listObj := dropdownComp.ChildByName("list"); listObj != nil {
							if list, ok := listObj.Data().(*GList); ok {
								c.SetList(list)

								// TypeScript版本第308行：给list添加点击事件监听
								// 监听list的StateChanged事件（当选择改变时触发）
								list.GComponent.GObject.On(laya.EventStateChanged, func(evt *laya.Event) {
									c.onListItemClick(evt)
								})
							}
						}

						// TypeScript版本第310-314行：设置dropdown和list之间的关系
						if c.list != nil {
							// list宽度跟随dropdown
							c.list.GComponent.GObject.AddRelation(dropdownComp.GObject, core.RelationTypeWidth, false)
							// list高度不跟随dropdown
							c.list.GComponent.GObject.RemoveRelation(dropdownComp.GObject, core.RelationTypeHeight)

							// dropdown高度跟随list
							dropdownComp.GObject.AddRelation(c.list.GComponent.GObject, core.RelationTypeHeight, false)
							// dropdown宽度不跟随list
							dropdownComp.GObject.RemoveRelation(c.list.GComponent.GObject, core.RelationTypeWidth)
						}

						// TypeScript版本第316行：监听dropdown的UNDISPLAY事件
						if disp := dropdownComp.DisplayObject(); disp != nil {
							disp.Dispatcher().On(laya.EventUndisplay, func(evt *laya.Event) {
								c.onPopupWinClosed(evt)
							})
						}
					}
				}
			} else {
				// 备用方案：保存信息，延迟创建
				if item := assets.GetItemByURL(*dropdownURL); item != nil {
					c.SetDropdownItem(item)
					c.SetDropdownURL(*dropdownURL)
				}
			}
		}
	}

	// 绑定事件监听器（使用sync.Once确保只绑定一次）
	c.bindEvents()

	return nil
}

func (c *GComboBox) bindEvents() {
	c.eventOnce.Do(func() {
		obj := c.GComponent.GObject
		if obj == nil {
			return
		}

		// TypeScript版本第319-321行：绑定自身事件
		obj.On(laya.EventRollOver, func(evt *laya.Event) {
			c.onRollOver(evt)
		})
		obj.On(laya.EventRollOut, func(evt *laya.Event) {
			c.onRollOut(evt)
		})
		obj.On(laya.EventMouseDown, func(evt *laya.Event) {
			c.onMouseDown(evt)
		})
	})
}

func (c *GComboBox) onRollOver(evt *laya.Event) {
	c.over = true
	if c.down || (c.dropdown != nil && c.dropdown.Parent() != nil) {
		return
	}
	c.setState(buttonStateOver)
}

func (c *GComboBox) onRollOut(evt *laya.Event) {
	c.over = false
	if c.down || (c.dropdown != nil && c.dropdown.Parent() != nil) {
		return
	}
	c.setState(buttonStateUp)
}

func (c *GComboBox) onMouseDown(evt *laya.Event) {
	if evt == nil {
		return
	}

	c.down = true

	// TypeScript版本第452行：调用GRoot.checkPopups关闭其他popup
	if root := core.Root(); root != nil && c.GComponent != nil && c.GComponent.GObject != nil {
		root.CheckPopups(c.GComponent.GObject.DisplayObject())
	}

	// TypeScript版本第456-457行：显示dropdown
	if c.dropdown != nil {
		c.showDropdown()
	}
}

func (c *GComboBox) setState(state string) {
	if c.buttonController != nil {
		// 检查状态是否存在
		for _, page := range c.buttonController.PageNames {
			if page == state {
				c.buttonController.SetSelectedPageName(state)
				break
			}
		}
	}
}

// onPopupWinClosed 处理dropdown关闭事件
// 对应TypeScript版本第411-416行
func (c *GComboBox) onPopupWinClosed(evt *laya.Event) {
	if c.over {
		c.setState(buttonStateOver)
	} else {
		c.setState(buttonStateUp)
	}
}

// onListItemClick 处理列表项点击事件（通过StateChanged事件）
// 对应TypeScript版本第418-429行的CLICK_ITEM处理逻辑
func (c *GComboBox) onListItemClick(evt *laya.Event) {
	if c == nil || c.list == nil {
		return
	}

	// 获取当前选中的索引
	selectedIndex := c.list.SelectedIndex()
	if selectedIndex < 0 {
		return
	}

	// TypeScript版本第423-424行：隐藏popup
	if c.dropdown != nil && c.dropdown.GObject != nil && c.dropdown.GObject.Parent() != nil {
		if root := core.Root(); root != nil {
			root.HidePopup(c.dropdown.GObject)
		}
	}

	// TypeScript版本第426-427行：设置selectedIndex
	c.selectedIndex = -1
	c.SetSelectedIndex(selectedIndex)

	// TypeScript版本第428行：派发STATE_CHANGED事件
	if c.GComponent != nil && c.GComponent.GObject != nil {
		c.GComponent.GObject.Emit(laya.EventStateChanged, nil)
	}
}

// ShowDropdown 公开方法，用于测试时手动触发下拉显示
func (c *GComboBox) ShowDropdown() {
	if c == nil {
		return
	}
	c.showDropdown()
}

func (c *GComboBox) showDropdown() {
	if c.list == nil || c.dropdown == nil {
		return
	}

	// 如果需要更新items
	if c.itemsUpdated {
		c.itemsUpdated = false

		// 清空列表
		itemCount := c.list.NumItems()
		for i := itemCount - 1; i >= 0; i-- {
			c.list.RemoveItemAt(i)
		}

		// 添加项目
		cnt := len(c.items)
		for i := 0; i < cnt; i++ {
			// 直接使用默认item模板
			defaultItem := c.list.DefaultItem()
			if defaultItem == "" {
				// 没有默认模板，无法创建item
				continue
			}

			item := c.list.getFromPool(defaultItem)
			if item == nil {
				// 尝试通过 FactoryObjectCreator 直接创建
				if creator := c.list.GetObjectCreator(); creator != nil {
					if obj := creator.CreateObject(defaultItem); obj != nil {
						item = obj
					}
				}
			}

			if item == nil {
				continue
			}

			if i < len(c.values) {
				item.SetName(c.values[i])
			} else {
				item.SetName("")
			}
			item.SetData(c.items[i])

			if i < len(c.icons) && c.icons[i] != "" {
				if loader, ok := item.Data().(*GLoader); ok {
					loader.SetURL(c.icons[i])
				} else if button, ok := item.Data().(*GButton); ok {
					button.SetIcon(c.icons[i])
				}
			}
			c.list.AddItem(item)
		}
	}

	c.list.SetSelectedIndex(-1)

	// 关键修复：确保 dropdown 尺寸正确计算
	// 在添加完列表项后，需要触发 list 的 updateBounds 来计算正确尺寸
	if c.list != nil && !c.list.IsVirtual() {
		c.list.GComponent.GObject.SetSize(c.GComponent.GObject.Width(), c.GComponent.GObject.Height())
		c.list.GComponent.SetBoundsChangedFlag()
		c.list.GComponent.EnsureBoundsCorrect()
		// 再次调用 updateBounds 确保尺寸正确
		c.list.GComponent.GObject.SetSize(c.GComponent.GObject.Width(), c.GComponent.GObject.Height())
	}

	// 使用 ComboBox 宽度和 list 实际高度设置 dropdown 尺寸
	dropdownWidth := c.GComponent.GObject.Width()
	dropdownHeight := c.dropdown.GObject.Height()
	// 如果 dropdown 高度为 0，尝试从 list 获取
	if dropdownHeight <= 0 && c.list != nil {
		dropdownHeight = c.list.GComponent.GObject.Height()
	}
	c.dropdown.GObject.SetSize(dropdownWidth, dropdownHeight)

	// 显示下拉框
	core.Root().TogglePopup(c.dropdown.GObject, c.GComponent.GObject, core.PopupDirection(c.popupDirection))
	if c.dropdown.Parent() != nil {
		c.setState(buttonStateDown)
	}
}
