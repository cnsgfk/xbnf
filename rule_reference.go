package xbnf

import (
	"fmt"
	"strings"
)

type ReferenceRule struct {
	ruleBase
	refName string
}

func (inst *ReferenceRule) desc() string {
	return inst.refName
}

func (inst *ReferenceRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	ruleRecord := grammar.GetRecord(inst.refName)
	if ruleRecord == nil {
		evalResult := &EvalResult{
			Error: fmt.Errorf("rule '%s' not defined", inst),
		}
		return evalResult
	}
	evalResult := ruleRecord.rule.Eval(grammar, charstream, flagLeadingSpaces)
	if evalResult.Node != nil {
		if inst.tokenized {
			if evalResult.Node.Tokenized { // tokenized && tokenized is non-tokenized
				evalResult.Node.Tokenized = false
			} else {
				evalResult.Node.Tokenized = true
			}
		}
		if inst.virtual {
			if evalResult.Node.Virtual { // virtual && virtual is non-virtual
				evalResult.Node.Virtual = false
			} else {
				evalResult.Node.Virtual = true
			}
		}
		if inst.nondata {
			if evalResult.Node.NonData { // nondata && nondata is data
				evalResult.Node.NonData = false
			} else {
				evalResult.Node.NonData = true
			}
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
