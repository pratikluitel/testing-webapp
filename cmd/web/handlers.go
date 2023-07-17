package main

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
	"webapp/pkg/data"
)

var pathToTemplates = "./templates/"

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var templateData = make(map[string]any)
	if app.Session.Exists(r.Context(), "test") {
		templateData["test"] = app.Session.GetString(r.Context(), "test")
	} else {
		app.Session.Put(r.Context(), "test", "Hit this page at "+time.Now().UTC().String())
	}
	_ = app.render(w, r, "home.page.gohtml", &TemplateData{Data: templateData})

}

func (app *application) Profile(w http.ResponseWriter, r *http.Request) {

	_ = app.render(w, r, "profile.page.gohtml", &TemplateData{})

}

type TemplateData struct {
	IP    string
	Data  map[string]any
	Error string
	Flash string
	User  data.User // currently authenticated user
}

func (app *application) render(w http.ResponseWriter, r *http.Request, tmpl string, td *TemplateData) error {
	// parse the template from disk
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, tmpl), path.Join(pathToTemplates, "base.layout.gohtml"))

	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return err
	}

	td.IP = app.ipFromContext(r.Context())

	td.Error = app.Session.PopString(r.Context(), "error")
	td.Flash = app.Session.PopString(r.Context(), "flash")

	// execute the template
	err = parsedTemplate.Execute(w, td)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	//validate data
	form := NewForm(r.PostForm)

	form.Required("email", "password")

	if !form.Valid() {
		// redirect to the login page with an error message
		app.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		app.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Println(password, user.FirstName)

	// authenticate user
	// if not authenticated, redirect with error
	if !app.authenticate(r, user, password) {
		app.Session.Put(r.Context(), "error", "Invalid login!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	// prevent session fixation attack
	// we renew session token every time page is reloaded
	_ = app.Session.RenewToken(r.Context())

	// store success message in session

	// redirect to some other page, a profile page
	app.Session.Put(r.Context(), "flash", "Successfully logged in!")
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}

func (app *application) authenticate(r *http.Request, user *data.User, password string) bool {
	if valid, err := user.PasswordMatches(password); err != nil || !valid {
		return false
	}
	app.Session.Put(r.Context(), "user", user)
	return true
}
