# 單二進制部署（推薦新手）

> 適用人群：完全新手，不想接觸 Docker 多容器編排，希望「一個二進制 + 一個 Redis 容器 + 一個網域」就能跑起來。

## 適用場景對比

| 部署方式 | 上手難度 | 容器數量 | 網域數 |
|---|---|---|---|
| **單二進制（本文）** | 低 | 1（Redis） | 1 |
| Docker Compose | 中 | 4-5 | 1-2 |
| aaPanel 手動部署 | 低-中 | 0（裸跑） | 1-2 |
| 手動原始碼部署 | 高 | 0 | 1-2 |

## 系統要求

- Linux x86_64 或 arm64
- Docker（僅用於跑 Redis 容器；如果你已有 Redis 服務可省略）
- 一個網域 + SSL 憑證（生產部署）
- 至少 512MB 記憶體

## 1. 下載

到 [GitHub Releases](https://github.com/dujiao-next/dujiao-next/releases) 找最新的 `dujiao-all_*.tar.gz`，按系統架構選：

```bash
# 例：Linux amd64
wget https://github.com/dujiao-next/dujiao-next/releases/download/vX.Y.Z/dujiao-all_vX.Y.Z_linux_amd64.tar.gz
tar -xzf dujiao-all_*.tar.gz
cd <解壓縮目錄>
```

## 2. 複製設定

```bash
cp config.yml.example config.yml
```

## 3. 必改欄位

打開 `config.yml`，按下表修改：

| 欄位 | 說明 | 範例值 |
|---|---|---|
| `jwt.secret` | 後台管理員 JWT 金鑰，**必改** | `openssl rand -hex 32` 輸出 |
| `user_jwt.secret` | 使用者 JWT 金鑰，**必改** | 同上，不同值 |
| `web.admin_path` | 後台存取路徑前綴，**強烈建議改** | `/dj-mgmt-7x9k2` |
| `redis.host` / `redis.port` | Redis 位址（分兩個欄位，預設 `127.0.0.1` + `6379`） | `127.0.0.1` + `6379` |
| `database.driver` / `database.dsn` | 資料庫（預設 SQLite 起步） | 見下方 |

### 關於 `web.admin_path`（重要）

預設值 `/admin` 是自動化掃描器的頭號目標。**強烈建議改成不易猜測的字串**：

```yaml
web:
  admin_path: "/dj-mgmt-7x9k2"   # 個人偏好的字串
```

這個路徑只是 SPA 入口的「門牌」，改了它不影響 admin API 介面；API 鑑權由 JWT + 限流保護。改路徑主要是過濾掉自動化掃描的雜訊。

### 關於資料庫

- **SQLite（預設）**：零設定，資料存在 `./db/dujiao.db`，單機夠用。
- **PostgreSQL（生產推薦）**：把 `database.driver` 改為 `postgres`，`database.dsn` 寫連接串。

## 4. 啟動 Redis

本套件附帶最小 Redis 設定：

```bash
docker compose up -d redis
```

如果你已有 Redis（系統服務或其他容器），改 `config.yml` 的 `redis.host` 和 `redis.port` 即可，不需要起新容器。

## 5. 啟動二進制

```bash
./dujiao-server
```

啟動日誌會顯示：
```
🚀 Dujiao-Next API 啟動中
...
Embedded SPAs: admin (/dj-mgmt-7x9k2), user (/)
```

二進制執行時會自動建立：
- `./db/`：SQLite 資料庫
- `./uploads/`：使用者上傳檔案
- `./logs/`：執行日誌

## 6. 存取

- **使用者端**：`http://<your-ip>:8080`
- **管理端**：`http://<your-ip>:8080/<web.admin_path>`（你剛才改的路徑）

首次登入用預設管理員帳號（在 `config.yml` 的 `bootstrap` 段設定）。**登入後立即修改密碼**。

## 7. 反向代理與 HTTPS（生產部署）

把單個網域 `shop.example.com` 轉發到二進制的 8080 連接埠（Nginx 範例）：

```nginx
server {
    listen 443 ssl http2;
    server_name shop.example.com;
    ssl_certificate     /path/to/fullchain.pem;
    ssl_certificate_key /path/to/privkey.pem;

    client_max_body_size 50m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 8. 系統服務（systemd）

先建立執行使用者（如果你打算用專用使用者跑服務）：
```bash
sudo useradd -r -s /sbin/nologin -d /opt/dujiao dujiao
sudo chown -R dujiao:dujiao /opt/dujiao
```

`/etc/systemd/system/dujiao.service`：
```ini
[Unit]
Description=Dujiao-Next Fullstack
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/dujiao
ExecStart=/opt/dujiao/dujiao-server
Restart=on-failure
User=dujiao

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now dujiao
sudo journalctl -u dujiao -f
```

## 9. 升級

1. `systemctl stop dujiao`
2. 備份：`cp -r db uploads /backup/`
3. 下載新版 tar.gz，僅替換 `dujiao-server` 二進制
4. `systemctl start dujiao`

資料庫遷移自動完成。

## 10. 從其他部署方式遷移

### 從 Docker Compose 遷移

1. 在 docker-compose 服務停機
2. 把現有 `db/`、`uploads/` 目錄拷到 fullstack 二進制執行目錄
3. 把 docker-compose 用的 `config.yml` 拷過來
4. 啟動 fullstack 二進制（注意設定 `web.admin_path`）

### 從原始碼手動部署遷移

同上，路徑和設定直接複用。

## 常見問題

### Q：admin SPA 載入報 404

確認 `config.yml` 的 `web.admin_path` 與瀏覽器存取路徑一致；如果改了 `web.admin_path`，必須重啟二進制讓新值生效。

### Q：日誌循環出現「`web.admin_path 仍为默认 /admin`」警告

> 註：日誌訊息為簡體中文，因為後端輸出的告警字串目前未做 i18n。

按 §3 的建議修改 `web.admin_path`，警告會消失。

### Q：fullstack 二進制和現有 Docker 映像可以混用嗎？

可以，但同一個資料庫不要同時被兩套部署連接。建議：要麼走 Docker 映像方案，要麼走單二進制方案，不混用。
