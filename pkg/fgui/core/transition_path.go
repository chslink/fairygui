package core

import (
	"math"

	"github.com/chslink/fairygui/pkg/fgui/tween"
)

type transitionPath struct {
	segments    []transitionPathSegment
	totalLength float64
}

type transitionPathSegment struct {
	kind   int
	points []vec2
	length float64
}

type vec2 struct {
	x float64
	y float64
}

func newTransitionPath(points []TransitionPathPoint) tween.Path {
	if len(points) < 2 {
		return nil
	}
	var segments []transitionPathSegment
	var spline []vec2
	prev := points[0]
	if prev.CurveType == transitionCurveCRSpline {
		spline = append(spline, vec2{prev.X, prev.Y})
	}
	total := 0.0
	for i := 1; i < len(points); i++ {
		current := points[i]
		if prev.CurveType != transitionCurveCRSpline {
			seg := transitionPathSegment{
				kind: prev.CurveType,
			}
			switch prev.CurveType {
			case transitionCurveStraight:
				seg.points = []vec2{
					{prev.X, prev.Y},
					{current.X, current.Y},
				}
			case transitionCurveBezier:
				seg.points = []vec2{
					{prev.X, prev.Y},
					{current.X, current.Y},
					{prev.CX1, prev.CY1},
				}
			case transitionCurveCubicBezier:
				seg.points = []vec2{
					{prev.X, prev.Y},
					{current.X, current.Y},
					{prev.CX1, prev.CY1},
					{prev.CX2, prev.CY2},
				}
			default:
				seg.points = []vec2{
					{prev.X, prev.Y},
					{current.X, current.Y},
				}
			}
			seg.length = distance(seg.points[0], seg.points[1])
			total += seg.length
			segments = append(segments, seg)
		}

		if current.CurveType != transitionCurveCRSpline {
			if len(spline) > 0 {
				spline = append(spline, vec2{current.X, current.Y})
				seg := buildSplineSegment(spline)
				total += seg.length
				segments = append(segments, seg)
				spline = spline[:0]
			}
		} else {
			spline = append(spline, vec2{current.X, current.Y})
		}
		prev = current
	}

	if len(spline) > 1 {
		seg := buildSplineSegment(spline)
		total += seg.length
		segments = append(segments, seg)
	}

	if len(segments) == 0 {
		return nil
	}
	return &transitionPath{
		segments:    segments,
		totalLength: total,
	}
}

func (p *transitionPath) PointAt(t float64) (float64, float64) {
	if p == nil || len(p.segments) == 0 {
		return 0, 0
	}
	if t <= 0 {
		return p.segments[0].pointAt(0)
	}
	if t >= 1 {
		last := p.segments[len(p.segments)-1]
		return last.pointAt(1)
	}

	length := t * p.totalLength
	for i, seg := range p.segments {
		length -= seg.length
		if length < 0 || (seg.length == 0 && i == len(p.segments)-1) {
			local := 1.0
			if seg.length > 0 {
				local = 1 + length/seg.length
			}
			if local < 0 {
				local = 0
			} else if local > 1 {
				local = 1
			}
			return seg.pointAt(local)
		}
	}
	last := p.segments[len(p.segments)-1]
	return last.pointAt(1)
}

func (s transitionPathSegment) pointAt(t float64) (float64, float64) {
	switch s.kind {
	case transitionCurveStraight:
		if len(s.points) < 2 {
			return 0, 0
		}
		start := s.points[0]
		end := s.points[1]
		return lerpVec(start, end, t)
	case transitionCurveBezier, transitionCurveCubicBezier:
		if len(s.points) < 3 {
			return 0, 0
		}
		start := s.points[0]
		end := s.points[1]
		cp0 := s.points[2]
		u := 1 - t
		if len(s.points) == 4 {
			cp1 := s.points[3]
			x := u*u*u*start.x + 3*u*u*t*cp0.x + 3*u*t*t*cp1.x + t*t*t*end.x
			y := u*u*u*start.y + 3*u*u*t*cp0.y + 3*u*t*t*cp1.y + t*t*t*end.y
			return x, y
		}
		x := u*u*start.x + 2*u*t*cp0.x + t*t*end.x
		y := u*u*start.y + 2*u*t*cp0.y + t*t*end.y
		return x, y
	case transitionCurveCRSpline:
		if len(s.points) < 4 {
			return 0, 0
		}
		segmentCount := len(s.points) - 3
		if segmentCount <= 0 {
			last := s.points[len(s.points)-1]
			return last.x, last.y
		}
		if t >= 1 {
			return catmullRom(s.points[len(s.points)-4], s.points[len(s.points)-3], s.points[len(s.points)-2], s.points[len(s.points)-1], 1)
		}
		pos := t * float64(segmentCount)
		index := int(math.Floor(pos))
		if index < 0 {
			index = 0
		}
		if index > segmentCount-1 {
			index = segmentCount - 1
		}
		local := pos - float64(index)
		return catmullRom(s.points[index], s.points[index+1], s.points[index+2], s.points[index+3], local)
	default:
		if len(s.points) == 0 {
			return 0, 0
		}
		last := s.points[len(s.points)-1]
		return last.x, last.y
	}
}

func buildSplineSegment(points []vec2) transitionPathSegment {
	if len(points) < 2 {
		return transitionPathSegment{
			kind:   transitionCurveCRSpline,
			points: append([]vec2(nil), points...),
		}
	}
	segPoints := make([]vec2, 0, len(points)+3)
	first := points[0]
	last := points[len(points)-1]
	segPoints = append(segPoints, first)
	segPoints = append(segPoints, points...)
	segPoints = append(segPoints, last, last)
	length := 0.0
	for i := 1; i < len(segPoints); i++ {
		length += distance(segPoints[i-1], segPoints[i])
	}
	return transitionPathSegment{
		kind:   transitionCurveCRSpline,
		points: segPoints,
		length: length,
	}
}

func distance(a, b vec2) float64 {
	dx := b.x - a.x
	dy := b.y - a.y
	return math.Hypot(dx, dy)
}

func lerpVec(a, b vec2, t float64) (float64, float64) {
	x := a.x + (b.x-a.x)*t
	y := a.y + (b.y-a.y)*t
	return x, y
}

func catmullRom(p0, p1, p2, p3 vec2, t float64) (float64, float64) {
	t0 := ((-t+2)*t - 1) * t * 0.5
	t1 := (((3*t-5)*t)*t + 2) * 0.5
	t2 := ((-3*t+4)*t + 1) * t * 0.5
	t3 := ((t - 1) * t * t) * 0.5
	x := p0.x*t0 + p1.x*t1 + p2.x*t2 + p3.x*t3
	y := p0.y*t0 + p1.y*t1 + p2.y*t2 + p3.y*t3
	return x, y
}
