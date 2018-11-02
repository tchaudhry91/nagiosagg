all: lint build-amd64 build-arm
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

lint:
	@echo ">> Grabbing Build Dependencies"
	@go get -v ./...
	@go get github.com/golang/lint/golint
	@go get github.com/fzipp/gocyclo
	GO_FILES=$(find . -iname '*.go' -type f | grep -v /vendor/)
	@test -z $(gofmt -s -l $GO_FILES)
	@echo ">> Running basic quality checks"
	@go vet ./...
	@golint $(go list ./...)

build-amd64:
	@echo ">> Grabbing Build Dependencies"
	@go get -v ./...
	@echo ">> Creating Release for AMD64"
	@GOOS=linux GOARCH=amd64 go build -o release/nagios-svc ./cmd

build-arm:
	@echo ">> Grabbing Build Dependencies"
	@go get -v ./...
	@echo ">> Creating Release for ARM"
	@GOOS=linux GOARCH=arm go build -o release/nagios-svc-arm ./cmd