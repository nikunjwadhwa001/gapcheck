// Package analysis compares a resume with a job description using an LLM
// and turns the model's reply into structured, validated feedback.
package analysis

// Request is the body the client sends to POST /api/analyze.
type Request struct {
	ResumeText string `json:"resume_text"`
	JDText     string `json:"jd_text"`
}

// Suggestion is one concrete improvement for the resume.
type Suggestion struct {
	Section string `json:"section"` // e.g. "Experience", "Skills", "Summary"
	Issue   string `json:"issue"`   // what is weak or missing
	Fix     string `json:"fix"`     // what to change, specifically
}

// Result is what /api/analyze returns on success.
type Result struct {
	Score           int          `json:"score"` // 0–100 match with the job description
	Summary         string       `json:"summary"`
	MatchedKeywords []string     `json:"matched_keywords"`
	MissingKeywords []string     `json:"missing_keywords"`
	Suggestions     []Suggestion `json:"suggestions"`
}
