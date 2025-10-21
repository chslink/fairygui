//go:build ebiten

package render

import (
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

const (
	loaderFillMethodNone       = int(widgets.LoaderFillMethodNone)
	loaderFillMethodHorizontal = int(widgets.LoaderFillMethodHorizontal)
	loaderFillMethodVertical   = int(widgets.LoaderFillMethodVertical)
	loaderFillMethodRadial90   = int(widgets.LoaderFillMethodRadial90)
	loaderFillMethodRadial180  = int(widgets.LoaderFillMethodRadial180)
	loaderFillMethodRadial360  = int(widgets.LoaderFillMethodRadial360)
)

const (
	loaderFillOriginTop = iota
	loaderFillOriginBottom
	loaderFillOriginLeft
	loaderFillOriginRight

	loaderFillOriginTopLeft     = loaderFillOriginTop
	loaderFillOriginTopRight    = loaderFillOriginBottom
	loaderFillOriginBottomLeft  = loaderFillOriginLeft
	loaderFillOriginBottomRight = loaderFillOriginRight
)

func renderLoader(target *ebiten.Image, loader *widgets.GLoader, atlas *AtlasManager, parentGeo ebiten.GeoM, alpha float64) error {
	if loader == nil {
		return nil
	}
	if comp := loader.Component(); comp != nil {
		return drawComponent(target, comp, atlas, parentGeo, alpha)
	}
	if item := loader.PackageItem(); item != nil {
		return renderLoaderPackageItem(target, loader, item, parentGeo, atlas, alpha)
	}
	return nil
}

func renderLoaderPackageItem(target *ebiten.Image, loader *widgets.GLoader, item *assets.PackageItem, parentGeo ebiten.GeoM, atlas *AtlasManager, alpha float64) error {
	if loader == nil || item == nil {
		return nil
	}
	resolved, err := atlas.ResolveSprite(item)
	if err != nil {
		return err
	}
	img, ok := resolved.(*ebiten.Image)
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

	method := int(loader.FillMethod())
	amount := loader.FillAmount()

	if grid := loader.Scale9Grid(); grid != nil {
		debugLabel := fmt.Sprintf("loader=%s item=%s", safeLoaderName(loader), item.ID)
		bounds := img.Bounds()
		left := int(grid.X)
		top := int(grid.Y)
		right := int(math.Max(0, float64(bounds.Dx())-float64(grid.X+grid.Width)))
		bottom := int(math.Max(0, float64(bounds.Dy())-float64(grid.Y+grid.Height)))
		slice := nineSlice{left: left, right: right, top: top, bottom: bottom}
		dstW, dstH := loader.ContentSize()
		if dstW <= 0 {
			dstW = float64(bounds.Dx())
		}
		if dstH <= 0 {
			dstH = float64(bounds.Dy())
		}
		if dstW > debugLargeDimensionLimit || dstH > debugLargeDimensionLimit || math.IsNaN(dstW) || math.IsNaN(dstH) || math.IsInf(dstW, 0) || math.IsInf(dstH, 0) {
			sx, sy := loader.ContentScale()
			ox, oy := loader.ContentOffset()
			log.Printf("[renderLoader] suspicious scale9 target: %s dst=(%.2f, %.2f) bounds=%v scale=(%.2f, %.2f) offset=(%.2f, %.2f) grid=%+v tile=%v method=%d amount=%.4f",
				debugLabel, dstW, dstH, bounds, sx, sy, ox, oy, *grid, loader.ScaleByTile(), method, amount)
		}
		drawNineSlice(target, geo, img, slice, dstW, dstH, alpha, loader.ScaleByTile(), loader.TileGridIndice(), debugLabel)
		return nil
	}

	if method == loaderFillMethodNone || amount <= 0 {
		return renderLoaderImage(target, loader, img, geo, alpha)
	}
	if amount >= 0.9999 {
		return renderLoaderImage(target, loader, img, geo, alpha)
	}

	dstW, dstH := loader.ContentSize()
	if dstW <= 0 {
		dstW = float64(img.Bounds().Dx())
	}
	if dstH <= 0 {
		dstH = float64(img.Bounds().Dy())
	}

	points := computeFillPoints(dstW, dstH, method, loader.FillOrigin(), loader.FillClockwise(), amount)
	if len(points) < 6 {
		return renderLoaderImage(target, loader, img, geo, alpha)
	}

	vertices := make([]ebiten.Vertex, len(points)/2)
	for i := 0; i < len(points); i += 2 {
		px := points[i]
		py := points[i+1]
		x, y := geo.Apply(px, py)
		vertices[i/2] = ebiten.Vertex{
			DstX:   float32(x),
			DstY:   float32(y),
			SrcX:   float32(px),
			SrcY:   float32(py),
			ColorR: 1,
			ColorG: 1,
			ColorB: 1,
			ColorA: float32(alpha),
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

func renderLoaderImage(target *ebiten.Image, loader *widgets.GLoader, img *ebiten.Image, geo ebiten.GeoM, alpha float64) error {
	renderImageWithGeo(target, img, geo, alpha)
	return nil
}

func renderImageWithGeo(target *ebiten.Image, img *ebiten.Image, geo ebiten.GeoM, alpha float64) {
	opts := &ebiten.DrawImageOptions{GeoM: geo}
	if alpha < 1 {
		opts.ColorM.Scale(1, 1, 1, alpha)
	}
	target.DrawImage(img, opts)
}

func safeLoaderName(loader *widgets.GLoader) string {
	if loader == nil {
		return "<nil>"
	}
	if name := loader.Name(); name != "" {
		return name
	}
	return loader.ID()
}

// computeFillPoints replicates FillUtils.fillImage for loader fills.
func computeFillPoints(w, h float64, method int, origin int, clockwise bool, amount float64) []float64 {
	if amount <= 0 {
		return nil
	}
	if amount >= 0.9999 {
		return []float64{0, 0, w, 0, w, h, 0, h}
	}

	switch method {
	case loaderFillMethodHorizontal:
		return fillHorizontal(w, h, origin, amount)
	case loaderFillMethodVertical:
		return fillVertical(w, h, origin, amount)
	case loaderFillMethodRadial90:
		return fillRadial90(w, h, origin, clockwise, amount)
	case loaderFillMethodRadial180:
		return fillRadial180(w, h, origin, clockwise, amount)
	case loaderFillMethodRadial360:
		return fillRadial360(w, h, origin, clockwise, amount)
	default:
		return nil
	}
}

func fillHorizontal(w, h float64, origin int, amount float64) []float64 {
	w2 := w * amount
	if origin == loaderFillOriginTop || origin == loaderFillOriginLeft {
		return []float64{0, 0, w2, 0, w2, h, 0, h}
	}
	return []float64{w, 0, w, h, w - w2, h, w - w2, 0}
}

func fillVertical(w, h float64, origin int, amount float64) []float64 {
	h2 := h * amount
	if origin == loaderFillOriginTop || origin == loaderFillOriginLeft {
		return []float64{0, 0, 0, h2, w, h2, w, 0}
	}
	return []float64{0, h, w, h, w, h - h2, 0, h - h2}
}

func fillRadial90(w, h float64, origin int, clockwise bool, amount float64) []float64 {
	origin &= 3
	if (clockwise && (origin == loaderFillOriginTopRight || origin == loaderFillOriginBottomLeft)) ||
		(!clockwise && (origin == loaderFillOriginTopLeft || origin == loaderFillOriginBottomRight)) {
		amount = 1 - amount
	}
	v := math.Tan(math.Pi / 2 * amount)
	h2 := w * v
	if h2 == 0 {
		h2 = 1e-6
	}
	v2 := (h2 - h) / h2

	switch origin {
	case loaderFillOriginTopLeft:
		if clockwise {
			if h2 <= h {
				return []float64{0, 0, w, h2, w, 0}
			}
			return []float64{0, 0, w * (1 - v2), h, w, h, w, 0}
		}
		if h2 <= h {
			return []float64{0, 0, w, h2, w, h, 0, h}
		}
		return []float64{0, 0, w * (1 - v2), h, 0, h}
	case loaderFillOriginTopRight:
		if clockwise {
			if h2 <= h {
				return []float64{w, 0, 0, h2, 0, h, w, h}
			}
			return []float64{w, 0, w * v2, h, w, h}
		}
		if h2 <= h {
			return []float64{w, 0, 0, h2, 0, 0}
		}
		return []float64{w, 0, w * v2, h, 0, h, 0, 0}
	case loaderFillOriginBottomLeft:
		if clockwise {
			if h2 <= h {
				return []float64{0, h, w, h - h2, w, 0, 0, 0}
			}
			return []float64{0, h, w * (1 - v2), 0, 0, 0}
		}
		if h2 <= h {
			return []float64{0, h, w, h - h2, w, h}
		}
		return []float64{0, h, w * (1 - v2), 0, w, 0, w, h}
	case loaderFillOriginBottomRight:
		if clockwise {
			if h2 <= h {
				return []float64{w, h, 0, h - h2, 0, h}
			}
			return []float64{w, h, w * v2, 0, 0, 0, 0, h}
		}
		if h2 <= h {
			return []float64{w, h, 0, h - h2, 0, 0, w, 0}
		}
		return []float64{w, h, w * v2, 0, w, 0}
	}
	return nil
}

func fillRadial180(w, h float64, origin int, clockwise bool, amount float64) []float64 {
	origin &= 3
	var points []float64
	switch origin {

	case loaderFillOriginTop:
		if amount <= 0.5 {
			amount = amount / 0.5
			points = fillRadial90(w/2, h, ternary(clockwise, loaderFillOriginTopLeft, loaderFillOriginTopRight), clockwise, amount)
			if clockwise {
				movePoints(points, w/2, 0)
			}
		} else {
			amount = (amount - 0.5) / 0.5
			points = fillRadial90(w/2, h, ternary(clockwise, loaderFillOriginTopRight, loaderFillOriginTopLeft), clockwise, amount)
			if clockwise {
				points = append(points, w, h, w, 0)
			} else {
				movePoints(points, w/2, 0)
				points = append(points, 0, h, 0, 0)
			}
		}
	case loaderFillOriginBottom:
		if amount <= 0.5 {
			amount = amount / 0.5
			points = fillRadial90(w/2, h, ternary(clockwise, loaderFillOriginBottomRight, loaderFillOriginBottomLeft), clockwise, amount)
			if !clockwise {
				movePoints(points, w/2, 0)
			}
		} else {
			amount = (amount - 0.5) / 0.5
			points = fillRadial90(w/2, h, ternary(clockwise, loaderFillOriginBottomLeft, loaderFillOriginBottomRight), clockwise, amount)
			if clockwise {
				movePoints(points, w/2, 0)
				points = append(points, 0, 0, 0, h)
			} else {
				points = append(points, w, 0, w, h)
			}
		}
	case loaderFillOriginLeft:
		if amount <= 0.5 {
			amount = amount / 0.5
			points = fillRadial90(w, h/2, ternary(clockwise, loaderFillOriginBottomLeft, loaderFillOriginTopLeft), clockwise, amount)
			if !clockwise {
				movePoints(points, 0, h/2)
			}
		} else {
			amount = (amount - 0.5) / 0.5
			points = fillRadial90(w, h/2, ternary(clockwise, loaderFillOriginTopLeft, loaderFillOriginBottomLeft), clockwise, amount)
			if clockwise {
				movePoints(points, 0, h/2)
				points = append(points, w, 0, 0, 0)
			} else {
				points = append(points, w, h, 0, h)
			}
		}
	case loaderFillOriginRight:
		if amount <= 0.5 {
			amount = amount / 0.5
			points = fillRadial90(w, h/2, ternary(clockwise, loaderFillOriginTopRight, loaderFillOriginBottomRight), clockwise, amount)
			if clockwise {
				movePoints(points, 0, h/2)
			}
		} else {
			amount = (amount - 0.5) / 0.5
			points = fillRadial90(w, h/2, ternary(clockwise, loaderFillOriginBottomRight, loaderFillOriginTopRight), clockwise, amount)
			if clockwise {
				points = append(points, 0, h, w, h)
			} else {
				movePoints(points, 0, h/2)
				points = append(points, 0, 0, w, 0)
			}
		}
	}
	return points
}

func fillRadial360(w, h float64, origin int, clockwise bool, amount float64) []float64 {
	origin &= 3
	if amount <= 0.5 {
		return fillRadial180(w, h, origin, clockwise, amount*2)
	}
	points := fillRadial180(w, h, oppositeOrigin(origin), !clockwise, (amount-0.5)*2)
	switch origin {
	case loaderFillOriginTop:
		points = append(points, 0, h, w, h)
	case loaderFillOriginBottom:
		points = append(points, w, 0, 0, 0)
	case loaderFillOriginLeft:
		points = append(points, w, h, w, 0)
	case loaderFillOriginRight:
		points = append(points, 0, 0, 0, h)
	}
	return points
}

func movePoints(points []float64, offsetX, offsetY float64) {
	for i := 0; i < len(points); i += 2 {
		points[i] += offsetX
		points[i+1] += offsetY
	}
}

func ternary(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}

func oppositeOrigin(origin int) int {
	switch origin {
	case loaderFillOriginTop:
		return loaderFillOriginBottom
	case loaderFillOriginBottom:
		return loaderFillOriginTop
	case loaderFillOriginLeft:
		return loaderFillOriginRight
	case loaderFillOriginRight:
		return loaderFillOriginLeft
	default:
		return origin
	}
}
