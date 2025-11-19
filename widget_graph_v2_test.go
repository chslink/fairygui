package fairygui

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestNewGraph 测试创建新的图形控件
func TestNewGraph(t *testing.T) {
	g := NewGraph()
	if g == nil {
		t.Fatal("NewGraph() returned nil")
	}

	// 检查默认属性
	if g.Object == nil {
		t.Error("Graph.Object is nil")
	}

	if g.GraphType() != GraphTypeRectangle {
		t.Errorf("默认图形类型不正确: got %v, want %v", g.GraphType(), GraphTypeRectangle)
	}

	// 检查默认颜色
	expLineColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	if g.LineColor() != expLineColor {
		t.Errorf("默认线条颜色不正确: got %+v, want %+v", g.LineColor(), expLineColor)
	}

	expFillColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if g.FillColor() != expFillColor {
		t.Errorf("默认填充颜色不正确: got %+v, want %+v", g.FillColor(), expFillColor)
	}

	if g.LineSize() != 1 {
		t.Errorf("默认线宽不正确: got %.1f, want 1", g.LineSize())
	}

	if g.Touchable() {
		t.Error("Graph 应该默认不拦截事件")
	}
}

// TestGraph_SetGraphType 测试设置图形类型
func TestGraph_SetGraphType(t *testing.T) {
	g := NewGraph()

	types := []GraphType{
		GraphTypeRectangle,
		GraphTypeRectangleCorner,
		GraphTypeCircle,
		GraphTypeEllipse,
		GraphTypePolygon,
		GraphTypeRegularPolygon,
	}

	for _, gt := range types {
		g.SetGraphType(gt)
		if g.GraphType() != gt {
			t.Errorf("图形类型不正确: got %v, want %v", g.GraphType(), gt)
		}
	}
}

// TestGraph_SetLineStyle 测试线条样式
func TestGraph_SetLineStyle(t *testing.T) {
	g := NewGraph()

	lineSize := 3.0
	lineColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	g.SetLineStyle(lineSize, lineColor)

	if g.LineSize() != lineSize {
		t.Errorf("线宽不正确: got %.1f, want %.1f", g.LineSize(), lineSize)
	}

	if g.LineColor() != lineColor {
		t.Errorf("线条颜色不正确: got %+v, want %+v", g.LineColor(), lineColor)
	}
}

// TestGraph_SetLineSize 测试设置线宽
func TestGraph_SetLineSize(t *testing.T) {
	g := NewGraph()

	sizes := []float64{0, 1, 2, 5, 10}

	for _, size := range sizes {
		g.SetLineSize(size)
		if g.LineSize() != size {
			t.Errorf("线宽不正确: got %.1f, want %.1f", g.LineSize(), size)
		}
	}

	// 测试负值（无边框）
	g.SetLineSize(-1)
	if g.LineSize() != -1 {
		t.Errorf("线宽不正确: got %.1f, want -1", g.LineSize())
	}
}

// TestGraph_SetLineColor 测试设置线条颜色
func TestGraph_SetLineColor(t *testing.T) {
	g := NewGraph()

	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
		{R: 255, G: 255, B: 255, A: 128},
		{R: 0, G: 0, B: 0, A: 0},
	}

	for _, c := range colors {
		g.SetLineColor(c)
		if g.LineColor() != c {
			t.Errorf("线条颜色不正确: got %+v, want %+v", g.LineColor(), c)
		}
	}
}

// TestGraph_SetStroke 测试描边
func TestGraph_SetStroke(t *testing.T) {
	g := NewGraph()

	strokeSize := 2.0
	strokeColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}

	g.SetStroke(strokeSize, strokeColor)

	if g.StrokeSize() != strokeSize {
		t.Errorf("描边大小不正确: got %.1f, want %.1f", g.StrokeSize(), strokeSize)
	}

	if g.StrokeColor() != strokeColor {
		t.Errorf("描边颜色不正确: got %+v, want %+v", g.StrokeColor(), strokeColor)
	}
}

// TestGraph_SetFillColor 测试填充颜色
func TestGraph_SetFillColor(t *testing.T) {
	g := NewGraph()

	fillColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	g.SetFillColor(fillColor, true)

	if g.FillColor() != fillColor {
		t.Errorf("填充颜色不正确: got %+v, want %+v", g.FillColor(), fillColor)
	}

	if !g.FillShape() {
		t.Error("FillShape 应该为 true")
	}

	// 测试不填充
	g.SetFillColor(fillColor, false)
	if g.FillShape() {
		t.Error("FillShape 应该为 false")
	}
}

// TestGraph_SetCornerRadius 测试圆角半径
func TestGraph_SetCornerRadius(t *testing.T) {
	g := NewGraph()

	radius := 10.0
	g.SetCornerRadius(radius)

	if g.CornerRadius() != radius {
		t.Errorf("圆角半径不正确: got %.1f, want %.1f", g.CornerRadius(), radius)
	}

	// 测试设置为圆角矩形类型
	g.SetGraphType(GraphTypeRectangleCorner)
	if g.GraphType() != GraphTypeRectangleCorner {
		t.Error("图形类型应该为 GraphTypeRectangleCorner")
	}
}

// TestGraph_SetRadius 测试圆形半径
func TestGraph_SetRadius(t *testing.T) {
	g := NewGraph()

	radius := 50.0
	g.SetRadius(radius)

	if g.RadiusX() != radius {
		t.Errorf("X半径不正确: got %.1f, want %.1f", g.RadiusX(), radius)
	}

	if g.RadiusY() != radius {
		t.Errorf("Y半径不正确: got %.1f, want %.1f", g.RadiusY(), radius)
	}

	// 测试设置为圆形类型
	g.SetGraphType(GraphTypeCircle)
	if g.GraphType() != GraphTypeCircle {
		t.Error("图形类型应该为 GraphTypeCircle")
	}
}

// TestGraph_SetRadiusXY 测试椭圆半径
func TestGraph_SetRadiusXY(t *testing.T) {
	g := NewGraph()

	radiusX := 60.0
	radiusY := 40.0
	g.SetRadiusXY(radiusX, radiusY)

	if g.RadiusX() != radiusX {
		t.Errorf("X半径不正确: got %.1f, want %.1f", g.RadiusX(), radiusX)
	}

	if g.RadiusY() != radiusY {
		t.Errorf("Y半径不正确: got %.1f, want %.1f", g.RadiusY(), radiusY)
	}

	// 测试设置为椭圆类型
	g.SetGraphType(GraphTypeEllipse)
	if g.GraphType() != GraphTypeEllipse {
		t.Error("图形类型应该为 GraphTypeEllipse")
	}
}

// TestGraph_SetSides 测试多边形边数
func TestGraph_SetSides(t *testing.T) {
	g := NewGraph()

	sides := 6
	g.SetSides(sides)

	if g.Sides() != sides {
		t.Errorf("边数不正确: got %d, want %d", g.Sides(), sides)
	}

	// 测试边数限制（最小3边）
	g.SetSides(2)
	if g.Sides() != 3 {
		t.Errorf("边数应该为最小值3: got %d", g.Sides())
	}
}

// TestGraph_SetRotation 测试旋转
func TestGraph_SetRotation(t *testing.T) {
	g := NewGraph()

	rotations := []float64{0, 45, 90, 180, 270, 360}

	for _, rotation := range rotations {
		g.SetRotation(rotation)
		if g.Rotation() != rotation {
			t.Errorf("旋转角度不正确: got %.1f, want %.1f", g.Rotation(), rotation)
		}
	}
}

// TestGraph_DrawRect 测试绘制矩形
func TestGraph_DrawRect(t *testing.T) {
	g := NewGraph()

	lineSize := 2.0
	lineColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	fillColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	g.DrawRect(lineSize, lineColor, fillColor)

	if g.GraphType() != GraphTypeRectangle {
		t.Error("图形类型应该为 GraphTypeRectangle")
	}

	if g.LineSize() != lineSize {
		t.Errorf("线宽不正确: got %.1f, want %.1f", g.LineSize(), lineSize)
	}

	if g.LineColor() != lineColor {
		t.Errorf("线条颜色不正确: got %+v, want %+v", g.LineColor(), lineColor)
	}

	if g.FillColor() != fillColor {
		t.Errorf("填充颜色不正确: got %+v, want %+v", g.FillColor(), fillColor)
	}

	if !g.FillShape() {
		t.Error("应该填充")
	}
}

// TestGraph_DrawRoundedRect 测试绘制圆角矩形
func TestGraph_DrawRoundedRect(t *testing.T) {
	g := NewGraph()

	lineSize := 2.0
	lineColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	fillColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	cornerRadius := 10.0

	g.DrawRoundedRect(lineSize, lineColor, fillColor, cornerRadius)

	if g.GraphType() != GraphTypeRectangleCorner {
		t.Error("图形类型应该为 GraphTypeRectangleCorner")
	}

	if g.CornerRadius() != cornerRadius {
		t.Errorf("圆角半径不正确: got %.1f, want %.1f", g.CornerRadius(), cornerRadius)
	}
}

// TestGraph_DrawCircle 测试绘制圆形
func TestGraph_DrawCircle(t *testing.T) {
	g := NewGraph()

	lineSize := 2.0
	lineColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	fillColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	g.DrawCircle(lineSize, lineColor, fillColor)

	if g.GraphType() != GraphTypeCircle {
		t.Error("图形类型应该为 GraphTypeCircle")
	}
}

// TestGraph_DrawEllipse 测试绘制椭圆
func TestGraph_DrawEllipse(t *testing.T) {
	g := NewGraph()

	lineSize := 2.0
	lineColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	fillColor := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	g.DrawEllipse(lineSize, lineColor, fillColor)

	if g.GraphType() != GraphTypeEllipse {
		t.Error("图形类型应该为 GraphTypeEllipse")
	}
}

// TestGraph_DrawPolygon 测试绘制多边形
func TestGraph_DrawPolygon(t *testing.T) {
	g := NewGraph()

	lineSize := 2.0
	lineColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	fillColor := color.RGBA{R: 255, G: 255, B: 0, A: 255}
	sides := 6

	g.DrawPolygon(lineSize, lineColor, fillColor, sides)

	if g.GraphType() != GraphTypePolygon {
		t.Error("图形类型应该为 GraphTypePolygon")
	}

	if g.Sides() != sides {
		t.Errorf("边数不正确: got %d, want %d", g.Sides(), sides)
	}
}

// TestGraph_DrawRegularPolygon 测试绘制正多边形
func TestGraph_DrawRegularPolygon(t *testing.T) {
	g := NewGraph()

	lineSize := 2.0
	lineColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	fillColor := color.RGBA{R: 255, G: 0, B: 255, A: 255}
	sides := 5
	cut := true

	g.DrawRegularPolygon(lineSize, lineColor, fillColor, sides, cut)

	if g.GraphType() != GraphTypeRegularPolygon {
		t.Error("图形类型应该为 GraphTypeRegularPolygon")
	}

	if g.Sides() != sides {
		t.Errorf("边数不正确: got %d, want %d", g.Sides(), sides)
	}

	if g.PolygonCut() != cut {
		t.Errorf("PolygonCut 不正确: got %v, want %v", g.PolygonCut(), cut)
	}
}

// TestGraph_Clear 测试清除图形
func TestGraph_Clear(t *testing.T) {
	g := NewGraph()

	// 先绘制一个图形
	g.DrawRect(2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// 验证填充
	if !g.FillShape() {
		t.Error("绘制后应该填充")
	}

	// 清除
	g.Clear()

	if g.FillShape() {
		t.Error("Clear 后应该不填充")
	}

	if g.LineSize() != 0 {
		t.Errorf("Clear 后线宽应该为 0: got %.1f", g.LineSize())
	}
}

// TestGraph_ObjectInterface 测试 Object 接口
func TestGraph_ObjectInterface(t *testing.T) {
	g := NewGraph()

	// 测试尺寸设置
	g.SetSize(100, 100)
	width, height := g.Size()
	if width != 100 || height != 100 {
		t.Errorf("尺寸不正确: got (%.1f, %.1f), want (100, 100)", width, height)
	}

	// 测试位置设置
	g.SetPosition(50, 50)
	x, y := g.Position()
	if x != 50 || y != 50 {
		t.Errorf("位置不正确: got (%.1f, %.1f), want (50, 50)", x, y)
	}

	// 测试可见性
	g.SetVisible(false)
	if g.Visible() {
		t.Error("设置 Visible(false) 失败")
	}

	// 测试透明度
	g.SetAlpha(0.5)
	if g.Alpha() != 0.5 {
		t.Errorf("透明度设置失败: got %.1f, want 0.5", g.Alpha())
	}
}

// TestGraph_Draw 测试绘制
func TestGraph_Draw(t *testing.T) {
	g := NewGraph()

	// 创建一个离屏图像用于测试绘制
	screen := ebiten.NewImage(800, 600)

	// 设置图形属性
	g.SetSize(100, 100)
	g.SetPosition(50, 50)
	g.SetVisible(true)

	// 绘制矩形
	g.DrawRect(2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, color.RGBA{R: 255, G: 0, B: 0, A: 128})

	// 绘制（不应该 panic）
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Draw 方法 panic: %v", r)
		}
	}()

	g.Draw(screen)
}

// TestGraph_Chaining 测试链式调用
func TestGraph_Chaining(t *testing.T) {
	g := NewGraph()

	// 测试方法链式调用（先设置类型）
	g.SetGraphType(GraphTypeCircle)
	g.SetLineSize(3)
	g.SetLineColor(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	g.SetFillColor(color.RGBA{R: 0, G: 255, B: 0, A: 128}, true)
	g.SetRadius(50)

	if g.GraphType() != GraphTypeCircle {
		t.Error("图形类型设置失败")
	}

	if g.LineSize() != 3 {
		t.Error("线宽设置失败")
	}

	if !g.FillShape() {
		t.Error("填充设置失败")
	}
}

// TestAssertGraph 测试类型断言
func TestAssertGraph(t *testing.T) {
	g := NewGraph()

	// 测试 AssertGraph
	result, ok := AssertGraph(g)
	if !ok {
		t.Error("AssertGraph 应该成功")
	}
	if result != g {
		t.Error("AssertGraph 返回的对象不正确")
	}

	// 测试 IsGraph
	if !IsGraph(g) {
		t.Error("IsGraph 应该返回 true")
	}

	// 测试不是 Graph 的情况
	obj := NewObject()
	_, ok = AssertGraph(obj)
	if ok {
		t.Error("AssertGraph 对非 Graph 对象应该失败")
	}

	if IsGraph(obj) {
		t.Error("IsGraph 对非 Graph 对象应该返回 false")
	}
}

// TestGraph_AllTypes 测试所有图形类型
func TestGraph_AllTypes(t *testing.T) {
	screen := ebiten.NewImage(800, 600)

	tests := []struct {
		name      string
		setupFunc func(*Graph)
	}{
		{
			name: "Rectangle",
			setupFunc: func(g *Graph) {
				g.DrawRect(2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			},
		},
		{
			name: "RoundedRect",
			setupFunc: func(g *Graph) {
				g.DrawRoundedRect(2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, color.RGBA{R: 0, G: 255, B: 0, A: 255}, 10)
			},
		},
		{
			name: "Circle",
			setupFunc: func(g *Graph) {
				g.DrawCircle(2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, color.RGBA{R: 0, G: 0, B: 255, A: 255})
			},
		},
		{
			name: "Ellipse",
			setupFunc: func(g *Graph) {
				g.DrawEllipse(2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, color.RGBA{R: 255, G: 255, B: 0, A: 255})
			},
		},
		{
			name: "Polygon",
			setupFunc: func(g *Graph) {
				g.DrawPolygon(2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, color.RGBA{R: 255, G: 0, B: 255, A: 255}, 6)
			},
		},
		{
			name: "RegularPolygon",
			setupFunc: func(g *Graph) {
				g.DrawRegularPolygon(2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, color.RGBA{R: 128, G: 128, B: 128, A: 255}, 5, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGraph()
			g.SetSize(100, 100)
			g.SetPosition(50, 50)

			tt.setupFunc(g)

			// 绘制（不应该 panic）
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Draw 方法 panic: %v", r)
				}
			}()

			g.Draw(screen)
		})
	}
}
