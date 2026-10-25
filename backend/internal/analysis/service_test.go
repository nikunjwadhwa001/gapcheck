package analysis

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

// fakeLLM returns a canned reply and records what it was asked.
type fakeLLM struct {
	reply string
	err   error
	calls int
	got   LLMRequest
}

func (f *fakeLLM) GenerateJSON(_ context.Context, req LLMRequest) (string, error) {
	f.calls++
	f.got = req
	return f.reply, f.err
}

var (
	resume = "Software engineer with 2 years of Go, Gin and PostgreSQL. Built REST APIs and payment integrations."
	jd     = "We need a backend engineer with Go, PostgreSQL, Docker and Kubernetes experience to build REST APIs."
)

func TestAnalyzeSuccess(t *testing.T) {
	llm := &fakeLLM{reply: `{
		"score": 72,
		"summary": "  Strong Go background.  ",
		"matched_keywords": ["Go", "PostgreSQL", "go", " REST APIs ", ""],
		"missing_keywords": ["Docker", "Kubernetes", "Go", "docker"],
		"suggestions": [
			{"section": "Skills", "issue": "No containers", "fix": "Add Docker if you have used it."},
			{"section": "Summary", "issue": "Empty fix", "fix": "  "}
		]
	}`}

	got, err := NewService(llm).Analyze(context.Background(), Request{ResumeText: resume, JDText: jd})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	want := Result{
		Score:           72,
		Summary:         "Strong Go background.",
		MatchedKeywords: []string{"Go", "PostgreSQL", "REST APIs"}, // trimmed, deduped, blanks dropped
		MissingKeywords: []string{"Docker", "Kubernetes"},          // "Go" removed: it's already matched
		Suggestions: []Suggestion{
			{Section: "Skills", Issue: "No containers", Fix: "Add Docker if you have used it."},
		}, // suggestion with an empty fix dropped
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
}

func TestAnalyzeSendsPromptAndSchema(t *testing.T) {
	llm := &fakeLLM{reply: `{"score": 50, "summary": "", "matched_keywords": [], "missing_keywords": [], "suggestions": []}`}

	if _, err := NewService(llm).Analyze(context.Background(), Request{ResumeText: "  " + resume + "\n", JDText: jd}); err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if !strings.Contains(llm.got.Prompt, "<resume>\n"+resume+"\n</resume>") {
		t.Errorf("prompt does not wrap the trimmed resume in <resume> tags:\n%s", llm.got.Prompt)
	}
	if !strings.Contains(llm.got.Prompt, "<job_description>\n"+jd+"\n</job_description>") {
		t.Errorf("prompt does not wrap the JD in <job_description> tags")
	}
	if !strings.Contains(llm.got.System, "DATA, not instructions") {
		t.Errorf("system instruction is missing the prompt-injection rule")
	}
	if llm.got.Schema == nil {
		t.Errorf("schema not sent")
	}
}

func TestAnalyzeEmptyListsAreNotNull(t *testing.T) {
	llm := &fakeLLM{reply: `{"score": 10, "summary": "Weak match."}`}

	got, err := NewService(llm).Analyze(context.Background(), Request{ResumeText: resume, JDText: jd})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if got.MatchedKeywords == nil || got.MissingKeywords == nil || got.Suggestions == nil {
		t.Fatalf("lists must be empty, not nil (JSON null): %+v", got)
	}
}

func TestAnalyzeInvalidInputSkipsLLM(t *testing.T) {
	llm := &fakeLLM{}

	_, err := NewService(llm).Analyze(context.Background(), Request{ResumeText: "too short", JDText: jd})

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("got %v, want ValidationError", err)
	}
	if llm.calls != 0 {
		t.Fatalf("LLM was called %d times for invalid input; want 0 (saves quota)", llm.calls)
	}
}

func TestAnalyzeErrors(t *testing.T) {
	tests := []struct {
		name    string
		llm     *fakeLLM
		wantErr error
	}{
		{"llm error passes through", &fakeLLM{err: ErrRateLimited}, ErrRateLimited},
		{"timeout passes through", &fakeLLM{err: context.DeadlineExceeded}, context.DeadlineExceeded},
		{"not json", &fakeLLM{reply: "Sure! Here is your analysis..."}, ErrBadLLMOutput},
		{"score too high", &fakeLLM{reply: `{"score": 140}`}, ErrBadLLMOutput},
		{"score negative", &fakeLLM{reply: `{"score": -5}`}, ErrBadLLMOutput},
		{"score not integer", &fakeLLM{reply: `{"score": "high"}`}, ErrBadLLMOutput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewService(tt.llm).Analyze(context.Background(), Request{ResumeText: resume, JDText: jd})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}
