.PHONY: build

build-local:
	@go build -o ./bin/main main.go

build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/main main.go
