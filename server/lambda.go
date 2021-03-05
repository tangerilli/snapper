package main

import (
	"errors"
	"log"
	"snapper"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	// ErrNameNotProvided is thrown when a name is not provided
	ErrNameNotProvided = errors.New("no content was provided in the HTTP body")

	InvocationCounter = 0
)

func Handler(request snapper.Request) (snapper.Response, error) {
	InvocationCounter++
	log.Printf("Received PDF generation request v2, invocations: %d", InvocationCounter)

	var err error

	if InvocationCounter == 1 {
		_, err = snapper.LaunchChrome(nil)
		if err != nil {
			log.Printf("Error launching chrome: %s", err)
			return snapper.Response{""}, err
		}

		/*
			defer func() {
				log.Println("Killing chrome process")
				if err := cmd.Process.Kill(); err != nil {
					log.Printf("Error killing chrome process: %s", err)
				}
			}()
		*/
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
		log.Printf("Error generating PDF: %s", err)
		return snapper.Response{""}, err
	}
	return snapper.Response{string(data)}, nil
}

func main() {
	lambda.Start(Handler)
}
