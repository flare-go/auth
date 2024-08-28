
# Auth 微服務模組

`auth` 是一個身份驗證微服務模組，提供了 gRPC 和 RESTful 兩種 API 接口，用於處理用戶的註冊、登錄、身份驗證、角色管理及權限控制等功能。該模組採用了 Casbin 作為授權工具，並使用 PASETO 進行安全的 Token 管理。未來將會加入 OAuth2 和 兩步驗證（2MFA）等功能。

## 目錄
- [功能特性](#功能特性)
- [安裝](#安裝)
- [使用方法](#使用方法)
    - [gRPC 接口](#grpc-接口)
    - [RESTful API](#restful-api)
- [資料庫結構](#資料庫結構)
- [系統架構](#系統架構)
- [未來計畫](#未來計畫)
- [技術細節](#技術細節)
- [開發指南](#開發指南)

## 功能特性
- **用戶註冊與登錄**: 提供用戶註冊、登錄、登出等基本功能。
- **Token 管理**: 使用 PASETO 生成與驗證安全的 Token。
- **角色與權限管理**: 支持角色創建、分配以及權限控制，並基於 Casbin 進行細粒度的授權管理。
- **gRPC 和 RESTful API**: 提供兩種 API 接口，方便整合到不同的系統中。
- **高安全性**: 採用了 Bcrypt 密碼哈希和 Ed25519 密鑰加密，保護用戶敏感信息。

## 安裝
在你的 Golang 專案中引入 `auth` 模組：

```bash
go get goflare.io/auth
```

## 使用方法

### gRPC 接口

`auth` 模組提供以下 gRPC 服務接口：

- **Register**: 用戶註冊
- **Login**: 用戶登錄
- **Logout**: 用戶登出
- **GenerateToken**: 生成新的 Token
- **RefreshToken**: 刷新 Token
- **ValidateToken**: 驗證 Token 的有效性

範例使用方法：

```go
package main

import (
  "context"
  "goflare.io/auth"
  "google.golang.org/grpc"
  "log"
)

func main() {
  conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
  if err != nil {
    log.Fatalf("did not connect: %v", err)
  }
  defer conn.Close()

  client := auth.NewAuthServiceClient(conn)

  response, err := client.Register(context.Background(), &auth.RegisterRequest{
    Username: "testuser",
    Password: "password123",
    Email:    "testuser@example.com",
  })
  if err != nil {
    log.Fatalf("could not register: %v", err)
  }

  log.Printf("Register Response: %v", response)
}
```

### RESTful API

提供的 RESTful API 與 gRPC 接口功能對應，可以使用任何 HTTP 客戶端進行調用，如 `curl` 或 `Postman`。

範例使用方法：
```bash
curl -X POST http://localhost:8080/auth/register -d '{
    "username": "testuser",
    "password": "password123",
    "email": "testuser@example.com"
}'
```

## 資料庫結構
`auth` 模組使用 PostgreSQL 進行數據存儲，以下是資料庫的主要結構設計：

- **users**: 存儲用戶基本信息。
- **roles**: 存儲系統角色。
- **permissions**: 存儲權限及其對應的資源與操作。
- **user_roles**: 記錄用戶與角色的對應關係。
- **role_permissions**: 記錄角色與權限的對應關係。

資料庫結構的 SQL 定義：
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE CHECK (position('@' in email) > 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    resource resource_type NOT NULL,
    action action_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE user_roles (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE role_permissions (
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
```

## 系統架構
`auth` 模組基於 Golang 開發，使用 Casbin 作為授權控制工具，並使用 PASETO 進行 Token 管理。所有 API 通過 gRPC 和 RESTful 接口暴露，方便不同系統的集成。

## 未來計畫
- **OAuth2 整合**: 將集成 OAuth2 來支持第三方登入認證。
- **2MFA 兩步驗證**: 提供更高級別的安全性，支持兩步驗證功能。

## 技術細節
- **Golang**: 使用 Go 1.23.0 作為主要開發語言。
- **Casbin**: 用於角色和權限的細粒度授權控制。
- **PASETO**: 採用 PASETO 作為 Token 的生成與驗證工具，提供高安全性的 Token 管理。
- **PostgreSQL**: 作為數據存儲的後端數據庫，採用 SQL 語句進行資料表管理。

## 開發指南
如需在本地進行開發和測試，可以使用以下命令啟動服務：

```bash
go run ./cmd/api/
```

在開發過程中，你可以根據需求修改 `casbin.conf` 來適配不同的授權需求，或者根據你的業務邏輯擴展 PASETO 的使用方法。
