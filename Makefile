.PHONY: build test lint fuzz-lexer clean

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run

fuzz-lexer:
	go test -fuzz=FuzzLexer ./compiler/lexer/ -fuzztime=60s

clean:
	rm -rf bin/
