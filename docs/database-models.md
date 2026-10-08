# 数据表模型

表映射统一平铺在 `internal/model/`，由 repository 返回，service 组装业务数据。认证上下文和共享业务类型放在 `internal/types/`，业务规则归对应 service，model 只保留数据表映射。

当前结构快照来自对 MySQL `information_schema` 的只读查询，包含 24 张表、256 个字段（含 Rails 元数据表）。快照不含业务记录；采集时间见其中的 `captured_at`。

`internal/model/testdata/schema_snapshot.json` 保存查询所得的字段、主键、索引和外键元数据，不包含连接凭据。模型测试使用该快照检查表和字段覆盖、SQL 类型、可空约束、默认值、主键、自增及普通/唯一索引（含顺序和前缀长度）。测试不依赖在线数据库；数据库结构变更时需要重新查询并更新快照和模型。

新增可空字段使用指针，JSON 使用 `json.RawMessage`（nil 表示 SQL NULL）。已有模型的值类型字段继续保留，兼容现有业务的零值处理；因此它们无法区分 SQL NULL 和零值。金额继续使用项目现有的 float64 表示，GORM 标签保留数据库的 decimal 精度。模型标签描述现有结构，不用于自动迁移数据库。

`user_assets` 是媒体附件表，`users_assets` 是资产分配表，分别使用 `UserAsset` 和 `UserAssetAssignment`；后者的 `asset_id` 在数据库中是 varchar。`ar_internal_metadata` 和 `schema_migrations` 使用字符串主键。

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

## 分析功能新增表

`migrations/20261006_insights.sql` 仅新增五张 InnoDB 表：项目、固定开销规则、账单分析标注、资产组合快照和商家汇总别名。现有财务表不变，服务不会自动迁移。部署新写入接口前需执行一次迁移。快照中的 `planned_migrations` 明确标记这些表的元数据来自迁移定义，尚未在线采集；原有 24 张表仍来自只读采集。部署后应重新采集并验证。

账单标注只用于归集与展示，不影响资产余额。分摊保存整数分的成员金额，资产组合快照保留每个叶子资产当时的余额及类型，历史缺失不补零。

项目参与人、起止日期及账单消费人由 `migrations/20261007_project_details.sql` 添加，仅扩展分析表，可空以兼容历史项目。项目起止日期是计划范围，不会隐藏已经归集的范围外账单。项目快捷记账在原财务事务内同时创建分析标注。

固定开销调度迁移 `migrations/20261007_fixed_cost_schedule.sql` 为 `insight_fixed_costs` 增加可空 `next_run_date`，并创建 `insight_fixed_cost_runs`。执行记录唯一键 `(fixed_cost_id, due_date)` 保存每期身份，账单删除不删除执行记录，避免定时检查恢复用户已删除的账单。规则日期推进和财务写入同事务，旧规则不默认生成账单。

项目外观迁移 `migrations/20261008_project_appearance.sql` 添加可空 `icon`（varchar(64)）和 `color`（varchar(7)）。旧项目留空并由客户端显示默认图标与绿色；编辑时省略字段保留已选外观，空字符串恢复默认。

## 日历手帐

`migrations/20261008_calendar_journal.sql` 新增 `calendar_journals`（`CalendarJournal`），唯一键为账簿、用户和日期。手帐为用户个人记录，同账簿其他成员不能读取或覆盖。日期保存为 YYYY-MM-DD 文本，避免时区转换。心情与 200 字日记独立于财务账单；零消费由用户确认，读取月度手帐时发现当天支出会持久撤销印章。收入、转账不撤销。迁移需要部署时执行，服务不会自动创建表。
