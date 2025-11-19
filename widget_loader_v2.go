package fairygui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// Loader - 加载器控件 V2 (简化实现)
// ============================================================================

type Loader struct {
	*Object

	url               string
	packageItem       PackageItem
	autoSize          bool
	fill              int // 0=none, 1=scaleFree, 2=scaleMatchHeight, 3=scaleMatchWidth
	align             TextAlign
	verticalAlign     VerticalAlign
}

// NewLoader 创建新的加载器
func NewLoader() *Loader {
	loader := &Loader{
		Object:        NewObject(),
		align:         TextAlignCenter,
		verticalAlign: VerticalAlignMiddle,
	}

	// 默认不拦截事件
	loader.SetTouchable(false)

	return loader
}

// ============================================================================
// URL 和 PackageItem
// ============================================================================

// SetURL 设置加载的 URL
func (l *Loader) SetURL(value string) {
	if l.url == value {
		return
	}

	l.url = value

	// 解析 URL 并加载资源
	if value == "" {
		l.packageItem = nil
	} else {
		l.loadFromURL(value)
	}

	l.updateGraphics()
}

// URL 返回 URL
func (l *Loader) URL() string {
	return l.url
}

// SetPackageItem 设置资源项
func (l *Loader) SetPackageItem(item PackageItem) {
	if l.packageItem == item {
		return
	}

	l.packageItem = item
	l.updateGraphics()
}

// PackageItem 返回资源项
func (l *Loader) PackageItem() PackageItem {
	return l.packageItem
}

// ============================================================================
// 布局参数
// ============================================================================

// SetAutoSize 设置自动大小
func (l *Loader) SetAutoSize(value bool) {
	if l.autoSize == value {
		return
	}
	l.autoSize = value
	l.updateGraphics()
}

// AutoSize 返回自动大小
func (l *Loader) AutoSize() bool {
	return l.autoSize
}

// Fill 返回填充模式
func (l *Loader) Fill() int {
	return l.fill
}

// SetFill 设置填充模式
func (l *Loader) SetFill(value int) {
	if l.fill == value {
		return
	}
	l.fill = value
	l.updateGraphics()
}

// Align 返回水平对齐
func (l *Loader) Align() TextAlign {
	return l.align
}

// SetAlign 设置水平对齐
func (l *Loader) SetAlign(value TextAlign) {
	if l.align == value {
		return
	}
	l.align = value
	l.updateGraphics()
}

// VerticalAlign 返回垂直对齐
func (l *Loader) VerticalAlign() VerticalAlign {
	return l.verticalAlign
}

// SetVerticalAlign 设置垂直对齐
func (l *Loader) SetVerticalAlign(value VerticalAlign) {
	if l.verticalAlign == value {
		return
	}
	l.verticalAlign = value
	l.updateGraphics()
}

// ============================================================================
// 内部方法
// ============================================================================

// loadFromURL 从 URL 加载内容
// URL 格式: ui://packageName/itemName
func (l *Loader) loadFromURL(url string) {
	// 解析 URL
	if len(url) > 6 && url[:6] == "ui://" {
		// 解析 fairygui 协议
		// TODO: 实现完整的 URL 解析
		fmt.Printf("加载 FairyGUI URL: %s\n", url)
	} else {
		// 普通资源路径
		// TODO: 加载资源
	}
}

// updateGraphics 更新图形
func (l *Loader) updateGraphics() {
	if l.packageItem == nil {
		l.SetTexture(nil)
		return
	}

	// 根据填充模式设置尺寸
	if l.autoSize {
		width := l.packageItem.Width()
		height := l.packageItem.Height()
		l.SetSize(float64(width), float64(height))
	}
}

// Draw 绘制
func (l *Loader) Draw(screen *ebiten.Image) {
	// 先调用父类绘制
	l.Object.Draw(screen)

	// TODO: 根据填充模式和对齐方式绘制内容
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertLoader 类型断言
func AssertLoader(obj DisplayObject) (*Loader, bool) {
	loader, ok := obj.(*Loader)
	return loader, ok
}

// IsLoader 检查是否是 Loader
func IsLoader(obj DisplayObject) bool {
	_, ok := obj.(*Loader)
	return ok
}
