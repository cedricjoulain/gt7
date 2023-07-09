package main

import (
	"html/template"
	"net/http"
)

// IndexPage template data
type IndexPage struct {
	Title template.HTML
	Data1 [][]float64
	Data2 [][]float64
}

// HandleHome Handle Home route
func HandleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Last-Modified", startTime.UTC().Format(http.TimeFormat))

	err := templates.ExecuteTemplate(w, "scatter.html", IndexPage{
		Title: "GT7 Telemetry",
		Data1: data1, // hugly global
		Data2: data2, // hugly global
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleFavicon Send images/favicon.ico directly from /favicon.ico
func HandleFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./images/favicon.ico")
}
