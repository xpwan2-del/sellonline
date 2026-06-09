---
outline: deep
---

# Upgrade & Migration

> Updated: 2026-03-28

This guide explains how to upgrade Dujiao-Next from an older version to a newer version.

---

## 1. Pre-Upgrade Preparation

### 1.1 Back Up Your Data

**You must back up the following before upgrading:**

- Database (SQLite file or PostgreSQL data)
- Configuration file (`config.yml`)
- Uploaded files directory (`uploads/`)

Refer to the [Backup & Recovery](/en/deploy/backup) guide.

### 1.2 Review the Changelog

Before upgrading, be sure to read the [Changelog](/intro/changelog) to understand:

- New features and configuration options
- Breaking changes
- Database schema changes
- New configuration fields

---

## 2. Manual Deployment Upgrade

### 2.1 Stop the Service

```bash
# Stop the backend service
systemctl stop dujiao-next
# Or manually stop the process
```

### 2.2 Back Up

```bash
# Back up the database (SQLite)
cp api/db/dujiao.db api/db/dujiao.db.bak.$(date +%Y%m%d)

# Back up the configuration
cp api/config.yml api/config.yml.bak

# Back up uploaded files
cp -r api/uploads api/uploads.bak
```

### 2.3 Update the Code

```bash
cd dujiao-studio
git pull origin main
```

### 2.4 Rebuild

```bash
# Backend
cd api && go build ./cmd/server ./cmd/seed

# Storefront
cd user && npm install && npm run build

# Admin panel
cd admin && npm install && npm run build
```

### 2.5 Update Configuration

Compare with `config.yml.example` to check for new configuration options and add them to `config.yml` as needed.

### 2.6 Start the Service

```bash
systemctl start dujiao-next
```

> Database schema changes are handled automatically by GORM auto-migration on startup -- no manual SQL execution is required.

---

## 3. Docker Compose Upgrade

### 3.1 Back Up

```bash
# Back up data volumes
docker compose exec api cp /app/db/dujiao.db /app/db/dujiao.db.bak

# Export configuration
docker compose cp api:/app/config.yml ./config.yml.bak
```

### 3.2 Update Images

```bash
# Pull the latest images
docker compose pull

# Or use a specific version
# Update the image version in docker-compose.yml
```

### 3.3 Restart Services

```bash
docker compose down
docker compose up -d
```

### 3.4 Check Logs

```bash
docker compose logs -f api
```

---

## 4. Post-Upgrade Verification

After the upgrade is complete, verify with the following steps:

1. **Admin login**: Confirm that administrators can log in to the admin panel
2. **Dashboard**: Check that data displays correctly
3. **Product list**: Verify product and inventory data is accurate
4. **Create a test order**: Place an order on the storefront and complete a payment test
5. **Payment callbacks**: Confirm payment callbacks are working properly
6. **Email notifications**: Confirm email delivery is functioning

---

## 5. Rollback

If issues arise after upgrading:

### 5.1 Manual Deployment Rollback

```bash
# Stop the service
systemctl stop dujiao-next

# Restore the database
cp api/db/dujiao.db.bak.YYYYMMDD api/db/dujiao.db

# Restore the configuration
cp api/config.yml.bak api/config.yml

# Switch back to the previous version
git checkout <previous-version-tag-or-commit>

# Rebuild and start
cd api && go build ./cmd/server
systemctl start dujiao-next
```

### 5.2 Docker Rollback

```bash
docker compose down

# Update the image version in docker-compose.yml to the previous version
# Restore the database backup

docker compose up -d
```

---

## 6. Important Notes

- When upgrading across multiple versions, it is recommended to upgrade one version at a time or carefully review the changelog for each version
- Auto-migration only adds new fields and tables -- it does not remove existing fields
- The first startup after an upgrade may be slower due to database migration
- New configuration fields typically have default values and will not prevent startup if omitted, but it is recommended to add them
