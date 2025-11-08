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
	// 默认按钮点击音效 URL
	ButtonSound string
	// 默认按钮点击音效音量（0-1）
	ButtonSoundVolumeScale float64
	// 默认右键菜单资源 URL
	PopupMenu string
}

// globalUIConfig 全局配置实例
var globalUIConfig = &UIConfig{
	HorizontalScrollBar:         "",
	VerticalScrollBar:           "",
	DefaultScrollBarDisplay:     ScrollBarDisplayVisible,
	DefaultScrollStep:           25,
	DefaultScrollTouchEffect:    true,
	DefaultScrollBounceEffect:   true,
	ImageFilter:                 ImageFilterLinear, // 默认使用线性过滤获得更好的抗锯齿效果
	PopupMenu:                   "",                // 默认右键菜单资源
	ButtonSound:                 "",                // 默认按钮点击音效
	ButtonSoundVolumeScale:      1,                 // 默认按钮点击音效音量
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

// SetDefaultButtonSound 设置默认按钮点击音效
// 对应 TypeScript 版本的 fgui.UIConfig.buttonSound
func SetDefaultButtonSound(soundURL string) {
	globalUIConfig.ButtonSound = soundURL
}

// SetDefaultPopupMenu 设置默认右键菜单资源
// 对应 TypeScript 版本的 fgui.UIConfig.popupMenu
func SetDefaultPopupMenu(menuURL string) {
	globalUIConfig.PopupMenu = menuURL
}
