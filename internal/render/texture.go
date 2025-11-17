package render

import (
	"image"

	"github.com/chslink/fairygui"
	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// 纹理渲染 - 九宫格和平铺
// ============================================================================

// DrawNineSliceTexture 绘制九宫格纹理。
//
// 九宫格将纹理分成 9 个部分：
//
//	+-----+-------+-----+
//	|  1  |   2   |  3  |  (top)
//	+-----+-------+-----+
//	|  4  |   5   |  6  |  (middle)
//	+-----+-------+-----+
//	|  7  |   8   |  9  |  (bottom)
//	+-----+-------+-----+
//	 left  center right
//
// 四个角（1, 3, 7, 9）保持原始尺寸不缩放。
// 边缘（2, 4, 6, 8）只在一个方向缩放。
// 中心（5）在两个方向缩放。
func (r *EbitenRenderer) DrawNineSliceTexture(
	screen *ebiten.Image,
	texture *ebiten.Image,
	options fairygui.DrawOptions,
	baseOpts *ebiten.DrawImageOptions,
) {
	ns := options.NineSlice
	if ns == nil {
		// 没有九宫格，直接绘制
		screen.DrawImage(texture, baseOpts)
		r.drawCalls++
		return
	}

	// 源纹理尺寸
	srcBounds := texture.Bounds()
	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())

	// 目标尺寸
	targetWidth := options.Width
	targetHeight := options.Height
	if targetWidth == 0 {
		targetWidth = srcW
	}
	if targetHeight == 0 {
		targetHeight = srcH
	}

	// 如果目标尺寸小于九宫格边距，直接缩放绘制
	if targetWidth < ns.Left+ns.Right || targetHeight < ns.Top+ns.Bottom {
		opts := *baseOpts
		scaleX := targetWidth / srcW
		scaleY := targetHeight / srcH
		opts.GeoM.Scale(scaleX, scaleY)
		screen.DrawImage(texture, &opts)
		r.drawCalls++
		return
	}

	// 计算九宫格各部分的尺寸
	// 源尺寸
	leftWidth := ns.Left
	rightWidth := ns.Right
	topHeight := ns.Top
	bottomHeight := ns.Bottom
	centerWidth := srcW - leftWidth - rightWidth
	middleHeight := srcH - topHeight - bottomHeight

	// 目标尺寸
	targetCenterWidth := targetWidth - leftWidth - rightWidth
	targetMiddleHeight := targetHeight - topHeight - bottomHeight

	// 绘制 9 个部分
	parts := []struct {
		srcX, srcY, srcW, srcH       float64 // 源矩形
		dstX, dstY, dstW, dstH       float64 // 目标矩形
		scaleX, scaleY               float64 // 缩放比例
	}{
		// 1. 左上角
		{0, 0, leftWidth, topHeight, 0, 0, leftWidth, topHeight, 1, 1},
		// 2. 顶部中间
		{leftWidth, 0, centerWidth, topHeight, leftWidth, 0, targetCenterWidth, topHeight, targetCenterWidth / centerWidth, 1},
		// 3. 右上角
		{srcW - rightWidth, 0, rightWidth, topHeight, targetWidth - rightWidth, 0, rightWidth, topHeight, 1, 1},
		// 4. 左中间
		{0, topHeight, leftWidth, middleHeight, 0, topHeight, leftWidth, targetMiddleHeight, 1, targetMiddleHeight / middleHeight},
		// 5. 中心
		{leftWidth, topHeight, centerWidth, middleHeight, leftWidth, topHeight, targetCenterWidth, targetMiddleHeight, targetCenterWidth / centerWidth, targetMiddleHeight / middleHeight},
		// 6. 右中间
		{srcW - rightWidth, topHeight, rightWidth, middleHeight, targetWidth - rightWidth, topHeight, rightWidth, targetMiddleHeight, 1, targetMiddleHeight / middleHeight},
		// 7. 左下角
		{0, srcH - bottomHeight, leftWidth, bottomHeight, 0, targetHeight - bottomHeight, leftWidth, bottomHeight, 1, 1},
		// 8. 底部中间
		{leftWidth, srcH - bottomHeight, centerWidth, bottomHeight, leftWidth, targetHeight - bottomHeight, targetCenterWidth, bottomHeight, targetCenterWidth / centerWidth, 1},
		// 9. 右下角
		{srcW - rightWidth, srcH - bottomHeight, rightWidth, bottomHeight, targetWidth - rightWidth, targetHeight - bottomHeight, rightWidth, bottomHeight, 1, 1},
	}

	for _, part := range parts {
		// 跳过空的部分
		if part.srcW <= 0 || part.srcH <= 0 || part.dstW <= 0 || part.dstH <= 0 {
			continue
		}

		// 创建子图像
		srcRect := image.Rect(
			int(part.srcX), int(part.srcY),
			int(part.srcX+part.srcW), int(part.srcY+part.srcH),
		)
		subImage := texture.SubImage(srcRect).(*ebiten.Image)

		// 创建绘制选项
		opts := *baseOpts
		opts.GeoM.Reset()

		// 应用缩放
		if part.scaleX != 1.0 || part.scaleY != 1.0 {
			opts.GeoM.Scale(part.scaleX, part.scaleY)
		}

		// 应用位置偏移
		opts.GeoM.Translate(options.X+part.dstX, options.Y+part.dstY)

		// 绘制
		screen.DrawImage(subImage, &opts)
		r.drawCalls++
	}
}

// DrawTilingTexture 绘制平铺纹理。
//
// 平铺模式会重复绘制纹理填满目标区域。
func (r *EbitenRenderer) DrawTilingTexture(
	screen *ebiten.Image,
	texture *ebiten.Image,
	options fairygui.DrawOptions,
	baseOpts *ebiten.DrawImageOptions,
) {
	// 源纹理尺寸
	srcBounds := texture.Bounds()
	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())

	// 目标尺寸
	targetWidth := options.Width
	targetHeight := options.Height
	if targetWidth == 0 {
		targetWidth = srcW
	}
	if targetHeight == 0 {
		targetHeight = srcH
	}

	// 计算需要平铺的次数
	tilesX := int(targetWidth/srcW) + 1
	tilesY := int(targetHeight/srcH) + 1

	for y := 0; y < tilesY; y++ {
		for x := 0; x < tilesX; x++ {
			// 计算当前 tile 的位置
			tileX := float64(x) * srcW
			tileY := float64(y) * srcH

			// 如果超出目标区域，跳过
			if tileX >= targetWidth || tileY >= targetHeight {
				continue
			}

			// 计算当前 tile 的实际尺寸（可能需要裁剪）
			tileW := srcW
			tileH := srcH
			if tileX+tileW > targetWidth {
				tileW = targetWidth - tileX
			}
			if tileY+tileH > targetHeight {
				tileH = targetHeight - tileY
			}

			// 创建子图像（如果需要裁剪）
			var subImage *ebiten.Image
			if tileW < srcW || tileH < srcH {
				srcRect := image.Rect(0, 0, int(tileW), int(tileH))
				subImage = texture.SubImage(srcRect).(*ebiten.Image)
			} else {
				subImage = texture
			}

			// 创建绘制选项
			opts := *baseOpts
			opts.GeoM.Reset()
			opts.GeoM.Translate(options.X+tileX, options.Y+tileY)

			// 绘制
			screen.DrawImage(subImage, &opts)
			r.drawCalls++
		}
	}
}

// DrawScaledTexture 绘制缩放纹理（普通模式）。
func (r *EbitenRenderer) DrawScaledTexture(
	screen *ebiten.Image,
	texture *ebiten.Image,
	options fairygui.DrawOptions,
	baseOpts *ebiten.DrawImageOptions,
) {
	opts := *baseOpts

	// 如果指定了宽高，应用缩放
	if options.Width > 0 || options.Height > 0 {
		srcBounds := texture.Bounds()
		srcW := float64(srcBounds.Dx())
		srcH := float64(srcBounds.Dy())

		width := options.Width
		height := options.Height
		if width == 0 {
			width = srcW
		}
		if height == 0 {
			height = srcH
		}

		scaleX := width / srcW
		scaleY := height / srcH
		opts.GeoM.Scale(scaleX, scaleY)
	}

	screen.DrawImage(texture, &opts)
	r.drawCalls++
}
