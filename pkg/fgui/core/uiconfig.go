package core

// ImageFilter represents the image filtering mode used when scaling images.
type ImageFilter int

const (
	// ImageFilterNearest uses nearest-neighbor filtering (crisp pixels).
	ImageFilterNearest ImageFilter = iota
	// ImageFilterLinear uses linear filtering (smooth interpolation).
	ImageFilterLinear
)

// UIConfig stores global FairyGUI configuration.
// 对应 TypeScript 版本 UIConfig.ts
type UIConfig struct {
	// 默认水平滚动条资源 URL
	HorizontalScrollBar string
	// 默认垂直滚动条资源 URL
	VerticalScrollBar string
	// 默认滚动条显示模式
	DefaultScrollBarDisplay ScrollBarDisplayType
	// 默认滚动步长（像素）
	DefaultScrollStep float64
	// 默认鼠标滚轮步长
	DefaultScrollTouchEffect bool
	DefaultScrollBounceEffect bool
	// 图片缩放时使用的全局Filter模式（默认使用线性过滤获得更好的抗锯齿效果）
	ImageFilter ImageFilter
}

// globalUIConfig 全局配置实例
var globalUIConfig = &UIConfig{
	HorizontalScrollBar:       "",
	VerticalScrollBar:         "",
	DefaultScrollBarDisplay:   ScrollBarDisplayVisible,
	DefaultScrollStep:         25,
	DefaultScrollTouchEffect:  true,
	DefaultScrollBounceEffect: true,
	ImageFilter:               ImageFilterLinear, // 默认使用线性过滤获得更好的抗锯齿效果
}

// GetUIConfig 返回全局 UIConfig 实例
func GetUIConfig() *UIConfig {
	return globalUIConfig
}

// SetDefaultScrollBars 设置默认滚动条资源 URL
func SetDefaultScrollBars(vertical, horizontal string) {
	globalUIConfig.VerticalScrollBar = vertical
	globalUIConfig.HorizontalScrollBar = horizontal
}

// SetImageFilter 设置图片缩放时使用的全局Filter模式
// filter 可以是 ImageFilterNearest 或 ImageFilterLinear
// 默认值是 ImageFilterLinear（更好的抗锯齿效果）
func SetImageFilter(filter ImageFilter) {
	globalUIConfig.ImageFilter = filter
}
