//go:build ebiten

package render

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

var (
	labelFont                    font.Face = basicfont.Face7x13
	debugNineSlice                         = os.Getenv("FGUI_DEBUG_NINESLICE") != ""
	debugNineSliceOverlayEnabled           = os.Getenv("FGUI_DEBUG_NINESLICE_OVERLAY") != ""
	lastNineSliceLog             sync.Map
	textureRectLog               sync.Map
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
	for _, child := range comp.Children() {
		if err := drawObject(target, child, atlas, parentGeo, parentAlpha); err != nil {
			return err
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

	local := ebiten.GeoM{}
	local.Reset()

	w := obj.Width()
	h := obj.Height()
	if w <= 0 {
		switch data := obj.Data().(type) {
		case *assets.PackageItem:
			if data != nil && data.Sprite != nil {
				w = float64(data.Sprite.Rect.Width)
			}
		case *widgets.GImage:
			if pkg := data.PackageItem(); pkg != nil && pkg.Sprite != nil {
				w = float64(pkg.Sprite.Rect.Width)
			}
		}
	}
	if h <= 0 {
		switch data := obj.Data().(type) {
		case *assets.PackageItem:
			if data != nil && data.Sprite != nil {
				h = float64(data.Sprite.Rect.Height)
			}
		case *widgets.GImage:
			if pkg := data.PackageItem(); pkg != nil && pkg.Sprite != nil {
				h = float64(pkg.Sprite.Rect.Height)
			}
		}
	}

	px, py := obj.Pivot()
	if !obj.PivotAsAnchor() {
		local.Translate(-px*w, -py*h)
	}

	sx, sy := obj.Scale()
	if sx != 0 || sy != 0 {
		local.Scale(sx, sy)
	}

	skewX, skewY := obj.Skew()
	if skewX != 0 || skewY != 0 {
		local.Skew(skewX, skewY)
	}

	rotation := obj.Rotation()
	if rotation != 0 {
		local.Rotate(rotation)
	}

	pos := obj.DisplayObject().Position()
	local.Translate(pos.X, pos.Y)

	combined := local
	combined.Concat(parentGeo)

	switch data := obj.Data().(type) {
	case *assets.PackageItem:
		if err := drawPackageItem(target, data, combined, atlas, alpha); err != nil {
			return err
		}
	case *widgets.GImage:
		if err := renderImageWidget(target, data, atlas, combined, alpha); err != nil {
			return err
		}
	case *core.GComponent:
		if err := drawComponent(target, data, atlas, combined, alpha); err != nil {
			return err
		}
	case string:
		if data != "" {
			if err := drawTextImage(target, combined, nil, data, alpha, obj.Width(), obj.Height()); err != nil {
				return err
			}
		}
	case *widgets.GTextField:
		if textValue := data.Text(); textValue != "" {
			if err := drawTextImage(target, combined, data, textValue, alpha, obj.Width(), obj.Height()); err != nil {
				return err
			}
		}
	case *widgets.GLabel:
		iconItem := data.IconItem()
		textMatrix := combined
		if iconItem != nil {
			iconGeo := combined
			if err := drawPackageItem(target, iconItem, iconGeo, atlas, alpha); err != nil {
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
			if err := drawTextImage(target, textMatrix, nil, textValue, alpha, obj.Width(), obj.Height()); err != nil {
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
	case *widgets.GLoader:
		if err := renderLoader(target, data, atlas, combined, alpha); err != nil {
			return err
		}
	case *widgets.GGraph:
		if err := renderGraph(target, data, combined, alpha); err != nil {
			return err
		}
	default:
		// Unsupported payloads are ignored for now.
	}

	return nil
}

func renderImageWidget(target *ebiten.Image, widget *widgets.GImage, atlas *AtlasManager, parentGeo ebiten.GeoM, alpha float64) error {
	if widget == nil {
		return nil
	}
	item := widget.PackageItem()
	if item == nil {
		return nil
	}
	spriteAny, err := atlas.ResolveSprite(item)
	if err != nil {
		return err
	}
	img, ok := spriteAny.(*ebiten.Image)
	if !ok || img == nil {
		return errors.New("render: atlas returned unexpected sprite type for image")
	}

	bounds := img.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())
	dstW := widget.Width()
	dstH := widget.Height()
	if dstW <= 0 {
		dstW = srcW
	}
	if dstH <= 0 {
		dstH = srcH
	}

	geo := parentGeo
	if sprite := item.Sprite; sprite != nil {
		if sprite.Offset.X != 0 || sprite.Offset.Y != 0 {
			geo.Translate(float64(sprite.Offset.X), float64(sprite.Offset.Y))
		}
	}

	if grid := item.Scale9Grid; grid != nil {
		left := clampFloat(float64(grid.X), 0, srcW)
		top := clampFloat(float64(grid.Y), 0, srcH)
		right := clampFloat(srcW-float64(grid.X+grid.Width), 0, srcW)
		bottom := clampFloat(srcH-float64(grid.Y+grid.Height), 0, srcH)
		slice := nineSlice{
			left:   int(left),
			right:  int(right),
			top:    int(top),
			bottom: int(bottom),
		}
		scaleByTile, tileGrid := widget.ScaleSettings()
		debugLabel := fmt.Sprintf("image=%s", item.ID)
		if debugNineSlice {
			logKey := fmt.Sprintf("scale9:%s:%.1fx%.1f:%d:%d:%d:%d:%t:%d", item.ID, dstW, dstH, slice.left, slice.right, slice.top, slice.bottom, scaleByTile, tileGrid)
			if prev, ok := lastNineSliceLog.Load(item.ID); !ok || prev != logKey {
				log.Printf("[render][9slice] id=%s dst=%.1fx%.1f src=%dx%d slice={L:%d R:%d T:%d B:%d} tile=%v grid=%d",
					item.ID, dstW, dstH, bounds.Dx(), bounds.Dy(), slice.left, slice.right, slice.top, slice.bottom, scaleByTile, tileGrid)
				lastNineSliceLog.Store(item.ID, logKey)
			}
		}
		drawNineSlice(target, geo, img, slice, dstW, dstH, alpha, scaleByTile, tileGrid, debugLabel)
		if debugNineSliceOverlayEnabled {
			drawNineSliceOverlay(target, geo, slice, dstW, dstH)
		}
		if debugNineSlice {
			x0, y0 := geo.Apply(0, 0)
			x1, y1 := geo.Apply(dstW, 0)
			x2, y2 := geo.Apply(dstW, dstH)
			x3, y3 := geo.Apply(0, dstH)
			rectKey := fmt.Sprintf("scale9rect:%s:%.1fx%.1f:%.1f,%.1f:%.1f,%.1f:%.1f,%.1f:%.1f,%.1f", item.ID, dstW, dstH, x0, y0, x1, y1, x2, y2, x3, y3)
			if prev, ok := textureRectLog.Load(item.ID); !ok || prev != rectKey {
				log.Printf("[render][rect] id=%s corners=(%.1f,%.1f)-(%.1f,%.1f)-(%.1f,%.1f)-(%.1f,%.1f)", item.ID, x0, y0, x1, y1, x2, y2, x3, y3)
				textureRectLog.Store(item.ID, rectKey)
			}
		}
		return nil
	}

	sx := 1.0
	sy := 1.0
	if srcW > 0 {
		sx = dstW / srcW
	}
	if srcH > 0 {
		sy = dstH / srcH
	}
	scaleGeo := geo
	scaleGeo.Scale(sx, sy)
	renderImageWithGeo(target, img, scaleGeo, alpha)
	if debugNineSlice {
		logKey := fmt.Sprintf("simple:%s:%.1fx%.1f:%.2f:%.2f", item.ID, dstW, dstH, sx, sy)
		if prev, ok := lastNineSliceLog.Load(item.ID); !ok || prev != logKey {
			log.Printf("[render][image] id=%s simple-scale dst=%.1fx%.1f src=%dx%d scale=(%.2f,%.2f)",
				item.ID, dstW, dstH, bounds.Dx(), bounds.Dy(), sx, sy)
			lastNineSliceLog.Store(item.ID, logKey)
		}
		x0, y0 := scaleGeo.Apply(0, 0)
		x1, y1 := scaleGeo.Apply(1, 0)
		x2, y2 := scaleGeo.Apply(1, 1)
		x3, y3 := scaleGeo.Apply(0, 1)
		rectKey := fmt.Sprintf("simpleRect:%s:%.1fx%.1f:%.2f:%.2f:%.1f,%.1f:%.1f,%.1f:%.1f,%.1f:%.1f,%.1f", item.ID, dstW, dstH, sx, sy, x0, y0, x1, y1, x2, y2, x3, y3)
		if prev, ok := textureRectLog.Load(item.ID); !ok || prev != rectKey {
			log.Printf("[render][rect] id=%s corners=(%.1f,%.1f)-(%.1f,%.1f)-(%.1f,%.1f)-(%.1f,%.1f)", item.ID, x0, y0, x1, y1, x2, y2, x3, y3)
			textureRectLog.Store(item.ID, rectKey)
		}
	}
	return nil
}

func renderGraph(target *ebiten.Image, graph *widgets.GGraph, parentGeo ebiten.GeoM, alpha float64) error {
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
	imgWidth := int(math.Ceil(w))
	imgHeight := int(math.Ceil(h))
	if imgWidth <= 0 || imgHeight <= 0 {
		return nil
	}
	tmp := ebiten.NewImage(imgWidth, imgHeight)
	switch graph.Type() {
	case widgets.GraphTypeRect, widgets.GraphTypeEmpty:
		if fillColor != nil {
			tint := applyAlpha(fillColor, alpha)
			vector.FillRect(tmp, 0, 0, float32(w), float32(h), tint, true)
		}
		if lineColor != nil && lineSize > 0 {
			tint := applyAlpha(lineColor, alpha)
			vector.StrokeRect(tmp, 0, 0, float32(w), float32(h), float32(lineSize), tint, true)
		}
	default:
		if fillColor != nil {
			tint := applyAlpha(fillColor, alpha)
			vector.FillRect(tmp, 0, 0, float32(w), float32(h), tint, true)
		}
		if lineColor != nil && lineSize > 0 {
			tint := applyAlpha(lineColor, alpha)
			vector.StrokeRect(tmp, 0, 0, float32(w), float32(h), float32(lineSize), tint, true)
		}
	}
	opts := &ebiten.DrawImageOptions{GeoM: parentGeo}
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

func drawTextImage(target *ebiten.Image, geo ebiten.GeoM, field *widgets.GTextField, value string, alpha float64, width, height float64) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	face := selectFontFace(field)
	lines := strings.Split(value, "\n")
	metrics := face.Metrics()
	ascent := metrics.Ascent.Ceil()
	descent := metrics.Descent.Ceil()
	lineHeight := ascent + descent
	if lineHeight <= 0 {
		lineHeight = metrics.Height.Ceil()
	}
	if lineHeight <= 0 {
		lineHeight = 1
	}
	letterSpacing := 0
	leading := 0
	align := widgets.TextAlignLeft
	valign := widgets.TextVerticalAlignTop
	col := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	if field != nil {
		if v := parseColor(field.Color()); v != nil {
			col = *v
		}
		if field.FontSize() > 0 {
			if sized := fontFaceForSize(field.FontSize()); sized != nil {
				face = sized
				metrics = face.Metrics()
				ascent = metrics.Ascent.Ceil()
				descent = metrics.Descent.Ceil()
				lineHeight = ascent + descent
				if lineHeight <= 0 {
					lineHeight = metrics.Height.Ceil()
				}
				if lineHeight <= 0 {
					lineHeight = 1
				}
			}
		}
		if field.LetterSpacing() != 0 {
			letterSpacing = field.LetterSpacing()
		}
		if field.Leading() != 0 {
			leading = field.Leading()
		}
		align = field.Align()
		valign = field.VerticalAlign()
	}
	maxWidth := 0.0
	for _, line := range lines {
		bound := text.BoundString(face, line)
		lineWidth := float64(bound.Dx())
		if letterSpacing != 0 && len(line) > 1 {
			lineWidth += float64(letterSpacing * (len(line) - 1))
		}
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	lineHeightWithLeading := lineHeight + leading
	textHeight := float64(lineHeight) + float64(lineHeightWithLeading-lineHeight)*(math.Max(0, float64(len(lines)-1)))
	if width <= 0 {
		width = maxWidth
	}
	if height <= 0 {
		height = textHeight
	}
	imgW := int(math.Ceil(width))
	imgH := int(math.Ceil(height))
	if imgW <= 0 {
		imgW = int(math.Max(1, maxWidth))
	}
	if imgW <= 0 {
		imgW = 1
	}
	if imgH <= 0 {
		imgH = int(math.Max(1, textHeight))
	}
	if imgH <= 0 {
		imgH = 1
	}
	textImg := ebiten.NewImage(imgW, imgH)
	startY := 0.0
	if valign == widgets.TextVerticalAlignMiddle {
		startY = (float64(imgH) - textHeight) / 2
	} else if valign == widgets.TextVerticalAlignBottom {
		startY = float64(imgH) - textHeight
	}
	if startY < 0 {
		startY = 0
	}
	drawer := font.Drawer{
		Dst:  textImg,
		Src:  image.NewUniform(col),
		Face: face,
	}
	for idx, line := range lines {
		bound := text.BoundString(face, line)
		lineWidth := float64(bound.Dx())
		if letterSpacing != 0 && len(line) > 1 {
			lineWidth += float64(letterSpacing * (len(line) - 1))
		}
		startX := 0.0
		switch align {
		case widgets.TextAlignCenter:
			startX = (float64(imgW) - lineWidth) / 2
		case widgets.TextAlignRight:
			startX = float64(imgW) - lineWidth
		default:
			startX = 0
		}
		if startX < 0 {
			startX = 0
		}
		y := startY + float64(idx)*float64(lineHeightWithLeading) + float64(ascent)
		drawer.Dot = fixed.Point26_6{
			X: fixed.Int26_6(startX * 64),
			Y: fixed.Int26_6(y * 64),
		}
		if letterSpacing == 0 {
			drawer.DrawString(line)
			continue
		}
		spacing := fixed.I(letterSpacing)
		runes := []rune(line)
		for i, r := range runes {
			drawer.DrawString(string(r))
			if i != len(runes)-1 {
				drawer.Dot.X += spacing
			}
		}
	}
	opts := &ebiten.DrawImageOptions{GeoM: geo}
	if alpha < 1 {
		opts.ColorM.Scale(1, 1, 1, alpha)
	}
	target.DrawImage(textImg, opts)
	return nil
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
		return nil
	}
	raw := strings.TrimSpace(value)
	if strings.HasPrefix(raw, "#") {
		raw = strings.TrimPrefix(raw, "#")
		switch len(raw) {
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
	return nil
}

func fontFaceCacheLookup(size int) font.Face {
	face, err := getFontFace(size)
	if err != nil {
		return nil
	}
	return face
}

func drawPackageItem(target *ebiten.Image, item *assets.PackageItem, geo ebiten.GeoM, atlas *AtlasManager, alpha float64) error {
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

	if sprite := item.Sprite; sprite != nil {
		if sprite.Offset.X != 0 || sprite.Offset.Y != 0 {
			geo.Translate(float64(sprite.Offset.X), float64(sprite.Offset.Y))
		}
	}

	opts := &ebiten.DrawImageOptions{
		GeoM: geo,
	}
	if alpha < 1 {
		opts.ColorM.Scale(1, 1, 1, alpha)
	}
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

func legacyDrawImageWithGeo(target *ebiten.Image, img *ebiten.Image, geo ebiten.GeoM, alpha float64) {
	renderImageWithGeo(target, img, geo, alpha)
}
