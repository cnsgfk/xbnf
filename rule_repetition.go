package xbnf

import (
	"fmt"
	"strconv"
	"strings"
)

type RepetitionRule struct {
	ruleBase
	rule IRule
	// we can specify how many times the language allows repeat
	// * means 0 or more time(s)
	// + means 1 or more time(s)
	// <min, max> means at least min times, and at most max times
	min uint
	max uint // 0 means unlimited or infinity
}

func (inst *RepetitionRule) desc() string {
	var desc string
	if inst.max == 0 {
		desc = fmt.Sprintf("%d or more time(s)", inst.min)
	} else {
		desc = fmt.Sprintf("%d to %d time(s)", inst.min, inst.max)
	}
	if inst.name != "" {
		return fmt.Sprintf("%s: %s", inst.name, desc)
	}
	return fmt.Sprintf("%s of %s", desc, inst.rule.desc())
}

// RepetitionRule Eval will always returns a result with a RepetitionNode, which may contain
// 0 or more children INode
func (inst *RepetitionRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	evalResult := &EvalResult{Virtual: inst.virtual, NonData: inst.nondata}
	evalResult.Sticky = true // will change to false if any of the child result is non-sticky
	node := &Node{RuleType: TypeRepetition, RuleName: inst.name, Virtual: inst.virtual, NonData: inst.nondata}
	cs := charstream
	var nodes []*Node // all found matched nodes
	rule := inst.rule
	for {
		cs = newCharstreamPrepend(cs, evalResult.CharsUnused)
		if cs.Peek() == EOFChar {
			break
		}
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
			evalResult.Sticky = false //change to false if any of the child result is non-sticky
		}

		if result.Node == nil {
			// we didn't find a match, should prepare the evalResult and exist
			evalResult.Error = result.Error
			evalResult.ErrIdx = charstream.Cursor()
			break
		}

		if node.Position == nil {
			node.Position = result.Node.Position
		}

		// got another match
		nodes = append(nodes, result.Node)
		if inst.max > 0 && len(nodes) == int(inst.max) { // found max number of matches
			break
		}
	}

	if len(nodes) < int(inst.min) {
		// eval fails, didn't found enough repetitions
		// all chars read should be unused
		evalResult.CharsUnused = evalResult.CharsRead
		if evalResult.Error == nil {
			evalResult.Error = fmt.Errorf("%s: %d less than minimal %d", inst.desc(), len(nodes), inst.min)
			evalResult.ErrIdx = charstream.Cursor()
		}
		return evalResult
	}

	node.ChildNodes = nodes
	evalResult.Node = node
	node.Sticky = evalResult.Sticky

	return evalResult
}

func (inst *RepetitionRule) String() string {
	var buf strings.Builder
	buf.WriteString(string(inst.annotation()))
	buf.WriteRune(RepetitionOpenSymbol)
	buf.WriteRune(' ')
	buf.WriteString(inst.rule.String())
	buf.WriteRune(' ')
	buf.WriteRune(RepetitionCloseSymbol)
	inst.repetitionSpecString(&buf)
	return buf.String()
}

func inRepetition(grammar *Grammar, name string, cs ICharstream) (*RepetitionRule, error) {
	reps := &RepetitionRule{}
	openChar := cs.Next()
	if openChar != '{' {
		return nil, fmt.Errorf("repetition must start with curly brace '{'")
	}
	rule, err := grammar.parse(name, cs, []rune{'}'})
	if err != nil {
		return nil, err
	}
	closeChar := cs.Next()
	if closeChar != '}' {
		return nil, fmt.Errorf("repetition must end with curly brace '}'")
	}
	// process repetition specification, could be a '*', '+', or 2 non-negative integers
	// enclosed in '<' and '>'
	char := cs.Peek()
	switch char {
	case '*':
		reps.min = 0
		reps.max = 0
		cs.Next()
	case '+':
		reps.min = 1
		reps.max = 0
		cs.Next()
	case '<':
		min, max, err := inRepeatSpec(cs)
		if err != nil {
			return nil, err
		}
		reps.min = min
		reps.max = max
	}
	reps.rule = rule
	return reps, nil
}

func inRepeatSpec(cs ICharstream) (uint, uint, error) {
	var min, max uint
	openChar := cs.Next()
	if openChar != '<' {
		return 0, 0, fmt.Errorf("repetition specification must start with a '<'")
	}
	// ignore space after '<'
	cs.SkipSpaces()
	// get min uint
	var buf strings.Builder
	hasCommas := true
	min = 0
EXIT_MIN:
	for {
		char := cs.Next()
		switch char {
		case ' ', '\t', '\n':
			continue
		case ',':
			break EXIT_MIN
		case '>':
			hasCommas = false
			break EXIT_MIN
		default:
			if '0' <= char && char <= '9' {
				buf.WriteRune(char)
				continue
			}
			return 0, 0, fmt.Errorf("invalid char in repetition specification: %c", char)
		}
	}
	minStr := buf.String()
	if len(minStr) > 0 {
		m, err := strconv.ParseUint(minStr, 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid repetition specification: %s", minStr)
		}
		min = uint(m)
	}
	max = 0
	if hasCommas {
		// get max uint
		buf.Reset()
	EXIT_MAX:
		for {
			char := cs.Next()
			switch char {
			case ' ', '\t', '\n':
				continue
			case '>':
				break EXIT_MAX
			default:
				if '0' <= char && char <= '9' {
					buf.WriteRune(char)
					continue
				}
				return 0, 0, fmt.Errorf("invalid char in repetition specification: %c", char)
			}
		}
		maxStr := buf.String()
		if len(maxStr) > 0 {
			m, err := strconv.ParseUint(maxStr, 10, 32)
			if err != nil {
				return 0, 0, fmt.Errorf("invalid repetition specification: %s", maxStr)
			}
			max = uint(m)
		}
	}
	if max != 0 && max < min {
		return 0, 0, fmt.Errorf("invalid repetition specification: max repeat less than min repeat: <%d,%d>", min, max)
	}
	return min, max, nil
}

func (inst *RepetitionRule) StringWithIndent(indent string) string {
	var buf strings.Builder
	buf.WriteString("repetition")
	inst.repetitionSpecString(&buf)
	buf.WriteRune(':')
	buf.WriteString(fmt.Sprintf("\n%s", inst.rule.StringWithIndent(indent)))
	result := buf.String()
	return strings.ReplaceAll(result, "\n", "\n"+indent)
}

func (inst *RepetitionRule) repetitionSpecString(buf *strings.Builder) {
	if inst.min == inst.max {
		if inst.min != 0 {
			buf.WriteString(fmt.Sprintf("<%d>", inst.min))
		}
	} else {
		if inst.max == 0 {
			if inst.min == 1 {
				buf.WriteRune('+')
			} else {
				buf.WriteString(fmt.Sprintf("<%d,0>", inst.min))
			}
		} else {
			buf.WriteString(fmt.Sprintf("<%d,%d>", inst.min, inst.max))
		}
	}
}
