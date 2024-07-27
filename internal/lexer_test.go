package ashudb

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToken_lexNumeric(t *testing.T) {
	tests := []struct {
		number bool
		value  string
	}{
		{
			number: true,
			value:  "105",
		},
		{
			number: true,
			value:  "105 ",
		},
		{
			number: true,
			value:  "123.",
		},
		{
			number: true,
			value:  "123.145",
		},
		{
			number: true,
			value:  "1e5",
		},
		{
			number: true,
			value:  "1.e21",
		},
		{
			number: true,
			value:  "1.1e2",
		},
		{
			number: true,
			value:  "1.1e-2",
		},
		{
			number: true,
			value:  "1.1e+2",
		},
		{
			number: true,
			value:  "1e-1",
		},
		{
			number: true,
			value:  ".1",
		},
		{
			number: true,
			value:  "4.",
		},
		// false tests
		{
			number: false,
			value:  "e4",
		},
		{
			number: false,
			value:  "1..",
		},
		{
			number: false,
			value:  "1ee4",
		},
		{
			number: false,
			value:  " 1",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexNumeric(test.value, Cursor{})
		assert.Equal(t, test.number, ok, test.value)
		if ok {
			assert.Equal(t, strings.TrimSpace(test.value), tok.Value, test.value)
		}
	}
}

func TestToken_lexString(t *testing.T) {
	tests := []struct {
		string bool
		value  string
	}{
		{
			string: true,
			value:  "'abc'",
		},
		{
			string: true,
			value:  "'a b'",
		},
		{
			string: true,
			value:  "'a' ",
		},
		{
			string: true,
			value:  "'a '' b'",
		},
		// false tests
		{
			string: false,
			value:  "'",
		},
		{
			string: false,
			value:  "",
		},
		{
			string: false,
			value:  " 'foo'",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexString(test.value, Cursor{})
		assert.Equal(t, test.string, ok, test.value)
		if ok {
			test.value = strings.TrimSpace(test.value)
			assert.Equal(t, test.value[1:len(test.value)-1], tok.Value, test.value)
		}
	}
}

func TestToken_lexIdentifier(t *testing.T) {
	tests := []struct {
		identifier bool
		input      string
		value      string
	}{
		{
			identifier: true,
			input:      "abc",
			value:      "abc",
		},
		{
			identifier: true,
			input:      "abc ",
			value:      "abc",
		},
		{
			identifier: true,
			input:      `" abc "`,
			value:      ` abc `,
		},
		{
			identifier: true,
			input:      "a9$",
			value:      "a9$",
		},
		{
			identifier: true,
			input:      "userName",
			value:      "username",
		},
		{
			identifier: true,
			input:      `"userName"`,
			value:      "userName",
		},
		// false tests
		{
			identifier: false,
			input:      `"`,
		},
		{
			identifier: false,
			input:      "_sadsfa",
		},
		{
			identifier: false,
			input:      "9sadsfa",
		},
		{
			identifier: false,
			input:      " abc",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexIdentifier(test.input, Cursor{})
		assert.Equal(t, test.identifier, ok, test.input)
		if ok {
			assert.Equal(t, test.value, tok.Value, test.input)
		}
	}
}

func TestToken_lexKeyword(t *testing.T) {
	tests := []struct {
		keyword bool
		value   string
	}{
		{
			keyword: true,
			value:   "select ",
		},
		{
			keyword: true,
			value:   "from",
		},
		{
			keyword: true,
			value:   "as",
		},
		{
			keyword: true,
			value:   "SELECT",
		},
		{
			keyword: true,
			value:   "into",
		},
		// false tests
		{
			keyword: false,
			value:   " into",
		},
		{
			keyword: false,
			value:   "flubbrety",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexKeyword(test.value, Cursor{})
		assert.Equal(t, test.keyword, ok, test.value)
		if ok {
			test.value = strings.TrimSpace(test.value)
			assert.Equal(t, strings.ToLower(test.value), tok.Value, test.value)
		}
	}
}

func TestLex(t *testing.T) {
	tests := []struct {
		input  string
		tokens []Token
		err    error
	}{
		{
			input: "select 1",
			tokens: []Token{
				{
					Loc:   Location{Col: 0, Line: 0},
					Value: string(SelectKeyword),
					Kind:  KeywordKind,
				},
				{
					Loc:   Location{Col: 7, Line: 0},
					Value: "1",
					Kind:  NumericKind,
				},
			},
			err: nil,
		},
		{
			input: "CREATE TABLE u (id INT, name TEXT)",
			tokens: []Token{
				{
					Loc:   Location{Col: 0, Line: 0},
					Value: string(CreateKeyword),
					Kind:  KeywordKind,
				},
				{
					Loc:   Location{Col: 7, Line: 0},
					Value: string(TableKeyword),
					Kind:  KeywordKind,
				},
				{
					Loc:   Location{Col: 13, Line: 0},
					Value: "u",
					Kind:  IdentifierKind,
				},
				{
					Loc:   Location{Col: 15, Line: 0},
					Value: "(",
					Kind:  SymbolKind,
				},
				{
					Loc:   Location{Col: 16, Line: 0},
					Value: "id",
					Kind:  IdentifierKind,
				},
				{
					Loc:   Location{Col: 19, Line: 0},
					Value: "int",
					Kind:  KeywordKind,
				},
				{
					Loc:   Location{Col: 22, Line: 0},
					Value: ",",
					Kind:  SymbolKind,
				},
				{
					Loc:   Location{Col: 24, Line: 0},
					Value: "name",
					Kind:  IdentifierKind,
				},
				{
					Loc:   Location{Col: 29, Line: 0},
					Value: "text",
					Kind:  KeywordKind,
				},
				{
					Loc:   Location{Col: 33, Line: 0},
					Value: ")",
					Kind:  SymbolKind,
				},
			},
		},
		{
			input: "insert into users values (105, 233)",
			tokens: []Token{
				{
					Loc:   Location{Col: 0, Line: 0},
					Value: string(InsertKeyword),
					Kind:  KeywordKind,
				},
				{
					Loc:   Location{Col: 7, Line: 0},
					Value: string(IntoKeyword),
					Kind:  KeywordKind,
				},
				{
					Loc:   Location{Col: 12, Line: 0},
					Value: "users",
					Kind:  IdentifierKind,
				},
				{
					Loc:   Location{Col: 18, Line: 0},
					Value: string(ValuesKeyword),
					Kind:  KeywordKind,
				},
				{
					Loc:   Location{Col: 25, Line: 0},
					Value: "(",
					Kind:  SymbolKind,
				},
				{
					Loc:   Location{Col: 26, Line: 0},
					Value: "105",
					Kind:  NumericKind,
				},
				{
					Loc:   Location{Col: 30, Line: 0},
					Value: ",",
					Kind:  SymbolKind,
				},
				{
					Loc:   Location{Col: 32, Line: 0},
					Value: "233",
					Kind:  NumericKind,
				},
				{
					Loc:   Location{Col: 36, Line: 0},
					Value: ")",
					Kind:  SymbolKind,
				},
			},
			err: nil,
		},
		{
			input: "SELECT id FROM users;",
			tokens: []Token{
				{
					Loc:   Location{Col: 0, Line: 0},
					Value: string(SelectKeyword),
					Kind:  KeywordKind,
				},
				{
					Loc:   Location{Col: 7, Line: 0},
					Value: "id",
					Kind:  IdentifierKind,
				},
				{
					Loc:   Location{Col: 10, Line: 0},
					Value: string(FromKeyword),
					Kind:  KeywordKind,
				},
				{
					Loc:   Location{Col: 15, Line: 0},
					Value: "users",
					Kind:  IdentifierKind,
				},
				{
					Loc:   Location{Col: 20, Line: 0},
					Value: ";",
					Kind:  SymbolKind,
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		tokens, err := lex(test.input)
		assert.Equal(t, test.err, err, test.input)
		assert.Equal(t, len(test.tokens), len(tokens), test.input)

		for i, tok := range tokens {
			assert.Equal(t, &test.tokens[i], tok, test.input)
		}
	}
}
