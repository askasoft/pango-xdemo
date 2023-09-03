package txts

import (
	"fmt"
	"io/fs"
	"testing"
)

func TestEmbedFS(t *testing.T) {
	fmt.Println("------------------------")
	fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
		fmt.Println(path, d.IsDir())
		return nil
	})
	fmt.Println("------------------------")
}
