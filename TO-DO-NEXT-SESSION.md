# Next Session To-Do

**Status:** Phase 4 (TUI & UX) is 100% complete and signed off!

## 1. Phase 5 Initialization (Current Priority)

Begin work on **Phase 5: The WhatsApp Daemon**.
The Junior Developer already has `JUNIOR_DEV_BRIEF.md` with instructions for:

- Creating `/internal/notify/`
- Installing `go.mau.fi/whatsmeow`
- Implementing QR login & session management (ensuring `whatsapp_store.db` is git-ignored)
- Providing a `cmd/wa_test/main.go` script to authenticate on terminal.

**First Action:** The Junior Developer should execute the brief, write the test script, and hand it back so the user can scan the WhatsApp QR code.

## 2. Optional UX Polish (Future Consideration)

- **Prompt before deleting local worksheets:** Change the automatic `worksheets/` directory cleanup after a Google Drive upload into an explicit user prompt: "Delete worksheets in folder? (Y/n)".
