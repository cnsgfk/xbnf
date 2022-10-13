package xbnf

import (
	"fmt"
	"strings"
)

const ()

type BlockRule struct {
	ruleBase
	open         IRule
	virtualClose bool // a ~ char in front of the close char >, indicates the chars used to evaluate the close rule will be reused
	close        IRule
	escape       IRule
	excludes     []IRule
}

func (inst *BlockRule) desc() string {
	if inst.name != "" {
		return inst.name
	}
	return fmt.Sprintf("%s to %s", inst.open.desc(), inst.close.desc())
}

func (inst *BlockRule) String() string {
	var buf strings.Builder
	buf.WriteString(string(inst.annotation()))
	buf.WriteRune(blockOpenSymbol)
	buf.WriteString(inst.open.String())

	if inst.escape != nil {
		buf.WriteRune(' ')
		buf.WriteString(inst.escape.String())
	}

	for _, exclude := range inst.excludes {
		buf.WriteString(" ^")
		buf.WriteString(exclude.String())
	}

	buf.WriteRune(' ')
	buf.WriteString(inst.close.String())
	if inst.virtualClose {
		buf.WriteRune(' ')
		buf.WriteRune(blockVirtualClose)
	}
	buf.WriteRune(blockCloseSymbol)
	return buf.String()
}

func (inst *BlockRule) StringWithIndent(indent string) string {
	return fmt.Sprintf("%s%s", indent, inst.String())
}

// Eval extract quote text from the charstream. If the charstream starts with
// openQuote and has a closeQuote somewhere, return the quoted text including
// the open and close quote. Return error if there is no quote text, and the
// chars read so far in order to find the quote.
func (inst *BlockRule) Eval(grammar *Grammar, charstream ICharstream, flagLeadingSpaces int) *EvalResult {
	evalResult := &EvalResult{
		Virtual: inst.virtual,
		NonData: inst.nondata,
		Sticky:  false,
	}
	node := &Node{
		RuleType: TypeBlock,
		RuleName: inst.name,
		Virtual:  inst.virtual,
		NonData:  inst.nondata,
		Sticky:   false, // block of text is always non-sticky,
	}

	cs := charstream
	var charsUsed []rune
	var charsUnused []rune

	// check the open chars of the block text
	result := inst.open.Eval(grammar, cs, flagLeadingSpaces)
	if result.Node != nil {
		node.ChildNodes = append(node.ChildNodes, result.Node) // the 1st child of block node is always open node
		charsUsed = append(charsUsed, result.CharsRead[:result.countCharsUsed()]...)
		charsUnused = result.CharsUnused
		node.Position = result.Node.Position
	} else {
		evalResult.Error = result.Error
		evalResult.CharsRead = result.CharsRead
		evalResult.CharsUnused = result.CharsUnused
		evalResult.ErrIdx = charstream.Cursor()
		return evalResult
	}

	// content child is the 2nd child of block node
	content := &Node{
		RuleType: TypeChars,
		Sticky:   true,
	}
	node.ChildNodes = append(node.ChildNodes, content)
	cs = newCharstreamPrepend(cs, charsUnused)
	content.Position = cs.Position()
	var closeResult *EvalResult
	for {
		// check if we have an escape as next match
		escaped := false
		if inst.escape != nil { // evaluation escape
			result = inst.escape.Eval(grammar, cs, NOT_SKIP) // do not skip space chars
			if result.Node != nil {
				charsUsed = append(charsUsed, result.CharsRead[:result.countCharsUsed()]...)
				charsUnused = result.CharsUnused
				content.Chars = append(content.Chars, result.charsUsed()...) // the escape char should be preserved
				// check if there is an escape token after escape token
				cs = newCharstreamPrepend(cs, charsUnused)
				result = inst.escape.Eval(grammar, cs, NOT_SKIP) // do not skip space chars
				if result.Node != nil {
					// see an escape token after an escape token
					charsUsed = append(charsUsed, result.CharsRead[:result.countCharsUsed()]...)
					charsUnused = result.CharsUnused
					content.Chars = append(content.Chars, result.charsUsed()...)
					continue
				} else {
					escaped = true
					charsUnused = result.CharsRead
				}
			} else {
				charsUnused = result.CharsRead
			}
		}

		// find exclude match(es)
		cs = newCharstreamPrepend(cs, charsUnused)
		if len(inst.excludes) > 0 {
			resultExclude := mostGreedy(grammar, cs, NOT_SKIP, inst.excludes)
			if resultExclude.Node != nil { // found one
				if escaped {
					charsUsed = append(charsUsed, resultExclude.charsUsed()...)
					charsUnused = resultExclude.CharsUnused
					content.Chars = append(content.Chars, resultExclude.charsUsed()...)
					continue
				}
				evalResult.Error = fmt.Errorf("text `%s` is not allowed in text block", string(resultExclude.charsUsed()))
				evalResult.ErrIdx = charstream.Cursor()
				evalResult.CharsRead = append(charsUsed, content.Chars...)
				evalResult.CharsRead = append(evalResult.CharsRead, resultExclude.CharsRead...)
				evalResult.CharsUnused = evalResult.CharsRead
				return evalResult
			}
			charsUnused = resultExclude.CharsRead
		}

		// check close rule
		cs = newCharstreamPrepend(cs, charsUnused)
		result := inst.close.Eval(grammar, cs, NOT_SKIP) // do not skip space
		charsUnused = result.CharsUnused
		if result.Node != nil {
			if escaped {
				charsUsed = append(charsUsed, result.charsUsed()...)
				content.Chars = append(content.Chars, result.charsUsed()...)
				continue
			}
			closeResult = result
			if !inst.virtualClose {
				node.ChildNodes = append(node.ChildNodes, result.Node) // the 3rd child of block node is always close node
			}
			charsUsed = append(charsUsed, content.Chars...)
			charsUsed = append(charsUsed, result.CharsRead[:result.countCharsUsed()]...)
			break
		}
		cs = newCharstreamPrepend(cs, charsUnused)
		char := cs.Peek()
		if char == EOFChar {
			charsUsed = append(charsUsed, content.Chars...)
			evalResult.Error = fmt.Errorf("missing %s at EOF", inst.close.desc())
			evalResult.ErrIdx = charstream.Cursor()
			break
		}
		char = cs.Next()
		content.Chars = append(content.Chars, char)
		charsUnused = nil
	}
	evalResult.CharsRead = append(charsUsed, charsUnused...)
	if evalResult.Error != nil {
		evalResult.CharsUnused = evalResult.CharsRead
	} else {
		if inst.virtualClose {
			evalResult.CharsUnused = closeResult.CharsRead
		} else {
			evalResult.CharsUnused = charsUnused
		}
		evalResult.Node = node
	}

	return evalResult
}

// This function expects '<' as the 1st char in the input charstream.
func inBlock(grammar *Grammar, name string, charstream ICharstream) (*BlockRule, error) {
	block := &BlockRule{}
	cs := charstream
	openChar := cs.Next()
	if openChar != blockOpenSymbol {
		return nil, fmt.Errorf("block must start with '%c' char", blockOpenSymbol)
	}

	// get the first open rule
	openRule, err := grammar.parse(name, cs, []rune{' ', blockCloseSymbol, blockVirtualClose}) // inside block rule, the rules are separated by a space(s)
	if err != nil {
		return nil, fmt.Errorf("missing block open rule: %s", err)
	}
	block.open = openRule

	// the optional escape rule
	cs.SkipSpaces()
	char := cs.Peek()
	if char != negationChar { // must be the escape rule or close rule
		rule, err := grammar.parse(name, cs, []rune{' ', blockCloseSymbol, blockVirtualClose})
		if err != nil {
			return nil, err
		}
		cs.SkipSpaces()
		peek := cs.Peek()
		if peek == blockCloseSymbol || peek == blockVirtualClose {
			peek = cs.Next() // consume the close/virtualclose char
			if peek == blockVirtualClose {
				cs.SkipSpaces()
				peek = cs.Next()
				if peek != blockCloseSymbol {
					return nil, fmt.Errorf("block must end with a '%c': encountered '%c'", blockCloseSymbol, peek)
				}
				block.virtualClose = true
			}
			block.close = rule
			return block, nil
		}
		block.escape = rule
	}

	// the 0 to many exclude rules
	for {
		cs.SkipSpaces()
		char := cs.Peek()
		if char != negationChar { // exclude rule must start with a negationChar
			break
		}
		cs.Next() // skip the negationChar
		excludeRule, err := grammar.parse(name, cs, []rune{' ', blockCloseSymbol, blockVirtualClose})
		if err != nil {
			return nil, err
		}
		block.excludes = append(block.excludes, excludeRule)
	}

	// now the close rule
	closeRule, err := grammar.parse(name, cs, []rune{blockCloseSymbol, blockVirtualClose})
	if err != nil {
		return nil, err
	}
	block.close = closeRule
	closeChar := cs.Next()              // consume the close char
	if closeChar == blockVirtualClose { // optional there may be a virtual symbol in front of close symbol
		block.virtualClose = true
		cs.SkipSpaces()
		closeChar = cs.Next()
	}
	if closeChar != blockCloseSymbol {
		return nil, fmt.Errorf("block must end with a '%c': encountered '%c'", blockCloseSymbol, closeChar)
	}
	return block, nil
}
