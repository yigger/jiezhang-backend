# API 接口文档

基于后端 `router.go` 和 handler 代码生成，反映实际后端接口。

## 1. 基础约定

- API Base：`{host}/api`
- 公共请求头：`Content-Type: application/json`
- 鉴权：除 `/check_openid` 外，所有接口需携带 `X-WX-Skey`（session token），由 `authMiddleware` 校验
- 鉴权失败统一返回：`{"status": 301, "msg": "session key overdue"}`

## 2. 鉴权与上传

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/check_openid` | 否 | 微信登录，Header: `X-WX-Code`，返回 `{"status": 200, "session": "..."}` |
| POST | `/upload` | 是 | 上传文件（FormData），字段：`file`(文件)、`type`、`statement_id`(选填) |

## 3. 首页

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/header` | 是 | 首页头部数据（收支趋势、预算等） |
| GET | `/index` | 是 | 首页账单列表，Query: `range`（today/yesterday/week/month/year） |
| GET | `/settings` | 是 | 用户设置，返回 `{user, version}`，`user.avatar_url` 已按 Rails avatar_path 逻辑处理 |

## 4. 用户

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/users` | 是 | 获取当前用户信息 |
| PUT | `/users/update_user` | 是 | 更新用户，Body: `{user: {theme_id, country, city, gender, language, province, bg_avatar_id, hidden_asset_money, avatar_url, nickname, bg_avatar}}`，所有字段可选 |
| POST | `/users/scan_login` | 是 | PC 扫码登录，Body: `{qr_code}` |

## 5. 账单（Statements）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/statements/categories` | 是 | 获取分类列表，Query: `type`（默认 expend） |
| GET | `/statements/assets` | 是 | 获取资产列表，Query: `type` |
| GET | `/statements/category_frequent` | 是 | 常用分类，Query: `type` |
| GET | `/statements/asset_frequent` | 是 | 常用资产，Query: `type` |
| GET | `/statements/default_category_asset` | 是 | 上次使用的分类和资产，Query: `type` |
| GET | `/statements` | 是 | 账单列表，Query: `start_date`, `end_date`, `limit`, `offset`, `category_ids`, `except_ids`, `order_by` |
| GET | `/statements/list_by_token` | 是 | 通过分享 token 查看账单，Query: `token`, `order_by` |
| GET | `/statements/:statementId` | 是 | 账单详情（含 upload_files、target_asset、can_edit 等） |
| POST | `/statements` | 是 | 创建账单，Body: `{statement: {type, amount, description, mood, category_id, asset_id, from_asset_id, to_asset_id, payee_id, target_object, location, nation, province, city, district, street, date, time}}`，amount 支持数字或字符串 |
| PUT | `/statements/:statementId` | 是 | 更新账单，Body 同上但所有字段可选（指针） |
| DELETE | `/statements/:statementId` | 是 | 删除账单 |
| DELETE | `/statements/:statementId/avatar` | 是 | 删除账单附件图片，Body: `{avatar_id}` |
| GET | `/search` | 是 | 搜索账单，Query: `keyword` |
| GET | `/statements/images` | 是 | 账单图片时间轴 |
| GET | `/statements/target_objects` | 是 | 获取目标对象列表，Query: `type` |
| POST | `/statements/generate_share_key` | 是 | 生成分享链接，Body: `{start_date, end_date, category_ids, except_statement_ids}` |
| POST | `/statements/export_check` | 是 | 导出次数检查 |
| GET | `/statements/export_excel` | 是 | 导出 Excel，Query: `range`，返回 .xlsx 文件 |

## 6. 分类（Categories）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/categories/category_list` | 是 | 分类列表（按 parent_id 展开），Query: `type`(默认 expend)、`parent_id`(默认 0) |
| GET | `/categories/parent` | 是 | 父级分类树，Query: `type`(默认 expend) |
| GET | `/categories/category_childs` | 是 | 子分类列表，Query: `parent_id`(必填) |
| GET | `/categories/category_statements` | 是 | 分类下的账单列表，Query: `category_id`(必填) |
| GET | `/categories/:id` | 是 | 分类详情 |
| POST | `/categories` | 是 | 创建分类，Body: `{category: {name, parent_id, icon_path, type}}` |
| PUT | `/categories/:id` | 是 | 更新分类，Body 同上 |
| DELETE | `/categories/:id` | 是 | 删除分类（含子分类和关联账单） |
| GET | `/icons/categories_with_url` | 否 | 获取分类图标列表 |

## 7. 资产（Assets）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/assets` | 是 | 资产列表/树，Query: `parent_id`(选填) |
| GET | `/assets/:id` | 是 | 资产详情 |
| POST | `/assets` | 是 | 创建资产，Body: `{wallet: {name, amount, parent_id, icon_path, remark, type}}` |
| PUT | `/assets/:id` | 是 | 更新资产，Body 同上 |
| DELETE | `/assets/:id` | 是 | 删除资产（含子资产和关联账单） |
| PUT | `/wallet/surplus` | 是 | 更新资产余额，Body: `{asset_id, amount}`，amount 支持数字或字符串 |
| GET | `/icons/assets_with_url` | 否 | 获取资产图标列表 |

## 8. 账本（Account Books）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/account_books` | 是 | 用户可访问的账本列表 |
| GET | `/account_books/:id` | 是 | 账本详情 |
| GET | `/account_books/types` | 是 | 账本类型列表 |
| GET | `/account_books/preset_categories` | 是 | 预设分类和资产模板，Query: `account_type` |
| POST | `/account_books` | 是 | 创建账本，Body: `{name, description, account_type, categories, assets}` |
| PUT | `/account_books/:id` | 是 | 更新账本，Body: `{name, description, account_type}` |
| PUT | `/account_books/:id/switch` | 是 | 切换默认账本 |
| DELETE | `/account_books/:id` | 是 | 删除账本 |

## 9. 财务/钱包（Finance）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/wallet` | 是 | 钱包总览 |
| GET | `/wallet/information` | 是 | 资产详情，Query: `asset_id`(必填) |
| GET | `/wallet/time_line` | 是 | 资产时间线，Query: `asset_id`(必填) |
| GET | `/wallet/statement_list` | 是 | 资产关联账单，Query: `asset_id`, `year`, `month`(均必填) |

## 10. 预算（Budgets）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/budgets` | 是 | 预算总览，Query: `year`, `month` |
| GET | `/budgets/parent` | 是 | 父级分类预算列表，Query: `year`, `month` |
| GET | `/budgets/:categoryId` | 是 | 分类预算详情，Query: `year`, `month` |
| PUT | `/budgets/0` | 是 | 更新预算，Body: `{type, amount, category_id}`，type="user" 为总预算，"category" 为分类预算，amount 支持数字或字符串，category_id 可选 |

## 11. 统计图表（Chart）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/chart/calendar_data` | 是 | 日历热力图，Query: `date`(YYYY-MM) |
| GET | `/chart/overview_header` | 是 | 概览头部数据，Query: `date`(YYYY-MM) |
| GET | `/chart/overview_statements` | 是 | 概览账单列表，Query: `date`, `type` |
| GET | `/chart/rate` | 是 | 收支占比，Query: `date`, `type` |

## 12. 超级流水（Super Statements）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/super_statements/time` | 是 | 月度汇总时间线，Query: `year`, `month`, `asset`, `asset_id`, `category_id`, `order_by` |
| GET | `/super_statements/list` | 是 | 流水列表，Query 同上 |

## 13. 超级图表（Super Chart）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/super_chart/header` | 是 | 图表头部数据，Query: `year`, `month` |
| GET | `/super_chart/get_pie_data` | 是 | 饼图数据，Query: `year`, `month`, `statement_type` |
| GET | `/super_chart/week_data` | 是 | 周数据，Query: `year`, `month` |
| GET | `/super_chart/line_chart` | 是 | 折线图数据，Query: `year`, `month` |
| GET | `/super_chart/categories_list` | 是 | 分类 Top 排行，Query: `year`, `month` |
| GET | `/super_chart/table_sumary` | 是 | 日汇总表，Query: `year`, `month` |

## 14. 消息（Messages）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/message` | 是 | 消息列表 |
| GET | `/message/:id` | 是 | 消息详情 |

## 15. 收款方（Payees）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/payees` | 是 | 收款方列表 |
| POST | `/payees` | 是 | 创建收款方，Body: `{name}` 或 `{payee: {name}}` |
| PUT | `/payees/:id` | 是 | 更新收款方 |
| DELETE | `/payees/:id` | 是 | 删除收款方 |

## 16. 好友协作（Friends）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/friends` | 是 | 协作者列表 |
| POST | `/friends/invite` | 是 | 邀请，Body: `{account_book_id, role}` |
| GET | `/friends/invite_information` | 否 | 查看邀请信息，Query: `invite_token` |
| POST | `/friends/accept_apply` | 是 | 接受邀请，Body: `{invite_token, nickname}` |
| PUT | `/friends/:collaboratorId` | 是 | 更新协作者，Body: `{account_book_id, role?, remark?}` |
| DELETE | `/friends/:collaboratorId` | 是 | 移除协作者 |

## 17. 反馈

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/settings/feedback` | 是 | 提交反馈，Body: `{content, type}` |

## 18. 与旧版 API.md 的差异

| 变更 | 说明 |
|------|------|
| **新增** `GET /categories/parent` | 父级分类树，前端已有调用 |
| **新增** `GET /categories/category_childs` | 按 parent_id 获取子分类 |
| **新增** `GET /categories/category_statements` | 分类下账单列表 |
| **补充** `PUT /assets/:id` | 旧文档未列出更新资产接口 |
| **补充** `PUT /account_books/:id` | 旧文档遗漏 |
| **补充** `PUT /account_books/:id/switch` | 旧文档遗漏 |
| **补充** `PUT /friends/:collaboratorId` | 旧文档遗漏 |
| **补充** BODY 详情 | 所有接口补充了完整的请求体结构 |
| **修正** amount 字段 | 多处 amount 支持数字或字符串（FlexibleAmount） |
