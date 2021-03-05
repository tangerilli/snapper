build-all: bin/pdfserver bin/linux/pdfserver bin/linux/lambda

bin/pdfserver: standalone.go src/snapper/pdfprinter.go src/snapper/chrome.go
	go build -o bin/pdfserver standalone.go

bin/linux/pdfserver: standalone.go src/snapper/pdfprinter.go src/snapper/chrome.go
	GOOS=linux go build -o bin/linux/pdfserver standalone.go

.PHONY: bin/linux/lambda
bin/linux/lambda: lambda.go src/snapper/pdfprinter.go src/snapper/chrome.go
	GOOS=linux go build -o bin/linux/lambda lambda.go

.PHONY: assets/headless-chrome
assets/headless-chrome:
	mkdir -p assets
	cd assets && rm -rf headless-chrome
	cd assets && rm stable-headless-chromium* || true

	# using a version newer than chrome 64 (e.g. chrome 65) currently does not work
	# v1.0.0-38/stable-headless-chromium-amazonlinux-2017-03.zip
	cd assets && wget https://github.com/adieuadieu/serverless-chrome/releases/download/v1.0.0-38/stable-headless-chromium-amazonlinux-2017-03.zip
	cd assets && unzip stable-headless-chromium-amazonlinux-2017-03

	# v1.0.0-39/stable-headless-chromium-amazonlinux-2017-03.zip
	# cd assets && wget https://github.com/adieuadieu/serverless-chrome/releases/download/v1.0.0-39/stable-headless-chromium-amazonlinux-2017-03.zip
	# cd assets && unzip stable-headless-chromium-amazonlinux-2017-03.zip

	# v1.0.0-57/stable-headless-chromium-amazonlinux-2.zip
	# cd assets && wget https://github.com/adieuadieu/serverless-chrome/releases/download/v1.0.0-57/stable-headless-chromium-amazonlinux-2.zip
	# cd assets && unzip stable-headless-chromium-amazonlinux-2.zip

	mkdir -p assets/headless-chrome && mv assets/headless-chromium assets/headless-chrome/headless_shell

.PHONY: snapper.zip
snapper.zip: bin/linux/lambda lambda.go src/snapper/pdfprinter.go src/snapper/chrome.go assets/headless-chrome
	cp bin/linux/lambda assets
	cd assets && rm -f snapper*.zip && zip -r snapper_v2.zip lambda headless-chrome

S3_BUCKET?=us-west-1.tony.angerilli.ca
S3_KEY?=lambda/snapper.zip
upload-lambda: snapper.zip
	aws s3 cp assets/snapper.zip s3://$(S3_BUCKET)/$(S3_KEY)

update-lambda: snapper.zip
	aws lambda update-function-code --function-name snapper-pdf-printer --zip-file fileb://./assets/snapper_v2.zip --region us-east-1

docker-image: bin/linux/pdfserver standalone.go src/snapper/pdfprinter.go src/snapper/chrome.go
	cp bin/linux/pdfserver assets
	cd assets && docker build -t tangerilli/snapper .

clean:
	rm -rf bin/*
