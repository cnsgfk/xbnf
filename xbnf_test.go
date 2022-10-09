package xbnf

import (
	"fmt"
	"strings"
	"testing"
)

func init() {
}

func testRule(t *testing.T, grammar *Grammar, ruleName string, sample string, expected string) {
	logSample := sample
	t.Logf("====> SAMPLE:%s*", logSample)
	record := grammar.GetRecord(ruleName)
	if record == nil {
		t.Errorf("XXXX> Failed: rule '%s' not defined", ruleName)
		return
	}
	rule := record.Rule()
	t.Logf("====> RULE  : %s = %s", ruleName, rule.String())
	cs := NewCharstreamFromString(sample)
	result := rule.Eval(grammar, cs, SUGGEST_SKIP)
	t.Logf("CharsRead:%s*", string(result.CharsRead))
	t.Logf("CharsUnused:%s*", string(result.CharsUnused))
	if result.Node == nil {
		if expected == "" {
			t.Logf("====> Passed")
		} else {
			t.Errorf("XXXX> Failed: %s", result.Error)
		}
		return
	}
	// t.Logf("Actual RAW Node: \n%s", result.Node.StringTree(nil))
	result.Node.MergeStickyNodes()
	result.Node.RemoveVirtualNodes()
	result.Node.RemoveRedundantNodes()
	t.Logf("Actual Node: \n%s", result.Node.StringTree(nil))
	actual := result.Node.Text()
	if expected != string(actual) {
		t.Errorf("XXXX> Failed: expected vs actual\n%s\n%s", expected, string(actual))
		return
	}
	result.Node.RemoveNonDataNodes()
	t.Logf("====> Passed: %s", string(actual))
}

type Expected interface {
	StringTree() string
	Verdict(t *testing.T, actualResult *EvalResult) error
}

type ExpectedEvalResult struct {
	CharsRead   string
	CharsUnused string
	Text        string
}

func (inst *ExpectedEvalResult) StringTree() string {
	var buf strings.Builder
	buf.WriteString("Expected EvalResult")
	buf.WriteString(fmt.Sprintf("\n├───CharsRead[%d]  :%s*", len(inst.CharsRead), inst.CharsRead))
	buf.WriteString(fmt.Sprintf("\n├───CharsUnused[%d]:%s*", len(inst.CharsUnused), inst.CharsUnused))
	buf.WriteString(fmt.Sprintf("\n└───Text:%s*", inst.Text))
	return buf.String()
}

func (inst *ExpectedEvalResult) Verdict(t *testing.T, actualResult *EvalResult) error {
	if inst.CharsRead != string(actualResult.CharsRead) {
		return fmt.Errorf("Unexpected CharsRead:%s*", string(actualResult.CharsRead))
	}
	if inst.CharsUnused != string(actualResult.CharsUnused) {
		return fmt.Errorf("Unexpected CharsUnused:%s*", string(actualResult.CharsUnused))
	}
	if actualResult.Node == nil {
		if inst.Text == "" {
			return nil
		}
		return fmt.Errorf("Expecting node text:%s*", inst.Text)
	}
	text := actualResult.Node.Text()
	if string(text) != inst.Text {
		return fmt.Errorf("Unexpecting node text:%s*", string(text))
	}

	return nil
}

type ExpectedText string

func (inst ExpectedText) StringTree() string {
	return fmt.Sprintf("Expected Text:%s*", inst)
}

func (inst ExpectedText) Verdict(t *testing.T, actualResult *EvalResult) error {
	if actualResult.Node == nil {
		if inst == "" {
			return nil
		}
		return fmt.Errorf("Expecting Text:%s*", inst)
	}
	text := actualResult.Node.Text()
	if string(text) != string(inst) {
		return fmt.Errorf("Unexpected Text:%s*", string(text))
	}

	return nil
}

func evalRule(t *testing.T, expectedResult Expected, grammar *Grammar, ruleName string, sample string) {
	logSample := sample
	//if len(logSample) > 40 {
	//	logSample = logSample[0:40] + "..."
	//}
	t.Logf("====> SAMPLE: \"%s\"", logSample)
	record := grammar.GetRecord(ruleName)
	if record == nil {
		t.Errorf("XXXX> Failed: rule '%s' not defined", ruleName)
		return
	}
	rule := record.rule
	t.Logf("====> RULE  : %s = %s", ruleName, rule.String())
	cs := NewCharstreamFromString(sample)
	result := rule.Eval(grammar, cs, SUGGEST_SKIP)
	verdict := expectedResult.Verdict(t, result)
	//result.MergeLeafNodes()
	treeConfig := DefaultNodeTreeConfig()
	treeConfig.PrintNonleafNodeText = true
	if verdict == nil {
		t.Logf("====> Passed: Actual EvalResult\n%s", result.StringTree(treeConfig))
	} else { // found node
		t.Errorf(fmt.Sprintf("XXXX> Failed: %s\n%s\nActual %s",
			verdict, expectedResult.StringTree(), result.StringTree(treeConfig)))
	}
}

func TestGrammarparse(t *testing.T) {
	tester := func(g *Grammar, ruleStr string, expect string) {
		t.Logf("====> Sample: %s", ruleStr)
		t.Logf("====> Expect: %s", expect)
		cs := NewCharstreamFromString(ruleStr)
		rule, err := g.parse("", cs, []rune{EOFChar})
		if err != nil {
			t.Logf("====> Result: %s", err)
			if expect == "" {
				t.Logf("====> Passed!")
				return
			}
			t.Errorf("****> Failed: unexpected error")
			return
		}
		ruleStr = rule.String()
		t.Logf("====> Result: %s - %T", ruleStr, rule)
		if expect == "" || ruleStr != expect {
			t.Errorf("****> Failed: unexpected result")
			return
		}
		t.Logf("====> Passed!")
	}
	grammar := NewGrammar()
	/*
	 */
	tester(grammar, `"0x"`, `"0x"`)
	tester(grammar, `"{" attr { ',' attr } "}"`, `"{" attr { ',' attr } "}"`)
	tester(grammar, `'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z'`,
		`'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z'`)
	tester(grammar, `"0x" { letter { letter | digit_dec | '_' } } end`, `"0x" { letter { letter | digit_dec | '_' } } end`)
	tester(grammar, `"0x" digit_hex { digit_hex }`, `"0x" digit_hex { digit_hex }`)
	tester(grammar, `"0x" { digit_hex}<2,>`, `"0x" { digit_hex }<2,0>`)
	tester(grammar, `"0x" { digit_hex}<2,0>`, `"0x" { digit_hex }<2,0>`)
	tester(grammar, `"0x"{digit_hex}`, `"0x" { digit_hex }`)
	tester(grammar, `"0x" { digit_hex}<4,3>`, ``)
	tester(grammar, `"now" | "fmt" "0x" digit_hex | "limit"`, `"now" | "fmt" "0x" digit_hex | "limit"`)
	tester(grammar, `"now" | "fmt" "0x" digit_hex`, `"now" | "fmt" "0x" digit_hex`)
	tester(grammar, `"0x" digit_hex`, `"0x" digit_hex`)
	tester(grammar, `"0x" digit_hex 'abc'`, `"0x" digit_hex 'abc'`)
	tester(grammar, `"now" | "fmt" | "limit"`, `"now" | "fmt" | "limit"`)
	tester(grammar, `"True" | "true" | "False" | "false"`, `"True" | "true" | "False" | "false"`)
	tester(grammar, `string_dq | string_sq | string_ml`, `string_dq | string_sq | string_ml`)
	tester(grammar, `term ("+" | "-") expr`, `term ( "+" | "-" ) expr`)
	tester(grammar, `[ "-" ] factor`, `[ "-" ] factor`)
	tester(grammar, `[ "-" | "+" ] factor`, `[ "-" | "+" ] factor`)
	tester(grammar, `[ currency ] digit_dec [ '.' {digit_dec}<2,4> ]`, `[ currency ] digit_dec [ '.' { digit_dec }<2,4> ]`)
	tester(grammar, `"0x" { letter { letter | digit_dec | '_' }<0,4> }+ end`, `"0x" { letter { letter | digit_dec | '_' }<0,4> }+ end`)
	tester(grammar, `"0x" { digit_hex }+`, `"0x" { digit_hex }+`)
	tester(grammar, `"0x" { digit_hex }* abc`, `"0x" { digit_hex } abc`)
	tester(grammar, `"0x" { digit_hex }<0,0> abc`, `"0x" { digit_hex } abc`)
	tester(grammar, ` ~"0x" { digit_hex }<0,0> abc`, `~"0x" { digit_hex } abc`)
	tester(grammar, `~( \u000A|EOF)`, `~( \u000A | EOF )`)
}

func TestGrammarParseRule(t *testing.T) {
	tester := func(t *testing.T, g *Grammar, ruleName string, ruleStr string, expect string) {
		t.Logf("====> Sample: %s = %s", ruleName, ruleStr)
		t.Logf("====> Expect: %s", expect)
		rule, err := g.ParseRule(ruleName, ruleStr)
		if err != nil {
			if expect == "" {
				t.Logf("====> Passed: %s", err)
				return
			}
			t.Errorf("****> Failed: unexpected error: %s", err)
			return
		}
		ruleStr = rule.String()
		t.Logf("====> Result: %s", ruleStr)
		if expect == "" || ruleStr != expect {
			t.Errorf("****> Failed: unexpected result")
			return
		}
		t.Logf("====> Passed!")
	}
	grammar := NewGrammar()
	t.Run("test1", func(t *testing.T) {
		tester(t, grammar, "string_sq", ` < '\'' '\\' '\'' >`, `<'\'' '\\' '\''>`)
	})
	t.Run("test2", func(t *testing.T) {
		tester(t, grammar, "line_comment", ` < '//' ( \u000A | EOF ) > `, `<'//' ( \u000A | EOF )>`)
	})
	t.Run("test3", func(t *testing.T) {
		tester(t, grammar, "string_ml", " <'`' '\\\\' '`'> ", "<'`' '\\\\' '`'>")
	})
	t.Run("test4", func(t *testing.T) {
		tester(t, grammar, "space", `~{ \u0020|\u0009|\u000A|\u000D}`, `~{ \u0020 | \u0009 | \u000A | \u000D }`)
	})
	t.Run("test5", func(t *testing.T) {
		tester(t, grammar, "baretext", ` <"" ',' !> `, `<"" ',' !>`)
	})
	t.Run("range1", func(t *testing.T) {
		tester(t, grammar, "range1", `'A'-'Z'`, `'A'-'Z'`)
	})
	t.Run("range2", func(t *testing.T) {
		tester(t, grammar, "range2", ` '0'-'9' `, `'0'-'9'`)
	})
	t.Run("range3", func(t *testing.T) {
		tester(t, grammar, "range3", ` '\u0041'-'Z' `, `\u0041-'Z'`)
	})
	t.Run("range4", func(t *testing.T) {
		tester(t, grammar, "range4", ` \u0041-'Z' `, `\u0041-'Z'`)
	})
	t.Run("range5", func(t *testing.T) {
		tester(t, grammar, "range5", ` 'A'-'\u005A' `, `'A'-\u005A`)
	})
	t.Run("range6", func(t *testing.T) {
		tester(t, grammar, "range6", ` 'A'-\u005A `, `'A'-\u005A`)
	})
	t.Run("range7", func(t *testing.T) {
		tester(t, grammar, "range7", ` '\u0041'-'\u005A' `, `\u0041-\u005A`)
	})
	t.Run("range8", func(t *testing.T) {
		tester(t, grammar, "range8", ` \u0041-'\u005A' `, `\u0041-\u005A`)
	})
	t.Run("range9", func(t *testing.T) {
		tester(t, grammar, "range9", ` '\u0041'-\u005A `, `\u0041-\u005A`)
	})
	t.Run("range10", func(t *testing.T) {
		tester(t, grammar, "range10", ` \u0041-\u005A `, `\u0041-\u005A`)
	})
	t.Run("choice1", func(t *testing.T) {
		tester(t, grammar, "letter_uppercase",
			`'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z'`,
			`'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z'`)
	})
	t.Run("choice2", func(t *testing.T) {
		tester(t, grammar, "ordered_choice",
			`'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' > 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z'`,
			`'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' > 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z'`)
	})
	t.Run("choice3", func(t *testing.T) {
		tester(t, grammar, "ordered_choice1",
			`'A' > 'B' | 'C' > 'D' | 'E' > 'F' `,
			`'A' > 'B' | 'C' > 'D' | 'E' > 'F'`)
	})

	t.Run("choice4", func(t *testing.T) {
		tester(t, grammar, "ordered_choice2",
			` "value" > <"" ( "," | ")" ) !> `,
			`"value" > <"" ( "," | ")" ) !>`)
	})

}

func TestGrammarAddRule(t *testing.T) {
	tester := func(g *Grammar, ruleDefinition string, expect string) {
		t.Logf("====> Sample: %s", ruleDefinition)
		t.Logf("====> Expect: %s", expect)
		rule, err := g.AddRule(ruleDefinition)
		if err != nil {
			t.Logf("====> Result: %s", err)
			if expect == "" {
				t.Logf("====> Passed!")
				return
			}
			t.Errorf("****> Failed: unexpected error")
			return
		}
		ruleStr := fmt.Sprintf("%s = %s", rule.Name(), rule.String())
		t.Logf("====> Result: %s", ruleStr)
		if expect == "" || ruleStr != expect {
			t.Errorf("****> Failed: unexpected result")
			return
		}
		t.Logf("====> Passed!")
	}
	g := NewGrammar()
	tester(g, `letter_uppercase    = 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z'`,
		`letter_uppercase = 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z'`)
	tester(g, `digit_dec           = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'`,
		`digit_dec = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'`)
	tester(g, `digit_hex           = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f'`,
		`digit_hex = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f'`)
	tester(g, `digit_oct           = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7'`, `digit_oct = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7'`)
	tester(g, `digit_bin           = '0' | '1'`, `digit_bin = '0' | '1'`)
	tester(g, `letter              = letter_lowercase | letter_uppercase`, `letter = letter_lowercase | letter_uppercase`)
	tester(g, `comment_block   = <'/*' '*/'>`, `comment_block = <'/*' '*/'>`)
}

func TestEvalRule(t *testing.T) {
	grammar, err := NewGrammarFromString(`
	letter_uppercase    = 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z'
	digit_dec           = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'
	digit_hex           = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f'
	digit_oct           = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7'
	digit_bin           = '0' | '1'
	letter              = letter_lowercase | letter_uppercase
	alphanumeric        = letter_lowercase | letter_uppercase | digit_dec
	letter_lowercase    = 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z'	
	identifier          = letter { alphanumeric | '_' }
	comment_line    = <'//' ( \u000A | EOF )>
	comment_block   = <'/*' '*/'>
	string   		= string_dq | string_sq | string_ml
	string_dq       = <'"' '\\' ^\u000A '"'>
	string_sq       = <\u0027 '\\' ^\u000A \u0027>
	string_ml       = <\u0060 '\\' \u0060>
	string_bq       = <'[' '\\' ']'>
	bool                = bool_tf | bool_yn
	bool_tf             = "True" | "true" | "False" | "false"
	bool_yn             = "Yes" | "yes" | "No" | "no"
	integer             = "" [ '+' | '-' ] (integer_dec | integer_oct | integer_hex | integer_bin )
	integer_dec         = digit_dec { digit_dec } 
	integer_hex         = '0x' digit_hex { digit_hex }
	integer_oct         = '0o' digit_oct { digit_oct }
	integer_bin         = '0b' digit_bin { digit_bin }
	literal             = string | integer
	int                 = integer
	names               = identifier { "" identifier }
	opr					= ( "+" | "-" )
	expr                = integer { opr literal }
	`)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	t.Logf("\nGrammar:%s", grammar.Serialize(false))
	t.Run("test1", func(t *testing.T) {
		evalRule(t, ExpectedText("12 + '12'"), grammar, "expr", " 12+  '12'abc")
	})
	t.Run("test2", func(t *testing.T) {
		evalRule(t, ExpectedText("myName_"), grammar, "identifier", "myName_ ")
	})
	t.Run("test3", func(t *testing.T) {
		evalRule(t, ExpectedText("0xF45A03F"), grammar, "integer_hex", "0xF45A03F")
	})
	t.Run("test4", func(t *testing.T) {
		evalRule(t, ExpectedText("1234567"), grammar, "integer_dec", "1234567")
	})
	t.Run("test5", func(t *testing.T) {
		evalRule(t, ExpectedText("0o1234567"), grammar, "integer_oct", "0o1234567")
	})
	t.Run("test6", func(t *testing.T) {
		evalRule(t, ExpectedText("0b10110010"), grammar, "integer_bin", "0b10110010234567")
	})
	t.Run("test7", func(t *testing.T) {
		evalRule(t, ExpectedText("true"), grammar, "bool_tf", "true")
	})
	t.Run("test8", func(t *testing.T) {
		evalRule(t, ExpectedText("Yes"), grammar, "bool_yn", "Yes")
	})
	t.Run("test9", func(t *testing.T) {
		evalRule(t, ExpectedText("No"), grammar, "bool", " No ")
	})
	t.Run("test10", func(t *testing.T) {
		evalRule(t, ExpectedText("0xF0F0F0"), grammar, "integer", "0xF0F0F0")
	})
	t.Run("test11", func(t *testing.T) {
		evalRule(t, ExpectedText("1 + 0xFF + 3"), grammar, "expr", "1+0xFF+3")
	})
	t.Run("test12", func(t *testing.T) {
		evalRule(t, ExpectedText("\"double quoted string\""), grammar, "string_dq", `"double quoted string"`)
	})
	t.Run("test13", func(t *testing.T) {
		evalRule(t, ExpectedText("[between square brackets]"), grammar, "string_bq", "[between square brackets]")
	})
	t.Run("test14", func(t *testing.T) {
		evalRule(t, ExpectedText("// comments here"), grammar, "comment_line", "  // comments here")
	})
	t.Run("test15", func(t *testing.T) {
		evalRule(t, ExpectedText("myName_"), grammar, "identifier", "myName_ abc")
	})
	t.Run("test16", func(t *testing.T) {
		evalRule(t, ExpectedText("-12"), grammar, "int", " -12 ")
	})
	t.Run("test17", func(t *testing.T) {
		evalRule(t, ExpectedText("-1"), grammar, "int", " -1 ")
	})
	t.Run("test18", func(t *testing.T) {
		evalRule(t, ExpectedText(""), grammar, "integer_dec", "F1 ")
	})
	t.Run("test19", func(t *testing.T) {
		evalRule(t, ExpectedText("/* comments here*/"), grammar, "comment_block", "  /* comments here*/  ")
	})
	t.Run("test20", func(t *testing.T) {
		evalRule(t, ExpectedText(""), grammar, "integer", "'12'")
	})
	t.Run("test21", func(t *testing.T) {
		evalRule(t, ExpectedText(""), grammar, "integer", "  '12'")
	})
	t.Run("test22", func(t *testing.T) {
		testRule(t, grammar, "expr", "1-12  +3  -14", "1 - 12 + 3 - 14")
	})
	t.Run("test23", func(t *testing.T) {
		evalRule(t, ExpectedText("14"), grammar, "integer_dec", "14")
	})
	t.Run("test24", func(t *testing.T) {
		evalRule(t, ExpectedText("sean ivy kevin"), grammar, "names", " sean ivy kevin ")
	})
	t.Run("test25", func(t *testing.T) {
		evalRule(t, ExpectedText("1"), grammar, "int", " 1 ")
	})
	t.Run("test26", func(t *testing.T) {
		evalRule(t, ExpectedText("0xFF + 45 - 0b0101"), grammar, "expr", " 0xFF+45-0b0101")
	})
	t.Run("test27", func(t *testing.T) {
		evalRule(t, ExpectedText("0xF45A"), grammar, "integer_hex", "0xF45A")
	})
	t.Run("test28", func(t *testing.T) {
		evalRule(t, ExpectedText("0xF45A + 12345"), grammar, "expr", "0xF45A+12345.67")
	})
}

func TestArithmeticRule(t *testing.T) {
	grammar, err := NewGrammarFromString(`
		SPACE   = \u0020  // space
		TAB     = \u0009  // horizontal tab
		NL      = \u000A  // new line
		CR      = \u000D  // carriage return
		space       = ~{ SPACE | TAB | NL | CR }
		digit_dec   = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'
		digit_hex   = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f'
		digit_oct   = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7'
		digit_bin   = '0' | '1'
		integer_dec = digit_dec { digit_dec } 
		integer_hex = "0x" digit_hex { digit_hex }
		integer_oct = "0o" digit_oct { digit_oct }
		integer_bin = "0b" digit_bin { digit_bin }
		float       = digit_dec '.' digit_dec { digit_dec }
		integer     = integer_dec | integer_oct | integer_hex | integer_bin
		literal		= integer | float
		factor		= literal | ( "(" expr ")" )
		term		= factor { ( "*" | "/" ) factor }
		expr		= term { ("+" | "-") term }
		exprs		= expr
		// spaces		= { {space}+ "efg" }
		spaces		= { { ~{'a'} }+ "efg" }
	`)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	t.Logf("\nGrammar:%s", grammar.Serialize(false))
	evalRule(t, ExpectedText("12"), grammar, "term", "12")
	evalRule(t, ExpectedText("12"), grammar, "term", " 12 ")
	evalRule(t, ExpectedText("12 * 3"), grammar, "term", " 12 *    3 ")
	evalRule(t, ExpectedText("1 / 12 * 12"), grammar, "term", "1/12*12")
	evalRule(t, ExpectedText("12"), grammar, "expr", "12")
	evalRule(t, ExpectedText("12 + 1"), grammar, "expr", "12 +   1")
	evalRule(t, ExpectedText("12"), grammar, "literal", "12 +   1")
	evalRule(t, ExpectedText("12 * 3 / 4"), grammar, "expr", "12*3/4")
	evalRule(t, ExpectedText("2 - 12 / 4"), grammar, "expr", "2-12/4 ")
	evalRule(t, ExpectedText("3 / 4 - 12 / 4"), grammar, "expr", " 3/4 -12/4 ")
	evalRule(t, ExpectedText("( 4 + 5 )"), grammar, "expr", "(4 +5) ")
	evalRule(t, ExpectedText("( 4 + 5 ) / 5"), grammar, "expr", " (4 +5) /5  ")
	evalRule(t, ExpectedText("124 * 23 - 23 + 3 / 4 - 12 / 4"), grammar, "expr", "124*23 - 23 + 3/4 -12/4 ")
	evalRule(t, ExpectedText("12 - 2 * 3"), grammar, "expr", " 12-2   *3  ")
	evalRule(t, ExpectedText(""), grammar, "spaces", "")
}

func TestBlock(t *testing.T) {
	grammar, err := NewGrammarFromString(`
		string_dq = <'"' '\\' '"'>
		string_sq = <"'" '\\' '\''>
		comment_line = < '//' ( \u000A | EOF ) >
		comment_ml = < '/*'  '*/' >
		string     = <\u0022 '\\' ^\u000A \u0022>
		string1     = <"\u0022" '\\' ^\u000A "\u0022"> // \u0022 is double quote
		key = < "" '\\' ^\u000A ('='|':') >
		value = < "" '\\' ~( \u000A | EOF )> // NewLine unicode is u000A
		keyempty = < "" ^'=' ~( \u000A | EOF )> 
		property = keyempty | ( key value )
		bool = "true" | "false"
		const = bool |string_dq | string_sq
		baretext = < "" ( ',' | EOF ) !>
		attr = string_dq ":" baretext
		attr1 = string_dq ":" ( bool > baretext )
		attrs = attr { #',' attr }
	`)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	t.Logf("\nGrammar:%s", grammar.Serialize(false))
	t.Run("t1.0", func(t *testing.T) {
		testRule(t, grammar, "string_dq", `"double-quote block"`, `"double-quote block"`)
	})
	t.Run("t2.0", func(t *testing.T) {
		testRule(t, grammar, "string_sq", `'single-quote block'`, `'single-quote block'`)
	})
	t.Run("t3", func(t *testing.T) {
		testRule(t, grammar, "comment_ml", "  /* the block \ncomment   */  ", "/* the block \ncomment   */")
	})
	t.Run("t4", func(t *testing.T) {
		testRule(t, grammar, "comment_ml", "  /* the block comment */  ", "/* the block comment */")
	})
	t.Run("t4", func(t *testing.T) {
		testRule(t, grammar, "string_dq", `"double-quote block"`, `"double-quote block"`)
	})
	t.Run("t5", func(t *testing.T) {
		testRule(t, grammar, "comment_line", `// comments here  `, `// comments here  `)
	})
	t.Run("t6", func(t *testing.T) {
		// string1 defines the open and close double quote as string, so need extra space before the close quote
		testRule(t, grammar, "string1", `  "/* the block comment */"`, `"/* the block comment */"`)
	})
	t.Run("t7", func(t *testing.T) {
		testRule(t, grammar, "string", ` " this \" escape "`, `" this \" escape "`)
	})
	t.Run("t8", func(t *testing.T) {
		testRule(t, grammar, "string", ` " this \" escape with \n new line"`, `" this \" escape with \n new line"`)
	})
	t.Run("t9", func(t *testing.T) {
		testRule(t, grammar, "string", ` " this \\" escape "`, `" this \\"`)
	})
	t.Run("t10", func(t *testing.T) {
		testRule(t, grammar, "key", " this key = that value ", `this key =`)
	})
	t.Run("t11", func(t *testing.T) {
		testRule(t, grammar, "key", " this \\= key = that value ", `this \= key =`)
	})
	t.Run("t12", func(t *testing.T) {
		testRule(t, grammar, "value", " this is a value ", `this is a value `)
	})
	t.Run("t13", func(t *testing.T) {
		testRule(t, grammar, "value", " this is \\\n a value ", "this is \\\n a value ")
	})
	t.Run("t14", func(t *testing.T) {
		testRule(t, grammar, "property", " key = value ", "key = value ")
	})
	t.Run("t15", func(t *testing.T) {
		testRule(t, grammar, "keyempty", " empty key ", `empty key `)
	})
	t.Run("t16", func(t *testing.T) {
		testRule(t, grammar, "keyempty", " empty key = ", ``)
	})
	t.Run("t17", func(t *testing.T) {
		testRule(t, grammar, "string", `"sun1.opacity = (sun1.opacity / 100) * 90;"`, `"sun1.opacity = (sun1.opacity / 100) * 90;"`)
	})
	t.Run("t.18.attr.1", func(t *testing.T) {
		testRule(t, grammar, "attr", ` "empty key": = ,`, `"empty key" : = `)
	})
	t.Run("t.19.attr.ambiguity", func(t *testing.T) {
		testRule(t, grammar, "attr1", ` "empty key": true ,`, `"empty key" : true`)
		testRule(t, grammar, "attr1", ` "empty key": true,`, `"empty key" : true`)
	})
	t.Run("t.20.attrs", func(t *testing.T) {
		testRule(t, grammar, "attrs", ` "empty key": = ,"kevin xie": 89~#`, `"empty key" : = , "kevin xie" : 89~#`)
	})
}

func TestChoice(t *testing.T) {
	grammar, err := NewGrammarFromString(`
		letter_uppercase    = 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z'
		letter_lowercase    = 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z'
		letter              = letter_lowercase | letter_uppercase
		digit               = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'
		alphanumeric        = letter_lowercase | letter_uppercase | digit
		identifier          = letter { alphanumeric | '_' }
		integer             = [ '-' ] digit { digit }
		decimal             = integer '.' { digit }
		number              = "" (integer | decimal) // the leading "" making number a token (non-sticky)
		bool                = "true" | "false"
		string_sq           = <'\u0027' '\u0027'> // \u0027 is single quote '
		string_dq           = <'\u0022' '\u0022'> // \u0022 is double quote '"' 
		string              = string_sq | string_dq	
		value_ambiguity		= bool | identifier | number | string
		value				= bool > identifier | number | string
	`)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	t.Logf("\nGrammar:%s", grammar.Serialize(false))
	t.Run("value1", func(t *testing.T) {
		testRule(t, grammar, "value_ambiguity", ` myname `, `myname`)
	})
	t.Run("value2", func(t *testing.T) {
		testRule(t, grammar, "value_ambiguity", ` 123.45 `, `123.45`)
	})
	t.Run("value3.0", func(t *testing.T) {
		testRule(t, grammar, "value_ambiguity", ` true `, ``)
		testRule(t, grammar, "value", ` true `, `true`)
	})
	t.Run("value3.1", func(t *testing.T) {
		testRule(t, grammar, "value_ambiguity", ` false `, ``)
		testRule(t, grammar, "value", ` false `, `false`)
	})
}

func TestVirtual(t *testing.T) {
	tester := func(t *testing.T, g *Grammar, ruleName string, sample string, nodeCount int) {
		logSample := sample
		if len(logSample) > 40 {
			logSample = logSample[0:40] + "..."
		}
		t.Logf("====> SAMPLE: \"%s\"", logSample)
		record := g.GetRecord(ruleName)
		if record == nil {
			t.Errorf("XXXX> Failed: rule '%s' not defined", ruleName)
			return
		}
		rule := record.rule
		t.Logf("====> RULE  : %s = %s", ruleName, rule.String())
		cs := NewCharstreamFromString(sample)
		result := rule.Eval(g, cs, SUGGEST_SKIP)
		if result.Node == nil {
			t.Errorf("XXXX> Failed: no node found: %s", result.Error)
			return
		}
		t.Logf("====> Raw Result: \n%s", result.Node.StringTree(nil))
		result.Node.RemoveVirtualNodes()
		//result.Node.MergeLeafNodes()
		t.Logf("====> Result after virtual node removed: \n%s", result.Node.StringTree(nil))
		count := result.Node.CountNodes()
		if count != nodeCount {
			t.Errorf("XXXX> Failed: node count not match: expected %d vs actual %d", nodeCount, count)
			return
		}
		t.Logf("===> Passed")
	}
	grammar, err := NewGrammarFromString(`
		end = ~(\u000A|EOF)
		bool = space ("true"|"false") end
		space = ~{ \u0020 | \u0009 | \u000A | \u000D } 
	`)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	tester(t, grammar, "end", "", 3)
	tester(t, grammar, "bool", " \n false", 4)
}

func TestNondata(t *testing.T) {
	tester := func(t *testing.T, g *Grammar, ruleName string, sample string, nodeCount int) {
		logSample := sample
		if len(logSample) > 40 {
			logSample = logSample[0:40] + "..."
		}
		t.Logf("====> SAMPLE: \"%s\"", logSample)
		record := g.GetRecord(ruleName)
		if record == nil {
			t.Errorf("XXXX> Failed: rule '%s' not defined", ruleName)
			return
		}
		rule := record.rule
		t.Logf("====> RULE  : %s = %s", ruleName, rule.String())
		cs := NewCharstreamFromString(sample)
		result := rule.Eval(g, cs, SUGGEST_SKIP)
		if result.Node == nil {
			t.Errorf("XXXX> Failed: no node found: %s", result.Error)
			return
		}
		t.Logf("====> Raw Result: \n%s", result.Node.StringTree(nil))
		result.Node.MergeStickyNodes()
		t.Logf("====> MergeStickyNodes: \n%s", result.Node.StringTree(nil))
		result.Node.RemoveVirtualNodes()
		result.Node.RemoveRedundantNodes()
		t.Logf("====> RemoveVirtualNodes: \n%s", result.Node.StringTree(nil))
		count := result.Node.CountTokens()
		if count != nodeCount {
			t.Errorf("XXXX> Failed: node count not match: expected %d vs actual %d", nodeCount, count)
			return
		}
		t.Logf("===> Passed")
	}
	grammar, err := NewGrammarFromString(`
		end = ~(\u000A|EOF)
		bool = space ("true"|"false")
		space = ~{ \u0020 | \u0009 | \u000A | \u000D }
		item = bool { #"," bool }
		array = #"[" { item } #"]"
	`)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	tester(t, grammar, "array", "[ false, true, false ]", 7)
}

func TestMergeStickyNodes(t *testing.T) {
	tester := func(t *testing.T, g *Grammar, sample string, stickyNodeCount int) {
		logSample := sample
		if len(logSample) > 40 {
			logSample = logSample[0:40] + "..."
		}
		t.Logf("====> SAMPLE: \"%s\"", logSample)
		cs := NewCharstreamFromString(sample)
		ast, err := g.Eval(cs, LevelBasic)
		if err != nil {
			if stickyNodeCount != 0 {
				t.Errorf("XXXX> Failed: %s", err)
				return
			}
			t.Logf("====> Passed with expect node count 0 and err: %s", err)
		}
		//t.Logf("AST Raw\n%s", ast.StringTree(nil))
		ast.MergeStickyNodes()
		t.Logf("AST after MergeLeafNodes()\n%s", ast.StringTree(nil))
		actualCount := ast.CountStickyNodes()
		if actualCount != stickyNodeCount {
			t.Errorf("XXXX> Failed: node count not match: expected %d vs actual %d", stickyNodeCount, actualCount)
			return
		}
		t.Logf("====> Passed")
	}
	grammar, err := NewGrammarFromString(`
		// unicode character constants
		ESC     = \u001B  // escape character
		SPACE   = \u0020  // space
		TAB     = \u0009  // horizontal tab
		NL      = \u000A  // new line
		CR      = \u000D  // carriage return
		digit      = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'
		integer    = digit { digit }
		float      = integer '.' digit { digit }
		number     = "" [ '-' ] integer | float
		string     = < '"' '\\' ^NL '"'>
		bool       = "true" | "false"
		space      = ~{ SPACE | TAB | NL | CR }       // space is virtual
		literal    = number | string | bool | "null"
		array      = #"[" space [ value { #"," value } ] #"]" 
		kv         = space string space ":" value 
		object     = #"{" [ kv { #"," kv } ] #"}"
		value      = space (literal | array | object) space
		json       = value // root node
	`)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	t.Logf("Grammar: %s", grammar.Serialize(false))
	t.Run("test1", func(t *testing.T) {
		tester(t, grammar, "123456", 0)
	})
	t.Run("test2", func(t *testing.T) {
		tester(t, grammar, " true", 0)
	})
	t.Run("test3", func(t *testing.T) {
		tester(t, grammar, ` - {"a":1,"b":2,"c":3 }`, 0)
	})
	t.Run("test4", func(t *testing.T) {
		tester(t, grammar, `  {"a":1,"b":2,"c":3 }  `, 0)
	})
	t.Run("test5", func(t *testing.T) {
		tester(t, grammar, "-12\n", 0)
	})
	t.Run("test6", func(t *testing.T) {
		tester(t, grammar, "- 12\n", 0)
	})
	t.Run("test7", func(t *testing.T) {
		tester(t, grammar, " 123456.7890", 0)
	})
	t.Run("test8", func(t *testing.T) {
		tester(t, grammar, " -12.120000012", 0)
	})
	t.Run("test9", func(t *testing.T) {
		tester(t, grammar, `"string one line "`, 0)
	})
	t.Run("test10", func(t *testing.T) {
		tester(t, grammar, `"string \n 2nd line "`, 0)
	})
	t.Run("test11", func(t *testing.T) {
		tester(t, grammar, `{}`, 0)
	})
	t.Run("test12", func(t *testing.T) {
		tester(t, grammar, `[12.34, false, null, "string value"]`, 0)
	})
	t.Run("value.object.0", func(t *testing.T) {
		testRule(t, grammar, "value", `		{"key":-1234.56}`, `{ "key" : -1234.56 }`) // with leading tab
	})
	t.Run("value.object.ml.0", func(t *testing.T) {
		testRule(t, grammar, "value", `
				{"k":-1.5}`, `{ "k" : -1.5 }`)
	})
	t.Run("value.object.ml.1", func(t *testing.T) {
		testRule(t, grammar, "value", `
				{"key":-1234.56}`, `{ "key" : -1234.56 }`)
	})
	t.Run("test13.1", func(t *testing.T) {
		tester(t, grammar, `{"a":[],"ab":
				{"key":-1234.56}}`, 0)
	})
	t.Run("test14", func(t *testing.T) {
		tester(t, grammar, `{"a":1,"b":2,"c":3 }`, 0)
	})
	t.Run("test15", func(t *testing.T) {
		tester(t, grammar, `{ "glossary": {
			"title": "example glossary",
			"GlossDiv": {
				"title": "S"
			}
		}
	}`, 0)
	})
	t.Run("test16", func(t *testing.T) {
		tester(t, grammar, `[]`, 0)
	})
	t.Run("test17", func(t *testing.T) {
		tester(t, grammar, "{ \"a\": [\n] }", 0)
	})
	t.Run("test18", func(t *testing.T) {
		tester(t, grammar, "{ \"a\": [\n{},\n{}\n] }", 0)
	})
	t.Run("test19", func(t *testing.T) {
		tester(t, grammar, `{
			"glossary": {
				"title": "example glossary",
				"GlossDiv": {
					"title": "S",
					"GlossList": {
						"GlossEntry": {
							"ID": "SGML",
							"SortAs": "SGML",
							"GlossTerm": "Standard Generalized Markup Language",
							"Acronym": "SGML",
							"Abbrev": "ISO 8879:1986",
							"GlossDef": {
								"para": "A meta-markup language, used to create markup languages such as DocBook.",
								"GlossSeeAlso": ["GML", "XML"]
							},
							"GlossSee": "markup"
						}
					}
				}
			}
		}`, 0)
	})
}

func TestEmbed(t *testing.T) {
	tester := func(t *testing.T, g *Grammar, ruleName string, sample string, nodeCount int) {
		logSample := sample
		if len(logSample) > 40 {
			logSample = logSample[0:40] + "..."
		}
		t.Logf("====> SAMPLE: \"%s\"", logSample)
		evalResult := g.EvalEmbed(ruleName, sample)
		node := evalResult.Node
		node.MergeStickyNodes()
		node.RemoveVirtualNodes()
		node.RemoveRedundantNodes()
		t.Logf("Result Node:\n%s", node.StringTree(nil))
		if len(node.ChildNodes) != nodeCount {
			t.Errorf("Failed: Expected node count %d; Actual node count %d", nodeCount, len(node.ChildNodes))
		} else {
			t.Logf("Passed")
		}
	}
	grammar, err := NewGrammarFromString(`
		var = '$' ('A'-'Z'|'a'-'z') { ('A'-'Z'|'a'-'z'|'0'-'9'|'_') }
		variable = '${' ('A'-'Z'|'a'-'z') { ('A'-'Z'|'a'-'z'|'0'-'9'|'_') } '}'
		variable1 = "${" ('A'-'Z'|'a'-'z') { ('A'-'Z'|'a'-'z'|'0'-'9'|'_') } "}"
	`)
	if err != nil {
		t.Errorf("Failed: %s", err)
		return
	}
	t.Logf("Grammar: %s", grammar.Serialize(false))
	t.Run("test0", func(t *testing.T) {
		tester(t, grammar, "var", "", 0)
	})
	t.Run("test1", func(t *testing.T) {
		tester(t, grammar, "var", "123456", 1)
	})
	t.Run("test2", func(t *testing.T) {
		tester(t, grammar, "var", "123$A456 $b100 another text", 5)
	})
	t.Run("test3", func(t *testing.T) {
		tester(t, grammar, "variable", "123${A456} ${b100} another text", 5)
	})
	t.Run("test4.0", func(t *testing.T) {
		tester(t, grammar, "variable1", " ${A456}", 2)
	})
	t.Run("test4.1", func(t *testing.T) {
		tester(t, grammar, "variable1", "123${A456} ${b100} another text", 5)
	})
}
