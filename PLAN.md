# Preply TUI Dashboard - Implementation Plan

This plan breaks down the project detailed in `AGENTS.md` into actionable microtasks. 

**Overview Reference:** The goal is to build a local, keyboard-navigated TUI application for managing ESL teaching workflows using Go, featuring strict package separation (`/cmd/app/`, `/internal/db/`, `/internal/calendar/`, `/internal/tui/`, `/internal/ai/`, `/internal/notify/`) and a Dual-Agent Protocol (Actor-Critic methodology) for code review.

---

## Phase 1: Foundation & Data Layer (SQLite)
*Ref: AGENTS.md > Phase 1*
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
*Ref: AGENTS.md > Phase 2*
- [x] **Task 2.1:** Create `/internal/calendar/` directory.
- [x] **Task 2.2:** Install Google Calendar Go SDK dependencies.
- [x] **Task 2.3:** **SECURITY PAUSE:** Stop and prompt user to download `credentials.json`, place it in the `config/` directory, and confirm it's git-ignored. *Wait for confirmation.*
- [x] **Task 2.4:** Implement OAuth2 authentication flow for Google Calendar.
- [x] **Task 2.5:** Implement event fetching logic for "Today" and "Tomorrow" filtering by a specific keyword.
- [x] **Task 2.6:** **Constraint Check:** Ensure all fetched events are strictly parsed and converted to WIB (Western Indonesian Time).
- [x] **Task 2.7:** **🛑 CHECKPOINT & REVIEW:** Hand off `/internal/calendar/` to the Senior Dev for review. Implement requested changes before proceeding.

> **Junior Dev Handoff Note (Phase 2 Complete):**
> Phase 2 is fully completed and all code review feedback from `reviews/TASK_2_FEEDBACK.md` has been successfully implemented. 
> Key changes made during review:
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
*Ref: AGENTS.md > Phase 3*
- [ ] **Task 3.1:** Create `/internal/ai/` directory.
- [ ] **Task 3.2:** Install dependencies (`github.com/joho/godotenv`, `github.com/atotto/clipboard`, Official Google GenAI SDK for Go).
- [ ] **Task 3.3:** **SECURITY PAUSE:** Implement `godotenv` loading. Stop and prompt user to create `.env` with `GEMINI_API_KEY`, and confirm it's git-ignored. *Wait for confirmation.*
- [ ] **Task 3.4:** Implement clipboard reading using `atotto/clipboard`.
- [ ] **Task 3.5:** Implement Gemini API client initialization.
- [ ] **Task 3.6:** Construct Gemini API prompt combining System Prompt, lesson metadata, and clipboard transcript.
- [ ] **Task 3.7:** Implement API call to Gemini and parse the response.
- [ ] **Task 3.8:** Implement saving the API response as a `.md` file in a local `/worksheets/` directory. Ensure directory exists.
- [ ] **Task 3.9:** **🛑 CHECKPOINT & REVIEW:** Hand off `/internal/ai/` to the Senior Dev for review. Implement requested changes before proceeding.

---

## Phase 4: The Bubble Tea Interface
*Ref: AGENTS.md > Phase 4*
- [ ] **Task 4.1:** Create `/internal/tui/` directory.
- [ ] **Task 4.2:** Install `github.com/charmbracelet/bubbletea` dependency.
- [ ] **Task 4.3:** Initialize Bubble Tea model, Init, Update, and View functions.
- [ ] **Task 4.4:** Implement the split-pane layout (Today on left, Tomorrow on right).
- [ ] **Task 4.5:** Implement keyboard navigation (Arrow keys to move between panes/items).
- [ ] **Task 4.6:** Implement action toggle logic (Press 'V' to toggle `vocab_sent` in the database for the selected lesson).
- [ ] **Task 4.7:** Implement AI Generation trigger (Press 'G' to open a text input form).
- [ ] **Task 4.8:** Integrate Phase 3 logic into 'G' action: read clipboard, call AI, link generated `.md` file to the lesson in the database.
- [ ] **Task 4.9:** **🛑 CHECKPOINT & REVIEW:** Hand off `/internal/tui/` to the Senior Dev for review. Implement requested changes before proceeding.

---

## Phase 5: The WhatsApp Daemon
*Ref: AGENTS.md > Phase 5*
- [ ] **Task 5.1:** Create `/internal/notify/` directory.
- [ ] **Task 5.2:** Install `go.mau.fi/whatsmeow` dependency.
- [ ] **Task 5.3:** **SECURITY PAUSE:** Implement terminal QR code login for `whatsmeow`. Ensure session database is git-ignored. Stop and prompt user to run daemon and scan QR code. *Wait for confirmation.*
- [ ] **Task 5.4:** Implement function to query DB and Calendar to format the daily message string ("Good morning! Here is today's schedule...").
- [ ] **Task 5.5:** Implement cron-like scheduler to trigger the formatted message daily.
- [ ] **Task 5.6:** **🛑 CHECKPOINT & REVIEW:** Hand off `/internal/notify/` to the Senior Dev for review. Implement requested changes before proceeding.

---

## Phase 6: Final Integration
*Ref: AGENTS.md > Core Architecture & Package Structure*
- [ ] **Task 6.1:** Create `/cmd/app/main.go`.
- [ ] **Task 6.2:** Wire up all packages (`db`, `calendar`, `tui`, `notify`) in `main.go`.
- [ ] **Task 6.3:** Final end-to-end testing of the complete workflow.
- [ ] **Task 6.4:** Final polish and clean up.