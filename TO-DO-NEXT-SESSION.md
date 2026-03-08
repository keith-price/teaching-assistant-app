# Next Session To-Do

**Status:** Phase 5 (The WhatsApp Daemon), including the DRY refactor, is **100% complete and signed off!**

## 1. Phase 6 Initialization (Current Priority)

Begin work on **Phase 6: Final Integration**.
The goal here is to wire everything together into a single, cohesive executable.

Here is your immediate checklist:

- [ ] Create `/cmd/app/main.go`.
- [ ] Initialize the Database (`internal/db`).
- [ ] Initialize the Calendar Fetcher (`internal/calendar`).
- [ ] Initialize the TUI (`internal/tui`).
- [ ] Initialize the AI Generator (`internal/ai`).
- [ ] Initialize the WhatsApp Daemon (`internal/notify`) and attach the cron scheduler.
- [ ] Ensure all systems (TUI blocking, Cron non-blocking) run simultaneously and gracefully shut down on exit.

_Note:_ Since we already have specific standalone tools for authentication (`cmd/auth/` and `cmd/wa_test/`), this main entry point should assume all dependencies (DB, tokens, env, session store) exist and gracefully degrade/prompt the user if they do not.

**First Action for Junior Dev:** Execute the above steps and test the end-to-end workflow (UI usage alongside notification scheduling) locally. Hand it back when you are satisfied with the integration.
