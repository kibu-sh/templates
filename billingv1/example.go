package billingv1

type Package struct {
	// Path is the full file path on disk
	Path string

	// Name is the fully qualified name of the package
	Name string

	//Ref is a reference to the package
	Ref any
}

type Type struct {
	//Name is the name of the type
	Name string

	//Doc is the documentation for the type
	Doc string

	// points to the package reference
	Ref any
}

// Method represents a function its arguments and return type
type Method struct {
	Name      string
	Doc       string
	Arguments []Type
	Results   []Type
}

// Endpoint
// is a reference to go code that can be executed over an rpc
type Endpoint struct {
	Name string

	// a comment line attached to the receiver method //kibu:*
	Directives []any
	Package    Package
	Methods    []Method
}
