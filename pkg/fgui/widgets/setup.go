package widgets

import (
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// SetupContext carries contextual information required by widget setup hooks.
type SetupContext struct {
	Owner        *assets.PackageItem
	Child        *assets.ComponentChild
	Parent       *core.GComponent
	Package      *assets.Package
	ResolvedItem *assets.PackageItem
	ResolveIcon  func(icon string) *assets.PackageItem
}

// BeforeAdder widgets mirror FairyGUI setup_beforeAdd.
type BeforeAdder interface {
	SetupBeforeAdd(ctx *SetupContext, buf *utils.ByteBuffer)
}

// AfterAdder widgets mirror FairyGUI setup_afterAdd.
type AfterAdder interface {
	SetupAfterAdd(ctx *SetupContext, buf *utils.ByteBuffer)
}
