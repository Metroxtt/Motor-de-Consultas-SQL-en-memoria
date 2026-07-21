package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/catalog"
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/engine"
	"github.com/Metroxtt/Motor-de-Consultas-SQL-en-memoria/internal/parser"
)

func main() {
	cat := catalog.NewCatalog()

	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			tableName := tableNameFromFile(arg)
			t, err := catalog.LoadCSVFile(tableName, arg, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error cargando %s: %v\n", arg, err)
				continue
			}
			cat.AddTable(t)
			fmt.Printf("Tabla %q cargada: %d filas, %d columnas\n",
				tableName, len(t.Rows), len(t.Schema.Columns))
		}
	}

	fmt.Println("Motor de Consultas SQL en Memoria")
	fmt.Println("Escriba una consulta SQL o 'exit' para salir.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("sql> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "exit" || strings.ToLower(input) == "quit" {
			fmt.Println("Adiós!")
			break
		}
		if strings.ToLower(input) == "tables" {
			printTables(cat)
			continue
		}

		executeQuery(input, cat)
	}
}

func executeQuery(input string, cat *catalog.Catalog) {
	node, err := parser.Parse(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error de sintaxis: %v\n", err)
		return
	}

	plan, err := engine.BuildPlan(node, cat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error de planificación: %v\n", err)
		return
	}
	defer plan.Close()

	table, _ := cat.GetTable(node.Table)
	headers := resolveHeaders(node, table)

	printRow(headers, false)
	printSeparator(headers)

	count := 0
	for {
		row, err := plan.Next()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error de ejecución: %v\n", err)
			return
		}
		if row == nil {
			break
		}

		vals := make([]string, len(headers))
		for i, h := range headers {
			vals[i] = formatValue(row[h])
		}
		printRow(vals, false)
		count++
	}

	fmt.Printf("(%d filas)\n\n", count)
}

func resolveHeaders(node *parser.SelectNode, table *catalog.Table) []string {
	if node.Columns == nil {
		headers := make([]string, len(table.Schema.Columns))
		for i, col := range table.Schema.Columns {
			headers[i] = col.Name
		}
		return headers
	}

	headers := make([]string, len(node.Columns))
	for i, col := range node.Columns {
		if ref, ok := col.(*parser.ColumnRefNode); ok {
			headers[i] = ref.Name
		} else {
			headers[i] = fmt.Sprintf("col_%d", i)
		}
	}
	return headers
}

func printRow(vals []string, header bool) {
	for i, v := range vals {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Printf("%-20s", v)
	}
	fmt.Println()
}

func printSeparator(vals []string) {
	for i := range vals {
		if i > 0 {
			fmt.Print("-+-")
		}
		fmt.Print(strings.Repeat("-", 20))
	}
	fmt.Println()
}

func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	return fmt.Sprintf("%v", v)
}

func printTables(cat *catalog.Catalog) {
	names := cat.TableNames()
	if len(names) == 0 {
		fmt.Println("No hay tablas cargadas.")
		return
	}
	for _, name := range names {
		t, _ := cat.GetTable(name)
		cols := make([]string, len(t.Schema.Columns))
		for i, c := range t.Schema.Columns {
			cols[i] = fmt.Sprintf("%s:%s", c.Name, c.Type)
		}
		fmt.Printf("  %s (%d filas): %s\n", name, len(t.Rows), strings.Join(cols, ", "))
	}
	fmt.Println()
}

func tableNameFromFile(path string) string {
	base := path
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	}
	if idx := strings.LastIndex(base, "\\"); idx >= 0 {
		base = base[idx+1:]
	}
	base = strings.TrimSuffix(base, ".csv")
	base = strings.TrimSuffix(base, ".CSV")
	return base
}
