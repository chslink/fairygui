package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// Graph 基础测试
// ============================================================================

func TestGraph_Creation(t *testing.T) {
	graph := fairygui.NewGraph()
	if graph == nil {
		t.Fatal("Expected non-nil graph")
	}
}

func TestGraph_Type(t *testing.T) {
	graph := fairygui.NewGraph()

	// 默认为 Empty
	if graph.Type() != fairygui.GraphTypeEmpty {
		t.Errorf("Expected default type to be Empty, got %d", graph.Type())
	}

	// 设置为矩形
	graph.SetType(fairygui.GraphTypeRect)
	if graph.Type() != fairygui.GraphTypeRect {
		t.Errorf("Expected type to be Rect, got %d", graph.Type())
	}

	// 设置为椭圆
	graph.SetType(fairygui.GraphTypeEllipse)
	if graph.Type() != fairygui.GraphTypeEllipse {
		t.Errorf("Expected type to be Ellipse, got %d", graph.Type())
	}
}

func TestGraph_LineSize(t *testing.T) {
	graph := fairygui.NewGraph()

	// 设置线条宽度
	graph.SetLineSize(2.5)
	if graph.LineSize() != 2.5 {
		t.Errorf("Expected LineSize to be 2.5, got %.2f", graph.LineSize())
	}
}

func TestGraph_LineColor(t *testing.T) {
	graph := fairygui.NewGraph()

	// 设置线条颜色
	graph.SetLineColor("#FF0000")
	if graph.LineColor() != "#FF0000" {
		t.Errorf("Expected LineColor to be '#FF0000', got '%s'", graph.LineColor())
	}
}

func TestGraph_FillColor(t *testing.T) {
	graph := fairygui.NewGraph()

	// 设置填充颜色
	graph.SetFillColor("#00FF00")
	if graph.FillColor() != "#00FF00" {
		t.Errorf("Expected FillColor to be '#00FF00', got '%s'", graph.FillColor())
	}
}

func TestGraph_Color(t *testing.T) {
	graph := fairygui.NewGraph()

	// Color 是 FillColor 的别名
	graph.SetColor("#0000FF")
	if graph.Color() != "#0000FF" {
		t.Errorf("Expected Color to be '#0000FF', got '%s'", graph.Color())
	}
	if graph.FillColor() != "#0000FF" {
		t.Errorf("Expected FillColor to be '#0000FF', got '%s'", graph.FillColor())
	}
}

func TestGraph_CornerRadius(t *testing.T) {
	graph := fairygui.NewGraph()

	// 设置圆角
	radii := []float64{10, 10, 10, 10}
	graph.SetCornerRadius(radii)

	gotRadii := graph.CornerRadius()
	if len(gotRadii) != 4 {
		t.Fatalf("Expected 4 corner radii, got %d", len(gotRadii))
	}

	for i, r := range radii {
		if gotRadii[i] != r {
			t.Errorf("Expected radius[%d] = %.0f, got %.0f", i, r, gotRadii[i])
		}
	}
}

func TestGraph_PolygonPoints(t *testing.T) {
	graph := fairygui.NewGraph()

	// 设置多边形顶点
	points := []float64{0, 0, 100, 0, 50, 100}
	graph.SetPolygonPoints(points)

	gotPoints := graph.PolygonPoints()
	if len(gotPoints) != 6 {
		t.Fatalf("Expected 6 points, got %d", len(gotPoints))
	}

	for i, p := range points {
		if gotPoints[i] != p {
			t.Errorf("Expected point[%d] = %.0f, got %.0f", i, p, gotPoints[i])
		}
	}
}

func TestGraph_RegularPolygon(t *testing.T) {
	graph := fairygui.NewGraph()

	// 设置正多边形
	graph.SetRegularPolygon(6, 0, nil)

	sides, angle, distances := graph.RegularPolygon()
	if sides != 6 {
		t.Errorf("Expected 6 sides, got %d", sides)
	}
	if angle != 0 {
		t.Errorf("Expected start angle 0, got %.0f", angle)
	}
	// 底层实现返回空切片而不是 nil
	if len(distances) != 0 {
		t.Errorf("Expected empty distances, got %v", distances)
	}
}

func TestGraph_DrawRect(t *testing.T) {
	graph := fairygui.NewGraph()

	// 绘制矩形
	graph.DrawRect(1, "#000000", "#FF0000", nil)

	// 验证类型和属性
	if graph.Type() != fairygui.GraphTypeRect {
		t.Errorf("Expected type to be Rect, got %d", graph.Type())
	}
	if graph.LineSize() != 1 {
		t.Errorf("Expected LineSize to be 1, got %.0f", graph.LineSize())
	}
	if graph.LineColor() != "#000000" {
		t.Errorf("Expected LineColor to be '#000000', got '%s'", graph.LineColor())
	}
	if graph.FillColor() != "#FF0000" {
		t.Errorf("Expected FillColor to be '#FF0000', got '%s'", graph.FillColor())
	}
}

func TestGraph_DrawEllipse(t *testing.T) {
	graph := fairygui.NewGraph()

	// 绘制椭圆
	graph.DrawEllipse(2, "#0000FF", "#00FF00")

	// 验证类型和属性
	if graph.Type() != fairygui.GraphTypeEllipse {
		t.Errorf("Expected type to be Ellipse, got %d", graph.Type())
	}
	if graph.LineSize() != 2 {
		t.Errorf("Expected LineSize to be 2, got %.0f", graph.LineSize())
	}
	if graph.LineColor() != "#0000FF" {
		t.Errorf("Expected LineColor to be '#0000FF', got '%s'", graph.LineColor())
	}
	if graph.FillColor() != "#00FF00" {
		t.Errorf("Expected FillColor to be '#00FF00', got '%s'", graph.FillColor())
	}
}

func TestGraph_DrawPolygon(t *testing.T) {
	graph := fairygui.NewGraph()

	// 绘制三角形
	points := []float64{0, 0, 100, 0, 50, 100}
	graph.DrawPolygon(1, "#000000", "#FFFF00", points)

	// 验证类型和属性
	if graph.Type() != fairygui.GraphTypePolygon {
		t.Errorf("Expected type to be Polygon, got %d", graph.Type())
	}

	gotPoints := graph.PolygonPoints()
	if len(gotPoints) != 6 {
		t.Fatalf("Expected 6 points, got %d", len(gotPoints))
	}
}

func TestGraph_DrawRegularPolygon(t *testing.T) {
	graph := fairygui.NewGraph()

	// 绘制六边形
	graph.DrawRegularPolygon(1, "#000000", "#FF00FF", 6, 0, nil)

	// 验证类型和属性
	if graph.Type() != fairygui.GraphTypeRegularPolygon {
		t.Errorf("Expected type to be RegularPolygon, got %d", graph.Type())
	}

	sides, _, _ := graph.RegularPolygon()
	if sides != 6 {
		t.Errorf("Expected 6 sides, got %d", sides)
	}
}

func TestGraph_Position(t *testing.T) {
	graph := fairygui.NewGraph()
	graph.SetPosition(100, 200)

	x, y := graph.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestGraph_Size(t *testing.T) {
	graph := fairygui.NewGraph()
	graph.SetSize(150, 120)

	w, h := graph.Size()
	if w != 150 || h != 120 {
		t.Errorf("Expected size (150, 120), got (%.0f, %.0f)", w, h)
	}
}

func TestGraph_Visible(t *testing.T) {
	graph := fairygui.NewGraph()

	// 默认可见
	if !graph.Visible() {
		t.Error("Expected graph to be visible by default")
	}

	// 隐藏
	graph.SetVisible(false)
	if graph.Visible() {
		t.Error("Expected graph to be hidden")
	}

	// 显示
	graph.SetVisible(true)
	if !graph.Visible() {
		t.Error("Expected graph to be visible")
	}
}

func TestGraph_Name(t *testing.T) {
	graph := fairygui.NewGraph()
	graph.SetName("MyGraph")

	if graph.Name() != "MyGraph" {
		t.Errorf("Expected name 'MyGraph', got '%s'", graph.Name())
	}
}

func TestGraph_Alpha(t *testing.T) {
	graph := fairygui.NewGraph()

	// 默认透明度为 1
	if graph.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", graph.Alpha())
	}

	// 设置半透明
	graph.SetAlpha(0.5)
	if graph.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", graph.Alpha())
	}
}

// ============================================================================
// RawGraph 访问测试
// ============================================================================

func TestGraph_RawGraph(t *testing.T) {
	graph := fairygui.NewGraph()
	raw := graph.RawGraph()

	if raw == nil {
		t.Error("Expected non-nil raw graph")
	}
}
