package core

import "github.com/chslink/fairygui/pkg/fgui/utils"

type ControllerAction struct {
	FromPages []string
	ToPages   []string
}

func (a *ControllerAction) Run(ctrl *Controller, prevPage, curPage string) {
	if a == nil {
		return
	}
	fromMatch := len(a.FromPages) == 0
	if !fromMatch {
		for _, p := range a.FromPages {
			if p == prevPage {
				fromMatch = true
				break
			}
		}
	}
	toMatch := len(a.ToPages) == 0
	if !toMatch {
		for _, p := range a.ToPages {
			if p == curPage {
				toMatch = true
				break
			}
		}
	}
	if fromMatch {
		a.Enter(ctrl)
	}
	if toMatch {
		a.Leave(ctrl)
	}
}

func (a *ControllerAction) Enter(ctrl *Controller)  {}
func (a *ControllerAction) Leave(ctrl *Controller)  {}

func (a *ControllerAction) Setup(buf *utils.ByteBuffer) {
	if a == nil || buf == nil {
		return
	}
	cnt := int(buf.ReadInt16())
	a.FromPages = make([]string, 0, cnt)
	for i := 0; i < cnt; i++ {
		s := buf.ReadS()
		if s != nil {
			a.FromPages = append(a.FromPages, *s)
		}
	}
	cnt = int(buf.ReadInt16())
	a.ToPages = make([]string, 0, cnt)
	for i := 0; i < cnt; i++ {
		s := buf.ReadS()
		if s != nil {
			a.ToPages = append(a.ToPages, *s)
		}
	}
}

func createControllerAction(actionType int) *ControllerAction {
	switch actionType {
	case 0:
		return NewPlayTransitionAction().ControllerAction
	case 1:
		return NewChangePageAction().ControllerAction
	default:
		return nil
	}
}
