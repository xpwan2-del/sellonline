---
outline: deep
---

# Notification Center

> Updated: 2026-03-28

The Dujiao-Next Notification Center supports automatic admin notifications via Email and Telegram when key business events occur.

---

## 1. Feature Overview

| Feature | Description |
|---------|-------------|
| Multi-channel notifications | Supports Email and Telegram notification channels |
| Event-driven | Order payment, recharge success, pending delivery reminders, exception alerts, etc. |
| Custom templates | Each event type supports custom notification title and content |
| Multi-language | Templates support Simplified Chinese, Traditional Chinese, and English |
| Deduplication | Prevents duplicate notifications for the same event within a short time window |
| Low-stock alerts | Automatically detects low product inventory and sends alerts |

---

## 2. Configuration Entry

Admin Panel > Settings > Notification Center

---

## 3. Notification Channels

### 3.1 Email Notifications

| Field | Description |
|-------|-------------|
| Enabled | Toggle for email notifications |
| Recipient list | Email addresses to receive notifications (multiple supported) |

> Email notifications depend on SMTP configuration. Please complete the mail server setup in Admin Panel > Settings > SMTP Email first.

### 3.2 Telegram Notifications

| Field | Description |
|-------|-------------|
| Enabled | Toggle for Telegram notifications |
| Chat ID list | Telegram Chat IDs to receive notifications (multiple supported) |

> A Telegram Bot Token must be configured in the system. You can use the Bot Token from Admin Panel > Settings > Telegram Login, or the channel client Token from the Telegram Bot service.

**How to get your Chat ID:**

1. Send a message to your Bot
2. Visit `https://api.telegram.org/bot{YOUR_BOT_TOKEN}/getUpdates`
3. Find the `chat.id` field in the response

---

## 4. Notification Events

### 4.1 Event Types

| Event | Identifier | Description |
|-------|------------|-------------|
| Wallet recharge success | `wallet_recharge_success` | Notifies when a user completes a wallet recharge |
| Order payment success | `order_paid_success` | Notifies when a user completes an order payment |
| Manual fulfillment pending | `manual_fulfillment_pending` | Alerts admin when a manual-delivery order is paid |
| Exception alert | `exception_alert` | Low stock, order backlog, and other anomalies |

Each event can be enabled or disabled independently.

### 4.2 Exception Alert Details

The system periodically checks the following metrics and triggers alerts when thresholds are exceeded:

| Alert Type | Description |
|------------|-------------|
| Product sold out | A product has 0 available stock |
| Low stock | Product stock is below the warning threshold |
| Pending payment backlog | Abnormal number of pending-payment orders |
| Payment failures | Abnormal number of payment failures |

Alert notifications include the list of affected products and specific data.

---

## 5. Custom Templates

Each notification event supports template customization in three languages (Simplified Chinese / Traditional Chinese / English).

### 5.1 Template Fields

Each template contains:

- **Title**: The notification title/subject
- **Content**: The detailed notification body

### 5.2 Template Variables

<div v-pre>

Use `{{variable_name}}` in templates to insert dynamic data:

**Wallet recharge success:**

| Variable | Description |
|----------|-------------|
| `{{customer_label}}` | Customer name and email |
| `{{customer_email}}` | Customer email |
| `{{recharge_no}}` | Recharge order number |
| `{{amount}}` | Recharge amount |
| `{{currency}}` | Currency |
| `{{payment_channel}}` | Payment channel |

**Order payment success:**

| Variable | Description |
|----------|-------------|
| `{{customer_label}}` | Customer name and email |
| `{{customer_email}}` | Customer email |
| `{{order_no}}` | Order number |
| `{{amount}}` | Order amount |
| `{{currency}}` | Currency |
| `{{payment_channel}}` | Payment channel |
| `{{items_summary}}` | Product list (with SKU details) |
| `{{fulfillment_items_summary}}` | Products pending manual fulfillment |
| `{{delivery_summary}}` | Delivery method summary |

**Manual fulfillment pending:**

| Variable | Description |
|----------|-------------|
| `{{customer_label}}` | Customer name and email |
| `{{customer_email}}` | Customer email |
| `{{order_no}}` | Order number |
| `{{order_status}}` | Current order status |
| `{{fulfillment_items_summary}}` | Products pending fulfillment |
| `{{delivery_summary}}` | Delivery information |

**Exception alert:**

| Variable | Description |
|----------|-------------|
| `{{alert_type}}` | Alert type |
| `{{alert_level}}` | Alert level (error / warning) |
| `{{alert_value}}` | Current value |
| `{{alert_threshold}}` | Threshold |
| `{{message}}` | Alert message |
| `{{affected_items_summary}}` | Affected product details |
| `{{affected_items_count}}` | Number of affected items |

</div>

---

## 6. Advanced Settings

| Field | Default | Description |
|-------|---------|-------------|
| Default language | `zh-CN` | Default language for notification templates |
| Deduplication interval | 300 seconds | Same event will not be sent again within this window (30--86400 seconds) |
| Low-stock alert interval | 1800 seconds | Minimum interval between low-stock alerts (60--604800 seconds) |
| Ignored products | None | Product IDs excluded from low-stock alert checks |

---

## 7. Notification Logs

Notification delivery history is available in the admin panel, including:

- Event type
- Delivery channel (Email / Telegram)
- Delivery status (Success / Failed)
- Delivery time

A test notification can be sent to verify that the configuration is correct.
