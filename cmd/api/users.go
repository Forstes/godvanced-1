package main

import (
	"errors"
	"net/http"
	"time"

	"godvanced.forstes.github.com/internal/data"
	"godvanced.forstes.github.com/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) listUserIkigaisHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SearchEmail string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.SearchEmail = app.readString(qs, "search", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "-created_at")
	input.Filters.SortSafelist = []string{"created_at", "-created_at", "email", "-email"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ikigais, metadata, err := app.models.Users.GetUserIkigais(input.SearchEmail, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"ikigais": ikigais, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Email:    input.Email,
		Name:     input.Name,
		Password: input.Password,
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	user.Password = string(hashedPass)

	err = app.models.Users.InsertUser(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	activationToken, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"activationToken": activationToken.Plaintext,
			"userID":          user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	token, err := app.generateJWT(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	cookie := &http.Cookie{Name: "auth_token", Value: token, Expires: time.Now().Add(app.config.jwtOptions.expires), HttpOnly: true}
	http.SetCookie(w, cookie)

	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "successfully registered"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.models.Users.GetUser(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		app.badRequestResponse(w, r, data.ErrInvalidCredentials)
	}

	token, err := app.generateJWT(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	cookie := &http.Cookie{Name: "auth_token", Value: token, Expires: time.Now().Add(app.config.jwtOptions.expires), HttpOnly: true}
	http.SetCookie(w, cookie)

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "successfully authorized"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		MaxAge: -1,
	}

	http.SetCookie(w, cookie)
	err := app.writeJSON(w, http.StatusOK, envelope{"message": "logged out"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = app.models.Users.UpdateUser(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
