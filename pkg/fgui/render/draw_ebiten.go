package render

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// clamp 限制数值在指定范围内
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

var (
	labelFont                    font.Face = basicfont.Face7x13
	debugNineSlice                         = os.Getenv("FGUI_DEBUG_NINESLICE") != ""
	debugNineSliceOverlayEnabled           = os.Getenv("FGUI_DEBUG_NINESLICE_OVERLAY") != ""
	lastNineSliceLog             sync.Map
	textureRectLog               sync.Map
	fontCacheSize                = 20 // 限制字体缓存大小，避免内存过度使用

	// graphRenderCache 缓存 GGraph 渲染结果，避免每帧重建
	graphRenderCache   = make(map[string]*ebiten.Image)
	graphRenderCacheMu sync.RWMutex
)

// SetTextFont overrides the default font used when drawing text-based widgets.
func SetTextFont(face font.Face) {
	if face != nil {
		labelFont = face
	}
}

// DrawComponent traverses the component hierarchy and draws the visible objects onto target.
// It currently supports image widgets; other widget types are skipped silently.
func DrawComponent(target *ebiten.Image, root *core.GComponent, atlas *AtlasManager) error {
	if target == nil {
		return errors.New("render: target image is nil")
	}
	if root == nil {
		return errors.New("render: component is nil")
	}
	if atlas == nil {
		return errors.New("render: atlas manager is nil")
	}

	var geo ebiten.GeoM
	geo.Reset()
	return drawComponent(target, root, atlas, geo, 1)
}

func drawComponent(target *ebiten.Image, comp *core.GComponent, atlas *AtlasManager, parentGeo ebiten.GeoM, parentAlpha float64) error {
	// 检查是否需要应用 scrollRect 裁剪
	display := comp.DisplayObject()
	container := comp.Container()

	// 关键修复：ScrollRect 可能设置在 container（maskContainer）上，而不是 display 上
	var scrollRect *laya.Rect
	if container != nil && container != display {
		// 优先检查 container 的 scrollRect（ScrollPane 场景）
		scrollRect = container.ScrollRect()
	}
	if scrollRect == nil {
		// 回退到检查 display 的 scrollRect（兼容其他场景）
		scrollRect = display.ScrollRect()
	}

	// 计算 container 相对于 display 的偏移
	containerGeo := parentGeo
	if container != nil && container != display {
		containerMatrix := container.LocalMatrix()
		containerTransform := ebiten.GeoM{}
		containerTransform.SetElement(0, 0, containerMatrix.A)
		containerTransform.SetElement(0, 1, containerMatrix.C)
		containerTransform.SetElement(0, 2, containerMatrix.Tx)
		containerTransform.SetElement(1, 0, containerMatrix.B)
		containerTransform.SetElement(1, 1, containerMatrix.D)
		containerTransform.SetElement(1, 2, containerMatrix.Ty)
		containerTransform.Concat(parentGeo)
		containerGeo = containerTransform
	}

	if scrollRect != nil {
		// 有 scrollRect，需要裁剪渲染
		return drawComponentWithClipping(target, comp, atlas, parentGeo, containerGeo, parentAlpha, scrollRect)
	}

	// 没有 scrollRect，正常渲染所有子对象
	for _, child := range comp.Children() {
		if child == nil {
			continue
		}
		if err := drawObject(target, child, atlas, containerGeo, parentAlpha); err != nil {
			return err
		}
	}

	// 渲染直接添加到 displayObject 的子对象（例如滚动条）
	if display != nil {
		extraCount := 0
		for _, childSprite := range display.Children() {
			if childSprite == container {
				continue
			}
			extraCount++

			if owner := childSprite.Owner(); owner != nil {
				var gobject *core.GObject
				if obj, ok := owner.(*core.GObject); ok {
					gobject = obj
				} else if gcomp, ok := owner.(*core.GComponent); ok {
					gobject = gcomp.GObject
				}

				if gobject != nil {
					// 检查是否在 Children() 列表中（避免重复渲染）
					isInChildren := false
					for _, child := range comp.Children() {
						if child == gobject {
							isInChildren = true
							break
						}
					}
					if isInChildren {
						continue
					}

					// 防止自渲染（避免无限循环）
					if gcomp, ok := owner.(*core.GComponent); ok {
						if gcomp == comp {
							continue
						}
					}

					// DEBUG: 只输出滚动条相关的日志
					isScrollBar := strings.Contains(strings.ToLower(gobject.Name()), "scroll") ||
						(gobject.Width() == 17 && gobject.Height() > 50) ||
						(gobject.Height() == 17 && gobject.Width() > 50)

					if isScrollBar {
						pos := childSprite.Position()
						dispVisible := childSprite.Visible()
						objVisible := gobject.Visible()
						fmt.Printf("[Render] ScrollBar: name='%s', pos=(%.1f,%.1f), dispVisible=%v, objVisible=%v, size=(%.1f,%.1f)\n",
							gobject.Name(), pos.X, pos.Y, dispVisible, objVisible, gobject.Width(), gobject.Height())
					}

					// 渲染额外的 GObject
					if err := drawObject(target, gobject, atlas, parentGeo, parentAlpha); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// drawComponentWithClipping 渲染带裁剪的组件
func drawComponentWithClipping(target *ebiten.Image, comp *core.GComponent, atlas *AtlasManager, parentGeo, containerGeo ebiten.GeoM, parentAlpha float64, scrollRect *laya.Rect) error {
	// 计算裁剪区域的实际尺寸
	clipW := int(math.Ceil(scrollRect.W))
	clipH := int(math.Ceil(scrollRect.H))

	if clipW <= 0 || clipH <= 0 {
		return nil
	}

	// 创建临时渲染目标
	tempTarget := ebiten.NewImage(clipW, clipH)
	defer func() {
		tempTarget.Dispose()
	}()

	// 创建内容偏移的变换矩阵
	contentGeo := ebiten.GeoM{}
	contentGeo.Translate(-scrollRect.X, -scrollRect.Y)

	// 渲染所有子对象到临时目标
	for _, child := range comp.Children() {
		if child == nil {
			continue
		}
		if err := drawObject(tempTarget, child, atlas, contentGeo, parentAlpha); err != nil {
			return err
		}
	}

	// 将裁剪后的内容绘制到最终目标
	finalGeo := parentGeo
	finalGeo.Translate(scrollRect.X, scrollRect.Y)

	opts := &ebiten.DrawImageOptions{
		GeoM: finalGeo,
	}

	if parentAlpha < 1.0 {
		opts.ColorScale.ScaleAlpha(float32(parentAlpha))
	}

	target.DrawImage(tempTarget, opts)

	// 渲染额外的 DisplayObject（如滚动条）
	display := comp.DisplayObject()
	container := comp.Container()
	if display != nil {
		for _, childSprite := range display.Children() {
			if childSprite == container {
				continue
			}
			if owner := childSprite.Owner(); owner != nil {
				var gobject *core.GObject
				if obj, ok := owner.(*core.GObject); ok {
					gobject = obj
				} else if gcomp, ok := owner.(*core.GComponent); ok {
					gobject = gcomp.GObject
				}

				if gobject != nil {
					// 检查是否在 Children() 列表中
					isInChildren := false
					for _, child := range comp.Children() {
						if child == gobject {
							isInChildren = true
							break
						}
					}
					if isInChildren {
						continue
					}

					// 防止自渲染
					if gcomp, ok := owner.(*core.GComponent); ok {
						if gcomp == comp {
							continue
						}
					}

					// DEBUG: 只输出滚动条相关的日志
					isScrollBar := strings.Contains(strings.ToLower(gobject.Name()), "scroll") ||
						(gobject.Width() == 17 && gobject.Height() > 50) ||
						(gobject.Height() == 17 && gobject.Width() > 50)

					if isScrollBar {
						pos := childSprite.Position()
						dispVisible := childSprite.Visible()
						objVisible := gobject.Visible()
						fmt.Printf("[Render/Clipping] ScrollBar: name='%s', pos=(%.1f,%.1f), dispVisible=%v, objVisible=%v, size=(%.1f,%.1f)\n",
							gobject.Name(), pos.X, pos.Y, dispVisible, objVisible, gobject.Width(), gobject.Height())
					}

					// 渲染额外的 GObject
					if err := drawObject(target, gobject, atlas, parentGeo, parentAlpha); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func drawObject(target *ebiten.Image, obj *core.GObject, atlas *AtlasManager, parentGeo ebiten.GeoM, parentAlpha float64) error {
	if obj == nil || !obj.Visible() {
		return nil
	}
	alpha := parentAlpha * obj.Alpha()
	if alpha <= 0 {
		return nil
	}

	sprite := obj.DisplayObject()
	localMatrix := sprite.LocalMatrix()
	combined := ebiten.GeoM{}
	combined.SetElement(0, 0, localMatrix.A)
	combined.SetElement(0, 1, localMatrix.C)
	combined.SetElement(0, 2, localMatrix.Tx)
	combined.SetElement(1, 0, localMatrix.B)
	combined.SetElement(1, 1, localMatrix.D)
	combined.SetElement(1, 2, localMatrix.Ty)
	combined.Concat(parentGeo)

	// ✅ 优先处理 Graphics 命令（统一渲染路径）
	gfx := sprite.Graphics()
	if gfx != nil && !gfx.IsEmpty() {
		// 遍历命令并使用专门的渲染器
		commands := gfx.Commands()
		for _, cmd := range commands {
			switch cmd.Type {
			case laya.GraphicsCommandTexture:
				// 使用 TextureRenderer 渲染纹理
				texRenderer := NewTextureRenderer(atlas)
				if err := texRenderer.Render(target, cmd.Texture, combined, alpha, sprite); err != nil {
					return err
				}
			case laya.GraphicsCommandRect, laya.GraphicsCommandEllipse,
				laya.GraphicsCommandPolygon, laya.GraphicsCommandPath,
				laya.GraphicsCommandLine, laya.GraphicsCommandPie:
				// 矢量命令：使用 renderGraphicsSprite（已有实现）
				if !renderGraphicsSprite(target, sprite, combined, alpha) {
					return fmt.Errorf("failed to render graphics command type %d", cmd.Type)
				}
				// ✅ 矢量命令处理完成后，继续渲染子对象
				// 需要处理所有可能包含子对象的容器类型
				switch data := obj.Data().(type) {
				case *core.GComponent:
					return drawComponent(target, data, atlas, combined, alpha)
				case *widgets.GButton:
					// 设置鼠标事件启用
					if sprite := obj.DisplayObject(); sprite != nil {
						sprite.SetMouseEnabled(true)
					}
					return drawComponent(target, data.GComponent, atlas, combined, alpha)
				case *widgets.GComboBox:
					if root := data.ComponentRoot(); root != nil {
						return drawComponent(target, root, atlas, combined, alpha)
					}
				case *widgets.GList:
					// 在渲染前检查虚拟列表是否需要刷新
					if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
						list.CheckVirtualList()
					}
					//fmt.Printf("[DEBUG GList] (vector path) rendering GList: name=%s, children=%d\n",
					//	obj.Name(), len(data.GComponent.Children()))
					return drawComponent(target, data.GComponent, atlas, combined, alpha)
				case *widgets.GTree:
					return drawComponent(target, data.GComponent, atlas, combined, alpha)
				}
				return nil
			}
		}
		// 命令处理完成，检查是否还需要递归渲染子对象
		// 需要处理所有可能包含子对象的容器类型
		switch data := obj.Data().(type) {
		case *core.GComponent:
			return drawComponent(target, data, atlas, combined, alpha)
		case *widgets.GButton:
			// GButton 内嵌 GComponent，渲染其子对象
			// 设置鼠标事件启用
			if sprite := obj.DisplayObject(); sprite != nil {
				sprite.SetMouseEnabled(true)
			}
			return drawComponent(target, data.GComponent, atlas, combined, alpha)
		case *widgets.GComboBox:
			// GComboBox 也是容器
			if root := data.ComponentRoot(); root != nil {
				return drawComponent(target, root, atlas, combined, alpha)
			}
		case *widgets.GList:
			// GList 也是容器
			//fmt.Printf("[DEBUG GList] rendering GList: name=%s, children=%d\n",
			//	obj.Name(), len(data.GComponent.Children()))
			return drawComponent(target, data.GComponent, atlas, combined, alpha)
		case *widgets.GTree:
			// GTree 也是容器
			return drawComponent(target, data.GComponent, atlas, combined, alpha)
		}
		return nil
	}

	// ✅ Widget 类型分发：处理各种 Widget 的专门渲染逻辑
	// Graphics 命令已在上面统一处理，这里处理特殊 Widget 和兼容路径
	w := obj.Width()
	h := obj.Height()
	if w <= 0 {
		if data, ok := obj.Data().(*assets.PackageItem); ok {
			if data != nil && data.Sprite != nil {
				w = float64(data.Sprite.Rect.Width)
			}
		}
	}
	if h <= 0 {
		if data, ok := obj.Data().(*assets.PackageItem); ok {
			if data != nil && data.Sprite != nil {
				h = float64(data.Sprite.Rect.Height)
			}
		}
	}

	switch data := obj.Data().(type) {
	case *assets.PackageItem:
		if err := drawPackageItem(target, data, combined, atlas, alpha, sprite); err != nil {
			return err
		}
	case *widgets.GImage:
		// ✅ GImage 已完全迁移到命令模式
		// updateGraphics() 总是生成 TextureCommand（除非 packageItem == nil）
		// 命令由统一的 TextureRenderer 处理
	case *widgets.GMovieClip:
		if err := renderMovieClipWidget(target, data, atlas, combined, alpha, sprite); err != nil {
			return err
		}
	case *core.GComponent:
		if err := drawComponent(target, data, atlas, combined, alpha); err != nil {
			return err
		}
	case string:
		if data != "" {
			if err := drawTextImage(target, combined, nil, data, alpha, obj.Width(), obj.Height(), atlas, sprite); err != nil {
				return err
			}
		}
	case *widgets.GTextInput:
		if textValue := data.Text(); textValue != "" {
			if err := drawTextImage(target, combined, data.GTextField, textValue, alpha, obj.Width(), obj.Height(), atlas, sprite); err != nil {
				return err
			}
		}
		// 绘制光标和选择区域
		if err := drawTextInputCursor(target, combined, data, alpha); err != nil {
			return err
		}
	case *widgets.GRichTextField:
		// 富文本控件：继承自 GTextField，需要特殊处理
		if textValue := data.Text(); textValue != "" {
			if err := drawTextImage(target, combined, data.GTextField, textValue, alpha, obj.Width(), obj.Height(), atlas, sprite); err != nil {
				return err
			}
		}
		// 启用鼠标交互以支持链接点击
		if sprite := obj.DisplayObject(); sprite != nil {
			sprite.SetMouseEnabled(true)
		}
	case *widgets.GTextField:
		if textValue := data.Text(); textValue != "" {
			if err := drawTextImage(target, combined, data, textValue, alpha, obj.Width(), obj.Height(), atlas, sprite); err != nil {
				return err
			}
		}
	case *widgets.GLabel:
		iconItem := data.IconItem()
		textMatrix := combined
		if iconItem != nil {
			iconGeo := combined
			if err := drawPackageItem(target, iconItem, iconGeo, atlas, alpha, sprite); err != nil {
				return err
			}
			if iconItem.Sprite != nil {
				shift := ebiten.GeoM{}
				shift.Translate(float64(iconItem.Sprite.Rect.Width)+4, 0)
				shift.Concat(combined)
				textMatrix = shift
			}
		}
		if textValue := data.Title(); textValue != "" {
			if err := drawTextImage(target, textMatrix, nil, textValue, alpha, obj.Width(), obj.Height(), atlas, sprite); err != nil {
				return err
			}
		}
	case *widgets.GButton:
		if tpl := data.TemplateComponent(); tpl != nil {
			if err := drawComponent(target, tpl, atlas, combined, alpha); err != nil {
				return err
			}
		} else if err := drawComponent(target, data.GComponent, atlas, combined, alpha); err != nil {
			return err
		}
		if sprite := obj.DisplayObject(); sprite != nil {
			sprite.SetMouseEnabled(true)
		}
	case *widgets.GLoader:
		// ⚠️ GLoader 已部分迁移到命令模式，这里是兼容路径
		// MovieClip、Component 和 FillMethod 仍使用旧渲染路径
		if gfx == nil || gfx.IsEmpty() {
			// 如果没有命令，使用旧渲染路径
			if err := renderLoader(target, data, atlas, combined, alpha); err != nil {
				return err
			}
		}
	case *widgets.GGraph:
		if err := renderGraph(target, data, combined, alpha, sprite); err != nil {
			return err
		}
	case *widgets.GList:
		// GList 内嵌 GComponent，渲染其子对象
		// 在渲染前检查虚拟列表是否需要刷新
		if list, ok := obj.Data().(*widgets.GList); ok && list.IsVirtual() {
			list.CheckVirtualList()
		}
		//fmt.Printf("[DEBUG GList] (widget path) rendering GList: name=%s, children=%d\n",
		//	obj.Name(), len(data.GComponent.Children()))
		if err := drawComponent(target, data.GComponent, atlas, combined, alpha); err != nil {
			return err
		}
	case *widgets.GComboBox:
		if root := data.ComponentRoot(); root != nil {
			if err := drawComponent(target, root, atlas, combined, alpha); err != nil {
				return err
			}
		}
	default:
		if sprite != nil {
			if renderGraphicsSprite(target, sprite, combined, alpha) {
				return nil
			}
		}
		// Unsupported payloads are ignored for now.
	}

	return nil
}

func renderMovieClipWidget(target *ebiten.Image, widget *widgets.GMovieClip, atlas *AtlasManager, parentGeo ebiten.GeoM, alpha float64, sprite *laya.Sprite) error {
	if widget == nil {
		return nil
	}
	frame := widget.CurrentFrame()
	if frame == nil {
		return nil
	}
	item := widget.PackageItem()

	alignWidth := 0
	alignHeight := 0
	if item != nil {
		if item.Width > 0 {
			alignWidth = item.Width
		}
		if item.Height > 0 {
			alignHeight = item.Height
		}
	}
	if alignWidth <= 0 {
		if frame.Width > 0 {
			alignWidth = frame.Width
		} else if frame.Sprite != nil {
			if frame.Sprite.OriginalSize.X > 0 {
				alignWidth = int(frame.Sprite.OriginalSize.X)
			} else if frame.Sprite.Rect.Width > 0 {
				alignWidth = frame.Sprite.Rect.Width
			}
		}
	}
	if alignHeight <= 0 {
		if frame.Height > 0 {
			alignHeight = frame.Height
		} else if frame.Sprite != nil {
			if frame.Sprite.OriginalSize.Y > 0 {
				alignHeight = int(frame.Sprite.OriginalSize.Y)
			} else if frame.Sprite.Rect.Height > 0 {
				alignHeight = frame.Sprite.Rect.Height
			}
		}
	}
	img, err := atlas.ResolveMovieClipFrameAligned(item, frame, alignWidth, alignHeight)
	useAligned := true
	if err != nil {
		useAligned = false
		img, err = atlas.ResolveMovieClipFrame(item, frame)
		if err != nil {
			return err
		}
	}
	if img == nil {
		return nil
	}

	tint := parseColor(widget.Color())

	baseWidth := 0.0
	baseHeight := 0.0
	if useAligned {
		if alignWidth > 0 {
			baseWidth = float64(alignWidth)
		}
		if alignHeight > 0 {
			baseHeight = float64(alignHeight)
		}
	} else {
		if item != nil {
			if item.Width > 0 {
				baseWidth = float64(item.Width)
			}
			if item.Height > 0 {
				baseHeight = float64(item.Height)
			}
		}
		if baseWidth <= 0 && frame.Sprite != nil && frame.Sprite.OriginalSize.X > 0 {
			baseWidth = float64(frame.Sprite.OriginalSize.X)
		}
		if baseHeight <= 0 && frame.Sprite != nil && frame.Sprite.OriginalSize.Y > 0 {
			baseHeight = float64(frame.Sprite.OriginalSize.Y)
		}
		if baseWidth <= 0 && frame.Width > 0 {
			baseWidth = float64(frame.Width)
		}
		if baseHeight <= 0 && frame.Height > 0 {
			baseHeight = float64(frame.Height)
		}
	}
	if baseWidth <= 0 {
		baseWidth = float64(img.Bounds().Dx())
	}
	if baseHeight <= 0 {
		baseHeight = float64(img.Bounds().Dy())
	}

	dstW := widget.Width()
	dstH := widget.Height()
	if dstW <= 0 {
		dstW = baseWidth
	}
	if dstH <= 0 {
		dstH = baseHeight
	}

	sx := 1.0
	sy := 1.0
	if baseWidth > 0 {
		sx = dstW / baseWidth
	}
	if baseHeight > 0 {
		sy = dstH / baseHeight
	}

	// 构建本地变换矩阵
	// 正确顺序：缩放 → 翻转 → frame offset → sprite offset → 父变换
	local := ebiten.GeoM{}

	// 1. 缩放到目标尺寸
	local.Scale(sx, sy)

	// 2. 翻转（在本地坐标系）
	switch widget.Flip() {
	case widgets.FlipTypeHorizontal:
		local.Scale(-1, 1)
		local.Translate(dstW, 0)
	case widgets.FlipTypeVertical:
		local.Scale(1, -1)
		local.Translate(0, dstH)
	case widgets.FlipTypeBoth:
		local.Scale(-1, -1)
		local.Translate(dstW, dstH)
	}

	// 3. frame offset（在翻转之后，避免被镜像）
	if !useAligned {
		offsetX := float64(frame.OffsetX) * sx
		offsetY := float64(frame.OffsetY) * sy
		local.Translate(offsetX, offsetY)
	}

	// 4. sprite offset（在翻转之后，不参与缩放 - 使用原始值）
	if !useAligned && frame.Sprite != nil {
		off := frame.Sprite.Offset
		if off.X != 0 || off.Y != 0 {
			local.Translate(float64(off.X), float64(off.Y))
		}
	}

	// 5. 应用父变换
	// ✅ 正确顺序：先 local 变换，再 parent 变换
	// 这样 sprite.rotation/scale 会作用于已缩放好的 MovieClip
	local.Concat(parentGeo)
	geo := local

	method, origin, clockwise, amount := widget.Fill()
	if method != 0 && amount > 0 && amount < 0.9999 {
		points := computeFillPoints(dstW, dstH, method, origin, clockwise, amount)
		if len(points) >= 6 {
			srcBounds := img.Bounds()
			srcW := float64(srcBounds.Dx())
			srcH := float64(srcBounds.Dy())
			var scaleSrcX, scaleSrcY float64
			if dstW > 0 {
				scaleSrcX = srcW / dstW
			}
			if dstH > 0 {
				scaleSrcY = srcH / dstH
			}
			var colorR, colorG, colorB, colorA float32 = 1, 1, 1, float32(alpha)
			if tint != nil {
				colorR = float32(tint.R) / 255
				colorG = float32(tint.G) / 255
				colorB = float32(tint.B) / 255
				colorA *= float32(tint.A) / 255
			}
			vertices := make([]ebiten.Vertex, len(points)/2)
			for i := 0; i < len(points); i += 2 {
				px := points[i]
				py := points[i+1]
				x, y := geo.Apply(px, py)
				srcX := px * scaleSrcX
				srcY := py * scaleSrcY
				vertices[i/2] = ebiten.Vertex{
					DstX:   float32(x),
					DstY:   float32(y),
					SrcX:   float32(srcX),
					SrcY:   float32(srcY),
					ColorR: colorR,
					ColorG: colorG,
					ColorB: colorB,
					ColorA: colorA,
				}
			}
			indices := make([]uint16, 0, (len(vertices)-2)*3)
			for i := 1; i < len(vertices)-1; i++ {
				indices = append(indices, 0, uint16(i), uint16(i+1))
			}
			options := &ebiten.DrawTrianglesOptions{}
			target.DrawTriangles(vertices, indices, img, options)
			return nil
		}
	}
	renderImageWithGeo(target, img, geo, alpha, tint, sprite)
	return nil
}

func renderGraph(target *ebiten.Image, graph *widgets.GGraph, parentGeo ebiten.GeoM, alpha float64, displaySprite *laya.Sprite) error {
	if target == nil || graph == nil {
		return nil
	}
	w := graph.GObject.Width()
	h := graph.GObject.Height()
	if w <= 0 || h <= 0 {
		return nil
	}
	fillColor := parseColor(graph.FillColor())
	lineColor := parseColor(graph.LineColor())
	lineSize := graph.LineSize()
	if (fillColor == nil || fillColor.A == 0) && (lineColor == nil || lineColor.A == 0 || lineSize <= 0) {
		return nil
	}
	strokePad := computeStrokePadding(lineColor, lineSize)
	imgWidth, imgHeight := ensureGraphCanvasSize(w, h, strokePad)
	if imgWidth <= 0 || imgHeight <= 0 {
		return nil
	}

	// 生成缓存键：基于 graph 对象指针、尺寸、类型和颜色
	graphType := graph.Type()
	cacheKey := fmt.Sprintf("graph_%p_%dx%d_%d_%s_%s_%.1f",
		graph, imgWidth, imgHeight, graphType, graph.FillColor(), graph.LineColor(), lineSize)

	// 尝试从缓存获取
	graphRenderCacheMu.RLock()
	tmp, cached := graphRenderCache[cacheKey]
	graphRenderCacheMu.RUnlock()

	if !cached {
		// 缓存未命中，创建新图像并渲染
		tmp = ebiten.NewImage(imgWidth, imgHeight)
		offsetX := strokePad
		offsetY := strokePad
		var drew bool
		switch graphType {
		case widgets.GraphTypeEmpty:
			// TypeScript 实现中为空图形不会绘制任何内容。
			return nil
		case widgets.GraphTypeRect:
			if radii := graph.CornerRadius(); len(radii) > 0 {
				var path vector.Path
				if buildRoundedRectPath(&path, w, h, radii, offsetX, offsetY) {
					drew = drawGraphPath(tmp, &path, fillColor, lineColor, lineSize, alpha)
				}
			}
			if !drew {
				if fillColor != nil {
					tint := applyAlpha(fillColor, alpha)
					vector.FillRect(tmp, float32(offsetX), float32(offsetY), float32(w), float32(h), tint, true)
					drew = true
				}
				if lineColor != nil && lineSize > 0 {
					tint := applyAlpha(lineColor, alpha)
					vector.StrokeRect(tmp, float32(offsetX), float32(offsetY), float32(w), float32(h), float32(lineSize), tint, true)
					drew = true
				}
			}
		case widgets.GraphTypeEllipse:
			var path vector.Path
			if buildEllipsePath(&path, w, h, offsetX, offsetY) {
				drew = drawGraphPath(tmp, &path, fillColor, lineColor, lineSize, alpha)
			}
		case widgets.GraphTypePolygon:
			points := graph.PolygonPoints()
			if len(points) >= 6 {
				var path vector.Path
				if buildPolygonPath(&path, points, offsetX, offsetY) {
					drew = drawGraphPath(tmp, &path, fillColor, lineColor, lineSize, alpha)
				}
			}
		case widgets.GraphTypeRegularPolygon:
			var path vector.Path
			sides, startAngle, distances := graph.RegularPolygon()
			if buildRegularPolygonPath(&path, w, h, sides, startAngle, distances, offsetX, offsetY) {
				drew = drawGraphPath(tmp, &path, fillColor, lineColor, lineSize, alpha)
			}
		default:
			if fillColor != nil {
				tint := applyAlpha(fillColor, alpha)
				vector.FillRect(tmp, float32(offsetX), float32(offsetY), float32(w), float32(h), tint, true)
				drew = true
			}
			if lineColor != nil && lineSize > 0 {
				tint := applyAlpha(lineColor, alpha)
				vector.StrokeRect(tmp, float32(offsetX), float32(offsetY), float32(w), float32(h), float32(lineSize), tint, true)
				drew = true
			}
		}
		if !drew {
			return nil
		}

		// 存入缓存
		graphRenderCacheMu.Lock()
		graphRenderCache[cacheKey] = tmp
		graphRenderCacheMu.Unlock()
	}

	// 使用缓存的 tmp 进行绘制
	geo := parentGeo
	if strokePad > 0 {
		geo = applyLocalOffset(geo, -strokePad, -strokePad)
	}
	opts := &ebiten.DrawImageOptions{GeoM: geo}
	applyColorEffects(opts, displaySprite)
	target.DrawImage(tmp, opts)
	return nil
}

func applyAlpha(src *color.NRGBA, alpha float64) color.NRGBA {
	if src == nil {
		return color.NRGBA{}
	}
	if alpha < 0 {
		alpha = 0
	} else if alpha > 1 {
		alpha = 1
	}
	out := *src
	out.A = uint8(math.Round(float64(out.A) * alpha))
	return out
}

func drawGraphPath(dst *ebiten.Image, path *vector.Path, fillColor, lineColor *color.NRGBA, lineSize float64, alpha float64) bool {
	if dst == nil || path == nil {
		return false
	}
	drew := false
	if fillColor != nil && fillColor.A > 0 {
		tint := applyAlpha(fillColor, alpha)
		var drawOpts vector.DrawPathOptions
		drawOpts.AntiAlias = true
		drawOpts.ColorScale.ScaleWithColor(tint)
		vector.FillPath(dst, path, nil, &drawOpts)
		drew = true
	}
	if lineColor != nil && lineColor.A > 0 && lineSize > 0 {
		tint := applyAlpha(lineColor, alpha)
		var strokeOpts vector.StrokeOptions
		strokeOpts.Width = float32(lineSize)
		strokeOpts.LineJoin = vector.LineJoinRound
		strokeOpts.LineCap = vector.LineCapRound
		var drawOpts vector.DrawPathOptions
		drawOpts.AntiAlias = true
		drawOpts.ColorScale.ScaleWithColor(tint)
		vector.StrokePath(dst, path, &strokeOpts, &drawOpts)
		drew = true
	}
	return drew
}

func localGeoMForObject(obj *core.GObject) ebiten.GeoM {
	var geo ebiten.GeoM
	geo.Reset()
	if obj == nil {
		return geo
	}
	if sprite := obj.DisplayObject(); sprite != nil {
		matrix := sprite.LocalMatrix()
		geo.SetElement(0, 0, matrix.A)
		geo.SetElement(0, 1, matrix.C)
		geo.SetElement(0, 2, matrix.Tx)
		geo.SetElement(1, 0, matrix.B)
		geo.SetElement(1, 1, matrix.D)
		geo.SetElement(1, 2, matrix.Ty)
	} else {
		geo.Translate(obj.X(), obj.Y())
	}
	return geo
}

func selectFontFace(field *widgets.GTextField) font.Face {
	if labelFont != nil {
		return labelFont
	}
	return basicfont.Face7x13
}

func fontFaceForSize(size int) font.Face {
	if size <= 0 {
		return nil
	}
	if face := fontFaceCacheLookup(size); face != nil {
		return face
	}
	return nil
}

func parseColor(value string) *color.NRGBA {
	if value == "" {
		// 改进：返回默认黑色而不是 nil
		return &color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}
	raw := strings.TrimSpace(value)
	lowered := strings.ToLower(raw)

	// 支持常见颜色名称
	colorNames := map[string]color.NRGBA{
		"black":   {R: 0, G: 0, B: 0, A: 255},
		"white":   {R: 255, G: 255, B: 255, A: 255},
		"red":     {R: 255, G: 0, B: 0, A: 255},
		"green":   {R: 0, G: 128, B: 0, A: 255},
		"blue":    {R: 0, G: 0, B: 255, A: 255},
		"yellow":  {R: 255, G: 255, B: 0, A: 255},
		"cyan":    {R: 0, G: 255, B: 255, A: 255},
		"magenta": {R: 255, G: 0, B: 255, A: 255},
		"silver":  {R: 192, G: 192, B: 192, A: 255},
		"gray":    {R: 128, G: 128, B: 128, A: 255},
		"grey":    {R: 128, G: 128, B: 128, A: 255},
		"maroon":  {R: 128, G: 0, B: 0, A: 255},
		"olive":   {R: 128, G: 128, B: 0, A: 255},
		"purple":  {R: 128, G: 0, B: 128, A: 255},
		"teal":    {R: 0, G: 128, B: 128, A: 255},
		"navy":    {R: 0, G: 0, B: 128, A: 255},
		"orange":  {R: 255, G: 165, B: 0, A: 255},
		"pink":    {R: 255, G: 192, B: 203, A: 255},
		"brown":   {R: 165, G: 42, B: 42, A: 255},
	}

	if namedColor, exists := colorNames[lowered]; exists {
		return &namedColor
	}

	// 支持透明颜色名称
	if lowered == "transparent" || lowered == "none" {
		return &color.NRGBA{R: 0, G: 0, B: 0, A: 0}
	}

	// 支持 rgb() 格式
	if strings.HasPrefix(lowered, "rgb(") && strings.HasSuffix(lowered, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(lowered, "rgb("), ")")
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			var r, g, b int
			var err error
			if r, err = strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
				if g, err = strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					if b, err = strconv.Atoi(strings.TrimSpace(parts[2])); err == nil {
						return &color.NRGBA{
							R: uint8(clamp(r, 0, 255)),
							G: uint8(clamp(g, 0, 255)),
							B: uint8(clamp(b, 0, 255)),
							A: 255,
						}
					}
				}
			}
		}
	}

	// 支持 rgba() 格式
	if strings.HasPrefix(lowered, "rgba(") && strings.HasSuffix(lowered, ")") {
		inner := strings.TrimSuffix(strings.TrimPrefix(lowered, "rgba("), ")")
		parts := strings.Split(inner, ",")
		if len(parts) == 4 {
			var r, g, b, a int
			var err error
			if r, err = strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
				if g, err = strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					if b, err = strconv.Atoi(strings.TrimSpace(parts[2])); err == nil {
						if a, err = strconv.Atoi(strings.TrimSpace(parts[3])); err == nil {
							return &color.NRGBA{
								R: uint8(clamp(r, 0, 255)),
								G: uint8(clamp(g, 0, 255)),
								B: uint8(clamp(b, 0, 255)),
								A: uint8(clamp(a, 0, 255)),
							}
						}
					}
				}
			}
		}
	}

	if strings.HasPrefix(lowered, "0x") {
		hex := raw[2:]
		switch len(hex) {
		case 3: // 0xRGB 格式
			if v, err := strconv.ParseUint(hex, 16, 32); err == nil {
				r := uint8((v >> 8) & 0xF)
				g := uint8((v >> 4) & 0xF)
				b := uint8(v & 0xF)
				return &color.NRGBA{
					R: r | r<<4,
					G: g | g<<4,
					B: b | b<<4,
					A: 0xff,
				}
			}
		case 6:
			if v, err := strconv.ParseUint(hex, 16, 32); err == nil {
				return &color.NRGBA{
					R: uint8(v >> 16),
					G: uint8(v >> 8),
					B: uint8(v),
					A: 0xff,
				}
			}
		case 8:
			if v, err := strconv.ParseUint(hex, 16, 32); err == nil {
				return &color.NRGBA{
					A: uint8(v >> 24),
					R: uint8(v >> 16),
					G: uint8(v >> 8),
					B: uint8(v),
				}
			}
		}
	}
	if strings.HasPrefix(raw, "#") {
		raw = strings.TrimPrefix(raw, "#")
		switch len(raw) {
		case 3: // #RGB 格式
			if v, err := strconv.ParseUint(raw, 16, 32); err == nil {
				r := uint8((v >> 8) & 0xF)
				g := uint8((v >> 4) & 0xF)
				b := uint8(v & 0xF)
				return &color.NRGBA{
					R: r | r<<4,
					G: g | g<<4,
					B: b | b<<4,
					A: 0xff,
				}
			}
		case 6:
			if v, err := strconv.ParseUint(raw, 16, 32); err == nil {
				return &color.NRGBA{
					R: uint8(v >> 16),
					G: uint8(v >> 8),
					B: uint8(v),
					A: 0xff,
				}
			}
		case 8:
			if v, err := strconv.ParseUint(raw, 16, 32); err == nil {
				return &color.NRGBA{
					A: uint8(v >> 24),
					R: uint8(v >> 16),
					G: uint8(v >> 8),
					B: uint8(v),
				}
			}
		}
	}
	if strings.HasPrefix(strings.ToLower(raw), "rgba") {
		start := strings.Index(raw, "(")
		end := strings.LastIndex(raw, ")")
		if start != -1 && end != -1 && end > start {
			body := raw[start+1 : end]
			parts := strings.Split(body, ",")
			if len(parts) == 4 {
				parseComponent := func(s string, scale float64) uint8 {
					val := strings.TrimSpace(s)
					if val == "" {
						return 0
					}
					if scale == 1 {
						if n, err := strconv.Atoi(val); err == nil {
							if n < 0 {
								n = 0
							} else if n > 255 {
								n = 255
							}
							return uint8(n)
						}
					} else {
						if f, err := strconv.ParseFloat(val, 64); err == nil {
							if f < 0 {
								f = 0
							} else if f > 1 {
								f = 1
							}
							return uint8(math.Round(f * scale))
						}
					}
					return 0
				}
				return &color.NRGBA{
					R: parseComponent(parts[0], 1),
					G: parseComponent(parts[1], 1),
					B: parseComponent(parts[2], 1),
					A: parseComponent(parts[3], 255),
				}
			}
		}
	}
	// 改进：如果无法解析颜色，返回默认黑色而不是 nil
	return &color.NRGBA{R: 0, G: 0, B: 0, A: 255}
}

// 改进的字体缓存机制
func fontFaceCacheLookup(size int) font.Face {
	// 首先检查系统字体缓存
	systemFontMu.RLock()
	if face, ok := systemFontCache[size]; ok {
		systemFontMu.RUnlock()
		return face
	}
	systemFontMu.RUnlock()

	// 如果缓存中没有，尝试加载
	face, err := getFontFace(size)
	if err != nil {
		// 回退到默认字体
		if labelFont != nil {
			return labelFont
		}
		return basicfont.Face7x13
	}

	// 缓存新加载的字体，但限制缓存大小
	systemFontMu.Lock()
	defer systemFontMu.Unlock()

	// 如果缓存超过限制，清理最旧的条目
	if len(systemFontCache) >= fontCacheSize {
		// 简单的LRU策略：删除第一个条目
		for k := range systemFontCache {
			delete(systemFontCache, k)
			break
		}
	}

	systemFontCache[size] = face
	return face
}

func drawPackageItem(target *ebiten.Image, item *assets.PackageItem, geo ebiten.GeoM, atlas *AtlasManager, alpha float64, sprite *laya.Sprite) error {
	if item == nil {
		return nil
	}
	spriteAny, err := atlas.ResolveSprite(item)
	if err != nil {
		return err
	}
	img, ok := spriteAny.(*ebiten.Image)
	if !ok || img == nil {
		return errors.New("render: atlas returned unexpected sprite type")
	}

	if spriteInfo := item.Sprite; spriteInfo != nil {
		if spriteInfo.Offset.X != 0 || spriteInfo.Offset.Y != 0 {
			geo.Translate(float64(spriteInfo.Offset.X), float64(spriteInfo.Offset.Y))
		}
	}

	opts := &ebiten.DrawImageOptions{
		GeoM: geo,
	}
	applyTintColor(opts, nil, alpha, sprite)
	target.DrawImage(img, opts)
	return nil
}

func legacyDrawLoader(target *ebiten.Image, loader *widgets.GLoader, atlas *AtlasManager, parentGeo ebiten.GeoM, alpha float64) error {
	return renderLoader(target, loader, atlas, parentGeo, alpha)
}

func legacyDrawLoaderPackageItem(target *ebiten.Image, loader *widgets.GLoader, item *assets.PackageItem, parentGeo ebiten.GeoM, atlas *AtlasManager, alpha float64) error {
	return renderLoaderPackageItem(target, loader, item, parentGeo, atlas, alpha)
	/*
		if loader == nil || item == nil {
			return nil
		}
		spriteAny, err := atlas.ResolveSprite(item)
		if err != nil {
			return err
		}
		img, ok := spriteAny.(*ebiten.Image)
		if !ok || img == nil {
			return errors.New("render: atlas returned unexpected sprite type")
		}

		geo := parentGeo
		sx, sy := loader.ContentScale()
		if sx == 0 {
			sx = 1
		}
		if sy == 0 {
			sy = 1
		}
		if sx != 1 || sy != 1 {
			geo.Scale(sx, sy)
		}
		if ox, oy := loader.ContentOffset(); ox != 0 || oy != 0 {
			geo.Translate(ox, oy)
		}
		if sprite := item.Sprite; sprite != nil {
			if sprite.Offset.X != 0 || sprite.Offset.Y != 0 {
				geo.Translate(float64(sprite.Offset.X), float64(sprite.Offset.Y))
			}
		}

		method := loader.FillMethod()
		amount := loader.FillAmount()
		if method == widgets.LoaderFillMethodNone || amount <= 0 {
			return drawLoaderImage(target, loader, img, geo, alpha)
		}

		if amount >= 0.9999 {
			return drawLoaderImage(target, loader, img, geo, alpha)
		}

		w, h := loader.ContentSize()
		if w <= 0 {
			w = float64(img.Bounds().Dx()) * sx
		}
		if h <= 0 {
			h = float64(img.Bounds().Dy()) * sy
		}

		points := computeFillPoints(w, h, method, loader.FillOrigin(), loader.FillClockwise(), amount)
		if len(points) < 6 {
			return drawLoaderImage(target, loader, img, geo, alpha)
			return nil
		}

		invSx := sx
		if invSx == 0 {
			invSx = 1
		}
		invSy := sy
		if invSy == 0 {
			invSy = 1
		}

		vertexCount := len(points) / 2
		vertices := make([]ebiten.Vertex, vertexCount)
		for i := 0; i < vertexCount; i++ {
			px := points[2*i]
			py := points[2*i+1]
			dx, dy := geo.Apply(px, py)
			vertices[i] = ebiten.Vertex{
				DstX:   float32(dx),
				DstY:   float32(dy),
				SrcX:   float32(px / invSx),
				SrcY:   float32(py / invSy),
				ColorR: 1,
				ColorG: 1,
				ColorB: 1,
				ColorA: float32(alpha),
			}
		}
		indices := make([]uint16, 0, (vertexCount-2)*3)
		for i := 1; i < vertexCount-1; i++ {
			indices = append(indices, 0, uint16(i), uint16(i+1))
		}
		opts := &ebiten.DrawTrianglesOptions{}
		target.DrawTriangles(vertices, indices, img, opts)
		return nil
	*/
}

// drawTextInputCursor 绘制文本输入框的光标和选择区域。
func drawTextInputCursor(target *ebiten.Image, geo ebiten.GeoM, input *widgets.GTextInput, alpha float64) error {
	if input == nil {
		return nil
	}

	// 只在获得焦点时绘制光标和选择
	if !input.IsFocused() {
		return nil
	}

	text := input.Text()
	runes := []rune(text)
	cursorPos := input.CursorPosition()
	selStart, selEnd := input.GetSelection()
	hasSelection := input.HasSelection()

	// 计算文本度量信息
	fontSize := float64(input.FontSize())
	if fontSize <= 0 {
		fontSize = 12
	}
	letterSpacing := float64(input.LetterSpacing())
	avgCharWidth := fontSize * 0.6

	// 计算光标或选择区域的 X 坐标
	calculateX := func(pos int) float64 {
		if pos < 0 {
			pos = 0
		}
		if pos > len(runes) {
			pos = len(runes)
		}

		x := 0.0
		for i := 0; i < pos; i++ {
			charWidth := avgCharWidth
			if i < len(runes) && runes[i] == ' ' {
				charWidth = fontSize * 0.3
			}
			x += charWidth
			if i < pos-1 {
				x += letterSpacing
			}
		}
		return x
	}

	// 绘制选择区域(如果有) - 使用 vector 绘制,避免创建临时图像
	if hasSelection && selStart != selEnd {
		selectionStartX := calculateX(selStart)
		selectionEndX := calculateX(selEnd)
		selectionWidth := selectionEndX - selectionStartX
		selectionHeight := fontSize * 1.2

		// 选择区域颜色(半透明蓝色)
		selectionColor := color.NRGBA{R: 51, G: 153, B: 255, A: uint8(100 * alpha)}

		// 应用变换
		x, y := geo.Apply(selectionStartX, 0)

		// 使用 vector 直接绘制矩形,不创建临时图像
		vector.DrawFilledRect(target, float32(x), float32(y), float32(selectionWidth), float32(selectionHeight), selectionColor, false)
	}

	// 绘制光标(如果可见) - 使用 vector 绘制,避免创建临时图像
	if input.IsCursorVisible() {
		cursorX := calculateX(cursorPos)
		cursorWidth := 1.0
		cursorHeight := fontSize * 1.2

		// 光标颜色(黑色)
		cursorColor := color.NRGBA{R: 0, G: 0, B: 0, A: uint8(255 * alpha)}

		// 应用变换
		x, y := geo.Apply(cursorX, 0)

		// 使用 vector 直接绘制矩形,不创建临时图像
		vector.DrawFilledRect(target, float32(x), float32(y), float32(cursorWidth), float32(cursorHeight), cursorColor, false)
	}

	return nil
}
