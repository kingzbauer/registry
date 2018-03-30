package main

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
}

func getAdd(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("../templates/index.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, map[string]interface{}{
		"errors":     nil,
		"categories": CategoryMap,
	})
}

func postAdd(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("../templates/index.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lisence := &LicenseEntry{}
	r.ParseForm()
	if err = decoder.Decode(lisence, r.PostForm); err != nil {
		tmpl.Execute(w, map[string]interface{}{
			"errors":     []string{err.Error()},
			"categories": CategoryMap,
		})
		return
	}

	if err = lisence.Save(); err != nil {
		tmpl.Execute(w, map[string]interface{}{
			"errors":     []string{err.Error()},
			"categories": CategoryMap,
		})
	}

	log.Warn("SAVED LICENSE")
	tmpl.Execute(w, map[string]interface{}{
		"errors":     nil,
		"categories": CategoryMap,
		"message":    "License created successfully",
	})
}

func licenseByCategory(w http.ResponseWriter, r *http.Request) {
	category := Category(chi.URLParam(r, "category"))

	tmpl, err := template.ParseFiles("../templates/license-category.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	licenses, err := LicenseByCategory(Category(category))
	if err != nil {
		tmpl.Execute(w, map[string]interface{}{
			"error": err,
		})
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"licenses": licenses,
		"category": CategoryMap[category],
	})
}
