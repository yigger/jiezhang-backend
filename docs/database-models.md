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
