package main

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
}

func main() {
	// close up the database
	defer close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/portal", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("../templates/index.gohtml")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, map[string]interface{}{
			"errors": nil,
		})
	})
	r.Post("/portal", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("../templates/index.gohtml")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// retrieve the form values
		name := r.FormValue("name")
		age := r.FormValue("age")
		if len(name) == 0 || len(age) == 0 {
			tmpl.Execute(w, map[string]interface{}{
				"errors": []string{"All the fields are required"},
			})
			return
		}

		_, fileHeader, _ := r.FormFile("file")
		_, err = newEntry(name, age, fileHeader)
		if err != nil {
			tmpl.Execute(w, map[string]interface{}{
				"errors": []string{err.Error()},
			})
			return
		}
		tmpl.Execute(w, map[string]interface{}{
			"message": "Entry saved successfully",
		})
	})

	// set up static files
	FileServer(r, "/assets/", http.Dir("../assets/"))

	log.Warn(http.ListenAndServe(":9003", r))
}

// FileServer adds a url to handle assets file in dev mode
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
