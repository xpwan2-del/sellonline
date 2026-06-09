---
outline: deep
---

# Wallet System

> Updated: 2026-03-28

Dujiao-Next includes a built-in wallet system. Users can top up their wallet balance via multiple payment methods, pay for orders using their balance, or redeem gift cards to add funds. Administrators can manually adjust user balances or refund orders to wallets.

---

## 1. Feature Overview

| Feature | Description |
|---------|-------------|
| Balance Top-Up | Users top up their wallet through configured payment channels |
| Balance Payment | Use wallet balance to pay for all or part of an order |
| Transaction History | Complete records for every top-up, purchase, and refund |
| Gift Card Redemption | Users enter a gift card code to add funds |
| Admin Adjustment | Administrators can increase or deduct user balances |
| Refund to Wallet | Administrators can refund paid order amounts back to the user's wallet |

---

## 2. User Operations

### 2.1 Top Up Wallet

1. Navigate to the "Wallet" page on the storefront
2. Enter the top-up amount (must be greater than 0, up to two decimal places)
3. Select a payment channel and complete the payment
4. Balance is credited automatically upon successful payment

> Unpaid top-up requests will automatically expire after the timeout period (default 15 minutes, same as order payment timeout).

### 2.2 Pay with Balance

The system automatically detects wallet balance when placing an order:

- **Balance >= Order Amount**: Full payment via balance with no online payment required
- **Balance < Order Amount**: Balance covers a partial amount; the remainder is paid through an online payment channel
- Guest users cannot use wallet payment

### 2.3 Redeem Gift Cards

1. Navigate to the "Wallet" page on the storefront
2. Enter the gift card code
3. The card's face value is automatically credited to the wallet upon successful redemption

> Gift card codes are case-insensitive. Used, disabled, or expired gift cards cannot be redeemed.

### 2.4 View Transaction History

The wallet page displays all historical transactions. Each record includes:

- Transaction type (Top-Up / Order Payment / Refund / Admin Adjustment / Gift Card Redemption)
- Amount and direction (Credit / Debit)
- Balance before and after the transaction
- Associated order number (if applicable)
- Timestamp and remarks

---

## 3. Transaction Types

| Type Identifier | Description | Direction |
|-----------------|-------------|-----------|
| `recharge` | User top-up | Credit |
| `order_pay` | Order balance payment | Debit |
| `admin_refund` | Admin refund to wallet | Credit |
| `admin_adjust` | Admin manual adjustment | Credit/Debit |
| `gift_card` | Gift card redemption | Credit |

---

## 4. Admin Operations

### 4.1 Adjust User Balance

In the admin panel under "User Management", select a user to manually increase or deduct their balance:

- Positive values increase the balance; negative values deduct
- Amount cannot be 0
- The resulting balance cannot be negative

### 4.2 Refund to Wallet

In the admin panel under "Order Management", paid orders support the "Refund to Wallet" action:

- Enter the refund amount (cannot exceed the refundable amount of the order)
- The refund amount is automatically credited to the user's wallet
- Multiple partial refunds are supported, as long as the cumulative total does not exceed the order's paid amount
- Guest orders do not support refund to wallet
- If the order has associated affiliate commissions, the commission will be deducted proportionally upon refund

---

## 5. Gift Card Management

### 5.1 Generate Gift Cards

Gift cards can be generated in bulk from the "Gift Cards" page in the admin panel:

- **Name**: Batch name for management purposes
- **Face Value**: Amount per gift card (greater than 0, up to two decimal places)
- **Quantity**: Up to 10,000 cards per batch
- **Expiration Date**: Optional; if not set, the card never expires

Each generated gift card receives a unique code (format: `GC` + timestamp + sequence number + random string).

### 5.2 Gift Card Statuses

| Status | Description |
|--------|-------------|
| `active` | Available, awaiting redemption |
| `redeemed` | Already redeemed by a user |
| `disabled` | Disabled, cannot be redeemed |

- Administrators can enable/disable gift cards in bulk
- Redeemed gift cards cannot have their status changed or be deleted

### 5.3 Export Gift Cards

Supported export formats:

- **TXT format**: One code per line
- **CSV format**: Includes all fields (ID, batch number, name, code, face value, status, redemption details, etc.)

---

## 6. Notes

- Wallet accounts are automatically created on the user's first visit; no manual activation is needed
- All monetary operations use idempotent references to prevent duplicate charges or credits
- Balance precision is two decimal places
- The default currency follows the system setting (e.g., CNY, USD, etc.)
