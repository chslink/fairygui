package core

var transitionSoundPlayer func(name string, volume float64)

// SetTransitionSoundPlayer 注册 Transition 播放时使用的声音回调。
func SetTransitionSoundPlayer(fn func(name string, volume float64)) {
	transitionSoundPlayer = fn
}

func playTransitionSound(name string, volume float64) {
	if name == "" {
		return
	}
	if transitionSoundPlayer != nil {
		transitionSoundPlayer(name, volume)
	}
}
