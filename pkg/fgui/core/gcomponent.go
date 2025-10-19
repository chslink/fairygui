package core

// GComponent is a container object capable of holding child objects.
type GComponent struct {
	*GObject
	children    []*GObject
	controllers []*Controller
}

// NewGComponent constructs an empty UI component.
func NewGComponent() *GComponent {
	base := NewGObject()
	return &GComponent{
		GObject: base,
	}
}

// AddChild appends a child to the component.
func (c *GComponent) AddChild(child *GObject) {
	c.AddChildAt(child, len(c.children))
}

// AddChildAt inserts a child at the given index.
func (c *GComponent) AddChildAt(child *GObject, index int) {
	if child == nil {
		return
	}
	if child.parent != nil {
		child.parent.RemoveChild(child)
	}
	if index < 0 {
		index = 0
	} else if index > len(c.children) {
		index = len(c.children)
	}
	child.parent = c
	c.children = append(c.children, nil)
	copy(c.children[index+1:], c.children[index:])
	c.children[index] = child
	if c.DisplayObject() != nil && child.DisplayObject() != nil {
		c.DisplayObject().AddChild(child.DisplayObject())
	}
}

// RemoveChild removes a child from the component.
func (c *GComponent) RemoveChild(child *GObject) {
	for i, current := range c.children {
		if current == child {
			c.RemoveChildAt(i)
			return
		}
	}
}

// RemoveChildAt removes the child at the given index.
func (c *GComponent) RemoveChildAt(index int) {
	if index < 0 || index >= len(c.children) {
		return
	}
	child := c.children[index]
	child.parent = nil
	if c.DisplayObject() != nil && child.DisplayObject() != nil {
		c.DisplayObject().RemoveChild(child.DisplayObject())
	}
	copy(c.children[index:], c.children[index+1:])
	c.children = c.children[:len(c.children)-1]
}

// Children returns a snapshot of the child slice.
func (c *GComponent) Children() []*GObject {
	snapshot := make([]*GObject, len(c.children))
	copy(snapshot, c.children)
	return snapshot
}

// ChildAt returns the child at the specified index.
func (c *GComponent) ChildAt(index int) *GObject {
	if index < 0 || index >= len(c.children) {
		return nil
	}
	return c.children[index]
}

// Controllers returns the controllers on this component.
func (c *GComponent) Controllers() []*Controller {
	return append([]*Controller(nil), c.controllers...)
}

// AddController appends a controller.
func (c *GComponent) AddController(ctrl *Controller) {
	if ctrl == nil {
		return
	}
	c.controllers = append(c.controllers, ctrl)
}
