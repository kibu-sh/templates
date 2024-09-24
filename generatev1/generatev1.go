package main

import (
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"golang.org/x/tools/go/packages"
)

func main() {
	// Change to the directory containing the target Go program
	pkgs, err := decorator.Load(&packages.Config{
		Tests: false,
		Dir:   "/Users/jqualls/projects/github.com/kibu-sh/templates/starter/src/backend/systems/billingv1",
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedModule |
			packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedDeps |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}, ".")
	checkErrFatal(err)

	analyzer := new(Analyzer)
	pkg := pkgs[0]
	for _, file := range pkg.Syntax {
		dst.Walk(analyzer, file)
	}
}

var _ dst.Visitor = (*Analyzer)(nil)

type Analyzer struct {
}

func (a *Analyzer) Visit(node dst.Node) dst.Visitor {
	iface, ok := node.(*dst.InterfaceType)
	if !ok {
		return a
	}

	_ = iface
	return a
}
