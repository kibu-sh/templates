package kibusvc

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var _ analysis.Fact = (*Module)(nil)

type Module struct {
	Name     string
	Services []*Service
}

func (m *Module) AFact() {}

type Operations struct {
	Name string
}

type Service struct {
	Operations []*Operations
}

var Analyzer = &analysis.Analyzer{
	Name:             "kibusvc",
	Doc:              "Analyzes go source code for kibu service definitions",
	RunDespiteErrors: true,
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	Run:              run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	walk := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}

	walk.Preorder(nodeFilter, func(n ast.Node) {
		decl := n.(*ast.GenDecl)
		_ = decl
		print()
	})
	return nil, nil
}
