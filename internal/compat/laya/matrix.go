package laya

import "math"

// Matrix represents a 2D affine transform:
// [ a c tx ]
// [ b d ty ]
// [ 0 0  1 ]
type Matrix struct {
	A  float64
	B  float64
	C  float64
	D  float64
	Tx float64
	Ty float64
}

// Identity returns the identity matrix.
func Identity() Matrix {
	return Matrix{A: 1, D: 1}
}

// Multiply returns the product of two matrices (m * other).
func (m Matrix) Multiply(other Matrix) Matrix {
	return Matrix{
		A:  m.A*other.A + m.C*other.B,
		B:  m.B*other.A + m.D*other.B,
		C:  m.A*other.C + m.C*other.D,
		D:  m.B*other.C + m.D*other.D,
		Tx: m.A*other.Tx + m.C*other.Ty + m.Tx,
		Ty: m.B*other.Tx + m.D*other.Ty + m.Ty,
	}
}

// Apply transforms the point using the matrix.
func (m Matrix) Apply(pt Point) Point {
	return Point{
		X: m.A*pt.X + m.C*pt.Y + m.Tx,
		Y: m.B*pt.X + m.D*pt.Y + m.Ty,
	}
}

// Invert returns the inverse matrix. The second return value is false if the matrix is singular.
func (m Matrix) Invert() (Matrix, bool) {
	det := m.A*m.D - m.B*m.C
	if math.Abs(det) < 1e-9 {
		return Matrix{}, false
	}
	invDet := 1.0 / det
	return Matrix{
		A:  m.D * invDet,
		B:  -m.B * invDet,
		C:  -m.C * invDet,
		D:  m.A * invDet,
		Tx: (m.C*m.Ty - m.D*m.Tx) * invDet,
		Ty: (m.B*m.Tx - m.A*m.Ty) * invDet,
	}, true
}
