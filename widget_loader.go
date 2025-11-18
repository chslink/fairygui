package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Loader - 简化的资源加载控件
// ============================================================================

// Loader 是简化的资源加载控件，包装了 pkg/fgui/widgets.GLoader。
type Loader struct {
	loader *widgets.GLoader
}

// NewLoader 创建一个新的资源加载控件。
//
// 示例：
//
//	loader := fairygui.NewLoader()
//	loader.SetURL("ui://Main/icon")
func NewLoader() *Loader {
	return &Loader{
		loader: widgets.NewLoader(),
	}
}

// URL 返回资源 URL。
func (l *Loader) URL() string {
	return l.loader.URL()
}

// SetURL 设置资源 URL。
//
// URL 格式: ui://packageName/itemName
//
// 示例：
//
//	loader.SetURL("ui://Main/icon")
func (l *Loader) SetURL(url string) {
	l.loader.SetURL(url)
}

// Color 返回颜色（十六进制格式）。
func (l *Loader) Color() string {
	return l.loader.Color()
}

// SetColor 设置颜色（十六进制格式）。
//
// 示例：
//
//	loader.SetColor("#FF0000")
func (l *Loader) SetColor(color string) {
	l.loader.SetColor(color)
}

// Playing 返回是否播放（用于 MovieClip）。
func (l *Loader) Playing() bool {
	return l.loader.Playing()
}

// SetPlaying 设置是否播放（用于 MovieClip）。
//
// 示例：
//
//	loader.SetPlaying(true)
func (l *Loader) SetPlaying(playing bool) {
	l.loader.SetPlaying(playing)
}

// Frame 返回当前帧（用于 MovieClip）。
func (l *Loader) Frame() int {
	return l.loader.Frame()
}

// SetFrame 设置当前帧（用于 MovieClip）。
//
// 示例：
//
//	loader.SetFrame(5)
func (l *Loader) SetFrame(frame int) {
	l.loader.SetFrame(frame)
}

// AutoSize 设置是否自动调整大小。
//
// 示例：
//
//	loader.SetAutoSize(true)
func (l *Loader) SetAutoSize(enabled bool) {
	l.loader.SetAutoSize(enabled)
}

// Fill 返回填充类型。
func (l *Loader) Fill() widgets.LoaderFillType {
	return l.loader.Fill()
}

// SetFill 设置填充类型。
//
// 示例：
//
//	loader.SetFill(widgets.LoaderFillScale)
func (l *Loader) SetFill(fill widgets.LoaderFillType) {
	l.loader.SetFill(fill)
}

// Align 返回水平对齐方式。
func (l *Loader) Align() widgets.LoaderAlign {
	return l.loader.Align()
}

// SetAlign 设置水平对齐方式。
//
// 示例：
//
//	loader.SetAlign(widgets.LoaderAlignCenter)
func (l *Loader) SetAlign(align widgets.LoaderAlign) {
	l.loader.SetAlign(align)
}

// SetVerticalAlign 设置垂直对齐方式。
//
// 示例：
//
//	loader.SetVerticalAlign(widgets.LoaderAlignMiddle)
func (l *Loader) SetVerticalAlign(align widgets.LoaderAlign) {
	l.loader.SetVerticalAlign(align)
}

// SetShrinkOnly 设置是否仅缩小。
//
// 示例：
//
//	loader.SetShrinkOnly(true)
func (l *Loader) SetShrinkOnly(enabled bool) {
	l.loader.SetShrinkOnly(enabled)
}

// FillMethod 返回填充方法。
func (l *Loader) FillMethod() int {
	return l.loader.FillMethod()
}

// SetFillMethod 设置填充方法。
//
// 示例：
//
//	loader.SetFillMethod(widgets.LoaderFillMethodHorizontal)
func (l *Loader) SetFillMethod(method int) {
	l.loader.SetFillMethod(method)
}

// FillOrigin 返回填充起点。
func (l *Loader) FillOrigin() int {
	return l.loader.FillOrigin()
}

// SetFillOrigin 设置填充起点。
//
// 示例：
//
//	loader.SetFillOrigin(0)
func (l *Loader) SetFillOrigin(origin int) {
	l.loader.SetFillOrigin(origin)
}

// FillClockwise 返回是否顺时针填充。
func (l *Loader) FillClockwise() bool {
	return l.loader.FillClockwise()
}

// SetFillClockwise 设置是否顺时针填充。
//
// 示例：
//
//	loader.SetFillClockwise(true)
func (l *Loader) SetFillClockwise(clockwise bool) {
	l.loader.SetFillClockwise(clockwise)
}

// FillAmount 返回填充量（0-1）。
func (l *Loader) FillAmount() float64 {
	return l.loader.FillAmount()
}

// SetFillAmount 设置填充量（0-1）。
//
// 示例：
//
//	loader.SetFillAmount(0.5)  // 填充 50%
func (l *Loader) SetFillAmount(amount float64) {
	l.loader.SetFillAmount(amount)
}

// SetScale9Grid 设置九宫格。
//
// 示例：
//
//	loader.SetScale9Grid(&assets.Rect{X: 10, Y: 10, Width: 20, Height: 20})
func (l *Loader) SetScale9Grid(grid *assets.Rect) {
	l.loader.SetScale9Grid(grid)
}

// SetScaleByTile 设置是否平铺缩放。
//
// 示例：
//
//	loader.SetScaleByTile(true)
func (l *Loader) SetScaleByTile(enabled bool) {
	l.loader.SetScaleByTile(enabled)
}

// Position 返回加载器位置。
func (l *Loader) Position() (x, y float64) {
	return l.loader.X(), l.loader.Y()
}

// SetPosition 设置加载器位置。
func (l *Loader) SetPosition(x, y float64) {
	l.loader.SetPosition(x, y)
}

// Size 返回加载器大小。
func (l *Loader) Size() (width, height float64) {
	return l.loader.Width(), l.loader.Height()
}

// SetSize 设置加载器大小。
func (l *Loader) SetSize(width, height float64) {
	l.loader.SetSize(width, height)
}

// Visible 返回加载器是否可见。
func (l *Loader) Visible() bool {
	return l.loader.Visible()
}

// SetVisible 设置加载器可见性。
func (l *Loader) SetVisible(visible bool) {
	l.loader.SetVisible(visible)
}

// Name 返回加载器名称。
func (l *Loader) Name() string {
	return l.loader.Name()
}

// SetName 设置加载器名称。
func (l *Loader) SetName(name string) {
	l.loader.SetName(name)
}

// Alpha 返回加载器透明度（0-1）。
func (l *Loader) Alpha() float64 {
	return l.loader.Alpha()
}

// SetAlpha 设置加载器透明度（0-1）。
func (l *Loader) SetAlpha(alpha float64) {
	l.loader.SetAlpha(alpha)
}

// RawLoader 返回底层的 widgets.GLoader 对象。
//
// 仅在需要访问底层 API 时使用。
func (l *Loader) RawLoader() *widgets.GLoader {
	return l.loader
}
