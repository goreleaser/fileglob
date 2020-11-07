package fileglob

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestGlob(t *testing.T) {
	t.Run("real", func(t *testing.T) {
		matches, err := Glob("*_test.go")
		require.NoError(t, err)
		require.Equal(t, []string{
			"glob_test.go",
			"prefix_test.go",
		}, matches)
	})

	t.Run("simple", func(t *testing.T) {
		matches, err := globInMemoryFs("./a/*/*", []string{
			"./c/file1.txt",
			"./a/nope/file1.txt",
			"./a/something",
			"./a/b/file1.txt",
			"./a/c/file2.txt",
			"./a/d/file1.txt",
		})
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file1.txt",
			"a/c/file2.txt",
			"a/d/file1.txt",
			"a/nope/file1.txt",
		}, matches)
	})

	t.Run("single file", func(t *testing.T) {
		matches, err := globInMemoryFs("a/b/*", []string{
			"./c/file1.txt",
			"./a/nope/file1.txt",
			"./a/b/file1.txt",
		})
		require.NoError(t, err)
		require.Equal(t, []string{"a/b/file1.txt"}, matches)
	})

	t.Run("double star", func(t *testing.T) {
		matches, err := globInMemoryFs("a/**/*", []string{
			"./a/nope.txt",
			"./a/d/file1.txt",
			"./a/e/f/file1.txt",
			"./a/b/file1.txt",
		})
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file1.txt",
			"a/d/file1.txt",
			"a/e/f/file1.txt",
		}, matches)
	})

	t.Run("direct match", func(t *testing.T) {
		matches, err := globInMemoryFs("a/b/c", []string{
			"./a/nope.txt",
			"./a/b/c",
		})
		require.NoError(t, err)
		require.Equal(t, []string{"a/b/c"}, matches)
	})

	t.Run("direct no match", func(t *testing.T) {
		matches, err := globInMemoryFs("a/b/d", []string{
			"./a/nope.txt",
			"./a/b/dc",
		})
		require.NoError(t, err)
		require.Equal(t, []string{}, matches)
	})

	t.Run("no matches", func(t *testing.T) {
		matches, err := globInMemoryFs("z/*", []string{
			"./a/nope.txt",
		})
		require.NoError(t, err)
		require.Empty(t, matches)
	})

	t.Run("empty folder", func(t *testing.T) {
		matches, err := globInMemoryFs("a*", []string{})
		require.NoError(t, err)
		require.Empty(t, matches)
	})

	t.Run("escaped asterisk", func(t *testing.T) {
		matches, err := globInMemoryFs("a/\\*/b", []string{
			"a/a/b",
			"a/*/b",
			"a/**/b",
		})
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/*/b",
		}, matches)
	})

	t.Run("escaped curly braces", func(t *testing.T) {
		matches, err := globInMemoryFs("\\{a,b\\}/c", []string{
			"a/c",
			"b/c",
			"{a,b}/c",
		})
		require.NoError(t, err)
		require.Equal(t, []string{
			"{a,b}/c",
		}, matches)
	})

	t.Run("invalid pattern", func(t *testing.T) {
		matches, err := globInMemoryFs("[*", []string{})
		require.EqualError(t, err, "unexpected end of input")
		require.Empty(t, matches)
	})

	t.Run("prefix is a file", func(t *testing.T) {
		matches, err := globInMemoryFs("ab/c/*", []string{
			"ab/c",
			"ab/d",
			"ab/e",
		})
		require.NoError(t, err)
		require.Empty(t, matches)
	})
}

func TestQuoteMeta(t *testing.T) {
	matches, err := globInMemoryFs(QuoteMeta("{a,b}/c"), []string{
		"a/c",
		"b/c",
		"{a,b}/c",
	})
	require.NoError(t, err)
	require.Equal(t, []string{
		"{a,b}/c",
	}, matches)
}

func globInMemoryFs(pattern string, files []string, options ...Options) ([]string, error) {
	var fs = afero.NewMemMapFs()
	if err := createFiles(fs, files); err != nil {
		return []string{}, err
	}
	return Glob(pattern, append(options, WithFs(fs))...)
}

func createFiles(fs afero.Fs, files []string) error {
	for _, file := range files {
		if _, err := fs.Create(file); err != nil {
			return err
		}
	}
	return nil
}
