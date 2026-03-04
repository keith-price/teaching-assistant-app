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
func CreateStudent(name, level, contactInfo string) (int64, error) {
	query := `INSERT INTO students (name, level, contact_info) VALUES (?, ?, ?)`
	result, err := DB.Exec(query, name, level, contactInfo)
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
func GetStudent(id int64) (*Student, error) {
	query := `SELECT id, name, level, contact_info, created_at FROM students WHERE id = ?`
	row := DB.QueryRow(query, id)

	var s Student
	err := row.Scan(&s.ID, &s.Name, &s.Level, &s.ContactInfo, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan student: %w", err)
	}

	return &s, nil
}

// GetAllStudents retrieves all students.
func GetAllStudents() ([]Student, error) {
	query := `SELECT id, name, level, contact_info, created_at FROM students ORDER BY name ASC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query students: %w", err)
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student
		err := rows.Scan(&s.ID, &s.Name, &s.Level, &s.ContactInfo, &s.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan student: %w", err)
		}
		students = append(students, s)
	}

	return students, nil
}

// UpdateStudent updates an existing student's details.
func UpdateStudent(id int64, name, level, contactInfo string) error {
	query := `UPDATE students SET name = ?, level = ?, contact_info = ? WHERE id = ?`
	_, err := DB.Exec(query, name, level, contactInfo, id)
	if err != nil {
		return fmt.Errorf("failed to update student: %w", err)
	}

	return nil
}

// DeleteStudent removes a student by ID.
func DeleteStudent(id int64) error {
	query := `DELETE FROM students WHERE id = ?`
	_, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}

	return nil
}
