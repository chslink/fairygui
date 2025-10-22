//go:build ebiten

package render

import (
	"errors"
	"fmt"
	"image/color"
	"log"
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
	combinedA := combined.Element(0, 0)
	combinedB := combined.Element(0, 1)
	combinedTx := combined.Element(0, 2)
	combinedC := combined.Element(1, 0)
	combinedD := combined.Element(1, 1)
	combinedTy := combined.Element(1, 2)
	log.Printf("[graph matrix] name=%s local=[[%.3f %.3f %.3f] [%.3f %.3f %.3f]] geo=[[%.3f %.3f %.3f] [%.3f %.3f %.3f]]", obj.Name(), localMatrix.A, localMatrix.C, localMatrix.Tx, localMatrix.B, localMatrix.D, localMatrix.Ty, combinedA, combinedB, combinedTx, combinedC, combinedD, combinedTy)

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
			if err := drawTextImage(target, combined, nil, data, alpha, obj.Width(), obj.Height(), atlas); err != nil {
				return err
			}
		}
	case *widgets.GTextField:
		if textValue := data.Text(); textValue != "" {
			if err := drawTextImage(target, combined, data, textValue, alpha, obj.Width(), obj.Height(), atlas); err != nil {
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
			if err := drawTextImage(target, textMatrix, nil, textValue, alpha, obj.Width(), obj.Height(), atlas); err != nil {
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
	if obj := graph.GObject; obj != nil {
		x := obj.X()
		y := obj.Y()
		px, py := obj.Pivot()
		skewX, skewY := obj.Skew()
		scaleX, scaleY := obj.Scale()
		offset := laya.Point{}
		displayPos := laya.Point{}
		localMatrix := laya.Matrix{}
		if sprite := obj.DisplayObject(); sprite != nil {
			displayPos = sprite.Position()
			offset = sprite.PivotOffset()
			localMatrix = sprite.LocalMatrix()
		}
		log.Printf("[graph debug] name=%s xy=(%.2f,%.2f) display=(%.2f,%.2f) offset=(%.2f,%.2f) size=(%.2f,%.2f) pivot=(%.3f,%.3f) anchor=%t scale=(%.3f,%.3f) rotation=%.3f skew=(%.3f,%.3f) localMatrix=[%.3f %.3f %.3f %.3f %.3f %.3f]",
			obj.Name(), x, y, displayPos.X, displayPos.Y, offset.X, offset.Y, obj.Width(), obj.Height(), px, py, obj.PivotAsAnchor(), scaleX, scaleY, obj.Rotation(), skewX, skewY,
			localMatrix.A, localMatrix.C, localMatrix.Tx, localMatrix.B, localMatrix.D, localMatrix.Ty)
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
	tmp := ebiten.NewImage(imgWidth, imgHeight)
	offsetX := strokePad
	offsetY := strokePad
	var drew bool
	switch graph.Type() {
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
	geo := parentGeo
	if strokePad > 0 {
		geo = applyLocalOffset(geo, -strokePad, -strokePad)
	}
	opts := &ebiten.DrawImageOptions{GeoM: geo}
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

func buildRoundedRectPath(path *vector.Path, width, height float64, radii []float64, offsetX, offsetY float64) bool {
	if path == nil || width <= 0 || height <= 0 {
		return false
	}
	corners := normalizeCornerRadii(width, height, radii)
	w := float32(width)
	h := float32(height)
	ox := float32(offsetX)
	oy := float32(offsetY)
	tl, tr, br, bl := corners[0], corners[1], corners[2], corners[3]

	path.MoveTo(ox+tl, oy)
	path.LineTo(ox+w-tr, oy)
	if tr > 0 {
		path.Arc(ox+w-tr, oy+tr, tr, -math.Pi/2, 0, vector.Clockwise)
	} else {
		path.LineTo(ox+w, oy)
	}
	path.LineTo(ox+w, oy+h-br)
	if br > 0 {
		path.Arc(ox+w-br, oy+h-br, br, 0, math.Pi/2, vector.Clockwise)
	} else {
		path.LineTo(ox+w, oy+h)
	}
	path.LineTo(ox+bl, oy+h)
	if bl > 0 {
		path.Arc(ox+bl, oy+h-bl, bl, math.Pi/2, math.Pi, vector.Clockwise)
	} else {
		path.LineTo(ox, oy+h)
	}
	path.LineTo(ox, oy+tl)
	if tl > 0 {
		path.Arc(ox+tl, oy+tl, tl, math.Pi, 3*math.Pi/2, vector.Clockwise)
	} else {
		path.LineTo(ox, oy)
	}
	path.Close()
	return true
}

func normalizeCornerRadii(width, height float64, raw []float64) [4]float32 {
	var vals [4]float32
	tmp := [4]float64{}
	for i := 0; i < 4; i++ {
		if i < len(raw) {
			tmp[i] = math.Max(0, raw[i])
		} else {
			tmp[i] = 0
		}
	}
	tmp[0], tmp[1] = clampRadiusPair(width, tmp[0], tmp[1])
	tmp[3], tmp[2] = clampRadiusPair(width, tmp[3], tmp[2])
	tmp[0], tmp[3] = clampRadiusPair(height, tmp[0], tmp[3])
	tmp[1], tmp[2] = clampRadiusPair(height, tmp[1], tmp[2])

	for i := 0; i < 4; i++ {
		vals[i] = float32(tmp[i])
	}
	return vals
}

func clampRadiusPair(limit, a, b float64) (float64, float64) {
	if limit <= 0 {
		return 0, 0
	}
	sum := a + b
	if sum > limit && sum > 0 {
		scale := limit / sum
		a *= scale
		b *= scale
	}
	return a, b
}

func buildEllipsePath(path *vector.Path, width, height float64, offsetX, offsetY float64) bool {
	if path == nil || width <= 0 || height <= 0 {
		return false
	}
	cx := float32(offsetX + width/2)
	cy := float32(offsetY + height/2)
	rx := float32(width / 2)
	ry := float32(height / 2)
	if rx <= 0 || ry <= 0 {
		return false
	}
	const kappa = 0.5522847498307936
	cxk := float32(kappa * width / 2)
	cyk := float32(kappa * height / 2)

	path.MoveTo(cx+rx, cy)
	path.CubicTo(cx+rx, cy+cyk, cx+cxk, cy+ry, cx, cy+ry)
	path.CubicTo(cx-cxk, cy+ry, cx-rx, cy+cyk, cx-rx, cy)
	path.CubicTo(cx-rx, cy-cyk, cx-cxk, cy-ry, cx, cy-ry)
	path.CubicTo(cx+cxk, cy-ry, cx+rx, cy-cyk, cx+rx, cy)
	path.Close()
	return true
}

func buildPolygonPath(path *vector.Path, points []float64, offsetX, offsetY float64) bool {
	if path == nil || len(points) < 6 || len(points)%2 != 0 {
		return false
	}
	ox := float32(offsetX)
	oy := float32(offsetY)
	path.MoveTo(float32(points[0])+ox, float32(points[1])+oy)
	for i := 2; i < len(points); i += 2 {
		path.LineTo(float32(points[i])+ox, float32(points[i+1])+oy)
	}
	path.Close()
	return true
}

func buildRegularPolygonPath(path *vector.Path, width, height float64, sides int, startAngle float64, distances []float64, offsetX, offsetY float64) bool {
	if path == nil || sides < 3 || width <= 0 || height <= 0 {
		return false
	}
	radius := math.Min(width, height) / 2
	if radius <= 0 {
		return false
	}
	centerX := float32(offsetX + width/2)
	centerY := float32(offsetY + height/2)
	angle := startAngle * math.Pi / 180
	step := 2 * math.Pi / float64(sides)
	for i := 0; i < sides; i++ {
		dist := 1.0
		if i < len(distances) && !math.IsNaN(distances[i]) && distances[i] > 0 {
			dist = distances[i]
		}
		x := centerX + float32(radius*dist*math.Cos(angle))
		y := centerY + float32(radius*dist*math.Sin(angle))
		if i == 0 {
			path.MoveTo(x, y)
		} else {
			path.LineTo(x, y)
		}
		angle += step
	}
	path.Close()
	return true
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

func computeStrokePadding(lineColor *color.NRGBA, lineSize float64) float64 {
	if lineColor == nil || lineColor.A == 0 || lineSize <= 0 {
		return 0
	}
	return math.Ceil(lineSize/2 + 1)
}

func ensureGraphCanvasSize(width, height float64, pad float64) (int, int) {
	w := int(math.Ceil(width + 2*pad))
	h := int(math.Ceil(height + 2*pad))
	if w <= 0 {
		w = 1
	}
	if h <= 0 {
		h = 1
	}
	return w, h
}

func applyLocalOffset(src ebiten.GeoM, dx, dy float64) ebiten.GeoM {
	if dx == 0 && dy == 0 {
		return src
	}
	a := src.Element(0, 0)
	b := src.Element(0, 1)
	c := src.Element(1, 0)
	d := src.Element(1, 1)
	tx := src.Element(0, 2)
	ty := src.Element(1, 2)

	tx = a*dx + b*dy + tx
	ty = c*dx + d*dy + ty

	src.SetElement(0, 2, tx)
	src.SetElement(1, 2, ty)
	return src
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
