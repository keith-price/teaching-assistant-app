# Phase 2 Code Review — Senior Dev Feedback

**Scope:** `internal/calendar/` package
**Verdict:** Good work overall — a few issues to address before moving to Phase 3.

---

## ✅ Strengths

1. **Clean struct-based design** — `Client` wraps the Google Calendar service, matching the `Store` pattern from Phase 1. No globals. Good consistency.
2. **WIB timezone handling is solid** — `getWIBLocation()` uses `time.FixedZone("WIB", 7*3600)` correctly. Time boundaries for today/tomorrow are constructed in WIB. RFC3339 results are converted via `.In(loc)`. This satisfies the critical constraint from AGENTS.md.
3. **`parseEventTime` is well thought out** — Handles both RFC3339 (timed events) and date-only strings (all-day events). Falls through gracefully.
4. **Good test coverage on the pure functions** — Table-driven tests for `parseEventTime` cover UTC→WIB conversion, offset-aware input, date-only input, and invalid input. `TestWIBLocation` verifies the zone name and offset. All tests pass.
5. **OAuth2 flow is properly structured** — Token caching with file permissions `0600` is a nice security touch.
6. **Error wrapping** — Consistent use of `fmt.Errorf("...: %w", err)` throughout.

---

## 🔴 Critical (Must Fix)

### 1. Silently skipping unparseable events is dangerous
**File:** `calendar.go:157-163`

When `parseEventTime` fails, the event is silently dropped with `continue`. The caller gets zero indication that events were lost. In a teaching dashboard, a silently missing lesson is worse than an error.

**Requested change:** At minimum, log a warning. Better yet, return an error or collect parsing errors:

```go
startTime, err := parseEventTime(startStr, wibLocation)
if err != nil {
    return nil, fmt.Errorf("unable to parse start time for event %q: %w", item.Summary, err)
}
```

If you want to be lenient and keep processing remaining events, collect warnings and return them alongside the events (e.g. via a multi-return or a log).

---

## 🟡 Important (Should Fix)

### 2. No pagination — only fetches first page of results
**File:** `calendar.go:132-139`

The Google Calendar API returns paginated results (default 250 events max per page). If the teacher has a busy schedule over 2 days, results could be truncated without any indication. For now this is unlikely to bite, but it's a latent bug.

**Requested change:** Add `.MaxResults(2500)` to the query or implement pagination using `NextPageToken`. The simple fix:

```go
eventsList, err := c.srv.Events.List("primary").
    ShowDeleted(false).
    SingleEvents(true).
    TimeMin(timeMin).
    TimeMax(timeMax).
    MaxResults(100).
    OrderBy("startTime").
    Q(keyword).
    Do()
```

This explicitly caps at 100 (plenty for 2 days) and documents the limit.

---

### 3. `credentials.json` path — not aligned with project decision
**File:** `calendar.go:32` (parameter name), PLAN.md Task 2.3

We agreed to move credentials into a `config/` directory (see `NEXT_STEPS_PHASE2.md`). The code accepts the path as a parameter (good!), but the PLAN.md checkpoint note at Task 2.3 still says "place it in the root". Ensure the calling code in `main.go` (Phase 6) will reference `config/credentials.json` and `config/token.json`.

**Requested change:** No code change needed in the calendar package itself (the path is parameterised). Just update the PLAN.md Task 2.3 text to reference `config/` instead of root, for consistency.

---

### 4. No test coverage for `FetchEvents` or OAuth flow
**File:** `calendar_test.go`

The tests only cover `parseEventTime` and `getWIBLocation` — which are purely internal helpers. The main public API (`NewClient`, `FetchEvents`) has zero test coverage. This is understandable since they need live Google API access, but there's an architectural issue: **the `Client` struct directly embeds `*calendar.Service`, making it impossible to mock.**

**Requested change:** Introduce an interface to make `FetchEvents` testable without live API calls:

```go
// EventFetcher allows mocking of the calendar service in tests.
type EventFetcher interface {
    FetchEvents(ctx context.Context, keyword string) ([]Event, error)
}
```

This ensures Phase 4 (TUI) and Phase 5 (WhatsApp daemon) can inject a mock calendar during testing. No need to write mock tests now — just prepare the interface.

---

## 🟢 Minor / Housekeeping

### 5. `getTokenFromWeb` uses a static state parameter
**File:** `calendar.go:74`

`oauth2.AccessTypeOffline` is correct, but `"state-token"` as the state parameter is a static string. Ideally this should be a random value verified on callback to prevent CSRF. For a local desktop app this is low risk, but worth noting.

**Requested change:** Add a comment acknowledging this is acceptable for a local-only app:

```go
// Note: Static state token is acceptable for local-only desktop OAuth flow.
authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
```

---

### 6. `saveToken` prints to stdout
**File:** `calendar.go:103`

`fmt.Printf("Saving credential file to: %s\n", path)` will interfere with the Bubble Tea TUI in Phase 4. TUI apps control stdout — random prints will corrupt the display.

**Requested change:** Remove the `Printf` or replace with a structured logger that can be silenced when the TUI is active. For now, simply remove it:

```go
func saveToken(path string, token *oauth2.Token) error {
    f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
    // ...
```

---

## Summary Checklist

| # | Issue | Severity | File |
|---|-------|----------|------|
| 1 | Silent skip of unparseable events | 🔴 Critical | `calendar.go:157-163` |
| 2 | No pagination on event listing | 🟡 Important | `calendar.go:132-139` |
| 3 | PLAN.md still references root for credentials | 🟡 Important | `PLAN.md:42` |
| 4 | No `EventFetcher` interface for testability | 🟡 Important | `calendar.go` |
| 5 | Static OAuth state token | 🟢 Minor | `calendar.go:74` |
| 6 | `fmt.Printf` in `saveToken` will break TUI | 🟢 Minor | `calendar.go:103` |

---

## Review Process Log

### Step 1 — Initial Review (2026-03-05)
**Reviewer:** Senior Dev (AI)
**Files reviewed:** `internal/calendar/calendar.go`, `internal/calendar/calendar_test.go`

- Identified 6 issues (1 critical, 3 important, 2 minor) — see sections above.
- All existing tests pass:

```
$ go test -v ./internal/calendar/
=== RUN   TestParseEventTime
=== RUN   TestParseEventTime/Valid_RFC3339_UTC_to_WIB
=== RUN   TestParseEventTime/Valid_RFC3339_WIB_with_offset
=== RUN   TestParseEventTime/Valid_All-Day_Event_(Date_only)
=== RUN   TestParseEventTime/Invalid_Time_String
--- PASS: TestParseEventTime (0.00s)
=== RUN   TestWIBLocation
--- PASS: TestWIBLocation (0.00s)
PASS
```

- Handed feedback report to Junior Dev for implementation.
