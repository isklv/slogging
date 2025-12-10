build:
	docker build -t app ./

compose up: build
	docker compose -f ./cmd/docker-compose.yml -p localtesting up