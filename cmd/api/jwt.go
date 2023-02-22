package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"godvanced.forstes.github.com/internal/data"
)

type jwtOptions struct {
	key     string
	expires time.Duration
}

func (app *application) generateJWT(user *data.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(app.config.jwtOptions.expires).Unix()
	claims["user"] = user.ID
	claims["email"] = user.Email
	claims["role"] = user.Role

	tokenString, err := token.SignedString([]byte(app.config.jwtOptions.key))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (app *application) extractToken(cookie *http.Cookie) (*jwt.Token, error) {
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(app.config.jwtOptions.key), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (app *application) extractClaims(cookie *http.Cookie) (jwt.MapClaims, error) {
	token, err := app.extractToken(cookie)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

func (app *application) verifyJWTMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("auth_token")
		if err != nil {
			app.forbiddenResponse(w, r, err)
			return
		}

		token, err := app.extractToken(cookie)
		if err != nil {
			app.forbiddenResponse(w, r, err)
			return
		}

		if !token.Valid {
			app.forbiddenResponse(w, r, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) JWTAdminOnlyMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("auth_token")
		if err != nil {
			app.forbiddenResponse(w, r, err)
			return
		}

		token, err := app.extractToken(cookie)
		if err != nil {
			app.forbiddenResponse(w, r, err)
			return
		}

		if !token.Valid {
			app.forbiddenResponse(w, r, err)
			return
		}

		claims, err := app.extractClaims(cookie)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		roleInt := int8(claims["role"].(float64))

		if roleInt != data.AdminRole {
			app.forbiddenResponse(w, r, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
