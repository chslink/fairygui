package core

import "github.com/chslink/fairygui/pkg/fgui/assets"

// OverflowType 是 assets.OverflowType 的别名
// 这样 core 包可以直接使用而无需导入 assets 包
type OverflowType = assets.OverflowType

const (
	// OverflowVisible 默认值，内容可见且不裁剪
	OverflowVisible = assets.OverflowTypeVisible

	// OverflowHidden 裁剪超出边界的内容
	// 使用 scrollRect 实现裁剪效果
	OverflowHidden = assets.OverflowTypeHidden

	// OverflowScroll 创建滚动区域
	// 由 SetupScroll 方法处理
	OverflowScroll = assets.OverflowTypeScroll
)

// Margin 定义组件的边距
// 用于 overflow 和布局计算
// 参考 TypeScript 版本：Margin 类
type Margin struct {
	Top    int
	Bottom int
	Left   int
	Right  int
}

// IsZero 检查 margin 是否全为零
func (m Margin) IsZero() bool {
	return m.Top == 0 && m.Bottom == 0 && m.Left == 0 && m.Right == 0
}
