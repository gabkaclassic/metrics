package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var InterruptCheckAnalyzer = &analysis.Analyzer{
	Name: "interruptCheck",
	Doc:  "check using panic or exit function, fatal logs",
	Run:  runAnalyzer,
}

func runAnalyzer(pass *analysis.Pass) (any, error) {

	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				if _, ok := pass.TypesInfo.Uses[fun].(*types.Builtin); ok && fun.Name == "panic" {
					pass.Reportf(fun.Pos(), "using builtin panic function")
					return true
				}
			case *ast.SelectorExpr:
				if pass.Pkg.Name() == "main" {
					return true
				}

				function, ok := pass.TypesInfo.Uses[fun.Sel].(*types.Func)
				if !ok {
					return true
				}

				if function.Name() == "Exit" && function.Pkg().Name() == "os" {
					pass.Reportf(fun.Sel.Pos(), "call exit function is not in main package")
					return true
				}
				if function.Name() == "Fatal" && function.Pkg().Name() == "log" {
					pass.Reportf(fun.Sel.Pos(), "fatal log is not in main package")
					return true
				}
			}

			return true
		})
	}

	return nil, nil
}
