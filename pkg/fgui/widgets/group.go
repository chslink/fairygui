package widgets

import "github.com/chslink/fairygui/pkg/fgui/core"

// GGroup mirrors FairyGUI 的组概念，仅充当布局元数据，不单独渲染。
type GGroup struct {
	*core.GObject
}

// NewGroup creates a new group widget backed by a bare GObject.
func NewGroup() *GGroup {
	obj := core.NewGObject()
	group := &GGroup{GObject: obj}
	obj.SetData(group)
	obj.SetTouchable(false)
	return group
}
