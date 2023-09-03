package tpls

import (
	"fmt"
	"io/fs"
	"testing"
)

func TestEmbedFS(t *testing.T) {
	fmt.Println("------------------------")
	fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
		fmt.Println(path)
		return nil
	})
	fmt.Println("------------------------")
}
