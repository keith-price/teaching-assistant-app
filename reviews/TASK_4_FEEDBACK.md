# Task 4 Feedback — Drive Folder Picker Enhancements

**Reviewer:** Senior Dev
**Date:** 2026-03-06
**Scope:** Drive scope upgrade, folder creation, breadcrumb navigation, subfolder upload

---

## Files Reviewed

| File                           | Verdict |
| ------------------------------ | ------- |
| `internal/drive/drive.go`      | ✅ Pass |
| `internal/drive/drive_test.go` | ✅ Pass |
| `internal/tui/model.go`        | ✅ Pass |
| `internal/tui/update.go`       | ✅ Pass |
| `internal/tui/view.go`         | ✅ Pass |
| `internal/tui/commands.go`     | ✅ Pass |
| `internal/tui/tui_test.go`     | ✅ Pass |
| `cmd/app/main.go`              | ✅ Pass |
| `cmd/auth/main.go`             | ✅ Pass |

## Review Notes

### `internal/drive/drive.go` — CreateFolder

- Correct use of `Fields("id, name")` to limit API response.
- `MimeType` and `Parents` correctly set. No issues.

### `internal/drive/drive_test.go`

- `DriveUploader` interface correctly extended with `CreateFolder`.
- `MockDriveClient.CreateFolder` appends to the slice and returns a predictable ID — good for testing.
- `TestMockCreateFolder` verifies both the returned folder and the side-effect on the mock's internal state.

### `internal/tui/model.go`

- `folderBreadcrumb` type cleanly defined with `id` and `name`.
- New fields (`folderPath`, `showCreateFolder`, `createFolderInput`) added to `Model` alongside existing folder picker state — logical grouping.
- `createFolderInput` correctly initialised in `initForm()` with placeholder and char limit.

### `internal/tui/commands.go`

- `createFolderCmd`: nil-check on `driveClient`, returns `folderCreatedMsg` with `parentID` for refresh — correct.
- `createSubfolderAndUploadCmd`: title sanitisation handles spaces, special chars, and empty-string fallback. Date-stamped subfolder name prevents collisions. Uploads both files into the subfolder. Teacher key upload correctly guarded by `teacherKeyContent != ""`.
- Old `uploadToDriveCmd` is fully removed — no dead code.

### `internal/tui/update.go` — handleFolderPickerKeys

- `showCreateFolder` check is the **first** branch — correct; prevents folder picker keys from leaking into the text input.
- Enter validates non-empty name before calling `createFolderCmd`.
- Esc cleanly cancels folder creation.
- Otherwise, keypress is forwarded to `createFolderInput.Update(msg)` — correct.
- Breadcrumb push on Enter: pushes `{id: m.folderParentID, name: selectedFolder.Name}` — correct.
- Breadcrumb pop on Backspace: pops last entry, restores `folderParentID` — correct. Empty stack falls back to "root".
- `n` key: sets `showCreateFolder`, resets input, focuses it — correct.
- `space`/`s`: bounds-checked, calls `createSubfolderAndUploadCmd` with correct args — correct.

### `internal/tui/update.go` — Update message handlers

- `folderCreatedMsg` handler: sets status, refreshes folder list for `msg.parentID` — correct.

### `internal/tui/view.go` — folderPickerView

- Breadcrumb bar built from `folderPath` with `> ` separators — clean.
- Create-folder input renders with Enter/Esc help text.
- Empty folder list shows "No subfolders. Press 'n' to create one." — good UX.
- Help text matches the full keybinding set.

### `cmd/app/main.go` & `cmd/auth/main.go`

- Both use `driveAPI.DriveScope` (not `DriveFileScope`) — correct.

### `internal/tui/tui_test.go`

- `TestFolderPickerCreateFolder`: verifies pressing `n` sets `showCreateFolder = true` — correct.
- `TestFolderPickerNavigation`: cursor bounds-checking still works with new fields — verified.
- All existing tests remain passing.

## Requested Changes

**None.** All implementations match the spec in `TO-DO-DRIVE-FOLDER-PICKER.md` and follow project conventions.

## Verification Log

```
$ go build ./...
✅ Compiles cleanly

$ go test ./... -count=1
?       teaching-assistant-app/cmd/app   [no test files]
?       teaching-assistant-app/cmd/auth  [no test files]
ok      teaching-assistant-app/internal/ai
ok      teaching-assistant-app/internal/db
ok      teaching-assistant-app/internal/drive
ok      teaching-assistant-app/internal/tui
✅ All tests pass
```

## Verdict

**✅ APPROVED — No changes required.**

> Keith: Before using the new folder browsing features, delete `config/token.json` and re-run `go run cmd/auth/main.go` to authorize with the broader `drive` scope.
