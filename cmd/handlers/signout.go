package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"pengoe/config"
	"pengoe/internal/router"
	"pengoe/internal/services"
	"pengoe/internal/token"
	"time"
)

/*
Signout signs the user out by deleting the refresh token.
Access token is cleared by the client.
*/
func Signout(w http.ResponseWriter, r *http.Request, p map[string]string) error {
	db, dbFound := r.Context().Value("db").(*sql.DB)
	if !dbFound {
		router.InternalError(w, r, p)
		return errors.New("Should use db middleware")
	}

	secure := config.Env.ENVIRONMENT == "production"
	var sameSite http.SameSite
	if config.Env.ENVIRONMENT == "production" {
		sameSite = http.SameSiteLaxMode
	} else {
		sameSite = http.SameSiteDefaultMode
	}

	// delete the session from the database
	sessionService := services.NewSessionService(db)

	// get old session form cookie
	session, sessionErr := sessionService.CheckFromCookie(r)
	if sessionErr != nil {
		return sessionErr
	}

	// delete the session from the database
	deleteErr := sessionService.Delete(session.Id)
	if deleteErr != nil {
		return deleteErr
	}

	// dete the server session
	serverSessionErr := token.Manager.Delete(session.Id)
	if serverSessionErr != nil {
		return serverSessionErr
	}

	// delete the session cookie from the client
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour).UTC(),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	w.Header().Set("HX-Redirect", "/signin")

	return nil
}
