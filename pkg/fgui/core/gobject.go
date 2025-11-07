package core

import (
	"fmt"
	"math"
	"strconv"
	"sync/atomic"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/gears"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// BlendMode represents sprite blending applied during rendering.
type BlendMode int

const (
	BlendModeNormal BlendMode = iota
	BlendModeAdd
)

func blendModeFromByte(value int) BlendMode {
	switch value {
	case 2:
		return BlendModeAdd
	default:
		return BlendModeNormal
	}
}

var gObjectCounter uint64

// GObject is the base building block for all FairyGUI entities.
type GObject struct {
	id         string
	name       string
	resourceID string
	display    *laya.Sprite

	x      float64
	y      float64
	width  float64
	height float64
	scaleX float64
	scaleY float64

	parent             *GComponent
	alpha              float64
	visible            bool
	internalVisible    bool // gear-controlled visibility (separate from user-controlled visible)
	touchable          bool
	grayed             bool
	rotation           float64
	skewX              float64
	skewY              float64
	pivotX             float64
	pivotY             float64
	pivotAsAnchor      bool
	data               any
	relations          *Relations
	dependents         []*RelationItem
	gears              [gears.SlotCount]gears.Gear
	gearLocked         bool
	handlingController bool
	rawWidth           float64
	rawHeight          float64
	sourceWidth        float64
	sourceHeight       float64
	initWidth          float64
	initHeight         float64
	minWidth           float64
	maxWidth           float64
	minHeight          float64
	maxHeight          float64
	props              map[gears.ObjectPropID]any
	tooltips           string
	group              *GObject
	colorFilter        [4]float64
	colorFilterEnabled bool
	shakeOffsetX       float64
	shakeOffsetY       float64
	customData         string
	blendMode          BlendMode
	sortingOrder       int // Z-order for rendering and interaction (0 = normal order)
}

type ownerSizeChanged interface {
	OwnerSizeChanged(oldWidth, oldHeight float64)
}

// NewGObject creates a base object with a backing sprite.
func NewGObject() *GObject {
	counter := atomic.AddUint64(&gObjectCounter, 1)
	display := laya.NewSprite()
	obj := &GObject{
		id:              fmt.Sprintf("gobj-%d", counter),
		display:         display,
		alpha:           1.0,
		visible:         true,
		internalVisible: true, // gear-controlled visibility默认为true
		touchable:       true,
		scaleX:          1.0,
		scaleY:          1.0,
		props:           make(map[gears.ObjectPropID]any),
	}
	display.SetOwner(obj)
	display.SetMouseEnabled(true)
	obj.relations = NewRelations(obj)
	return obj
}

// ID returns the unique identifier.
func (g *GObject) ID() string {
	return g.id
}

// Name returns the display name.
func (g *GObject) Name() string {
	return g.name
}

// ResourceID returns the package-specific identifier assigned by the builder.
func (g *GObject) ResourceID() string {
	return g.resourceID
}

// Parent returns the parent component, if any.
func (g *GObject) Parent() *GComponent {
	return g.parent
}

// SetName updates the display name.
func (g *GObject) SetName(name string) {
	g.name = name
	if g.display != nil {
		g.display.SetName(name)
	}
}

// SetResourceID records the package-scoped identifier (e.g., "n1_rftu").
func (g *GObject) SetResourceID(id string) {
	g.resourceID = id
}

// DisplayObject exposes the underlying compat sprite.
func (g *GObject) DisplayObject() *laya.Sprite {
	return g.display
}

// SetPosition moves the object within its parent coordinate space.
func (g *GObject) SetPosition(x, y float64) {
	dx := x - g.x
	dy := y - g.y
	if dx == 0 && dy == 0 {
		return
	}
	g.x = x
	g.y = y
	g.refreshTransform()

	// 如果是 GGroup，需要同步移动子元素
	// 参考 TypeScript 版本 GObject.ts setXY (97-110行)
	// if (this instanceof GGroup) this.moveChildren(dx, dy);
	if groupMover, ok := g.data.(interface{ MoveChildren(dx, dy float64) }); ok {
		groupMover.MoveChildren(dx, dy)
	}

	g.updateGear(gears.IndexXY)
	g.notifyDependentsXY(dx, dy)
}

// SetSize updates width and height.
func (g *GObject) SetSize(width, height float64) {
	oldWidth := g.width
	oldHeight := g.height
	g.width = width
	g.height = height
	g.rawWidth = width
	g.rawHeight = height
	if g.display != nil {
		g.display.SetSize(width, height)
	}
	g.refreshTransform()
	dw := width - oldWidth
	dh := height - oldHeight
	g.updateGear(gears.IndexSize)
	if (dw != 0 || dh != 0) && g.relations != nil {
		g.relations.OnOwnerSizeChanged(dw, dh, g.pivotAsAnchor)
		// 通知父组件边界可能变化（对应 TypeScript GObject.ts:228）
		if g.parent != nil {
			g.parent.SetBoundsChangedFlag()
		}
		// 通知组 (group) 边界可能变化
		if g.group != nil {
			if groupComp, ok := g.group.data.(*GComponent); ok {
				groupComp.SetBoundsChangedFlag()
			}
		}
	}
	g.notifyDependentsSize(dw, dh)

	if handler, ok := g.data.(ownerSizeChanged); ok {

		handler.OwnerSizeChanged(oldWidth, oldHeight)
	}
}

// SetMinSize stores the minimum width and height constraints.
func (g *GObject) SetMinSize(minWidth, minHeight float64) {
	if g == nil {
		return
	}
	if minWidth < 0 {
		minWidth = 0
	}
	if minHeight < 0 {
		minHeight = 0
	}
	g.minWidth = minWidth
	g.minHeight = minHeight
}

// MinSize returns the minimum width and height constraints.
func (g *GObject) MinSize() (float64, float64) {
	if g == nil {
		return 0, 0
	}
	return g.minWidth, g.minHeight
}

// SetMaxSize stores the maximum width and height constraints (0 means no limit).
func (g *GObject) SetMaxSize(maxWidth, maxHeight float64) {
	if g == nil {
		return
	}
	if maxWidth < 0 {
		maxWidth = 0
	}
	if maxHeight < 0 {
		maxHeight = 0
	}
	g.maxWidth = maxWidth
	g.maxHeight = maxHeight
}

// MaxSize returns the maximum width and height constraints (0 means no limit).
func (g *GObject) MaxSize() (float64, float64) {
	if g == nil {
		return 0, 0
	}
	return g.maxWidth, g.maxHeight
}

// SetScale updates the scaling factors on both axes.
func (g *GObject) SetScale(scaleX, scaleY float64) {
	g.scaleX = scaleX
	g.scaleY = scaleY
	if g.display != nil {
		g.display.SetScale(scaleX, scaleY)
	}
	g.refreshTransform()
	g.updateGear(gears.IndexSize)
}

// Scale returns the scaling factors.
func (g *GObject) Scale() (float64, float64) {
	return g.scaleX, g.scaleY
}

// SetRotation stores the rotation in degrees and mirrors it to the sprite.
// Note: FairyGUI uses degrees, not radians (consistent with TypeScript version).
func (g *GObject) SetRotation(degrees float64) {
	g.rotation = degrees
	if g.display != nil {
		g.display.SetRotation(degrees)
	}
	g.refreshTransform()
}

// Rotation returns the current rotation in degrees.
// Note: FairyGUI uses degrees, not radians (consistent with TypeScript version).
func (g *GObject) Rotation() float64 {
	return g.rotation
}

// SetSkew updates skew factors in degrees (FairyGUI uses degrees, not radians).
// Note: Consistent with TypeScript version, skew values are in degrees.
func (g *GObject) SetSkew(skewX, skewY float64) {
	g.skewX = skewX
	g.skewY = skewY
	if g.display != nil {
		g.display.SetSkew(-skewX, skewY)
	}
	g.refreshTransform()
}

// Skew returns the skew factors in degrees.
// Note: FairyGUI uses degrees, not radians (consistent with TypeScript version).
func (g *GObject) Skew() (float64, float64) {
	return g.skewX, g.skewY
}

// SetPivot configures the normalized pivot point.
func (g *GObject) SetPivot(px, py float64) {
	g.SetPivotWithAnchor(px, py, false)
}

// SetPivotWithAnchor configures the pivot and whether it acts as an anchor.
func (g *GObject) SetPivotWithAnchor(px, py float64, asAnchor bool) {
	g.pivotX = px
	g.pivotY = py
	g.pivotAsAnchor = asAnchor
	if g.display != nil {
		g.display.SetPivotWithAnchor(px, py, asAnchor)
	}
	g.refreshTransform()
}

// Pivot returns the normalized pivot point.
func (g *GObject) Pivot() (float64, float64) {
	return g.pivotX, g.pivotY
}

// PivotAsAnchor reports whether the pivot acts as an anchor.
func (g *GObject) PivotAsAnchor() bool {
	return g.pivotAsAnchor
}

// SetAlpha adjusts transparency.
func (g *GObject) SetAlpha(alpha float64) {
	g.alpha = alpha
	if g.display != nil {
		g.display.SetAlpha(alpha)
	}
}

// Alpha returns the current alpha value.
func (g *GObject) Alpha() float64 {
	return g.alpha
}

// SetVisible toggles visibility.
func (g *GObject) SetVisible(visible bool) {
	if g.visible == visible {
		return
	}
	g.visible = visible
	// 检查 data 是否实现了自定义的 HandleVisibleChanged
	// 这样可以让 GGroup 等子类覆盖行为
	if handler, ok := g.data.(interface{ HandleVisibleChanged() }); ok {
		handler.HandleVisibleChanged()
	} else {
		g.HandleVisibleChanged()
	}
}

// Visible reports whether the object is actually visible.
// This considers both user-controlled visibility and gear-controlled visibility.
// 参考 TypeScript 版本 GObject.ts internalVisible getter
func (g *GObject) Visible() bool {
	if !g.visible {
		return false
	}
	if !g.internalVisible {
		return false
	}
	// 检查 group 可见性
	if g.group != nil && g.group.display != nil {
		return g.group.display.Visible()
	}
	return true
}

// HandleVisibleChanged 更新 displayObject 的可见性
// 参考 TypeScript 版本 GObject.ts internalVisible getter (455-469行)
// 最终可见性 = internalVisible(gear控制) && visible(用户控制) && group可见性
func (g *GObject) HandleVisibleChanged() {
	if g == nil || g.display == nil {
		return
	}
	// 计算最终可见性：internalVisible && visible && group可见性
	finalVisible := g.internalVisible && g.visible
	if finalVisible && g.group != nil {
		// group 的实际可见性 = group.internalVisible && group.visible
		// 参考 TypeScript 版本 GObject.ts internalVisible getter
		if g.group.display != nil {
			finalVisible = g.group.display.Visible()
		}
	}
	g.display.SetVisible(finalVisible)
}

// isGroupVisible 递归检查 group 链的可见性
func (g *GObject) isGroupVisible(group *GObject) bool {
	if group == nil {
		return true
	}
	if !group.visible {
		return false
	}
	if group.group != nil {
		return g.isGroupVisible(group.group)
	}
	return true
}

// Touchable returns whether the object reacts to user input.
func (g *GObject) Touchable() bool {
	return g.touchable
}

// SetTouchable updates the touchable flag.
func (g *GObject) SetTouchable(touchable bool) {
	g.touchable = touchable
	if g.display != nil {
		g.display.SetMouseEnabled(touchable)
	}
}

// On registers an event listener against the underlying display object.
func (g *GObject) On(evt laya.EventType, fn laya.Listener) {
	if g == nil || g.display == nil || fn == nil {
		return
	}
	g.display.Dispatcher().On(evt, fn)
}

// Once registers a one-shot listener against the underlying display object.
func (g *GObject) Once(evt laya.EventType, fn laya.Listener) {
	if g == nil || g.display == nil || fn == nil {
		return
	}
	g.display.Dispatcher().Once(evt, fn)
}

// Off removes a previously registered listener.
func (g *GObject) Off(evt laya.EventType, fn laya.Listener) {
	if g == nil || g.display == nil || fn == nil {
		return
	}
	g.display.Dispatcher().Off(evt, fn)
}

// Emit dispatches an event from the underlying display object.
func (g *GObject) Emit(evt laya.EventType, data any) {
	if g == nil || g.display == nil {
		return
	}
	g.display.Dispatcher().Emit(evt, data)
}

// Grayed reports whether the object is displayed in a grayed state.
func (g *GObject) Grayed() bool {
	return g.grayed
}

// SetGrayed toggles the grayed state.
func (g *GObject) SetGrayed(value bool) {
	g.grayed = value
	if g.display != nil {
		g.display.SetGray(value)
	}
}

// SetColorFilter 记录颜色滤镜参数（用于 Transition 等场景）。
func (g *GObject) SetColorFilter(r, gFactor, b, a float64) {
	g.colorFilter = [4]float64{r, gFactor, b, a}
	g.colorFilterEnabled = !(math.Abs(r) <= 1e-6 && math.Abs(gFactor) <= 1e-6 && math.Abs(b) <= 1e-6 && math.Abs(a) <= 1e-6)
	if g.display != nil {
		g.display.SetColorFilter(r, gFactor, b, a)
	}
}

// ClearColorFilter 移除当前颜色滤镜。
func (g *GObject) ClearColorFilter() {
	g.colorFilter = [4]float64{}
	g.colorFilterEnabled = false
	if g.display != nil {
		g.display.ClearColorFilter()
	}
}

// ColorFilter 返回颜色滤镜参数及是否启用。
func (g *GObject) ColorFilter() (enabled bool, values [4]float64) {
	return g.colorFilterEnabled, g.colorFilter
}

// SetBlendMode updates the rendering blend mode for the object.
func (g *GObject) SetBlendMode(mode BlendMode) {
	g.blendMode = mode
	if g.display != nil {
		g.display.SetBlendMode(laya.BlendMode(mode))
	}
}

// BlendMode returns the current blend mode applied to this object.
func (g *GObject) BlendMode() BlendMode {
	return g.blendMode
}

// X returns the local X position.
func (g *GObject) X() float64 {
	return g.x
}

// Y returns the local Y position.
func (g *GObject) Y() float64 {
	return g.y
}

// Width returns the current width.
func (g *GObject) Width() float64 {
	g.EnsureSizeCorrect()
	if g.relations != nil && g.relations.sizeDirty {
		g.relations.EnsureRelationsSizeCorrect()
	}
	return g.width
}

// Height returns the current height.
func (g *GObject) Height() float64 {
	g.EnsureSizeCorrect()
	if g.relations != nil && g.relations.sizeDirty {
		g.relations.EnsureRelationsSizeCorrect()
	}
	return g.height
}

// ActualWidth returns the width taking scale into account.
// 对应 TypeScript: get actualWidth(): number { return this.width * Math.abs(this._scaleX); }
func (g *GObject) ActualWidth() float64 {
	return g.Width() * math.Abs(g.scaleX)
}

// ActualHeight returns the height taking scale into account.
// 对应 TypeScript: get actualHeight(): number { return this.height * Math.abs(this._scaleY); }
func (g *GObject) ActualHeight() float64 {
	return g.Height() * math.Abs(g.scaleY)
}

// EnsureSizeCorrect ensures the size is up-to-date.
// This is a hook method that can be overridden by subclasses (e.g., GTextField for AutoSize).
// 对应 TypeScript: public ensureSizeCorrect(): void {}
func (g *GObject) EnsureSizeCorrect() {
	// Base implementation is empty - subclasses can override
}

// ParentSize reports the dimensions of the parent component, when available.
func (g *GObject) ParentSize() (float64, float64) {
	if g == nil || g.parent == nil {
		return 0, 0
	}
	return g.parent.Width(), g.parent.Height()
}

// RawWidth returns the unclamped width recorded during the last SetSize call.
func (g *GObject) RawWidth() float64 {
	return g.rawWidth
}

// RawHeight returns the unclamped height recorded during the last SetSize call.
func (g *GObject) RawHeight() float64 {
	return g.rawHeight
}

// SetSourceSize stores the design-time size declared by the package.
func (g *GObject) SetSourceSize(width, height float64) {
	g.sourceWidth = width
	g.sourceHeight = height
}

// SourceSize returns the source width/height pair.
func (g *GObject) SourceSize() (float64, float64) {
	return g.sourceWidth, g.sourceHeight
}

// SourceWidth returns the stored source width.
func (g *GObject) SourceWidth() float64 {
	return g.sourceWidth
}

// SourceHeight returns the stored source height.
func (g *GObject) SourceHeight() float64 {
	return g.sourceHeight
}

// SetInitSize stores the initial size used when constructing the object.
func (g *GObject) SetInitSize(width, height float64) {
	g.initWidth = width
	g.initHeight = height
}

// InitSize returns the stored initial width/height pair.
func (g *GObject) InitSize() (float64, float64) {
	return g.initWidth, g.initHeight
}

// InitWidth returns the stored initial width.
func (g *GObject) InitWidth() float64 {
	return g.initWidth
}

// InitHeight returns the stored initial height.
func (g *GObject) InitHeight() float64 {
	return g.initHeight
}

func (g *GObject) ensureProps() {
	if g.props == nil {
		g.props = make(map[gears.ObjectPropID]any)
	}
}

// GetProp retrieves the property value referenced by the provided identifier.
func (g *GObject) GetProp(id gears.ObjectPropID) any {
	if g == nil {
		return nil
	}
	data := g.Data()
	switch id {
	case gears.ObjectPropIDText:
		switch v := data.(type) {
		case textAccessor:
			return v.Text()
		case titleAccessor:
			return v.Title()
		}
	case gears.ObjectPropIDIcon:
		if v, ok := data.(iconAccessor); ok {
			return v.Icon()
		}
	case gears.ObjectPropIDColor:
		switch v := data.(type) {
		case colorAccessor:
			return v.Color()
		case titleColorAccessor:
			return v.TitleColor()
		}
	case gears.ObjectPropIDOutlineColor:
		switch v := data.(type) {
		case outlineColorAccessor:
			return v.OutlineColor()
		case titleOutlineColorAccessor:
			return v.TitleOutlineColor()
		}
	case gears.ObjectPropIDFontSize:
		switch v := data.(type) {
		case fontSizeAccessor:
			return v.FontSize()
		case titleFontSizeAccessor:
			return v.TitleFontSize()
		}
	case gears.ObjectPropIDSelected:
		if v, ok := data.(selectedAccessor); ok {
			return v.Selected()
		}
	case gears.ObjectPropIDPlaying:
		if v, ok := data.(playingAccessor); ok {
			return v.Playing()
		}
	case gears.ObjectPropIDFrame:
		switch v := data.(type) {
		case frameAccessor:
			return v.Frame()
		}
	case gears.ObjectPropIDTimeScale:
		if v, ok := data.(timeScaleAccessor); ok {
			return v.TimeScale()
		}
	case gears.ObjectPropIDDeltaTime:
		if v, ok := data.(deltaTimeAccessor); ok {
			return v.DeltaTime()
		}
	default:
		return g.props[id]
	}
	if g.props != nil {
		if val, ok := g.props[id]; ok {
			return val
		}
	}
	switch id {
	case gears.ObjectPropIDText, gears.ObjectPropIDIcon, gears.ObjectPropIDColor, gears.ObjectPropIDOutlineColor:
		return ""
	case gears.ObjectPropIDSelected, gears.ObjectPropIDPlaying:
		return false
	case gears.ObjectPropIDFontSize, gears.ObjectPropIDFrame:
		return 0
	case gears.ObjectPropIDTimeScale:
		return 1.0
	case gears.ObjectPropIDDeltaTime:
		return 0.0
	default:
		return nil
	}
}

// SetProp updates the property referenced by the provided identifier.
func (g *GObject) SetProp(id gears.ObjectPropID, value any) {
	if g == nil {
		return
	}
	g.ensureProps()
	data := g.Data()
	switch id {
	case gears.ObjectPropIDText:
		str := toString(value)
		switch v := data.(type) {
		case textAccessor:
			v.SetText(str)
		case titleAccessor:
			v.SetTitle(str)
		}
		g.props[id] = str
	case gears.ObjectPropIDIcon:
		str := toString(value)
		if v, ok := data.(iconAccessor); ok {
			v.SetIcon(str)
		}
		g.props[id] = str
	case gears.ObjectPropIDColor:
		str := toString(value)
		switch v := data.(type) {
		case colorAccessor:
			v.SetColor(str)
		case titleColorAccessor:
			v.SetTitleColor(str)
		}
		g.props[id] = str
	case gears.ObjectPropIDOutlineColor:
		str := toString(value)
		switch v := data.(type) {
		case outlineColorAccessor:
			v.SetOutlineColor(str)
		case titleOutlineColorAccessor:
			v.SetTitleOutlineColor(str)
		}
		g.props[id] = str
	case gears.ObjectPropIDFontSize:
		size, ok := toInt(value)
		if !ok {
			size = 0
		}
		switch v := data.(type) {
		case fontSizeAccessor:
			v.SetFontSize(size)
		case titleFontSizeAccessor:
			v.SetTitleFontSize(size)
		}
		g.props[id] = size
	case gears.ObjectPropIDSelected:
		selected := toBool(value)
		if v, ok := data.(selectedAccessor); ok {
			v.SetSelected(selected)
		}
		g.props[id] = selected
	case gears.ObjectPropIDPlaying:
		playing := toBool(value)
		if v, ok := data.(playingAccessor); ok {
			v.SetPlaying(playing)
		}
		g.props[id] = playing
	case gears.ObjectPropIDFrame:
		frame, ok := toInt(value)
		if !ok {
			frame = 0
		}
		if v, ok := data.(frameAccessor); ok {
			v.SetFrame(frame)
		}
		g.props[id] = frame
	case gears.ObjectPropIDTimeScale:
		scale, ok := toFloat(value)
		if !ok {
			scale = 1
		}
		if v, ok := data.(timeScaleAccessor); ok {
			v.SetTimeScale(scale)
		}
		g.props[id] = scale
	case gears.ObjectPropIDDeltaTime:
		delta, ok := toFloat(value)
		if !ok {
			delta = 0
		}
		if v, ok := data.(deltaTimeAccessor); ok {
			v.SetDeltaTime(delta)
		}
		g.props[id] = delta
	default:
		g.props[id] = value
	}
}

func (g *GObject) applyShake(deltaX, deltaY float64) {
	g.SetPosition(g.X()-g.shakeOffsetX+deltaX, g.Y()-g.shakeOffsetY+deltaY)
	g.shakeOffsetX = deltaX
	g.shakeOffsetY = deltaY
}

func (g *GObject) clearShake() {
	if g.shakeOffsetX != 0 || g.shakeOffsetY != 0 {
		g.SetPosition(g.X()-g.shakeOffsetX, g.Y()-g.shakeOffsetY)
		g.shakeOffsetX = 0
		g.shakeOffsetY = 0
	}
}

func (g *GObject) xMin() float64 {
	if g == nil {
		return 0
	}
	if g.pivotAsAnchor {
		return g.x - g.width*g.pivotX
	}
	return g.x
}

func (g *GObject) setXMin(value float64) {
	if g == nil {
		return
	}
	if g.pivotAsAnchor {
		g.SetPosition(value+g.width*g.pivotX, g.y)
	} else {
		g.SetPosition(value, g.y)
	}
}

func (g *GObject) yMin() float64 {
	if g == nil {
		return 0
	}
	if g.pivotAsAnchor {
		return g.y - g.height*g.pivotY
	}
	return g.y
}

func (g *GObject) setYMin(value float64) {
	if g == nil {
		return
	}
	if g.pivotAsAnchor {
		g.SetPosition(g.x, value+g.height*g.pivotY)
	} else {
		g.SetPosition(g.x, value)
	}
}

// SetData assigns arbitrary user data to the object.
func (g *GObject) SetData(value any) {
	g.data = value
}

// Data returns the user data associated with the object.
func (g *GObject) Data() any {
	return g.data
}

// Tooltips returns the tooltip text associated with this object.
func (g *GObject) Tooltips() string {
	if g == nil {
		return ""
	}
	return g.tooltips
}

// SetTooltips updates the tooltip text stored on this object.
func (g *GObject) SetTooltips(value string) {
	if g == nil {
		return
	}
	g.tooltips = value
}

// SetCustomData stores arbitrary package-defined metadata string.
func (g *GObject) SetCustomData(value string) {
	if g == nil {
		return
	}
	g.customData = value
}

// CustomData returns the package-defined metadata string.
func (g *GObject) CustomData() string {
	if g == nil {
		return ""
	}
	return g.customData
}

// SetupBeforeAdd 从buffer读取并应用基础属性
// 完全对应 TypeScript 版本 GObject.setup_beforeAdd (GObject.ts:985-1056)
// 这是组件构建的核心方法，在第一次添加到显示列表前调用
func (g *GObject) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if g == nil || buf == nil {
		return
	}

	// 保存当前位置，函数结束时恢复（避免影响后续读取）
	saved := buf.Pos()
	defer func() {
		_ = buf.SetPos(saved)
	}()

	// ts: buffer.seek(beginPos, 0); buffer.skip(5);
	if !buf.Seek(beginPos, 0) {
		return
	}
	if err := buf.Skip(5); err != nil {
		return
	}

	// ts: this._id = buffer.readS(); this._name = buffer.readS();
	if id := buf.ReadS(); id != nil {
		g.resourceID = *id
	} else {
		g.resourceID = ""
	}
	if name := buf.ReadS(); name != nil {
		g.name = *name
	} else {
		g.name = ""
	}

	// ts: f1 = buffer.getInt32(); f2 = buffer.getInt32(); this.setXY(f1, f2);
	x := float64(buf.ReadInt32())
	y := float64(buf.ReadInt32())
	g.SetPosition(x, y)

	// ts: if (buffer.readBool()) { this.initWidth = ...; this.setSize(...); }
	if buf.ReadBool() {
		g.initWidth = float64(buf.ReadInt32())
		g.initHeight = float64(buf.ReadInt32())
		g.SetSize(g.initWidth, g.initHeight)
	}

	// ts: if (buffer.readBool()) { this.minWidth = ...; }
	if buf.ReadBool() {
		minW := float64(buf.ReadInt32())
		maxW := float64(buf.ReadInt32())
		minH := float64(buf.ReadInt32())
		maxH := float64(buf.ReadInt32())
		g.SetMinSize(minW, minH)
		g.SetMaxSize(maxW, maxH)
	}

	// ts: if (buffer.readBool()) { this.setScale(f1, f2); }
	if buf.ReadBool() {
		scaleX := float64(buf.ReadFloat32())
		scaleY := float64(buf.ReadFloat32())
		g.SetScale(scaleX, scaleY)
	}

	// ts: if (buffer.readBool()) { this.setSkew(f1, f2); }
	if buf.ReadBool() {
		skewX := float64(buf.ReadFloat32())
		skewY := float64(buf.ReadFloat32())
		g.SetSkew(skewX, skewY)
	}

	// ts: if (buffer.readBool()) { this.setPivot(f1, f2, buffer.readBool()); }
	if buf.ReadBool() {
		pivotX := float64(buf.ReadFloat32())
		pivotY := float64(buf.ReadFloat32())
		asAnchor := buf.ReadBool()
		g.SetPivotWithAnchor(pivotX, pivotY, asAnchor)
	}

	// ts: f1 = buffer.getFloat32(); if (f1 != 1) this.alpha = f1;
	alpha := float64(buf.ReadFloat32())
	if alpha != 1.0 {
		g.SetAlpha(alpha)
	}

	// ts: f1 = buffer.getFloat32(); if (f1 != 0) this.rotation = f1;
	rotation := float64(buf.ReadFloat32())
	if rotation != 0 {
		g.SetRotation(rotation)
	}

	// ts: if (!buffer.readBool()) this.visible = false;
	if !buf.ReadBool() {
		g.SetVisible(false)
	}

	// ts: if (!buffer.readBool()) this.touchable = false;
	if !buf.ReadBool() {
		g.SetTouchable(false)
	}

	// ts: if (buffer.readBool()) this.grayed = true;
	if buf.ReadBool() {
		g.SetGrayed(true)
	}

	// ts: var bm = buffer.readByte(); if (BlendMode[bm]) this.blendMode = BlendMode[bm];
	bm := buf.ReadByte()
	if bm != 0 {
		g.SetBlendMode(blendModeFromByte(int(bm)))
	}

	// ts: var filter = buffer.readByte(); if (filter == 1) { ... }
	filter := int(buf.ReadByte())
	if filter == 1 {
		r := buf.ReadFloat32()
		gr := buf.ReadFloat32()
		b := buf.ReadFloat32()
		a := buf.ReadFloat32()
		g.colorFilter = [4]float64{float64(r), float64(gr), float64(b), float64(a)}
		g.colorFilterEnabled = true
	}

	// ts: var str = buffer.readS(); if (str != null) this.data = str;
	if customData := buf.ReadS(); customData != nil {
		g.SetCustomData(*customData)
	}
}

// ApplyComponentChild copies common transform/appearance data from component child metadata.
func (g *GObject) ApplyComponentChild(child *assets.ComponentChild) {
	if g == nil || child == nil {
		return
	}
	if child.Name != "" {
		g.SetName(child.Name)
	}
	g.SetPosition(float64(child.X), float64(child.Y))
	// 对应 TypeScript 版本 GObject.ts:998-1001
	// 只有当buffer中明确设置了尺寸时才调用SetSize
	// 避免对自动尺寸组件调用SetSize(0, 0)
	if child.Width > 0 || child.Height > 0 {
		g.SetSize(float64(child.Width), float64(child.Height))
	}
	g.SetMinSize(float64(child.MinWidth), float64(child.MinHeight))
	g.SetMaxSize(float64(child.MaxWidth), float64(child.MaxHeight))
	g.SetScale(float64(child.ScaleX), float64(child.ScaleY))
	g.SetSkew(float64(child.SkewX)*math.Pi/180.0, float64(child.SkewY)*math.Pi/180.0)
	if child.PivotAnchor || child.PivotX != 0 || child.PivotY != 0 {
		g.SetPivotWithAnchor(float64(child.PivotX), float64(child.PivotY), child.PivotAnchor)
	}
	g.SetAlpha(float64(child.Alpha))
	g.SetRotation(float64(child.Rotation) * math.Pi / 180.0)
	g.SetVisible(child.Visible)
	g.SetTouchable(child.Touchable)
	g.SetGrayed(child.Grayed)
	g.SetCustomData(child.Data)
	g.SetBlendMode(blendModeFromByte(child.BlendMode))
}

// SetupAfterAdd mirrors FairyGUI 的 setup_afterAdd 逻辑，读取提示、分组和控制器默认页等信息。
func (g *GObject) SetupAfterAdd(parent *GComponent, buf *utils.ByteBuffer, start int) {
	if g == nil || buf == nil || start < 0 {
		return
	}
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if buf.Seek(start, 1) {
		if buf.Remaining() >= 2 {
			if tip := buf.ReadS(); tip != nil {
				g.SetTooltips(*tip)
			}
		}
		if buf.Remaining() >= 2 {
			groupIndex := int(buf.ReadInt16())
			if groupIndex >= 0 && parent != nil {
				if groupObj := parent.ChildAt(groupIndex); groupObj != nil {
					g.SetGroup(groupObj)
				}
			}
		}
	}
	component := componentFromObject(g)
	if component == nil && parent != nil && parent.GObject == g {
		component = parent
	}
	if component != nil && buf.Seek(start, 4) {
		if buf.Remaining() >= 2 {
			_ = buf.ReadInt16() // scroll pane page controller index (未实现)
		}
		if buf.Remaining() >= 2 {
			cnt := int(buf.ReadInt16())
			for i := 0; i < cnt; i++ {
				if buf.Remaining() < 4 {
					break
				}
				ctrlName := buf.ReadS()
				pageID := buf.ReadS()
				if ctrlName == nil || pageID == nil {
					continue
				}
				if ctrl := component.ControllerByName(*ctrlName); ctrl != nil {
					ctrl.SetSelectedPageID(*pageID)
				}
			}
		}
		if buf.Version >= 2 && buf.Remaining() >= 2 {
			assignments := int(buf.ReadInt16())
			for i := 0; i < assignments; i++ {
				if buf.Remaining() < 6 {
					break
				}
				targetPath := buf.ReadS()
				propID := buf.ReadInt16()
				value := buf.ReadS()
				if targetPath == nil || value == nil {
					continue
				}
				if target := FindChildByPath(component, *targetPath); target != nil {
					target.SetProp(gears.ObjectPropID(propID), *value)
				}
			}
		}
	}
}

// Group returns the group object that owns this node, if any.
func (g *GObject) Group() *GObject {
	if g == nil {
		return nil
	}
	return g.group
}

// SetGroup assigns the group associated with this object.
func (g *GObject) SetGroup(group *GObject) {
	if g == nil {
		return
	}
	g.group = group
	// 设置 group 后，需要重新计算可见性
	// 因为实际可见性 = 自己的 visible && group 的可见性
	g.HandleVisibleChanged()
}

// Relations returns the relation set associated with this object.
func (g *GObject) Relations() *Relations {
	if g == nil {
		return nil
	}
	return g.relations
}

// AddRelation registers a relation between this object and the target one.
func (g *GObject) AddRelation(target *GObject, relation RelationType, usePercent bool) {
	if g == nil {
		return
	}
	if g.relations == nil {
		g.relations = NewRelations(g)
	}
	g.relations.Add(target, relation, usePercent)
}

// RemoveRelation removes the relation from this object to the target.
func (g *GObject) RemoveRelation(target *GObject, relation RelationType) {
	if g == nil || g.relations == nil {
		return
	}
	g.relations.Remove(target, relation)
}

// RemoveRelations clears all relations referencing the target object.
func (g *GObject) RemoveRelations(target *GObject) {
	if g == nil || g.relations == nil {
		return
	}
	g.relations.ClearFor(target)
}

func (g *GObject) addRelationDependent(item *RelationItem) {
	if g == nil || item == nil {
		return
	}
	g.dependents = append(g.dependents, item)
}

func (g *GObject) removeRelationDependent(item *RelationItem) {
	if g == nil || item == nil {
		return
	}
	for i, dep := range g.dependents {
		if dep == item {
			g.dependents = append(g.dependents[:i], g.dependents[i+1:]...)
			break
		}
	}
}

func (g *GObject) notifyDependentsXY(dx, dy float64) {
	if g == nil || len(g.dependents) == 0 || (dx == 0 && dy == 0) {
		return
	}
	snapshot := append([]*RelationItem(nil), g.dependents...)
	for _, dep := range snapshot {
		if dep == nil || dep.owner == nil {
			continue
		}
		if rel := dep.owner.Relations(); rel != nil && rel.handling == g {
			continue
		}
		dep.onTargetXYChanged(dx, dy)
	}
}

func (g *GObject) notifyDependentsSize(dw, dh float64) {
	if g == nil || len(g.dependents) == 0 || (dw == 0 && dh == 0) {
		return
	}
	snapshot := append([]*RelationItem(nil), g.dependents...)
	for _, dep := range snapshot {
		if dep == nil || dep.owner == nil {
			continue
		}
		if rel := dep.owner.Relations(); rel != nil && rel.handling == g {
			continue
		}
		dep.onTargetSizeChanged(dw, dh)
	}
}

func (g *GObject) updateGearFromRelationsSafe(index int, dx, dy float64) {
	if g == nil || index < 0 || index >= gears.SlotCount {
		return
	}
	gear := g.gears[index]
	if gear == nil {
		return
	}
	gear.UpdateFromRelations(dx, dy)
}

// refreshTransform updates the display object's position based on current transform state.
func (g *GObject) refreshTransform() {
	if g.display == nil {
		return
	}
	// 直接传递位置给Sprite，不做任何调整
	// Sprite的localMatrix会根据pivotAsAnchor正确解释这个位置的含义
	g.display.SetPosition(g.x, g.y)
}

// SetGearLocked toggles the guard that prevents recursive gear updates.
func (g *GObject) SetGearLocked(locked bool) {
	if g == nil {
		return
	}
	g.gearLocked = locked
}

// GearLocked reports whether the object is currently suppressing gear updates.
func (g *GObject) GearLocked() bool {
	if g == nil {
		return false
	}
	return g.gearLocked
}

// GetGear fetches (and lazily creates) the gear stored at the given index.
func (g *GObject) GetGear(index int) gears.Gear {
	if g == nil || index < 0 || index >= gears.SlotCount {
		return nil
	}
	gear := g.gears[index]
	if gear == nil {
		gear = gears.Create(g, index)
		if gear == nil {
			return nil
		}
		g.gears[index] = gear
		gear.UpdateState()
	}
	return gear
}

func (g *GObject) updateGear(index int) {
	if g == nil || g.gearLocked || index < 0 || index >= gears.SlotCount {
		return
	}
	if gear := g.gears[index]; gear != nil && gear.Controller() != nil {
		gear.UpdateState()
	}
}

// HandleControllerChanged applies the gear state associated with the specified controller.
func (g *GObject) HandleControllerChanged(ctrl *Controller) {
	if g == nil || ctrl == nil {
		return
	}
	if g.handlingController {
		return
	}

	g.handlingController = true
	for _, gear := range g.gears {
		if gear != nil && gear.Controller() == ctrl {
			gear.Apply()
		}
	}
	g.handlingController = false

	// 在控制器改变后，需要重新计算 GearDisplay 和 GearDisplay2 的组合可见性
	// 参考 TypeScript 版本 GObject.ts handleControllerChanged (899行)
	g.CheckGearDisplay()
}

// CheckGearDisplay 计算 GearDisplay 和 GearDisplay2 的组合可见性
// 参考 TypeScript 版本 GObject.ts checkGearDisplay (617-634行)
// 更新 internalVisible 字段，然后调用 HandleVisibleChanged 更新实际显示
func (g *GObject) CheckGearDisplay() {
	if g == nil {
		return
	}
	if g.handlingController {
		return
	}

	// 1. 先获取 GearDisplay (index 0) 的 connected 状态
	connected := true
	if gearDisplay := g.gears[gears.IndexDisplay]; gearDisplay != nil {
		if gd, ok := gearDisplay.(interface{ Connected() bool }); ok {
			connected = gd.Connected()
		}
	}

	// 2. 如果有 GearDisplay2 (index 8)，用它的 evaluate 方法计算最终可见性
	if gearDisplay2 := g.gears[gears.IndexDisplay2]; gearDisplay2 != nil {
		if gd2, ok := gearDisplay2.(interface{ Evaluate(bool) bool }); ok {
			connected = gd2.Evaluate(connected)
		}
	}

	// 3. 更新 internalVisible（gear控制的可见性）
	// 注意：不直接修改 g.visible（用户控制的可见性）
	if g.internalVisible != connected {
		g.internalVisible = connected

		// 调用 HandleVisibleChanged 重新计算最终可见性
		// 检查 data 是否实现了自定义的 HandleVisibleChanged（如 GGroup）
		if handler, ok := g.data.(interface{ HandleVisibleChanged() }); ok {
			handler.HandleVisibleChanged()
		} else {
			g.HandleVisibleChanged()
		}
	}
}

// CheckGearController reports whether the specified gear slot is driven by the controller.
func (g *GObject) CheckGearController(index int, ctrl *Controller) bool {
	if g == nil || index < 0 || index >= gears.SlotCount || ctrl == nil {
		return false
	}
	gear := g.gears[index]
	return gear != nil && gear.Controller() == ctrl
}

type textAccessor interface {
	Text() string
	SetText(string)
}

type titleAccessor interface {
	Title() string
	SetTitle(string)
}

type iconAccessor interface {
	Icon() string
	SetIcon(string)
}

type colorAccessor interface {
	Color() string
	SetColor(string)
}

type outlineColorAccessor interface {
	OutlineColor() string
	SetOutlineColor(string)
}

type fontSizeAccessor interface {
	FontSize() int
	SetFontSize(int)
}

type titleColorAccessor interface {
	TitleColor() string
	SetTitleColor(string)
}

type titleOutlineColorAccessor interface {
	TitleOutlineColor() string
	SetTitleOutlineColor(string)
}

type titleFontSizeAccessor interface {
	TitleFontSize() int
	SetTitleFontSize(int)
}

type selectedAccessor interface {
	Selected() bool
	SetSelected(bool)
}

type playingAccessor interface {
	Playing() bool
	SetPlaying(bool)
}

type frameAccessor interface {
	Frame() int
	SetFrame(int)
}

type timeScaleAccessor interface {
	TimeScale() float64
	SetTimeScale(float64)
}

type deltaTimeAccessor interface {
	DeltaTime() float64
	SetDeltaTime(float64)
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
	case fmt.Stringer:
		return v.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func toBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case *bool:
		if v != nil {
			return *v
		}
	case int:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	default:
		return false
	}
	return false
}

func toInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	case *int:
		if v != nil {
			return *v, true
		}
	case *int32:
		if v != nil {
			return int(*v), true
		}
	case *int64:
		if v != nil {
			return int(*v), true
		}
	default:
		return 0, false
	}
	return 0, false
}

func toFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		if v == "" {
			return 0, false
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	case *float64:
		if v != nil {
			return *v, true
		}
	case *float32:
		if v != nil {
			return float64(*v), true
		}
	case *int:
		if v != nil {
			return float64(*v), true
		}
	case *int32:
		if v != nil {
			return float64(*v), true
		}
	case *int64:
		if v != nil {
			return float64(*v), true
		}
	}
	return 0, false
}

// SortingOrder returns the Z-order value used for rendering and interaction.
// Objects with higher values are rendered later (on top).
// Default is 0, which means use insertion order.
func (g *GObject) SortingOrder() int {
	if g == nil {
		return 0
	}
	return g.sortingOrder
}

// SetSortingOrder updates the Z-order value and repositions the object within its parent's children array.
// Objects with sortingOrder > 0 are maintained in ascending order at the end of the children array.
func (g *GObject) SetSortingOrder(value int) {
	if g == nil {
		return
	}
	if value < 0 {
		value = 0
	}
	if g.sortingOrder == value {
		return
	}
	oldValue := g.sortingOrder
	g.sortingOrder = value
	if g.parent != nil {
		g.parent.childSortingOrderChanged(g, oldValue, value)
	}
}
