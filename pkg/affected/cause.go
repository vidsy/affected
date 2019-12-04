package affected

import "github.com/vidsy/affected/pkg/module"

// Cause is why a package has been marked as affected
type Cause struct {
	Package    *module.Package   // The package that has modififcations
	ImportPath module.ImportPath // The import graph to that package
}
