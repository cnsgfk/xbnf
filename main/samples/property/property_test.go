package property_test

import (
	"testing"

	"github.com/cnsgfk/xbnf"
)

func evalGrammar(t *testing.T, grammar *xbnf.Grammar, caseIsGood bool, caseSample string) {
	logSample := caseSample
	if len(logSample) > 40 {
		logSample = logSample[0:40] + "..."
	}
	t.Logf("============================================================")
	t.Logf("====> sample: [%s]", logSample)
	cs := xbnf.NewCharstreamFromString(caseSample)
	treeConf := xbnf.DefaultNodeTreeConfig()
	treeConf.PrintNonleafNodeText = false
	treeConf.VerboseNodeText = true
	//treeConf.PrintRuleType = true
	ast, err := grammar.Eval(cs, xbnf.LevelBasic)
	//t.Logf(string(ast.Text()))
	t.Logf("%s", ast.StringTree(nil))
	if caseIsGood {
		if err != nil {
			t.Errorf("****> Failed: %s; Result: \n%s", err, ast.StringTree(treeConf))
			return
		}
		ast.RemoveVirtualNodes()
		ast.MergeStickyNodes()
		t.Logf(string(ast.Text()))
		astString := ast.StringTree(treeConf)
		t.Logf("====> Passed: Result: \n%s", astString)
		return
	}
	if err == nil {
		t.Errorf("****> Failed: missing expected error; Result: \n%s", ast.StringTree(treeConf))
		return
	}
	t.Logf("====> Passed: Result: %s\n%s", err, ast.StringTree(treeConf))
	return
}

func TestProperty(t *testing.T) {
	file := "property.xbnf"
	grammar, err := xbnf.NewGrammarFromFile(file)
	if err != nil {
		t.Errorf("Failed: invalid xbnf file: %s", err)
		return
	}
	t.Logf("Grammar: %s", grammar.Serialize(false))
	evalGrammar(t, grammar, true, "123456")
	evalGrammar(t, grammar, true, "colon_key:http://www.google.com")
	evalGrammar(t, grammar, true, `empty key
		    abc    = 123456
	    	url		=    http://www.google.com`)
	evalGrammar(t, grammar, true, `
		empty key
		abc	= 123456
	    url	=    http://www.google.com
	    colon_Key	:    http://www.google.com
		multi_line = line1 \
continue second line \
3rd line
	`)
	evalGrammar(t, grammar, true, `
db.url=localhost
db.user=mkyong
db.password=password	
	`)
}
