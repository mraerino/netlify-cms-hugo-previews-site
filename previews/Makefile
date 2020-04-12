.PHONY: deps clean build

deps:
	go mod download

clean:
	rm -rf ./dist/previews

build:
	GOOS=linux GOARCH=amd64 go build -o dist/previews ./handler.go
