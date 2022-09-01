package xbnf

type SpaceProperties struct{}

var whiteSpaceMap = map[rune]*SpaceProperties{
	'\u0009': nil, // TAB
	'\u000A': nil, // LF - LineFeed
	'\u000B': nil, // VT - Vertical Tab
	'\u000C': nil, // FF - Form feed
	'\u000D': nil, // CR - Carriage return
	'\u0020': nil, // ' ' - space
	'\u0085': nil, // NEL - Next line
	'\u00A0': nil, //
	'\u1680': nil,
	'\u2000': nil,
	'\u2001': nil,
	'\u2002': nil,
	'\u2003': nil,
	'\u2004': nil,
	'\u2005': nil,
	'\u2006': nil,
	'\u2007': nil,
	'\u2008': nil,
	'\u2009': nil,
	'\u200A': nil,
	'\u2028': nil,
	'\u2029': nil,
	'\u202F': nil,
	'\u205F': nil,
	'\u3000': nil,
}

var whiteSpaces []rune

func init() {
	for key := range whiteSpaceMap {
		whiteSpaces = append(whiteSpaces, key)
	}
}

func IsWhiteSpace(char rune) bool {
	_, is := whiteSpaceMap[char]
	return is
}
