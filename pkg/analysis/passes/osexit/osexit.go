package osexit

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "checking os.Exit() calls in the main function",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			if f, ok := node.(*ast.FuncDecl); ok && f.Name.Name == "main" {
				for _, st := range f.Body.List {
					switch x := st.(type) {
					case *ast.ExprStmt:
						call := x.X.(*ast.CallExpr).Fun.(*ast.SelectorExpr)
						if pkg, ok := call.X.(*ast.Ident); ok {
							packageName := pkg.Name
							funcName := call.Sel.Name
							if packageName == "os" && funcName == "Exit" {
								pass.Reportf(call.Pos(), "using os.Exit in main function")
							}
						}
					}
				}
			}
			return true
		})
	}

	return nil, nil
}
