//go:build ebiten
// +build ebiten

package render

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestIntegratedRenderer 测试集成渲染器
func TestIntegratedRenderer(t *testing.T) {
	renderer := NewIntegratedRenderer(nil)

	// 测试 BeginFrame/EndFrame
	renderer.BeginFrame()
	renderer.EndFrame()

	// 测试 DrawImage
	img := ebiten.NewImage(100, 100)
	defer img.Dispose()

	geo := ebiten.GeoM{}
	geo.Translate(10, 10)

	opts := &ebiten.DrawImageOptions{
		GeoM: geo,
	}

	renderer.DrawImage(img, geo, opts)
	renderer.EndFrame()

	// 验证统计
	stats := renderer.GetStats()
	if stats.TotalDrawCalls == 0 {
		t.Error("DrawCall 计数应该大于 0")
	}

	// 测试临时图像缓存
	tempImg, err := renderer.GetTemporaryImage(64, 64)
	if err != nil {
		t.Fatalf("获取临时图像失败: %v", err)
	}

	renderer.ReleaseTemporaryImage(tempImg)

	// 测试绘制参数缓存
	drawParams := renderer.GetDrawParams(img, opts.ColorScale, opts.Blend, opts.Filter)
	if drawParams == nil {
		t.Error("绘制参数不应该为 nil")
	}

	t.Log("✓ 集成渲染器测试通过")
}

// TestRenderStatsManager 测试渲染统计管理器
func TestRenderStatsManager(t *testing.T) {
	manager := NewRenderStatsManager(100)

	// 开始帧统计
	manager.StartFrame(1)
	time.Sleep(10 * time.Millisecond)

	// 结束帧统计
	manager.EndFrame(100, 500, 10, 85.5)

	// 获取当前统计
	current := manager.GetCurrentStats()
	if current.FrameNumber != 1 {
		t.Errorf("帧号错误: 期望 1，实际 %d", current.FrameNumber)
	}

	if current.DrawCalls != 100 {
		t.Errorf("DrawCalls 错误: 期望 100，实际 %d", current.DrawCalls)
	}

	// 记录内存信息
	manager.RecordMemory(1024*1024, 1, 100, 50, 60)

	// 验证内存统计
	current = manager.GetCurrentStats()
	if current.MemoryAlloc != 1024*1024 {
		t.Errorf("内存分配错误: 期望 1024*1024，实际 %d", current.MemoryAlloc)
	}

	// 获取性能报告
	report := manager.GetPerformanceReport()
	if report.Average.DrawCalls != 100 {
		t.Errorf("平均 DrawCalls 错误: 期望 100，实际 %d", report.Average.DrawCalls)
	}

	t.Log("✓ 渲染统计管理器测试通过")
}

// TestPerformanceAnalyzer 测试性能分析器
func TestPerformanceAnalyzer(t *testing.T) {
	config := DefaultAnalysisConfig
	config.SampleInterval = 10 * time.Millisecond
	config.MinSamples = 10

	analyzer := NewPerformanceAnalyzer(config)

	// 添加样本
	for i := 0; i < 20; i++ {
		sample := PerformanceSample{
			Timestamp:    time.Now(),
			FrameTime:    10 * time.Millisecond * time.Duration(i+1),
			DrawCalls:    100 + i*10,
			Triangles:    500 + i*50,
			Batches:      10 + i,
			MemoryAlloc:  uint64(1024 * (i + 1)),
			CPUUsage:     50.0 + float64(i),
			GPUUsage:     60.0 + float64(i),
			CacheHitRate: 80.0 - float64(i),
			BatchEfficiency: 0.8,
		}
		analyzer.AddSample(sample)
	}

	// 执行分析
	result := analyzer.Analyze()

	// 验证分析结果
	if result.TotalSamples != 20 {
		t.Errorf("样本数错误: 期望 20，实际 %d", result.TotalSamples)
	}

	if result.FrameTimeStats.Mean <= 0 {
		t.Error("平均帧时间应该大于 0")
	}

	if result.DrawCallStats.Mean <= 0 {
		t.Error("平均 DrawCalls 应该大于 0")
	}

	if len(result.Hotspots) == 0 {
		t.Log("未检测到性能热点（正常情况）")
	}

	if len(result.Recommendations) == 0 {
		t.Log("未生成优化建议（正常情况）")
	}

	// 验证性能等级
	if result.Grade.OverallScore < 0 || result.Grade.OverallScore > 100 {
		t.Errorf("性能等级评分超出范围: %f", result.Grade.OverallScore)
	}

	t.Log("✓ 性能分析器测试通过")
}

// TestIntegratedRendererObjectPool 测试集成渲染器对象池
func TestIntegratedRendererObjectPool(t *testing.T) {
	// 从对象池获取渲染器
	r1 := GetIntegratedRenderer(nil)
	r2 := GetIntegratedRenderer(nil)

	if r1 == r2 {
		t.Error("从对象池获取的渲染器应该是不同的实例")
	}

	// 使用渲染器
	r1.BeginFrame()
	r1.EndFrame()

	// 返回对象池
	PutIntegratedRenderer(r1)
	PutIntegratedRenderer(r2)

	// 再次获取，应该能正常使用
	r3 := GetIntegratedRenderer(nil)
	if r3 == nil {
		t.Error("从对象池获取的渲染器不应该为 nil")
	}

	PutIntegratedRenderer(r3)

	t.Log("✓ 集成渲染器对象池测试通过")
}

// TestPhase3Optimization 测试 Phase 3 优化效果
func TestPhase3Optimization(t *testing.T) {
	renderer := NewIntegratedRenderer(nil)

	// 模拟渲染大量对象
	renderer.BeginFrame()

	target := ebiten.NewImage(1000, 1000)
	defer target.Dispose()

	img := ebiten.NewImage(64, 64)
	defer img.Dispose()

	// 渲染 1000 个对象
	for i := 0; i < 1000; i++ {
		geo := ebiten.GeoM{}
		geo.Translate(float64(i%100), float64(i/100))

		opts := &ebiten.DrawImageOptions{
			GeoM: geo,
		}

		renderer.DrawImage(img, geo, opts)

		// 每 10 个对象使用一次临时图像
		if i%10 == 0 {
			tempImg, _ := renderer.GetTemporaryImage(32, 32)
			renderer.ReleaseTemporaryImage(tempImg)
		}
	}

	start := time.Now()
	err := renderer.EndFrame()
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}

	// 验证统计
	stats := renderer.GetStats()
	t.Logf("渲染 1000 个对象用时: %v", elapsed)
	t.Logf("DrawCalls: %d, Batches: %d",
		stats.TotalDrawCalls, stats.BatchCount)

	// 性能分析
	analyzer := NewPerformanceAnalyzer(DefaultAnalysisConfig)
	sample := PerformanceSample{
		Timestamp:      time.Now(),
		FrameTime:      elapsed,
		DrawCalls:      int(stats.TotalDrawCalls),
		Triangles:      int(stats.TotalTriangles),
		Batches:        int(stats.BatchCount),
		CacheHitRate:   0, // 这里不计算缓存命中率
		BatchEfficiency: 0.8,
	}
	analyzer.AddSample(sample)
	result := analyzer.Analyze()

	t.Log("=== Phase 3 优化效果 ===")
	t.Logf("性能等级: %s (%.2f分)", result.Grade.Letters, result.Grade.OverallScore)
	t.Logf("平均帧时间: %.2fms", result.FrameTimeStats.Mean.Seconds()*1000)
	t.Logf("平均 DrawCalls: %.2f", result.DrawCallStats.Mean)

	// 验证优化效果
	if stats.TotalDrawCalls > 0 && stats.BatchCount > 0 {
		// 批处理应该有效
		t.Log("✓ Phase 3 优化测试通过")
	} else {
		t.Error("批处理渲染器未正常工作")
	}
}
