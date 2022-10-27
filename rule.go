package xbnf

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

const (
	stateMatched    = iota // evaluation process found a match
	stateDismatched        // rule evaluation process found a dismatch
	statePending           // rule evaluation process can't make a decision yet, need more input
)

// The type of a node, which is decided by the type of rule that generates the node
type Type string

// Rule types
const (
	TypeEOF         = Type("EOF") // Predefined rule type
	TypeChar        = Type("char")
	TypeRange       = Type("range")
	TypeChars       = Type("chars")
	TypeString      = Type("string")
	TypeGroup       = Type("group")
	TypeOption      = Type("option")
	TypeBlock       = Type("block")
	TypeReference   = Type("reference")
	TypeRepetition  = Type("repetition")
	TypeChoice      = Type("choice")
	TypeConcatenate = Type("concatenate")

	TypeEmbed = Type("embed") // just for node parsed using EvalEmbed
	TypeText  = Type("text")  // just for free text node parsed using EvalEmbed
)

// leading spaces handling flag
const (
	SUGGEST_NOT_SKIP = 0
	SUGGEST_SKIP     = 1
	NOT_SKIP         = 3 // used mostly in explicit rule evaluation call
)

const negationChar = '^' // char can be places in front of a rule defintion to means negation

type EvalResult struct {
	Node *Node

	Error  error
	ErrIdx int // the index where the error occurs

	// total chars read from the Charstream, including those CharsUnused, if there
	// is any, at the end
	CharsRead []rune

	// In order to evaluate against the rule, the chars read from Charstream but not used by
	// the Node. This is for continuous rule evaluations on a Charstream
	CharsUnused []rune

	// during parsing, whether space(s) are allowed between the node of this result and its
	// right neighber rule
	Sticky bool
}

func (inst *EvalResult) MergeStickyNodes() {
	if inst.Node == nil {
		return
	}
	inst.Node.MergeStickyNodes()
}

func (inst *EvalResult) countCharsUsed() int {
	return len(inst.CharsRead) - len(inst.CharsUnused)
}

func (inst *EvalResult) charsUsed() []rune {
	return inst.CharsRead[:len(inst.CharsRead)-len(inst.CharsUnused)]
}

func (inst *EvalResult) StringTree(config *NodeTreeConfig) string {
	var buf strings.Builder
	buf.WriteString("EvalResult")
	buf.WriteString(fmt.Sprintf("\n├───CharsRead[%d]  :%s*", len(inst.CharsRead), string(inst.CharsRead)))
	buf.WriteString(fmt.Sprintf("\n├───CharsUnused[%d]:%s*", len(inst.CharsUnused), string(inst.CharsUnused)))
	if inst.Node == nil {
		buf.WriteString("\n└───Node: nil")
		if inst.Error != nil {
			buf.WriteString(fmt.Sprintf("; %s", inst.Error))
		}
	} else {
		nodeStr := inst.Node.StringTree(config)
		nodeStr = strings.ReplaceAll(nodeStr, "\n", "\n    ")
		buf.WriteString(fmt.Sprintf("\n└───Node: %s", nodeStr))
	}
	return buf.String()
}

type output struct {
	text  []rune // the text has been evaluate so far
	state int    // the current state of the evaluation process
	err   error  // if there is error, used only when the state == stateDismatched
}

type input struct {
	char   rune
	wait   *sync.WaitGroup
	output chan *output
}

type virtual bool

func (inst *virtual) IsVirtual() bool {
	return bool(*inst)
}

func (inst *virtual) setVirtual(isVirtual bool) {
	*inst = virtual(isVirtual)
}

type IRule interface {
	setName(ruleName string)
	setVirtual(isVirtual bool)
	setNonData(isNonData bool)

	desc() string

	// Returns the name of the rule.
	Name() string

	IsVirtual() bool

	IsNonData() bool

	// Eval evaluates a charstream according to this rule. An EvalResult is returned. If resulting
	// Node is nil, it means this rule is not match, in which case an error may also be return if
	// the rule must be matched, cases such as open quote without close quote, certain unique words must
	// followed by specific text, etc.
	// The flagLeadingSpaces demands how the rule should handle leading white spaces in the stream:
	// 0 - suggest NOT to skip leading spaces, 1 - suggest to skip leading spaces, 3 - DO NOT skip
	// leading space
	Eval(grammar *Grammar, charStram ICharstream, flagLeadingSpaces int) *EvalResult

	// Evaluate input according to this rule. This is meant to be run in its own goroutine
	// for concurrent evalution of an input for multiple rules
	// Evaluate(*Grammar, chan *input)

	// Returns the rule definition string in xbnf format
	String() string

	StringWithIndent(indent string) string
}

type ruleBase struct {
	name    string
	virtual bool
	nondata bool
	//negated bool
}

func (inst *ruleBase) Name() string {
	return inst.name
}

func (inst *ruleBase) setName(ruleName string) {
	inst.name = ruleName
}

func (inst *ruleBase) IsVirtual() bool {
	return inst.virtual
}

func (inst *ruleBase) setVirtual(isVirtual bool) {
	inst.virtual = isVirtual
}

func (inst *ruleBase) IsNonData() bool {
	return inst.nondata
}

func (inst *ruleBase) setNonData(isNonData bool) {
	inst.nondata = isNonData
}

func (inst *ruleBase) Eval(grammar *Grammar, cs ICharstream) *EvalResult {
	return &EvalResult{Error: errors.New("Eval() not implemented")}
}

func (inst *ruleBase) String() string {
	return "#!NotImplemented!"
}
func (inst *ruleBase) annotation() []rune {
	var annotation []rune
	if inst.virtual {
		annotation = append(annotation, VirtualSymbol)
	}
	if inst.nondata {
		annotation = append(annotation, NonDataSymbol)
	}
	//if inst.negated {
	//	annotation = append(annotation, NegatedSymbol)
	//}
	return annotation
}

func (inst *ruleBase) StringWithIndent(indent string) string {
	return "#!NotImplemented!"
}

func leadingWhiteSpace(input []rune) []rune {
	var wspaces []rune
	for _, char := range input {
		if IsWhiteSpace(char) {
			wspaces = append(wspaces, char)
			continue
		}
		break
	}
	return wspaces
}

// Given a ruleName as is in the xbnf rule definition, checks if the name
// is valid. Return err is not. The ruleName may optionally prefixed with
// an VirtualChar, meaning the rule defined is virtual. The return
// bool is true if the name is virtual.
func validateRuleName(ruleName string) (validRuleName string, err error) {
	//if strings.HasPrefix(ruleName, string(VirtualSymbol)) {
	//	isVirtual = true
	//	ruleName = ruleName[1:]
	//}
	if ruleName == "" {
		err = fmt.Errorf("empty rule name")
		return
	}
	firstChar := ruleName[0]
	if !('a' <= firstChar && firstChar <= 'z') && !('A' <= firstChar && firstChar <= 'Z') {
		// must start with a letter
		err = fmt.Errorf("rule name must starts with a letter character: '%c' is invalid", firstChar)
		return
	}
	for _, char := range ruleName[1:] {
		switch {
		case ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_' || ('0' <= char && char <= '9'):
		default:
			err = fmt.Errorf("rule name must consist of letters, 0-9 digits, or '_': '%s' is invalid", ruleName)
			return
		}
	}
	validRuleName = ruleName
	return
}

// Given a list of rules, find the most greedy match. If no rule matches, return a evalResult with
// nil node.
func mostGreedy(grammar *Grammar, charStream ICharstream, flagLeadingSpaces int, rules []IRule) *EvalResult {
	evalResult := &EvalResult{}
	var resultsMatched []*EvalResult
	var maxCharsRead []rune // the max chars read to evaluate all rules
	cs := charStream
	for _, rule := range rules {
		cs := newCharstreamPrepend(cs, maxCharsRead)
		result := rule.Eval(grammar, cs, flagLeadingSpaces)
		if len(maxCharsRead) < len(result.CharsRead) {
			maxCharsRead = result.CharsRead
		}
		if result.Node != nil {
			resultsMatched = append(resultsMatched, result)
		}
	}
	matches := len(resultsMatched)
	var resultFound *EvalResult
	if matches == 0 {
		// no match, all CharsRead are unused
		evalResult.CharsRead = maxCharsRead
		evalResult.CharsUnused = evalResult.CharsRead
		return evalResult
	} else if matches == 1 {
		resultFound = resultsMatched[0]
	} else { // we found multiple matches, get the one that is most greedy, ie. uses most chars
		maxUsed := 0
		var maxResult *EvalResult
		for _, result := range resultsMatched {
			countCharsUsed := result.countCharsUsed()
			if countCharsUsed < maxUsed {
				continue
			}
			if countCharsUsed == maxUsed && maxResult != nil {
				// we found ambiguity
				var buf strings.Builder
				buf.WriteString("ambiguity found:")
				buf.WriteString(fmt.Sprintf("\t\n%s", string(maxResult.Node.Text())))
				buf.WriteString(fmt.Sprintf("\t\n%s", string(result.Node.Text())))
				evalResult.Error = errors.New(buf.String())
				return evalResult
			}
			maxUsed = countCharsUsed
			maxResult = result
		}
		resultFound = maxResult
	}
	evalResult.CharsRead = maxCharsRead
	evalResult.Node = resultFound.Node
	evalResult.Sticky = resultFound.Sticky
	evalResult.CharsUnused = evalResult.CharsRead[resultFound.countCharsUsed():]
	return evalResult
}
