# Phase 3 — Deploy (free tier)

**Goal:** GapCheck on the internet at zero cost. The Flutter web app on GitHub Pages (`meetoova.me/gapcheck/`), the Go API on Render's free tier, both deploying automatically from `main`.

---

## 1. The deployed architecture

```
  git push main
      │
      ├─► GitHub Actions: CI (ci.yml)  ── Go + Flutter tests, every push
      │
      ├─► GitHub Actions: Pages (pages.yml)   only when app/** changes
      │      flutter test → flutter build web --base-href /gapcheck/
      │      --dart-define=API_BASE_URL=https://gapcheck-api.onrender.com
      │      → upload artifact → deploy to Pages
      │                                   ▼
      │                     http://meetoova.me/gapcheck/  (static files)
      │
      └─► Render (render.yaml)                 only when backend/** changes,
             waits for CI to pass → go build → start → GET /health ok → swap
                                              ▼
                             https://gapcheck-api.onrender.com  (Go API + Gemini key)
```

The browser loads the app from Pages, then calls the API on Render directly. The Gemini key exists only in Render's environment settings.

## 2. Backend: production readiness

### `cmd/server/main.go`
- **`http.Server` with timeouts** instead of `r.Run()` (which has none):
  - `ReadHeaderTimeout: 10s` — a client that sends headers very slowly can't hold a connection open forever (Slowloris protection).
  - `ReadTimeout: 30s`, `IdleTimeout: 60s`.
  - `WriteTimeout: LLM_TIMEOUT + 15s` — an analysis legitimately takes up to `LLM_TIMEOUT`, so the write limit must be longer.
- **Graceful shutdown:** `signal.NotifyContext` waits for `SIGINT`/`SIGTERM`. Render sends `SIGTERM` on every redeploy. `srv.Shutdown` stops accepting new connections and lets in-flight requests finish (up to 25s, under Render's 30s kill deadline). Without this, a redeploy would cut off anyone mid-analysis.
- Verified locally: SIGTERM → "shutting down, giving requests up to 25s to finish" → clean exit.

### `internal/server/router.go`
- `r.SetTrustedProxies(nil)` — Gin no longer believes `X-Forwarded-For`, which any client can fake. Gin's startup warning is gone. Phase 5 (per-IP rate limiting) will check which header Render's proxy actually sets and trust only that.
- **Release mode** comes from `GIN_MODE=release` (Gin reads it itself): no debug route dump, less logging overhead.

## 3. `render.yaml` — infrastructure as code

A Render **Blueprint**: the whole service defined in the repo instead of clicked together in a dashboard.

| Field | Value | Why |
|---|---|---|
| `type` / `runtime` | `web` / `go` | Native Go build, no Dockerfile needed |
| `plan` | `free` | Must be explicit — omitted means a paid plan |
| `region` | `singapore` | Closest Render region to India |
| `rootDir` | `backend` | The Go module lives there |
| `buildCommand` | `go build -trimpath -ldflags="-s -w" -o bin/server ./cmd/server` | `-trimpath` drops local paths from the binary; `-s -w` strips debug symbols (smaller) |
| `startCommand` | `./bin/server` | Render sets `PORT` (10000); we already read it |
| `healthCheckPath` | `/health` | Render only routes traffic to a new deploy once this returns 200 — zero-downtime deploys |
| `autoDeployTrigger` | `checksPass` | Deploy only after GitHub CI passes for that commit |
| `buildFilter.paths` | `backend/**` | Changing the app or docs doesn't redeploy the API (paths are repo-root relative) |
| `GEMINI_API_KEY` | `sync: false` | Secret: Render asks for it once in its UI; it is never in the repo |
| `GIN_MODE`, `GEMINI_MODEL`, `LLM_TIMEOUT` | plain values | Non-secret config, versioned with the code |
| `ALLOWED_ORIGINS` | `http://meetoova.me,https://meetoova.me,https://nikunjwadhwa001.github.io` | Both schemes, because meetoova.me has no HTTPS certificate yet |

**Free tier facts** (from Render's docs): spins down after **15 minutes** idle, takes about **1 minute** to wake, 750 instance-hours a month (one always-on service uses ~744). With **no payment method on file**, Render suspends instead of billing — so it can never cost money.

## 4. `.github/workflows/pages.yml` — deploy the web app

- Triggers on pushes to `main` that touch `app/**` (or the workflow itself), plus a manual **Run workflow** button.
- **build job:** checkout → Flutter 3.47.6 → `flutter test` (never deploy a broken app) → `flutter build web --release --base-href /gapcheck/ --dart-define=API_BASE_URL=…` → `upload-pages-artifact`.
  - `--base-href /gapcheck/`: the site lives in a sub-folder, so every asset URL needs that prefix. Without it the page loads blank.
  - `API_BASE_URL` comes from a repository **variable** if set, else `https://gapcheck-api.onrender.com`. It's not a secret — the browser sees it anyway.
- **deploy job:** `deploy-pages` publishes the artifact. Uses OIDC (`id-token: write`) to prove it's allowed, instead of a stored token.
- `permissions` are the minimum: read code, write Pages, mint an ID token.
- `concurrency: pages` with `cancel-in-progress: false` — two quick pushes queue instead of racing.
- Hash routing (Phase 2) means refreshes on `/#/results` never 404 on Pages.

## 5. App: free-tier cold starts

Render's free server sleeps after 15 idle minutes and takes ~1 minute to wake. Two changes so users aren't confused:

1. **Wake-up ping.** `InputScreen.initState` calls `api.wakeUp()` — a fire-and-forget `GET /health`. Pasting a resume and job description takes longer than a minute, so by the time the user presses Analyze the server is usually awake. Errors are swallowed; the real request reports problems properly.
2. **Honest loading message.** `_Loading` became a `StatefulWidget` with a `Timer`: after 8 seconds the hint changes to "The free server is waking up after a quiet spell. This first analysis can take up to a minute." The timer is cancelled in `dispose`.
3. The client's **90s timeout** (Phase 2) already covers a ~60s cold start plus the analysis.

## 6. Tests

- `api_client_test.dart`: `wakeUp` sends `GET /health` and never throws, even when the server refuses.
- `widget_test.dart`: the app pings `/health` once on start (recorded with `MockClient`, no real network); the cold-start message appears after 9 seconds of loading and not before.
- Totals: **20 Flutter tests**, all Go tests passing. Production web build checked locally: `<base href="/gapcheck/">`, Render URL compiled in, no `localhost` left.

## 7. HTTPS on meetoova.me

`meetoova.me` currently serves **HTTP only**: GitHub Pages hasn't issued a certificate for it (`https_enforced: false`; the HTTPS certificate doesn't match the name). The DNS A records already point at GitHub (185.199.108–111.153), so the fix is in the user site's settings: **nikunjwadhwa001.github.io → Settings → Pages → Enforce HTTPS** (if greyed out, remove and re-add the custom domain to trigger a certificate).

Until then the app still works and resumes are still encrypted in transit, because the browser sends them to the **HTTPS** Render API. Only the app's own files travel over plain HTTP. CORS allows both `http://` and `https://meetoova.me` so nothing breaks when HTTPS is switched on.

---

## How to explain Phase 3 in 30 seconds

> "Deployment is fully automated and free. A GitHub Actions workflow tests and builds the Flutter web app with the right base path and API URL, then publishes it to GitHub Pages using OIDC instead of a stored token. The Go API is defined as a Render Blueprint in the repo: native Go build, a health check for zero-downtime deploys, deploys only after CI passes, and only when backend files change. The Gemini key is a Render secret, never in the repo. For production I added server timeouts against slow clients, graceful shutdown on SIGTERM so redeploys don't cut off analyses, release mode, and no trusted proxies. Render's free tier sleeps after 15 minutes, so the app pings /health on load to wake it while the user pastes, and explains the delay if it's still waking."
