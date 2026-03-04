# Phase 1 Code Review — Senior Dev Feedback

**Scope:** `internal/db/` package + `cmd/testdb/main.go`
**Verdict:** Solid foundation — needs fixes before moving to Phase 2.

---

## 🔴 Critical (Must Fix)

### 1. Global `DB` variable is a concurrency & testability hazard
**File:** `db.go:11`

A package-level `var DB *sql.DB` creates tight coupling. Every CRUD function reaches into global state, which makes unit testing painful and introduces risks when this evolves into a TUI app with concurrent goroutines (e.g. the WhatsApp daemon in Phase 5).

**Requested change:** Wrap the connection in a struct and make CRUD functions methods on it.

```go
// db.go
type Store struct {
    db *sql.DB
}

func NewStore(dataSourceName string) (*Store, error) {
    conn, err := sql.Open("sqlite", dataSourceName)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    if err = conn.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    s := &Store{db: conn}
    return s, s.createTables()
}

func (s *Store) Close() {
    if s.db != nil {
        s.db.Close()
    }
}
```

Then CRUD functions become methods: `func (s *Store) CreateStudent(...)`, etc.

---

### 2. Missing `rows.Err()` check after iteration
**Files:** `students.go:68`, `lessons.go:70`, `materials.go:48`

Every `for rows.Next()` loop **must** be followed by a `rows.Err()` check. Without it, a mid-iteration I/O error is silently swallowed.

**Requested change:** Add after every `rows.Next()` loop:

```go
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("error iterating rows: %w", err)
}
```

---

### 3. Foreign key enforcement is off by default in SQLite
**File:** `db.go`

SQLite does **not** enforce `FOREIGN KEY` constraints unless you explicitly enable them per connection. Your schema declares them, but they are currently no-ops.

**Requested change:** Execute this immediately after opening the connection:

```go
if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
    return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
}
```

---

## 🟡 Important (Should Fix)

### 4. `UpdateStudent` doesn't verify the row exists
**File:** `students.go:73-81`

`DB.Exec` on `UPDATE ... WHERE id = ?` with a non-existent ID succeeds silently (0 rows affected). The caller has no way to distinguish "updated" from "not found".

**Requested change:** Check `RowsAffected()`:

```go
func (s *Store) UpdateStudent(id int64, name, level, contactInfo string) error {
    query := `UPDATE students SET name = ?, level = ?, contact_info = ? WHERE id = ?`
    result, err := s.db.Exec(query, name, level, contactInfo, id)
    if err != nil {
        return fmt.Errorf("failed to update student: %w", err)
    }
    n, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to check rows affected: %w", err)
    }
    if n == 0 {
        return fmt.Errorf("student with id %d not found", id)
    }
    return nil
}
```

Apply the same pattern to `DeleteStudent`, `DeleteLesson`, `DeleteMaterial`, and `ToggleVocabSent`.

---

### 5. `ToggleVocabSent` doesn't actually toggle — it's a setter
**File:** `lessons.go:75`

The function name says "toggle" but it takes an explicit `bool` — that's a **set**, not a toggle. Either rename it or make it a real toggle.

**Requested change (option A — true toggle, preferred for the TUI 'V' key):**

```go
func (s *Store) ToggleVocabSent(id int64) error {
    query := `UPDATE lessons SET vocab_sent = NOT vocab_sent WHERE id = ?`
    // ... + RowsAffected check
}
```

This aligns perfectly with the Phase 4 spec: *"Pressing 'V' toggles `vocab_sent`."*

---

### 6. `GetAllLessons` will be insufficient — need a date-filtered query
**File:** `lessons.go:53`

Phase 4 requires a split-pane view of **Today** and **Tomorrow** lessons. A blanket `GetAllLessons` won't cut it once the database grows. You'll want a filtered query sooner rather than later.

**Requested change:** Add a date-range query now so Phase 4 doesn't need to refactor:

```go
func (s *Store) GetLessonsByDateRange(start, end time.Time) ([]Lesson, error) {
    query := `SELECT id, student_id, start_time, end_time, vocab_sent, notes, created_at
              FROM lessons WHERE start_time >= ? AND start_time < ? ORDER BY start_time ASC`
    // ...
}
```

---

## 🟢 Minor / Housekeeping

### 7. Schema should include `UNIQUE` constraint on student name
**File:** `db.go:30-36`

Two students with the same name will cause silent data issues. At minimum, make `name` unique or add a secondary identifier.

**Requested change:** `name TEXT NOT NULL UNIQUE` in the students table.

---

### 8. Test script should be a proper `_test.go` file
**File:** `cmd/testdb/main.go`

A `main()` that runs assertions isn't idiomatic Go. This should be a `TestXxx` function using `testing.T` so it integrates with `go test` and gives proper pass/fail output.

**Requested change:** Replace `cmd/testdb/main.go` with `internal/db/db_test.go` using the standard `testing` package. Use `:memory:` as the SQLite data source for speed and no cleanup. Example structure:

```go
package db_test

import (
    "testing"
    "teaching-assistant-app/internal/db"
)

func setupTestStore(t *testing.T) *db.Store {
    t.Helper()
    store, err := db.NewStore(":memory:")
    if err != nil {
        t.Fatalf("failed to init test db: %v", err)
    }
    t.Cleanup(func() { store.Close() })
    return store
}

func TestStudentCRUD(t *testing.T) {
    store := setupTestStore(t)
    // ... assertions using t.Errorf / t.Fatalf
}
```

---

### 9. `CloseDB` / `Close` should return an error
**File:** `db.go:68-72`

`sql.DB.Close()` returns an error. Silently discarding it is a minor code smell.

**Requested change:**

```go
func (s *Store) Close() error {
    if s.db != nil {
        return s.db.Close()
    }
    return nil
}
```

---

## Summary Checklist

| # | Issue | Severity | Files Affected |
|---|-------|----------|----------------|
| 1 | Replace global `DB` with `Store` struct | 🔴 Critical | All `internal/db/*.go` |
| 2 | Add `rows.Err()` checks | 🔴 Critical | `students.go`, `lessons.go`, `materials.go` |
| 3 | Enable `PRAGMA foreign_keys` | 🔴 Critical | `db.go` |
| 4 | Check `RowsAffected` on update/delete | 🟡 Important | `students.go`, `lessons.go`, `materials.go` |
| 5 | Rename or fix `ToggleVocabSent` | 🟡 Important | `lessons.go` |
| 6 | Add `GetLessonsByDateRange` | 🟡 Important | `lessons.go` |
| 7 | `UNIQUE` constraint on student name | 🟢 Minor | `db.go` |
| 8 | Convert test script to `_test.go` | 🟢 Minor | `cmd/testdb/main.go` → `internal/db/db_test.go` |
| 9 | Return error from `Close` | 🟢 Minor | `db.go` |

---

## Review Process Log

### Step 1 — Initial Review (2026-03-04)
**Reviewer:** Senior Dev (AI)
**Files reviewed:** `internal/db/db.go`, `students.go`, `lessons.go`, `materials.go`, `cmd/testdb/main.go`

- Identified 9 issues (3 critical, 3 important, 3 minor) — see sections above.
- Handed feedback report to Junior Dev for implementation.

### Step 2 — Junior Dev Implementation
The Junior Dev implemented **all 9 items** (critical, important, *and* minor):

| # | Item | What was done |
|---|------|---------------|
| 1 | `Store` struct | Replaced global `var DB *sql.DB` with `Store` struct. All CRUD functions converted to methods on `*Store`. `InitDB()` → `NewStore()`. |
| 2 | `rows.Err()` | Added after every `rows.Next()` loop in `GetAllStudents`, `GetAllLessons`, `GetLessonsByDateRange`, `GetMaterialsByLesson`. |
| 3 | FK pragma | `PRAGMA foreign_keys = ON` added in `NewStore()` immediately after `Ping()`. |
| 4 | `RowsAffected` | Applied to `UpdateStudent`, `DeleteStudent`, `ToggleVocabSent`, `DeleteLesson`, `DeleteMaterial`. All return "not found" error when 0 rows affected. |
| 5 | True toggle | `ToggleVocabSent` now uses `NOT vocab_sent` in SQL — no bool parameter. |
| 6 | Date-range query | `GetLessonsByDateRange(start, end time.Time)` added to `lessons.go` with `rows.Err()` check. |
| 7 | `UNIQUE` constraint | Schema updated: `name TEXT NOT NULL UNIQUE`. |
| 8 | Proper tests | `cmd/testdb/main.go` deleted. New `internal/db/db_test.go` created with `TestStudentCRUD`, `TestLessonCRUD`, `TestMaterialCRUD` using `:memory:` DB and `t.Cleanup`. Also tests the duplicate-name constraint. |
| 9 | `Close()` returns error | Signature changed to `func (s *Store) Close() error`. Test cleanup checks the returned error. |

### Step 3 — Re-Review & Verification (2026-03-04)
**Reviewer:** Senior Dev (AI)

Re-read all files in `internal/db/` and confirmed each item was correctly implemented. Ran test suite:

```
$ go test -v ./internal/db/
=== RUN   TestStudentCRUD
--- PASS: TestStudentCRUD (0.00s)
=== RUN   TestLessonCRUD
--- PASS: TestLessonCRUD (0.00s)
=== RUN   TestMaterialCRUD
--- PASS: TestMaterialCRUD (0.00s)
PASS
```

**Verdict:** ✅ **All items resolved. Phase 1 approved — cleared to proceed to Phase 2.**

---

## Current Project Status

| Phase | Status | Notes |
|-------|--------|-------|
| Phase 1: Foundation & Data Layer | ✅ Complete | Reviewed, revised, re-reviewed, all tests passing |
| Phase 2: Google Calendar Fetcher | ⬜ Not started | Next up |
| Phase 3: AI & Clipboard Pipeline | ⬜ Not started | |
| Phase 4: Bubble Tea Interface | ⬜ Not started | |
| Phase 5: WhatsApp Daemon | ⬜ Not started | |
| Phase 6: Final Integration | ⬜ Not started | |
