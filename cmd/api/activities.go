package main

import (
	"errors"
	"net/http"

	"godvanced.forstes.github.com/internal/data"
	"godvanced.forstes.github.com/internal/validator"
)

func (app *application) createActivityHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	claims, err := app.extractClaims(cookie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	userID := int64(claims["user"].(float64))

	var input struct {
		Name         string  `json:"name"`
		AnswerPoints []int16 `json:"answer_points"`
		AnswersSum   int16   `json:"answers_sum"`
		Status       int16   `json:"status"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	activity := &data.Activity{
		Name:         input.Name,
		AnswerPoints: input.AnswerPoints,
		AnswersSum:   input.AnswersSum,
		Status:       input.Status,
	}
	activity.UserID = userID
	if activity.AnswerPoints == nil {
		activity.AnswerPoints = make([]int16, 0)
	}

	v := validator.New()

	if data.ValidateActivity(v, activity); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Activities.InsertActivity(activity)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"activity": activity}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateActivityHandler(w http.ResponseWriter, r *http.Request) {
	// Getting userID and Role
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	claims, err := app.extractClaims(cookie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	userID := int64(claims["user"].(float64))
	roleInt := int8(claims["role"].(float64))

	// Getting Activity
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	activity, err := app.models.Activities.GetActivity(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// User can't edit not his own activities, admin can edit anyway
	if activity.UserID != userID && roleInt != data.AdminRole {
		app.forbiddenResponse(w, r, err)
		return
	}

	var input struct {
		Name         *string `json:"name"`
		AnswerPoints []int16 `json:"answer_points"`
		AnswersSum   *int16  `json:"answers_sum"`
		Status       *int16  `json:"status"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		activity.Name = *input.Name
	}
	if input.AnswerPoints != nil {
		activity.AnswerPoints = input.AnswerPoints
	}
	if input.AnswersSum != nil {
		activity.AnswersSum = *input.AnswersSum
	}
	if input.Status != nil {
		activity.Status = *input.Status
	}

	v := validator.New()
	if data.ValidateActivity(v, activity); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Activities.UpdateActivity(activity)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"activity": activity}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	claims, err := app.extractClaims(cookie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	userID := int64(claims["user"].(float64))

	v := validator.New()
	app.getUserActivities(userID, v, w, r)
}

func (app *application) listUserActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	v := validator.New()
	qs := r.URL.Query()

	userID := app.readInt64(qs, "user_id", 1, v)
	app.getUserActivities(userID, v, w, r)
}

func (app *application) getUserActivities(userId int64, v *validator.Validator, w http.ResponseWriter, r *http.Request) {

	var input struct {
		UserID int64
		data.Filters
	}

	qs := r.URL.Query()

	input.UserID = userId
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "-created_at")
	input.Filters.SortSafelist = []string{"created_at", "-created_at", "email", "-email"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	activities, metadata, err := app.models.Activities.GetActivities(input.UserID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"activities": activities, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
