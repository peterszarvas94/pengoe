package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/mail"
	"pengoe/internal/logger"
	"pengoe/internal/router"
	"pengoe/internal/services"
	"pengoe/internal/utils"
	"pengoe/web/templates/pages"

	"github.com/a-h/templ"
)

/*
SignupPage handles the GET request to /signup.
*/
func SignupPage(w http.ResponseWriter, r *http.Request, p map[string]string) error {
	redirect, found := r.Context().Value("redirect").(string)
	if !found {
		router.InternalError(w, r, p)
		return errors.New("Should use redirect middleware")
	}

	data := pages.SignupProps{
		Title:         "pengoe - Sign in",
		Description:   "Sign in to pengoe",
		RedirectUrl:   redirect,
		Username:      "",
		Email:         "",
		Firstname:     "",
		Lastname:      "",
		UsernameCheck: "",
		UsernameError: "",
		EmailCheck:    "",
		EmailError:    "",
	}

	component := pages.Signup(data)
	handler := templ.Handler(component)
	handler.ServeHTTP(w, r)

	return nil
}

/*
Signup handles the POST request to /signup.
*/
func Signup(w http.ResponseWriter, r *http.Request, p map[string]string) error {
	redirect, found := r.Context().Value("redirect").(string)
	if !found {
		router.InternalError(w, r, p)
		return errors.New("Should use redirect middleware")
	}
	db, found := r.Context().Value("db").(*sql.DB)
	if !found {
		router.InternalError(w, r, p)
		return errors.New("Should use db middleware")
	}

	err := r.ParseForm()
	if err != nil {
		router.InternalError(w, r, p)
		return err
	}

	form := r.Form

	username := html.EscapeString(form.Get("username"))
	if username == "" {
		router.BadRequest(w, r, p)
		return errors.New("Username is required")
	}

	email := html.EscapeString(form.Get("email"))
	if email == "" {
		router.BadRequest(w, r, p)
		return errors.New("Email is required")
	}

	firstname := html.EscapeString(form.Get("firstname"))
	if firstname == "" {
		router.BadRequest(w, r, p)
		return errors.New("Firstname is required")
	}

	lastname := html.EscapeString(form.Get("lastname"))
	if lastname == "" {
		router.BadRequest(w, r, p)
		return errors.New("Lastname is required")
	}

	password := html.EscapeString(form.Get("password"))
	if password == "" {
		router.BadRequest(w, r, p)
		return errors.New("Password is required")
	}

	// create user service
	userService := services.NewUserService(db)

	// add user
	id := utils.NewUUID("usr")

	err = userService.Signup(id, username, email, firstname, lastname, password)

	// unsuccessful signup, render signup page with error message
	if err != nil {
		log := logger.Get()

		log.Error(err.Error())

		_, parseErr := mail.ParseAddress(email)
		_, usernameQueryErr := userService.GetByUsername(username)
		_, emailQueryErr := userService.GetByEmail(email)

		emailInvalid := parseErr != nil
		usernameExists := usernameQueryErr == nil
		emailExists := emailQueryErr == nil

		usernameError := ""
		usernameCheck := ""
		emailError := ""
		emailCheck := ""

		if emailInvalid {
			emailError = "Invalid email"
			emailCheck = "incorrect"
		}

		if usernameExists {
			usernameError = "Username already exists"
			usernameCheck = "incorrect"
		} else {
			usernameCheck = "correct"
		}

		if emailExists {
			emailError = "Email already exists"
			emailCheck = "incorrect"
		} else if !emailInvalid {
			emailCheck = "correct"
		}

		w.WriteHeader(http.StatusConflict)

		data := pages.SignupProps{
			Title:         "pengoe - Sign in",
			Description:   "Sign in to pengoe",
			RedirectUrl:   redirect,
			Firstname:     firstname,
			Lastname:      lastname,
			Username:      username,
			UsernameCheck: usernameCheck,
			UsernameError: usernameError,
			Email:         email,
			EmailCheck:    emailCheck,
			EmailError:    emailError,
		}

		component := pages.Signup(data)
		handler := templ.Handler(component)
		handler.ServeHTTP(w, r)

		return err
	}

	// successful signup, redirect to signin page
	http.Redirect(
		w,
		r,
		fmt.Sprintf("/signin?redirect=%s", redirect),
		http.StatusSeeOther,
	)

	return nil
}
