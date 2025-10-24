package core

// TransitionInfo 暂存 FairyGUI Transition 的基础元数据，后续可用于构建完整的动画系统。
type TransitionInfo struct {
	Name          string
	Options       int
	AutoPlay      bool
	AutoPlayTimes int
	AutoPlayDelay float64
	ItemCount     int
	Items         []TransitionItem
	TotalDuration float64
}

// TransitionAction 描述 item 的行为类型（对齐 FairyGUI ActionType 枚举）。
type TransitionAction int

const (
	TransitionActionXY TransitionAction = iota
	TransitionActionSize
	TransitionActionScale
	TransitionActionPivot
	TransitionActionAlpha
	TransitionActionRotation
	TransitionActionColor
	TransitionActionAnimation
	TransitionActionVisible
	TransitionActionSound
	TransitionActionTransition
	TransitionActionShake
	TransitionActionColorFilter
	TransitionActionSkew
	TransitionActionText
	TransitionActionIcon
	TransitionActionUnknown
)

const (
	transitionCurveCRSpline = iota
	transitionCurveBezier
	transitionCurveCubicBezier
	transitionCurveStraight
)

// TransitionValue 存储单个值段的解析结果。
type TransitionValue struct {
	B1 bool
	B2 bool
	B3 bool

	F1 float64
	F2 float64
	F3 float64
	F4 float64

	Color uint32

	Playing   bool
	Frame     int
	Visible   bool
	DeltaTime float64

	Sound  string
	Volume float64

	TransName string
	PlayTimes int

	Amplitude float64
	Duration  float64
	OffsetX   float64
	OffsetY   float64

	Text string
}

// TransitionPathPoint 记录路径插值点。
type TransitionPathPoint struct {
	CurveType int
	X         float64
	Y         float64
	CX1       float64
	CY1       float64
	CX2       float64
	CY2       float64
}

// TransitionTween 捕获 tween 段的配置。
type TransitionTween struct {
	Duration float64
	EaseType int
	Repeat   int
	Yoyo     bool
	EndLabel string
	Path     []TransitionPathPoint
	Start    TransitionValue
	End      TransitionValue
}

// TransitionItem 对应 TS Item 结构，包含单条动作或 tween。
type TransitionItem struct {
	Time     float64
	TargetID string
	Type     TransitionAction
	Label    string
	Tween    *TransitionTween
	Value    TransitionValue
}
