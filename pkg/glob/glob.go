// Package glob provides utilities for matching file globs. This package is heavily inspired by
// github.com/go-godo/godo which had done a lot of the hard work. However we did not want to depend
// upon this library as a dependency, since it provides other unneeded functionality.
package glob

import (
	"bytes"
	"regexp"
	"strings"
)

const (
	// NotSlash is any rune but path separator.
	notSlash = "[^/]"
	// AnyRune is zero or more non-path separators.
	anyRune = notSlash + "*"
	// ZeroOrMoreDirectories is used by ** patterns.
	zeroOrMoreDirectories = `(?:[.{}\w\-\ ]+\/)*`
)

// IncludeDefault returns the default set of default include globs
func IncludeDefault() []string {
	return []string{
		"/**/*.go",
		"/**/go.mod",
		"/**/go.sum",
	}
}

// ExcludeDefault returns the default set of default exlude globs
func ExcludeDefault() []string {
	return []string{
		"/**/*_test.go",
	}
}

// Match matches a glob to a subject, returning bool
func Match(glob, subject string) (bool, error) {
	re, err := Regexp(glob)
	if err != nil {
		return false, err
	}

	return re.MatchString(subject), nil
}

// Include returns a filtered slice of strings by the globs that should be included
func Include(in []string, globs ...string) []string {
	out := make([]string, 0)

	for _, v := range in {
		for _, g := range globs {
			if m, err := Match(g, v); m && err == nil {
				out = append(out, v)
				break
			}
		}
	}

	return out
}

// Exclude returns a filtered slice of strings by the globs that should be excluded
func Exclude(in []string, globs ...string) []string {
	for _, x := range Include(in, globs...) {
		for i, v := range in {
			if v == x {
				in = append(in[:i], in[i+2:]...)
				break
			}
		}
	}

	return in
}

// Regexp builds a regular expression for the given glob
func Regexp(glob string) (*regexp.Regexp, error) {
	var (
		group bool
		skip  int8
	)

	b := &buffer{new(bytes.Buffer)}
	b.WriteString("^")

	for pos, char := range glob {
		if skip > 0 {
			skip--
			continue
		}

		switch char {
		case '\\', '$', '^', '+', '.', '(', ')', '=', '!', '|':
			if err := b.WriteRunes('\\', char); err != nil {
				return nil, err
			}
		case '/':
			skip = forwardSlash(b, pos, glob)
		case '?':
			b.WriteRune('.')
		case '[', ']':
			b.WriteRune(char)
		case '{':
			skip, group = openBrace(b, pos, glob, group)
		case '}':
			group = closeBrace(b, group)
		case ',':
			if err := comma(b, char, group); err != nil {
				return nil, err
			}
		case '*':
			skip = asterix(b, pos, glob)
		default:
			b.WriteRune(char)
		}
	}

	b.WriteString("$")

	return regexp.Compile(b.String())
}

type buffer struct {
	*bytes.Buffer
}

func (b *buffer) WriteRunes(runes ...rune) error {
	for _, r := range runes {
		if _, err := b.Buffer.WriteRune(r); err != nil {
			return err
		}
	}

	return nil
}

func forwardSlash(b *buffer, pos int, glob string) int8 {
	b.WriteRune('/')

	rest := glob[pos:]
	if strings.HasPrefix(rest, "/**/") {
		b.WriteString(zeroOrMoreDirectories)
		return 3 // nolint: mnd
	}

	if rest == "/**" {
		b.WriteString(".*")
		return 2 // nolint: mmd
	}

	return 0
}

func openBrace(b *buffer, pos int, glob string, group bool) (int8, bool) {
	if pos < len(glob)-1 {
		if glob[pos+1:pos+2] == "{" {
			b.WriteString("\\{")

			return 1, group // nolint: mmd
		}
	}

	b.WriteRune('(')

	return 0, true
}

func closeBrace(b *buffer, group bool) bool {
	if group {
		b.WriteRune(')')
	} else {
		b.WriteRune('}')
	}

	return false
}

func comma(b *buffer, char rune, group bool) error {
	if group {
		b.WriteRune('|')

		return nil
	}

	return b.WriteRunes('\\', char)
}

func asterix(b *buffer, pos int, glob string) int8 {
	rest := glob[pos:]
	if strings.HasPrefix(rest, "**/") {
		b.WriteString(zeroOrMoreDirectories)

		return 2 // nolint: mmd
	}

	b.WriteString(anyRune)

	return 0
}
