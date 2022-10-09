package xbnf

import (
	"testing"
)

func TestCharstreamNextString(t *testing.T) {
	cs, charSequence, _ := getCharStream()
	var idx int
	t.Logf("ICharStream type: %T", cs)
	for idx = 0; ; idx++ {
		char := cs.Next()
		if char == EOFChar {
			break
		}
		expectChar := charSequence[idx]
		if expectChar != char {
			t.Errorf("Failed: At %d index: expect '%c' vs actual '%c'", idx, expectChar, char)
			return
		}
	}
	t.Logf("Passed: read %d char(s), all matched expectation", idx)
}

func TestCharstreamNextPrepend(t *testing.T) {
	css, charSequence, _ := getCharStream()
	var prepend []rune
	for i := 0; i < len(charSequence)/2; i++ {
		char := css.Next()
		prepend = append(prepend, char)
	}
	cs := newCharstreamPrepend(css, prepend)
	t.Logf("ICharStream type: %T", cs)
	var idx int
	for idx = 0; ; idx++ {
		char := cs.Next()
		if char == EOFChar {
			break
		}
		expectChar := charSequence[idx]
		if expectChar != char {
			t.Errorf("Failed: At %d index: expect '%c' vs actual '%c'", idx, expectChar, char)
			return
		}
	}
	t.Logf("Passed: read %d char(s), all matched expectation", idx)
}

func TestCharstreamSkipSpaces(t *testing.T) {
	tester := func(t *testing.T, example string, spaceCount int, firstNonSpace rune) {
		t.Logf("====> Sample:%s*", example)
		cs := NewCharstreamFromString(example)
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
	tester := func(cs ICharstream, matchWord string) {
		t.Logf("trying match   : '%s'", matchWord)
		matches, succeeded := cs.Match([]rune(matchWord))
		matchStr := string(matches)
		if succeeded {
			t.Logf("match succeeded: '%s'", matchStr)
		} else {
			t.Logf("match fail     : '%s'", matchStr)
		}
	}
	tester(NewCharstreamFromString("This is my name 'Kevin Xie'"), "This is m")
	tester(NewCharstreamFromString("This is my name 'Kevin Xie'"), "This Is m")
	tester(newCharstreamPrepend(NewCharstreamFromString("This is my name 'Kevin Xie'"), []rune("Prepend Text: ")), "Prepend Tex")
	tester(newCharstreamPrepend(NewCharstreamFromString("This is my name 'Kevin Xie'"), []rune("Prepend Text: ")), "Prepend Text: This ")
	tester(newCharstreamPrepend(NewCharstreamFromString("This is my name 'Kevin Xie'"), []rune("Prepend Text: ")), "Prepend Text: This M")
}

func TestCharstreamPositionString(t *testing.T) {
	cs, charSequence, charMap := getCharStream()
	t.Logf("ICharStream type: %T", cs)
	var pos *Position
	var charRead rune
	var charAtPos rune
	// before reading anything
	pos = cs.Position()
	if pos == nil && len(charSequence) != 0 {
		t.Errorf("Failed: expect L1:1 but get %s", pos.String())
		return
	}
	for {
		charRead = cs.Next()
		if charRead == EOFChar {
			break
		}
		charReadStr := string(charRead)
		if charRead == '\n' {
			charReadStr = "\\n"
		}
		charAtPos = charMap[pos.Line-1][pos.Col-1]
		if charAtPos != charRead {
			charAtPosStr := string(charAtPos)
			if charAtPos == '\n' {
				charAtPosStr = "\\n"
			}
			t.Errorf("Failed: char read is '%s' vs char at %s is '%s'", charReadStr, pos.String(), charAtPosStr)
			return
		}
		pos = cs.Position()
	}
	t.Logf("Passed: every char read matched char get by position")
}

func TestCharstreamPositionPrepend(t *testing.T) {
	css, charSequence, charMap := getCharStream()
	var prepend []rune
	for i := 0; i < len(charSequence)/2; i++ {
		char := css.Next()
		prepend = append(prepend, char)
	}
	cs := newCharstreamPrepend(css, prepend)
	t.Logf("ICharStream type: %T", cs)
	var pos *Position
	var charRead rune
	var charAtPos rune
	// before reading anything
	pos = cs.Position()
	if pos == nil && len(charSequence) != 0 {
		t.Errorf("Failed: expect L1:1 but get %s", pos.String())
		return
	}
	for {
		charRead = cs.Next()
		if charRead == EOFChar {
			break
		}
		charReadStr := string(charRead)
		if charRead == '\n' {
			charReadStr = "\\n"
		}
		charAtPos = charMap[pos.Line-1][pos.Col-1]
		if charAtPos != charRead {
			charAtPosStr := string(charAtPos)
			if charAtPos == '\n' {
				charAtPosStr = "\\n"
			}
			t.Errorf("Failed: char read is '%s' vs char at %s is '%s'", charReadStr, pos.String(), charAtPosStr)
			return
		}
		pos = cs.Position()
	}
	t.Logf("Passed: every char read matched char get by position")
}

func TestCharstreamLookupString(t *testing.T) {
	cs, charSequence, charMap := getCharStream()
	for cs.Next() != EOFChar {
	}
	t.Logf("ICharStream type: %T", cs)
	tester := func(t *testing.T, idx int) {
		pos := cs.PositionLookup(idx)
		if idx < 0 || idx >= len(charSequence) {
			if pos == nil {
				t.Logf("Passed: %d index doesn't have a position", idx)
			} else {
				t.Errorf("Failed: %d index expects nil position, but got %s", idx, pos.String())
			}
			return
		}
		if pos == nil {
			t.Errorf("Failed: %d index expects a position, but got nil", idx)
			return
		}
		charByMap := string(charMap[pos.Line-1][pos.Col-1])
		if charByMap == "\n" {
			charByMap = "\\n"
		}
		charBySeq := string(charSequence[idx])
		if charBySeq == "\n" {
			charBySeq = "\\n"
		}
		if charBySeq != charByMap {
			t.Errorf("Failed: char at %d index is '%s' vs '%s' at position %s", idx, charBySeq, charByMap, pos.String())
			return
		}
		t.Logf("Passed: char at %d index is '%s' matches '%s' at position %s", idx, charBySeq, charByMap, pos.String())
	}
	tester(t, -10)
	tester(t, 0)
	tester(t, 1)
	tester(t, 2)
	tester(t, 10)
	tester(t, 50)
	tester(t, len(charSequence)-1)
	tester(t, len(charSequence))
	tester(t, len(charSequence)+1)
	tester(t, 1000)
}

func TestCharstreamLookupPrepend(t *testing.T) {
	css, charSequence, charMap := getCharStream()
	var prepend []rune
	for i := 0; i < len(charSequence)/2; i++ {
		char := css.Next()
		prepend = append(prepend, char)
	}
	cs := newCharstreamPrepend(css, prepend)
	for cs.Next() != EOFChar {
	}
	t.Logf("ICharStream type: %T", cs)
	tester := func(t *testing.T, idx int) {
		pos := cs.PositionLookup(idx)
		if idx < 0 || idx >= len(charSequence) {
			if pos == nil {
				t.Logf("Passed: %d index doesn't have a position", idx)
			} else {
				t.Errorf("Failed: %d index expects nil position, but got %s", idx, pos.String())
			}
			return
		}
		if pos == nil {
			t.Errorf("Failed: %d index expects a position, but got nil", idx)
			return
		}
		charByMap := string(charMap[pos.Line-1][pos.Col-1])
		if charByMap == "\n" {
			charByMap = "\\n"
		}
		charBySeq := string(charSequence[idx])
		if charBySeq == "\n" {
			charBySeq = "\\n"
		}
		if charBySeq != charByMap {
			t.Errorf("Failed: char at %d index is '%s' vs '%s' at position %s", idx, charBySeq, charByMap, pos.String())
			return
		}
		t.Logf("Passed: char at %d index is '%s' matches '%s' at position %s", idx, charBySeq, charByMap, pos.String())
	}
	tester(t, -10)
	tester(t, 0)
	tester(t, 1)
	tester(t, 2)
	tester(t, 10)
	tester(t, 50)
	tester(t, len(charSequence)-1)
	tester(t, len(charSequence))
	tester(t, len(charSequence)+1)
	tester(t, 1000)
}

func getCharStream() (cs ICharstream, charSequence []rune, charMap [][]rune) {
	var data [][]rune = [][]rune{
		[]rune("\n"),                       // L1
		[]rune("ThisIsJustALineExample\n"), //L2
		[]rune("\n"),                       //L3
		[]rune("\n"),                       //L4
		[]rune("\n"),                       //L5
		[]rune("LinesAfter3BlankLines\n"),  //L6
		[]rune("\n"),                       //L7
		[]rune("FirstLineInAThreeLineSequence\n"), //L8
		[]rune("2ndLine\n"),                       //L9
		[]rune("3rdLine\n"),                       //L10
		[]rune("\n"),                              //L11
		[]rune("FinalLineFollowingANewLine\n"),    //L12 - last \n is L12:27
		[]rune("\n"),                              // L13
	}
	var tmpLine []rune
	for _, line := range data {
		tmpLine = nil
		charSequence = append(charSequence, line...)
		tmpLine = append(tmpLine, line...)
		charMap = append(charMap, tmpLine)
	}
	cs = NewCharstreamFromString(string(charSequence))
	return
}
