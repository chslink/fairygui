package scenes

import (
	"fmt"
	"strings"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// ErrMissingComponent 在 Demo 加载过程中找不到目标组件时返回。
type ErrMissingComponent struct {
	Package string
	Target  string
}

func newMissingComponentError(pkg, target string) error {
	return ErrMissingComponent{Package: pkg, Target: target}
}

func (e ErrMissingComponent) Error() string {
	target := e.Target
	if target == "" {
		target = "<unknown>"
	}
	return fmt.Sprintf("scene: package %s missing component %s", e.Package, target)
}

// chooseComponent 根据候选名称挑选组件，若候选为空则返回包内第一个组件。
func chooseComponent(pkg *assets.Package, candidates ...string) *assets.PackageItem {
	if pkg == nil {
		return nil
	}
	for _, name := range candidates {
		if name == "" {
			continue
		}
		if item := pkg.ItemByName(name); item != nil && item.Type == assets.PackageItemTypeComponent && item.Component != nil {
			return item
		}
	}
	if len(candidates) > 0 {
		return nil
	}
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Component != nil {
			return item
		}
	}
	return nil
}

func describeCandidates(candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	return strings.Join(candidates, ", ")
}
