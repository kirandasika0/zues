build:
	go build -o zues main.go
run:
	make build
	./zues
clean:
	rm zues
