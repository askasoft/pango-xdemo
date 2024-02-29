package web

import (
	"fmt"
	"io/fs"
	"testing"
)

func TestEmbedFS_Statics(t *testing.T) {
	for k, v := range Statics {
		fmt.Println("------", k, "------------------")
		fs.WalkDir(v, ".", func(path string, d fs.DirEntry, err error) error {
			fmt.Println(path)
			return nil
		})
		fmt.Println("------------------------")
	}
}

func TestEmbedFS_Assets(t *testing.T) {
	fmt.Println("------------------------")
	fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
		fmt.Println(path)
		return nil
	})
	fmt.Println("------------------------")
}
