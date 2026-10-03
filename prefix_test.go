package fileglob

import (
	"testing"

	"github.com/gobwas/glob/syntax/ast"
	"github.com/gobwas/glob/syntax/lexer"
	"github.com/matryer/is"
)

func TestStaticPrefix(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		pattern string
		prefix  string
	}{
		{"/foo/b*ar/baz", "/foo"},
		{"foo/bar", "foo/bar"},
		{"/foo/bar/{b,p}az", "/foo/bar"},
		{"*/foo", "."},
		{"./", "."},
		{"fo\\*o/bar/b*z", "fo*o/bar"},
		{"/\\{foo\\}/bar", "/{foo}/bar"},
		{"C:/Path/To/Some/File", "C:/Path/To/Some/File"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.pattern, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			prefix, err := staticPrefix(testCase.pattern)
			is.NoErr(err)
			is.Equal(testCase.prefix, prefix)
		})
	}
}

func TestContainsMatchers(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		pattern          string
		containsMatchers bool
	}{
		{"/a/*/b", true},
		{"\\{a\\}/\\*/", false},
		{"a/b/c", false},
		{"", false},
		{"\\*/\\?", false},
		{"*/\\?", true},
		{"\\{*\\}", true},
		{"{a,b}/c", true},
		{"\\{\\\\[a-z]", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.pattern, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			_, err := ast.Parse(lexer.NewLexer(testCase.pattern))
			is.NoErr(err)
			is.Equal(testCase.containsMatchers, ContainsMatchers(testCase.pattern))
		})
	}
}

func TestValidPattern(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		pattern string
		valid   bool
	}{
		{"/a/*/b", true},
		{"{a[", false},
		{"[*]", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.pattern, func(t *testing.T) {
			t.Parallel()
			is.New(t).Equal(testCase.valid, ValidPattern(testCase.pattern) == nil)
		})
	}
}

func TestMaxPathSeparators(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		pattern string
		count   int
	}{
		{"", 0},
		{".", 0},
		{"blocked", 0},
		{"*.txt", 0},
		{"a/?/file.txt", 2},
		{"a/*/*.txt", 2},
		{"**", -1},
		{"a/**/file.txt", -1},
		{`a/\*\*/file.txt`, 2},
		{`a\/file.txt`, 1},
		{"{a,b/c}", 1},
		{"{a/b,c/d}.txt", 1},
		{"{a,{b/c,d/e/f}}", 2},
		{"{a/b,c}/{d,e/f}", 3},
		{"{a,{b/**,c}}", -1},
		{"a[/]b", 1},
		{"[!/]file.txt", 0},
		{"[abc]", 0},
		{"[!abc]", 1},
		{"[.-0]", 1},
		{"[!.-0]", 0},
		{"[a-z]", 0},
		{"[!a-z]", 1},
	} {
		t.Run(tc.pattern, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			node, err := ast.Parse(lexer.NewLexer(tc.pattern))
			is.NoErr(err)
			is.Equal(tc.count, maxPathSeparators(node))
		})
	}
}
