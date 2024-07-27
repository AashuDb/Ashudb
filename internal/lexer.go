package ashudb

import (
	"fmt"
	"strings"
)

type Location struct {
	Line uint
	Col  uint
}

type Keyword string

const (
	SelectKeyword Keyword = "select"
	FromKeyword   Keyword = "from"
	AsKeyword     Keyword = "as"
	TableKeyword  Keyword = "table"
	CreateKeyword Keyword = "create"
	WhereKeyword  Keyword = "where"
	InsertKeyword Keyword = "insert"
	IntoKeyword   Keyword = "into"
	ValuesKeyword Keyword = "values"
	IntKeyword    Keyword = "int"
	TextKeyword   Keyword = "text"
)

type Symbol string

const (
	SemicolonSymbol  Symbol = ";"
	AsteriskSymbol   Symbol = "*"
	CommaSymbol      Symbol = ","
	LeftParenSymbol  Symbol = "("
	RightParenSymbol Symbol = ")"
)

type TokenKind uint

const (
	KeywordKind TokenKind = iota
	SymbolKind
	IdentifierKind
	StringKind
	NumericKind
)

type Token struct {
	Value string
	Kind  TokenKind
	Loc   Location
}

type Cursor struct {
	Pointer uint
	Loc     Location
}

func (t *Token) equals(other *Token) bool {
	return t.Value == other.Value && t.Kind == other.Kind
}

type lexer func(string, Cursor) (*Token, Cursor, bool)

func lex(source string) ([]*Token, error) {
	tokens := []*Token{}
	cur := Cursor{}

lex:
	for cur.Pointer < uint(len(source)) {
		lexers := []lexer{lexKeyword, lexSymbol, lexString, lexNumeric, lexIdentifier}
		for _, l := range lexers {
			if token, newCursor, ok := l(source, cur); ok {
				cur = newCursor

				// Omit nil tokens for valid, but empty syntax like newlines
				if token != nil {
					tokens = append(tokens, token)
				}

				continue lex
			}
		}

		hint := ""
		if len(tokens) > 0 {
			hint = " after " + tokens[len(tokens)-1].Value
		}
		return nil, fmt.Errorf("unable to lex token%s, at %d:%d", hint, cur.Loc.Line, cur.Loc.Col)
	}

	return tokens, nil
}

func lexNumeric(source string, ic Cursor) (*Token, Cursor, bool) {
	cur := ic

	periodFound := false
	expMarkerFound := false

	for ; cur.Pointer < uint(len(source)); cur.Pointer++ {
		c := source[cur.Pointer]
		cur.Loc.Col++

		isDigit := c >= '0' && c <= '9'
		isPeriod := c == '.'
		isExpMarker := c == 'e'

		// Must start with a digit or period
		if cur.Pointer == ic.Pointer {
			if !isDigit && !isPeriod {
				return nil, ic, false
			}

			periodFound = isPeriod
			continue
		}

		if isPeriod {
			if periodFound {
				return nil, ic, false
			}

			periodFound = true
			continue
		}

		if isExpMarker {
			if expMarkerFound {
				return nil, ic, false
			}

			// No periods allowed after expMarker
			periodFound = true
			expMarkerFound = true

			// expMarker must be followed by digits
			if cur.Pointer == uint(len(source)-1) {
				return nil, ic, false
			}

			cNext := source[cur.Pointer+1]
			if cNext == '-' || cNext == '+' {
				cur.Pointer++
				cur.Loc.Col++
			}

			continue
		}

		if !isDigit {
			break
		}
	}

	// No characters accumulated
	if cur.Pointer == ic.Pointer {
		return nil, ic, false
	}

	return &Token{
		Value: source[ic.Pointer:cur.Pointer],
		Loc:   ic.Loc,
		Kind:  NumericKind,
	}, cur, true
}

func lexCharacterDelimited(source string, ic Cursor, delimiter byte) (*Token, Cursor, bool) {
	cur := ic

	if len(source[cur.Pointer:]) == 0 {
		return nil, ic, false
	}

	if source[cur.Pointer] != delimiter {
		return nil, ic, false
	}

	cur.Loc.Col++
	cur.Pointer++

	var value []byte
	for ; cur.Pointer < uint(len(source)); cur.Pointer++ {
		c := source[cur.Pointer]

		if c == delimiter {
			// SQL escapes are via double characters, not backslash.
			if cur.Pointer+1 >= uint(len(source)) || source[cur.Pointer+1] != delimiter {
				cur.Pointer++
				return &Token{
					Value: string(value),
					Loc:   ic.Loc,
					Kind:  StringKind,
				}, cur, true
			}
			// else {
			// 	value = append(value, delimiter)
			// 	cur.Pointer++
			// 	cur.Loc.Col++
			// }
		}

		value = append(value, c)
		cur.Loc.Col++
	}

	return nil, ic, false
}

func lexString(source string, ic Cursor) (*Token, Cursor, bool) {
	return lexCharacterDelimited(source, ic, '\'')
}

func lexSymbol(source string, ic Cursor) (*Token, Cursor, bool) {
	c := source[ic.Pointer]
	cur := ic
	// Will get overwritten later if not an ignored syntax
	cur.Pointer++
	cur.Loc.Col++

	switch c {
	// Syntax that should be thrown away
	case '\n':
		cur.Loc.Line++
		cur.Loc.Col = 0
		fallthrough
	case '\t':
		fallthrough
	case ' ':
		return nil, cur, true
	}

	// Syntax that should be kept
	symbols := []Symbol{
		CommaSymbol,
		LeftParenSymbol,
		RightParenSymbol,
		SemicolonSymbol,
		AsteriskSymbol,
	}

	var options []string
	for _, s := range symbols {
		options = append(options, string(s))
	}

	// Use `ic`, not `cur`
	match := longestMatch(source, ic, options)
	// Unknown character
	if match == "" {
		return nil, ic, false
	}

	cur.Pointer = ic.Pointer + uint(len(match))
	cur.Loc.Col = ic.Loc.Col + uint(len(match))

	return &Token{
		Value: match,
		Loc:   ic.Loc,
		Kind:  SymbolKind,
	}, cur, true
}

func lexKeyword(source string, ic Cursor) (*Token, Cursor, bool) {
	cur := ic
	keywords := []Keyword{
		SelectKeyword,
		FromKeyword,
		AsKeyword,
		TableKeyword,
		CreateKeyword,
		WhereKeyword,
		InsertKeyword,
		IntoKeyword,
		ValuesKeyword,
		IntKeyword,
		TextKeyword,
	}

	var options []string
	for _, k := range keywords {
		options = append(options, string(k))
	}

	match := longestMatch(source, ic, options)
	if match == "" {
		return nil, ic, false
	}

	cur.Pointer = ic.Pointer + uint(len(match))
	cur.Loc.Col = ic.Loc.Col + uint(len(match))

	return &Token{
		Value: match,
		Kind:  KeywordKind,
		Loc:   ic.Loc,
	}, cur, true
}

// longestMatch iterates through a source string starting at the given
// cursor to find the longest matching substring among the provided
// options
func longestMatch(source string, ic Cursor, options []string) string {
	var value []byte
	var skipList []int
	var match string

	cur := ic

	for cur.Pointer < uint(len(source)) {

		value = append(value, strings.ToLower(string(source[cur.Pointer]))...)
		cur.Pointer++

	match:
		for i, option := range options {
			for _, skip := range skipList {
				if i == skip {
					continue match
				}
			}

			// Deal with cases like INT vs INTO
			if option == string(value) {
				skipList = append(skipList, i)
				if len(option) > len(match) {
					match = option
				}

				continue
			}

			sharesPrefix := string(value) == option[:cur.Pointer-ic.Pointer]
			tooLong := len(value) > len(option)
			if tooLong || !sharesPrefix {
				skipList = append(skipList, i)
			}
		}

		if len(skipList) == len(options) {
			break
		}
	}

	return match
}

func lexIdentifier(source string, ic Cursor) (*Token, Cursor, bool) {
	// Handle separately if is a double-quoted identifier
	if token, newCursor, ok := lexCharacterDelimited(source, ic, '"'); ok {
		return token, newCursor, true
	}

	cur := ic

	c := source[cur.Pointer]
	// Other characters count too, big ignoring non-ascii for now
	isAlphabetical := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	if !isAlphabetical {
		return nil, ic, false
	}
	cur.Pointer++
	cur.Loc.Col++

	value := []byte{c}
	for ; cur.Pointer < uint(len(source)); cur.Pointer++ {
		c = source[cur.Pointer]

		// Other characters count too, big ignoring non-ascii for now
		isAlphabetical := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
		isNumeric := c >= '0' && c <= '9'
		if isAlphabetical || isNumeric || c == '$' || c == '_' {
			value = append(value, c)
			cur.Loc.Col++
			continue
		}

		break
	}

	if len(value) == 0 {
		return nil, ic, false
	}

	return &Token{
		// Unquoted dentifiers are case-insensitive
		Value: strings.ToLower(string(value)),
		Loc:   ic.Loc,
		Kind:  IdentifierKind,
	}, cur, true
}
