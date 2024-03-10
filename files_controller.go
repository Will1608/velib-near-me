package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type FilesController struct{}

func (c *FilesController) Show(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("name")
	var f []byte
	var err error
	if filename == "leaflet.css" || filename == "leaflet.js" || filename == "velib.png" || filename == "pin.png" {
		f, err = os.ReadFile(filename)
		if err != nil {
			defer handleHttpError(w, err)
			return
		}
	} else {
		defer handleHttpError(w, errors.New(fmt.Sprintf("unrecongnized filename: %s", filename)))
		return
	}

	switch filepath.Ext(filename) {
	case ".js":
		w.Header().Add("content-type", "application/javascript; charset=utf-9")
	case ".css":
		w.Header().Add("content-type", "text/css; charset=utf-9")
	case ".png":
		w.Header().Add("content-type", "image/png")
	}

	_, err = w.Write(f)
	if err != nil {
		defer handleHttpError(w, err)
		return
	}
}
