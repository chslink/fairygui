package widgets

import (
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// GroupLayoutType mirrors FairyGUI GroupLayoutType 枚举。
type GroupLayoutType int

const (
	GroupLayoutTypeNone GroupLayoutType = iota
	GroupLayoutTypeHorizontal
	GroupLayoutTypeVertical
)

// GGroup mirrors FairyGUI 的组概念，仅充当布局元数据，不单独渲染。
type GGroup struct {
	*core.GObject

	layout            GroupLayoutType
	lineGap           int
	columnGap         int
	excludeInvisibles bool
	autoSizeDisabled  bool
	mainGridIndex     int
	mainGridMinSize   int
	updating          int // 位标志，用于避免递归更新
}

// NewGroup creates a new group widget backed by a bare GObject.
func NewGroup() *GGroup {
	obj := core.NewGObject()
	group := &GGroup{
		GObject:         obj,
		layout:          GroupLayoutTypeNone,
		mainGridIndex:   -1,
		mainGridMinSize: 50,
	}
	obj.SetData(group)
	obj.SetTouchable(false)
	return group
}

// Layout 当前布局方式。
func (g *GGroup) Layout() GroupLayoutType {
	return g.layout
}

// LineGap 行间距。
func (g *GGroup) LineGap() int {
	return g.lineGap
}

// ColumnGap 列间距。
func (g *GGroup) ColumnGap() int {
	return g.columnGap
}

// ExcludeInvisibles 返回是否排除不可见子元素。
func (g *GGroup) ExcludeInvisibles() bool {
	return g.excludeInvisibles
}

// AutoSizeDisabled 返回自动尺寸是否被禁用。
func (g *GGroup) AutoSizeDisabled() bool {
	return g.autoSizeDisabled
}

// MainGridIndex 返回主要网格索引。
func (g *GGroup) MainGridIndex() int {
	return g.mainGridIndex
}

// MainGridMinSize 返回主要网格最小尺寸。
func (g *GGroup) MainGridMinSize() int {
	return g.mainGridMinSize
}

// SetupBeforeAdd 解析组在加入父节点前的配置。
func (g *GGroup) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if g == nil || buf == nil {
		return
	}

	// 首先调用父类GObject处理基础属性
	g.GObject.SetupBeforeAdd(buf, beginPos)

	// 然后处理GGroup特定属性（block 5）
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(beginPos, 5) || buf.Remaining() <= 0 {
		return
	}
	g.layout = clampGroupLayout(GroupLayoutType(buf.ReadByte()))
	if buf.Remaining() >= 4 {
		g.lineGap = int(buf.ReadInt32())
	}
	if buf.Remaining() >= 4 {
		g.columnGap = int(buf.ReadInt32())
	}
	if buf.Version >= 2 {
		if buf.Remaining() > 0 {
			g.excludeInvisibles = buf.ReadBool()
		}
		if buf.Remaining() > 0 {
			g.autoSizeDisabled = buf.ReadBool()
		}
		if buf.Remaining() >= 2 {
			g.mainGridIndex = int(buf.ReadInt16())
		}
	}
}

// SetupAfterAdd 在加入父节点后触发
// 参考 TypeScript 版本 GGroup.ts setup_afterAdd (441-446行)
func (g *GGroup) SetupAfterAdd(ctx *SetupContext, buf *utils.ByteBuffer) {
	// 调用父类的 SetupAfterAdd（处理 gears 等）
	// 注意：GObject.SetupAfterAdd 需要父组件参数，但这里我们通过 g.GObject.Parent() 获取
	if g.GObject != nil && g.GObject.Parent() != nil && buf != nil {
		// 这里需要 beginPos，但我们没有，所以跳过调用父类方法
		// 实际上 GGroup 不需要父类的 SetupAfterAdd 处理
	}

	// 关键：如果 Group 不可见，需要同步子元素的可见性
	// 这确保了在 gear 应用后，子元素能够正确隐藏
	if g.GObject != nil && !g.GObject.Visible() {
		g.HandleVisibleChanged()
	}
}

// HandleVisibleChanged 覆盖 GObject 的实现
// Group 的可见性改变时，同步所有属于该 Group 的子元素
// 参考 TypeScript 版本 GGroup.ts handleVisibleChanged (414-424行)
func (g *GGroup) HandleVisibleChanged() {
	if g == nil || g.GObject == nil {
		return
	}

	// 首先更新 Group 自己的 displayObject
	g.GObject.HandleVisibleChanged()

	// 然后同步所有属于该 Group 的子元素
	parent := g.GObject.Parent()
	if parent == nil {
		return
	}

	for _, child := range parent.Children() {
		if child != nil && child.Group() == g.GObject {
			child.HandleVisibleChanged()
		}
	}
}

// MoveChildren 移动所有属于该 Group 的子元素
// 当 Group 的位置改变时自动调用
// 参考 TypeScript 版本 GGroup.ts moveChildren (238-250行)
func (g *GGroup) MoveChildren(dx, dy float64) {
	if g == nil || g.GObject == nil {
		return
	}

	// 检查是否正在更新中（避免递归）
	if (g.updating & 1) != 0 {
		return
	}

	parent := g.GObject.Parent()
	if parent == nil {
		return
	}

	// 设置更新标志
	g.updating |= 1
	defer func() { g.updating &^= 1 }()

	// 遍历父组件的所有子元素，移动属于该 Group 的子元素
	for _, child := range parent.Children() {
		if child != nil && child.Group() == g.GObject {
			child.SetPosition(child.X()+dx, child.Y()+dy)
		}
	}
}

func clampGroupLayout(layout GroupLayoutType) GroupLayoutType {
	if layout < GroupLayoutTypeNone || layout > GroupLayoutTypeVertical {
		return GroupLayoutTypeNone
	}
	return layout
}
