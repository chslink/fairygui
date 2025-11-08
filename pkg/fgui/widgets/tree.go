package widgets

import (
	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// GTree 实现基于 GList 的树形控件，支持节点展开、插入与选择。
type GTree struct {
	*GList
	indent                int
	clickToExpand         int
	rootNode              *GTreeNode
	treeNodeRender        TreeNodeRender
	treeNodeWillExpand    TreeNodeWillExpand
	expandedStatusInEvent bool
}

const defaultTreeIndent = 15

// NewTree 创建空树。
func NewTree() *GTree {
	list := NewList()
	tree := &GTree{
		GList:         list,
		indent:        defaultTreeIndent,
		clickToExpand: 0,
	}
	tree.rootNode = NewTreeNode(true, "")
	tree.rootNode.level = 0
	tree.rootNode.expanded = true
	tree.rootNode.setTree(tree)
	return tree
}

// SetupBeforeAdd 解析树控件配置。
func (t *GTree) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if t == nil {
		return
	}
	// 首先调用父类 GList 处理列表和基础属性
	if t.GList != nil {
		t.GList.SetupBeforeAdd(buf, beginPos)
	}
	if buf == nil {
		return
	}
	// 然后处理 GTree 特定属性（block 9）
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(beginPos, 9) {
		return
	}
	if buf.Remaining() >= 4 {
		t.indent = int(buf.ReadInt32())
	}
	if buf.Remaining() > 0 {
		t.clickToExpand = int(buf.ReadByte())
	}
}

// RootNode 返回树的根节点。
func (t *GTree) RootNode() *GTreeNode {
	if t == nil {
		return nil
	}
	return t.rootNode
}

// Indent 返回当前缩进宽度。
func (t *GTree) Indent() int {
	if t == nil {
		return defaultTreeIndent
	}
	return t.indent
}

// SetIndent 设置缩进宽度并刷新可见节点。
func (t *GTree) SetIndent(value int) {
	if t == nil {
		return
	}
	if value < 0 {
		value = 0
	}
	if t.indent == value {
		return
	}
	t.indent = value
	t.refreshVisible()
}

// ClickToExpand 返回点击展开模式。
func (t *GTree) ClickToExpand() int {
	if t == nil {
		return 0
	}
	return t.clickToExpand
}

// SetClickToExpand 设置点击展开模式（0=禁用，1=单击切换，2=双击）。
func (t *GTree) SetClickToExpand(value int) {
	if t == nil {
		return
	}
	t.clickToExpand = value
}

// SetTreeNodeRender 注册节点渲染回调。
func (t *GTree) SetTreeNodeRender(fn TreeNodeRender) {
	if t == nil {
		return
	}
	t.treeNodeRender = fn
}

// SetTreeNodeWillExpand 注册节点展开通知回调。
func (t *GTree) SetTreeNodeWillExpand(fn TreeNodeWillExpand) {
	if t == nil {
		return
	}
	t.treeNodeWillExpand = fn
}

// GetSelectedNode 返回首个选中节点。
func (t *GTree) GetSelectedNode() *GTreeNode {
	if t == nil {
		return nil
	}
	index := t.SelectedIndex()
	if index < 0 {
		return nil
	}
	return t.nodeFromIndex(index)
}

// GetSelectedNodes 返回所有选中节点。
func (t *GTree) GetSelectedNodes() []*GTreeNode {
	if t == nil {
		return nil
	}
	indices := t.SelectedIndices()
	if len(indices) == 0 {
		return nil
	}
	nodes := make([]*GTreeNode, 0, len(indices))
	for _, idx := range indices {
		if node := t.nodeFromIndex(idx); node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// SelectNode 选中指定节点。
func (t *GTree) SelectNode(node *GTreeNode, scrollToView bool) {
	if t == nil || node == nil {
		return
	}
	for parent := node.parent; parent != nil && parent != t.rootNode; parent = parent.parent {
		parent.SetExpanded(true)
	}
	if node.cell == nil {
		t.ensureNodeCell(node)
	}
	index := t.indexOfNode(node)
	if index >= 0 {
		t.SetSelectedIndex(index)
	}
}

// UnselectNode 取消节点选中。
func (t *GTree) UnselectNode(node *GTreeNode) {
	if t == nil || node == nil {
		return
	}
	if node.cell == nil {
		return
	}
	if idx := t.indexOfNode(node); idx >= 0 {
		t.RemoveSelection(idx)
	}
}

// ExpandAll 展开所有节点。
func (t *GTree) ExpandAll(folder *GTreeNode) {
	if t == nil {
		return
	}
	if folder == nil {
		folder = t.rootNode
	}
	if folder == nil {
		return
	}
	if folder != t.rootNode {
		folder.SetExpanded(true)
	}
	for _, child := range folder.children {
		if child != nil && child.IsFolder() {
			t.ExpandAll(child)
		}
	}
}

// CollapseAll 折叠所有节点。
func (t *GTree) CollapseAll(folder *GTreeNode) {
	if t == nil {
		return
	}
	if folder == nil {
		folder = t.rootNode
	}
	if folder == nil {
		return
	}
	if folder != t.rootNode {
		folder.SetExpanded(false)
	}
	for _, child := range folder.children {
		if child != nil && child.IsFolder() {
			t.CollapseAll(child)
		}
	}
}

func (t *GTree) afterInserted(node *GTreeNode) {
	if t == nil || node == nil || node == t.rootNode {
		return
	}
	cell := t.ensureNodeCell(node)
	if cell == nil {
		return
	}
	index := t.getInsertIndexForNode(node)
	obj := cell.GObject
	currentIdx := t.indexOfNode(node)
	if currentIdx >= 0 {
		if currentIdx != index {
			t.RemoveItemAt(currentIdx)
			if currentIdx < index {
				index--
			}
			t.InsertItemAt(obj, index)
		}
	} else {
		t.InsertItemAt(obj, index)
	}
	t.attachNodeEvents(node)
	t.renderNode(node)
	if node.IsFolder() && node.Expanded() {
		t.checkChildren(node, index)
	}
}

func (t *GTree) afterRemoved(node *GTreeNode) {
	if t == nil || node == nil {
		return
	}
	t.removeNode(node)
}

func (t *GTree) afterExpanded(node *GTreeNode) {
	if t == nil || node == nil {
		return
	}
	if t.treeNodeWillExpand != nil {
		t.treeNodeWillExpand(node, true)
	}
	t.renderNode(node)
	if node == t.rootNode {
		t.checkChildren(node, -1)
		return
	}
	if node.cell == nil {
		return
	}
	index := t.indexOfNode(node)
	t.checkChildren(node, index)
}

func (t *GTree) afterCollapsed(node *GTreeNode) {
	if t == nil || node == nil {
		return
	}
	if t.treeNodeWillExpand != nil {
		t.treeNodeWillExpand(node, false)
	}
	t.renderNode(node)
	if node != t.rootNode {
		t.hideFolderNode(node)
	}
}

func (t *GTree) afterMoved(node *GTreeNode) {
	if t == nil || node == nil {
		return
	}
	start := t.indexOfNode(node)
	if start == -1 {
		return
	}
	var end int
	if node.IsFolder() {
		end = t.getFolderEndIndex(start, node.Level())
	} else {
		end = start + 1
	}
	insertIndex := t.getInsertIndexForNode(node)
	if insertIndex < start {
		for i := start; i < end; i++ {
			obj := t.items[i]
			t.RemoveItemAt(i)
			t.InsertItemAt(obj, insertIndex+(i-start))
		}
	} else {
		for i := end - 1; i >= start; i-- {
			obj := t.items[i]
			t.RemoveItemAt(i)
			t.InsertItemAt(obj, insertIndex+(i-start)-(end-start))
		}
	}
}

func (t *GTree) ensureNodeCell(node *GTreeNode) *core.GComponent {
	if node == nil || node == t.rootNode {
		return nil
	}
	if node.cell != nil {
		return node.cell
	}
	if node.cell == nil {
		comp := core.NewGComponent()
		comp.GObject.SetData(node)
		node.SetCell(comp)
	}
	return node.cell
}

func (t *GTree) renderNode(node *GTreeNode) {
	if node == nil || node.cell == nil {
		return
	}
	if indent := node.level - 1; indent >= 0 {
		if indentObj := node.cell.ChildByName("indent"); indentObj != nil {
			indentObj.SetSize(float64(indent*t.indent), indentObj.Height())
		}
	}
	if ctrl := node.cell.ControllerByName("expanded"); ctrl != nil {
		if node.Expanded() {
			ctrl.SetSelectedIndex(1)
		} else {
			ctrl.SetSelectedIndex(0)
		}
	}
	if ctrl := node.cell.ControllerByName("leaf"); ctrl != nil {
		if node.IsFolder() {
			ctrl.SetSelectedIndex(0)
		} else {
			ctrl.SetSelectedIndex(1)
		}
	}
	if t.treeNodeRender != nil {
		t.treeNodeRender(node, node.cell)
		return
	}
	data := node.cell.GObject.Data()
	switch v := data.(type) {
	case *GButton:
		v.SetTitle(node.text)
		v.SetIcon(node.icon)
	case *GLabel:
		v.SetTitle(node.text)
		v.SetIcon(node.icon)
	case interface{ SetTitle(string) }:
		v.SetTitle(node.text)
	case interface{ SetText(string) }:
		v.SetText(node.text)
	}
}

func (t *GTree) getInsertIndexForNode(node *GTreeNode) int {
	if node == nil {
		return 0
	}
	prev := node.GetPrevSibling()
	if prev == nil {
		prev = node.parent
	}
	insertIndex := -1
	if prev != nil {
		insertIndex = t.indexOfNode(prev)
	}
	insertIndex++
	myLevel := node.Level()
	for i := insertIndex; i < len(t.items); i++ {
		test := t.nodeFromIndex(i)
		if test == nil {
			continue
		}
		if test.Level() <= myLevel {
			break
		}
		insertIndex++
	}
	if insertIndex < 0 {
		insertIndex = 0
	}
	if insertIndex > len(t.items) {
		insertIndex = len(t.items)
	}
	return insertIndex
}

func (t *GTree) checkChildren(folder *GTreeNode, index int) int {
	if folder == nil || !folder.IsFolder() {
		return index
	}
	for _, child := range folder.children {
		if child == nil {
			continue
		}
		index++
		cell := t.ensureNodeCell(child)
		if cell != nil && t.indexOfNode(child) == -1 {
			t.InsertItemAt(cell.GObject, index)
			t.attachNodeEvents(child)
			t.renderNode(child)
		}
		if child.IsFolder() && child.Expanded() {
			index = t.checkChildren(child, index)
		}
	}
	return index
}

func (t *GTree) hideFolderNode(folder *GTreeNode) {
	if folder == nil || !folder.IsFolder() {
		return
	}
	for _, child := range folder.children {
		if child == nil {
			continue
		}
		if child.cell != nil {
			if idx := t.indexOfNode(child); idx >= 0 {
				t.RemoveItemAt(idx)
			}
		}
		if child.IsFolder() && child.Expanded() {
			t.hideFolderNode(child)
		}
	}
}

func (t *GTree) removeNode(node *GTreeNode) {
	if node == nil {
		return
	}
	if node.cell != nil {
		obj := node.cell.GObject
		if node.clickFn != nil {
			obj.Off(laya.EventClick, node.clickFn)
			node.clickFn = nil
		}
		if node.mouseDown != nil {
			obj.Off(laya.EventMouseDown, node.mouseDown)
			node.mouseDown = nil
		}
		if idx := t.indexOfNode(node); idx >= 0 {
			t.RemoveItemAt(idx)
		}
	}
	if node.IsFolder() {
		for _, child := range node.children {
			t.removeNode(child)
		}
	}
}

func (t *GTree) indexOfNode(node *GTreeNode) int {
	if node == nil || node.cell == nil {
		return -1
	}
	return t.indexOf(node.cell.GObject)
}

func (t *GTree) nodeFromIndex(index int) *GTreeNode {
	if index < 0 || index >= len(t.items) {
		return nil
	}
	item := t.items[index]
	if item == nil {
		return nil
	}
	if node, ok := item.Data().(*GTreeNode); ok {
		return node
	}
	return nil
}

func (t *GTree) attachNodeEvents(node *GTreeNode) {
	if node == nil || node.cell == nil {
		return
	}
	obj := node.cell.GObject
	if node.mouseDown != nil {
		obj.Off(laya.EventMouseDown, node.mouseDown)
	}
	mouseDown := func(evt *laya.Event) {
		t.expandedStatusInEvent = node.Expanded()
	}
	node.mouseDown = mouseDown
	obj.On(laya.EventMouseDown, mouseDown)

	if node.clickFn != nil {
		obj.Off(laya.EventClick, node.clickFn)
	}
	click := func(evt *laya.Event) {
		t.handleCellClick(node)
	}
	node.clickFn = click
	obj.On(laya.EventClick, click)
}

func (t *GTree) handleCellClick(node *GTreeNode) {
	if t == nil || node == nil || !node.IsFolder() {
		return
	}
	if t.clickToExpand == 0 {
		return
	}
	if node.Expanded() != t.expandedStatusInEvent {
		return
	}
	switch t.clickToExpand {
	case 1:
		node.SetExpanded(!node.Expanded())
	case 2:
		node.SetExpanded(!node.Expanded())
	}
}

func (t *GTree) getFolderEndIndex(startIndex, level int) int {
	for i := startIndex + 1; i < len(t.items); i++ {
		test := t.nodeFromIndex(i)
		if test == nil {
			continue
		}
		if test.Level() <= level {
			return i
		}
	}
	return len(t.items)
}

func (t *GTree) refreshVisible() {
	for _, item := range t.items {
		if item == nil {
			continue
		}
		if node, ok := item.Data().(*GTreeNode); ok {
			t.renderNode(node)
		}
	}
}
