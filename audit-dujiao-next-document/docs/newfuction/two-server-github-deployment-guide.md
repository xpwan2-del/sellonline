# sellonline 统一仓库与三套正式环境部署说明

## 这份文档的目的

这份文档记录当前真实部署方案，避免再次出现“推错分支、推错 Vercel 项目、推错服务器”的问题。

核心目标：

```text
同一套代码
三套正式环境独立配置
指定哪个环境，就只更新哪个环境
代码更新不能覆盖数据库、卡密、支付、邮箱、上传文件和后台配置
```

核心原则：

```text
代码统一
配置分离
数据库分离
分支隔离
Vercel 项目隔离
部署前先确认目标
```

## 当前三套正式环境

### 正式环境 A：toplenged

```text
实例：instance-20260606-005744
区域：asia-northeast1-a
内网 IP：10.146.0.3
公网 IP：34.104.155.75
域名：toplenged.com / www.toplenged.com
API 域名：https://api.toplenged.com
admin 域名：https://admin.toplenged.com
用户端 Vercel 项目：https://vercel.com/xpwan1-7539s-projects/toplenged
用户端部署分支：deploy/toplenged
```

### 正式环境 B：888tech

```text
实例：instance-888tech
原故障实例：instance-20260607-114457
原故障公网 IP：34.59.247.41
区域：asia-northeast1-b
内网 IP：10.146.0.4
公网 IP：35.243.99.93
域名：888tech.club / www.888tech.club
API 域名：https://api.888tech.club
admin 域名：https://admin.888tech.club
用户端 Vercel 项目：https://vercel.com/xpwan1-7539s-projects/888tech
用户端部署分支：deploy/888tech
```

说明：`35.243.99.93 / instance-888tech` 已经代替原来的 `34.59.247.41 / instance-20260607-114457`，现在是 888tech 的正式运行服务器。

### 正式环境 C：top-ai

```text
实例：instance-top-ai
原名：instance-20260609-111237
区域：asia-northeast1-c
内网 IP：10.146.0.2
公网 IP：35.243.101.220
域名：top-ai.band / www.top-ai.band
API 域名：https://api.top-ai.band
admin 域名：https://admin.top-ai.band
用户端 Vercel 项目：https://vercel.com/xpwan1-7539s-projects/top-ai
用户端部署分支：deploy/top-ai
```

## 统一 GitHub 仓库

当前统一仓库：

```text
https://github.com/xpwan2-del/sellonline
```

标准目录：

```text
sellonline/
  audit-dujiao-next-api/
  audit-dujiao-next-admin/
  audit-dujiao-next-user/
  audit-dujiao-next-document/
  scripts/
  vercel.json
  .gitignore
```

说明：

```text
audit-dujiao-next-api      API 后端
audit-dujiao-next-admin    admin 后台前端
audit-dujiao-next-user     用户端前端，供 toplenged / 888tech / top-ai 三个 Vercel 项目共用
audit-dujiao-next-document 文档
scripts                    通用脚本，包含 Vercel 分支保护脚本
```

不要长期维护多个用户端 GitHub 仓库。以后标准仓库是 `xpwan2-del/sellonline`。

注意：`deploy/toplenged`、`deploy/888tech`、`deploy/top-ai` 主要用于用户端 Vercel 部署。服务器端 API/admin 是否从这个仓库拉取，必须看当前服务器部署方式，不能因为仓库里有目录就直接整包覆盖服务器。

## 当前已经确认过的问题

### 1. 不能直接推 main

```text
sellonline/main
```

导致：

```text
toplenged Vercel 项目也被触发
888tech Vercel 项目也被触发
top-ai Vercel 项目也被触发
```

这说明：只连接同一个 GitHub 仓库还不够，必须做分支隔离和 Vercel 构建保护。

### 2. Vercel 项目名和旧部署名容易混淆

```text
toplenged 项目以前叫 project-2f6ez。
Vercel 部署 URL 里仍可能出现 project-2f6ez。
```

这不是另一个项目。看到 `project-2f6ez.vercel.app` 时，先确认它是不是 `toplenged` 项目的历史部署名，不要误以为有两个 toplenged 项目。

### 3. Promote 和 Redeploy 不是一回事

```text
Promote：把某个已经构建好的部署切成正式域名正在使用的版本。
Redeploy：重新构建一次，常用于改了 Vercel 环境变量以后让新变量进入前端包。
Rollback：把正式域名回到旧部署。
```

如果生产项目处于 Instant Rollback 状态，新部署 Ready 以后，正式域名可能仍然停在旧版本。必须在正确部署上点 `Promote`，正式域名才会切过去。

## Vercel 分支隔离方案

不能用同一个 `main` 分支同时触发三个 Vercel 项目。

当前推荐分支：

```text
deploy/toplenged  只部署 toplenged Vercel 项目
deploy/888tech    只部署 888tech Vercel 项目
deploy/top-ai     只部署 top-ai Vercel 项目
main              代码整合分支，不应该直接作为三套环境的自动部署入口
```

以后更新哪个项目，就推哪个分支：

```bash
# 只更新 toplenged 前端
git push origin HEAD:deploy/toplenged

# 只更新 888tech 前端
git push origin HEAD:deploy/888tech

# 只更新 top-ai 前端
git push origin HEAD:deploy/top-ai
```

说明：上面的 `origin` 必须指向 `https://github.com/xpwan2-del/sellonline`。如果当前本地仓库的 remote 名叫 `sellonline`，就把命令里的 `origin` 换成 `sellonline`。

禁止直接推 `main` 让 Vercel 自动部署正式项目。

## Vercel 项目必须配置的环境变量

三个正式 Vercel 项目都必须设置 `DEPLOY_BRANCH`，用来决定自己只允许哪个分支部署。

### toplenged 项目

```text
Project：toplenged
Root Directory：audit-dujiao-next-user
VITE_API_BASE_URL：https://api.toplenged.com
DEPLOY_BRANCH：deploy/toplenged
```

### 888tech 项目

```text
Project：888tech
Root Directory：audit-dujiao-next-user
VITE_API_BASE_URL：https://api.888tech.club
DEPLOY_BRANCH：deploy/888tech
```

### top-ai 项目

```text
Project：top-ai
Root Directory：audit-dujiao-next-user
VITE_API_BASE_URL：https://api.top-ai.band
DEPLOY_BRANCH：deploy/top-ai
```

环境变量位置：

```text
Vercel Project -> Settings -> Environment Variables
```

注意：

```text
DEPLOY_BRANCH 不是代码写死项目
DEPLOY_BRANCH 是每个 Vercel 项目自己的部署规则
同一份代码读取不同项目的 DEPLOY_BRANCH，决定是否跳过构建
VITE_API_BASE_URL 是前端构建时写进 JS 包的值
改了 VITE_API_BASE_URL 以后，必须 Redeploy，旧部署不会自动变
```

三个环境的 API 地址必须严格对应：

```text
toplenged -> https://api.toplenged.com
888tech   -> https://api.888tech.club
top-ai    -> https://api.top-ai.band
```

禁止把 `toplenged`、`888tech`、`top-ai` 任意一个项目的 `VITE_API_BASE_URL` 填成另一个项目的 API。

## Vercel Root Directory 和构建设置

三个正式 Vercel 项目的 Root Directory 都必须是：

```text
audit-dujiao-next-user
```

如果 Root Directory 是 `./`，Vercel 会在仓库根目录执行构建，可能报：

```text
sh: line 1: vite: command not found
Error: Command "vite build" exited with 127
```

正确设置：

```text
Framework Preset：Vite（推荐）
Root Directory：audit-dujiao-next-user
Install Command：默认 npm install 即可
Build Command：npm run build
Output Directory：dist
Node.js Version：24.x
```

说明：

```text
用户端本质是 Vite 项目，所以新项目推荐选 Vite。
如果已有正式项目 Framework Preset 是 Other，但 Build Command 和 Output Directory 正常，也能运行。
不要为了统一 Framework Preset，在无关发布里顺手改正在正常运行的正式项目。
```

## Vercel 分支保护脚本

仓库中应该存在：

```text
scripts/vercel-ignore-by-branch.mjs
audit-dujiao-next-user/scripts/vercel-ignore-by-branch.mjs
vercel.json
audit-dujiao-next-user/vercel.json
```

脚本逻辑：

```text
如果 DEPLOY_BRANCH 没设置：跳过部署
如果当前 Git 分支 != DEPLOY_BRANCH：跳过部署
如果当前 Git 分支 == DEPLOY_BRANCH：允许部署
```

Vercel 的规则是：

```text
ignoreCommand 退出码 0 = 跳过构建
ignoreCommand 退出码 1 = 继续构建
```

脚本必须保持通用，不能写死 `toplenged`、`888tech`、`top-ai`。

示例逻辑：

```js
const currentBranch = process.env.VERCEL_GIT_COMMIT_REF || ''
const deployBranch = process.env.DEPLOY_BRANCH || ''

if (!deployBranch) process.exit(0)
if (currentBranch !== deployBranch) process.exit(0)
process.exit(1)
```

`audit-dujiao-next-user/vercel.json` 必须保留 SPA rewrite，同时加入 `ignoreCommand`：

```json
{
  "ignoreCommand": "node scripts/vercel-ignore-by-branch.mjs",
  "rewrites": [
    {
      "source": "/(.*)",
      "destination": "/index.html"
    }
  ]
}
```

## 用户端 API 地址规则

用户端前端不能写死任何一个域名。

正确规则：

```text
代码读取 VITE_API_BASE_URL
Vercel 每个项目设置自己的 VITE_API_BASE_URL
```

禁止：

```text
代码里写死 https://api.888tech.club
代码里写死 https://api.toplenged.com
代码里写死 https://api.top-ai.band
根据 window.location.hostname 写 if/else 切 API
```

检查文件：

```text
audit-dujiao-next-user/src/config/api.ts
audit-dujiao-next-user/src/api/client.ts
```

如果发现硬编码生产域名，必须先修掉，再部署。

部署后必须验证当前前端包实际请求的 API：

```bash
# 查看首页 HTML 引用的前端入口 JS
curl -sS https://www.toplenged.com/ | grep -Eo 'assets/index-[^" ]+\.js'

# 如果页面为空，用浏览器 Console / Network 看实际请求域名
# toplenged 正确请求应该是 https://api.toplenged.com/api/v1/...
# 888tech 正确请求应该是 https://api.888tech.club/api/v1/...
# top-ai 正确请求应该是 https://api.top-ai.band/api/v1/...
```

如果前端请求了错误 API：

```text
1. 去对应 Vercel 项目 Settings -> Environment Variables
2. 修改 VITE_API_BASE_URL
3. 保存 Production 环境变量
4. 回到 Deployments
5. 对目标 commit 执行 Redeploy
6. Redeploy 成功后，检查正式域名是否切到新部署
7. 如果正式域名仍停在旧部署或 Rollback 状态，再 Promote 到 Production
```

## 后台站点配置规则

这些内容不能写死在代码里：

```text
站点名称
站点网址
用户前端域名
admin 域名
API 域名
邮箱配置
支付通道配置
返佣配置
公告
hero
商品数据
卡密数据
```

它们应该来自：

```text
每台服务器自己的数据库
每台服务器自己的后台设置
每台服务器自己的 .env / config.yml
每个 Vercel 项目自己的环境变量
```

例如邀请链接不能写死 `toplenged.com`，应该读取后台 `site_config.brand.site_url`。

## 什么可以提交到 GitHub

可以提交：

```text
API 源码
admin 源码
用户端源码
文档
package.json
package-lock.json
go.mod
go.sum
vercel.json
通用脚本
示例配置
.env.example
```

## 什么绝对不能提交到 GitHub

不能提交：

```text
.env
.env.local
.env.*
.vercel
数据库文件
*.db
*.sqlite
*.sqlite3
uploads
storage
logs
node_modules
dist
.cache
.playwright-cli
deploy-artifacts
codex_ssh_keys
SSH 私钥
Vercel token
支付密钥
邮箱授权码
JWT 密钥
真实 Nginx 配置
真实 systemd 配置
真实用户数据
真实订单数据
真实充值记录
卡密数据
```

提交前必须检查：

```bash
git status
git diff --cached
git ls-files
```

## 服务器端部署原则

API/admin 服务器部署代码时，只能更新：

```text
API 编译后的程序文件
admin 构建后的静态文件
必要的源码版本
必要的数据库结构迁移
```

不能覆盖：

```text
数据库文件
真实 config.yml
真实 .env
uploads
storage
logs
Nginx 真实配置
systemd 真实配置
后台设置
支付配置
邮箱配置
商品数据
卡密数据
```

如果涉及数据库结构变化，必须先备份每台服务器自己的数据库，再在每台服务器自己的数据库上做结构迁移。

不能把任何一台服务器的数据库复制覆盖到另一台。

当前服务器端部署目录规则：

```text
toplenged 服务器：/opt/dujiao-test
888tech 服务器：部署前先 SSH 只读确认目录，未确认目录前禁止替换文件
top-ai 服务器：部署前先 SSH 只读确认目录，未确认目录前禁止替换文件
```

说明：`/opt/dujiao-test` 只是 toplenged 服务器上的历史目录名，不代表它是另一个可部署目标。

服务器上通常由 `docker-compose.yml` 管理：

```text
api   -> 只更新 API 源码/镜像，挂载服务器自己的 config 和 data
admin -> 推荐上传本地已 build 的 dist，用 nginx 预构建镜像部署
redis -> 不因代码部署重建，除非明确需要
```

部署前必须备份：

```text
当前 API 目录
当前 admin 目录
当前服务器自己的 data/db
```

允许的安全做法：

```text
1. 本地构建并测试 API/admin
2. API 只打源码包，不包含 .env、db、uploads、logs
3. admin 用本地 npm run build 后的 dist 打预构建包
4. 上传到目标服务器 backups 目录
5. 服务器 hostname 校验通过后，再替换 app/audit-dujiao-next-api 和 app/audit-dujiao-next-admin
6. docker-compose build api admin
7. docker-compose up -d api admin
8. docker-compose ps 验证容器 Up
```

不推荐在服务器上重新跑 admin 的 Node/Tailwind 构建。之前遇到过服务器构建依赖问题，admin 静态前端用本地 `dist` 预构建更稳。

## 正式环境部署前检查

正式环境 A：

```text
hostname 对应 instance-20260606-005744
公网 IP：34.104.155.75
域名：toplenged.com
```

正式环境 B：

```text
hostname 对应 instance-888tech
公网 IP：35.243.99.93
域名：888tech.club
```

正式环境 C：

```text
hostname 应该对应 instance-top-ai
原名：instance-20260609-111237
公网 IP：35.243.101.220
域名：top-ai.band
```

注意：如果 SSH 后 `hostname` 仍返回 `instance-20260609-111237`，必须先确认它就是 `35.243.101.220` 这台 top-ai 服务器，不能自行假定后继续部署。

任何正式部署前必须确认：

```bash
hostname
pwd
git remote -v
git status
git rev-parse --short HEAD
```

不确认目标服务器，不允许部署。

## 正式前端发布流程

正式发布必须先明确目标环境。

### 发布 toplenged 前端

```bash
git push origin HEAD:deploy/toplenged
```

发布后必须确认：

```text
Vercel Project：toplenged
Branch：deploy/toplenged
Commit：本次推送的 commit
Status：Ready
VITE_API_BASE_URL：https://api.toplenged.com
```

如果这个项目之前做过 Instant Rollback，新的部署 Ready 后不一定自动接管 `www.toplenged.com`。必须打开该部署，点 `Promote`，确认它会 alias 到：

```text
toplenged.com
www.toplenged.com
project-2f6ez.vercel.app
```

`project-2f6ez.vercel.app` 是 toplenged 项目的历史项目地址，不代表另一个项目。

### 发布 888tech 前端

```bash
git push origin HEAD:deploy/888tech
```

发布后必须确认：

```text
Vercel Project：888tech
Branch：deploy/888tech
Commit：本次推送的 commit
Status：Ready
VITE_API_BASE_URL：https://api.888tech.club
```

如果项目处于 Instant Rollback 状态，新的部署 Ready 后同样要手动 `Promote`。

### 发布 top-ai 前端

```bash
git push origin HEAD:deploy/top-ai
```

发布后必须确认：

```text
Vercel Project：top-ai
Branch：deploy/top-ai
Commit：本次推送的 commit
Status：Ready
VITE_API_BASE_URL：https://api.top-ai.band
```

如果项目处于 Instant Rollback 状态，新的部署 Ready 后同样要手动 `Promote`。

发布前必须确认：

```text
toplenged 项目 DEPLOY_BRANCH=deploy/toplenged
toplenged 项目 VITE_API_BASE_URL=https://api.toplenged.com
888tech 项目 DEPLOY_BRANCH=deploy/888tech
888tech 项目 VITE_API_BASE_URL=https://api.888tech.club
top-ai 项目 DEPLOY_BRANCH=deploy/top-ai
top-ai 项目 VITE_API_BASE_URL=https://api.top-ai.band
Root Directory=audit-dujiao-next-user
```

不要一次盲目更新多个正式项目。推荐先更新一个，验证无误，再更新另一个。

正式前端发布后的最低验证：

```bash
# 确认正式域名当前指向哪个 Vercel 部署
npx -y vercel@latest inspect https://www.toplenged.com --scope xpwan1-7539s-projects
npx -y vercel@latest inspect https://www.888tech.club --scope xpwan1-7539s-projects
npx -y vercel@latest inspect https://www.top-ai.band --scope xpwan1-7539s-projects

# 确认 API 有数据
curl -sS https://api.toplenged.com/api/v1/public/products | head -c 500
curl -sS https://api.888tech.club/api/v1/public/products | head -c 500
curl -sS https://api.top-ai.band/api/v1/public/products | head -c 500
```

## Vercel 误触发生产后的处理

如果发现某个环境提交误触发了其它正式项目：

```text
1. 立刻停止新的推送和部署
2. 查看 Vercel Deployments，确认误触发的新部署 URL
3. 找到上一版正常部署 URL
4. 对生产项目执行 rollback
5. 验证生产域名是否回到旧版本
6. 修复分支隔离和 DEPLOY_BRANCH 后，才能继续测试部署
```

回滚命令示例：

```bash
npx -y vercel@latest rollback <上一版部署URL> --scope xpwan1-7539s-projects --yes --timeout 5m
```

回滚后要检查：

```bash
curl -sSI https://www.toplenged.com/
curl -sSI https://www.888tech.club/
curl -sSI https://www.top-ai.band/
```

注意：

```text
Rollback 只是把正式域名临时切回旧部署。
修复问题后，新的部署 Ready 了还需要 Promote，才能取消当前 Instant Rollback。
如果只 Redeploy 但不 Promote，正式域名可能仍然停在旧部署。
```

## 常见故障和处理

### 页面显示 Site，商品和文章都是空

优先判断为前端没有拿到 `/public/config`、`/public/products`、`/public/posts`。

检查：

```text
浏览器 Console / Network
请求域名是否正确
请求是否被 CORS 拦截
请求返回的是 JSON 还是 index.html
```

处理：

```text
如果请求到了其它环境的 API：改 VITE_API_BASE_URL 为当前环境 API，然后 Redeploy。
如果请求到了 www.<域名>/api/v1：说明 VITE_API_BASE_URL 没进构建，检查 Vercel 环境变量和 Redeploy。
如果 api.<域名> 有 JSON 但页面空：看 Console 是否有字段解析错误或 CORS 错误。
```

### API 有数据，前端仍然暂无商品

必须分别验证：

```bash
curl -sS https://api.toplenged.com/api/v1/public/products | head -c 500
curl -sS https://www.toplenged.com/ | grep -Eo 'assets/index-[^" ]+\.js'
npx -y vercel@latest inspect https://www.toplenged.com --scope xpwan1-7539s-projects
```

把上面的域名替换成当前目标环境域名。如果 API 有数据，但用户端域名的 inspect 显示仍是旧部署，去 Vercel 对新部署点 `Promote`。

### 改了 Vercel 环境变量但页面没变

这是正常的。Vite 前端环境变量是构建时写入 JS 包的。

```text
改环境变量 -> 必须 Redeploy -> 新部署 Ready -> 必要时 Promote
```

## 数据库迁移和数据覆盖的区别

数据库迁移是：

```text
新增字段
新增索引
新增表
调整字段默认值
```

数据库数据是：

```text
商品
卡密
公告
hero
用户
订单
支付记录
充值记录
后台设置
邮箱配置
返佣配置
```

部署代码时可以做必要的数据库结构迁移，但不能复制或覆盖业务数据。

例如：

```text
products.sold_count_offset
```

这是结构字段，可以在每台服务器自己的数据库上新增。

但是不能把任何一台服务器的数据库复制覆盖到另一台服务器。

## 最重要的禁止事项

```text
禁止推 main 来部署任何正式项目
禁止让 toplenged/888tech/top-ai 三个项目监听同一个部署分支
禁止没有 DEPLOY_BRANCH 保护就连接同一个 GitHub 仓库
禁止把正式环境 A 数据库覆盖正式环境 B
禁止把正式环境 B 数据库覆盖正式环境 A
禁止把正式环境 C 数据库覆盖正式环境 A 或 B
禁止把正式环境 A 或 B 数据库覆盖正式环境 C
禁止把本地数据库上传覆盖服务器数据库
禁止把 .env 提交到 GitHub
禁止把品牌、域名、支付、邮箱写死在代码里
禁止为了某个域名写代码特殊判断
禁止更新代码时顺手改生产数据
禁止从本地整包上传覆盖服务器项目目录
禁止使用没有明确排除规则的 rsync --delete
禁止 scp 整个项目目录覆盖服务器
```

## 下次让 Codex 部署前必须先读

下次部署前，对 Codex 说：

```text
先读 audit-dujiao-next-document/docs/newfuction/two-server-github-deployment-guide.md，再部署。
```

Codex 必须先回答并确认：

```text
这次目标是 toplenged / 888tech / top-ai 哪一个？
目标 Git 分支是什么？
目标 Vercel 项目是什么？
目标 Vercel 项目的 VITE_API_BASE_URL 应该是什么？
目标 Vercel 项目的 DEPLOY_BRANCH 应该是什么？
目标服务器 IP 是什么？
目标服务器 hostname 必须是什么？
是否需要数据库迁移？
是否已经备份？
哪些文件允许修改？
哪些环境绝对不能碰？
这次是否需要 Promote？
这次是否需要 Redeploy？
```

没有确认这些，不允许开始部署。

## 一句话结论

以后要做到“指定哪个项目，就只更新哪个项目”，靠的是：

```text
统一仓库 xpwan2-del/sellonline
每个 Vercel 项目设置自己的 DEPLOY_BRANCH
代码里的 ignoreCommand 通用读取 DEPLOY_BRANCH
部署时只推对应 deploy/* 分支
服务器端配置和数据库永远独立
```
