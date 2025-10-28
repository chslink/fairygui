package widgets

import (
	"log"
	"math"
)

// refreshVirtualList 刷新虚拟列表 - 对应 TypeScript 版本的 _refreshVirtualList
func (l *GList) refreshVirtualList() {
	if !l.virtual {
		return
	}

	log.Printf("🔄 refreshVirtualList 开始")

	// 检查显示对象是否存在
	displayObj := l.GComponent.GObject.DisplayObject()
	if displayObj == nil {
		log.Printf("❌ DisplayObject为nil")
		return
	}

	// 调试信息
	if l.creator == nil {
		log.Printf("❌ creator为nil，无法创建对象")
		return
	}

	// 获取视图尺寸
	viewWidth := l.getViewWidth()
	viewHeight := l.getViewHeight()
	log.Printf("   视图尺寸: %dx%d", viewWidth, viewHeight)
	log.Printf("   列表尺寸: %.0fx%.0f", l.GComponent.Width(), l.GComponent.Height())
	if scrollPane := l.GComponent.ScrollPane(); scrollPane != nil {
		log.Printf("   ScrollPane视图: %.0fx%.0f", scrollPane.ViewWidth(), scrollPane.ViewHeight())
	}

	layoutChanged := l.virtualListChanged == 2
	l.virtualListChanged = 0
	l.eventLocked = true

	log.Printf("   layoutChanged=%v, realNumItems=%d", layoutChanged, l.realNumItems)

	// 计算每行项目数
	if layoutChanged {
		l.calculateLineItemCount()
		log.Printf("   每行项目数: %d", l.curLineItemCount)
	}

	// 计算内容尺寸
	var contentWidth, contentHeight float64
	if l.realNumItems > 0 {
		contentWidth, contentHeight = l.calculateContentSize()
		log.Printf("   内容尺寸: %.0fx%.0f", contentWidth, contentHeight)
	}

	// 处理对齐
	l.handleAlign(contentWidth, contentHeight)

	// 设置ScrollPane的内容尺寸
	// 这是关键：ScrollPane需要知道内容总尺寸才能显示滚动条
	if scrollPane := l.GComponent.ScrollPane(); scrollPane != nil {
		log.Printf("   设置ScrollPane内容尺寸: %.0fx%.0f", contentWidth, contentHeight)
		scrollPane.SetContentSize(contentWidth, contentHeight)
		log.Printf("   ScrollPane状态: viewSize=%.0fx%.0f, contentSize=%.0fx%.0f",
			scrollPane.ViewWidth(), scrollPane.ViewHeight(),
			scrollPane.ContentSize().X, scrollPane.ContentSize().Y)
	}

	// 处理滚动 - 关键修复：强制更新，确保第一次初始化时也能创建子组件
	log.Printf("   调用handleScroll(true)")
	l.handleScroll(true)

	childCount := len(l.GComponent.Children())
	log.Printf("✅ refreshVirtualList 完成，子对象数=%d", childCount)

	l.eventLocked = false

	// 处理拱形顺序
	l.handleArchOrder1()
	l.handleArchOrder2()
}

// calculateLineItemCount 计算每行项目数
func (l *GList) calculateLineItemCount() {
	switch l.layout {
	case ListLayoutTypeSingleColumn, ListLayoutTypeSingleRow:
		l.curLineItemCount = 1
	case ListLayoutTypeFlowHorizontal:
		if l.columnCount > 0 {
			l.curLineItemCount = l.columnCount
		} else {
			// 根据视图宽度计算
			viewWidth := l.getViewWidth()
			if viewWidth > 0 && l.itemSize.X > 0 {
				l.curLineItemCount = int(math.Floor((float64(viewWidth) + float64(l.columnGap)) / (float64(l.itemSize.X) + float64(l.columnGap))))
				if l.curLineItemCount <= 0 {
					l.curLineItemCount = 1
				}
			} else {
				l.curLineItemCount = 1
			}
		}
	case ListLayoutTypeFlowVertical:
		if l.lineCount > 0 {
			l.curLineItemCount = l.lineCount
		} else {
			// 根据视图高度计算
			viewHeight := l.getViewHeight()
			if viewHeight > 0 && l.itemSize.Y > 0 {
				l.curLineItemCount = int(math.Floor((float64(viewHeight) + float64(l.lineGap)) / (float64(l.itemSize.Y) + float64(l.lineGap))))
				if l.curLineItemCount <= 0 {
					l.curLineItemCount = 1
				}
			} else {
				l.curLineItemCount = 1
			}
		}
	case ListLayoutTypePagination:
		// 水平方向
		if l.columnCount > 0 {
			l.curLineItemCount = l.columnCount
		} else {
			viewWidth := l.getViewWidth()
			if viewWidth > 0 && l.itemSize.X > 0 {
				l.curLineItemCount = int(math.Floor((float64(viewWidth) + float64(l.columnGap)) / (float64(l.itemSize.X) + float64(l.columnGap))))
				if l.curLineItemCount <= 0 {
					l.curLineItemCount = 1
				}
			} else {
				l.curLineItemCount = 1
			}
		}

		// 垂直方向
		if l.lineCount > 0 {
			l.curLineItemCount2 = l.lineCount
		} else {
			viewHeight := l.getViewHeight()
			if viewHeight > 0 && l.itemSize.Y > 0 {
				l.curLineItemCount2 = int(math.Floor((float64(viewHeight) + float64(l.lineGap)) / (float64(l.itemSize.Y) + float64(l.lineGap))))
				if l.curLineItemCount2 <= 0 {
					l.curLineItemCount2 = 1
				}
			} else {
				l.curLineItemCount2 = 1
			}
		}
	}
}

// calculateContentSize 计算内容尺寸
func (l *GList) calculateContentSize() (float64, float64) {
	var contentWidth, contentHeight float64

	if l.realNumItems == 0 {
		return 0, 0
	}

	// 计算总行数
	lineCount := int(math.Ceil(float64(l.realNumItems) / float64(l.curLineItemCount)))

	switch l.layout {
	case ListLayoutTypeSingleColumn, ListLayoutTypeFlowHorizontal:
		// 计算高度
		for i := 0; i < lineCount; i++ {
			lineHeight := 0
			startIdx := i * l.curLineItemCount
			endIdx := startIdx + l.curLineItemCount
			if endIdx > l.realNumItems {
				endIdx = l.realNumItems
			}

			for j := startIdx; j < endIdx; j++ {
				ii := l.virtualItems[j]
				if ii != nil && ii.height > lineHeight {
					lineHeight = ii.height
				}
			}

			// 如果没有实际的项目高度，使用预估的项目尺寸
			if lineHeight == 0 && l.itemSize.Y > 0 {
				lineHeight = int(l.itemSize.Y)
			}

			if i > 0 {
				contentHeight += float64(l.lineGap)
			}
			contentHeight += float64(lineHeight)
		}

		// 计算宽度
		if l.autoResizeItem {
			contentWidth = float64(l.getViewWidth())
		} else {
			// 取第一行的宽度
			for i := 0; i < l.curLineItemCount && i < l.realNumItems; i++ {
				ii := l.virtualItems[i]
				itemWidth := 0
				if ii != nil && ii.width > 0 {
					itemWidth = ii.width
				} else if l.itemSize.X > 0 {
					itemWidth = int(l.itemSize.X)
				}
				contentWidth += float64(itemWidth)
				if i > 0 {
					contentWidth += float64(l.columnGap)
				}
			}
		}

	case ListLayoutTypeSingleRow, ListLayoutTypeFlowVertical:
		// 计算宽度
		for i := 0; i < lineCount; i++ {
			lineWidth := 0
			startIdx := i * l.curLineItemCount
			endIdx := startIdx + l.curLineItemCount
			if endIdx > l.realNumItems {
				endIdx = l.realNumItems
			}

			for j := startIdx; j < endIdx; j++ {
				ii := l.virtualItems[j]
				if ii != nil && ii.width > lineWidth {
					lineWidth = ii.width
				}
			}

			if i > 0 {
				contentWidth += float64(l.columnGap)
			}
			contentWidth += float64(lineWidth)
		}

		// 计算高度
		if l.autoResizeItem {
			contentHeight = float64(l.getViewHeight())
		} else {
			// 取第一行的高度
			for i := 0; i < l.curLineItemCount && i < l.realNumItems; i++ {
				ii := l.virtualItems[i]
				if ii != nil {
					contentHeight += float64(ii.height)
					if i > 0 {
						contentHeight += float64(l.lineGap)
					}
				}
			}
		}

	case ListLayoutTypePagination:
		// 分页模式，需要特殊处理
		pageCount := int(math.Ceil(float64(l.realNumItems) / float64(l.curLineItemCount*l.curLineItemCount2)))

		// 计算总宽度（页数 * 视图宽度）
		contentWidth = float64(pageCount * l.getViewWidth())

		// 计算高度
		if l.autoResizeItem {
			contentHeight = float64(l.getViewHeight())
		} else {
			// 计算每页的最大高度
			for page := 0; page < pageCount; page++ {
				pageHeight := 0
				startIdx := page * l.curLineItemCount * l.curLineItemCount2
				endIdx := startIdx + l.curLineItemCount*l.curLineItemCount2
				if endIdx > l.realNumItems {
					endIdx = l.realNumItems
				}

				for i := startIdx; i < endIdx; i += l.curLineItemCount {
					lineHeight := 0
					lineEnd := i + l.curLineItemCount
					if lineEnd > endIdx {
						lineEnd = endIdx
					}

					for j := i; j < lineEnd; j++ {
						ii := l.virtualItems[j]
						if ii != nil && ii.height > lineHeight {
							lineHeight = ii.height
						}
					}

					if i > startIdx {
						pageHeight += l.lineGap
					}
					pageHeight += lineHeight
				}

				if pageHeight > int(contentHeight) {
					contentHeight = float64(pageHeight)
				}
			}
		}
	}

	return contentWidth, contentHeight
}

// handleAlign 处理对齐
func (l *GList) handleAlign(contentWidth, contentHeight float64) {
	if !l.virtual {
		return
	}

	var newOffsetX, newOffsetY float64

	// 垂直对齐
	viewHeight := float64(l.getViewHeight())
	if contentHeight < viewHeight {
		switch l.verticalAlign {
		case LoaderAlignMiddle:
			newOffsetY = math.Floor((viewHeight - contentHeight) / 2)
		case LoaderAlignBottom:
			newOffsetY = viewHeight - contentHeight
		}
	}

	// 水平对齐
	viewWidth := float64(l.getViewWidth())
	if contentWidth < viewWidth {
		switch l.align {
		case LoaderAlignCenter:
			newOffsetX = math.Floor((viewWidth - contentWidth) / 2)
		case LoaderAlignRight:
			newOffsetX = viewWidth - contentWidth
		}
	}

	// 应用偏移
	if newOffsetX != 0 || newOffsetY != 0 {
		// 注意：laya.Sprite 没有 SetX/SetY 方法，需要使用 SetMatrix 或其他方式
		// 这里暂时跳过偏移设置
	}
}

// handleScroll 处理滚动 - 对应 TypeScript 版本的 handleScroll
func (l *GList) handleScroll(forceUpdate bool) {
	if !l.virtual || l.realNumItems == 0 {
		return
	}

	// 根据布局类型调用不同的滚动处理函数
	switch l.layout {
	case ListLayoutTypeSingleColumn, ListLayoutTypeFlowHorizontal:
		l.handleScroll1(forceUpdate)
	case ListLayoutTypeSingleRow, ListLayoutTypeFlowVertical:
		l.handleScroll2(forceUpdate)
	case ListLayoutTypePagination:
		l.handleScroll3(forceUpdate)
	}
}

// handleScroll1 处理单列/水平流式布局的滚动
func (l *GList) handleScroll1(forceUpdate bool) {
	// 获取滚动位置 - 对应 TypeScript 版本 GList.ts:1380
	var pos int
	scrollPane := l.GComponent.ScrollPane()
	if scrollPane != nil {
		pos = int(scrollPane.PosY())
	} else {
		pos = 0 // 默认位置
	}

	viewHeight := l.getViewHeight()
	if viewHeight == 0 {
		return
	}

	max := pos + viewHeight

	// 计算新的起始索引
	newFirstIndex := l.getIndexByPos(pos, true)
	if newFirstIndex < 0 {
		newFirstIndex = 0
	}

	// 如果范围没有改变且不需要强制更新，则返回
	if !forceUpdate && newFirstIndex == l.firstIndex {
		return
	}

	// 更新起始索引
	oldFirstIndex := l.firstIndex
	l.firstIndex = newFirstIndex

	// 处理虚拟项 - 传入pos和max用于位置计算
	l.updateVirtualItems1(oldFirstIndex, newFirstIndex, pos, max, forceUpdate)
}

// handleScroll2 处理单行/垂直流式布局的滚动
func (l *GList) handleScroll2(forceUpdate bool) bool {
	// 获取滚动位置
	scrollPane := l.GComponent.ScrollPane()
	if scrollPane == nil {
		return false
	}

	pos := int(scrollPane.PosX())
	viewWidth := l.getViewWidth()
	if viewWidth == 0 {
		return false
	}

	// 计算新的起始索引
	newFirstIndex := l.getIndexByPos(pos, true)
	if newFirstIndex < 0 {
		newFirstIndex = 0
	}

	// 计算结束索引
	newLastIndex := l.getIndexByPos(pos + viewWidth, false)
	if newLastIndex > l.realNumItems-1 {
		newLastIndex = l.realNumItems - 1
	}

	// 如果范围没有改变且不需要强制更新，则返回
	if !forceUpdate && newFirstIndex == l.firstIndex {
		return false
	}

	// 更新起始索引
	oldFirstIndex := l.firstIndex
	l.firstIndex = newFirstIndex

	// 处理虚拟项
	l.updateVirtualItems(oldFirstIndex, newFirstIndex, newLastIndex, false, forceUpdate)

	return true
}

// handleScroll3 处理分页布局的滚动
func (l *GList) handleScroll3(forceUpdate bool) {
	// 获取滚动位置
	scrollPane := l.GComponent.ScrollPane()
	if scrollPane == nil {
		return
	}

	posX := int(scrollPane.PosX())
	posY := int(scrollPane.PosY())
	viewWidth := l.getViewWidth()
	viewHeight := l.getViewHeight()

	if viewWidth == 0 || viewHeight == 0 {
		return
	}

	// 计算页索引
	pageIndex := posX / viewWidth

	// 计算页内的起始索引
	pageStartIndex := pageIndex * l.curLineItemCount * l.curLineItemCount2

	// 计算页内的起始行和列
	startRow := posY / (viewHeight / l.curLineItemCount2)
	startCol := 0

	// 计算新的起始索引
	newFirstIndex := pageStartIndex + startRow*l.curLineItemCount + startCol
	if newFirstIndex < 0 {
		newFirstIndex = 0
	}
	if newFirstIndex > l.realNumItems-1 {
		newFirstIndex = l.realNumItems - 1
	}

	// 计算结束索引（显示一页的内容）
	newLastIndex := newFirstIndex + l.curLineItemCount*l.curLineItemCount2 - 1
	if newLastIndex > l.realNumItems-1 {
		newLastIndex = l.realNumItems - 1
	}

	// 如果范围没有改变且不需要强制更新，则返回
	if !forceUpdate && newFirstIndex == l.firstIndex {
		return
	}

	// 更新起始索引
	oldFirstIndex := l.firstIndex
	l.firstIndex = newFirstIndex

	// 处理虚拟项
	l.updateVirtualItems(oldFirstIndex, newFirstIndex, newLastIndex, true, forceUpdate)
}

// getIndexByPos 根据位置获取索引
func (l *GList) getIndexByPos(pos int, first bool) int {
	if l.realNumItems == 0 {
		return -1
	}

	// 简化实现：根据位置估算索引
	if l.layout == ListLayoutTypeSingleColumn || l.layout == ListLayoutTypeFlowHorizontal {
		// 垂直滚动
		if l.itemSize.Y > 0 {
			itemHeight := int(l.itemSize.Y) + l.lineGap
			if itemHeight <= 0 {
				itemHeight = int(l.itemSize.Y) + 1
			}
			index := pos / itemHeight
			if !first {
				index = (pos + int(l.itemSize.Y)) / itemHeight
			}
			if l.loop {
				index = index % l.numItems
			}
			// 确保索引在有效范围内
			if index < 0 {
				index = 0
			}
			if index >= l.realNumItems {
				index = l.realNumItems - 1
			}
			return index
		}
	} else {
		// 水平滚动
		if l.itemSize.X > 0 {
			itemWidth := int(l.itemSize.X) + l.columnGap
			if itemWidth <= 0 {
				itemWidth = int(l.itemSize.X) + 1
			}
			index := pos / itemWidth
			if !first {
				index = (pos + int(l.itemSize.X)) / itemWidth
			}
			if l.loop {
				index = index % l.numItems
			}
			// 确保索引在有效范围内
			if index < 0 {
				index = 0
			}
			if index >= l.realNumItems {
				index = l.realNumItems - 1
			}
			return index
		}
	}

	return 0
}

// updateVirtualItems1 更新虚拟项（垂直滚动）
// 对应 TypeScript 版本 GList.ts:1379-1531 (handleScroll1)
func (l *GList) updateVirtualItems1(oldFirstIndex, newFirstIndex, pos, max int, forceUpdate bool) {
	// 重置所有更新标记
	l.ResetAllUpdateFlags()

	// 计算结束索引 - 对应 TypeScript 版本行1409
	curY := pos
	curIndex := newFirstIndex
	for curIndex < l.realNumItems && curY < max {
		ii := l.virtualItems[curIndex]
		if ii == nil {
			ii = &ItemInfo{}
			l.virtualItems[curIndex] = ii
		}
		if ii.height == 0 && l.itemSize.Y > 0 {
			ii.height = int(l.itemSize.Y)
		}
		curY += ii.height + l.lineGap
		curIndex++
	}
	newLastIndex := curIndex - 1
	if newLastIndex > l.realNumItems-1 {
		newLastIndex = l.realNumItems - 1
	}

	// 初始化位置 - 对应 TypeScript 版本行1398
	curX := 0
	curY = pos

	log.Printf("   updateVirtualItems1: newFirstIndex=%d, newLastIndex=%d, pos=%d, max=%d",
		newFirstIndex, newLastIndex, pos, max)
	log.Printf("   初始位置: curX=%d, curY=%d", curX, curY)

	// 更新可见项目
	itemCount := 0
	for curIndex = newFirstIndex; curIndex <= newLastIndex && curIndex < l.realNumItems; curIndex++ {
		ii := l.virtualItems[curIndex]
		if ii == nil {
			ii = &ItemInfo{}
			l.virtualItems[curIndex] = ii
		}

		needRender := false

		// 检查是否需要创建新对象
		if ii.obj == nil {
			needRender = true

			// 获取项目URL
			url := l.defaultItem
			if l.itemProvider != nil {
				providedURL := l.itemProvider(curIndex % l.numItems)
				if providedURL != "" {
					url = providedURL
				}
			}

			// 从对象池获取对象
			if l.pool != nil {
				ii.obj = l.pool.GetObject(url)
			}

			// 如果池中没有，创建新对象
			if ii.obj == nil {
				// 使用对象创建器创建对象
				if l.creator != nil {
					ii.obj = l.creator.CreateObject(url)
				}

				if ii.obj == nil {
					// 无法创建对象，跳过此项
					continue
				}
			}

			// 添加为子对象
			if ii.obj != nil {
				l.GComponent.AddChild(ii.obj)
				l.items = append(l.items, ii.obj)
				l.attachItemClick(ii.obj)
			}
		} else if forceUpdate {
			needRender = true
		}

		// 渲染项目
		if needRender && l.itemRenderer != nil {
			l.itemRenderer(curIndex%l.numItems, ii.obj)

			// 更新尺寸信息
			if ii.obj != nil {
				ii.width = int(ii.obj.Width())
				ii.height = int(ii.obj.Height())
			}
		}

		// 标记为已更新
		l.MarkItemUpdated(ii)

		// 设置位置 - 对应 TypeScript 版本行1494
		if ii.obj != nil {
			ii.obj.SetPosition(float64(curX), float64(curY))
			if itemCount < 5 {
				log.Printf("      item[%d]: pos=(%d,%d), size=%dx%d", curIndex, curX, curY, ii.width, ii.height)
			}
		}

		itemCount++

		// 更新位置 - 对应 TypeScript 版本行1498-1503
		curX += ii.width + l.columnGap

		if curIndex%l.curLineItemCount == l.curLineItemCount-1 {
			curX = 0
			curY += ii.height + l.lineGap
		}
	}

	log.Printf("   updateVirtualItems1 完成: 创建了%d个子对象", itemCount)

	// 清理未使用的对象
	for i := 0; i < len(l.virtualItems); i++ {
		ii := l.virtualItems[i]
		if ii != nil && !l.IsItemUpdated(ii) && ii.obj != nil {
			// 保存选择状态
			if button, ok := ii.obj.Data().(*GButton); ok {
				ii.selected = button.Selected()
			}

			// 从items数组中移除对象
			for j, item := range l.items {
				if item == ii.obj {
					l.items = append(l.items[:j], l.items[j+1:]...)
					break
				}
			}

			// 返回对象池
			if l.pool != nil {
				l.pool.ReturnObject(ii.obj)
			}

			// 从父对象移除
			l.GComponent.RemoveChild(ii.obj)
			ii.obj = nil
		}
	}
}

// updateVirtualItems 更新虚拟项（旧版本，保留用于其他布局）
func (l *GList) updateVirtualItems(oldFirstIndex, newFirstIndex, newLastIndex int, forward bool, forceUpdate bool) {
	// 重置所有更新标记
	l.ResetAllUpdateFlags()

	// 处理重用索引
	reuseIndex := oldFirstIndex
	if forward {
		reuseIndex = oldFirstIndex + (newLastIndex - newFirstIndex + 1)
	}

	// 更新可见项目
	curX, curY := 0, 0
	max := 0

	for curIndex := newFirstIndex; curIndex <= newLastIndex; curIndex++ {
		ii := l.virtualItems[curIndex]
		if ii == nil {
			ii = &ItemInfo{}
			l.virtualItems[curIndex] = ii
		}

		needRender := false

		// 检查是否需要创建新对象
		if ii.obj == nil {
			needRender = true

			// 获取项目URL
			url := l.defaultItem
			if l.itemProvider != nil {
				providedURL := l.itemProvider(curIndex % l.numItems)
				if providedURL != "" {
					url = providedURL
				}
			}

			// 从对象池获取对象
			if l.pool != nil {
				ii.obj = l.pool.GetObject(url)
			}

			// 如果池中没有，创建新对象
			if ii.obj == nil {
				// 使用对象创建器创建对象
				if l.creator != nil {
					ii.obj = l.creator.CreateObject(url)
				}

				if ii.obj == nil {
					// 无法创建对象，跳过此项
					continue
				}
			}

			// 添加为子对象 - 这里是问题所在！
			// 虚拟列表创建的对象必须正确添加到items数组中
			if ii.obj != nil {
				l.GComponent.AddChild(ii.obj)
				// 关键修复：将对象添加到items数组并附加点击事件
				l.items = append(l.items, ii.obj)
				l.attachItemClick(ii.obj)
			}
		} else if forceUpdate {
			needRender = true
		}

		// 渲染项目
		if needRender && l.itemRenderer != nil {
			l.itemRenderer(curIndex%l.numItems, ii.obj)

			// 更新尺寸信息
			if ii.obj != nil {
				ii.width = int(ii.obj.Width())
				ii.height = int(ii.obj.Height())
			}
		}

		// 标记为已更新
		l.MarkItemUpdated(ii)

		// 设置位置
		if ii.obj != nil {
			ii.obj.SetPosition(float64(curX), float64(curY))
		}

		// 更新位置
		if l.layout == ListLayoutTypeSingleColumn || l.layout == ListLayoutTypeFlowHorizontal {
			// 垂直布局
			curX += ii.width + l.columnGap
			if curIndex%l.curLineItemCount == l.curLineItemCount-1 {
				curX = 0
				curY += ii.height + l.lineGap
				if curIndex == newFirstIndex {
					max += ii.height
				}
			}
		} else {
			// 水平布局
			curY += ii.height + l.lineGap
			if curIndex%l.curLineItemCount == l.curLineItemCount-1 {
				curY = 0
				curX += ii.width + l.columnGap
				if curIndex == newFirstIndex {
					max += ii.width
				}
			}
		}
	}

	// 清理未使用的对象
	for i := reuseIndex; i < len(l.virtualItems); i++ {
		ii := l.virtualItems[i]
		if ii != nil && !l.IsItemUpdated(ii) && ii.obj != nil {
			// 保存选择状态
			if button, ok := ii.obj.Data().(*GButton); ok {
				ii.selected = button.Selected()
			}

			// 关键修复：从items数组中移除对象
			for j, item := range l.items {
				if item == ii.obj {
					// 移除并保持数组连续性
					l.items = append(l.items[:j], l.items[j+1:]...)
					break
				}
			}

			// 返回对象池
			if l.pool != nil {
				l.pool.ReturnObject(ii.obj)
			}

			// 从父对象移除
			l.GComponent.RemoveChild(ii.obj)
			ii.obj = nil
		}
	}
}

// handleArchOrder1 处理拱形顺序（垂直）
func (l *GList) handleArchOrder1() {
	if l.childrenOrder != ListChildrenRenderOrderArch {
		return
	}

	scrollPane := l.GComponent.ScrollPane()
	if scrollPane == nil {
		return
	}

	mid := int(scrollPane.PosY()) + l.getViewHeight()/2
	minDist := math.MaxInt32
	apexIndex := 0

	children := l.GComponent.Children()
	for i, child := range children {
		if child == nil {
			continue
		}

		if !l.foldInvisible || child.Visible() {
			dist := int(math.Abs(float64(mid) - child.Y() - child.Height()/2))
			if dist < minDist {
				minDist = dist
				apexIndex = i
			}
		}
	}

	l.apexIndex = apexIndex
}

// handleArchOrder2 处理拱形顺序（水平）
func (l *GList) handleArchOrder2() {
	if l.childrenOrder != ListChildrenRenderOrderArch {
		return
	}

	scrollPane := l.GComponent.ScrollPane()
	if scrollPane == nil {
		return
	}

	mid := int(scrollPane.PosX()) + l.getViewWidth()/2
	minDist := math.MaxInt32
	apexIndex := 0

	children := l.GComponent.Children()
	for i, child := range children {
		if child == nil {
			continue
		}

		if !l.foldInvisible || child.Visible() {
			dist := int(math.Abs(float64(mid) - child.X() - child.Width()/2))
			if dist < minDist {
				minDist = dist
				apexIndex = i
			}
		}
	}

	l.apexIndex = apexIndex
}

// getViewWidth 获取视图宽度
func (l *GList) getViewWidth() int {
	if scrollPane := l.GComponent.ScrollPane(); scrollPane != nil {
		return int(scrollPane.ViewWidth())
	}
	return int(l.GComponent.Width())
}

// getViewHeight 获取视图高度
func (l *GList) getViewHeight() int {
	if scrollPane := l.GComponent.ScrollPane(); scrollPane != nil {
		return int(scrollPane.ViewHeight())
	}
	return int(l.GComponent.Height())
}

// refreshVirtualList 公共方法，供外部调用
func (l *GList) refreshVirtualListPublic() {
	l.refreshVirtualList()
}

// handleScrollPublic 公共方法，供外部调用
func (l *GList) handleScrollPublic(forceUpdate bool) {
	l.handleScroll(forceUpdate)
}