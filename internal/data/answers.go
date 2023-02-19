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

type Answer struct {
	ID         int    `json:"id"`
	QuestionId int    `json:"-"`
	Title      string `json:"title"`
	Points     int16  `json:"points"`
}

func ValidateAnswer(v *validator.Validator, answer *Answer) {
	v.Check(answer.Title != "", "title", "must be provided")
	v.Check(len(answer.Title) <= 500, "title", "must not be more than 500 bytes long")
}

type AnswerModel struct {
	DB *pgxpool.Pool
}

func (m AnswerModel) InsertAnswers(answers []*Answer) ([]*Answer, error) {
	query := `INSERT INTO answers (question_id, title, points) VALUES`

	valuesStr := ""
	args := []any{}
	count := 1
	for _, ans := range answers {
		valuesStr += fmt.Sprintf(" ($%d, $%d, $%d),", count, count+1, count+2)
		args = append(args, ans.QuestionId, ans.Title, ans.Points)
		count += 3
	}

	valuesStr = valuesStr[:len(valuesStr)-1]
	query += valuesStr + " RETURNING id"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	answersRes := []*Answer{}

	for rows.Next() {
		var answer Answer
		err := rows.Scan(
			&answer.ID,
		)
		if err != nil {
			return nil, err
		}
		answersRes = append(answersRes, &answer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return answersRes, nil
}

func (m AnswerModel) InsertAnswer(answer *Answer) error {
	query := `INSERT INTO answers (title, points) VALUES ($1, $2) RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRow(ctx, query, answer.Title, answer.Points).Scan(&answer.ID)
}

func (m AnswerModel) GetAnswer(id int64) (*Answer, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, title, points
		FROM answers
		WHERE id = $1`

	var answer Answer

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, id).Scan(
		&answer.ID,
		&answer.Title,
		&answer.Points,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &answer, nil
}

func (m AnswerModel) GetAllAnswers() ([]*Answer, error) {
	query :=
		`SELECT id, title, points
		FROM answers
		ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	answers := []*Answer{}

	for rows.Next() {
		var answer Answer
		err := rows.Scan(
			&answer.ID,
			&answer.Title,
			&answer.Points,
		)
		if err != nil {
			return nil, err
		}
		answers = append(answers, &answer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return answers, nil
}

func (m AnswerModel) UpdateAnswer(answer *Answer) error {
	query := `
		UPDATE answers
		SET title = $1, points = $2
		WHERE id = $3`

	args := []any{
		answer.Title,
		answer.Points,
		answer.ID,
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

func (m AnswerModel) DeleteAnswer(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM answers WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
