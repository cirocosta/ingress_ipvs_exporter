all: install

install:
	go install -v

test:
	go test -v ./...

fmt:
	go fmt ./...
