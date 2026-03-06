# Bug Fix Plan: Worksheet Split & Google Docs Upload

**Date:** 2026-03-06
**Assigned To:** Junior Dev
**Status:** NOT STARTED — Awaiting permission to begin
**Priority:** High — Both bugs affect the core worksheet delivery workflow

---

## ⚠️ Before You Write a Single Line of Code

1. **Read `AGENTS.md`** — Understand the Dual-Agent Protocol.
2. **Read `PLAN.md`** — Review all Phase 4 handoff notes and the Senior Dev sign-off.
3. **Study these files thoroughly** — You will be modifying all of them:

   **`internal/ai/ai.go`** — Current `splitResponse` function and `GenerateWorksheet`. You will debug the delimiter parsing.

   **`internal/ai/prompt.md`** — The embedded system prompt sent to Gemini. You may need to further refine delimiter instructions.

   **`internal/ai/ai_test.go`** — Current tests for `splitResponse`. You will add new test cases.

   **`internal/drive/drive.go`** — Current `UploadFile` function uses `text/markdown` MimeType. You will change it to upload as Google Docs.

   **`internal/drive/drive_test.go`** — You will update the mock and tests for the new upload behaviour.

4. **Do not start coding until Keith gives explicit permission.**

---

## Context: What Changed and Why

### Bug 1: Worksheet Not Splitting

During today's session, the Senior Dev made `splitResponse` more resilient to handle missing delimiters from Gemini. The function now tries multiple delimiter variants and falls back gracefully. However, the result is that the worksheet and teacher key are **no longer being separated** — the full response arrives as one block.

**Root cause investigation needed:** The split logic itself may be correct but Gemini may have changed its output format after the prompt update. The issue is likely one of:

- (a) Gemini is embedding the delimiters inside markdown code fences (e.g., ` ```markdown ... [END STUDENT WORKSHEET] ... ``` `), so `strings.Index` finds them but the surrounding fences corrupt the split.
- (b) Gemini is omitting `[END STUDENT WORKSHEET]` but including `[BEGIN TEACHER KEY]`, and the fallback split isn't extracting the teacher key correctly.
- (c) The `[BEGIN STUDENT WORKSHEET]` stripping logic is too aggressive and removing content.

### Bug 2: Files Upload as Markdown, Not Google Docs

Currently `UploadFile` in `internal/drive/drive.go` sets `MimeType: "text/markdown"` on the file metadata. This creates a raw `.md` file in Google Drive. Keith needs the files as native **Google Docs** so they are immediately readable and editable in Drive.

---

## Your Task List

Complete these in order. Do **not** skip ahead.

### Task B.1 — Diagnose the Split Issue

1. **Generate a worksheet** using the app (`go run cmd/app/main.go`). Use any article text.

2. **Find the locally saved worksheet file** in the `/worksheets/` directory. Open it and look for:
   - Does the file contain `[BEGIN STUDENT WORKSHEET]` or `[END STUDENT WORKSHEET]`?
   - Does the file contain `[BEGIN TEACHER KEY]` or `[END TEACHER KEY]`?
   - Is the teacher key embedded inside the worksheet file, or is there a separate `_teacher_key.md` file?

3. **Check the raw Gemini output.** To do this, temporarily add a debug log in `GenerateWorksheet` in `internal/ai/ai.go`:

   ```go
   // Add this TEMPORARILY after line: text := result.Text()
   os.WriteFile("debug_gemini_response.txt", []byte(text), 0644)
   ```

   Run the app again, generate a worksheet, then open `debug_gemini_response.txt` to see **exactly** what Gemini returned. This will tell you whether the delimiters are present and in what format.

4. **Remove the debug log** after diagnosing.

### Task B.2 — Fix `splitResponse` Based on Diagnosis

Based on what you found in Task B.1, fix `splitResponse` in `internal/ai/ai.go`. Here are the likely fixes:

**If Gemini wraps output in markdown code fences:**

The raw response may look like:

````
```markdown
[BEGIN STUDENT WORKSHEET]
...content...
[END STUDENT WORKSHEET]
```(end fence)
````

In this case, add a preprocessing step at the top of `splitResponse` to strip wrapping code fences:

````go
func splitResponse(fullResponse string) (worksheet string, teacherKey string, err error) {
    // Strip wrapping markdown code fences if present
    cleaned := fullResponse
    cleaned = strings.TrimSpace(cleaned)
    if strings.HasPrefix(cleaned, "```") {
        // Remove opening fence (e.g., ```markdown\n)
        firstNewline := strings.Index(cleaned, "\n")
        if firstNewline != -1 {
            cleaned = cleaned[firstNewline+1:]
        }
        // Remove closing fence
        if strings.HasSuffix(strings.TrimSpace(cleaned), "```") {
            cleaned = strings.TrimSpace(cleaned)
            cleaned = cleaned[:len(cleaned)-3]
            cleaned = strings.TrimSpace(cleaned)
        }
    }
    fullResponse = cleaned

    // ... rest of existing splitResponse logic
}
````

**If the delimiters are present but the teacher key isn't being extracted:**

Verify that `extractTeacherKey` is correctly finding `[BEGIN TEACHER KEY]` in the remaining text after `[END STUDENT WORKSHEET]`. Add a test case (see Task B.4).

### Task B.3 — Convert Upload to Google Docs Format

1. **In `internal/drive/drive.go`**, modify `UploadFile`:

   ```go
   // UploadFile uploads content as a Google Doc to the specified Drive folder.
   func (c *Client) UploadFile(ctx context.Context, folderID, filename, content string) error {
       // Strip .md extension for Google Docs — they don't need file extensions
       docName := strings.TrimSuffix(filename, ".md")

       f := &drive.File{
           Name:     docName,
           MimeType: "application/vnd.google-apps.document", // Google Docs native format
           Parents:  []string{folderID},
       }
       _, err := c.srv.Files.Create(f).Media(strings.NewReader(content)).Do()
       if err != nil {
           return fmt.Errorf("unable to upload file: %w", err)
       }
       return nil
   }
   ```

   **How this works:** When you set the file metadata `MimeType` to `application/vnd.google-apps.document` and provide text/markdown content via `.Media()`, the Google Drive API **automatically converts** the content into a native Google Doc. The markdown headings, bold, lists, etc. will be rendered as formatted Google Docs content.

2. Run `go build ./...` — must compile.

### Task B.4 — Add Test Cases

1. **In `internal/ai/ai_test.go`**, add these test cases to `TestSplitResponse`:

   ````go
   // Test with markdown code fence wrapping
   fencedResp := "```markdown\n[BEGIN STUDENT WORKSHEET]\nWorksheet content\n[END STUDENT WORKSHEET]\n[BEGIN TEACHER KEY]\nKey content\n[END TEACHER KEY]\n```"
   wsFenced, tkFenced, errFenced := splitResponse(fencedResp)
   if errFenced != nil {
       t.Errorf("Unexpected error for fenced response: %v", errFenced)
   }
   if !strings.Contains(wsFenced, "Worksheet content") {
       t.Errorf("Expected worksheet content in fenced response, got: %s", wsFenced)
   }
   if !strings.Contains(tkFenced, "Key content") {
       t.Errorf("Expected teacher key in fenced response, got: %s", tkFenced)
   }

   // Test with bold-wrapped delimiters
   boldResp := "**[BEGIN STUDENT WORKSHEET]**\nBold worksheet\n**[END STUDENT WORKSHEET]**\n**[BEGIN TEACHER KEY]**\nBold key\n**[END TEACHER KEY]**"
   wsBold, tkBold, errBold := splitResponse(boldResp)
   if errBold != nil {
       t.Errorf("Unexpected error for bold response: %v", errBold)
   }
   if !strings.Contains(wsBold, "Bold worksheet") {
       t.Errorf("Expected worksheet content in bold response, got: %s", wsBold)
   }
   if !strings.Contains(tkBold, "Bold key") {
       t.Errorf("Expected teacher key in bold response, got: %s", tkBold)
   }

   // Test fallback: [BEGIN TEACHER KEY] present but [END STUDENT WORKSHEET] missing
   noEndResp := "Worksheet content\n[BEGIN TEACHER KEY]\nTeacher content\n[END TEACHER KEY]"
   wsNoEnd, tkNoEnd, errNoEnd := splitResponse(noEndResp)
   if errNoEnd != nil {
       t.Errorf("Unexpected error: %v", errNoEnd)
   }
   if !strings.Contains(wsNoEnd, "Worksheet content") {
       t.Errorf("Expected worksheet content, got: %s", wsNoEnd)
   }
   if !strings.Contains(tkNoEnd, "Teacher content") {
       t.Errorf("Expected teacher key content, got: %s", tkNoEnd)
   }
   ````

2. Run `go test ./internal/ai/ -v` — all tests must pass.

### Task B.5 — Verify End-to-End

1. `go build ./...` — must compile cleanly.
2. `go test ./... -count=1` — all tests must pass.
3. **Manual test:**
   - Run `go run cmd/app/main.go`
   - Press `g`, fill in the form, press `Ctrl+S`
   - Verify the preview shows **only the student worksheet** (no teacher key mixed in)
   - Press `y`, navigate to a Drive folder, press Space
   - Open Google Drive in a browser and verify:
     - A subfolder was created with the lesson title
     - Inside the subfolder there are **two Google Docs** (not `.md` files)
     - One is the student worksheet, the other is the teacher key
     - Both are formatted as proper Google Docs (not raw markdown text)
4. Also verify local files: check the `/worksheets/` directory for both `_worksheet.md` and `_teacher_key.md` files.

### Task B.6 — 🛑 CHECKPOINT & REVIEW

1. `go build ./...` — must compile cleanly.
2. `go test ./... -count=1` — all tests must pass.
3. `go mod tidy`
4. Update `PLAN.md` — add a handoff note under Phase 4 summarising the bug fixes.
5. Hand off the following to the Senior Dev for review:
   - `internal/ai/ai.go` (split fix)
   - `internal/ai/ai_test.go` (new test cases)
   - `internal/drive/drive.go` (Google Docs upload)

---

## Quality Checklist

- [ ] `debug_gemini_response.txt` has been deleted (do NOT commit debug files)
- [ ] `splitResponse` correctly separates worksheet and teacher key for all delimiter variants
- [ ] `splitResponse` handles missing delimiters gracefully (no errors returned to user)
- [ ] `UploadFile` creates native Google Docs, not `.md` files
- [ ] Google Doc names do not have `.md` extension
- [ ] Both local save and Drive upload produce two separate documents
- [ ] All existing tests still pass
- [ ] New test cases cover fenced, bold, and missing delimiter scenarios

---

## Reminder

**Do not start coding until you have:**

1. Read `AGENTS.md`, `PLAN.md`, and `reviews/TASK_4_FEEDBACK.md`.
2. Studied the current code in `internal/ai/ai.go`, `internal/drive/drive.go`, and their test files.
3. Summarised your understanding of today's tasks back to Keith.
4. Received Keith's explicit permission to begin.
