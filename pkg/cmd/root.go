package cmd

import (
	"github.com/spf13/cobra"
)

const long = `Affected detects services that have directly or indirectly been modified via
library changes. It can output the results to multiple formats and locations. Use the -f/--format
flag to give the formats you would like. Use the -a/-b flag to change the commits to compare.`

const afterOffset = 1

// Options configures how the command is run
type Options struct {
	// Persistent options
	Format               string
	Module               string
	CommitA              string
	CommitB              string
	Discard              bool
	IncludeGlobs         []string
	ExcludeGlobs         []string
	OverrideIncludeGlobs bool
	OverrideExcludeGlobs bool

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
			return Run(opts)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.Format, "format", "f", "json", "e.g text/json/json-minified")
	cmd.PersistentFlags().StringVarP(&opts.CommitA, "a", "a", "origin/master", "Commit A")
	cmd.PersistentFlags().StringVarP(&opts.CommitB, "b", "b", "HEAD", "Commit B")
	cmd.PersistentFlags().BoolVarP(&opts.Discard, "discard", "d", false, "Discard output")
	cmd.PersistentFlags().StringArrayVarP(&opts.IncludeGlobs, "include", "i", []string{}, "File name globs to include")
	cmd.PersistentFlags().StringArrayVarP(&opts.ExcludeGlobs, "exclude", "x", []string{}, "File name globs to exclude")
	cmd.PersistentFlags().BoolVar(&opts.OverrideIncludeGlobs, "override-include-globs", false, "Default include globs will be omitted, only globs you provide will be used")
	cmd.PersistentFlags().BoolVar(&opts.OverrideExcludeGlobs, "override-exclude-globs", false, "Default exclude globs will be omitted, only globs you provide will be used")

	cmd.AddCommand(GroupCmd(opts))

	return cmd
}
