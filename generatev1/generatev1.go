package main

import (
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/pb33f/libopenapi/orderedmap"
	"golang.org/x/tools/go/packages"
	"os"
)

// Struct Definitions

type Package struct {
	Name string
}

type Type struct {
	Package *Package
	Name    string
}

type Decoration struct{}
type Comment struct{}

// Endpoint is anything tagged with one of
//
//	(kibu:service, kibu:activity, kibu:workflow)
type Endpoint struct {
	Package     *Package
	Name        string
	Type        *Type
	Comments    *orderedmap.Map[string, *Comment]
	Operations  *orderedmap.Map[string, *Operation]
	Decorations *orderedmap.Map[string, *Decoration]
}

// Operation is a method on a type or a raw function
// any public receiver or interface methods found on those tagged as endpoints
type Operation struct {
	Package  *Package
	Name     string
	Type     *Type
	Comments *orderedmap.Map[string, *Comment]
	Args     *orderedmap.Map[string, *Type]
	Results  *orderedmap.Map[string, *Type]
}

func checkErrFatal(err error) {
	if err != nil {
		panic(err)
	}
}

// Main Function

func main() {
	// Change to the directory containing the target Go program
	err := os.Chdir("/Users/jqualls/projects/github.com/kibu-sh/templates/starter/src/backend/systems/billingv1")
	checkErrFatal(err)

	pkgs, err := decorator.Load(&packages.Config{
		Dir:  "./",
		Mode: packages.LoadAllSyntax | packages.Ty,
	}, ".")
	checkErrFatal(err)

	var analyzer Analyzer
	pkg := pkgs[0]
	for _, file := range pkg.Syntax {
		dst.Walk(analyzer, file)
	}
}

var _ dst.Visitor = (*Analyzer)(nil)

type Analyzer struct {
}

func (a Analyzer) Visit(node dst.Node) dst.Visitor {
	iface, ok := node.(*dst.InterfaceType)
	if !ok {
		return a
	}

	_ = iface
	return a
}

// analyzeFile parses a DST file and extracts endpoints
//func analyzeFile(file *dst.File) ([]*Endpoint, error) {
//	var endpoints []*Endpoint
//	_ = &Package{Name: file.Name.Name}
//
//	for _, decl := range file.Decls {
//		genDecl, ok := decl.(*dst.GenDecl)
//		if !ok || genDecl.Tok != token.TYPE {
//			continue
//		}
//
//		for _, spec := range genDecl.Specs {
//			typeSpec, ok := spec.(*dst.TypeSpec)
//			if !ok {
//				continue
//			}
//
//			// Check if the type is an interface
//			iface, ok := typeSpec.Type.(*dst.InterfaceType)
//			if !ok {
//				continue
//			}
//
//			decorations := iface.Decorations()
//			_ = decorations
//		}
//	}
//
//	return endpoints, nil
//}
