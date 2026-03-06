package calendar

import (
	"context"
	"fmt"
	"time"

	"teaching-assistant-app/internal/auth"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// EventFetcher allows mocking of the calendar service in tests.
type EventFetcher interface {
	FetchEvents(ctx context.Context, keyword string) ([]Event, error)
}

// Client is a wrapper around the Google Calendar Service
type Client struct {
	srv *calendar.Service
}

// Event represents a parsed calendar event.
type Event struct {
	ID        string
	Title     string
	StartTime time.Time
	EndTime   time.Time
}

// NewClient initializes the Google Calendar client.
// It will prompt the user in the terminal to authorize the app if a valid token is not found.
func NewClient(ctx context.Context, credentialsFile, tokenFile string) (*Client, error) {
	httpClient, err := auth.GetHTTPClient(ctx, credentialsFile, tokenFile,
		calendar.CalendarReadonlyScope,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get http client: %w", err)
	}

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %w", err)
	}

	return &Client{srv: srv}, nil
}

// getWIBLocation returns the Western Indonesian Time (WIB) location (UTC+7).
func getWIBLocation() *time.Location {
	return time.FixedZone("WIB", 7*3600)
}

// FetchEvents fetches today's and tomorrow's events containing the specific keyword.
// Time boundaries and parsed event times are strictly converted to WIB (UTC+7).
func (c *Client) FetchEvents(ctx context.Context, keyword string) ([]Event, error) {
	wibLocation := getWIBLocation()
	nowWIB := time.Now().In(wibLocation)

	// Today's start in WIB (00:00:00)
	todayStart := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), 0, 0, 0, 0, wibLocation)
	// Tomorrow's end in WIB (Start of Day + 2 days = 00:00:00 day after tomorrow)
	tomorrowEnd := todayStart.AddDate(0, 0, 2)

	// timeMin and timeMax require RFC3339 format
	timeMin := todayStart.Format(time.RFC3339)
	timeMax := tomorrowEnd.Format(time.RFC3339)

	eventsList, err := c.srv.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin).
		TimeMax(timeMax).
		MaxResults(100).
		OrderBy("startTime").
		Q(keyword).
		Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %w", err)
	}

	var events []Event
	for _, item := range eventsList.Items {
		// Event could be an all-day event (Date) or specific time (DateTime)
		startStr := item.Start.DateTime
		if startStr == "" {
			startStr = item.Start.Date
		}
		endStr := item.End.DateTime
		if endStr == "" {
			endStr = item.End.Date
		}

		startTime, err := parseEventTime(startStr, wibLocation)
		if err != nil {
			return nil, fmt.Errorf("unable to parse start time for event %q: %w", item.Summary, err)
		}

		endTime, err := parseEventTime(endStr, wibLocation)
		if err != nil {
			return nil, fmt.Errorf("unable to parse end time for event %q: %w", item.Summary, err)
		}

		events = append(events, Event{
			ID:        item.Id,
			Title:     item.Summary,
			StartTime: startTime,
			EndTime:   endTime,
		})
	}

	return events, nil
}

// parseEventTime attempts to parse an RFC3339 datetime string.
// If it fails, it tries parsing as a date-only string ("2006-01-02").
// The resulting time is always returned in the specified location.
func parseEventTime(timeStr string, loc *time.Location) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t.In(loc), nil
	}

	t, err = time.ParseInLocation("2006-01-02", timeStr, loc)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time string %q", timeStr)
}
