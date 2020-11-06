package zglob

import (
	"fmt"
	"os"
	"path"

	"github.com/gobwas/glob"
	"github.com/spf13/afero"
)

func Glob(pattern string) ([]string, error) {
	return GlobWithFs(afero.NewOsFs(), pattern)
}

func GlobWithFs(fs afero.Fs, pattern string) ([]string, error) {
	var matches []string

	matcher, err := glob.Compile(pattern)
	if err != nil {
		return matches, err
	}

	prefix, err := staticPrefix(pattern)
	if err != nil {
		return nil, fmt.Errorf("determine static prefix: %w", err)
	}

	prefixInfo, err := fs.Stat(prefix)
	if os.IsNotExist(err) {
		// if the prefix does not exist, the whole
		// glob pattern won't match anything
		return []string{}, nil
	} else if err != nil {
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

	return matches, afero.Walk(fs, prefix, func(currentPath string, info os.FileInfo,
		err error) error {
		if err != nil {
			return err
		}

		if !matcher.Match(info.Name()) {
			return nil
		}

		if info.IsDir() {
			// a direct match on a directory implies that all files inside match
			filesInDir, err := filesInDirectory(fs, path.Join(currentPath, info.Name()))
			if err != nil {
				return err
			}

			matches = append(matches, filesInDir...)
		} else {
			matches = append(matches, path.Join(currentPath, info.Name()))
		}

		return nil
	})
}

func filesInDirectory(fs afero.Fs, dir string) ([]string, error) {
	files := []string{}

	err := afero.Walk(fs, dir, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, path.Join(currentPath, info.Name()))

		return nil
	})

	return files, err
}
