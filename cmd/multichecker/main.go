package main

import (
	"github.com/renatus-cartesius/metricserv/pkg/analysis/passes/osexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"strings"
)

func main() {

	analyzers := make([]*analysis.Analyzer, 0)

	// standard golang.org/x/tools/go/analysis/passes analyzers
	analyzers = append(analyzers, printf.Analyzer)
	analyzers = append(analyzers, shadow.Analyzer)
	analyzers = append(analyzers, structtag.Analyzer)
	analyzers = append(analyzers, lostcancel.Analyzer)
	analyzers = append(analyzers, shift.Analyzer)
	analyzers = append(analyzers, sortslice.Analyzer)
	analyzers = append(analyzers, unmarshal.Analyzer)
	analyzers = append(analyzers, unusedwrite.Analyzer)
	analyzers = append(analyzers, unusedresult.Analyzer)

	// staticcheck analyzers with SA prefix
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// stylecheck analyzers with S1 prefix
	for _, v := range stylecheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "ST") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// custom analyzer
	analyzers = append(analyzers, osexit.Analyzer)

	multichecker.Main(
		analyzers...,
	)
}
