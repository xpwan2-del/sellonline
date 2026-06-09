---
outline: deep
---

# 升級與遷移

> 更新時間：2026-03-28

本指南介紹如何將 Dujiao-Next 從舊版本升級到新版本。

---

## 1. 升級前準備

### 1.1 備份資料

**必須在升級前備份以下內容：**

- 資料庫（SQLite 檔案或 PostgreSQL 資料）
- 設定檔（`config.yml`）
- 上傳檔案目錄（`uploads/`）

參考 [備份與還原](/zh-hant/deploy/backup) 指南。

### 1.2 檢視更新日誌

升級前務必閱讀 [更新日誌](/zh-hant/intro/changelog)，了解：

- 新增功能和設定項
- 破壞性變更（Breaking Changes）
- 資料庫結構變更
- 設定檔新增欄位

---

## 2. 手動部署升級

### 2.1 停止服務

```bash
# 停止後端服務
systemctl stop dujiao-next
# 或手動停止程序
```

### 2.2 備份

```bash
# 備份資料庫（SQLite）
cp api/db/dujiao.db api/db/dujiao.db.bak.$(date +%Y%m%d)

# 備份設定
cp api/config.yml api/config.yml.bak

# 備份上傳檔案
cp -r api/uploads api/uploads.bak
```

### 2.3 更新程式碼

```bash
cd dujiao-studio
git pull origin main
```

### 2.4 重新構建

```bash
# 後端
cd api && go build ./cmd/server ./cmd/seed

# 前台
cd user && npm install && npm run build

# 管理後台
cd admin && npm install && npm run build
```

### 2.5 更新設定

對比 `config.yml.example` 檢查是否有新增設定項，按需新增到 `config.yml`。

### 2.6 啟動服務

```bash
systemctl start dujiao-next
```

> 資料庫結構變更會在啟動時由 GORM 自動遷移完成，無需手動執行 SQL。

---

## 3. Docker Compose 升級

### 3.1 備份

```bash
# 備份資料卷
docker compose exec api cp /app/db/dujiao.db /app/db/dujiao.db.bak

# 匯出設定
docker compose cp api:/app/config.yml ./config.yml.bak
```

### 3.2 更新映像

```bash
# 拉取最新映像
docker compose pull

# 或使用指定版本
# 修改 docker-compose.yml 中的 image 版本號
```

### 3.3 重啟服務

```bash
docker compose down
docker compose up -d
```

### 3.4 檢查日誌

```bash
docker compose logs -f api
```

---

## 4. 升級後驗證

升級完成後按以下步驟驗證：

1. **後台登入**：確認管理員可正常登入後台
2. **儀表板**：檢查資料是否正常顯示
3. **商品列表**：確認商品和庫存資料正確
4. **建立測試訂單**：前台下單並完成支付測試
5. **支付回呼**：確認支付回呼正常運作
6. **郵件通知**：確認郵件發送功能正常

---

## 5. 回滾

如果升級後出現問題：

### 5.1 手動部署回滾

```bash
# 停止服務
systemctl stop dujiao-next

# 還原資料庫
cp api/db/dujiao.db.bak.YYYYMMDD api/db/dujiao.db

# 還原設定
cp api/config.yml.bak api/config.yml

# 切換回舊版本程式碼
git checkout <舊版本tag或commit>

# 重新構建並啟動
cd api && go build ./cmd/server
systemctl start dujiao-next
```

### 5.2 Docker 回滾

```bash
docker compose down

# 修改 docker-compose.yml 中的映像版本為舊版本
# 還原資料庫備份

docker compose up -d
```

---

## 6. 注意事項

- 跨多個版本升級時，建議逐版本升級或仔細閱讀每個版本的更新日誌
- 資料庫自動遷移只增加欄位/表，不會刪除已有欄位
- 升級後首次啟動可能稍慢（執行資料庫遷移）
- 設定檔新增欄位通常有預設值，不加也不影響啟動，但建議補全
