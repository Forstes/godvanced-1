package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/register", app.register)
	router.HandlerFunc(http.MethodPost, "/v1/login", app.login)
	router.HandlerFunc(http.MethodGet, "/v1/logout", app.logout)
	router.HandlerFunc(http.MethodPut, "/v1/user/activated", app.activateUserHandler)

	router.Handler(http.MethodGet, "/v1/questions", app.verifyJWTMiddleware(http.HandlerFunc(app.listQuestionsHandler)))
	router.Handler(http.MethodGet, "/v1/ikigais", app.JWTAdminOnlyMiddleware(http.HandlerFunc(app.listUserIkigaisHandler)))
	router.Handler(http.MethodPost, "/v1/questions", app.JWTAdminOnlyMiddleware(http.HandlerFunc(app.createQuestionHandler)))
	router.Handler(http.MethodPatch, "/v1/questions/:id", app.JWTAdminOnlyMiddleware(http.HandlerFunc(app.updateQuestionHandler)))
	router.Handler(http.MethodDelete, "/v1/questions/:id", app.JWTAdminOnlyMiddleware(http.HandlerFunc(app.deleteQuestionHandler)))

	router.Handler(http.MethodGet, "/v1/answers", app.verifyJWTMiddleware(http.HandlerFunc(app.listAnswersHandler)))
	router.Handler(http.MethodPost, "/v1/answers", app.JWTAdminOnlyMiddleware(http.HandlerFunc(app.createAnswerHandler)))
	router.Handler(http.MethodPut, "/v1/answers/:id", app.JWTAdminOnlyMiddleware(http.HandlerFunc(app.updateAnswerHandler)))
	router.Handler(http.MethodDelete, "/v1/answers/:id", app.JWTAdminOnlyMiddleware(http.HandlerFunc(app.deleteAnswerHandler)))

	router.Handler(http.MethodGet, "/v1/admin/activities", app.JWTAdminOnlyMiddleware(http.HandlerFunc(app.listUserActivitiesHandler)))
	router.Handler(http.MethodGet, "/v1/activities", app.verifyJWTMiddleware(http.HandlerFunc(app.listActivitiesHandler)))
	router.Handler(http.MethodPost, "/v1/activities", app.verifyJWTMiddleware(http.HandlerFunc(app.createActivityHandler)))
	router.Handler(http.MethodPatch, "/v1/activities/:id", app.verifyJWTMiddleware(http.HandlerFunc(app.updateActivityHandler)))

	return app.recoverPanic(app.rateLimit(router))
}
