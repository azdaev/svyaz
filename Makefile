.PHONY: run build setup

setup:
	@test -d node_modules || npm install
	@test -f static/css/hiq.min.css || cp node_modules/hiq/dist/hiq.min.css static/css/hiq.min.css

run: setup
	go run ./cmd/server

build: setup
	go build -o bin/server ./cmd/server
