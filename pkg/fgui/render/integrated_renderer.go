package render

import (
	"fmt"
	"sync"
	"time"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

// IntegratedRenderer 集成渲染器
// 整合所有优化组件，提供统一的渲染接口
type IntegratedRenderer struct {
	// 核心组件
	atlas          *AtlasManager
	batchRenderer  *BatchRenderer
	renderCtx      *RenderContext
	imageCache     *TemporaryImageCache
	textureManager *TextureManager

	// 统计信息
	stats IntegratedStats

	// 同步锁
	mu sync.Mutex
}

// IntegratedStats 集成渲染器统计
type IntegratedStats struct {
	TotalDrawCalls    int64
	TotalTriangles    int64
	BatchCount        int64
	CacheHits         int64
	CacheMisses       int64
	FrameCount        int64
	AvgFrameTime      float64
	LastFrameTime     float64
	LastFrameAt       time.Time
}

// NewIntegratedRenderer 创建集成渲染器
func NewIntegratedRenderer(loader assets.Loader) *IntegratedRenderer {
	return &IntegratedRenderer{
		atlas:          NewAtlasManager(loader),
		batchRenderer:  NewBatchRenderer(),
		renderCtx:      NewRenderContext(),
		imageCache:     NewTemporaryImageCache(),
		textureManager: NewTextureManager(DefaultTextureManagerConfig),
		stats: IntegratedStats{
			LastFrameAt: time.Now(),
		},
	}
}

// BeginFrame 开始渲染帧
func (r *IntegratedRenderer) BeginFrame() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 重置渲染上下文
	r.renderCtx.Begin()

	// 重置批处理器
	r.batchRenderer.batches = make(map[string][]BatchCommand)
	r.batchRenderer.standalone = r.batchRenderer.standalone[:0]

	// 重置统计
	r.stats.TotalDrawCalls = 0
	r.stats.TotalTriangles = 0
	r.stats.BatchCount = 0

	// 更新帧时间
	now := time.Now()
	if r.stats.FrameCount > 0 {
		elapsed := now.Sub(r.stats.LastFrameAt)
		// 指数移动平均
		if r.stats.AvgFrameTime == 0 {
			r.stats.AvgFrameTime = float64(elapsed.Nanoseconds())
		} else {
			r.stats.AvgFrameTime = r.stats.AvgFrameTime*0.9 + float64(elapsed.Nanoseconds())*0.1
		}
	}
	r.stats.LastFrameAt = now
}

// EndFrame 结束渲染帧
func (r *IntegratedRenderer) EndFrame() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 如果有命令需要执行，获取默认渲染目标
	if len(r.batchRenderer.batches) > 0 || len(r.batchRenderer.standalone) > 0 {
		target := r.atlas.atlasImages["default"]
		if target != nil {
			r.batchRenderer.Flush(target)
		}
	}

	// 清理临时图像缓存
	r.imageCache.Cleanup()

	// 更新统计
	batchStats := r.batchRenderer.GetStats()
	r.stats.BatchCount = int64(batchStats.BatchCount)

	clippingStats := r.imageCache.GetStats()
	r.stats.CacheHits = clippingStats.CacheHit
	r.stats.CacheMisses = clippingStats.CacheMiss

	r.stats.FrameCount++

	return nil
}

// DrawImage 绘制图像（使用批处理）
func (r *IntegratedRenderer) DrawImage(img *ebiten.Image, geo ebiten.GeoM, opts *ebiten.DrawImageOptions) {
	if img == nil || opts == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 应用渲染上下文状态
	r.renderCtx.ApplyToOptions(opts)

	// 添加到批处理器
	r.batchRenderer.AddCommand(img, geo, opts)

	// 更新统计
	r.stats.TotalDrawCalls++
}

// DrawTriangles 绘制三角形（使用批处理）
func (r *IntegratedRenderer) DrawTriangles(vertices []ebiten.Vertex, indices []uint16, texture *ebiten.Image) {
	if len(vertices) == 0 || len(indices) == 0 || texture == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 添加到批处理器的独立命令
	r.batchRenderer.AddTrianglesCommand(vertices, indices, texture)

	// 更新统计
	r.stats.TotalTriangles += int64(len(indices) / 3)
}

// GetTemporaryImage 获取临时图像（使用缓存）
func (r *IntegratedRenderer) GetTemporaryImage(width, height int) (*ebiten.Image, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	img, err := r.imageCache.GetOrCreate(width, height)
	if err != nil {
		return nil, err
	}

	// 更新统计
	clippingStats := r.imageCache.GetStats()
	_ = clippingStats

	return img, nil
}

// ReleaseTemporaryImage 释放临时图像
func (r *IntegratedRenderer) ReleaseTemporaryImage(img *ebiten.Image) {
	if img == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.imageCache.Release(img)
}

// GetDrawParams 获取绘制参数（使用缓存）
func (r *IntegratedRenderer) GetDrawParams(img *ebiten.Image, colorScale ebiten.ColorScale, blend ebiten.Blend, filter ebiten.Filter) *DrawParams {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.atlas.GetDrawParams(img, colorScale, blend, filter)
}

// GetAtlasManager 获取图集管理器
func (r *IntegratedRenderer) GetAtlasManager() *AtlasManager {
	return r.atlas
}

// GetRenderContext 获取渲染上下文
func (r *IntegratedRenderer) GetRenderContext() *RenderContext {
	return r.renderCtx
}

// GetStats 获取统计信息
func (r *IntegratedRenderer) GetStats() IntegratedStats {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.stats
}

// GetFPS 获取当前 FPS
func (r *IntegratedRenderer) GetFPS() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stats.AvgFrameTime == 0 {
		return 0
	}

	return 1000.0 / (r.stats.AvgFrameTime / 1000000.0)
}

// GetBatchStats 获取批处理统计
func (r *IntegratedRenderer) GetBatchStats() BatchStats {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.batchRenderer.GetStats()
}

// GetClippingStats 获取剪裁统计
func (r *IntegratedRenderer) GetClippingStats() ClippingStats {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.imageCache.GetStats()
}

// Cleanup 清理资源
// IntegratedRendererPool 集成渲染器对象池
var IntegratedRendererPool = sync.Pool{
	New: func() interface{} {
		return NewIntegratedRenderer(nil)
	},
}

// GetIntegratedRenderer 从对象池获取
func GetIntegratedRenderer(loader assets.Loader) *IntegratedRenderer {
	r := IntegratedRendererPool.Get().(*IntegratedRenderer)
	if loader != nil {
		r.atlas.loader = loader
	}
	return r
}

// PutIntegratedRenderer 返回对象池
func PutIntegratedRenderer(r *IntegratedRenderer) {
	if r == nil {
		return
	}
	r.Cleanup()
	IntegratedRendererPool.Put(r)
}

// AcquireTexture 获取 GPU 纹理
func (r *IntegratedRenderer) AcquireTexture(key string, width, height int, format TextureFormat) (*GPUTexture, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.textureManager == nil {
		return nil, fmt.Errorf("texture manager not initialized")
	}

	return r.textureManager.Acquire(key, width, height, format)
}

// ReleaseTexture 释放 GPU 纹理
func (r *IntegratedRenderer) ReleaseTexture(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.textureManager == nil {
		return fmt.Errorf("texture manager not initialized")
	}

	return r.textureManager.Release(key)
}

// GetTextureManager 获取纹理管理器
func (r *IntegratedRenderer) GetTextureManager() *TextureManager {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.textureManager
}

// Cleanup 清理资源（更新以包含纹理管理器）
func (r *IntegratedRenderer) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 清理缓存
	if r.imageCache != nil {
		r.imageCache.Cleanup()
	}

	// 关闭纹理管理器
	if r.textureManager != nil {
		r.textureManager.Close()
	}

	// 重置状态
	if r.renderCtx != nil {
		r.renderCtx.End()
	}
}
