BINARY=github-pr-exporter
BUILD_FLAGS=-ldflags="-s -w"
PROJECT=circa10a/github-pr-exporter
VERSION=0.2.0

build:
	go build -o $(BINARY)

run:
	go run . --config examples/config.yaml

compile:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/linux/amd64/$(BINARY)
	GOOS=linux GOARCH=arm go build $(BUILD_FLAGS) -o bin/linux/arm/$(BINARY)
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/linux/arm64/$(BINARY)
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/darwin/amd64/$(BINARY)
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/windows/amd64/$(BINARY).exe

package:
	GOOS=linux GOARCH=amd64 tar -czf bin/$(BINARY)_$(VERSION)_linux_amd64.tar.gz -C bin/linux/amd64 $(BINARY)
	GOOS=linux GOARCH=arm tar -czf bin/$(BINARY)_$(VERSION)_linux_arm.tar.gz -C bin/linux/arm $(BINARY)
	GOOS=linux GOARCH=arm64 tar -czf bin/$(BINARY)_$(VERSION)_linux_arm64.tar.gz -C bin/linux/arm64 $(BINARY)
	GOOS=darwin GOARCH=amd64 tar -czf bin/$(BINARY)_$(VERSION)_darwin_amd64.tar.gz -C bin/darwin/amd64 $(BINARY)
	GOOS=windows GOARCH=amd64 zip -j bin/$(BINARY)_$(VERSION)_windows_amd64.zip bin/windows/amd64/$(BINARY).exe

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
