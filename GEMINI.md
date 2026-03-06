# Preply TUI Dashboard

## Project Overview
The Preply TUI Dashboard is a local, keyboard-navigated Terminal User Interface (TUI) application designed for managing daily ESL teaching workflows. 
It features schedule management via Google Calendar, local data ownership via SQLite, AI-assisted worksheet generation via the Gemini API, Google Drive integration for document uploads, and planned automated daily WhatsApp briefings.

### Tech Stack
*   **Language:** Go (Golang)
*   **UI Framework:** Bubble Tea (`github.com/charmbracelet/bubbletea`)
*   **Database:** SQLite (`modernc.org/sqlite`, pure Go)
*   **APIs:** Google Calendar SDK, Google Drive SDK, Official Google GenAI SDK for Go (Gemini API)
*   **Authentication:** Shared OAuth package for Google services
*   **Other:** `godotenv` for secrets, `gomarkdown/markdown` for document conversion, `whatsmeow` (planned for WhatsApp).

### Architecture
The project strictly separates concerns into specific packages:
*   `/cmd/app/` - Main entry point for the TUI application.
*   `/cmd/auth/` - Standalone CLI for handling Google OAuth flows interactively.
*   `/internal/db/` - SQLite connection, schema definitions, and CRUD operations.
*   `/internal/calendar/` - Google Calendar fetching and parsing.
*   `/internal/tui/` - Bubble Tea models, views, and update logic.
*   `/internal/ai/` - Gemini API integration and worksheet generation.
*   `/internal/drive/` - Google Drive upload and folder management.
*   `/internal/auth/` - Shared OAuth token caching and client initialization.
*   `/internal/notify/` - (Planned) WhatsApp connection and scheduled messaging.

## Building and Running

### Prerequisites
1.  **Environment Variables:** Create a `.env` file in the root directory with `GEMINI_API_KEY=your_actual_key_here`.
2.  **Google Credentials:** Place your Google `credentials.json` in the `config/` directory.

### Commands
*   **Authentication:** Before running the main app, authorize Google Calendar and Drive scopes:
    ```bash
    go run cmd/auth/main.go
    ```
*   **Run the Application:**
    ```bash
    go run cmd/app/main.go
    ```
*   **Run Tests:**
    ```bash
    go test ./... -count=1
    ```

## Development Conventions

This project employs a **Dual-Agent Protocol (Actor-Critic methodology)**:
*   **Junior Dev (CLI):** Acts as the Lead Implementer. Responsible for writing code, running local tests, managing dependencies, and documenting code.
*   **Senior Dev (Main Chat):** Acts as the Architectural Overseer & Lead Reviewer. Responsible for reviewing code, ensuring security/privacy standards, validating testing strategy, and signing off on merges.

### Guidelines
1.  **Test-Driven Development:** Unit tests must be written and executed successfully before code is submitted for review.
2.  **Code Reviews:** Feedback from the Senior Dev is stored in the `reviews/` directory. After implementing feedback, update the handoff notes in `PLAN.md`.
3.  **Security & Privacy:** 
    *   Never log or commit Personally Identifiable Information (PII) or API secrets.
    *   Ensure `*.db`, `.env`, `config/credentials.json`, `config/token.json`, and the `worksheets/` directory remain in `.gitignore`.
4.  **Documentation:** Write idiomatic Go documentation (doc comments for exported types/functions) inline as you code.
5.  **Refactoring:** Proactively manage technical debt and ensure clean architectural separation between packages.
