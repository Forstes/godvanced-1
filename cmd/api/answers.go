package main

import (
	"errors"
	"fmt"
	"net/http"

	"godvanced.forstes.github.com/internal/data"
	"godvanced.forstes.github.com/internal/validator"
)

func (app *application) createAnswerHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string `json:"title"`
		Points int16  `json:"points"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	answer := &data.Answer{
		Title:  input.Title,
		Points: input.Points,
	}

	v := validator.New()

	if data.ValidateAnswer(v, answer); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Answers.InsertAnswer(answer)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/answers/%d", answer.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"answer": answer}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listAnswersHandler(w http.ResponseWriter, r *http.Request) {
	answers, err := app.models.Answers.GetAllAnswers()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"answers": answers}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateAnswerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	answer, err := app.models.Answers.GetAnswer(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title  *string `json:"title"`
		Points *int16  `json:"points"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	answer.Title = *input.Title
	answer.Points = *input.Points

	v := validator.New()
	if data.ValidateAnswer(v, answer); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Answers.UpdateAnswer(answer)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"answer": answer}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteAnswerHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Answers.DeleteAnswer(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "answer successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
