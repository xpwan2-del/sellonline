# 一级代理准入与购买归属修改计划

## 目标

废掉当前“普通用户自助开通推广返利”的旧模式，调整成“平台只允许一级代理商”的模式。

最终规则：

```text
平台生成代理邀请码
用户用平台代理邀请码注册
每个平台代理邀请码只能使用一次
用完立即失效
平台代理邀请码不设置过期时间
注册成功后，该用户成为一级代理商
一级代理商拥有自己的购买邀请链接：/?aff=代理码
买家通过代理商链接进入商城并购买
订单归属给这个一级代理商
买家以后复购，即使没有再次带 aff，也继续归属原代理商
佣金只给一级代理商
不做二级分佣
不做三级分佣
```

这里必须把两个邀请码分清楚：

```text
平台代理邀请码
用途：让用户注册成为一级代理商
本计划采用 URL 参数：agent_invite

代理商购买邀请链接
用途：让买家购买时归属到某个一级代理商
现有 URL 参数：aff
```

`agent_invite` 和 `aff` 不能混用。

## 不做什么

本次不是二级分佣，也不是多级分销。

本次不做：

- A 邀请 B，B 邀请 C，然后 C 下单时 A 和 B 同时拿佣金。
- 代理商之间上下级分佣。
- 给普通用户开放“自己开通推广返利”入口。
- 给已经注册好的普通账号补发平台代理邀请码并升级成代理商。
- 后台手动把普通账号改成代理商。
- 改支付渠道。
- 改钱包充值。
- 改优惠券。
- 改订单状态流转。
- 改卡密发放。
- 改商品库存。
- 改提现审核核心流程。
- 改现有商品级返佣计算方式。
- 改现有平台默认返佣计算方式。

## 原来的返佣系统怎么处理

原来的返佣计算系统不推倒重写。

保留这些已经做好的能力：

- 商品级返佣比例。
- 平台默认返佣比例。
- 一笔订单生成佣金汇总。
- 商品佣金明细 `affiliate_commission_items`。
- 佣金状态：待确认、可提现、已提现、已失效。
- 退款、取消订单后的佣金处理。
- 后台佣金记录。
- 后台返佣统计报表。
- 用户个人返佣报表。
- 提现申请和后台提现审核。

真正要撤销的是：

```text
旧规则，必须关闭：普通用户自助开通推广入口
```

真正要新增的是：

```text
平台代理邀请码
买家和一级代理商的永久归属关系
```

所以这次改动的核心不是“重新算佣金”，而是“谁有资格成为代理商”和“买家的订单归属给谁”。

## 精确改动数量

### 数据表

涉及 3 张表：

```text
新增 2 张表
1. affiliate_invite_codes
2. affiliate_customer_relations

修改 1 张表
3. affiliate_profiles
```

本计划不修改 `orders` 表。

原因：

当前订单已经有归属快照：

```text
orders.affiliate_profile_id
orders.affiliate_code
```

这两个字段足够记录“这笔订单归属哪个代理商”。

`affiliate_customer_relations` 用来记录“这个买家长期归属哪个代理商”。订单创建时读取这个关系，再把结果写入订单快照即可。

### 后台页面

涉及 6 个后台页面：

```text
新增 1 个页面
1. src/views/admin/AffiliateInviteCodes.vue

调整 5 个现有页面
2. src/views/admin/AffiliateSettings.vue
3. src/views/admin/AffiliateUsers.vue
4. src/views/admin/AffiliateReports.vue
5. src/views/admin/AffiliateCommissions.vue
6. src/views/admin/UserDetail.vue
```

### 用户前端页面

涉及 3 个用户前端页面/模块：

```text
1. src/views/auth/Register.vue
2. src/views/personal/AffiliatePanel.vue
3. src/views/personal/affiliate/AffiliateReportSection.vue
```

### API 接口

新增 6 个接口。

修改 16 个现有接口或入口逻辑。

新增接口：

```text
1. GET   /api/v1/admin/affiliate-invite-codes
2. POST  /api/v1/admin/affiliate-invite-codes
3. PATCH /api/v1/admin/affiliate-invite-codes/:id/status
4. GET   /api/v1/admin/affiliate-invite-codes/:id/usage
5. GET   /api/v1/admin/affiliates/:id/customers
6. GET   /api/v1/public/affiliate-invite-codes/:code/check
```

修改现有入口：

```text
1. POST /api/v1/auth/register
2. POST /api/v1/affiliate/open
3. GET  /api/v1/affiliate/dashboard
4. GET  /api/v1/affiliate/report/summary
5. GET  /api/v1/affiliate/report/commissions
6. POST /api/v1/public/affiliate/click
7. POST /api/v1/orders
8. POST /api/v1/orders/create-and-pay
9. POST /api/v1/guest/orders
10. POST /api/v1/guest/orders/create-and-pay
11. GET /api/v1/admin/settings/affiliate
12. PUT /api/v1/admin/settings/affiliate
13. GET /api/v1/admin/affiliates/users
14. GET /api/v1/admin/affiliates/reports/summary
15. GET /api/v1/admin/affiliates/reports/commissions
16. POST /api/v1/channel/affiliate/open
```

后台返佣报表接口不改变路由，但响应字段和筛选逻辑要调整：

```text
GET /api/v1/admin/affiliates/reports/summary
GET /api/v1/admin/affiliates/reports/commissions
```

## 数据库设计

### affiliate_invite_codes

用途：

平台生成“代理准入邀请码”。用户只有使用有效的邀请码注册，才会成为一级代理商。

每个平台代理邀请码只能使用一次。

用完后状态变为 `used`，不能再次注册。

平台代理邀请码不设置过期时间。

字段：

```text
id
code
status
used_by_user_id
used_at
created_by_admin_id
remark
created_at
updated_at
deleted_at
```

字段说明：

```text
code
唯一邀请码，例如 AGT8K2Q9

status
active / used / disabled

used_by_user_id
使用这个邀请码注册成功的用户 ID，可以为空

used_at
使用时间，可以为空

created_by_admin_id
创建这个邀请码的管理员 ID，可以为空

remark
后台备注
```

本计划使用索引：

```text
unique index idx_affiliate_invite_codes_code on code
index idx_affiliate_invite_codes_status on status
```

### affiliate_customer_relations

用途：

记录“买家长期归属哪个一级代理商”。

字段：

```text
id
customer_user_id
affiliate_profile_id
source_affiliate_code
source_order_id
source_type
bound_at
created_at
updated_at
deleted_at
```

字段说明：

```text
customer_user_id
买家用户 ID
一个买家只能长期归属一个代理商

affiliate_profile_id
一级代理商档案 ID

source_affiliate_code
第一次绑定时使用的代理商购买邀请码，也就是 aff

source_order_id
第一次产生绑定关系的订单 ID，可以为空

source_type
绑定来源：
register
order

bound_at
绑定时间
```

本计划使用索引：

```text
unique index idx_affiliate_customer_relations_customer_user_id on customer_user_id
index idx_affiliate_customer_relations_affiliate_profile_id on affiliate_profile_id
```

本计划不加数据库强外键。

原因：

项目当前更多是通过模型、索引、服务层校验来保证关系。这里继续按项目现有方式做，避免因为强外键影响用户删除、订单清理、测试数据重置。

### affiliate_profiles

现有代理/推广档案表继续使用。

新增字段：

```text
source
invite_code_id
```

字段说明：

```text
source
代理商来源：
platform_invite = 平台邀请码注册

invite_code_id
用户注册成为代理商时使用的平台代理邀请码 ID
```

本计划不支持后台手动开通代理商。

本计划不支持已经注册好的普通账号补用平台代理邀请码。

原因：

```text
一级代理身份必须在注册时确定
只有使用平台代理邀请码注册成功的人才是一级代理商
```

由于当前环境都是测试数据，不需要为历史测试数据做复杂兼容。

测试环境可以清理旧的 `affiliate_profiles`、`affiliate_clicks`、`affiliate_commissions`、`affiliate_commission_items`、`affiliate_withdraw_requests`、`affiliate_customer_relations` 后重新造数。

生产环境上线时仍然要使用非破坏式迁移：

```text
新增字段
新增表
不直接删除历史订单
不直接删除历史用户
```

## 业务规则

### 注册成为一级代理商

用户注册时可以带：

```text
agent_invite=平台代理邀请码
```

规则：

```text
有有效 agent_invite
→ 注册普通用户
→ 自动创建 affiliate_profiles
→ 用户成为一级代理商
→ invite_code.status = used
→ invite_code.used_by_user_id = 新注册用户 ID
→ invite_code.used_at = 当前时间

没有 agent_invite
→ 只注册普通用户
→ 不创建 affiliate_profiles

agent_invite 无效、停用、已经使用
→ 注册失败

已经注册好的普通账号后来拿到 agent_invite
→ 不支持补开通
→ 必须重新注册新账号
```

这里要明确：

```text
agent_invite 只决定是否成为一级代理商
不决定买家订单归属
```

注册请求不能同时使用：

```text
agent_invite
aff
```

如果两个同时出现，注册失败。

原因：

```text
agent_invite 是平台代理准入
aff 是买家归属
同一个注册动作不能既注册代理商，又注册成某个代理商名下买家
```

### 注册成为代理商名下买家

普通买家注册时可以带：

```text
aff=代理商购买邀请码
```

规则：

```text
有有效 aff
没有 agent_invite
→ 注册普通用户
→ 不创建 affiliate_profiles
→ 创建 affiliate_customer_relations
→ source_type = register
→ customer_user_id = 新注册用户 ID
→ affiliate_profile_id = aff 对应的一级代理商

没有 aff
没有 agent_invite
→ 注册普通用户
→ 不创建 affiliate_profiles
→ 不创建 affiliate_customer_relations
```

这里要明确：

```text
aff 只决定买家归属
不决定是否成为一级代理商
```

### 普通用户不能自己开通推广

现有：

```text
POST /api/v1/affiliate/open
```

调整后：

```text
已经是代理商
→ 返回已有代理档案

不是代理商
→ 返回错误：需要使用平台代理邀请码重新注册
```

用户前端不再给普通用户展示“开通推广返利”按钮。

### 代理商购买邀请链接

一级代理商拥有购买邀请链接：

```text
https://商城域名/?aff=代理码
```

`aff` 继续走现有前端工具：

```text
audit-dujiao-next-user/src/utils/affiliate.ts
```

这个工具只负责买家购买归属，不负责平台代理注册。

### 买家首次绑定

买家首次绑定可以发生在两个时机：

```text
1. 普通买家注册时带有效 aff
2. 登录买家第一次下单时带有效 aff
```

创建后：

```text
后续这个买家所有订单
即使没有 aff
也归属给同一个一级代理商
```

### 买家归属不能被自动覆盖

如果买家已经绑定代理商 A：

```text
买家后来又从代理商 B 的 ?aff=xxx 进入
不能自动改绑到 B
```

原因：

这是一级代理模式，必须保证代理商客户归属稳定。

本计划不提供后台改绑。

原因：

```text
买家归属永久固定
后台误改会影响历史归属和后续佣金
```

### 游客订单

游客没有 `customer_user_id`，所以不能建立永久归属关系。

游客订单规则：

```text
游客订单带有效 aff
→ 当前订单可以归属代理商
→ 可以生成佣金
→ 不创建 affiliate_customer_relations

游客订单不带 aff
→ 不归属代理商
```

如果以后游客注册成用户，再通过代理商链接下单，才建立永久归属。

“只做当次归因”的意思：

```text
游客 C 不登录，直接下单
这次订单请求带了 B 的 affiliate_code
affiliate_visitor_key 只用于辅助点击归因
→ 这笔订单归属 B
→ 这笔订单可以给 B 生成佣金

但系统不把游客永久绑定给 B
→ C 下次还是游客、但没有带 aff
→ 下次订单不自动归属 B
```

原因：

```text
游客没有 user_id
没有稳定账号身份可以绑定
```

所以永久归属只使用：

```text
登录用户的 user_id
```

### 只做一级佣金

订单只找一个代理商：

```text
orders.affiliate_profile_id
```

佣金只生成给这一个代理商。

不查代理商上级。

不新增 `parent_affiliate_profile_id`。

不新增多级分佣明细表。

## 后端文件归属

### 模型目录

目录：

```text
audit-dujiao-next-api/internal/models
```

新增文件：

```text
internal/models/affiliate_invite_code.go
internal/models/affiliate_customer_relation.go
```

修改文件：

```text
internal/models/affiliate_profile.go
internal/models/db.go
```

修改原因：

```text
affiliate_invite_code.go
定义平台代理邀请码表模型

affiliate_customer_relation.go
定义买家和代理商长期归属关系模型

affiliate_profile.go
增加代理商来源字段

db.go
把新增模型加入自动迁移/初始化列表
```

不要把新表模型写进 `affiliate_service.go` 或 handler 文件里。

### Repository 目录

目录：

```text
audit-dujiao-next-api/internal/repository
```

新增文件：

```text
internal/repository/affiliate_invite_code_repository.go
internal/repository/affiliate_customer_relation_repository.go
```

修改文件：

```text
internal/repository/affiliate_repository.go
internal/repository/affiliate_report_repository.go
```

修改原因：

```text
affiliate_invite_code_repository.go
只负责邀请码 CRUD、标记已使用、状态查询

affiliate_customer_relation_repository.go
只负责买家归属关系查询、创建

affiliate_repository.go
保留代理档案查询和状态管理，增加按 source/invite_code_id 查询

affiliate_report_repository.go
报表需要统计代理商客户数、复购订单时再补查询
```

不要把 SQL 堆到 Vue 页面。

不要把归属关系 SQL 写在 handler 里。

### Service 目录

目录：

```text
audit-dujiao-next-api/internal/service
```

新增文件：

```text
internal/service/affiliate_invite_service.go
internal/service/affiliate_customer_relation_service.go
```

修改文件：

```text
internal/service/affiliate_service.go
internal/service/user_auth_service.go
internal/service/order_service.go
internal/service/affiliate_report_service.go
internal/service/affiliate_setting.go
internal/service/errors.go
```

修改原因：

```text
affiliate_invite_service.go
负责平台代理邀请码生成、校验、标记已使用、停用

affiliate_customer_relation_service.go
负责买家和代理商的长期归属绑定、查询

affiliate_service.go
保留佣金计算主流程
调整 ResolveOrderAffiliateSnapshot：
先查买家长期归属
再看订单传入的 aff
最后写订单归属快照

user_auth_service.go
注册时接收平台代理邀请码，注册成功后创建代理档案

order_service.go
创建订单时继续写 orders.affiliate_profile_id / orders.affiliate_code
不改支付和订单状态逻辑

affiliate_report_service.go
用户个人报表和后台报表增加客户归属、复购口径

affiliate_setting.go
配置项增加“是否允许普通用户自助开通”的关闭语义，本计划按关闭处理

errors.go
增加邀请码无效、邀请码已使用、普通用户不可自助开通等错误
```

佣金计算仍然留在：

```text
internal/service/affiliate_service.go
```

尤其是商品级佣金计算相关逻辑不能搬家。

### Handler 目录

目录：

```text
audit-dujiao-next-api/internal/http/handlers
```

新增文件：

```text
internal/http/handlers/admin/admin_affiliate_invite_code.go
internal/http/handlers/admin/admin_affiliate_customer.go
```

修改文件：

```text
internal/http/handlers/public/user_auth.go
internal/http/handlers/public/affiliate.go
internal/http/handlers/public/public.go
internal/http/handlers/admin/admin_affiliate.go
internal/http/handlers/admin/admin_affiliate_manage.go
internal/http/handlers/admin/admin_affiliate_report.go
internal/http/handlers/channel/channel_affiliate.go
```

修改原因：

```text
admin_affiliate_invite_code.go
后台管理平台代理邀请码

admin_affiliate_customer.go
后台查看某个代理商名下买家

public/user_auth.go
注册请求增加 agent_invite、affiliate_code、affiliate_visitor_key 字段，传给 UserAuthService.Register
agent_invite 用于注册一级代理商
affiliate_code 用于注册普通买家并建立代理商归属

public/affiliate.go
普通用户自助开通逻辑改成受限制
public affiliate click 继续只处理买家 aff 点击

public/public.go
订单创建请求已经接收 affiliate_code / affiliate_visitor_key
需要确保下单入口继续传给 OrderService

admin_affiliate.go
后台返佣设置返回新配置字段

admin_affiliate_manage.go
代理商列表增加来源、邀请码信息、客户数

admin_affiliate_report.go
报表筛选和返回字段增加客户归属/复购统计

channel_affiliate.go
限制 Channel 端自助开通推广入口，防止普通 Channel 用户绕过平台代理邀请码创建代理档案
```

不要在 handler 里直接写复杂 SQL。

### Router 和 Container

修改文件：

```text
audit-dujiao-next-api/internal/router/router.go
audit-dujiao-next-api/internal/provider/container.go
```

修改原因：

```text
router.go
注册新增后台和公开接口

container.go
注入新增 repository 和 service
```

新增后台权限：

```text
GET:/admin/affiliate-invite-codes
POST:/admin/affiliate-invite-codes
PATCH:/admin/affiliate-invite-codes/:id/status
GET:/admin/affiliate-invite-codes/:id/usage
GET:/admin/affiliates/:id/customers
```

### DTO

目录：

```text
audit-dujiao-next-api/internal/dto
```

修改文件：

```text
internal/dto/affiliate.go
internal/dto/order.go
```

修改原因：

```text
affiliate.go
增加邀请码、代理商来源、客户归属、个人报表字段

order.go
确认订单请求/响应里的 affiliate_code 表示购买归属
affiliate_visitor_key 只用于辅助点击归因
```

## Admin 前端文件归属

### 路由和菜单

修改文件：

```text
audit-dujiao-next-admin/src/router/index.ts
audit-dujiao-next-admin/src/layouts/AdminLayout.vue
audit-dujiao-next-admin/src/i18n/index.ts
```

新增菜单：

```text
推广返利
- 返利设置
- 代理邀请码
- 返利用户
- 佣金记录
- 返佣统计报表
- 提现审核
```

新增路由：

```text
/affiliates/invite-codes
```

权限：

```text
GET:/admin/affiliate-invite-codes
```

### API 封装

修改文件：

```text
audit-dujiao-next-admin/src/api/admin.ts
audit-dujiao-next-admin/src/api/types.ts
```

新增方法：

```text
getAffiliateInviteCodes
createAffiliateInviteCode
updateAffiliateInviteCodeStatus
getAffiliateInviteCodeUsage
getAffiliateCustomers
```

新增类型：

```text
AdminAffiliateInviteCode
AdminAffiliateInviteCodeUsage
AdminAffiliateCustomerRelation
```

### 新增页面

新增文件：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateInviteCodes.vue
```

页面功能：

```text
查看平台代理邀请码
创建平台代理邀请码
启用/停用未使用的邀请码
查看邀请码使用记录
显示是否已使用、使用人、使用时间、备注
```

页面不做：

```text
不展示订单卡密
不展示支付渠道密钥
不修改佣金金额
```

### 现有后台页面调整

修改文件：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateSettings.vue
```

调整内容：

```text
说明普通用户不能自助开通推广
保留平台默认返佣比例配置
保留提现配置
```

修改文件：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateUsers.vue
```

调整内容：

```text
显示代理商来源
显示使用的平台代理邀请码
显示名下买家数量
增加查看名下买家入口
```

修改文件：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateReports.vue
```

调整内容：

```text
保留当前返佣统计报表
增加代理商客户数
增加复购订单数
增加新客订单/复购订单筛选或展示
```

修改文件：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateCommissions.vue
```

调整内容：

```text
佣金记录里显示买家是否来自长期归属
显示订单使用的 affiliate_code 快照
不展示卡密
```

修改文件：

```text
audit-dujiao-next-admin/src/views/admin/UserDetail.vue
```

调整内容：

```text
用户详情里显示：
如果该用户是代理商，显示代理商信息
如果该用户是买家，显示归属代理商
```

## 用户前端文件归属

### 注册页

修改文件：

```text
audit-dujiao-next-user/src/views/auth/Register.vue
audit-dujiao-next-user/src/api/auth.ts
audit-dujiao-next-user/src/api/types.ts
audit-dujiao-next-user/src/i18n/index.ts
```

调整内容：

```text
注册页读取 URL 参数 agent_invite
注册提交时带 agent_invite
如果邀请码有效，注册后自动成为一级代理商
如果邀请码无效，提示用户邀请码不可用
注册页也读取已有 ?aff 归因
普通买家注册时提交 affiliate_code、affiliate_visitor_key
普通买家注册成功后永久归属该 aff 对应的一级代理商
agent_invite 和 aff 同时存在时阻止注册并提示参数冲突
```

不要用 `aff` 作为平台代理邀请码。

### 买家购买归属

保留现有文件：

```text
audit-dujiao-next-user/src/utils/affiliate.ts
audit-dujiao-next-user/src/views/Checkout.vue
audit-dujiao-next-user/src/api/order.ts
```

调整内容：

```text
继续读取 ?aff=代理码
继续把 affiliate_code、affiliate_visitor_key 带到下单接口
不要在这里处理 agent_invite
```

本计划单独缓存平台代理邀请码时新增：

```text
audit-dujiao-next-user/src/utils/affiliateInvite.ts
```

它只处理：

```text
agent_invite
```

不处理：

```text
aff
```

### 个人推广返利页

修改文件：

```text
audit-dujiao-next-user/src/views/personal/AffiliatePanel.vue
audit-dujiao-next-user/src/views/personal/affiliate/AffiliateReportSection.vue
audit-dujiao-next-user/src/api/affiliate.ts
audit-dujiao-next-user/src/api/types.ts
audit-dujiao-next-user/src/i18n/index.ts
```

调整内容：

```text
普通用户：
不展示开通按钮
显示只有使用平台代理邀请码注册的新账号才能成为一级代理商

一级代理商：
显示自己的购买邀请链接
显示自己的返佣报表
显示自己的下游订单返佣明细
显示名下客户数、新客订单、复购订单
```

用户端不能看到：

```text
卡密
交付内容
兑换码
激活码
密钥
下载链接中的私密凭证
商品成本价
买家完整邮箱
买家完整手机号
其他一级代理商数据
```

这是硬规则：

```text
代理商用户前端报表接口不能返回卡密字段
代理商用户前端页面不能展示卡密字段
AffiliateReportSection.vue 只能展示订单号、商品名、SKU 名、数量、订单金额、佣金基数、比例、佣金、状态、时间
不能展示 order_items 里的交付内容
不能展示 card / secret / code / token / license / delivery_content 这类敏感字段
```

也就是说，不能只在页面上隐藏卡密。

API 返回给代理商前端的数据里，也不能包含卡密。

## API 设计

### 新增后台接口

```text
GET /api/v1/admin/affiliate-invite-codes
```

用途：

```text
后台分页查询平台代理邀请码
```

筛选：

```text
keyword
status
page
page_size
```

```text
POST /api/v1/admin/affiliate-invite-codes
```

用途：

```text
后台创建平台代理邀请码
```

请求字段：

```text
code 可选，不传则系统生成
remark
```

```text
PATCH /api/v1/admin/affiliate-invite-codes/:id/status
```

用途：

```text
启用、停用未使用的平台代理邀请码
status = used 的邀请码不能重新启用
```

```text
GET /api/v1/admin/affiliate-invite-codes/:id/usage
```

用途：

```text
查看某个平台代理邀请码被哪个用户使用
```

因为一个邀请码只能使用一次，所以返回最多一条使用记录。

```text
GET /api/v1/admin/affiliates/:id/customers
```

用途：

```text
查看某个一级代理商名下买家
```

本计划不提供后台解绑和改绑接口。

原因：

```text
买家归属永久固定
后台只能查看归属关系
不能因为后台误操作影响已有佣金归属
```

### 新增公开接口

```text
GET /api/v1/public/affiliate-invite-codes/:code/check
```

用途：

```text
注册页检查平台代理邀请码是否有效
```

返回：

```text
valid
message
```

这个接口只能返回是否有效，不能返回管理员信息、使用人列表等后台数据。

### 修改注册接口

现有：

```text
POST /api/v1/auth/register
```

新增请求字段：

```text
agent_invite
affiliate_code
affiliate_visitor_key
```

逻辑：

```text
agent_invite 为空
affiliate_code 为空
→ 普通注册
→ 不创建 affiliate_profiles
→ 不创建 affiliate_customer_relations

agent_invite 有值且有效
affiliate_code 为空
→ 普通注册
→ 创建 affiliate_profiles
→ source = platform_invite
→ invite_code_id = 对应邀请码 ID
→ 邀请码标记为 used

agent_invite 有值但无效
→ 注册失败

agent_invite 为空
affiliate_code 有值且有效
→ 普通注册
→ 不创建 affiliate_profiles
→ 创建 affiliate_customer_relations
→ source_type = register

agent_invite 有值
affiliate_code 也有值
→ 注册失败

已经存在的普通账号
→ 不能通过该接口补开通代理商
```

### 修改自助开通接口

现有：

```text
POST /api/v1/affiliate/open
```

调整：

```text
已经是代理商
→ 返回已有代理商档案

不是代理商
→ 返回错误
```

错误含义：

```text
普通用户不能自助开通，需要使用平台代理邀请码重新注册
```

### 修改订单归属逻辑

涉及入口：

```text
POST /api/v1/orders
POST /api/v1/orders/create-and-pay
POST /api/v1/guest/orders
POST /api/v1/guest/orders/create-and-pay
```

服务层归属规则：

```text
登录用户下单：
1. 先查 affiliate_customer_relations.customer_user_id
2. 如果已有长期归属，使用长期归属
3. 如果没有长期归属，再看请求里的 affiliate_code
4. affiliate_code 有效时，创建长期归属
5. 把最终代理商写入 orders.affiliate_profile_id 和 orders.affiliate_code

游客下单：
1. 没有 customer_user_id
2. 只看本次请求里的 affiliate_code
3. 有效则写入 orders.affiliate_profile_id 和 orders.affiliate_code
4. 不创建 affiliate_customer_relations
```

### 修改报表接口

现有后台报表：

```text
GET /api/v1/admin/affiliates/reports/summary
GET /api/v1/admin/affiliates/reports/commissions
```

增加返回字段：

```text
customer_count
new_customer_order_count
repeat_customer_order_count
```

现有用户个人报表：

```text
GET /api/v1/affiliate/report/summary
GET /api/v1/affiliate/report/commissions
```

增加返回字段：

```text
customer_count
new_customer_order_count
repeat_customer_order_count
```

用户个人接口必须从登录用户解析自己的 `affiliate_profile_id`。

不能允许用户传别人的 `affiliate_profile_id`。

## 当前现有路由位置

后端路由当前在：

```text
audit-dujiao-next-api/internal/router/router.go
```

现有公开点击接口：

```text
POST /api/v1/public/affiliate/click
```

现有注册接口：

```text
POST /api/v1/auth/register
```

现有用户推广接口：

```text
POST /api/v1/affiliate/open
GET  /api/v1/affiliate/dashboard
GET  /api/v1/affiliate/report/summary
GET  /api/v1/affiliate/report/commissions
GET  /api/v1/affiliate/commissions
GET  /api/v1/affiliate/withdraws
POST /api/v1/affiliate/withdraws
```

现有后台推广接口：

```text
GET   /api/v1/admin/settings/affiliate
PUT   /api/v1/admin/settings/affiliate
GET   /api/v1/admin/affiliates/users
PATCH /api/v1/admin/affiliates/users/:id/status
PATCH /api/v1/admin/affiliates/users/batch-status
GET   /api/v1/admin/affiliates/commissions
GET   /api/v1/admin/affiliates/reports/summary
GET   /api/v1/admin/affiliates/reports/commissions
GET   /api/v1/admin/affiliates/withdraws
POST  /api/v1/admin/affiliates/withdraws/:id/reject
POST  /api/v1/admin/affiliates/withdraws/:id/pay
```

Channel 端也有推广接口：

```text
POST /api/v1/channel/affiliate/open
POST /api/v1/channel/affiliate/click
GET  /api/v1/channel/affiliate/dashboard
GET  /api/v1/channel/affiliate/commissions
GET  /api/v1/channel/affiliate/withdraws
POST /api/v1/channel/affiliate/withdraws
```

本计划不能让 Channel 端成为普通用户自助开通推广的绕过入口。

原因：

最终规则是平台只允许一级代理商：

```text
只有使用平台代理邀请码注册成功的新账号才是一级代理商
普通用户不能自己开通推广返利
```

所以：

```text
POST /api/v1/channel/affiliate/open
```

不能继续直接给普通 Channel 用户创建 `affiliate_profiles`。

处理规则：

```text
如果 Channel 对应的用户已经是一级代理商
→ 返回已有代理档案

如果 Channel 对应的用户不是一级代理商
→ 返回错误
→ 提示必须通过平台代理邀请码注册新账号成为一级代理商
```

本计划暂不重做 Channel 端完整页面和流程，但 API 层必须和商城用户端保持同一条准入规则，不能留下自助开通漏洞。

## 测试计划

### 后端测试

新增或修改测试文件：

```text
audit-dujiao-next-api/internal/service/affiliate_invite_service_test.go
audit-dujiao-next-api/internal/service/affiliate_customer_relation_service_test.go
audit-dujiao-next-api/internal/service/user_auth_service_test.go
audit-dujiao-next-api/internal/service/order_service_test.go
audit-dujiao-next-api/internal/service/affiliate_service_test.go
audit-dujiao-next-api/internal/service/affiliate_report_service_test.go
audit-dujiao-next-api/internal/repository/affiliate_invite_code_repository_test.go
audit-dujiao-next-api/internal/repository/affiliate_customer_relation_repository_test.go
audit-dujiao-next-api/internal/http/handlers/channel/channel_affiliate_test.go
```

必须覆盖：

```text
平台代理邀请码创建成功
平台代理邀请码停用后不可用
平台代理邀请码使用一次后不可用
带有效 agent_invite 注册后自动成为代理商
不带 agent_invite 注册后只是普通用户
已经注册的普通账号不能补用 agent_invite 升级为代理商
普通买家带有效 aff 注册后绑定代理商
agent_invite 和 aff 同时提交时注册失败
普通用户调用 /affiliate/open 被拒绝
已是代理商调用 /affiliate/open 返回已有档案
Channel 用户不是一级代理商，调用 /channel/affiliate/open 被拒绝
Channel 用户已经是一级代理商，调用 /channel/affiliate/open 返回已有档案
买家第一次带 aff 下单后绑定代理商
买家复购不带 aff 仍然归属原代理商
买家已归属 A 后带 B 的 aff 下单不会自动改绑
游客带 aff 下单只归属当前订单，不创建长期关系
订单支付后仍然按现有商品级返佣生成佣金
退款后佣金仍然按现有逻辑失效或调整
后台报表只能看到归属后的订单和佣金
用户个人报表只能看到自己的代理数据
```

### Admin 前端测试

执行：

```text
npm run build
```

页面检查：

```text
代理邀请码页面能打开
能创建邀请码
能停用未使用的邀请码
已使用的邀请码不能重新启用
返利用户列表显示来源
返佣统计报表仍可筛选
佣金记录仍可查看商品明细
用户详情能看到代理归属
```

### 用户前端测试

执行：

```text
npm run build
```

页面检查：

```text
/auth/register?agent_invite=xxx 能注册代理商
普通注册不会成为代理商
普通老账号不能补用 agent_invite 成为代理商
/?aff=代理码 仍然能记录买家归属
Checkout 仍然提交 affiliate_code
普通用户个人中心不展示开通按钮
代理商个人中心能看到购买邀请链接和个人返佣报表
用户端不展示卡密
```

## 上线顺序

### 测试环境

因为当前都是测试数据，可以按这个顺序：

```text
1. API 新增表和字段
2. API 调整注册、下单归属、报表，并同步限制 /affiliate/open 和 /channel/affiliate/open 自助开通入口
3. Admin 增加代理邀请码页面
4. User 增加 agent_invite 注册逻辑
5. 清理旧测试返佣数据
6. 创建平台代理邀请码
7. 用平台代理邀请码注册一级代理商
8. 用一级代理商 ?aff 链接注册/购买测试
9. 验证佣金、报表、提现
```

可清理的测试数据范围：

```text
affiliate_customer_relations
affiliate_invite_codes
affiliate_withdraw_requests
affiliate_commission_items
affiliate_commissions
affiliate_clicks
affiliate_profiles
测试订单
测试用户
```

清理前必须确认只在测试环境执行。

### 生产环境

生产环境不能直接清表。

生产环境顺序：

```text
1. 先上数据库新增表和新增字段
2. 再上 API
3. 再上 Admin
4. 再上 User
5. 后台创建新的平台代理邀请码
6. 新用户按新规则进入
7. 老数据是否保留，由上线前单独确认
```

## 风险控制

必须遵守：

```text
不改支付 provider
不改支付回调签名校验
不改卡密解密和交付
不改库存扣减
不改钱包充值
不改优惠券
不改订单状态枚举
不改提现审核状态枚举
```

订单归属只允许在订单创建阶段决定：

```text
orders.affiliate_profile_id
orders.affiliate_code
```

订单支付成功后只根据订单快照生成佣金。

不要在支付回调里重新解析 `aff`。

不要在佣金计算时重新查 URL 参数。

## 验收标准

这个改造完成后，必须满足：

```text
普通用户不能自己开通推广返利
只有平台代理邀请码注册的人才是一级代理商
一级代理商有自己的 ?aff 购买邀请链接
买家第一次通过 ?aff 下单后长期归属这个代理商
买家复购继续给同一个代理商算佣金
买家不会被其他代理商链接自动抢走
游客订单只做当前订单归属
佣金仍然按商品级返佣和平台默认返佣计算
后台能看平台代理邀请码、代理商、佣金、报表、提现
代理商能在用户前端看自己的个人返佣报表
用户前端个人返佣报表不泄露卡密、交付内容、兑换码、激活码、密钥、私密下载凭证
用户前端个人返佣报表接口响应里也不能包含这些敏感字段
其他订单、支付、库存、钱包、优惠券功能不受影响
```
