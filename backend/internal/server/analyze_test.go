package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikunjwadhwa001/gapcheck/backend/internal/analysis"
)

// fakeAnalyzer returns whatever the test sets.
type fakeAnalyzer struct {
	res analysis.Result
	err error
}

func (f fakeAnalyzer) Analyze(context.Context, analysis.Request) (analysis.Result, error) {
	return f.res, f.err
}

func post(t *testing.T, an Analyzer, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := NewRouter(Config{Analyzer: an, LLMTimeout: time.Second})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

const validBody = `{"resume_text": "resume", "jd_text": "job"}`

func TestAnalyzeOK(t *testing.T) {
	want := analysis.Result{
		Score:           81,
		Summary:         "Good fit.",
		MatchedKeywords: []string{"Go"},
		MissingKeywords: []string{"Docker"},
		Suggestions:     []analysis.Suggestion{{Section: "Skills", Issue: "x", Fix: "y"}},
	}
	w := post(t, fakeAnalyzer{res: want}, validBody)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	var got analysis.Result
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if got.Score != 81 || got.MissingKeywords[0] != "Docker" {
		t.Fatalf("got %+v", got)
	}
}

func TestAnalyzeErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		err      error
		wantCode int
		wantErr  string
	}{
		{"not json", `resume please`, nil, 400, "invalid_json"},
		{"validation", validBody, &analysis.ValidationError{Field: "jd_text", Msg: "is required"}, 400, "invalid_input"},
		{"timeout", validBody, fmt.Errorf("llm call: %w", context.DeadlineExceeded), 504, "timeout"},
		{"rate limited", validBody, fmt.Errorf("llm call: %w", analysis.ErrRateLimited), 429, "rate_limited"},
		{"bad llm output", validBody, analysis.ErrBadLLMOutput, 502, "upstream_error"},
		{"other llm error", validBody, errors.New("gemini returned 500"), 502, "upstream_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := post(t, fakeAnalyzer{err: tt.err}, tt.body)
			if w.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d (body %s)", w.Code, tt.wantCode, w.Body)
			}
			var got struct {
				Error apiError `json:"error"`
			}
			json.Unmarshal(w.Body.Bytes(), &got)
			if got.Error.Code != tt.wantErr {
				t.Fatalf("error code = %q, want %q", got.Error.Code, tt.wantErr)
			}
			if got.Error.Message == "" {
				t.Fatalf("error message is empty")
			}
		})
	}
}

func TestAnalyzeValidationErrorNamesField(t *testing.T) {
	w := post(t, fakeAnalyzer{err: &analysis.ValidationError{Field: "resume_text", Msg: "is too short"}}, validBody)
	if !strings.Contains(w.Body.String(), `"field":"resume_text"`) {
		t.Fatalf("body = %s", w.Body)
	}
}

func TestAnalyzeInternalDetailsNotLeaked(t *testing.T) {
	w := post(t, fakeAnalyzer{err: errors.New("gemini returned 400: API key AIza-secret not valid")}, validBody)
	if strings.Contains(w.Body.String(), "AIza") || strings.Contains(w.Body.String(), "gemini") {
		t.Fatalf("internal error leaked to client: %s", w.Body)
	}
}

func TestAnalyzeBodyTooLarge(t *testing.T) {
	big := `{"resume_text": "` + strings.Repeat("a", maxBodyBytes) + `", "jd_text": "x"}`
	w := post(t, fakeAnalyzer{}, big)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", w.Code)
	}
}
