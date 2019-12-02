package module

import (
	"errors"
	"fmt"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// DefaultPackageLoader is the default package loader
func DefaultPackageLoader() PackageLoader {
	return PackageLoaderFunc(func(module string) (Packages, error) {
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedTypes,
		}

		lpkgs, err := packages.Load(cfg, module+"/...")
		if err != nil {
			return nil, err
		}

		pkgs := Packages{}
		pkgs = pkgs.AddNew(module, lpkgs...)

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

// ErrSkipPackage will stop the package walker waking the package
var ErrSkipPackage = errors.New("skip walking this package")

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

// LoadPackages uses the default package loader to load the modules packages
func LoadPackages(module string) (Packages, error) {
	return DefaultPackageLoader().Load(module)
}

// NewPackage creates a new *Package from a *packages.Package
func NewPackage(module string, pkg *packages.Package) *Package {
	return &Package{
		ID:      pkg.ID,
		Dir:     Dir(pkg),
		Parents: make([]*Package, 0),
		Imports: make([]*Package, 0),

		pkg: pkg,
	}
}

// Package represnets a module package
type Package struct {
	ID      string     // Module ID (the import path)
	Dir     string     // Directory the package is located in
	Parents []*Package // Packages that import this package
	Imports []*Package // Packages imported by this package

	pkg *packages.Package // Raw package
}

// Packages holds the modules package structure which can be traversed
type Packages map[string]*Package

// Add adds Packages
func (p Packages) Add(pkgs ...*Package) Packages {
	for _, pkg := range pkgs {
		p[pkg.ID] = pkg
	}

	p.relate()
	p.prune()

	return p
}

// AddNew creates a new Package and adds it
func (p Packages) AddNew(module string, in ...*packages.Package) Packages {
	pkgs := make([]*Package, len(in))

	for i, pkg := range in {
		pkgs[i] = NewPackage(module, pkg)
	}

	return p.Add(pkgs...)
}

// Find will traverse the packages and find the package, privde a FindPackageFunc used to decide if
// a package has been founnd, if no package is found the return value will be nil
func (p Packages) Find(cmp FindPackageFunc) *Package {
	for _, pkg := range p {
		if v := find(pkg, cmp); v != nil {
			return v
		}
	}

	return nil
}

func (p Packages) String() string {
	var s string

	var visit func(*Package, string)

	visit = func(p *Package, indent string) {
		s += fmt.Sprintf("\n%s %s", indent, p.ID)

		for _, i := range p.Imports {
			visit(i, indent+">")
		}
	}

	for _, v := range p {
		visit(v, ">")
	}

	return s
}

func find(p *Package, cmp FindPackageFunc) *Package {
	if cmp(p) {
		return p
	}

	for _, pkg := range p.Imports {
		return find(pkg, cmp)
	}

	return nil
}

func (p Packages) prune() {
	for k, v := range p {
		if len(v.Parents) > 0 {
			delete(p, k)
		}
	}
}

func (p Packages) relate() {
	for _, parent := range p {
		for _, imp := range parent.pkg.Imports {
			if child, ok := p[imp.ID]; ok {
				parent.Imports = append(parent.Imports, child)
				child.Parents = append(child.Parents, parent)
			}
		}
	}
}

// A PackageLoader uses go tooling to load packages in given module
type PackageLoader interface {
	Load(module string) (Packages, error)
}

// PackageLoaderFunc is an adaptor allowing methods to act as a PackageLoader
type PackageLoaderFunc func(string) (Packages, error)

// Load loads packages for a module
func (fn PackageLoaderFunc) Load(m string) (Packages, error) {
	return fn(m)
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
