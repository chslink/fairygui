package widgets

import (
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// ObjectCreator 对象创建器接口，用于创建新对象
type ObjectCreator interface {
	CreateObject(url string) *core.GObject
}

// GObjectPool 对象池管理，用于虚拟列表中的对象重用
// 严格遵循TypeScript版本的实现设计
type GObjectPool struct {
	pool        map[string][]*core.GObject // URL -> 对象数组
	count       int                          // 总数
	creator     ObjectCreator                // 对象创建器
}

// NewGObjectPool 创建新的对象池
func NewGObjectPool() *GObjectPool {
	return &GObjectPool{
		pool: make(map[string][]*core.GObject),
	}
}

// NewGObjectPoolWithCreator 使用对象创建器创建新的对象池
func NewGObjectPoolWithCreator(creator ObjectCreator) *GObjectPool {
	return &GObjectPool{
		pool:    make(map[string][]*core.GObject),
		creator: creator,
	}
}

// Clear 清空对象池并释放所有对象
// 注意：与TypeScript版本不同，Go版本GObject没有Dispose方法，所以这里只从父对象移除
func (p *GObjectPool) Clear() {
	for _, arr := range p.pool {
		for _, obj := range arr {
			if obj != nil {
				// 从父对象移除
				if parent := obj.Parent(); parent != nil {
					parent.RemoveChild(obj)
				}
				// 注意：理想情况下应该调用Dispose方法，但当前GObject没有实现
			}
		}
	}
	p.pool = make(map[string][]*core.GObject)
	p.count = 0
}

// Count 返回池中对象总数
func (p *GObjectPool) Count() int {
	return p.count
}

// GetObject 从池中获取对象，如果没有则创建新对象
// 注意：TypeScript版本使用UIPackage.normalizeURL和UIPackage.createObjectFromURL
// Go版本使用注入的ObjectCreator来替代UIPackage.createObjectFromURL
func (p *GObjectPool) GetObject(url string) *core.GObject {
	// TypeScript版本会标准化URL，但Go版本暂时直接使用
	if url == "" {
		return nil
	}

	arr := p.pool[url]
	if len(arr) > 0 {
		p.count--
		// 从数组头部取出对象（与TypeScript的shift()行为一致）
		obj := arr[0]
		p.pool[url] = arr[1:]
		return obj
	}

	// 池中没有可用对象，需要创建新对象
	if p.creator != nil {
		obj := p.creator.CreateObject(url)
		if obj != nil {
			return obj
		}
	}

	return nil
}

// ReturnObject 将对象返回池中
// 注意：TypeScript版本使用obj.resourceURL，但Go版本的GObject没有这个属性
// 这里使用obj.Name()作为临时替代，实际应该在GObject中添加ResourceURL属性
func (p *GObjectPool) ReturnObject(obj *core.GObject) {
	if obj == nil {
		return
	}

	// 从父对象移除
	if parent := obj.Parent(); parent != nil {
		parent.RemoveChild(obj)
	}

	// 注意：这里是关键差异点，TypeScript版本使用resourceURL
	// 理想情况下，应该给GObject添加ResourceURL()方法
	url := obj.Name() // 临时替代方案
	if url == "" {
		return
	}

	arr := p.pool[url]
	if arr == nil {
		arr = make([]*core.GObject, 0)
		p.pool[url] = arr
	}

	p.count++
	// 添加到数组尾部（与TypeScript的push()行为一致）
	p.pool[url] = append(arr, obj)
}

// RemoveObject 从池中移除指定对象
// 注意：这是Go版本额外添加的方法，TypeScript版本没有
func (p *GObjectPool) RemoveObject(obj *core.GObject) {
	if obj == nil {
		return
	}

	url := obj.Name() // 同样使用Name作为临时替代
	if url == "" {
		return
	}

	arr := p.pool[url]
	if len(arr) == 0 {
		return
	}

	// 从数组中移除指定对象
	for i, o := range arr {
		if o == obj {
			p.pool[url] = append(arr[:i], arr[i+1:]...)
			p.count--
			break
		}
	}
}