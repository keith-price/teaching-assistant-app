# To-Do (Next Session)

## 1. Fix Remaining Google Docs Formatting Issues

The markdown to HTML conversion is working, and the documents are successfully splitting and uploading as native Google Docs. However, the styling in Google Docs still has some issues:

- **Title Formatting:** The title line (`Title: Iran school...`) is rendering as regular body text. It needs to be formatted as an actual `<h1>` or Title style.
- **List Spacing:** The numbered lists (e.g., in the "WARMER" and "KEY VOCABULARY" sections) have too much spacing or incorrect indentation based on the screenshot provided by Keith.
- **Bold Text in Lists:** Some bold formatting in lists (like `____**: a detailed study...`) is rendering the actual asterisks (`**`) instead of bolding the text. The markdown parser might be failing on list items that start with underscores or asterisks.

**Action Item:**
We need to revisit `internal/drive/drive.go` and potentially the prompt in `internal/ai/prompt.md` to ensure the generated markdown is clean, and the CSS injected into the HTML payload perfectly mimics standard Google Docs styling.

---

## 2. Navigate Upwards in Google Drive Folder Picker

Currently, the TUI folder picker only allows selecting a folder and navigating _down_ into subfolders (`enter` key). There is a `backspace` key implemented to navigate up the breadcrumb trail, but Keith reported that it's not working as expected, or the functionality needs to be more explicitly surfaced to the user.

**Action Item:**

- Review `handleFolderPickerKeys` in `internal/tui/update.go`.
- Ensure the `backspace` navigation correctly lists the parent directory.
- Update the TUI footer/hints to clearly show `[backspace] go up`.

---

## 3. Phase 5: The WhatsApp Daemon

Once the bugs above are fixed, we are ready to move on to Phase 5.

- Keith needs to delete `config/token.json` and re-run `go run cmd/auth/main.go` to authorize both Calendar and Drive scopes before we begin.
- Refer to `PLAN.md` for the Phase 5 tasks.
