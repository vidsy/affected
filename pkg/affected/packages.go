package affected

import (
	"encoding/json"

	"github.com/vidsy/affected/pkg/glob"
	"github.com/vidsy/affected/pkg/module"
	"github.com/vidsy/affected/pkg/vcs"
	"github.com/vidsy/affected/pkg/vcs/git"
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
	VCS           vcs.ModifiedDirectoriesDetector // Verson control system to find mofified directories
	PackageLoader module.PackageLoader            // Package loader
	IncludeGlobs  []string                        // Filename globs to include
	ExcludeGlobs  []string                        // Filename globs to exclude
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
func Packages(mod, a, b string, opts ...PackagesOption) ([]Package, error) {
	g, err := git.New()
	if err != nil {
		return nil, err
	}

	o := &PackagesOptions{
		VCS:           g,
		PackageLoader: module.DefaultPackageLoader(),
		IncludeGlobs:  glob.IncludeDefault(),
		ExcludeGlobs:  glob.ExcludeDefault(),
	}

	for _, opt := range opts {
		opt(o)
	}

	dirs, err := o.VCS.ModifiedDirectories(a, b,
		vcs.ModifiedDirectoriesIncludeGlobs(o.IncludeGlobs...),
		vcs.ModifiedDirectoriesExcludeGlobs(o.ExcludeGlobs...))
	if err != nil {
		return nil, err
	}

	graph, err := o.PackageLoader.Load(mod)
	if err != nil {
		return nil, err
	}

	return packages(graph, dirs...), nil
}

func packages(g module.Graph, dirs ...string) []Package {
	m := make(map[string]*Package)

	for _, dir := range dirs {
		if modified := g.Find(module.FindPackageByDir(dir)); modified != nil {
			for pkg := range g {

				path := g.ImportPath(pkg, modified)
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
