package widgets

import (
	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// TreeNodeRender describes a callback invoked when a tree node's cell needs updating.
type TreeNodeRender func(node *GTreeNode, cell *core.GComponent)

// TreeNodeWillExpand notifies listeners before a node toggles its expanded state.
type TreeNodeWillExpand func(node *GTreeNode, expanded bool)

// GTreeNode mirrors FairyGUI 的 GTreeNode，支持层级关系与展开状态。
type GTreeNode struct {
	data any

	parent    *GTreeNode
	children  []*GTreeNode
	isFolder  bool
	expanded  bool
	level     int
	tree      *GTree
	cell      *core.GComponent
	resURL    string
	text      string
	icon      string
	clickFn   laya.Listener
	mouseDown laya.Listener
}

// NewTreeNode 构造一颗树节点，hasChild 为 true 时初始化为文件夹节点。
func NewTreeNode(hasChild bool, resURL string) *GTreeNode {
	node := &GTreeNode{
		isFolder: hasChild,
		resURL:   resURL,
	}
	if hasChild {
		node.children = make([]*GTreeNode, 0)
	}
	return node
}

// IsFolder 返回节点是否为文件夹类型。
func (n *GTreeNode) IsFolder() bool {
	return n != nil && n.isFolder
}

// Expanded 返回节点是否处于展开状态。
func (n *GTreeNode) Expanded() bool {
	if n == nil {
		return false
	}
	return n.expanded
}

// SetExpanded 修改节点展开状态。
func (n *GTreeNode) SetExpanded(value bool) {
	if n == nil || !n.IsFolder() || n.expanded == value {
		return
	}
	n.expanded = value
	if n.tree != nil {
		if value {
			n.tree.afterExpanded(n)
		} else {
			n.tree.afterCollapsed(n)
		}
	}
}

// Parent 返回父节点。
func (n *GTreeNode) Parent() *GTreeNode {
	if n == nil {
		return nil
	}
	return n.parent
}

// Tree 返回所属的树。
func (n *GTreeNode) Tree() *GTree {
	if n == nil {
		return nil
	}
	return n.tree
}

// Level 返回节点层级（根节点为 0）。
func (n *GTreeNode) Level() int {
	if n == nil {
		return 0
	}
	return n.level
}

// Text 返回节点标题。
func (n *GTreeNode) Text() string {
	if n == nil {
		return ""
	}
	return n.text
}

// SetText 更新节点标题。
func (n *GTreeNode) SetText(value string) {
	if n == nil {
		return
	}
	if n.text == value {
		return
	}
	n.text = value
	if n.tree != nil {
		n.tree.renderNode(n)
	}
}

// Icon 返回节点图标 URL。
func (n *GTreeNode) Icon() string {
	if n == nil {
		return ""
	}
	return n.icon
}

// SetIcon 更新节点图标 URL。
func (n *GTreeNode) SetIcon(value string) {
	if n == nil {
		return
	}
	if n.icon == value {
		return
	}
	n.icon = value
	if n.tree != nil {
		n.tree.renderNode(n)
	}
}

// Data 返回节点绑定的数据对象。
func (n *GTreeNode) Data() any {
	if n == nil {
		return nil
	}
	return n.data
}

// SetData 绑定任意数据对象。
func (n *GTreeNode) SetData(value any) {
	if n == nil {
		return
	}
	n.data = value
}

// ResURL 返回节点的模板资源地址。
func (n *GTreeNode) ResURL() string {
	if n == nil {
		return ""
	}
	return n.resURL
}

// SetResURL 更新节点模板资源地址。
func (n *GTreeNode) SetResURL(value string) {
	if n == nil {
		return
	}
	n.resURL = value
}

// Cell 返回节点对应的组件实例。
func (n *GTreeNode) Cell() *core.GComponent {
	if n == nil {
		return nil
	}
	return n.cell
}

// SetCell 绑定节点的组件实例。
func (n *GTreeNode) SetCell(comp *core.GComponent) {
	if n == nil {
		return
	}
	n.cell = comp
	if comp != nil && comp.GObject != nil {
		comp.GObject.SetData(n)
	}
}

// NumChildren 返回子节点数量。
func (n *GTreeNode) NumChildren() int {
	if n == nil || len(n.children) == 0 {
		return 0
	}
	return len(n.children)
}

// Children 返回子节点切片的副本。
func (n *GTreeNode) Children() []*GTreeNode {
	if n == nil || len(n.children) == 0 {
		return nil
	}
	out := make([]*GTreeNode, len(n.children))
	copy(out, n.children)
	return out
}

// ChildAt 返回指定索引的子节点。
func (n *GTreeNode) ChildAt(index int) *GTreeNode {
	if n == nil || index < 0 || index >= len(n.children) {
		return nil
	}
	return n.children[index]
}

// GetChildIndex 返回子节点所在索引。
func (n *GTreeNode) GetChildIndex(child *GTreeNode) int {
	if n == nil || child == nil {
		return -1
	}
	for idx, current := range n.children {
		if current == child {
			return idx
		}
	}
	return -1
}

// AddChild 追加子节点。
func (n *GTreeNode) AddChild(child *GTreeNode) *GTreeNode {
	if n == nil {
		return nil
	}
	return n.AddChildAt(child, len(n.children))
}

// AddChildAt 在指定位置插入子节点。
func (n *GTreeNode) AddChildAt(child *GTreeNode, index int) *GTreeNode {
	if n == nil || child == nil {
		return nil
	}
	if child.parent == n {
		n.SetChildIndex(child, index)
		return child
	}
	if child.parent != nil {
		child.parent.RemoveChild(child)
	}
	if !n.isFolder {
		n.isFolder = true
		n.children = make([]*GTreeNode, 0)
	}
	if index < 0 || index > len(n.children) {
		index = len(n.children)
	}
	n.children = append(n.children, nil)
	copy(n.children[index+1:], n.children[index:])
	n.children[index] = child
	child.parent = n
	child.level = n.level + 1
	child.setTree(n.tree)
	if n.tree != nil && (n == n.tree.rootNode || (n.cell != nil && n.cell.GObject.Parent() != nil && n.expanded)) {
		n.tree.afterInserted(child)
	}
	return child
}

// RemoveChild 移除指定子节点。
func (n *GTreeNode) RemoveChild(child *GTreeNode) *GTreeNode {
	if n == nil || child == nil {
		return nil
	}
	if idx := n.GetChildIndex(child); idx != -1 {
		return n.RemoveChildAt(idx)
	}
	return nil
}

// RemoveChildAt 按索引移除子节点。
func (n *GTreeNode) RemoveChildAt(index int) *GTreeNode {
	if n == nil || index < 0 || index >= len(n.children) {
		return nil
	}
	child := n.children[index]
	copy(n.children[index:], n.children[index+1:])
	n.children = n.children[:len(n.children)-1]
	child.parent = nil
	child.setTree(nil)
	if n.tree != nil {
		n.tree.afterRemoved(child)
	}
	return child
}

// RemoveChildren 批量移除子节点。
func (n *GTreeNode) RemoveChildren(begin, end int) {
	if n == nil || begin < 0 || end < begin {
		return
	}
	if end >= len(n.children) {
		end = len(n.children) - 1
	}
	for i := end; i >= begin; i-- {
		n.RemoveChildAt(i)
	}
}

// SetChildIndex 调整子节点顺序。
func (n *GTreeNode) SetChildIndex(child *GTreeNode, index int) {
	if n == nil || child == nil {
		return
	}
	oldIndex := n.GetChildIndex(child)
	if oldIndex == -1 || oldIndex == index {
		return
	}
	if index < 0 {
		index = 0
	} else if index >= len(n.children) {
		index = len(n.children) - 1
	}
	copy(n.children[oldIndex:], n.children[oldIndex+1:])
	n.children = n.children[:len(n.children)-1]
	n.children = append(n.children, nil)
	copy(n.children[index+1:], n.children[index:])
	n.children[index] = child
	if n.tree != nil {
		n.tree.afterMoved(child)
	}
}

// GetPrevSibling 返回前一个兄弟节点。
func (n *GTreeNode) GetPrevSibling() *GTreeNode {
	if n == nil || n.parent == nil {
		return nil
	}
	idx := n.parent.GetChildIndex(n)
	if idx <= 0 {
		return nil
	}
	return n.parent.children[idx-1]
}

// GetNextSibling 返回后一个兄弟节点。
func (n *GTreeNode) GetNextSibling() *GTreeNode {
	if n == nil || n.parent == nil {
		return nil
	}
	idx := n.parent.GetChildIndex(n)
	if idx < 0 || idx >= n.parent.NumChildren()-1 {
		return nil
	}
	return n.parent.children[idx+1]
}

// ExpandToRoot 展开当前节点至根节点。
func (n *GTreeNode) ExpandToRoot() {
	for p := n; p != nil; p = p.parent {
		p.SetExpanded(true)
	}
}

func (n *GTreeNode) setTree(tree *GTree) {
	if n.tree == tree {
		return
	}
	n.tree = tree
	if tree != nil && n.expanded && tree.treeNodeWillExpand != nil {
		tree.treeNodeWillExpand(n, true)
	}
	if len(n.children) == 0 {
		return
	}
	for _, child := range n.children {
		if child == nil {
			continue
		}
		child.level = n.level + 1
		child.setTree(tree)
	}
}
