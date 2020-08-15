package wdav

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathSplit(t *testing.T) {
	data := []struct {
		path  string
		parts []string
	}{
		{"path/to/file", []string{"path", "to", "file"}},
		{"/path/to/file", []string{"path", "to", "file"}},
		{"/path/to'a file with spaces /file", []string{"path", "to'a file with spaces ", "file"}},
		{"", []string{}},
		{"/", []string{}},
	}
	for i, d := range data {
		actual := splitPath(d.path)
		assert.Equal(t, d.parts, actual, "#%d input='%s'", i, d.path)
	}
}

func TestSplitDirfromName(t *testing.T) {
	data := []struct {
		path []string
		dir  []string
		name string
	}{
		{[]string{"path", "to", "file"}, []string{"path", "to"}, "file"},
		{[]string{"file"}, []string{}, "file"},
		{[]string{}, []string{}, ""},
	}
	for i, d := range data {
		dir, name := splitDirFromName(d.path)
		if name != d.name {
			t.Errorf("#%d: Bad name returned: expected '%s', got '%s'", i, d.name, name)
		}
		if fmt.Sprintf("%v", dir) != fmt.Sprintf("%v", d.dir) {
			t.Errorf("#%d: Bad path returned: expected '%v', got '%v'", i, d.dir, dir)
		}
	}
}
