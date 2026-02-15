.PHONY: build install test lint clean uninstall fmt

build:
	mkdir -p build
	go build -o build/boxed ./cmd/boxed
	go build -o build/boxed-xbar ./cmd/boxed-xbar

install: build
	mkdir -p ~/bin
	mkdir -p ~/Library/Application\ Support/xbar/plugins/
	cp build/boxed ~/bin/boxed
	cp build/boxed-xbar ~/Library/Application\ Support/xbar/plugins/boxed.1s.cgo
	mkdir -p ~/.local/lib/boxed/sounds
	cp sounds/*.ogg ~/.local/lib/boxed/sounds/

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -r build

uninstall:
	rm ~/bin/boxed
	rm ~/Library/Application\ Support/xbar/plugins/boxed.1s.cgo
	rm -r ~/.local/lib/boxed/

fmt:
	gofmt -w .
