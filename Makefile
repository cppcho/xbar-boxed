.PHONY: build install test lint clean

build:
	go build -o boxed ./cmd/boxed
	go build -o boxed-xbar ./cmd/boxed-xbar

install: build
	cp boxed ~/bin/boxed
	cp boxed-xbar ~/Library/Application\ Support/xbar/plugins/boxed.1s
	mkdir -p ~/.local/lib/boxed/sounds
	cp sounds/*.ogg ~/.local/lib/boxed/sounds/

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -f boxed boxed-xbar
