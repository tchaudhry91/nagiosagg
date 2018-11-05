DOCKER_IMAGE         ?= tchaudhry/nagios-svc
DOCKER_IMAGE_TAG     ?= master
DOCKER_IMAGE_TAG_ARM ?= armhf

build: lint test build-amd64 build-arm 
docker-local: build-amd64 docker-amd64 build-arm docker-arm
docker-drone: docker-standalone
docker-push: docker-push-amd64 docker-push-armhf

clean:
	@echo ">> Cleaning Release"
	@rm -rf release
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG)
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG_ARM)

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
	@GO_FILES=$(find . -iname '*.go' -type f | grep -v /vendor/) test -z $(gofmt -s -l $GO_FILES)
	@echo ">> Running basic quality checks"
	@go vet ./...
	@golint $(go list ./...)

build-amd64:
	@echo ">> Grabbing Build Dependencies"
	@go get -v ./...
	@echo ">> Creating Release for AMD64"
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o release/nagios-svc ./cmd

docker-amd64:
	@echo ">> Building an AMD64 Docker image"
	@docker build -f Dockerfile-amd64 -t $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) .

docker-push-amd64:
	@docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
	@docker push $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG)

build-arm:
	@echo ">> Grabbing Build Dependencies"
	@go get -v ./...
	@echo ">> Register binfmt qemu bin for arm builder"
	@echo ">> Creating Release for ARM"
	@GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -o release/nagios-svc-arm ./cmd

docker-arm:
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@echo ">> Building an armhf docker image"
	@docker build -f Dockerfile-ARM -t $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG_ARM) .

docker-push-armhf:
	@docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
	@docker push $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG_ARM)

docker-standalone:
	@echo ">> Building inside docker"
	@docker build -f Dockerfile-standalone -t $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) .
