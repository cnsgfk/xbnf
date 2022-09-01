package xbnf

import (
	"fmt"
	"strings"
)

type TerminalStringRule struct {
	ruleBase
	text []rune
}

func (inst *TerminalStringRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	evalResult := &EvalResult{
		Virtual: inst.virtual,
		NonData: inst.nondata,
		Sticky:  false,
	}
	if charstream.Peek() == EOFChar {
		if inst.name != "" {
			evalResult.Error = fmt.Errorf("missing %s", inst.name)
		}
		return evalResult
	}
	node := &Node{
		RuleType: TypeString,
		RuleName: inst.name,
		Virtual:  inst.virtual,
		NonData:  inst.nondata,
		Sticky:   false,
	}

	// String rule always skip leading white spaces unless explicitly ask not to, ie. flagLeadingSpaces=NOT_SKIP(3)
	text := inst.text
	if flagLeadingSpaces == SUGGEST_SKIP || flagLeadingSpaces == SUGGEST_NOT_SKIP {
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
	node.Chars = append(node.Chars, inst.text...)
	evalResult.Node = node
	return evalResult

}

// Returns the rule definition string in xbnf format
func (inst *TerminalStringRule) String() string {
	str := string(inst.text)
	return fmt.Sprintf("%s\"%s\"", string(inst.annotation()), str)
}

func (inst *TerminalStringRule) StringWithIndent(indent string) string {
	return fmt.Sprintf("token:%s", inst.String())
}

func inString(name string, cs ICharstream) (*TerminalStringRule, error) {
	var buf strings.Builder
	var definedUnicodes []rune
	openChar := cs.Next()
	if openChar != '"' {
		return nil, fmt.Errorf("terminal string must start with a double quote")
	}
	var char rune
	for {
		char = cs.Next()
		if char == EOFChar {
			return nil, fmt.Errorf("terminal string must end with a double quote")
		}
		if char == openChar {
			break
		}
		if char == '\\' { // escape char
			if cs.Peek() == 'u' { // it's an unicode escape (with leading '\u')
				unicode, err := parseUnicodeEscape(cs)
				if err != nil {
					return nil, err
				}
				char = unicode
				definedUnicodes = append(definedUnicodes, char)
			} else { // a bare \ just escape the next char from participate in boundary check
				char = cs.Next()
			}
		}
		buf.WriteRune(char)
	}
	rule := &TerminalStringRule{}
	rule.text = []rune(buf.String())
	return rule, nil
}
