package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Loader 基础测试
// ============================================================================

func TestLoader_Creation(t *testing.T) {
	loader := fairygui.NewLoader()
	if loader == nil {
		t.Fatal("Expected non-nil loader")
	}
}

func TestLoader_URL(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置 URL
	loader.SetURL("ui://Main/icon")
	if loader.URL() != "ui://Main/icon" {
		t.Errorf("Expected URL 'ui://Main/icon', got '%s'", loader.URL())
	}
}

func TestLoader_Color(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置颜色
	loader.SetColor("#FF0000")
	if loader.Color() != "#FF0000" {
		t.Errorf("Expected color '#FF0000', got '%s'", loader.Color())
	}
}

func TestLoader_Playing(t *testing.T) {
	loader := fairygui.NewLoader()

	// 默认播放
	if !loader.Playing() {
		t.Error("Expected loader to be playing by default")
	}

	// 停止播放
	loader.SetPlaying(false)
	if loader.Playing() {
		t.Error("Expected loader to be stopped")
	}

	// 开始播放
	loader.SetPlaying(true)
	if !loader.Playing() {
		t.Error("Expected loader to be playing")
	}
}

func TestLoader_Frame(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置帧
	loader.SetFrame(5)
	if loader.Frame() != 5 {
		t.Errorf("Expected frame 5, got %d", loader.Frame())
	}
}

func TestLoader_AutoSize(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置自动调整大小
	loader.SetAutoSize(true)
	loader.SetAutoSize(false)

	// 注意：无法直接验证，只验证调用不会出错
}

func TestLoader_Fill(t *testing.T) {
	loader := fairygui.NewLoader()

	// 默认为 None
	if loader.Fill() != widgets.LoaderFillNone {
		t.Errorf("Expected default fill None, got %d", loader.Fill())
	}

	// 设置填充类型
	loader.SetFill(widgets.LoaderFillScale)
	if loader.Fill() != widgets.LoaderFillScale {
		t.Errorf("Expected fill Scale, got %d", loader.Fill())
	}
}

func TestLoader_Align(t *testing.T) {
	loader := fairygui.NewLoader()

	// 默认左对齐
	if loader.Align() != widgets.LoaderAlignLeft {
		t.Errorf("Expected default align Left, got '%s'", loader.Align())
	}

	// 设置居中对齐
	loader.SetAlign(widgets.LoaderAlignCenter)
	if loader.Align() != widgets.LoaderAlignCenter {
		t.Errorf("Expected align Center, got '%s'", loader.Align())
	}
}

func TestLoader_VerticalAlign(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置垂直对齐
	loader.SetVerticalAlign(widgets.LoaderAlignMiddle)

	// 注意：GLoader 可能没有 VerticalAlign getter
	// 只验证设置不会出错
}

func TestLoader_ShrinkOnly(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置仅缩小
	loader.SetShrinkOnly(true)

	// 注意：无法直接验证，只验证调用不会出错
}

func TestLoader_FillMethod(t *testing.T) {
	loader := fairygui.NewLoader()

	// 默认为 None
	if loader.FillMethod() != int(widgets.LoaderFillMethodNone) {
		t.Errorf("Expected default fill method None, got %d", loader.FillMethod())
	}

	// 设置填充方法
	loader.SetFillMethod(int(widgets.LoaderFillMethodHorizontal))
	if loader.FillMethod() != int(widgets.LoaderFillMethodHorizontal) {
		t.Errorf("Expected fill method Horizontal, got %d", loader.FillMethod())
	}
}

func TestLoader_FillOrigin(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置填充起点
	loader.SetFillOrigin(1)
	if loader.FillOrigin() != 1 {
		t.Errorf("Expected fill origin 1, got %d", loader.FillOrigin())
	}
}

func TestLoader_FillClockwise(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置顺时针填充
	loader.SetFillClockwise(true)
	if !loader.FillClockwise() {
		t.Error("Expected fill clockwise to be true")
	}

	loader.SetFillClockwise(false)
	if loader.FillClockwise() {
		t.Error("Expected fill clockwise to be false")
	}
}

func TestLoader_FillAmount(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置填充量
	loader.SetFillAmount(0.5)
	if loader.FillAmount() != 0.5 {
		t.Errorf("Expected fill amount 0.5, got %.2f", loader.FillAmount())
	}

	// 设置边界值
	loader.SetFillAmount(0)
	if loader.FillAmount() != 0 {
		t.Errorf("Expected fill amount 0, got %.2f", loader.FillAmount())
	}

	loader.SetFillAmount(1)
	if loader.FillAmount() != 1 {
		t.Errorf("Expected fill amount 1, got %.2f", loader.FillAmount())
	}
}

func TestLoader_Scale9Grid(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置九宫格（不会 panic 即可）
	loader.SetScale9Grid(nil)
}

func TestLoader_ScaleByTile(t *testing.T) {
	loader := fairygui.NewLoader()

	// 设置平铺缩放（不会 panic 即可）
	loader.SetScaleByTile(true)
	loader.SetScaleByTile(false)
}

func TestLoader_Position(t *testing.T) {
	loader := fairygui.NewLoader()
	loader.SetPosition(100, 200)

	x, y := loader.Position()
	if x != 100 || y != 200 {
		t.Errorf("Expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestLoader_Size(t *testing.T) {
	loader := fairygui.NewLoader()
	loader.SetSize(120, 80)

	w, h := loader.Size()
	if w != 120 || h != 80 {
		t.Errorf("Expected size (120, 80), got (%.0f, %.0f)", w, h)
	}
}

func TestLoader_Visible(t *testing.T) {
	loader := fairygui.NewLoader()

	// 默认可见
	if !loader.Visible() {
		t.Error("Expected loader to be visible by default")
	}

	// 隐藏
	loader.SetVisible(false)
	if loader.Visible() {
		t.Error("Expected loader to be hidden")
	}

	// 显示
	loader.SetVisible(true)
	if !loader.Visible() {
		t.Error("Expected loader to be visible")
	}
}

func TestLoader_Name(t *testing.T) {
	loader := fairygui.NewLoader()
	loader.SetName("MyLoader")

	if loader.Name() != "MyLoader" {
		t.Errorf("Expected name 'MyLoader', got '%s'", loader.Name())
	}
}

func TestLoader_Alpha(t *testing.T) {
	loader := fairygui.NewLoader()

	// 默认透明度为 1
	if loader.Alpha() != 1.0 {
		t.Errorf("Expected default alpha 1.0, got %.2f", loader.Alpha())
	}

	// 设置半透明
	loader.SetAlpha(0.5)
	if loader.Alpha() != 0.5 {
		t.Errorf("Expected alpha 0.5, got %.2f", loader.Alpha())
	}
}

// ============================================================================
// RawLoader 访问测试
// ============================================================================

func TestLoader_RawLoader(t *testing.T) {
	loader := fairygui.NewLoader()
	raw := loader.RawLoader()

	if raw == nil {
		t.Error("Expected non-nil raw loader")
	}
}
