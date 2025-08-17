package strutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/askasoft/pango/fsu"
)

func testFilename(name string) string {
	return filepath.Join("testdata", name)
}

func testReadFile(t *testing.T, name string) []byte {
	fn := testFilename(name)
	bs, err := fsu.ReadFile(fn)
	if err != nil {
		t.Fatalf("Failed to read file %q: %v", fn, err)
	}
	return bs
}

func TestEllipsiz(t *testing.T) {
	cs := []string{"ellipsiz"}

	for i, c := range cs {
		s := string(testReadFile(t, c+".org.txt"))
		w := string(testReadFile(t, c+".exp.txt"))
		a := Ellipsiz(s, 50)

		if w != a {
			t.Errorf("[%d] Ellipsiz(%q):\n  GOT: %q\n WANT: %q\n", i, c, a, w)
			fsu.WriteString(testFilename(c+".out"), a, fsu.FileMode(0660))
		} else {
			os.Remove(testFilename(c + ".out"))
		}
	}
}

func TestNextKeyword(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantKey    string
		wantRest   string
		wantQuoted bool
	}{
		{"Empty input", "", "", "", false},
		{"Only spaces", "    ", "", "", false},
		{"Single word", "hello", "hello", "", false},
		{"Multiple words", "hello world", "hello", " world", false},
		{"Quoted word", `"hello world" test`, "hello world", " test", true},
		{"Quoted no close", `"hello world`, `"hello`, " world", false},
		{"No space", "nowordboundary", "nowordboundary", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, rest, quoted := NextKeyword(tt.input)
			if tt.wantKey != key || tt.wantRest != rest || tt.wantQuoted != quoted {
				t.Fatalf("NextKeyword(%q) = (%q, %q, %v), want (%q, %q, %v)", tt.input,
					key, rest, quoted, tt.wantKey, tt.wantRest, tt.wantQuoted,
				)
			}
		})
	}
}
