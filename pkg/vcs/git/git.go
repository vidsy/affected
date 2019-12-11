package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VCS provides functionality for the git version control system
type VCS struct {
	RepositoryDir string // Repository directory
}

// ModifiedDirectories returns a slice of directories that have modification between two commits
func (vcs *VCS) ModifiedDirectories(a, b string) ([]string, error) {
	files, err := vcs.ModifiedFiles(a, b)
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

// ModifiedFiles returns a set of modified files between two git commits
func (vcs *VCS) ModifiedFiles(a, b string) ([]string, error) {
	lines, err := vcs.diff(a, b, "--name-only")
	if err != nil {
		return nil, err
	}

	files := make([]string, len(lines))

	for i, line := range lines {
		abs, err := filepath.Abs(filepath.Join(vcs.RepositoryDir, line))
		if err != nil {
			return nil, err
		}

		files[i] = abs
	}

	return files, nil
}

func (vcs *VCS) diff(a, b string, flags ...string) ([]string, error) {
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
