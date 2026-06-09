# 商品展示销量与手机端卡片优化修改计划

## 目标

本计划包含两个相关但可以分步实施的前端展示优化：

- 商品公开接口增加展示销量 `sold_count`。
- 商品首页、商品中心、商品详情页在用户端展示“已售 X 件”。
- 手机端商品卡片视觉更充实，减少过大的左右留白，提升商品卡片在移动端的存在感。
- 商品展示图统一按 `1:1` 正方形完整展示，避免 1080x1080 商品图被横图容器裁剪。

核心原则：

```text
前台展示销量 = 真实已售数量 + 后台手动加成数量
```

这样既保留真实业务数据，又允许运营在后台调整前台展示氛围。

## 总修改清单

本功能需要修改三个项目。

### API 项目

项目：

```text
audit-dujiao-next-api
```

预计修改文件：

```text
internal/models/product.go
internal/dto/product.go
internal/http/handlers/public/public.go
internal/http/handlers/admin/admin_product.go
internal/service/product_service.go
```

可能需要同步检查：

```text
internal/service/product_service_test.go
internal/http/handlers/public/public_stock_test.go
```

API 要完成的事情：

- `products` 表新增 `sold_count_offset` 字段。
- 沿用现有 `CreateProductRequest -> service.CreateProductInput -> models.Product` 链路保存 `sold_count_offset`。
- 商品公开列表接口返回 `sold_count`。
- 商品公开详情接口返回 `sold_count`。
- `sold_count = real_sold_count + sold_count_offset`。
- 自动发货商品真实已售优先使用卡密已使用数量。
- 手动发货商品真实已售使用 SKU 已售数量汇总。

### Admin 项目

项目：

```text
audit-dujiao-next-admin
```

预计修改文件：

```text
src/api/types.ts
src/views/admin/components/ProductEditModal.vue
src/i18n/index.ts
```

可能需要同步检查：

```text
src/views/admin/Products.vue
src/api/admin.ts
```

说明：

- `src/api/admin.ts` 已经用 `createProduct(data: Partial<AdminProduct>)` 和 `updateProduct(id, data)` 传 payload，通常不需要新增 API 方法。
- `Products.vue` 是列表入口，商品新增/编辑弹窗实际在 `ProductEditModal.vue`。

Admin 要完成的事情：

- 商品类型 `AdminProduct` 增加 `sold_count_offset`。
- 商品编辑弹窗增加“销量加成”输入框。
- 编辑商品时回显 `sold_count_offset`。
- 保存商品时提交 `sold_count_offset`。
- 增加中文、繁体、英文文案。

### User 用户端项目

项目：

```text
audit-dujiao-next-user
```

预计修改文件：

```text
src/components/ProductCard.vue
src/components/Footer.vue
src/views/ProductDetail.vue
src/views/Products.vue
src/views/Home.vue
src/i18n/index.ts
```

可能需要同步检查：

```text
src/api/types.ts
```

说明：当前用户端商品数据在 `ProductCard.vue`、`Home.vue`、`Products.vue`、`ProductDetail.vue` 中主要以 `any` 传递，`src/api/types.ts` 暂无统一 PublicProduct 类型。如果实施时不新增公共商品类型，可以不修改该文件。

User 要完成的事情：

- 商品卡片显示“已售 X 件”。
- 商品详情页显示“已售 X 件”。
- 商品卡片主图完整显示 1:1 商品展示图。
- 商品详情页主图完整显示 1:1 商品展示图。
- 页脚版权品牌名从后台品牌配置读取，不再硬编码 `Toplenged`。
- 手机端商品卡片减少左右留白。
- 手机端商品标题允许两行。
- 手机端商品卡片显示简短描述。
- 首页精选商品和商品中心列表保持一致的手机端密度。

## 现有结构核对结论

本计划必须沿用现有代码结构，不新增无关页面或绕开现有链路。

## 实施纪律

开发时必须先按真实项目结构定位文件，不能为了匹配文档而新建重复文件。

必须遵守：

- 如果计划中列出的文件和实际项目结构不完全一致，优先服从实际项目结构。
- 如果某个文件不存在，先查找现有同职责文件，不允许直接新建同名或相似文件。
- API 字段必须沿现有 handler -> service -> model 链路传递。
- 公开商品响应必须沿现有 `decoratePublicProduct -> toProductResp -> dto.ProductResp` 链路输出。
- Admin 表单必须沿现有 `ProductEditModal.vue` 的 `form / resetForm / populateForm / handleSubmit` 修改。
- User 商品卡片必须沿现有 `ProductCard.vue` 修改。
- User 商品详情主图必须沿现有 `ProductImageGallery.vue` 修改。
- 文案必须进入现有 `src/i18n/index.ts`，不允许散落硬编码中文。
- 不允许按商品名、slug、图片路径写特殊判断。
- 不允许新建与现有职责重复的组件、接口文件、页面文件。
- 不允许把 repository 查询写进 handler。
- 不允许把展示销量用于订单、支付、库存、佣金、财务统计。

### API 现有链路

商品新增/编辑现有链路是：

```text
internal/http/handlers/admin/admin_product.go
CreateProductRequest
    -> service.CreateProductInput
    -> internal/service/product_service.go
    -> models.Product
```

因此 `sold_count_offset` 需要依次补进：

```text
CreateProductRequest
CreateProductInput
models.Product
ProductService.Create
ProductService.Update
```

公开商品返回现有链路是：

```text
ProductService.ListPublic / GetPublicBySlug
    -> ProductService.ApplyAutoStockCounts
    -> public.go decoratePublicProduct
    -> publicProductView.toProductResp
    -> dto.ProductResp
```

因此 `sold_count` 不应另起一个接口，应该在现有公开商品列表和详情响应里返回。

### Admin 现有链路

商品编辑现有链路是：

```text
src/views/admin/Products.vue
    -> src/views/admin/components/ProductEditModal.vue
    -> adminAPI.createProduct / adminAPI.updateProduct
```

因此后台只需要在现有 `ProductEditModal.vue` 的：

```text
form
resetForm
populateForm
handleSubmit payload
template 表单区域
```

中加入 `sold_count_offset`，不应该新建商品编辑页面。

### User 现有链路

用户端商品卡片现有链路是：

```text
Home.vue / Products.vue
    -> ProductCard.vue
```

商品详情现有链路是：

```text
ProductDetail.vue
    -> productAPI.detail(slug)
```

因此“已售 X 件”应放在：

```text
ProductCard.vue
ProductDetail.vue
```

手机端外层留白和网格密度应放在：

```text
Home.vue
Products.vue
```

不应该新建独立商品卡片组件。

商品图片展示现有链路是：

```text
ProductCard.vue
ProductImageGallery.vue
```

因此商品图比例和裁剪方式必须在现有组件中统一调整，不允许按商品名、slug、图片路径做特殊判断。

页脚版权现有链路是：

```text
Footer.vue
    -> appStore.config.brand.site_name
    -> brandSiteName
    -> i18n footer.rights
```

因此页脚品牌名必须复用现有 `brandSiteName`，不能把 `Toplenged` 直接替换成另一个硬编码品牌名。

## 当前问题

### 手机端商品卡片偏空

当前用户端商品列表在手机端主要表现为：

- 页面容器左右留白偏大。
- 商品卡片是两列布局，但每张卡片内容偏少。
- 手机端标题只显示一行。
- 手机端商品描述被隐藏。
- 卡片内价格、库存、购买信息视觉重量偏弱。

结果是：即使商品数量不少，手机端页面仍然容易显得空、散、小。

### 商品展示图被裁剪

当前商品展示图素材标准为：

```text
1080 x 1080
```

也就是 `1:1` 正方形商品信息图。

但当前前端商品图展示使用的是：

```text
aspect-[4/3] + object-cover
```

这会导致正方形商品图被横向容器强制裁剪，出现展示不完整的问题。

本问题不是某一个商品的问题，而是商品图片展示规范和前端容器比例不一致。

### 前台没有展示已售数量

当前商品卡片和商品详情页主要展示：

- 商品标题
- 分类
- 标签
- 价格
- 库存状态
- 发货方式

但没有展示：

```text
已售 X 件
```

对于售卖型页面来说，销量数字能提升用户对商品热度和可信度的感知。

### 页脚版权品牌名硬编码

当前用户端页脚存在硬编码：

```text
© 2026 Toplenged. All rights reserved.
```

位置：

```text
audit-dujiao-next-user/src/components/Footer.vue
```

这个属于硬编码问题。

同一个 `Footer.vue` 文件里已经存在：

```text
brandSiteName
```

并且它已经从后台站点配置读取：

```text
appStore.config.brand.site_name
```

因此正确修复方式不是把 `Toplenged` 写死替换成 `888tech`，而是：

- 品牌名使用现有 `brandSiteName`。
- 年份使用当前年份或组件内 `currentYear` 计算值。
- 权利声明使用现有 i18n 文案 `footer.rights`。

目标效果：

```text
© 当前年份 后台配置的站点名. All rights reserved.
```

这样测试环境、正式环境、以后换品牌时都不需要改代码。

## 明确边界

本次计划不修改真实交易逻辑。

不做：

- 不按商品名写特殊逻辑。
- 不按 slug 写特殊逻辑。
- 不按图片文件名或路径写特殊逻辑。
- 不修改订单创建逻辑。
- 不修改支付逻辑。
- 不修改优惠券逻辑。
- 不修改卡密交付逻辑。
- 不修改真实库存扣减逻辑。
- 不修改真实销售统计口径。
- 不把展示销量当成财务统计。

展示销量只用于前台 UI 展示，不参与：

- 订单统计
- 财务统计
- 佣金计算
- 库存计算
- 卡密发放
- 商品是否售罄判断

## 数据来源设计

### 真实已售数量

真实已售数量根据商品发货类型计算。

#### 自动发货商品

自动发货商品通常使用卡密库存。

真实已售数量来自：

```text
card_secrets
```

统计条件：

```text
product_id = 当前商品 ID
status = used
```

如果存在 SKU，需要按商品下所有 SKU 汇总。

#### 手动发货商品

手动发货商品真实已售数量来自：

```text
product_skus.manual_stock_sold
```

如果商品存在多个 SKU，需要汇总该商品下所有启用或存在的 SKU：

```text
SUM(product_skus.manual_stock_sold)
```

说明：当前公开商品详情已经返回 SKU 的 `manual_stock_sold`，但没有返回商品级最终展示销量 `sold_count`。

#### 上游商品

上游商品如果没有本地明确已售字段，建议先按本地订单或 SKU 统计逻辑处理。

第一阶段可以采用保守策略：

```text
上游商品真实已售 = 0
最终展示销量 = sold_count_offset
```

如果后续需要精确统计，再单独增加上游订单销量统计。

## 新增字段设计

建议在 `products` 表新增字段：

```text
sold_count_offset INTEGER NOT NULL DEFAULT 0
```

含义：

```text
展示销量手动加成
```

示例：

```text
真实已售 = 4
后台填写 sold_count_offset = 100
前台展示 sold_count = 104
```

字段规则：

- 默认值为 `0`。
- 不允许负数。
- 只用于展示。
- 不参与库存、订单、支付、佣金计算。

## API 修改计划

### 模型层

修改：

```text
audit-dujiao-next-api/internal/models/product.go
```

在 `Product` 模型中增加：

```go
SoldCountOffset int `gorm:"not null;default:0" json:"sold_count_offset"`
```

需要确认自动迁移是否会给 SQLite 添加该字段。

### DTO 层

修改：

```text
audit-dujiao-next-api/internal/dto/product.go
```

在公开商品响应中增加：

```go
RealSoldCount   int64 `json:"real_sold_count"`
SoldCountOffset int   `json:"sold_count_offset"`
SoldCount       int64 `json:"sold_count"`
```

说明：

- `real_sold_count`：真实已售数量。
- `sold_count_offset`：后台手动加成。
- `sold_count`：前台最终展示数量。

前台只需要使用：

```text
sold_count
```

### 商品公开接口计算

修改：

```text
audit-dujiao-next-api/internal/http/handlers/public/public.go
```

在商品公开响应装饰逻辑中计算：

```text
sold_count = real_sold_count + sold_count_offset
```

其中：

- 自动发货商品优先使用卡密 `used` 数量。
- 手动发货商品使用 SKU `manual_stock_sold` 汇总。
- 如果真实已售无法计算，默认 `0`。
- 如果 `sold_count_offset < 0`，归一化为 `0`。

注意：当前后端内部已经存在自动发货库存统计概念，例如 `AutoStockSold`，但公开商品 DTO 尚未输出最终展示销量。实现时可以复用现有库存装饰逻辑，避免重复查询或重复统计。

具体建议：

- `ProductService.ApplyAutoStockCounts` 已经把 `card_secrets` 的 `used` 状态汇总到 `Product.AutoStockSold` 和 `ProductSKU.AutoStockSold`。
- 自动发货商品的 `real_sold_count` 优先读取 `product.AutoStockSold`。
- 手动发货商品的 `real_sold_count` 优先汇总 active SKU 的 `ManualStockSold`；如果没有 SKU，再回退到商品级 `ManualStockSold`。
- 不要在前端自己从 SKU 猜最终销量，统一由 API 返回 `sold_count`。

### 商品管理接口

需要确认 admin 商品新增/编辑接口当前使用的请求 DTO。

预计修改位置：

```text
audit-dujiao-next-api/internal/http/handlers/admin/admin_product.go
audit-dujiao-next-api/internal/service/product_service.go
```

新增字段：

```text
sold_count_offset
```

保存规则：

- 未传时默认 `0`。
- 传空或非法值时按 `0` 处理，或返回参数错误。
- 不允许负数。

## Admin 修改计划

目标：后台商品新增/编辑页允许管理员设置销量加成。

预计修改：

```text
audit-dujiao-next-admin/src/views/admin/components/ProductEditModal.vue
```

`Products.vue` 是商品列表页，实际新增/编辑商品表单在 `ProductEditModal.vue`。

现有 `ProductEditModal.vue` 已经有：

```text
form
resetForm
populateForm
handleSubmit
```

实现时应在这些现有结构中增加 `sold_count_offset`，不要新建表单组件。

新增表单项：

```text
销量加成
```

建议说明文案：

```text
前台展示销量 = 真实已售数量 + 销量加成；该数值仅用于展示，不影响库存和订单统计。
```

表单规则：

- 数字输入框。
- 最小值 `0`。
- 默认 `0`。
- 保存时提交 `sold_count_offset`。
- 编辑商品时回显当前 `sold_count_offset`。

管理后台商品列表是否展示该字段可以分两步：

第一阶段：

- 只在编辑表单中展示和修改。

第二阶段：

- 商品列表中增加一列：

```text
展示销量
```

展示：

```text
真实已售 + 加成 = 最终展示
```

## 用户端修改计划

### 商品图片完整展示

当前商品图标准：

```text
1080 x 1080
```

因此用户端商品主图展示规范应统一为：

```text
aspect-square + object-contain
```

而不是：

```text
aspect-[4/3] + object-cover
```

修改：

```text
audit-dujiao-next-user/src/components/ProductCard.vue
audit-dujiao-next-user/src/components/product/ProductImageGallery.vue
```

具体要求：

- 商品卡片主图容器使用正方形比例。
- 商品详情页主图容器使用正方形比例。
- 图片使用 `object-contain`，保证完整显示。
- 容器保留统一背景色，避免图片未铺满时显得突兀。
- 不能根据商品标题、slug、图片路径写条件判断。
- 不新建专门给某个商品使用的图片组件。

注意：

- 如果未来上传横图，`object-contain` 会出现上下或左右留白。
- 由于当前商品展示图统一是 1080x1080，第一阶段按正方图规范处理。
- 如果未来需要同时支持横图和正方图，应设计通用的图片展示策略，而不是按单个商品硬编码。

### 商品卡片显示已售

修改：

```text
audit-dujiao-next-user/src/components/ProductCard.vue
```

显示逻辑：

```text
如果 product.sold_count > 0，则显示 已售 X 件
```

建议展示位置：

- 手机端：价格上方或价格右侧的小标签。
- 桌面端：价格区域旁边或库存标签旁边。

示例：

```text
已售 104
```

注意：

- `sold_count = 0` 时可以不显示，避免页面显得冷清。
- 不要显示 `real_sold_count` 和 `sold_count_offset`。
- 前台只显示最终的 `sold_count`。

### 商品详情页显示已售

修改：

```text
audit-dujiao-next-user/src/views/ProductDetail.vue
```

建议在标题下方的标签区增加：

```text
已售 X 件
```

位置建议：

```text
游客可购 / 自动交付 / 库存紧张 / 已售 X 件
```

显示规则：

```text
product.sold_count > 0
```

### 类型定义

如果用户端后续新增统一商品类型定义，需要增加：

```text
sold_count?: number
real_sold_count?: number
sold_count_offset?: number
```

预计位置：

```text
audit-dujiao-next-user/src/api/types.ts
```

但当前用户端商品数据主要以 `any` 传递，这一项不是必改项。

### 页脚版权配置化

修改：

```text
audit-dujiao-next-user/src/components/Footer.vue
```

当前问题：

```text
&copy; 2026 Toplenged. All rights reserved.
```

修复要求：

- 删除硬编码品牌名 `Toplenged`。
- 删除硬编码年份 `2026`。
- 品牌名使用现有 `brandSiteName`。
- 年份使用当前年份，例如组件内 computed 或常量 `currentYear`。
- `All rights reserved` 使用现有 `t('footer.rights')`。

禁止：

- 不允许写死为 `888tech`。
- 不允许按域名判断显示哪个品牌。
- 不允许新增独立配置文件绕过后台 `site_config.brand.site_name`。
- 不允许把版权文案散落成硬编码字符串。

建议实现形态：

```text
© {{ currentYear }} {{ brandSiteName }}. {{ t('footer.rights') }}.
```

## 手机端商品卡片优化计划

### 商品中心页面

修改：

```text
audit-dujiao-next-user/src/views/Products.vue
```

当前手机端主要是：

```text
container px-4
grid grid-cols-2 gap-3
```

建议调整：

```text
手机端容器左右间距从 px-4 降到 px-2.5 或 px-3
手机端商品网格 gap 从 gap-3 调整为 gap-2.5 或 gap-2
```

目标：

- 卡片更接近屏幕边缘。
- 两列商品更大。
- 页面看起来更满。

### 首页精选商品

修改：

```text
audit-dujiao-next-user/src/views/Home.vue
```

首页精选商品同样使用 `ProductCard.vue`，需要同步优化外层网格。

建议：

```text
移动端精选商品区域左右留白减少
grid gap 减小
```

### 商品卡片组件

修改：

```text
audit-dujiao-next-user/src/components/ProductCard.vue
```

建议调整：

- 手机端标题从一行改为两行。
- 手机端显示一行简短描述。
- 手机端卡片内边距保持紧凑但不要太空。
- 手机端价格区域更突出。
- 手机端显示“已售 X 件”。
- 手机端商品图按正方形完整展示，不裁剪核心信息。

当前：

```text
标题 line-clamp-1
描述 hidden md:block
卡片 p-3
```

建议：

```text
标题 line-clamp-2
描述手机端 line-clamp-1
价格和已售信息放在底部同一视觉区
```

示例结构：

```text
商品图
分类 / 标签
标题（最多两行）
简短描述（一行）
游客可购 / 自动交付
价格        已售 104
购物车按钮
```

## 推荐实施顺序

### 第一步：API 字段与计算

- 增加 `products.sold_count_offset`。
- 计算真实已售。
- 返回 `real_sold_count`、`sold_count_offset`、`sold_count`。
- 确认商品列表和详情接口都返回新字段。

### 第二步：Admin 可编辑销量加成

- 商品编辑表单增加“销量加成”。
- 保存和回显 `sold_count_offset`。
- 确认不会影响其他商品字段保存。

### 第三步：用户端展示已售

- 商品卡片显示 `sold_count`。
- 商品详情页显示 `sold_count`。
- `sold_count = 0` 时隐藏。

### 第四步：商品图完整展示与手机端卡片视觉优化

- 商品卡片主图改为正方形完整展示。
- 商品详情页主图改为正方形完整展示。
- 调整商品中心手机端容器边距。
- 调整首页精选商品手机端网格间距。
- 调整 `ProductCard.vue` 手机端排版密度。

### 第五步：页脚版权配置化

- 修改 `Footer.vue` 的版权行。
- 确认品牌名来自 `brandSiteName`。
- 确认权利声明来自 `footer.rights`。
- 搜索用户端构建产物，确认页脚不再出现硬编码 `© 2026 Toplenged. All rights reserved.`。

### 第六步：移动端实际截图确认

必须用手机宽度检查：

```text
375 x 812
390 x 844
430 x 932
```

检查重点：

- 左右留白是否减少。
- 商品卡片是否更大。
- 商品 1080x1080 展示图是否完整显示。
- 商品详情页主图是否完整显示。
- 标题是否不挤压。
- 价格和已售信息是否清楚。
- 卡片高度是否一致。
- 两列商品是否没有文字重叠。

## 测试计划

### API 测试

检查商品列表：

```text
GET /api/v1/public/products
```

检查商品详情：

```text
GET /api/v1/public/products/:slug
```

确认返回：

```json
{
  "real_sold_count": 4,
  "sold_count_offset": 100,
  "sold_count": 104
}
```

### Admin 测试

测试流程：

1. 打开商品编辑页。
2. 修改“销量加成”为 `100`。
3. 保存商品。
4. 重新打开编辑页，确认回显 `100`。
5. 调用公开商品接口，确认 `sold_count_offset = 100`。

异常测试：

- 填 `0`。
- 填空。
- 填负数。
- 填非数字。

### 用户端测试

检查页面：

```text
/
/products
/products/:slug
```

确认：

- 商品卡片显示“已售 X 件”。
- 商品详情页显示“已售 X 件”。
- 商品卡片图片完整显示，没有裁掉 1080x1080 图里的文字和边缘信息。
- 商品详情页主图完整显示，没有裁掉 1080x1080 图里的文字和边缘信息。
- 页脚版权显示后台配置的站点名，不显示硬编码 `Toplenged`。
- 页脚版权年份不是写死的旧年份。
- `sold_count = 0` 时不显示。
- 手机端卡片比原来更充实。
- 桌面端布局不被破坏。

## 部署计划

如果要部署到当前生产服务器，必须只操作：

```text
instance-20260607-114457
34.59.247.41
```

用户端部署目标应为当前正式前端项目：

```text
Vercel project-66n51
888tech.club / www.888tech.club
```

部署前必须执行主机保护：

```bash
hostname
```

必须返回：

```text
instance-20260607-114457
```

否则停止。

推荐部署顺序：

1. 部署 API。
2. 确认数据库字段已存在。
3. 确认公开商品接口返回 `sold_count`。
4. 部署 admin。
5. 在 admin 中设置销量加成。
6. 部署用户端。
7. 检查首页、商品中心、商品详情页。

## 风险评估

风险等级：中等偏低。

低风险原因：

- 不改真实订单流程。
- 不改支付。
- 不改卡密交付。
- 不改库存扣减。
- 新字段默认 `0`，不影响旧商品。

主要风险：

- 数据库新增字段需要确认 SQLite 自动迁移是否成功。
- 后端计算真实已售时要避免重复统计 SKU 与商品级卡密。
- Admin 保存商品时不能漏掉新字段。
- 商品图片从 `object-cover` 改为 `object-contain` 后，需要确认不同背景色下视觉是否自然。
- 前端手机端样式调整要防止文字溢出和卡片高度混乱。
- 页脚版权必须接后台品牌配置，不能为了当前生产域名临时写死 `888tech`。

## 最终建议

建议按完整小功能处理，不要只在前端写死“已售数量”。

正确路径是：

```text
真实已售 + 后台手动加成 -> API sold_count -> 用户端展示
```

这样后续运营想调整展示销量，只需要在 admin 修改“销量加成”，不会影响真实订单、库存和卡密。
