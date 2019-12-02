package affected

import (
	"encoding/json"

	"github.com/vidsy/affected/module"
	"github.com/vidsy/affected/vcs"
	"github.com/vidsy/affected/vcs/git"
)

// Cause is why a package has been marked as affected
type Cause struct {
	Package *module.Package   // The package that has modififcations
	Imports []*module.Package // The import graph to that package
}

// Package represents a modified package
type Package struct {
	*module.Package

	cause []*module.Package
}

// Cause returns why a package is modified based on it's imports
func (p *Package) Cause() []Cause {
	cause := make([]Cause, 0)

	for _, c := range p.cause {
		imports := make([]*module.Package, 0)

		walk := func(cause *module.Package) func(i *module.Package) error {
			return func(i *module.Package) error {
				imports = append(imports, i)

				if i.ID == cause.ID {
					return module.ErrSkipPackage
				}

				return nil
			}
		}

		if err := module.Walk(p.Package, module.WalkImports, walk(c)); err != nil {
			return nil
		}

		cause = append(cause, Cause{
			Package: c,
			Imports: imports,
		})
	}

	return cause
}

// MarshalJSON marshals an Package effected package to JSON which will contain cause and imports
func (p *Package) MarshalJSON() ([]byte, error) {
	causes := make([]map[string]interface{}, 0, len(p.cause))

	for _, pkg := range p.Cause() {
		imports := make([]string, len(pkg.Imports))
		for i, imp := range pkg.Imports {
			imports[i] = imp.ID
		}

		causes = append(causes, map[string]interface{}{
			"package":   pkg.Package.ID,
			"directory": pkg.Package.Dir,
			"imports":   imports,
		})
	}

	return json.Marshal(map[string]interface{}{
		"package": p.ID,
		"causes":  causes,
	})
}

// PackagesOptions holds condifuration for loading pacakges and modified directories
type PackagesOptions struct {
	VCS vcs.ModifiedDirectoriesDetector // Verson control system to find mofified directories
	PL  module.PackageLoader            // Package loader
}

// Packages returns packages affected by two different commits either directly or indirectly
func Packages(mod, a, b string) ([]*Package, error) {
	g, err := git.New()
	if err != nil {
		return nil, err
	}

	opts := &PackagesOptions{
		VCS: g,
		PL:  module.DefaultPackageLoader(),
	}

	// Get a list of modified directories from VCS
	dirs, err := opts.VCS.ModifiedDirectories(a, b)
	if err != nil {
		return nil, err
	}

	// Load packages in this module
	pkgs, err := opts.PL.Load(mod)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*Package)

	for _, dir := range dirs {
		if modified := pkgs.Find(module.FindPackageByDir(dir)); modified != nil {
			err := module.Walk(modified, module.WalkParents, func(p *module.Package) error {
				if len(p.Parents) == 0 { // If this package is not imported by any other packages it is by definition a root node
					affected, ok := m[p.ID]
					if !ok {
						affected = &Package{
							Package: p,
						}

						m[p.ID] = affected
					}

					affected.cause = append(affected.cause, modified)

					return module.ErrSkipPackage
				}

				return nil
			})

			if err != nil {
				return nil, err
			}
		}
	}

	affected := make([]*Package, 0, len(m))
	for _, pkg := range m {
		affected = append(affected, pkg)
	}

	return affected, nil
}
