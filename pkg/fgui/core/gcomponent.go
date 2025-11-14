package core

import (
	"math"

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
	margin             Margin
	overflow           OverflowType
	boundsChanged      bool // 边界是否需要重新计算（对应 TypeScript _boundsChanged）
	trackBounds        bool // 是否跟踪子对象边界（对应 TypeScript _trackBounds）
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
	// 如果有ScrollPane，应该返回ScrollPane管理的container
	// 因为ScrollPane会将原始container移动到自己的mask容器中
	if c.scrollPane != nil && c.scrollPane.container != nil {
		return c.scrollPane.container
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

// SetupOverflow 配置组件的 overflow 行为
// 参考 TypeScript 版本：GComponent.ts setupOverflow (746-762行)
func (c *GComponent) SetupOverflow(overflow OverflowType) {
	if c == nil {
		return
	}
	c.overflow = overflow

	if overflow == OverflowHidden {
		// 如果 display 和 container 是同一个对象，创建新的 container
		if c.display == c.container {
			c.container = laya.NewSprite()
			c.display.AddChild(c.container)
		}
		c.UpdateMask()
		c.container.SetPosition(float64(c.margin.Left), float64(c.margin.Top))
	} else if c.margin.Left != 0 || c.margin.Top != 0 {
		// 即使不是 Hidden，如果有 margin 也需要独立的 container
		if c.display == c.container {
			c.container = laya.NewSprite()
			c.display.AddChild(c.container)
		}
		c.container.SetPosition(float64(c.margin.Left), float64(c.margin.Top))
	}
}

// UpdateMask 更新裁剪矩形（用于 overflow=hidden）
// 参考 TypeScript 版本：GComponent.ts updateMask (724-734行)
func (c *GComponent) UpdateMask() {
	if c == nil || c.display == nil {
		return
	}

	// 创建裁剪矩形
	rect := &laya.Rect{
		X: float64(c.margin.Left),
		Y: float64(c.margin.Top),
		W: c.width - float64(c.margin.Right),
		H: c.height - float64(c.margin.Bottom),
	}

	// 应用到 displayObject
	c.display.SetScrollRect(rect)
}

// SetMargin 设置组件的边距
func (c *GComponent) SetMargin(margin Margin) {
	if c == nil {
		return
	}
	c.margin = margin
}

// Margin 返回组件的边距
func (c *GComponent) Margin() Margin {
	if c == nil {
		return Margin{}
	}
	return c.margin
}

// Overflow 返回组件的 overflow 类型
func (c *GComponent) Overflow() OverflowType {
	if c == nil {
		return OverflowVisible
	}
	return c.overflow
}

// SetSize overrides GObject.SetSize to同步滚动视图。
// 参考 TypeScript 版本：GComponent.ts handleSizeChanged (764-774行)
func (c *GComponent) SetSize(width, height float64) {
	if c == nil {
		return
	}
	oldWidth := c.Width()
	oldHeight := c.Height()
	c.GObject.SetSize(width, height)

	// 尺寸改变时需要更新
	if width != oldWidth || height != oldHeight {
		if c.scrollPane != nil {
			c.scrollPane.OnOwnerSizeChanged()
		} else if c.display != nil && c.display.ScrollRect() != nil {
			// 如果有 scrollRect（overflow=hidden），更新裁剪区域
			c.UpdateMask()
		}
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

	// 通知边界可能变化（对应 TypeScript GComponent.ts:111）
	c.SetBoundsChangedFlag()
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

	// 通知边界可能变化（对应 TypeScript GComponent.ts:163）
	c.SetBoundsChangedFlag()
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
// 对应 TypeScript 版本的 getChild
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

// GetChild is an alias for ChildByName to match TypeScript API.
// 对应 TypeScript 版本的 getChild
func (c *GComponent) GetChild(name string) *GObject {
	return c.ChildByName(name)
}

// GetChildAt returns the child at the specified index.
// 对应 TypeScript 版本的 getChildAt
func (c *GComponent) GetChildAt(index int) *GObject {
	if index < 0 || index >= len(c.children) {
		return nil
	}
	return c.children[index]
}

// GetController is an alias for ControllerByName to match TypeScript API.
// 对应 TypeScript 版本的 getController
func (c *GComponent) GetController(name string) *Controller {
	return c.ControllerByName(name)
}

// ControllerByName is an alias for GetController.
// 对应 TypeScript 版本的 getController
func (c *GComponent) ControllerByName(name string) *Controller {
	if name == "" {
		return nil
	}
	for _, controller := range c.controllers {
		if controller != nil && controller.Name == name {
			return controller
		}
	}
	return nil
}

// NumChildren returns the number of children.
// 对应 TypeScript 版本的 numChildren
func (c *GComponent) NumChildren() int {
	return len(c.children)
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

// IsChildInView checks if a child object is visible in the viewport.
// This is a helper method for scroll-related functionality.
func (c *GComponent) IsChildInView(child *GObject) bool {
	if c == nil || child == nil || !child.Visible() {
		return false
	}

	// Get child bounds
	childX := child.X()
	childY := child.Y()
	childW := child.Width()
	childH := child.Height()

	// Get view bounds (taking margin into account)
	viewX := float64(c.margin.Left)
	viewY := float64(c.margin.Top)
	viewW := c.width - float64(c.margin.Left+c.margin.Right)
	viewH := c.height - float64(c.margin.Top+c.margin.Bottom)

	// Check if child overlaps with viewport
	return !(childX+childW <= viewX || childX >= viewX+viewW || 
			childY+childH <= viewY || childY >= viewY+viewH)
}

// GetFirstChildInView returns the index of the first child that is visible in the viewport.
func (c *GComponent) GetFirstChildInView() int {
	if c == nil {
		return -1
	}
	
	for i, child := range c.children {
		if c.IsChildInView(child) {
			return i
		}
	}
	return -1
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

	// 应用 mask 到 displayObject（对应 TypeScript GComponent.ts setMask）
	display := c.DisplayObject()
	if display == nil {
		return
	}

	if mask == nil {
		display.SetMask(nil)
		return
	}

	// 获取 mask 对象的 displayObject
	maskSprite := mask.DisplayObject()
	if maskSprite == nil {
		return
	}

	// 设置 mask（暂不支持 reversed，需要 blendMode 支持）
	display.SetMask(maskSprite)
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
	c.SetupTransitions(buf, start)
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

// SetBoundsChangedFlag marks the bounds as needing recalculation.
// 对应 TypeScript: public setBoundsChangedFlag(): void (GComponent.ts:797-807)
func (c *GComponent) SetBoundsChangedFlag() {
	if c == nil {
		return
	}
	// 只有在有 scrollPane 或 trackBounds 时才需要标记
	if c.scrollPane == nil && !c.trackBounds {
		return
	}

	if !c.boundsChanged {
		c.boundsChanged = true
		// TypeScript 使用 Laya.timer.callLater(this, this.ensureBoundsCorrect)
		// Go 版本我们不使用 callLater，而是在需要时同步调用
	}
}

// EnsureBoundsCorrect ensures all child sizes are correct and updates bounds if needed.
// 对应 TypeScript: public ensureBoundsCorrect(): void (GComponent.ts:821-832)
func (c *GComponent) EnsureBoundsCorrect() {
	if c == nil {
		return
	}

	// 确保所有子对象的尺寸正确
	for _, child := range c.children {
		if child != nil {
			child.EnsureSizeCorrect()
		}
	}

	// 如果边界标记为已变化，更新边界
	if c.boundsChanged {
		c.UpdateBounds()
	}
}

// UpdateBounds calculates the bounding box from children and updates content size.
// 对应 TypeScript: protected updateBounds(): void (GComponent.ts:834-862)
func (c *GComponent) UpdateBounds() {
	if c == nil {
		return
	}

	var ax, ay, aw, ah float64

	if len(c.children) > 0 {
		// 初始化为最大/最小值
		ax = math.MaxFloat64
		ay = math.MaxFloat64
		ar := -math.MaxFloat64 // right edge
		ab := -math.MaxFloat64 // bottom edge

		// 遍历所有子对象计算边界
		for _, child := range c.children {
			if child == nil {
				continue
			}

			x := child.X()
			if x < ax {
				ax = x
			}

			y := child.Y()
			if y < ay {
				ay = y
			}

			r := x + child.ActualWidth()
			if r > ar {
				ar = r
			}

			b := y + child.ActualHeight()
			if b > ab {
				ab = b

			}
		}

		aw = ar - ax
		ah = ab - ay
	}

	c.SetBounds(ax, ay, aw, ah)
}

// SetBounds updates the content size for scrollPane if present.
// 对应 TypeScript: public setBounds(ax: number, ay: number, aw: number, ah: number): void (GComponent.ts:864-869)
func (c *GComponent) SetBounds(ax, ay, aw, ah float64) {
	if c == nil {
		return
	}

	c.boundsChanged = false

	// 如果有 ScrollPane，设置内容尺寸
	if c.scrollPane != nil {
		newWidth := math.Round(ax + aw)
		newHeight := math.Round(ay + ah)

		// 关键修复：如果当前 contentSize 已经合理（大于等于 viewSize），
		// 并且新计算的高度明显变小，说明可能是子对象未完全创建（GList 虚拟化等），
		// 此时保留原有的高度，但仍然更新宽度
		currentSize := c.scrollPane.ContentSize()
		viewHeight := c.scrollPane.ViewHeight()

		// 如果满足以下条件，保留当前高度：
		// 1. 当前高度 >= viewHeight（需要滚动条）
		// 2. 新高度 < viewHeight（不需要滚动条）
		// 3. 新高度明显小于当前高度（减少超过 20%）
		if currentSize.Y >= viewHeight &&
			newHeight < viewHeight &&
			newHeight < currentSize.Y*0.8 {
			newHeight = currentSize.Y // 保留原高度
		}

		// TypeScript: this._scrollPane.setContentSize(Math.round(ax + aw), Math.round(ay + ah))
		c.scrollPane.SetContentSize(newWidth, newHeight)
	}
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

// handleGrayedChanged 传播灰色状态到子组件
// 参考 TypeScript 版本：GComponent.ts handleGrayedChanged (776-788行)
func (c *GComponent) handleGrayedChanged() {
	if c == nil {
		return
	}
	// 首先尝试使用 "grayed" 控制器（如果存在）
	grayedController := c.ControllerByName("grayed")
	if grayedController != nil {
		if c.Grayed() {
			grayedController.SetSelectedIndex(1)
		} else {
			grayedController.SetSelectedIndex(0)
		}
		return
	}
	// 如果没有 "grayed" 控制器，将 grayed 状态传播到所有子组件
	v := c.Grayed()
	cnt := len(c.children)
	for i := 0; i < cnt; i++ {
		if child := c.children[i]; child != nil {
			child.SetGrayed(v)
		}
	}
}
