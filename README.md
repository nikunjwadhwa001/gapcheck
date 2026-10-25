# GapCheck

Score your resume against a job description. GapCheck highlights missing keywords and suggests targeted improvements, using LLM-based analysis that returns structured feedback rather than free-form text.

> Status: early development (Phase 1 — analyze API).

## Stack

| Part | Tech | Hosting (free tier) |
|---|---|---|
| App | Flutter (web + Android) | GitHub Pages |
| API | Go + Gin | Render |
| LLM | Google Gemini API | — |

The Gemini API key lives only on the backend. The web app never sees it.

## Repo layout

```
backend/   Go API (cmd/server entrypoint, internal/ packages)
app/       Flutter client
.github/   CI workflows
docs/      Phase-by-phase write-ups of what was built and why
```

## Running locally

**Backend**

```bash
cd backend
cp .env.example .env   # add your free Gemini key
go run ./cmd/server    # http://localhost:8080/health
go test ./...
```

**Try the API**

```bash
curl -s -X POST localhost:8080/api/analyze \
  -H 'Content-Type: application/json' \
  -d '{"resume_text": "<paste resume>", "jd_text": "<paste job description>"}'
```

Returns `score`, `summary`, `matched_keywords`, `missing_keywords` and `suggestions`. Errors always look like `{"error": {"code": "...", "message": "..."}}`.

**App**

```bash
cd app
flutter pub get
flutter run -d chrome
flutter test
```

## Roadmap

- [x] Phase 0 — Repo scaffolding and CI
- [x] Phase 1 — `/api/analyze`: resume + JD → structured score, keywords, suggestions
- [ ] Phase 2 — Flutter UI (input, results)
- [ ] Phase 3 — Deploy (GitHub Pages + Render)
- [ ] Phase 4 — PDF / DOCX resume upload
- [ ] Phase 5 — Deterministic keyword score, rate limiting, caching

## License

[MIT](LICENSE)
