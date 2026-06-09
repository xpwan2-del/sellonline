---
outline: deep
---

# Card Secret Management

> Updated: 2026-03-28

Card secrets are the core of auto-delivery products in Dujiao-Next. After a user completes payment, the system automatically assigns a card secret from inventory and displays it to the user. This guide covers importing, managing, and delivering card secrets.

---

## 1. Basic Concepts

| Concept | Description |
|---------|-------------|
| Card Secret | Digital product content to be delivered (e.g., activation codes, account credentials, CDKeys) |
| Batch | A set of card secrets created by a single import operation, for grouped management |
| Auto Delivery | System automatically assigns and displays card secrets after payment |
| Manual Delivery | Admin manually enters delivery content (does not use card secret inventory) |

---

## 2. Card Secret Statuses

| Status | Identifier | Description |
|--------|------------|-------------|
| Available | `available` | Pending assignment, can be used by new orders |
| Reserved | `reserved` | Linked to an order, awaiting payment completion |
| Used | `used` | Delivered to the user |

### Status Flow

```
Import -> Available
            | User places order
          Reserved
            | Payment succeeds & auto delivery
          Used

Payment timeout / Order canceled -> Released back to Available
```

---

## 3. Importing Card Secrets

### 3.1 Manual Input

From the "Card Secret Management" page in admin:

1. Select a product (must be auto-delivery type)
2. Select a SKU (if the product has multiple SKUs)
3. Enter card secrets in the text box, one per line
4. Optionally fill in a batch name and notes
5. Click "Import"

> The system automatically removes blank lines, trims whitespace, and deduplicates entries.

### 3.2 CSV Import

Bulk import via CSV file is supported:

**CSV Format Requirements:**

- Must contain a `secret` column (header is case-insensitive)
- One card secret per row
- Blank lines and duplicates are automatically skipped
- Supports UTF-8 encoding (BOM header handled automatically)

**CSV Example:**

```csv
secret
ABCD-1234-EFGH-5678
IJKL-9012-MNOP-3456
QRST-7890-UVWX-1234
```

### 3.3 Batch Information

Each import automatically generates a batch record containing:

- Batch number (auto-generated, format: `BATCH-{timestamp}-{random}`)
- Import source (manual / CSV)
- Total card secret count
- Status breakdown

---

## 4. Managing Card Secrets

### 4.1 Viewing Card Secrets

Multi-dimensional filtering is supported:

| Filter | Description |
|--------|-------------|
| Product | Filter by associated product |
| SKU | Filter by SKU variant |
| Batch | Filter by import batch |
| Status | Filter by available / reserved / used |
| Keyword | Fuzzy search by card secret content |
| Batch Number | Search by batch number |

### 4.2 Bulk Operations

- **Bulk Status Change**: Select multiple card secrets and change their status in bulk
- **Bulk Delete**: Delete selected card secrets (not recommended for used ones)
- **Batch Operations**: Apply status changes to an entire batch

### 4.3 Exporting Card Secrets

Two export formats are supported:

| Format | Description |
|--------|-------------|
| TXT | One card secret per line, suitable for simple use |
| CSV | Full fields included (ID, secret, status, product, SKU, order, batch, timestamps, etc.) |

---

## 5. Inventory Statistics

### 5.1 Product Level

The product list and detail pages show:

- **Total Stock**: Total number of card secrets
- **Available Stock**: Number of card secrets with `available` status
- **Reserved**: Number of card secrets linked to pending orders
- **Sold**: Number of card secrets delivered to users

### 5.2 Batch Level

The batch list displays statistics for each batch:

- Batch number and name
- Import source and time
- Count by status
- Associated product and SKU

---

## 6. Auto Delivery Process

After a user's payment succeeds, the system automatically performs the following steps:

1. Confirms the order status is "Paid"
2. Verifies all product items have delivery type `auto`
3. Assigns card secrets by product and SKU:
   - Prioritizes reserved (`reserved`) card secrets
   - Supplements from available (`available`) card secrets if needed
4. Marks assigned card secrets as used (`used`)
5. Creates a delivery record with card secret content displayed to the user
6. Updates order status to "Completed"

> If available card secrets are insufficient, auto delivery will fail. The admin must replenish card secrets and handle the order manually.

---

## 7. Important Notes

- Auto-delivery products must have card secrets imported before they can be sold
- Card secret content is not encrypted at rest during import; ensure proper data security
- Duplicate card secrets are automatically skipped during bulk import
- When a product has multiple SKUs, card secrets are managed independently per SKU
- Regularly check inventory and replenish card secrets for products that are running low
- The dashboard displays alerts when inventory is low
