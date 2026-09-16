package sqlvet

import (
	"fmt"
	"go/ast"
	"regexp"
	"slices"
	"strings"
)

type rule interface {
	ID() string
	Check(fileContext) []diagnostic
}

type queryRule struct {
	id    string
	check func(queryCall) string
}

func (r queryRule) ID() string { return r.id }

func (r queryRule) Check(context fileContext) []diagnostic {
	var diagnostics []diagnostic
	for _, query := range context.queries {
		if message := r.check(query); message != "" {
			diagnostics = append(diagnostics, diagnostic{pos: query.pos, message: message})
		}
	}
	return diagnostics
}

func allRules() []rule {
	return []rule{
		queryRule{id: "SQL001", check: checkSelectStar},
		loopQueryRule{},
		queryRule{id: "SQL003", check: checkForeignKeyIndex},
		queryRule{id: "SQL004", check: checkLeftJoin},
		transactionRule{},
		queryRule{id: "SQL006", check: checkWhereFunction},
		queryRule{id: "SQL007", check: checkHardcodedValue},
		queryRule{id: "SQL008", check: checkDistinct},
		queryRule{id: "SQL009", check: checkUnboundedSelect},
		explainResultRule{},
	}
}

var selectStarPattern = regexp.MustCompile(`(?i)\bselect\s+\*`)

func checkSelectStar(query queryCall) string {
	if selectStarPattern.MatchString(query.sql) {
		return "avoid SELECT *; select only the columns you need"
	}
	return ""
}

type loopQueryRule struct{}

func (loopQueryRule) ID() string { return "SQL002" }

func (loopQueryRule) Check(context fileContext) []diagnostic {
	var diagnostics []diagnostic
	for _, query := range context.queries {
		if !slices.Contains([]string{"Query", "QueryContext", "QueryRow", "QueryRowContext"}, query.method) {
			continue
		}
		for parent := context.parents[query.call]; parent != nil; parent = context.parents[parent] {
			switch parent.(type) {
			case *ast.ForStmt, *ast.RangeStmt:
				diagnostics = append(diagnostics, diagnostic{query.pos, "avoid database queries inside loops; batch or join instead"})
				parent = nil
			case *ast.FuncDecl, *ast.FuncLit:
				parent = nil
			}
		}
	}
	return diagnostics
}

var (
	createTablePattern = regexp.MustCompile(`(?i)\bcreate\s+table\b`)
	foreignKeyPattern  = regexp.MustCompile(`(?is)\bforeign\s+key\s*\(\s*([a-z_][a-z0-9_$]*)\s*\)`)
)

func checkForeignKeyIndex(query queryCall) string {
	if !createTablePattern.MatchString(query.sql) {
		return ""
	}
	match := foreignKeyPattern.FindStringSubmatch(query.sql)
	if len(match) == 0 {
		return ""
	}
	column := match[1]
	indexPattern := regexp.MustCompile(`(?is)(?:\bcreate\s+(?:unique\s+)?index\b[^;]*|\b(?:primary\s+key|unique)\b\s*)\(\s*` + regexp.QuoteMeta(column) + `(?:\s*[,)]|\s+(?:asc|desc)\s*[,)])`)
	if indexPattern.MatchString(query.sql) {
		return ""
	}
	return fmt.Sprintf("foreign key %s has no supporting index in this DDL statement", column)
}

var (
	leftJoinPattern = regexp.MustCompile(`(?is)\bleft(?:\s+outer)?\s+join\s+[a-z_][a-z0-9_$.]*\s+(?:as\s+)?([a-z_][a-z0-9_$]*)\s+on\b`)
	wherePattern    = regexp.MustCompile(`(?i)\bwhere\b`)
)

func checkLeftJoin(query queryCall) string {
	join := leftJoinPattern.FindStringSubmatch(query.sql)
	whereIndex := wherePattern.FindStringIndex(query.sql)
	if len(join) == 0 || whereIndex == nil {
		return ""
	}
	alias := join[1]
	where := query.sql[whereIndex[0]:]
	columnPattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(alias) + `\.[a-z_][a-z0-9_$]*\b`)
	if !columnPattern.MatchString(where) {
		return ""
	}
	nullCheckPattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(alias) + `\.[a-z_][a-z0-9_$]*\s+is\s+null\b`)
	if nullCheckPattern.MatchString(where) {
		return ""
	}
	return fmt.Sprintf("LEFT JOIN alias %s is null-rejected in WHERE; use INNER JOIN or move the predicate", alias)
}

type transactionRule struct{}

func (transactionRule) ID() string { return "SQL005" }

var writePattern = regexp.MustCompile(`(?i)^\s*(?:insert|update|delete|merge|replace)\b`)

func (transactionRule) Check(context fileContext) []diagnostic {
	writesByFunction := make(map[*ast.FuncDecl][]queryCall)
	for _, query := range context.queries {
		if query.receiverType != "DB" || !writePattern.MatchString(query.sql) {
			continue
		}
		function := ancestor[*ast.FuncDecl](context, query.call)
		if function != nil {
			writesByFunction[function] = append(writesByFunction[function], query)
		}
	}
	var diagnostics []diagnostic
	for _, writes := range writesByFunction {
		for _, query := range writes[1:] {
			diagnostics = append(diagnostics, diagnostic{query.pos, "multiple writes through *sql.DB should be wrapped in a transaction"})
		}
	}
	return diagnostics
}

var whereFunctionPattern = regexp.MustCompile(`(?is)\bwhere\b.*?\b(date|lower|upper|trim|substr|substring|coalesce)\s*\(\s*[a-z_][a-z0-9_$.]*\s*\)`)

func checkWhereFunction(query queryCall) string {
	match := whereFunctionPattern.FindStringSubmatch(query.sql)
	if len(match) == 0 {
		return ""
	}
	return fmt.Sprintf("avoid applying %s to a column in WHERE; use a searchable range predicate", strings.ToUpper(match[1]))
}

var hardcodedPredicatePattern = regexp.MustCompile(`(?is)\bwhere\b.*?(?:=|<>|!=|<=|>=|<|>)\s*('(?:''|[^'])*'|[0-9]+(?:\.[0-9]+)?)`)

func checkHardcodedValue(query queryCall) string {
	match := hardcodedPredicatePattern.FindStringSubmatch(query.sql)
	if len(match) == 0 {
		return ""
	}
	return fmt.Sprintf("parameterize literal %s in the SQL statement", match[1])
}

var distinctPattern = regexp.MustCompile(`(?i)\bselect\s+distinct\b`)

func checkDistinct(query queryCall) string {
	if distinctPattern.MatchString(query.sql) {
		return "review SELECT DISTINCT; fix duplicate joins or data at the source"
	}
	return ""
}

var (
	selectPattern    = regexp.MustCompile(`(?is)^\s*(?:/\*.*?\*/\s*)*select\b`)
	boundPattern     = regexp.MustCompile(`(?i)\b(?:limit|fetch\s+(?:first|next))\b|^\s*select\s+top\b`)
	aggregatePattern = regexp.MustCompile(`(?i)^\s*select\s+(?:count|exists|max|min|sum|avg)\s*\(`)
)

func checkUnboundedSelect(query queryCall) string {
	if strings.HasPrefix(query.method, "QueryRow") || !selectPattern.MatchString(query.sql) || boundPattern.MatchString(query.sql) || aggregatePattern.MatchString(query.sql) {
		return ""
	}
	return "bound SELECT results with LIMIT, FETCH, TOP, or pagination"
}

type explainResultRule struct{}

func (explainResultRule) ID() string { return "SQL010" }

var explainPattern = regexp.MustCompile(`(?i)^\s*explain\b`)

func (explainResultRule) Check(context fileContext) []diagnostic {
	var diagnostics []diagnostic
	for _, query := range context.queries {
		if !explainPattern.MatchString(query.sql) {
			continue
		}
		if _, discarded := context.parents[query.call].(*ast.ExprStmt); discarded {
			diagnostics = append(diagnostics, diagnostic{query.pos, "EXPLAIN result is discarded; inspect the returned execution plan"})
		}
	}
	return diagnostics
}
