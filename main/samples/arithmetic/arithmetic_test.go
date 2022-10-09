package arithmetic_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/cnsgfk/xbnf"
)

func evalRule(t *testing.T, grammar *xbnf.Grammar, ruleName string, sample string, expected string) {
	logSample := sample
	t.Logf("====> SAMPLE:%s*", logSample)
	record := grammar.GetRecord(ruleName)
	if record == nil {
		t.Errorf("XXXX> Failed: rule '%s' not defined", ruleName)
		return
	}
	rule := record.Rule()
	t.Logf("====> RULE  : %s = %s", ruleName, rule.String())
	cs := xbnf.NewCharstreamFromString(sample)
	result := rule.Eval(grammar, cs, xbnf.SUGGEST_SKIP)
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
	//result.Node.MergeStickyNodes()
	//result.Node.RemoveVirtualNodes()
	//result.Node.RemoveRedundantNodes()
	t.Logf("Actual Node: \n%s", result.Node.StringTree(nil))
	actual := result.Node.Text()
	if expected != string(actual) {
		t.Errorf("XXXX> Failed: expected vs actual\n%s\n%s", expected, string(actual))
		return
	}
	result.Node.RemoveNonDataNodes()
	t.Logf("====> Passed: %s", string(actual))
}

func evalRuleCS(t *testing.T, grammar *xbnf.Grammar, ruleName string, cs xbnf.ICharstream, expected string) {
	record := grammar.GetRecord(ruleName)
	if record == nil {
		t.Errorf("XXXX> Failed: rule '%s' not defined", ruleName)
		return
	}
	rule := record.Rule()
	t.Logf("====> RULE  : %s = %s", ruleName, rule.String())
	result := rule.Eval(grammar, cs, xbnf.SUGGEST_SKIP)
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
	//result.Node.MergeStickyNodes()
	//result.Node.RemoveVirtualNodes()
	//result.Node.RemoveRedundantNodes()
	t.Logf("Actual Node: \n%s", result.Node.StringTree(nil))
	actual := result.Node.Text()
	if expected != string(actual) {
		t.Errorf("XXXX> Failed: expected vs actual\n%s\n%s", expected, string(actual))
		return
	}
	result.Node.RemoveNonDataNodes()
	t.Logf("====> Passed: %s", string(actual))
}

func TestMain(t *testing.T) {
	tester := func(t *testing.T, grammar *xbnf.Grammar, sampleName string) {
		t.Logf("====> Sample: %s", sampleName)
		sampleFile := sampleName + ".txt"
		outputFile := sampleName + ".output"
		sample, err := ioutil.ReadFile(sampleFile)
		if err != nil {
			t.Errorf("Failed: %s", fmt.Errorf("can't real file: %s", err))
			return
		}
		t.Logf("====> Input file     : %s\n%s", sampleFile, string(sample))
		t.Logf("====> Expected Output: %s", outputFile)
		output, err := ioutil.ReadFile(outputFile)
		if err != nil {
			t.Errorf("Failed: %s", fmt.Errorf("can't real file: %s", err))
			return
		}
		outputText := string(output)
		outputText = strings.ReplaceAll(outputText, "\r\n", "\n") // convert from windows to unix newline format
		charstream := xbnf.NewCharstreamFromString(string(sample))
		ast, err := grammar.Eval(charstream, xbnf.LevelBasic)
		if err != nil {
			t.Errorf("Failed: %s", err)
			return
		}
		ast.RemoveVirtualNodes()
		t.Logf("Result Tree:\n%s", ast.StringTree(nil))
		text := string(ast.Text())
		if text != outputText {
			t.Logf("Expected Output: \n%s", outputText)
			t.Errorf("Failed: unexpect output\n%s", string(text))
			return
		}
		t.Logf("Passed: actual output\n%s", string(text))
	}
	file := "arithmetic.xbnf"
	g, err := xbnf.NewGrammarFromFile(file)
	if err != nil {
		t.Errorf("Failed: invalid xbnf file: %s", err)
		return
	}
	t.Logf("Grammar\n%s", g.Serialize(false))
	t.Run("letter0.0", func(t *testing.T) {
		cs := xbnf.NewCharstreamFromString("age")
		evalRuleCS(t, g, "letter_lowercase", cs, "a")
	})
	t.Run("letter0.1", func(t *testing.T) {
		cs := xbnf.NewCharstreamFromString("Age")
		evalRuleCS(t, g, "letter", cs, "A")
	})
	t.Run("variable0", func(t *testing.T) {
		evalRule(t, g, "variable", "Age", "Age")
	})
	t.Run("factor0", func(t *testing.T) {
		evalRule(t, g, "factor", "1\nA", "1")
	})
	t.Run("factor1", func(t *testing.T) {
		evalRule(t, g, "factor", " Age", "Age")
	})
	t.Run("literal", func(t *testing.T) {
		evalRule(t, g, "literal", "1\nA", "1")
	})
	t.Run("float0", func(t *testing.T) {
		evalRule(t, g, "float", "1\nA", "")
	})
	t.Run("float1", func(t *testing.T) {
		evalRule(t, g, "float", "1234.56", "1234.56")
	})
	t.Run("float2", func(t *testing.T) {
		evalRule(t, g, "float", ".56", ".56")
	})
	t.Run("float3", func(t *testing.T) {
		evalRule(t, g, "float", "-1234.56", "-1234.56")
	})
	t.Run("float4", func(t *testing.T) {
		evalRule(t, g, "float", "-.56", "-.56")
	})
	t.Run("term0", func(t *testing.T) {
		evalRule(t, g, "term", " Age", "Age")
	})
	t.Run("expr1.0", func(t *testing.T) {
		evalRule(t, g, "expr", "123 + 456", "123 + 456")
	})
	t.Run("expr1.1", func(t *testing.T) {
		evalRule(t, g, "expr", "123 + -456", "123 + -456")
	})
	t.Run("expr1.2", func(t *testing.T) {
		evalRule(t, g, "expr", "123 + -456", "123 + -456")
	})
	t.Run("expr1.3", func(t *testing.T) {
		evalRule(t, g, "expr", "A+123.45+456.78--34", "A + 123.45 + 456.78 - -34")
	})
	t.Run("expr1.4", func(t *testing.T) {
		evalRule(t, g, "expr", "123", "123")
	})
	t.Run("expr2.0", func(t *testing.T) {
		evalRule(t, g, "expr", "123 - Age*456", "123 - Age * 456")
	})
	t.Run("expr2.1", func(t *testing.T) {
		evalRule(t, g, "expr", "(123 - Age)      *456", "( 123 - Age ) * 456")
	})
	t.Run("expr2.2", func(t *testing.T) {
		evalRule(t, g, "expr", "123 - Age", "123 - Age")
	})
	t.Run("expr2.3", func(t *testing.T) {
		evalRule(t, g, "expr", "1\nA", "1")
	})
	t.Run("exprs1.0", func(t *testing.T) {
		evalRule(t, g, "exprs", "A+B\nC+D", "A + B\nC + D")
	})
	t.Run("exprs1.1", func(t *testing.T) {
		evalRule(t, g, "exprs", "A   \n\nB", "A\nB")
	})
	t.Run("exprs1.2", func(t *testing.T) {
		evalRule(t, g, "exprs", "A   \n\nB\n123", "A\nB\n123")
	})
	t.Run("exprs1.6", func(t *testing.T) {
		evalRule(t, g, "exprs", "(123-Age)*456\n1+1", "( 123 - Age ) * 456\n1 + 1")
	})
	t.Run("exprs2.0", func(t *testing.T) {
		evalRule(t, g, "exprs", "1\nA", "1\nA")
	})
	t.Run("exprs2.1", func(t *testing.T) {
		evalRule(t, g, "exprs", "1+1\nA*B", "1 + 1\nA * B")
	})
	t.Run("exprs2.2", func(t *testing.T) {
		evalRule(t, g, "exprs", "(123-Age)*456\n1+1\nA*B", "( 123 - Age ) * 456\n1 + 1\nA * B")
	})
	t.Run("exprs3", func(t *testing.T) {
		evalRule(t, g, "exprs", "123.45 + 456.78 - -34\n1+1\nA+3*B", "123.45 + 456.78 - -34\n1 + 1\nA + 3 * B")
	})
	t.Run("sample1", func(t *testing.T) {
		tester(t, g, "sample1")
	})
	t.Run("sample2", func(t *testing.T) {
		tester(t, g, "sample2")
	})
	t.Run("sample3", func(t *testing.T) {
		tester(t, g, "sample3")
	})
}
