package xbnf

import "sync"

// special predefined chars
const (
	EOFChar = -1
)

type Position struct {
	Line uint // starts at 0
	Col  uint // starts at 0
}

func (inst *Position) Forward(chars []rune) {
	for _, char := range chars {
		if char == '\n' {
			inst.Line = inst.Line + 1
			inst.Col = 0
			continue
		}
		inst.Col = inst.Col + 1
	}
}

type ICharstream interface {
	// return a char (unicode point)
	Next() rune

	// Preview next char, the cursor underneath the stream should not move
	Peek() rune

	// Matches the input rune slice, if succeeded, return the match rune list. If not,
	// return false and the slice of rune read in order to try the match
	Match([]rune) ([]rune, bool)

	SkipSpaces() []rune

	Position() Position

	Index() int // return current char index
}

func NewCharstreamString(input string) ICharstream {
	cs := &CharstreamString{}
	cs.cursor = 0
	cs.chars = []rune(input)
	return cs
}

// CharstreamString is thread-safe
type CharstreamString struct {
	position Position
	cursor   int    // cursor starts at 0
	chars    []rune // chars should not be nil
	lock     sync.RWMutex
}

func (inst *CharstreamString) Position() Position {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	return inst.position
}

func (inst *CharstreamString) Index() int {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	return inst.cursor
}

func (inst *CharstreamString) Next() rune {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	return inst.next()
}

func (inst *CharstreamString) next() rune {
	if len(inst.chars) == 0 {
		return EOFChar
	}
	if inst.cursor >= len(inst.chars) {
		return EOFChar
	}
	idx := inst.cursor
	inst.cursor = inst.cursor + 1
	char := inst.chars[idx]
	if char == '\n' {
		inst.position.Line = inst.position.Line + 1
	} else {
		inst.position.Col = inst.position.Col + 1
	}
	return char
}

func (inst *CharstreamString) Peek() rune {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	if len(inst.chars) == 0 {
		return EOFChar
	}
	if inst.cursor >= len(inst.chars) {
		return EOFChar
	}
	return inst.chars[inst.cursor]
}

// Read all chars from the stream until a none-whitespace character. Return nil slice if the
// current char is a none-whitespace
func (inst *CharstreamString) SkipSpaces() []rune {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	beginCur := inst.cursor
	i := inst.cursor
	countNewline := uint(0)
	lastNewline := 0
	for i = inst.cursor; i < len(inst.chars); i++ {
		if !IsWhiteSpace(inst.chars[i]) {
			break
		}
		if inst.chars[i] == '\n' {
			countNewline++
			lastNewline = i
		}
	}
	if i > beginCur {
		inst.cursor = i
		spaces := inst.chars[beginCur:inst.cursor]
		if countNewline > 0 {
			inst.position.Line = inst.position.Line + countNewline
			inst.position.Col = uint(inst.cursor - lastNewline)
		} else {
			inst.position.Col = inst.position.Col + uint(len(spaces))
		}
		return spaces
	}
	return nil
}

func (inst *CharstreamString) Match(input []rune) ([]rune, bool) {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	if input == nil || len(input) == 0 {
		return nil, true
	}
	var result []rune
	for _, inputChar := range input {
		char := inst.next()
		if char == EOFChar {
			return result, false // not match
		}
		result = append(result, char)
		if char != inputChar {
			return result, false
		}
	}
	return result, true
}

func NewCharstreamPrepend(baseCharstream ICharstream, prepend []rune) ICharstream {
	if len(prepend) == 0 {
		return baseCharstream
	}

	csp := &CharstreamPrepend{}
	csp.charstream = baseCharstream
	if csp.charstream == nil {
		csp.charstream = NewCharstreamString("")
	}
	csp.prepend = prepend
	return csp
}

// CharstreamPrepend is thread-safe
type CharstreamPrepend struct {
	position   Position
	prepend    []rune
	charstream ICharstream
	cursor     int
	lock       sync.Mutex
}

func (inst *CharstreamPrepend) Next() rune {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	return inst.next()
}

func (inst *CharstreamPrepend) next() rune {
	if inst.prepend != nil && inst.cursor < len(inst.prepend) {
		inst.cursor = inst.cursor + 1
		return inst.prepend[inst.cursor-1]
	}
	char := inst.charstream.Next()
	if char != EOFChar {
		inst.cursor = inst.cursor + 1
	}
	return char
}

func (inst *CharstreamPrepend) Peek() rune {
	inst.lock.Lock()
	defer inst.lock.Unlock()
	if inst.prepend != nil && inst.cursor < len(inst.prepend) {
		return inst.prepend[inst.cursor]
	}
	return inst.charstream.Peek()
}

func (inst *CharstreamPrepend) peek() rune {
	if inst.prepend != nil && inst.cursor < len(inst.prepend) {
		return inst.prepend[inst.cursor]
	}
	return inst.charstream.Peek()
}

func (inst *CharstreamPrepend) SkipSpaces() []rune {
	inst.lock.Lock()
	defer inst.lock.Unlock()
	var charsRead []rune
	for {
		char := inst.peek()
		if char == EOFChar {
			break
		}
		if !IsWhiteSpace(char) {
			break
		}
		charsRead = append(charsRead, inst.next())
	}
	return charsRead
}

func (inst *CharstreamPrepend) Match(input []rune) ([]rune, bool) {
	inst.lock.Lock()
	defer inst.lock.Unlock()
	if input == nil || len(input) == 0 {
		return nil, true
	}
	var result []rune
	for _, inputChar := range input {
		char := inst.next()
		if char == EOFChar {
			return result, false // not match
		}
		result = append(result, char)
		if char != inputChar {
			return result, false
		}
	}
	return result, true
}

func (inst *CharstreamPrepend) Position() Position {
	//inst.lock.Lock()
	//defer inst.lock.Unlock()

	return inst.charstream.Position()
}

func (inst *CharstreamPrepend) Index() int {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	return inst.Index() - len(inst.prepend) + inst.cursor
}
