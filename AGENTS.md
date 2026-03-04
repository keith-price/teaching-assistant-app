\# Project Master Plan: Preply TUI Dashboard (Dual-Agent Protocol)



\## 1. Project Overview

\*\*Goal:\*\* Build a local, keyboard-navigated Terminal User Interface (TUI) application for managing daily ESL teaching workflows. 

\*\*Primary User:\*\* An ESL teacher managing Preply schedules, requiring local data ownership, AI-assisted worksheet generation, and automated daily WhatsApp briefings.



\## 2. Tech Stack \& Libraries

\* \*\*Language:\*\* Go (Golang)

\* \*\*UI Framework:\*\* `github.com/charmbracelet/bubbletea`

\* \*\*Database:\*\* `github.com/mattn/go-sqlite3` (Local SQLite)

\* \*\*Calendar API:\*\* Official Google Calendar Go SDK

\* \*\*AI Integration:\*\* Official Google GenAI SDK for Go (Gemini API) \& `github.com/joho/godotenv` (for secret management)

\* \*\*Clipboard Management:\*\* `github.com/atotto/clipboard`

\* \*\*WhatsApp API:\*\* `go.mau.fi/whatsmeow`



\## 3. Core Architecture \& Package Structure

The app must be strictly separated into isolated packages to facilitate clean AI generation and review.

\* `/cmd/app/` - The main entry point.

\* `/internal/db/` - SQLite connection, schema, and CRUD.

\* `/internal/calendar/` - Google Calendar OAuth and fetching.

\* `/internal/tui/` - Bubble Tea models, views, and update loops.

\* `/internal/ai/` - Gemini API calls, `.env` loading, and clipboard integration.

\* `/internal/notify/` - WhatsApp connection and scheduled messaging.



\## 4. The Dual-Agent Workflow Protocol

This project will be built using an \*\*Actor-Critic\*\* methodology. 

\* \*\*The Actor (Junior Dev):\*\* The AI running \*inside\* the IDE. Responsible for writing the code, running local tests, and prompting the user for manual security steps.

\* \*\*The Critic (Senior Dev):\*\* A separate AI instance running \*outside\* the IDE. Responsible for reviewing the code against Go best practices and generating a Markdown feedback report.



\*\*The Loop:\*\* For every Phase below, the Junior will build the feature. The Junior will ask the user to test it. The user will pass the code to the Senior for review. The Junior will implement the Senior's requested changes before moving to the next Phase.



---



\## 5. Phase-by-Phase Implementation Plan



\### Phase 1: Foundation \& Data Layer (SQLite)

1\.  Initialize the Go module and create `/internal/db/`.

2\.  \*\*SECURITY STEP 1:\*\* Junior Dev must create a `.gitignore` file in the project root and immediately add `.env`, `credentials.json`, `token.json`, and `\*.db` to it. 

3\.  Design a schema with tables for: `students`, `lessons`, and `materials`.

4\.  Write basic CRUD functions to insert and retrieve this data.

5\.  Write a quick CLI test script to verify the database works.

6\.  \*\*🛑 CHECKPOINT \& REVIEW:\*\* Pass `/internal/db/` to the Senior Dev.



\### Phase 2: Google Calendar Fetcher

1\.  Create `/internal/calendar/` and implement the OAuth2 flow.

2\.  \*\*SECURITY PAUSE:\*\* Junior Dev must stop and prompt the user: \*"Please download your Google Calendar `credentials.json` and place it in the root folder. Confirm it is ignored by git. Do not paste the contents here."\* Wait for user confirmation.

3\.  Create a function to fetch "Today's" and "Tomorrow's" events containing a specific keyword.

4\.  \*\*Critical Timezone Constraint:\*\* All fetched events must be strictly parsed and converted to \*\*WIB (Western Indonesian Time)\*\*. Do not rely on UTC defaults.

5\.  \*\*🛑 CHECKPOINT \& REVIEW:\*\* Pass `/internal/calendar/` to the Senior Dev.



\### Phase 3: AI \& Clipboard Pipeline

1\.  Create `/internal/ai/`.

2\.  \*\*SECURITY PAUSE:\*\* Junior Dev must implement `godotenv` to load environment variables, then stop and prompt the user: \*"Please create a `.env` file in your root directory and add `GEMINI\_API\_KEY=your\_actual\_key\_here`. Confirm it is ignored by git. Do not paste your key in this chat."\* Wait for user confirmation.

3\.  Implement `atotto/clipboard` to read raw transcripts from the OS clipboard.

4\.  \*\*Prompt Injection:\*\* Construct a Gemini API payload combining the System Prompt, the lesson metadata (level/time), and the clipboard transcript.

5\.  Save the API response as a `.md` file in a local `/worksheets/` directory.

6\.  \*\*🛑 CHECKPOINT \& REVIEW:\*\* Pass `/internal/ai/` to the Senior Dev.



\### Phase 4: The Bubble Tea Interface

1\.  Create `/internal/tui/` and build a split-pane layout (Today on the left, Tomorrow on the right).

2\.  Implement keyboard navigation (Arrow keys).

3\.  Implement action toggles (Pressing 'V' toggles `vocab\_sent` in the database).

4\.  Implement the AI Generation trigger (Pressing 'G' opens a text input form, reads the clipboard, calls Phase 3, and links the file to the lesson).

5\.  \*\*🛑 CHECKPOINT \& REVIEW:\*\* Pass `/internal/tui/` to the Senior Dev.



\### Phase 5: The WhatsApp Daemon

1\.  Create `/internal/notify/` using `whatsmeow`.

2\.  \*\*SECURITY PAUSE:\*\* Junior Dev must implement the terminal QR code login, save the session locally (ensure the session database is in `.gitignore`), and prompt the user: \*"Please run the daemon and scan the QR code with your phone to authenticate."\*

3\.  Create a formatting function that queries the DB and Calendar to build a text string: \*"Good morning! Here is today's schedule..."\*

4\.  Implement a cron-like scheduler to trigger this message daily.

5\.  \*\*🛑 CHECKPOINT \& REVIEW:\*\* Pass `/internal/notify/` to the Senior Dev.

