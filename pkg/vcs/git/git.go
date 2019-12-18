package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vidsy/affected/pkg/glob"
	"github.com/vidsy/affected/pkg/vcs"
)

var _ vcs.ModifiedDirectoriesDetector = new(VCS)

// VCS provides functionality for the git version control system
type VCS struct {
	RepositoryDir string // Repository directory
}

// ModifiedDirectories returns a slice of directories that have modification between two commits
// If no globs are provided then all mofified files within the directory will result in the
// directory as being modified
func (v *VCS) ModifiedDirectories(a, b string, opts ...vcs.ModifiedDirectoriesOption) ([]string, error) {
	files, err := v.ModifiedFiles(a, b, opts...)
	if err != nil {
		return nil, err
	}

	m := make(map[string]struct{})

	for _, file := range files {
		if _, err := os.Stat(file); err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		d := filepath.Dir(file)

		if _, ok := m[d]; !ok {
			m[d] = struct{}{}
		}
	}

	dirs := make([]string, 0, len(m))
	for k := range m {
		dirs = append(dirs, k)
	}

	return dirs, nil
}

// ModifiedFiles returns a set of modified files between two git commits, if no globs are provided
// all files will be marked as modified
func (v *VCS) ModifiedFiles(a, b string, opts ...vcs.ModifiedDirectoriesOption) ([]string, error) {
	lines, err := v.diff(a, b, "--name-only")
	if err != nil {
		return nil, err
	}

	o := &vcs.ModifiedDirectoriesOptions{}

	for _, opt := range opts {
		opt(o)
	}

	files := make([]string, len(lines))

	for i, line := range lines {
		abs, err := filepath.Abs(filepath.Join(v.RepositoryDir, line))
		if err != nil {
			return nil, err
		}

		files[i] = abs
	}

	if len(o.IncludeGlobs) > 0 {
		files = glob.Include(files, o.IncludeGlobs...)
	}

	if len(o.ExcludeGlobs) > 0 {
		files = glob.Exclude(files, o.ExcludeGlobs...)
	}

	return files, nil
}

func (v *VCS) diff(a, b string, flags ...string) ([]string, error) {
	args := append(
		[]string{"diff", fmt.Sprintf("%s..%s", a, b)},
		flags...)

	cmd := exec.Command("git", args...)

	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	lines := make([]string, 0)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, cmd.Wait()
}

// New constructs a new git VCS
func New() (*VCS, error) {
	vcs := &VCS{}

	if vcs.RepositoryDir == "" {
		out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			return nil, err
		}

		dir, err := filepath.EvalSymlinks(strings.TrimSpace(string(out)))
		if err != nil {
			return nil, err
		}

		vcs.RepositoryDir = dir
	}

	return vcs, nil
}
