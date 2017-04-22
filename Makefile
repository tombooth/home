.PHONY: clean deps

bin/home: $(shell find . -name '*.go')
	GODEBUG=cgocheck=0 go build -o bin/home cmd/*.go

deps:
	go get ./...

clean:
	rm -rf bin/
