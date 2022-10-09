package xbnf

import "fmt"

type EOFRule struct {
	ruleBase
}

func EOF() *EOFRule {
	rule := &EOFRule{}
	rule.nondata = true
	rule.virtual = true
	return rule
}

func (inst *EOFRule) Name() string {
	return string(TypeEOF)
}

// Returns the rule definition string in xbnf format
func (inst *EOFRule) String() string {

	return string(inst.annotation()) + string(TypeEOF)
}

func (inst *EOFRule) StringWithIndent(indent string) string {
	return string(TypeEOF)
}

func (inst *EOFRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	evalResult := &EvalResult{
		Virtual: true,
		NonData: true,
		Sticky:  true,
	}
	if flagLeadingSpaces == 1 {
		evalResult.CharsRead = charstream.SkipSpaces()
	}
	if EOFChar == charstream.Peek() {
		evalResult.Node = &Node{
			RuleType: TypeEOF,
			RuleName: string(TypeEOF),
			Virtual:  inst.virtual,
			NonData:  inst.nondata,
			Sticky:   true,
			Chars:    []rune{EOFChar},
			Position: charstream.Position(),
		}
	} else {
		evalResult.CharsUnused = evalResult.CharsRead
		evalResult.Error = fmt.Errorf("missing EOF")
	}

	return evalResult
}
