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

// SetupAfterAdd 在加入父节点后触发。当前实现无需额外处理。
func (g *GGroup) SetupAfterAdd(_ *SetupContext, _ *utils.ByteBuffer) {}

func clampGroupLayout(layout GroupLayoutType) GroupLayoutType {
	if layout < GroupLayoutTypeNone || layout > GroupLayoutTypeVertical {
		return GroupLayoutTypeNone
	}
	return layout
}
