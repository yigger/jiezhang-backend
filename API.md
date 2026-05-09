# API 接口清单（前端项目视角）

本文档基于当前前端代码中的实际调用整理，覆盖 `src/api` 与业务页面中使用到的后端接口。

## 1. 基础约定

- API Base URL：`{host}/api`
- `host` 来源：
  - 开发：`src/config/config.ts` 的 `dev_host`
  - 生产：`src/config/config.ts` 的 `production_host`
- 公共请求头：
  - `content-type: application/json`
  - `X-WX-APP-ID: {appid}`
  - `X-WX-Skey: {session}`（鉴权后注入）

## 2. 鉴权与上传下载

| 方法 | 路径 | 用途 | 请求参数 |
|---|---|---|---|
| POST | `/check_openid` | 通过微信登录 code 换取 session | Header: `X-WX-Code`, `X-WX-APP-ID` |
| POST | `/upload` | 上传文件（头像/账单图片） | FormData + 文件字段名 `file` |
| GET | `/statements/export_excel` | 导出账单 Excel | Query: `range` |

补充说明：

- `session` 本地缓存键：`access_token_data`，有效期 2 小时。
- 若业务响应 `data.status === 301`，前端会清空 token 并自动重试（最多 5 次）。

## 3. 首页与聚合

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/header` | `api.main.header()` | 无 |
| GET | `/index` | `api.main.statements(range)` | Query: `range` |

## 4. 用户

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/settings` | `api.users.getSettingsData()` | 无 |
| GET | `/users` | `api.users.getUserInfo()` | 无 |
| PUT | `/users/update_user` | `api.users.updateUserInfo(params)` | Body: `{ user: params }` |
| POST | `/users/scan_login` | `api.users.loginPc(code)` | Body: `{ qr_code }` |

## 5. 账单（Statements）

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/statements/categories` | `api.statements.categoriesWithForm(type)` | Query: `type` |
| GET | `/statements/assets` | `api.statements.assetsWithForm(params)` | Query: 可选 |
| GET | `/statements/category_frequent` | `api.statements.categoryFrequent(type)` | Query: `type` |
| GET | `/statements/asset_frequent` | `api.statements.assetFrequent()` | 无 |
| GET | `/statements` | `api.statements.list(params)` | Query: 列表筛选参数 |
| GET | `/statements/list_by_token` | `api.statements.getListByToken(token, orderBy)` | Query: `token`, `order_by` |
| POST | `/statements` | `api.statements.create(data)` | Body: `{ statement: data }` |
| PUT | `/statements/:statementId` | `api.statements.update(statementId, data)` | Body: `{ statement: data }` |
| GET | `/statements/:statementId` | `api.statements.getStatement(statementId)` | Path: `statementId` |
| DELETE | `/statements/:statementId` | `api.statements.deleteStatement(statementId)` | Path: `statementId` |
| GET | `/search` | `api.statements.searchStatements(keyword)` | Query: `keyword` |
| GET | `/statements/images` | `api.statements.getStatementImages()` | 无 |
| POST | `/statements/generate_share_key` | `api.statements.generateShareToken(params)` | Body: 分享参数 |
| POST | `/statements/export_check` | `api.statements.pre_check_export()` | Body: `{}` |
| GET | `/statements/target_objects` | `api.statements.targetObjects(statementType)` | Query: `type` |
| DELETE | `/statements/:statementId/avatar` | `api.statements.removeAvatar(statementId, avatarId)` | Body: `{ avatar_id }` |
| GET | `/statements/default_category_asset` | `api.statements.defaultCategoryAsset(statementType)` | Query: `type` |

## 6. 分类（Categories）

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/categories/category_list` | `api.categories.getSettingList(...)` | Query: `type`, `parent_id` |
| GET | `/categories/:id` | `api.categories.getCategoryDetail(id)` | Path: `id` |
| DELETE | `/categories/:id` | `api.categories.deleteCategory(id)` | Path: `id` |
| GET | `/icons/categories_with_url` | `api.categories.getCategoryIcon()` | 无 |
| PUT | `/categories/:id` | `api.categories.updateCategory(id, data)` | Body: 业务对象 |
| POST | `/categories` | `api.categories.create(data)` | Body: 业务对象 |

## 7. 资产（Assets）

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/assets` | `api.assets.getSettingList({ parentId })` | Query: `parent_id` |
| GET | `/assets/:id` | `api.assets.getAssetDetail(id)` | Path: `id` |
| DELETE | `/assets/:id` | `api.assets.deleteAsset(id)` | Path: `id` |
| GET | `/icons/assets_with_url` | `api.assets.getAssetIcon()` | 无 |
| PUT | `/assets/:id` | `api.assets.updateAsset(id, data)` | Body: `{ wallet: data }` |
| POST | `/assets` | `api.assets.create(data)` | Body: `{ wallet: data }` |
| PUT | `/wallet/surplus` | `api.assets.updateAssetAmount(id, amount)` | Body: `{ asset_id, amount }` |

## 8. 账本（Account Books）

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/account_books` | `api.account_books.getAccountBooks()` | 无 |
| GET | `/account_books/:id` | `api.account_books.getAccountBook(id)` | Path: `id` |
| GET | `/account_books/types` | `api.account_books.getAccountBookTypes()` | 无 |
| GET | `/account_books/preset_categories` | `api.account_books.getCategoriesList({ accountType })` | Query: `account_type` |
| PUT | `/account_books/:id/switch` | `api.account_books.updateDefaultAccount(accountBook)` | Path: `accountBook.id` |
| POST | `/account_books` | `api.account_books.create(data)` | Body: 账本对象 |
| PUT | `/account_books/:id` | `api.account_books.update(id, data)` | Body: 账本对象 |
| DELETE | `/account_books/:id` | `api.account_books.destroy(id)` | Path: `id` |

## 9. 财务（Finance）

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/wallet` | `api.finances.index()` | 无 |
| GET | `/wallet/information` | `api.finances.getAssetDetail(assetId)` | Query: `asset_id` |
| GET | `/wallet/time_line` | `api.finances.getAssetTimeline(assetId)` | Query: `asset_id` |
| GET | `/wallet/statement_list` | `api.finances.getAssetStatements(params)` | Query: `asset_id`, `year`, `month` |
| PUT | `/users/update_user` | `api.finances.updateAmountVisible({ visible })` | Body: `{ user: { hidden_asset_money } }` |

## 10. 预算（Budgets）

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/budgets` | `api.budgets.getSummary({ year, month })` | Query: `year`, `month` |
| GET | `/budgets/parent` | `api.budgets.getParentList({ year, month })` | Query: `year`, `month` |
| GET | `/budgets/:categoryId` | `api.budgets.getCategoryBudget(...)` | Path: `category_id`; Query: `year`, `month` |
| PUT | `/budgets/0` | `api.budgets.updateRootAmount({ amount })` | Body: `{ type: 'user', amount }` |
| PUT | `/budgets/0` | `api.budgets.updateCategoryAmount(...)` | Body: `{ type: 'category', category_id, amount }` |

## 11. 统计（Statistic）

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/chart/calendar_data` | `api.statistics.getCalendarData(date)` | Query: `date` |
| GET | `/chart/overview_header` | `api.statistics.getOverviewHeader(date)` | Query: `date` |
| GET | `/chart/overview_statements` | `api.statistics.getOverviewStatements(date)` | Query: `date` |
| GET | `/chart/rate` | `api.statistics.getRate(date, type)` | Query: `date`, `type` |

## 12. 超级图表与流水

### 12.1 Super Statements

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/super_statements/time` | `api.superStatements.getTime()` | 无 |
| GET | `/super_statements/list` | `api.superStatements.getStatements(params)` | Query: 年月等筛选参数 |

### 12.2 Super Chart

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/super_chart/header` | `api.superCharts.getHeader(params)` | Query: 筛选参数 |
| GET | `/super_chart/get_pie_data` | `api.superCharts.getPieData(params)` | Query: 筛选参数 |
| GET | `/super_chart/week_data` | `api.superCharts.getWeekData(params)` | Query: 筛选参数 |
| GET | `/super_chart/line_chart` | `api.superCharts.getLineData({ year })` | Query: `year` |
| GET | `/super_chart/categories_list` | `api.superCharts.getCategoriesTop(...)` | Query: `year`, `month` |
| GET | `/super_chart/table_sumary` | `api.superCharts.getTableSumary(...)` | Query: `year`, `month` |

## 13. 消息

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/message` | `api.messages.getList()` | 无 |
| GET | `/message/:id` | `api.messages.getMessage(id)` | Path: `id` |

## 14. 收款方（Payees）

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/payees` | `api.payees.list()` | Query: `account_book_id` |
| POST | `/payees` | `api.payees.create(payee)` | Body: `{ account_book_id, payee }` |
| PUT | `/payees/:id` | `api.payees.update(payeeId, payee)` | Body: `{ account_book_id, payee }` |
| DELETE | `/payees/:id` | `api.payees.delete(payee)` | Body: `{ account_book_id }` |

## 15. 好友协作（Friends）

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| GET | `/friends` | `api.friends.list({ account_book_id })` | Query: `account_book_id` |
| POST | `/friends/invite` | `api.friends.invite(data)` | Body: `account_book_id`, `role` |
| GET | `/friends/invite_information` | `api.friends.information(token)` | Query: `invite_token` |
| POST | `/friends/accept_apply` | `api.friends.accept(token, nickname)` | Body: `invite_token`, `nickname` |
| DELETE | `/friends/:collaboratorId` | `api.friends.remove(data)` | Path: `collaborator_id`; Body: `account_book_id` |
| PUT | `/friends/:collaboratorId` | `api.friends.update(data)` | Path: `collaborator_id`; Body: 业务对象 |

## 16. 其他

| 方法 | 路径 | 对应方法 | 参数 |
|---|---|---|---|
| POST | `/settings/feedback` | `api.chaos.submitFeedback({ content })` | Body: `{ content, type: 0 }` |

## 17. 响应结构（前端约定）

- `src/api/http-result.ts` 将 Taro 响应包装为：
  - `status`：HTTP 状态码（取 `statusCode`）
  - `data`：服务端响应体
  - `header`：响应头
  - `isSuccess`：`status === 200`
- 业务代码中常见读取方式：`const { data } = await api.xxx()`，通常进一步判断 `data.status`、`data.msg`、`data.data`。

## 18. 备注

- 本文档反映的是“前端当前代码已调用/封装”的接口，不包含后端潜在但未被本项目调用的接口。
- 若后续新增 `src/api/logic/*` 方法，建议同步更新本文档对应模块。
