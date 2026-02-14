.PHONY: install sync test lint format

install:
	./install.sh

sync:
	uv sync

test:
	uv run pytest tests/ -v

lint:
	uvx ruff check .

format:
	uvx ruff format .
