package vcs

// ModifiedDirectoriesOptions holds optional configuration for returning modified directories
type ModifiedDirectoriesOptions struct {
	IncludeGlobs []string
	ExcludeGlobs []string
}

// ModifiedDirectoriesOption updates ModifiedDirectoriesOptions
type ModifiedDirectoriesOption func(*ModifiedDirectoriesOptions)

// ModifiedDirectoriesIncludeGlobs sets the include globs
func ModifiedDirectoriesIncludeGlobs(globs ...string) ModifiedDirectoriesOption {
	return func(opts *ModifiedDirectoriesOptions) {
		opts.IncludeGlobs = globs
	}
}

// ModifiedDirectoriesExcludeGlobs sets the exclude globs
func ModifiedDirectoriesExcludeGlobs(globs ...string) ModifiedDirectoriesOption {
	return func(opts *ModifiedDirectoriesOptions) {
		opts.ExcludeGlobs = globs
	}
}

// A ModifiedDirectoriesDetector can detect modified directories
type ModifiedDirectoriesDetector interface {
	ModifiedDirectories(a, b string, opts ...ModifiedDirectoriesOption) ([]string, error)
}

// A ModifiedFilesDetector can detect modified files.
type ModifiedFilesDetector interface {
	ModifiedFiles(a, b string, opts ...ModifiedDirectoriesOption) ([]string, error)
}

// FileAtRefReader reads a file at a given ref, e.g a branch or commit.
type FileAtRefReader interface {
	ReadFileAtRef(ref, name string) ([]byte, error)
}
