package xbnf

import (
	"fmt"
	"strings"
)

type ConcatenateRule struct {
	ruleBase
	rules []IRule
}

func (inst *ConcatenateRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	evalResult := &EvalResult{Virtual: inst.virtual, NonData: inst.nondata}
	evalResult.Sticky = true // will change to false if any of the child result is non-sticky
	node := &Node{RuleType: TypeConcatenate, RuleName: inst.name, Virtual: inst.virtual, NonData: inst.nondata}
	cs := charstream
	for _, rule := range inst.rules {
		cs = NewCharstreamPrepend(cs, evalResult.CharsUnused)
		result := rule.Eval(grammar, cs, flagLeadingSpaces)
		//
		// Update evalResult.CharsRead and evalResult.CharsUnused
		// Note the result.CharsRead may be less or more than current evalResult.CharsUnused, and the evelResult.CharsRead
		// contains all current evalResult.CharsUnused, if result.CharsRead less than current charsUnused, don't need to
		// update evalResult.CharsRead, otherwise, need to add the extra chars to evalResult.CharsRead
		// ┄┄┄┼┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄────────────┤ evalResult.CharsRead
		//    ├─────────────────────┬──────────┤ evalResult.CharsUnused
		//                          ├──────────┤ unusedLeft = evalResult.CharsUnused - result.CharsUsed
		//    ├─────────────────────┼─────────────────────┤ result.CharsRead
		//                          ├─────────────────────┤ result.CharsUnused
		if len(result.CharsRead) > len(evalResult.CharsUnused) {
			extraChars := result.CharsRead[len(evalResult.CharsUnused):]
			evalResult.CharsRead = append(evalResult.CharsRead, extraChars...)
			evalResult.CharsUnused = result.CharsUnused
		} else {
			resultCharsUsed := result.charsUsed()
			evalResult.CharsUnused = evalResult.CharsUnused[len(resultCharsUsed):] // unusedLeft
			if len(evalResult.CharsUnused) < len(resultCharsUsed) {
				evalResult.CharsUnused = result.CharsUnused
			}
		}

		if !result.Sticky {
			evalResult.Sticky = false
		}

		if result.Node == nil {
			// one of the rule doesn't match, resulting this rule does not match
			evalResult.CharsUnused = evalResult.CharsRead
			if inst.name != "" {
				evalResult.Error = fmt.Errorf("%s: %s", inst.name, result.Error)
			} else {
				evalResult.Error = result.Error
			}
			return evalResult
		}

		// option node should be treated specially to decide whether or not next eval should skip leading spaces
		// when option rule doesn't match, the hintSkipLeadingSpaces should remain the same
		if result.Node.RuleType != TypeOption ||
			(len(result.Node.ChildNodes) > 0 || len(result.Node.Chars) > 0) {
			// it's not an option node, or matched
			if result.Sticky {
				flagLeadingSpaces = SUGGEST_NOT_SKIP
			} else {
				flagLeadingSpaces = SUGGEST_SKIP
			}
		}

		node.ChildNodes = append(node.ChildNodes, result.Node)
	}

	// all child rules matched
	node.Sticky = evalResult.Sticky
	evalResult.Node = node
	return evalResult
}

func (inst *ConcatenateRule) String() string {
	var buf strings.Builder
	if len(inst.rules) < 2 {
		return "invalid concatenate rule: # of rules less than 2"
	}
	annotations := inst.annotation()
	if len(annotations) > 0 {
		buf.WriteString(string(annotations))
		buf.WriteRune(GroupOpenSymbol)
	}
	buf.WriteString(inst.rules[0].String())
	for _, rule := range inst.rules[1:] {
		buf.WriteRune(' ')
		buf.WriteString(rule.String())
	}
	if len(annotations) > 0 {
		buf.WriteRune(GroupCloseSymbol)
	}
	return buf.String()
}

func (inst *ConcatenateRule) StringWithIndent(indent string) string {
	var buf strings.Builder
	if len(inst.rules) < 2 {
		return "invalid concatenate rule: # of rules less than 2"
	}
	buf.WriteString("concatenate:")
	for _, rule := range inst.rules {
		buf.WriteString(fmt.Sprintf("\n%s", rule.StringWithIndent(indent)))
	}
	result := buf.String()
	return strings.ReplaceAll(result, "\n", "\n"+indent)
}
