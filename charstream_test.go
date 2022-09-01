package xbnf_test

import (
	"testing"

	"github.com/cnsgfk/xbnf"
)

func TestCharstreamNext(t *testing.T) {
	tester := func(cs xbnf.ICharstream) {
		t.Logf("----")
		for idx := 0; ; idx++ {
			char := cs.Next()
			if char == xbnf.EOFChar {
				break
			}
			t.Logf("%d: %c", idx, char)
		}
	}
	tester(xbnf.NewCharstreamString("This is my name 'Kevin Xie'"))
	tester(xbnf.NewCharstreamPrepend(xbnf.NewCharstreamString("This is my name 'Kevin Xie'"), []rune("Prepend Text: ")))
}

func TestCharstreamSkipSpaces(t *testing.T) {
	tester := func(t *testing.T, example string, spaceCount int, firstNonSpace rune) {
		t.Logf("====> Sample:%s*", example)
		cs := xbnf.NewCharstreamString(example)
		t.Logf("====> expect %d leading space(s), and first non-space char '%c'", spaceCount, firstNonSpace)
		spaces := cs.SkipSpaces()
		if spaceCount != len(spaces) || cs.Peek() != firstNonSpace {
			t.Errorf("****> Failed: %d leading space(s); first non-space char '%c'", len(spaces), firstNonSpace)
			return
		}
		t.Logf("====> Passed: %d leading space(s); first non-space char '%c'", len(spaces), firstNonSpace)
	}
	t.Run("test1", func(t *testing.T) {
		tester(t, " is my name 'Kevin Xie'", 1, 'i')
	})
	t.Run("test2", func(t *testing.T) {
		tester(t, "\t \n name", 4, 'n')
	})
	t.Run("test3", func(t *testing.T) {
		tester(t, "    This is my name 'Kevin Xie'", 4, 'T')
	})
}
func TestCharstreamMatch(t *testing.T) {
	tester := func(cs xbnf.ICharstream, matchWord string) {
		t.Logf("trying match   : '%s'", matchWord)
		matches, succeeded := cs.Match([]rune(matchWord))
		matchStr := string(matches)
		if succeeded {
			t.Logf("match succeeded: '%s'", matchStr)
		} else {
			t.Logf("match fail     : '%s'", matchStr)
		}
	}
	tester(xbnf.NewCharstreamString("This is my name 'Kevin Xie'"), "This is m")
	tester(xbnf.NewCharstreamString("This is my name 'Kevin Xie'"), "This Is m")
	tester(xbnf.NewCharstreamPrepend(xbnf.NewCharstreamString("This is my name 'Kevin Xie'"), []rune("Prepend Text: ")), "Prepend Tex")
	tester(xbnf.NewCharstreamPrepend(xbnf.NewCharstreamString("This is my name 'Kevin Xie'"), []rune("Prepend Text: ")), "Prepend Text: This ")
	tester(xbnf.NewCharstreamPrepend(xbnf.NewCharstreamString("This is my name 'Kevin Xie'"), []rune("Prepend Text: ")), "Prepend Text: This M")
}
