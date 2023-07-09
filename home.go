package main

import (
	"html/template"
	"net/http"
)

// IndexPage template data
type IndexPage struct {
	Title template.HTML
	Data  [][]float64
}

// HandleHome Handle Home route
func HandleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Last-Modified", startTime.UTC().Format(http.TimeFormat))

	err := templates.ExecuteTemplate(w, "scatter.html", IndexPage{
		Title: "GT7 Telemetry",
		Data:  data, // hugly global
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleFavicon Send images/favicon.ico directly from /favicon.ico
func HandleFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./images/favicon.ico")
}
