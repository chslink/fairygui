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
type UIConfig struct {
	HorizontalScrollBar           string
	VerticalScrollBar             string
	DefaultScrollBarDisplay       ScrollBarDisplayType
	DefaultScrollStep             float64
	DefaultScrollTouchEffect      bool
	DefaultScrollBounceEffect     bool
	ImageFilter                   ImageFilter
	ButtonSound                   string
	ButtonSoundVolumeScale        float64
	PopupMenu                     string
	PopupMenuSeperator            string
	GlobalModalWaiting            string
	WindowModalWaiting            string
	BringWindowToFrontOnClick     bool
	FrameTimeForAsyncUIConstruction float64
}

var globalUIConfig = &UIConfig{
	HorizontalScrollBar:              "",
	VerticalScrollBar:                "",
	DefaultScrollBarDisplay:          ScrollBarDisplayVisible,
	DefaultScrollStep:                25,
	DefaultScrollTouchEffect:         true,
	DefaultScrollBounceEffect:        true,
	ImageFilter:                      ImageFilterLinear,
	PopupMenu:                        "",
	ButtonSound:                      "",
	ButtonSoundVolumeScale:           1,
	BringWindowToFrontOnClick:        true,
	FrameTimeForAsyncUIConstruction:  0.002,
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
