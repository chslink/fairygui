package laya

import (
	"math"
	"testing"
)

// TestPivotRotationTransform 测试轴心点旋转变换的正确性
func TestPivotRotationTransform(t *testing.T) {
	tests := []struct {
		name          string
		width         float64
		height        float64
		pivotX        float64
		pivotY        float64
		posX          float64
		posY          float64
		rotation      float64
		pivotAsAnchor bool
		wantCenterX   float64 // 旋转后轴心点的全局坐标
		wantCenterY   float64
	}{
		{
			name:          "Center pivot, 90 degree rotation",
			width:         100,
			height:        50,
			pivotX:        0.5,
			pivotY:        0.5,
			posX:          200,
			posY:          300,
			rotation:      math.Pi / 2, // 90 degrees
			pivotAsAnchor: false,
			wantCenterX:   250, // pivotAsAnchor=false: posX是左上角位置，pivot在(200+50, 300+25)
			wantCenterY:   325,
		},
		{
			name:          "Top-left pivot, 45 degree rotation",
			width:         100,
			height:        100,
			pivotX:        0,
			pivotY:        0,
			posX:          100,
			posY:          100,
			rotation:      math.Pi / 4, // 45 degrees
			pivotAsAnchor: false,
			wantCenterX:   100, // 轴心点在左上角，应该保持在 (100, 100)
			wantCenterY:   100,
		},
		{
			name:          "Bottom-right pivot, 180 degree rotation",
			width:         100,
			height:        50,
			pivotX:        1.0,
			pivotY:        1.0,
			posX:          200,
			posY:          300,
			rotation:      math.Pi, // 180 degrees
			pivotAsAnchor: false,
			wantCenterX:   300, // 轴心点在右下角，应该保持在 (200, 300)
			wantCenterY:   350,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprite := NewSprite()
			sprite.SetSize(tt.width, tt.height)
			sprite.SetPivotWithAnchor(tt.pivotX, tt.pivotY, tt.pivotAsAnchor)
			sprite.SetPosition(tt.posX, tt.posY)
			sprite.SetRotation(tt.rotation)

			// 计算轴心点在局部坐标系中的位置
			localPivotX := tt.pivotX * tt.width
			localPivotY := tt.pivotY * tt.height

			// 将局部轴心点转换到全局坐标
			globalPivot := sprite.LocalToGlobal(Point{X: localPivotX, Y: localPivotY})

			// 检查轴心点是否在预期位置
			// 轴心点旋转后应该保持在原位置
			tolerance := 1e-6
			if math.Abs(globalPivot.X-tt.wantCenterX) > tolerance {
				t.Errorf("pivot global X = %v, want %v (diff: %v)",
					globalPivot.X, tt.wantCenterX, globalPivot.X-tt.wantCenterX)
			}
			if math.Abs(globalPivot.Y-tt.wantCenterY) > tolerance {
				t.Errorf("pivot global Y = %v, want %v (diff: %v)",
					globalPivot.Y, tt.wantCenterY, globalPivot.Y-tt.wantCenterY)
			}

			// 打印调试信息
			matrix := sprite.LocalMatrix()
			t.Logf("Matrix: [[%.6f %.6f %.6f] [%.6f %.6f %.6f]]",
				matrix.A, matrix.C, matrix.Tx,
				matrix.B, matrix.D, matrix.Ty)
			t.Logf("Global pivot: (%.6f, %.6f)", globalPivot.X, globalPivot.Y)
			t.Logf("Position: (%.6f, %.6f)", sprite.Position().X, sprite.Position().Y)
			t.Logf("RawPosition: (%.6f, %.6f)", sprite.rawPosition.X, sprite.rawPosition.Y)
			t.Logf("PivotOffset: (%.6f, %.6f)", sprite.pivotOffset.X, sprite.pivotOffset.Y)
		})
	}
}

// TestTopLeftCornerAfterRotation 测试旋转后左上角的位置
func TestTopLeftCornerAfterRotation(t *testing.T) {
	sprite := NewSprite()
	sprite.SetSize(100, 50)
	sprite.SetPivotWithAnchor(0.5, 0.5, false) // 中心轴
	sprite.SetPosition(200, 300)
	sprite.SetRotation(90) // 90 degrees (FairyGUI uses degrees, not radians)

	// 旋转前左上角在局部坐标 (0, 0)
	// 旋转后左上角应该移动
	topLeft := sprite.LocalToGlobal(Point{X: 0, Y: 0})

	// 计算预期位置：
	// pivotAsAnchor=false: SetPosition(200,300)表示左上角在(200,300)
	// pivot在中心(0.5,0.5)，即(50,25)，所以pivot全局位置在(250,325)
	// 左上角相对pivot为 (-50, -25)
	// 旋转90度后相对pivot为 (25, -50)
	// 全局坐标为 (250+25, 325-50) = (275, 275)
	wantX := 275.0
	wantY := 275.0

	tolerance := 1e-6
	if math.Abs(topLeft.X-wantX) > tolerance || math.Abs(topLeft.Y-wantY) > tolerance {
		t.Errorf("top-left after rotation = (%.6f, %.6f), want (%.6f, %.6f)",
			topLeft.X, topLeft.Y, wantX, wantY)
	}
}
