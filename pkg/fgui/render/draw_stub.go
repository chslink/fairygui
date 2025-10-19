//go:build !ebiten

package render

import (
	"errors"

	"github.com/chslink/fairygui/pkg/fgui/core"
)

// DrawComponent is unavailable without the ebiten tag.
func DrawComponent(_ any, _ *core.GComponent, _ *AtlasManager) error {
	return errors.New("render: DrawComponent requires ebiten build tag")
}
