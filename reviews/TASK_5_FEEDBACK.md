# Task 5 Feedback — The WhatsApp Daemon

**Reviewer:** Senior Dev
**Date:** 2026-03-08
**Scope:** WhatsApp client initialization, background cron scheduling, and message generation.

---

## Files Reviewed

| File                           | Verdict         |
| ------------------------------ | --------------- |
| `internal/notify/whatsapp.go`  | ✅ Pass         |
| `internal/notify/scheduler.go` | 🛠️ Needs Change |
| `cmd/wa_test/main.go`          | ✅ Pass         |

## Review Notes

### `internal/notify/whatsapp.go` & Dependencies

- **Excellent architectural decision** using `modernc.org/sqlite` over `mattn/go-sqlite3`. This maintains our strict CGO-free requirement from Phase 1.
- `InitWhatsApp` successfully provisions the local session database and the directory creation is robust.
- The `.gitignore` update for `config/whatsapp_store.db*` is correct and verified. This prevents the user's WhatsApp session keys from leaking.
- `cmd/wa_test/main.go` functions exactly as required for the initial terminal QR login.

### `internal/notify/scheduler.go`

- **Timezone:** Correctly hardcoded to WIB (`UTC+7`) for both the Cron scheduler and the `time.Now()` evaluations.
- **Message Construction (`BuildDailyScheduleMessage`):** The logic flows nicely, separating Calendar events and DB lessons.

**DRY Principle Violation Identified:**
In `BuildDailyScheduleMessage`, you have written:

```go
todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
tomorrowStart := todayStart.AddDate(0, 0, 1)
```

And then you manually iterate over lessons and call `s.dbStore.GetStudent(lesson.StudentID)`.

If you look at `internal/tui/commands.go` in `fetchLessonsCmd`, we already have this exact date boundary logic and the exact same `n+1` student fetching loop to populate the TUI panes.

While the _string formatting_ (TUI lipgloss vs WhatsApp plain text) is necessarily different, the _data fetching_ is identical.

## Requested Changes

1. **Refactor Lesson Fetching (DRY):**
   Please create a shared helper function in `internal/db/` or `internal/tui/` (or a new shared package if necessary, though `db` makes the most sense if it just returns a struct combining Lesson and Student) that returns the fully populated schedule for a given day.

   Currently, both `fetchLessonsCmd` (in the TUI) and `BuildDailyScheduleMessage` (in the Notify daemon) are doing the exact same `GetLessonsByDateRange` followed by a loop calling `GetStudent`. This needs to be consolidated so we aren't repeating the student hydration logic.

## Verification Log

```bash
$ go build ./...
✅ Compiles cleanly

$ go test ./... -count=1
✅ All tests pass
```

## Verdict

**⚠️ CHANGES REQUESTED**

Please fix the DRY violation regarding the daily lesson + student hydration logic. Once fixed, update this feedback document with your handoff note and I will re-review.

> **Junior Dev Handoff Note (Phase 5 Refactor Complete):**
> I have fixed the DRY violation by doing the following:
>
> 1.  Added a new struct `LessonWithStudent` to `internal/db/lessons.go`.
> 2.  Added a new method `GetLessonsWithStudentByDateRange(start, end time.Time) ([]LessonWithStudent, error)` to `internal/db/lessons.go`. This method fetches the lessons and handles the `n+1` student hydration in one place.
> 3.  Refactored `fetchLessonsCmd` in `internal/tui/commands.go` to use this new database method, entirely removing the `tui.LessonView` struct in favor of the shared `db.LessonWithStudent`.
> 4.  Refactored `BuildDailyScheduleMessage` in `internal/notify/scheduler.go` to use the new database method, removing the manual looping and `GetStudent` calls.
> 5.  All tests pass (`go test ./...`) and the project builds successfully.
>
> Handing this back to you for final verification and sign-off!

---

## Final Verification (Senior Dev)

**Date:** 2026-03-08

The DRY violation has been successfully resolved. `internal/db/lessons.go` now properly handles the student hydration logic centrally via `GetLessonsWithStudentByDateRange`. Both the TUI and WhatsApp Notification modules consume this effectively.

```bash
$ go build ./...
✅ Compiles cleanly

$ go test ./... -count=1
✅ All tests pass
```

**✅ Phase 5 is SIGNED OFF.**

We are ready to proceed to Phase 6: Final Integration.
