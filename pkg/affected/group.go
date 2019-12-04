package affected

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

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

// GroupByPkgPrefix groups packages by a prefix
func GroupByPkgPrefix(prefix string) GroupFunc {
	return func(pkg *Package) (string, bool) {
		if strings.HasPrefix(pkg.ID, prefix) {
			return prefix, true
		}

		return "", false
	}
}

// GroupByPkgAfterPrefix returns a GroupFunc that will group affected packages by a prefix, the group
// name will be the n path element after the prefix, for example:
// - foo.com/pkg/a
// - foo.com/pkg/b
// If the given prefix is foo.com/pkg and the element given is 0
// If the number of path elements after the prefix is < the given n the package will not be grouped
func GroupByPkgAfterPrefix(prefix string, n int) GroupFunc {
	return func(pkg *Package) (string, bool) {
		if strings.HasPrefix(pkg.ID, prefix) {
			parts := strings.Split(strings.TrimLeft(strings.TrimPrefix(pkg.ID, prefix), "/"), "/")
			if len(parts) >= n {
				return strings.Join([]string{prefix, parts[n]}, "/"), true
			}
		}

		return "", false
	}
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
