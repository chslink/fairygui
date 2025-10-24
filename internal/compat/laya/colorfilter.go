package laya

import "math"

var identityColorMatrix = [20]float64{
	1, 0, 0, 0, 0,
	0, 1, 0, 0, 0,
	0, 0, 1, 0, 0,
	0, 0, 0, 1, 0,
}

const (
	lumaR = 0.299
	lumaG = 0.587
	lumaB = 0.114
)

func computeColorMatrix(brightness, contrast, saturation, hue float64) ([20]float64, bool) {
	if almostZero(brightness) && almostZero(contrast) && almostZero(saturation) && almostZero(hue) {
		return identityColorMatrix, false
	}
	m := identityColorMatrix
	adjustHue(&m, hue)
	adjustContrast(&m, contrast)
	adjustBrightness(&m, brightness)
	adjustSaturation(&m, saturation)
	return m, true
}

func adjustBrightness(matrix *[20]float64, value float64) {
	v := clamp(value, 1) * 255
	multiplyMatrix(matrix, [20]float64{
		1, 0, 0, 0, v,
		0, 1, 0, 0, v,
		0, 0, 1, 0, v,
		0, 0, 0, 1, 0,
	})
}

func adjustContrast(matrix *[20]float64, value float64) {
	v := clamp(value, 1)
	s := v + 1
	o := 128 * (1 - s)
	multiplyMatrix(matrix, [20]float64{
		s, 0, 0, 0, o,
		0, s, 0, 0, o,
		0, 0, s, 0, o,
		0, 0, 0, 1, 0,
	})
}

func adjustSaturation(matrix *[20]float64, value float64) {
	v := clamp(value, 1)
	v += 1
	invSat := 1 - v
	invLumR := invSat * lumaR
	invLumG := invSat * lumaG
	invLumB := invSat * lumaB
	multiplyMatrix(matrix, [20]float64{
		invLumR + v, invLumG, invLumB, 0, 0,
		invLumR, invLumG + v, invLumB, 0, 0,
		invLumR, invLumG, invLumB + v, 0, 0,
		0, 0, 0, 1, 0,
	})
}

func adjustHue(matrix *[20]float64, value float64) {
	v := clamp(value, 1)
	v *= math.Pi
	cosVal := math.Cos(v)
	sinVal := math.Sin(v)
	multiplyMatrix(matrix, [20]float64{
		(lumaR + (cosVal * (1 - lumaR))) + (sinVal * -lumaR), (lumaG + (cosVal * -lumaG)) + (sinVal * -lumaG), (lumaB + (cosVal * -lumaB)) + (sinVal * (1 - lumaB)), 0, 0,
		(lumaR + (cosVal * -lumaR)) + (sinVal * 0.143), (lumaG + (cosVal * (1 - lumaG))) + (sinVal * 0.14), (lumaB + (cosVal * -lumaB)) + (sinVal * -0.283), 0, 0,
		(lumaR + (cosVal * -lumaR)) + (sinVal * -(1 - lumaR)), (lumaG + (cosVal * -lumaG)) + (sinVal * lumaG), (lumaB + (cosVal * (1 - lumaB))) + (sinVal * lumaB), 0, 0,
		0, 0, 0, 1, 0,
	})
}

func multiplyMatrix(base *[20]float64, other [20]float64) {
	var result [20]float64
	for y := 0; y < 4; y++ {
		for x := 0; x < 5; x++ {
			idx := y*5 + x
			result[idx] = other[y*5+0]*base[x] + other[y*5+1]*base[x+5] + other[y*5+2]*base[x+10] + other[y*5+3]*base[x+15]
			if x == 4 {
				result[idx] += other[y*5+4]
			}
		}
	}
	*base = result
}

func clamp(value, limit float64) float64 {
	if value > limit {
		return limit
	}
	if value < -limit {
		return -limit
	}
	return value
}

func almostZero(value float64) bool {
	return math.Abs(value) <= 1e-6
}
