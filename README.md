# jiezhang-backend

基于 Go、Gin、GORM 和 MySQL 的记账小程序后端，支持账本、收支记录、资产、预算、统计，以及支付宝和微信历史账单导入。

## 快速开始

安装 Git 和 Docker（含 Compose）。Windows / macOS 推荐使用 Docker Desktop，Windows 需启用 Linux 容器。

```bash
git clone https://github.com/yigger/jiezhang-backend.git
cd jiezhang-backend
docker compose up -d
```

首次启动会下载依赖、初始化 MySQL 和 Redis，并创建开发用户与账本（ID 均为 `1`），无需配置 `.env`。首次编译需要一些时间，可通过下面的日志命令查看进度。

- API：<http://localhost:10240>
- Swagger：<http://localhost:10240/swagger/index.html>

修改 Go 源码后，Air 会自动编译并重启服务，无需重新构建镜像。

小程序真实登录需要在根目录 `.env` 配置对应的 `MINIPROGRAM_APPID` 和 `MINIPROGRAM_SECRET`。修改后执行 `docker compose up -d --force-recreate api`，普通 `restart` 不会更新容器环境变量。开发鉴权不代替 `/api/check_openid` 的微信登录流程。

## 常用命令

```bash
docker compose ps                     # 查看运行状态
docker compose logs -f api            # 查看服务日志
docker compose restart api           # 重启服务
docker compose up -d --build api      # 修改 Dockerfile 后重新构建
docker compose down                  # 停止环境，保留数据
```

MySQL 默认映射到 `127.0.0.1:13306`，Redis 映射到 `127.0.0.1:16379`。可用 `DEV_HTTP_PORT`、`DEV_MYSQL_PORT`、`DEV_REDIS_PORT` 调整宿主机端口。数据库密码首次启动时随机生成，保存在 Docker 数据卷中。

## 本地运行（可选）

需要 Go **1.26.3+** 和 MySQL，Redis 可选。所有命令在项目根目录执行。

```bash
cp .env.example .env
```

编辑 `.env`，填写 `MYSQL_DSN`、`MINIPROGRAM_APPID`、`MINIPROGRAM_SECRET` 和 `SESSION_TOKEN_SECRET`。不使用 Redis 时，将 `REDIS_URL` 留空。完整配置见 [`.env.example`](.env.example)，进程环境变量优先于 `.env`。

创建并选择数据库后，按文件名升序执行 [migrations](migrations) 下的 SQL，再启动服务：

```bash
go mod download
go run .
```

服务不会自动建表或迁移。`00000000_initial_schema.sql` 仅用于空库；已有数据库只执行尚未应用的增量迁移。Compose 会在首次初始化数据卷时自动完成这些步骤。

开发模式 `ENV=dev` 使用用户 ID `1`，仍需该用户具有账本权限。对外部署应设置 `ENV=production`、`GIN_MODE=release`，并配置真实的小程序凭据和 `PUBLIC_BASE_URL`。

## 历史账单导入

需要宿主机安装 Go，并配置目标数据库的 `MYSQL_DSN`。支持已解压的支付宝 CSV / XLSX 和微信 XLSX。

<details>
<summary>宿主机连接 Compose MySQL（PowerShell）</summary>

在当前终端执行，使用 Compose 的开发密码和映射端口：

```powershell
$billImportPassword = (docker compose exec -T mysql cat /run/dev-secrets/mysql-password).Trim()
$billImportPort = ((docker compose port mysql 3306) -split ':')[-1].Trim()
$env:MYSQL_DSN = "jiezhang:${billImportPassword}@tcp(127.0.0.1:${billImportPort})/jiezhang?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
```

</details>

```bash
go run ./cmd/bill-import --file "./tmp/支付宝.csv" --file "./tmp/微信.xlsx" --user-id 1 --book-id 1
```

替换为实际文件路径和有权限的用户、账本 ID。默认仅预览，确认后添加 `--apply` 写入。历史补录仅处理普通收入和支出，不改变现有资产余额。自定义分类和资产映射见 [mapping.example.json](cmd/bill-import/mapping.example.json)，通过 `--mapping` 指定。

## 开发与文档

安装 Go 和 Make 后，可运行：

```bash
make test       # 运行测试
make swagger    # 生成接口文档
make build      # 编译服务和导入工具到 bin/
make check      # Swagger、测试、vet 和编译
```

仅使用 Docker 时，也可在容器内运行测试：

```bash
docker compose exec api /usr/local/bin/dev-entrypoint.sh go test ./...
```

- [项目修改约定](AGENTS.md)
- [数据库模型](docs/database-models.md)
- [接口文档](docs/swagger/swagger.yaml)
