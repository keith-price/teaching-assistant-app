# Preply TUI Dashboard - Implementation Plan

This plan breaks down the project detailed in `AGENTS.md` into actionable microtasks.

**Overview Reference:** The goal is to build a local, keyboard-navigated TUI application for managing ESL teaching workflows using Go, featuring strict package separation (`/cmd/app/`, `/internal/db/`, `/internal/calendar/`, `/internal/tui/`, `/internal/ai/`, `/internal/notify/`) and a Dual-Agent Protocol (Actor-Critic methodology) for code review.

---

## Phase 1: Foundation & Data Layer (SQLite)

_Ref: AGENTS.md > Phase 1_

- [x] **Task 1.1:** Initialize the Go module (`go mod init`).
- [x] **Task 1.2:** Create the directory structure for `/internal/db/`.
- [x] **Task 1.3:** **SECURITY STEP 1:** Create `.gitignore` in project root. Add `.env`, `credentials.json`, `token.json`, and `*.db`.
- [x] **Task 1.4:** Install SQLite dependency (`go get github.com/mattn/go-sqlite3`).
- [x] **Task 1.5:** Design database schema (`students`, `lessons`, `materials` tables).
- [x] **Task 1.6:** Implement database connection function.
- [x] **Task 1.7:** Implement CRUD operations for `students`.
- [x] **Task 1.8:** Implement CRUD operations for `lessons`.
- [x] **Task 1.9:** Implement CRUD operations for `materials`.
- [x] **Task 1.10:** Write a CLI test script to verify database connectivity and CRUD operations.
- [x] **Task 1.11:** **🛑 CHECKPOINT & REVIEW:** Hand off `/internal/db/` to the Senior Dev for review. Implement requested changes before proceeding.

> **Junior Dev Handoff Note (Phase 1 Complete):**
> Phase 1 is fully completed and all code review feedback from `reviews/TASK_1_FEEDBACK.md` has been successfully implemented.
> Key changes made during review:
>
> - Removed global `DB` variable; introduced `Store` struct for better testability and concurrency.
> - Added missing `rows.Err()` checks after all iterations.
> - Enabled `PRAGMA foreign_keys = ON` for SQLite.
> - Added `RowsAffected()` checks on all UPDATE/DELETE queries.
> - Converted `ToggleVocabSent` to a true boolean toggle.
> - Added `GetLessonsByDateRange` preemptively for Phase 4 split-pane support.
> - Replaced the CLI test script with idiomatic Go tests in `internal/db/db_test.go`.
>
> **Next steps:** Start working on **Phase 2: Google Calendar Fetcher**, beginning with Task 2.1.

---

## Phase 2: Google Calendar Fetcher

_Ref: AGENTS.md > Phase 2_

- [x] **Task 2.1:** Create `/internal/calendar/` directory.
- [x] **Task 2.2:** Install Google Calendar Go SDK dependencies.
- [x] **Task 2.3:** **SECURITY PAUSE:** Stop and prompt user to download `credentials.json`, place it in the `config/` directory, and confirm it's git-ignored. _Wait for confirmation._
- [x] **Task 2.4:** Implement OAuth2 authentication flow for Google Calendar.
- [x] **Task 2.5:** Implement event fetching logic for "Today" and "Tomorrow" filtering by a specific keyword.
- [x] **Task 2.6:** **Constraint Check:** Ensure all fetched events are strictly parsed and converted to WIB (Western Indonesian Time).
- [x] **Task 2.7:** **🛑 CHECKPOINT & REVIEW:** Hand off `/internal/calendar/` to the Senior Dev for review. Implement requested changes before proceeding.

> **Junior Dev Handoff Note (Phase 2 Complete):**
> Phase 2 is fully completed and all code review feedback from `reviews/TASK_2_FEEDBACK.md` has been successfully implemented.
> Key changes made during review:
>
> - Modified error handling in `parseEventTime` loops to return errors instead of silently skipping events.
> - Added `.MaxResults(100)` to the Google Calendar API query to ensure explicit limits.
> - Updated `PLAN.md` references from project root to the `config/` directory for `credentials.json`.
> - Created an `EventFetcher` interface inside `calendar.go` to make `FetchEvents` mockable for future tests.
> - Documented the acceptable use of a static state token for local OAuth.
> - Removed a TUI-breaking `fmt.Printf` log from the token-saving logic.
>
> **Next steps:** Start working on **Phase 3: AI & Clipboard Pipeline**, beginning with Task 3.1.

---

## Phase 3: AI & Clipboard Pipeline

_Ref: AGENTS.md > Phase 3_

- [x] **Task 3.1:** Create `/internal/ai/` directory.
- [x] **Task 3.2:** Install dependencies (`github.com/joho/godotenv`, `github.com/atotto/clipboard`, Official Google GenAI SDK for Go).
- [x] **Task 3.3:** **SECURITY PAUSE:** Implement `godotenv` loading. Stop and prompt user to create `.env` with `GEMINI_API_KEY`, and confirm it's git-ignored. _Wait for confirmation._
- [x] **Task 3.4:** Implement clipboard reading using `atotto/clipboard`.
- [x] **Task 3.5:** Implement Gemini API client initialization.
- [x] **Task 3.6:** Construct Gemini API prompt combining System Prompt, lesson metadata, and clipboard transcript.
- [x] **Task 3.7:** Implement API call to Gemini and parse the response.
- [x] **Task 3.8:** Implement saving the API response as a `.md` file in a local `/worksheets/` directory. Ensure directory exists. **Add `/worksheets/` to `.gitignore`** — these files contain private student data and must never be committed.

> _[Decision — 2026-03-06] Confirmed with Keith: the `/worksheets/` directory is git-ignored. Generated worksheets contain private student lesson transcripts and personalised AI content. They must remain local-only._

- [x] **Task 3.9:** **🛑 CHECKPOINT & REVIEW:** Hand off `/internal/ai/` to the Senior Dev for review. Implement requested changes before proceeding.

> **Junior Dev Handoff Note (Phase 3 Complete):**
> Phase 3 development is complete.
> Key implementation details:
>
> - Created `internal/ai/ai.go` using a struct-based design (`Generator`).
> - Implemented a `ClipboardReader` interface to ensure the clipboard reading is fully testable and mockable.
> - Handled empty clipboard states explicitly (returning an error instead of passing empty text to Gemini).
> - Created a centralized `SystemPrompt` constant to cleanly build the request to the official Google GenAI Go SDK.
> - Auto-creates the `/worksheets/` directory and securely writes the generated markdown files. Added `/worksheets/` to `.gitignore` as requested.
> - Tested successfully via `go test ./internal/ai/` utilizing a mock clipboard.
>
> **Junior Dev Handoff Note (Phase 3 Review Fixes Complete):**
> All code review feedback from `reviews/TASK_3_FEEDBACK.md` has been successfully implemented.
> Key changes made during review:
>
> - Migrated from prompt concatenation to using `SystemInstruction: genai.NewContentFromText(SystemPrompt, genai.RoleUser)`.
> - Kept the `prompt.md` filename casing consistent to avoid Linux/CI `go:embed` errors.
> - Prevented `SaveWorksheet` from overwriting same-day duplicates by appending a time component `2006-01-02_1504` to the filename.
> - Required `baseDir` in `SaveWorksheet` avoiding fragile relative directory creation and used `t.TempDir()` in tests to prevent pollution.
> - Added missing `lessonTime` and `lessonType` into `constructPrompt` matching the system prompt parameters.
> - Made `DefaultModel = "gemini-2.5-flash"` configurable via a package-level constant.
> - Safely verified `result.Candidates` length before calling `result.Text()`.
>
> **Next steps:** Waiting for Senior Dev review of Phase 3 (`/internal/ai/`).

---

## Phase 4: The Bubble Tea Interface

_Ref: AGENTS.md > Phase 4_

- [x] **Task 4.1:** Create `/internal/tui/` directory.
- [x] **Task 4.2:** Install `github.com/charmbracelet/bubbletea` dependency.
- [x] **Task 4.3:** Initialize Bubble Tea model, Init, Update, and View functions.
- [x] **Task 4.4:** Implement the split-pane layout (Today on left, Tomorrow on right).
- [x] **Task 4.5:** Implement keyboard navigation (Arrow keys to move between panes/items).
- [x] **Task 4.6:** Implement action toggle logic (Press 'V' to toggle `vocab_sent` in the database for the selected lesson).
- [x] **Task 4.7:** Implement AI Generation trigger (Press 'G' to open a text input form).
- [x] **Task 4.8:** Integrate Phase 3 logic into 'G' action: read clipboard, call AI, link generated `.md` file to the lesson in the database.
- [x] **Task 4.9:** **🛑 CHECKPOINT & REVIEW:** Hand off `/internal/tui/` to the Senior Dev for review. Implement requested changes before proceeding.

> **Junior Dev Handoff Note (Phase 4 Refactor Complete):**
> Phase 4 development and subsequent refactoring are complete.
> Key implementation details from Refactor:
>
> - Decoupled worksheets from individual students and lessons, making them standalone teaching assets.
> - Removed OS clipboard dependency; input is now provided via a TUI textarea.
> - AI generation produces two separate documents (Student Worksheet and Teacher Key) from a single Gemini API response using prompt delimiters.
> - Implemented Google Drive integration (`internal/drive/`) allowing users to upload generated documents to selected Drive folders.
> - Created a shared `internal/auth/` package for consistent OAuth flow across Google Calendar and Drive.
> - Overhauled the TUI form with a comprehensive 5-field standalone interface (Title, Level, Type, Duration, Text).
> - Added preview and folder selection workflows wrapped safely in asynchronous `tea.Cmd`.
>
> **Next steps:** Waiting for Senior Dev review of Phase 4 (`internal/tui/`, `internal/ai/`, `internal/drive/`, `internal/auth/`).

> **Junior Dev Handoff Note (Phase 4 Refactor Bug Fixes Complete):**
> All critical bugs identified in Task R.8 by the Senior Dev have been fixed:
>
> - **Bug 1 (OAuth Deadlock):** Refactored `internal/auth/auth.go` to safely load cached tokens inside the TUI (`GetHTTPClient`) without prompting. Created a separate interactive browser flow (`AuthorizeInteractively`) triggered via a new standalone CLI (`cmd/auth/main.go`). The app no longer freezes if tokens are missing.
> - **Bug 2 (Enter Key):** Fixed the `case "enter"` block in `internal/tui/update.go` to allow newlines in the source text area.

> **Junior Dev Handoff Note (Drive Folder Picker Enhancements Complete):**
> All tasks for improving the Google Drive integration have been successfully implemented:
>
> - Upgraded OAuth scope from `drive.file` to the full `drive` scope in `main.go` and `cmd/auth/main.go` to enable folder browsing.
> - Added `CreateFolder` API to the Drive client (`internal/drive/drive.go`).
> - Enhanced the TUI folder picker to support breadcrumb navigation, allowing users to drill into folders and navigate back up.
> - Added inline folder creation functionality (press `n` to create a new folder directly within the picker).
> - Updated the upload logic so that selecting a target folder now automatically creates a lesson-specific subfolder (using a sanitized title) and uploads both the worksheet and teacher key into it.
>
> **Next steps:** Handing back to Senior Dev for final verification of Phase 4 refactor, bug fixes, and Drive folder picker improvements.

> ✅ Phase 4 (including Drive Folder Picker enhancements) **SIGNED OFF**.
>
> **What was reviewed:**
>
> - `internal/auth/auth.go` — `GetHTTPClient` no longer calls `fmt.Scan`; `AuthorizeInteractively` correctly isolated for CLI-only use.
> - `cmd/auth/main.go` — Clean standalone OAuth CLI tool with both Calendar + Drive scopes.
> - `cmd/app/main.go` — Graceful degradation when token is missing; scope upgraded to `driveAPI.DriveScope`.
> - `internal/tui/update.go` — Enter key fix; `handleFolderPickerKeys` correctly handles breadcrumbs, folder creation, and subfolder upload.
> - `internal/ai/ai.go` — Clipboard dependency removed. `splitResponse`, `SaveDocuments`, and `constructPrompt` all refactored correctly.
> - `internal/calendar/calendar.go` — Successfully migrated to shared `auth` package.
> - `internal/drive/drive.go` — Clean implementation with `ListFolders`, `FindFolder`, `UploadFile`, and `CreateFolder`.
> - `internal/tui/model.go` — `folderBreadcrumb` type and new state fields (`folderPath`, `showCreateFolder`, `createFolderInput`) are correct.
> - `internal/tui/commands.go` — `createFolderCmd` and `createSubfolderAndUploadCmd` both correct; old `uploadToDriveCmd` fully removed.
> - `internal/tui/view.go` — Breadcrumb bar, create-folder input, empty-state message, and updated help text all correct.
> - `internal/tui/tui_test.go` — Tests cover G key, selector cycling, preview state transitions, folder picker navigation, and folder creation trigger.
>
> **Verification results:**
>
> - `go build ./...` — ✅ compiles cleanly
> - `go test ./... -count=1` — ✅ all tests pass
>
> **Full review:** See `reviews/TASK_4_FEEDBACK.md` for detailed line-by-line notes.
>
> **Next steps:** Ready for **Phase 5: The WhatsApp Daemon** (Task 5.1). Before starting Phase 5, Keith should delete `config/token.json` and re-run `go run cmd/auth/main.go` to authorize both Calendar and Drive scopes.

---

## Phase 5: The WhatsApp Daemon

_Ref: AGENTS.md > Phase 5_

- [ ] **Task 5.1:** Create `/internal/notify/` directory.
- [ ] **Task 5.2:** Install `go.mau.fi/whatsmeow` dependency.
- [ ] **Task 5.3:** **SECURITY PAUSE:** Implement terminal QR code login for `whatsmeow`. Ensure session database is git-ignored. Stop and prompt user to run daemon and scan QR code. _Wait for confirmation._
- [ ] **Task 5.4:** Implement function to query DB and Calendar to format the daily message string ("Good morning! Here is today's schedule...").
- [ ] **Task 5.5:** Implement cron-like scheduler to trigger the formatted message daily.
- [ ] **Task 5.6:** **🛑 CHECKPOINT & REVIEW:** Hand off `/internal/notify/` to the Senior Dev for review. Implement requested changes before proceeding.

---

## Phase 6: Final Integration

_Ref: AGENTS.md > Core Architecture & Package Structure_

- [ ] **Task 6.1:** Create `/cmd/app/main.go`.
- [ ] **Task 6.2:** Wire up all packages (`db`, `calendar`, `tui`, `notify`) in `main.go`.
- [ ] **Task 6.3:** Final end-to-end testing of the complete workflow.
- [ ] **Task 6.4:** Final polish and clean up.
