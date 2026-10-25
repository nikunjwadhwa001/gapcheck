package analysis

import (
	"errors"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	ok := strings.Repeat("a", MinTextLen)

	tests := []struct {
		name      string
		req       Request
		wantField string // "" means valid
	}{
		{"valid", Request{ResumeText: ok, JDText: ok}, ""},
		{"missing resume", Request{ResumeText: "", JDText: ok}, "resume_text"},
		{"short resume", Request{ResumeText: "Go dev", JDText: ok}, "resume_text"},
		{"long resume", Request{ResumeText: strings.Repeat("a", MaxTextLen+1), JDText: ok}, "resume_text"},
		{"missing jd", Request{ResumeText: ok, JDText: ""}, "jd_text"},
		// Limits count characters, not bytes: 50 Hindi letters are 150 bytes.
		{"multi-byte at minimum", Request{ResumeText: strings.Repeat("क", MinTextLen), JDText: ok}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantField == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("got %v, want ValidationError", err)
			}
			if ve.Field != tt.wantField {
				t.Fatalf("field = %q, want %q", ve.Field, tt.wantField)
			}
		})
	}
}
