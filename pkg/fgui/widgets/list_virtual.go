package widgets

import (
	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// ItemInfo 虚拟列表项信息，对应 TypeScript 版本的 ItemInfo
type ItemInfo struct {
	obj        *core.GObject // 实际对象
	width      int           // 计算宽度
	height     int           // 计算高度
	selected   bool          // 选择状态
	updateFlag int           // 更新标记，用于标识item是否在本次处理中已经被重用了
}

// VirtualListConfig 虚拟列表配置
type VirtualListConfig struct {
	virtual         bool           // 是否启用虚拟化
	loop            bool           // 是否循环
	numItems        int            // 数据项总数
	realNumItems    int            // 实际项数（循环模式）
	firstIndex      int            // 左上角索引
	curLineItemCount int           // 每行项目数
	curLineItemCount2 int          // 只用在页面模式，表示垂直方向的项目数
	itemSize        *laya.Point    // 项目尺寸
	virtualListChanged int         // 1-内容改变, 2-尺寸改变
	virtualItems    []*ItemInfo    // 虚拟项数组
	itemInfoVer     int            // 项信息版本
	eventLocked     bool           // 事件锁定
}

// NewVirtualListConfig 创建虚拟列表配置
func NewVirtualListConfig() *VirtualListConfig {
	return &VirtualListConfig{
		itemSize:     &laya.Point{},
		virtualItems: make([]*ItemInfo, 0),
	}
}

// ResetVirtualState 重置虚拟状态
func (c *VirtualListConfig) ResetVirtualState() {
	c.firstIndex = 0
	c.curLineItemCount = 0
	c.curLineItemCount2 = 0
	c.virtualListChanged = 0
	c.itemInfoVer = 0
}

// CheckVirtualList 检查虚拟列表是否需要刷新
func (c *VirtualListConfig) CheckVirtualList() bool {
	return c.virtualListChanged != 0
}

// SetVirtualListChangedFlag 设置虚拟列表改变标记
func (c *VirtualListConfig) SetVirtualListChangedFlag(layoutChanged bool) {
	if layoutChanged {
		c.virtualListChanged = 2
	} else if c.virtualListChanged == 0 {
		c.virtualListChanged = 1
	}
}

// GetItemInfo 获取指定索引的项信息
func (c *VirtualListConfig) GetItemInfo(index int) *ItemInfo {
	if index < 0 || index >= len(c.virtualItems) {
		return nil
	}
	return c.virtualItems[index]
}

// EnsureVirtualItems 确保虚拟项数组大小
func (c *VirtualListConfig) EnsureVirtualItems(count int) {
	currentCount := len(c.virtualItems)
	if currentCount < count {
		// 扩展数组
		for i := currentCount; i < count; i++ {
			c.virtualItems = append(c.virtualItems, &ItemInfo{})
		}
	} else if currentCount > count {
		// 缩容数组
		c.virtualItems = c.virtualItems[:count]
	}
}

// ResetAllUpdateFlags 重置所有更新标记
func (c *VirtualListConfig) ResetAllUpdateFlags() {
	c.itemInfoVer++
}

// IsItemUpdated 检查项是否已更新
func (c *VirtualListConfig) IsItemUpdated(item *ItemInfo) bool {
	return item.updateFlag == c.itemInfoVer
}

// MarkItemUpdated 标记项为已更新
func (c *VirtualListConfig) MarkItemUpdated(item *ItemInfo) {
	item.updateFlag = c.itemInfoVer
}