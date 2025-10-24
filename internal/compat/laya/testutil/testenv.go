package testutil

import (
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
)

// StageEnv bundles helpers to drive a stage inside tests.
type StageEnv struct {
	T         *testing.T
	Stage     *laya.Stage
	Time      time.Duration
	lastMouse laya.MouseState
}

// NewStageEnv creates a new environment with a stage of fixed size.
func NewStageEnv(t *testing.T, width, height int) *StageEnv {
	t.Helper()
	return &StageEnv{
		T:     t,
		Stage: laya.NewStage(width, height),
	}
}

// Advance progresses the simulated time and updates the stage with the given mouse state.
func (env *StageEnv) Advance(delta time.Duration, mouse laya.MouseState) {
	env.T.Helper()
	env.Time += delta
	env.lastMouse = mouse
	env.Stage.Update(delta, mouse)
}

// AdvanceInput progresses the simulated time with a full input state.
func (env *StageEnv) AdvanceInput(delta time.Duration, input laya.InputState) {
	env.T.Helper()
	env.Time += delta
	env.lastMouse = input.Mouse
	env.Stage.UpdateInput(delta, input)
}

// Scheduler returns the stage scheduler for chaining timer assertions.
func (env *StageEnv) Scheduler() *laya.Scheduler {
	return env.Stage.Scheduler()
}

// FlushScheduler advances the scheduler by delta without changing mouse state.
func (env *StageEnv) FlushScheduler(delta time.Duration) {
	env.Advance(delta, env.StageMouse())
}

// StageMouse returns the last mouse state applied to the stage.
func (env *StageEnv) StageMouse() laya.MouseState {
	state := env.lastMouse
	state.WheelX = 0
	state.WheelY = 0
	return state
}
