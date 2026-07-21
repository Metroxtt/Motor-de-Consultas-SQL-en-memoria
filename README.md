# Motor de Consultas SQL en Memoria

Motor SQL reducido en Go que carga datos desde CSV, parsea consultas SQL y las ejecuta en memoria usando el modelo de iteradores (Volcano).

## Estructura del proyecto

```
.
├── cmd/            # Punto de entrada (CLI/REPL)
├── internal/       # Paquetes internos
│   ├── lexer/      # Tokenizador SQL
│   ├── parser/     # Parser → AST
│   ├── engine/     # Motor de ejecución (operadores)
│   └── catalog/    # Catálogo de tablas y esquemas
├── data/           # Archivos CSV de ejemplo
└── docs/           # Gramática EBFN, bitácora de decisiones
```

## Compilación y ejecución

```bash
go build -o sqlengine ./cmd/...
./sqlengine
```

## Pruebas

```bash
go test ./...
go test -race ./...
```
