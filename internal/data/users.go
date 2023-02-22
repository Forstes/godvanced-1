package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"godvanced.forstes.github.com/internal/validator"
)

type User struct {
	ID        int       `json:"id"`
	Role      int       `json:"-"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"-"`
}

type UserIkigai struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Ikigai string `json:"ikigai"`
}

const (
	UserRole  = 0
	AdminRole = 1
)

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(validator.Matches(user.Email, validator.EmailRX), "email", "incorrect format")
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 64, "name", "must not be more than 64 bytes long")
	v.Check(len(user.Password) >= 8, "password", "must contain at least 8 characters")
	v.Check(len(user.Password) <= 32, "password", "must not exceed max length")
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m UserModel) InsertUser(user *User) error {
	query := `INSERT INTO users (email, name, password) VALUES ($1, $2, $3) RETURNING id, role`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRow(ctx, query, user.Email, user.Name, user.Password).Scan(&user.ID, &user.Role)
}

func (m UserModel) GetUser(email string) (*User, error) {
	query := `
		SELECT id, role, email, name, password, created_at
		FROM users
		WHERE email = $1`

	user := User{Email: email}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Role,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m UserModel) UpdateUser(user *User) error {
	query := `
		UPDATE users
		SET email = $1, name = $2, password = $3
		WHERE id = $4`

	args := []any{
		user.Email,
		user.Name,
		user.Password,
		user.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.Exec(ctx, query, args...)
	if err != nil {
		switch {
		case result.RowsAffected() == 0:
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetUserIkigais(searchEmail string, filters Filters) ([]*UserIkigai, Metadata, error) {
	query := fmt.Sprintf(
		`SELECT count(*) OVER(), u.id, u.email, u.name, a.name 
		FROM users u
		JOIN activities a ON (a.user_id = u.id AND a.answers_sum = (SELECT MAX(answers_sum) FROM activities WHERE user_id = u.id)) 
		WHERE (u.email ILIKE $1 OR $1 = '')
		ORDER BY u.%s %s, u.id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, searchEmail, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	ikigais := []*UserIkigai{}

	for rows.Next() {
		var ikigai UserIkigai

		err := rows.Scan(
			&totalRecords,
			&ikigai.UserID,
			&ikigai.Email,
			&ikigai.Name,
			&ikigai.Ikigai,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		ikigais = append(ikigais, &ikigai)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return ikigais, metadata, nil
}
