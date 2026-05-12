package core

import "github.com/chslink/fairygui/pkg/fgui/utils"

type PlayTransitionAction struct {
	*ControllerAction

	TransitionName string
	PlayTimes      int
	Delay          float64
	StopOnExit     bool

	currentTransition *Transition
}

func NewPlayTransitionAction() *PlayTransitionAction {
	return &PlayTransitionAction{
		ControllerAction: &ControllerAction{},
		PlayTimes:        1,
	}
}

func (a *PlayTransitionAction) Enter(ctrl *Controller) {
	if a == nil || ctrl == nil || ctrl.Parent() == nil {
		return
	}
	comp := ctrl.Parent()
	trans := comp.Transition(a.TransitionName)
	if trans == nil {
		return
	}
	a.currentTransition = trans
	if a.Delay > 0 {
		// TODO: implement delayed play via scheduler
		trans.Play(a.PlayTimes, a.Delay)
	} else {
		trans.Play(a.PlayTimes, 0)
	}
}

func (a *PlayTransitionAction) Leave(ctrl *Controller) {
	if a == nil || !a.StopOnExit {
		return
	}
	if a.currentTransition != nil {
		a.currentTransition.Stop(true)
		a.currentTransition = nil
	}
}

func (a *PlayTransitionAction) Setup(buf *utils.ByteBuffer) {
	if a == nil || buf == nil {
		return
	}
	a.ControllerAction.Setup(buf)
	s := buf.ReadS()
	if s != nil {
		a.TransitionName = *s
	}
	a.PlayTimes = int(buf.ReadInt32())
	a.Delay = float64(buf.ReadFloat32())
	a.StopOnExit = buf.ReadBool()
}
