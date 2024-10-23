BINARY=github-pr-exporter
BUILD_FLAGS=-ldflags="-s -w"
PROJECT=circa10a/github-pr-exporter
VERSION=0.0.0-dev

build:
	go build -o $(BINARY)

run:
	go run . --config examples/config.yaml

clean:
	go clean
	rm -rf $(BINARY) bin/

lint:
	golangci-lint run -v

release: clean compile package

docker-build:
	docker build -t $(PROJECT):$(VERSION) .

docker-run:
	docker run --rm -v $(shell pwd)/config.yaml:/config.yaml $(PROJECT):$(VERSION)

docker-release: docker-build
docker-release:
	echo "${DOCKER_PASSWORD}" | docker login -u ${DOCKER_USERNAME} --password-stdin
	docker push $(PROJECT):$(VERSION)
