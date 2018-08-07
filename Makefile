TIME=$(shell date +"%d-%m-%y")
SHA = $(shell git rev-parse HEAD)
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o zues -ldflags "-s -w -X zues.main.BuildTime=$(TIME) -X zues.main.GitSHA=$(SHA)" main.go
run:
	DOCKER_ENV=true ./zues --addr=:11000
clean:
	rm zues
