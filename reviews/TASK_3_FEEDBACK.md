# Phase 3 Code Review — Senior Dev Feedback

**Scope:** `internal/ai/` package
**Verdict:** Solid implementation — well structured, good use of `//go:embed`, and the `PROMPT.md` separation was a smart call. A few issues to address before moving to Phase 4.

---

## ✅ Strengths

1. **`//go:embed prompt.md` is excellent** — Embedding the prompt as a separate `PROMPT.md` file via Go's `embed` directive keeps `ai.go` clean while ensuring the prompt is compiled into the binary (no runtime file I/O, no missing-file errors in production). The prompt itself is well-structured with clear ESL pedagogy.
2. **`ClipboardReader` interface** — Exactly what was requested. Enables full testability without touching the real OS clipboard. Matches the `EventFetcher` pattern from Phase 2.
3. **Struct-based `Generator` design** — Consistent with the `Store` pattern from Phase 1. No globals. Dependencies are injected via the constructor.
4. **Empty clipboard handling** — Returns a clear error instead of sending empty text to Gemini. Good defensive programming.
5. **Error wrapping** — Consistent `fmt.Errorf("...: %w", err)` throughout. Keeps the error chain intact for callers.
6. **`.gitignore` updated correctly** — `/worksheets/` is properly ignored as per the decision documented in `PLAN.md`.
7. **`godotenv.Load()` failure is non-fatal** — Silently falling through with `_ = godotenv.Load()` is correct — the API key may come from the actual environment rather than a `.env` file. Clean approach.

---

## 🔴 Critical (Must Fix)

### 1. System prompt should use `genai.SystemInstruction`, not prompt concatenation
**File:** `ai.go:84-89`

The system prompt is currently concatenated into the user prompt:

```go
fullPrompt := SystemPrompt + "\n\n" + prompt
result, err := g.client.Models.GenerateContent(ctx, "gemini-2.0-flash", genai.Text(fullPrompt), nil)
```

The Google GenAI Go SDK supports proper system instructions via `GenerateContentConfig`. System instructions are treated differently by the model — they have higher authority and are not "confused" with user content. Concatenating them into the user prompt weakens instruction-following and risks prompt injection from transcript content.

**Requested change:**

```go
result, err := g.client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), &genai.GenerateContentConfig{
    SystemInstruction: genai.NewContentFromText(SystemPrompt, genai.RoleUser),
})
```

This properly separates the system instruction channel from the user content channel.

---

### 2. `PROMPT.md` filename casing — `//go:embed` references `prompt.md` but file is `PROMPT.md`
**File:** `ai.go:16`

The embed directive says:

```go
//go:embed prompt.md
var SystemPrompt string
```

But the file in the directory listing earlier showed `PROMPT.md` (uppercase). On Linux (and in CI), Go's `embed` is **case-sensitive** — this will fail to compile. It works on Windows only because the filesystem is case-insensitive.

**Requested change:** Either:
- (a) **Rename the file** to `prompt.md` (lowercase) to match the directive, OR
- (b) **Change the directive** to `//go:embed PROMPT.md`

Option (a) is preferred — lowercase filenames are the Go convention for non-exported resources.

---

## 🟡 Important (Should Fix)

### 3. `SaveWorksheet` overwrites files without warning
**File:** `ai.go:111-115`

If the teacher generates a worksheet for `"Bob Smith"` twice on the same day, the second call silently overwrites the first. This could cause data loss.

**Requested change:** Either:
- (a) Check if the file exists first and append a counter suffix (e.g. `2026-03-06_Bob_Smith_lesson_2.md`), OR
- (b) Use a timestamp with time component (e.g. `2026-03-06_1430_Bob_Smith_lesson.md`)

Option (b) is simpler and more informative:

```go
dateStr := time.Now().Format("2006-01-02_1504")
```

---

### 4. `SaveWorksheet` uses a relative path — fragile when called from different working directories
**File:** `ai.go:104`

```go
dir := "worksheets"
```

This creates `worksheets/` relative to whatever the current working directory happens to be when the binary runs. If the TUI is launched from a different directory (or via a shortcut), the worksheets end up in an unexpected location.

**Requested change:** Accept the base directory as a parameter, or resolve it relative to the executable/project root:

```go
func (g *Generator) SaveWorksheet(content, studentName, baseDir string) (string, error) {
    dir := filepath.Join(baseDir, "worksheets")
    // ...
}
```

This also makes the function testable with temp directories (see item #6).

---

### 5. `constructPrompt` doesn't include lesson time or lesson type
**File:** `ai.go:71-73`

The `PROMPT.md` system prompt explicitly expects four input parameters: `TARGET LEVEL`, `LESSON TIME`, `LESSON TYPE`, and `TOPIC/SOURCE`. But `constructPrompt` only sends `studentName`, `studentLevel`, and `transcript`. Lesson time and lesson type are missing.

**Requested change:** Add `lessonTime` and `lessonType` parameters:

```go
func constructPrompt(studentName, studentLevel, lessonTime, lessonType, transcript string) string {
    return fmt.Sprintf(
        "Student Name: %s\nTarget Level: %s\nLesson Time: %s minutes\nLesson Type: %s\n\nTranscript:\n%s",
        studentName, studentLevel, lessonTime, lessonType, transcript,
    )
}
```

Update `GenerateWorksheet` and `GenerateAndSaveWorksheet` signatures accordingly. This ensures the prompt template and the code are aligned.

---

### 6. `TestSaveWorksheet` writes to the real `worksheets/` directory
**File:** `ai_test.go:68-94`

The test creates a real file in `worksheets/` and only cleans up the file — not the directory. This:
- Pollutes the project directory with a `worksheets/` folder during test runs
- Could conflict with other tests in CI
- Doesn't test the `MkdirAll` path properly

**Requested change:** Use `t.TempDir()` for test isolation:

```go
func TestSaveWorksheet(t *testing.T) {
    tmpDir := t.TempDir()
    // Pass tmpDir as baseDir (after fixing item #4)
    // ...
}
```

This is automatically cleaned up by the test framework.

---

## 🟢 Minor / Housekeeping

### 7. `genai.Text()` return value — verify `result.Text()` isn't deprecated
**File:** `ai.go:89-94`

Verify that `result.Text()` is the correct accessor for the current SDK version (`google.golang.org/genai v1.49.0`). Some versions use `result.Candidates[0].Content.Parts[0]` instead. If `Text()` is a convenience method that panics on empty results, the nil/empty check on line 95 may not catch all edge cases.

**Requested change:** Check the SDK docs. If `Text()` can panic, wrap it in a safety check:

```go
if result == nil || len(result.Candidates) == 0 {
    return "", fmt.Errorf("received empty response from Gemini API")
}
text := result.Text()
```

---

### 8. Model name should be a configurable constant
**File:** `ai.go:89`

`"gemini-2.0-flash"` is hardcoded in the function body. When Gemini updates models, this requires editing function logic.

**Requested change:** Extract to a package-level constant and update to `gemini-2.5-flash` (confirmed with Keith — free tier covers our usage and the stronger reasoning benefits C1/C2 worksheet quality):

```go
const DefaultModel = "gemini-2.5-flash"
```

Or accept it as a `Generator` field set during construction.

---

## Summary Checklist

| # | Issue | Severity | File |
|---|-------|----------|------|
| 1 | Use `SystemInstruction` instead of prompt concatenation | 🔴 Critical | `ai.go:84-89` |
| 2 | `PROMPT.md` filename casing mismatch with `//go:embed` | 🔴 Critical | `ai.go:16` |
| 3 | `SaveWorksheet` silently overwrites same-day duplicates | 🟡 Important | `ai.go:111` |
| 4 | Relative `worksheets/` path is fragile | 🟡 Important | `ai.go:104` |
| 5 | `constructPrompt` missing `lessonTime` and `lessonType` | 🟡 Important | `ai.go:71-73` |
| 6 | Tests write to real directory, not `t.TempDir()` | 🟡 Important | `ai_test.go:68-94` |
| 7 | Verify `result.Text()` safety for current SDK version | 🟢 Minor | `ai.go:94` |
| 8 | Model name should be a constant, not inline string | 🟢 Minor | `ai.go:89` |

---

## Review Process Log

### Step 1 — Initial Review (2026-03-06)
**Reviewer:** Senior Dev (AI)
**Files reviewed:** `internal/ai/ai.go`, `internal/ai/ai_test.go`, `internal/ai/PROMPT.md`

- Identified 8 issues (2 critical, 4 important, 2 minor) — see sections above.
- All existing tests pass:

```
$ go test -v ./internal/ai/
=== RUN   TestConstructPrompt
--- PASS: TestConstructPrompt (0.00s)
=== RUN   TestClipboardReaderErrors
--- PASS: TestClipboardReaderErrors (0.00s)
=== RUN   TestSaveWorksheet
--- PASS: TestSaveWorksheet (0.09s)
PASS
```

- `go vet ./internal/ai/` — clean, no issues.
- Full project tests (`go test ./...`) — all pass, no regressions.
- Handed feedback report to Junior Dev for implementation.

### Step 2 — Verification of Review Fixes (2026-03-06)
**Reviewer:** Senior Dev (AI)
**Files reviewed:** `internal/ai/ai.go`, `internal/ai/ai_test.go`, `internal/ai/prompt.md`

**All 8 items verified as resolved:**

| # | Issue | Status | Verification |
|---|-------|--------|-------------|
| 1 | Use `SystemInstruction` | ✅ Fixed | Line 91: `SystemInstruction: genai.NewContentFromText(SystemPrompt, genai.RoleUser)` |
| 2 | `prompt.md` filename casing | ✅ Fixed | File is now lowercase `prompt.md`, matching `//go:embed prompt.md` |
| 3 | Overwrite protection | ✅ Fixed | Line 116: timestamp format is `2006-01-02_1504` (includes time) |
| 4 | Relative path → `baseDir` param | ✅ Fixed | Line 110: `SaveWorksheet(content, studentName, baseDir string)` with `filepath.Join` |
| 5 | Missing `lessonTime`/`lessonType` | ✅ Fixed | Line 74: `constructPrompt(studentName, studentLevel, lessonTime, lessonType, transcript)` |
| 6 | Tests use `t.TempDir()` | ✅ Fixed | `ai_test.go:82`: `tmpDir := t.TempDir()` |
| 7 | `result.Candidates` safety check | ✅ Fixed | Lines 97-99: nil and empty candidates check before `result.Text()` |
| 8 | Model name constant | ✅ Fixed | Line 20: `const DefaultModel = "gemini-2.5-flash"` (upgraded per Keith's decision) |

**Test results — all pass, no regressions:**

```
$ go test ./... -count=1
ok  teaching-assistant-app/internal/ai       0.09s
ok  teaching-assistant-app/internal/db       0.53s
```

```
$ go vet ./internal/ai/
(clean — no issues)
```

**Verdict:** Phase 3 is **signed off**. Ready to proceed to Phase 4.
