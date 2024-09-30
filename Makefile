PROTOC:=protoc
PROTO_PATH:=$(HOME)/go/src/github.com/koopa0/auth/
GOOGLE_PROTOBUF_PATH:=$(GOPATH)/pkg/mod/github.com/google/protobuf@v5.27.3+incompatible/src/
GO_OUT:=proto/pb
PROTO_FILES:=proto/*.proto

# Google Cloud 相關變數
PROJECT_ID := valiant-realm-435619-g6
REGION := asia-east1
REPOSITORY := goflare-micro
IMAGE_NAME := broker-service
TAG := latest

.PHONY: all run test db-up db-down redis-up redis-down nsq-up nsq-down build docker-build docker-push k8s-deploy k8s-delete clean proto sqlc-generate migrate-up migrate-down gcp-setup gcp-clean gcp-login gcp-config gcp-push gcp-list gcp-check-permissions gcp-add-permissions

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

proto:
	$(PROTOC) -I $(PROTO_PATH) \
	-I $(GOOGLE_PROTOBUF_PATH) \
	--proto_path=proto --go_out=$(GO_OUT) --go_opt=paths=source_relative \
	--go-grpc_out=$(GO_OUT) --go-grpc_opt=paths=source_relative \
	$(PROTO_FILES)

# Google Cloud 相關命令
gcp-setup: gcp-clean gcp-login gcp-config

gcp-clean:
	@echo "清理舊的配置..."
	gcloud config unset auth/access_token_file
	rm -f "$(HOME)/Library/Application Support/cloud-code/intellij/access_token_file"

gcp-login:
	@echo "設置應用默認憑證..."
	gcloud auth application-default login
	@echo "登錄主帳戶..."
	gcloud auth login

gcp-config:
	@echo "配置項目和 Docker..."
	gcloud config set project $(PROJECT_ID)
	gcloud auth configure-docker $(REGION)-docker.pkg.dev

gcp-push:
	@echo "推送 Docker 映像到 Artifact Registry..."
	docker push $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)/$(IMAGE_NAME):$(TAG)

gcp-list:
	@echo "列出 Artifact Registry 中的映像..."
	gcloud artifacts docker images list $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPOSITORY)

gcp-check-permissions:
	@echo "檢查 IAM 權限..."
	gcloud projects get-iam-policy $(PROJECT_ID) --format=json | jq '.bindings[] | select(.members[] | contains("koopapapa@gmail.com"))'

gcp-add-permissions:
	@echo "添加 Artifact Registry 管理員權限..."
	gcloud projects add-iam-policy-binding $(PROJECT_ID) --member=user:koopapapa@gmail.com --role=roles/artifactregistry.admin

help:
	@echo "可用的命令："
	@echo "  make run               - 運行應用程序"
	@echo "  make test              - 運行測試"
	@echo "  make db-up             - 啟動數據庫容器"
	@echo "  make db-down           - 停止數據庫容器"
	@echo "  make redis-up          - 啟動 Redis 容器"
	@echo "  make redis-down        - 停止 Redis 容器"
	@echo "  make nsq-up            - 啟動 NSQ 容器"
	@echo "  make nsq-down          - 停止 NSQ 容器"
	@echo "  make build             - 構建應用程序"
	@echo "  make docker-build      - 構建 Docker 映像"
	@echo "  make docker-push       - 推送 Docker 映像"
	@echo "  make k8s-deploy        - 部署到 Kubernetes"
	@echo "  make k8s-delete        - 從 Kubernetes 刪除部署"
	@echo "  make clean             - 清理項目"
	@echo "  make sqlc-generate     - 生成 SQLC 代碼"
	@echo "  make migrate-up        - 運行數據庫遷移（向上）"
	@echo "  make migrate-down      - 運行數據庫遷移（向下）"
	@echo "  make proto             - 生成 Protocol Buffers 代碼"
	@echo "  make gcp-setup         - 設置 Google Cloud 環境"
	@echo "  make gcp-push          - 推送 Docker 映像到 Artifact Registry"
	@echo "  make gcp-list          - 列出 Artifact Registry 中的映像"
	@echo "  make gcp-check-permissions - 檢查 GCP IAM 權限"
	@echo "  make gcp-add-permissions  - 添加 Artifact Registry 管理員權限"