package docutil

import (
	"bytes"

	"github.com/askasoft/pango/doc/htmlx"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/wcu"
	"golang.org/x/net/html"
)

const (
	CharsetDetectLength = 4096
)

func ParseHTMLFile(name string, charsets ...string) (*html.Node, error) {
	return htmlx.ParseHTMLFile(name, CharsetDetectLength, charsets...)
}

func DetectAndReadFile(filename string, charsets ...string) ([]byte, string, error) {
	return wcu.DetectAndReadFile(filename, CharsetDetectLength, charsets...)
}

func ReadTextFromTextData(data []byte) string {
	r, _, err := wcu.DetectAndTransform(bytes.NewReader(data), CharsetDetectLength, false)
	if err == nil {
		bs, err := iox.ReadAll(r)
		if err == nil {
			return str.UnsafeString(bs)
		}
	}
	return str.UnsafeString(data)
}

func ReadTextFromTextFile(filename string, charset string) (string, error) {
	bs, _, err := DetectAndReadFile(filename, charset)
	if err != nil {
		return "", err
	}
	return str.UnsafeString(bs), err
}
