package projectlint

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// IntNameAnalyzer enforces the project rule that variables with the built-in
// int type start with a capital I.
var IntNameAnalyzer *analysis.Analyzer = &analysis.Analyzer{
	Name: "intname",
	Doc:  "requires built-in int variables to start with a capital I",
	Run:  runIntName,
}

// ExplicitInitAnalyzer enforces explicit types and initializers for var declarations.
var ExplicitInitAnalyzer *analysis.Analyzer = &analysis.Analyzer{
	Name: "explicitinit",
	Doc:  "requires var declarations to explicitly declare a type and initialize variables",
	Run:  runExplicitInit,
}

func runIntName(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if !isSourceFile(pass, file) {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.FuncDecl:
				checkFuncType(pass, n.Type)
			case *ast.FuncLit:
				checkFuncType(pass, n.Type)
			case *ast.ValueSpec:
				for _, name := range n.Names {
					checkIntName(pass, name, pass.TypesInfo.Defs[name])
				}
			case *ast.AssignStmt:
				if n.Tok != token.DEFINE {
					return true
				}
				for _, expr := range n.Lhs {
					name, ok := expr.(*ast.Ident)
					if !ok {
						continue
					}
					checkIntName(pass, name, pass.TypesInfo.Defs[name])
				}
			case *ast.RangeStmt:
				if n.Tok != token.DEFINE {
					return true
				}
				checkRangeName(pass, n.Key)
				checkRangeName(pass, n.Value)
			}

			return true
		})
	}

	return nil, nil
}

func checkFuncType(pass *analysis.Pass, funcType *ast.FuncType) {
	if funcType == nil {
		return
	}
	checkFieldList(pass, funcType.Params)
	checkFieldList(pass, funcType.Results)
}

func checkFieldList(pass *analysis.Pass, fields *ast.FieldList) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		for _, name := range field.Names {
			checkIntName(pass, name, pass.TypesInfo.Defs[name])
		}
	}
}

func checkRangeName(pass *analysis.Pass, expr ast.Expr) {
	name, ok := expr.(*ast.Ident)
	if !ok {
		return
	}
	checkIntName(pass, name, pass.TypesInfo.Defs[name])
}

func checkIntName(pass *analysis.Pass, name *ast.Ident, obj types.Object) {
	variable, ok := obj.(*types.Var)
	if !ok || !types.Identical(variable.Type(), types.Typ[types.Int]) {
		return
	}

	if name == nil || name.Name == "_" || strings.HasPrefix(name.Name, "I") {
		return
	}

	pass.Reportf(name.Pos(), "int variable %q must start with I", name.Name)
}

func runExplicitInit(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if !isSourceFile(pass, file) {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			decl, ok := node.(*ast.GenDecl)
			if !ok || decl.Tok != token.VAR {
				return true
			}

			for _, spec := range decl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					if name.Name == "_" {
						continue
					}
					if valueSpec.Type == nil {
						pass.Reportf(name.Pos(), "variable %q must declare an explicit type", name.Name)
					}
					if len(valueSpec.Values) == 0 {
						pass.Reportf(name.Pos(), "variable %q must be explicitly initialized", name.Name)
					}
				}
			}

			return false
		})
	}

	return nil, nil
}

func isSourceFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(pass.Fset.PositionFor(file.Pos(), false).Filename, ".go")
}
