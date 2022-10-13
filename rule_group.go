package xbnf

import (
	"fmt"
	"strings"
)

type GroupRule struct {
	ruleBase
	rule IRule
}

func (inst *GroupRule) desc() string {
	if inst.name != "" {
		return inst.name
	}
	return inst.rule.desc()
}

func (inst *GroupRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	node := &Node{RuleType: TypeGroup, RuleName: inst.name, Virtual: inst.virtual, NonData: inst.nondata}
	evalResult := inst.rule.Eval(grammar, charstream, flagLeadingSpaces)
	if evalResult.Node == nil {
		evalResult.CharsUnused = evalResult.CharsRead
		return evalResult
	}
	node.Sticky = evalResult.Sticky
	node.ChildNodes = append(node.ChildNodes, evalResult.Node)
	evalResult.Node = node
	node.Position = evalResult.Node.Position
	return evalResult
}

func inGroup(grammar *Grammar, name string, cs ICharstream) (*GroupRule, error) {
	group := &GroupRule{}
	openChar := cs.Next()
	if openChar != '(' {
		return nil, fmt.Errorf("group must start with a round bracket '('")
	}
	rule, err := grammar.parse(name, cs, []rune{')'})
	if err != nil {
		return nil, err
	}
	closeChar := cs.Next()
	if closeChar != ')' {
		return nil, fmt.Errorf("group must end with a round bracket ')'")
	}
	group.rule = rule
	return group, nil
}
func (inst *GroupRule) String() string {
	var buf strings.Builder
	buf.WriteString(string(inst.annotation()))
	buf.WriteRune(GroupOpenSymbol)
	buf.WriteRune(' ')
	buf.WriteString(inst.rule.String())
	buf.WriteRune(' ')
	buf.WriteRune(GroupCloseSymbol)
	return buf.String()
}

func (inst *GroupRule) StringWithIndent(indent string) string {
	var buf strings.Builder
	buf.WriteString("group:")
	buf.WriteString(fmt.Sprintf("\n%s", inst.rule.StringWithIndent(indent)))
	result := buf.String()
	return strings.ReplaceAll(result, "\n", "\n"+indent)
}
