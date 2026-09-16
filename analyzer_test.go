package sqlvet_test

import (
	"testing"

	"github.com/stuft2/sqlvet"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()
	tests := []struct {
		rule    string
		fixture string
	}{
		{rule: "SQL001", fixture: "rule001"},
		{rule: "SQL002", fixture: "rule002"},
		{rule: "SQL003", fixture: "rule003"},
		{rule: "SQL004", fixture: "rule004"},
		{rule: "SQL005", fixture: "rule005"},
		{rule: "SQL006", fixture: "rule006"},
		{rule: "SQL007", fixture: "rule007"},
		{rule: "SQL008", fixture: "rule008"},
		{rule: "SQL009", fixture: "rule009"},
		{rule: "SQL010", fixture: "rule010"},
	}

	for _, test := range tests {
		t.Run(test.rule, func(t *testing.T) {
			t.Parallel()
			settings := sqlvet.Settings{Rules: []string{test.rule}}
			analysistest.Run(t, testdata, sqlvet.NewAnalyzer(settings), test.fixture)
		})
	}
}
