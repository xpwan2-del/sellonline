# 用户订单与充值统计数量修复计划

## 目标

修复用户端个人中心里订单、充值统计卡片数量不准确的问题。

当前问题：

```text
用户端订单/充值列表是分页加载。
页面顶部统计卡片现在按当前页列表数据计算数量。
如果用户有很多订单或充值记录，当前页只加载 20 条，统计卡片就只统计这 20 条，数量会偏小。
```

正确目标：

```text
列表继续分页加载。
统计卡片单独调用 API 聚合接口。
统计接口按当前用户的全量订单/充值记录计算。
统计结果不受分页影响。
```

本计划只修复：

```text
用户订单统计数量
用户充值统计数量
个人中心订单总数展示
```

不在本计划中重复修改：

```text
邮箱后缀注册限制
素材库批量删除
卡密导出
后台批发价
前台批发价展示
批发价计费一致性
```

上述功能已在本地代码中发现对应实现，本计划不重复搬官方补丁，避免覆盖现有定制代码。

## 官方更新参考

官方相关更新是：

```text
API:
76641c2 fix: 用户订单统计卡片数量bug

User:
efd6268 fix: 用户订单统计卡片数量bug
```

该修复的核心不是 admin 仪表盘，而是用户端个人中心的订单/充值统计卡片。

## 修改范围

开发前必须再次用搜索确认本地是否已经存在以下符号：

```text
orders/stats
wallet/recharges/stats
OrderStats
MyWalletRechargeStats
StatsOrdersByUser
StatsUserRechargeOrders
orderStats
rechargeStats
```

如果开发时这些符号已经存在，必须先阅读现有实现，再决定是补齐遗漏还是调整调用，不能重复新增一套同名或相似逻辑。

## 代码架构与落地纪律

本计划实施时，必须严格按照当前项目已有代码架构、代码风格和分层标准修改，不能为了套官方提交而乱堆代码。

必须遵守：

```text
先读本地现有代码，再动手修改。
以本地项目真实结构为准，不以官方文件路径机械覆盖。
只在现有职责文件中补齐逻辑，不新建重复职责文件。
不新建重复 API 模块。
不新建重复 service。
不新建重复 repository。
不把业务查询写进 handler。
不把数据库查询写进前端。
不把用户端统计逻辑放进 admin。
不把 admin 仪表盘统计接口拿给用户端复用。
不复制官方整段代码覆盖本地定制逻辑。
不改与本计划无关的邮箱后缀、素材库、卡密导出、批发价、支付、订单创建逻辑。
```

如果官方代码与本地代码有差异，处理原则是：

```text
保留本地已有定制。
只吸收本次统计修复所需的最小逻辑。
函数名、错误处理、响应格式、API 封装方式必须贴合本地现有写法。
测试优先验证本地业务行为，而不是机械追求与官方 diff 完全一致。
```

本计划的正确落地方式是：

```text
API 按 router -> handler -> service -> repository 补齐 stats 链路。
User 按 api 封装 -> store/page 状态 -> template 展示补齐 stats 使用。
数据库只读聚合，不做 schema 迁移。
部署只更新目标环境代码，不覆盖数据库、上传文件、配置文件或域名配置。
```

### API 项目

项目：

```text
audit-dujiao-next-api
```

预计修改文件：

```text
internal/router/router.go
internal/http/handlers/public/order.go
internal/http/handlers/public/wallet.go
internal/service/order_service_query.go
internal/service/wallet_service.go
internal/repository/order_repository.go
internal/repository/wallet_repository.go
```

可能新增或补充测试：

```text
internal/repository/order_repository_test.go
internal/repository/wallet_repository_test.go
internal/http/handlers/public/*_test.go
```

### User 用户端项目

项目：

```text
audit-dujiao-next-user
```

预计修改文件：

```text
src/api/order.ts
src/api/wallet.ts
src/stores/userProfile.ts
src/views/PersonalCenter.vue
src/views/personal/OrdersPanel.vue
```

不需要修改：

```text
audit-dujiao-next-admin
```

原因：本次修复是用户个人中心统计卡片，不是后台 admin 仪表盘。

## API 设计

新增两个用户鉴权接口：

```text
GET /api/v1/orders/stats
GET /api/v1/wallet/recharges/stats
```

两个接口必须挂在现有用户鉴权路由组中：

```text
apiV1.Group("")
  .Use(UserJWTAuthMiddleware(...))
```

不能挂到 admin 路由，不能做成公开接口。

### 订单统计接口

接口：

```text
GET /api/v1/orders/stats
```

支持参数：

```text
order_no
```

统计规则：

```text
只统计当前登录用户 user_id。
只统计父订单 parent_id IS NULL。
不受分页影响。
不应用 status 筛选，因为统计目的就是返回各状态分布。
可以复用 order_no 关键词筛选。
```

返回结构。

注意：以下结构是 `response.Success` 返回体里的 `data` 内容，不是绕过项目响应规范直接返回裸 JSON。

```json
{
  "total": 12,
  "by_status": {
    "pending_payment": 1,
    "paid": 2,
    "delivered": 5,
    "completed": 4
  }
}
```

### 充值统计接口

接口：

```text
GET /api/v1/wallet/recharges/stats
```

支持参数：

```text
recharge_no
```

统计规则：

```text
只统计当前登录用户 user_id。
不受分页影响。
不应用 status 筛选。
可以复用 recharge_no 关键词筛选。
```

返回结构。

注意：以下结构是 `response.Success` 返回体里的 `data` 内容，不是绕过项目响应规范直接返回裸 JSON。

```json
{
  "total": 3,
  "by_status": {
    "pending": 1,
    "success": 2
  }
}
```

## API 现有代码结构要求

必须遵循现有分层：

```text
router
  -> handler
  -> service
  -> repository
```

不允许：

```text
handler 直接写数据库查询
service 直接拼 SQL
为了一个统计接口新建重复 repository
把统计逻辑写到 admin 模块
把当前页数据拿来伪装全量统计
```

### 订单 API 链路

现有订单列表链路：

```text
internal/router/router.go
  user.GET("/orders", publicHandler.ListOrders)

internal/http/handlers/public/order.go
  Handler.ListOrders

internal/service/order_service_query.go
  OrderService.ListOrdersByUser

internal/repository/order_repository.go
  OrderRepository.ListByUser
```

新增统计链路应为：

```text
internal/router/router.go
  user.GET("/orders/stats", publicHandler.OrderStats)

internal/http/handlers/public/order.go
  Handler.OrderStats

internal/service/order_service_query.go
  OrderService.StatsOrdersByUser

internal/repository/order_repository.go
  OrderRepository.StatsByUser
```

### 充值 API 链路

现有充值列表链路：

```text
internal/router/router.go
  user.GET("/wallet/recharges", publicHandler.ListMyWalletRecharges)

internal/http/handlers/public/wallet.go
  Handler.ListMyWalletRecharges

internal/service/wallet_service.go
  WalletService.ListUserRechargeOrders

internal/repository/wallet_repository.go
  WalletRepository.ListRechargeOrdersAdmin
```

新增统计链路应为：

```text
internal/router/router.go
  user.GET("/wallet/recharges/stats", publicHandler.MyWalletRechargeStats)

internal/http/handlers/public/wallet.go
  Handler.MyWalletRechargeStats

internal/service/wallet_service.go
  WalletService.StatsUserRechargeOrders

internal/repository/wallet_repository.go
  WalletRepository.StatsRechargeOrders
```

## User 用户端结构要求

必须沿用现有用户端结构：

```text
src/api/order.ts
  userOrderAPI

src/api/wallet.ts
  walletAPI

src/stores/userProfile.ts
  useUserProfileStore

src/views/PersonalCenter.vue
  个人中心总览

src/views/personal/OrdersPanel.vue
  订单和充值列表页
```

不允许：

```text
在组件里手写 fetch 绕过 api 封装
新建重复订单 API 文件
用当前页 orders.length 继续当全量统计
把充值统计混进订单统计
把 admin 仪表盘统计接口拿来给用户端使用
```

### 用户端 API 封装

`src/api/order.ts` 增加：

```text
userOrderAPI.stats(params)
```

对应：

```text
GET /orders/stats
```

`src/api/wallet.ts` 增加：

```text
walletAPI.rechargeStats(params)
```

对应：

```text
GET /wallet/recharges/stats
```

### PersonalCenter.vue

当前问题：

```text
订单数展示使用 recentOrders.length。
recentOrders 只是最近几条订单，不是总订单数。
```

计划：

```text
userProfileStore 增加 ordersTotal。
loadRecentOrders 读取订单列表接口 pagination.total。
PersonalCenter.vue 使用 ordersTotal 展示订单总数。
```

### OrdersPanel.vue

当前问题：

```text
pendingPaymentCount 从当前页 orders 里 filter。
finishedCount 从当前页 orders 里 filter。
rechargePendingCount 从当前页 rechargeOrders 里 filter。
```

计划：

```text
增加 orderStats。
增加 rechargeStats。
loadOrders 完成列表加载后，调用 loadOrderStats。
loadRechargeOrders 完成列表加载后，调用 loadRechargeStats。
pendingPaymentCount 从 orderStats["pending_payment"] 读取。
finishedCount 从 delivered/completed/partially_refunded/refunded 的全量统计求和。
rechargePendingCount 从 rechargeStats["pending"] 读取。
```

筛选规则：

```text
订单 stats 只跟随 order_no 关键词。
订单 stats 不跟随 status。
充值 stats 只跟随 recharge_no 关键词。
充值 stats 不跟随 status。
```

这样切换状态筛选时，统计卡片仍然代表当前关键词下的全量状态分布，而不是当前状态或当前页。

## 数据库影响

本计划不需要改数据库结构。

不新增表：

```text
不新增 orders 字段
不新增 wallet_recharge_orders 字段
不新增 settings 字段
```

只读取现有表：

```text
orders
wallet_recharge_orders
```

不修改业务数据：

```text
不修改订单
不修改充值记录
不修改支付记录
不修改用户
不修改钱包余额
```

## 测试计划

### API 测试

至少验证：

```text
订单 stats 只统计当前用户。
订单 stats 只统计 parent_id IS NULL 的父订单。
订单 stats 不受分页影响。
订单 stats 不应用 status 筛选。
订单 stats 可以按 order_no 关键词过滤。
充值 stats 只统计当前用户。
充值 stats 不受分页影响。
充值 stats 不应用 status 筛选。
充值 stats 可以按 recharge_no 关键词过滤。
```

### User 前端测试

至少验证：

```text
个人中心顶部订单总数等于 pagination.total，不再等于 recentOrders.length。
订单列表第一页只有 20 条时，待支付/已完成统计仍按全量返回。
切换订单状态筛选后，统计卡片不被当前页覆盖。
输入订单号关键词后，统计卡片跟随关键词变化。
充值列表第一页只有 20 条时，待处理统计仍按全量返回。
切换充值状态筛选后，统计卡片不被当前页覆盖。
输入充值单号关键词后，统计卡片跟随关键词变化。
```

### 回归测试

必须确认：

```text
订单列表分页正常。
订单详情正常。
取消订单正常。
充值列表分页正常。
充值详情正常。
主动检查充值支付状态正常。
游客订单不受影响。
admin 后台不受影响。
```

## 部署注意事项

该功能涉及：

```text
API 后端
用户端 Vercel 前端
```

不涉及：

```text
admin 前端
数据库迁移
服务器配置
域名配置
Caddy/Nginx 配置
```

部署时必须按现有三环境规则执行，不能推错分支。

参考：

```text
audit-dujiao-next-document/docs/newfuction/two-server-github-deployment-guide.md
```

测试环境优先：

```text
服务器：34.146.83.35 / instance-testsell
用户端 Vercel 项目：test
用户端分支：deploy/test
```

正式环境必须单独确认后再发布：

```text
toplenged -> deploy/toplenged
888tech   -> deploy/888tech
```

不允许因为修 API 就覆盖数据库或上传文件。

## 风险点

### 风险 1：路由顺序

新增：

```text
/orders/stats
/wallet/recharges/stats
```

必须放在动态路由之前：

```text
/orders/:order_no
/wallet/recharges/:recharge_no
```

否则 `stats` 可能被当成 `order_no` 或 `recharge_no`。

### 风险 2：统计口径

订单统计必须只统计父订单：

```text
parent_id IS NULL
```

否则父子订单会重复计数。

### 风险 3：状态筛选

统计接口不能应用当前 status 筛选。

原因：

```text
统计卡片显示的是各状态分布。
如果应用 status=pending_payment，其他状态都会变成 0，统计卡片没有意义。
```

### 风险 4：前端异步覆盖

`loadOrders` 和 `loadOrderStats` 都是异步。

需要避免：

```text
旧请求后返回，覆盖新关键词的统计结果。
```

如果实际开发中发现快速输入关键词会造成闪动，可以沿用现有 `debounceAsync` 机制，或在响应前核对当前筛选值。

## 验收标准

功能完成后必须满足：

```text
API 存在 /orders/stats。
API 存在 /wallet/recharges/stats。
未登录访问 stats 接口会被用户鉴权拦截。
订单 stats 返回当前用户全量父订单状态分布。
充值 stats 返回当前用户全量充值状态分布。
用户端个人中心订单总数不再使用 recentOrders.length。
用户端订单统计卡片不再使用当前页 orders.filter。
用户端充值统计卡片不再使用当前页 rechargeOrders.filter。
不修改数据库结构。
不影响 admin。
不重复改已经存在的邮箱后缀、素材批量删除、卡密导出、批发价功能。
```
