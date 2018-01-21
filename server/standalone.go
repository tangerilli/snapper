package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"snapper"
)

type HTMLPdfInput struct {
	Html *string `json:"html"`
}

func PdfHTMLHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var input HTMLPdfInput
	err := decoder.Decode(&input)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid request body")
		log.Println("Invalid request body")
		return
	}

	data, err := snapper.PrintPdfFromHtml("http://localhost:9222", *input.Html)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error generating PDF")
		fmt.Fprintf(w, "Error generating PDF")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Snapper!")
}

// TODO
// Depending on options, either launch chrome or connect to an existing instance
// run an HTTP server which accepts requests to generate PDF's
func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", HomeHandler)

	s := r.PathPrefix("/pdf").Methods("POST").Subrouter()
	// s.HandleFunc("/url/", ScreenshotURLHandler)
	s.HandleFunc("/html/", PdfHTMLHandler)

	http.Handle("/", r)

	handler := cors.Default().Handler(r)
	log.Fatal(http.ListenAndServe(":8088", handler))
}
