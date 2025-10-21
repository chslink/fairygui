package scenes

import (
	"context"

	"github.com/chslink/fairygui/pkg/fgui/core"
)

// BagDemo is a skeletal port of the TypeScript bag demo.
// It currently loads the Bag package main component without interactive behaviour.
type BagDemo struct {
	component *core.GComponent
}

func (s *BagDemo) Name() string {
	return "BagDemo"
}

func (s *BagDemo) Load(ctx context.Context, mgr *Manager) (*core.GComponent, error) {
	env := mgr.Environment()
	pkg, err := env.Package(ctx, "Bag")
	if err != nil {
		return nil, err
	}
	item := chooseComponent(pkg, "Main")
	if item == nil {
		return nil, newMissingComponentError("Bag", "Main")
	}
	component, err := env.Factory.BuildComponent(ctx, pkg, item)
	if err != nil {
		return nil, err
	}
	s.component = component
	return component, nil
}

func (s *BagDemo) Dispose() {
	s.component = nil
}
