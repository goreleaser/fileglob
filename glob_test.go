package fileglob

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestGlob(t *testing.T) { // nolint:funlen
	t.Run("real", func(t *testing.T) {
		matches, err := Glob("*_test.go")
		require.NoError(t, err)
		require.Equal(t, []string{
			"glob_test.go",
			"prefix_test.go",
		}, matches)
	})

	t.Run("simple", func(t *testing.T) {
		matches, err := Glob("./a/*/*", WithFs(testFs(t, []string{
			"./c/file1.txt",
			"./a/nope/file1.txt",
			"./a/something",
			"./a/b/file1.txt",
			"./a/c/file2.txt",
			"./a/d/file1.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file1.txt",
			"a/c/file2.txt",
			"a/d/file1.txt",
			"a/nope/file1.txt",
		}, matches)
	})

	t.Run("single file", func(t *testing.T) {
		matches, err := Glob("a/b/*", WithFs(testFs(t, []string{
			"./c/file1.txt",
			"./a/nope/file1.txt",
			"./a/b/file1.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{"a/b/file1.txt"}, matches)
	})

	t.Run("super asterisk", func(t *testing.T) {
		matches, err := Glob("a/**/*", WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/d/file1.txt",
			"./a/e/f/file1.txt",
			"./a/b/file1.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file1.txt",
			"a/d/file1.txt",
			"a/e/f/file1.txt",
		}, matches)
	})

	t.Run("alternative matchers", func(t *testing.T) {
		matches, err := Glob("a/{b,d}/file.txt", WithFs(testFs(t, []string{
			"a/b/file.txt",
			"a/c/file.txt",
			"a/d/file.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file.txt",
			"a/d/file.txt",
		}, matches)
	})

	t.Run("character list and range matchers", func(t *testing.T) {
		matches, err := Glob("[!bc]/[a-z]/file[01].txt", WithFs(testFs(t, []string{
			"a/b/file0.txt",
			"a/c/file1.txt",
			"a/d/file2.txt",
			"a/0/file.txt",
			"b/b/file0.txt",
			"c/b/file0.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file0.txt",
			"a/c/file1.txt",
		}, matches)
	})

	t.Run("nested matchers", func(t *testing.T) {
		matches, err := Glob("{a,[0-9]b}.txt", WithFs(testFs(t, []string{
			"a.txt",
			"b.txt",
			"1b.txt",
			"ab.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"1b.txt",
			"a.txt",
		}, matches)
	})

	t.Run("single symbol wildcards", func(t *testing.T) {
		matches, err := Glob("a?.txt", WithFs(testFs(t, []string{
			"a.txt",
			"a1.txt",
			"ab.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a1.txt",
			"ab.txt",
		}, matches)
	})

	t.Run("direct match", func(t *testing.T) {
		matches, err := Glob("a/b/c", WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/b/c",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{"a/b/c"}, matches)
	})

	t.Run("direct match wildcard", func(t *testing.T) {
		matches, err := Glob(QuoteMeta("a/b/c{a"), WithFs(testFs(t, []string{
			"./a/nope.txt",
			QuoteMeta("a/b/c{a"),
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{"a/b/c\\a{"}, matches)
	})

	t.Run("direct no match", func(t *testing.T) {
		matches, err := Glob("a/b/d", WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/b/dc",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{}, matches)
	})

	t.Run("direct no match escaped wildcards", func(t *testing.T) {
		matches, err := Glob(QuoteMeta("a/b/c{a"), WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/b/dc",
		}, nil)))
		require.EqualError(t, err, "file does not exist")
		require.Empty(t, matches)
	})

	t.Run("no matches", func(t *testing.T) {
		matches, err := Glob("z/*", WithFs(testFs(t, []string{
			"./a/nope.txt",
		}, nil)))
		require.NoError(t, err)
		require.Empty(t, matches)
	})

	t.Run("empty folder", func(t *testing.T) {
		matches, err := Glob("a*", WithFs(testFs(t, nil, nil)))
		require.NoError(t, err)
		require.Empty(t, matches)
	})

	t.Run("escaped asterisk", func(t *testing.T) {
		matches, err := Glob("a/\\*/b", WithFs(testFs(t, []string{
			"a/a/b",
			"a/*/b",
			"a/**/b",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/*/b",
		}, matches)
	})

	t.Run("escaped curly braces", func(t *testing.T) {
		matches, err := Glob("\\{a,b\\}/c", WithFs(testFs(t, []string{
			"a/c",
			"b/c",
			"{a,b}/c",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"{a,b}/c",
		}, matches)
	})

	t.Run("invalid pattern", func(t *testing.T) {
		matches, err := Glob("[*", WithFs(testFs(t, nil, nil)))
		require.EqualError(t, err, "compile glob pattern: unexpected end of input")
		require.Empty(t, matches)
	})

	t.Run("prefix is a file", func(t *testing.T) {
		matches, err := Glob("ab/c/*", WithFs(testFs(t, []string{
			"ab/c",
			"ab/d",
			"ab/e",
		}, nil)))
		require.NoError(t, err)
		require.Empty(t, matches)
	})

	t.Run("match files in directories", func(t *testing.T) {
		matches, err := Glob("/a/{b,c}", WithFs(testFs(t, []string{
			"/a/b/d",
			"/a/b/e/f",
			"/a/c",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"/a/b/d",
			"/a/b/e/f",
			"/a/c",
		}, matches)
	})

	t.Run("match directories directly", func(t *testing.T) {
		matches, err := Glob("/a/{b,c}", MatchDirectories(true), WithFs(testFs(t, []string{
			"/a/b/d",
			"/a/b/e/f",
			"/a/c",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"/a/b",
			"/a/c",
		}, matches)
	})

	t.Run("match empty directory", func(t *testing.T) {
		matches, err := Glob("/a/{b,c}", MatchDirectories(true), WithFs(testFs(t, []string{
			"/a/b",
		}, []string{
			"/a/c",
		})))
		require.NoError(t, err)
		require.Equal(t, []string{
			"/a/b",
			"/a/c",
		}, matches)
	})

	t.Run("pattern ending with star and subdir", func(t *testing.T) {
		matches, err := Glob("a/*", WithFs(testFs(t, []string{
			"./a/1.txt",
			"./a/2.txt",
			"./a/3.txt",
			"./a/b/4.txt",
		}, []string{
			"a",
			"a/b",
		})))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/1.txt",
			"a/2.txt",
			"a/3.txt",
			"a/b/4.txt",
		}, matches)
	})
}

func TestQuoteMeta(t *testing.T) {
	matches, err := Glob(QuoteMeta("{a,b}/c"), WithFs(testFs(t, []string{
		"a/c",
		"b/c",
		"{a,b}/c",
	}, nil)))
	require.NoError(t, err)
	require.Equal(t, []string{
		"{a,b}/c",
	}, matches)
}

func testFs(tb testing.TB, files, dirs []string) afero.Fs {
	tb.Helper()

	fs := afero.NewMemMapFs()

	for _, file := range files {
		if _, err := fs.Create(filepath.FromSlash(file)); err != nil {
			require.NoError(tb, err)
		}
	}

	for _, dir := range dirs {
		if err := fs.MkdirAll(filepath.FromSlash(dir), 0664); err != nil {
			require.NoError(tb, err)
		}
	}

	return fs
}
