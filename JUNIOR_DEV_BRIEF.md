# Phase 5: The WhatsApp Daemon

**To:** Junior Developer
**From:** Senior Developer
**Date:** 2026-03-07

All Phase 4 bug fixes are complete. We are now officially beginning Phase 5. The goal of this phase is to integrate WhatsApp so we can send automated daily schedules to the teacher.

**Your Action Plan:**

1.  **Project Structure:**
    Create a new directory: `/internal/notify/`

2.  **Dependencies:**
    Install the WhatsApp Go library:
    `go get go.mau.fi/whatsmeow`
    _(Note: You will also likely need its dependencies like `go.mau.fi/libsignal/logger` and a sqlite3 driver for its session storage, e.g., `github.com/mattn/go-sqlite3` which we already have)._

3.  **Authentication & Session Management:**
    Create `internal/notify/whatsapp.go`.
    Implement a function to initialize the `whatsmeow` client.
    Crucially, it must support terminal QR code login.

    _Security Requirement:_ The `whatsmeow` client requires a database to store its session keys (so the user doesn't have to scan the QR code every single time they start the app). Configure the client to save its session data to a local file, like `whatsapp_store.db` (or inside the `config/` directory).

    **You must add this session database file to `.gitignore` immediately so the user's WhatsApp keys are never committed to version control.**

4.  **CLI Test Script:**
    Create a temporary script at `cmd/wa_test/main.go` that initializes your new package and triggers the QR code prompt in the terminal.

**Handoff:**
Ensure the dependencies are installed and the code compiles. Give me the brief back once the CLI test script is ready so the user (Keith) can scan the QR code to authenticate the daemon.
