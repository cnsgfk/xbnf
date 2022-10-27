package xbnf

import (
	"fmt"
	"strconv"
)

type TerminalCharRule struct {
	ruleBase
	text             rune
	definedAsUnicode bool
}

func (inst *TerminalCharRule) desc() string {
	if inst.name == "" {
		return "'" + string(inst.text) + "'"
	}
	return inst.name
}
func (inst *TerminalCharRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	evalResult := &EvalResult{
		Sticky: true,
	}
	if charstream.Peek() == EOFChar {
		evalResult.Error = fmt.Errorf("missing %s at EOF", inst.desc())
		evalResult.ErrIdx = charstream.Cursor()
		return evalResult
	}
	node := &Node{
		RuleType: TypeChar,
		RuleName: inst.name,
		Virtual:  inst.virtual,
		NonData:  inst.nondata,
		Sticky:   true,
	}

	startPos := charstream.Position()

	if flagLeadingSpaces == SUGGEST_SKIP {
		skippedSpaces := charstream.SkipSpaces()
		evalResult.CharsRead = append(evalResult.CharsRead, skippedSpaces...)
		// inst.text may be a whitespace, so we need to check if it's among the skippedSpaces
		if IsWhiteSpace(inst.text) && len(skippedSpaces) > 0 {
			for i, wspace := range skippedSpaces {
				if wspace == inst.text {
					node.Chars = append(node.Chars, inst.text)
					node.Position = charstream.PositionLookup(charstream.Cursor() - len(skippedSpaces) + i)
					evalResult.Node = node
					evalResult.CharsUnused = skippedSpaces[i+1:]
					return evalResult
				}
			}
			evalResult.CharsUnused = evalResult.CharsRead
			evalResult.Error = fmt.Errorf("missing %s at %s", inst.desc(), startPos.String())
			evalResult.ErrIdx = charstream.Cursor()
			return evalResult
		}
	}

	char := charstream.Peek()
	if char != inst.text {
		evalResult.CharsUnused = evalResult.CharsRead
		evalResult.Error = fmt.Errorf("missing %s at %s", inst.desc(), startPos.String())
		evalResult.ErrIdx = charstream.Cursor()
		return evalResult
	}
	node.Position = charstream.Position()
	char = charstream.Next()
	evalResult.CharsRead = append(evalResult.CharsRead, char)
	node.Chars = append(node.Chars, char)
	evalResult.Node = node
	return evalResult
}

// Returns the rule definition string in xbnf format
func (inst *TerminalCharRule) String() string {
	annotations := string(inst.annotation())
	if inst.definedAsUnicode {
		return fmt.Sprintf("%s\\u%04X", annotations, inst.text)
	}
	if inst.text == '\\' {
		return annotations + "'\\\\'" // output double back-slash \\
	}
	if inst.text == '\'' {
		return annotations + "'\\''"
	}
	return annotations + "'" + string(inst.text) + "'"
}

func (inst *TerminalCharRule) StringWithIndent(indent string) string {
	return fmt.Sprintf("char:%s", inst.String())
}

// inChar expects the 1st char in the input charstream is a single quote `'`.
func inChar(cs ICharstream) (IRule, error) {
	var buf []rune
	var definedUnicodes []rune
	openChar := cs.Next()
	if openChar != '\'' {
		return nil, fmt.Errorf("terminal char must start with a single quote")
	}
	for {
		char := cs.Next()
		if char == EOFChar {
			return nil, fmt.Errorf("terminal char must end with a single quote")
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
		buf = append(buf, char)
	}
	switch len(buf) {
	case 0:
		return nil, fmt.Errorf("terminal char(s) must contains a least 1 character")
	case 1:
		// check if there is a CharRangeSymbol follow
		if cs.Peek() == CharRangeSymbol { // this is a range rule
			return inRange(cs, buf[0], len(definedUnicodes) > 0)
		}
		rule := &TerminalCharRule{}
		rule.text = buf[0]
		rule.definedAsUnicode = len(definedUnicodes) > 0
		return rule, nil
	default:
		rule := &TerminalCharsRule{}
		rule.text = buf
		return rule, nil
	}
}

// Creates a TerminalChar rule for an unicode escape. This function expects the 1st char
// from the input stream is a '\' char.
func inUnicode(cs ICharstream) (IRule, error) {
	openChar := cs.Next()
	uChar := cs.Peek()
	if openChar != '\\' || uChar != 'u' {
		return nil, fmt.Errorf("unicode terminal must start with '\\u' and followed with 4 hex chars")
	}

	unicode, err := parseUnicodeEscape(cs)
	if err != nil {
		return nil, err
	}

	if cs.Peek() == CharRangeSymbol { // this is a range rule
		return inRange(cs, unicode, true)
	}

	rule := &TerminalCharRule{}
	rule.definedAsUnicode = true
	rule.text = unicode
	return rule, nil
}

// In XBNF definition, the back-splash character is used to escape special characters
// that won't be possible to be included in a terminal string or char sequence. For example,
// the char sequence 'didn\'t' won't be possible without the escape char '\'. When escape
// char appears, the following char is treat as is, ie won't be used to test boundary of
// string or char sequance.

// Unicode escape starts with '\u', before calling this function, the leading '\' must
// already be consumed from the input cs. This function expect a 'u' as the 1st char
// from the stream. If not, an error will be returned, and the 1st char remains in the
// input stream (not consumed)
func parseUnicodeEscape(cs ICharstream) (rune, error) {
	if cs.Peek() != 'u' {
		return 0, fmt.Errorf("unicode escape must start with '\\u'")
	}
	cs.Next() // consume the 'u'
	var hexChars []rune
	for i := 0; i < 4; i++ { // read 4 chars from the stream, unicode escape is 4 hex chars
		char := cs.Next()
		if char == EOFChar || !(('0' <= char && char <= '9') ||
			('a' <= char && char <= 'f') ||
			('A' <= char && char <= 'F')) {
			return 0, fmt.Errorf("unicode escape must start with '\\u' and followed with 4 hex chars")
		}
		hexChars = append(hexChars, char)
	}
	codepoint, err := strconv.ParseInt(string(hexChars), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid unicode hex value: %s", string(hexChars))
	}
	return rune(codepoint), nil
}
