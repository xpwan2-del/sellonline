# 推广返利展示优化与客户列表修改计划

## 目标

本计划只做推广返利相关的展示和查询优化，不改变返佣计算规则。

本次需求包含 4 个功能点：

```text
1. Admin 后台平台代理邀请码列表增加“复制邀请链接”按钮
2. Admin 后台返佣报表买家邮箱不再脱敏
3. 用户前端推广返利报表买家邮箱不再脱敏
4. 用户前端推广返利页面改成 Tabs，并新增“我的客户”列表
```

本计划依赖现有规则：

```text
平台代理邀请码使用 agent_invite
一级代理商购买邀请链接使用 aff
平台只允许一级代理商
只有使用平台代理邀请码注册的新账号才是一级代理商
买家通过一级代理商购买邀请链接注册或购买后，归属该一级代理商
佣金只给一级代理商
不做二级分佣
不做三级分佣
```

## 明确边界

本次不修改：

```text
返佣比例计算
商品级返佣规则
平台默认返佣规则
订单创建逻辑
订单支付逻辑
退款和取消订单后的佣金失效逻辑
提现申请逻辑
提现审核逻辑
钱包充值逻辑
卡密发放逻辑
库存逻辑
平台代理邀请码使用规则
买家永久归属规则
```

本次必须继续保证：

```text
用户前端只能看到自己的推广返利数据
用户前端不能查看其他代理商数据
用户前端不能展示卡密
用户前端不能展示交付内容
用户前端不能展示成本价
用户前端不能展示后台管理字段
```

本次允许展示完整买家邮箱，但只限于：

```text
Admin 后台返佣报表
用户前端当前代理商自己的下游订单返佣明细
用户前端当前代理商自己的佣金记录
```

完整买家邮箱不等于开放订单交付内容。

## 文件归属和代码架构硬性要求

本次开发必须先检查现有项目结构，再决定具体落点。

禁止：

```text
随手新建和现有命名风格不一致的目录
把 API 查询 SQL 写在 handler 里
把复杂页面逻辑全部堆进一个 Vue 文件
为了省事复制一套和现有 service/repository 重复的逻辑
把 Admin 页面文件放到 User 项目
把 User 页面文件放到 Admin 项目
把公共类型随便放在页面目录
把测试工具目录提交进 Git
```

必须遵守当前项目分层：

```text
API:
handler 只负责参数解析、鉴权上下文、调用 service、返回 response
service 负责业务规则、权限边界、组合 repository
repository 负责数据库查询
dto 负责响应结构
router 只注册路由

Admin:
页面放在 src/views/admin
接口封装放在 src/api
类型按现有结构放在 src/api/types.ts 或项目现有类型文件
i18n 文案放在 src/i18n/index.ts

User:
个人中心页面放在 src/views/personal
推广返利子组件放在 src/views/personal/affiliate
接口封装放在 src/api
类型按现有结构放在 src/api/types.ts 或项目现有类型文件
i18n 文案放在 src/i18n/index.ts
```

如果开发前发现实际文件名和本计划列出的文件名不完全一致，必须按现有项目真实结构归类，不允许为了匹配文档而新建重复文件。

开发前必须检查：

```text
1. 现有 Admin 代理邀请码页面文件名和接口封装位置
2. 现有 Admin 返佣报表页面文件名和接口封装位置
3. 现有 User 推广返利页面是否已经拆子组件
4. 现有 API affiliate report handler/service/repository/dto 文件名
5. 现有 affiliate_customer_relations repository/service 是否已经存在
6. 现有分页、筛选、response、错误码写法
7. 现有测试文件命名和测试 fixture 写法
```

本计划列出的文件路径是归属建议，不是允许乱建文件的理由。

最终代码必须符合现有项目架构。

## 需求 1：后台平台代理邀请码复制邀请链接

### 现状

后台已经有平台代理邀请码列表。

目前表格里只展示邀请码本身，例如：

```text
42AHT6FC
```

运营需要手动拼接注册链接：

```text
https://toplenged.com/auth/register?agent_invite=42AHT6FC
```

### 修改目标

在单个邀请码表格行最后增加一个操作按钮：

```text
复制邀请链接
```

点击后自动复制完整开户链接：

```text
https://toplenged.com/auth/register?agent_invite=邀请码
```

### 页面草图

```text
推广返利 / 代理邀请码
────────────────────────────────────────────────────────────────────────
[生成邀请码] [刷新]

ID   邀请码      状态     使用人              使用时间       创建时间       操作
#1   42AHT6FC    未使用   -                   -              2026/6/7       [复制邀请链接] [停用]
#2   J65MXCQJ    已使用   user@example.com    2026/6/7       2026/6/7       [复制邀请链接]

点击“复制邀请链接”后复制：
https://toplenged.com/auth/register?agent_invite=42AHT6FC
```

### 邀请链接域名来源

不能使用 Admin 当前运行时域名拼接注册链接。

原因：

Admin 后台域名是：

```text
https://admin.toplenged.com
```

用户注册链接应该是商城前端域名：

```text
https://toplenged.com
```

第一版必须固定使用商城前端域名：

```text
https://toplenged.com
```

本次不要从 Admin 当前域名推导，也不要依赖后台品牌配置。

第一版可以在该页面使用：

```text
const storefrontBaseURL = 'https://toplenged.com'
```

后续如果要改成配置化，必须确保配置值指向商城前端域名，而不是 Admin 后台域名。

禁止：

```text
https://admin.toplenged.com/auth/register?agent_invite=邀请码
```

### 文件归属

Admin 前端：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateInviteCodes.vue
```

可能涉及：

```text
audit-dujiao-next-admin/src/i18n/index.ts
```

不需要新增 API。

不需要修改数据库。

## 需求 2：Admin 后台返佣报表买家邮箱不脱敏

### 现状

Admin 后台返佣统计报表中，买家邮箱当前展示为脱敏格式：

```text
2***@qq.com
w***@163.com
```

### 修改目标

Admin 后台返佣统计报表明细直接展示完整买家邮箱：

```text
2622418533@qq.com
wangxue_fengcb@163.com
```

### 影响页面

```text
audit-dujiao-next-admin/src/views/admin/AffiliateReports.vue
```

如果 Admin 佣金记录页面也使用同一个买家字段，并且当前也脱敏，需要一起检查：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateCommissions.vue
```

### API 字段处理

优先方案：

```text
API 返回完整 buyer_email
Admin 直接展示 buyer_email
```

如果当前 API 只有脱敏后的字段，例如：

```text
buyer_email_masked
```

则新增字段：

```text
buyer_email
```

并保留旧字段，避免影响其他旧页面。

不推荐把脱敏字段直接改名。

### 文件归属

API 可能涉及：

```text
audit-dujiao-next-api/internal/dto/affiliate_report.go
audit-dujiao-next-api/internal/repository/affiliate_report_repository.go
audit-dujiao-next-api/internal/service/affiliate_report_service.go
audit-dujiao-next-api/internal/http/handlers/admin/admin_affiliate_report.go
```

实际以当前项目文件名为准。

Admin 前端可能涉及：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateReports.vue
audit-dujiao-next-admin/src/api/affiliate.ts
audit-dujiao-next-admin/src/api/types.ts
```

如果类型定义集中在别的文件，按现有项目结构修改，不新增乱放文件。

## 需求 3：用户前端推广返利买家邮箱不脱敏

### 现状

用户前端“我的推广返利报表”里，下游订单返佣明细显示：

```text
买家 w***@163.com
```

佣金记录里没有完整买家邮箱。

### 修改目标

用户前端只对当前登录一级代理商展示其名下订单买家完整邮箱：

```text
买家 wangxue_fengcb@163.com
买家 2622418533@qq.com
```

### 安全边界

允许展示：

```text
买家邮箱
订单号
订单金额
佣金金额
返佣比例
佣金状态
返佣来源
商品佣金明细
下单时间
```

禁止展示：

```text
卡密
交付内容
下载链接
供应商成本
后台备注
风控字段
买家密码
买家钱包余额
买家手机号，除非未来明确要求
```

### API 字段处理

用户报表接口当前为：

```text
GET /api/v1/affiliate/report/summary
GET /api/v1/affiliate/report/commissions
```

本次可在佣金明细响应中增加：

```text
buyer_email
```

保留旧字段：

```text
buyer_email_masked
```

用户前端改用：

```text
buyer_email
```

如果 `buyer_email` 为空，再 fallback 到 `buyer_email_masked`。

### 文件归属

API 可能涉及：

```text
audit-dujiao-next-api/internal/dto/affiliate_report.go
audit-dujiao-next-api/internal/repository/affiliate_report_repository.go
audit-dujiao-next-api/internal/service/affiliate_report_service.go
audit-dujiao-next-api/internal/http/handlers/public/affiliate_report.go
```

用户前端可能涉及：

```text
audit-dujiao-next-user/src/views/personal/AffiliatePanel.vue
audit-dujiao-next-user/src/api/affiliate.ts
audit-dujiao-next-user/src/api/types.ts
```

如果现有用户返佣报表已经拆成子组件，则应放在现有子组件目录，不在页面根部堆代码。

## 需求 4：用户前端推广返利页面改成 Tabs

### 现状

当前“推广返利”页面从上到下连续展示多个模块：

```text
我的推广返利报表
下游订单返佣明细
申请提现
佣金记录
提现记录
```

页面太长，不方便看。

### 修改目标

改成 Tab 结构：

```text
推广返利
仅展示返佣信息，不展示卡密/交付内容。

[数据总览] [下游订单] [我的客户] [佣金记录] [申请提现] [提现记录]
```

### Tab 1：数据总览

```text
推广返利
────────────────────────────────────────
[数据总览] [下游订单] [我的客户] [佣金记录] [申请提现] [提现记录]

当前 Tab：数据总览
────────────────────────────────────────
[今天] [本周] [本月] [上月]
[开始日期] [结束日期] [全部状态] [查询]

总佣金        可提现        待确认        已提现
634.39        534.39        0.00          0.00

名下客户      新客订单      复购订单      有效订单
1             1             2             3

点击数        转化率        平均佣金
1             300.00%       211.46

佣金趋势
05-31  0.00
06-01  0.00
06-06  394.39
06-07  240.00

返佣来源拆分
平台默认返佣    449.65
单商品返佣      184.74
```

### Tab 2：下游订单

```text
[数据总览] [下游订单] [我的客户] [佣金记录] [申请提现] [提现记录]

当前 Tab：下游订单
────────────────────────────────────────
[开始日期] [结束日期] [订单号] [买家邮箱] [状态] [查询] [重置]

订单号              买家邮箱                  订单金额   返佣比例   佣金    来源       状态     时间
DJ20260607...       2622418533@qq.com         100.00    5.00%      5.00    单商品返佣 可提现   2026/6/7
DJ20260607...       454742995@qq.com          560.00    15.00%     60.00   单商品返佣 可提现   2026/6/7

[商品明细]
```

这里的“下游订单”本质仍然展示返佣明细，不展示订单交付内容。

### Tab 3：我的客户

新增。

展示当前一级代理商名下客户列表。

```text
[数据总览] [下游订单] [我的客户] [佣金记录] [申请提现] [提现记录]

当前 Tab：我的客户
────────────────────────────────────────
客户列表

[客户邮箱/昵称] [开始注册时间] [结束注册时间] [查询] [重置]

客户ID   邮箱                     昵称             注册时间              最近下单时间          订单数   贡献佣金
#11      2622418533@qq.com        2622418533       2026/6/7 15:23       2026/6/7 15:45       3       70.00
#13      454742995@qq.com         454742995        2026/6/7 15:25       -                    0       0.00
```

### Tab 4：佣金记录

```text
[数据总览] [下游订单] [我的客户] [佣金记录] [申请提现] [提现记录]

当前 Tab：佣金记录
────────────────────────────────────────
[刷新]

订单号              买家邮箱                  综合返利比例   佣金     状态     时间
DJ20260607...       2622418533@qq.com         5.00%          5.00     可提现   2026/6/7
DJ20260607...       454742995@qq.com          15.00%         60.00    可提现   2026/6/7

[明细]
```

### Tab 5：申请提现

```text
[数据总览] [下游订单] [我的客户] [佣金记录] [申请提现] [提现记录]

当前 Tab：申请提现
────────────────────────────────────────
可提现金额：534.39

提现金额
[请输入提现金额]

提现渠道
[请输入或选择提现渠道]

提现账号
[请输入提现账号]

[提交申请]
```

### Tab 6：提现记录

```text
[数据总览] [下游订单] [我的客户] [佣金记录] [申请提现] [提现记录]

当前 Tab：提现记录
────────────────────────────────────────
[刷新]

金额       渠道       状态       申请时间
100.00     支付宝     待审核     2026/6/7 11:58:45
```

### 前端组件归属

不建议继续把所有逻辑堆在：

```text
audit-dujiao-next-user/src/views/personal/AffiliatePanel.vue
```

推荐拆成现有页面下的子组件目录：

```text
audit-dujiao-next-user/src/views/personal/affiliate/
```

建议组件：

```text
AffiliateReportOverview.vue
AffiliateDownstreamOrders.vue
AffiliateCustomerList.vue
AffiliateCommissionList.vue
AffiliateWithdrawForm.vue
AffiliateWithdrawList.vue
```

如果当前项目已经存在类似子组件，优先复用现有文件名和结构。

`AffiliatePanel.vue` 只负责：

```text
页面标题
代理商资格判断
Tab 状态
基础布局
把数据请求分发给子组件
```

不要把所有表格、筛选、提现表单都塞在一个文件里。

## 需求 4 的后端接口：新增我的客户接口

### 业务含义

“我的客户”只展示当前登录一级代理商名下的客户。

客户来源：

```text
affiliate_customer_relations
```

关系含义：

```text
affiliate_profile_id = 当前登录代理商的 affiliate_profiles.id
customer_user_id = 下游客户 users.id
```

不展示游客当次归因订单。

原因：

```text
游客没有固定 user_id
游客只做当次订单归因
不建立永久客户关系
```

### 新增接口

建议新增：

```text
GET /api/v1/affiliate/customers
```

该接口只读取当前登录用户自己的代理档案。

该接口不允许传入：

```text
affiliate_profile_id
user_id
agent_id
```

原因：

用户前端不能通过传参查看其他代理商客户。

### 查询参数

```text
keyword
registered_from
registered_to
page
page_size
```

字段含义：

```text
keyword          客户邮箱或昵称，模糊搜索
registered_from  客户注册开始时间
registered_to    客户注册结束时间
page             页码
page_size        每页数量
```

### 响应字段

```text
id
email
display_name
registered_at
last_order_at
order_count
commission_amount
```

示例：

```json
{
  "id": 11,
  "email": "2622418533@qq.com",
  "display_name": "2622418533",
  "registered_at": "2026-06-07T15:23:00+08:00",
  "last_order_at": "2026-06-07T15:45:00+08:00",
  "order_count": 3,
  "commission_amount": "70.00"
}
```

### SQL 查询原则

按项目现有 repository/service/handler 分层实现。

不要在 handler 里直接拼 SQL。

推荐查询来源：

```text
affiliate_customer_relations
users
orders
affiliate_commissions
```

核心条件：

```text
affiliate_customer_relations.affiliate_profile_id = 当前登录代理商 profile id
affiliate_customer_relations.customer_user_id = users.id
users.deleted_at IS NULL
```

统计字段：

```text
last_order_at       当前代理商名下该客户的最近支付订单时间
order_count         当前代理商名下该客户的有效订单数
commission_amount   当前代理商从该客户订单获得的佣金合计
```

有效订单必须沿用返佣报表现有口径。

第一版原则：

```text
不要在本接口里重新定义一套有效订单状态
不要硬写和返佣报表不同的订单状态条件
优先复用现有返佣报表 repository/service 的状态条件或 helper
```

具体状态常量必须按项目现有 constants 使用，不在 SQL 里硬写魔法字符串。

### API 文件归属

API 建议新增或扩展：

```text
audit-dujiao-next-api/internal/dto/affiliate_customer.go
audit-dujiao-next-api/internal/repository/affiliate_customer_relation_repository.go
audit-dujiao-next-api/internal/service/affiliate_service.go
audit-dujiao-next-api/internal/http/handlers/public/affiliate_customer.go
audit-dujiao-next-api/internal/router/router.go
```

如果现有项目已有客户 relation service/repository，应优先扩展现有文件。

不要新建和现有命名风格不一致的目录。

## API 修改清单

### API 开发前文件归类检查

开始写 API 代码前，必须先在 `audit-dujiao-next-api` 中确认真实文件位置。

必须检查：

```text
internal/router/router.go
internal/http/handlers/public/affiliate_report.go
internal/http/handlers/admin/admin_affiliate_report.go
internal/service/affiliate_report_service.go
internal/repository/affiliate_report_repository.go
internal/repository/affiliate_customer_relation_repository.go
internal/dto/affiliate.go
internal/dto/affiliate_report.go
```

如果某个文件不存在，不要立刻新建同名文件。

处理规则：

```text
1. 先查项目里现有 affiliate/customer/report 相关文件
2. 如果已有同类文件，优先扩展现有文件
3. 只有确实没有合适归属时，才按项目命名风格新增文件
4. 新增文件必须放在对应层目录：dto/repository/service/handler
5. 不允许把 repository 查询写进 handler
6. 不允许把 response DTO 写在 repository
7. 不允许新建和现有命名冲突的重复文件
```

### 新增接口

```text
GET /api/v1/affiliate/customers
```

### 修改接口

```text
GET /api/v1/admin/affiliates/reports/commissions
GET /api/v1/affiliate/report/commissions
GET /api/v1/affiliate/commissions
```

修改内容：

```text
增加或改用完整 buyer_email
保留 buyer_email_masked 兼容旧展示
不返回卡密/交付内容
```

如果 Admin 和 User 报表共用同一套 DTO，需要在 DTO 中同时保留：

```text
buyer_email
buyer_email_masked
```

如果某些接口不需要买家邮箱，不强行新增。

## 前端修改清单

### Admin

### Admin 开发前文件归类检查

开始写 Admin 代码前，必须先在 `audit-dujiao-next-admin` 中确认真实文件位置。

必须检查：

```text
src/views/admin/AffiliateInviteCodes.vue
src/views/admin/AffiliateReports.vue
src/views/admin/AffiliateCommissions.vue
src/api/affiliate.ts
src/api/types.ts
src/i18n/index.ts
```

处理规则：

```text
1. 页面修改必须留在 src/views/admin
2. 邀请码复制按钮放在 AffiliateInviteCodes.vue 的表格操作列
3. 返佣报表邮箱展示放在 AffiliateReports.vue 对应明细列
4. 接口请求必须通过 src/api 现有封装
5. 文案必须进 src/i18n/index.ts，不要写散落的硬编码中文
6. 不要把 Admin 组件放到 User 项目
7. 不要新建无关 component 目录
```

```text
audit-dujiao-next-admin/src/views/admin/AffiliateInviteCodes.vue
```

修改：

```text
表格操作列新增“复制邀请链接”
点击后复制完整注册链接
复制成功后显示 toast/提示
```

可能修改：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateReports.vue
audit-dujiao-next-admin/src/views/admin/AffiliateCommissions.vue
audit-dujiao-next-admin/src/api/affiliate.ts
audit-dujiao-next-admin/src/api/types.ts
audit-dujiao-next-admin/src/i18n/index.ts
```

### User

### User 开发前文件归类检查

开始写 User 代码前，必须先在 `audit-dujiao-next-user` 中确认真实文件位置。

必须检查：

```text
src/views/personal/AffiliatePanel.vue
src/views/personal/affiliate/
src/api/affiliate.ts
src/api/types.ts
src/i18n/index.ts
```

处理规则：

```text
1. 个人中心推广返利入口仍然由 AffiliatePanel.vue 管理
2. 如果已有 src/views/personal/affiliate 目录，必须复用
3. 新 Tab 子组件必须放在 src/views/personal/affiliate
4. 不要把 Tab 子组件放到 src/components 下，除非项目已有同类约定
5. API 请求必须通过 src/api/affiliate.ts 封装
6. 类型必须放到现有 types 文件，不要散落在页面里
7. i18n 文案必须放到 src/i18n/index.ts
8. 不要把 User 页面文件放进 Admin 项目
```

```text
audit-dujiao-next-user/src/views/personal/AffiliatePanel.vue
```

建议拆分：

```text
audit-dujiao-next-user/src/views/personal/affiliate/AffiliateReportOverview.vue
audit-dujiao-next-user/src/views/personal/affiliate/AffiliateDownstreamOrders.vue
audit-dujiao-next-user/src/views/personal/affiliate/AffiliateCustomerList.vue
audit-dujiao-next-user/src/views/personal/affiliate/AffiliateCommissionList.vue
audit-dujiao-next-user/src/views/personal/affiliate/AffiliateWithdrawForm.vue
audit-dujiao-next-user/src/views/personal/affiliate/AffiliateWithdrawList.vue
```

可能修改：

```text
audit-dujiao-next-user/src/api/affiliate.ts
audit-dujiao-next-user/src/api/types.ts
audit-dujiao-next-user/src/i18n/index.ts
```

## 数据库变更

本次不新增数据库表。

本次不修改数据库表结构。

“我的客户”读取已有：

```text
affiliate_customer_relations
users
orders
affiliate_commissions
```

如果当前实现还没有 `affiliate_customer_relations`，则必须先完成：

```text
affiliate-platform-agent-invite-plan.md
```

本计划不重复定义该表结构。

## 测试计划

### Admin 邀请码复制

测试：

```text
1. 打开后台代理邀请码列表
2. 点击某条邀请码的“复制邀请链接”
3. 粘贴结果应为 https://toplenged.com/auth/register?agent_invite=邀请码
4. 已使用邀请码也可以复制链接
5. 停用邀请码也可以复制链接，复制只是复制链接，不改变状态
```

### Admin 买家邮箱完整展示

测试：

```text
1. 打开后台返佣统计报表
2. 查询已有佣金明细
3. 买家邮箱显示完整邮箱
4. 商品明细仍然只展示佣金商品行
5. 不展示卡密
6. 不展示交付内容
```

### User 买家邮箱完整展示

测试：

```text
1. 用一级代理商账号登录商城前端
2. 打开个人中心 / 推广返利
3. 下游订单 Tab 显示完整买家邮箱
4. 佣金记录 Tab 显示完整买家邮箱
5. 普通非代理用户不能看到返佣报表数据
6. 不能看到卡密或交付内容
```

### User Tabs

测试：

```text
1. 推广返利页面默认进入数据总览 Tab
2. 切换下游订单 Tab 不丢失登录态
3. 切换我的客户 Tab 能正常加载客户列表
4. 切换佣金记录 Tab 能正常加载佣金记录
5. 切换申请提现 Tab 可以提交提现申请
6. 切换提现记录 Tab 可以查看提现历史
7. 移动端 Tab 不挤压、不换行混乱
```

### 我的客户

测试：

```text
1. 一级代理商可以看到自己名下客户
2. 客户邮箱完整展示
3. keyword 可按邮箱搜索
4. keyword 可按昵称搜索
5. registered_from + registered_to 可按注册时间筛选
6. 组合查询可同时生效
7. 不能传 affiliate_profile_id 查看别人客户
8. 游客订单不会出现在客户列表
9. 客户订单数和贡献佣金与返佣报表口径一致
```

### 回归测试

必须确认：

```text
注册流程正常
平台代理邀请码注册一级代理商正常
购买邀请链接 aff 归属正常
下单正常
支付正常
返佣计算正常
提现申请正常
后台提现审核正常
后台佣金记录正常
后台返佣统计报表原有筛选正常
商品级返佣筛选正常
```

## 风险点

### 风险 1：完整买家邮箱暴露范围过大

控制方式：

```text
只在 Admin 和当前代理商自己的返佣报表里展示完整邮箱
普通用户不能访问别人的返佣报表
接口按当前登录 user_id 查代理档案，不接受外部 affiliate_profile_id
```

### 风险 2：Tabs 拆分时把提现逻辑弄乱

控制方式：

```text
提现申请和提现记录只做组件移动
不改变原接口
不改变原提交字段
不改变原状态展示
```

### 风险 3：客户列表统计口径和报表不一致

控制方式：

```text
客户列表中的 order_count 和 commission_amount 使用现有报表 service/repository 的状态口径
不要在新 SQL 中硬写一套不同状态
```

### 风险 4：复制链接域名错误

控制方式：

```text
不能使用 admin.toplenged.com 拼注册链接
注册链接必须使用 toplenged.com
```

## 开发顺序

建议按以下顺序开发：

```text
1. API：完整买家邮箱字段
2. Admin：返佣报表展示完整买家邮箱
3. Admin：邀请码列表复制完整注册链接
4. API：新增 /api/v1/affiliate/customers
5. User：新增 affiliate customers API 类型和方法
6. User：推广返利页面拆 Tabs
7. User：新增我的客户 Tab
8. User：下游订单和佣金记录展示完整买家邮箱
9. 跑 API 单元测试
10. 跑 Admin/User build
11. 部署测试环境
12. 用真实账号完成回归验证
```

不要先大规模重构页面。

每一步完成后都要确认不影响现有功能。

## 验收标准

完成后必须满足：

```text
后台平台代理邀请码可以一键复制完整注册链接
后台返佣报表买家邮箱完整显示
用户前端返佣报表买家邮箱完整显示
用户前端推广返利页面变成 Tabs
用户前端新增我的客户 Tab
我的客户支持邮箱/昵称和注册时间组合查询
普通用户不能看代理商报表
代理商不能看其他代理商客户
卡密和交付内容不泄露
商品级返佣和平台默认返佣计算不变
提现功能不变
```
