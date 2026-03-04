package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Student struct {
	ID          int64
	Name        string
	Level       string
	ContactInfo string
	CreatedAt   time.Time
}

// CreateStudent adds a new student to the database.
func (s *Store) CreateStudent(name, level, contactInfo string) (int64, error) {
	query := `INSERT INTO students (name, level, contact_info) VALUES (?, ?, ?)`
	result, err := s.db.Exec(query, name, level, contactInfo)
	if err != nil {
		return 0, fmt.Errorf("failed to insert student: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	return id, nil
}

// GetStudent retrieves a student by ID.
func (s *Store) GetStudent(id int64) (*Student, error) {
	query := `SELECT id, name, level, contact_info, created_at FROM students WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var st Student
	err := row.Scan(&st.ID, &st.Name, &st.Level, &st.ContactInfo, &st.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan student: %w", err)
	}

	return &st, nil
}

// GetAllStudents retrieves all students.
func (s *Store) GetAllStudents() ([]Student, error) {
	query := `SELECT id, name, level, contact_info, created_at FROM students ORDER BY name ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query students: %w", err)
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var st Student
		err := rows.Scan(&st.ID, &st.Name, &st.Level, &st.ContactInfo, &st.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan student: %w", err)
		}
		students = append(students, st)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return students, nil
}

// UpdateStudent updates an existing student's details.
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

// DeleteStudent removes a student by ID.
func (s *Store) DeleteStudent(id int64) error {
	query := `DELETE FROM students WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
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