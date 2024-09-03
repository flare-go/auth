# 第一階段：構建階段
FROM golang:1.23 as builder

# 設定工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum 來利用 Docker 的構建快取機制
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製其餘的應用程式原始碼
COPY . .

# 編譯應用程式，啟用優化標誌
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/api ./cmd/api

# 第二階段：運行階段
FROM alpine:3.18

# 設定工作目錄
WORKDIR /app

# 從構建階段複製編譯後的二進制文件
COPY --from=builder /app/api .

# 複製配置文件到容器中
COPY casbin.conf .
COPY local.env .

# 設定默認環境變量
ENV ECHO_MODE=release

# 開放應用程式需要的端口（假設 Echo 預設使用 8000 端口）
EXPOSE 8000

# 設定容器啟動時運行的指令
CMD ["./api"]
