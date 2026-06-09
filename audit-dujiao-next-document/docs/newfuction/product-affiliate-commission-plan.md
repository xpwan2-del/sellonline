# 商品级推广返利配置修改计划

## 目标

将当前推广返利从“全局统一返利比例”升级为“商品级返利比例覆盖”。

修改完成后，系统需要支持：

- 全局返利比例继续作为默认比例。
- 单个商品可以单独设置返利比例。
- 一个订单内包含多个商品、多个数量、不同返利比例时，可以正确计算佣金。
- 后台可以配置商品返利比例。
- 后台和用户端可以清楚看到“综合返利比例”和商品佣金明细。
- 现有提现、退款、订单取消逻辑保持可用。

本计划只定义“佣金怎么算”。

代理商准入、买家归属、游客当次归因、普通用户不能自助开通推广，以 `affiliate-platform-agent-invite-plan.md` 为准。

也就是说：

```text
商品级返佣计划
→ 决定订单归属到代理商后，按商品行怎么算佣金

一级代理准入与购买归属计划
→ 决定谁是一级代理商、买家订单归属给谁
```

## 明确边界

本次只做商品级返佣，不做 SKU 级返佣。

也就是说，一个商品下所有 SKU 使用同一个返佣配置。

示例：

```text
商品 A
- SKU: 月卡
- SKU: 季卡
- SKU: 年卡
```

如果商品 A 设置：

```text
参与推广返利 = 开启
单商品返利比例 = 10%
```

那么商品 A 下所有 SKU 都按 10% 计算返利。

本次不支持：

```text
月卡 10%
季卡 15%
年卡 20%
```

因此不会修改 `product_skus` 表，也不会在 SKU 编辑区域增加返利配置。

## 当前逻辑

当前返利比例来自全局配置：

```text
settings.affiliate_config.commission_rate
```

当前佣金计算逻辑：

```text
佣金基数 = 订单中已开启返利商品的成交金额合计
佣金金额 = 佣金基数 * 全局返利比例 / 100
```

其中商品是否参与返利由 `products.is_affiliate_enabled` 控制。

当前佣金记录表 `affiliate_commissions` 是订单级汇总记录，一笔订单通常生成一条佣金记录。

本计划不改变订单归属来源。

订单是否归属代理商，仍以订单创建阶段写入的：

```text
orders.affiliate_profile_id
orders.affiliate_code
```

为准。

## 数据库变更

本次数据库结构变更涉及 2 张表：

```text
1. products
2. affiliate_commission_items
```

### products

新增字段：

```text
affiliate_commission_rate DECIMAL(10,2) NULL
```

字段含义：

```text
NULL  = 使用全局默认返利比例
0.00  = 该商品返利 0%
10.00 = 该商品按 10% 返利
20.00 = 该商品按 20% 返利
```

与现有字段、全局返利设置组合后的业务含义：

```text
is_affiliate_enabled = false
→ 该商品不参与返利
→ 即使设置了 affiliate_commission_rate，也不产生佣金

is_affiliate_enabled = true
affiliate_commission_rate = 10.00
→ 使用商品单独 10%
→ 不受全局 enabled 是否开启影响

is_affiliate_enabled = true
affiliate_commission_rate = NULL
全局 enabled = true
全局 commission_rate = 20.00
→ 使用全局默认 20%

is_affiliate_enabled = true
affiliate_commission_rate = NULL
全局 enabled = false
→ 没有可用返利比例，不产生佣金
```

这里的全局 `enabled` 指现有返利设置中的字段：

```text
settings.affiliate_config.enabled
```

也就是后台“推广返利设置”页面当前的“启用推广返利”开关。

它不是新增字段，也不是 `products` 表字段。

本次改造后，该字段语义需要调整为：

```text
是否启用平台默认返利比例
```

全局返利设置中的 `enabled` 只控制平台默认返利比例是否可用，不是整个推广返利系统的总开关。

商品单独返利比例优先级高于全局默认比例。

如果全局 enabled = false，但商品设置了单独返利比例：

```text
商品 affiliate_commission_rate = 10.00
→ 使用商品单独 10%，正常产生佣金
```

如果商品没有设置单独返利比例，才会判断全局 enabled 和全局 commission_rate。

后台页面必须提示：

```text
开启商品返利后，如果未设置单商品返利比例，将按照平台默认返利比例计算。
```

不在 `product_skus` 表增加任何返利字段。

### affiliate_commission_items

新增佣金明细表：

```text
affiliate_commission_items
```

字段：

```text
id
affiliate_commission_id
order_item_id
product_id
base_amount
rate_percent
commission_amount
created_at
updated_at
deleted_at
```

字段说明：

```text
id
主键

affiliate_commission_id
关联 affiliate_commissions.id，表示属于哪一条佣金汇总

order_item_id
关联 order_items.id，表示是哪一个订单商品行产生的佣金

product_id
关联 products.id，方便按商品统计返利

base_amount
该商品行佣金基数

rate_percent
该商品行实际使用的返利比例

commission_amount
该商品行产生的佣金金额

created_at
创建时间

updated_at
更新时间

deleted_at
软删除时间
```

建议索引：

```text
idx_affiliate_commission_items_commission_id
affiliate_commission_id

idx_affiliate_commission_items_order_item_id
order_item_id

idx_affiliate_commission_items_product_id
product_id

idx_affiliate_commission_items_deleted_at
deleted_at
```

建议唯一约束：

```text
affiliate_commission_id + order_item_id
```

用于防止同一条订单商品行在同一条佣金汇总下重复生成佣金明细。

## 不改结构但参与逻辑的表

以下表不新增字段、不删除字段，但会参与计算或展示：

```text
affiliate_commissions
orders
order_items
affiliate_profiles
affiliate_withdraw_requests
```

### affiliate_commissions

继续作为订单级佣金汇总表。

字段含义调整为：

```text
base_amount
= 所有 affiliate_commission_items.base_amount 合计

commission_amount
= 所有 affiliate_commission_items.commission_amount 合计

rate_percent
= 综合返利比例
= commission_amount / base_amount * 100
```

页面展示时必须使用“综合返利比例”，不能只显示一个百分比，避免用户误解。

### order_items

继续作为订单商品行快照。

佣金明细按 `order_items` 维度生成，不按商品数量逐个拆分。

例如订单购买：

```text
A 商品 x 3
B 商品 x 1
C 商品 x 10
```

如果订单中有 3 条 `order_items`，则生成 3 条 `affiliate_commission_items`，不是生成 14 条。

## 后端模型变更

新增模型：

```text
models.AffiliateCommissionItem
```

修改模型：

```text
models.Product
```

新增字段：

```text
AffiliateCommissionRate
```

该字段需要能表达 NULL，因此不能用普通非指针数值字段。

自动迁移中加入：

```text
AffiliateCommissionItem
```

## 后端接口变更

商品创建、编辑接口增加字段：

```text
affiliate_commission_rate
```

影响范围：

```text
CreateProductRequest
UpdateProductRequest
CreateProductInput
UpdateProductInput
AdminProduct 响应类型
```

校验规则：

```text
允许 NULL
允许 0-100
最多两位小数
小于 0 拒绝
大于 100 拒绝
```

## 佣金计算逻辑

订单支付成功后，按订单商品行逐条计算佣金。

计算规则：

```text
如果商品未开启 is_affiliate_enabled
→ 跳过

如果商品开启返利，且 affiliate_commission_rate 有值
→ 使用商品单独比例

如果商品开启返利，且 affiliate_commission_rate 为 NULL
→ 判断全局 enabled

如果商品开启返利，且 affiliate_commission_rate 为 NULL，且全局 enabled = true
→ 使用全局默认比例 commission_rate

如果商品开启返利，且 affiliate_commission_rate 为 NULL，且全局 enabled = false
→ 没有可用返利比例，跳过

base_amount = order_item.total_price - order_item.coupon_discount_amount
commission_amount = base_amount * 实际比例 / 100

如果实际比例 <= 0 或 commission_amount <= 0
→ 该订单商品行不生成佣金明细
```

最终生成：

```text
1 条 affiliate_commissions 汇总记录
N 条 affiliate_commission_items 明细记录
```

其中 N 等于参与返利的订单商品行数量。

## 多商品订单示例

全局默认返利比例：

```text
20%
```

订单商品：

```text
A 商品 x 3
成交金额 300
优惠分摊 30
商品单独返利 10%

B 商品 x 1
成交金额 80
优惠分摊 0
商品单独返利 NULL，使用全局 20%

C 商品 x 10
成交金额 500
优惠分摊 50
商品单独返利 5%
```

生成明细：

```text
A: base_amount = 270, rate = 10%, commission = 27.00
B: base_amount = 80,  rate = 20%, commission = 16.00
C: base_amount = 450, rate = 5%,  commission = 22.50
```

汇总：

```text
base_amount = 270 + 80 + 450 = 800.00
commission_amount = 27 + 16 + 22.50 = 65.50
综合返利比例 = 65.50 / 800.00 * 100 = 8.19%
```

页面展示：

```text
佣金基数：800.00
综合返利比例：8.19%
佣金金额：65.50
```

展开明细：

```text
商品    数量    佣金基数    返利比例    佣金
A       x3      270.00      10%        27.00
B       x1      80.00       20%        16.00
C       x10     450.00      5%         22.50
合计            800.00      综合8.19%  65.50
```

## 提现拆分逻辑

现有提现逻辑在提现金额小于某条可提现佣金金额时，会拆分 `affiliate_commissions` 记录。

引入明细表后，提现拆分时必须同步拆分明细。

处理规则：

```text
原佣金记录拆成：
1. 提现绑定部分
2. 剩余可提现部分

affiliate_commission_items 也按金额比例拆分成两组
分别挂到两条 affiliate_commissions 下
```

目标：

```text
每条 affiliate_commissions 的 commission_amount
必须等于它名下 affiliate_commission_items.commission_amount 合计
```

## 退款与取消订单逻辑

本次不做商品级退款改造。

也就是说，本次不新增：

```text
order_refund_record_items
```

退款记录仍沿用当前订单级退款结构，只记录：

```text
order_id
amount
type
remark
```

因此系统无法知道部分退款具体退的是哪一个商品。

在这种限制下，退款佣金处理规则为：

```text
订单取消
→ 未进入提现流程的佣金作废

订单全额退款
→ 未进入提现流程的佣金作废

订单部分退款
→ 未进入提现流程的佣金按整单退款比例扣减
→ affiliate_commission_items 也按相同比例扣减

佣金已进入提现流程
→ 沿用现有规则，不回滚
```

部分退款按整单比例扣减的原因：

```text
当前退款记录没有 order_item_id / product_id / 退款数量
所以无法判断退款具体对应哪个商品
```

如果未来要做到“退哪个商品就扣哪个商品的佣金”，需要另做商品级退款改造。

订单取消、全额退款或部分退款扣减到 0 后，继续以 `affiliate_commissions.status` 作为佣金是否有效的状态来源。

`affiliate_commission_items` 不单独增加状态字段。

处理方式：

```text
affiliate_commissions.status = rejected
→ 该汇总下所有明细视为无效
```

这样可以避免新增一套明细状态机。

## 后台页面变更

一次性修改以下后台页面。

### 商品编辑页

文件：

```text
audit-dujiao-next-admin/src/views/admin/components/ProductEditModal.vue
```

新增商品级配置：

```text
推广返利
[开关] 参与推广返利

单商品返利比例（%）
[        ]
开启商品返利后，如果未设置单商品返利比例，将按照平台默认返利比例计算。
填写后该商品按单商品比例计算返利。
```

位置：

```text
商品级成本价下面
排序权重上面
```

不放到 SKU 配置区域。

交互：

```text
参与推广返利关闭
→ 单商品返利比例输入框禁用或隐藏

参与推广返利开启
→ 显示单商品返利比例输入框
→ 留空 = 使用平台默认返利比例
→ 填 0 = 该商品不产生佣金
→ 填 20 = 按该商品成交金额 20% 返利
```

### 商品列表页

文件：

```text
audit-dujiao-next-admin/src/views/admin/Products.vue
```

展示返利状态：

```text
返利关闭
使用全局默认
单独 10%
```

如果全局默认比例可用，可以显示：

```text
全局默认 20%
单独 10%
```

### 返利设置页

文件：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateSettings.vue
```

文案调整：

```text
返利比例（%）
```

改为：

```text
默认返利比例（%）
```

说明调整为：

```text
未单独配置商品比例时使用
```

现有开关文案：

```text
启用推广返利
```

需要改为：

```text
启用平台默认返利
```

说明：

```text
关闭后，未单独设置商品返利比例的商品不会产生佣金；已设置单商品返利比例的商品仍按商品比例计算。
```

该开关只控制平台默认返利比例。

它不控制普通用户是否可以自助开通推广。

普通用户不能自助开通推广，这个规则以 `affiliate-platform-agent-invite-plan.md` 为准。

### 管理端佣金记录页

文件：

```text
audit-dujiao-next-admin/src/views/admin/AffiliateCommissions.vue
```

列表字段调整：

```text
返利比例
```

改为：

```text
综合返利比例
```

增加商品佣金明细查看能力。

明细字段：

```text
商品
SKU
数量
佣金基数
返利比例
佣金
```

## 用户端页面变更

购买页、商品页不展示商品返利配置。

本商品级返佣计划本身不要求修改 Checkout。

如果同时实现一级代理准入与购买归属计划，Checkout 是否需要继续提交 `affiliate_code`，以 `affiliate-platform-agent-invite-plan.md` 为准。

仅按商品级返佣任务本身，不修改：

```text
ProductDetail.vue
ProductCard.vue
ProductListItem.vue
ProductQuickBuy.vue
Cart.vue
Checkout.vue
```

用户中心推广返利页需要兼容明细展示。

文件：

```text
audit-dujiao-next-user/src/views/personal/AffiliatePanel.vue
```

展示规则：

```text
佣金记录显示最终佣金金额
比例字段显示“综合返利比例”
可展开查看商品明细
```

如果不展示明细，也至少不能把综合比例误写成普通返利比例。

## 接口返回设计

管理端佣金列表或详情需要返回明细：

```text
commission_items: [
  {
    order_item_id,
    product_id,
    product_title,
    sku_snapshot,
    quantity,
    base_amount,
    rate_percent,
    commission_amount
  }
]
```

其中：

```text
quantity
product_title
sku_snapshot
```

可以从 `order_items` 关联读取，不一定冗余存入 `affiliate_commission_items`。

## 测试覆盖

需要覆盖以下场景：

```text
1. 商品未开启返利，不生成佣金
2. 商品开启返利，商品比例 NULL，使用全局比例
3. 商品开启返利，商品比例 10%，覆盖全局比例
4. 一单多商品、多数量、不同比例，生成多条明细和一条汇总
5. 综合返利比例计算正确
6. 订单取消后，未进入提现流程的汇总佣金 rejected，明细仍可用于后台查账
7. 订单全额退款后，未进入提现流程的汇总佣金 rejected，明细仍可用于后台查账
8. 订单部分退款后，未进入提现流程的佣金和明细按整单退款比例扣减
9. 已进入提现流程的佣金遇到退款时沿用现有规则，不回滚
10. 提现拆分佣金时，明细同步拆分，金额对得上
11. 后台商品编辑保存后，单商品比例能正确回显
12. 后台商品列表能显示返利关闭、全局默认、单独比例
13. 管理端佣金记录能查看商品佣金明细
14. 用户端推广佣金记录不会误把综合比例显示成普通比例
```

## 一次性完成标准

本次修改完成时，需要同时满足：

```text
数据库字段和新表已自动迁移
商品后台能配置商品级返利比例
订单支付后能生成汇总佣金和商品佣金明细
多商品、多数量、不同比例订单计算正确
退款、取消、提现拆分逻辑不破坏金额一致性
部分退款按整单比例扣减佣金，不做商品级精准退款扣佣
后台佣金记录能查明细
前后台文案明确区分“默认返利比例”和“综合返利比例”
不引入 SKU 级返佣配置
```
