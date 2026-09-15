# 项目修改约定

## 入口与结构

- 唯一入口是根目录 `main.go`。启动、资源关闭、依赖组装和路由配置集中在 `internal/bootstrap/app.go`，组装函数为 `setupRoutes`。
- 使用显式构造函数和具名变量表达依赖，不使用 dig、运行时 DI 容器或全局业务服务注册表。不再创建 `cmd/server`、`bootstrap/modules` 或单独的 setup/wiring 文件。
- 当前路径约定优先于外部架构 skill 中旧的 handler/repository/DTO 路径；分层隔离原则仍然适用。

## 分层与类型归属

依赖方向：`controller → service/<业务> → repo 接口`；`repo/mysql` 实现接口。

- `internal/controller` 按业务文件平铺，只处理传输参数、调用 Service 和映射错误，不访问 Repo、ORM 或 SQL。
- `internal/service/<业务>` 负责业务校验、权限、计算及用例编排。通过接口访问仓储和外部能力，不导入 Gin、GORM、SQL、Controller、配置包或 Infrastructure 实现。
- `internal/repo` 平铺仓储接口、复杂查询条件和 SQL 聚合结果；简单查询直接传参数，避免为少量必填参数新增 Filter。
- `internal/repo/mysql` 执行 GORM/SQL、行锁和持久化事务，不依赖 Service、Controller 或 Types，不组装 HTTP 响应。
- `internal/model` 只放统一的数据表映射，按表平铺，供 Repo 和 Service 复用。不要恢复 model/domain，也不要新增同一表的业务专用 Row/Record 副本。普通查询返回表模型，SQL 聚合结果另放 Repo。
- `internal/types` 平铺 HTTP 请求/响应、认证上下文和共享业务类型；仅单个用例使用的内部类型可留在对应 Service。业务规则放 Service。
- Infrastructure 实现外部能力，通过启动代码注入。共享组装逻辑放 Service 的合适位置，不为复用格式化代码而调用另一个 Service。
- 不引入无实际用途的 BaseService、通用仓储或额外目录层级。

## 必须保持的行为

- 保留现有接口方法、路径、字段和兼容错误格式，除非任务明确要求变更；HTTP 200 可能承载业务失败。
- 账单和资产余额必须原子提交或回滚。更新、删除在事务内锁定账单后检查归属，部分更新在锁内合并；余额规则由 Service 计算，Repo 应用增减值，响应查询在提交后执行。
- 不弱化账本权限、Session 校验或文件 URL 签名检查。`ENV=dev` 使用用户 ID 1 的开发鉴权，`GIN_MODE` 不控制该行为。
- 表模型变更以真实 schema 为依据，不凭命名猜测字段类型。新增可空字段保留 NULL；已有零值兼容和 float64 金额表示的变动需同时审查业务影响。
- 服务不执行 AutoMigrate。不要把 `internal/model/testdata/schema_snapshot.json` 当作迁移或初始化脚本；它仅包含 schema 元数据，禁止写入凭据或业务记录。

## 验证与文档

- 修改 Go 代码后运行 gofmt、相关测试和 `go test ./...`。不要删除或放宽分层、路由、模型测试来绕过失败。
- Swagger 基础注解在 `main.go`，接口注解在 Controller。接口变更后运行 `make swagger`，提交 `docs/swagger` 生成文件，不手改生成结果。
- 有意改变 API 方法/路径时，同时审查 `internal/router/testdata/routes.golden`。对交付改动运行 `make check`（Swagger、测试、vet、编译）和 `git diff --check`。
- schema 变更同步更新模型、元数据快照及 `internal/model/schema_test.go` 中的覆盖清单。SQL mock 和快照测试不等于在线数据库集成测试，报告验证范围时明确区分。
- 启动、配置或结构变化同步更新 README；表映射约定更新 `docs/database-models.md`。不维护重复的手写 API 说明或阶段性重构记录。
- 不打印或提交 `.env`、密钥、数据库连接凭据、业务数据。保留工作区已有修改，不顺手重置无关文件。
