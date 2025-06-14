package args

import (
	"testing"
)

func TestKeywords_Contains(t *testing.T) {
	tests := []struct {
		name     string
		keywords Keywords
		input    string
		expected bool
	}{
		{"Empty list", Keywords{}, "go", false},
		{"Exact match", Keywords{"go", "lang"}, "go", true},
		{"Case-insensitive", Keywords{"GoLang"}, "golang", true},
		{"Partial match", Keywords{"gram"}, "programming", true},
		{"No match", Keywords{"code", "run"}, "build", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.keywords.Contains(tt.input)
			if a != tt.expected {
				t.Errorf("Keywords.Contains(%q) = %v, want %v", tt.input, a, tt.expected)
			}
		})
	}
}

func TestKeywords_ContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		keywords Keywords
		input    []string
		expected bool
	}{
		{"Empty keywords", Keywords{}, []string{"go"}, false},
		{"Empty input", Keywords{"go"}, []string{}, false},
		{"One matches", Keywords{"go", "lang"}, []string{"build", "go"}, true},
		{"Case-insensitive match", Keywords{"GoLang"}, []string{"run", "GOLANG"}, true},
		{"Partial match", Keywords{"gram"}, []string{"programming"}, true},
		{"No matches", Keywords{"run", "code"}, []string{"compile", "link"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.keywords.ContainsAny(tt.input...)
			if a != tt.expected {
				t.Errorf("Keywords.ContainsAny(%v) = %v, want %v", tt.input, a, tt.expected)
			}
		})
	}
}

func TestParseKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Keywords
	}{
		{"Empty input", "", Keywords{}},
		{"Single word", "hello", Keywords{"hello"}},
		{"Multiple words", "go lang", Keywords{"go", "lang"}},
		{"Quoted and unquoted", `"go lang" test`, Keywords{"go lang", "test"}},
		{"Mixed spacing", `   go    "hello world"   test  `, Keywords{"go", "hello world", "test"}},
		{"Incomplete quote", `"hello world`, Keywords{`"hello`, `world`}}, // because quote doesn't close
		{"Only spaces", "     ", Keywords{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := ParseKeywords(tt.input).String()
			w := Keywords(tt.expected).String()
			if w != a {
				t.Errorf("ParseKeywords(%v) = %q, want %q", tt.input, a, w)
			}
		})
	}
}
