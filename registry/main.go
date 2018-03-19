package main

import (
	"fmt"
	"html/template"
	"image/png"
	"net/http"
	"os"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
)

var host string

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)

	host = os.Getenv("QR_HOST")
}

func main() {
	// close up the database
	defer closeDB()

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

		_, fileHeader, err := r.FormFile("file")
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

	r.Get("/list", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("../templates/list.gohtml")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var entries []*Entry
		if entries, err = list(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, map[string]interface{}{
			"list": entries,
		})
	})

	r.Get("/entry/{id}", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("../templates/entry.gohtml")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id := chi.URLParam(r, "id")
		var entry *Entry
		if entry, err = get(id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		tmpl.Execute(w, map[string]interface{}{
			"entry": entry,
		})
	})

	r.Get("/file/{id}", func(w http.ResponseWriter, r *http.Request) {
		f, err := getFile(chi.URLParam(r, "id"))
		if err != nil || f == nil {
			errMsg := "No file found"
			if err != nil {
				errMsg = err.Error()
			}
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", f.Name))
		w.Header().Set("Content-Type", f.ContentType)
		w.Write(f.Content)
	})

	r.Get("/qr/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		URL := fmt.Sprintf("%s/entry/%s", host, id)

		// set the headers
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", id))
		w.Header().Set("Content-Type", "image/png")

		qrCode, _ := qr.Encode(URL, qr.M, qr.Auto)
		qrCode, _ = barcode.Scale(qrCode, 200, 200)

		png.Encode(w, qrCode)
	})

	// set up static files
	FileServer(r, "/assets/", http.Dir("../assets/"))

	log.Warn(http.ListenAndServe("0.0.0.0:9003", r))
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
