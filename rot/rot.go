package rot

import (
	"flag"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "rot",
		Doc:   "Makes sure that a variable is defined right before it is used.",
		Run:   run,
		Flags: flag.FlagSet{},
	}
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			analyzeFunction(pass, fn)
		}
	}
	return nil, nil
}

type blockCtx struct {
	parent *blockCtx
	owner  ast.Stmt
	stmts  [][]ast.Stmt
}

type stmtInfo struct {
	block *blockCtx
	index int
}

type declInfo struct {
	name  string
	pos   token.Pos
	block *blockCtx
	index int
}

func (ctx *blockCtx) stmtAt(idx int) ast.Stmt {
	if ctx == nil || idx < 0 {
		return nil
	}
	if idx >= len(ctx.stmts) {
		return nil
	}
	if len(ctx.stmts[idx]) == 0 {
		return nil
	}
	return ctx.stmts[idx][0]
}

func (ctx *blockCtx) stmtsInRange(start, end int) []ast.Stmt {
	if ctx == nil {
		return nil
	}
	if end <= start+1 {
		return nil
	}
	if end > len(ctx.stmts) {
		return nil
	}
	var out []ast.Stmt
	for i := start + 1; i < end; i++ {
		out = append(out, ctx.stmts[i]...)
	}
	return out
}

type analyzer struct {
	pass       *analysis.Pass
	parents    map[ast.Node]ast.Node
	stmtInfo   map[ast.Stmt]*stmtInfo
	synthInfo  map[ast.Stmt]*stmtInfo
	decls      map[types.Object]*declInfo
	seen       map[types.Object]bool
	violations map[types.Object]bool
	forPost    map[ast.Stmt]struct{}
	caseBlocks map[*ast.CaseClause]*blockCtx
}

func analyzeFunction(pass *analysis.Pass, fn *ast.FuncDecl) {
	parents := buildParents(fn.Body)
	builder := newContextBuilder()
	builder.buildBlock(fn.Body, nil, nil)

	a := &analyzer{
		pass:       pass,
		parents:    parents,
		stmtInfo:   builder.stmtInfo,
		synthInfo:  builder.synthInfo,
		decls:      make(map[types.Object]*declInfo),
		seen:       make(map[types.Object]bool),
		violations: make(map[types.Object]bool),
		forPost:    builder.forPost,
		caseBlocks: builder.caseBlocks,
	}

	a.collectDecls(fn.Body)
	if len(a.decls) == 0 {
		return
	}
	a.inspectUses(fn.Body)

	for obj, decl := range a.decls {
		if !a.violations[obj] {
			continue
		}
		pass.Reportf(decl.pos, "variable %s should be declared right before it is used", decl.name)
	}
}

func (a *analyzer) collectDecls(node ast.Node) {
	ast.Inspect(node, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.DeclStmt:
			gen, ok := stmt.Decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				return true
			}
			for _, spec := range gen.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					a.recordDecl(name, stmt)
				}
			}
		case *ast.AssignStmt:
			if stmt.Tok != token.DEFINE {
				return true
			}
			for _, expr := range stmt.Lhs {
				ident, ok := expr.(*ast.Ident)
				if !ok {
					continue
				}
				a.recordDeclFromAssign(ident, stmt)
			}
		case *ast.RangeStmt:
			if stmt.Tok != token.DEFINE {
				return true
			}
			if ident, ok := stmt.Key.(*ast.Ident); ok {
				a.recordDecl(ident, stmt)
			}
			if ident, ok := stmt.Value.(*ast.Ident); ok {
				a.recordDecl(ident, stmt)
			}
		case *ast.CaseClause:
			if obj := a.pass.TypesInfo.Implicits[stmt]; obj != nil {
				v, ok := obj.(*types.Var)
				if !ok {
					return true
				}
				block := a.caseBlocks[stmt]
				if block == nil {
					return true
				}
				a.decls[v] = &declInfo{
					name:  v.Name(),
					pos:   v.Pos(),
					block: block,
					index: -1,
				}
			}
		}
		return true
	})
}

func (a *analyzer) recordDeclWithObject(ident *ast.Ident, stmt ast.Stmt, obj types.Object) {
	if ident == nil || ident.Name == "_" {
		return
	}
	if obj == nil {
		return
	}
	if _, ok := obj.(*types.Var); !ok {
		return
	}
	info := a.contextInfo(stmt)
	if info == nil || info.block == nil {
		return
	}
	index := info.index
	if clause, ok := info.block.owner.(*ast.CommClause); ok && clause.Comm == stmt {
		index--
	}
	a.decls[obj] = &declInfo{
		name:  ident.Name,
		pos:   ident.Pos(),
		block: info.block,
		index: index,
	}
}

func (a *analyzer) recordDecl(ident *ast.Ident, stmt ast.Stmt) {
	obj := a.pass.TypesInfo.Defs[ident]
	a.recordDeclWithObject(ident, stmt, obj)
}

func (a *analyzer) recordDeclFromAssign(ident *ast.Ident, stmt *ast.AssignStmt) {
	if ident == nil || ident.Name == "_" {
		return
	}
	obj := a.pass.TypesInfo.Defs[ident]
	if obj == nil {
		if !a.isTypeSwitchAssign(stmt) {
			return
		}
		obj = a.pass.TypesInfo.ObjectOf(ident)
	}
	a.recordDeclWithObject(ident, stmt, obj)
}

func (a *analyzer) isTypeSwitchAssign(assign *ast.AssignStmt) bool {
	parent, ok := a.parents[assign]
	if !ok {
		return false
	}
	ts, ok := parent.(*ast.TypeSwitchStmt)
	if !ok {
		return false
	}
	return ts.Assign == assign
}

func (a *analyzer) inspectUses(node ast.Node) {
	ast.Inspect(node, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if ident.Name == "_" {
			return true
		}
		if def := a.pass.TypesInfo.Defs[ident]; def != nil {
			return true
		}
		obj := a.pass.TypesInfo.ObjectOf(ident)
		if obj == nil {
			return true
		}
		decl, ok := a.decls[obj]
		if !ok {
			return true
		}
		if a.seen[obj] {
			return true
		}
		stmt, info := a.enclosingStmt(ident)
		if stmt == nil || info == nil {
			a.seen[obj] = true
			return true
		}
		if _, skip := a.forPost[stmt]; skip {
			return true
		}
		if !a.pathOK(decl, stmt, info) {
			a.violations[obj] = true
		}
		a.seen[obj] = true
		return true
	})
}

func (a *analyzer) enclosingStmt(node ast.Node) (ast.Stmt, *stmtInfo) {
	for n := node; n != nil; n = a.parents[n] {
		if stmt, ok := n.(ast.Stmt); ok {
			if info := a.contextInfo(stmt); info != nil {
				return stmt, info
			}
		}
	}
	return nil, nil
}

func (a *analyzer) pathOK(decl *declInfo, stmt ast.Stmt, info *stmtInfo) bool {
	block := info.block
	idx := info.index
	for {
		if block == nil {
			return false
		}
		if block == decl.block {
			if idx <= decl.index+1 {
				return true
			}
			if a.onlyDeclarationsOrErrorChecksBetween(decl, block, decl.index, idx) {
				return true
			}
			return false
		}
		if idx != 0 && !a.allowInnerBlockGap(decl, block) {
			return false
		}
		owner := block.owner
		if owner == nil {
			block = block.parent
			continue
		}
		ownerInfo := a.contextInfo(owner)
		if ownerInfo == nil {
			return false
		}
		block = ownerInfo.block
		idx = ownerInfo.index
	}
}

func (a *analyzer) allowInnerBlockGap(decl *declInfo, block *blockCtx) bool {
	if decl == nil || decl.block == nil || block == nil {
		return false
	}
	owner := block.owner
	if owner == nil {
		return false
	}
	switch owner.(type) {
	case *ast.ForStmt, *ast.RangeStmt:
		declStmt := decl.block.stmtAt(decl.index)
		if declStmt == nil {
			return false
		}
		return declStmt != owner
	default:
		return false
	}
}

func (a *analyzer) onlyDeclarationsOrErrorChecksBetween(decl *declInfo, block *blockCtx, from, to int) bool {
	if block == nil {
		return false
	}
	if to <= from+1 {
		return true
	}
	if to > len(block.stmts) {
		return false
	}
	for i := from + 1; i < to; i++ {
		for _, stmt := range block.stmts[i] {
			if !isDeclarationOrEmpty(stmt) && !a.isErrorCheck(stmt, decl) {
				return false
			}
		}
	}
	return true
}

func (a *analyzer) isErrorCheck(stmt ast.Stmt, decl *declInfo) bool {
	// Get the declaration statement for the variable we're tracking
	declStmt := decl.block.stmtAt(decl.index)
	if declStmt == nil {
		return false
	}

	// Check if the declaration statement is an assignment that also declares an error variable
	var errObj types.Object
	assignStmt, ok := declStmt.(*ast.AssignStmt)
	if !ok || assignStmt.Tok != token.DEFINE {
		return false
	}

	// Find the error variable declared in the same assignment
	for _, expr := range assignStmt.Lhs {
		ident, ok := expr.(*ast.Ident)
		if !ok {
			continue
		}
		if ident.Name == "err" {
			obj := a.pass.TypesInfo.Defs[ident]
			if obj != nil {
				errObj = obj
				break
			}
		}
	}

	if errObj == nil {
		return false
	}

	// Check if the statement is validating this error
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		// Check for function calls like require.NoError(t, err), assert.NoError(t, err)
		call, ok := s.X.(*ast.CallExpr)
		if !ok {
			return false
		}
		return a.isErrorCheckCall(call, errObj)
	case *ast.IfStmt:
		// Check for if err != nil { ... }
		return a.isErrorCheckIf(s, errObj)
	default:
		return false
	}
}

func (a *analyzer) isErrorCheckCall(call *ast.CallExpr, errObj types.Object) bool {
	if len(call.Args) == 0 {
		return false
	}

	// Check if the last argument is the error variable we're tracking
	lastArg := call.Args[len(call.Args)-1]
	ident, ok := lastArg.(*ast.Ident)
	if !ok {
		return false
	}

	obj := a.pass.TypesInfo.ObjectOf(ident)
	if obj != errObj {
		return false
	}

	// Check if it's a known error-checking function
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	selIdent := sel.Sel

	// Common error-checking function names
	errorCheckNames := map[string]bool{
		"NoError":      true,
		"Error":        true,
		"ErrorIs":      true,
		"ErrorAs":      true,
		"Nil":          true,
		"NotNil":       true,
		"Fatal":        true,
		"Fatalf":       true,
		"FailNow":      true,
		"Must":         true,
		"MustNotError": true,
	}

	return errorCheckNames[selIdent.Name]
}

func (a *analyzer) isErrorCheckIf(ifStmt *ast.IfStmt, errObj types.Object) bool {
	// Check if the condition is checking err != nil or err == nil
	binary, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return false
	}

	// Check for err != nil or err == nil patterns
	if binary.Op != token.NEQ && binary.Op != token.EQL {
		return false
	}

	// Check if one side is nil and the other is the error variable
	// nil is a predeclared identifier, so we check by name
	if ident, ok := binary.X.(*ast.Ident); ok {
		obj := a.pass.TypesInfo.ObjectOf(ident)
		if obj == errObj {
			// Check if the other side is nil
			if identY, ok := binary.Y.(*ast.Ident); ok && identY.Name == "nil" {
				return true
			}
		}
	}

	if ident, ok := binary.Y.(*ast.Ident); ok {
		obj := a.pass.TypesInfo.ObjectOf(ident)
		if obj == errObj {
			// Check if the other side is nil
			if identX, ok := binary.X.(*ast.Ident); ok && identX.Name == "nil" {
				return true
			}
		}
	}

	return false
}

func onlyDeclarationsBetween(block *blockCtx, from, to int) bool {
	if block == nil {
		return false
	}
	if to <= from+1 {
		return true
	}
	if to > len(block.stmts) {
		return false
	}
	for i := from + 1; i < to; i++ {
		for _, stmt := range block.stmts[i] {
			if !isDeclarationOrEmpty(stmt) {
				return false
			}
		}
	}
	return true
}

func isDeclarationOrEmpty(stmt ast.Stmt) bool {
	switch s := stmt.(type) {
	case *ast.EmptyStmt:
		return true
	case *ast.DeclStmt:
		return true
	case *ast.AssignStmt:
		return s.Tok == token.DEFINE
	default:
		return false
	}
}

func (a *analyzer) contextInfo(stmt ast.Stmt) *stmtInfo {
	if info, ok := a.stmtInfo[stmt]; ok {
		return info
	}
	return a.synthInfo[stmt]
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

type contextBuilder struct {
	stmtInfo   map[ast.Stmt]*stmtInfo
	synthInfo  map[ast.Stmt]*stmtInfo
	forPost    map[ast.Stmt]struct{}
	caseBlocks map[*ast.CaseClause]*blockCtx
}

func newContextBuilder() *contextBuilder {
	return &contextBuilder{
		stmtInfo:   make(map[ast.Stmt]*stmtInfo),
		synthInfo:  make(map[ast.Stmt]*stmtInfo),
		forPost:    make(map[ast.Stmt]struct{}),
		caseBlocks: make(map[*ast.CaseClause]*blockCtx),
	}
}

func (cb *contextBuilder) infoFor(stmt ast.Stmt) *stmtInfo {
	if stmt == nil {
		return nil
	}
	if info := cb.stmtInfo[stmt]; info != nil {
		return info
	}
	return cb.synthInfo[stmt]
}

func (cb *contextBuilder) buildBlock(block *ast.BlockStmt, parent *blockCtx, owner ast.Stmt) *blockCtx {
	if block == nil {
		return parent
	}
	ctx := &blockCtx{
		parent: parent,
		owner:  owner,
		stmts:  make([][]ast.Stmt, len(block.List)),
	}
	for i, stmt := range block.List {
		ctx.stmts[i] = append(ctx.stmts[i], stmt)
		cb.stmtInfo[stmt] = &stmtInfo{block: ctx, index: i}
		cb.processStmt(stmt, ctx)
	}
	return ctx
}

func (cb *contextBuilder) processStmt(stmt ast.Stmt, ctx *blockCtx) {
	switch s := stmt.(type) {
	case *ast.LabeledStmt:
		cb.stmtInfo[s.Stmt] = cb.infoFor(stmt)
		cb.processStmt(s.Stmt, ctx)
	case *ast.BlockStmt:
		cb.buildBlock(s, ctx, stmt)
	case *ast.IfStmt:
		if s.Init != nil {
			if owner := cb.infoFor(stmt); owner != nil {
				cb.addSynthetic(s.Init, owner.block, owner.index-1)
			}
		}
		cb.buildBlock(s.Body, ctx, stmt)
		cb.buildElse(s.Else, ctx, stmt)
	case *ast.ForStmt:
		if s.Init != nil {
			if owner := cb.infoFor(stmt); owner != nil {
				cb.addSynthetic(s.Init, owner.block, owner.index-1)
			}
		}
		if s.Post != nil {
			if owner := cb.infoFor(stmt); owner != nil {
				cb.addSynthetic(s.Post, owner.block, owner.index)
				cb.forPost[s.Post] = struct{}{}
			}
		}
		cb.buildBlock(s.Body, ctx, stmt)
	case *ast.RangeStmt:
		cb.buildBlock(s.Body, ctx, stmt)
	case *ast.SwitchStmt:
		if s.Init != nil {
			if owner := cb.infoFor(stmt); owner != nil {
				cb.addSynthetic(s.Init, owner.block, owner.index-1)
			}
		}
		cb.buildBlock(s.Body, ctx, stmt)
	case *ast.TypeSwitchStmt:
		if s.Init != nil {
			if owner := cb.infoFor(stmt); owner != nil {
				cb.addSynthetic(s.Init, owner.block, owner.index-1)
			}
		}
		if s.Assign != nil {
			if owner := cb.infoFor(stmt); owner != nil {
				cb.addSynthetic(s.Assign, owner.block, owner.index-1)
			}
		}
		cb.buildBlock(s.Body, ctx, stmt)
	case *ast.SelectStmt:
		cb.buildBlock(s.Body, ctx, stmt)
	case *ast.CaseClause:
		cb.buildClauseBody(s, ctx)
	case *ast.CommClause:
		cb.buildCommClause(s, ctx)
	}
}

func (cb *contextBuilder) buildElse(stmt ast.Stmt, ctx *blockCtx, owner ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		cb.buildBlock(s, ctx, owner)
	case *ast.IfStmt:
		info := cb.infoFor(owner)
		if info != nil {
			cb.synthInfo[s] = &stmtInfo{block: info.block, index: info.index}
		}
		cb.processStmt(s, ctx)
	}
}

func (cb *contextBuilder) buildClauseBody(clause *ast.CaseClause, parent *blockCtx) {
	ctx := &blockCtx{
		parent: parent,
		owner:  clause,
		stmts:  make([][]ast.Stmt, len(clause.Body)),
	}
	cb.caseBlocks[clause] = ctx
	for i, stmt := range clause.Body {
		ctx.stmts[i] = append(ctx.stmts[i], stmt)
		cb.stmtInfo[stmt] = &stmtInfo{block: ctx, index: i}
		cb.processStmt(stmt, ctx)
	}
}

func (cb *contextBuilder) buildCommClause(clause *ast.CommClause, parent *blockCtx) {
	size := len(clause.Body)
	if clause.Comm != nil && size == 0 {
		size = 1
	}
	ctx := &blockCtx{
		parent: parent,
		owner:  clause,
		stmts:  make([][]ast.Stmt, size),
	}
	idx := 0
	if clause.Comm != nil {
		if idx < len(ctx.stmts) {
			ctx.stmts[idx] = append(ctx.stmts[idx], clause.Comm)
		}
		cb.stmtInfo[clause.Comm] = &stmtInfo{block: ctx, index: idx}
		cb.processStmt(clause.Comm, ctx)
	}
	for _, stmt := range clause.Body {
		if idx >= len(ctx.stmts) {
			ctx.stmts = append(ctx.stmts, nil)
		}
		ctx.stmts[idx] = append(ctx.stmts[idx], stmt)
		cb.stmtInfo[stmt] = &stmtInfo{block: ctx, index: idx}
		cb.processStmt(stmt, ctx)
		idx++
	}
}

func (cb *contextBuilder) addSynthetic(stmt ast.Stmt, block *blockCtx, index int) {
	if block == nil {
		return
	}
	cb.synthInfo[stmt] = &stmtInfo{block: block, index: index}
	cb.processStmt(stmt, block)
}
