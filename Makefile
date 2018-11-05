build: lint test build-amd64 build-arm 
docker-local: build-amd64 docker-amd64 build-arm docker-arm
docker-drone: docker-standalone

clean:
	@echo ">> Cleaning Release"
	@rm -rf release
	@docker rmi tchaudhry/nagios-svc:master
	@docker rmi tchaudhry/nagios-svc:armhf

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
	@go get golang.org/x/lint/golint
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

docker-amd64:
	@echo ">> Building an AMD64 Docker image"
	@docker build -f Dockerfile-amd64 -t tchaudhry/nagios-svc:master .

build-arm:
	@echo ">> Grabbing Build Dependencies"
	@go get -v ./...
	@echo ">> Register binfmt qemu bin for arm builder"
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@echo ">> Creating Release for ARM"
	@GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -o release/nagios-svc-arm ./cmd

docker-arm:
	@echo ">> Building an armhf docker image"
	@docker build -f Dockerfile-ARM -t tchaudhry/nagios-svc:armhf .

docker-standalone:
	@echo ">> Building inside docker"
	@docker build -f Dockerfile-standalone -t tchaudhry/nagios-svc:master .