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

type Question struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	VideoUrl string    `json:"video_url"`
	Answers  []*Answer `json:"answers"`
}

func ValidateQuestion(v *validator.Validator, question *Question) {
	v.Check(question.Title != "", "title", "must be provided")
	v.Check(len(question.Title) <= 500, "title", "must not be more than 500 bytes long")
}

type QuestionModel struct {
	DB *pgxpool.Pool
}

func (m QuestionModel) InsertQuestion(question *Question) error {
	query :=
		`WITH qrow AS (INSERT INTO questions (title) VALUES ($1) RETURNING id)
		INSERT INTO answers (question_id, title, points) VALUES`

	valuesStr := ""
	args := []any{question.Title}
	i := 2
	for _, ans := range question.Answers {
		valuesStr += fmt.Sprintf(" ((SELECT * FROM qrow), $%d, $%d),", i, i+1)
		args = append(args, ans.Title, ans.Points)
		i += 2
	}

	valuesStr = valuesStr[:len(valuesStr)-1]
	query += valuesStr + " RETURNING id"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	i = 0
	for rows.Next() {
		err := rows.Scan(
			&question.Answers[i].ID,
		)
		if err != nil {
			return nil
		}
		i += 1
	}

	if err = rows.Err(); err != nil {
		return nil
	}

	return nil
}

func (m QuestionModel) GetQuestion(id int64) (*Question, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, title, video_url
		FROM questions
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var question Question

	err := m.DB.QueryRow(ctx, query, id).Scan(
		&question.ID,
		&question.Title,
		&question.VideoUrl,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &question, nil
}

func (m QuestionModel) GetAllQuestions(filters Filters) ([]*Question, Metadata, error) {
	query :=
		`SELECT count(*) OVER(), q.id, q.title, q.video_url, a.id, a.title, a.points 
		FROM questions q
		JOIN answers a ON a.question_id = q.id
		ORDER BY q.id ASC, a.id ASC
		LIMIT $1 OFFSET $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	questions := []*Question{}

	for rows.Next() {
		var question Question
		var answer Answer

		err := rows.Scan(
			&totalRecords,
			&question.ID,
			&question.Title,
			&question.VideoUrl,
			&answer.ID,
			&answer.Title,
			&answer.Points,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// if same questions - just add answer to slice, otherwise - add new question and add answer
		if len(questions) == 0 || question.ID != questions[len(questions)-1].ID {
			questions = append(questions, &question)
		}
		questions[len(questions)-1].Answers = append(questions[len(questions)-1].Answers, &answer)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return questions, metadata, nil
}

func (m QuestionModel) UpdateQuestion(question *Question) error {
	query := `
		UPDATE questions
		SET title = $1, video_url = $2
		WHERE id = $3`

	args := []any{
		question.Title,
		question.VideoUrl,
		question.ID,
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

func (m QuestionModel) DeleteQuestion(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM questions WHERE id = $1`

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
