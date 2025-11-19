package fairygui

import (
	"sync"
)

// ============================================================================
// ControllerImpl - 控制器实现
// ============================================================================

type ControllerImpl struct {
	name          string
	pageNames     []string
	pageIDs       []string
	selectedIndex int
	selectedPageID string

	mu        sync.RWMutex
	listeners []ControllerSelectionListener
	parent    Component
}

// ControllerSelectionListener 控制器选择监听器
type ControllerSelectionListener func(Controller)

// NewController 创建控制器
func NewController(name string) *ControllerImpl {
	return &ControllerImpl{
		name:          name,
		selectedIndex: 0,
		listeners:     make([]ControllerSelectionListener, 0),
	}
}

// ============================================================================
// Controller 接口实现
// ============================================================================

func (c *ControllerImpl) Name() string                                    { return c.name }
func (c *ControllerImpl) SelectedIndex() int                              { return c.selectedIndex }
func (c *ControllerImpl) SelectedPage() string                            { return c.SelectedPage() }
func (c *ControllerImpl) PageCount() int                                  { return len(c.pageNames) }

func (c *ControllerImpl) SetSelectedIndex(index int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.pageNames) || c.selectedIndex == index {
		return
	}
	c.selectedIndex = index
	if index < len(c.pageIDs) {
		c.selectedPageID = c.pageIDs[index]
	}
	c.notifyListeners()
}

func (c *ControllerImpl) SetSelectedPage(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, page := range c.pageNames {
		if page == name && i != c.selectedIndex {
			c.selectedIndex = i
			c.selectedPageID = c.pageIDs[i]
			c.notifyListeners()
			return
		}
	}
}

func (c *ControllerImpl) OnChanged(handler func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.listeners = append(c.listeners, func(Controller) { handler() })
}

// ============================================================================
// 扩展方法
// ============================================================================

func (c *ControllerImpl) PageNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.pageNames))
	copy(result, c.pageNames)
	return result
}

func (c *ControllerImpl) SetPageNames(names []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pageNames = make([]string, len(names))
	copy(c.pageNames, names)
	if c.selectedIndex >= len(c.pageNames) {
		c.selectedIndex = len(c.pageNames) - 1
	}
}

func (c *ControllerImpl) AddPage(name, id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pageNames = append(c.pageNames, name)
	c.pageIDs = append(c.pageIDs, id)
}

func (c *ControllerImpl) SelectedPageID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.selectedPageID
}

func (c *ControllerImpl) SetSelectedPageID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, pageID := range c.pageIDs {
		if pageID == id {
			c.selectedIndex = i
			c.selectedPageID = id
			c.notifyListeners()
			return
		}
	}
}

func (c *ControllerImpl) SetSelectedPageName(name string) {
	for i, page := range c.pageNames {
		if page == name {
			c.SetSelectedIndex(i)
			return
		}
	}
}

func (c *ControllerImpl) SetOppositePageID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, pageID := range c.pageIDs {
		if pageID != id && i < len(c.pageNames) {
			c.selectedIndex = i
			c.selectedPageID = pageID
			c.notifyListeners()
			return
		}
	}
}

func (c *ControllerImpl) Parent() Component                    { return c.parent }
func (c *ControllerImpl) SetParent(parent Component)            { c.parent = parent }
func (c *ControllerImpl) AddSelectionListener(l ControllerSelectionListener) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := len(c.listeners)
	c.listeners = append(c.listeners, l)
	return id
}
func (c *ControllerImpl) RemoveSelectionListener(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if id >= 0 && id < len(c.listeners) {
		c.listeners[id] = nil
	}
}

func (c *ControllerImpl) notifyListeners() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, l := range c.listeners {
		if l != nil {
			l(c)
		}
	}
}
