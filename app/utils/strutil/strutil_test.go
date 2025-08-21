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
