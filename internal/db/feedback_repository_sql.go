package db

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type SQLFeedbackRepository struct {
	db *sql.DB
}

func NewSQLFeedbackRepository(db *sql.DB) *SQLFeedbackRepository {
	return &SQLFeedbackRepository{db: db}
}

func (r *SQLFeedbackRepository) CreateFeedback(userID, message, category string, rating int) (*Feedback, error) {
	if message == "" {
		return nil, errors.New("message cannot be empty")
	}

	id := "fb-" + uuid.New().String()[:8]
	query := `INSERT INTO feedback (id, user_id, message, category, rating)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, user_id, message, category, rating, created_at`

	f := &Feedback{}
	err := r.db.QueryRow(query, id, userID, message, category, rating).Scan(
		&f.ID,
		&f.UserID,
		&f.Message,
		&f.Category,
		&f.Rating,
		&f.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *SQLFeedbackRepository) ListFeedback() ([]*Feedback, error) {
	query := `SELECT id, user_id, message, category, rating, created_at FROM feedback`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Feedback
	for rows.Next() {
		f := &Feedback{}
		err := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.Message,
			&f.Category,
			&f.Rating,
			&f.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return []*Feedback{}, nil
	}
	return result, nil
}
