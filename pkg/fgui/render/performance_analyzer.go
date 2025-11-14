package render

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// PerformanceAnalyzer 性能分析器
// 用于分析渲染性能瓶颈，提供详细的性能报告
type PerformanceAnalyzer struct {
	mu sync.Mutex

	// 分析配置
	config AnalysisConfig

	// 采样数据
	samples []PerformanceSample

	// 分析结果
	analysisResult *AnalysisResult
}

// AnalysisConfig 分析配置
type AnalysisConfig struct {
	SampleInterval     time.Duration // 采样间隔
	MinSamples         int           // 最小样本数
	AnalysisWindow     time.Duration // 分析窗口
	EnableCPUProfiling bool          // 启用 CPU 分析
	EnableMemoryProfiling bool       // 启用内存分析
}

// DefaultAnalysisConfig 默认分析配置
var DefaultAnalysisConfig = AnalysisConfig{
	SampleInterval:     10 * time.Millisecond, // 10ms 采样一次
	MinSamples:         100,                   // 最小 100 个样本
	AnalysisWindow:     5 * time.Second,       // 5 秒分析窗口
	EnableCPUProfiling: true,
	EnableMemoryProfiling: false,
}

// PerformanceSample 性能样本
type PerformanceSample struct {
	Timestamp       time.Time
	FrameTime       time.Duration
	DrawCalls       int
	Triangles       int
	Batches         int
	MemoryAlloc     uint64
	CPUUsage        float64 // CPU 使用率 (0-100)
	GPUUsage        float64 // GPU 使用率 (0-100)
	CacheHitRate    float64
	BatchEfficiency float64 // 批处理效率 (0-1)
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	GeneratedAt      time.Time
	TotalSamples     int
	AnalysisWindow   time.Duration

	// 统计信息
	FrameTimeStats   DurationStats
	DrawCallStats    IntStats
	TriangleStats    IntStats
	BatchStats       IntStats
	MemoryStats      Uint64Stats
	CPUUsageStats    FloatStats
	GPUUsageStats    FloatStats
	CacheStats       FloatStats

	// 性能热点
	Hotspots []PerformanceHotspot

	// 建议
	Recommendations []PerformanceRecommendation

	// 性能评级
	Grade PerformanceGrade
}

// DurationStats 时间统计
type DurationStats struct {
	Min       time.Duration
	Max       time.Duration
	Mean      time.Duration
	Median    time.Duration
	StdDev    time.Duration
	P95       time.Duration
	P99       time.Duration
}

// IntStats 整数统计
type IntStats struct {
	Min    int
	Max    int
	Mean   float64
	Median float64
	StdDev float64
	P95    float64
	P99    float64
}

// Uint64Stats 无符号64位整数统计
type Uint64Stats struct {
	Min    uint64
	Max    uint64
	Mean   float64
	Median float64
	StdDev float64
	P95    float64
	P99    float64
}

// FloatStats 浮点数统计
type FloatStats struct {
	Min    float64
	Max    float64
	Mean   float64
	Median float64
	StdDev float64
	P95    float64
	P99    float64
}

// PerformanceHotspot 性能热点
type PerformanceHotspot struct {
	Name        string
	Category    HotspotCategory
	Impact      float64 // 影响程度 (0-100)
	Description string
	Value       float64
	Threshold   float64
}

// HotspotCategory 热点类别
type HotspotCategory int

const (
	HotspotCategoryCPU     HotspotCategory = iota
	HotspotCategoryGPU
	HotspotCategoryMemory
	HotspotCategoryDrawCalls
	HotspotCategoryBatching
	HotspotCategoryCache
)

// String 转换为字符串
func (c HotspotCategory) String() string {
	switch c {
	case HotspotCategoryCPU:
		return "CPU"
	case HotspotCategoryGPU:
		return "GPU"
	case HotspotCategoryMemory:
		return "Memory"
	case HotspotCategoryDrawCalls:
		return "DrawCalls"
	case HotspotCategoryBatching:
		return "Batching"
	case HotspotCategoryCache:
		return "Cache"
	default:
		return "Unknown"
	}
}

// PerformanceRecommendation 性能建议
type PerformanceRecommendation struct {
	Priority   RecommendationPriority
	Title      string
	Description string
	Impact     string
	Effort     string
	Category   HotspotCategory
}

// RecommendationPriority 建议优先级
type RecommendationPriority int

const (
	PriorityHigh   RecommendationPriority = iota
	PriorityMedium
	PriorityLow
)

// String 转换为字符串
func (p RecommendationPriority) String() string {
	switch p {
	case PriorityHigh:
		return "High"
	case PriorityMedium:
		return "Medium"
	case PriorityLow:
		return "Low"
	default:
		return "Unknown"
	}
}

// PerformanceGrade 性能等级
type PerformanceGrade struct {
	OverallScore float64 // 总体评分 (0-100)
	Letters      string  // 字母等级 (A, B, C, D, F)
	Color        string  // 颜色 (green, yellow, red)
	Description  string
}

// NewPerformanceAnalyzer 创建性能分析器
func NewPerformanceAnalyzer(config AnalysisConfig) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		config:  config,
		samples: make([]PerformanceSample, 0, 10000),
	}
}

// AddSample 添加性能样本
func (a *PerformanceAnalyzer) AddSample(sample PerformanceSample) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.samples = append(a.samples, sample)

	// 保持样本数量在合理范围内
	maxSamples := int(a.config.AnalysisWindow / a.config.SampleInterval)
	if len(a.samples) > maxSamples {
		a.samples = a.samples[len(a.samples)-maxSamples:]
	}
}

// Analyze 执行性能分析
func (a *PerformanceAnalyzer) Analyze() *AnalysisResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.samples) < a.config.MinSamples {
		return &AnalysisResult{
			GeneratedAt: time.Now(),
			TotalSamples: len(a.samples),
		}
	}

	// 筛选分析窗口内的样本
	windowStart := time.Now().Add(-a.config.AnalysisWindow)
	filteredSamples := make([]PerformanceSample, 0, len(a.samples))
	for _, sample := range a.samples {
		if sample.Timestamp.After(windowStart) {
			filteredSamples = append(filteredSamples, sample)
		}
	}

	result := &AnalysisResult{
		GeneratedAt:    time.Now(),
		TotalSamples:   len(filteredSamples),
		AnalysisWindow: a.config.AnalysisWindow,
	}

	// 分析各项指标
	result.FrameTimeStats = a.analyzeDurations(filteredSamples, func(s PerformanceSample) time.Duration { return s.FrameTime })
	result.DrawCallStats = a.analyzeInts(filteredSamples, func(s PerformanceSample) int { return s.DrawCalls })
	result.TriangleStats = a.analyzeInts(filteredSamples, func(s PerformanceSample) int { return s.Triangles })
	result.BatchStats = a.analyzeInts(filteredSamples, func(s PerformanceSample) int { return s.Batches })
	result.MemoryStats = a.analyzeUint64s(filteredSamples, func(s PerformanceSample) uint64 { return s.MemoryAlloc })
	result.CPUUsageStats = a.analyzeFloats(filteredSamples, func(s PerformanceSample) float64 { return s.CPUUsage })
	result.GPUUsageStats = a.analyzeFloats(filteredSamples, func(s PerformanceSample) float64 { return s.GPUUsage })
	result.CacheStats = a.analyzeFloats(filteredSamples, func(s PerformanceSample) float64 { return s.CacheHitRate })

	// 找出性能热点
	result.Hotspots = a.findHotspots(filteredSamples)

	// 生成建议
	result.Recommendations = a.generateRecommendations(result)

	// 计算性能等级
	result.Grade = a.calculateGrade(result)

	a.analysisResult = result
	return result
}

// analyzeDurations 分析时间数据
func (a *PerformanceAnalyzer) analyzeDurations(samples []PerformanceSample, selector func(PerformanceSample) time.Duration) DurationStats {
	if len(samples) == 0 {
		return DurationStats{}
	}

	values := make([]time.Duration, len(samples))
	for i, sample := range samples {
		values[i] = selector(sample)
	}

	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })

	n := len(values)
	sum := time.Duration(0)
	for _, v := range values {
		sum += v
	}

	mean := sum / time.Duration(n)
	median := values[n/2]
	p95 := values[int(float64(n)*0.95)]
	p99 := values[int(float64(n)*0.99)]

	// 计算标准差
	var variance time.Duration
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	stdDev := time.Duration(math.Sqrt(float64(variance) / float64(n)))

	return DurationStats{
		Min:    values[0],
		Max:    values[n-1],
		Mean:   mean,
		Median: median,
		StdDev: stdDev,
		P95:    p95,
		P99:    p99,
	}
}

// analyzeInts 分析整数数据
func (a *PerformanceAnalyzer) analyzeInts(samples []PerformanceSample, selector func(PerformanceSample) int) IntStats {
	if len(samples) == 0 {
		return IntStats{}
	}

	values := make([]float64, len(samples))
	for i, sample := range samples {
		values[i] = float64(selector(sample))
	}

	floatStats := a.calculateFloatStats(values)
	return IntStats{
		Min:    int(floatStats.Min),
		Max:    int(floatStats.Max),
		Mean:   floatStats.Mean,
		Median: floatStats.Median,
		StdDev: floatStats.StdDev,
		P95:    floatStats.P95,
		P99:    floatStats.P99,
	}
}

// analyzeUint64s 分析无符号64位整数
func (a *PerformanceAnalyzer) analyzeUint64s(samples []PerformanceSample, selector func(PerformanceSample) uint64) Uint64Stats {
	if len(samples) == 0 {
		return Uint64Stats{}
	}

	values := make([]float64, len(samples))
	for i, sample := range samples {
		values[i] = float64(selector(sample))
	}

	stats := a.calculateFloatStats(values)
	return Uint64Stats{
		Min:    uint64(stats.Min),
		Max:    uint64(stats.Max),
		Mean:   stats.Mean,
		Median: stats.Median,
		StdDev: stats.StdDev,
		P95:    stats.P95,
		P99:    stats.P99,
	}
}

// analyzeFloats 分析浮点数
func (a *PerformanceAnalyzer) analyzeFloats(samples []PerformanceSample, selector func(PerformanceSample) float64) FloatStats {
	if len(samples) == 0 {
		return FloatStats{}
	}

	values := make([]float64, len(samples))
	for i, sample := range samples {
		values[i] = selector(sample)
	}

	return a.calculateFloatStats(values)
}

// calculateFloatStats 计算浮点数统计
func (a *PerformanceAnalyzer) calculateFloatStats(values []float64) FloatStats {
	if len(values) == 0 {
		return FloatStats{}
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}

	mean := sum / float64(n)
	median := sorted[n/2]
	p95 := sorted[int(float64(n)*0.95)]
	p99 := sorted[int(float64(n)*0.99)]

	// 计算标准差
	variance := 0.0
	for _, v := range sorted {
		diff := v - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(n))

	return FloatStats{
		Min:    sorted[0],
		Max:    sorted[n-1],
		Mean:   mean,
		Median: median,
		StdDev: stdDev,
		P95:    p95,
		P99:    p99,
	}
}

// findHotspots 找出性能热点
func (a *PerformanceAnalyzer) findHotspots(samples []PerformanceSample) []PerformanceHotspot {
	var hotspots []PerformanceHotspot

	// 分析帧时间
	frameTimeStats := a.analyzeDurations(samples, func(s PerformanceSample) time.Duration { return s.FrameTime })
	if frameTimeStats.Mean > 16*time.Millisecond { // 超过 60fps
		hotspots = append(hotspots, PerformanceHotspot{
			Name:        "High Frame Time",
			Category:    HotspotCategoryCPU,
			Impact:      math.Min(100, float64(frameTimeStats.Mean)/16*100),
			Description: fmt.Sprintf("平均帧时间 %.2fms 超过 60fps 要求", frameTimeStats.Mean.Seconds()*1000),
			Value:       frameTimeStats.Mean.Seconds() * 1000,
			Threshold:   16.0,
		})
	}

	// 分析 DrawCall
	drawCallStats := a.analyzeInts(samples, func(s PerformanceSample) int { return s.DrawCalls })
	if drawCallStats.Mean > 100 {
		hotspots = append(hotspots, PerformanceHotspot{
			Name:        "High Draw Calls",
			Category:    HotspotCategoryDrawCalls,
			Impact:      math.Min(100, drawCallStats.Mean),
			Description: fmt.Sprintf("平均 DrawCall 数 %d 过高", int(drawCallStats.Mean)),
			Value:       drawCallStats.Mean,
			Threshold:   100.0,
		})
	}

	// 分析批处理效率
	batchEfficiency := 0.0
	if len(samples) > 0 {
		totalBatches := 0
		totalDrawCalls := 0
		for _, sample := range samples {
			totalBatches += sample.Batches
			totalDrawCalls += sample.DrawCalls
		}
		if totalDrawCalls > 0 {
			batchEfficiency = float64(totalBatches) / float64(totalDrawCalls)
		}
	}
	if batchEfficiency < 0.5 {
		hotspots = append(hotspots, PerformanceHotspot{
			Name:        "Low Batch Efficiency",
			Category:    HotspotCategoryBatching,
			Impact:      (0.5 - batchEfficiency) * 200,
			Description: fmt.Sprintf("批处理效率 %.2f%% 过低", batchEfficiency*100),
			Value:       batchEfficiency * 100,
			Threshold:   50.0,
		})
	}

	// 分析缓存命中率
	cacheStats := a.analyzeFloats(samples, func(s PerformanceSample) float64 { return s.CacheHitRate })
	if cacheStats.Mean < 70 {
		hotspots = append(hotspots, PerformanceHotspot{
			Name:        "Low Cache Hit Rate",
			Category:    HotspotCategoryCache,
			Impact:      (70 - cacheStats.Mean),
			Description: fmt.Sprintf("缓存命中率 %.2f%% 过低", cacheStats.Mean),
			Value:       cacheStats.Mean,
			Threshold:   70.0,
		})
	}

	// 按影响排序
	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].Impact > hotspots[j].Impact
	})

	return hotspots
}

// generateRecommendations 生成性能建议
func (a *PerformanceAnalyzer) generateRecommendations(result *AnalysisResult) []PerformanceRecommendation {
	var recommendations []PerformanceRecommendation

	for _, hotspot := range result.Hotspots {
		switch hotspot.Category {
		case HotspotCategoryCPU:
			recommendations = append(recommendations, PerformanceRecommendation{
				Priority:   PriorityHigh,
				Title:      "优化 CPU 计算",
				Description: "帧时间过长，建议使用批处理和对象池优化",
				Impact:     "显著提升帧率",
				Effort:     "中等",
				Category:   HotspotCategoryCPU,
			})
		case HotspotCategoryDrawCalls:
			recommendations = append(recommendations, PerformanceRecommendation{
				Priority:   PriorityHigh,
				Title:      "减少 DrawCall 数",
				Description: "DrawCall 数过高，建议使用批处理渲染器",
				Impact:     "显著提升性能",
				Effort:     "低",
				Category:   HotspotCategoryDrawCalls,
			})
		case HotspotCategoryBatching:
			recommendations = append(recommendations, PerformanceRecommendation{
				Priority:   PriorityMedium,
				Title:      "改进批处理策略",
				Description: "批处理效率低，建议优化材质分组",
				Impact:     "中等提升",
				Effort:     "中等",
				Category:   HotspotCategoryBatching,
			})
		case HotspotCategoryCache:
			recommendations = append(recommendations, PerformanceRecommendation{
				Priority:   PriorityMedium,
				Title:      "优化缓存策略",
				Description: "缓存命中率低，建议调整缓存大小和生命周期",
				Impact:     "中等提升",
				Effort:     "中等",
				Category:   HotspotCategoryCache,
			})
		}
	}

	return recommendations
}

// calculateGrade 计算性能等级
func (a *PerformanceAnalyzer) calculateGrade(result *AnalysisResult) PerformanceGrade {
	score := 100.0

	// 根据热点扣分
	for _, hotspot := range result.Hotspots {
		score -= hotspot.Impact * 0.5
	}

	// 归一化
	score = math.Max(0, math.Min(100, score))

	var letters string
	var color string
	var description string

	if score >= 90 {
		letters = "A"
		color = "green"
		description = "优秀"
	} else if score >= 80 {
		letters = "B"
		color = "lightgreen"
		description = "良好"
	} else if score >= 70 {
		letters = "C"
		color = "yellow"
		description = "一般"
	} else if score >= 60 {
		letters = "D"
		color = "orange"
		description = "较差"
	} else {
		letters = "F"
		color = "red"
		description = "需要优化"
	}

	return PerformanceGrade{
		OverallScore: score,
		Letters:      letters,
		Color:        color,
		Description:  description,
	}
}

// FormatReport 格式化分析报告
func (r *AnalysisResult) FormatReport() string {
	var s strings.Builder

	s.WriteString("=== 性能分析报告 ===\n")
	s.WriteString(fmt.Sprintf("生成时间: %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))
	s.WriteString(fmt.Sprintf("样本数: %d\n", r.TotalSamples))
	s.WriteString(fmt.Sprintf("分析窗口: %v\n\n", r.AnalysisWindow))

	// 帧时间统计
	s.WriteString("帧时间统计:\n")
	s.WriteString(fmt.Sprintf("  平均: %.2fms\n", r.FrameTimeStats.Mean.Seconds()*1000))
	s.WriteString(fmt.Sprintf("  中位数: %.2fms\n", r.FrameTimeStats.Median.Seconds()*1000))
	s.WriteString(fmt.Sprintf("  P95: %.2fms\n", r.FrameTimeStats.P95.Seconds()*1000))
	s.WriteString(fmt.Sprintf("  P99: %.2fms\n", r.FrameTimeStats.P99.Seconds()*1000))
	s.WriteString("\n")

	// DrawCall 统计
	s.WriteString("DrawCall 统计:\n")
	s.WriteString(fmt.Sprintf("  平均: %.2f\n", r.DrawCallStats.Mean))
	s.WriteString(fmt.Sprintf("  中位数: %.2f\n", r.DrawCallStats.Median))
	s.WriteString(fmt.Sprintf("  最大: %d\n", int(r.DrawCallStats.Max)))
	s.WriteString("\n")

	// 批处理统计
	s.WriteString("批处理统计:\n")
	s.WriteString(fmt.Sprintf("  平均: %.2f\n", r.BatchStats.Mean))
	s.WriteString(fmt.Sprintf("  中位数: %.2f\n", r.BatchStats.Median))
	s.WriteString("\n")

	// 缓存统计
	s.WriteString("缓存统计:\n")
	s.WriteString(fmt.Sprintf("  命中率: %.2f%%\n", r.CacheStats.Mean))
	s.WriteString("\n")

	// 性能热点
	if len(r.Hotspots) > 0 {
		s.WriteString("性能热点:\n")
		for _, hotspot := range r.Hotspots {
			categoryName := hotspot.Category.String()
			s.WriteString(fmt.Sprintf("  [%s] %s (影响: %.2f)\n",
				categoryName, hotspot.Name, hotspot.Impact))
			s.WriteString(fmt.Sprintf("    %s\n", hotspot.Description))
		}
		s.WriteString("\n")
	}

	// 建议
	if len(r.Recommendations) > 0 {
		s.WriteString("优化建议:\n")
		for _, rec := range r.Recommendations {
			priorityName := rec.Priority.String()
			s.WriteString(fmt.Sprintf("  [%s] %s\n", priorityName, rec.Title))
			s.WriteString(fmt.Sprintf("    %s\n", rec.Description))
			s.WriteString(fmt.Sprintf("    预期效果: %s, 实施难度: %s\n", rec.Impact, rec.Effort))
		}
		s.WriteString("\n")
	}

	// 性能等级
	s.WriteString("性能等级:\n")
	s.WriteString(fmt.Sprintf("  总体评分: %.2f\n", r.Grade.OverallScore))
	s.WriteString(fmt.Sprintf("  等级: %s (%s)\n", r.Grade.Letters, r.Grade.Color))
	s.WriteString(fmt.Sprintf("  评价: %s\n", r.Grade.Description))

	return s.String()
}
