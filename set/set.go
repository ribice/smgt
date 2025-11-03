package set

import (
	"flag"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "set",
		Doc:   "Detects map[string]bool values that are only assigned the constant true and recommends map[string]struct{} instead.",
		Run:   run,
		Flags: flag.FlagSet{},
	}
}

type mapUsage struct {
	onlyTrue bool
	sawWrite bool
}

func run(pass *analysis.Pass) (any, error) {
	usages := collectCandidates(pass)
	if len(usages) == 0 {
		return nil, nil
	}

	a := &analyzer{
		pass:   pass,
		usages: usages,
	}

	for _, file := range pass.Files {
		ast.Inspect(file, a.inspect)
	}

	for obj, usage := range usages {
		if usage.onlyTrue && usage.sawWrite {
			pass.Reportf(obj.Pos(), "map[string]bool variable %s is used as a set; use map[string]struct{} instead", obj.Name())
		}
	}

	return nil, nil
}

type analyzer struct {
	pass   *analysis.Pass
	usages map[types.Object]*mapUsage
}

func collectCandidates(pass *analysis.Pass) map[types.Object]*mapUsage {
	usages := make(map[types.Object]*mapUsage)
	for ident, obj := range pass.TypesInfo.Defs {
		if obj == nil {
			continue
		}
		v, ok := obj.(*types.Var)
		if !ok {
			continue
		}
		if pass.Pkg != nil && v.Pkg() != pass.Pkg {
			continue
		}
		if !isStringBoolMap(v.Type()) {
			continue
		}
		if !identIsValid(ident) {
			continue
		}
		usages[obj] = &mapUsage{
			onlyTrue: true,
		}
	}
	return usages
}

func (a *analyzer) inspect(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.AssignStmt:
		a.handleAssignStmt(n)
	case *ast.RangeStmt:
		// no-op; included to avoid extra allocations in default branch
	case *ast.ValueSpec:
		a.handleValueSpec(n)
	}
	return true
}

func (a *analyzer) handleAssignStmt(stmt *ast.AssignStmt) {
	// Handle assignments to map indexes.
	for i, lhs := range stmt.Lhs {
		index, ok := lhs.(*ast.IndexExpr)
		if !ok {
			continue
		}
		target := a.mapObject(index.X)
		if target == nil {
			continue
		}
		usage := a.usages[target]
		if usage == nil {
			continue
		}
		if stmt.Tok == token.DEFINE {
			// map indexes cannot appear on the left-hand side of :=
			continue
		}
		value := a.valueForIndex(stmt, i)
		if value == nil {
			usage.onlyTrue = false
			continue
		}
		usage.sawWrite = true
		if !a.isConstTrue(value) {
			usage.onlyTrue = false
		}
	}

	// Handle direct assignments of composite literals.
	for i, lhs := range stmt.Lhs {
		obj := a.objectForAssignLHS(lhs)
		if obj == nil {
			continue
		}
		usage := a.usages[obj]
		if usage == nil {
			continue
		}
		value := a.valueForIndex(stmt, i)
		if value == nil {
			continue
		}
		a.handleAssignedValue(usage, value)
	}
}

func (a *analyzer) handleValueSpec(spec *ast.ValueSpec) {
	for i, name := range spec.Names {
		obj := a.pass.TypesInfo.Defs[name]
		if obj == nil {
			continue
		}
		usage := a.usages[obj]
		if usage == nil {
			continue
		}
		if len(spec.Values) == 0 {
			continue
		}
		value := spec.Values[0]
		if len(spec.Values) == len(spec.Names) {
			value = spec.Values[i]
		}
		a.handleAssignedValue(usage, value)
	}
}

func (a *analyzer) handleAssignedValue(usage *mapUsage, expr ast.Expr) {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return
	}
	if !isStringBoolMap(a.pass.TypesInfo.TypeOf(lit)) {
		return
	}
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		usage.sawWrite = true
		if !a.isConstTrue(kv.Value) {
			usage.onlyTrue = false
		}
	}
}

func (a *analyzer) valueForIndex(stmt *ast.AssignStmt, idx int) ast.Expr {
	if len(stmt.Rhs) == 0 {
		return nil
	}
	if len(stmt.Rhs) == 1 {
		return stmt.Rhs[0]
	}
	if idx < len(stmt.Rhs) {
		return stmt.Rhs[idx]
	}
	return nil
}

func (a *analyzer) objectForAssignLHS(expr ast.Expr) types.Object {
	switch e := expr.(type) {
	case *ast.Ident:
		if obj := a.pass.TypesInfo.Defs[e]; obj != nil {
			return obj
		}
		return a.pass.TypesInfo.Uses[e]
	case *ast.SelectorExpr:
		if obj := a.pass.TypesInfo.Uses[e.Sel]; obj != nil {
			return obj
		}
		if sel := a.pass.TypesInfo.Selections[e]; sel != nil {
			return sel.Obj()
		}
	}
	return nil
}

func (a *analyzer) mapObject(expr ast.Expr) types.Object {
	switch e := expr.(type) {
	case *ast.Ident:
		if obj := a.pass.TypesInfo.Uses[e]; obj != nil {
			return obj
		}
		return a.pass.TypesInfo.Defs[e]
	case *ast.SelectorExpr:
		if obj := a.pass.TypesInfo.Uses[e.Sel]; obj != nil {
			return obj
		}
		if sel := a.pass.TypesInfo.Selections[e]; sel != nil {
			return sel.Obj()
		}
	}
	return nil
}

func (a *analyzer) isConstTrue(expr ast.Expr) bool {
	tv, ok := a.pass.TypesInfo.Types[expr]
	if !ok {
		return false
	}
	if tv.Value == nil {
		return false
	}
	if tv.Value.Kind() != constant.Bool {
		return false
	}
	return tv.Value.String() == "true"
}

func isStringBoolMap(typ types.Type) bool {
	if typ == nil {
		return false
	}
	t := typ.Underlying()
	m, ok := t.(*types.Map)
	if !ok {
		return false
	}
	key, ok := m.Key().Underlying().(*types.Basic)
	if !ok || key.Kind() != types.String {
		return false
	}
	elem, ok := m.Elem().Underlying().(*types.Basic)
	if !ok || elem.Kind() != types.Bool {
		return false
	}
	return true
}

func identIsValid(ident *ast.Ident) bool {
	if ident == nil {
		return false
	}
	if ident.Name == "_" {
		return false
	}
	return ident.NamePos.IsValid()
}
