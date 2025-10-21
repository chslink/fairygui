package core

// GComponent is a container object capable of holding child objects.
type GComponent struct {
	*GObject
	children        []*GObject
	controllers     []*Controller
	opaque          bool
	mask            *GObject
	maskReversed    bool
	hitTest         HitTest
	transitions     []TransitionInfo
	transitionList  []*Transition
	transitionCache map[string]*Transition
}

// HitTestMode enumerates supported hit-test strategies.
type HitTestMode int

const (
	// HitTestModeNone indicates no custom hit test.
	HitTestModeNone HitTestMode = iota
	// HitTestModePixel references a pixel mask defined in package data.
	HitTestModePixel
	// HitTestModeChild references another child as the hit area.
	HitTestModeChild
)

// HitTest captures metadata required to perform hit testing.
type HitTest struct {
	Mode       HitTestMode
	ItemID     string
	OffsetX    int
	OffsetY    int
	ChildIndex int
}

// NewGComponent constructs an empty UI component.
func NewGComponent() *GComponent {
	base := NewGObject()
	return &GComponent{
		GObject: base,
		hitTest: HitTest{Mode: HitTestModeNone},
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
	for _, ctrl := range c.controllers {
		if ctrl != nil {
			child.HandleControllerChanged(ctrl)
		}
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

// ChildByName returns the first child with the given name.
func (c *GComponent) ChildByName(name string) *GObject {
	if name == "" {
		return nil
	}
	for _, child := range c.children {
		if child != nil && child.Name() == name {
			return child
		}
	}
	return nil
}

// Controllers returns the controllers on this component.
func (c *GComponent) Controllers() []*Controller {
	return append([]*Controller(nil), c.controllers...)
}

// Transitions 返回从构建数据解析出的 Transition 元数据快照。
func (c *GComponent) Transitions() []TransitionInfo {
	if c == nil || len(c.transitions) == 0 {
		return nil
	}
	out := make([]TransitionInfo, len(c.transitions))
	copy(out, c.transitions)
	return out
}

// Transition 返回指定名称的运行时 Transition。
func (c *GComponent) Transition(name string) *Transition {
	if c == nil || name == "" || c.transitionCache == nil {
		return nil
	}
	return c.transitionCache[name]
}

// TransitionAt 返回指定索引的运行时 Transition。
func (c *GComponent) TransitionAt(index int) *Transition {
	if c == nil || index < 0 || index >= len(c.transitionList) {
		return nil
	}
	return c.transitionList[index]
}

// RuntimeTransitions 返回运行时 Transition 的快照。
func (c *GComponent) RuntimeTransitions() []*Transition {
	if c == nil || len(c.transitionList) == 0 {
		return nil
	}
	out := make([]*Transition, len(c.transitionList))
	copy(out, c.transitionList)
	return out
}

// AddTransition 追加一条 Transition 元数据。
func (c *GComponent) AddTransition(info TransitionInfo) {
	if c == nil {
		return
	}
	index := len(c.transitions)
	c.transitions = append(c.transitions, info)
	c.upsertTransition(index, info)
}

func (c *GComponent) upsertTransition(index int, info TransitionInfo) {
	if c.transitionList == nil {
		c.transitionList = make([]*Transition, index+1)
	} else if len(c.transitionList) <= index {
		for len(c.transitionList) <= index {
			c.transitionList = append(c.transitionList, nil)
		}
	}
	tx := c.transitionList[index]
	if tx != nil {
		tx.reset(info)
	} else {
		tx = newTransition(c, info)
		c.transitionList[index] = tx
	}
	if key := info.Name; key != "" {
		if c.transitionCache == nil {
			c.transitionCache = make(map[string]*Transition)
		}
		c.transitionCache[key] = tx
	}
	if info.AutoPlay {
		tx.Play(info.AutoPlayTimes, info.AutoPlayDelay)
	}
}

// ControllerByName looks up a controller with the given name.
func (c *GComponent) ControllerByName(name string) *Controller {
 for _, ctrl := range c.controllers {
  if ctrl != nil && ctrl.Name == name {
   return ctrl
  }
 }
 return nil
}

// ControllerAt returns the controller at the provided index.
func (c *GComponent) ControllerAt(index int) *Controller {
	if c == nil || index < 0 || index >= len(c.controllers) {
		return nil
	}
	return c.controllers[index]
}

// AddController appends a controller.
func (c *GComponent) AddController(ctrl *Controller) {
	if ctrl == nil {
		return
	}
	if ctrl.Parent() != nil && ctrl.Parent() != c {
		ctrl.Parent().RemoveController(ctrl)
	}
	c.controllers = append(c.controllers, ctrl)
	ctrl.attach(c)
}

// RemoveController detaches the given controller from this component.
func (c *GComponent) RemoveController(ctrl *Controller) {
	if ctrl == nil {
		return
	}
	for i, stored := range c.controllers {
		if stored == ctrl {
			ctrl.detach(c)
			copy(c.controllers[i:], c.controllers[i+1:])
			c.controllers = c.controllers[:len(c.controllers)-1]
			for _, child := range c.children {
				if child != nil {
					child.HandleControllerChanged(ctrl)
				}
			}
			return
		}
	}
}

// ApplyAllControllers re-applies every controller on each child.
func (c *GComponent) ApplyAllControllers() {
	for _, ctrl := range c.controllers {
		c.applyController(ctrl)
	}
}

func (c *GComponent) applyController(ctrl *Controller) {
	if c == nil || ctrl == nil {
		return
	}
	for _, child := range c.children {
		if child != nil {
			child.HandleControllerChanged(ctrl)
		}
	}
}

// SetOpaque records whether the component blocks pointer events by default.
func (c *GComponent) SetOpaque(value bool) {
	if c == nil {
		return
	}
	c.opaque = value
}

// Opaque reports whether the component blocks pointer events.
func (c *GComponent) Opaque() bool {
	if c == nil {
		return false
	}
	return c.opaque
}

// SetMask assigns the child used as a mask along with inversion flag.
func (c *GComponent) SetMask(mask *GObject, reversed bool) {
	if c == nil {
		return
	}
	c.mask = mask
	c.maskReversed = reversed
}

// Mask returns the mask object and whether the mask is inverted.
func (c *GComponent) Mask() (*GObject, bool) {
	if c == nil {
		return nil, false
	}
	return c.mask, c.maskReversed
}

// SetHitTest stores the hit-test metadata associated with this component.
func (c *GComponent) SetHitTest(data HitTest) {
	if c == nil {
		return
	}
	c.hitTest = data
}

// HitTest returns the hit-test metadata.
func (c *GComponent) HitTest() HitTest {
	if c == nil {
		return HitTest{Mode: HitTestModeNone}
	}
	return c.hitTest
}
