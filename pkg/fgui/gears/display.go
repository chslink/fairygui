package gears

import (
	"fmt"

	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// GearDisplay controls an object's visibility across controller pages.
type GearDisplay struct {
	Base

	pages          []string
	visibleCounter int
	lockToken      uint32
}

// NewGearDisplay constructs a GearDisplay for the provided owner.
func NewGearDisplay(owner Owner) *GearDisplay {
	return &GearDisplay{
		Base:           NewBase(owner, IndexDisplay),
		visibleCounter: 1,
		lockToken:      1,
	}
}

// SetController assigns the driving controller.
func (g *GearDisplay) SetController(ctrl Controller) {
	if g == nil {
		return
	}
	g.Base.SetController(ctrl)
	g.Apply()
	// 手动触发可见性更新，因为 Apply() 只更新 visibleCounter
	// 需要调用 CheckGearDisplay() 来更新 internalVisible 和实际显示
	if owner := g.Owner(); owner != nil {
		if obj, ok := owner.(interface{ CheckGearDisplay() }); ok {
			obj.CheckGearDisplay()
		}
	}
}

// Setup initialises the gear from serialized data.
func (g *GearDisplay) Setup(buffer *utils.ByteBuffer, resolver ControllerResolver) {
	if g == nil || buffer == nil {
		return
	}
	ctrlIndex := int(buffer.ReadInt16())
	var ctrl Controller
	if resolver != nil && ctrlIndex >= 0 {
		ctrl = resolver.ControllerAt(ctrlIndex)
	}
	g.SetController(ctrl)
	count := int(buffer.ReadInt16())
	if count > 0 {
		raw := buffer.ReadSArray(count)
		pages := make([]string, 0, count)
		for _, entry := range raw {
			if entry != nil {
				pages = append(pages, *entry)
			} else {
				pages = append(pages, "")
			}
		}
		g.SetPages(pages)
	} else {
		g.SetPages(nil)
	}
}

// UpdateState is a no-op for GearDisplay since visibility is binary.
func (g *GearDisplay) UpdateState() {
	// No state caching required; visibility is derived from controller pages.
}

// Apply evaluates the selected page and toggles owner visibility.
// 参考 TypeScript 版本 GearDisplay.ts apply() 方法
// 只更新 visibleCounter，不调用 SetVisible
// 实际可见性由 CheckGearDisplay() 统一计算
func (g *GearDisplay) Apply() {
	if g == nil || g.Owner() == nil {
		return
	}
	visible := true
	ctrl := g.Controller()
	if ctrl != nil {
		if len(g.pages) > 0 {
			// GearDisplay支持两种pages格式：
			// 1. 页面ID字符串（如"up", "down", "over"）- 用于手动设置
			// 2. 页面索引字符串（如"0", "1", "2"）- 来自二进制数据
			// 参考TypeScript版本GearDisplay.ts apply()方法
			selectedIndex := ctrl.SelectedIndex()
			pageID := ctrl.SelectedPageID()

			visible = false
			// 先尝试作为页面ID匹配
			for _, candidate := range g.pages {
				if candidate == pageID {
					visible = true
					break
				}
			}
			// 如果没匹配到，尝试作为索引号匹配
			if !visible {
				currentPageStr := fmt.Sprintf("%d", selectedIndex)
				for _, candidate := range g.pages {
					if candidate == currentPageStr {
						visible = true
						break
					}
				}
			}
		}
	}
	if visible {
		g.visibleCounter = 1
	} else {
		g.visibleCounter = 0
	}
	// 注意：不调用 SetVisible，由 CheckGearDisplay 统一处理可见性更新
}

// UpdateFromRelations keeps compatibility with the Gear interface but has no effect.
func (g *GearDisplay) UpdateFromRelations(dx, dy float64) {
	// Visibility does not depend on relation deltas.
}

// SetPages stores the allowed page IDs for which the owner stays visible.
func (g *GearDisplay) SetPages(pages []string) {
	if g == nil {
		return
	}
	g.pages = append([]string(nil), pages...)
	g.Apply()
}

// Pages returns a copy of the stored page IDs.
func (g *GearDisplay) Pages() []string {
	if g == nil {
		return nil
	}
	return append([]string(nil), g.pages...)
}

// Connected mirrors FairyGUI semantics, signalling whether the object should remain visible.
// 如果对象有displayLock，即使visibleCounter为0也返回true，防止在动画期间被隐藏
// 参考 TypeScript 版本 GearDisplay.ts connected 属性的实现
func (g *GearDisplay) Connected() bool {
	if g == nil {
		return true
	}
	// 如果没有控制器，始终可见
	if g.Controller() == nil {
		return true
	}
	// 如果有displayLock，强制保持可见（用于tween动画期间）
	if owner, ok := g.Owner().(interface{ hasActiveDisplayLock() bool }); ok {
		if owner.hasActiveDisplayLock() {
			return true
		}
	}
	// 否则根据visibleCounter决定
	return g.visibleCounter > 0
}

// AddLock increments the visibility lock counter and returns a pseudo token.
func (g *GearDisplay) AddLock() uint32 {
	if g == nil {
		return 0
	}
	g.visibleCounter++
	g.lockToken++
	if g.lockToken == 0 {
		g.lockToken = 1
	}
	return g.lockToken
}

// ReleaseLock decrements the visibility lock counter when the token matches.
func (g *GearDisplay) ReleaseLock(token uint32) {
	if g == nil {
		return
	}
	if token == g.lockToken && g.visibleCounter > 0 {
		g.visibleCounter--
	}
}
