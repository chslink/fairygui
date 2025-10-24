package core

import (
	"sync"
	"time"
)

type tickerFunc func(time.Duration)

var (
	tickerMutex sync.Mutex
	tickerSeq   uint64
	tickers     = make(map[uint64]tickerFunc)
)

// RegisterTicker registers a callback that receives the frame delta during Root.Advance.
// It returns a function that removes the ticker when invoked.
func RegisterTicker(fn func(time.Duration)) func() {
	if fn == nil {
		return func() {}
	}
	tickerMutex.Lock()
	defer tickerMutex.Unlock()
	tickerSeq++
	id := tickerSeq
	tickers[id] = fn
	return func() {
		tickerMutex.Lock()
		delete(tickers, id)
		tickerMutex.Unlock()
	}
}

func tickAll(delta time.Duration) {
	tickerMutex.Lock()
	if len(tickers) == 0 {
		tickerMutex.Unlock()
		return
	}
	snapshot := make([]tickerFunc, 0, len(tickers))
	for _, fn := range tickers {
		snapshot = append(snapshot, fn)
	}
	tickerMutex.Unlock()
	for _, fn := range snapshot {
		if fn != nil {
			fn(delta)
		}
	}
}
