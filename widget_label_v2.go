package fairygui

// ============================================================================
// Label - 标签控件 V2
// ============================================================================

type Label struct {
	*ComponentImpl

	// 资源相关
	packageItem *PackageItemWrapper
	template    *ComponentImpl

	// 子对象
	titleObject DisplayObject
	iconObject  DisplayObject

	// 数据
	title             string
	icon              string
	iconItem          *PackageItemWrapper
	resource          string
	titleColor        string
	titleOutlineColor string
	titleFontSize     int
}

// NewLabel 创建新的标签
func NewLabel() *Label {
	label := &Label{
		ComponentImpl:     NewComponent(),
		titleColor:        "#ffffff",
		titleFontSize:     12,
	}

	return label
}

// ============================================================================
// 资源相关
// ============================================================================

// SetPackageItem 设置资源项
func (l *Label) SetPackageItem(item PackageItem) {
	if item == nil {
		l.packageItem = nil
		return
	}
	if wrapper, ok := item.(*PackageItemWrapper); ok {
		l.packageItem = wrapper
	}
}

// PackageItem 返回资源项
func (l *Label) PackageItem() PackageItem {
	return l.packageItem
}

// SetTemplateComponent 设置模板组件
func (l *Label) SetTemplateComponent(comp *ComponentImpl) {
	if l.template != nil {
		l.RemoveChild(l.template)
	}
	l.template = comp
	if comp != nil {
		comp.SetPosition(0, 0)
		l.AddChild(comp)
	}
	l.applyTitleState()
	l.applyIconState()
	l.applyTitleFormatting()
}

// TemplateComponent 返回模板组件
func (l *Label) TemplateComponent() *ComponentImpl {
	return l.template
}

// ============================================================================
// 文本相关
// ============================================================================

// SetTitle 设置标题文本
func (l *Label) SetTitle(value string) {
	l.title = value
	l.applyTitleState()
}

// Title 返回标题文本
func (l *Label) Title() string {
	return l.title
}

// SetTitleColor 设置标题颜色
func (l *Label) SetTitleColor(value string) {
	l.titleColor = value
	l.applyTitleFormatting()
}

// TitleColor 返回标题颜色
func (l *Label) TitleColor() string {
	return l.titleColor
}

// SetTitleOutlineColor 设置标题描边颜色
func (l *Label) SetTitleOutlineColor(value string) {
	l.titleOutlineColor = value
	l.applyTitleFormatting()
}

// TitleOutlineColor 返回标题描边颜色
func (l *Label) TitleOutlineColor() string {
	return l.titleOutlineColor
}

// SetTitleFontSize 设置标题字体大小
func (l *Label) SetTitleFontSize(size int) {
	l.titleFontSize = size
	l.applyTitleFormatting()
}

// TitleFontSize 返回标题字体大小
func (l *Label) TitleFontSize() int {
	return l.titleFontSize
}

// ============================================================================
// 图标相关
// ============================================================================

// SetIcon 设置图标
func (l *Label) SetIcon(value string) {
	l.icon = value
	l.applyIconState()
}

// Icon 返回图标
func (l *Label) Icon() string {
	return l.icon
}

// SetIconItem 设置图标项
func (l *Label) SetIconItem(item PackageItem) {
	if item == nil {
		l.iconItem = nil
		return
	}
	if wrapper, ok := item.(*PackageItemWrapper); ok {
		l.iconItem = wrapper
	}
}

// IconItem 返回图标项
func (l *Label) IconItem() PackageItem {
	return l.iconItem
}

// ============================================================================
// 资源相关
// ============================================================================

// SetResource 设置资源
func (l *Label) SetResource(res string) {
	l.resource = res
}

// Resource 返回资源
func (l *Label) Resource() string {
	return l.resource
}

// ============================================================================
// 子对象管理
// ============================================================================

// SetTitleObject 设置标题对象
func (l *Label) SetTitleObject(obj DisplayObject) {
	l.titleObject = obj
	l.applyTitleState()
	l.applyTitleFormatting()
}

// TitleObject 返回标题对象
func (l *Label) TitleObject() DisplayObject {
	return l.titleObject
}

// SetIconObject 设置图标对象
func (l *Label) SetIconObject(obj DisplayObject) {
	l.iconObject = obj
	l.applyIconState()
}

// IconObject 返回图标对象
func (l *Label) IconObject() DisplayObject {
	return l.iconObject
}

// ============================================================================
// 内部方法
// ============================================================================

func (l *Label) applyTitleState() {
	if l.titleObject == nil {
		return
	}

	text := l.title

	// 尝试不同类型
	if tf, ok := l.titleObject.(*TextField); ok {
		tf.SetText(text)
	} else if comp, ok := l.titleObject.(*ComponentImpl); ok {
		comp.SetData(text)
	} else if label, ok := l.titleObject.(*Label); ok {
		label.SetTitle(text)
	} else if btn, ok := l.titleObject.(*Button); ok {
		btn.SetTitle(text)
	}
}

func (l *Label) applyIconState() {
	if l.iconObject == nil {
		return
	}

	icon := l.icon

	// 尝试不同类型
	if loader, ok := l.iconObject.(*Loader); ok {
		loader.SetURL(icon)
	} else if btn, ok := l.iconObject.(*Button); ok {
		btn.SetIcon(icon)
	} else if label, ok := l.iconObject.(*Label); ok {
		label.SetIcon(icon)
	} else if comp, ok := l.iconObject.(*ComponentImpl); ok {
		comp.SetData(icon)
	}
}

func (l *Label) applyTitleFormatting() {
	if l.titleObject == nil {
		return
	}

	// 应用文本样式
	if tf, ok := l.titleObject.(*TextField); ok {
		if l.titleColor != "" {
			tf.SetColor(l.titleColor)
		}
		if l.titleOutlineColor != "" {
			// 注意：TextField 可能需要添加 SetOutlineColor 方法
			// tf.SetOutlineColor(l.titleOutlineColor)
		}
		if l.titleFontSize != 0 {
			tf.SetFontSize(l.titleFontSize)
		}
	} else if lbl, ok := l.titleObject.(*Label); ok {
		if l.titleColor != "" {
			lbl.SetTitleColor(l.titleColor)
		}
		if l.titleOutlineColor != "" {
			lbl.SetTitleOutlineColor(l.titleOutlineColor)
		}
		if l.titleFontSize != 0 {
			lbl.SetTitleFontSize(l.titleFontSize)
		}
	} else if btn, ok := l.titleObject.(*Button); ok {
		if l.titleColor != "" {
			btn.SetTitleColor(l.titleColor)
		}
		// TODO: Button 不支持描边颜色设置
		if l.titleFontSize != 0 {
			btn.SetTitleFontSize(l.titleFontSize)
		}
	}
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertLabel 类型断言
func AssertLabel(obj DisplayObject) (*Label, bool) {
	label, ok := obj.(*Label)
	return label, ok
}

// IsLabel 检查是否是 Label
func IsLabel(obj DisplayObject) bool {
	_, ok := obj.(*Label)
	return ok
}
