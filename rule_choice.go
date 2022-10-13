package xbnf

import (
	"errors"
	"fmt"
	"strings"
)

type ChoiceRule struct {
	ruleBase
	groups [][]IRule // list of order groups, each group is a list of non-ordered choice
}

func (inst *ChoiceRule) desc() string {
	if inst.name != "" {
		return inst.name
	}
	var rules []IRule
	for _, g := range inst.groups {
		for _, r := range g {
			rules = append(rules, r)
		}
	}
	var buf strings.Builder
	count := len(rules)
	buf.WriteString(rules[0].desc())
	for i := 1; i < count-1; i++ {
		buf.WriteString(", " + rules[i].desc())
	}
	buf.WriteString(" and " + rules[count-1].desc())
	return buf.String()
}

func (inst *ChoiceRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	evalResult := &EvalResult{Virtual: inst.virtual, NonData: inst.nondata}
	node := &Node{RuleType: TypeChoice, RuleName: inst.name, Virtual: inst.virtual, NonData: inst.nondata}
	cs := charstream
	sticky := true
	var resultsMatched []*EvalResult
	var resultFound *EvalResult
	var maxErr error
	var maxErrCursor int
	//startPos := charstream.Position()
	for _, group := range inst.groups {
		resultsMatched = nil
		resultFound = nil
		for _, rule := range group {
			cs = newCharstreamPrepend(charstream, evalResult.CharsRead)
			result := rule.Eval(grammar, cs, flagLeadingSpaces)
			if !result.Sticky { // as long as one of the choices is non-sticky, the result should be non-sticky
				sticky = false
			}
			if len(evalResult.CharsRead) < len(result.CharsRead) {
				evalResult.CharsRead = result.CharsRead
			}
			if result.Node != nil {
				resultsMatched = append(resultsMatched, result)
			} else {
				if result.ErrIdx > maxErrCursor {
					maxErrCursor = result.ErrIdx
					maxErr = result.Error
				}
			}
		}
		matches := len(resultsMatched)
		if matches == 0 {
			// no match in this group, continue to next group
			continue
		} else if matches == 1 {
			resultFound = resultsMatched[0]
			break
		} else { // we found multiple matches, get the one that is most greedy, ie. uses most chars
			maxUsed := 0
			for _, result := range resultsMatched {
				countCharsUsed := result.countCharsUsed()
				if countCharsUsed < maxUsed {
					continue
				}
				if countCharsUsed == maxUsed {
					if resultFound == nil {
						resultFound = result
						continue
					}
					// we found ambiguity
					var buf strings.Builder
					if inst.name == "" {
						buf.WriteString("ambiguity found:")
					} else {
						buf.WriteString(fmt.Sprintf("rule %s: ambiguity found", inst.name))
					}
					buf.WriteString(fmt.Sprintf("\t\n%s: %s", resultFound.Node.RuleName, string(resultFound.Node.Text())))
					buf.WriteString(fmt.Sprintf("\t\n%s: %s", result.Node.RuleName, string(result.Node.Text())))
					evalResult.Error = errors.New(buf.String())
					return evalResult
				}
				maxUsed = countCharsUsed
				resultFound = result
			}
			break
		}
	}
	evalResult.Sticky = sticky
	if resultFound != nil { // we found a result
		node.ChildNodes = append(node.ChildNodes, resultFound.Node)
		node.Sticky = evalResult.Sticky
		node.Position = resultFound.Node.Position
		evalResult.Node = node
		evalResult.CharsUnused = evalResult.CharsRead[resultFound.countCharsUsed():]
	} else {
		evalResult.CharsUnused = evalResult.CharsRead
		evalResult.ErrIdx = maxErrCursor
		evalResult.Error = fmt.Errorf("%s: %s", inst.desc(), maxErr)
	}
	return evalResult
}

// inChoice expects the 1st non-space char be '|' or '>'.
func inChoice(grammar *Grammar, name string, firstRule IRule, cs ICharstream) (*ChoiceRule, error) {
	rules := []IRule{firstRule}
	groups := [][]IRule{}
EXIT:
	for {
		char := cs.Peek()
		switch char {
		case ' ': // ignore space
			cs.Next()
			continue
		case ChoiceSymbol:
			cs.Next()
			rule, err := grammar.parseOne(name, cs)
			if err != nil {
				return nil, err
			}
			rules = append(rules, rule)
		case ChoiceOrderSymbol: // time to create a new group
			if len(rules) == 0 {
				return nil, fmt.Errorf("no rule found for choice group before '%c'", ChoiceOrderSymbol)
			}
			groups = append(groups, rules)
			rules = nil
			cs.Next()
			rule, err := grammar.parseOne(name, cs)
			if err != nil {
				return nil, err
			}
			rules = append(rules, rule)
		default:
			break EXIT
		}
	}
	groups = append(groups, rules)
	if len(groups) < 1 || (len(groups) == 1 && len(groups[0]) < 2) {
		return nil, fmt.Errorf("choice rule must has at least 2 rules as choices")
	}
	ruleChoice := &ChoiceRule{}
	ruleChoice.groups = groups
	return ruleChoice, nil
}

func (inst *ChoiceRule) String() string {
	var buf strings.Builder

	annotations := inst.annotation()
	if len(annotations) > 0 {
		buf.WriteString(string(annotations))
		buf.WriteRune(GroupOpenSymbol)
	}

	group0 := inst.groups[0]
	groups := inst.groups[1:]

	rule0 := group0[0]
	rules := group0[1:]
	buf.WriteString(rule0.String())
	for _, rule := range rules {
		buf.WriteRune(' ')
		buf.WriteRune(ChoiceSymbol)
		buf.WriteRune(' ')
		buf.WriteString(rule.String())
	}
	for _, group := range groups {
		buf.WriteRune(' ')
		buf.WriteRune(ChoiceOrderSymbol)
		buf.WriteRune(' ')
		rule0 := group[0]
		rules := group[1:]
		buf.WriteString(rule0.String())
		for _, rule := range rules {
			buf.WriteRune(' ')
			buf.WriteRune(ChoiceSymbol)
			buf.WriteRune(' ')
			buf.WriteString(rule.String())
		}
	}

	if len(annotations) > 0 {
		buf.WriteRune(GroupCloseSymbol)
	}

	return buf.String()
}

func (inst *ChoiceRule) StringWithIndent(indent string) string {
	var buf strings.Builder
	buf.WriteString("choice:")
	for _, group := range inst.groups {
		for _, rule := range group {
			buf.WriteString(fmt.Sprintf("\n%s", rule.StringWithIndent(indent)))
		}
	}
	result := buf.String()
	return strings.ReplaceAll(result, "\n", "\n"+indent)
}
