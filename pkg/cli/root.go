package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/vidsy/affected/pkg/affected"
	"github.com/vidsy/affected/pkg/module"
)

const long = `Affected detects services that have directly or indirectly been modified via
library changes. It can output the results to multiple formats and locations. Use the -f/--format
flag to give the formats you would like. Use the -a/-b flag to change the commits to compare.`

const afterOffset = 1

// Options configures how the command is run
type Options struct {
	// Persistent options
	Format  string
	Module  string
	CommitA string
	CommitB string
	Discard bool

	// Grouping options
	GroupByPkgPrefix string
	GroupByAfter     int
}

// RootCmd returns the root CLI command
func RootCmd() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:     "affected",
		Short:   "Detects packages affected changes to other packages via their improts and vcs.",
		Long:    long,
		Example: "affected -f json -a origin/master -b HEAD > affected.json",
		RunE: func(*cobra.Command, []string) error {
			return Run(Writer(opts), opts.Format, opts.CommitA, opts.CommitB, nil)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.Format, "format", "f", "json", "e.g text/json/json-minified")
	cmd.PersistentFlags().StringVarP(&opts.CommitA, "a", "a", "origin/master", "Commit A")
	cmd.PersistentFlags().StringVarP(&opts.CommitB, "b", "b", "HEAD", "Commit B")
	cmd.PersistentFlags().BoolVarP(&opts.Discard, "discard", "d", false, "Discard output")

	cmd.AddCommand(GroupCmd(opts))

	return cmd
}

// GroupCmd returns the group sub command which allows packages to be grouped
func GroupCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "group",
		Short:   "Group affected packages by a user defined group",
		Example: "affected group --pkg-prefix foo.com/pkg --after 1 -f json -a origin/master -b HEAD > affected.json",
		RunE: func(*cobra.Command, []string) error {
			var fn affected.GroupFunc
			if fn = GroupFunc(opts); fn == nil {
				return errors.New("no group function provided")
			}

			return Run(Writer(opts), opts.Format, opts.CommitA, opts.CommitB, fn)
		},
	}

	cmd.Flags().StringVar(&opts.GroupByPkgPrefix, "pkg-prefix", "", "Group by package prefix")
	cmd.Flags().IntVar(&opts.GroupByAfter, "after", 0, "Group after n (one-based numbering)")

	return cmd
}

// Run exectues the affected tool with the given options
func Run(w io.Writer, f, a, b string, fn affected.GroupFunc) error {
	var v interface{}

	// Figure out module path from go.mod
	m, err := module.Path()
	if err != nil {
		return err
	}

	// Load affected packages
	pkgs, err := affected.Packages(m, a, b)
	if err != nil {
		return err
	}

	v = pkgs

	// If we are grouping group packages by the grouping function
	if fn != nil {
		v = affected.GroupPackages(fn, pkgs...)
	}

	// Write the value to the correct format to the given writer
	switch f {
	case "json":
		return WriteJSON(w, v, true)
	case "json-minified":
		return WriteJSON(w, v, false)
	case "text":
		return WriteText(w, v)
	default:
		return errors.New("unsupported format")
	}
}

// GroupFunc returns the correct group function based on CLI arguments, if no group function can be
// determined nil will be returned
func GroupFunc(opts *Options) affected.GroupFunc {
	var fn affected.GroupFunc

	if opts.GroupByPkgPrefix != "" {
		fn = affected.GroupByPkgPrefix(opts.GroupByPkgPrefix)
		if opts.GroupByAfter > 0 {
			fn = affected.GroupByPkgAfterPrefix(opts.GroupByPkgPrefix, opts.GroupByAfter-afterOffset)
		}
	}

	return fn
}

// Writer returns a writer based on the options, defaults to stdout
func Writer(opts *Options) io.Writer {
	var w io.Writer = os.Stdout

	if opts.Discard {
		w = ioutil.Discard
	}

	// Add suoport for writing directly to file with -v output for also writing to stdout
	// at the same time via a T writer

	return w
}

// WriteJSON writes the JSON output to the writer
func WriteJSON(w io.Writer, v interface{}, indent bool) error {
	var marshal func(interface{}) ([]byte, error) = json.Marshal
	if indent {
		marshal = func(interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "  ")
		}
	}

	b, err := marshal(v)
	if err != nil {
		return err
	}

	if _, err := w.Write(b); err != nil {
		return err
	}

	return nil
}

// WriteText writes the string output to the writer
func WriteText(w io.Writer, v interface{}) error {
	if _, err := fmt.Fprintln(w, v); err != nil {
		return err
	}

	return nil
}
