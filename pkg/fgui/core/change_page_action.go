package core

import "github.com/chslink/fairygui/pkg/fgui/utils"

type ChangePageAction struct {
	*ControllerAction

	ObjectID       string
	ControllerName string
	TargetPage     string
}

func NewChangePageAction() *ChangePageAction {
	return &ChangePageAction{
		ControllerAction: &ControllerAction{},
	}
}

func (a *ChangePageAction) Enter(ctrl *Controller) {
	if a == nil || ctrl == nil || ctrl.Parent() == nil {
		return
	}
	var target *GComponent
	if a.ObjectID != "" {
		obj := FindChildByPath(ctrl.Parent(), a.ObjectID)
		if obj != nil {
			target = obj.AsComponent()
		}
	} else {
		target = ctrl.Parent()
	}
	if target == nil {
		return
	}
	targetCtrl := target.ControllerByName(a.ControllerName)
	if targetCtrl == nil {
		return
	}
	switch a.TargetPage {
	case "~1":
		targetCtrl.SetSelectedIndex(ctrl.SelectedIndex())
	case "~2":
		targetCtrl.SetSelectedPageName(ctrl.SelectedPageName())
	default:
		targetCtrl.SetSelectedPageID(a.TargetPage)
	}
}

func (a *ChangePageAction) Leave(ctrl *Controller) {}

func (a *ChangePageAction) Setup(buf *utils.ByteBuffer) {
	if a == nil || buf == nil {
		return
	}
	a.ControllerAction.Setup(buf)
	s := buf.ReadS()
	if s != nil {
		a.ObjectID = *s
	}
	s = buf.ReadS()
	if s != nil {
		a.ControllerName = *s
	}
	s = buf.ReadS()
	if s != nil {
		a.TargetPage = *s
	}
}
