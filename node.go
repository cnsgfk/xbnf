package xbnf

import (
	"fmt"
	"strings"
)

type NodeTreeConfig struct {
	PrintRuleType        bool
	PrintNonleafNodeText bool
	VerboseNodeText      bool
}

func DefaultNodeTreeConfig() *NodeTreeConfig {
	return &NodeTreeConfig{
		PrintRuleType:        true,
		VerboseNodeText:      false, // used only when PrintNonleafNodeText is true
		PrintNonleafNodeText: false,
	}
}

// Node is the tree node in a AST tree
type Node struct {
	RuleType   Type
	RuleName   string
	Sticky     bool
	Virtual    bool
	NonData    bool
	Chars      []rune
	ChildNodes []*Node
}

func (inst *Node) CountNodes() int {
	if inst == nil {
		return 0
	}
	count := 1
	for _, child := range inst.ChildNodes {
		count = count + child.CountNodes()
	}
	return count
}

func (inst *Node) CountTokens() int {
	if inst == nil {
		return 0
	}
	if len(inst.ChildNodes) == 0 {
		return 1
	}
	count := 0
	for _, child := range inst.ChildNodes {
		count = count + child.CountTokens()
	}
	return count
}

func (inst *Node) CountStickyNodes() int {
	if inst == nil {
		return 0
	}
	count := 0
	if inst.Sticky {
		count++
	}
	for _, child := range inst.ChildNodes {
		count = count + child.CountStickyNodes()
	}
	return count
}

// Remove virtual nodes in the tree rooted at this node
func (inst *Node) RemoveVirtualNodes() {
	if len(inst.ChildNodes) == 0 {
		return
	}
	var nonVirtual []*Node
	for _, node := range inst.ChildNodes {
		if node.Virtual {
			continue
		}
		node.RemoveVirtualNodes()
		nonVirtual = append(nonVirtual, node)
	}
	inst.ChildNodes = nonVirtual
}

// MergeStickyNodes should be the 1st step to simplify an AST. It merge adjacent sticky nodes into
// one single node. This operation is basically a tokenization operation.
func (inst *Node) MergeStickyNodes() {
	childrenCount := len(inst.ChildNodes)
	if childrenCount == 0 { // Do nothing on leaf node
		return
	}

	// for block node, we don't want to merge the open/close with content even when
	// they are all sticky. We want to always keep open/content/close node separated
	if inst.RuleType == TypeBlock {
		for _, node := range inst.ChildNodes {
			node.MergeStickyNodes()
			node.Sticky = false
		}
		inst.Sticky = false
		return
	}

	inst.ChildNodes = MergeStickyNodes(inst.ChildNodes)
}

// MergeStickyNodes merges nodes in a slice of nodes if a node is sticky and its previous sibling is
// also sticky, the node will be dropped and its text will be merged into previous sibling's text.
func MergeStickyNodes(nodes []*Node) []*Node {
	if len(nodes) == 0 {
		// Do nothing on empty input
		return nodes
	}

	/**
	// what's this for??
	if len(nodes) == 1 {
		if nodes[0].RuleName != "" {
			nodes[0].MergeStickyNodes()
		} else {
			nodes[0].Chars = nodes[0].Text()
			nodes[0].ChildNodes = nil
		}
		nodes[0].Sticky = false
		return nodes
	}
	*/

	var children []*Node
	var prevNode *Node
	for _, node := range nodes {
		if node.Sticky {
			// the node is sticky node that can be merged to previous sticky node if there is one
			nodeText := node.Text()
			if prevNode != nil && prevNode.Sticky {
				// previous node is steaky, node and prevNode can be merge
				prevNode.Chars = append(prevNode.Chars, nodeText...)
				continue
			}
			// all children of a sticky node can be drop, we only need to keep the text
			node.Chars = nodeText
			node.ChildNodes = nil
			children = append(children, node)
			prevNode = node
			continue
		}
		// now node is non-stick, can't be merged to previous sibling. we need to change the
		// prevNode to non-sticky, in case that this node is a virtual node and is removed in
		// RemoveVirtual(), its left and right sibling will be combined unintentionally.
		if prevNode != nil {
			prevNode.Sticky = false
		}
		prevNode = nil
		node.MergeStickyNodes() // do merge on child node first
		children = append(children, node)
	}
	if prevNode != nil {
		prevNode.Sticky = false
	}
	return children
}

func (inst *Node) RemoveNonDataNodes() {
	if len(inst.ChildNodes) == 0 {
		return
	}
	var dataNodes []*Node
	for _, node := range inst.ChildNodes {
		if node.NonData {
			continue
		}
		node.RemoveNonDataNodes()
		dataNodes = append(dataNodes, node)
	}
	inst.ChildNodes = dataNodes
}

// RemoveRedundantNodes removes redundant nodes in the tree rooted at this node. A node is redundant when:
// 1. It doesn't have a name,
// 2. It has only one child
func (inst *Node) RemoveRedundantNodes() {
	if len(inst.ChildNodes) == 0 {
		// Do nothing on leaf node
		return
	}

	for _, child := range inst.ChildNodes {
		child.RemoveRedundantNodes()
	}

	// has only 1 child node and the only child node does NOT have a ruleName, premote the
	// childNodes of the child node as direct children of this node. ie, grandchildren become children
	if len(inst.ChildNodes) == 1 && inst.ChildNodes[0].RuleName == "" {
		// this node only have 1 child, and the child doesn't have a name, we can promote its
		// text (leaf) or children to this node
		onlyChild := inst.ChildNodes[0]
		if len(onlyChild.Chars) > 0 { // leaf node
			inst.Chars = onlyChild.Chars
			inst.ChildNodes = nil
		} else { // intermediate node that can be dropped after all its children promoted
			inst.ChildNodes = onlyChild.ChildNodes
		}
	}

	var children []*Node
	// if a child does not have ruleName, and it's not leaf node (ie, has child nodes), premote
	// its children to this node's children. ie. grandchildren becomes children
	for _, child := range inst.ChildNodes { // range over slice in golang is order guaranteed
		// remove the child if it is empty and unnamed after virtual node removed
		if len(child.Text()) == 0 && child.RuleName == "" {
			continue
		}

		if child.RuleName == "" && len(child.ChildNodes) > 0 {
			children = append(children, child.ChildNodes...)
		} else {
			children = append(children, child) // this child is leaf node
		}
	}
	inst.ChildNodes = children

	// remove child if just only 1 child, and it's unamed, and it's a leaf node
	if len(inst.ChildNodes) == 1 && inst.ChildNodes[0].RuleName == "" && len(inst.ChildNodes[0].ChildNodes) == 0 {
		inst.Chars = inst.ChildNodes[0].Chars
		inst.ChildNodes = nil
	}
}

// The input config must not be nil, or this call will panic
func (inst *Node) header(config *NodeTreeConfig) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%s", inst.RuleName))
	if config.PrintRuleType {
		buf.WriteString(fmt.Sprintf("/%s", inst.RuleType))
	}
	if inst.Virtual {
		buf.WriteRune(VirtualSymbol)
	}
	if inst.NonData {
		buf.WriteRune(NonDataSymbol)
	}
	if inst.Sticky {
		buf.WriteRune(StickySymbol)
	}
	return buf.String()
}

func (inst *Node) StringTree(config *NodeTreeConfig) string {
	if config == nil {
		config = DefaultNodeTreeConfig()
	}
	var buf strings.Builder
	if len(inst.ChildNodes) == 0 || config.PrintNonleafNodeText {
		text := inst.Text()
		//if !config.VerboseNodeText && len(text) > 40 {
		//	text = text[:40]
		//	text = append(text, []rune("...")...)
		//}
		buf.WriteString(fmt.Sprintf("%s: >%s<", inst.header(config), string(text)))
	} else {
		buf.WriteString(fmt.Sprintf("%s", inst.header(config)))
	}
	if len(inst.ChildNodes) > 0 {
		// these are node types that have child nodes
		switch inst.RuleType {
		case TypeGroup, TypeOption, TypeBlock, TypeRepetition, TypeChoice, TypeConcatenate, TypeEmbed:
			size := len(inst.ChildNodes)
			switch size {
			case 0:
			case 1:
				nodeStr := inst.ChildNodes[0].StringTree(config)
				nodeStr = strings.ReplaceAll(nodeStr, "\n", "\n   ")
				buf.WriteString(fmt.Sprintf("\n└──%s", nodeStr))
			default:
				for _, node := range inst.ChildNodes[0 : size-1] {
					nodeStr := node.StringTree(config)
					nodeStr = strings.ReplaceAll(nodeStr, "\n", "\n│  ")
					buf.WriteString(fmt.Sprintf("\n├──%s", nodeStr))
				}
				nodeStr := inst.ChildNodes[size-1].StringTree(config)
				nodeStr = strings.ReplaceAll(nodeStr, "\n", "\n   ")
				buf.WriteString(fmt.Sprintf("\n└──%s", nodeStr))
			}
		}
	}
	return buf.String()
}

func (inst *Node) Text() []rune {
	if len(inst.Chars) > 0 {
		// this is a leaf node may be a product of MergeLeaf ...
		if len(inst.Chars) == 1 && inst.Chars[0] == EOFChar {
			return nil
		}
		return inst.Chars
	}
	var text []rune
	var prevNode *Node
	var prevText []rune
	for _, node := range inst.ChildNodes {
		nodeText := node.Text()
		if inst.RuleType != TypeBlock && // for block node, we don't want to insert a space between its open/content/close nodes
			len(prevText) > 0 && !IsWhiteSpace(prevText[len(prevText)-1]) && // if the proceeding char is a whitespace, we don't need to add a space
			len(nodeText) > 0 && !IsWhiteSpace(nodeText[0]) && // if the 1st char is a white space, we don't need to add a space
			(!prevNode.Sticky || !node.Sticky) {
			text = append(text, ' ')
		}
		text = append(text, nodeText...)
		prevText = text
		prevNode = node
	}
	return text
}
