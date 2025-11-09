package widgets

import "github.com/chslink/fairygui/pkg/fgui/utils"

// ExtensionConstructor 用于需要从构造数据读取属性的特殊组件
// 这些组件延迟到完整构建完成后再调用 ConstructExtension
// 对应 TypeScript 版本的 constructExtension(buffer: ByteBuffer) 方法
type ExtensionConstructor interface {
	// ConstructExtension 在组件完整构建后、SetupAfterAdd 之前调用
	// 用于读取部分 6（section 6）的扩展属性并绑定事件
	ConstructExtension(buf *utils.ByteBuffer) error
}
