package fairygui

import "sync"

// handlerWrapper 包装事件处理器，添加 ID 用于移除
type handlerWrapper struct {
	id      int
	handler EventHandler
}

// basicEventDispatcher 是 EventDispatcher 的基础实现。
type basicEventDispatcher struct {
	mu        sync.RWMutex
	handlers  map[string][]handlerWrapper
	nextID    int
}

// newBasicEventDispatcher 创建一个新的事件分发器。
func newBasicEventDispatcher() *basicEventDispatcher {
	return &basicEventDispatcher{
		handlers: make(map[string][]handlerWrapper),
		nextID:   1,
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
		d.handlers = make(map[string][]handlerWrapper)
	}

	// 分配唯一 ID
	id := d.nextID
	d.nextID++

	wrapper := handlerWrapper{
		id:      id,
		handler: handler,
	}

	d.handlers[eventType] = append(d.handlers[eventType], wrapper)

	// 返回取消函数
	return func() {
		d.removeByID(eventType, id)
	}
}

// removeByID 通过 ID 移除处理器
func (d *basicEventDispatcher) removeByID(eventType string, id int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	handlers, ok := d.handlers[eventType]
	if !ok {
		return
	}

	// 找到并移除对应 ID 的处理器
	for i, wrapper := range handlers {
		if wrapper.id == id {
			d.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	// 如果没有处理器了，删除键
	if len(d.handlers[eventType]) == 0 {
		delete(d.handlers, eventType)
	}
}

// Off 移除事件处理器。
// 注意：由于 Go 函数无法可靠比较，建议使用 On 返回的 cancel 函数来移除处理器
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

	// 尝试通过函数地址匹配（不完全可靠）
	// 更好的做法是使用 On 返回的 cancel 函数
	for i, wrapper := range handlers {
		// 尝试地址比较（有限的可靠性）
		h1 := wrapper.handler
		h2 := handler
		if &h1 == &h2 {
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
	handlersCopy := make([]handlerWrapper, len(handlers))
	copy(handlersCopy, handlers)
	d.mu.RUnlock()

	// 调用所有处理器
	for _, wrapper := range handlersCopy {
		wrapper.handler(event)

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
