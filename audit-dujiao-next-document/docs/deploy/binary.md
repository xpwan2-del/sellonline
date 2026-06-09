# 单二进制部署（推荐小白）

> 适用人群：完全新手，不想接触 Docker 多容器编排，希望"一个二进制 + 一个 Redis 容器 + 一个域名"就能跑起来。

## 适用场景对比

| 部署方式 | 上手难度 | 容器数量 | 域名数 |
|---|---|---|---|
| **单二进制（本文）** | 低 | 1（Redis） | 1 |
| Docker Compose | 中 | 4-5 | 1-2 |
| aaPanel 手动部署 | 低-中 | 0（裸跑） | 1-2 |
| 手动源码部署 | 高 | 0 | 1-2 |

## 系统要求

- Linux x86_64 或 arm64
- Docker（仅用于跑 Redis 容器；如果你已有 Redis 服务可省略）
- 一个域名 + SSL 证书（生产部署）
- 至少 512MB 内存

## 1. 下载

到 [GitHub Releases](https://github.com/dujiao-next/dujiao-next/releases) 找最新的 `dujiao-all_*.tar.gz`，按系统架构选：

```bash
# 例：Linux amd64
wget https://github.com/dujiao-next/dujiao-next/releases/download/vX.Y.Z/dujiao-all_vX.Y.Z_linux_amd64.tar.gz
tar -xzf dujiao-all_*.tar.gz
cd <解压目录>
```

## 2. 复制配置

```bash
cp config.yml.example config.yml
```

## 3. 必改字段

打开 `config.yml`，按下表修改：

| 字段 | 说明 | 示例值 |
|---|---|---|
| `jwt.secret` | 后台管理员 JWT 密钥，**必改** | `openssl rand -hex 32` 输出 |
| `user_jwt.secret` | 用户 JWT 密钥，**必改** | 同上，不同值 |
| `web.admin_path` | 后台访问路径前缀，**强烈建议改** | `/dj-mgmt-7x9k2` |
| `redis.host` / `redis.port` | Redis 地址（分两个字段，默认 `127.0.0.1` + `6379`） | `127.0.0.1` + `6379` |
| `database.driver` / `database.dsn` | 数据库（默认 SQLite 起步） | 见下方 |

### 关于 `web.admin_path`（重要）

默认值 `/admin` 是自动化扫描器的头号目标。**强烈建议改成不易猜测的字符串**：

```yaml
web:
  admin_path: "/dj-mgmt-7x9k2"   # 个人偏好的字符串
```

这个路径只是 SPA 入口的"门牌"，改了它不影响 admin API 接口；API 鉴权由 JWT + 限流保护。改路径主要是过滤掉自动化扫描的噪音。

### 关于数据库

- **SQLite（默认）**：零配置，数据存在 `./db/dujiao.db`，单机够用。
- **PostgreSQL（生产推荐）**：把 `database.driver` 改为 `postgres`，`database.dsn` 写连接串。

## 4. 启动 Redis

本包附带最小 Redis 配置：

```bash
docker compose up -d redis
```

如果你已有 Redis（系统服务或其他容器），改 `config.yml` 的 `redis.host` 和 `redis.port` 即可，不需要起新容器。

## 5. 启动二进制

```bash
./dujiao-server
```

启动日志会显示：
```
🚀 Dujiao-Next API 启动中
...
Embedded SPAs: admin (/dj-mgmt-7x9k2), user (/)
```

二进制运行时会自动创建：
- `./db/`：SQLite 数据库
- `./uploads/`：用户上传文件
- `./logs/`：运行日志

## 6. 访问

- **用户端**：`http://<your-ip>:8080`
- **管理端**：`http://<your-ip>:8080/<web.admin_path>`（你刚才改的路径）

首次登录用默认管理员账号（在 `config.yml` 的 `bootstrap` 段配置）。**登录后立即修改密码**。

## 7. 反代与 HTTPS（生产部署）

把单个域名 `shop.example.com` 转发到二进制的 8080 端口（Nginx 示例）：

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

## 8. 系统服务（systemd）

先创建运行用户（如果你打算用专用用户跑服务）：
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

## 9. 升级

1. `systemctl stop dujiao`
2. 备份：`cp -r db uploads /backup/`
3. 下载新版 tar.gz，仅替换 `dujiao-server` 二进制
4. `systemctl start dujiao`

数据库迁移自动完成。

## 10. 从其他部署方式迁移

### 从 Docker Compose 迁移

1. 在 docker-compose 服务停机
2. 把现有 `db/`、`uploads/` 目录拷到 fullstack 二进制运行目录
3. 把 docker-compose 用的 `config.yml` 拷过来
4. 启动 fullstack 二进制（注意配置 `web.admin_path`）

### 从源码手动部署迁移

同上，路径和配置直接复用。

## 常见问题

### Q：admin SPA 加载报 404

确认 `config.yml` 的 `web.admin_path` 与浏览器访问路径一致；如果改了 `web.admin_path`，必须重启二进制让新值生效。

### Q：日志循环出现 "web.admin_path 仍为默认 /admin" 警告

按 §3 的建议修改 `web.admin_path`，警告会消失。

### Q：fullstack 二进制和现有 Docker 镜像可以混用吗？

可以，但同一个数据库不要同时被两套部署连接。建议：要么走 Docker 镜像方案，要么走单二进制方案，不混用。
