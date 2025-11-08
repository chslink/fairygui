package core

var (
	transitionSoundPlayer func(name string, volume float64)
	buttonSoundPlayer     func(name string, volume float64)
)

// SetTransitionSoundPlayer 注册 Transition 播放时使用的声音回调。
func SetTransitionSoundPlayer(fn func(name string, volume float64)) {
	transitionSoundPlayer = fn
}

// SetButtonSoundPlayer 注册按钮点击音效播放回调。
// 参见 TypeScript 版本: 相当于在应用层设置 UIPackage.inst.setButtonSoundPlayer
func SetButtonSoundPlayer(fn func(name string, volume float64)) {
	buttonSoundPlayer = fn
}

func playTransitionSound(name string, volume float64) {
	if name == "" {
		return
	}
	if transitionSoundPlayer != nil {
		transitionSoundPlayer(name, volume)
	}
}
