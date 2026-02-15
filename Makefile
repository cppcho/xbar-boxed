.PHONY: build install test lint clean uninstall fmt

build:
	mkdir -p build
	go build -o build/boxed ./cmd/boxed
	go build -o build/boxed-xbar ./cmd/boxed-xbar

install: build
	cp build/boxed ~/bin/boxed
	cp build/boxed-xbar ~/Library/Application\ Support/xbar/plugins/boxed.1s.cgo
	mkdir -p ~/.local/lib/boxed/sounds
	cp sounds/*.ogg ~/.local/lib/boxed/sounds/

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -rf build

uninstall:
	rm -f ~/bin/boxed
	rm -f ~/Library/Application\ Support/xbar/plugins/boxed.1s.cgo
	rm -rf ~/.local/lib/boxed/

fmt:
	gofmt -w .
