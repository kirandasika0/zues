build:
	go build main.go

run:
	make build
	./main
clean:
	rm main