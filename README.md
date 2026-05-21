# jiezhang-backend

> Go + Gin 实现的记账小程序后端

## 技术栈

| 组件 | 选型 |
|------|------|
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) v1.12 |
| ORM | [GORM](https://gorm.io) v1.31 + MySQL 驱动 |
| 数据库 | MySQL 8.x（复用 Rails 项目的库） |
| 缓存 | Redis（可选，未配置时用内存缓存） |
| 微信登录 | `jscode2session` API |
| Excel 导出 | [excelize](https://github.com/xuri/excelize) v2 |
| 签名 URL | HMAC-SHA256（防私密文件直接访问） |

## 快速开始

### 环境要求

- Go 1.26+
- MySQL 8.x（已有 Rails 数据库可直接复用）
- Redis（可选，建议生产环境配置）

### 启动

```bash
# 1. 克隆项目
git clone https://github.com/yigger/jiezhang-backend.git
cd jiezhang-backend

# 2. 创建环境配置
cp .env.example .env
# 编辑 .env，填入真实的数据库连接、小程序密钥等

# 3. 安装依赖并运行
go mod tidy
go run .
```

服务默认监听 `http://localhost:10240`。

### 环境变量

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `MYSQL_DSN` | ✅ | — | MySQL 连接串 |
| `MINIPROGRAM_APPID` | ✅ | — | 微信小程序 AppID |
| `MINIPROGRAM_SECRET` | ✅ | — | 微信小程序 Secret |
| `SESSION_TOKEN_SECRET` | ✅ | — | 会话签名密钥 |
| `PORT` | 否 | `10240` | 监听端口 |
| `PUBLIC_BASE_URL` | 否 | `http://localhost:<PORT>` | 生成图片/文件链接的域名 |
| `REDIS_URL` | 否 | — | Redis 连接（不配则用内存缓存） |
| `GIN_MODE` | 否 | `debug` | `debug` / `release` / `test` |

## 项目结构

```
.
├── main.go                          # 入口
├── cmd/server/main.go               # 备选入口（内容相同）
├── go.mod
├── .env.example                     # 环境变量模板
├── API.md                           # 完整 API 接口文档
├── public/                          # 静态文件（图片、上传附件）
│   ├── images/
│   │   ├── asset/                   # 资产图标
│   │   └── category/                # 分类图标
│   └── private/                     # 私有文件（含签名鉴权）
└── internal/
    ├── bootstrap/
    │   ├── app.go                   # 应用启动组装（依赖注入）
    │   └── modules/                 # 15 个模块组装器
    │       ├── auth_module.go
    │       ├── user_module.go
    │       ├── home_module.go
    │       ├── statement_module.go
    │       ├── ...（每领域一个文件）
    │       └── super_module.go
    ├── config/
    │   └── config.go               # 配置加载（.env + 环境变量）
    ├── domain/                      # 领域实体
    │   ├── user.go
    │   ├── statement.go
    │   ├── category.go
    │   ├── account_book.go
    │   └── payee.go
    ├── repository/                  # 数据访问接口定义
    │   ├── user_repository.go
    │   ├── statement_repository.go
    │   ├── ...（每聚合一个接口）
    │   └── mysql/                   # GORM / MySQL 实现
    │       ├── user_repository.go
    │       ├── statement_repository.go
    │       └── ...
    ├── service/                     # 业务逻辑层
    │   ├── statement_service.go
    │   ├── statement_export_excel.go
    │   ├── home_service.go
    │   ├── ...（每领域一个服务）
    │   ├── auth/
    │   │   └── check_openid_service.go
    │   ├── statement/
    │   │   ├── types.go             # DTO + 响应类型
    │   │   └── mapper.go            # DB row → DTO 映射
    │   └── helper/
    │       └── statement_helper.go
    ├── http/
    │   ├── router/
    │   │   └── router.go           # 83 个路由注册
    │   ├── handler/                 # HTTP 处理层
    │   │   ├── auth_handler.go
    │   │   ├── statements_handler.go
    │   │   ├── ...（每领域一个 handler）
    │   │   └── request_context.go   # 请求上下文提取（currentUser 等）
    │   ├── middleware/
    │   │   ├── api_v1_auth.go       # Session 鉴权
    │   │   ├── signed_url.go        # 私有文件签名校验
    │   │   └── access_log.go        # 请求日志
    │   └── dto/                     # 请求体 DTO
    │       ├── statement.go
    │       ├── asset.go
    │       └── ...
    └── infrastructure/              # 基础设施
        ├── db/mysql.go              # MySQL 连接
        ├── sessioncache/            # 缓存抽象
        │   ├── cache.go             # Cache 接口
        │   └── redis_cache.go       # Redis 实现
        ├── signedurl/signer.go      # HMAC URL 签名
        ├── urlbuilder/              # 公开 URL 构建
        └── wechat/client.go         # 微信 API 客户端
```

## 架构

### 分层架构

```
┌─────────────────────────────────────────────┐
│  handler (HTTP)                              │  ← 请求绑定、响应序列化
├─────────────────────────────────────────────┤
│  service (业务逻辑)                           │  ← 核心逻辑、协调编排
├─────────────────────────────────────────────┤
│  repository (数据访问接口)                     │  ← 接口定义（无外部依赖）
│  repository/mysql (GORM 实现)                 │  ← SQL 查询、事务
├─────────────────────────────────────────────┤
│  domain (领域实体)                            │  ← 纯结构体，无 ORM 标签
├─────────────────────────────────────────────┤
│  infrastructure (基础设施)                    │  ← DB、缓存、微信 SDK、签名
└─────────────────────────────────────────────┘
```

### 设计约定

- **依赖注入**：`bootstrap/app.go` 和 `bootstrap/modules/*.go` 手动组装依赖，不使用 DI 框架
- **接口隔离**：`repository/` 下定义接口，`repository/mysql/` 下实现，service 只依赖接口
- **领域与持久化分离**：`domain/` 的实体没有 GORM 标签，`repository/mysql/` 有独立的 model 结构体
- **DTO 层**：`http/dto/` 定义请求体结构，与领域实体彻底解耦
- **模块化**：每个业务领域（auth、statement、category 等）有独立的 handler、service、repository 模块构造函数

### 特殊实现

- **用户创建回调**：对标 Rails `UserAble.after_create :initialize_user`，新用户自动创建默认账本、资产（4 父 + 子）、收支分类（11 父 + 子），设置 `uid = id + 10000`
- **特殊账单类型**：transfer / repayment / loan_in / loan_out / reimburse / payment_proxy 自动查找对应的 `special_type` 分类
- **签名 URL**：`/private/*/statements/*` 路径的附件链接附带 HMAC 签名 + 过期时间（默认 1 小时），中间件校验通过后才返回文件
- **FlexibleAmount**：`amount` 字段同时接受 JSON 字符串和数字

## 如何加入开发

### 1. 理解代码流程（以"创建账单"为例）

```
客户端 POST /api/statements
  → router.go 路由到 statementsHandler.Create
  → handler: 绑定 DTO → 调用 service
  → service: normalizeStatementWriteInput（校验 + 特殊分类查找 + 余额计算）
  → repository: MySQL INSERT
  → service: 回查 detail → rowMapper.ToListItem → 返回 JSON
```

### 2. 新增一个业务领域

```bash
# 以"标签"为例，需要创建以下文件：
internal/domain/tag.go                          # 领域实体
internal/repository/tag_repository.go           # 接口定义
internal/repository/mysql/tag_repository.go     # MySQL 实现
internal/service/tag_service.go                 # 业务逻辑
internal/http/handler/tag_handler.go            # HTTP 处理
internal/http/dto/tag.go                        # 请求 DTO
internal/bootstrap/modules/tag_module.go        # 依赖组装

# 然后在以下文件中注册：
internal/bootstrap/app.go                       # 调用 BuildTagModule
internal/http/router/router.go                  # 注册路由
```

### 3. 代码规范

- 错误处理：repository 层定义 `var ErrXxxNotFound`，service 层用 `errors.Is` 判断
- 事务：需要事务的操作在 repository 的 MySQL 实现中用 `db.Transaction()`
- 日志：使用 Go 标准 `log` 包，关键节点打日志
- 避免 N+1：列表查询尽量用 JOIN，不用循环单条查询

### 4. API 文档

详见 [API.md](API.md)，覆盖全部 83 个接口，含请求参数和响应格式。

## 部署

```bash
# 编译二进制
go build -o jiezhang-server .

# 直接运行
./jiezhang-server

# 或配合 systemd / supervisor 守护进程
# 要求：
# - .env 文件与二进制同目录（或通过环境变量注入）
# - MySQL 可达
# - Redis 可达（可选）
# - public/ 目录存在（用于静态文件服务）
```
