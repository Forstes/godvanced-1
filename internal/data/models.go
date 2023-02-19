package data

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRecordNotFound     = errors.New("record not found")
	ErrEditConflict       = errors.New("edit conflict")
	ErrInvalidCredentials = errors.New("wrong user credentials")
)

type Models struct {
	Users      UserModel
	Activities ActivityModel
	Questions  QuestionModel
	Answers    AnswerModel
}

func NewModels(db *pgxpool.Pool) Models {
	return Models{
		Users:      UserModel{DB: db},
		Activities: ActivityModel{DB: db},
		Questions:  QuestionModel{DB: db},
		Answers:    AnswerModel{DB: db},
	}
}
