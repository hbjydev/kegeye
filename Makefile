.PHONY: build clean deploy

build: deps
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w -extldflags '-static'" -o bin/sls sls/main.go

deps:
	go mod download

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
