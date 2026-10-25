# Phase 2 — The Flutter UI

**Goal:** a real app on top of the API. Paste a resume and a job description, press Analyze, and see a score ring, keyword chips and suggestion cards. Every API error code gets a clear screen. Verified end to end in a browser against the real Go API and Gemini.

---

## 1. What the user sees

```
 /  Input screen                         /#/results  Results screen
┌──────────────────────────────┐        ┌──────────────────────────────┐
│ Your resume │ Job description│        │  (65)  Partial match         │
│             │                │  ───►  │        summary…              │
│ 494/20,000  │ 401/20,000     │        │ Missing (9):  gRPC  Redis …  │
│ [Analyze match] Try an example│       │ Matching (4): Go  SQL …      │
└──────────────────────────────┘        │ Suggested edits: [cards]     │
                                        └──────────────────────────────┘
```

- Side by side on wide screens, stacked on phones (`LayoutBuilder`, breakpoint 760px).
- **Analyze** is disabled until both boxes have 50–20,000 characters — the same limits as the backend, so bad input never costs a request.
- **Try an example** fills in a made-up resume and job description, so anyone can try the app in one click.

## 2. File layout

```
app/lib/
├── main.dart                 ProviderScope + MaterialApp.router + themes
├── router.dart               GoRouter: "/" and "/results"
├── theme.dart                Material 3 theme + semantic match colours
├── data/
│   ├── models.dart           AnalysisResult, Suggestion, MatchBand, ApiException, limits
│   ├── api_client.dart       GapCheckApi: POST /api/analyze, error parsing
│   └── sample.dart           example texts
├── state/
│   └── analysis_controller.dart   Riverpod providers + AnalysisController
├── screens/
│   ├── input_screen.dart
│   └── results_screen.dart
└── widgets/
    ├── score_ring.dart       animated CustomPainter gauge
    ├── keyword_section.dart  green/amber chips
    ├── suggestion_card.dart
    └── error_view.dart       error code → title, icon, actions
```

Layers: **data** (talks to the API) → **state** (holds what's happening) → **screens/widgets** (draw it). Screens never call HTTP directly.

## 3. Data layer

### `models.dart`
- Dart mirrors of the Go types. `fromJson` is defensive: a missing or `null` list becomes `[]`, a missing string becomes `''`. The backend already guarantees this, but the app doesn't crash if it ever changes.
- `MatchBand` — Excellent (90+), Strong (70–89), Partial (50–69), Weak (<50). The **same thresholds as the backend prompt**, so the label always agrees with what the model was told.
- `ApiException {code, message, field, statusCode}` — one error type for the whole app. `code` is the backend's code (`invalid_input`, `rate_limited`, `timeout`, `upstream_error`, …) plus two client-only ones: `network` (no response at all) and `unknown` (a response we can't parse).
- `minTextLength = 50`, `maxTextLength = 20000` — duplicated from Go on purpose; tests on both sides pin them.

### `api_client.dart`
- `API_BASE_URL` is a **compile-time constant** (`String.fromEnvironment`), set with `--dart-define`. Defaults to `http://localhost:8080`. In Phase 3 the GitHub Pages build points it at Render.
- `analyze()` POSTs JSON, then:
  - 200 → `AnalysisResult.fromJson`
  - error body `{"error": {...}}` → `ApiException` with the server's code and message
  - non-JSON body (e.g. an HTML 502 page from a proxy) → `unknown`
  - `http.ClientException` (offline, server down, CORS blocked) → `network`
  - no reply within **90s** → `timeout`. Longer than the server's own 30s, with room for Render's free tier waking from sleep.
- Takes an optional `http.Client`, so tests pass `MockClient` — **dependency injection again**, same idea as the Go `LLM` interface.

## 4. State — Riverpod

```dart
final apiProvider = Provider<GapCheckApi>(...);
final analysisProvider = NotifierProvider<AnalysisController, AsyncValue<AnalysisResult>?>(...);
```

The analysis state has exactly four shapes:

| State | Meaning | Results screen shows |
|---|---|---|
| `null` | nothing analysed yet (e.g. someone opened `/#/results` directly) | "No analysis yet" + button back |
| `AsyncLoading` | request in flight | spinner + "Comparing your resume with the job…" |
| `AsyncData(result)` | success | score ring, chips, cards |
| `AsyncError(ApiException)` | failure | `ErrorView` for that code |

- `AnalysisController.analyze()` sets `AsyncLoading`, then `AsyncValue.guard(...)` runs the API call and turns a thrown `ApiException` into `AsyncError` instead of crashing.
- It remembers the last input, so **Try again** (`retry()`) resends it without the user retyping.
- `AsyncValue` is a **sealed class**, so the results screen uses a Dart 3 `switch` expression over it. The compiler checks every case is handled.
- **Why Riverpod?** State lives outside widgets, so the input screen starts the analysis and the results screen reads it, without passing callbacks around. And tests override providers (`apiProvider.overrideWithValue`, `analysisProvider.overrideWith`) to put a screen in any state instantly.

## 5. Navigation — GoRouter

- Two routes: `/` and `/results`. Analyze calls `context.push('/results')`, so the input screen stays underneath with the text still in it. Back returns to it.
- **Hash URLs** (`/#/results`, Flutter web's default) — on purpose for GitHub Pages, which only serves real files. A refresh on `/results` would 404 there; `/#/results` always loads `index.html`.
- The router is created once in `_GapCheckAppState`, so navigation state survives rebuilds.

## 6. Screens and widgets

- **`InputScreen`** (`ConsumerStatefulWidget`): owns two `TextEditingController`s (disposed properly). `ListenableBuilder` rebuilds only the counter and the button as you type, not the whole screen. Helper text says how many more characters are needed; error text appears above 20,000.
- **`ResultsScreen`** (`ConsumerWidget`): the `switch` above. Content capped at 860px wide so lines stay readable on big screens.
- **`ScoreRing`**: `TweenAnimationBuilder` animates 0 → score over 900ms; a `CustomPainter` draws a grey track and a coloured arc starting at 12 o'clock. Wrapped in `Semantics` so screen readers say "Match score 65 out of 100".
- **`KeywordSection`**: amber chips for missing, green for matched, with counts in the title.
- **`SuggestionCard`**: section pill → issue (muted) → fix (bold, with an arrow). The fix is what the user acts on, so it gets the emphasis.
- **`ErrorView`**: maps code → title + icon + whether retry makes sense. Input problems (`invalid_input`, `invalid_json`, `body_too_large`) show only **Edit inputs**, because retrying the same input would fail the same way. Everything else offers **Try again**. The server's message is shown as is — it was written for users in Phase 1.
- **`theme.dart`**: Material 3 from an indigo seed, light and dark. `MatchColors` are separate semantic colours (green/amber/red) tuned per brightness, so "good" and "missing" never depend on the brand colour.

## 7. Backend change: CORS

The web app runs on `localhost:3000`; the API on `localhost:8080`. Different ports = different **origins**, and browsers block cross-origin requests unless the server opts in.

- New `server/cors.go` middleware:
  - If the request's `Origin` is in `ALLOWED_ORIGINS`, reply with `Access-Control-Allow-Origin: <that origin>` plus allowed methods and headers.
  - Unknown origins get no header, so the browser blocks them.
  - `OPTIONS` requests (the browser's **preflight** check before a JSON POST) are answered with 204 right in the middleware.
  - Always sends `Vary: Origin`, so caches don't serve one origin's response to another.
  - `Access-Control-Max-Age: 600` lets the browser skip the preflight for 10 minutes.
- Written by hand (~30 lines) instead of a library, so every header is visible and explainable.
- 4 new Go tests: allowed origin, trimmed config, unknown origin, preflight.

**Gotcha found while testing:** the plan was port 5000, but on macOS **AirPlay Receiver** listens on 5000, so `localhost:5000` returned AirPlay's 403. Dev port moved to **3000** (`.env.example` and the README updated).

## 8. Tests (17 Flutter + 4 new Go)

| File | What it proves |
|---|---|
| `test/models_test.dart` | JSON parsing; null/missing lists become `[]`; score bands match the backend thresholds exactly |
| `test/api_client_test.dart` | Request method, URL, headers, body; success parsing; error body → code + field; 429 kept; HTML error page → `unknown`; connection failure → `network`; slow server → `timeout` (all with `MockClient`, no server) |
| `test/widget_test.dart` | App starts on input screen with Analyze disabled; button enables only when both texts are ≥50 chars; "41 more" helper text; Try an example fills both; results screen shows score/band/chips/suggestion; loading state; rate-limit error offers retry; input error offers only edit; opening `/results` directly shows a way back |
| `backend/internal/server/cors_test.go` | Allowed origin echoed; config trimmed; unknown origin gets nothing; preflight → 204 |

One gotcha: widget tests default to a 3× pixel ratio, so a 1200×1600 test window is only 400×533 logical pixels and the Analyze button was off screen. Setting `devicePixelRatio = 1` fixed it.

## 9. End-to-end check (25 Oct 2026)

Ran the real Go API (with the Gemini key) and the built web app in a browser:

1. Input screen rendered; **Try an example** filled both boxes (494 and 401 characters).
2. **Analyze match** → results in a few seconds: **65, Partial match**, 9 missing keywords (gRPC, Redis, Docker, Kubernetes, GitHub Actions, Prometheus, Grafana, AWS, Kafka), 4 matched (Go, PostgreSQL, REST API design, SQL), suggestion cards.
3. Phone width (375px): score card, summary and chips wrap correctly.
4. Stopped the API and analysed again → **"Can't reach the server"** with Try again / Edit inputs. No crash.

---

## How to explain Phase 2 in 30 seconds

> "The Flutter app has three layers: a data layer with typed models and an API client, Riverpod state, and screens. The analysis state is an AsyncValue — loading, data or error — and the results screen switches over that sealed type, so every state is handled at compile time. Every backend error code maps to a specific screen: input problems offer 'edit', transient ones offer 'retry'. Validation limits match the backend so bad input never costs a request. The API client takes an injected HTTP client, so 17 tests run without a server. For the browser to call the API I added a small hand-written CORS middleware with origin allow-listing and preflight handling, and verified the whole flow end to end in a browser."
