---
outline: deep
---

# 升级与迁移

> 更新时间：2026-03-28

本指南介绍如何将 Dujiao-Next 从旧版本升级到新版本。

---

## 1. 升级前准备

### 1.1 备份数据

**必须在升级前备份以下内容：**

- 数据库（SQLite 文件或 PostgreSQL 数据）
- 配置文件（`config.yml`）
- 上传文件目录（`uploads/`）

参考 [备份与恢复](/deploy/backup) 指南。

### 1.2 查看更新日志

升级前务必阅读 [更新日志](/intro/changelog)，了解：

- 新增功能和配置项
- 破坏性变更（Breaking Changes）
- 数据库结构变更
- 配置文件新增字段

---

## 2. 手动部署升级

### 2.1 停止服务

```bash
# 停止后端服务
systemctl stop dujiao-next
# 或手动停止进程
```

### 2.2 备份

```bash
# 备份数据库（SQLite）
cp api/db/dujiao.db api/db/dujiao.db.bak.$(date +%Y%m%d)

# 备份配置
cp api/config.yml api/config.yml.bak

# 备份上传文件
cp -r api/uploads api/uploads.bak
```

### 2.3 更新代码

```bash
cd dujiao-studio
git pull origin main
```

### 2.4 重新构建

```bash
# 后端
cd api && go build ./cmd/server ./cmd/seed

# 前台
cd user && npm install && npm run build

# 管理后台
cd admin && npm install && npm run build
```

### 2.5 更新配置

对比 `config.yml.example` 检查是否有新增配置项，按需添加到 `config.yml`。

### 2.6 启动服务

```bash
systemctl start dujiao-next
```

> 数据库结构变更会在启动时由 GORM 自动迁移完成，无需手动执行 SQL。

---

## 3. Docker Compose 升级

### 3.1 备份

```bash
# 备份数据卷
docker compose exec api cp /app/db/dujiao.db /app/db/dujiao.db.bak

# 导出配置
docker compose cp api:/app/config.yml ./config.yml.bak
```

### 3.2 更新镜像

```bash
# 拉取最新镜像
docker compose pull

# 或使用指定版本
# 修改 docker-compose.yml 中的 image 版本号
```

### 3.3 重启服务

```bash
docker compose down
docker compose up -d
```

### 3.4 检查日志

```bash
docker compose logs -f api
```

---

## 4. 升级后验证

升级完成后按以下步骤验证：

1. **后台登录**：确认管理员可正常登录后台
2. **仪表盘**：检查数据是否正常显示
3. **商品列表**：确认商品和库存数据正确
4. **创建测试订单**：前台下单并完成支付测试
5. **支付回调**：确认支付回调正常工作
6. **邮件通知**：确认邮件发送功能正常

---

## 5. 回滚

如果升级后出现问题：

### 5.1 手动部署回滚

```bash
# 停止服务
systemctl stop dujiao-next

# 恢复数据库
cp api/db/dujiao.db.bak.YYYYMMDD api/db/dujiao.db

# 恢复配置
cp api/config.yml.bak api/config.yml

# 切换回旧版本代码
git checkout <旧版本tag或commit>

# 重新构建并启动
cd api && go build ./cmd/server
systemctl start dujiao-next
```

### 5.2 Docker 回滚

```bash
docker compose down

# 修改 docker-compose.yml 中的镜像版本为旧版本
# 恢复数据库备份

docker compose up -d
```

---

## 6. 注意事项

- 跨多个版本升级时，建议逐版本升级或仔细阅读每个版本的更新日志
- 数据库自动迁移只增加字段/表，不会删除已有字段
- 升级后首次启动可能稍慢（执行数据库迁移）
- 配置文件新增字段通常有默认值，不加也不影响启动，但建议补全
