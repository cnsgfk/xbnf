package xbnf

import (
	"fmt"
	"strings"
)

const (
	LevelRaw       = -1 // raw ast tree without any simplification operation
	LevelBasic     = 0  // MergeStickyNodes performed
	LevelNoVertual = 1  // Vertual Nodes removed
	LevelDataOnly  = 2  // NonData nodes removed
)

type RuleName string

func (inst RuleName) Name() string {
	return string(inst)
}

func (inst RuleName) header(nodeType string) string {
	var buf strings.Builder
	if inst == "" {
		buf.WriteString(fmt.Sprintf("/%s", nodeType))
	} else {
		buf.WriteString(fmt.Sprintf("%s/%s", inst, nodeType))
	}
	return buf.String()
}

type Sticky bool

func (inst Sticky) isSticky() bool {
	return bool(inst)
}

type AST struct {
	Filename string // if available
	Nodes    []*Node
}

func (inst *AST) RemoveVirtualNodes() {
	if len(inst.Nodes) == 0 {
		return
	}
	var nonVirtual []*Node
	for _, node := range inst.Nodes {
		if node.Virtual {
			continue
		}
		node.RemoveVirtualNodes()
		nonVirtual = append(nonVirtual, node)
	}
	inst.Nodes = nonVirtual
}

func (inst *AST) RemoveNonDataNodes() {
	if len(inst.Nodes) == 0 {
		return
	}
	var dataNodes []*Node
	for _, node := range inst.Nodes {
		if node.NonData {
			continue
		}
		node.RemoveNonDataNodes()
		dataNodes = append(dataNodes, node)
	}
	inst.Nodes = dataNodes
}

func (inst *AST) RemoveRedundantNodes() {
	if len(inst.Nodes) == 0 {
		return
	}
	for _, node := range inst.Nodes {
		if node.NonData {
			continue
		}
		node.RemoveRedundantNodes()
	}
}
func (inst *AST) MergeStickyNodes() {
	if inst == nil {
		return
	}
	if len(inst.Nodes) == 0 {
		return
	}
	inst.Nodes = MergeStickyNodes(inst.Nodes)
}

func (inst *AST) StringTree(config *NodeTreeConfig) string {
	if inst == nil {
		return "nil AST"
	}
	var buf strings.Builder
	buf.WriteString("Abstract Syntax Tree")
	buf.WriteString(fmt.Sprintf("\n├─ file:%s", inst.Filename))
	size := len(inst.Nodes)
	switch size {
	case 0:
		buf.WriteString(fmt.Sprintf("\n└─ Nodes: %d", len(inst.Nodes)))
	case 1:
		buf.WriteString(fmt.Sprintf("\n├─ Nodes: %d", len(inst.Nodes)))
		nodeStr := inst.Nodes[0].StringTree(config)
		nodeStr = strings.ReplaceAll(nodeStr, "\n", "\n    ")
		buf.WriteString(fmt.Sprintf("\n└───%s", nodeStr))
	default:
		buf.WriteString(fmt.Sprintf("\n├─ Nodes: %d", len(inst.Nodes)))
		for _, node := range inst.Nodes[0 : size-1] {
			nodeStr := node.StringTree(config)
			nodeStr = strings.ReplaceAll(nodeStr, "\n", "\n│   ")
			buf.WriteString(fmt.Sprintf("\n├───%s", nodeStr))
		}
		nodeStr := inst.Nodes[size-1].StringTree(config)
		nodeStr = strings.ReplaceAll(nodeStr, "\n", "\n    ")
		buf.WriteString(fmt.Sprintf("\n└───%s", nodeStr))
	}
	return buf.String()
}

func (inst *AST) Text() []rune {
	if inst == nil {
		return nil
	}
	var text []rune
	for _, node := range inst.Nodes {
		text = append(text, node.Text()...)
	}
	return text
}

func (inst *AST) CountNodes() int {
	if inst == nil {
		return 0
	}
	count := 0
	for _, node := range inst.Nodes {
		count = count + node.CountNodes()
	}
	return count
}

func (inst *AST) CountTokens() int {
	if inst == nil {
		return 0
	}
	count := 0
	for _, node := range inst.Nodes {
		count = count + node.CountTokens()
	}
	return count
}

func (inst *AST) CountStickyNodes() int {
	if inst == nil {
		return 0
	}
	count := 0
	for _, node := range inst.Nodes {
		count = count + node.CountStickyNodes()
	}
	return count
}
