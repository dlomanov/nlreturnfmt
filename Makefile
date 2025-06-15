LOCAL_BIN := $(shell pwd)/bin

bin:
	mkdir -p bin

build:
	go build -o bin/nlreturnfmt ./cmd/nlreturnfmt

run: build
	./bin/nlreturnfmt

