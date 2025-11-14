package render

import (
	"fmt"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// TextureCompressor 纹理压缩器
// 提供多种压缩算法，优化 GPU 内存使用
type TextureCompressor struct {
	pool     *sync.Pool
	config   CompressionConfig
	stats    CompressionStats
}

// CompressionConfig 压缩配置
type CompressionConfig struct {
	EnableFastCompression   bool          // 启用快速压缩
	EnableQualityCompression bool        // 启用高质量压缩
	MaxCompressionTime      time.Duration // 最大压缩时间
	QualityLevel            float64       // 质量级别 (0-1)
}

// DefaultCompressionConfig 默认压缩配置
var DefaultCompressionConfig = CompressionConfig{
	EnableFastCompression:   true,
	EnableQualityCompression: false,
	MaxCompressionTime:      100 * time.Millisecond,
	QualityLevel:            0.8,
}

// CompressionStats 压缩统计
type CompressionStats struct {
	TotalCompressions    int64         // 总压缩次数
	TotalDecompressions  int64         // 总解压次数
	TotalCompressedSize  int64         // 压缩后总大小
	TotalOriginalSize    int64         // 压缩前总大小
	CompressionRatio     float64       // 压缩率
	AverageCompressionTime time.Duration // 平均压缩时间
	SpaceSaved           int64         // 节省的空间
}

// NewTextureCompressor 创建纹理压缩器
func NewTextureCompressor(config CompressionConfig) *TextureCompressor {
	return &TextureCompressor{
		pool: &sync.Pool{
			New: func() interface{} {
				return &CompressionJob{
					ID:        0,
					Source:    nil,
					Target:    nil,
					Format:    TextureFormatRGBA8888,
					Quality:   0.8,
					CreatedAt: time.Now(),
				}
			},
		},
		config: config,
		stats: CompressionStats{},
	}
}

// Compress 压缩纹理
// 注意：这是模拟实现，实际压缩需要 GPU API 或外部库
func (tc *TextureCompressor) Compress(image *ebiten.Image, format TextureFormat, quality float64) (*CompressedTexture, error) {
	if image == nil {
		return nil, fmt.Errorf("image is nil")
	}

	startTime := time.Now()

	// 模拟压缩过程
	// 在真实实现中，这里会调用 GPU 压缩 API 或算法库
	width, height := image.Size()
	originalSize := int64(width * height * 4) // RGBA8888

	// 模拟压缩时间
	compressionDelay := 10*time.Millisecond + time.Duration(quality*50)*time.Millisecond
	time.Sleep(compressionDelay)

	compressed := &CompressedTexture{
		Format:      format,
		Width:       width,
		Height:      height,
		OriginalSize: originalSize,
		Compressed:  true,
		Quality:     quality,
		CreatedAt:   time.Now(),
	}

	// 计算压缩后大小（模拟）
	switch format {
	case TextureFormatDXT1:
		compressed.CompressedSize = originalSize / 6 // 6:1 压缩
	case TextureFormatDXT5:
		compressed.CompressedSize = originalSize / 4 // 4:1 压缩
	case TextureFormatETC2:
		compressed.CompressedSize = originalSize / 4
	case TextureFormatASTC:
		compressed.CompressedSize = originalSize / 8 // 8:1 压缩
	case TextureFormatRGBA8888:
		compressed.CompressedSize = originalSize
	case TextureFormatRGB888:
		compressed.CompressedSize = (originalSize / 4) * 3
	default:
		compressed.CompressedSize = originalSize / 2
	}

	// 更新统计
	elapsed := time.Since(startTime)
	tc.stats.TotalCompressions++
	tc.stats.TotalCompressedSize += compressed.CompressedSize
	tc.stats.TotalOriginalSize += originalSize
	tc.stats.AverageCompressionTime = time.Duration(
		(int64(tc.stats.AverageCompressionTime)*(tc.stats.TotalCompressions-1) + int64(elapsed)) / tc.stats.TotalCompressions)
	tc.stats.SpaceSaved += originalSize - compressed.CompressedSize

	if tc.stats.TotalOriginalSize > 0 {
		tc.stats.CompressionRatio = float64(tc.stats.TotalCompressedSize) / float64(tc.stats.TotalOriginalSize)
	}

	return compressed, nil
}

// Decompress 解压纹理
func (tc *TextureCompressor) Decompress(compTex *CompressedTexture) (*ebiten.Image, error) {
	if compTex == nil {
		return nil, fmt.Errorf("compressed texture is nil")
	}

	startTime := time.Now()

	// 模拟解压过程
	time.Sleep(5 * time.Millisecond) // 解压通常比压缩快

	// 创建解压后的图像
	image := ebiten.NewImage(compTex.Width, compTex.Height)

	// 更新统计
	elapsed := time.Since(startTime)
	tc.stats.TotalDecompressions++
	tc.stats.AverageCompressionTime = time.Duration(
		(int64(tc.stats.AverageCompressionTime)*(tc.stats.TotalDecompressions-1) + int64(elapsed)) / tc.stats.TotalDecompressions)

	return image, nil
}

// GetStats 获取压缩统计
func (tc *TextureCompressor) GetStats() CompressionStats {
	return tc.stats
}

// Close 关闭压缩器
func (tc *TextureCompressor) Close() {
	// 纹理压缩器没有需要关闭的资源
	// 此方法为未来扩展预留
}

// CompressedTexture 压缩纹理
type CompressedTexture struct {
	Format         TextureFormat
	Width, Height  int
	OriginalSize   int64
	CompressedSize int64
	Compressed     bool
	Quality        float64
	CreatedAt      time.Time
}

// String 返回压缩信息
func (ct *CompressedTexture) String() string {
	ratio := float64(ct.OriginalSize) / float64(ct.CompressedSize)
	return fmt.Sprintf("CompressedTexture(%s, %dx%d, %.2fx compression, Quality=%.2f)",
		ct.Format, ct.Width, ct.Height, ratio, ct.Quality)
}

// GetCompressionRatio 获取压缩率
func (ct *CompressedTexture) GetCompressionRatio() float64 {
	if ct.CompressedSize == 0 {
		return 0
	}
	return float64(ct.OriginalSize) / float64(ct.CompressedSize)
}

// CompressionJob 压缩任务
type CompressionJob struct {
	ID        int64
	Source    *ebiten.Image
	Target    *CompressedTexture
	Format    TextureFormat
	Quality   float64
	CreatedAt time.Time
	mu        sync.Mutex
	Result    <-chan CompressionResult
	Done      <-chan struct{}
}

// CompressionResult 压缩结果
type CompressionResult struct {
	Texture *CompressedTexture
	Error   error
}

// TextureLoader 异步纹理加载器
type TextureLoader struct {
	manager    *TextureManager
	compressor *TextureCompressor
	jobQueue   chan LoadJob
	workerPool chan chan LoadJob
	quit       chan bool
	numWorkers int
}

// LoadJob 加载任务
type LoadJob struct {
	Key        string
	URL        string
	Width      int
	Height     int
	Format     TextureFormat
	Callback   func(*GPUTexture, error)
}

// NewTextureLoader 创建纹理加载器
func NewTextureLoader(manager *TextureManager, compressor *TextureCompressor, numWorkers int) *TextureLoader {
	if numWorkers <= 0 {
		numWorkers = 4
	}

	return &TextureLoader{
		manager:    manager,
		compressor: compressor,
		jobQueue:   make(chan LoadJob, 100),
		workerPool: make(chan chan LoadJob, numWorkers),
		quit:       make(chan bool),
		numWorkers: numWorkers,
	}
}

// Start 启动加载器
func (tl *TextureLoader) Start() {
	// 启动工作线程
	for i := 0; i < tl.numWorkers; i++ {
		workerChan := make(chan LoadJob, 10)
		tl.workerPool <- workerChan
		go tl.worker(workerChan)
	}
}

// worker 工作线程
func (tl *TextureLoader) worker(jobChan chan LoadJob) {
	for {
		select {
		case job := <-jobChan:
			tl.processJob(job)
		case <-tl.quit:
			return
		}
	}
}

// processJob 处理任务
func (tl *TextureLoader) processJob(job LoadJob) {
	// 生成缓存键
	key := fmt.Sprintf("%s_%dx%d_%s", job.URL, job.Width, job.Height, job.Format)

	// 尝试从管理器获取
	texture, err := tl.manager.Acquire(key, job.Width, job.Height, job.Format)
	if err != nil {
		if job.Callback != nil {
			job.Callback(nil, err)
		}
		return
	}

	if texture != nil {
		if job.Callback != nil {
			job.Callback(texture, nil)
		}
		return
	}

	// 如果压缩启用，尝试压缩
	if tl.compressor != nil {
		compTex, err := tl.compressor.Compress(nil, job.Format, 0.8)
		if err == nil && compTex != nil {
			// 模拟返回压缩纹理
			if job.Callback != nil {
				job.Callback(texture, nil)
			}
			return
		}
	}

	// 如果都无法处理，返回错误
	if job.Callback != nil {
		job.Callback(nil, fmt.Errorf("failed to load texture: %s", job.URL))
	}
}

// LoadAsync 异步加载纹理
func (tl *TextureLoader) LoadAsync(job LoadJob) {
	select {
	case tl.jobQueue <- job:
	default:
		// 队列满，丢弃任务
		if job.Callback != nil {
			job.Callback(nil, fmt.Errorf("job queue is full"))
		}
	}
}

// Close 关闭加载器
func (tl *TextureLoader) Close() {
	close(tl.quit)
}

// GetStats 获取加载器统计
func (tl *TextureLoader) GetStats() TextureLoaderStats {
	return TextureLoaderStats{
		QueueLength:   len(tl.jobQueue),
		ActiveWorkers: tl.numWorkers,
	}
}

// TextureLoaderStats 加载器统计
type TextureLoaderStats struct {
	QueueLength   int
	ActiveWorkers int
}
