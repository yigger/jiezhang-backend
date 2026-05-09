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
│   │   └── app.go
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   └── user.go
│   ├── http/
│   │   ├── handler/
│   │   ├── middleware/
│   │   └── router/
│   ├── infrastructure/
│   │   ├── db/
│   │   │   └── mysql.go
│   │   ├── sessioncache/
│   │   │   └── cache.go
│   │   └── wechat/
│   │       └── client.go
│   ├── repository/
│   │   ├── user_repository.go
│   │   └── mysql/
│   │       └── user_repository.go
│   └── service/
│       ├── user_service.go
│       └── auth/
│           └── check_openid_service.go
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
MINIPROGRAM_APPID=your_miniprogram_appid
MINIPROGRAM_SECRET=your_miniprogram_secret
SESSION_TOKEN_SECRET=replace_with_a_long_random_secret
```

说明：

- `config.Load()` 会先尝试读取根目录 `.env`
- 已存在的系统环境变量优先生效（不会被 `.env` 覆盖）
- `MYSQL_DSN`、`MINIPROGRAM_APPID`、`MINIPROGRAM_SECRET`、`SESSION_TOKEN_SECRET` 均为必填

## 启动

```bash
go mod tidy
go run ./cmd/server
```

## check_openid（Go 版）

- 路径：`POST /api/v1/check_openid`
- 请求头：`X-WX-Code: <wx_login_code>`
- 成功响应：`{"status":200,"session":"<third_session>"}`
- 失败响应：`{"status":401,"msg":"登录失败"}`

示例：

```bash
curl -X POST http://localhost:10240/api/v1/check_openid \
  -H "X-WX-Code: YOUR_WECHAT_LOGIN_CODE"
```

实现行为对齐 Rails `login_controller.rb`：

1. 调微信 `jscode2session`（超时重试）
2. 用 `openid` 查找或创建用户
3. 命中缓存则直接返回 session
4. 否则生成 `third_session`，写回用户并缓存 2 天

## 当前状态

- API 路由已按 `API.md` 注册
- 未实现逻辑的方法统一返回 `501 not implemented`
- `GET /api/v1/users` 与 `POST /api/v1/check_openid` 已接入 MySQL + 业务逻辑
