package core

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
}

// globalUIConfig 全局配置实例
var globalUIConfig = &UIConfig{
	HorizontalScrollBar:       "",
	VerticalScrollBar:         "",
	DefaultScrollBarDisplay:   ScrollBarDisplayVisible,
	DefaultScrollStep:         25,
	DefaultScrollTouchEffect:  true,
	DefaultScrollBounceEffect: true,
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
