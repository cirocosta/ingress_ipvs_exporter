all: install

install:
	go install -v

fmt:
	go fmt ./...
