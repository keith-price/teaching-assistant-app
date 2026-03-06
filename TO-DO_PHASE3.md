# Phase 3: AI & Clipboard Pipeline — Junior Dev Instructions

**Date:** 2026-03-06  
**Assigned Phase:** Phase 3 (Tasks 3.1–3.9)  
**Status:** NOT STARTED — Awaiting permission to begin

---

## ⚠️ Before You Write a Single Line of Code

1. **Read `AGENTS.md`** — Understand the full project architecture, the tech stack, the dual-agent review protocol, and the review conventions. Pay special attention to:
   - Section 2 (Tech Stack & Libraries) for the exact dependencies you'll use.
   - Section 4 (Dual-Agent Workflow Protocol) for how your work will be reviewed.
   - Section 5, Phase 3 for the feature specification.

2. **Read `PLAN.md`** — Review the entire plan, not just Phase 3. Understand:
   - What was built in Phases 1 & 2 (especially the `Store` struct pattern in `/internal/db/` and the `EventFetcher` interface pattern in `/internal/calendar/`).
   - The handoff notes from Phases 1 & 2 — these document key architectural decisions you must follow.
   - The exact task list for Phase 3 (Tasks 3.1–3.9).

3. **Review the existing code** — Before creating `/internal/ai/`, study the patterns already established:
   - `internal/db/db.go` — How the `Store` struct wraps the database connection.
   - `internal/calendar/calendar.go` — How the `EventFetcher` interface was created for testability.
   - Follow these same patterns (struct-based design, interfaces for external dependencies, proper error handling).

4. **Review the Senior Dev feedback** — Read `reviews/TASK_1_FEEDBACK.md` and `reviews/TASK_2_FEEDBACK.md` to understand the quality bar and avoid repeating past mistakes.

5. **Clarify your tasks** — Write a brief summary of what you plan to do today and present it to Keith for approval. **Do not begin coding until you receive explicit permission.**

---

## Your Task List

Complete these in order. Do **not** skip ahead.

### Task 3.1 — Create `/internal/ai/` directory
- Create the package directory.
- Create the initial `ai.go` file with `package ai`.

### Task 3.2 — Install dependencies
Install all three required dependencies:
- `github.com/joho/godotenv` — for loading `.env` files.
- `github.com/atotto/clipboard` — for reading OS clipboard.
- The **Official Google GenAI SDK for Go** (Gemini API) — check the latest import path. Do NOT use a third-party wrapper.

### Task 3.3 — 🔒 SECURITY PAUSE
This is a mandatory stop.
1. Implement `godotenv` to load the `.env` file from the project root.
2. **STOP and prompt Keith:**
   > *"Please create a `.env` file in the project root and add `GEMINI_API_KEY=your_actual_key_here`. Confirm it is ignored by git. Do not paste your key in this chat."*
3. **Wait for Keith's confirmation before proceeding.** Do not continue until he confirms.
4. Verify that `.env` is already listed in `.gitignore` (it is — but confirm it).

### Task 3.4 — Implement clipboard reading
- Create a function that reads raw text from the OS clipboard using `atotto/clipboard`.
- Handle the case where the clipboard is empty or unreadable — return a clear error, do not silently proceed.
- Consider creating an interface (like `EventFetcher` in Phase 2) so clipboard reading is mockable in tests.

### Task 3.5 — Implement Gemini API client initialisation
- Load the API key from the environment variable `GEMINI_API_KEY`.
- Initialise the Gemini client.
- Follow the `Store` struct pattern — wrap the client in a struct (e.g., `AIClient` or `Generator`).
- Handle missing API key gracefully with a clear error message.

### Task 3.6 — Construct the Gemini prompt
- Build a prompt that combines:
  1. A **System Prompt** defining the AI's role (e.g., "You are an ESL teaching assistant that generates student worksheets...").
  2. **Lesson metadata** — student level, lesson time, student name.
  3. **Clipboard transcript** — the raw text read from the clipboard.
- The system prompt should instruct Gemini to output a well-structured Markdown worksheet.
- Keep the prompt template as a constant or configurable string, not buried in logic.

### Task 3.7 — Implement the API call and parse the response
- Send the constructed prompt to Gemini.
- Parse and extract the text response.
- Handle API errors gracefully (rate limits, network failures, empty responses).
- Do NOT use `fmt.Printf` or `fmt.Println` for logging — this will break the TUI in Phase 4. Use `log` or return errors.

### Task 3.8 — Save response as a `.md` file
- Save the worksheet to `/worksheets/` directory.
- Auto-create the `/worksheets/` directory if it doesn't exist (`os.MkdirAll`).
- Use a descriptive filename format, e.g. `worksheets/2026-03-06_StudentName_lesson.md`.
- Return the file path so Phase 4 can link it to the lesson in the database.
- **Decision needed:** Confirm with Keith whether `/worksheets/` should be added to `.gitignore` or tracked in git.

### Task 3.9 — 🛑 CHECKPOINT & REVIEW
- Write unit tests in `internal/ai/ai_test.go`.
- Ensure all tests pass (`go test ./internal/ai/`).
- Update `PLAN.md`: mark Tasks 3.1–3.8 as `[x]` and add a handoff note block (follow the same format as Phases 1 & 2).
- Hand off `/internal/ai/` to the Senior Dev for review.

---

## Quality Checklist (Apply to Every File)

- [ ] No global variables — use struct-based design.
- [ ] Interfaces for external dependencies (clipboard, Gemini API) — for testability.
- [ ] All errors are returned, never silently swallowed.
- [ ] No `fmt.Print*` calls — these break the TUI.
- [ ] `rows.Err()` checks after any iteration (if applicable).
- [ ] Functions are small and single-purpose.
- [ ] Code comments explain *why*, not *what*.
- [ ] Tests cover happy path AND error cases.

---

## Reminder

**Do not start coding until you have:**
1. Read `AGENTS.md`, `PLAN.md`, and both review feedback files.
2. Summarised your understanding of today's tasks back to Keith.
3. Received Keith's explicit permission to begin.

Good luck. Build it clean.
