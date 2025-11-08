package audio

import (
	"bytes"
	"context"
	"sync"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	// defaultSampleRate 默认采样率
	defaultSampleRate = 48000
)

var (
	// instance 音频播放器单例
	instance *AudioPlayer
	once     sync.Once

	// 缓存的音频数据，key: 文件路径或URL，value: 音频字节数据
	audioCache = make(map[string][]byte)
	cacheMutex sync.RWMutex

	// globalLoader 全局资源加载器，用于自动加载音效数据
	globalLoader assets.Loader
)

// AudioPlayer 音频播放器
type AudioPlayer struct {
	audioContext *audio.Context
}

// GetInstance 获取音频播放器单例
func GetInstance() *AudioPlayer {
	once.Do(func() {
		instance = &AudioPlayer{
			audioContext: audio.NewContext(defaultSampleRate),
		}
	})
	return instance
}

// Init 初始化音频播放器
// sampleRate 采样率，默认48000
func (p *AudioPlayer) Init(sampleRate int) {
	if p.audioContext != nil {
		return
	}
	if sampleRate <= 0 {
		sampleRate = defaultSampleRate
	}
	p.audioContext = audio.NewContext(sampleRate)
}

// SetLoader 设置全局资源加载器
// 用于自动从包中加载音效数据
func SetLoader(loader assets.Loader) {
	globalLoader = loader
}

// SetVolume 设置全局音量
// volume 音量，范围0-1
func (p *AudioPlayer) SetVolume(volume float64) {
	// Ebiten的音频系统没有全局音量设置
	// 需要在播放时指定音量
}

// LoadFile 从文件路径加载音频文件
// 支持 MP3、Wav、Ogg 格式
func (p *AudioPlayer) LoadFile(filePath string) error {
	// 优先从缓存获取
	cacheMutex.RLock()
	if _, ok := audioCache[filePath]; ok {
		cacheMutex.RUnlock()
		return nil // 已缓存
	}
	cacheMutex.RUnlock()

	// TODO: 实现文件读取
	// 目前需要预先注册音频数据
	return nil
}

// RegisterAudioData 注册音频数据到缓存
// filePath 文件标识，可以是URL、文件路径或包资源ID
// data 音频字节数据
func (p *AudioPlayer) RegisterAudioData(filePath string, data []byte) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	audioCache[filePath] = data
}

// Play 播放音效
// filePath 音频文件标识，可以是URL（ui://package/item）或文件路径
// volume 音量，范围0-1
func (p *AudioPlayer) Play(filePath string, volume float64) {
	if p.audioContext == nil {
		p.Init(defaultSampleRate)
	}

	// 获取音频数据
	cacheMutex.RLock()
	data, ok := audioCache[filePath]
	cacheMutex.RUnlock()

	if !ok {
		// 缓存中没有，尝试自动从包中加载
		if globalLoader != nil {
			p.tryAutoLoadAndPlay(filePath, volume)
		}
		return
	}

	// 异步播放音效
	go func() {
		p.playBytes(data, volume)
	}()
}

// tryAutoLoadAndPlay 尝试自动从包中加载音效并播放
func (p *AudioPlayer) tryAutoLoadAndPlay(url string, volume float64) {
	// 检查是否是FairyGUI的URL格式（ui://package/item）
	if !hasAudioDataInCache(url) && shouldTryLoadFromPackage(url) {
		// 尝试从包中查找音效资源
		if item := assets.GetItemByURL(url); item != nil && item.File != "" {
			// 找到音效资源，尝试加载
			go func() {
				// 使用全局Loader加载音频文件
				ctx := context.Background()
				if data, err := globalLoader.LoadOne(ctx, item.File, assets.ResourceSound); err == nil {
					// 加载成功，注册到缓存并播放
					p.RegisterAudioData(url, data)
					// 重新获取数据并播放
					cacheMutex.RLock()
					audioData := audioCache[url]
					cacheMutex.RUnlock()
					if audioData != nil {
						p.playBytes(audioData, volume)
					}
				}
			}()
		}
	}
}

// hasAudioDataInCache 检查缓存中是否有音频数据
func hasAudioDataInCache(key string) bool {
	cacheMutex.RLock()
	_, ok := audioCache[key]
	cacheMutex.RUnlock()
	return ok
}

// shouldTryLoadFromPackage 检查是否应该尝试从包中加载
func shouldTryLoadFromPackage(url string) bool {
	// 只有FairyGUI格式的URL才尝试从包中加载
	return len(url) > 5 && url[:5] == "ui://"
}

func (p *AudioPlayer) playBytes(data []byte, volume float64) {
	// 尝试解码音频
	var player *audio.Player
	var err error

	// 首先尝试Wav格式（最简单）
	if player, err = p.tryDecodeWav(data, volume); err == nil {
		player.Play()
		return
	}

	// 尝试MP3格式
	if player, err = p.tryDecodeMP3(data, volume); err == nil {
		player.Play()
		return
	}

	// 尝试Ogg格式
	if player, err = p.tryDecodeOgg(data, volume); err == nil {
		player.Play()
		return
	}

	// 所有格式都失败了
	// TODO: 可以添加日志记录
}

func (p *AudioPlayer) tryDecodeWav(data []byte, volume float64) (*audio.Player, error) {
	s, err := wav.DecodeF32(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	player, err := p.audioContext.NewPlayerF32(s)
	if err != nil {
		return nil, err
	}
	player.SetVolume(volume)
	return player, nil
}

func (p *AudioPlayer) tryDecodeMP3(data []byte, volume float64) (*audio.Player, error) {
	s, err := mp3.DecodeF32(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	player, err := p.audioContext.NewPlayerF32(s)
	if err != nil {
		return nil, err
	}
	player.SetVolume(volume)
	return player, nil
}

func (p *AudioPlayer) tryDecodeOgg(data []byte, volume float64) (*audio.Player, error) {
	s, err := vorbis.DecodeF32(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	player, err := p.audioContext.NewPlayerF32(s)
	if err != nil {
		return nil, err
	}
	player.SetVolume(volume)
	return player, nil
}
