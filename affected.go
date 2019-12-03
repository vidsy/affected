package affected

import (
	"encoding/json"

	"github.com/vidsy/affected/module"
	"github.com/vidsy/affected/vcs"
	"github.com/vidsy/affected/vcs/git"
)

// Cause is why a package has been marked as affected
type Cause struct {
	Package    *module.Package   // The package that has modififcations
	ImportPath module.ImportPath // The import graph to that package
}

// Package represents a modified package
type Package struct {
	*module.Package

	Causes []Cause
}

// MarshalJSON marshals a package with its causes into a json structure
func (p *Package) MarshalJSON() ([]byte, error) {
	causes := make([]map[string]interface{}, len(p.Causes))

	for i, cause := range p.Causes {
		causes[i] = map[string]interface{}{
			"package": cause.Package,
			"imports": cause.ImportPath,
		}
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

// SubjectFunc sets what packages should be trated as the subjects that will be affected by
// modifications in other packages
type SubjectFunc func(p *module.Package) bool

// NoParents will result in all top level packages being analysed for modifications
func NoParents(p *module.Package) bool {
	return len(p.Parents) == 0
}

// Packages returns packages affected by two different commits either directly or indirectly
func Packages(mod, a, b string, fn SubjectFunc) ([]*Package, error) {
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

	// Load package graph
	graph, err := opts.PL.Load(mod)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*Package)

	for _, dir := range dirs {
		if modified := graph.Find(module.FindPackageByDir(dir)); modified != nil {
			err := module.Walk(modified, module.WalkParents, func(p *module.Package) error {
				if fn(p) {
					affected, ok := m[p.ID]
					if !ok {
						affected = &Package{
							Package: p,
						}

						m[p.ID] = affected
					}

					affected.Causes = append(affected.Causes, Cause{
						Package:    modified,
						ImportPath: graph.ImportPath(affected.Package, modified),
					})

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
