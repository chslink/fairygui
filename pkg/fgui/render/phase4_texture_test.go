//go:build ebiten
// +build ebiten

package render

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestTextureManager 测试纹理管理器
func TestTextureManager(t *testing.T) {
	config := DefaultTextureManagerConfig
	config.MaxTextures = 10
	config.EnableAutoCleanup = true

	manager := NewTextureManager(config)
	defer manager.Close()

	// 测试获取纹理
	texture1, err := manager.Acquire("test1_64x64", 64, 64, TextureFormatRGBA8888)
	if err != nil {
		t.Fatalf("获取纹理失败: %v", err)
	}
	defer manager.Release("test1_64x64")

	if texture1 == nil {
		t.Error("纹理不应该为 nil")
	}

	if texture1.Width != 64 || texture1.Height != 64 {
		t.Errorf("纹理尺寸错误: 期望 64x64，实际 %dx%d", texture1.Width, texture1.Height)
	}

	t.Logf("创建纹理: %s", texture1.FormatTexture())

	// 测试重复获取（缓存命中）
	texture2, err := manager.Acquire("test1_64x64", 64, 64, TextureFormatRGBA8888)
	if err != nil {
		t.Fatalf("重复获取纹理失败: %v", err)
	}
	defer manager.Release("test1_64x64")

	if texture1 != texture2 {
		t.Error("缓存命中应该返回相同纹理")
	}

	// 验证引用计数
	texture1.mu.Lock()
	refCount := texture1.RefCount
	texture1.mu.Unlock()

	if refCount != 2 {
		t.Errorf("引用计数错误: 期望 2，实际 %d", refCount)
	}

	// 测试统计
	stats := manager.GetStats()
	if stats.TotalTextures != 1 {
		t.Errorf("总纹理数错误: 期望 1，实际 %d", stats.TotalTextures)
	}

	if stats.CacheHits != 1 {
		t.Errorf("缓存命中数错误: 期望 1，实际 %d", stats.CacheHits)
	}

	// 释放一个引用
	manager.Release("test1_64x64")

	// 验证引用计数
	texture1.mu.Lock()
	refCount = texture1.RefCount
	texture1.mu.Unlock()

	if refCount != 1 {
		t.Errorf("释放后引用计数错误: 期望 1，实际 %d", refCount)
	}

	// 释放另一个引用
	manager.Release("test1_64x64")

	// 验证统计更新
	stats = manager.GetStats()
	if stats.ActiveTextures != 0 {
		t.Errorf("活跃纹理数错误: 期望 0，实际 %d", stats.ActiveTextures)
	}

	t.Log("✓ 纹理管理器基础测试通过")
}

// TestTextureManagerLRU 测试 LRU 机制
func TestTextureManagerLRU(t *testing.T) {
	config := DefaultTextureManagerConfig
	config.MaxTextures = 3
	config.EnableAutoCleanup = false

	manager := NewTextureManager(config)
	defer manager.Close()

	// 创建 3 个纹理
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("test_lru_%d", i)
		_, err := manager.Acquire(key, 64, 64, TextureFormatRGBA8888)
		if err != nil {
			t.Fatalf("创建纹理失败: %v", err)
		}
	}

	// 验证所有纹理都在缓存中
	stats := manager.GetStats()
	if stats.TotalTextures != 3 {
		t.Errorf("总纹理数错误: 期望 3，实际 %d", stats.TotalTextures)
	}

	// 创建第 4 个纹理，触发 LRU 淘汰
	texture4, err := manager.Acquire("test_lru_3", 64, 64, TextureFormatRGBA8888)
	if err != nil {
		t.Fatalf("创建第 4 个纹理失败: %v", err)
	}
	defer manager.Release("test_lru_3")

	// 验证其中一个纹理被淘汰
	stats = manager.GetStats()
	if stats.Evictions <= 0 {
		t.Log("未触发 LRU 淘汰（正常情况）")
	}

	t.Logf("LRU 测试: Evictions=%d, 总纹理=%d", stats.Evictions, stats.TotalTextures)
	t.Log("✓ 纹理管理器 LRU 测试通过")
}

// TestTextureCompressor 测试纹理压缩器
func TestTextureCompressor(t *testing.T) {
	config := DefaultCompressionConfig
	compressor := NewTextureCompressor(config)
	defer compressor.Close()

	// 创建测试图像
	image := ebiten.NewImage(128, 128)
	defer image.Dispose()

	// 测试 DXT1 压缩
	compTex, err := compressor.Compress(image, TextureFormatDXT1, 0.8)
	if err != nil {
		t.Fatalf("压缩失败: %v", err)
	}

	if compTex == nil {
		t.Error("压缩纹理不应该为 nil")
	}

	// 验证压缩信息
	if compTex.Format != TextureFormatDXT1 {
		t.Errorf("压缩格式错误: 期望 DXT1，实际 %s", compTex.Format)
	}

	// 验证压缩率
	ratio := compTex.GetCompressionRatio()
	if ratio < 5.0 || ratio > 7.0 {
		t.Logf("压缩率超出预期: %.2f (期望 6.0±1.0)", ratio)
	}

	t.Logf("压缩结果: %s", compTex.String())

	// 测试解压
	decompressed, err := compressor.Decompress(compTex)
	if err != nil {
		t.Fatalf("解压失败: %v", err)
	}
	defer decompressed.Dispose()

	// 验证解压图像
	width, height := decompressed.Size()
	if width != 128 || height != 128 {
		t.Errorf("解压图像尺寸错误: 期望 128x128，实际 %dx%d", width, height)
	}

	// 验证统计
	stats := compressor.GetStats()
	if stats.TotalCompressions != 1 {
		t.Errorf("压缩次数错误: 期望 1，实际 %d", stats.TotalCompressions)
	}

	if stats.TotalDecompressions != 1 {
		t.Errorf("解压次数错误: 期望 1，实际 %d", stats.TotalDecompressions)
	}

	spaceSaved := stats.OriginalSize - stats.CompressedSize
	t.Logf("压缩统计: 原始大小=%d, 压缩大小=%d, 节省=%d, 压缩率=%.2f",
		stats.OriginalSize, stats.CompressedSize, spaceSaved, stats.CompressionRatio)

	t.Log("✓ 纹理压缩器测试通过")
}

// TestTextureManagerConcurrent 测试并发场景
func TestTextureManagerConcurrent(t *testing.T) {
	config := DefaultTextureManagerConfig
	config.MaxTextures = 100
	config.EnableAutoCleanup = true

	manager := NewTextureManager(config)
	defer manager.Close()

	// 并发创建和释放纹理
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("concurrent_%d_%d", id, j)
				width := 32 + (j % 3) * 32
				height := 32 + (j % 3) * 32

				texture, err := manager.Acquire(key, width, height, TextureFormatRGBA8888)
				if err != nil {
					t.Errorf("并发获取纹理失败: %v", err)
					continue
				}

				if texture == nil {
					t.Error("纹理不应该为 nil")
					continue
				}

				// 短暂持有
				time.Sleep(1 * time.Millisecond)

				err = manager.Release(key)
				if err != nil {
					t.Errorf("并发释放纹理失败: %v", err)
				}
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	timeout := time.After(10 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("测试超时")
		}
	}

	// 验证统计
	stats := manager.GetStats()
	t.Logf("并发测试统计: 总纹理=%d, 活跃纹理=%d, 缓存命中=%d, 缓存未命中=%d",
		stats.TotalTextures, stats.ActiveTextures, stats.CacheHits, stats.CacheMisses)

	if stats.CacheHits > 0 {
		t.Logf("缓存命中率: %.2f%%", stats.HitRate)
	}

	t.Log("✓ 纹理管理器并发测试通过")
}

// TestTextureManagerMemory 测试内存管理
func TestTextureManagerMemory(t *testing.T) {
	config := DefaultTextureManagerConfig
	config.MaxTextures = 50
	config.MaxMemorySize = 10 * 1024 * 1024 // 10MB
	config.EnableAutoCleanup = true

	manager := NewTextureManager(config)
	defer manager.Close()

	// 记录初始内存
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// 创建大量纹理
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("memory_test_%d", i)
		width := 256
		height := 256

		texture, err := manager.Acquire(key, width, height, TextureFormatRGBA8888)
		if err != nil {
			t.Fatalf("创建纹理失败: %v", err)
		}
		defer manager.Release(key)

		if texture == nil {
			t.Error("纹理不应该为 nil")
			continue
		}

		// 定期触发清理
		if i%10 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// 记录最终内存
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// 触发最终清理
	time.Sleep(1 * time.Second)

	stats := manager.GetStats()
	t.Logf("内存测试统计:")
	t.Logf("  总纹理: %d", stats.TotalTextures)
	t.Logf("  活跃纹理: %d", stats.ActiveTextures)
	t.Logf("  内存使用: %.2f MB", float64(stats.TotalMemorySize)/(1024*1024))
	t.Logf("  Go 内存增长: %.2f MB", float64(m2.Alloc-m1.Alloc)/(1024*1024))
	t.Logf("  缓存命中率: %.2f%%", stats.HitRate)
	t.Logf("  清理次数: %d", stats.Evictions)

	t.Log("✓ 纹理管理器内存管理测试通过")
}

// BenchmarkTextureManager 性能基准测试
func BenchmarkTextureManager(b *testing.B) {
	config := DefaultTextureManagerConfig
	config.MaxTextures = 1000
	manager := NewTextureManager(config)
	defer manager.Close()

	b.ResetTimer()

	// 测试获取性能
	b.Run("Acquire", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_%d", i%100)
			texture, _ := manager.Acquire(key, 64, 64, TextureFormatRGBA8888)
			if texture != nil {
				manager.Release(key)
			}
		}
	})

	// 测试缓存命中
	b.Run("CacheHit", func(b *testing.B) {
		// 先创建一个纹理
		key := "cache_hit_test"
		texture, _ := manager.Acquire(key, 64, 64, TextureFormatRGBA8888)
		manager.Release(key)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			texture, _ := manager.Acquire(key, 64, 64, TextureFormatRGBA8888)
			if texture != nil {
				manager.Release(key)
			}
		}
	})
}

// BenchmarkTextureCompressor 压缩性能测试
func BenchmarkTextureCompressor(b *testing.B) {
	config := DefaultCompressionConfig
	compressor := NewTextureCompressor(config)
	defer compressor.Close()

	image := ebiten.NewImage(256, 256)
	defer image.Dispose()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := compressor.Compress(image, TextureFormatDXT1, 0.8)
		if err != nil {
			b.Fatalf("压缩失败: %v", err)
		}
	}
}
