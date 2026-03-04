package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Lesson struct {
	ID        int64
	StudentID int64
	StartTime time.Time
	EndTime   time.Time
	VocabSent bool
	Notes     string
	CreatedAt time.Time
}

// CreateLesson adds a new lesson to the database.
func (s *Store) CreateLesson(studentID int64, startTime, endTime time.Time, notes string) (int64, error) {
	query := `INSERT INTO lessons (student_id, start_time, end_time, notes) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, studentID, startTime, endTime, notes)
	if err != nil {
		return 0, fmt.Errorf("failed to insert lesson: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	return id, nil
}

// GetLesson retrieves a lesson by ID.
func (s *Store) GetLesson(id int64) (*Lesson, error) {
	query := `SELECT id, student_id, start_time, end_time, vocab_sent, notes, created_at FROM lessons WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var l Lesson
	err := row.Scan(&l.ID, &l.StudentID, &l.StartTime, &l.EndTime, &l.VocabSent, &l.Notes, &l.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan lesson: %w", err)
	}

	return &l, nil
}

// GetAllLessons retrieves all lessons.
func (s *Store) GetAllLessons() ([]Lesson, error) {
	query := `SELECT id, student_id, start_time, end_time, vocab_sent, notes, created_at FROM lessons ORDER BY start_time ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query lessons: %w", err)
	}
	defer rows.Close()

	var lessons []Lesson
	for rows.Next() {
		var l Lesson
		err := rows.Scan(&l.ID, &l.StudentID, &l.StartTime, &l.EndTime, &l.VocabSent, &l.Notes, &l.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lesson: %w", err)
		}
		lessons = append(lessons, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return lessons, nil
}

// GetLessonsByDateRange retrieves lessons within a specific date range.
func (s *Store) GetLessonsByDateRange(start, end time.Time) ([]Lesson, error) {
	query := `SELECT id, student_id, start_time, end_time, vocab_sent, notes, created_at
	          FROM lessons WHERE start_time >= ? AND start_time < ? ORDER BY start_time ASC`
	rows, err := s.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query lessons by date range: %w", err)
	}
	defer rows.Close()

	var lessons []Lesson
	for rows.Next() {
		var l Lesson
		err := rows.Scan(&l.ID, &l.StudentID, &l.StartTime, &l.EndTime, &l.VocabSent, &l.Notes, &l.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lesson: %w", err)
		}
		lessons = append(lessons, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return lessons, nil
}

// ToggleVocabSent toggles the vocab_sent status of a lesson.
func (s *Store) ToggleVocabSent(id int64) error {
	query := `UPDATE lessons SET vocab_sent = NOT vocab_sent WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to toggle lesson vocab status: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("lesson with id %d not found", id)
	}

	return nil
}

// DeleteLesson removes a lesson by ID.
func (s *Store) DeleteLesson(id int64) error {
	query := `DELETE FROM lessons WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lesson: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("lesson with id %d not found", id)
	}

	return nil
}