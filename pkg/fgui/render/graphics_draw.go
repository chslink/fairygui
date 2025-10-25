package render

import (
	"fmt"
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chslink/fairygui/internal/compat/laya"
)

// graphicsCache 缓存渲染后的 Graphics 结果，避免每帧重建
var (
	graphicsCache   = make(map[string]*ebiten.Image)
	graphicsCacheMu sync.RWMutex
)

func renderGraphicsSprite(target *ebiten.Image, sprite *laya.Sprite, parentGeo ebiten.GeoM, alpha float64) bool {
	if target == nil || sprite == nil {
		return false
	}
	gfx := sprite.Graphics()
	if gfx == nil || gfx.IsEmpty() {
		return false
	}

	commands := gfx.Commands()
	if len(commands) == 0 {
		return false
	}

	bounds := gfx.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return false
	}

	maxStroke := maxStrokeWidth(commands)
	pad := math.Ceil(maxStroke/2 + 1)

	canvasW := int(math.Ceil(bounds.W + pad*2))
	canvasH := int(math.Ceil(bounds.H + pad*2))
	if canvasW <= 0 || canvasH <= 0 {
		return false
	}

	// 生成缓存键：基于 sprite 的唯一标识和命令序列
	cacheKey := fmt.Sprintf("gfx_%p_%d_%dx%d", sprite, len(commands), canvasW, canvasH)

	// 尝试从缓存获取
	graphicsCacheMu.RLock()
	canvas, cached := graphicsCache[cacheKey]
	graphicsCacheMu.RUnlock()

	if !cached {
		// 缓存未命中，创建新图像
		canvas = ebiten.NewImage(canvasW, canvasH)
		offsetX := bounds.X - pad
		offsetY := bounds.Y - pad

		for _, cmd := range commands {
			switch cmd.Type {
			case laya.GraphicsCommandRect:
				drawRectCommand(canvas, &cmd, offsetX, offsetY, alpha)
			case laya.GraphicsCommandEllipse:
				drawEllipseCommand(canvas, &cmd, offsetX, offsetY, alpha)
			case laya.GraphicsCommandPolygon:
				drawPolygonCommand(canvas, &cmd, offsetX, offsetY, alpha)
			case laya.GraphicsCommandPath:
				drawPathCommand(canvas, &cmd, offsetX, offsetY, alpha)
			case laya.GraphicsCommandLine:
				drawLineCommand(canvas, cmd.Line, offsetX, offsetY, alpha)
			case laya.GraphicsCommandPie:
				drawPieCommand(canvas, cmd.Pie, offsetX, offsetY, alpha)
			default:
				// Texture and other commands are handled elsewhere; ignore for now.
			}
		}

		// 存入缓存
		graphicsCacheMu.Lock()
		graphicsCache[cacheKey] = canvas
		graphicsCacheMu.Unlock()
	}

	// 使用缓存的 canvas 进行绘制
	offsetX := bounds.X - pad
	offsetY := bounds.Y - pad
	geo := applyLocalOffset(parentGeo, offsetX, offsetY)
	opts := &ebiten.DrawImageOptions{GeoM: geo}
	applyColorEffects(opts, sprite)
	if alpha < 0 {
		alpha = 0
	} else if alpha > 1 {
		alpha = 1
	}
	opts.ColorM.Scale(1, 1, 1, alpha)
	target.DrawImage(canvas, opts)
	return true
}

func drawRectCommand(dst *ebiten.Image, cmd *laya.GraphicsCommand, offsetX, offsetY, globalAlpha float64) {
	if dst == nil || cmd == nil || cmd.Rect == nil {
		return
	}
	x := cmd.Rect.X - offsetX
	y := cmd.Rect.Y - offsetY
	w := cmd.Rect.W
	h := cmd.Rect.H
	if w <= 0 || h <= 0 {
		return
	}
	fill := colorFromFill(cmd.Rect.Fill, globalAlpha)
	stroke := colorFromStroke(cmd.Rect.Stroke, globalAlpha)
	if cmd.Rect.Radii != nil && len(cmd.Rect.Radii) > 0 {
		var path vector.Path
		if buildRoundedRectPath(&path, w, h, cmd.Rect.Radii, x, y) {
			drawPath(dst, &path, fill, stroke, strokeWidth(cmd.Rect.Stroke))
		}
		return
	}
	if fill != nil {
		vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), *fill, true)
	}
	if stroke != nil && strokeWidth(cmd.Rect.Stroke) > 0 {
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), float32(strokeWidth(cmd.Rect.Stroke)), *stroke, true)
	}
}

func drawEllipseCommand(dst *ebiten.Image, cmd *laya.GraphicsCommand, offsetX, offsetY, globalAlpha float64) {
	if dst == nil || cmd == nil || cmd.Ellipse == nil {
		return
	}
	w := cmd.Ellipse.Rx * 2
	h := cmd.Ellipse.Ry * 2
	x := cmd.Ellipse.Cx - cmd.Ellipse.Rx - offsetX
	y := cmd.Ellipse.Cy - cmd.Ellipse.Ry - offsetY
	if w <= 0 || h <= 0 {
		return
	}
	var path vector.Path
	if !buildEllipsePath(&path, float64(w), float64(h), float64(x), float64(y)) {
		return
	}
	fill := colorFromFill(cmd.Ellipse.Fill, globalAlpha)
	stroke := colorFromStroke(cmd.Ellipse.Stroke, globalAlpha)
	drawPath(dst, &path, fill, stroke, strokeWidth(cmd.Ellipse.Stroke))
}

func drawPolygonCommand(dst *ebiten.Image, cmd *laya.GraphicsCommand, offsetX, offsetY, globalAlpha float64) {
	if dst == nil || cmd == nil || cmd.Polygon == nil {
		return
	}
	var path vector.Path
	if !buildPolygonPath(&path, cmd.Polygon.Points, offsetX, offsetY) {
		return
	}
	fill := colorFromFill(cmd.Polygon.Fill, globalAlpha)
	stroke := colorFromStroke(cmd.Polygon.Stroke, globalAlpha)
	drawPath(dst, &path, fill, stroke, strokeWidth(cmd.Polygon.Stroke))
}

func drawPathCommand(dst *ebiten.Image, cmd *laya.GraphicsCommand, offsetX, offsetY, globalAlpha float64) {
	if dst == nil || cmd == nil || cmd.Path == nil {
		return
	}
	var path vector.Path
	if !buildGraphicsPath(&path, cmd.Path, offsetX, offsetY) {
		return
	}
	fill := colorFromFill(cmd.Path.Fill, globalAlpha)
	stroke := colorFromStroke(cmd.Path.Stroke, globalAlpha)
	drawPath(dst, &path, fill, stroke, strokeWidth(cmd.Path.Stroke))
}

func drawLineCommand(dst *ebiten.Image, line *laya.LineCommand, offsetX, offsetY, globalAlpha float64) {
	if dst == nil || line == nil || line.Stroke == nil {
		return
	}
	color := colorFromStroke(line.Stroke, globalAlpha)
	if color == nil {
		return
	}
	width := strokeWidth(line.Stroke)
	vector.StrokeLine(dst, float32(line.X0-offsetX), float32(line.Y0-offsetY), float32(line.X1-offsetX), float32(line.Y1-offsetY), float32(width), *color, true)
}

func drawPieCommand(dst *ebiten.Image, pie *laya.PieCommand, offsetX, offsetY, globalAlpha float64) {
	if dst == nil || pie == nil {
		return
	}
	var path vector.Path
	if !buildPiePath(&path, pie.Cx, pie.Cy, pie.Radius, pie.Start, pie.End, offsetX, offsetY) {
		return
	}
	fill := colorFromFill(pie.Fill, globalAlpha)
	stroke := colorFromStroke(pie.Stroke, globalAlpha)
	drawPath(dst, &path, fill, stroke, strokeWidth(pie.Stroke))
}

func drawPath(dst *ebiten.Image, path *vector.Path, fill, stroke *color.NRGBA, width float64) {
	if dst == nil || path == nil {
		return
	}
	if fill != nil {
		var opts vector.DrawPathOptions
		opts.AntiAlias = true
		opts.ColorScale.ScaleWithColor(*fill)
		vector.FillPath(dst, path, nil, &opts)
	}
	if stroke != nil && width > 0 {
		var strokeOpts vector.StrokeOptions
		strokeOpts.Width = float32(width)
		strokeOpts.LineJoin = vector.LineJoinRound
		strokeOpts.LineCap = vector.LineCapRound
		var drawOpts vector.DrawPathOptions
		drawOpts.AntiAlias = true
		drawOpts.ColorScale.ScaleWithColor(*stroke)
		vector.StrokePath(dst, path, &strokeOpts, &drawOpts)
	}
}

func buildGraphicsPath(path *vector.Path, command *laya.PathCommand, offsetX, offsetY float64) bool {
	if path == nil || command == nil || len(command.Commands) == 0 {
		return false
	}
	x := command.OffsetX - offsetX
	y := command.OffsetY - offsetY
	path.Reset()
	for i, cmd := range command.Commands {
		switch cmd.Op {
		case laya.PathOpMoveTo:
			if len(cmd.Args) >= 2 {
				path.MoveTo(float32(x+cmd.Args[0]), float32(y+cmd.Args[1]))
			}
		case laya.PathOpLineTo:
			if len(cmd.Args) >= 2 {
				path.LineTo(float32(x+cmd.Args[0]), float32(y+cmd.Args[1]))
			}
		case laya.PathOpArcTo:
			if len(cmd.Args) >= 4 {
				path.LineTo(float32(x+cmd.Args[2]), float32(y+cmd.Args[3]))
			}
		case laya.PathOpClosePath:
			if i > 0 {
				path.Close()
			}
		}
	}
	return true
}

func maxStrokeWidth(commands []laya.GraphicsCommand) float64 {
	max := 0.0
	for _, cmd := range commands {
		switch cmd.Type {
		case laya.GraphicsCommandRect:
			if w := strokeWidth(cmd.Rect.Stroke); w > max {
				max = w
			}
		case laya.GraphicsCommandEllipse:
			if w := strokeWidth(cmd.Ellipse.Stroke); w > max {
				max = w
			}
		case laya.GraphicsCommandPolygon:
			if w := strokeWidth(cmd.Polygon.Stroke); w > max {
				max = w
			}
		case laya.GraphicsCommandPath:
			if w := strokeWidth(cmd.Path.Stroke); w > max {
				max = w
			}
		case laya.GraphicsCommandLine:
			if w := strokeWidth(cmd.Line.Stroke); w > max {
				max = w
			}
		case laya.GraphicsCommandPie:
			if w := strokeWidth(cmd.Pie.Stroke); w > max {
				max = w
			}
		}
	}
	return max
}

func strokeWidth(stroke *laya.StrokeStyle) float64 {
	if stroke == nil || stroke.Width <= 0 {
		return 0
	}
	return stroke.Width
}

func colorFromFill(fill *laya.FillStyle, globalAlpha float64) *color.NRGBA {
	if fill == nil || fill.Color == "" {
		return nil
	}
	base := parseColor(fill.Color)
	if base == nil {
		return nil
	}
	alpha := float64(base.A) / 255 * globalAlpha
	if alpha <= 0 {
		return nil
	}
	out := *base
	out.A = uint8(math.Max(0, math.Min(255, alpha*255)))
	return &out
}

func colorFromStroke(stroke *laya.StrokeStyle, globalAlpha float64) *color.NRGBA {
	if stroke == nil || stroke.Color == "" || stroke.Width <= 0 {
		return nil
	}
	base := parseColor(stroke.Color)
	if base == nil {
		return nil
	}
	alpha := float64(base.A) / 255 * globalAlpha
	if alpha <= 0 {
		return nil
	}
	out := *base
	out.A = uint8(math.Max(0, math.Min(255, alpha*255)))
	return &out
}
