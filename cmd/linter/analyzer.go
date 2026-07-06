package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var PanicCheckAnalyzer = &analysis.Analyzer{
	Name: "panicCheck",
	Doc:  "check for unchecked errors",
	Run:  run,
}

func isPanic(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if ident.Name == "panic" {
			return true
		}
	}
	return false
}

func isLogFatalOrOsExit(call *ast.CallExpr) bool {
	if selExpr, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok {
			if ident.Name == "log" && selExpr.Sel.Name == "Fatal" {
				return true
			}
			if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
				return true
			}
		}
	}
	return false
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		currentPackage := pass.Pkg.Name()
		var currentFunc *ast.FuncDecl

		ast.Inspect(file, func(node ast.Node) bool {
			var exprStmt *ast.ExprStmt
			var call *ast.CallExpr
			var ok bool

			if funcDecl, ok := node.(*ast.FuncDecl); ok {
				currentFunc = funcDecl
			}

			if exprStmt, ok = node.(*ast.ExprStmt); !ok {
				return true
			}

			if call, ok = exprStmt.X.(*ast.CallExpr); !ok {
				return true
			}

			if isPanic(call) {
				pass.Reportf(exprStmt.Pos(), "используется встроенная функция panic")
			}

			if isLogFatalOrOsExit(call) && (currentPackage != "main" || currentFunc.Name.Name != "main") {
				pass.Reportf(
					exprStmt.Pos(),
					"используется log.Fatal и/или os.Exit вне функции main пакета main",
				)
			}
			return true
		})
	}
	return nil, nil
}
