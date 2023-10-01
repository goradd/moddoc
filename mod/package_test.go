package mod

import (
	"reflect"
	"testing"
)

func Test_parseCommentFlags(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		wantNewText string
		wantFlags   map[string]string
	}{
		{"empty", "", "", nil},
		{"noflags", "abc", "abc", nil},
		{"noflags 2 lines", "abc\njkl", "abc\njkl", nil},
		{"bool flag", "doc: abc", "", map[string]string{"abc": ""}},
		{"val flag", "doc: abc=def", "", map[string]string{"abc": "def"}},
		{"2 lines", "doc: abc=def\ndoc:ghi", "", map[string]string{"abc": "def", "ghi": ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewText, gotFlags := parseCommentFlags(tt.text)
			if gotNewText != tt.wantNewText {
				t.Errorf("parseCommentFlags() gotNewText = %v, want %v", gotNewText, tt.wantNewText)
			}
			if !reflect.DeepEqual(gotFlags, tt.wantFlags) {
				t.Errorf("parseCommentFlags() gotFlags = %v, want %v", gotFlags, tt.wantFlags)
			}
		})
	}
}
