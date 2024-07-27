package ashudb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		source string
		ast    *Ast
	}{
		{
			source: "INSERT INTO users VALUES (105, 233);",
			ast: &Ast{
				Statements: []*Statement{
					{
						Kind: InsertKind,
						InsertStatement: &InsertStatement{
							table: Token{
								Loc:   Location{Col: 12, Line: 0},
								Kind:  IdentifierKind,
								Value: "users",
							},
							values: &[]*expression{
								{
									literal: &Token{
										Loc:   Location{Col: 26, Line: 0},
										Kind:  NumericKind,
										Value: "105",
									},
									kind: literalKind,
								},
								{
									literal: &Token{
										Loc:   Location{Col: 32, Line: 0},
										Kind:  NumericKind,
										Value: "233",
									},
									kind: literalKind,
								},
							},
						},
					},
				},
			},
		},
		{
			source: "CREATE TABLE users (id INT, name TEXT);",
			ast: &Ast{
				Statements: []*Statement{
					{
						Kind: CreateTableKind,
						CreateTableStatement: &CreateTableStatement{
							name: Token{
								Loc:   Location{Col: 13, Line: 0},
								Kind:  IdentifierKind,
								Value: "users",
							},
							cols: &[]*columnDefinition{
								{
									name: Token{
										Loc:   Location{Col: 20, Line: 0},
										Kind:  IdentifierKind,
										Value: "id",
									},
									datatype: Token{
										Loc:   Location{Col: 23, Line: 0},
										Kind:  KeywordKind,
										Value: "int",
									},
								},
								{
									name: Token{
										Loc:   Location{Col: 28, Line: 0},
										Kind:  IdentifierKind,
										Value: "name",
									},
									datatype: Token{
										Loc:   Location{Col: 33, Line: 0},
										Kind:  KeywordKind,
										Value: "text",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			source: "SELECT *, exclusive;",
			ast: &Ast{
				Statements: []*Statement{
					{
						Kind: SelectKind,
						SelectStatement: &SelectStatement{
							item: &[]*selectItem{
								{
									asterisk: true,
								},
								{
									exp: &expression{
										kind: literalKind,
										literal: &Token{
											Loc:   Location{Col: 10, Line: 0},
											Kind:  IdentifierKind,
											Value: "exclusive",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			source: "SELECT id, name AS fullname FROM users;",
			ast: &Ast{
				Statements: []*Statement{
					{
						Kind: SelectKind,
						SelectStatement: &SelectStatement{
							item: &[]*selectItem{
								{
									exp: &expression{
										kind: literalKind,
										literal: &Token{
											Loc:   Location{Col: 7, Line: 0},
											Kind:  IdentifierKind,
											Value: "id",
										},
									},
								},
								{
									exp: &expression{
										kind: literalKind,
										literal: &Token{
											Loc:   Location{Col: 11, Line: 0},
											Kind:  IdentifierKind,
											Value: "name",
										},
									},
									as: &Token{
										Loc:   Location{Col: 19, Line: 0},
										Kind:  IdentifierKind,
										Value: "fullname",
									},
								},
							},
							from: &fromItem{
								table: &Token{
									Loc:   Location{Col: 33, Line: 0},
									Kind:  IdentifierKind,
									Value: "users",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		ast, err := Parse(test.source)
		assert.Nil(t, err, test.source)
		assert.Equal(t, test.ast, ast, test.source)
	}
}
