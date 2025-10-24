package render

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"sync"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type nineSlice struct {
	left   int
	right  int
	top    int
	bottom int
}

var (
	solidImageOnce sync.Once
	solidUnitImage *ebiten.Image
	segmentLog     sync.Map
	centerLog      sync.Map
)

const debugLargeDimensionLimit = 8192.0

func drawNineSlice(target *ebiten.Image, baseGeo ebiten.GeoM, img *ebiten.Image, slice nineSlice, dstW, dstH float64, alpha float64, tint *color.NRGBA, scaleByTile bool, tileGrid int, sprite *laya.Sprite, debugLabel string) {
	if target == nil || img == nil {
		return
	}
	if dstW <= 0 || dstH <= 0 {
		return
	}

	bounds := img.Bounds()
	if bounds.Empty() {
		return
	}
	if debugLabel == "" {
		debugLabel = "<unnamed-loader>"
	}

	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())

	left := clampFloat(float64(slice.left), 0, srcW)
	right := clampFloat(float64(slice.right), 0, srcW-left)
	top := clampFloat(float64(slice.top), 0, srcH)
	bottom := clampFloat(float64(slice.bottom), 0, srcH-top)

	dstCols := splitSegments(dstW, left, right)
	dstRows := splitSegments(dstH, top, bottom)
	srcCols := [4]float64{0, left, srcW - right, srcW}
	srcRows := [4]float64{0, top, srcH - bottom, srcH}

	tileCenter := scaleByTile || tileGrid != 0
	if debugNineSlice {
		key := fmt.Sprintf("%s:%.1fx%.1f", debugLabel, dstW, dstH)
		if _, ok := segmentLog.Load(key); !ok {
			log.Printf("[render][9slice] %s segments dstRows=[%.2f %.2f %.2f %.2f] dstCols=[%.2f %.2f %.2f %.2f]",
				debugLabel,
				dstRows[0], dstRows[1], dstRows[2], dstRows[3],
				dstCols[0], dstCols[1], dstCols[2], dstCols[3],
			)
			segmentLog.Store(key, true)
		}
	}

	if dstW > debugLargeDimensionLimit || dstH > debugLargeDimensionLimit || math.IsNaN(dstW) || math.IsNaN(dstH) || math.IsInf(dstW, 0) || math.IsInf(dstH, 0) {
		log.Printf("[drawNineSlice] suspicious dst size: %s dst=(%.2f, %.2f) src=%v slice=%+v tileCenter=%v tileGrid=%d",
			debugLabel, dstW, dstH, bounds, slice, tileCenter, tileGrid)
	}

	for yi := 0; yi < 3; yi++ {
		sy0 := srcRows[yi]
		sy1 := srcRows[yi+1]
		srcHeight := sy1 - sy0

		dy0 := dstRows[yi]
		dy1 := dstRows[yi+1]
		dstHeight := dy1 - dy0

		if srcHeight <= 0 || dstHeight <= 0 {
			continue
		}

		for xi := 0; xi < 3; xi++ {
			sx0 := srcCols[xi]
			sx1 := srcCols[xi+1]
			srcWidth := sx1 - sx0

			dx0 := dstCols[xi]
			dx1 := dstCols[xi+1]
			dstWidth := dx1 - dx0

			if srcWidth <= 0 || dstWidth <= 0 {
				continue
			}

			if dstWidth > debugLargeDimensionLimit || dstHeight > debugLargeDimensionLimit || math.IsNaN(dstWidth) || math.IsNaN(dstHeight) || math.IsInf(dstWidth, 0) || math.IsInf(dstHeight, 0) {
				log.Printf("[drawNineSlice] suspicious patch: %s dstPatch=(%.2f, %.2f) srcPatch=(%.2f, %.2f)-(%.2f, %.2f) tileCenter=%v",
					debugLabel, dstWidth, dstHeight, sx0, sy0, sx1, sy1, tileCenter)
			}

			if debugNineSlice && xi == 1 && yi == 1 {
				centerKey := fmt.Sprintf("%s:%.1fx%.1f", debugLabel, dstWidth, dstHeight)
				if _, logged := centerLog.Load(centerKey); !logged {
					log.Printf("[render][9slice] %s center src=(%.2f,%.2f)-(%.2f,%.2f) dst=%.2fx%.2f", debugLabel, sx0, sy0, sx1, sy1, dstWidth, dstHeight)
					centerLog.Store(centerKey, true)
				}
			}
			if tileCenter && xi == 1 && yi == 1 {
				tileImagePatch(target, baseGeo, img, sx0, sy0, srcWidth, srcHeight, dx0, dy0, dstWidth, dstHeight, alpha, tint, sprite, debugLabel)
				continue
			}

			drawImagePatch(target, baseGeo, img, sx0, sy0, srcWidth, srcHeight, dx0, dy0, dstWidth, dstHeight, alpha, tint, sprite, debugLabel)
		}
	}
}

func drawNineSliceOverlay(target *ebiten.Image, baseGeo ebiten.GeoM, slice nineSlice, dstW, dstH float64) {
	if !debugNineSliceOverlayEnabled {
		return
	}
	left := float64(slice.left)
	right := dstW - float64(slice.right)
	top := float64(slice.top)
	bottom := dstH - float64(slice.bottom)
	if right < left {
		right = left
	}
	if bottom < top {
		bottom = top
	}

	transform := func(x, y float64) (float64, float64) {
		return baseGeo.Apply(x, y)
	}

	drawRect := func(x, y, w, h float64, col color.Color) {
		x0, y0 := transform(x, y)
		x1, y1 := transform(x+w, y)
		x2, y2 := transform(x+w, y+h)
		x3, y3 := transform(x, y+h)
		ebitenutil.DrawLine(target, x0, y0, x1, y1, col)
		ebitenutil.DrawLine(target, x1, y1, x2, y2, col)
		ebitenutil.DrawLine(target, x2, y2, x3, y3, col)
		ebitenutil.DrawLine(target, x3, y3, x0, y0, col)
	}

	drawRect(0, 0, dstW, dstH, color.RGBA{0x40, 0x80, 0xff, 0xff})
	drawRect(left, top, right-left, bottom-top, color.RGBA{0xff, 0xd7, 0x00, 0xff})
}

func drawImagePatch(target *ebiten.Image, baseGeo ebiten.GeoM, img *ebiten.Image, sx0, sy0, sw, sh, dx, dy, dw, dh, alpha float64, tint *color.NRGBA, sprite *laya.Sprite, debugLabel string) {
	if dw <= 0 || dh <= 0 || sw <= 0 || sh <= 0 {
		return
	}

	bounds := img.Bounds()
	x0 := bounds.Min.X + int(math.Round(sx0))
	y0 := bounds.Min.Y + int(math.Round(sy0))
	x1 := bounds.Min.X + int(math.Round(sx0+sw))
	y1 := bounds.Min.Y + int(math.Round(sy0+sh))

	if x0 < bounds.Min.X {
		x0 = bounds.Min.X
	}
	if y0 < bounds.Min.Y {
		y0 = bounds.Min.Y
	}
	if x1 > bounds.Max.X {
		x1 = bounds.Max.X
	}
	if y1 > bounds.Max.Y {
		y1 = bounds.Max.Y
	}
	if x1 <= x0 || y1 <= y0 {
		return
	}

	actualSW := float64(x1 - x0)
	actualSH := float64(y1 - y0)
	if actualSW <= 0 || actualSH <= 0 {
		return
	}

	subImg, ok := img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
	if !ok {
		return
	}

	local := ebiten.GeoM{}
	local.Scale(dw/actualSW, dh/actualSH)
	local.Translate(dx, dy)

	geo := local
	geo.Concat(baseGeo)

	if dw > debugLargeDimensionLimit || dh > debugLargeDimensionLimit || math.IsNaN(dw) || math.IsNaN(dh) || math.IsInf(dw, 0) || math.IsInf(dh, 0) {
		log.Printf("[drawImagePatch] suspicious dst patch: %s dst=(%.2f, %.2f) actualSrc=(%.2f, %.2f) srcRect=(%d,%d)-(%d,%d)",
			debugLabel, dw, dh, actualSW, actualSH, x0, y0, x1, y1)
	}

	opts := &ebiten.DrawImageOptions{GeoM: geo}
	applyTintColor(opts, tint, alpha, sprite)
	target.DrawImage(subImg, opts)
}

func tileImagePatch(target *ebiten.Image, baseGeo ebiten.GeoM, img *ebiten.Image, sx0, sy0, sw, sh, dx, dy, dw, dh, alpha float64, tint *color.NRGBA, sprite *laya.Sprite, debugLabel string) {
	if sw <= 0 || sh <= 0 || dw <= 0 || dh <= 0 {
		return
	}

	cols := int(math.Ceil(dw / sw))
	rows := int(math.Ceil(dh / sh))

	if cols*rows > 2048 {
		log.Printf("[tileImagePatch] large tiling grid: %s cols=%d rows=%d srcPatch=(%.2f, %.2f)-(%.2f, %.2f) dst=(%.2f, %.2f)-(%.2f, %.2f)",
			debugLabel, cols, rows, sx0, sy0, sx0+sw, sy0+sh, dx, dy, dx+dw, dy+dh)
	}

	for y := 0; y < rows; y++ {
		offsetY := float64(y) * sh
		remainingH := math.Min(sh, dh-offsetY)
		if remainingH <= 0 {
			continue
		}

		for x := 0; x < cols; x++ {
			offsetX := float64(x) * sw
			remainingW := math.Min(sw, dw-offsetX)
			if remainingW <= 0 {
				continue
			}

			drawImagePatch(target, baseGeo, img, sx0, sy0, remainingW, remainingH, dx+offsetX, dy+offsetY, remainingW, remainingH, alpha, tint, sprite, debugLabel)
		}
	}
}

func applyTintColor(opts *ebiten.DrawImageOptions, tint *color.NRGBA, alpha float64, sprite *laya.Sprite) {
	if opts == nil {
		return
	}
	scaleR, scaleG, scaleB := 1.0, 1.0, 1.0
	scaleA := alpha
	if tint != nil {
		scaleR = float64(tint.R) / 255.0
		scaleG = float64(tint.G) / 255.0
		scaleB = float64(tint.B) / 255.0
		scaleA *= float64(tint.A) / 255.0
	}
	if scaleA < 0 {
		scaleA = 0
	} else if scaleA > 1 {
		scaleA = 1
	}
	opts.ColorM.Scale(scaleR, scaleG, scaleB, scaleA)
	applyColorEffects(opts, sprite)
}

func drawColor9Slice(target *ebiten.Image, baseGeo ebiten.GeoM, w, h float64, slice nineSlice, col color.Color) {
	if target == nil {
		return
	}
	if w <= 0 || h <= 0 {
		return
	}

	r, g, b, a := col.RGBA()
	if a == 0 {
		return
	}

	left := math.Max(float64(slice.left), 0)
	right := math.Max(float64(slice.right), 0)
	top := math.Max(float64(slice.top), 0)
	bottom := math.Max(float64(slice.bottom), 0)

	dstCols := splitSegments(w, left, right)
	dstRows := splitSegments(h, top, bottom)

	cr := float64(r) / 65535
	cg := float64(g) / 65535
	cb := float64(b) / 65535
	ca := float64(a) / 65535

	for yi := 0; yi < 3; yi++ {
		dy0 := dstRows[yi]
		dy1 := dstRows[yi+1]
		dstHeight := dy1 - dy0
		if dstHeight <= 0 {
			continue
		}
		for xi := 0; xi < 3; xi++ {
			dx0 := dstCols[xi]
			dx1 := dstCols[xi+1]
			dstWidth := dx1 - dx0
			if dstWidth <= 0 {
				continue
			}
			drawSolidPatch(target, baseGeo, dx0, dy0, dstWidth, dstHeight, cr, cg, cb, ca)
		}
	}
}

func drawSolidPatch(target *ebiten.Image, baseGeo ebiten.GeoM, dx, dy, dw, dh, r, g, b, a float64) {
	if dw <= 0 || dh <= 0 || a == 0 {
		return
	}

	img := solidImage()

	local := ebiten.GeoM{}
	local.Scale(dw, dh)
	local.Translate(dx, dy)

	geo := local
	geo.Concat(baseGeo)

	opts := &ebiten.DrawImageOptions{GeoM: geo}
	opts.ColorM.Scale(r, g, b, a)
	target.DrawImage(img, opts)
}

func splitSegments(total, leading, trailing float64) [4]float64 {
	if total <= 0 {
		return [4]float64{0, 0, 0, 0}
	}

	leading = math.Max(leading, 0)
	trailing = math.Max(trailing, 0)

	sum := leading + trailing
	if sum > total && sum > 0 {
		scale := total / sum
		leading *= scale
		trailing = total - leading
	}

	middleStart := leading
	middleEnd := total - trailing
	if middleEnd < middleStart {
		mid := (middleStart + middleEnd) / 2
		middleStart = mid
		middleEnd = mid
	}

	return [4]float64{0, middleStart, middleEnd, total}
}

func clampFloat(v, min, max float64) float64 {
	if max < min {
		max = min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func solidImage() *ebiten.Image {
	solidImageOnce.Do(func() {
		solidUnitImage = ebiten.NewImage(1, 1)
		solidUnitImage.Fill(color.White)
	})
	return solidUnitImage
}

// tileImagePatchWithFlip renders a tiled pattern with flip effects applied per-tile.
func tileImagePatchWithFlip(target *ebiten.Image, posGeo, flipGeo ebiten.GeoM, img *ebiten.Image, sx0, sy0, sw, sh, dx, dy, dw, dh, alpha float64, tint *color.NRGBA, sprite *laya.Sprite, debugLabel string) {
	if sw <= 0 || sh <= 0 || dw <= 0 || dh <= 0 {
		return
	}

	// 获取源图像区域
	bounds := img.Bounds()
	subX0 := bounds.Min.X + int(math.Round(sx0))
	subY0 := bounds.Min.Y + int(math.Round(sy0))
	subX1 := bounds.Min.X + int(math.Round(sx0+sw))
	subY1 := bounds.Min.Y + int(math.Round(sy0+sh))

	// 边界检查
	if subX0 < bounds.Min.X {
		subX0 = bounds.Min.X
	}
	if subY0 < bounds.Min.Y {
		subY0 = bounds.Min.Y
	}
	if subX1 > bounds.Max.X {
		subX1 = bounds.Max.X
	}
	if subY1 > bounds.Max.Y {
		subY1 = bounds.Max.Y
	}
	if subX1 <= subX0 || subY1 <= subY0 {
		return
	}

	sourceImg, ok := img.SubImage(image.Rect(subX0, subY0, subX1, subY1)).(*ebiten.Image)
	if !ok {
		return
	}

	// 检查是否需要翻转
	needFlip := false
	flipScaleX := flipGeo.Element(0, 0)
	flipScaleY := flipGeo.Element(1, 1)
	if flipScaleX < 0 || flipScaleY < 0 {
		needFlip = true
	}

	// 如果需要翻转，先创建翻转后的源图像
	var processedImg *ebiten.Image
	if needFlip {
		processedImg = ebiten.NewImage(int(sw), int(sh))
		opts := &ebiten.DrawImageOptions{}
		// 构建围绕中心点的翻转变换
		opts.GeoM.Translate(-sw/2, -sh/2)
		opts.GeoM.Scale(flipScaleX, flipScaleY)
		opts.GeoM.Translate(sw/2, sh/2)
		processedImg.DrawImage(sourceImg, opts)
	} else {
		processedImg = sourceImg
	}

	// 计算需要平铺的行列数
	cols := int(math.Ceil(dw / sw))
	rows := int(math.Ceil(dh / sh))

	if cols*rows > 2048 {
		log.Printf("[tileImagePatchWithFlip] large tiling grid: %s cols=%d rows=%d srcPatch=(%.2f, %.2f)-(%.2f, %.2f) dst=(%.2f, %.2f)-(%.2f, %.2f)",
			debugLabel, cols, rows, sx0, sy0, sx0+sw, sy0+sh, dx, dy, dx+dw, dy+dh)
	}

	// 平铺渲染每个tile
	for y := 0; y < rows; y++ {
		offsetY := float64(y) * sh

		for x := 0; x < cols; x++ {
			offsetX := float64(x) * sw

			// 检查这个平铺块是否在目标区域内
			if offsetX >= dw || offsetY >= dh {
				continue
			}

			// 计算这个平铺块的渲染区域
			renderW := sw
			renderH := sh

			// 如果平铺块超出边界，只渲染可见部分
			if offsetX+renderW > dw {
				renderW = dw - offsetX
			}
			if offsetY+renderH > dh {
				renderH = dh - offsetY
			}

			if renderW <= 0 || renderH <= 0 {
				continue
			}

			// 从源图像中裁剪需要的部分（边缘切割而非压缩）
			tileBounds := processedImg.Bounds()
			cropX0 := tileBounds.Min.X
			cropY0 := tileBounds.Min.Y
			cropX1 := tileBounds.Min.X + int(math.Round(renderW))
			cropY1 := tileBounds.Min.Y + int(math.Round(renderH))

			// 边界检查
			if cropX1 > tileBounds.Max.X {
				cropX1 = tileBounds.Max.X
			}
			if cropY1 > tileBounds.Max.Y {
				cropY1 = tileBounds.Max.Y
			}

			if cropX1 <= cropX0 || cropY1 <= cropY0 {
				continue
			}

			croppedTile := processedImg.SubImage(image.Rect(cropX0, cropY0, cropX1, cropY1)).(*ebiten.Image)
			if croppedTile == nil {
				continue
			}

			// 为每个平铺块计算变换（不缩放，直接平移到目标位置）
			local := ebiten.GeoM{}
			local.Translate(offsetX, offsetY)
			local.Concat(posGeo)

			opts := &ebiten.DrawImageOptions{}
			opts.GeoM = local
			// 翻转已在预处理阶段应用，这里只应用颜色效果
			applyTintColor(opts, tint, alpha, sprite)

			target.DrawImage(croppedTile, opts)
		}
	}
}

// drawImagePatchWithGeo renders an image patch using a pre-computed geometry transformation.
// Unlike drawImagePatch, this function doesn't build its own transformation matrix
// but uses the provided one directly.
func drawImagePatchWithGeo(target *ebiten.Image, geo ebiten.GeoM, img *ebiten.Image, sx0, sy0, sw, sh, dw, dh, alpha float64, tint *color.NRGBA, sprite *laya.Sprite, debugLabel string) {
	if dw <= 0 || dh <= 0 || sw <= 0 || sh <= 0 {
		return
	}

	bounds := img.Bounds()
	x0 := bounds.Min.X + int(math.Round(sx0))
	y0 := bounds.Min.Y + int(math.Round(sy0))
	x1 := bounds.Min.X + int(math.Round(sx0+sw))
	y1 := bounds.Min.Y + int(math.Round(sy0+sh))

	if x0 < bounds.Min.X {
		x0 = bounds.Min.X
	}
	if y0 < bounds.Min.Y {
		y0 = bounds.Min.Y
	}
	if x1 > bounds.Max.X {
		x1 = bounds.Max.X
	}
	if y1 > bounds.Max.Y {
		y1 = bounds.Max.Y
	}
	if x1 <= x0 || y1 <= y0 {
		return
	}

	actualSW := float64(x1 - x0)
	actualSH := float64(y1 - y0)
	if actualSW <= 0 || actualSH <= 0 {
		return
	}

	subImg, ok := img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
	if !ok {
		return
	}

	if dw > debugLargeDimensionLimit || dh > debugLargeDimensionLimit || math.IsNaN(dw) || math.IsNaN(dh) || math.IsInf(dw, 0) || math.IsInf(dh, 0) {
		log.Printf("[drawImagePatchWithGeo] suspicious dst patch: %s dst=(%.2f, %.2f) actualSrc=(%.2f, %.2f) srcRect=(%d,%d)-(%d,%d)",
			debugLabel, dw, dh, actualSW, actualSH, x0, y0, x1, y1)
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM = geo
	applyTintColor(opts, tint, alpha, sprite)
	target.DrawImage(subImg, opts)
}
