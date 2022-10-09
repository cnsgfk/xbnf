package xbnf

import (
	"fmt"
	"sync"
)

// special predefined chars
const (
	EOFChar = -1
)

// Position represents a character location in a stream in line/col format.
// The Line and Col both start at 1
type Position struct {
	Line int
	Col  int
}

func (inst *Position) String() string {
	if inst == nil {
		return "L0:0"
	}
	return fmt.Sprintf("L%d:%d", inst.Line, inst.Col)
}

func (inst *Position) is(line int, col int) bool {
	if inst == nil {
		return line == 0 && col == 0
	}
	return inst.Line == line && inst.Col == col
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

	// Position returns the current cursor position in the stream. The char at current
	// position is the char to be read next. Return nil if the cursor is at EOF
	Position() *Position

	// PositionLookup translates an index in the stream to a position object. Return nil if
	// the index is less than 0 or is equal or beyond the current cursor in the stream.
	PositionLookup(idx int) *Position

	Cursor() int // return current cursor which is the 0-base index position in the stream.
}

type positionMap struct {
	// the index of all newline '\n' chars in order, if the last idx is the same as
	// the 2nd last index, it means the last read char is a '\n'
	lines []int
}

func (inst *positionMap) update(idx int, chars ...rune) {
	if len(inst.lines) == 0 {
		inst.lines = []int{-1} // beginning value
	}
	for _, char := range chars {
		if idx > inst.lines[len(inst.lines)-1] {
			// this char at idx has NOT been read and updated
			if char == '\n' {
				inst.lines[len(inst.lines)-1] = idx  // set last line's last char to '\n' index
				inst.lines = append(inst.lines, idx) // create a new line with last char also to '\n' index
			} else {
				inst.lines[len(inst.lines)-1] = idx // update the last line's last char idx
			}
		}
		idx++
	}
}

func (inst *positionMap) lookup(idx int) *Position {
	if idx < 0 {
		return nil
	}

	countLines := len(inst.lines)
	if countLines == 0 {
		return &Position{Line: 1, Col: 1} // index 0 is always L1:1
	}

	maxLineIdx := countLines - 1

	// char at idx hasn't been read yet
	if idx > inst.lines[maxLineIdx] {
		return nil
	} else if maxLineIdx == 0 {
		// only 1 line read
		return &Position{Line: 1, Col: idx + 1}
	}

	// the case of "last read char is '\n'"
	if maxLineIdx > 0 && inst.lines[maxLineIdx] == inst.lines[maxLineIdx-1] {
		// the last line is not a real line, but the case indicator
		maxLineIdx = maxLineIdx - 1
	}

	lineIdx := 0
	var col int
	// do reverse linear lookup (TODO: binary lookup should be faster)
	for i := maxLineIdx; i >= 0; i-- {
		if idx >= inst.lines[i] {
			lineIdx = i
			break
		}
	}

	// now the idx is garrenttee >= the lineIdx
	// we found a line
	//if lineIdx == 0 {
	//	// if it's the 1st line, the col is the idx + 1
	//	col = idx + 1
	//} else {
	//	// if it has a previous line, the col is
	if idx == inst.lines[lineIdx] {
		if lineIdx > 0 {
			col = idx - inst.lines[lineIdx-1]
		} else {
			col = idx + 1
		}
	} else {
		lineIdx = lineIdx + 1
		col = idx - inst.lines[lineIdx-1]
	}
	//}

	return &Position{Line: lineIdx + 1, Col: col}
}

/*
*

	func NewCharstreamFromFile(file string) (ICharstream, error) {
		reader, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		runeReader := bufio.NewReader(reader)
		cs := &CharstreamFile{}
		cs.source = File(file)
		cs.runeReader = runeReader
		return cs
	}

	type CharstreamFile struct {
		lock       sync.RWMutex
		lines      []uint64    // array of lines, the value is last char 0-base index in the line, the newline \n char is not included in any line
		source     interface{} // could be File/String
		runeReader *bufio.Reader
		cursor     uint64
	}

	func (inst *CharstreamFile) Position() Position {
		inst.lock.Lock()
		defer inst.lock.Unlock()

		return inst.position
	}

	func (inst *CharstreamFile) Index() int {
		inst.lock.Lock()
		defer inst.lock.Unlock()

		return inst.cursor
	}

	func (inst *CharstreamFile) Next() rune {
		inst.lock.Lock()
		defer inst.lock.Unlock()

		return inst.next()
	}

	func (inst *CharstreamFile) next() rune {
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

	func (inst *CharstreamFile) Peek() rune {
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

	func (inst *CharstreamFile) SkipSpaces() []rune {
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

	func (inst *CharstreamFile) Match(input []rune) ([]rune, bool) {
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
*/
type File string
type String string

func NewCharstreamFromString(input string) ICharstream {
	cs := &CharstreamString{}
	cs.cursor = 0
	cs.chars = []rune(input)
	cs.positionMap.lines = []int{-1}
	return cs
}

// CharstreamString is thread-safe
type CharstreamString struct {
	positionMap positionMap
	cursor      int    // cursor starts at 0
	chars       []rune // chars should not be nil
	lock        sync.RWMutex
}

func (inst *CharstreamString) Position() *Position {
	inst.lock.RLock()
	defer inst.lock.RUnlock()

	if inst.cursor >= len(inst.chars) {
		// cursor is at EOF
		return nil
	}

	if inst.cursor == 0 {
		// first position is always L1:1
		return &Position{Line: 1, Col: 1}
	}

	inst.positionMap.update(inst.cursor, inst.chars[inst.cursor])

	return inst.positionMap.lookup(inst.cursor)
}

// Return the position of char index
func (inst *CharstreamString) PositionLookup(idx int) *Position {
	inst.lock.RLock()
	defer inst.lock.RUnlock()

	if idx == 0 {
		return &Position{Line: 1, Col: 1}
	}

	if idx >= inst.cursor || idx >= len(inst.chars) {
		return nil
	}

	return inst.positionMap.lookup(idx)
}

func (inst *CharstreamString) Cursor() int {
	inst.lock.RLock()
	defer inst.lock.RUnlock()

	return inst.cursor
}

func (inst *CharstreamString) Next() rune {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	return inst.next()
}

// internal next, no locking, the caller need to make sure the inst is lock
func (inst *CharstreamString) next() rune {
	if len(inst.chars) == 0 {
		return EOFChar
	}
	if inst.cursor >= len(inst.chars) {
		return EOFChar
	}

	idx := inst.cursor
	char := inst.chars[idx]
	inst.cursor = inst.cursor + 1
	inst.positionMap.update(idx, char)

	return char
}

func (inst *CharstreamString) Peek() rune {
	inst.lock.RLock()
	defer inst.lock.RUnlock()

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
	for inst.cursor < len(inst.chars) {
		char := inst.chars[inst.cursor]
		if !IsWhiteSpace(char) {
			break
		}
		inst.cursor = inst.cursor + 1
	}

	if inst.cursor > beginCur {
		spaces := inst.chars[beginCur:inst.cursor]
		inst.positionMap.update(beginCur, spaces...)
		return spaces
	}

	return nil
}

// Matches the input rune slice, if succeeded, return the match rune list. If not,
// return false and the slice of rune read in order to try the match
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

// The prepend MUST be chars read just before the cursor.
func newCharstreamPrepend(baseCharstream ICharstream, prepend []rune) ICharstream {
	if len(prepend) == 0 {
		return baseCharstream
	}

	csp := &CharstreamPrepend{}
	csp.charstream = baseCharstream
	if csp.charstream == nil {
		csp.charstream = NewCharstreamFromString("")
	}
	csp.prepend = prepend
	return csp
}

// CharstreamPrepend is thread-safe
type CharstreamPrepend struct {
	prepend    []rune
	charstream ICharstream
	cursor     int // this cursor is up to only len(prepend)
	lock       sync.RWMutex
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
	return inst.charstream.Next()
}

func (inst *CharstreamPrepend) Peek() rune {
	inst.lock.RLock()
	defer inst.lock.RUnlock()
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

func (inst *CharstreamPrepend) Position() *Position {
	if inst.cursor >= len(inst.prepend) {
		return inst.charstream.Position()
	}
	idx := inst.charstream.Cursor() - len(inst.prepend) + inst.cursor
	return inst.charstream.PositionLookup(idx)
}

// Return the position of char index
func (inst *CharstreamPrepend) PositionLookup(idx int) *Position {
	return inst.charstream.PositionLookup(idx)
}

func (inst *CharstreamPrepend) Cursor() int {
	inst.lock.RLock()
	defer inst.lock.RUnlock()

	return inst.charstream.Cursor() - len(inst.prepend) + inst.cursor
}
