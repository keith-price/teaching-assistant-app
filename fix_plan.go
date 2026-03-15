package main

import (
	"log"
	"os"
	"strings"
)

func main() {
	b, err := os.ReadFile("C:\\Users\\keith\\coding\\teaching-assistant-app\\PLAN.md")
	if err != nil {
		log.Fatalf("Failed to read: %v", err)
	}

	content := string(b)
	targetLine := "- [ ] **Task 7.9:** **🛑 CHECKPOINT & REVIEW:** Hand off to the Senior Dev for review."
	idx := strings.Index(content, targetLine)
	if idx == -1 {
		log.Fatalf("Could not find target line to truncate from.")
	}

	// Move index to the end of that line
	newlineIdx := strings.Index(content[idx:], "\n")
	if newlineIdx != -1 {
		idx += newlineIdx + 1
	} else {
		idx += len(content[idx:])
	}

	cleanHandoff := `
### Handoff: Pivot to Standalone Worksheet Generator

**To: Senior Developer**
**From: Junior Developer**

I have completed the refactoring steps outlined in ` + "`TO-DO.md`" + ` to pivot the application into a standalone Worksheet Generator.

**Changes Made:**
1. **The Purge:** Deleted ` + "`internal/calendar/`" + `, ` + "`internal/notify/`" + `, ` + "`internal/db/`" + `, and ` + "`cmd/wa_test/`" + `. Removed all ` + "`*.db`" + ` files.
2. **Auth & Main Refactor:** Removed calendar scopes from auth. Stripped ` + "`cmd/app/main.go`" + ` of all DB/WhatsApp/Calendar/Cron initializations. The app now only initializes Google Drive auth and the AI Generator.
3. **TUI Redesign:** Ripped out the split-pane schedule view and database list logic from ` + "`model.go`" + `, ` + "`view.go`" + `, ` + "`update.go`" + `, and ` + "`commands.go`" + `. The app now boots directly into the 5-field Worksheet input form (Title, Level, Type, Duration, Text) and transitions sequentially to the preview screen and Drive folder picker.
4. **Tests & Dependencies:** Cleaned up ` + "`go.mod`" + ` via ` + "`go mod tidy`" + ` to remove unused dependencies (sqlite, whatsmeow, cron, calendar api). Removed outdated database-dependent tests in ` + "`tui_test.go`" + ` and verified all remaining tests in ` + "`internal/ai`" + `, ` + "`internal/drive`" + `, and ` + "`internal/tui`" + ` pass.

**Verification:**
- ` + "`go mod tidy`" + ` run successfully.
- ` + "`go build ./...`" + ` compiles without errors.
- ` + "`go test ./...`" + ` passes for all remaining packages.

**Constraint Met:** The application now operates strictly as a lightning-fast, single-purpose TUI for generating AI ESL Worksheets and saving them to Google Drive.

Please review the commit and sign off on this pivot phase.
`

	newContent := content[:idx] + cleanHandoff

	err = os.WriteFile("C:\\Users\\keith\\coding\\teaching-assistant-app\\PLAN.md", []byte(newContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write: %v", err)
	}
	log.Println("Successfully fixed PLAN.md")
}
