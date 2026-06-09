---
outline: deep
---

# Membership Levels

> Updated: 2026-03-28

Dujiao-Next supports a membership level system that provides exclusive pricing and discounts for users at different spending tiers. Membership levels can be upgraded automatically (based on cumulative top-ups/spending) or assigned manually by administrators.

---

## 1. Feature Overview

| Feature | Description |
|---------|-------------|
| Level Management | Create multiple levels with discount rates and upgrade thresholds |
| Level-Exclusive Pricing | Set member-only prices for specific products/SKUs |
| Auto Upgrade | Automatically upgrade when cumulative top-ups or spending reach the threshold |
| Manual Assignment | Administrators can directly set a user's level |
| Level Discount | A global discount rate automatically applied to all products |

---

## 2. Level Configuration

### 2.1 Creating Levels

Create levels from the "Membership Levels" page in the admin panel. Each level includes:

| Field | Description |
|-------|-------------|
| Name | Level name (supports multiple languages) |
| Slug | Unique identifier, e.g., `default`, `silver`, `gold`, `diamond` |
| Icon | Emoji or icon URL |
| Discount Rate | A value from 0 to 100. 100 = no discount, 90 = 10% off, 80 = 20% off |
| Top-Up Threshold | Cumulative top-up amount required for auto upgrade (0 = top-ups not considered) |
| Spending Threshold | Cumulative spending amount required for auto upgrade (0 = spending not considered) |
| Sort Weight | Higher values indicate higher levels |
| Default Level | The level automatically assigned to new users (only one allowed) |
| Enabled | Toggle on/off |

### 2.2 Level Examples

```
Default Member (default)   Discount Rate 100  Sort 0   -> No discount
Silver Member  (silver)    Discount Rate 95   Sort 10  -> Top-up >= 500 or Spending >= 1000
Gold Member    (gold)      Discount Rate 90   Sort 20  -> Top-up >= 2000 or Spending >= 5000
Diamond Member (diamond)   Discount Rate 85   Sort 30  -> Top-up >= 5000 or Spending >= 10000
```

---

## 3. Price Calculation

Member prices are determined by the following priority:

### 3.1 Priority

1. **SKU-Level Exclusive Price**: A member price set for a specific SKU (highest priority)
2. **Product-Level Exclusive Price**: A member price set for the product
3. **Global Discount Rate**: Calculated using the level's discount rate -> `original price x discount rate / 100`

> Member pricing only takes effect when the member price is lower than the original price; member prices will never exceed the original price.

### 3.2 Setting Exclusive Prices

From the "Membership Level Pricing" page in the admin panel:

- Select a membership level and product
- Set a unified member price for the product
- Or set individual member prices for specific SKUs under the product
- SKU-level settings take priority over product-level settings

---

## 4. Auto Upgrade

### 4.1 Upgrade Conditions

A user is upgraded when **either** of the following conditions is met:

- Cumulative top-up amount >= the level's "Top-Up Threshold" (effective when threshold > 0)
- Cumulative spending amount >= the level's "Spending Threshold" (effective when threshold > 0)

### 4.2 Upgrade Timing

The system automatically checks and upgrades after the following operations:

- Successful wallet top-up
- Successful order payment

### 4.3 Upgrade Rules

- **Upgrade only, never downgrade**: Levels only go up and will not decrease due to subsequent spending changes
- **Skip-level upgrades**: If the cumulative amount directly meets a higher level's threshold, intermediate levels are skipped
- The system matches from highest to lowest sort weight, selecting the highest level whose conditions are met

---

## 5. Admin Operations

### 5.1 Manual Level Assignment

In the admin panel under "User Management", administrators can directly set a user's membership level without threshold restrictions.

### 5.2 Batch Backfill

For existing users who were registered before the system upgrade, use the "Batch Backfill Default Level" feature to assign the default level to all users who have not yet been assigned a level.

---

## 6. Notes

- Each system can only have one default level
- The default level cannot be deleted
- Level slugs must be unique and cannot be duplicated
- A discount rate of 100 or above means the level has no discount
- A threshold of 0 means that condition is not used for upgrade evaluation
- Membership discounts are applied before activity pricing; both can stack
