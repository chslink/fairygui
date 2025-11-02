package core

// Controller models a simple controller with page list and selection state.
type Controller struct {
	Name           string
	AutoRadio      bool
	PageNames      []string
	PageIDs        []string
	selectedIndex  int
	previousIndex  int
	parent         *GComponent
	changing       bool
	nextListenerID int
	listeners      map[int]func(*Controller)
}

// NewController constructs a controller.
func NewController(name string) *Controller {
	return &Controller{Name: name, selectedIndex: -1, previousIndex: -1}
}

// Parent returns the component that owns this controller.
func (c *Controller) Parent() *GComponent {
	if c == nil {
		return nil
	}
	return c.parent
}

// SetPages replaces the controller pages and normalises selection.
func (c *Controller) SetPages(pageIDs, pageNames []string) {
	if c == nil {
		return
	}
	if pageIDs != nil {
		c.PageIDs = append(make([]string, 0, len(pageIDs)), pageIDs...)
	} else {
		c.PageIDs = c.PageIDs[:0]
	}
	if pageNames != nil {
		c.PageNames = append(make([]string, 0, len(pageNames)), pageNames...)
	} else {
		c.PageNames = c.PageNames[:0]
	}
	oldIndex := c.selectedIndex
	// oldPageID := ""
	// if oldIndex >= 0 && oldIndex < len(c.PageIDs) {
	// 	oldPageID = c.PageIDs[oldIndex]
	// }
	c.normalizeSelection()
	if oldIndex != c.selectedIndex {
		c.previousIndex = oldIndex
		c.applySelection()
		c.notifySelectionChanged()
	}
}

// PageCount returns the number of configured pages.
func (c *Controller) PageCount() int {
	if c == nil {
		return 0
	}
	return len(c.PageIDs)
}

// SelectedIndex returns the active page index.
func (c *Controller) SelectedIndex() int {
	if c == nil {
		return -1
	}
	c.normalizeSelection()
	return c.selectedIndex
}

// PreviousIndex returns the previously selected page index.
func (c *Controller) PreviousIndex() int {
	if c == nil {
		return -1
	}
	return c.previousIndex
}

// SetSelectedIndex selects the page by index.
func (c *Controller) SetSelectedIndex(index int) {
	if c == nil {
		return
	}
	if len(c.PageIDs) == 0 {
		index = -1
	} else {
		if index < 0 {
			index = 0
		}
		if index >= len(c.PageIDs) {
			index = len(c.PageIDs) - 1
		}
	}
	if index == c.selectedIndex {
		return
	}

	// 记录旧状态
	// oldIndex := c.selectedIndex
	// oldPageID := ""
	// if oldIndex >= 0 && oldIndex < len(c.PageIDs) {
	// 	oldPageID = c.PageIDs[oldIndex]
	// }

	c.previousIndex = c.selectedIndex
	c.selectedIndex = index
	c.applySelection()
	c.notifySelectionChanged()
}

// SelectedPageID returns the identifier for the active page.
func (c *Controller) SelectedPageID() string {
	if c == nil {
		return ""
	}
	index := c.SelectedIndex()
	if index < 0 || index >= len(c.PageIDs) {
		return ""
	}
	return c.PageIDs[index]
}

// SelectedPageName returns the name for the active page.
func (c *Controller) SelectedPageName() string {
	if c == nil {
		return ""
	}
	index := c.SelectedIndex()
	if index < 0 || index >= len(c.PageNames) {
		return ""
	}
	return c.PageNames[index]
}

// SetSelectedPageID selects the page matching the provided ID.
func (c *Controller) SetSelectedPageID(id string) {
	if c == nil || len(c.PageIDs) == 0 {
		c.SetSelectedIndex(-1)
		return
	}
	target := -1
	for idx, stored := range c.PageIDs {
		if stored == id {
			target = idx
			break
		}
	}
	if target == -1 {
		target = 0
	}
	c.SetSelectedIndex(target)
}

// SetSelectedPageName selects the page matching the provided name.
func (c *Controller) SetSelectedPageName(name string) {
	if c == nil || len(c.PageNames) == 0 {
		c.SetSelectedIndex(-1)
		return
	}
	target := -1
	for idx, stored := range c.PageNames {
		if stored == name {
			target = idx
			break
		}
	}
	if target == -1 {
		target = 0
	}
	c.SetSelectedIndex(target)
}

// SetOppositePageID switches to the opposite page for toggle-style behavior.
// 参考 TypeScript 原版：Controller.ts oppositePageId setter
// 如果传入的 pageID 在控制器页面列表中的索引 > 0，切换到 page 0
// 否则，如果有多个页面，切换到 page 1
func (c *Controller) SetOppositePageID(id string) {
	if c == nil || len(c.PageIDs) == 0 {
		return
	}
	index := -1
	for idx, stored := range c.PageIDs {
		if stored == id {
			index = idx
			break
		}
	}
	if index > 0 {
		c.SetSelectedIndex(0)
	} else if len(c.PageIDs) > 1 {
		c.SetSelectedIndex(1)
	}
}

func (c *Controller) attach(parent *GComponent) {
	if c == nil {
		return
	}
	c.parent = parent
	c.normalizeSelection()
	c.applySelection()
}

func (c *Controller) detach(parent *GComponent) {
	if c == nil || c.parent != parent {
		return
	}
	c.parent = nil
}

func (c *Controller) applySelection() {
	if c == nil {
		return
	}
	if c.parent == nil {
		return
	}
	if c.changing {
		return
	}
	c.changing = true
	c.parent.applyController(c)
	c.changing = false
}

func (c *Controller) normalizeSelection() {
	if c == nil {
		return
	}
	if len(c.PageIDs) == 0 {
		c.selectedIndex = -1
		return
	}
	if c.selectedIndex >= len(c.PageIDs) {
		c.selectedIndex = len(c.PageIDs) - 1
		return
	}
	if c.selectedIndex < 0 {
		c.selectedIndex = 0
	}
}

// AddSelectionListener registers a callback fired when the selected index changes.
// It returns a listener ID that can be used to remove the listener later.
func (c *Controller) AddSelectionListener(listener func(*Controller)) int {
	if c == nil || listener == nil {
		return 0
	}
	if c.listeners == nil {
		c.listeners = make(map[int]func(*Controller))
	}
	c.nextListenerID++
	id := c.nextListenerID
	c.listeners[id] = listener
	return id
}

// RemoveSelectionListener removes the listener registered with the provided ID.
func (c *Controller) RemoveSelectionListener(id int) {
	if c == nil || id == 0 || len(c.listeners) == 0 {
		return
	}
	delete(c.listeners, id)
	if len(c.listeners) == 0 {
		c.listeners = nil
	}
}

func (c *Controller) notifySelectionChanged() {
	if c == nil || len(c.listeners) == 0 {
		return
	}
	// Copy listeners to avoid mutation during iteration.
	tmp := make([]func(*Controller), 0, len(c.listeners))
	for _, fn := range c.listeners {
		tmp = append(tmp, fn)
	}
	for _, fn := range tmp {
		fn(c)
	}
}
