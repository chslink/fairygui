package laya

import "math"

// GraphicsCommandType 枚举 Graphics 中记录的绘制命令类别。
type GraphicsCommandType int

const (
	GraphicsCommandUnknown GraphicsCommandType = iota
	GraphicsCommandPath
	GraphicsCommandRect
	GraphicsCommandEllipse
	GraphicsCommandPolygon
	GraphicsCommandTexture
	GraphicsCommandLine
	GraphicsCommandPie
)

// FillStyle 描述填充颜色。
type FillStyle struct {
	Color string
}

// StrokeStyle 描述描边颜色及线宽。
type StrokeStyle struct {
	Color string
	Width float64
}

// GraphicsPathOp 表示路径命令。
type GraphicsPathOp int

const (
	PathOpMoveTo GraphicsPathOp = iota
	PathOpLineTo
	PathOpArcTo
	PathOpClosePath
)

// GraphicsPathCommand 记录单条路径命令及其参数。
type GraphicsPathCommand struct {
	Op   GraphicsPathOp
	Args []float64
}

// PathCommand 保存 drawPath 的参数。
type PathCommand struct {
	OffsetX  float64
	OffsetY  float64
	Commands []GraphicsPathCommand
	Fill     *FillStyle
	Stroke   *StrokeStyle
}

// RectCommand 记录矩形绘制参数。
type RectCommand struct {
	X      float64
	Y      float64
	W      float64
	H      float64
	Radii  []float64
	Fill   *FillStyle
	Stroke *StrokeStyle
}

// EllipseCommand 记录椭圆/圆绘制参数。
type EllipseCommand struct {
	Cx     float64
	Cy     float64
	Rx     float64
	Ry     float64
	Fill   *FillStyle
	Stroke *StrokeStyle
}

// PolygonCommand 记录多边形绘制参数。
type PolygonCommand struct {
	Points []float64
	Fill   *FillStyle
	Stroke *StrokeStyle
}

// TextureCommandMode 描述纹理绘制模式。
type TextureCommandMode int

const (
	TextureModeSimple TextureCommandMode = iota
	TextureModeScale9
	TextureModeTile
)

// TextureCommand 记录纹理绘制参数。
type TextureCommand struct {
	Texture        any
	Mode           TextureCommandMode
	Dest           Rect
	OffsetX        float64
	OffsetY        float64
	Scale9Grid     *Rect
	ScaleByTile    bool
	TileGridIndice int
	Color          string
	ScaleX         float64
	ScaleY         float64
}

// GraphicsCommand 表示一条绘制命令。
type GraphicsCommand struct {
	Type    GraphicsCommandType
	Path    *PathCommand
	Rect    *RectCommand
	Ellipse *EllipseCommand
	Polygon *PolygonCommand
	Texture *TextureCommand
	Line    *LineCommand
	Pie     *PieCommand
}

// Graphics 模拟 Laya.Graphics，记录绘制命令序列。
type Graphics struct {
	commands []GraphicsCommand
	version  uint64
}

// LineCommand 记录线段绘制参数。
type LineCommand struct {
	X0     float64
	Y0     float64
	X1     float64
	Y1     float64
	Stroke *StrokeStyle
}

// PieCommand 记录扇形绘制参数（角度为度）。
type PieCommand struct {
	Cx     float64
	Cy     float64
	Radius float64
	Start  float64
	End    float64
	Fill   *FillStyle
	Stroke *StrokeStyle
}

// NewGraphics 创建 Graphics。
func NewGraphics() *Graphics {
	return &Graphics{}
}

// Clear 清空当前所有命令。
func (g *Graphics) Clear() {
	if g == nil {
		return
	}
	if len(g.commands) == 0 {
		return
	}
	g.commands = nil
	g.version++
}

// DrawRect 记录矩形绘制命令。
func (g *Graphics) DrawRect(x, y, w, h float64, fill *FillStyle, stroke *StrokeStyle) {
	if g == nil {
		return
	}
	cmd := GraphicsCommand{
		Type: GraphicsCommandRect,
		Rect: &RectCommand{
			X:      x,
			Y:      y,
			W:      w,
			H:      h,
			Radii:  nil,
			Fill:   cloneFill(fill),
			Stroke: cloneStroke(stroke),
		},
	}
	g.commands = append(g.commands, cmd)
	g.version++
}

// DrawRoundRect 记录带圆角的矩形绘制命令。
func (g *Graphics) DrawRoundRect(x, y, w, h float64, radii []float64, fill *FillStyle, stroke *StrokeStyle) {
	if g == nil {
		return
	}
	cmd := GraphicsCommand{
		Type: GraphicsCommandRect,
		Rect: &RectCommand{
			X:      x,
			Y:      y,
			W:      w,
			H:      h,
			Radii:  copyFloat64s(radii),
			Fill:   cloneFill(fill),
			Stroke: cloneStroke(stroke),
		},
	}
	g.commands = append(g.commands, cmd)
	g.version++
}

// DrawEllipse 记录椭圆绘制命令。
func (g *Graphics) DrawEllipse(cx, cy, rx, ry float64, fill *FillStyle, stroke *StrokeStyle) {
	if g == nil {
		return
	}
	cmd := GraphicsCommand{
		Type: GraphicsCommandEllipse,
		Ellipse: &EllipseCommand{
			Cx:     cx,
			Cy:     cy,
			Rx:     rx,
			Ry:     ry,
			Fill:   cloneFill(fill),
			Stroke: cloneStroke(stroke),
		},
	}
	g.commands = append(g.commands, cmd)
	g.version++
}

// DrawPolygon 记录多边形绘制命令。
func (g *Graphics) DrawPolygon(points []float64, fill *FillStyle, stroke *StrokeStyle) {
	if g == nil || len(points) < 6 {
		return
	}
	cmd := GraphicsCommand{
		Type: GraphicsCommandPolygon,
		Polygon: &PolygonCommand{
			Points: copyFloat64s(points),
			Fill:   cloneFill(fill),
			Stroke: cloneStroke(stroke),
		},
	}
	g.commands = append(g.commands, cmd)
	g.version++
}

// DrawLine 记录直线绘制命令。
func (g *Graphics) DrawLine(x0, y0, x1, y1 float64, stroke *StrokeStyle) {
	if g == nil || stroke == nil {
		return
	}
	cmd := GraphicsCommand{
		Type: GraphicsCommandLine,
		Line: &LineCommand{
			X0:     x0,
			Y0:     y0,
			X1:     x1,
			Y1:     y1,
			Stroke: cloneStroke(stroke),
		},
	}
	g.commands = append(g.commands, cmd)
	g.version++
}

// DrawPie 记录扇形绘制命令。
func (g *Graphics) DrawPie(cx, cy, radius, startAngle, endAngle float64, fill *FillStyle, stroke *StrokeStyle) {
	if g == nil {
		return
	}
	cmd := GraphicsCommand{
		Type: GraphicsCommandPie,
		Pie: &PieCommand{
			Cx:     cx,
			Cy:     cy,
			Radius: radius,
			Start:  startAngle,
			End:    endAngle,
			Fill:   cloneFill(fill),
			Stroke: cloneStroke(stroke),
		},
	}
	g.commands = append(g.commands, cmd)
	g.version++
}

// DrawPath 记录路径命令。
func (g *Graphics) DrawPath(offsetX, offsetY float64, commands []GraphicsPathCommand, fill *FillStyle, stroke *StrokeStyle) {
	if g == nil || len(commands) == 0 {
		return
	}
	copied := make([]GraphicsPathCommand, len(commands))
	for i, cmd := range commands {
		copied[i] = GraphicsPathCommand{
			Op:   cmd.Op,
			Args: copyFloat64s(cmd.Args),
		}
	}
	cmd := GraphicsCommand{
		Type: GraphicsCommandPath,
		Path: &PathCommand{
			OffsetX:  offsetX,
			OffsetY:  offsetY,
			Commands: copied,
			Fill:     cloneFill(fill),
			Stroke:   cloneStroke(stroke),
		},
	}
	g.commands = append(g.commands, cmd)
	g.version++
}

// DrawTexture 记录纹理绘制命令。
func (g *Graphics) DrawTexture(cmd TextureCommand) {
	if g == nil || cmd.Texture == nil {
		return
	}
	copy := TextureCommand{
		Texture:        cmd.Texture,
		Mode:           cmd.Mode,
		Dest:           cmd.Dest,
		OffsetX:        cmd.OffsetX,
		OffsetY:        cmd.OffsetY,
		ScaleByTile:    cmd.ScaleByTile,
		TileGridIndice: cmd.TileGridIndice,
		Color:          cmd.Color,
		ScaleX:         cmd.ScaleX,
		ScaleY:         cmd.ScaleY,
	}
	if cmd.Scale9Grid != nil {
		rect := *cmd.Scale9Grid
		copy.Scale9Grid = &rect
	}
	if copy.ScaleX == 0 {
		copy.ScaleX = 1
	}
	if copy.ScaleY == 0 {
		copy.ScaleY = 1
	}
	g.commands = append(g.commands, GraphicsCommand{
		Type:    GraphicsCommandTexture,
		Texture: &copy,
	})
	g.version++
}

// Commands 返回命令序列（不要修改返回 slice）。
func (g *Graphics) Commands() []GraphicsCommand {
	if g == nil {
		return nil
	}
	return g.commands
}

// Bounds 返回绘制命令的包围盒。
func (g *Graphics) Bounds() Rect {
	if g == nil {
		return Rect{}
	}
	return computeBounds(g)
}

// IsEmpty 判断是否无命令。
func (g *Graphics) IsEmpty() bool {
	return g == nil || len(g.commands) == 0
}

// Version 返回命令序列版本号。
func (g *Graphics) Version() uint64 {
	if g == nil {
		return 0
	}
	return g.version
}

// GraphicsHitArea 命中测试区域。
type HitArea struct {
	graphics        *Graphics
	cachedVersion   uint64
	cachedBounds    Rect
	hasCachedBounds bool
}

// NewHitArea 创建命中区域。
func NewHitArea() *HitArea {
	return &HitArea{}
}

// SetGraphics 将命中区域绑定到指定 Graphics。
func (h *HitArea) SetGraphics(gfx *Graphics) {
	h.graphics = gfx
	h.cachedVersion = 0
	h.hasCachedBounds = false
}

// Contains 判断本地坐标点是否命中。
func (h *HitArea) Contains(x, y float64) bool {
	if h == nil || h.graphics == nil {
		return false
	}
	if h.cachedVersion != h.graphics.Version() || !h.hasCachedBounds {
		h.cachedBounds = computeBounds(h.graphics)
		h.cachedVersion = h.graphics.Version()
		h.hasCachedBounds = true
	}
	if !h.cachedBounds.Contains(Point{X: x, Y: y}) {
		return false
	}
	// 进一步依据命令类型精细判断。
	for _, cmd := range h.graphics.Commands() {
		switch cmd.Type {
		case GraphicsCommandRect:
			if rectContains(cmd.Rect, x, y) {
				return true
			}
		case GraphicsCommandEllipse:
			if ellipseContains(cmd.Ellipse, x, y) {
				return true
			}
		case GraphicsCommandPolygon:
			if polygonContains(cmd.Polygon, x, y) {
				return true
			}
		case GraphicsCommandPath:
			if pathContains(cmd.Path, x, y) {
				return true
			}
		case GraphicsCommandTexture:
			if textureContains(cmd.Texture, x, y) {
				return true
			}
		case GraphicsCommandLine:
			if lineContains(cmd.Line, x, y) {
				return true
			}
		case GraphicsCommandPie:
			if pieContains(cmd.Pie, x, y) {
				return true
			}
		}
	}
	return false
}

func computeBounds(gfx *Graphics) Rect {
	bounds := Rect{}
	first := true
	for _, cmd := range gfx.Commands() {
		var cmdBounds Rect
		switch cmd.Type {
		case GraphicsCommandRect:
			cmdBounds = Rect{X: cmd.Rect.X, Y: cmd.Rect.Y, W: cmd.Rect.W, H: cmd.Rect.H}
		case GraphicsCommandEllipse:
			cmdBounds = Rect{
				X: cmd.Ellipse.Cx - cmd.Ellipse.Rx,
				Y: cmd.Ellipse.Cy - cmd.Ellipse.Ry,
				W: cmd.Ellipse.Rx * 2,
				H: cmd.Ellipse.Ry * 2,
			}
		case GraphicsCommandPolygon:
			cmdBounds = polygonBounds(cmd.Polygon.Points)
		case GraphicsCommandPath:
			cmdBounds = pathBounds(cmd.Path)
		case GraphicsCommandTexture:
			if cmd.Texture != nil {
				cmdBounds = Rect{
					X: cmd.Texture.Dest.X + cmd.Texture.OffsetX,
					Y: cmd.Texture.Dest.Y + cmd.Texture.OffsetY,
					W: cmd.Texture.Dest.W,
					H: cmd.Texture.Dest.H,
				}
			}
		case GraphicsCommandLine:
			cmdBounds = lineBounds(cmd.Line)
		case GraphicsCommandPie:
			cmdBounds = pieBounds(cmd.Pie)
		default:
			continue
		}
		if first {
			bounds = cmdBounds
			first = false
		} else {
			bounds = unionRect(bounds, cmdBounds)
		}
	}
	if first {
		return Rect{}
	}
	return bounds
}

func unionRect(a, b Rect) Rect {
	minX := math.Min(a.X, b.X)
	minY := math.Min(a.Y, b.Y)
	maxX := math.Max(a.Right(), b.Right())
	maxY := math.Max(a.Bottom(), b.Bottom())
	return Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

func rectContains(rect *RectCommand, x, y float64) bool {
	if rect == nil {
		return false
	}
	if x < rect.X || x > rect.X+rect.W || y < rect.Y || y > rect.Y+rect.H {
		return false
	}
	if len(rect.Radii) == 0 {
		return true
	}
	rTopLeft := cornerRadius(rect.Radii, 0)
	rTopRight := cornerRadius(rect.Radii, 1)
	rBottomRight := cornerRadius(rect.Radii, 2)
	rBottomLeft := cornerRadius(rect.Radii, 3)

	if rTopLeft > 0 {
		cx := rect.X + rTopLeft
		cy := rect.Y + rTopLeft
		if x < cx && y < cy && !pointInCircle(x, y, cx, cy, rTopLeft) {
			return false
		}
	}
	if rTopRight > 0 {
		cx := rect.X + rect.W - rTopRight
		cy := rect.Y + rTopRight
		if x > cx && y < cy && !pointInCircle(x, y, cx, cy, rTopRight) {
			return false
		}
	}
	if rBottomRight > 0 {
		cx := rect.X + rect.W - rBottomRight
		cy := rect.Y + rect.H - rBottomRight
		if x > cx && y > cy && !pointInCircle(x, y, cx, cy, rBottomRight) {
			return false
		}
	}
	if rBottomLeft > 0 {
		cx := rect.X + rBottomLeft
		cy := rect.Y + rect.H - rBottomLeft
		if x < cx && y > cy && !pointInCircle(x, y, cx, cy, rBottomLeft) {
			return false
		}
	}
	return true
}

func ellipseContains(circle *EllipseCommand, x, y float64) bool {
	if circle == nil || circle.Rx == 0 || circle.Ry == 0 {
		return false
	}
	dx := (x - circle.Cx) / circle.Rx
	dy := (y - circle.Cy) / circle.Ry
	return dx*dx+dy*dy <= 1
}

func polygonContains(poly *PolygonCommand, x, y float64) bool {
	if poly == nil || len(poly.Points) < 6 {
		return false
	}
	hit := false
	j := len(poly.Points) - 2
	for i := 0; i < len(poly.Points); i += 2 {
		x0 := poly.Points[i]
		y0 := poly.Points[i+1]
		x1 := poly.Points[j]
		y1 := poly.Points[j+1]
		intersect := ((y0 > y) != (y1 > y)) &&
			x < (x1-x0)*(y-y0)/(y1-y0+1e-9)+x0
		if intersect {
			hit = !hit
		}
		j = i
	}
	return hit
}

func pathContains(path *PathCommand, x, y float64) bool {
	if path == nil {
		return false
	}
	bounds := pathBounds(path)
	return bounds.Contains(Point{X: x, Y: y})
}

func textureContains(tex *TextureCommand, x, y float64) bool {
	if tex == nil {
		return false
	}
	minX := tex.Dest.X + tex.OffsetX
	minY := tex.Dest.Y + tex.OffsetY
	maxX := minX + tex.Dest.W
	maxY := minY + tex.Dest.H
	return x >= minX && x <= maxX && y >= minY && y <= maxY
}

func polygonBounds(points []float64) Rect {
	if len(points) < 2 {
		return Rect{}
	}
	minX, maxX := points[0], points[0]
	minY, maxY := points[1], points[1]
	for i := 2; i < len(points); i += 2 {
		if points[i] < minX {
			minX = points[i]
		}
		if points[i] > maxX {
			maxX = points[i]
		}
		if points[i+1] < minY {
			minY = points[i+1]
		}
		if points[i+1] > maxY {
			maxY = points[i+1]
		}
	}
	return Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

func pathBounds(path *PathCommand) Rect {
	if path == nil || len(path.Commands) == 0 {
		return Rect{}
	}
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	update := func(px, py float64) {
		x := path.OffsetX + px
		y := path.OffsetY + py
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	for _, cmd := range path.Commands {
		switch cmd.Op {
		case PathOpMoveTo, PathOpLineTo:
			if len(cmd.Args) >= 2 {
				update(cmd.Args[0], cmd.Args[1])
			}
		case PathOpArcTo:
			if len(cmd.Args) >= 4 {
				update(cmd.Args[0], cmd.Args[1])
				update(cmd.Args[2], cmd.Args[3])
			}
		}
	}
	if minX == math.MaxFloat64 {
		return Rect{}
	}
	return Rect{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
}

func copyFloat64s(src []float64) []float64 {
	if len(src) == 0 {
		return nil
	}
	dst := make([]float64, len(src))
	copy(dst, src)
	return dst
}

func cloneFill(src *FillStyle) *FillStyle {
	if src == nil {
		return nil
	}
	out := *src
	return &out
}

func cloneStroke(src *StrokeStyle) *StrokeStyle {
	if src == nil {
		return nil
	}
	out := *src
	return &out
}

func cornerRadius(values []float64, index int) float64 {
	if index < len(values) {
		return math.Max(0, values[index])
	}
	if len(values) > 0 {
		return math.Max(0, values[len(values)-1])
	}
	return 0
}

func pointInCircle(x, y, cx, cy, radius float64) bool {
	dx := x - cx
	dy := y - cy
	return dx*dx+dy*dy <= radius*radius
}

func lineBounds(line *LineCommand) Rect {
	if line == nil {
		return Rect{}
	}
	minX := math.Min(line.X0, line.X1)
	maxX := math.Max(line.X0, line.X1)
	minY := math.Min(line.Y0, line.Y1)
	maxY := math.Max(line.Y0, line.Y1)
	width := maxX - minX
	height := maxY - minY
	stroke := 0.0
	if line.Stroke != nil && line.Stroke.Width > 0 {
		stroke = line.Stroke.Width
	}
	pad := stroke / 2
	return Rect{X: minX - pad, Y: minY - pad, W: width + stroke, H: height + stroke}
}

func lineContains(line *LineCommand, x, y float64) bool {
	if line == nil || line.Stroke == nil || line.Stroke.Width <= 0 {
		return false
	}
	lengthSquared := (line.X1-line.X0)*(line.X1-line.X0) + (line.Y1-line.Y0)*(line.Y1-line.Y0)
	if lengthSquared == 0 {
		return math.Hypot(x-line.X0, y-line.Y0) <= line.Stroke.Width/2
	}
	t := ((x-line.X0)*(line.X1-line.X0) + (y-line.Y0)*(line.Y1-line.Y0)) / lengthSquared
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	projX := line.X0 + t*(line.X1-line.X0)
	projY := line.Y0 + t*(line.Y1-line.Y0)
	dist := math.Hypot(x-projX, y-projY)
	return dist <= line.Stroke.Width/2
}

func pieBounds(pie *PieCommand) Rect {
	if pie == nil {
		return Rect{}
	}
	r := pie.Radius
	return Rect{X: pie.Cx - r, Y: pie.Cy - r, W: r * 2, H: r * 2}
}

func pieContains(pie *PieCommand, x, y float64) bool {
	if pie == nil {
		return false
	}
	if !pointInCircle(x, y, pie.Cx, pie.Cy, pie.Radius) {
		return false
	}
	angle := math.Atan2(y-pie.Cy, x-pie.Cx) * 180 / math.Pi
	if angle < 0 {
		angle += 360
	}
	start := normalizeAngle(pie.Start)
	end := normalizeAngle(pie.End)
	if start <= end {
		return angle >= start && angle <= end
	}
	return angle >= start || angle <= end
}

func normalizeAngle(a float64) float64 {
	angle := math.Mod(a, 360)
	if angle < 0 {
		angle += 360
	}
	return angle
}
