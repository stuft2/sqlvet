# sqlvet

`sqlvet` is a Go static analyzer for SQL query anti-patterns. It inspects
compile-time string constants passed to query methods on `database/sql.DB`,
`database/sql.Tx`, and `database/sql.Conn`.

The analyzer implements all ten rules from
[`docs/sql-rules.md`](docs/sql-rules.md):

| Rule | Check |
| --- | --- |
| `SQL001` | `SELECT *` |
| `SQL002` | Query calls inside Go loops |
| `SQL003` | Foreign keys without a supporting index in the same DDL string |
| `SQL004` | `LEFT JOIN` aliases null-rejected by `WHERE` |
| `SQL005` | Multiple writes through `*sql.DB` in one function |
| `SQL006` | Functions applied to columns in `WHERE` |
| `SQL007` | Literal values in `WHERE` predicates |
| `SQL008` | `SELECT DISTINCT` |
| `SQL009` | Multi-row `SELECT` queries without a result bound |
| `SQL010` | Discarded `EXPLAIN` result rows |

## Develop

Run the analyzer tests and build the standalone command:

```sh
task test
task build
```

Run the standalone analyzer against a Go package:

```sh
go run ./cmd/sqlvet ./path/to/package/...
```

## Use with golangci-lint

This repository uses golangci-lint's module plugin system. Build the custom
binary using the checked-in `.custom-gcl.yml`:

```sh
task build:golangci-lint
```

Then copy `.golangci.example.yml` to the project being checked and run:

```sh
./.bin/custom-golangci-lint run --config .golangci.example.yml ./...
```

The build configuration pins golangci-lint `v2.9.0`. Update that version
deliberately when upgrading the custom runner.

## Current scope

Only statically known strings passed directly to supported `database/sql`
methods are inspected. Queries assembled at runtime and calls through custom
database interfaces are intentionally skipped.

Some review concerns cannot be proven locally. In particular, sqlvet cannot
inspect a production schema or infer business intent. `SQL003`, `SQL004`,
`SQL005`, and `SQL010` therefore check narrow, actionable signals rather than
claiming that every index, join, transaction boundary, or execution plan is
correct. Enable or disable individual rule IDs in the golangci-lint settings to
match the project.
