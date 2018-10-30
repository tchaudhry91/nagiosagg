all: build-amd64 build-arm
build: build-amd64

clean:
	@echo ">> Cleaning Release"
	@rm -rf release

test:
	@echo ">> Grabbing Test Dependencies"
	@go get -t ./...
	@echo ">> Running Tests"
	@go test -v ./...

test-short:
	@echo ">> Grabbing Test Dependencies"
	@go get -t ./...
	@echo ">> Running Short Tests"
	@go test -v -short ./...

build-amd64:
	@echo ">> Creating Release for AMD64"
	@GOOS=linux GOARCH=amd64 go build -o release/nagios-svc ./cmd

build-arm:
	@echo ">> Creating Release for ARM"
	@GOOS=linux GOARCH=arm go build -o release/nagios-svc-arm ./cmd