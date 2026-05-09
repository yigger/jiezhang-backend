# jiezhang-backend (Gin)

这是一个用于从 Ruby/Rails 迁移到 Go + Gin 的学习脚手架。

## 项目结构

```text
.
├── cmd/
│   └── server/
│       └── main.go                        # 推荐入口（生产常用）
├── internal/
│   ├── bootstrap/
│   │   └── app.go                         # 组装应用（配置、路由、中间件、依赖注入）
│   ├── config/
│   │   └── config.go                      # 环境变量配置
│   ├── domain/
│   │   └── user.go                        # 领域模型（不依赖 Gin/ORM）
│   ├── http/
│   │   ├── handler/
│   │   │   ├── health_handler.go
│   │   │   ├── hello_handler.go
│   │   │   └── user_handler.go
│   │   ├── middleware/
│   │   │   └── access_log.go
│   │   └── router/
│   │       └── router.go
│   ├── repository/
│   │   ├── user_repository.go             # repository 接口定义
│   │   └── memory/
│   │       └── user_repository.go         # 内存实现（用于学习/本地开发）
│   └── service/
│       ├── hello_service.go
│       └── user_service.go
├── main.go                                # 兼容入口（go run .）
├── go.mod
└── README.md
```

## 为什么这样分层

- `handler`：处理 HTTP 输入输出（参数、状态码、JSON）
- `service`：放业务规则，不依赖 Gin
- `repository`：定义数据库访问接口，并可替换具体实现
- `domain`：业务实体定义，避免被 HTTP/ORM 污染
- `router`：集中管理路由和版本分组
- `bootstrap`：应用装配层，负责把各模块连起来
- `internal`：限制外部模块直接 import，保持边界清晰

这对应 Ruby 常见思路：

- Controller ~= `handler`
- Service Object ~= `service`
- ActiveRecord 查询职责 ~= `repository`
- Model（领域对象） ~= `domain`
- routes.rb ~= `router`
- config/initializers ~= `bootstrap + config`

## 当前用户接口（内存仓储）

- `GET /api/v1/users`：用户列表
- `GET /api/v1/users/:id`：用户详情
- `POST /api/v1/users`：创建用户

请求示例：

```bash
curl http://localhost:8080/api/v1/users
curl http://localhost:8080/api/v1/users/1
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Jie","email":"jie@example.com"}'
```

## 运行

### 1) 安装依赖

```bash
go mod tidy
```

### 2) 启动服务

推荐入口：

```bash
go run ./cmd/server
```

兼容入口：

```bash
go run .
```

### 3) 基础请求示例

```bash
curl http://localhost:8080/ping
curl http://localhost:8080/healthz
curl "http://localhost:8080/api/v1/hello?name=jie"
```

## 环境变量

- `PORT`：服务端口，默认 `8080`
- `GIN_MODE`：`debug` / `release` / `test`，默认 `debug`
- `APP_NAME`：服务名，默认 `jiezhang-backend`

示例：

```bash
APP_NAME=jiezhang-api PORT=9090 GIN_MODE=release go run ./cmd/server
```

## 下一步建议（接数据库）

1. 新建 `internal/repository/mysql/user_repository.go`（GORM/SQL 实现）
2. 在 `bootstrap/app.go` 中把 `memory.NewUserRepository()` 切换成 MySQL 实现
3. 增加迁移工具（`migrate` 或 GORM migration）和 `users` 表
4. 给 `user_service` 和 `user_handler` 增加单元测试/接口测试
