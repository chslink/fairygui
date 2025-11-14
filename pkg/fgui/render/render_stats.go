package render

import (
	"fmt"
	"sync"
	"time"
)

// RenderStatsManager 渲染统计管理器
// 实时监控渲染性能，提供详细的统计数据
type RenderStatsManager struct {
	mu sync.Mutex

	// 历史记录
	history []FrameStats

	// 当前统计
	current FrameStats

	// 配置
	maxHistorySize int

	// 性能阈值
	thresholds PerformanceThresholds
}

// FrameStats 单帧统计
type FrameStats struct {
	FrameNumber      int64
	Timestamp        time.Time
	DrawCalls        int
	Triangles        int
	Batches          int
	CacheHitRate     float64
	FrameTime        time.Duration
	MemoryAlloc      uint64
	GCNum            uint32
	ObjectCount      int
	ActiveImages     int
	TotalImages      int
}

// PerformanceThresholds 性能阈值配置
type PerformanceThresholds struct {
	MaxFrameTime       time.Duration // 最大帧时间（60fps = 16.67ms）
	MinFPS             float64       // 最小 FPS
	MaxDrawCalls       int           // 最大 DrawCall 数
	MaxCacheMissRate   float64       // 最大缓存未命中率
	WarningFPS         float64       // 警告 FPS
}

// DefaultPerformanceThresholds 默认性能阈值
var DefaultPerformanceThresholds = PerformanceThresholds{
	MaxFrameTime:       16 * time.Millisecond, // 60 FPS
	MinFPS:             30.0,                  // 最低 30 FPS
	MaxDrawCalls:       1000,                  // 最大 1000 DrawCall
	MaxCacheMissRate:   30.0,                  // 最大 30% 缓存未命中
	WarningFPS:         50.0,                  // 警告 50 FPS
}

// NewRenderStatsManager 创建渲染统计管理器
func NewRenderStatsManager(maxHistorySize int) *RenderStatsManager {
	if maxHistorySize <= 0 {
		maxHistorySize = 300 // 默认保存 5 秒（60fps * 5）
	}

	return &RenderStatsManager{
		history:         make([]FrameStats, 0, maxHistorySize),
		maxHistorySize:  maxHistorySize,
		thresholds:      DefaultPerformanceThresholds,
		current:         FrameStats{},
	}
}

// StartFrame 开始帧统计
func (m *RenderStatsManager) StartFrame(frameNumber int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.current = FrameStats{
		FrameNumber: frameNumber,
		Timestamp:   time.Now(),
	}
}

// EndFrame 结束帧统计
func (m *RenderStatsManager) EndFrame(drawCalls, triangles, batches int, cacheHitRate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.current.DrawCalls = drawCalls
	m.current.Triangles = triangles
	m.current.Batches = batches
	m.current.CacheHitRate = cacheHitRate
	m.current.FrameTime = time.Since(m.current.Timestamp)

	// 添加到历史记录
	m.history = append(m.history, m.current)

	// 保持历史记录大小
	if len(m.history) > m.maxHistorySize {
		m.history = m.history[1:]
	}
}

// RecordMemory 记录内存信息
func (m *RenderStatsManager) RecordMemory(memAlloc uint64, gcNum uint32, objectCount, activeImages, totalImages int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.current.MemoryAlloc = memAlloc
	m.current.GCNum = gcNum
	m.current.ObjectCount = objectCount
	m.current.ActiveImages = activeImages
	m.current.TotalImages = totalImages
}

// GetCurrentStats 获取当前帧统计
func (m *RenderStatsManager) GetCurrentStats() FrameStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.current
}

// GetHistory 获取历史统计
func (m *RenderStatsManager) GetHistory() []FrameStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 返回副本
	history := make([]FrameStats, len(m.history))
	copy(history, m.history)
	return history
}

// GetAverageStats 获取平均统计（最近 N 帧）
func (m *RenderStatsManager) GetAverageStats(frameCount int) FrameStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	if frameCount <= 0 || len(m.history) == 0 {
		return FrameStats{}
	}

	// 限制帧数
	if frameCount > len(m.history) {
		frameCount = len(m.history)
	}

	// 计算平均值
	startIdx := len(m.history) - frameCount
	var total FrameStats
	count := float64(frameCount)

	for i := startIdx; i < len(m.history); i++ {
		stats := m.history[i]
		total.DrawCalls += stats.DrawCalls
		total.Triangles += stats.Triangles
		total.Batches += stats.Batches
		total.FrameTime += stats.FrameTime
		total.MemoryAlloc += stats.MemoryAlloc
	}

	total.DrawCalls /= int(count)
	total.Triangles /= int(count)
	total.Batches /= int(count)
	total.FrameTime = time.Duration(float64(total.FrameTime) / count)
	total.MemoryAlloc /= uint64(count)

	// 计算平均 FPS
	if total.FrameTime > 0 {
		fps := 1.0 / total.FrameTime.Seconds()
		total.FrameTime = time.Duration(float64(time.Second) / fps)
	}

	return total
}

// GetPerformanceReport 获取性能报告
func (m *RenderStatsManager) GetPerformanceReport() PerformanceReport {
	m.mu.Lock()
	defer m.mu.Unlock()

	report := PerformanceReport{
		Thresholds: m.thresholds,
	}

	if len(m.history) == 0 {
		return report
	}

	// 计算统计信息
	avgStats := m.GetAverageStats(60) // 最近 60 帧平均
	report.Average = avgStats

	// 当前帧
	report.Current = m.current

	// 计算 FPS
	if avgStats.FrameTime > 0 {
		report.AverageFPS = 1.0 / avgStats.FrameTime.Seconds()
		report.CurrentFPS = 1.0 / m.current.FrameTime.Seconds()
	}

	// 最小/最大 FPS
	var minFPS, maxFPS float64
	var minFrameTime, maxFrameTime time.Duration
	var minDrawCalls, maxDrawCalls int

	for _, stats := range m.history {
		if stats.FrameTime > 0 {
			fps := 1.0 / stats.FrameTime.Seconds()
			if minFPS == 0 || fps < minFPS {
				minFPS = fps
				minFrameTime = stats.FrameTime
			}
			if fps > maxFPS {
				maxFPS = fps
				maxFrameTime = stats.FrameTime
			}
		}

		if minDrawCalls == 0 || stats.DrawCalls < minDrawCalls {
			minDrawCalls = stats.DrawCalls
		}
		if stats.DrawCalls > maxDrawCalls {
			maxDrawCalls = stats.DrawCalls
		}
	}

	report.MinFPS = minFPS
	report.MaxFPS = maxFPS
	report.MinFrameTime = minFrameTime
	report.MaxFrameTime = maxFrameTime
	report.MinDrawCalls = minDrawCalls
	report.MaxDrawCalls = maxDrawCalls

	// 检查性能警告
	report.Warnings = m.checkWarnings(avgStats, report.AverageFPS)

	return report
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	Thresholds    PerformanceThresholds
	Current       FrameStats
	Average       FrameStats
	CurrentFPS    float64
	AverageFPS    float64
	MinFPS        float64
	MaxFPS        float64
	MinFrameTime  time.Duration
	MaxFrameTime  time.Duration
	MinDrawCalls  int
	MaxDrawCalls  int
	Warnings      []PerformanceWarning
}

// PerformanceWarning 性能警告
type PerformanceWarning struct {
	Level   WarningLevel
	Message string
	Value   float64
	Threshold float64
}

// WarningLevel 警告级别
type WarningLevel int

const (
	WarningLevelInfo    WarningLevel = iota
	WarningLevelWarning
	WarningLevelError
)

// String 转换为字符串
func (w WarningLevel) String() string {
	switch w {
	case WarningLevelInfo:
		return "INFO"
	case WarningLevelWarning:
		return "WARNING"
	case WarningLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// checkWarnings 检查性能警告
func (m *RenderStatsManager) checkWarnings(stats FrameStats, avgFPS float64) []PerformanceWarning {
	var warnings []PerformanceWarning

	// 检查 FPS
	if avgFPS < m.thresholds.WarningFPS {
		warnings = append(warnings, PerformanceWarning{
			Level:      WarningLevelWarning,
			Message:    fmt.Sprintf("Average FPS (%.2f) below warning threshold (%.2f)", avgFPS, m.thresholds.WarningFPS),
			Value:      avgFPS,
			Threshold:  m.thresholds.WarningFPS,
		})
	}

	if avgFPS < m.thresholds.MinFPS {
		warnings = append(warnings, PerformanceWarning{
			Level:      WarningLevelError,
			Message:    fmt.Sprintf("Average FPS (%.2f) below minimum threshold (%.2f)", avgFPS, m.thresholds.MinFPS),
			Value:      avgFPS,
			Threshold:  m.thresholds.MinFPS,
		})
	}

	// 检查 DrawCall 数
	if stats.DrawCalls > m.thresholds.MaxDrawCalls {
		warnings = append(warnings, PerformanceWarning{
			Level:      WarningLevelWarning,
			Message:    fmt.Sprintf("DrawCalls (%d) above threshold (%d)", stats.DrawCalls, m.thresholds.MaxDrawCalls),
			Value:      float64(stats.DrawCalls),
			Threshold:  float64(m.thresholds.MaxDrawCalls),
		})
	}

	// 检查缓存命中率
	cacheMissRate := 100.0 - stats.CacheHitRate
	if cacheMissRate > m.thresholds.MaxCacheMissRate {
		warnings = append(warnings, PerformanceWarning{
			Level:      WarningLevelInfo,
			Message:    fmt.Sprintf("Cache miss rate (%.2f%%) above threshold (%.2f%%)", cacheMissRate, m.thresholds.MaxCacheMissRate),
			Value:      cacheMissRate,
			Threshold:  m.thresholds.MaxCacheMissRate,
		})
	}

	// 检查帧时间
	if stats.FrameTime > m.thresholds.MaxFrameTime {
		warnings = append(warnings, PerformanceWarning{
			Level:      WarningLevelWarning,
			Message:    fmt.Sprintf("Frame time (%.2fms) above threshold (%.2fms)", stats.FrameTime.Seconds()*1000, m.thresholds.MaxFrameTime.Seconds()*1000),
			Value:      stats.FrameTime.Seconds() * 1000,
			Threshold:  m.thresholds.MaxFrameTime.Seconds() * 1000,
		})
	}

	return warnings
}

// FormatReport 格式化报告为字符串
func (p PerformanceReport) FormatReport() string {
	var s string

	s += "=== 渲染性能报告 ===\n\n"

	// 当前帧
	s += "当前帧:\n"
	s += fmt.Sprintf("  FPS: %.2f\n", p.CurrentFPS)
	s += fmt.Sprintf("  帧时间: %.2fms\n", p.Current.FrameTime.Seconds()*1000)
	s += fmt.Sprintf("  DrawCalls: %d\n", p.Current.DrawCalls)
	s += fmt.Sprintf("  三角形: %d\n", p.Current.Triangles)
	s += fmt.Sprintf("  批处理: %d\n", p.Current.Batches)
	s += fmt.Sprintf("  缓存命中率: %.2f%%\n", p.Current.CacheHitRate)
	s += "\n"

	// 平均帧
	s += "平均帧 (最近60帧):\n"
	s += fmt.Sprintf("  FPS: %.2f\n", p.AverageFPS)
	s += fmt.Sprintf("  帧时间: %.2fms\n", p.Average.FrameTime.Seconds()*1000)
	s += fmt.Sprintf("  DrawCalls: %d\n", p.Average.DrawCalls)
	s += fmt.Sprintf("  三角形: %d\n", p.Average.Triangles)
	s += fmt.Sprintf("  批处理: %d\n", p.Average.Batches)
	s += fmt.Sprintf("  缓存命中率: %.2f%%\n", p.Average.CacheHitRate)
	s += "\n"

	// 范围
	s += "范围:\n"
	s += fmt.Sprintf("  FPS: %.2f - %.2f\n", p.MinFPS, p.MaxFPS)
	s += fmt.Sprintf("  帧时间: %.2fms - %.2fms\n", p.MinFrameTime.Seconds()*1000, p.MaxFrameTime.Seconds()*1000)
	s += fmt.Sprintf("  DrawCalls: %d - %d\n", p.MinDrawCalls, p.MaxDrawCalls)
	s += "\n"

	// 警告
	if len(p.Warnings) > 0 {
		s += "警告:\n"
		for _, w := range p.Warnings {
			s += fmt.Sprintf("  [%s] %s\n", w.Level.String(), w.Message)
		}
	} else {
		s += "警告: 无\n"
	}

	return s
}

// RenderProfiler 渲染性能分析器
type RenderProfiler struct {
	mu      sync.Mutex
	enabled bool
	events  []ProfilerEvent
}

// ProfilerEvent 性能分析事件
type ProfilerEvent struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// NewRenderProfiler 创建渲染性能分析器
func NewRenderProfiler() *RenderProfiler {
	return &RenderProfiler{
		enabled: true,
		events:  make([]ProfilerEvent, 0, 100),
	}
}

// StartEvent 开始事件
func (p *RenderProfiler) StartEvent(name string) {
	if !p.enabled {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.events = append(p.events, ProfilerEvent{
		Name:      name,
		StartTime: time.Now(),
	})
}

// EndEvent 结束事件
func (p *RenderProfiler) EndEvent(name string) {
	if !p.enabled {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 找到最后一个匹配的事件
	for i := len(p.events) - 1; i >= 0; i-- {
		if p.events[i].Name == name && p.events[i].EndTime.IsZero() {
			p.events[i].EndTime = time.Now()
			p.events[i].Duration = p.events[i].EndTime.Sub(p.events[i].StartTime)
			break
		}
	}
}

// GetEvents 获取所有事件
func (p *RenderProfiler) GetEvents() []ProfilerEvent {
	p.mu.Lock()
	defer p.mu.Unlock()

	events := make([]ProfilerEvent, len(p.events))
	copy(events, p.events)
	return events
}

// Clear 清除事件
func (p *RenderProfiler) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.events = p.events[:0]
}

// Enable 启用/禁用
func (p *RenderProfiler) Enable(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.enabled = enabled
	if !enabled {
		p.events = p.events[:0]
	}
}
