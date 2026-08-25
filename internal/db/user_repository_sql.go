package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/realestate-trust/monorepo/internal/core"
)

type SQLUserRepository struct {
	db *sql.DB
}

func NewSQLUserRepository(db *sql.DB) *SQLUserRepository {
	return &SQLUserRepository{db: db}
}

func (r *SQLUserRepository) CreateUser(email, passwordHash, name string, role core.UserRole) (*User, error) {
	id := "usr-" + email
	query := `INSERT INTO users (id, email, password_hash, full_name, role)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, email, password_hash, full_name, role, created_at`

	user := &User{}
	var roleStr string
	err := r.db.QueryRow(query, id, email, passwordHash, name, string(role)).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&roleStr,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.Role = core.UserRole(roleStr)
	return user, nil
}

func (r *SQLUserRepository) GetUser(id string) (*User, error) {
	query := `SELECT id, email, password_hash, full_name, role, created_at FROM users WHERE id = $1`
	user := &User{}
	var roleStr string
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&roleStr,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	user.Role = core.UserRole(roleStr)
	return user, nil
}

func (r *SQLUserRepository) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, email, password_hash, full_name, role, created_at FROM users WHERE email = $1`
	user := &User{}
	var roleStr string
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&roleStr,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	user.Role = core.UserRole(roleStr)
	return user, nil
}

func (r *SQLUserRepository) SubmitKYC(userID, docType, docRef string) (*KYCVerification, error) {
	// First, clean up any existing verification for this user to match mock behavior
	_, err := r.db.Exec(`DELETE FROM kyc_verifications WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}

	encDocRef, err := EncryptKYC(docRef)
	if err != nil {
		return nil, err
	}

	id := "kyc-" + userID
	query := `INSERT INTO kyc_verifications (id, user_id, document_type, document_reference, status)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, user_id, document_type, document_reference, status, verified_at, created_at`

	kyc := &KYCVerification{}
	err = r.db.QueryRow(query, id, userID, docType, encDocRef, "PENDING").Scan(
		&kyc.ID,
		&kyc.UserID,
		&kyc.DocumentType,
		&kyc.DocumentReference,
		&kyc.Status,
		&kyc.VerifiedAt,
		&kyc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	decDocRef, err := DecryptKYC(kyc.DocumentReference)
	if err != nil {
		return nil, err
	}
	kyc.DocumentReference = decDocRef

	return kyc, nil
}

func (r *SQLUserRepository) GetKYCStatus(userID string) (string, *time.Time, error) {
	query := `SELECT status, verified_at FROM kyc_verifications WHERE user_id = $1`
	var status string
	var verifiedAt *time.Time
	err := r.db.QueryRow(query, userID).Scan(&status, &verifiedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "PENDING", nil, nil
		}
		return "PENDING", nil, err
	}
	return status, verifiedAt, nil
}

func (r *SQLUserRepository) ListUsers() ([]*User, error) {
	query := `SELECT id, email, password_hash, full_name, role, created_at FROM users`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []*User
	for rows.Next() {
		user := &User{}
		var roleStr string
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.FullName,
			&roleStr,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		user.Role = core.UserRole(roleStr)
		result = append(result, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return []*User{}, nil
	}
	return result, nil
}

func (r *SQLUserRepository) CreateSession(userID string) (string, error) {
	token := uuid.NewString()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	query := `INSERT INTO refresh_token_sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, token, userID, expiresAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (r *SQLUserRepository) ValidateSession(token string) (string, error) {
	query := `SELECT user_id, expires_at FROM refresh_token_sessions WHERE token = $1`
	var userID string
	var expiresAt time.Time
	err := r.db.QueryRow(query, token).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("session not found")
		}
		return "", err
	}
	if time.Now().After(expiresAt) {
		return "", errors.New("session expired")
	}
	return userID, nil
}

func (r *SQLUserRepository) RevokeSession(token string) error {
	query := `DELETE FROM refresh_token_sessions WHERE token = $1`
	_, err := r.db.Exec(query, token)
	return err
}

func (r *SQLUserRepository) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}
