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

func Handler(request snapper.Request) (snapper.Response, error) {
	log.Println("Received PDF generation request")
	_, err := snapper.LaunchChrome(nil)
	if err != nil {
		log.Printf("Error launching chrome: %s", err)
		return snapper.Response{""}, err
	}
	chromeDebugUrl := "http://localhost:9222"

	var data []byte
	options := request.Options
	if options == nil {
		options = new(snapper.Options)
	}
	snapper.SetDefaultOptions(options)
	if request.Html != nil {
		data, err = snapper.PrintPdfFromHtml(chromeDebugUrl, options, *request.Html)
	} else {
		data, err = snapper.PrintPdfFromUrl(chromeDebugUrl, options, *request.Url)
	}
	if err != nil {
		log.Println("Error generating PDF")
		return snapper.Response{""}, err
	}
	return snapper.Response{string(data)}, nil
}

func main() {
	lambda.Start(Handler)
}
