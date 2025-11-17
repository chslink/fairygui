package fairygui

import "sync"

// basicEventDispatcher 是 EventDispatcher 的基础实现。
type basicEventDispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

// newBasicEventDispatcher 创建一个新的事件分发器。
func newBasicEventDispatcher() *basicEventDispatcher {
	return &basicEventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

// On 注册事件处理器。
func (d *basicEventDispatcher) On(eventType string, handler EventHandler) func() {
	if handler == nil {
		return func() {}
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.handlers == nil {
		d.handlers = make(map[string][]EventHandler)
	}

	d.handlers[eventType] = append(d.handlers[eventType], handler)

	// 返回取消函数
	return func() {
		d.Off(eventType, handler)
	}
}

// Off 移除事件处理器。
func (d *basicEventDispatcher) Off(eventType string, handler EventHandler) {
	if handler == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	handlers, ok := d.handlers[eventType]
	if !ok {
		return
	}

	// 移除处理器
	for i, h := range handlers {
		// 比较函数指针（注意：这在 Go 中不是完全可靠的）
		if &h == &handler {
			d.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	// 如果没有处理器了，删除键
	if len(d.handlers[eventType]) == 0 {
		delete(d.handlers, eventType)
	}
}

// Once 注册一次性事件处理器。
func (d *basicEventDispatcher) Once(eventType string, handler EventHandler) {
	if handler == nil {
		return
	}

	var executed bool
	var cancel func()
	wrapper := func(event Event) {
		if executed {
			return
		}
		executed = true
		handler(event)
		if cancel != nil {
			cancel()
		}
	}

	cancel = d.On(eventType, wrapper)
}

// Emit 触发事件。
func (d *basicEventDispatcher) Emit(event Event) {
	if event == nil {
		return
	}

	d.mu.RLock()
	handlers, ok := d.handlers[event.Type()]
	if !ok {
		d.mu.RUnlock()
		return
	}

	// 复制处理器列表，避免在处理过程中被修改
	handlersCopy := make([]EventHandler, len(handlers))
	copy(handlersCopy, handlers)
	d.mu.RUnlock()

	// 调用所有处理器
	for _, handler := range handlersCopy {
		handler(event)

		// 如果事件传播被停止，退出
		if event.IsPropagationStopped() {
			break
		}
	}
}

// HasListener 返回是否有指定类型的事件监听器。
func (d *basicEventDispatcher) HasListener(eventType string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	handlers, ok := d.handlers[eventType]
	return ok && len(handlers) > 0
}
