package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"teaching-assistant-app/internal/ai"
	"teaching-assistant-app/internal/db"
	"teaching-assistant-app/internal/drive"

	tea "github.com/charmbracelet/bubbletea"
)

// Messages
type lessonsFetchedMsg struct {
	today    []LessonView
	tomorrow []LessonView
}
type vocabToggledMsg struct {
	lessonID int64
}

type worksheetPreviewMsg struct {
	worksheet  string
	teacherKey string
	level      string
	lessonType string
	title      string
}

type foldersListedMsg struct {
	folders []drive.Folder
}

type folderCreatedMsg struct {
	folder   drive.Folder
	parentID string
}

type uploadCompleteMsg struct {
	folderName string
}

type errMsg struct {
	err error
}

func fetchLessonsCmd(store *db.Store) tea.Cmd {
	return func() tea.Msg {
		if store == nil {
			return errMsg{fmt.Errorf("store is nil")}
		}
		loc := time.FixedZone("WIB", 7*3600)
		now := time.Now().In(loc)
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		tomorrowStart := todayStart.AddDate(0, 0, 1)
		dayAfterTomorrowStart := todayStart.AddDate(0, 0, 2)

		todayLessons, err := store.GetLessonsByDateRange(todayStart, tomorrowStart)
		if err != nil {
			return errMsg{fmt.Errorf("failed to fetch today's lessons: %w", err)}
		}

		tomorrowLessons, err := store.GetLessonsByDateRange(tomorrowStart, dayAfterTomorrowStart)
		if err != nil {
			return errMsg{fmt.Errorf("failed to fetch tomorrow's lessons: %w", err)}
		}

		// fetch student details for each
		var todayViews []LessonView
		for _, l := range todayLessons {
			student, err := store.GetStudent(l.StudentID)
			if err != nil || student == nil {
				continue
			}
			todayViews = append(todayViews, LessonView{Lesson: l, Student: *student})
		}

		var tomorrowViews []LessonView
		for _, l := range tomorrowLessons {
			student, err := store.GetStudent(l.StudentID)
			if err != nil || student == nil {
				continue
			}
			tomorrowViews = append(tomorrowViews, LessonView{Lesson: l, Student: *student})
		}

		return lessonsFetchedMsg{today: todayViews, tomorrow: tomorrowViews}
	}
}

func toggleVocabCmd(store *db.Store, lessonID int64) tea.Cmd {
	return func() tea.Msg {
		if store == nil {
			return errMsg{fmt.Errorf("store is nil")}
		}
		err := store.ToggleVocabSent(lessonID)
		if err != nil {
			return errMsg{fmt.Errorf("failed to toggle vocab: %w", err)}
		}
		return vocabToggledMsg{lessonID}
	}
}

func generateWorksheetCmd(generator *ai.Generator, level, duration, lessonType, sourceText, lessonTitle string) tea.Cmd {
	return func() tea.Msg {
		if generator == nil {
			return errMsg{fmt.Errorf("generator is nil")}
		}
		ctx := context.Background()
		worksheet, teacherKey, err := generator.GenerateWorksheet(ctx, level, duration, lessonType, sourceText, lessonTitle)
		if err != nil {
			return errMsg{fmt.Errorf("AI generation failed: %w", err)}
		}
		return worksheetPreviewMsg{worksheet: worksheet, teacherKey: teacherKey, level: level, lessonType: lessonType, title: lessonTitle}
	}
}

func saveLocallyCmd(worksheet, teacherKey, level, lessonType, title, baseDir string) tea.Cmd {
	return func() tea.Msg {
		_, _, err := ai.SaveDocuments(worksheet, teacherKey, level, lessonType, title, baseDir)
		if err != nil {
			return errMsg{fmt.Errorf("failed to save locally: %w", err)}
		}
		return nil
	}
}

func listDriveFoldersCmd(driveClient *drive.Client, parentID string) tea.Cmd {
	return func() tea.Msg {
		if driveClient == nil {
			return errMsg{fmt.Errorf("driveClient is nil")}
		}
		ctx := context.Background()
		folders, err := driveClient.ListFolders(ctx, parentID)
		if err != nil {
			return errMsg{fmt.Errorf("failed to list Drive folders: %w", err)}
		}
		return foldersListedMsg{folders: folders}
	}
}

// createFolderCmd creates a new folder and refreshes the listing.
func createFolderCmd(driveClient *drive.Client, parentID, name string) tea.Cmd {
	return func() tea.Msg {
		if driveClient == nil {
			return errMsg{fmt.Errorf("driveClient is nil")}
		}
		ctx := context.Background()
		folder, err := driveClient.CreateFolder(ctx, parentID, name)
		if err != nil {
			return errMsg{fmt.Errorf("failed to create folder: %w", err)}
		}
		return folderCreatedMsg{folder: *folder, parentID: parentID}
	}
}

func createSubfolderAndUploadCmd(driveClient *drive.Client, parentID, worksheetContent, teacherKeyContent, level, lessonType, lessonTitle, baseDir string) tea.Cmd {
	return func() tea.Msg {
		if driveClient == nil {
			return errMsg{fmt.Errorf("driveClient is nil")}
		}
		ctx := context.Background()

		// Sanitise title for folder name
		safeTitle := strings.ReplaceAll(lessonTitle, " ", "_")
		safeTitle = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
				return r
			}
			return -1
		}, safeTitle)
		if safeTitle == "" {
			safeTitle = "lesson"
		}
		dateStr := time.Now().Format("2006-01-02_1504")
		subfolderName := fmt.Sprintf("%s_%s_%s_%s", level, lessonType, safeTitle, dateStr)

		// Create subfolder
		subfolder, err := driveClient.CreateFolder(ctx, parentID, subfolderName)
		if err != nil {
			return errMsg{fmt.Errorf("failed to create subfolder: %w", err)}
		}

		// Upload both files into the subfolder
		wsFilename := fmt.Sprintf("%s_%s_%s_%s_worksheet.md", level, lessonType, safeTitle, dateStr)
		tkFilename := fmt.Sprintf("%s_%s_%s_%s_teacher_key.md", level, lessonType, safeTitle, dateStr)

		err = driveClient.UploadFile(ctx, subfolder.ID, wsFilename, worksheetContent)
		if err != nil {
			return errMsg{fmt.Errorf("failed to upload worksheet: %w", err)}
		}

		if teacherKeyContent != "" {
			err = driveClient.UploadFile(ctx, subfolder.ID, tkFilename, teacherKeyContent)
			if err != nil {
				return errMsg{fmt.Errorf("failed to upload teacher key: %w", err)}
			}
		}

		// Cleanup local worksheets directory
		if baseDir != "" {
			worksheetsDir := filepath.Join(baseDir, "worksheets")
			entries, err := os.ReadDir(worksheetsDir)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						os.Remove(filepath.Join(worksheetsDir, entry.Name()))
					}
				}
			}
		}

		return uploadCompleteMsg{folderName: subfolderName}
	}
}
