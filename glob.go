package fileglob

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/spf13/afero"
)

type FileSystem afero.Fs

// globOptions allowed to be passed to Glob.
type globOptions struct {
	fs afero.Fs

	// if matchDirectories directly is set to true  a matching directory will
	// be treated just like a matching file. If set to false, a matching directory
	// will auto-match all files inside instead of the directory itself.
	matchDirectoriesDirectly bool

	separator rune
}

type OptFunc func(opts *globOptions) error

// WithFs allows to provide another afero.Fs implementation to Glob.
func WithFs(fs FileSystem) OptFunc {
	return func(opts *globOptions) error {
		opts.fs = fs
		return nil
	}
}

// MatchDirectories determines weather a matching directory should
// result in only the folder name itself being returned (true) or
// in all files inside that folder being returned (false).
func MatchDirectories(v bool) OptFunc {
	return func(opts *globOptions) error {
		opts.matchDirectoriesDirectly = v
		return nil
	}
}

func WithSeparator(sep rune) OptFunc {
	return func(opts *globOptions) error {
		opts.separator = sep
		return nil
	}
}

// QuoteMeta returns a string that quotes all glob pattern meta characters
// inside the argument text; For example, QuoteMeta(`{foo*}`) returns `\{foo\*\}`.
func QuoteMeta(pattern string) string {
	return glob.QuoteMeta(pattern)
}

// Glob returns all files that match the given pattern in the current directory.
func Glob(pattern string, opts ...OptFunc) ([]string, error) { // nolint:funlen
	options, err := compileOptions(opts)
	if err != nil {
		return nil, fmt.Errorf("compile options: %w", err)
	}

	pattern = strings.TrimPrefix(pattern, "./")

	var fs = options.fs
	var matches []string

	matcher, err := glob.Compile(pattern, options.separator)
	if err != nil {
		return matches, fmt.Errorf("compile glob pattern: %w", err)
	}

	prefix, err := staticPrefix(pattern, options.separator)
	if err != nil {
		return nil, fmt.Errorf("cannot determine static prefix: %w", err)
	}

	prefixInfo, err := fs.Stat(prefix)
	if os.IsNotExist(err) {
		// if the prefix does not exist, the whole
		// glob pattern won't match anything
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("stat prefix: %w", err)
	}

	if !prefixInfo.IsDir() {
		// if the prefix is a file, it either has to be
		// the only match, or nothing matches at all
		if matcher.Match(prefix) {
			return []string{prefix}, nil
		}

		return []string{}, nil
	}

	return matches, afero.Walk(fs, prefix, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !matcher.Match(path) {
			return nil
		}

		if info.IsDir() {
			if options.matchDirectoriesDirectly {
				matches = append(matches, path)
				return nil
			}

			// a direct match on a directory implies that all files inside
			// match if options.matchFolders is false
			filesInDir, err := filesInDirectory(fs, path)
			if err != nil {
				return err
			}

			matches = append(matches, filesInDir...)
			return filepath.SkipDir
		}

		matches = append(matches, path)

		return nil
	})
}

func compileOptions(optFuncs []OptFunc) (*globOptions, error) {
	var opts = &globOptions{
		fs:        afero.NewOsFs(),
		separator: filepath.Separator,
	}

	for _, optFunc := range optFuncs {
		err := optFunc(opts)
		if err != nil {
			return nil, fmt.Errorf("applying options: %w", err)
		}
	}

	return opts, nil
}

func filesInDirectory(fs afero.Fs, dir string) ([]string, error) {
	var files []string

	return files, afero.Walk(fs, dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
}
