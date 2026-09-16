// Package sqlvet defines a Go analyzer for SQL embedded in Go source.
package sqlvet

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"
)

const analyzerName = "sqlvet"

// Settings configures the analyzer. An empty Rules list enables all rules.
type Settings struct {
	Rules []string `json:"rules"`
}

// NewAnalyzer constructs an analyzer. Invalid rule names are reported when the
// analyzer runs; plugin users should use New, which validates settings eagerly.
func NewAnalyzer(settings Settings) *analysis.Analyzer {
	analyzer, err := New(settings)
	if err != nil {
		return &analysis.Analyzer{Name: analyzerName, Doc: "checks SQL queries embedded in Go source", Run: func(*analysis.Pass) (any, error) { return nil, err }}
	}
	return analyzer
}

// New constructs an analyzer after validating its settings.
func New(settings Settings) (*analysis.Analyzer, error) {
	rules, err := enabledRules(settings.Rules)
	if err != nil {
		return nil, err
	}
	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  "checks SQL queries embedded in Go source",
		Run: func(pass *analysis.Pass) (any, error) {
			run(pass, rules)
			return nil, nil
		},
		RunDespiteErrors: true,
	}, nil
}

type diagnostic struct {
	pos     token.Pos
	message string
}

type queryCall struct {
	call         *ast.CallExpr
	method       string
	pos          token.Pos
	receiverType string
	sql          string
}

type fileContext struct {
	file    *ast.File
	parents map[ast.Node]ast.Node
	pass    *analysis.Pass
	queries []queryCall
}

func run(pass *analysis.Pass, rules []rule) {
	for _, file := range pass.Files {
		context := inspectFile(pass, file)
		for _, currentRule := range rules {
			for _, issue := range currentRule.Check(context) {
				pass.Reportf(issue.pos, "%s: %s", currentRule.ID(), issue.message)
			}
		}
	}
}

func inspectFile(pass *analysis.Pass, file *ast.File) fileContext {
	context := fileContext{file: file, parents: make(map[ast.Node]ast.Node), pass: pass}
	var stack []ast.Node
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			context.parents[node] = stack[len(stack)-1]
		}
		stack = append(stack, node)
		if call, ok := node.(*ast.CallExpr); ok {
			if query, ok := extractQuery(pass, call); ok {
				context.queries = append(context.queries, query)
			}
		}
		return true
	})
	return context
}

func extractQuery(pass *analysis.Pass, call *ast.CallExpr) (queryCall, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return queryCall{}, false
	}
	receiverType, ok := databaseSQLReceiver(pass.TypesInfo.Selections[selector])
	if !ok {
		return queryCall{}, false
	}
	argument, ok := queryArgument(selector.Sel.Name)
	if !ok || argument >= len(call.Args) {
		return queryCall{}, false
	}
	expression := call.Args[argument]
	value := pass.TypesInfo.Types[expression].Value
	if value == nil || value.Kind() != constant.String {
		return queryCall{}, false
	}
	return queryCall{call: call, method: selector.Sel.Name, pos: expression.Pos(), receiverType: receiverType, sql: constant.StringVal(value)}, true
}

func databaseSQLReceiver(selection *types.Selection) (string, bool) {
	if selection == nil || selection.Obj().Pkg() == nil || selection.Obj().Pkg().Path() != "database/sql" {
		return "", false
	}
	receiver := selection.Recv()
	if pointer, ok := receiver.(*types.Pointer); ok {
		receiver = pointer.Elem()
	}
	named, ok := receiver.(*types.Named)
	if !ok || named.Obj().Pkg() == nil || named.Obj().Pkg().Path() != "database/sql" {
		return "", false
	}
	name := named.Obj().Name()
	return name, slices.Contains([]string{"DB", "Tx", "Conn"}, name)
}

func queryArgument(method string) (int, bool) {
	switch method {
	case "Exec", "Prepare", "Query", "QueryRow":
		return 0, true
	case "ExecContext", "PrepareContext", "QueryContext", "QueryRowContext":
		return 1, true
	default:
		return 0, false
	}
}

func ancestor[T ast.Node](context fileContext, node ast.Node) T {
	for parent := context.parents[node]; parent != nil; parent = context.parents[parent] {
		if match, ok := parent.(T); ok {
			return match
		}
	}
	var zero T
	return zero
}

func enabledRules(names []string) ([]rule, error) {
	available := allRules()
	if len(names) == 0 {
		return available, nil
	}
	byID := make(map[string]rule, len(available))
	for _, currentRule := range available {
		byID[currentRule.ID()] = currentRule
	}
	rules := make([]rule, 0, len(names))
	for _, name := range names {
		currentRule, ok := byID[name]
		if !ok {
			return nil, fmt.Errorf("unknown sqlvet rule %q", name)
		}
		rules = append(rules, currentRule)
	}
	return rules, nil
}
