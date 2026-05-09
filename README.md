# jiezhang-backend (Gin)

这是一个用于从 Ruby/Rails 迁移到 Go + Gin 的学习脚手架。

## 项目结构

```text
.
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── bootstrap/
│   │   └── app.go                         # 应用装配（MySQL + repository 注入）
│   ├── config/
│   │   └── config.go                      # 从根目录 .env + 环境变量加载配置
│   ├── domain/
│   │   └── user.go
│   ├── http/
│   │   ├── handler/
│   │   ├── middleware/
│   │   └── router/
│   ├── infrastructure/
│   │   └── db/
│   │       └── mysql.go                   # GORM MySQL 连接初始化
│   ├── repository/
│   │   ├── user_repository.go             # 接口定义
│   │   ├── memory/
│   │   └── mysql/
│   │       └── user_repository.go         # GORM 实现（含 users AutoMigrate）
│   └── service/
├── .env.example
├── API.md
├── main.go
└── go.mod
```

## 配置（.env）

在项目根目录创建 `.env`（可从 `.env.example` 复制）：

```bash
cp .env.example .env
```

默认示例：

```env
APP_NAME=jiezhang-backend
PORT=10240
GIN_MODE=debug
MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/jiezhang?charset=utf8mb4&parseTime=True&loc=Local
```

说明：

- `config.Load()` 会先尝试读取根目录 `.env`
- 已存在的系统环境变量优先生效（不会被 `.env` 覆盖）
- `MYSQL_DSN` 现在是必填；未配置会启动失败

## 启动

```bash
go mod tidy
go run ./cmd/server
```

## 请求生命周期（以 GET /api/users 为例）

1. **程序启动与依赖装配**
- 入口：`cmd/server/main.go`
- 调用 `bootstrap.NewApp()`：
  - `config.Load()` 读取 `.env`
  - `db.NewMySQL(cfg.MySQLDSN)` 建立 GORM 连接
  - `mysql.NewUserRepository(mysqlDB)` 初始化仓储（并 AutoMigrate）
  - `service.NewUserService(userRepo)` 初始化业务层
  - `handler.NewUserHandler(userService)` 和 `handler.NewUsersAPIHandler(...)`
  - `router.Register(engine, usersHandler)` 注册路由

2. **请求进入 Gin**
- 客户端请求：`GET /api/users`
- Gin 先经过全局中间件：
  - `gin.Recovery()`
  - `middleware.AccessLog()`

3. **路由匹配**
- `internal/http/router/router.go` 中 `api.GET("/users", usersHandler.GetUserInfo)` 命中。

4. **Handler 层（HTTP 适配层）**
- `UsersAPIHandler.GetUserInfo` 目前转调 `UserHandler.List`。
- `UserHandler.List` 只做 HTTP 职责：
  - 取 request context
  - 调用 service
  - 把结果转成 JSON + 状态码

5. **Service 层（业务规则层）**
- `UserService.List(ctx)` 处理业务流程。
- 当前实现是调用 repository，不直接依赖 Gin/GORM。

6. **Repository 层（数据访问层）**
- `mysql.UserRepository.List(ctx)` 用 GORM 查询 `users` 表。
- 将数据库模型转换为 `domain.User` 返回。

7. **响应返回**
- Handler 将数据写回客户端（`200` + `{"data": ...}`）。
- 中间件记录请求日志，请求结束。

## 这一套分层的意义

- Handler：只关心 HTTP
- Service：只关心业务
- Repository：只关心数据存取
- Domain：只关心业务对象

你后续实现其它接口（如 `POST /api/statements`）也按同样生命周期走就行。

## 当前状态

- API 路由已按 `API.md` 注册
- 未实现逻辑的方法统一返回 `501 not implemented`
- `GET /api/users` 已切到 MySQL repository（GORM）
