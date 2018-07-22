TIME=$(shell date +"%d-%m-%y")
build:
	go build -o zues -ldflags "-w -s -X main.BuildTime=$(TIME)" main.go
run:
	DOCKER_ENV=true ./zues
clean:
	rm zues
