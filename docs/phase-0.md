# Phase 0 — Scaffolding and CI

**Goal:** an empty-but-working skeleton. Both halves of the project (Go API and Flutter app) build, have one passing test each, and GitHub checks every push automatically. No product features yet.

**Commit:** `e3ee5e5` — "Phase 0: scaffold Go backend, Flutter app, and CI"

---

## 1. The big picture

```
            Browser (Flutter web, hosted on GitHub Pages)
                         │  HTTPS / JSON
                         ▼
            Go + Gin API (hosted on Render)  ── holds the secret Gemini key
                         │
                         ▼
                 Google Gemini API (LLM)
```

- **Why a separate backend at all?** GitHub Pages only serves static files, and anything shipped to the browser is public. If the Gemini API key lived in the Flutter app, anyone could open DevTools and steal it. The Go backend keeps the key secret and is also where validation, prompt building, and response parsing happen.
- **Why a monorepo?** Frontend and backend change together (the JSON contract between them). One repo means one PR can change both sides, one CI pipeline, one README.

## 2. Repo layout

```
gapcheck/
├── backend/                 Go API
│   ├── cmd/server/main.go   entrypoint (starts the HTTP server)
│   ├── internal/server/     routes + handlers (+ tests)
│   ├── go.mod / go.sum      dependencies (Gin)
│   └── .env.example         config template, no secrets
├── app/                     Flutter client (web + Android)
│   ├── lib/main.dart        the app
│   ├── test/widget_test.dart
│   └── pubspec.yaml         Flutter dependencies
├── .github/workflows/ci.yml GitHub Actions pipeline
├── docs/                    these phase write-ups
├── .gitignore, README.md, LICENSE
```

## 3. Backend (Go + Gin)

### `cmd/server/main.go` — the entrypoint
- Reads the `PORT` environment variable, falls back to `8080`.
- Builds the router and starts listening.
- **Why read `PORT` from env?** Hosting platforms like Render tell your app which port to use through `PORT`. Hard-coding it would break deployment. This follows the "12-factor app" idea: configuration comes from the environment, not from code.
- **Why `cmd/` and `internal/`?** Standard Go project layout. `cmd/<name>` holds runnable programs; `internal/` holds packages that Go *forbids* other modules from importing, so it's the app's private code.

### `internal/server/router.go` — routes
- `NewRouter()` creates a Gin engine and registers routes. Right now there is one: `GET /health` → `{"status":"ok"}`.
- Uses `gin.New()` plus explicit `Logger()` and `Recovery()` middleware (rather than `gin.Default()`) so it's visible exactly what runs on each request:
  - **Logger** prints one line per request (method, path, status, latency).
  - **Recovery** catches a panic in any handler and returns HTTP 500 instead of crashing the whole server.
- **Why a health endpoint?** Hosting platforms and uptime monitors ping it to know the service is alive. It's also the simplest possible proof that the server runs.
- **Why a `NewRouter()` function instead of building routes in `main`?** So tests can create the exact same router without starting a real server.

### `internal/server/router_test.go` — the test
- Uses Go's `httptest` package: creates a fake request to `/health`, runs it through the router *in memory* (no network, no port), and checks the status code is 200 and the body is `{"status":"ok"}`.
- `gin.SetMode(gin.TestMode)` silences Gin's debug output during tests.

### `go.mod` / `go.sum`
- `go.mod` declares the module name `github.com/nikunjwadhwa001/gapcheck/backend` (matches the GitHub URL, which is the Go convention) and the Go version (1.25).
- `go.sum` stores checksums of every dependency so builds are reproducible and tamper-evident.

### `.env.example`
- A template listing the config the backend will need: `PORT`, `GEMINI_API_KEY` (Phase 1), `ALLOWED_ORIGINS` (Phase 3, for CORS).
- The real `.env` is git-ignored, so secrets never reach GitHub. The example file *is* committed so anyone cloning the repo knows what to set.

## 4. App (Flutter)

- Created with `flutter create --platforms web,android`. Web is what gets hosted on GitHub Pages; Android comes for free from the same code (an APK is an optional Phase 6 item).
- Package id: `com.nikunjwadhwa.gapcheck`.

### `lib/main.dart`
- Replaced Flutter's default counter demo with a simple placeholder: "GapCheck — Score your resume against a job description — Coming soon".
- `GapCheckApp` sets up Material 3 theming from one seed colour (indigo), with both a light and a dark theme. The OS setting picks which one.
- `HomePage` is a `StatelessWidget` — it has no state yet. Real screens, Riverpod (state management), and GoRouter (navigation) come in Phase 2.

### `test/widget_test.dart`
- A widget test: renders the app in a fake screen and asserts the title and tagline text appear. Proves the app builds and renders without errors.

## 5. CI — `.github/workflows/ci.yml`

GitHub Actions runs this on every push to `main` and on every pull request. Free for public repos. Two jobs run **in parallel**:

| Job | Steps | What each step catches |
|---|---|---|
| **Backend (Go)** | `gofmt` check → `go vet` → `go test` | unformatted code → suspicious code (e.g. wrong printf args) → broken behaviour |
| **App (Flutter)** | `dart format` check → `flutter analyze` → `flutter test` → `flutter build web` | unformatted code → lint/type issues → broken widgets → app that doesn't compile for web |

- `permissions: contents: read` — the workflow token can only read the repo. Least privilege: a compromised step can't push code.
- Go version is read from `go.mod` and Flutter is pinned to `3.47.6`, so CI uses the same versions as your laptop. Dependency caching makes reruns faster.
- First run: Backend 55s, App 2m16s, both green.

## 6. Repo hygiene files

- **`.gitignore`** (root): ignores `.env` files (secrets), Go build outputs, and editor/OS junk. Flutter's own `app/.gitignore` (generated) ignores `build/`, `.dart_tool/`, etc.
- **`README.md`**: what the project is, stack table, how to run each part locally, and a phase roadmap with checkboxes.
- **`LICENSE`**: MIT — anyone can use the code with attribution. Standard for portfolio projects.

## 7. GitHub setup

- Public repo `nikunjwadhwa001/gapcheck` (public = free Actions minutes and free Pages).
- **Pages source = GitHub Actions**: a workflow (added in Phase 3) will build Flutter web and publish it. Because the account has a custom domain, the site will live at `meetoova.me/gapcheck/`.

---

## How to explain Phase 0 in 30 seconds

> "I set up GapCheck as a monorepo with a Go/Gin API and a Flutter client. The API follows the standard `cmd/` + `internal/` layout, reads its config from environment variables so it deploys cleanly on Render, and exposes a health check covered by an in-memory `httptest` test. The Flutter app targets web and Android from one codebase. GitHub Actions runs formatting, static analysis, tests, and a web build on both halves for every push, with a read-only token. The LLM API key will only ever live on the backend, because GitHub Pages is static and anything in the frontend is public."
