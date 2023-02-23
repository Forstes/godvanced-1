package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"godvanced.forstes.github.com/internal/validator"
)

type Activity struct {
	ID           int64   `json:"id"`
	UserID       int64   `json:"-"`
	Name         string  `json:"name"`
	AnswerPoints []int16 `json:"answer_points"`
	AnswersSum   int16   `json:"answers_sum"`
	Status       int16   `json:"status"`
}

const (
	Ikigai = 0
	Tool   = 1
	Trash  = 2
)

func ValidateActivity(v *validator.Validator, activity *Activity) {
	v.Check(activity.Name != "", "name", "must be provided")
	v.Check(len(activity.Name) <= 64, "name", "must not be more than 64 bytes long")
	v.Check(activity.AnswerPoints != nil, "answer_points", "must be provided")
	v.Check(activity.AnswersSum >= 0, "answer_points", "must be positive value")
	v.Check(activity.Status >= 0 && activity.Status <= 2, "status", "should be equal 0, 1, or 2")
}

type ActivityModel struct {
	DB *pgxpool.Pool
}

func (m ActivityModel) GetActivity(id int64) (*Activity, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query :=
		`SELECT user_id, name, answer_points, answers_sum, status
		FROM activities
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var activity Activity

	err := m.DB.QueryRow(ctx, query, id).Scan(
		&activity.UserID,
		&activity.Name,
		&activity.AnswerPoints,
		&activity.AnswersSum,
		&activity.Status,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &activity, nil
}

func (m ActivityModel) GetActivities(userID int64, filters Filters) ([]*Activity, Metadata, error) {
	query := `
		SELECT count(*) OVER(), id, name, answer_points, answers_sum, status
		FROM activities
		WHERE user_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, userID, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	activities := []*Activity{}

	for rows.Next() {
		var activity Activity

		err := rows.Scan(
			&totalRecords,
			&activity.ID,
			&activity.Name,
			&activity.AnswerPoints,
			&activity.AnswersSum,
			&activity.Status,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		activities = append(activities, &activity)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return activities, metadata, nil
}

func (m ActivityModel) InsertActivity(activity *Activity) error {
	query := `INSERT INTO activities (user_id, name, answer_points, answers_sum, status) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		activity.UserID, activity.Name, activity.AnswerPoints, activity.AnswersSum, activity.Status,
	}

	return m.DB.QueryRow(ctx, query, args...).Scan(&activity.ID)
}

func (m ActivityModel) UpdateActivity(activity *Activity) error {
	query :=
		`UPDATE activities
		SET name = $1, answer_points = $2, answers_sum = $3, status = $4
		WHERE id = $5`

	args := []any{
		activity.Name, activity.AnswerPoints, activity.AnswersSum, activity.Status, activity.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
