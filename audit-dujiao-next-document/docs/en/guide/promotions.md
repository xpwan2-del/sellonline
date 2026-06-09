---
outline: deep
---

# Coupons & Activity Pricing

> Updated: 2026-03-28

Dujiao-Next provides two promotional mechanisms: **Coupons** (manually entered by users) and **Activity Pricing** (automatically applied when conditions are met). Both can be stacked with membership level discounts.

---

## 1. Discount Application Order

When a user places an order, the system calculates the final price in the following order:

```
Base price (or SKU price)
    |
Membership level discount (auto-applied)
    |
Activity pricing (auto-applied when conditions are met)
    |
Coupon discount (manually applied by user)
    |
Final payment amount
```

> Activity pricing and membership pricing use whichever is lower; coupons are then applied on top of the activity price.

---

## 2. Coupons

### 2.1 Coupon Types

| Type | Identifier | Calculation |
|------|------------|-------------|
| Fixed amount discount | `fixed` | Deducts a fixed amount from the order |
| Percentage discount | `percent` | Deducts a percentage of the order amount |

### 2.2 Creating Coupons

Create coupons from the "Coupons" page in the admin panel. Configuration options include:

| Field | Description |
|-------|-------------|
| Coupon Code | Unique identifier that users enter to apply the coupon |
| Discount Type | `fixed` (fixed amount) or `percent` (percentage) |
| Discount Value | The fixed amount or discount percentage |
| Minimum Spend | The subtotal of eligible products must reach this amount |
| Maximum Discount | Upper limit on discount amount for percentage-type coupons |
| Total Usage Limit | Maximum number of uses site-wide; `0` for unlimited |
| Per-User Limit | Maximum number of uses per user; `0` for unlimited |
| Applicable Products | Specify which products the coupon applies to (all or selected products) |
| Validity Period | Start time and end time |
| Enabled | Toggle on/off |

### 2.3 Usage Rules

When a user enters a coupon code at checkout, the system validates in this order:

1. Does the coupon code exist and is it enabled?
2. Is the current time within the validity period (start time <= now <= end time)?
3. Has the site-wide usage limit been reached?
4. Has the current user's usage limit been reached (registered users only)?
5. Does the subtotal of eligible products meet the minimum spend requirement?

**Discount Calculation:**

- **Fixed amount**: Deducts the specified amount directly
- **Percentage**: Calculates the percentage of the eligible products' subtotal, capped by the "Maximum Discount" limit
- The discount amount will never exceed the subtotal of eligible products

### 2.4 Applicable Scope

The coupon's "Applicable Products" setting determines which products participate in the discount calculation:

- Only the subtotal of in-scope products is used to evaluate the "Minimum Spend" requirement
- Only in-scope product amounts are included in the discount calculation
- Products outside the scope in the same order are unaffected

---

## 3. Activity Pricing

### 3.1 Activity Pricing Types

| Type | Identifier | Calculation |
|------|------------|-------------|
| Percentage discount | `percent` | Unit price = original price x (100 - value) / 100 |
| Fixed amount discount | `fixed` | Unit price = original price - value (minimum 0) |
| Special price | `special_price` | Unit price = specified value |

### 3.2 Creating Activity Pricing

Create activity pricing rules from the "Activity Pricing" page in the admin panel. Configuration options include:

| Field | Description |
|-------|-------------|
| Name | Activity name (supports multiple languages) |
| Applicable Products | Specify which products the rule applies to |
| Discount Type | `percent` / `fixed` / `special_price` |
| Discount Value | The numeric value for the corresponding type |
| Threshold Amount | The purchase subtotal (unit price x quantity) must reach this amount for the rule to take effect |
| Validity Period | Start time and end time |
| Enabled | Toggle on/off |

### 3.3 Tiered Discounts

Multiple activity rules can be configured for the same product to create tiered discounts:

```
Rule A: Threshold $50, discount 1%
Rule B: Threshold $150, discount 2%
Rule C: Threshold $500, discount 5%
```

The system matches thresholds from highest to lowest, applying the highest tier whose condition is met:

- Purchase subtotal $49 -> No discount
- Purchase subtotal $100 -> Matches Rule A (1%)
- Purchase subtotal $200 -> Matches Rule B (2%)
- Purchase subtotal $600 -> Matches Rule C (5%)

### 3.4 Application Logic

- Activity pricing is matched on a **per-product basis**; each product is calculated independently
- The system displays activity rules on the product detail page, even if the current quantity does not meet the threshold
- Price precision is two decimal places
- Activity pricing rules of the same type do not stack; the best matching rule is used for each product

---

## 4. Best Practices

- Coupons are well suited for flash sales, new user acquisition, and targeted product promotions
- Activity pricing is ideal for tiered pricing and bulk purchase discounts
- Both can be used simultaneously: activity pricing automatically reduces the unit price, and coupons provide an additional reduction on top
- It is recommended to periodically clean up expired coupons and activity pricing rules to keep lists manageable
