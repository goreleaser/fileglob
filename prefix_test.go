package fileglob

import "testing"

func TestStaticPrefix(t *testing.T) {
	var testCases = []struct {
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
	}

	for _, testCase := range testCases {
		prefix, err := staticPrefix(testCase.pattern)
		if err != nil {
			t.Errorf("staticPrefix: %v", err)
		}
		if prefix != testCase.prefix {
			t.Errorf("prefixPrefix returned %q instead of %q for pattern %q",
				prefix, testCase.prefix, testCase.pattern)
		}
	}
}
