# Phase 7: Daily Briefing Status Toggle

**To:** Junior Developer
**From:** Senior Developer
**Date:** 2026-03-08

## Overview

The Phase 6 WhatsApp Daemon works great, but the user has flagged an architectural issue: the app resets or sends the daily briefing every time the daemon connects or the cron job fires, without tracking if it has _already_ sent a briefing for the current day.

We need to add a "Briefing Sent: [True/False]" toggle directly into the BubbleTea TUI so the user can see its status and manually override it if necessary, and we need the background daemon to respect this state.

## Tasks to Implement

### 1. Database Layer update (`internal/db/`)

Since the briefing is a daily global action (not tied to a specific lesson or student), we need a place to store this state persistently.

- [ ] Create a new table or configuration mechanism in `internal/db/db.go` to store "app state" metrics. A simple key-value `app_settings` table might be best: `(key TEXT PRIMARY KEY, value TEXT)`.
- [ ] Add methods to the `db.Store` to `SetDailyBriefingSent(date string, sent bool)` and `HasDailyBriefingBeenSent(date string) bool`.

### 2. TUI Update (`internal/tui/`)

- [ ] In `view.go`, update the header or footer status bar to display the briefing status for the current day: e.g., `[Briefing Sent: ✅]` or `[Briefing Sent: ❌]`.
- [ ] **Crucial Requirement:** The UI must automatically reflect `❌` (False) at the start of a new day. The status should be keyed to the _current date_ in WIB, not just a global boolean.
- [ ] Pick a keyboard shortcut (e.g., `b`) and update `update.go` to toggle this true/false state in the database for _today's date_ when pressed, immediately reflecting the change in the UI.

### 3. WhatsApp Scheduler Update (`internal/notify/scheduler.go`)

- [ ] Update `Start()`: Add a "catch-up" mechanism. When the app boots, check `HasDailyBriefingBeenSent(today)`. If it hasn't been sent, send it _immediately_ instead of waiting for the 7:00 AM cron tick.
- [ ] Update `SendDailyBriefing(ctx context.Context)`:
  - Query the database using `HasDailyBriefingBeenSent` for today's date.
  - If `true`, return early/silently. Do not send another WhatsApp message.
  - If `false`, send the message through `wa.Client.SendMessage`, and then immediately update the database by calling `SetDailyBriefingSent(today, true)`.

### 4. Google Calendar Auto-Sync (`internal/tui/`, `internal/db/`)

The TUI currently displays `No lessons scheduled.` because it only reads from the local SQLite database. We need to automatically sync down Google Calendar events so the user doesn't have to double-enter their schedule.

- [ ] Modify `cmd/app/main.go` or `internal/tui/Init()` to trigger a sync on launch.
- [ ] Create a `SyncCalendarEvents(store *db.Store, calClient *calendar.Client)` function.
- [ ] This function should:
  - Fetch today's & tomorrow's events from Calendar using the `Preply` keyword.
  - Check if these events already exist in the local SQLite `lessons` table (you may need to add a `calendar_event_id TEXT UNIQUE` column to the `lessons` table in `db.go` to prevent duplicates).
  - If they don't exist, create a generic "Synced Student" in the `students` table (if one doesn't exist yet) and insert the lesson into the local database mappings.
- [ ] Ensure the TUI gracefully re-renders `todayLessons` and `tomorrowLessons` after the sync completes so the panes populate automatically.

Get this implemented and hand it back to me for review!
