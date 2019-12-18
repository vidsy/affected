package cmd

import (
	"errors"

	"github.com/vidsy/affected/pkg/affected"
	"github.com/vidsy/affected/pkg/module"
)

// Run exectues the affected tool with the given options
func Run(opts *Options) error {
	// Figure out module path from go.mod if not provided by the user
	if opts.Module == "" {
		m, err := module.Path()
		if err != nil {
			return err
		}

		opts.Module = m
	}

	popts := []affected.PackagesOption{}

	if len(opts.IncludeGlobs) > 0 {
		fn := affected.WithAppendIncludeGlobs(opts.IncludeGlobs...)
		if opts.OverrideIncludeGlobs {
			fn = affected.WithIncludeFileGlobs(opts.IncludeGlobs...)
		}

		popts = append(popts, fn)
	}

	if len(opts.ExcludeGlobs) > 0 {
		fn := affected.WithAppendExcludeGlobs(opts.ExcludeGlobs...)
		if opts.OverrideExcludeGlobs {
			fn = affected.WithExcludeFileGlobs(opts.ExcludeGlobs...)
		}

		popts = append(popts, fn)
	}

	// Load affected packages
	pkgs, err := affected.Packages(opts.Module, opts.CommitA, opts.CommitB, popts...)
	if err != nil {
		return err
	}

	var v interface{} = pkgs

	// If we are grouping group packages by the grouping function
	if fn := GroupFunc(opts); fn != nil {
		v = affected.GroupPackages(fn, pkgs...)
	}

	// Write the value to the correct format to the given writer
	w := Writer(opts)
	switch opts.Format {
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
