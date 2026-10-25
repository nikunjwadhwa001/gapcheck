// Package gemini calls Google's Gemini REST API. It implements
// analysis.LLM, so the rest of the app never depends on Gemini directly.
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nikunjwadhwa001/gapcheck/backend/internal/analysis"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// maxResponseBytes caps how much of Gemini's reply we read into memory.
const maxResponseBytes = 1 << 20 // 1 MB

type Client struct {
	apiKey  string
	model   string
	baseURL string
	http    *http.Client
}

// New returns a client for the given model, e.g. "gemini-3.5-flash-lite".
// Timeouts come from the context passed to each call.
func New(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultBaseURL,
		http:    &http.Client{},
	}
}

// Wire format for models.generateContent (only the fields we use).

type part struct {
	Text string `json:"text"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type generationConfig struct {
	Temperature      float64        `json:"temperature"`
	ResponseMimeType string         `json:"responseMimeType"`
	ResponseSchema   map[string]any `json:"responseSchema,omitempty"`
}

type generateRequest struct {
	SystemInstruction *content         `json:"systemInstruction,omitempty"`
	Contents          []content        `json:"contents"`
	GenerationConfig  generationConfig `json:"generationConfig"`
}

type generateResponse struct {
	Candidates []struct {
		Content      content `json:"content"`
		FinishReason string  `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback *struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// GenerateJSON sends one prompt and returns the model's JSON text.
func (c *Client) GenerateJSON(ctx context.Context, req analysis.LLMRequest) (string, error) {
	body, err := json.Marshal(generateRequest{
		SystemInstruction: &content{Parts: []part{{Text: req.System}}},
		Contents:          []content{{Role: "user", Parts: []part{{Text: req.Prompt}}}},
		GenerationConfig: generationConfig{
			Temperature:      0.2, // low = consistent scores for the same input
			ResponseMimeType: "application/json",
			ResponseSchema:   req.Schema,
		},
	})
	if err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent", c.baseURL, c.model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// Header, not query string, so the key never shows up in URL logs.
	httpReq.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("call gemini: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var e errorResponse
		_ = json.Unmarshal(data, &e) // best effort; message may be empty
		if resp.StatusCode == http.StatusTooManyRequests {
			return "", fmt.Errorf("%w: %s", analysis.ErrRateLimited, e.Error.Message)
		}
		return "", fmt.Errorf("gemini returned %d %s: %s", resp.StatusCode, e.Error.Status, e.Error.Message)
	}

	var gr generateResponse
	if err := json.Unmarshal(data, &gr); err != nil {
		return "", fmt.Errorf("decode gemini response: %w", err)
	}
	if gr.PromptFeedback != nil && gr.PromptFeedback.BlockReason != "" {
		return "", fmt.Errorf("gemini blocked the prompt: %s", gr.PromptFeedback.BlockReason)
	}
	if len(gr.Candidates) == 0 {
		return "", fmt.Errorf("gemini returned no candidates")
	}

	cand := gr.Candidates[0]
	// STOP is a normal finish. Anything else (MAX_TOKENS, SAFETY, ...) means
	// the JSON is probably cut off or missing.
	if cand.FinishReason != "" && cand.FinishReason != "STOP" {
		return "", fmt.Errorf("gemini stopped early: %s", cand.FinishReason)
	}

	var text strings.Builder
	for _, p := range cand.Content.Parts {
		text.WriteString(p.Text)
	}
	if text.Len() == 0 {
		return "", fmt.Errorf("gemini returned an empty reply")
	}
	return text.String(), nil
}
