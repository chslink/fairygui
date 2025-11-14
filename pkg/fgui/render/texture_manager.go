package render

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// GPUTexture GPU 纹理对象
// 封装 Ebiten 图像，提供引用计数和生命周期管理
type GPUTexture struct {
	// 核心属性
	Image     *ebiten.Image // Ebiten 图像对象
	Width     int           // 纹理宽度
	Height    int           // 纹理高度
	Format    TextureFormat // 纹理格式
	Key       string        // 缓存键

	// 生命周期管理
	RefCount     int64         // 引用计数
	LastAccess   time.Time     // 最后访问时间
	CreatedAt    time.Time     // 创建时间
	IsCompressed bool          // 是否压缩

	// 统计信息
	AccessCount   int64 // 访问次数
	MemorySize    int64 // 内存大小（字节）

	// 同步
	mu sync.Mutex
}

// TextureFormat 纹理格式
type TextureFormat int

const (
	TextureFormatRGBA8888 TextureFormat = iota
	TextureFormatRGBA4444
	TextureFormatRGB888
	TextureFormatRGB565
	TextureFormatRGBA16F   // 浮点格式
	TextureFormatDXT1      // 压缩格式
	TextureFormatDXT5      // 压缩格式
	TextureFormatETC2      // 移动端压缩格式
	TextureFormatASTC      // 现代压缩格式
)

// String 返回格式名称
func (f TextureFormat) String() string {
	switch f {
	case TextureFormatRGBA8888:
		return "RGBA8888"
	case TextureFormatRGBA4444:
		return "RGBA4444"
	case TextureFormatRGB888:
		return "RGB888"
	case TextureFormatRGB565:
		return "RGB565"
	case TextureFormatRGBA16F:
		return "RGBA16F"
	case TextureFormatDXT1:
		return "DXT1"
	case TextureFormatDXT5:
		return "DXT5"
	case TextureFormatETC2:
		return "ETC2"
	case TextureFormatASTC:
		return "ASTC"
	default:
		return "Unknown"
	}
}

// CalculateMemorySize 计算内存大小
func (t *GPUTexture) CalculateMemorySize() int64 {
	if t.MemorySize > 0 {
		return t.MemorySize
	}

	// 基于格式计算内存大小
	bytesPerPixel := 4 // 默认 RGBA8888
	switch t.Format {
	case TextureFormatRGBA8888:
		bytesPerPixel = 4
	case TextureFormatRGBA4444, TextureFormatRGB565:
		bytesPerPixel = 2
	case TextureFormatRGB888:
		bytesPerPixel = 3
	case TextureFormatRGBA16F:
		bytesPerPixel = 8
	case TextureFormatDXT1:
		bytesPerPixel = 4 // 6:1 压缩率
	case TextureFormatDXT5:
		bytesPerPixel = 8 // 4:1 压缩率
	case TextureFormatETC2:
		bytesPerPixel = 4
	case TextureFormatASTC:
		bytesPerPixel = 2 // 高压缩率
	}

	t.MemorySize = int64(t.Width) * int64(t.Height) * int64(bytesPerPixel)
	return t.MemorySize
}

// TextureManager 纹理管理器
// 提供 GPU 纹理的创建、获取、释放和生命周期管理
type TextureManager struct {
	// 核心组件
	textures map[string]*GPUTexture // 纹理缓存
	lruList  *list.List             // LRU 链表
	lruMap   map[*GPUTexture]*list.Element // LRU 映射
	pool     *sync.Pool            // 对象池

	// 配置
	config   TextureManagerConfig
	mu       sync.RWMutex

	// 统计
	stats TextureManagerStats

	// 自动清理
	ticker   *time.Ticker
	quit     chan bool
}

// TextureManagerConfig 纹理管理器配置
type TextureManagerConfig struct {
	MaxTextures        int           // 最大纹理数
	MaxMemorySize      int64         // 最大内存（字节）
	MaxTextureSize     int           // 最大纹理尺寸
	LRUCleanupInterval time.Duration // LRU 清理间隔
	EnableCompression  bool          // 启用压缩
	EnableAutoCleanup  bool          // 启用自动清理
	CompressionFormat  TextureFormat // 压缩格式
}

// DefaultTextureManagerConfig 默认配置
var DefaultTextureManagerConfig = TextureManagerConfig{
	MaxTextures:         1024,
	MaxMemorySize:      512 * 1024 * 1024, // 512MB
	MaxTextureSize:     4096,              // 4K
	LRUCleanupInterval: 5 * time.Second,   // 5秒
	EnableCompression:  false,
	EnableAutoCleanup:  true,
	CompressionFormat:  TextureFormatDXT1,
}

// TextureManagerStats 纹理管理器统计
type TextureManagerStats struct {
	TotalTextures      int64         // 总纹理数
	ActiveTextures     int64         // 活跃纹理数
	TotalMemorySize    int64         // 总内存大小
	CacheHits          int64         // 缓存命中
	CacheMisses        int64         // 缓存未命中
	Compressions       int64         // 压缩次数
	Decompressions     int64         // 解压次数
	Evictions          int64         // 淘汰次数
	HitRate            float64       // 命中率
	LastCleanupAt      time.Time     // 上次清理时间
}

// NewTextureManager 创建纹理管理器
func NewTextureManager(config TextureManagerConfig) *TextureManager {
	if config.MaxTextures <= 0 {
		config.MaxTextures = DefaultTextureManagerConfig.MaxTextures
	}
	if config.MaxMemorySize <= 0 {
		config.MaxMemorySize = DefaultTextureManagerConfig.MaxMemorySize
	}

	tm := &TextureManager{
		textures: make(map[string]*GPUTexture),
		lruList:  list.New(),
		lruMap:   make(map[*GPUTexture]*list.Element),
		config:   config,
		stats: TextureManagerStats{
			LastCleanupAt: time.Now(),
		},
		quit: make(chan bool),
	}

	// 创建对象池
	tm.pool = &sync.Pool{
		New: func() interface{} {
			return &GPUTexture{
				RefCount:   0,
				CreatedAt:  time.Now(),
				LastAccess: time.Now(),
			}
		},
	}

	// 启动自动清理
	if config.EnableAutoCleanup {
		tm.startAutoCleanup()
	}

	return tm
}

// Acquire 获取纹理
func (tm *TextureManager) Acquire(key string, width, height int, format TextureFormat) (*GPUTexture, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 查找缓存
	if texture, ok := tm.textures[key]; ok {
		// 缓存命中
		texture.mu.Lock()
		texture.RefCount++
		texture.LastAccess = time.Now()
		texture.AccessCount++
		texture.mu.Unlock()

		tm.stats.CacheHits++
		tm.stats.ActiveTextures++
		return texture, nil
	}

	// 缓存未命中，创建新纹理
	tm.stats.CacheMisses++

	// 检查是否需要创建
	if len(tm.textures) >= tm.config.MaxTextures {
		// 尝试清理 LRU
		tm.evictLRU(1)
	}

	// 创建新纹理
	texture := tm.createTexture(key, width, height, format)
	if texture == nil {
		return nil, fmt.Errorf("failed to create texture: %s", key)
	}

	// 添加到缓存
	tm.textures[key] = texture
	tm.lruList.PushBack(texture)

	tm.stats.TotalTextures++
	tm.stats.ActiveTextures++

	return texture, nil
}

// Release 释放纹理
func (tm *TextureManager) Release(key string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	texture, ok := tm.textures[key]
	if !ok {
		return fmt.Errorf("texture not found: %s", key)
	}

	texture.mu.Lock()
	texture.RefCount--
	if texture.RefCount < 0 {
		texture.RefCount = 0
	}
	texture.mu.Unlock()

	tm.stats.ActiveTextures--

	// 如果引用计数为0，可以考虑释放或保留
	if texture.RefCount == 0 && tm.shouldEvict(texture) {
		tm.removeTexture(key)
	}

	return nil
}

// createTexture 创建纹理
func (tm *TextureManager) createTexture(key string, width, height int, format TextureFormat) *GPUTexture {
	// 检查尺寸限制
	if width > tm.config.MaxTextureSize || height > tm.config.MaxTextureSize {
		return nil
	}

	// 检查内存限制
	estimatedSize := int64(width) * int64(height) * 4 // 估算 RGBA
	if tm.stats.TotalMemorySize+estimatedSize > tm.config.MaxMemorySize {
		// 尝试清理内存
		tm.evictLRU(0)
	}

	// 创建 Ebiten 图像
	image := ebiten.NewImage(width, height)
	if image == nil {
		return nil
	}

	// 获取纹理对象
	texture := tm.pool.Get().(*GPUTexture)
	*texture = GPUTexture{
		Image:       image,
		Width:       width,
		Height:      height,
		Format:      format,
		Key:         key,
		RefCount:    1,
		LastAccess:  time.Now(),
		CreatedAt:   time.Now(),
		IsCompressed: tm.config.EnableCompression && format != TextureFormatRGBA8888,
		AccessCount: 1,
		mu:          sync.Mutex{},
	}

	// 计算内存大小
	texture.CalculateMemorySize()
	tm.stats.TotalMemorySize += texture.MemorySize

	return texture
}

// removeTexture 移除纹理
func (tm *TextureManager) removeTexture(key string) {
	texture, ok := tm.textures[key]
	if !ok {
		return
	}

	// 从 LRU 列表移除
	if element, ok := tm.lruMap[texture]; ok {
		tm.lruList.Remove(element)
		delete(tm.lruMap, texture)
	}

	// 释放 Ebiten 图像
	if texture.Image != nil {
		texture.Image.Dispose()
	}

	// 减少内存统计
	tm.stats.TotalMemorySize -= texture.MemorySize
	tm.stats.TotalTextures--
	tm.stats.ActiveTextures--

	// 从缓存移除
	delete(tm.textures, key)

	// 返回对象池
	tm.pool.Put(texture)
}

// evictLRU 淘汰 LRU 纹理
func (tm *TextureManager) evictLRU(count int) int {
	evicted := 0
	for {
		if evicted >= count && count > 0 {
			break
		}

		// 获取最久未访问的纹理
		element := tm.lruList.Front()
		if element == nil {
			break
		}

		texture := element.Value.(*GPUTexture)

		// 检查是否可以淘汰
		if texture.RefCount == 0 {
			tm.removeTexture(texture.Key)
			evicted++
			tm.stats.Evictions++
		} else {
			// 移动到后面（避免淘汰正在使用的）
			tm.lruList.MoveToBack(element)
			if count == 0 {
				// 只淘汰一个，返回
				break
			}
		}

		// 如果没有可淘汰的纹理，退出
		if tm.lruList.Len() == 0 {
			break
		}

		// 检查内存是否充足
		if tm.stats.TotalMemorySize < tm.config.MaxMemorySize/2 {
			break
		}
	}

	return evicted
}

// shouldEvict 判断是否应该淘汰
func (tm *TextureManager) shouldEvict(texture *GPUTexture) bool {
	// 检查内存限制
	if tm.stats.TotalMemorySize < tm.config.MaxMemorySize {
		return false
	}

	// 检查访问时间
	age := time.Since(texture.LastAccess)
	if age < 10*time.Second {
		return false
	}

	// 检查访问频率
	if texture.AccessCount > 100 {
		return false
	}

	return true
}

// startAutoCleanup 启动自动清理
func (tm *TextureManager) startAutoCleanup() {
	tm.ticker = time.NewTicker(tm.config.LRUCleanupInterval)
	go func() {
		for {
			select {
			case <-tm.ticker.C:
				tm.performCleanup()
			case <-tm.quit:
				tm.ticker.Stop()
				return
			}
		}
	}()
}

// performCleanup 执行清理
func (tm *TextureManager) performCleanup() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	evicted := tm.evictLRU(10)
	tm.stats.LastCleanupAt = time.Now()

	// 更新命中率
	if tm.stats.CacheHits+tm.stats.CacheMisses > 0 {
		tm.stats.HitRate = float64(tm.stats.CacheHits) / float64(tm.stats.CacheHits+tm.stats.CacheMisses) * 100
	}

	if evicted > 0 {
		fmt.Printf("TextureManager: Cleaned up %d textures\n", evicted)
	}
}

// GetStats 获取统计信息
func (tm *TextureManager) GetStats() TextureManagerStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	stats := tm.stats
	if stats.CacheHits+stats.CacheMisses > 0 {
		stats.HitRate = float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses) * 100
	}

	return stats
}

// Close 关闭纹理管理器
func (tm *TextureManager) Close() {
	if tm.ticker != nil {
		tm.quit <- true
		tm.ticker.Stop()
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 清理所有纹理
	for key := range tm.textures {
		tm.removeTexture(key)
	}
}

// FormatTexture 格式化纹理信息
func (t *GPUTexture) FormatTexture() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	return fmt.Sprintf("Texture(%s: %dx%d, RefCount=%d, Memory=%.2fMB, Access=%d)",
		t.Format, t.Width, t.Height, t.RefCount, float64(t.MemorySize)/(1024*1024), t.AccessCount)
}
