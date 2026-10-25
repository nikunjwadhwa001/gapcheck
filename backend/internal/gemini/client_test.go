package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nikunjwadhwa001/gapcheck/backend/internal/analysis"
)

// newTestClient points a client at a fake Gemini server.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New("test-key", "test-model")
	c.baseURL = srv.URL
	return c
}

var req = analysis.LLMRequest{
	System: "be a recruiter",
	Prompt: "compare these",
	Schema: map[string]any{"type": "OBJECT"},
}

func TestGenerateJSONSendsCorrectRequest(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/models/test-model:generateContent" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("x-goog-api-key"); got != "test-key" {
			t.Errorf("api key header = %q", got)
		}
		if r.URL.RawQuery != "" {
			t.Errorf("query string should be empty (key must not be in the URL), got %q", r.URL.RawQuery)
		}

		var body generateRequest
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("bad request body: %v", err)
		}
		if body.SystemInstruction.Parts[0].Text != "be a recruiter" {
			t.Errorf("system instruction not sent")
		}
		if body.Contents[0].Parts[0].Text != "compare these" {
			t.Errorf("prompt not sent")
		}
		if body.GenerationConfig.ResponseMimeType != "application/json" {
			t.Errorf("responseMimeType = %q", body.GenerationConfig.ResponseMimeType)
		}
		if body.GenerationConfig.ResponseSchema["type"] != "OBJECT" {
			t.Errorf("schema not sent")
		}

		w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"score\":"},{"text":"80}"}]},"finishReason":"STOP"}]}`))
	})

	got, err := c.GenerateJSON(context.Background(), req)
	if err != nil {
		t.Fatalf("GenerateJSON: %v", err)
	}
	if got != `{"score":80}` {
		t.Fatalf("got %q; parts should be joined", got)
	}
}

func TestGenerateJSONErrors(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr error  // checked with errors.Is when set
		wantMsg string // substring of the error otherwise
	}{
		{"rate limited", 429, `{"error":{"message":"quota exceeded","status":"RESOURCE_EXHAUSTED"}}`, analysis.ErrRateLimited, ""},
		{"bad key", 400, `{"error":{"message":"API key not valid","status":"INVALID_ARGUMENT"}}`, nil, "API key not valid"},
		{"server error", 500, `oops`, nil, "500"},
		{"blocked prompt", 200, `{"promptFeedback":{"blockReason":"SAFETY"}}`, nil, "blocked"},
		{"no candidates", 200, `{"candidates":[]}`, nil, "no candidates"},
		{"cut off", 200, `{"candidates":[{"content":{"parts":[{"text":"{\"sco"}]},"finishReason":"MAX_TOKENS"}]}`, nil, "MAX_TOKENS"},
		{"empty reply", 200, `{"candidates":[{"content":{"parts":[]},"finishReason":"STOP"}]}`, nil, "empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			})
			_, err := c.GenerateJSON(context.Background(), req)
			if err == nil {
				t.Fatal("expected an error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
			if tt.wantMsg != "" && !strings.Contains(err.Error(), tt.wantMsg) {
				t.Fatalf("error %q does not mention %q", err, tt.wantMsg)
			}
		})
	}
}

func TestGenerateJSONRespectsContextTimeout(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.GenerateJSON(ctx, req)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want context.DeadlineExceeded", err)
	}
}
