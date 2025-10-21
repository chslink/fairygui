package tween

import "math"

// EaseType mirrors FairyGUI's EaseType.
type EaseType int

const (
	EaseTypeLinear EaseType = iota
	EaseTypeSineIn
	EaseTypeSineOut
	EaseTypeSineInOut
	EaseTypeQuadIn
	EaseTypeQuadOut
	EaseTypeQuadInOut
	EaseTypeCubicIn
	EaseTypeCubicOut
	EaseTypeCubicInOut
	EaseTypeQuartIn
	EaseTypeQuartOut
	EaseTypeQuartInOut
	EaseTypeQuintIn
	EaseTypeQuintOut
	EaseTypeQuintInOut
	EaseTypeExpoIn
	EaseTypeExpoOut
	EaseTypeExpoInOut
	EaseTypeCircIn
	EaseTypeCircOut
	EaseTypeCircInOut
	EaseTypeElasticIn
	EaseTypeElasticOut
	EaseTypeElasticInOut
	EaseTypeBackIn
	EaseTypeBackOut
	EaseTypeBackInOut
	EaseTypeBounceIn
	EaseTypeBounceOut
	EaseTypeBounceInOut
)

const (
	piOverTwo = math.Pi * 0.5
	twoPi     = math.Pi * 2
)

func easeValue(ease EaseType, time, duration, overshootOrAmplitude, period float64) float64 {
	switch ease {
	case EaseTypeLinear:
		return time / duration
	case EaseTypeSineIn:
		return -math.Cos(time/duration*piOverTwo) + 1
	case EaseTypeSineOut:
		return math.Sin(time / duration * piOverTwo)
	case EaseTypeSineInOut:
		return -0.5 * (math.Cos(math.Pi*time/duration) - 1)
	case EaseTypeQuadIn:
		t := time / duration
		return t * t
	case EaseTypeQuadOut:
		t := time / duration
		return -(t * (t - 2))
	case EaseTypeQuadInOut:
		t := time / duration * 0.5
		if t < 1 {
			return 0.5 * t * t
		}
		t--
		return -0.5 * (t*(t-2) - 1)
	case EaseTypeCubicIn:
		t := time / duration
		return t * t * t
	case EaseTypeCubicOut:
		t := time/duration - 1
		return t*t*t + 1
	case EaseTypeCubicInOut:
		t := time / duration * 0.5
		if t < 1 {
			return 0.5 * t * t * t
		}
		t -= 2
		return 0.5 * (t*t*t + 2)
	case EaseTypeQuartIn:
		t := time / duration
		return t * t * t * t
	case EaseTypeQuartOut:
		t := time/duration - 1
		return -(t*t*t*t - 1)
	case EaseTypeQuartInOut:
		t := time / duration * 0.5
		if t < 1 {
			return 0.5 * t * t * t * t
		}
		t -= 2
		return -0.5 * (t*t*t*t - 2)
	case EaseTypeQuintIn:
		t := time / duration
		return t * t * t * t * t
	case EaseTypeQuintOut:
		t := time/duration - 1
		return t*t*t*t*t + 1
	case EaseTypeQuintInOut:
		t := time / duration * 0.5
		if t < 1 {
			return 0.5 * t * t * t * t * t
		}
		t -= 2
		return 0.5 * (t*t*t*t*t + 2)
	case EaseTypeExpoIn:
		if time == 0 {
			return 0
		}
		return math.Pow(2, 10*(time/duration-1))
	case EaseTypeExpoOut:
		if time == duration {
			return 1
		}
		return -math.Pow(2, -10*time/duration) + 1
	case EaseTypeExpoInOut:
		if time == 0 {
			return 0
		}
		if time == duration {
			return 1
		}
		t := time / duration * 0.5
		if t < 1 {
			return 0.5 * math.Pow(2, 10*(t-1))
		}
		t--
		return 0.5 * (-math.Pow(2, -10*t) + 2)
	case EaseTypeCircIn:
		t := time / duration
		return -(math.Sqrt(1-t*t) - 1)
	case EaseTypeCircOut:
		t := time/duration - 1
		return math.Sqrt(1 - t*t)
	case EaseTypeCircInOut:
		t := time / duration * 0.5
		if t < 1 {
			return -0.5 * (math.Sqrt(1-t*t) - 1)
		}
		t -= 2
		return 0.5 * (math.Sqrt(1-t*t) + 1)
	case EaseTypeElasticIn:
		if time == 0 {
			return 0
		}
		t := time/duration - 1
		if t == 0 {
			return 0
		}
		if period == 0 {
			period = duration * 0.3
		}
		var s float64
		if overshootOrAmplitude < 1 {
			overshootOrAmplitude = 1
			s = period / 4
		} else {
			s = period / twoPi * math.Asin(1/overshootOrAmplitude)
		}
		return -(overshootOrAmplitude * math.Pow(2, 10*t) * math.Sin((t*duration-s)*twoPi/period))
	case EaseTypeElasticOut:
		if time == 0 {
			return 0
		}
		t := time / duration
		if t == 1 {
			return 1
		}
		if period == 0 {
			period = duration * 0.3
		}
		var s float64
		if overshootOrAmplitude < 1 {
			overshootOrAmplitude = 1
			s = period / 4
		} else {
			s = period / twoPi * math.Asin(1/overshootOrAmplitude)
		}
		return overshootOrAmplitude*math.Pow(2, -10*t)*math.Sin((t*duration-s)*twoPi/period) + 1
	case EaseTypeElasticInOut:
		if time == 0 {
			return 0
		}
		t := time / duration * 0.5
		if t == 1 {
			return 1
		}
		if period == 0 {
			period = duration * (0.3 * 1.5)
		}
		var s float64
		if overshootOrAmplitude < 1 {
			overshootOrAmplitude = 1
			s = period / 4
		} else {
			s = period / twoPi * math.Asin(1/overshootOrAmplitude)
		}
		if t < 1 {
			t--
			return -0.5 * (overshootOrAmplitude * math.Pow(2, 10*t) * math.Sin((t*duration-s)*twoPi/period))
		}
		t--
		return overshootOrAmplitude*math.Pow(2, -10*t)*math.Sin((t*duration-s)*twoPi/period)*0.5 + 1
	case EaseTypeBackIn:
		t := time / duration
		return t * t * ((overshootOrAmplitude+1)*t - overshootOrAmplitude)
	case EaseTypeBackOut:
		t := time/duration - 1
		return t*t*((overshootOrAmplitude+1)*t+overshootOrAmplitude) + 1
	case EaseTypeBackInOut:
		overshootOrAmplitude *= 1.525
		t := time / duration * 0.5
		if t < 1 {
			return 0.5 * (t * t * (((overshootOrAmplitude + 1) * t) - overshootOrAmplitude))
		}
		t -= 2
		return 0.5 * (t*t*((overshootOrAmplitude+1)*t+overshootOrAmplitude) + 2)
	case EaseTypeBounceIn:
		return 1 - bounceEaseOut(duration-time, duration)
	case EaseTypeBounceOut:
		return bounceEaseOut(time, duration)
	case EaseTypeBounceInOut:
		if time < duration*0.5 {
			return bounceEaseIn(time*2, duration) * 0.5
		}
		return bounceEaseOut(time*2-duration, duration)*0.5 + 0.5
	default:
		// fallback to quad out
		t := time / duration
		return -(t * (t - 2))
	}
}

func bounceEaseIn(time, duration float64) float64 {
	return 1 - bounceEaseOut(duration-time, duration)
}

func bounceEaseOut(time, duration float64) float64 {
	t := time / duration
	if t < (1 / 2.75) {
		return 7.5625 * t * t
	}
	if t < (2 / 2.75) {
		t -= 1.5 / 2.75
		return 7.5625*t*t + 0.75
	}
	if t < (2.5 / 2.75) {
		t -= 2.25 / 2.75
		return 7.5625*t*t + 0.9375
	}
	t -= 2.625 / 2.75
	return 7.5625*t*t + 0.984375
}

func bounceEaseInOut(time, duration float64) float64 {
	if time < duration*0.5 {
		return bounceEaseIn(time*2, duration) * 0.5
	}
	return bounceEaseOut(time*2-duration, duration)*0.5 + 0.5
}
