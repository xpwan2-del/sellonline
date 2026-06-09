# Single Binary Deployment (Recommended for Beginners)

> Who this is for: Complete beginners who don't want to deal with multi-container Docker orchestration and prefer to run everything with "one binary + one Redis container + one domain".

## When to Choose This

| Deployment Method | Difficulty | Container Count | Domain Count |
|---|---|---|---|
| **Single Binary (this guide)** | Low | 1 (Redis) | 1 |
| Docker Compose | Medium | 4-5 | 1-2 |
| aaPanel Manual Deployment | Low-Medium | 0 (bare metal) | 1-2 |
| Manual Source Build | High | 0 | 1-2 |

## System Requirements

- Linux x86_64 or arm64
- Docker (only required for running the Redis container; can be skipped if you already have a Redis service)
- One domain + SSL certificate (for production deployment)
- At least 512MB of memory

## 1. Download

Go to [GitHub Releases](https://github.com/dujiao-next/dujiao-next/releases) and find the latest `dujiao-all_*.tar.gz`. Pick the one matching your system architecture:

```bash
# Example: Linux amd64
wget https://github.com/dujiao-next/dujiao-next/releases/download/vX.Y.Z/dujiao-all_vX.Y.Z_linux_amd64.tar.gz
tar -xzf dujiao-all_*.tar.gz
cd <extracted-directory>
```

## 2. Copy Configuration

```bash
cp config.yml.example config.yml
```

## 3. Required Configuration Changes

Open `config.yml` and update the fields below:

| Field | Description | Example Value |
|---|---|---|
| `jwt.secret` | Admin JWT secret. **Must change** | Output of `openssl rand -hex 32` |
| `user_jwt.secret` | User JWT secret. **Must change** | Same as above, but a different value |
| `web.admin_path` | Admin URL prefix. **Strongly recommended to change** | `/dj-mgmt-7x9k2` |
| `redis.host` / `redis.port` | Redis address (two fields, default `127.0.0.1` + `6379`) | `127.0.0.1` + `6379` |
| `database.driver` / `database.dsn` | Database (defaults to SQLite) | See below |

### About `web.admin_path` (Important)

The default value `/admin` is the number-one target for automated scanners. **Strongly recommended to change it to a hard-to-guess string**:

```yaml
web:
  admin_path: "/dj-mgmt-7x9k2"   # A string of your choosing
```

This path is just the "doorplate" for the SPA entry. Changing it does not affect the admin API endpoints; API authorization is protected by JWT and rate limiting. The main reason to change this path is to filter out noise from automated scanners.

### Database Options

- **SQLite (default)**: Zero configuration. Data is stored in `./db/dujiao.db`. Sufficient for a single-machine setup.
- **PostgreSQL (recommended for production)**: Set `database.driver` to `postgres` and `database.dsn` to your connection string.

## 4. Start Redis

This package ships with a minimal Redis configuration:

```bash
docker compose up -d redis
```

If you already have Redis running (as a system service or another container), just update `redis.host` and `redis.port` in `config.yml` — there's no need to start a new container.

## 5. Start the Binary

```bash
./dujiao-server
```

The startup log will show:
```
🚀 Dujiao-Next API 启动中
...
Embedded SPAs: admin (/dj-mgmt-7x9k2), user (/)
```

When running, the binary automatically creates:
- `./db/`: SQLite database
- `./uploads/`: User-uploaded files
- `./logs/`: Runtime logs

## 6. Access the Site

- **User frontend**: `http://<your-ip>:8080`
- **Admin panel**: `http://<your-ip>:8080/<web.admin_path>` (the path you just configured)

For the first login, use the default admin account (configured in the `bootstrap` section of `config.yml`). **Change the password immediately after logging in.**

## 7. Reverse Proxy & HTTPS (Production Deployment)

Forward a single domain `shop.example.com` to port 8080 of the binary (Nginx example):

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

## 8. systemd Service

First create the runtime user (if you plan to run the service under a dedicated account):
```bash
sudo useradd -r -s /sbin/nologin -d /opt/dujiao dujiao
sudo chown -R dujiao:dujiao /opt/dujiao
```

`/etc/systemd/system/dujiao.service`:
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

## 9. Upgrade

1. `systemctl stop dujiao`
2. Back up: `cp -r db uploads /backup/`
3. Download the new tar.gz and replace only the `dujiao-server` binary
4. `systemctl start dujiao`

Database migrations run automatically.

## 10. Migrating from Other Deployments

### From Docker Compose

1. Stop the docker-compose services
2. Copy the existing `db/` and `uploads/` directories into the fullstack binary's working directory
3. Copy over the `config.yml` used by docker-compose
4. Start the fullstack binary (be sure to configure `web.admin_path`)

### From Manual Source Build

Same as above — paths and configuration can be reused directly.

## FAQ

### Q: The admin SPA returns a 404 when loading

Make sure the `web.admin_path` in `config.yml` matches the path you're using in the browser. If you changed `web.admin_path`, you must restart the binary for the new value to take effect.

### Q: The log keeps showing the warning `web.admin_path 仍为默认 /admin`

> Note: the warning is emitted in Simplified Chinese — the backend log strings are not yet i18n-aware. Grep for the literal Chinese text above.

Follow the recommendation in §3 and change `web.admin_path`. The warning will go away.

### Q: Can the fullstack binary and an existing Docker image be used together?

Yes, but the same database must not be connected to by both deployments at the same time. Recommendation: pick either the Docker image approach or the single binary approach — don't mix them.
