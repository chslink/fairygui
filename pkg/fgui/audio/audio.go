package audio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	audio2 "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	// defaultSampleRate 默认采样率 - 48000Hz 是 Ebiten 推荐的标准采样率
	// 音频解码时会自动重采样到此采样率，确保播放正确
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
	audioContext *audio2.Context
}

// GetInstance 获取音频播放器单例
func GetInstance() *AudioPlayer {
	once.Do(func() {
		instance = &AudioPlayer{
			audioContext: audio2.NewContext(defaultSampleRate),
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
	p.audioContext = audio2.NewContext(sampleRate)
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
	go p.playBytes(data, volume)
}

// tryAutoLoadAndPlay 尝试自动从包中加载音效并播放
func (p *AudioPlayer) tryAutoLoadAndPlay(url string, volume float64) {
	// 检查是否已在缓存中
	if hasAudioDataInCache(url) {
		return
	}

	// 异步加载并播放
	go func() {
		ctx := context.Background()
		data, err := globalLoader.LoadOne(ctx, url, assets.ResourceSound)
		if err != nil {
			fmt.Printf("[WARN] Failed to load audio %s: %v\n", url, err)
			return
		}

		// 加载成功，注册到缓存并播放
		p.RegisterAudioData(url, data)
		p.playBytes(data, volume)
	}()
}

// hasAudioDataInCache 检查缓存中是否有音频数据
func hasAudioDataInCache(key string) bool {
	cacheMutex.RLock()
	_, ok := audioCache[key]
	cacheMutex.RUnlock()
	return ok
}

// playBytes 解码并播放音频数据
// 自动检测音频格式并重采样到正确的采样率
func (p *AudioPlayer) playBytes(data []byte, volume float64) {
	// 获取目标采样率
	targetSampleRate := p.audioContext.SampleRate()

	// 尝试不同的音频格式解码
	var stream io.ReadSeeker
	var err error

	// 1. 尝试 WAV 格式
	if stream, err = p.decodeWAV(data, targetSampleRate); err == nil {
		p.playStream(stream, volume)
		return
	}

	// 2. 尝试 MP3 格式
	if stream, err = p.decodeMP3(data, targetSampleRate); err == nil {
		p.playStream(stream, volume)
		return
	}

	// 3. 尝试 Ogg 格式
	if stream, err = p.decodeOgg(data, targetSampleRate); err == nil {
		p.playStream(stream, volume)
		return
	}

	fmt.Printf("[ERROR] Failed to decode audio: all formats failed\n")
}

// playStream 播放音频流
// DecodeWithSampleRate 已经处理了重采样，这里直接播放即可
func (p *AudioPlayer) playStream(stream io.ReadSeeker, volume float64) {
	// 创建播放器
	player, err := p.audioContext.NewPlayer(stream)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create player: %v\n", err)
		return
	}

	// 设置音量并播放
	player.SetVolume(volume)
	player.Play()

	// 播放完成后自动清理
	// 注意：Ebiten 的 Player 会在播放结束后自动释放资源
}

// decodeWAV 解码 WAV 格式音频
// DecodeWithSampleRate 会自动将音频重采样到指定的采样率
func (p *AudioPlayer) decodeWAV(data []byte, sampleRate int) (io.ReadSeeker, error) {
	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return stream, nil
}

// decodeMP3 解码 MP3 格式音频
// DecodeWithSampleRate 会自动将音频重采样到指定的采样率
func (p *AudioPlayer) decodeMP3(data []byte, sampleRate int) (io.ReadSeeker, error) {
	stream, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return stream, nil
}

// decodeOgg 解码 Ogg 格式音频
// DecodeWithSampleRate 会自动将音频重采样到指定的采样率
func (p *AudioPlayer) decodeOgg(data []byte, sampleRate int) (io.ReadSeeker, error) {
	stream, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return stream, nil
}
