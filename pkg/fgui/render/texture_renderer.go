package render

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/hajimehoshi/ebiten/v2"
)

// TextureRenderer 负责渲染纹理命令（GImage、GLoader、GMovieClip 等）
type TextureRenderer struct {
	atlas *AtlasManager
}

// NewTextureRenderer 创建纹理渲染器
func NewTextureRenderer(atlas *AtlasManager) *TextureRenderer {
	return &TextureRenderer{atlas: atlas}
}

// Render 渲染纹理命令
func (r *TextureRenderer) Render(
	target *ebiten.Image,
	cmd *laya.TextureCommand,
	geo ebiten.GeoM,
	alpha float64,
	sprite *laya.Sprite,
) error {
	if target == nil || cmd == nil || cmd.Texture == nil {
		return nil
	}

	// 解析 PackageItem
	item, ok := cmd.Texture.(*assets.PackageItem)
	if !ok {
		return errors.New("texture_renderer: texture is not a PackageItem")
	}

	// 从 atlas 加载纹理
	spriteAny, err := r.atlas.ResolveSprite(item)
	if err != nil {
		return err
	}
	img, ok := spriteAny.(*ebiten.Image)
	if !ok || img == nil {
		return errors.New("texture_renderer: atlas returned unexpected sprite type")
	}

	// 解析颜色覆盖
	tint := parseColor(cmd.Color)

	// 计算源尺寸和目标尺寸
	bounds := img.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())
	dstW := cmd.Dest.W
	dstH := cmd.Dest.H
	if dstW <= 0 {
		dstW = srcW
	}
	if dstH <= 0 {
		dstH = srcH
	}

	// 获取 sprite offset（图像裁剪后的偏移）
	var spriteOffsetX, spriteOffsetY float64
	if spriteInfo := item.Sprite; spriteInfo != nil {
		spriteOffsetX = float64(spriteInfo.Offset.X)
		spriteOffsetY = float64(spriteInfo.Offset.Y)
	}

	// 根据模式选择渲染方法
	switch cmd.Mode {
	case laya.TextureModeScale9:
		return r.renderScale9(target, img, item, geo, dstW, dstH, srcW, srcH, alpha, tint, cmd, sprite, spriteOffsetX, spriteOffsetY)
	case laya.TextureModeTile:
		return r.renderTiled(target, img, geo, dstW, dstH, bounds, alpha, tint, cmd, sprite, spriteOffsetX, spriteOffsetY)
	default: // TextureModeSimple
		return r.renderSimple(target, img, geo, dstW, dstH, srcW, srcH, alpha, tint, sprite, cmd, spriteOffsetX, spriteOffsetY)
	}
}

// renderSimple 简单缩放渲染
func (r *TextureRenderer) renderSimple(
	target *ebiten.Image,
	img *ebiten.Image,
	parentGeo ebiten.GeoM,
	dstW, dstH, srcW, srcH float64,
	alpha float64,
	tint *color.NRGBA,
	sprite *laya.Sprite,
	cmd *laya.TextureCommand,
	spriteOffsetX, spriteOffsetY float64,
) error {
	if sprite != nil {
		if owner := sprite.Owner(); owner != nil {
			if gobj, ok := owner.(*core.GObject); ok {
				name := gobj.Name()
				if name == "n1" || name == "n2" || name == "n3" {
				}
			}
		}
	}

	// 构建完整的本地变换矩阵
	// 顺序：缩放到目标尺寸 → 翻转 → sprite offset → 命令偏移 → 父变换
	localGeo := ebiten.GeoM{}

	// 1. 计算并应用缩放到目标尺寸
	sx := 1.0
	sy := 1.0
	if srcW > 0 {
		sx = dstW / srcW
	}
	if srcH > 0 {
		sy = dstH / srcH
	}
	if sx != 1.0 || sy != 1.0 {
		localGeo.Scale(sx, sy)
	}

	// 2. 应用翻转
	if cmd.ScaleX != 1 || cmd.ScaleY != 1 {
		localGeo.Scale(cmd.ScaleX, cmd.ScaleY)
	}

	// 3. 应用 sprite offset（在翻转之后，避免被镜像）
	if spriteOffsetX != 0 || spriteOffsetY != 0 {
		localGeo.Translate(spriteOffsetX, spriteOffsetY)
	}

	// 4. 应用命令偏移（Widget 层指定的额外偏移）
	if cmd.OffsetX != 0 || cmd.OffsetY != 0 {
		localGeo.Translate(cmd.OffsetX, cmd.OffsetY)
	}

	// 5. 最后应用父变换
	localGeo.Concat(parentGeo)

	renderImageWithGeo(target, img, localGeo, alpha, tint, sprite)
	return nil
}

// renderScale9 九宫格渲染
func (r *TextureRenderer) renderScale9(
	target *ebiten.Image,
	img *ebiten.Image,
	item *assets.PackageItem,
	parentGeo ebiten.GeoM,
	dstW, dstH, srcW, srcH float64,
	alpha float64,
	tint *color.NRGBA,
	cmd *laya.TextureCommand,
	sprite *laya.Sprite,
	spriteOffsetX, spriteOffsetY float64,
) error {
	grid := cmd.Scale9Grid
	if grid == nil && item.Scale9Grid != nil {
		grid = &laya.Rect{
			X: float64(item.Scale9Grid.X),
			Y: float64(item.Scale9Grid.Y),
			W: float64(item.Scale9Grid.Width),
			H: float64(item.Scale9Grid.Height),
		}
	}
	if grid == nil {
		// 没有九宫格信息，降级到简单渲染
		return r.renderSimple(target, img, parentGeo, dstW, dstH, srcW, srcH, alpha, tint, sprite, cmd, spriteOffsetX, spriteOffsetY)
	}

	// 计算九宫格切片
	left := clampFloat(grid.X, 0, srcW)
	top := clampFloat(grid.Y, 0, srcH)
	right := clampFloat(srcW-grid.X-grid.W, 0, srcW)
	bottom := clampFloat(srcH-grid.Y-grid.H, 0, srcH)

	slice := nineSlice{
		left:   int(left),
		right:  int(right),
		top:    int(top),
		bottom: int(bottom),
	}

	debugLabel := fmt.Sprintf("texture=%s", item.ID)
	if debugNineSlice {
		logKey := fmt.Sprintf("scale9:%s:%.1fx%.1f:%d:%d:%d:%d:%t:%d",
			item.ID, dstW, dstH, slice.left, slice.right, slice.top, slice.bottom,
			cmd.ScaleByTile, cmd.TileGridIndice)
		if prev, ok := lastNineSliceLog.Load(item.ID); !ok || prev != logKey {
			log.Printf("[texture_renderer][9slice] id=%s dst=%.1fx%.1f src=%dx%d slice={L:%d R:%d T:%d B:%d} tile=%v grid=%d",
				item.ID, dstW, dstH, img.Bounds().Dx(), img.Bounds().Dy(),
				slice.left, slice.right, slice.top, slice.bottom, cmd.ScaleByTile, cmd.TileGridIndice)
			lastNineSliceLog.Store(item.ID, logKey)
		}
	}

	// 构建本地变换矩阵
	// 顺序：翻转 → sprite offset → 命令偏移 → 父变换
	localGeo := ebiten.GeoM{}

	// 1. 应用翻转
	if cmd.ScaleX != 1 || cmd.ScaleY != 1 {
		localGeo.Scale(cmd.ScaleX, cmd.ScaleY)
	}

	// 2. 应用 sprite offset（在翻转之后，避免被镜像）
	if spriteOffsetX != 0 || spriteOffsetY != 0 {
		localGeo.Translate(spriteOffsetX, spriteOffsetY)
	}

	// 3. 应用命令偏移
	if cmd.OffsetX != 0 || cmd.OffsetY != 0 {
		localGeo.Translate(cmd.OffsetX, cmd.OffsetY)
	}

	// 4. 应用父变换
	localGeo.Concat(parentGeo)

	drawNineSlice(target, localGeo, img, slice, dstW, dstH, alpha, tint,
		cmd.ScaleByTile, cmd.TileGridIndice, sprite, debugLabel)

	if debugNineSliceOverlayEnabled {
		drawNineSliceOverlay(target, localGeo, slice, dstW, dstH)
	}

	return nil
}

// renderTiled 平铺渲染
func (r *TextureRenderer) renderTiled(
	target *ebiten.Image,
	img *ebiten.Image,
	parentGeo ebiten.GeoM,
	dstW, dstH float64,
	bounds image.Rectangle,
	alpha float64,
	tint *color.NRGBA,
	cmd *laya.TextureCommand,
	sprite *laya.Sprite,
	spriteOffsetX, spriteOffsetY float64,
) error {
	debugLabel := fmt.Sprintf("tile-texture")
	if debugNineSlice {
		log.Printf("[texture_renderer][tile] dst=%.1fx%.1f src=%dx%d tile=%v grid=%d",
			dstW, dstH, bounds.Dx(), bounds.Dy(), cmd.ScaleByTile, cmd.TileGridIndice)
	}

	// 构建本地变换矩阵
	// 顺序：sprite offset → 父变换
	// 注意：翻转在 tileImagePatchWithFlip 内部围绕中心处理，
	// 不需要 cmd.OffsetX/Y，因为那会导致位置错误
	localGeo := ebiten.GeoM{}

	// 1. 应用 sprite offset
	if spriteOffsetX != 0 || spriteOffsetY != 0 {
		localGeo.Translate(spriteOffsetX, spriteOffsetY)
	}

	// 2. 翻转变换（单独构建，传递给平铺函数）
	flipGeo := ebiten.GeoM{}
	if cmd.ScaleX != 1 || cmd.ScaleY != 1 {
		flipGeo.Scale(cmd.ScaleX, cmd.ScaleY)
	}

	// 注意：不应用 cmd.OffsetX/Y，因为翻转在平铺函数内部围绕中心处理

	// 3. 应用父变换
	localGeo.Concat(parentGeo)

	// 平铺渲染
	tileImagePatchWithFlip(target, localGeo, flipGeo, img,
		0, 0, float64(bounds.Dx()), float64(bounds.Dy()),
		0, 0, dstW, dstH,
		alpha, tint, sprite, debugLabel)

	return nil
}
