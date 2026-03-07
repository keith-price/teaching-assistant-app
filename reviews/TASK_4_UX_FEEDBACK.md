# Task 4 Feedback — UX Enhancements & Bug Fixes

**Reviewer:** Senior Dev
**Date:** 2026-03-07
**Scope:** Google Docs formatting, Drive folder navigation, loading spinners, local worksheet cleanup, and TUI list pagination.

---

## Files Reviewed

| File                       | Verdict |
| -------------------------- | ------- |
| `internal/ai/ai.go`        | ✅ Pass |
| `internal/ai/prompt.md`    | ✅ Pass |
| `internal/drive/drive.go`  | ✅ Pass |
| `internal/tui/model.go`    | ✅ Pass |
| `internal/tui/update.go`   | ✅ Pass |
| `internal/tui/view.go`     | ✅ Pass |
| `internal/tui/commands.go` | ✅ Pass |

## Review Notes

### 1. Google Docs Formatting (Markdown to HTML)

- `internal/ai/ai.go`: The overly greedy regex issue stripping the `#` from the title heading was successfully resolved.
- `internal/ai/prompt.md`: The system prompt correctly enforces the `# 3. READING` structure, explicitly forbids bolding the blank gaps (`________`), and now enforces the required CEFR length and complexity constraints.
- `internal/drive/drive.go`: The injected CSS inside `UploadFile` correctly neutralizes the Gomarkdown `<p>` tag margins inside `<li>` elements, resolving the list spacing issues.

### 2. Drive Folder Picker Navigation

- `internal/tui/update.go`: The manual navigation stack logic handles traversing upwards (`Backspace`) beautifully. Returning `m.folderParentID` to `"root"` safely resolves the silent API failure.

### 3. Progress Spinners

- `internal/tui/model.go`: The `spinner.Model` is correctly initialized via Bubbletea.
- `internal/tui/view.go`: The formatting for prepend-spinner to the status message looks clean.
- `internal/tui/update.go`: Flags `m.uploading` and `m.generating` appropriately tied to API lifecycle events and handle state transitions/cancellations effectively.

### 4. Local Worksheets Cleanup

- `internal/tui/commands.go`: `os.ReadDir` and `os.RemoveAll` logic inside `createSubfolderAndUploadCmd` cleanly wipes specific generated files after successful Drive upload without deleting the `worksheets/` directory itself.

### 5. TUI List Pagination

- `internal/tui/view.go`: The array slicing inside `renderPane` is extremely tight. The dynamic `maxVisible` calculation bounded to `m.height` ensures resizing the terminal window doesn't break the UI.

### 6. Spinner Alignment & Emoji Redundancy

- `internal/tui/update.go` and `internal/tui/commands.go`: Removed the `⏳` emoji from all status message string literals. The spinner now aligns correctly with the baseline text font and looks significantly cleaner without redundant emojis.

## Verification Log

```
$ go test ./... -v
✅ All tests pass

$ go build ./cmd/app/main.go
✅ Compiles cleanly
```

## Verdict

**✅ APPROVED — No changes required.**

Phase 4 (including all iterations and bug fixes) is completely signed off. We are formally ready to begin Phase 5: The WhatsApp Daemon.
