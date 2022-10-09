package xbnf

import (
	"fmt"
	"strings"
)

type TerminalCharsRule struct {
	ruleBase
	text []rune
}

func (inst *TerminalCharsRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	evalResult := &EvalResult{
		Virtual: inst.virtual,
		NonData: inst.nondata,
		Sticky:  true,
	}
	if charstream.Peek() == EOFChar {
		if inst.name != "" {
			evalResult.Error = fmt.Errorf("missing %s", inst.name)
		}
		return evalResult
	}
	node := &Node{
		RuleType: TypeChars,
		RuleName: inst.name,
		Virtual:  inst.virtual,
		NonData:  inst.nondata,
		Sticky:   true,
	}

	startCursor := charstream.Cursor()
	text := inst.text
	if flagLeadingSpaces == SUGGEST_SKIP {
		leadingWSpaces := leadingWhiteSpace(text)
		skippedWSpaces := charstream.SkipSpaces()
		evalResult.CharsRead = append(evalResult.CharsRead, skippedWSpaces...)

		if len(skippedWSpaces) < len(leadingWSpaces) {
			// for sure not match
			evalResult.CharsUnused = evalResult.CharsRead
			if inst.name != "" {
				evalResult.Error = fmt.Errorf("missing %s", inst.name)
			}
			return evalResult
		}
		startCursor = startCursor + len(skippedWSpaces)
		// now we need to check if the leadingWSpace matches the end of skippedWSpace
		skippedWSpaces = skippedWSpaces[len(skippedWSpaces)-len(leadingWSpaces):]
		for i, skippedWSpace := range skippedWSpaces {
			if skippedWSpace != leadingWSpaces[i] {
				// not match
				evalResult.CharsUnused = evalResult.CharsRead
				if inst.name != "" {
					evalResult.Error = fmt.Errorf("missing %s", inst.name)
				}
				return evalResult
			}
		}
		// now the leadingWSpaces matched, check the rest of text
		text = inst.text[len(leadingWSpaces):]
	}

	charsRead, succeeded := charstream.Match(text)
	evalResult.CharsRead = append(evalResult.CharsRead, charsRead...)
	if !succeeded {
		evalResult.CharsUnused = evalResult.CharsRead
		if inst.name != "" {
			evalResult.Error = fmt.Errorf("missing %s", inst.name)
		}
		return evalResult
	}
	node.Position = charstream.PositionLookup(startCursor)
	node.Chars = append(node.Chars, inst.text...)
	evalResult.Node = node
	return evalResult
}

// Returns the rule definition string in xbnf format
func (inst *TerminalCharsRule) String() string {
	annotations := string(inst.annotation())
	str := string(inst.text)
	if strings.ContainsRune(str, '\'') {
		str = strings.ReplaceAll(str, "'", "\\'")
	}
	return fmt.Sprintf("%s'%s'", annotations, str)
}

func (inst *TerminalCharsRule) StringWithIndent(indent string) string {
	return fmt.Sprintf("char:%s", inst.String())
}
