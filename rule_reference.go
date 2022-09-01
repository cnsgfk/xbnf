package xbnf

import (
	"fmt"
	"strings"
)

type ReferenceRule struct {
	ruleBase
	refName string
}

func (inst *ReferenceRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	ruleRecord := grammar.GetRecord(inst.refName)
	if ruleRecord == nil {
		evalResult := &EvalResult{
			Virtual: inst.virtual,
			NonData: inst.nondata,
			Error:   fmt.Errorf("rule '%s' not defined", inst),
		}
		return evalResult
	}
	evalResult := ruleRecord.rule.Eval(grammar, charstream, flagLeadingSpaces)
	if inst.virtual {
		var isVirtual bool
		if evalResult.Virtual {
			isVirtual = false
		} else {
			isVirtual = true
		}
		evalResult.Virtual = isVirtual
		if evalResult.Node != nil {
			evalResult.Node.Virtual = isVirtual
		}
	}
	if inst.nondata {
		var isNonData bool
		if evalResult.NonData {
			isNonData = false
		} else {
			isNonData = true
		}
		evalResult.NonData = isNonData
		if evalResult.Node != nil {
			evalResult.Node.NonData = isNonData
		}
	}
	return evalResult
}

// Returns the rule definition string in xbnf format
func (inst ReferenceRule) String() string {
	annotation := inst.annotation()
	return fmt.Sprintf("%s%s", string(annotation), inst.refName)
}

func (inst ReferenceRule) StringWithIndent(indent string) string {
	return fmt.Sprintf("reference:%s", inst.String())
}

// The name is the name of rule under construction.
func inReference(grammar *Grammar, name string, cs ICharstream) (IRule, error) {
	var buf strings.Builder
	cs.SkipSpaces()
	char := cs.Next()
	// first char must be a letter or underscore
	if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_') {
		return nil, fmt.Errorf("rule name reference must be start with a letter or an underscore")
	}
	buf.WriteRune(char)
	for {
		char = cs.Peek()
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_' || (char >= '0' && char <= '9') {
			buf.WriteRune(char)
			cs.Next()
			continue
		}
		break
	}
	ruleName := buf.String()
	if ruleName == "EOF" {
		return EOF(), nil
	}
	grammar.AddUsage(ruleName, name)
	rule := &ReferenceRule{}
	rule.refName = ruleName
	return rule, nil
}
