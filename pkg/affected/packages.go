package affected

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/vidsy/affected/pkg/glob"
	"github.com/vidsy/affected/pkg/module"
	"github.com/vidsy/affected/pkg/vcs"
	"github.com/vidsy/affected/pkg/vcs/git"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

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
	VCS interface {
		vcs.ModifiedDirectoriesDetector
		vcs.ModifiedFilesDetector
		vcs.FileAtRefReader
	}
	GraphConstructor module.GraphConstructor // Graph constructor
	PackageLoader    module.PackageLoader    // Package loader
	IncludeGlobs     []string                // Filename globs to include
	ExcludeGlobs     []string                // Filename globs to exclude
}

// PackagesOption configures packages options
type PackagesOption func(*PackagesOptions)

// WithIncludeFileGlobs sets packages to also mark packages with changes to the given filename globs
// regardless of if any go files have changed
func WithIncludeFileGlobs(v ...string) PackagesOption {
	return func(o *PackagesOptions) {
		o.IncludeGlobs = v
	}
}

// WithAppendIncludeGlobs appends with given globs to the globs already on the options
func WithAppendIncludeGlobs(with ...string) PackagesOption {
	return func(o *PackagesOptions) {
		o.IncludeGlobs = append(o.IncludeGlobs, with...)
	}
}

// WithExcludeFileGlobs sets packages to also mark packages with changes to the given filename globs
// regardless of if any go files have changed
func WithExcludeFileGlobs(v ...string) PackagesOption {
	return func(o *PackagesOptions) {
		o.ExcludeGlobs = v
	}
}

// WithAppendExcludeGlobs appends with given globs to the globs already on the options
func WithAppendExcludeGlobs(with ...string) PackagesOption {
	return func(o *PackagesOptions) {
		o.ExcludeGlobs = append(o.ExcludeGlobs, with...)
	}
}

// NoParents will result in all top level packages being analysed for modifications
func NoParents(p *module.Package) bool {
	return len(p.Parents) == 0
}

// Packages returns a slice of packages affected by direct or indirect changes, use PackageOptions
// to overide defautlt behaviour
func Packages(name, a, b string, opts ...PackagesOption) ([]Package, error) {
	g, err := git.New()
	if err != nil {
		return nil, err
	}

	o := &PackagesOptions{
		VCS:              g,
		GraphConstructor: module.DefaultGraphConstructor(),
		PackageLoader:    module.DefaultPackageLoader(),
		IncludeGlobs:     glob.IncludeDefault(),
		ExcludeGlobs:     glob.ExcludeDefault(),
	}

	for _, opt := range opts {
		opt(o)
	}

	pkgs, err := o.PackageLoader.Load(name)
	if err != nil {
		return nil, err
	}

	// NOTE: There maybe a better way of resolving an absolute go file path to a package import path
	// But at time of writting I could not find one. Ideally we would resolve package path while
	// looping over the files which would be more efficient but since I can't figure out how to
	// resolve a go file path to a package improt path this is the best I could come up with.

	dirs := make(map[string]*packages.Package)

	for _, pkg := range pkgs {
		if len(pkg.GoFiles) > 0 {
			dirs[filepath.Dir(pkg.GoFiles[0])] = pkg
		}
	}

	files, err := o.VCS.ModifiedFiles(a, b,
		vcs.ModifiedDirectoriesIncludeGlobs(o.IncludeGlobs...),
		vcs.ModifiedDirectoriesExcludeGlobs(o.ExcludeGlobs...))
	if err != nil {
		return nil, err
	}

	var modified []*packages.Package

	seen := make(map[string]struct{})

	for _, file := range files {
		if _, err := os.Stat(file); err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		switch {
		case strings.Contains(file, "go.mod"):
			updated, err := diffModfile(o.VCS, o.PackageLoader, a, b, file)
			if err != nil {
				return nil, err
			}

			// Add packages to the packages used to build the import graph
			pkgs = append(pkgs, updated...)

			// Add packages to modified packages
			modified = append(modified, pkgs...)
		case strings.HasSuffix(file, ".go"):
			dir := filepath.Dir(file)

			if _, seen := seen[dir]; seen {
				continue
			}

			pkg, ok := dirs[dir]
			if !ok {
				continue
			}

			seen[dir] = struct{}{}
			modified = append(modified, pkg)
		}
	}

	// Build the graph
	graph := module.NewGraph(pkgs...)

	// Return packages affected by modified packages
	return affected(graph, modified...), nil
}

func affected(graph module.Graph, pkgs ...*packages.Package) []Package {
	m := make(map[string]*Package)

	for _, pkg := range pkgs {
		if modified := graph.Find(module.FindPackageByID(pkg.ID)); modified != nil {
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

	return affected
}

// diffModfile diffs a go.mod file between two refs returning a slice modified modules and a slice of
// all packages modifiled, for example, if the module "github.com/aws/aws-sdk-go" version has
// changed all packages within that module are considered as modified.
func diffModfile(r vcs.FileAtRefReader, l module.PackageLoader, refA, refB, path string) ([]*packages.Package, error) {
	data, err := r.ReadFileAtRef(refA, filepath.Base(path))
	if err != nil {
		return nil, err
	}

	modfileA, err := modfile.Parse(path, data, nil)
	if err != nil {
		return nil, err
	}

	data, err = r.ReadFileAtRef(refB, filepath.Base(path))
	if err != nil {
		return nil, err
	}

	modfileB, err := modfile.Parse(path, data, nil)
	if err != nil {
		return nil, err
	}

	require := make(map[string]string)

	for _, module := range modfileA.Require {
		require[module.Mod.Path] = module.Mod.Version
	}

	for _, module := range modfileB.Require {
		v, ok := require[module.Mod.Path]
		if !ok { // new module
			continue
		}

		if module.Mod.Version == v { // no change to version, remove from the map
			delete(require, module.Mod.Path)
		}
	}

	var modules []string

	for module := range require {
		modules = append(modules, module)
	}

	packages, err := l.Load(modules...)
	if err != nil {
		return nil, err
	}

	return packages, nil
}
