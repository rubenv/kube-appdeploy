.PHONY: all get-deps test

all: get-deps test

get-deps:
	go get -t ./...

test:
	go test -v ./...
