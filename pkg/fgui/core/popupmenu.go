package core

type popupMenuItem struct {
	caption   string
	handler   func()
	checkable bool
	checked   bool
	visible   bool
	grayed    bool
	separator bool
}

type PopupMenu struct {
	contentPane *GComponent
	items       []*popupMenuItem
	resourceURL string
}

// NewPopupMenu creates a popup menu.
func NewPopupMenu(resourceURL string) *PopupMenu {
	if resourceURL == "" {
		resourceURL = GetUIConfig().PopupMenu
	}
	return &PopupMenu{
		items:       make([]*popupMenuItem, 0),
		resourceURL: resourceURL,
	}
}

func (m *PopupMenu) AddItem(caption string, handler func()) {
	m.items = append(m.items, &popupMenuItem{
		caption: caption,
		handler: handler,
		visible: true,
	})
}

func (m *PopupMenu) AddSeparator() {
	m.items = append(m.items, &popupMenuItem{separator: true})
}

func (m *PopupMenu) SetItemVisible(index int, visible bool) {
	if index >= 0 && index < len(m.items) {
		m.items[index].visible = visible
	}
}

func (m *PopupMenu) SetItemGrayed(index int, grayed bool) {
	if index >= 0 && index < len(m.items) {
		m.items[index].grayed = grayed
	}
}

func (m *PopupMenu) SetItemCheckable(index int, checkable bool) {
	if index >= 0 && index < len(m.items) {
		m.items[index].checkable = checkable
	}
}

func (m *PopupMenu) SetItemChecked(index int, checked bool) {
	if index >= 0 && index < len(m.items) {
		m.items[index].checked = checked
	}
}

func (m *PopupMenu) RemoveItem(index int) {
	if index >= 0 && index < len(m.items) {
		m.items = append(m.items[:index], m.items[index+1:]...)
	}
}

func (m *PopupMenu) ClearItems() {
	m.items = m.items[:0]
}

func (m *PopupMenu) ItemCount() int {
	return len(m.items)
}

func (m *PopupMenu) Show(target *GObject, dir PopupDirection) {
	if m == nil {
		return
	}
	m.buildContent()
	if m.contentPane != nil {
		Root().ShowPopup(m.contentPane.GObject, target, dir)
	}
}

func (m *PopupMenu) Hide() {
	if m.contentPane != nil {
		Root().HidePopup(m.contentPane.GObject)
	}
}

func (m *PopupMenu) buildContent() {
	if m.contentPane != nil || len(m.items) == 0 {
		return
	}
	m.contentPane = NewGComponent()

	y := float64(0)
	itemHeight := float64(30)
	maxWidth := float64(150)
	for i, item := range m.items {
		if !item.visible {
			continue
		}
		if item.separator {
			sep := NewGObject()
			sep.SetSize(maxWidth, 2)
			sep.SetPosition(0, y)
			m.contentPane.AddChild(sep)
			y += 2
			continue
		}
		btn := NewGComponent()
		btn.SetSize(maxWidth, itemHeight)
		btn.SetPosition(0, y)

		idx := i
		btn.OnClick(func() {
			if idx < len(m.items) && m.items[idx].handler != nil && !m.items[idx].grayed {
				m.items[idx].handler()
			}
			m.Hide()
		})
		m.contentPane.AddChild(btn.GObject)
		y += itemHeight
	}
	m.contentPane.SetSize(maxWidth, y)
}

func (m *PopupMenu) AsComponent() *GComponent {
	m.buildContent()
	return m.contentPane
}
