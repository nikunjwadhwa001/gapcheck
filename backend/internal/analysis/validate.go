package analysis

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Length limits, in characters, for each text field. Too short can't be
// analysed meaningfully; too long wastes LLM quota (~20k chars ≈ 8 pages).
const (
	MinTextLen = 50
	MaxTextLen = 20000
)

// ValidationError means the client sent bad input. The handler turns it
// into HTTP 400 and shows Field and Msg to the user.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return e.Field + " " + e.Msg
}

// normalized returns a copy with surrounding whitespace removed.
func (r Request) normalized() Request {
	return Request{
		ResumeText: strings.TrimSpace(r.ResumeText),
		JDText:     strings.TrimSpace(r.JDText),
	}
}

// Validate checks both fields are present and within the length limits.
func (r Request) Validate() error {
	if err := checkText("resume_text", r.ResumeText); err != nil {
		return err
	}
	return checkText("jd_text", r.JDText)
}

func checkText(field, text string) error {
	n := utf8.RuneCountInString(text)
	switch {
	case n == 0:
		return &ValidationError{Field: field, Msg: "is required"}
	case n < MinTextLen:
		return &ValidationError{Field: field, Msg: fmt.Sprintf("is too short (minimum %d characters)", MinTextLen)}
	case n > MaxTextLen:
		return &ValidationError{Field: field, Msg: fmt.Sprintf("is too long (maximum %d characters)", MaxTextLen)}
	}
	return nil
}
