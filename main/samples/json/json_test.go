package json_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/cnsgfk/xbnf"
)

func TestJSON(t *testing.T) {
	tester := func(t *testing.T, grammar *xbnf.Grammar, sampleName string) {
		t.Logf("====> Sample: %s", sampleName)
		sampleFile := sampleName + ".json"
		outputFile := sampleName + ".output"
		t.Logf("====> JSON file      : %s", sampleFile)
		t.Logf("====> Expected Output: %s", outputFile)
		sample, err := ioutil.ReadFile(sampleFile)
		if err != nil {
			t.Errorf("Failed: %s", fmt.Errorf("can't real file: %s", err))
			return
		}
		output, err := ioutil.ReadFile(outputFile)
		if err != nil {
			t.Errorf("Failed: %s", fmt.Errorf("can't real file: %s", err))
			return
		}
		outputText := string(output)
		charstream := xbnf.NewCharstreamString(string(sample))
		ast, err := grammar.Eval(charstream, xbnf.LevelRaw)
		if err != nil {
			t.Errorf("Failed: %s", err)
			return
		}
		//outputText = strings.ReplaceAll(outputText, "\n", "")
		ast.MergeStickyNodes()
		ast.RemoveVirtualNodes()
		ast.RemoveRedundantNodes()
		t.Logf("Result Tree:\n%s", ast.StringTree(nil))
		text := ast.Text()
		if string(text) != outputText {
			t.Logf("Expected Output: \n%s", outputText)
			t.Errorf("Failed: unexpect output\n%s", string(text))
			return
		}
		t.Logf("Passed: actual output\n%s", string(text))
	}
	file := "json.xbnf"
	g, err := xbnf.NewGrammarFromFile(file)
	if err != nil {
		t.Errorf("Failed: invalid xbnf file: %s", err)
		return
	}
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

func TestNondata(t *testing.T) {
	tester := func(t *testing.T, grammar *xbnf.Grammar, sampleName string) {
		t.Logf("====> Sample: %s", sampleName)
		sampleFile := sampleName + ".json"
		outputFile := sampleName + ".output"
		t.Logf("====> JSON file      : %s", sampleFile)
		t.Logf("====> Expected Output: %s", outputFile)
		sample, err := ioutil.ReadFile(sampleFile)
		if err != nil {
			t.Errorf("Failed: %s", fmt.Errorf("can't real file: %s", err))
			return
		}
		output, err := ioutil.ReadFile(outputFile)
		if err != nil {
			t.Errorf("Failed: %s", fmt.Errorf("can't real file: %s", err))
			return
		}
		outputText := string(output)
		charstream := xbnf.NewCharstreamString(string(sample))
		ast, err := grammar.Eval(charstream, xbnf.LevelRaw)
		if err != nil {
			t.Errorf("Failed: %s", err)
			return
		}
		ast.RemoveVirtualNodes()
		ast.MergeStickyNodes()
		//ast.RemoveNonDataNodes()
		ast.RemoveRedundantNodes()
		t.Logf("Result Tree:\n%s", ast.StringTree(nil))
		text := ast.Text()
		if string(text) != outputText {
			t.Logf("Expected Output: \n%s", outputText)
			t.Errorf("Failed: unexpect output\n%s", string(text))
			return
		}
		t.Logf("Passed: actual output\n%s", string(text))
	}
	file := "json.xbnf"
	g, err := xbnf.NewGrammarFromFile(file)
	if err != nil {
		t.Errorf("Failed: invalid xbnf file: %s", err)
		return
	}
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
