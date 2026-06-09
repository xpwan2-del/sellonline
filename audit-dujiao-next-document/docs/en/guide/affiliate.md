---
outline: deep
---

# Affiliate Program

> Updated: 2026-03-28

Dujiao-Next includes a built-in affiliate program that allows users to invite others to purchase through referral links and earn commissions. The system supports automatic attribution, commission calculation, a confirmation period mechanism, and withdrawal review.

---

## 1. Feature Overview

| Feature | Description |
|---------|-------------|
| Affiliate Activation | Users apply to become affiliates and receive a unique referral code |
| Referral Links | Track visitor sources via links or referral codes |
| Auto Attribution | Orders placed within a 30-day window are automatically attributed to the affiliate |
| Commission Calculation | Commission records are automatically generated after order payment |
| Confirmation Period | Commissions become withdrawable only after the confirmation period |
| Withdrawal Management | Users apply for withdrawals; administrators review them |
| Affiliate Dashboard | View clicks, orders, commissions, and other statistics |

---

## 2. Becoming an Affiliate

### 2.1 User Side

1. Navigate to the "Affiliate Center" on the storefront
2. Click "Become an Affiliate"
3. The system automatically generates an 8-character referral code (uppercase letters + digits)

> The affiliate feature must be enabled by the administrator in the admin panel before users can activate it.

### 2.2 Admin Configuration

Configure under "Settings -> Affiliate Settings" in the admin panel:

| Field | Description |
|-------|-------------|
| Enabled | Master toggle for the affiliate feature |
| Commission Rate | Commission percentage of the order amount |
| Confirmation Days | Waiting period (in days) from "Pending Confirmation" to "Available for Withdrawal" (0 = immediately available) |
| Minimum Withdrawal Amount | Minimum threshold for withdrawal requests |
| Withdrawal Channels | List of supported withdrawal methods |

---

## 3. Referral Links

The affiliate's unique referral link format:

```
https://your-domain/?aff=REFERRAL_CODE
```

For example, if the referral code is `AB3CD5EF`, the referral link would be:

```
https://your-domain/?aff=AB3CD5EF
```

Users can also share their referral code manually; other users simply need to visit the site with the `?aff=REFERRAL_CODE` parameter.

---

## 4. Visitor Attribution

### 4.1 Attribution Rules

The system associates affiliates with ordering users through the following methods:

1. **Browser fingerprint matching**: After a visitor clicks a referral link, any orders placed in the same browser within 30 days are automatically attributed
2. **Referral code matching**: Direct attribution via the `aff` parameter in the URL

> Browser fingerprint matching takes priority (using the most recent valid click), followed by referral code matching.

### 4.2 Deduplication

- Repeated clicks by the same visitor on the same page within 10 minutes are not recorded
- Users cannot be their own affiliate (self-referral is invalid)

---

## 5. Commission Calculation

### 5.1 Calculation Rules

Commissions are automatically generated after successful order payment:

```
Commission base = Sum(product totals - coupon discounts)  // Only products with affiliate enabled
Commission amount = Commission base x Commission rate / 100
```

- Only products with the "Allow Affiliate" option enabled participate in commission calculation
- Commissions are generated at the order level; each order produces at most one commission record
- Duplicate commissions are not generated for the same order

### 5.2 Commission Statuses

| Status | Identifier | Description |
|--------|------------|-------------|
| Pending Confirmation | `pending_confirm` | Waiting for the confirmation period to end |
| Available | `available` | Confirmation period has passed; eligible for withdrawal |
| Withdrawn | `withdrawn` | Included in a withdrawal request |
| Rejected | `rejected` | Invalidated due to order cancellation/refund |

### 5.3 Status Flow

```
Order paid -> Pending Confirmation
                | (confirmation period expires)
              Available
                | (user requests withdrawal)
              Withdrawn

Order cancelled/refunded -> Rejected (skips commissions already in withdrawal process)
```

---

## 6. Withdrawal Process

### 6.1 User Application

1. View the available balance in the "Affiliate Center"
2. Click "Request Withdrawal"
3. Enter the withdrawal amount (must be >= minimum withdrawal amount)
4. Select a withdrawal channel and provide payment details
5. Submit and wait for administrator review

> When a request is submitted, the system locks the corresponding commission amount to the withdrawal request. If the last commission record exceeds the required amount, the system automatically splits it.

### 6.2 Admin Review

Process requests under "Affiliate Management -> Withdrawal Management" in the admin panel:

- **Approve**: Marks commissions as "Withdrawn"; the administrator completes the payment offline
- **Reject**: Provide a rejection reason; locked commissions are released back to "Available" status

---

## 7. Affiliate Dashboard

Users can view the following in the "Affiliate Center" on the storefront:

| Data | Description |
|------|-------------|
| Referral Code | Personal unique referral code |
| Referral Link | Complete referral URL |
| Total Clicks | Number of times the referral link has been clicked |
| Valid Orders | Number of completed orders generated through referrals |
| Conversion Rate | Valid orders / Total clicks x 100% |
| Pending Commission | Total commission amount still in the confirmation period |
| Available Commission | Total confirmed commission amount eligible for withdrawal |
| Withdrawn Commission | Total commission amount withdrawn historically |

---

## 8. Order Refunds and Commissions

- **Order Cancellation**: Associated commissions are automatically marked as "Rejected"
- **Partial Refund**: Commissions are proportionally reduced based on the refund ratio
  - Refund ratio = Refund amount / Remaining refundable amount
  - Commission base and commission amount are reduced proportionally
  - If the commission is reduced to 0 or below, it is automatically marked as "Rejected"
- Commissions already in the withdrawal process are not affected by order cancellations

---

## 9. Notes

- The affiliate feature must be enabled by the administrator before users can use it
- Referral codes are 8-character combinations of uppercase letters and digits (excluding easily confused characters I and O)
- The attribution window is 30 days; click records beyond this period are no longer used for attribution
- The commission base deducts coupon discounts but does not deduct wallet payment portions
- Products must have the "Allow Affiliate" option enabled individually in the admin panel
