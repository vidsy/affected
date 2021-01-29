package module

import (
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// ErrSkipPackage will stop the package walker waking the package
var ErrSkipPackage = errors.New("skip walking this package")

type (
	// A GraphConstructor uses go tooling to load packages in given module
	GraphConstructor interface {
		Construct(modules ...string) (Graph, error)
	}

	// A PackageLoader loads packages for one or more modules.
	PackageLoader interface {
		Load(modules ...string) ([]*packages.Package, error)
	}
)

// DefaultGraphConstructor is the default graph constructor
func DefaultGraphConstructor() GraphConstructor {
	return GraphConstructorFunc(func(modules ...string) (Graph, error) {
		pkgs, err := DefaultPackageLoader().Load(modules...)
		if err != nil {
			return nil, err
		}

		return NewGraph(pkgs...), nil
	})
}

// DefaultPackageLoader is the default package loader
func DefaultPackageLoader() PackageLoader {
	return PackageLoaderFunc(func(modules ...string) ([]*packages.Package, error) {
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedTypes,
		}

		for i := range modules {
			modules[i] = fmt.Sprintf("%s/...", modules[i])
		}

		pkgs, err := packages.Load(cfg, modules...)
		if err != nil {
			return nil, err
		}

		return pkgs, nil
	})
}

// Dir returns the package directly based on GoFiles
func Dir(pkg *packages.Package) string {
	var dir string

	if len(pkg.GoFiles) > 0 {
		dir = filepath.Dir(pkg.GoFiles[0])
		if e, err := filepath.EvalSymlinks(dir); err == nil {
			dir = e
		}
	}

	return dir
}

// A FindPackageFunc is used to compare a package during a Find
type FindPackageFunc func(*Package) bool

// FindPackageByID finds packages by their ID (import path)
func FindPackageByID(id string) FindPackageFunc {
	return func(p *Package) bool {
		return p.ID == id
	}
}

// FindPackageByDir finds packages by their source directory based on their .go files
func FindPackageByDir(dir string) FindPackageFunc {
	return func(p *Package) bool {
		return p.Dir == dir
	}
}

// GraphConstructorFunc is an adaptor allowing methods to act as a GraphConstructor
type GraphConstructorFunc func(...string) (Graph, error)

// Construct loads packages for a module
func (fn GraphConstructorFunc) Construct(modules ...string) (Graph, error) {
	return fn(modules...)
}

// Graph is the package import graph
type Graph map[*Package][]*Package

// NewGraph constructs a new package graph
func NewGraph(in ...*packages.Package) Graph {
	g := Graph{}

	pkgs := make([]*Package, len(in))

	for i, pkg := range in {
		pkgs[i] = NewPackage(pkg)
	}

	g.relate(pkgs...)

	return g
}

// Find finds a package in the graph
func (g Graph) Find(fn FindPackageFunc) *Package {
	for p := range g {
		if fn(p) {
			return p
		}
	}

	return nil
}

// ImportPath returns the shortest import path betwween two packages, if no path exists the return
// value will be nil
func (g Graph) ImportPath(start, end *Package) ImportPath {
	p := importPath(g, start, end, make(ImportPath, 0))

	if len(p) == 0 {
		return nil
	}

	return p
}

func importPath(g Graph, start, end *Package, p ImportPath) ImportPath {
	if _, exist := g[start]; !exist {
		return p
	}

	p = append(p, start)

	if start == end {
		return p
	}

	shortest := make([]*Package, 0)

	for _, node := range g[start] {
		if !p.HasNode(node) {
			newPath := importPath(g, node, end, p)
			if len(newPath) > 0 {
				if len(shortest) == 0 || (len(newPath) < len(shortest)) {
					shortest = newPath
				}
			}
		}
	}

	return shortest
}

func (g Graph) relate(pkgs ...*Package) {
	for _, pkg := range pkgs {
		children := make([]*Package, 0)

		for _, imp := range pkg.pkg.Imports {
			for _, p := range pkgs {
				if p.ID == imp.ID {
					children = append(children, p)
				}
			}
		}

		g[pkg] = children
	}
}

// ImportPath holds the import path between two packages
type ImportPath []*Package

// HasNode checks if the path alreadt has the node
func (p ImportPath) HasNode(pkg *Package) bool {
	for _, v := range p {
		if pkg == v {
			return true
		}
	}

	return false
}

// ConstructGraph uses the default graph constructor to load the modules packages
func ConstructGraph(modules ...string) (Graph, error) {
	return DefaultGraphConstructor().Construct(modules...)
}

// NewPackage creates a new *Package from a *packages.Package
func NewPackage(pkg *packages.Package) *Package {
	return &Package{
		ID:      pkg.ID,
		Dir:     Dir(pkg),
		Parents: make([]*Package, 0),
		Imports: make([]*Package, 0),

		pkg: pkg,
	}
}

// PackageLoaderFunc is an adaptor allowing methods to act as a PackageLoader.
type PackageLoaderFunc func(modules ...string) ([]*packages.Package, error)

// Load loads packages for a module.
func (fn PackageLoaderFunc) Load(modules ...string) ([]*packages.Package, error) {
	return fn(modules...)
}

// Package represnets a module package
type Package struct {
	ID      string     `json:"package"`   // Module ID (the import path)
	Dir     string     `json:"directory"` // Directory the package is located in
	Parents []*Package `json:"-"`         // Packages that import this package
	Imports []*Package `json:"-"`         // Packages imported by this package

	pkg *packages.Package // Raw package
}

// WalkDirection is a direction in which we can traverse the packages
type WalkDirection int8

// Walk directions is the direction in which we walk the package tree
const (
	WalkParents WalkDirection = iota + 1
	WalkImports
)

// A WalkFunc is called for
type WalkFunc func(*Package) error

// Walk walks along a packages parents or children calling the given WalkFunc
func Walk(p *Package, d WalkDirection, fn WalkFunc) error {
	var pkgs []*Package

	if err := fn(p); err != nil {
		if err == ErrSkipPackage {
			return nil
		}

		return err
	}

	switch d {
	case WalkParents:
		pkgs = p.Parents
	case WalkImports:
		pkgs = p.Imports
	default:
		return errors.New("invalid walk direction")
	}

	for _, pkg := range pkgs {
		if err := Walk(pkg, d, fn); err != nil {
			return err
		}
	}

	return nil
}
