# GapCheck

Score your resume against a job description. GapCheck highlights missing keywords and suggests targeted improvements, using LLM-based analysis that returns structured feedback rather than free-form text.

**Live:** http://meetoova.me/gapcheck/ · API: https://gapcheck-api.onrender.com/health

> Status: early development (Phase 3 — deployed). The free API server sleeps after 15 idle minutes, so the first analysis after a quiet spell can take up to a minute.

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
flutter run -d chrome --web-port 3000   # the API allows this origin via ALLOWED_ORIGINS
flutter test
```

The app calls `http://localhost:8080` by default. Point it elsewhere with `--dart-define=API_BASE_URL=https://...`. Port 3000 rather than 5000, because macOS AirPlay Receiver already uses 5000.

## Deployment

- **Web app:** `.github/workflows/pages.yml` tests, builds and publishes `app/` to GitHub Pages on every push to `main` that changes `app/**`. The API URL comes from the `API_BASE_URL` repository variable (default `https://gapcheck-api.onrender.com`).
- **API:** `render.yaml` defines a free Render web service. Render deploys `backend/` after CI passes. `GEMINI_API_KEY` is set in Render's dashboard, never in the repo.

## Roadmap

- [x] Phase 0 — Repo scaffolding and CI
- [x] Phase 1 — `/api/analyze`: resume + JD → structured score, keywords, suggestions
- [x] Phase 2 — Flutter UI (input, results)
- [x] Phase 3 — Deploy (GitHub Pages + Render)
- [ ] Phase 4 — PDF / DOCX resume upload
- [ ] Phase 5 — Deterministic keyword score, rate limiting, caching

## License

[MIT](LICENSE)
