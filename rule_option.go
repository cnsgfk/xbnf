package xbnf

import (
	"fmt"
	"strings"
)

type OptionRule struct {
	ruleBase
	rule IRule
}

func (inst *OptionRule) desc() string {
	if inst.name != "" {
		return fmt.Sprintf("optional %s", inst.name)
	}
	return fmt.Sprintf("optional %s", inst.rule.desc())
}

// Eval with option rule always produce a OptionNode, whose Child node may be nil, meaning
// optional token not found
func (inst *OptionRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	node := &Node{
		RuleType:  TypeOption,
		RuleName:  inst.name,
		Tokenized: inst.tokenized,
		Virtual:   inst.virtual,
		NonData:   inst.nondata,
	}
	evalResult := inst.rule.Eval(grammar, charstream, flagLeadingSpaces)
	node.Sticky = evalResult.Sticky
	if evalResult.Node == nil {
		evalResult.Node = node // OptionRule rule should always generate a node
		return evalResult
	}
	node.ChildNodes = append(node.ChildNodes, evalResult.Node)
	node.Position = evalResult.Node.Position
	evalResult.Node = node
	return evalResult
}
func inOption(grammar *Grammar, name string, cs ICharstream) (*OptionRule, error) {
	option := &OptionRule{}
	openChar := cs.Next()
	if openChar != '[' {
		return nil, fmt.Errorf("option must start with a square bracket '['")
	}
	rule, err := grammar.parse(name, cs, []rune{']'})
	if err != nil {
		return nil, err
	}
	closeChar := cs.Next()
	if closeChar != ']' {
		return nil, fmt.Errorf("option must end with a square bracket ']'")
	}
	option.rule = rule
	return option, nil
}
func (inst *OptionRule) String() string {
	var buf strings.Builder
	buf.WriteString(string(inst.annotation()))
	buf.WriteRune(OptionOpenSymbol)
	buf.WriteRune(' ')
	buf.WriteString(inst.rule.String())
	buf.WriteRune(' ')
	buf.WriteRune(OptionCloseSymbol)
	return buf.String()
}

func (inst *OptionRule) StringWithIndent(indent string) string {
	var buf strings.Builder
	buf.WriteString("option:")
	buf.WriteString(fmt.Sprintf("\n%s", inst.rule.StringWithIndent(indent)))
	result := buf.String()
	return strings.ReplaceAll(result, "\n", "\n"+indent)
}
