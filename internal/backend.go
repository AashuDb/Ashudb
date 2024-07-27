package ashudb

import "errors"

type ColumnType uint

const (
	TextType ColumnType = iota
	IntType
)

type Cell interface {
	AsText() string
	AsInt() int32
}

type Results struct {
	Columns []struct {
		Type ColumnType
		Name string
	}
	Rows [][]Cell
}

var (
	ErrTableDoesNotExist  = errors.New("table does not exist")
	ErrColumnDoesNotExist = errors.New("column does not exist")
	ErrInvalidSelectItem  = errors.New("select item is not valid")
	ErrInvalidDatatype    = errors.New("invalid datatype")
	ErrMissingValues      = errors.New("missing values")
)

type Backend interface {
	CreateTable(*CreateTableStatement) error
	Insert(*InsertStatement) error
	Select(*SelectStatement) (*Results, error)
}
