.PHONY: all run test db-up db-down redis-up redis-down nsq-up nsq-down build docker-build docker-push k8s-deploy k8s-delete clean

all: run

run:
	go run cmd/api/main.go

test:
	go test ./...

db-up:
	docker-compose up -d db

db-down:
	docker-compose down

redis-up:
	docker-compose up -d redis

redis-down:
	docker-compose down

nsq-up:
	docker-compose up -d nsq

nsq-down:
	docker-compose down

build:
	go build -o neomart ./cmd/api

docker-build:
	docker build -t your-org/neomart:latest .

docker-push:
	docker push your-org/neomart:latest

k8s-deploy:
	kubectl apply -f k8s/

k8s-delete:
	kubectl delete -f k8s/

clean:
	go clean
	docker-compose down
	kubectl delete -f k8s/

sqlc-generate:
	sqlc generate

migrate-up:
	migrate -database ${POSTGRESQL_URL} -path ./migrations up

migrate-down:
	migrate -database ${POSTGRESQL_URL} -path ./migrations down

gcp-auth:
	gcloud auth login

gcp-unset-token:
	gcloud config unset auth/access_token_file

gcp-auth-application:
	gcloud auth application-default login