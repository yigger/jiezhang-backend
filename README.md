# jiezhang-backend

基于 Go、Gin、GORM 和 MySQL 的记账小程序后端，提供账本、账单、资产、分类、预算、统计、分享、上传及 Excel 导出接口。Redis 用于缓存，微信小程序登录通过 `jscode2session` 完成。

## 本地启动

需要 Go 1.26.3 或更高版本、可访问的 MySQL，以及微信小程序配置。Redis 可选。

```bash
cp .env.example .env
# 编辑 .env，替换数据库连接、小程序配置和会话密钥
# 不使用 Redis 时，将 REDIS_URL 留空
go mod download
go run .
```

在项目根目录运行。默认地址为 `http://localhost:10240`，接口文档为 [Swagger UI](http://localhost:10240/swagger/)，原始描述为 `/swagger/doc.json`。

服务连接已有数据库，**不会自动建表或执行迁移**。分析功能新增的五张表需在部署前执行一次 [`migrations/20261006_insights.sql`](migrations/20261006_insights.sql)，再执行 [`migrations/20261007_project_details.sql`](migrations/20261007_project_details.sql) 补项目参与人、时间范围和消费人字段，再执行 [`migrations/20261007_fixed_cost_schedule.sql`](migrations/20261007_fixed_cost_schedule.sql) 添加固定开销调度字段和执行记录，再执行 [`migrations/20261008_project_appearance.sql`](migrations/20261008_project_appearance.sql) 添加项目图标与颜色；这些迁移仅创建或扩展分析表，保留现有账单与资产数据。仓库未提供完整的建库与初始化数据脚本；首次接手需取得测试数据库及所需初始数据。表模型清单与字段兼容说明见 [数据库模型](docs/database-models.md)。

## 配置

配置从当前工作目录的 `.env` 加载，已设置的进程环境变量优先。不要提交真实凭据。

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `MYSQL_DSN` | 无，必填 | MySQL DSN；参考 `.env.example`，保留时间解析配置 |
| `MINIPROGRAM_APPID` | 无，必填 | 微信小程序 AppID |
| `MINIPROGRAM_SECRET` | 无，必填 | 微信小程序 Secret |
| `SESSION_TOKEN_SECRET` | 无，必填 | Token 和 URL 签名使用的密钥 |
| `ENV` | `dev` | `dev` 使用开发鉴权；部署时设为 `production` |
| `GIN_MODE` | `debug` | 部署时设为 `release`；此配置不会关闭开发鉴权 |
| `PORT` | `10240` | HTTP 监听端口 |
| `PUBLIC_BASE_URL` | `http://localhost:<PORT>` | 返回给客户端的文件链接前缀 |
| `REDIS_URL` | 空 | 留空使用内存缓存；连接失败也会记录日志并回退内存缓存 |

`ENV=dev` 时，鉴权直接读取用户 ID 1，跳过 AppID 和 Session 校验；测试库需要该用户及其可访问的账本。开发模式仍会校验账本权限。**对外部署必须显式设置 `ENV=production`**，仅修改 `GIN_MODE` 不够。

非开发模式通过 `X-WX-APP-ID` 和 `X-WX-Skey` 鉴权。登录接口使用 `X-WX-Code`，具体参数见 Swagger。账本从 `account_book_id` 查询参数或用户默认账本获取。部分兼容接口用 HTTP 200 返回业务失败，客户端还需检查响应中的 `status`、`msg` 或 `message`。

内存缓存不跨实例共享，进程重启会丢失缓存内容；多实例部署应配置共享 Redis，并检查启动日志是否发生回退。配置结构中保留的 `APP_NAME`、`MCP_API_KEY` 当前不控制已注册 HTTP 接口；MCP 适配器尚未接入启动路由。

## 代码导航

```text
main.go                     唯一启动入口、Swagger 基础注解
internal/
  bootstrap/app.go          配置校验、资源生命周期、setupRoutes 显式组装依赖
  config/                   环境配置读取
  controller/               HTTP 参数绑定、调用 service、错误映射、接口注解
  service/<业务>/           业务规则、权限校验、用例编排
  repo/                     仓储接口、复杂查询条件、SQL 聚合结果
  repo/mysql/               GORM/SQL 实现和事务
  model/                    全部数据表映射，按表平铺
  types/                    请求、响应、认证上下文及共享业务类型
  router/                   路由、Swagger、路由契约测试
  middleware/               HTTP 鉴权适配、日志、文件 URL 签名检查
  infrastructure/           数据库连接、缓存、微信、文件、Token 和 Excel 实现
  mcp/                      MCP 适配器；当前没有注册 HTTP 入口
```

依赖方向是 `controller → service → repo 接口`，`repo/mysql` 实现接口。所有依赖通过构造函数显式传入，在 `bootstrap/app.go` 的 `setupRoutes` 中组装，不使用运行时 DI 容器。Controller、Repo、Model、Types 按文件平铺，Service 按业务拆包。

仓储返回统一表模型；Service 读取关联数据并组装响应。共享账单组装逻辑在 `service/helper/statement_assembler.go`，响应映射在 `service/statement/mapper.go`。不要为同一张表再定义业务专用的 Row/Record 副本。

账单写入由 `service/statement.Writer` 发起事务。更新、删除先锁定当前账单并校验归属；更新在锁内合并部分字段。Service 计算余额增减值，MySQL 实现在同一事务内写入账单和资产余额，查询响应在提交后进行。

## 开发与验证

```bash
make test       # go test ./...
make swagger    # 从 main.go、Controller 注解生成 docs/swagger
make vet        # go vet ./...
make build      # 本机编译，输出 bin/jiezhang-server
make build-server # Linux amd64，输出 bin/linux-amd64/jiezhang-backend
make check      # 生成 Swagger、全量测试、vet、编译
```

新增业务时，按需要补充表模型、Repo 接口与 MySQL 实现、Service、Types 和 Controller，再在 `setupRoutes` 注入依赖、在 Router 注册路由。详细修改约定见 [AGENTS.md](AGENTS.md)。

接口文档以 Controller 注解及生成的 [Swagger](docs/swagger/swagger.yaml) 为准。不要手改生成文件；修改接口后运行 `make swagger` 并提交产物。有意变更方法或路径时，同时审查 `internal/router/testdata/routes.golden`。

现有检查包括路由与 Swagger 一致性、分层依赖、数据库结构快照、权限、余额规则、事务提交/回滚、分享 Token、文件和 Excel。分层检查位于 `internal/repo/dependencies_test.go`。测试使用 SQL mock 和结构快照，不会连接真实 MySQL，也不能代替数据库集成测试。涉及写入链路时，仍需在测试数据库回归账单创建、更新、删除及对应余额变化。

## 运行与部署

```bash
make build
# 在配置所在的工作目录启动，也可以通过进程环境变量注入配置
./bin/jiezhang-server
```

在 macOS 上为 Linux 服务器编译时，使用服务器构建目标（关闭 CGO）：

```bash
make build-server                   # 服务器 uname -m 返回 x86_64
make build-server SERVER_ARCH=arm64  # 服务器 uname -m 返回 aarch64
```

将对应的 `bin/linux-<架构>/jiezhang-backend` 上传到服务器，在配置所在目录执行 `./jiezhang-backend`。本机与服务器产物分目录保存，避免上传 macOS 可执行文件。

设置 `ENV=production`、`GIN_MODE=release` 和客户端可访问的 `PUBLIC_BASE_URL`。服务从当前工作目录读取 `.env`，文件存储也相对于当前工作目录；启动会创建 `public/`，运行账号需要写权限。部署时保留并持久化上传文件，不要把工作目录切换到临时目录。

应用收到 SIGINT/SIGTERM 后执行 HTTP 优雅关闭并释放缓存与数据库连接。首次上线或变更持久化逻辑前，在测试环境验证业务链路；`make check` 不包含部署和真实数据库回归。

分析功能的 MySQL 集成测试只连接显式指定的临时本地实例，使用 `/tmp/jiezhang-insights-db.*` 下的 socket，并创建和清理独立测试 schema：

```bash
INSIGHTS_TEST_MYSQL_SOCKET=/tmp/jiezhang-insights-db.example/mysql.sock go test ./internal/service/insights -run TestMySQLInsightLifecycle -count=1
```

未提供该 socket 时测试跳过；普通 `make check` 不连接业务数据库。

### 固定开销自动记账

服务运行时启动 cron（`CRON_TZ=Asia/Shanghai 55 23 * * *`）：每晚 23:55 检查已经确认启用的固定开销，将到期支出从所选钱包入账。首次日期按保存时的当月指定日计算；如果当月日期已过，向后移动规则周期（月）；以后每期都从计划月份计算，31 日在短月份取月底，下一期恢复 31 日。暂停后不生成账单，再次启用从当前日期重新排期，不补暂停期间。

重启时补查已经结束的日期（23:55 前只补到昨天），遗漏期数逐期处理。账单、钱包余额、固定开销标注、执行记录与下一次日期使用同一个财务事务；账簿行锁及 `(fixed_cost_id, due_date)` 唯一约束保护重复检查和多实例执行。同一规则同一计划月份只生成一次，修改日期或删除已生成账单也不会重复生成当期。分类、钱包或创建人权限失效时回滚该期，日志报告错误，其他规则继续处理。现有预算规则的 `next_run_date` 留空，用户重新确认分类和钱包后才开始自动记账。必须先执行迁移再启动新版服务；服务不会自动迁移数据库。

日历手帐上线前执行 `migrations/20261008_calendar_journal.sql`。手帐按用户、账簿、日期隔离；清空内容可重复保存，不修改财务数据。元数据快照中的新表按迁移定义记录，尚未在线验证。
