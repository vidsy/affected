package affected

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vidsy/affected/module"
	"github.com/vidsy/affected/vcs"
	"github.com/vidsy/affected/vcs/git"
)

// Cause is why a package has been marked as affected
type Cause struct {
	Package    *module.Package   // The package that has modififcations
	ImportPath module.ImportPath // The import graph to that package
}

// A Group holds a group of affected packages
type Group struct {
	Name     string
	Packages []Package
	Causes   []Cause
}

// MarshalJSON marshales groups of affected packages to json
func (g Group) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"group":    g.Name,
		"packages": g.Packages,
	})
}

// Groups holds groups of grouped affected packages
type Groups []Group

func (g Groups) String() string {
	w := new(bytes.Buffer)

	for n, group := range g {
		if n > 0 {
			fmt.Fprint(w, "\n")
		}

		title := fmt.Sprintf("Group: %s", group.Name)

		fmt.Fprintln(w, title)
		fmt.Fprintln(w, strings.Repeat("-", len(title)))

		for _, pkg := range group.Packages {
			fmt.Fprintln(w, "- Package: ", pkg.ID)

			for _, cause := range pkg.Causes {
				fmt.Fprintln(w, " - Caused By:", cause.Package.ID)

				for i, pkg := range cause.ImportPath {
					fmt.Fprintln(w, fmt.Sprintf("  %s %s", strings.Repeat(">", i), pkg.ID))
				}
			}
		}
	}

	return w.String()
}

// A GroupFunc determines a packages group name and if it should be grouped
type GroupFunc func(*Package) (string, bool)

// GroupPackages groups affected packages into groups determined by the GroupFunc
func GroupPackages(fn GroupFunc, pkgs ...Package) Groups {
	gm := make(map[string]*Group)

	for _, pkg := range pkgs {
		pkg := pkg

		if name, ok := fn(&pkg); ok {
			g, ok := gm[name]
			if !ok {
				g = &Group{
					Name: name,
				}

				gm[name] = g
			}

			g.Packages = append(g.Packages, pkg)
			g.Causes = append(g.Causes, pkg.Causes...)
		}
	}

	groups := make(Groups, 0, len(gm))
	for _, g := range gm {
		groups = append(groups, *g)
	}

	return groups
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
func Packages(mod, a, b string) ([]Package, error) {
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
			for pkg := range graph {
				path := graph.ImportPath(pkg, modified)
				if len(path) > 0 {
					affected, ok := m[pkg.ID]
					if !ok {
						affected = &Package{
							Package: pkg,
						}

						m[pkg.ID] = affected
					}

					affected.Causes = append(affected.Causes, Cause{
						Package:    modified,
						ImportPath: path,
					})
				}
			}
		}
	}

	affected := make([]Package, 0, len(m))
	for _, pkg := range m {
		affected = append(affected, *pkg)
	}

	return affected, nil
}
