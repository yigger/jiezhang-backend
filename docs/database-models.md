# 数据表模型

表映射统一平铺在 `internal/model/`，由 repository 返回，service 组装业务数据。认证上下文和共享业务类型放在 `internal/types/`，业务规则归对应 service，model 只保留数据表映射。

当前结构快照来自对 MySQL `information_schema` 的只读查询，包含 24 张表、256 个字段（含 Rails 元数据表）。快照不含业务记录；采集时间见其中的 `captured_at`。

`internal/model/testdata/schema_snapshot.json` 保存查询所得的字段、主键、索引和外键元数据，不包含连接凭据。模型测试使用该快照检查表和字段覆盖、SQL 类型、可空约束、默认值、主键、自增及普通/唯一索引（含顺序和前缀长度）。测试不依赖在线数据库；数据库结构变更时需要重新查询并更新快照和模型。

新增可空字段使用指针，JSON 使用 `json.RawMessage`（nil 表示 SQL NULL）。已有模型的值类型字段继续保留，兼容现有业务的零值处理；因此它们无法区分 SQL NULL 和零值。金额继续使用项目现有的 float64 表示，GORM 标签保留数据库的 decimal 精度。模型标签描述现有结构，不用于自动迁移数据库。

`user_assets` 是媒体附件表，`users_assets` 是资产分配表，分别使用 `UserAsset` 和 `UserAssetAssignment`；后者的 `asset_id` 在数据库中是 varchar。`ar_internal_metadata` 和 `schema_migrations` 使用字符串主键。

全部表结构统一维护在 `migrations/`。[`00000000_initial_schema.sql`](../migrations/00000000_initial_schema.sql) 是最早的结构基线，包含原有 24 张核心表的显式 DDL；按文件名顺序执行其余增量迁移后，得到当前完整的 31 张表。基线只用于空库，不包含凭据、用户或交易记录，也不从元数据快照初始化。已有数据库只执行尚未应用的增量迁移。

Docker 的 [`01-migrations.sql`](../docker/mysql/01-migrations.sql) 从基线开始执行上述迁移，再由独立的 [`02-dev-seed.sql`](../docker/mysql/02-dev-seed.sql) 添加合成的开发用户、账本、分类和零余额资产。这些脚本只在新建的 Compose MySQL 数据卷初始化时执行，不属于运行时 AutoMigrate。新增迁移应同步更新 Docker 初始化清单；已有卷仍需手动执行迁移。

| 数据表 | Go 模型 |
| --- | --- |
| `account_books` | [`AccountBook`](../internal/model/account_book.go) |
| `account_book_collaborators` | [`AccountBookCollaborator`](../internal/model/account_book_collaborator.go) |
| `account_book_friend_refs` | [`AccountBookFriendRef`](../internal/model/account_book_friend_ref.go) |
| `ar_internal_metadata` | [`ARInternalMetadata`](../internal/model/ar_internal_metadata.go) |
| `assets` | [`Asset`](../internal/model/asset.go) |
| `asset_logs` | [`AssetLog`](../internal/model/asset_log.go) |
| `asset_snapshots` | [`AssetSnapshot`](../internal/model/asset_snapshot.go) |
| `bonus_points_logs` | [`BonusPointsLog`](../internal/model/bonus_points_log.go) |
| `categories` | [`Category`](../internal/model/category.go) |
| `error_logs` | [`ErrorLog`](../internal/model/error_log.go) |
| `feedbacks` | [`Feedback`](../internal/model/feedback.go) |
| `friends` | [`Friend`](../internal/model/friend.go) |
| `friend_applies` | [`FriendApply`](../internal/model/friend_apply.go) |
| `messages` | [`Message`](../internal/model/message.go) |
| `month_charts` | [`MonthChart`](../internal/model/month_chart.go) |
| `operate_logs` | [`OperateLog`](../internal/model/operate_log.go) |
| `payees` | [`Payee`](../internal/model/payee.go) |
| `pre_orders` | [`PreOrder`](../internal/model/pre_order.go) |
| `recommends` | [`Recommend`](../internal/model/recommend.go) |
| `schema_migrations` | [`SchemaMigration`](../internal/model/schema_migration.go) |
| `statements` | [`Statement`](../internal/model/statement.go) |
| `users` | [`User`](../internal/model/user.go) |
| `user_assets` | [`UserAsset`](../internal/model/user_asset.go) |
| `users_assets` | [`UserAssetAssignment`](../internal/model/user_asset_assignment.go) |

## 历史补录来源与余额行为

[`migrations/20261011_statement_import.sql`](../migrations/20261011_statement_import.sql) 给 `statements` 添加可空 `balance_applied`（`*bool`），NULL 兼容普通账单，false 表示历史补录，创建、更新和删除均不改变资产余额。补录文件没有钱包历史余额，现有 `residue` 保持零值，不代表真实历史余额；原有金额 float64 和 nullable/零值兼容保持不变。

导入不新增表。现有 `statements` 增加可空 `import_key`、`import_fingerprint`（均为 `*string`、ascii_bin 的 char(64)）：前者是平台和原订单号的 SHA-256，后者是原始流水内容的 SHA-256（时间、整数分金额、交易对方等，独立于资产/分类映射）。唯一索引 `idx_statements_import_key` 为 `(account_book_id, user_id, import_key)`；普通账单保留 NULL，可重复存在。两个指纹和账单、自动创建的分类/资产同事务提交，编辑时保留原值以识别重导及来源内容冲突；删除账单会删除指纹，再次导入允许重新创建。时间、金额、交易对方相同的不同订单不会被误合并。元数据快照按迁移定义增加覆盖，并标记为 planned migration，尚未在线采集验证；服务不执行 AutoMigrate。

## 分析功能新增表

`migrations/20261006_insights.sql` 仅新增五张 InnoDB 表：项目、固定开销规则、账单分析标注、资产组合快照和商家汇总别名。现有财务表不变，服务不会自动迁移。部署新写入接口前需执行一次迁移。快照中的 `planned_migrations` 明确标记这些表的元数据来自迁移定义，尚未在线采集；原有 24 张表仍来自只读采集。部署后应重新采集并验证。

账单标注只用于归集与展示，不影响资产余额。分摊保存整数分的成员金额，资产组合快照保留每个叶子资产当时的余额及类型，历史缺失不补零。

项目参与人、起止日期及账单消费人由 `migrations/20261007_project_details.sql` 添加，仅扩展分析表，可空以兼容历史项目。项目起止日期是计划范围，不会隐藏已经归集的范围外账单。项目快捷记账在原财务事务内同时创建分析标注。

固定开销调度迁移 `migrations/20261007_fixed_cost_schedule.sql` 为 `insight_fixed_costs` 增加可空 `next_run_date`，并创建 `insight_fixed_cost_runs`。执行记录唯一键 `(fixed_cost_id, due_date)` 保存每期身份，账单删除不删除执行记录，避免定时检查恢复用户已删除的账单。规则日期推进和财务写入同事务，旧规则不默认生成账单。

项目外观迁移 `migrations/20261008_project_appearance.sql` 添加可空 `icon`（varchar(64)）和 `color`（varchar(7)）。旧项目留空并由客户端显示默认图标与绿色；编辑时省略字段保留已选外观，空字符串恢复默认。

## 日历手帐

`migrations/20261008_calendar_journal.sql` 新增 `calendar_journals`（`CalendarJournal`），唯一键为账簿、用户和日期。手帐为用户个人记录，同账簿其他成员不能读取或覆盖。日期保存为 YYYY-MM-DD 文本，避免时区转换。心情与 200 字日记独立于财务账单；零消费由用户确认，读取月度手帐时发现当天支出会持久撤销印章。收入、转账不撤销。迁移需要部署时执行，服务不会自动创建表。

`users` 的用户名称使用 `nickname`，不映射 `name` 列。依据线上 1054 错误确认的结构，移除旧快照中多出的 `users.name`，避免登录会话更新及新用户创建写入不存在的列；此修正无需数据库迁移。
