package main

import (
	"github.com/stuft2/sqlvet"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(sqlvet.NewAnalyzer(sqlvet.Settings{}))
}
