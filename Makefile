.PHONY: build run tidy vet clean

build:
	go build -o bin/server .

run: build
	PORT=8080 ./bin/server

tidy:
	go mod tidy

vet:
	go vet ./...

clean:
	rm -rf bin/
