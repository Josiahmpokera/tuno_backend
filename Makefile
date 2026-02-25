.PHONY: run build docker-up docker-down docker-build docker-run tidy

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

docker-up:
	docker-compose up -d postgres redis

docker-down:
	docker-compose down

docker-build:
	docker-compose build

docker-run:
	docker-compose up --build

tidy:
	go mod tidy
