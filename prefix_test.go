package fileglob

import (
	"testing"

	"github.com/gobwas/glob/syntax/ast"
	"github.com/gobwas/glob/syntax/lexer"
	"github.com/matryer/is"
)

func TestJoinParentPattern(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		prefix   string
		relative string
		pattern  string
	}{
		{"UNC wildcard", "//host/share", "*.txt", "//host/share/*.txt"},
		{"UNC empty suffix", "//host/share", "", "//host/share"},
		{"UNC trailing separator", "//host/share/", "*.txt", "//host/share/*.txt"},
		{"drive root empty suffix", "C:/", "", "C:/"},
		{"drive root wildcard", "C:/", "*.txt", "C:/*.txt"},
		{"Unix root empty suffix", "/", "", "/"},
		{"Unix root wildcard", "/", "*.txt", "/*.txt"},
		{"escaped UNC path", `//host/share/build\[1\]`, `file\[1\].txt`, `//host/share/build\[1\]/file\[1\].txt`},
		{"escaped drive path", `C:/build\[1\]`, "file?.txt", `C:/build\[1\]/file?.txt`},
		{"escaped Unix path", `/build\[1\]`, "*.txt", `/build\[1\]/*.txt`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			is.New(t).Equal(tc.pattern, joinParentPattern(tc.prefix, tc.relative))
		})
	}
}

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
		{`/build\[1\]/*.txt`, "/build[1]"},
		{`/build\[1\]/file.txt`, "/build[1]/file.txt"},
		{`/build\[1\]/file\[1\].txt`, "/build[1]/file[1].txt"},
		{`/build\\1/*.txt`, `/build\1`},
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
