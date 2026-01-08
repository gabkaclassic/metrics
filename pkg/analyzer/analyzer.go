// Package analyzer provides static analysis checks for detecting
// unsafe process-terminating constructs in Go code.
//
// The analyzer reports:
//   - usage of the builtin panic function
//   - calls to os.Exit outside the main package
//   - calls to log.Fatal outside the main package
//
// Intended to be used as a SAST rule to discourage abrupt termination
// of program execution in library and non-entrypoint code.
package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// InterruptCheckAnalyzer reports usages of panic, os.Exit, and log.Fatal
// according to the following rules:
//
//   - builtin panic is always reported
//   - os.Exit is reported if used outside package main
//   - log.Fatal is reported if used outside package main
//
// The analyzer relies on go/types information to accurately distinguish
// builtin functions and imported package symbols.
var InterruptCheckAnalyzer = &analysis.Analyzer{

	Name: "interruptCheck",
	Doc:  "check using panic or exit function, fatal logs",
	Run:  runAnalyzer,
}

// runAnalyzer walks through all AST files in the analyzed package,
// inspects function call expressions, and reports diagnostics
// when forbidden interrupting constructs are detected.
//
// It uses type information from analysis.Pass to:
//   - identify builtin functions (panic)
//   - resolve selector expressions to concrete package functions
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
