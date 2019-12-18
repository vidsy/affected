package cli

import (
	"github.com/spf13/cobra"
	"github.com/vidsy/affected/pkg/affected"
)

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

// GroupCmd returns the group sub command which allows packages to be grouped
func GroupCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "group",
		Short:   "Group affected packages by a user defined group",
		Example: "affected group --pkg-prefix foo.com/pkg --after 1 -f json -a origin/master -b HEAD > affected.json",
		RunE: func(*cobra.Command, []string) error {
			return Run(opts)
		},
	}

	cmd.Flags().StringVar(&opts.GroupByPkgPrefix, "pkg-prefix", "", "Group by package prefix")
	cmd.Flags().IntVar(&opts.GroupByAfter, "after", 0, "Group after n (one-based numbering)")

	return cmd
}
