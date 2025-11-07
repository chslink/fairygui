package widgets

import (
	"math"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestScrollBarSyncsWithScrollPane(t *testing.T) {
	owner := core.NewGComponent()
	owner.SetSize(100, 100)
	pane := owner.EnsureScrollPane(core.ScrollTypeVertical)
	pane.SetContentSize(100, 300)

	sb := NewScrollBar()
	sb.SetSize(20, 100)
	sb.SetTemplateComponent(makeScrollBarTemplate(true))
	sb.SetScrollPane(pane, true)

	pane.SetPos(0, 100, false)
	if diff := math.Abs(sb.scrollPerc - 0.5); diff > 0.01 {
		t.Fatalf("expected scrollPerc around 0.5, got %.2f", sb.scrollPerc)
	}
	length := sb.bar.Height() - sb.extraMargin()
	gripLen := sb.grip.Height()
	expected := length * (pane.ViewHeight() / 300)
	if diff := math.Abs(gripLen - expected); diff > 1 {
		t.Fatalf("expected grip length %.2f got %.2f", expected, gripLen)
	}
}

func makeScrollBarTemplate(vertical bool) *core.GComponent {
	tmpl := core.NewGComponent()
	tmpl.SetSize(20, 100)

	bar := core.NewGObject()
	bar.SetName("bar")
	if vertical {
		bar.SetSize(10, 90)
	} else {
		bar.SetSize(90, 10)
	}
	tmpl.AddChild(bar)

	grip := core.NewGObject()
	grip.SetName("grip")
	if vertical {
		grip.SetSize(10, 30)
	} else {
		grip.SetSize(30, 10)
	}
	tmpl.AddChild(grip)

	return tmpl
}

// TestScrollBarGripSizeCalculation 验证滑块尺寸计算逻辑（基于实际日志数据）
func TestScrollBarGripSizeCalculation(t *testing.T) {
	tests := []struct {
		name         string
		barLength    float64 // bar 的长度
		viewSize     float64 // 视口大小
		contentSize  float64 // 内容大小
		vertical     bool    // 是否垂直
		expectedGrip float64 // 期望的 grip 长度
	}{
		{
			name:         "Single Column 列表 - 垂直滚动条",
			barLength:    342.0,                          // 从日志: total=342.0
			viewSize:     422.0,                          // 从日志: viewSize.Y=422.0
			contentSize:  630.0,                          // 从日志: contentSize.Y=630.0
			vertical:     true,                           // 垂直滚动条
			expectedGrip: 342.0 * (422.0 / 630.0),        // 342 * 0.670 ≈ 229.14
		},
		{
			name:         "水平滚动列表",
			barLength:    497.0,                          // 从日志: total=497.0
			viewSize:     571.0,                          // 从日志: viewSize.X=571.0
			contentSize:  630.0,                          // 从日志: contentSize.X=630.0
			vertical:     false,                          // 水平滚动条
			expectedGrip: 497.0 * (571.0 / 630.0),        // 497 * 0.906 ≈ 450.5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 ScrollPane
			owner := core.NewGComponent()
			if tt.vertical {
				owner.SetSize(163, tt.viewSize)
			} else {
				owner.SetSize(tt.viewSize, 126)
			}
			pane := owner.EnsureScrollPane(core.ScrollTypeVertical)

			// 设置内容大小
			if tt.vertical {
				pane.SetContentSize(163, tt.contentSize)
			} else {
				pane.SetContentSize(tt.contentSize, 126)
			}

			// 创建 ScrollBar 和模板
			sb := NewScrollBar()
			tmpl := core.NewGComponent()
			if tt.vertical {
				tmpl.SetSize(20, tt.barLength)
			} else {
				tmpl.SetSize(tt.barLength, 20)
			}

			bar := core.NewGObject()
			bar.SetName("bar")
			if tt.vertical {
				bar.SetSize(10, tt.barLength)
			} else {
				bar.SetSize(tt.barLength, 10)
			}
			tmpl.AddChild(bar)

			grip := core.NewGObject()
			grip.SetName("grip")
			if tt.vertical {
				grip.SetSize(10, 50) // 初始尺寸
			} else {
				grip.SetSize(50, 10)
			}
			tmpl.AddChild(grip)

			sb.SetTemplateComponent(tmpl)
			sb.SetScrollPane(pane, tt.vertical)

			// 获取实际的 grip 尺寸
			var actualGripSize float64
			if tt.vertical {
				actualGripSize = sb.grip.Height()
			} else {
				actualGripSize = sb.grip.Width()
			}

			// 验证尺寸（允许 2.0 的误差）
			tolerance := 2.0
			if diff := math.Abs(actualGripSize - tt.expectedGrip); diff > tolerance {
				t.Errorf("grip size = %.2f, expected %.2f (diff = %.2f, displayPerc = %.3f)",
					actualGripSize, tt.expectedGrip, diff, sb.displayPerc)

				// 输出调试信息
				t.Logf("Debug info:")
				t.Logf("  barLength = %.2f", tt.barLength)
				t.Logf("  viewSize = %.2f", tt.viewSize)
				t.Logf("  contentSize = %.2f", tt.contentSize)
				t.Logf("  displayPerc = %.3f (expected %.3f)", sb.displayPerc, tt.viewSize/tt.contentSize)
				t.Logf("  fixedGrip = %v", sb.fixedGrip)
			}
		})
	}
}

// TestScrollBarGripMinSize 验证滑块最小尺寸约束
func TestScrollBarGripMinSize(t *testing.T) {
	owner := core.NewGComponent()
	owner.SetSize(100, 400)
	pane := owner.EnsureScrollPane(core.ScrollTypeVertical)

	// 设置极大的内容尺寸，导致 displayPerc 极小
	pane.SetContentSize(100, 5000) // displayPerc = 400/5000 = 0.08

	sb := NewScrollBar()
	tmpl := makeScrollBarTemplate(true)
	sb.SetTemplateComponent(tmpl)
	sb.SetScrollPane(pane, true)

	// grip 应该被限制为最小尺寸（minSize）
	gripHeight := sb.grip.Height()
	minExpected := sb.minSize()

	if gripHeight < minExpected {
		t.Errorf("grip height = %.2f, should be at least %.2f (minSize)",
			gripHeight, minExpected)
	}
}

// TestScrollBarWithoutScrollPane 验证未绑定 ScrollPane 时的行为
func TestScrollBarWithoutScrollPane(t *testing.T) {
	sb := NewScrollBar()
	tmpl := makeScrollBarTemplate(true)
	sb.SetTemplateComponent(tmpl)

	initialHeight := sb.grip.Height()

	// 尝试更新（应该被跳过，因为 target=nil）
	sb.updateGrip()

	// grip 尺寸不应改变
	if sb.grip.Height() != initialHeight {
		t.Errorf("grip height changed from %.2f to %.2f when target=nil",
			initialHeight, sb.grip.Height())
	}
}
