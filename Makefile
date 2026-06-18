.PHONY: setup build test test-watch lint run tidy check compose-up compose-down

setup:
	mise run setup

build:
	go build -o spotify-playlist ./cmd/spotify/

test:
	mise run test

test-watch:
	mise run test-watch

lint:
	mise run lint

run:
	mise run run

tidy:
	mise run tidy

check:
	mise run check

compose-up:
	docker compose up -d

compose-down:
	docker compose down