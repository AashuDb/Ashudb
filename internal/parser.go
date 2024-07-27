package ashudb

import (
	"errors"
	"fmt"
)

type Ast struct {
	Statements []*Statement
}

type AstKind uint

const (
	SelectKind AstKind = iota
	CreateTableKind
	InsertKind
)

type expressionKind uint

const (
	literalKind expressionKind = iota
)

type expression struct {
	literal *Token
	kind    expressionKind
}

type selectItem struct {
	exp      *expression
	asterisk bool
	as       *Token
}

type fromItem struct {
	table *Token
}

type columnDefinition struct {
	name     Token
	datatype Token
}

type Statement struct {
	SelectStatement      *SelectStatement
	CreateTableStatement *CreateTableStatement
	InsertStatement      *InsertStatement
	Kind                 AstKind
}

type InsertStatement struct {
	table  Token
	values *[]*expression
}

type CreateTableStatement struct {
	name Token
	cols *[]*columnDefinition
}

type SelectStatement struct {
	item *[]*selectItem
	from *fromItem
}

func tokenFromKeyword(k Keyword) Token {
	return Token{
		Kind:  KeywordKind,
		Value: string(k),
	}
}

func tokenFromSymbol(s Symbol) Token {
	return Token{
		Kind:  SymbolKind,
		Value: string(s),
	}
}

func expectToken(tokens []*Token, cursor uint, t Token) bool {
	if cursor >= uint(len(tokens)) {
		return false
	}

	return t.equals(tokens[cursor])
}

func helpMessage(tokens []*Token, cursor uint, msg string) {
	var c *Token
	if cursor < uint(len(tokens)) {
		c = tokens[cursor]
	} else {
		c = tokens[cursor-1]
	}

	fmt.Printf("[%d,%d]: %s, got: %s\n", c.Loc.Line, c.Loc.Col, msg, c.Value)
}

func Parse(source string) (*Ast, error) {
	tokens, err := lex(source)
	if err != nil {
		return nil, err
	}

	a := Ast{}
	cursor := uint(0)
	for cursor < uint(len(tokens)) {
		stmt, newCursor, ok := parseStatement(tokens, cursor, tokenFromSymbol(SemicolonSymbol))
		if !ok {
			helpMessage(tokens, cursor, "Expected statement")
			return nil, errors.New("failed to parse, expected statement")
		}
		cursor = newCursor

		a.Statements = append(a.Statements, stmt)

		atLeastOneSemicolon := false
		for expectToken(tokens, cursor, tokenFromSymbol(SemicolonSymbol)) {
			cursor++
			atLeastOneSemicolon = true
		}

		if !atLeastOneSemicolon {
			helpMessage(tokens, cursor, "Expected semi-colon delimiter between statements")
			return nil, errors.New("missing semi-colon between statements")
		}
	}

	return &a, nil
}

func parseStatement(tokens []*Token, initialCursor uint, delimiter Token) (*Statement, uint, bool) {
	cursor := initialCursor

	// Look for a SELECT statement
	slct, newCursor, ok := parseSelectStatement(tokens, cursor, delimiter)
	if ok {
		return &Statement{
			Kind:            SelectKind,
			SelectStatement: slct,
		}, newCursor, true
	}

	// Look for a INSERT statement
	inst, newCursor, ok := parseInsertStatement(tokens, cursor, delimiter)
	if ok {
		return &Statement{
			Kind:            InsertKind,
			InsertStatement: inst,
		}, newCursor, true
	}

	// Look for a CREATE statement
	crtTbl, newCursor, ok := parseCreateTableStatement(tokens, cursor, delimiter)
	if ok {
		return &Statement{
			Kind:                 CreateTableKind,
			CreateTableStatement: crtTbl,
		}, newCursor, true
	}

	return nil, initialCursor, false
}

func parseExpressions(tokens []*Token, initialCursor uint, delimiters []Token) (*[]*expression, uint, bool) {
	cursor := initialCursor

	exps := []*expression{}
outer:
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		// Look for delimiter
		current := tokens[cursor]
		for _, delimiter := range delimiters {
			if delimiter.equals(current) {
				break outer
			}
		}

		// Look for comma
		if len(exps) > 0 {
			if !expectToken(tokens, cursor, tokenFromSymbol(CommaSymbol)) {
				helpMessage(tokens, cursor, "Expected comma")
				return nil, initialCursor, false
			}

			cursor++
		}

		// Look for expression
		exp, newCursor, ok := parseExpression(tokens, cursor, tokenFromSymbol(CommaSymbol))
		if !ok {
			helpMessage(tokens, cursor, "Expected expression")
			return nil, initialCursor, false
		}
		cursor = newCursor

		exps = append(exps, exp)
	}

	return &exps, cursor, true
}

func parseExpression(tokens []*Token, initialCursor uint, _ Token) (*expression, uint, bool) {
	cursor := initialCursor

	kinds := []TokenKind{IdentifierKind, NumericKind, StringKind}
	for _, kind := range kinds {
		t, newCursor, ok := parseToken(tokens, cursor, kind)
		if ok {
			return &expression{
				literal: t,
				kind:    literalKind,
			}, newCursor, true
		}
	}

	return nil, initialCursor, false
}

func parseSelectItem(tokens []*Token, initialCursor uint, delimiters []Token) (*[]*selectItem, uint, bool) {
	cursor := initialCursor

	s := []*selectItem{}
outer:
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		current := tokens[cursor]
		for _, delimiter := range delimiters {
			if delimiter.equals(current) {
				break outer
			}
		}

		if len(s) > 0 {
			if !expectToken(tokens, cursor, tokenFromSymbol(CommaSymbol)) {
				helpMessage(tokens, cursor, "Expected comma")
				return nil, initialCursor, false
			}

			cursor++
		}

		var si selectItem
		if expectToken(tokens, cursor, tokenFromSymbol(AsteriskSymbol)) {
			si = selectItem{asterisk: true}
			cursor++
		} else {
			exp, newCursor, ok := parseExpression(tokens, cursor, tokenFromSymbol(CommaSymbol))
			if !ok {
				helpMessage(tokens, cursor, "Expected expression")
				return nil, initialCursor, false
			}

			cursor = newCursor
			si.exp = exp

			if expectToken(tokens, cursor, tokenFromKeyword(AsKeyword)) {
				cursor++

				id, newCursor, ok := parseToken(tokens, cursor, IdentifierKind)
				if !ok {
					helpMessage(tokens, cursor, "Expected identifier after AS")
					return nil, initialCursor, false
				}

				cursor = newCursor
				si.as = id
			}
		}

		s = append(s, &si)
	}

	return &s, cursor, true
}

func parseFromItem(tokens []*Token, initialCursor uint, _ Token) (*fromItem, uint, bool) {
	ident, newCursor, ok := parseToken(tokens, initialCursor, IdentifierKind)
	if !ok {
		return nil, initialCursor, false
	}

	return &fromItem{table: ident}, newCursor, true
}

func parseSelectStatement(tokens []*Token, initialCursor uint, delimiter Token) (*SelectStatement, uint, bool) {
	cursor := initialCursor
	if !expectToken(tokens, cursor, tokenFromKeyword(SelectKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	slct := SelectStatement{}

	exps, newCursor, ok := parseSelectItem(tokens, cursor, []Token{tokenFromKeyword(FromKeyword), delimiter})
	if !ok {
		return nil, initialCursor, false
	}

	slct.item = exps
	cursor = newCursor

	if expectToken(tokens, cursor, tokenFromKeyword(FromKeyword)) {
		cursor++

		from, newCursor, ok := parseFromItem(tokens, cursor, delimiter)
		if !ok {
			helpMessage(tokens, cursor, "Expected FROM token")
			return nil, initialCursor, false
		}

		slct.from = from
		cursor = newCursor
	}

	return &slct, cursor, true
}

func parseToken(tokens []*Token, initialCursor uint, kind TokenKind) (*Token, uint, bool) {
	cursor := initialCursor

	if cursor >= uint(len(tokens)) {
		return nil, initialCursor, false
	}

	current := tokens[cursor]
	if current.Kind == kind {
		return current, cursor + 1, true
	}

	return nil, initialCursor, false
}

func parseInsertStatement(tokens []*Token, initialCursor uint, delimiter Token) (*InsertStatement, uint, bool) {
	cursor := initialCursor

	// Look for INSERT
	if !expectToken(tokens, cursor, tokenFromKeyword(InsertKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	// Look for INTO
	if !expectToken(tokens, cursor, tokenFromKeyword(IntoKeyword)) {
		helpMessage(tokens, cursor, "Expected into")
		return nil, initialCursor, false
	}
	cursor++

	// Look for table name
	table, newCursor, ok := parseToken(tokens, cursor, IdentifierKind)
	if !ok {
		helpMessage(tokens, cursor, "Expected table name")
		return nil, initialCursor, false
	}
	cursor = newCursor

	// Look for VALUES
	if !expectToken(tokens, cursor, tokenFromKeyword(ValuesKeyword)) {
		helpMessage(tokens, cursor, "Expected VALUES")
		return nil, initialCursor, false
	}
	cursor++

	// Look for left paren
	if !expectToken(tokens, cursor, tokenFromSymbol(LeftParenSymbol)) {
		helpMessage(tokens, cursor, "Expected left paren")
		return nil, initialCursor, false
	}
	cursor++

	// Look for expression list
	values, newCursor, ok := parseExpressions(tokens, cursor, []Token{tokenFromSymbol(RightParenSymbol)})
	if !ok {
		return nil, initialCursor, false
	}
	cursor = newCursor

	// Look for right paren
	if !expectToken(tokens, cursor, tokenFromSymbol(RightParenSymbol)) {
		helpMessage(tokens, cursor, "Expected right paren")
		return nil, initialCursor, false
	}
	cursor++

	return &InsertStatement{
		table:  *table,
		values: values,
	}, cursor, true
}

func parseCreateTableStatement(tokens []*Token, initialCursor uint, delimiter Token) (*CreateTableStatement, uint, bool) {
	cursor := initialCursor

	if !expectToken(tokens, cursor, tokenFromKeyword(CreateKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	if !expectToken(tokens, cursor, tokenFromKeyword(TableKeyword)) {
		return nil, initialCursor, false
	}
	cursor++

	name, newCursor, ok := parseToken(tokens, cursor, IdentifierKind)
	if !ok {
		helpMessage(tokens, cursor, "Expected table name")
		return nil, initialCursor, false
	}
	cursor = newCursor

	if !expectToken(tokens, cursor, tokenFromSymbol(LeftParenSymbol)) {
		helpMessage(tokens, cursor, "Expected left parenthesis")
		return nil, initialCursor, false
	}
	cursor++

	cols, newCursor, ok := parseColumnDefinitions(tokens, cursor, tokenFromSymbol(RightParenSymbol))
	if !ok {
		return nil, initialCursor, false
	}
	cursor = newCursor

	if !expectToken(tokens, cursor, tokenFromSymbol(RightParenSymbol)) {
		helpMessage(tokens, cursor, "Expected right parenthesis")
		return nil, initialCursor, false
	}
	cursor++

	return &CreateTableStatement{
		name: *name,
		cols: cols,
	}, cursor, true
}

func parseColumnDefinitions(tokens []*Token, initialCursor uint, delimiter Token) (*[]*columnDefinition, uint, bool) {
	cursor := initialCursor

	cds := []*columnDefinition{}
	for {
		if cursor >= uint(len(tokens)) {
			return nil, initialCursor, false
		}

		// Look for a delimiter
		current := tokens[cursor]
		if delimiter.equals(current) {
			break
		}

		// Look for a comma
		if len(cds) > 0 {
			if !expectToken(tokens, cursor, tokenFromSymbol(CommaSymbol)) {
				helpMessage(tokens, cursor, "Expected comma")
				return nil, initialCursor, false
			}

			cursor++
		}

		// Look for a column name
		id, newCursor, ok := parseToken(tokens, cursor, IdentifierKind)
		if !ok {
			helpMessage(tokens, cursor, "Expected column name")
			return nil, initialCursor, false
		}
		cursor = newCursor

		// Look for a column type
		ty, newCursor, ok := parseToken(tokens, cursor, KeywordKind)
		if !ok {
			helpMessage(tokens, cursor, "Expected column type")
			return nil, initialCursor, false
		}
		cursor = newCursor

		cds = append(cds, &columnDefinition{
			name:     *id,
			datatype: *ty,
		})
	}

	return &cds, cursor, true
}
