// Package plugin registers sqlvet with golangci-lint's module plugin system.
package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/stuft2/sqlvet"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("sqlvet", newPlugin)
}

type sqlvetPlugin struct {
	settings sqlvet.Settings
}

func newPlugin(rawSettings any) (register.LinterPlugin, error) {
	settings, err := register.DecodeSettings[sqlvet.Settings](rawSettings)
	if err != nil {
		return nil, err
	}

	if _, err := sqlvet.New(settings); err != nil {
		return nil, err
	}

	return &sqlvetPlugin{settings: settings}, nil
}

func (p *sqlvetPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	analyzer, err := sqlvet.New(p.settings)
	if err != nil {
		return nil, err
	}

	return []*analysis.Analyzer{analyzer}, nil
}

func (*sqlvetPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
