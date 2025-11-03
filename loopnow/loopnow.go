package loopnow

import (
	"flag"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "loopnow",
		Doc:   "Flags calls to time.Now() inside loops and suggests hoisting them to reduce system calls.",
		Run:   run,
		Flags: flag.FlagSet{},
	}
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		parents := buildParents(file)
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !isTimeNowCall(pass, call) {
				return true
			}
			if !inLoop(call, parents) {
				return true
			}
			pass.Reportf(call.Fun.Pos(), "time.Now should not be called inside loops; compute the value outside the loop")
			return true
		})
	}
	return nil, nil
}

func isTimeNowCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "Now" {
		return false
	}
	obj := pass.TypesInfo.Uses[sel.Sel]
	fn, ok := obj.(*types.Func)
	if !ok {
		return false
	}
	pkg := fn.Pkg()
	if pkg == nil {
		return false
	}
	if pkg.Path() != "time" {
		return false
	}
	return true
}

func inLoop(node ast.Node, parents map[ast.Node]ast.Node) bool {
	for parent := parents[node]; parent != nil; parent = parents[parent] {
		switch parent.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			return true
		}
	}
	return false
}

func buildParents(root ast.Node) map[ast.Node]ast.Node {
	parents := make(map[ast.Node]ast.Node)
	var stack []ast.Node
	ast.Inspect(root, func(n ast.Node) bool {
		if n != nil {
			if len(stack) > 0 {
				parents[n] = stack[len(stack)-1]
			}
			stack = append(stack, n)
			return true
		}
		stack = stack[:len(stack)-1]
		return false
	})
	return parents
}
