package main

import (
	"fmt"
	"strings"

	ashudb "github.com/aashudb/ashudb/internal"
)

func main() {
	mb := ashudb.NewMemoryBackend()

	// reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to AshuDB.")
	for {
		fmt.Print("# ")
		// text, _ := reader.ReadString('\n')
		text := "INSERT 'a '' b'"
		text = strings.Replace(text, "\n", "", -1)

		ast, err := ashudb.Parse(text)
		if err != nil {
			panic(err)
		}

		for _, stmt := range ast.Statements {
			switch stmt.Kind {
			case ashudb.CreateTableKind:
				err = mb.CreateTable(ast.Statements[0].CreateTableStatement)
				if err != nil {
					panic(err)
				}
				fmt.Println("huss")
			case ashudb.InsertKind:
				err = mb.Insert(stmt.InsertStatement)
				if err != nil {
					panic(err)
				}

				fmt.Println("huss")
			case ashudb.SelectKind:
				results, err := mb.Select(stmt.SelectStatement)
				if err != nil {
					panic(err)
				}

				for _, col := range results.Columns {
					fmt.Printf("| %s ", col.Name)
				}
				fmt.Println("|")

				for i := 0; i < 20; i++ {
					fmt.Printf("=")
				}
				fmt.Println()

				for _, result := range results.Rows {
					fmt.Printf("|")

					for i, cell := range result {
						typ := results.Columns[i].Type
						s := ""
						switch typ {
						case ashudb.IntType:
							s = fmt.Sprintf("%d", cell.AsInt())
						case ashudb.TextType:
							s = cell.AsText()
						}

						fmt.Printf(" %s | ", s)
					}

					fmt.Println()
				}

				fmt.Println("huss")
			}
		}
	}
}
