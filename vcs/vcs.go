package vcs

import "github.com/vidsy/affected/vcs/git"

var _ ModifiedDirectoriesDetector = new(git.VCS)

// A ModifiedDirectoriesDetector can detect modified directories
type ModifiedDirectoriesDetector interface {
	ModifiedDirectories(a, b string) ([]string, error)
}
