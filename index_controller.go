package main

import (
	"html/template"
	"log"
	"net/http"
)

type IndexController struct{}

func (i *IndexController) Show(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		log.Print(err)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		log.Print(err)
		return
	}
}
