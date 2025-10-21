package scenes

import (
	"context"

	"github.com/chslink/fairygui/pkg/fgui/core"
)

type simpleScene struct {
	name        string
	packageName string
	candidates  []string
	component   *core.GComponent
}

// NewSimpleScene 构建一个仅需加载指定包组件的简单 Demo 场景。
func NewSimpleScene(name, packageName string, candidates ...string) Scene {
	if len(candidates) == 0 {
		candidates = []string{"Main"}
	}
	names := make([]string, len(candidates))
	copy(names, candidates)
	return &simpleScene{
		name:        name,
		packageName: packageName,
		candidates:  names,
	}
}

func (s *simpleScene) Name() string {
	return s.name
}

func (s *simpleScene) Load(ctx context.Context, mgr *Manager) (*core.GComponent, error) {
	env := mgr.Environment()
	pkg, err := env.Package(ctx, s.packageName)
	if err != nil {
		return nil, err
	}
	item := chooseComponent(pkg, s.candidates...)
	if item == nil {
		return nil, newMissingComponentError(s.packageName, describeCandidates(s.candidates))
	}
	component, err := env.Factory.BuildComponent(ctx, pkg, item)
	if err != nil {
		return nil, err
	}
	s.component = component
	return component, nil
}

func (s *simpleScene) Dispose() {
	s.component = nil
}
