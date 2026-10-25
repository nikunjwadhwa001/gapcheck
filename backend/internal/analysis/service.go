package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// LLM is anything that can answer a prompt with JSON matching a schema.
// The real implementation is the Gemini client; tests use a fake.
type LLM interface {
	GenerateJSON(ctx context.Context, req LLMRequest) (string, error)
}

// LLMRequest is everything an LLM needs for one call.
type LLMRequest struct {
	System string         // role and rules for the model
	Prompt string         // the user's resume and job description
	Schema map[string]any // the JSON shape the reply must follow
}

var (
	// ErrRateLimited means the LLM provider refused because of quota.
	ErrRateLimited = errors.New("llm rate limit reached")
	// ErrBadLLMOutput means the model replied, but not with usable JSON.
	ErrBadLLMOutput = errors.New("llm returned an unusable response")
)

// Service runs an analysis: validate → prompt → LLM → parse → clean up.
type Service struct {
	llm LLM
}

func NewService(llm LLM) *Service {
	return &Service{llm: llm}
}

// Analyze compares the resume with the job description.
func (s *Service) Analyze(ctx context.Context, req Request) (Result, error) {
	req = req.normalized()
	if err := req.Validate(); err != nil {
		return Result{}, err
	}

	raw, err := s.llm.GenerateJSON(ctx, LLMRequest{
		System: systemInstruction,
		Prompt: buildPrompt(req),
		Schema: resultSchema,
	})
	if err != nil {
		return Result{}, fmt.Errorf("llm call: %w", err)
	}

	var res Result
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrBadLLMOutput, err)
	}
	// Schemas are a strong hint, not a guarantee, so check the important parts.
	if res.Score < 0 || res.Score > 100 {
		return Result{}, fmt.Errorf("%w: score %d is outside 0-100", ErrBadLLMOutput, res.Score)
	}
	return clean(res), nil
}

// clean tidies the model's lists so the UI can trust them.
func clean(res Result) Result {
	res.Summary = strings.TrimSpace(res.Summary)
	res.MatchedKeywords = dedupe(res.MatchedKeywords, nil)
	// A keyword can't be both matched and missing; matched wins.
	res.MissingKeywords = dedupe(res.MissingKeywords, res.MatchedKeywords)

	suggestions := make([]Suggestion, 0, len(res.Suggestions))
	for _, s := range res.Suggestions {
		s.Section = strings.TrimSpace(s.Section)
		s.Issue = strings.TrimSpace(s.Issue)
		s.Fix = strings.TrimSpace(s.Fix)
		if s.Fix != "" {
			suggestions = append(suggestions, s)
		}
	}
	res.Suggestions = suggestions
	return res
}

// dedupe trims items and drops blanks, case-insensitive repeats, and
// anything in exclude. It never returns nil, so JSON shows [] not null.
func dedupe(items, exclude []string) []string {
	seen := make(map[string]bool, len(items)+len(exclude))
	for _, e := range exclude {
		seen[strings.ToLower(strings.TrimSpace(e))] = true
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		it = strings.TrimSpace(it)
		key := strings.ToLower(it)
		if it == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, it)
	}
	return out
}
