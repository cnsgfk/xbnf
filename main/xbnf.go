package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cnsgfk/xbnf"
)

func main() {
	// process command line arguments
	ruleFile := flag.String("xbnf", "", "Optional - The XBNF file with a set of rules to be added to the grammar")
	treeNodeType := flag.Bool("showNodeType", false, "Show node type in the AST tree")
	help := flag.Bool("help", false, "Print help message")
	var rules multi
	flag.Var(&rules, "rule", "Optional - Add a rule in the grammar. ")
	var texts multi
	flag.Var(&texts, "text", "Optional - text to be parsed by the grammar")
	var textFiles multi
	flag.Var(&textFiles, "file", "Optional - file to be parsed by the grammar")
	flag.Parse()

	if *help || len(os.Args) <= 1 {
		printHelp()
		os.Exit(0)
	}

	var grammar *xbnf.Grammar

	if (ruleFile == nil || *ruleFile == "") && len(rules) <= 0 {
		fmt.Printf("ERROR: Command line arguments must contains at least one of -xbnf or -rule\n")
		printHelp()
		return
	} else {
		fmt.Printf("Adding rule(s) from XBNF file: %s\n", *ruleFile)
		g, err := xbnf.NewGrammarFromFile(*ruleFile)
		if err != nil {
			fmt.Printf("ERROR: %s", err)
			return
		}
		grammar = g
	}

	if grammar == nil {
		grammar = xbnf.NewGrammar()
	}

	for _, rule := range rules {
		fmt.Printf("Adding rule: %s\n", rule)
		_, err := grammar.AddRule(rule)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return
		}
	}

	err := grammar.Validate()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	grammarStr := grammar.Serialize(false)
	grammarStr = strings.ReplaceAll(grammarStr, "\n", "\n    ")
	fmt.Printf("Grammar:\n    %s", grammarStr)

	treeConf := xbnf.DefaultNodeTreeConfig()
	treeConf.PrintRuleType = *treeNodeType

	for _, text := range texts {
		fmt.Printf("\nParsing text: %s\n", text)
		cs := xbnf.NewCharstreamFromString(text)
		ast, err := grammar.Eval(cs, xbnf.LevelBasic)
		if err != nil {
			fmt.Printf("ERROR: %s", err)
			return
		}
		ast.RemoveVirtualNodes()
		ast.RemoveNonDataNodes()
		ast.RemoveRedundantNodes()
		fmt.Printf("%s\n", ast.StringTree(treeConf))
	}

	for _, textfile := range textFiles {
		fmt.Printf("\nParsing file: %s\n", textfile)
		text, err := ioutil.ReadFile(textfile)
		if err != nil {
			fmt.Printf("ERROR: %s", err)
			return
		}
		cs := xbnf.NewCharstreamFromString(string(text))
		ast, err := grammar.Eval(cs, xbnf.LevelBasic)
		if err != nil {
			fmt.Printf("ERROR: %s", err)
			return
		}
		ast.RemoveVirtualNodes()
		ast.MergeStickyNodes()
		ast.RemoveNonDataNodes()
		ast.RemoveRedundantNodes()
		fmt.Printf("%s\n", ast.StringTree(treeConf))
	}
}

func printHelp() {
	flag.CommandLine.Usage()
}

type multi []string

func (inst *multi) Set(value string) error {
	*inst = append(*inst, value)
	return nil
}

func (inst *multi) String() string {
	return fmt.Sprintf("%v", *inst)
}
