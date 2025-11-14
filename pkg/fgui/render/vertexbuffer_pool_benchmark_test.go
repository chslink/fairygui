package render

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// BenchmarkVertexBufferPool 测试对象池的性能提升
// 对比使用对象池 vs 直接分配的性能差异
func BenchmarkVertexBufferPool(b *testing.B) {
	b.Run("WithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// 直接分配顶点缓冲区
			vertices := make([]ebiten.Vertex, 256)
			indices := make([]uint16, 256)
			_ = vertices
			_ = indices
		}
	})

	b.Run("WithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// 使用对象池
			vb := GetVertexBuffer()
			_ = vb.Vertices
			_ = vb.Indices
			PutVertexBuffer(vb)
		}
	})

	b.Run("WithPool_Realistic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			vb := GetVertexBuffer()
			// 模拟真实的填充操作
			vb.Vertices = vb.Vertices[:100]
			for i := 0; i < 100; i++ {
				vb.Vertices[i] = ebiten.Vertex{
					DstX:   float32(i),
					DstY:   float32(i),
					SrcX:   float32(i),
					SrcY:   float32(i),
					ColorR: 1,
					ColorG: 1,
					ColorB: 1,
					ColorA: 1,
				}
			}
			vb.Indices = make([]uint16, 300)
			for i := 0; i < 300; i++ {
				vb.Indices[i] = uint16(i % 256)
			}
			PutVertexBuffer(vb)
		}
	})
}

// BenchmarkDrawParamsCache 测试 DrawParams 缓存的性能提升
func BenchmarkDrawParamsCache(b *testing.B) {
	atlas := NewAtlasManager(nil)

	b.Run("WithoutCache", func(b *testing.B) {
		img := ebiten.NewImage(100, 100)
		defer img.Dispose()

		for i := 0; i < b.N; i++ {
			// 每次都创建新的 DrawImageOptions
			opts := &ebiten.DrawImageOptions{
				ColorScale: ebiten.ColorScale{},
				Blend:      ebiten.Blend{},
				Filter:     ebiten.FilterNearest,
			}
			_ = opts
		}
	})

	b.Run("WithCache", func(b *testing.B) {
		img := ebiten.NewImage(100, 100)
		defer img.Dispose()

		for i := 0; i < b.N; i++ {
			// 使用缓存的 DrawParams
			params := atlas.GetDrawParams(img, ebiten.ColorScale{}, ebiten.Blend{}, ebiten.FilterNearest)
			_ = params
		}
	})

	b.Run("WithCache_DifferentImages", func(b *testing.B) {
		// 创建多个图像模拟真实场景
		images := make([]*ebiten.Image, 10)
		for i := range images {
			images[i] = ebiten.NewImage(100, 100)
			defer images[i].Dispose()
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 随机选择图像
			img := images[i%len(images)]
			params := atlas.GetDrawParams(img, ebiten.ColorScale{}, ebiten.Blend{}, ebiten.FilterNearest)
			_ = params
		}
	})
}

// BenchmarkMovieClipRendering 测试 MovieClip 渲染性能
func BenchmarkMovieClipRendering(b *testing.B) {
	// 创建模拟的测试数据（控制在对象池容量范围内）
	// 对象池初始化容量为 256，索引最多 256，所以顶点最多 86（(86-2)*3=252）
	points := make([]float64, 160) // 160/2 = 80 顶点
	for i := range points {
		points[i] = float64(i)
	}

	b.Run("WithObjectPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			vb := GetVertexBuffer()
			// 模拟 MovieClip 填充渲染
			vertexCount := len(points) / 2
			vb.Vertices = vb.Vertices[:vertexCount]
			for i := 0; i < vertexCount; i++ {
				vb.Vertices[i] = ebiten.Vertex{
					DstX:   float32(i),
					DstY:   float32(i),
					SrcX:   float32(i),
					SrcY:   float32(i),
					ColorR: 1,
					ColorG: 1,
					ColorB: 1,
					ColorA: 1,
				}
			}
			vb.Indices = vb.Indices[:(vertexCount-2)*3]
			PutVertexBuffer(vb)
		}
	})

	b.Run("WithoutObjectPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// 直接分配
			vertexCount := len(points) / 2
			vertices := make([]ebiten.Vertex, vertexCount)
			indices := make([]uint16, (vertexCount-2)*3)
			for i := 0; i < vertexCount; i++ {
				vertices[i] = ebiten.Vertex{
					DstX:   float32(i),
					DstY:   float32(i),
					SrcX:   float32(i),
					SrcY:   float32(i),
					ColorR: 1,
					ColorG: 1,
					ColorB: 1,
					ColorA: 1,
				}
			}
			_ = vertices
			_ = indices
		}
	})
}

// BenchmarkMemoryAllocation 内存分配测试
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("ObjectPool_HeapAlloc", func(b *testing.B) {
		var totalAlloc int64
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vb := GetVertexBuffer()
			vb.Vertices = vb.Vertices[:100]
			vb.Indices = make([]uint16, 300)
			// 模拟使用
			_ = len(vb.Vertices)
			_ = len(vb.Indices)
			PutVertexBuffer(vb)
			totalAlloc += int64(b.N) // 简化统计
		}
		_ = totalAlloc
	})

	b.Run("DirectAllocation_HeapAlloc", func(b *testing.B) {
		var totalAlloc int64
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			vertices := make([]ebiten.Vertex, 100)
			indices := make([]uint16, 300)
			_ = len(vertices)
			_ = len(indices)
			totalAlloc += int64(b.N)
		}
		_ = totalAlloc
	})
}
