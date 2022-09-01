package xbnf

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
)

const (
	VirtualSymbol         = '~'
	StickySymbol          = '+'
	NonDataSymbol         = '#'
	NegatedSymbol         = '^'
	ChoiceSymbol          = '|'
	ChoiceOrderSymbol     = '>'
	RepetitionOpenSymbol  = '{'
	RepetitionCloseSymbol = '}'
	OptionOpenSymbol      = '['
	OptionCloseSymbol     = ']'
	GroupOpenSymbol       = '('
	GroupCloseSymbol      = ')'
	blockOpenSymbol       = '<'
	blockCloseSymbol      = '>'
	blockVirtualClose     = '!'
	CharRangeSymbol       = '-'
	CharsSymbol           = '\''
	EscapeSymbol          = '\\'
)

type RuleRecord struct {
	name string
	rule IRule
	line int
}

func (inst *RuleRecord) Rule() IRule {
	return inst.rule
}

func NewGrammar() *Grammar {
	grammar := &Grammar{}
	grammar.ruleRecords = make(map[string]*RuleRecord)
	grammar.nonRoots = make(map[string][]string) // a rule is used by some other rule(s)
	grammar.terminals = make(map[string]*RuleRecord)
	grammar.rootRules = make(map[string]*RuleRecord)
	return grammar
}

type Grammar struct {
	fileName    string
	ruleRecords map[string]*RuleRecord // all rules along with name an line # defined in file
	nonRoots    map[string][]string    // key is rule that is referenced by at least 1 child rule
	terminals   map[string]*RuleRecord // all rules that has no child rule
	rootRules   map[string]*RuleRecord // rules that are not referenced by any rule
	maxLine     int
	//lock        sync.RWMutex
	validated bool
}

func (inst *Grammar) GetRecord(ruleName string) *RuleRecord {
	//inst.lock.RLock()
	//defer inst.lock.RUnlock()
	return inst.ruleRecords[ruleName]
}

func (inst *Grammar) GetRule(ruleName string) IRule {
	//inst.lock.RLock()
	//defer inst.lock.RUnlock()
	var rule IRule
	record, exists := inst.ruleRecords[ruleName]
	if exists {
		rule = record.rule
	}
	return rule
}

func (inst *Grammar) ParseRule(name string, ruleStr string) (IRule, error) {
	//ruleName, isVirtual, err := validateRuleName(name)
	ruleName, err := validateRuleName(name)
	if err != nil {
		return nil, err
	}
	ruleRecord, exists := inst.ruleRecords[ruleName]
	if exists {
		return nil, fmt.Errorf("rule '%s' already defined at line %d", name, ruleRecord.line)
	}
	ruleStr = strings.TrimSpace(ruleStr)
	var rule IRule
	cs := NewCharstreamString(ruleStr)
	if err != nil {
		return nil, err
	}
	r, err := inst.parse(ruleName, cs, []rune{EOFChar})
	if err != nil {
		return nil, err
	}
	rule = r
	rule.setName(ruleName)
	// rule.setVirtual(isVirtual)

	record := &RuleRecord{}
	inst.maxLine = inst.maxLine + 1
	record.line = inst.maxLine
	record.rule = rule
	record.name = ruleName
	inst.ruleRecords[ruleName] = record

	return rule, nil
}

// ruleDefintion contains rulename and definition string
func (inst *Grammar) AddRule(ruleDefinition string) (IRule, error) {
	//inst.lock.Lock()
	//defer inst.lock.Unlock()
	tokens := strings.SplitN(ruleDefinition, "=", 2)
	if len(tokens) != 2 {
		return nil, fmt.Errorf("invalid(missing =): %s", ruleDefinition)
	}
	name := strings.TrimSpace(tokens[0])
	ruleStr := strings.TrimSpace(tokens[1])
	return inst.ParseRule(name, ruleStr)
}

func (inst *Grammar) Serialize(withDetail bool) string {
	//inst.lock.RLock()
	//defer inst.lock.RUnlock()
	var maxNameLen int
	byLine := make(map[int]*RuleRecord)
	var lines []int
	for _, record := range inst.ruleRecords {
		if len(record.name) > maxNameLen {
			maxNameLen = len(record.name)
		}
		byLine[record.line] = record
		lines = append(lines, record.line)
	}
	maxNameLen = maxNameLen + 4 // all for [T] to indicate it's a terminal rule and optional ! for virtual
	sort.Ints(lines)
	nameFmtStr := fmt.Sprintf("%%-%ds", maxNameLen)
	var buf strings.Builder
	for _, line := range lines {
		record := byLine[line]
		buf.WriteString(fmt.Sprintf("L%04d: ", record.line))
		ruleName := record.name
		if record.rule.IsVirtual() {
			ruleName = "!" + ruleName
		}
		_, isTerminal := inst.terminals[record.name]
		if isTerminal {
			buf.WriteString(fmt.Sprintf(nameFmtStr, ruleName+"[T]"))
		} else {
			buf.WriteString(fmt.Sprintf(nameFmtStr, ruleName))
		}
		buf.WriteString(" = ")
		buf.WriteString(record.rule.String())
		buf.WriteString("\n    ")
		if withDetail {
			ruleStr := record.rule.StringWithIndent("    ")
			ruleStr = strings.ReplaceAll(ruleStr, "\n", "\n    ")
			buf.WriteString(ruleStr)
			buf.WriteRune('\n')
		}
	}
	return inst.StringRuleRelation(nameFmtStr) + "Rules:\n    " + buf.String()
}

func (inst *Grammar) StringRuleRelation(nameFmtStr string) string {
	var bufRoot strings.Builder
	var bufUsage strings.Builder
	bufRoot.WriteString("Root Rules: ")
	bufUsage.WriteString("Rule Reference Map: \n")
	for name := range inst.ruleRecords {
		children, exists := inst.nonRoots[name]
		if exists {
			bufUsage.WriteString(fmt.Sprintf("    "+nameFmtStr+": ", name))
			for _, user := range children {
				bufUsage.WriteString(fmt.Sprintf("%s,", user))
			}
			bufUsage.WriteRune('\n')
		} else {
			bufRoot.WriteString(fmt.Sprintf("%s,", name))
		}
	}
	return bufRoot.String() + "\n" + bufUsage.String()
}

func (inst *Grammar) AddUsage(ruleName string, userRule string) {
	//inst.lock.Lock()
	//defer inst.lock.Unlock()
	inst.nonRoots[ruleName] = append(inst.nonRoots[ruleName], userRule)
}

func (inst *Grammar) Validate() error {
	//inst.lock.Lock()
	//defer inst.lock.Unlock()
	nonTerminals := make(map[string]byte)
	for name, children := range inst.nonRoots {
		_, exists := inst.ruleRecords[name]
		if !exists {
			return fmt.Errorf("rule name '%s' referenced but not defined", name)
		}
		for _, child := range children {
			nonTerminals[child] = 0
		}
	}
	for name, record := range inst.ruleRecords {
		_, exists := nonTerminals[name]
		if exists {
			continue // none terminal
		}
		inst.terminals[name] = record
	}
	for name, ruleRecord := range inst.ruleRecords {
		_, exists := inst.nonRoots[name]
		if !exists {
			inst.rootRules[name] = ruleRecord
		}
	}
	return nil
}

// EvalEmbed parse a string by a rule in a fashion that it evals the rule at the 1st char, if match, it
// continues eval starting at the 1st char right after the matched chars. If not match, it starts match
// starting on the next char, and so on. A node is returned when there is at least 1 match, and has type
// "embed", the child nodes of the return node contains the rule matched node(s) and leaf node contains
// the unmatched texts
func (inst *Grammar) EvalEmbed(ruleName string, sample string) *EvalResult {
	evalResult := &EvalResult{}
	record := inst.GetRecord(ruleName)
	if record == nil {
		evalResult.Error = fmt.Errorf("rule '%s' not defined", ruleName)
		return evalResult
	}
	rule := record.Rule()
	node := &Node{RuleType: TypeEmbed, RuleName: ruleName}
	evalResult.Node = node
	cs := NewCharstreamString(sample)
	var result *EvalResult
	var text []rune
	for {
		result = rule.Eval(inst, cs, NOT_SKIP)
		cs = NewCharstreamPrepend(cs, result.CharsUnused)
		if result.Node == nil { // no match
			char := cs.Next()
			if char == EOFChar {
				if len(text) > 0 {
					textNode := &Node{RuleType: TypeText, RuleName: "_"}
					textNode.Chars = text
					text = nil
					node.ChildNodes = append(node.ChildNodes, textNode)
				}
				break
			}
			text = append(text, char)
		} else { // find a match
			if len(text) > 0 {
				textNode := &Node{RuleType: TypeText, RuleName: "_"}
				textNode.Chars = text
				text = nil
				node.ChildNodes = append(node.ChildNodes, textNode)
			}
			node.ChildNodes = append(node.ChildNodes, result.Node)
		}
	}
	return evalResult
}

func (inst *Grammar) EvalRule(ruleName string, sample string) *EvalResult {
	evalResult := &EvalResult{}
	record := inst.GetRecord(ruleName)
	if record == nil {
		evalResult.Error = fmt.Errorf("rule '%s' not defined", ruleName)
		return evalResult
	}
	rule := record.Rule()
	cs := NewCharstreamString(sample)
	return rule.Eval(inst, cs, SUGGEST_SKIP)
}

func (inst *Grammar) Eval(charstream ICharstream, simplifyLevel int) (*AST, error) {
	ast, err := inst.EvalRaw(charstream)
	if err != nil {
		return nil, err
	}
	if simplifyLevel == LevelRaw {
		return ast, nil
	}
	switch simplifyLevel {
	case LevelDataOnly:
		ast.MergeStickyNodes()
		ast.RemoveVirtualNodes()
		ast.RemoveNonDataNodes()
	case LevelNoVertual:
		ast.MergeStickyNodes()
		ast.RemoveVirtualNodes()
	default: // default is LevelBasic
		ast.MergeStickyNodes()
	}
	ast.RemoveRedundantNodes()
	return ast, nil
}

// Evaluate is the driver of parsing.
func (inst *Grammar) EvalRaw(charstream ICharstream) (*AST, error) {
	if charstream.Peek() == EOFChar {
		return nil, fmt.Errorf("EOF encountered")
	}
	ast := &AST{}
	cs := charstream
	var charsUnused []rune
	flagLeadingSpaces := SUGGEST_SKIP
	for {
		cs = NewCharstreamPrepend(cs, charsUnused)
		if cs.Peek() == EOFChar {
			break
		}
		resultsMatched := make(map[string]*EvalResult)
		resultsError := make(map[string]*EvalResult)
		var maxCharsRead []rune
		for name, ruleRecord := range inst.rootRules {
			cs = NewCharstreamPrepend(cs, maxCharsRead)
			result := ruleRecord.rule.Eval(inst, cs, flagLeadingSpaces)
			if len(maxCharsRead) < len(result.CharsRead) {
				maxCharsRead = result.CharsRead
			}
			if result.Node != nil {
				resultsMatched[name] = result
			}
			if result.Error != nil {
				resultsError[name] = result
			}
		}
		matches := len(resultsMatched)
		switch matches {
		case 0:
			var buf strings.Builder
			for _, result := range resultsError {
				buf.WriteString(fmt.Sprintf("%s; ", result.Error))
			}
			return nil, fmt.Errorf("no matches found: %s", buf.String())
		case 1:
			var result *EvalResult
			for _, r := range resultsMatched {
				result = r
			}
			ast.Nodes = append(ast.Nodes, result.Node)
			if result.Sticky {
				flagLeadingSpaces = SUGGEST_NOT_SKIP
			} else {
				flagLeadingSpaces = SUGGEST_SKIP
			}
			charsUnused = maxCharsRead[result.countCharsUsed():]
		default: // we found more than 1 matches, the most greedy one wins
			maxCountCharsUsed := 0
			var maxResult *EvalResult
			for _, result := range resultsMatched {
				countCharsUsed := result.countCharsUsed()
				if countCharsUsed == maxCountCharsUsed && maxResult != nil {
					return nil, fmt.Errorf("ambiguity found: %s(\"%s\") vs %s(\"%s\")",
						result.Node.RuleName, string(result.CharsRead),
						maxResult.Node.RuleName, string(maxResult.CharsRead))
				}
				if countCharsUsed >= maxCountCharsUsed {
					maxResult = result
					maxCountCharsUsed = countCharsUsed
				}
			}
			ast.Nodes = append(ast.Nodes, maxResult.Node)
			if maxResult.Sticky {
				flagLeadingSpaces = SUGGEST_NOT_SKIP
			} else {
				flagLeadingSpaces = SUGGEST_SKIP
			}
			charsUnused = maxCharsRead[maxResult.countCharsUsed():]
		}
		cs = charstream
	}
	if len(ast.Nodes) == 0 {
		return nil, fmt.Errorf("no AST node found")
	}
	return ast, nil
}

// parse a rule using the chars up to the terminator chars. The terminator chars won't be
// consumed after this call.
func (inst *Grammar) parse(name string, cs ICharstream, terminators []rune) (IRule, error) {
	var rules []IRule
	isVirtual := false
	isNonData := false
	cs.SkipSpaces() // spaces are insignificant at the begining of block rule
EXIT:
	for {
		char := cs.Peek()
		if char == EOFChar { // end of file
			break
		}
		for _, terminator := range terminators {
			if terminator == char {
				break EXIT
			}
		}

		// skip spaces
		//case ' ', '\n', '\t', '\r':
		if IsWhiteSpace(char) {
			cs.Next()
			continue
		}

		switch char {
		case VirtualSymbol, NonDataSymbol: // rule annotations: virtual, nondata, negated
			if char == VirtualSymbol {
				isVirtual = true
			} else {
				isNonData = true
			}
			cs.Next()
			// rule annotations can not appear before the ChoiceSymbol
			if cs.Peek() == ChoiceSymbol {
				return nil, fmt.Errorf("annotation symbol '%c' can not appear before choice symbol '%c'", char, ChoiceSymbol)
			}
			continue
		case '\'':
			r, err := inChar(cs)
			if err != nil {
				return nil, err
			}
			r.setVirtual(isVirtual)
			isVirtual = false
			r.setNonData(isNonData)
			isNonData = false
			rules = append(rules, r)
		case '\\':
			r, err := inUnicode(cs)
			if err != nil {
				return nil, err
			}
			r.setVirtual(isVirtual)
			isVirtual = false
			r.setNonData(isNonData)
			isNonData = false
			rules = append(rules, r)
		case '"':
			r, err := inString(name, cs)
			if err != nil {
				return nil, err
			}
			r.setVirtual(isVirtual)
			isVirtual = false
			r.setNonData(isNonData)
			isNonData = false
			rules = append(rules, r)
		case ChoiceSymbol, ChoiceOrderSymbol: // '|' or '>'
			size := len(rules)
			switch size {
			case 0:
				return nil, fmt.Errorf("Choice operator must in between rules")
			case 1:
				rule, err := inChoice(inst, name, rules[0], cs)
				if err != nil {
					return nil, err
				}
				rule.setVirtual(isVirtual)
				isVirtual = false
				rule.setNonData(isNonData)
				isNonData = false
				rules[0] = rule
			default:
				// last rule will participate in the choice, and be replaced by the choice rule
				rule, err := inChoice(inst, name, rules[size-1], cs)
				if err != nil {
					return nil, err
				}
				rule.setVirtual(isVirtual)
				isVirtual = false
				rule.setNonData(isNonData)
				isNonData = false
				rules[size-1] = rule
			}
		case OptionOpenSymbol: // '['
			rule, err := inOption(inst, name, cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			isVirtual = false
			rule.setNonData(isNonData)
			isNonData = false
			rules = append(rules, rule)
		case RepetitionOpenSymbol: // '{'
			rule, err := inRepetition(inst, name, cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			isVirtual = false
			rule.setNonData(isNonData)
			isNonData = false
			rules = append(rules, rule)
		case blockOpenSymbol: // '<'
			rule, err := inBlock(inst, name, cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			isVirtual = false
			rule.setNonData(isNonData)
			isNonData = false
			rules = append(rules, rule)
		case GroupOpenSymbol: //'('
			rule, err := inGroup(inst, name, cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			isVirtual = false
			rule.setNonData(isNonData)
			isNonData = false
			rules = append(rules, rule)
		case '/': // must be comments
			var comment []rune
			comment = append(comment, char)
			cs.Next()
			if cs.Next() != '/' {
				return nil, fmt.Errorf("invalid char '%c'", char)
			}
			comment = append(comment, char)
			for {
				char = cs.Next()
				if char == '\n' || char == EOFChar {
					break
				}
				comment = append(comment, char)
			}
			break EXIT
		default:
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_' {
				rule, err := inReference(inst, name, cs)
				if err != nil {
					return nil, err
				}
				rule.setVirtual(isVirtual)
				isVirtual = false
				rule.setNonData(isNonData)
				isNonData = false
				rules = append(rules, rule)
			} else {
				return nil, fmt.Errorf("invalid char '%c'", char)
			}
		}
	}
	switch len(rules) {
	case 0:
		return nil, fmt.Errorf("no rule found")
	case 1:
		return rules[0], nil
	default:
		rule := &ConcatenateRule{}
		rule.rules = rules
		return rule, nil
	}
}

// parseOne get one rule from the stream and return the rule, even if there are
// more content in the stream
func (inst *Grammar) parseOne(name string, cs ICharstream) (IRule, error) {
	isVirtual := false
	isNonData := false
	for {
		char := cs.Peek()
		switch char {
		case ' ', '\n', '\t', '\r':
			cs.Next()
			continue
		case EOFChar:
			return nil, fmt.Errorf("EOF encountered, no more rule")
		case VirtualSymbol, NonDataSymbol:
			if char == VirtualSymbol {
				isVirtual = true
			} else {
				isNonData = true
			}
			cs.Next()
			// rule annotations can not appear before the ChoiceSymbol
			if cs.Peek() == ChoiceSymbol {
				return nil, fmt.Errorf("annotation symbol '%c' can not appear before choice symbol '%c'", char, ChoiceSymbol)
			}
		case '\'':
			rule, err := inChar(cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			rule.setNonData(isNonData)
			return rule, nil
		case '\\':
			rule, err := inUnicode(cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			rule.setNonData(isNonData)
			return rule, nil
		case '"':
			rule, err := inString(name, cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			rule.setNonData(isNonData)
			return rule, nil
		case '[':
			rule, err := inOption(inst, name, cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			rule.setNonData(isNonData)
			return rule, nil
		case '(':
			rule, err := inGroup(inst, name, cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			rule.setNonData(isNonData)
			return rule, nil
		case '{':
			rule, err := inRepetition(inst, name, cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			rule.setNonData(isNonData)
			return rule, nil
		case blockOpenSymbol: // '<'
			rule, err := inBlock(inst, name, cs)
			if err != nil {
				return nil, err
			}
			rule.setVirtual(isVirtual)
			rule.setNonData(isNonData)
			return rule, nil
		default:
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_' {
				rule, err := inReference(inst, name, cs)
				if err != nil {
					return nil, err
				}
				rule.setVirtual(isVirtual)
				rule.setNonData(isNonData)
				return rule, nil
			}
			return nil, fmt.Errorf("invalid char '%c' for any rule", char)
		}
	}
}

func NewGrammarFromFile(filexbnf string) (*Grammar, error) {
	xbnfText, err := ioutil.ReadFile(filexbnf)
	if err != nil {
		return nil, fmt.Errorf("can't real file: %s", err)
	}
	grammar, err := NewGrammarFromString(string(xbnfText))
	if err == nil {
		grammar.fileName = filexbnf
		return grammar, nil
	}
	err = fmt.Errorf("%s: %s", filexbnf, err)
	return nil, err
}

func NewGrammarFromString(grammarText string) (*Grammar, error) {
	grammar := NewGrammar()
	lines := strings.Split(grammarText, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "//") {
			grammar.maxLine = i + 1
			continue // empty line or comments
		}
		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) != 2 {
			return nil, fmt.Errorf("L#%d: invalid(missing =): %s", i+1, line)
		}
		name := strings.TrimSpace(tokens[0])
		ruleStr := strings.TrimSpace(tokens[1])
		_, err := grammar.ParseRule(name, ruleStr)
		if err != nil {
			return nil, fmt.Errorf("L#%d: rule [%s] - %s", i+1, name, err)
		}
		grammar.maxLine = i + 1
	}
	err := grammar.Validate()
	if err != nil {
		return nil, err
	}
	return grammar, nil
}
