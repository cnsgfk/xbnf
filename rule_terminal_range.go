package xbnf

import (
	"fmt"
)

type TerminalRangeRule struct {
	ruleBase
	begin          rune // begin must less or equal end character
	beginAsUnicode bool
	end            rune
	endAsUnicode   bool
}

func (inst *TerminalRangeRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
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

	if flagLeadingSpaces == SUGGEST_SKIP {
		skippedWSpaces := charstream.SkipSpaces()
		evalResult.CharsRead = append(evalResult.CharsRead, skippedWSpaces...)
		for i, char := range skippedWSpaces {
			if inst.begin <= char && char >= inst.end {
				// matched
				evalResult.CharsUnused = skippedWSpaces[i+1:]
				node.Chars = append(node.Chars, char)
				evalResult.Node = node
				return evalResult
			}
		}
	}

	char := charstream.Peek()
	if inst.begin <= char && char <= inst.end {
		// matched
		char = charstream.Next()
		evalResult.CharsRead = append(evalResult.CharsRead, char)
		node.Chars = append(node.Chars, char)
		evalResult.Node = node
		return evalResult
	}

	evalResult.CharsUnused = evalResult.CharsRead
	if inst.name == "" {
		evalResult.Error = fmt.Errorf("missing %s", inst.String())
	} else {
		evalResult.Error = fmt.Errorf("%s:missing %s", inst.name, inst.String())
	}
	return evalResult
}

// Returns the rule definition string in xbnf format
func (inst *TerminalRangeRule) String() string {
	annotations := string(inst.annotation())
	beginStr := ""
	endStr := ""
	if inst.beginAsUnicode {
		beginStr = fmt.Sprintf("%s\\u%04X", annotations, inst.begin)
	} else {
		if inst.begin == '\'' {
			beginStr = "'\\''"
		} else {
			beginStr = "'" + string(inst.begin) + "'"
		}
	}
	if inst.endAsUnicode {
		endStr = fmt.Sprintf("%s\\u%04X", annotations, inst.end)
	} else {
		if inst.end == '\'' {
			endStr = "'\\''"
		} else {
			endStr = "'" + string(inst.end) + "'"
		}
	}

	return fmt.Sprintf("%s%s%s%s", annotations, beginStr, string(CharRangeSymbol), endStr)
}

func (inst *TerminalRangeRule) StringWithIndent(indent string) string {
	return fmt.Sprintf("%s:%s", TypeRange, inst.String())
}

// inRange expects the 1st char in the input charstream is a CharRangeSymbol '-'.
func inRange(cs ICharstream, beginChar rune, beginAsUnicode bool) (IRule, error) {
	var endAsUnicode bool
	var endChar rune
	char := cs.Next()
	if char != CharRangeSymbol {
		return nil, fmt.Errorf("char range must be 2 chars connected by a range symbol '-'")
	}
	endChar = cs.Next() // expecting a open "'" or "\"
	if endChar == CharsSymbol {
		endChar = cs.Next()
		if endChar == EscapeSymbol { // escape char
			if cs.Peek() == 'u' { // it's an unicode escape (with leading '\u')
				unicode, err := parseUnicodeEscape(cs)
				if err != nil {
					return nil, err
				}
				endChar = unicode
				endAsUnicode = true
			} else {
				return nil, fmt.Errorf("char range: end char missing unicode escape \\u")
			}
		}
		char = cs.Next() // expecting a closing "'"
		if char != CharsSymbol {
			return nil, fmt.Errorf("char range must be 2 chars connected by a range symbol '-'")
		}
	} else if endChar == EscapeSymbol {
		unicode, err := parseUnicodeEscape(cs)
		if err != nil {
			return nil, err
		}
		endChar = unicode
		endAsUnicode = true
	} else {
		return nil, fmt.Errorf("char range must be 2 chars connected by a range symbol '-': invalid 2nd char")
	}

	rule := &TerminalRangeRule{}
	rule.begin = beginChar
	rule.beginAsUnicode = beginAsUnicode
	rule.end = endChar
	rule.endAsUnicode = endAsUnicode
	return rule, nil
}
