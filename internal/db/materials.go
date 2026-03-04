package db

import (
	"fmt"
	"time"
)

type Material struct {
	ID        int64
	LessonID  int64
	FilePath  string
	CreatedAt time.Time
}

// CreateMaterial adds a new material linked to a lesson.
func (s *Store) CreateMaterial(lessonID int64, filePath string) (int64, error) {
	query := `INSERT INTO materials (lesson_id, file_path) VALUES (?, ?)`
	result, err := s.db.Exec(query, lessonID, filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to insert material: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	return id, nil
}

// GetMaterialsByLesson retrieves materials associated with a lesson ID.
func (s *Store) GetMaterialsByLesson(lessonID int64) ([]Material, error) {
	query := `SELECT id, lesson_id, file_path, created_at FROM materials WHERE lesson_id = ?`
	rows, err := s.db.Query(query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to query materials: %w", err)
	}
	defer rows.Close()

	var materials []Material
	for rows.Next() {
		var m Material
		err := rows.Scan(&m.ID, &m.LessonID, &m.FilePath, &m.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan material: %w", err)
		}
		materials = append(materials, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return materials, nil
}

// DeleteMaterial removes a material by ID.
func (s *Store) DeleteMaterial(id int64) error {
	query := `DELETE FROM materials WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete material: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("material with id %d not found", id)
	}

	return nil
}