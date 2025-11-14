package render

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// TemporaryImage 临时图像缓存
type TemporaryImage struct {
	Image   *ebiten.Image
	Width   int
	Height  int
	Used    bool
	LastUse int64
}

// TemporaryImageCache 临时图像缓存池
type TemporaryImageCache struct {
	cache     map[string][]*TemporaryImage
	counter   int64
	maxImages int
	stats     ClippingStats
}

// ClippingStats 剪裁统计
type ClippingStats struct {
	TotalImages int
	ActiveImages int
	CacheHit    int64
	CacheMiss   int64
	HitRate     float64
}

// NewTemporaryImageCache 创建缓存
func NewTemporaryImageCache() *TemporaryImageCache {
	return &TemporaryImageCache{
		cache:     make(map[string][]*TemporaryImage),
		maxImages: 64,
	}
}

// GetOrCreate 获取或创建临时图像
func (c *TemporaryImageCache) GetOrCreate(width, height int) (*ebiten.Image, error) {
	if width <= 0 || height <= 0 {
		return nil, nil
	}

	key := fmt.Sprintf("%dx%d", width, height)

	// 尝试从缓存获取
	for _, tempImg := range c.cache[key] {
		if !tempImg.Used {
			tempImg.Used = true
			tempImg.LastUse = c.counter
			c.stats.CacheHit++
			return tempImg.Image, nil
		}
	}

	// 缓存未命中，创建新图像
	c.stats.CacheMiss++
	c.counter++

	img := ebiten.NewImage(width, height)
	c.addToCache(key, img, width, height)

	return img, nil
}

// addToCache 添加到缓存
func (c *TemporaryImageCache) addToCache(key string, img *ebiten.Image, width, height int) {
	tempImg := &TemporaryImage{
		Image:   img,
		Width:   width,
		Height:  height,
		Used:    true,
		LastUse: c.counter,
	}

	c.cache[key] = append(c.cache[key], tempImg)
	c.stats.TotalImages++
	c.stats.ActiveImages++
}

// Release 释放图像
func (c *TemporaryImageCache) Release(img *ebiten.Image) {
	if img == nil {
		return
	}

	for _, images := range c.cache {
		for _, tempImg := range images {
			if tempImg.Image == img {
				tempImg.Used = false
				c.stats.ActiveImages--
				return
			}
		}
	}
}

// Cleanup 清理未使用的图像
func (c *TemporaryImageCache) Cleanup() {
	for key, images := range c.cache {
		var activeImages []*TemporaryImage
		for _, img := range images {
			if img.Used {
				activeImages = append(activeImages, img)
			} else {
				img.Image.Dispose()
				c.stats.TotalImages--
			}
		}
		if len(activeImages) == 0 {
			delete(c.cache, key)
		} else {
			c.cache[key] = activeImages
		}
	}

	// 更新统计
	if c.stats.CacheHit+c.stats.CacheMiss > 0 {
		c.stats.HitRate = float64(c.stats.CacheHit) / float64(c.stats.CacheHit+c.stats.CacheMiss) * 100
	}
}

// GetStats 获取统计
func (c *TemporaryImageCache) GetStats() ClippingStats {
	return c.stats
}
