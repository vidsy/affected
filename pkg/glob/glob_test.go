package glob

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInclude(t *testing.T) {
	testCases := map[string]struct {
		files    []string
		globs    []string
		expected []string
	}{
		"ReturnsDefaultIncludeGlobsRelativePaths": {
			files: []string{
				"go.mod",
				"go.sum",
				"foo.go",
				"foo_test.go",
				"bar/bar.go",
				"bar/bar_test.go",
				"README.md",
			},
			globs: IncludeDefault(),
			expected: []string{
				"go.mod",
				"go.sum",
				"foo.go",
				"foo_test.go",
				"bar/bar.go",
				"bar/bar_test.go",
			},
		},
		"ReturnsDefaultIncludeGlobsAbsolutePaths": {
			files: []string{
				"/root/go.mod",
				"/root/go.sum",
				"/root/foo.go",
				"/root/foo_test.go",
				"/root/bar/bar.go",
				"/root/bar/bar_test.go",
				"/root/README.md",
			},
			globs: IncludeDefault(),
			expected: []string{
				"/root/go.mod",
				"/root/go.sum",
				"/root/foo.go",
				"/root/foo_test.go",
				"/root/bar/bar.go",
				"/root/bar/bar_test.go",
			},
		},
	}

	for name, testCase := range testCases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, Include(tc.files, tc.globs...))
		})
	}
}

func TestExclude(t *testing.T) {
	testCases := map[string]struct {
		files    []string
		globs    []string
		expected []string
	}{
		"RemovesDefaultExcludedGlobsRelativePaths": {
			files: []string{
				"foo.go",
				"foo_test.go",
				"bar/bar.go",
				"bar/bar_test.go",
			},
			globs: ExcludeDefault(),
			expected: []string{
				"foo.go",
				"bar/bar.go",
			},
		},
		"RemovesDefaultExcludedGlobsAbsolutePaths": {
			files: []string{
				"/root/foo.go",
				"/root/foo_test.go",
				"/root/bar/bar.go",
				"/root/bar/bar_test.go",
			},
			globs: ExcludeDefault(),
			expected: []string{
				"/root/foo.go",
				"/root/bar/bar.go",
			},
		},
	}

	for name, testCase := range testCases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, Exclude(tc.files, tc.globs...))
		})
	}
}
