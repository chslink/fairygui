package fairygui

import (
	"sync"
)

// ============================================================================
// TreeNode - 树节点
// ============================================================================

type TreeNode struct {
	// 节点数据
	data interface{}

	// 节点文本
	text string

	// 图标
	icon string

	// 子节点
	children []*TreeNode

	// 父节点
	parent *TreeNode

	// 展开状态
	expanded bool

	// 是否可选中
	selectable bool

	// 是否可见
	visible bool

	// 层级
	level int

	// 索引（在扁平化列表中的位置）
	index int

	// 内部状态
	mu sync.RWMutex
}

// NewTreeNode 创建新的树节点
func NewTreeNode(text string) *TreeNode {
	return &TreeNode{
		text:       text,
		children:   make([]*TreeNode, 0),
		expanded:   false,
		selectable: true,
		visible:    true,
		level:      0,
		index:      -1,
	}
}

// SetData 设置节点数据
func (n *TreeNode) SetData(data interface{}) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.data = data
}

// Data 返回节点数据
func (n *TreeNode) Data() interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.data
}

// SetText 设置节点文本
func (n *TreeNode) SetText(text string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.text = text
}

// Text 返回节点文本
func (n *TreeNode) Text() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.text
}

// SetIcon 设置节点图标
func (n *TreeNode) SetIcon(icon string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.icon = icon
}

// Icon 返回节点图标
func (n *TreeNode) Icon() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.icon
}

// SetExpanded 设置展开状态
func (n *TreeNode) SetExpanded(expanded bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.expanded = expanded
}

// IsExpanded 返回展开状态
func (n *TreeNode) IsExpanded() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.expanded
}

// SetSelectable 设置是否可选中
func (n *TreeNode) SetSelectable(selectable bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.selectable = selectable
}

// IsSelectable 返回是否可选中
func (n *TreeNode) IsSelectable() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.selectable
}

// SetVisible 设置是否可见
func (n *TreeNode) SetVisible(visible bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.visible = visible
}

// IsVisible 返回是否可见
func (n *TreeNode) IsVisible() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.visible
}

// Level 返回节点层级
func (n *TreeNode) Level() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.level
}

// SetParent 设置父节点
func (n *TreeNode) SetParent(parent *TreeNode) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.parent = parent
	if parent != nil {
		n.level = parent.level + 1
	} else {
		n.level = 0
	}
}

// Parent 返回父节点
func (n *TreeNode) Parent() *TreeNode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.parent
}

// AddChild 添加子节点
func (n *TreeNode) AddChild(child *TreeNode) {
	if child == nil {
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	child.SetParent(n)
	n.children = append(n.children, child)
}

// RemoveChild 移除子节点
func (n *TreeNode) RemoveChild(child *TreeNode) {
	if child == nil {
		return
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	for i, c := range n.children {
		if c == child {
			n.children = append(n.children[:i], n.children[i+1:]...)
			child.SetParent(nil)
			break
		}
	}
}

// RemoveChildAt 移除指定索引的子节点
func (n *TreeNode) RemoveChildAt(index int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if index >= 0 && index < len(n.children) {
		child := n.children[index]
		n.children = append(n.children[:index], n.children[index+1:]...)
		child.SetParent(nil)
	}
}

// RemoveChildren 移除所有子节点
func (n *TreeNode) RemoveChildren() {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, child := range n.children {
		child.SetParent(nil)
	}
	n.children = make([]*TreeNode, 0)
}

// Children 返回所有子节点
func (n *TreeNode) Children() []*TreeNode {
	n.mu.RLock()
	defer n.mu.RUnlock()

	result := make([]*TreeNode, len(n.children))
	copy(result, n.children)
	return result
}

// NumChildren 返回子节点数量
func (n *TreeNode) NumChildren() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.children)
}

// GetChildAt 获取指定索引的子节点
func (n *TreeNode) GetChildAt(index int) *TreeNode {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if index >= 0 && index < len(n.children) {
		return n.children[index]
	}
	return nil
}

// Expand 展开节点
func (n *TreeNode) Expand() {
	n.SetExpanded(true)
}

// Collapse 折叠节点
func (n *TreeNode) Collapse() {
	n.SetExpanded(false)
}

// Toggle 切换展开/折叠
func (n *TreeNode) Toggle() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.expanded = !n.expanded
}

// HasChildren 返回是否有子节点
func (n *TreeNode) HasChildren() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.children) > 0
}

// IsLeaf 返回是否是叶子节点
func (n *TreeNode) IsLeaf() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.children) == 0
}

// IndexOfChild 返回子节点的索引
func (n *TreeNode) IndexOfChild(child *TreeNode) int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for i, c := range n.children {
		if c == child {
			return i
		}
	}
	return -1
}

// Root 返回根节点
func (n *TreeNode) Root() *TreeNode {
	root := n
	for root.Parent() != nil {
		root = root.Parent()
	}
	return root
}

// IsAncestorOf 返回是否是另一个节点的祖先
func (n *TreeNode) IsAncestorOf(descendant *TreeNode) bool {
	for node := descendant; node != nil; node = node.Parent() {
		if node == n {
			return true
		}
	}
	return false
}

// IsDescendantOf 返回是否是另一个节点的后代
func (n *TreeNode) IsDescendantOf(ancestor *TreeNode) bool {
	return ancestor.IsAncestorOf(n)
}

// ============================================================================
// TreeView - 树形视图
// ============================================================================

type TreeView struct {
	*List

	// 树的根节点
	rootNode *TreeNode

	// 可见节点列表（扁平化）
	visibleNodes []*TreeNode

	// 列表项到节点的映射
	itemToNode map[*ComponentImpl]*TreeNode

	// 节点到列表项的映射
	nodeToItem map[*TreeNode]*ComponentImpl

	// 缩进宽度
	indentWidth float64

	// 是否显示图标
	showIcons bool

	// 是否显示箭头
	showArrows bool

	// 箭头图标
	arrowDownIcon string
	arrowRightIcon string

	// 展开/折叠事件
	onExpand EventHandler
	onCollapse EventHandler

	// 节点选择事件
	onNodeSelect EventHandler

	// 内部状态
	nodeMutex sync.RWMutex
}

// NewTreeView 创建新的树视图
func NewTreeView() *TreeView {
	tree := &TreeView{
		List:           NewList(),
		rootNode:       NewTreeNode("Root"),
		visibleNodes:   make([]*TreeNode, 0),
		itemToNode:     make(map[*ComponentImpl]*TreeNode),
		nodeToItem:     make(map[*TreeNode]*ComponentImpl),
		indentWidth:    20.0,
		showIcons:      true,
		showArrows:     true,
		arrowDownIcon:  "▼",
		arrowRightIcon: "▶",
	}

	// 设置 List 的 item renderer
	tree.List.itemRenderer = tree.renderNode

	// 隐藏根节点（默认）
	tree.rootNode.SetVisible(false)

	return tree
}

// ============================================================================
// 节点管理
// ============================================================================

// RootNode 返回根节点
func (tv *TreeView) RootNode() *TreeNode {
	tv.nodeMutex.RLock()
	defer tv.nodeMutex.RUnlock()
	return tv.rootNode
}

// SetRootNode 设置根节点
func (tv *TreeView) SetRootNode(node *TreeNode) {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()

	// 清除所有子节点
	tv.rootNode.RemoveChildren()
	tv.clearMappings()

	// 设置新的根节点
	if node != nil {
		tv.rootNode = node
		tv.rootNode.SetParent(nil)
	} else {
		tv.rootNode = NewTreeNode("Root")
	}

	tv.updateVisibleNodes()
}

// AddNode 添加节点
func (tv *TreeView) AddNode(parent *TreeNode, node *TreeNode) {
	if node == nil {
		return
	}

	if parent == nil {
		parent = tv.rootNode
	}

	parent.AddChild(node)
	tv.updateVisibleNodes()
}

// RemoveNode 移除节点
func (tv *TreeView) RemoveNode(node *TreeNode) {
	if node == nil || node == tv.rootNode {
		return
	}

	if parent := node.Parent(); parent != nil {
		parent.RemoveChild(node)
		tv.updateVisibleNodes()
	}
}

// RemoveNodeAt 移除指定位置的节点
func (tv *TreeView) RemoveNodeAt(parent *TreeNode, index int) {
	if parent == nil {
		parent = tv.rootNode
	}

	parent.RemoveChildAt(index)
	tv.updateVisibleNodes()
}

// ClearNodes 清除所有节点
func (tv *TreeView) ClearNodes() {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()

	tv.rootNode.RemoveChildren()
	tv.clearMappings()
	tv.visibleNodes = make([]*TreeNode, 0)
	tv.updateVisibleNodes()
}

// GetNodeAt 获取指定索引的节点
func (tv *TreeView) GetNodeAt(index int) *TreeNode {
	tv.nodeMutex.RLock()
	defer tv.nodeMutex.RUnlock()

	if index >= 0 && index < len(tv.visibleNodes) {
		return tv.visibleNodes[index]
	}
	return nil
}

// GetItemNode 获取列表项对应的节点
func (tv *TreeView) GetItemNode(item *ComponentImpl) *TreeNode {
	tv.nodeMutex.RLock()
	defer tv.nodeMutex.RUnlock()

	if node, ok := tv.itemToNode[item]; ok {
		return node
	}
	return nil
}

// GetNodeItem 获取节点对应的列表项
func (tv *TreeView) GetNodeItem(node *TreeNode) *ComponentImpl {
	tv.nodeMutex.RLock()
	defer tv.nodeMutex.RUnlock()

	if item, ok := tv.nodeToItem[node]; ok {
		return item
	}
	return nil
}

// ExpandNode 展开节点
func (tv *TreeView) ExpandNode(node *TreeNode) {
	if node == nil || !node.HasChildren() {
		return
	}

	node.Expand()
	tv.updateVisibleNodes()

	// 触发事件
	tv.emitExpand(node)
}

// CollapseNode 折叠节点
func (tv *TreeView) CollapseNode(node *TreeNode) {
	if node == nil {
		return
	}

	node.Collapse()
	tv.updateVisibleNodes()

	// 触发事件
	tv.emitCollapse(node)
}

// ToggleNode 切换节点展开/折叠
func (tv *TreeView) ToggleNode(node *TreeNode) {
	if node == nil || !node.HasChildren() {
		return
	}

	node.Toggle()
	tv.updateVisibleNodes()

	if node.IsExpanded() {
		tv.emitExpand(node)
	} else {
		tv.emitCollapse(node)
	}
}

// ExpandAll 展开所有节点
func (tv *TreeView) ExpandAll() {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()

	tv.expandAllNodes(tv.rootNode)
	tv.updateVisibleNodes()
}

// expandAllNodes 递归展开所有节点
func (tv *TreeView) expandAllNodes(node *TreeNode) {
	if node == nil {
		return
	}

	node.Expand()
	for _, child := range node.Children() {
		tv.expandAllNodes(child)
	}
}

// CollapseAll 折叠所有节点（除根节点外）
func (tv *TreeView) CollapseAll() {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()

	children := tv.rootNode.Children()
	for _, child := range children {
		tv.collapseAllNodes(child)
	}
	tv.updateVisibleNodes()
}

// collapseAllNodes 递归折叠所有节点
func (tv *TreeView) collapseAllNodes(node *TreeNode) {
	if node == nil {
		return
	}

	node.Collapse()
	for _, child := range node.Children() {
		tv.collapseAllNodes(child)
	}
}

// ============================================================================
// 缩进和图标
// ============================================================================

// SetIndentWidth 设置缩进宽度
func (tv *TreeView) SetIndentWidth(width float64) {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()
	tv.indentWidth = width
}

// IndentWidth 返回缩进宽度
func (tv *TreeView) IndentWidth() float64 {
	tv.nodeMutex.RLock()
	defer tv.nodeMutex.RUnlock()
	return tv.indentWidth
}

// SetShowIcons 设置是否显示图标
func (tv *TreeView) SetShowIcons(show bool) {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()
	tv.showIcons = show
}

// ShowIcons 返回是否显示图标
func (tv *TreeView) ShowIcons() bool {
	tv.nodeMutex.RLock()
	defer tv.nodeMutex.RUnlock()
	return tv.showIcons
}

// SetShowArrows 设置是否显示箭头
func (tv *TreeView) SetShowArrows(show bool) {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()
	tv.showArrows = show
}

// ShowArrows 返回是否显示箭头
func (tv *TreeView) ShowArrows() bool {
	tv.nodeMutex.RLock()
	defer tv.nodeMutex.RUnlock()
	return tv.showArrows
}

// SetArrowIcons 设置箭头图标
func (tv *TreeView) SetArrowIcons(downIcon, rightIcon string) {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()
	tv.arrowDownIcon = downIcon
	tv.arrowRightIcon = rightIcon
}

// ============================================================================
// 节点选择
// ============================================================================

// SetSelectedNode 设置选中的节点
func (tv *TreeView) SetSelectedNode(node *TreeNode) {
	if node == nil || !node.IsSelectable() {
		tv.List.SetSelectedIndex(-1)
		return
	}

	index := tv.getNodeIndex(node)
	if index >= 0 {
		tv.List.SetSelectedIndex(index)
		tv.emitNodeSelect(node)
	}
}

// GetSelectedNode 返回选中的节点
func (tv *TreeView) GetSelectedNode() *TreeNode {
	index := tv.List.SelectedIndex()
	if index < 0 {
		return nil
	}
	return tv.GetNodeAt(index)
}

// SelectNode 选择节点
func (tv *TreeView) SelectNode(node *TreeNode, scrollToView bool) {
	if node == nil || !node.IsSelectable() {
		return
	}

	index := tv.getNodeIndex(node)
	if index >= 0 {
		tv.List.AddSelection(index, scrollToView)
		tv.emitNodeSelect(node)
	}
}

// DeselectNode 取消选择节点
func (tv *TreeView) DeselectNode(node *TreeNode) {
	index := tv.getNodeIndex(node)
	if index >= 0 {
		tv.List.RemoveSelection(index)
	}
}

// getNodeIndex 获取节点在可见列表中的索引
func (tv *TreeView) getNodeIndex(node *TreeNode) int {
	tv.nodeMutex.RLock()
	defer tv.nodeMutex.RUnlock()

	for i, n := range tv.visibleNodes {
		if n == node {
			return i
		}
	}
	return -1
}

// ============================================================================
// 内部方法
// ============================================================================

// updateVisibleNodes 更新可见节点列表
func (tv *TreeView) updateVisibleNodes() {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()

	tv.visibleNodes = make([]*TreeNode, 0)
	result := &tv.visibleNodes
	tv.collectVisibleNodes(tv.rootNode, result)

	// 更新 List 的项数量
	tv.List.SetNumItems(len(tv.visibleNodes))
}

// collectVisibleNodes 收集可见节点（递归）
func (tv *TreeView) collectVisibleNodes(node *TreeNode, result *[]*TreeNode) {
	if node == nil || (!node.IsVisible() && node != tv.rootNode) {
		return
	}

	// 根节点通常不显示
	if node != tv.rootNode {
		*result = append(*result, node)
	}

	// 如果节点展开，添加子节点
	if node.IsExpanded() {
		for _, child := range node.Children() {
			tv.collectVisibleNodes(child, result)
		}
	}
}

// clearMappings 清除映射关系
func (tv *TreeView) clearMappings() {
	tv.itemToNode = make(map[*ComponentImpl]*TreeNode)
	tv.nodeToItem = make(map[*TreeNode]*ComponentImpl)
}

// renderNode 渲染节点（作为 List 的 itemRenderer）
func (tv *TreeView) renderNode(index int, item *ComponentImpl) {
	if index < 0 || index >= len(tv.visibleNodes) {
		return
	}

	node := tv.visibleNodes[index]

	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()

	// 建立映射
	tv.itemToNode[item] = node
	tv.nodeToItem[node] = item

	// 设置节点数据
	item.SetData(node)

	// 设置文本
	if tf, ok := item.GetChildByName("title").(*TextField); ok {
		tf.SetText(node.Text())
	}

	// 设置图标
	if tv.showIcons && node.Icon() != "" {
		if icon, ok := item.GetChildByName("icon").(*Image); ok {
			// 这里应该加载图标资源
			_ = icon
		}
	}

	// 设置箭头
	if tv.showArrows {
		if arrow, ok := item.GetChildByName("arrow").(*ComponentImpl); ok {
			if node.HasChildren() {
				if node.IsExpanded() {
					// 展开状态
					if tf, ok := arrow.GetChildByName("title").(*TextField); ok {
						tf.SetText(tv.arrowDownIcon)
					}
				} else {
					// 折叠状态
					if tf, ok := arrow.GetChildByName("title").(*TextField); ok {
						tf.SetText(tv.arrowRightIcon)
					}
				}
				arrow.SetVisible(true)
			} else {
				arrow.SetVisible(false)
			}
		}
	}

	// 设置缩进
	indent := float64(node.Level()) * tv.indentWidth
	_, currentY := item.Position()
	item.SetPosition(indent, currentY)
}

// handleNodeClick 处理节点点击
func (tv *TreeView) handleNodeClick(node *TreeNode) {
	if node == nil {
		return
	}

	if node.HasChildren() {
		tv.ToggleNode(node)
	}
}

// ============================================================================
// 事件处理
// ============================================================================

// OnExpand 设置展开事件
func (tv *TreeView) OnExpand(handler EventHandler) {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()
	tv.onExpand = handler
}

// OnCollapse 设置折叠事件
func (tv *TreeView) OnCollapse(handler EventHandler) {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()
	tv.onCollapse = handler
}

// OnNodeSelect 设置节点选择事件
func (tv *TreeView) OnNodeSelect(handler EventHandler) {
	tv.nodeMutex.Lock()
	defer tv.nodeMutex.Unlock()
	tv.onNodeSelect = handler
}

// emitExpand 触发展开事件
func (tv *TreeView) emitExpand(node *TreeNode) {
	tv.nodeMutex.RLock()
	handler := tv.onExpand
	tv.nodeMutex.RUnlock()

	if handler != nil {
		handler(NewUIEvent("expand", tv, node))
	}

	tv.Emit(NewUIEvent("node_expand", tv, node))
}

// emitCollapse 触发折叠事件
func (tv *TreeView) emitCollapse(node *TreeNode) {
	tv.nodeMutex.RLock()
	handler := tv.onCollapse
	tv.nodeMutex.RUnlock()

	if handler != nil {
		handler(NewUIEvent("collapse", tv, node))
	}

	tv.Emit(NewUIEvent("node_collapse", tv, node))
}

// emitNodeSelect 触发节点选择事件
func (tv *TreeView) emitNodeSelect(node *TreeNode) {
	tv.nodeMutex.RLock()
	handler := tv.onNodeSelect
	tv.nodeMutex.RUnlock()

	if handler != nil {
		handler(NewUIEvent("node_select", tv, node))
	}

	tv.Emit(NewUIEvent("node_select", tv, node))
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertTreeView 类型断言
func AssertTreeView(obj DisplayObject) (*TreeView, bool) {
	tree, ok := obj.(*TreeView)
	return tree, ok
}

// IsTreeView 检查是否是 TreeView
func IsTreeView(obj DisplayObject) bool {
	_, ok := obj.(*TreeView)
	return ok
}
