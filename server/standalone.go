package main

import (
	"encoding/json"
	"flag"
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

type Response struct {
	PdfData string `json:"pdfData"`
}

// bad global variable, bad
var chromeDebugUrl string

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

	data, err := snapper.PrintPdfFromHtml(chromeDebugUrl, *input.Html)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error generating PDF: %s\n", err)
		fmt.Fprintf(w, "Error generating PDF")
		return
	}

	response := Response{string(data)}
	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Snapper!")
}

func launchHttpServer(port int) {
	r := mux.NewRouter()

	r.HandleFunc("/", HomeHandler)

	s := r.PathPrefix("/pdf").Methods("POST").Subrouter()
	s.HandleFunc("/html/", PdfHTMLHandler)

	http.Handle("/", r)

	handler := cors.Default().Handler(r)
	log.Println("Listening for requests on port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}

func main() {
	launchChrome := flag.Bool("launch-chrome", false, "Set to true to automatically launch chrome")
	chromePath := flag.String("chrome-path", "", "The path to the chrome binary, if launching chrome")
	httpPort := flag.Int("http-port", 8088, "The port the HTTP server will listen on")
	chromeDebugUrlArg := flag.String("chrome-debug-url", "http://localhost:9222", "Where to find the chrome instance to be used for printing")

	flag.Parse()

	if *launchChrome {
		_, err := snapper.LaunchChrome(chromePath)
		if err != nil {
			log.Fatalf("Could not launch chrome: %s", err)
		}
	}

	chromeDebugUrl = *chromeDebugUrlArg
	launchHttpServer(*httpPort)
}
