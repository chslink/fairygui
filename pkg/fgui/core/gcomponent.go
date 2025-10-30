package core

import (
	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// GComponent is a container object capable of holding child objects.
type GComponent struct {
	*GObject
	container          *laya.Sprite
	scrollPane         *ScrollPane
	children           []*GObject
	controllers        []*Controller
	opaque             bool
	mask               *GObject
	maskReversed       bool
	hitTest            HitTest
	transitions        []TransitionInfo
	transitionList     []*Transition
	transitionCache    map[string]*Transition
	sortingChildCount  int // Number of children with sortingOrder > 0
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
	comp := &GComponent{
		GObject:   base,
		container: base.DisplayObject(),
		hitTest:   HitTest{Mode: HitTestModeNone},
		opaque:    false, // 修复：默认opaque=false，与TypeScript版本一致，避免容器拦截子元素事件
	}
	base.SetData(comp)

	// 参考 TypeScript 原版：GComponent.ts createDisplayObject()
	// opaque=false时，mouseThrough=true，让事件穿透到子元素
	if sprite := base.DisplayObject(); sprite != nil {
		sprite.SetMouseEnabled(true)
		sprite.SetMouseThrough(true) // opaque=false → mouseThrough=true
	}

	return comp
}

// ComponentRoot satisfies core.ComponentAccessor.
func (c *GComponent) ComponentRoot() *GComponent {
	return c
}

func (c *GComponent) childContainer() *laya.Sprite {
	if c == nil {
		return nil
	}
	if c.container == nil {
		c.container = c.DisplayObject()
	}
	return c.container
}

func (c *GComponent) ensureContainer() *laya.Sprite {
	if c == nil {
		return nil
	}
	display := c.DisplayObject()
	if c.container == nil {
		c.container = display
		return c.container
	}
	if c.container != display || display == nil {
		return c.container
	}
	newContainer := laya.NewSprite()
	newContainer.SetOwner(c)
	children := display.Children()
	for _, child := range children {
		display.RemoveChild(child)
		newContainer.AddChild(child)
	}
	display.AddChild(newContainer)
	c.container = newContainer
	return c.container
}

// Container returns the internal display container holding child sprites。
func (c *GComponent) Container() *laya.Sprite {
	return c.childContainer()
}

// EnsureContainer guarantees the component uses a dedicated child container and returns it。
func (c *GComponent) EnsureContainer() *laya.Sprite {
	return c.ensureContainer()
}

// SetupScroll parses scroll pane配置并绑定到组件上。
func (c *GComponent) SetupScroll(buf *utils.ByteBuffer) {
	if c == nil || buf == nil {
		return
	}
	scrollTypeValue := int(buf.ReadByte())
	scrollBarDisplay := int(buf.ReadByte())
	flags := buf.ReadInt32()
	if buf.ReadBool() {
		_ = buf.ReadInt32() // scrollBarMargin.top
		_ = buf.ReadInt32() // scrollBarMargin.bottom
		_ = buf.ReadInt32() // scrollBarMargin.left
		_ = buf.ReadInt32() // scrollBarMargin.right
	}
	vtScrollBarRes := buf.ReadS()  // 垂直滚动条资源 URL
	hzScrollBarRes := buf.ReadS()  // 水平滚动条资源 URL
	headerRes := buf.ReadS()       // header 资源 URL
	footerRes := buf.ReadS()       // footer 资源 URL
	_ = headerRes
	_ = footerRes


	mode := ScrollTypeBoth
	switch scrollTypeValue {
	case int(ScrollTypeHorizontal):
		mode = ScrollTypeHorizontal
	case int(ScrollTypeVertical):
		mode = ScrollTypeVertical
	}
	pane := c.EnsureScrollPane(mode)
	if pane == nil {
		return
	}

	// 保存滚动条配置
	pane.scrollBarDisplay = scrollBarDisplay
	if vtScrollBarRes != nil {
		pane.vtScrollBarURL = *vtScrollBarRes
	}
	if hzScrollBarRes != nil {
		pane.hzScrollBarURL = *hzScrollBarRes
	}

	// 解析 flags
	pane.SetMouseWheelEnabled(scrollBarDisplay != int(ScrollBarDisplayHidden))
	if flags&1 != 0 {
		pane.displayOnLeft = true
	}
	if flags&2 != 0 {
		pane.snapToItem = true
	}
	if flags&8 != 0 {
		pane.pageMode = true
		pane.pageSize = laya.Point{}
	}
	if flags&16 != 0 {
		pane.touchEffect = true
	} else if flags&32 != 0 {
		pane.touchEffect = false
	}
	if flags&256 != 0 {
		pane.inertiaDisabled = true
	}
	if flags&1024 != 0 {
		pane.floating = true
	}
	pane.OnOwnerSizeChanged()
}

// SetSize overrides GObject.SetSize to同步滚动视图。
func (c *GComponent) SetSize(width, height float64) {
	if c == nil {
		return
	}
	oldWidth := c.Width()
	oldHeight := c.Height()
	c.GObject.SetSize(width, height)
	if c.scrollPane != nil && (width != oldWidth || height != oldHeight) {
		c.scrollPane.OnOwnerSizeChanged()
	}
}

// AddChild appends a child to the component.
func (c *GComponent) AddChild(child *GObject) {
	c.AddChildAt(child, len(c.children))
}

// AddChildAt inserts a child at the given index.
// If child has sortingOrder > 0, the index is adjusted to maintain sorted order.
func (c *GComponent) AddChildAt(child *GObject, index int) {
	if child == nil {
		return
	}
	if child.parent != nil {
		child.parent.RemoveChild(child)
	}

	cnt := len(c.children)
	if index < 0 {
		index = 0
	} else if index > cnt {
		index = cnt
	}

	// Handle sortingOrder: children with sortingOrder > 0 are kept sorted at the end
	if child.sortingOrder != 0 {
		c.sortingChildCount++
		index = c.getInsertPosForSortingChild(child)
	} else if c.sortingChildCount > 0 {
		// Non-sorting children must be inserted before sorting children
		if index > (cnt - c.sortingChildCount) {
			index = cnt - c.sortingChildCount
		}
	}

	child.parent = c
	c.children = append(c.children, nil)
	copy(c.children[index+1:], c.children[index:])
	c.children[index] = child
	if container := c.childContainer(); container != nil && child.DisplayObject() != nil {
		container.AddChild(child.DisplayObject())
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

	// Update sorting child count if necessary
	if child.sortingOrder != 0 {
		c.sortingChildCount--
	}

	if container := c.childContainer(); container != nil && child.DisplayObject() != nil {
		container.RemoveChild(child.DisplayObject())
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
	if c.opaque == value {
		return
	}
	c.opaque = value
	// 参考 TypeScript 原版：GComponent.ts set opaque()
	// opaque = true 时，mouseThrough = false（拦截事件）
	// opaque = false 时，mouseThrough = true（穿透事件）
	if sprite := c.DisplayObject(); sprite != nil {
		sprite.SetMouseThrough(!value)
	}
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

// SetupBeforeAdd parses component-level scroll/mask/overflow metadata.
// 先调用父类处理基础属性，然后处理组件特定属性（mask, hitTest等）
func (c *GComponent) SetupBeforeAdd(buf *utils.ByteBuffer, start int, resolver MaskResolver) {
	if c == nil || buf == nil || start < 0 {
		return
	}

	// 首先调用父类GObject.SetupBeforeAdd处理基础属性（位置、尺寸、旋转等）
	// 这对应TypeScript版本中隐式的继承链调用
	c.GObject.SetupBeforeAdd(buf, start)

	// 然后处理组件特定属性（mask, hitTest等）
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(start, 4) {
		return
	}
	if err := buf.Skip(2); err != nil || buf.Remaining() < 1+2+2+4+4 {
		return
	}
	c.SetOpaque(buf.ReadBool())
	maskIndex := int(buf.ReadInt16())
	reversed := false
	var mask *GObject
	if maskIndex >= 0 {
		if buf.Remaining() > 0 {
			reversed = buf.ReadBool()
		}
		if resolver != nil {
			mask = resolver.MaskChild(maskIndex)
		}
	}
	c.SetMask(mask, reversed)
	hitMode := HitTest{Mode: HitTestModeNone}
	hitID := buf.ReadS()
	offsetX := int(buf.ReadInt32())
	offsetY := int(buf.ReadInt32())
	if hitID != nil && *hitID != "" {
		hitMode = HitTest{Mode: HitTestModePixel, ItemID: *hitID, OffsetX: offsetX, OffsetY: offsetY}
	} else if offsetX != 0 && offsetY != -1 {
		hitMode = HitTest{Mode: HitTestModeChild, OffsetX: offsetX, ChildIndex: offsetY}
	}
	c.SetHitTest(hitMode)
}

// SetupAfterAdd applies scroll/mask hit-test data and transitions.
func (c *GComponent) SetupAfterAdd(buf *utils.ByteBuffer, start int, resolver MaskResolver, pixelResolver PixelHitResolver) {
	if c == nil || buf == nil || start < 0 {
		return
	}
	saved := buf.Pos()
	defer buf.SetPos(saved)
	c.SetupBeforeAdd(buf, start, resolver)
	hit := c.HitTest()
	if pixelResolver != nil {
		var data *assets.PixelHitTestData
		if hit.Mode == HitTestModePixel && hit.ItemID != "" {
			data = pixelResolver.PixelData(hit.ItemID)
		}
		pixelResolver.Configure(c, hit, data)
	}
	c.setupTransitions(buf, start)
}

// getInsertPosForSortingChild finds the correct insertion position for a child with sortingOrder > 0.
// Children with sortingOrder are maintained in ascending order at the end of the children array.
func (c *GComponent) getInsertPosForSortingChild(target *GObject) int {
	if c == nil || target == nil {
		return 0
	}
	cnt := len(c.children)
	for i := 0; i < cnt; i++ {
		child := c.children[i]
		if child == target {
			continue
		}
		if target.sortingOrder < child.sortingOrder {
			return i
		}
	}
	return cnt
}

// childSortingOrderChanged is called when a child's sortingOrder changes.
// It repositions the child within the children array to maintain sorted order.
func (c *GComponent) childSortingOrderChanged(child *GObject, oldValue, newValue int) {
	if c == nil || child == nil {
		return
	}

	// Find current index
	oldIndex := -1
	for i, ch := range c.children {
		if ch == child {
			oldIndex = i
			break
		}
	}
	if oldIndex == -1 {
		return
	}

	// Handle transition from/to sortingOrder = 0
	if newValue == 0 {
		// Child no longer has sortingOrder, move to end of non-sorting section
		c.sortingChildCount--
		newIndex := len(c.children) - c.sortingChildCount - 1
		c.setChildIndex(child, oldIndex, newIndex)
	} else {
		if oldValue == 0 {
			// Child now has sortingOrder for the first time
			c.sortingChildCount++
		}
		// Find new position based on sortingOrder
		newIndex := c.getInsertPosForSortingChild(child)
		if oldIndex < newIndex {
			c.setChildIndex(child, oldIndex, newIndex-1)
		} else {
			c.setChildIndex(child, oldIndex, newIndex)
		}
	}
}

// setChildIndex moves a child from oldIndex to newIndex in the children array.
func (c *GComponent) setChildIndex(child *GObject, oldIndex, newIndex int) {
	if c == nil || child == nil || oldIndex == newIndex {
		return
	}
	if oldIndex < 0 || oldIndex >= len(c.children) {
		return
	}
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= len(c.children) {
		newIndex = len(c.children) - 1
	}

	// Remove from old position
	copy(c.children[oldIndex:], c.children[oldIndex+1:])
	c.children = c.children[:len(c.children)-1]

	// Insert at new position
	c.children = append(c.children, nil)
	copy(c.children[newIndex+1:], c.children[newIndex:])
	c.children[newIndex] = child

	// Update display object order if needed
	if container := c.childContainer(); container != nil && child.DisplayObject() != nil {
		container.RemoveChild(child.DisplayObject())
		container.AddChild(child.DisplayObject())
	}
}

// ViewWidth returns the width of the scrollable view port if available, otherwise the component width.
// 对应 TypeScript 版本的 get viewWidth() 方法 (GComponent.ts:871-876)
func (c *GComponent) ViewWidth() float64 {
	if c == nil {
		return 0
	}
	if c.scrollPane != nil {
		return c.scrollPane.ViewWidth()
	}
	return c.Width()
}

// ViewHeight returns the height of the scrollable view port if available, otherwise the component height.
// 对应 TypeScript 版本的 get viewHeight() 方法 (GComponent.ts:885-890)
func (c *GComponent) ViewHeight() float64 {
	if c == nil {
		return 0
	}
	if c.scrollPane != nil {
		return c.scrollPane.ViewHeight()
	}
	return c.Height()
}

// MaskResolver resolves mask children by index.
type MaskResolver interface {
	MaskChild(index int) *GObject
}

// PixelHitResolver resolves pixel hit data and configures the render system.
type PixelHitResolver interface {
	PixelData(itemID string) *assets.PixelHitTestData
	Configure(comp *GComponent, hit HitTest, data *assets.PixelHitTestData)
}
