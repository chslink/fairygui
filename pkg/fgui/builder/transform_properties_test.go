package builder

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// TestGraphSceneTransformProperties 验证 Graph 场景中对象的变换属性
// 测试 n11 和 trapezoid 的旋转、轴心、倾斜属性是否与配置对应
//
// 参考 Demo_Graph.xml:
// - n11: pivot="0.5,0.5" skew="0,30" rotation="-30"
// - trapezoid: pivot="0,0.5" skew="34,0"
func TestGraphSceneTransformProperties(t *testing.T) {
	// 加载 Basics.fui 包
	fuiPath := filepath.Join("..", "..", "..", "demo", "assets", "Basics.fui")
	fuiData, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 .fui 文件: %v", err)
	}

	pkg, err := assets.ParsePackage(fuiData, "demo/assets/Basics")
	if err != nil {
		t.Fatalf("解析 .fui 文件失败: %v", err)
	}

	// 查找 Demo_Graph 组件
	var graphItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_Graph" {
			graphItem = item
			break
		}
	}

	if graphItem == nil {
		t.Fatalf("未找到 Demo_Graph 组件")
	}

	// 构建 Demo_Graph 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	graph, err := factory.BuildComponent(ctx, pkg, graphItem)
	if err != nil {
		t.Fatalf("构建 Demo_Graph 组件失败: %v", err)
	}

	// 测试 n11 的变换属性
	t.Run("n11_TransformProperties", func(t *testing.T) {
		n11 := graph.ChildByName("n11")
		if n11 == nil {
			t.Fatalf("未找到 n11 对象")
		}

		// 验证 Pivot (0.5, 0.5)
		pivotX, pivotY := n11.Pivot()
		expectedPivotX, expectedPivotY := 0.5, 0.5
		if !floatEquals(pivotX, expectedPivotX, 0.001) || !floatEquals(pivotY, expectedPivotY, 0.001) {
			t.Errorf("n11 Pivot 不正确: 期望(%.3f,%.3f), 实际(%.3f,%.3f)",
				expectedPivotX, expectedPivotY, pivotX, pivotY)
		} else {
			t.Logf("✓ n11 Pivot 正确: (%.3f,%.3f)", pivotX, pivotY)
		}

		// 验证 Rotation (-30 degrees)
		// 注意：FairyGUI 使用角度，不是弧度（与 TypeScript 版本一致）
		rotation := n11.Rotation()
		expectedRotation := -30.0
		if !floatEquals(rotation, expectedRotation, 0.001) {
			t.Errorf("n11 Rotation 不正确: 期望%.1f°, 实际%.1f°",
				expectedRotation, rotation)
		} else {
			t.Logf("✓ n11 Rotation 正确: %.1f°", rotation)
		}

		// 验证 Skew (0, 30)
		skewX, skewY := n11.Skew()
		expectedSkewX, expectedSkewY := 0.0, 30.0
		if !floatEquals(skewX, expectedSkewX, 0.001) || !floatEquals(skewY, expectedSkewY, 0.001) {
			t.Errorf("n11 Skew 不正确: 期望(%.1f,%.1f), 实际(%.1f,%.1f)",
				expectedSkewX, expectedSkewY, skewX, skewY)
		} else {
			t.Logf("✓ n11 Skew 正确: (%.1f,%.1f)", skewX, skewY)
		}

		// 验证 PivotAsAnchor (pivot 属性不带 anchor="true"，应该为 false)
		if n11.PivotAsAnchor() {
			t.Errorf("n11 PivotAsAnchor 应该是 false，实际: true")
		} else {
			t.Logf("✓ n11 PivotAsAnchor 正确: false")
		}
	})

	// 测试 trapezoid 的变换属性
	t.Run("trapezoid_TransformProperties", func(t *testing.T) {
		trapezoid := graph.ChildByName("trapezoid")
		if trapezoid == nil {
			t.Fatalf("未找到 trapezoid 对象")
		}

		// 验证 Pivot (0, 0.5)
		pivotX, pivotY := trapezoid.Pivot()
		expectedPivotX, expectedPivotY := 0.0, 0.5
		if !floatEquals(pivotX, expectedPivotX, 0.001) || !floatEquals(pivotY, expectedPivotY, 0.001) {
			t.Errorf("trapezoid Pivot 不正确: 期望(%.3f,%.3f), 实际(%.3f,%.3f)",
				expectedPivotX, expectedPivotY, pivotX, pivotY)
		} else {
			t.Logf("✓ trapezoid Pivot 正确: (%.3f,%.3f)", pivotX, pivotY)
		}

		// 验证 Rotation (XML 中没有 rotation 属性，应该为 0)
		// 注意：FairyGUI 使用角度，不是弧度（与 TypeScript 版本一致）
		rotation := trapezoid.Rotation()
		expectedRotation := 0.0
		if !floatEquals(rotation, expectedRotation, 0.001) {
			t.Errorf("trapezoid Rotation 不正确: 期望%.1f°, 实际%.1f°",
				expectedRotation, rotation)
		} else {
			t.Logf("✓ trapezoid Rotation 正确: %.1f°", rotation)
		}

		// 验证 Skew (34, 0)
		skewX, skewY := trapezoid.Skew()
		expectedSkewX, expectedSkewY := 34.0, 0.0
		if !floatEquals(skewX, expectedSkewX, 0.001) || !floatEquals(skewY, expectedSkewY, 0.001) {
			t.Errorf("trapezoid Skew 不正确: 期望(%.1f,%.1f), 实际(%.1f,%.1f)",
				expectedSkewX, expectedSkewY, skewX, skewY)
		} else {
			t.Logf("✓ trapezoid Skew 正确: (%.1f,%.1f)", skewX, skewY)
		}

		// 验证 PivotAsAnchor (pivot 属性不带 anchor="true"，应该为 false)
		if trapezoid.PivotAsAnchor() {
			t.Errorf("trapezoid PivotAsAnchor 应该是 false，实际: true")
		} else {
			t.Logf("✓ trapezoid PivotAsAnchor 正确: false")
		}
	})

	// 测试对比：验证有 anchor="true" 的对象
	t.Run("radial_AnchorProperty", func(t *testing.T) {
		radial := graph.ChildByName("radial")
		if radial == nil {
			t.Skipf("未找到 radial 对象（可选测试）")
			return
		}

		// radial 的 XML: pivot="0.5,0.5" anchor="true"
		// 验证 PivotAsAnchor 应该为 true
		if !radial.PivotAsAnchor() {
			t.Errorf("radial PivotAsAnchor 应该是 true (XML 中有 anchor=\"true\")，实际: false")
		} else {
			t.Logf("✓ radial PivotAsAnchor 正确: true (有 anchor 属性)")
		}

		// 验证 Pivot (0.5, 0.5)
		pivotX, pivotY := radial.Pivot()
		expectedPivotX, expectedPivotY := 0.5, 0.5
		if !floatEquals(pivotX, expectedPivotX, 0.001) || !floatEquals(pivotY, expectedPivotY, 0.001) {
			t.Errorf("radial Pivot 不正确: 期望(%.3f,%.3f), 实际(%.3f,%.3f)",
				expectedPivotX, expectedPivotY, pivotX, pivotY)
		} else {
			t.Logf("✓ radial Pivot 正确: (%.3f,%.3f)", pivotX, pivotY)
		}
	})
}

// floatEquals 比较两个浮点数是否相等（在误差范围内）
func floatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}
