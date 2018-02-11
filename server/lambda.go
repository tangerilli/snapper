package main

import (
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"snapper"
)

var (
	// ErrNameNotProvided is thrown when a name is not provided
	ErrNameNotProvided = errors.New("no content was provided in the HTTP body")
)

// TODO: Add print parameters
type Request struct {
	Html string `json:"html"`
}

type Response struct {
	PdfData string `json:"pdfData"`
}

func Handler(request Request) (Response, error) {
	log.Println("Received PDF generation request")
	_, err := snapper.LaunchChrome(nil)
	if err != nil {
		log.Printf("Error launching chrome: %s", err)
		return Response{""}, err
	}

	data, err := snapper.PrintPdfFromHtml("http://localhost:9222", request.Html)
	if err != nil {
		log.Println("Error generating PDF")
		return Response{""}, err
	}
	return Response{string(data)}, nil
}

func main() {
	lambda.Start(Handler)
}
