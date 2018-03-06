build-all: bin/pdfserver bin/linux/pdfserver bin/linux/lambda

bin/pdfserver: standalone.go src/snapper/pdfprinter.go src/snapper/chrome.go
	go build -o bin/pdfserver standalone.go

bin/linux/pdfserver: standalone.go src/snapper/pdfprinter.go src/snapper/chrome.go
	GOOS=linux go build -o bin/linux/pdfserver standalone.go

bin/linux/lambda: lambda.go src/snapper/pdfprinter.go src/snapper/chrome.go
	GOOS=linux go build -o bin/linux/lambda lambda.go

snapper.zip: bin/linux/lambda lambda.go src/snapper/pdfprinter.go src/snapper/chrome.go
	cp bin/linux/lambda assets
	cd assets && rm -f snapper.zip && zip -r snapper.zip lambda headless-chrome

deploy-lambda: snapper.zip
	aws s3 cp assets/snapper.zip s3://us-west-1.tony.angerilli.ca/lambda/snapper.zip
	python deploy.py

update-lambda: snapper.zip
	aws lambda update-function-code --function-name snapper-us-west-1-lambda-PdfPrinter-1DL19DSUOD7H0 --zip-file fileb://./assets/snapper.zip --region us-west-1

docker-image: bin/linux/pdfserver standalone.go src/snapper/pdfprinter.go src/snapper/chrome.go
	cp bin/linux/pdfserver assets
	cd assets && docker build -t tangerilli/snapper .

clean:
	rm -rf bin/*