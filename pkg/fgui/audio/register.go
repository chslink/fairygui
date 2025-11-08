package audio

import (
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// RegisterAsDefaultButtonSoundPlayer 将音频播放器注册为FairyGUI的默认按钮音效播放器
// 这使得所有按钮点击都会播放配置的音效
func RegisterAsDefaultButtonSoundPlayer() {
	player := GetInstance()

	// 注册按钮音效播放器
	core.SetButtonSoundPlayer(func(name string, volume float64) {
		player.Play(name, volume)
	})
}

// RegisterAudio 注册音频数据
// 便捷函数，用于注册音频资源
func RegisterAudio(name string, data []byte) {
	GetInstance().RegisterAudioData(name, data)
}
