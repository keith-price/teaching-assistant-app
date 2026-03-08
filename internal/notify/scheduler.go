package notify

import (
	"context"
	"fmt"
	"strings"
	"time"

	"teaching-assistant-app/internal/calendar"
	"teaching-assistant-app/internal/db"

	"github.com/robfig/cron/v3"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
	waProto "go.mau.fi/whatsmeow/binary/proto"
)

// Scheduler handles the automated messaging daemon.
type Scheduler struct {
	waClient *WhatsAppClient
	dbStore  *db.Store
	calFetch calendar.EventFetcher
	cron     *cron.Cron
}

// NewScheduler creates a new notification scheduler.
func NewScheduler(waClient *WhatsAppClient, dbStore *db.Store, calFetch calendar.EventFetcher) *Scheduler {
	// Set cron to run with WIB timezone
	loc := time.FixedZone("WIB", 7*3600)
	c := cron.New(cron.WithLocation(loc))

	return &Scheduler{
		waClient: waClient,
		dbStore:  dbStore,
		calFetch: calFetch,
		cron:     c,
	}
}

// Start begins the background cron scheduler.
func (s *Scheduler) Start() error {
	// Schedule to run every morning at 7:00 AM WIB
	_, err := s.cron.AddFunc("0 7 * * *", func() {
		err := s.SendDailyBriefing(context.Background())
		if err != nil {
			fmt.Printf("Failed to send daily briefing: %v\n", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to schedule cron job: %w", err)
	}

	s.cron.Start()
	return nil
}

// Stop halts the scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// SendDailyBriefing constructs the daily schedule text and sends it via WhatsApp.
func (s *Scheduler) SendDailyBriefing(ctx context.Context) error {
	msg, err := s.BuildDailyScheduleMessage(ctx)
	if err != nil {
		return fmt.Errorf("failed to build message: %w", err)
	}

	// Assuming we send the message to the authenticated user's own number.
	ownID := s.waClient.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("whatsapp client not logged in")
	}

	// Convert the device JID to a user JID (removing the device specific parts)
	targetJID := types.NewJID(ownID.User, ownID.Server)

	waMsg := &waProto.Message{
		Conversation: proto.String(msg),
	}

	_, err = s.waClient.Client.SendMessage(ctx, targetJID, waMsg)
	if err != nil {
		return fmt.Errorf("failed to send whatsapp message: %w", err)
	}

	fmt.Println("Daily briefing sent successfully.")
	return nil
}

// BuildDailyScheduleMessage queries the Calendar and Database to format the daily schedule.
func (s *Scheduler) BuildDailyScheduleMessage(ctx context.Context) (string, error) {
	loc := time.FixedZone("WIB", 7*3600)
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	tomorrowStart := todayStart.AddDate(0, 0, 1)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Good morning! Here is today's schedule (%s):\n\n", todayStart.Format("Mon, Jan 2")))

	// 1. Fetch from Google Calendar (e.g., using "Preply" keyword as an example)
	events, err := s.calFetch.FetchEvents(ctx, "Preply")
	if err != nil {
		return "", fmt.Errorf("failed to fetch calendar events: %w", err)
	}

	sb.WriteString("🗓️ *Calendar Events:*\n")
	eventCount := 0
	for _, event := range events {
		// Filter events that start today
		if event.StartTime.After(todayStart) && event.StartTime.Before(tomorrowStart) {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", event.StartTime.Format("15:04"), event.Title))
			eventCount++
		}
	}
	if eventCount == 0 {
		sb.WriteString("No calendar events today.\n")
	}

	sb.WriteString("\n")

	// 2. Fetch from Local Database
	lessonsWithStudent, err := s.dbStore.GetLessonsWithStudentByDateRange(todayStart, tomorrowStart)
	if err != nil {
		return "", fmt.Errorf("failed to fetch lessons from db: %w", err)
	}

	sb.WriteString("📚 *Local Database Lessons:*\n")
	if len(lessonsWithStudent) == 0 {
		sb.WriteString("No registered lessons in the local database for today.\n")
	} else {
		for _, ls := range lessonsWithStudent {
			studentInfo := ls.Student.Name
			if studentInfo == "" {
				studentInfo = fmt.Sprintf("Student ID: %d", ls.Lesson.StudentID)
			}

			vocabStatus := "❌ Vocab Pending"
			if ls.Lesson.VocabSent {
				vocabStatus = "✅ Vocab Sent"
			}
			
			sb.WriteString(fmt.Sprintf("- %s: %s (%s)\n", ls.Lesson.StartTime.Format("15:04"), studentInfo, vocabStatus))
		}
	}

	sb.WriteString("\nHave a great day teaching!")
	return sb.String(), nil
}
