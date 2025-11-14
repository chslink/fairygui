package render

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// DebugConfig 调试配置
type DebugConfig struct {
	Enabled              bool
	ShowFPS              bool
	ShowDrawCalls        bool
	ShowTriangles        bool
	ShowMemoryUsage      bool
	ShowCacheStats       bool
	ShowPerformanceWarnings bool
	LogToFile            bool
	LogFilePath          string
	UpdateInterval       time.Duration
}

// DefaultDebugConfig 默认调试配置
var DefaultDebugConfig = DebugConfig{
	Enabled:                   false,
	ShowFPS:                   true,
	ShowDrawCalls:             true,
	ShowTriangles:             false,
	ShowMemoryUsage:           true,
	ShowCacheStats:            true,
	ShowPerformanceWarnings:   true,
	LogToFile:                 false,
	LogFilePath:               "render-debug.log",
	UpdateInterval:            time.Second,
}

// RenderDebugger 渲染调试器
type RenderDebugger struct {
	mu      sync.Mutex
	config  DebugConfig
	logFile *os.File

	// 统计
	statsManager *RenderStatsManager
	profiler     *RenderProfiler

	// 控制
	running      bool
	lastUpdate   time.Time
	frameCounter int64
}

// NewRenderDebugger 创建渲染调试器
func NewRenderDebugger(config DebugConfig) (*RenderDebugger, error) {
	var logFile *os.File
	if config.LogToFile {
		var err error
		logFile, err = os.Create(config.LogFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create log file: %w", err)
		}
	}

	return &RenderDebugger{
		config:       config,
		logFile:      logFile,
		statsManager: NewRenderStatsManager(300),
		profiler:     NewRenderProfiler(),
		running:      config.Enabled,
		lastUpdate:   time.Now(),
	}, nil
}

// Start 开始调试
func (d *RenderDebugger) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.running = true
	d.lastUpdate = time.Now()
	d.frameCounter = 0

	if d.config.LogToFile && d.logFile != nil {
		d.logFile.WriteString(fmt.Sprintf("=== 渲染调试开始: %s ===\n", time.Now().Format("2006-01-02 15:04:05")))
	}
}

// Stop 停止调试
func (d *RenderDebugger) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.running = false

	if d.logFile != nil {
		d.logFile.WriteString(fmt.Sprintf("=== 渲染调试结束: %s ===\n\n", time.Now().Format("2006-01-02 15:04:05")))
		d.logFile.Sync()
	}
}

// Close 关闭调试器
func (d *RenderDebugger) Close() {
	d.Stop()

	if d.logFile != nil {
		d.logFile.Close()
	}
}

// Update 更新调试信息
func (d *RenderDebugger) Update(stats IntegratedStats) {
	if !d.running {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.frameCounter++

	// 更新统计
	now := time.Now()
	elapsed := now.Sub(d.lastUpdate)

	if elapsed >= d.config.UpdateInterval {
		d.outputDebugInfo(stats)
		d.lastUpdate = now
	}
}

// outputDebugInfo 输出调试信息
func (d *RenderDebugger) outputDebugInfo(stats IntegratedStats) {
	var output strings.Builder

	// 标题
	output.WriteString(fmt.Sprintf("=== 帧 #%d (%.2f FPS) ===\n", stats.FrameCount, d.calculateFPS(stats)))

	// FPS
	if d.config.ShowFPS {
		output.WriteString(fmt.Sprintf("FPS: %.2f\n", d.calculateFPS(stats)))
	}

	// DrawCall
	if d.config.ShowDrawCalls {
		output.WriteString(fmt.Sprintf("DrawCalls: %d\n", stats.TotalDrawCalls))
	}

	// 三角形
	if d.config.ShowTriangles {
		output.WriteString(fmt.Sprintf("Triangles: %d\n", stats.TotalTriangles))
	}

	// 批处理
	output.WriteString(fmt.Sprintf("Batches: %d\n", stats.BatchCount))

	// 缓存统计
	if d.config.ShowCacheStats {
		cacheHitRate := 0.0
		if stats.CacheHits+stats.CacheMisses > 0 {
			cacheHitRate = float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses) * 100
		}
		output.WriteString(fmt.Sprintf("CacheHit: %d, CacheMiss: %d, HitRate: %.2f%%\n",
			stats.CacheHits, stats.CacheMisses, cacheHitRate))
	}

	// 内存使用
	if d.config.ShowMemoryUsage {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		output.WriteString(fmt.Sprintf("Memory: Alloc=%d KB, TotalAlloc=%d KB, NumGC=%d\n",
			m.Alloc/1024, m.TotalAlloc/1024, m.NumGC))
	}

	// 性能警告
	if d.config.ShowPerformanceWarnings {
		report := d.statsManager.GetPerformanceReport()
		if len(report.Warnings) > 0 {
			output.WriteString("\n警告:\n")
			for _, w := range report.Warnings {
				output.WriteString(fmt.Sprintf("  [%s] %s\n", w.Level.String(), w.Message))
			}
		}
	}

	output.WriteString("\n")

	// 输出到控制台
	fmt.Print(output.String())

	// 输出到文件
	if d.logFile != nil {
		d.logFile.WriteString(output.String())
		d.logFile.Sync()
	}
}

// calculateFPS 计算 FPS
func (d *RenderDebugger) calculateFPS(stats IntegratedStats) float64 {
	if stats.AvgFrameTime == 0 {
		return 0
	}
	return 1000.0 / (stats.AvgFrameTime / 1000000.0)
}

// RecordFrame 记录帧统计
func (d *RenderDebugger) RecordFrame(drawCalls, triangles, batches int, cacheHitRate float64) {
	if !d.running {
		return
	}

	d.statsManager.EndFrame(drawCalls, triangles, batches, cacheHitRate)
}

// GetStatsManager 获取统计管理器
func (d *RenderDebugger) GetStatsManager() *RenderStatsManager {
	return d.statsManager
}

// GetProfiler 获取性能分析器
func (d *RenderDebugger) GetProfiler() *RenderProfiler {
	return d.profiler
}

// SetConfig 更新配置
func (d *RenderDebugger) SetConfig(config DebugConfig) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 如果日志配置变化，重新打开文件
	if config.LogToFile != d.config.LogToFile || config.LogFilePath != d.config.LogFilePath {
		if d.logFile != nil {
			d.logFile.Close()
			d.logFile = nil
		}
		if config.LogToFile {
			var err error
			d.logFile, err = os.Create(config.LogFilePath)
			if err != nil {
				return fmt.Errorf("failed to create log file: %w", err)
			}
		}
	}

	d.config = config
	return nil
}

// RenderDebugOverlay 渲染调试覆盖层
// 在屏幕上显示调试信息
type RenderDebugOverlay struct {
	mu      sync.Mutex
	enabled bool
	font    *ebiten.Image
	fontMap map[rune][4]int // 字符到像素位置的映射
	scale   float64
	color   [4]byte
}

// NewRenderDebugOverlay 创建调试覆盖层
func NewRenderDebugOverlay(scale float64) *RenderDebugOverlay {
	return &RenderDebugOverlay{
		scale: scale,
		color: [4]byte{255, 255, 255, 255}, // 白色
		// TODO: 初始化字体
	}
}

// Enable 启用/禁用
func (o *RenderDebugOverlay) Enable(enabled bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.enabled = enabled
}

// Render 渲染覆盖层
func (o *RenderDebugOverlay) Render(target *ebiten.Image, stats IntegratedStats, x, y float64) error {
	if !o.enabled || target == nil {
		return nil
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	// TODO: 实现字体渲染
	// 这里应该使用 Ebiten 的字体系统渲染文本

	return nil
}

// DebugCommand 调试命令
type DebugCommand struct {
	Name        string
	Description string
	Handler     func(args []string) error
}

// DebugCommandRegistry 调试命令注册表
type DebugCommandRegistry struct {
	mu       sync.Mutex
	commands map[string]DebugCommand
}

// NewDebugCommandRegistry 创建命令注册表
func NewDebugCommandRegistry() *DebugCommandRegistry {
	return &DebugCommandRegistry{
		commands: make(map[string]DebugCommand),
	}
}

// Register 注册命令
func (r *DebugCommandRegistry) Register(cmd DebugCommand) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd.Name] = cmd
}

// Execute 执行命令
func (r *DebugCommandRegistry) Execute(name string, args []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cmd, ok := r.commands[name]
	if !ok {
		return fmt.Errorf("unknown command: %s", name)
	}

	return cmd.Handler(args)
}

// ListCommands 列出所有命令
func (r *DebugCommandRegistry) ListCommands() []DebugCommand {
	r.mu.Lock()
	defer r.mu.Unlock()

	commands := make([]DebugCommand, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}

	return commands
}

// GlobalDebugger 全局调试器实例
var GlobalDebugger *RenderDebugger

// InitGlobalDebugger 初始化全局调试器
func InitGlobalDebugger(config DebugConfig) error {
	var err error
	GlobalDebugger, err = NewRenderDebugger(config)
	return err
}

// GetGlobalDebugger 获取全局调试器
func GetGlobalDebugger() *RenderDebugger {
	return GlobalDebugger
}

// StartDebug 启动调试
func StartDebug() {
	if GlobalDebugger != nil {
		GlobalDebugger.Start()
	}
}

// StopDebug 停止调试
func StopDebug() {
	if GlobalDebugger != nil {
		GlobalDebugger.Stop()
	}
}

// UpdateDebug 更新调试
func UpdateDebug(stats IntegratedStats) {
	if GlobalDebugger != nil {
		GlobalDebugger.Update(stats)
	}
}
