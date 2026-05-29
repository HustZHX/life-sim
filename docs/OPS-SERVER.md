# 服务器运维与发布流程

本文档沉淀 **阿里云 ECS 生产环境** 的日常运维：热更、全量拉取部署、SSH 访问与常见问题。首次部署仍见 [DEPLOY.md](../DEPLOY.md)；Agent 自动化见 [AGENT_DEPLOY.md](../AGENT_DEPLOY.md)。

> **安全**：`.env`、API Key、进门密码、JWT、SSH 私钥 **不得写入 Git**。下文仅记录路径与命令模板。

---

## 1. 生产环境一览

| 项 | 值 |
|----|-----|
| 云厂商 | 阿里云 ECS |
| 公网 IP | `47.108.84.163` |
| 访问地址 | http://47.108.84.163 |
| 系统 | Alibaba Cloud Linux 8（与 Ubuntu 用法相同，Docker 部署无差异） |
| 代码目录 | `/opt/life-sim` |
| 仓库 | https://github.com/HustZHX/life-sim.git（公开） |
| 编排 | Docker Compose（`docker-compose.yml`） |
| 对外端口 | **仅 80**（nginx 前端）；后端 `8080` 仅容器内 `expose` |
| 数据卷 | `lifesim-data` → 容器内 `/data/lifesim.db`（`docker compose down` **不会**删库） |

### 1.1 SSH

- 用户：`root`
- 主机：`47.108.84.163`
- 认证：**SSH 私钥**（由运维本地保管，勿提交仓库）
- 示例（将 `PATH_TO_KEY.pem` 换成本机私钥路径）：

```powershell
ssh -i "PATH_TO_KEY.pem" -o StrictHostKeyChecking=accept-new root@47.108.84.163
```

### 1.2 生产 `.env`

- 路径：`/opt/life-sim/.env`
- 权限建议：`chmod 600`
- 模板：项目根目录 [`.env.example`](../.env.example)
- 生产需启用：`AUTH_ENABLED=true`，并配置 `JWT_SECRET`、`SITE_ACCESS_CODE`、`AUTH_BOOTSTRAP_USERS`、`CORS_ORIGINS=http://47.108.84.163` 等（详见 DEPLOY.md）

### 1.3 容器与健康检查

```bash
cd /opt/life-sim
docker compose ps
curl -s -o /dev/null -w "health:%{http_code}\n" http://127.0.0.1/health
curl -s http://127.0.0.1/api/v1/auth/status
```

期望：`health:200`；鉴权开启时 `auth/status` 含 `enabled:true`。

---

## 2. 两种发布方式怎么选

| 方式 | 适用场景 | 代码来源 | 服务器 Git 状态 |
|------|----------|----------|-----------------|
| **热更** | 急需验证、尚未 push，或服务器工作区有未提交改动导致 `git pull` 失败 | 本机已改文件 → `scp` 覆盖 | 可脏，不依赖 pull |
| **全量更新** | 已 `git push`，希望服务器与远程分支一致 | `git pull` + 重建镜像 | 建议干净；有本地改动需先处理 |

原则：

- **日常推荐**：本地提交并推送 → 服务器 **全量更新**。
- **临时救火**：本地改完先 **热更** 验证，确认后再走提交推送 + 全量更新，避免服务器与 Git 长期分叉。

---

## 3. 热更流程（不改 Git，不 push）

将本机指定文件同步到服务器对应路径，在服务器上 **只重建受影响服务**（多数情况仅 `frontend`）。

### 3.1 本机（Windows PowerShell）

**1）同步文件**（按需增删路径；路径含空格必须加引号）：

```powershell
$KEY = "PATH_TO_KEY.pem"
$HOST = "root@47.108.84.163"
$REMOTE = "/opt/life-sim"

# 示例：仅前端组件
scp -i $KEY ".\frontend\src\components\TimelineAxis.vue" "${HOST}:${REMOTE}/frontend/src/components/TimelineAxis.vue"

# 示例：前端 + 后端各一个文件
scp -i $KEY ".\backend\store\timelines.go" "${HOST}:${REMOTE}/backend/store/timelines.go"
```

**2）在服务器构建并重启**：

```powershell
ssh -i $KEY -o StrictHostKeyChecking=accept-new $HOST "cd /opt/life-sim && docker compose up --build -d"
```

**3）验证**：

```powershell
ssh -i $KEY $HOST "curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1/health"
```

### 3.2 仅改前端时的注意点

- 必须执行 `docker compose up --build -d`（或至少 `docker compose build frontend && docker compose up -d frontend`），否则 nginx 静态资源不会更新。
- 浏览器可能缓存旧 JS/CSS：强刷或无痕窗口。

### 3.3 仅改后端时

- 同步 `backend/` 下变更文件后同样 `docker compose up --build -d`。
- 后端重启会将在途异步任务标为失败，需在前端重新提交生成（见 [DEVELOPMENT.md](../DEVELOPMENT.md)）。

### 3.4 热更风险

- 服务器 `/opt/life-sim` 与 Git 远程 **不一致**，后续 `git pull` 可能冲突。
- 热更后应尽快 **提交推送 + 全量更新**，让服务器回到「以 Git 为准」的状态。

---

## 4. 全量更新流程（提交 → 推送 → 服务器拉取部署）

### 4.1 本机：提交并推送

```powershell
cd C:\Users\Administrator\Documents\life-sim   # 改为你的本机路径

git status -sb
git add <相关文件>
git commit -m "简述本次变更目的"
git push
```

推送前确认：

- [ ] 未包含 `.env`、私钥、真实密码
- [ ] 前端 `npm run build` 通过（可选但推荐）

当前常用远程分支示例：`feature/mobile-layout-mode`（以你实际分支为准）。

### 4.2 服务器：拉取并重建

```bash
cd /opt/life-sim

# 查看是否有未提交改动（有则见 4.3）
git status -sb

git pull
docker compose up --build -d
docker compose ps
curl -s http://127.0.0.1/health
```

首次在该机部署或换分支时：

```bash
cd /opt
git clone https://github.com/HustZHX/life-sim.git life-sim
cd life-sim
cp .env.example .env
# 编辑 .env 填入生产配置
chmod 600 .env
docker compose up --build -d
```

### 4.3 服务器 `git pull` 失败：有本地改动

典型报错：`cannot pull with rebase: You have unstaged changes`

**方案 A（推荐）**：若本地改动应废弃，与远程对齐：

```bash
cd /opt/life-sim
git stash push -u -m "server-local-before-pull"
git pull
docker compose up --build -d
```

**方案 B**：保留服务器改动、暂不 pull → 继续用 **热更** 从本机 `scp`，直到本机 push 后再 stash/reset 对齐。

**方案 C**：确认服务器改动可丢：

```bash
git fetch origin
git reset --hard origin/<你的分支名>
docker compose up --build -d
```

> 执行 `reset --hard` 前确认不会误删仅在服务器上的 `.env`（`.env` 应在 `.gitignore` 中，一般不受影响）。

### 4.4 全量更新后验证清单

- [ ] `docker compose ps`：`backend`、`frontend` 均为 `running`
- [ ] `curl http://127.0.0.1/health` → 200
- [ ] 浏览器访问 http://47.108.84.163：进门 → 登录 → 打开人物时间轴
- [ ] 公网 **不能** 访问 `http://47.108.84.163:8080`
- [ ] 涉及世界线/时间轴 UI 时，强刷缓存后再看

---

## 5. 常用运维命令（服务器上）

```bash
cd /opt/life-sim

# 日志
docker compose logs -f backend --tail 100
docker compose logs -f frontend --tail 50

# 仅重启（不重建镜像，改 .env 时可用）
docker compose restart

# 停止（不删数据卷）
docker compose down

# 重新构建并后台运行
docker compose up --build -d

# 查看镜像与容器占用
docker compose images
docker system df
```

---

## 6. 构建加速（国内服务器）

在阿里云等国内环境，首次 `docker compose build` 可能因访问 `proxy.golang.org` / npm 官方源超时失败。项目已在 **`backend/Dockerfile`** 使用：

```dockerfile
ENV GOPROXY=https://goproxy.cn,direct
```

若前端构建超时，可在 **`frontend/Dockerfile`** 的 `npm ci` 前增加（与服务器热更实践一致）：

```dockerfile
RUN npm config set registry https://registry.npmmirror.com
```

修改 Dockerfile 后需提交 Git 或 scp 到服务器再 `docker compose up --build -d`。

---

## 7. 故障排查速查

| 现象 | 可能原因 | 处理 |
|------|----------|------|
| 外网打不开 | 安全组未放行 80 | 阿里云控制台入方向 TCP 80 |
| 502 / 白屏 | 容器未起来 | `docker compose logs frontend backend` |
| 热更后页面无变化 | 未 rebuild 前端或浏览器缓存 | `up --build -d` + 强刷 |
| `git pull` 失败 | 服务器有未提交文件 | 见 §4.3 |
| 世界线不显示 | 列表 API 未带 `world_line` | 确认已部署含 `ListTimelines` 修复的后端 |
| 进门/登录 429 | 鉴权限流 | 等待约 1 分钟后重试 |
| AI 失败 | Key 或出网 | 检查 `.env` 中 `DEEPSEEK_API_KEY`；`curl https://api.deepseek.com` |
| `npm run build` 本地失败 | 依赖未装全 | `cd frontend && npm ci && npm run build` |

---

## 8. 与仓库其他文档的关系

| 文档 | 用途 |
|------|------|
| [DEPLOY.md](../DEPLOY.md) | 首次公网部署、`.env`、安全组、HTTPS 后续 |
| [AGENT_DEPLOY.md](../AGENT_DEPLOY.md) | 给 Agent 的逐步部署清单 |
| [DEVELOPMENT.md](../DEVELOPMENT.md) | 本地开发、重启、数据兼容 |
| [docs/superpowers/plans/](../superpowers/plans/) | 功能实现计划（非运维手册） |

---

## 9. 变更记录（运维相关）

| 日期 | 说明 |
|------|------|
| 2026-05 | 阿里云 ECS 生产部署；双重鉴权；仓库公开；`/opt/life-sim` + Docker Compose |
| 2026-05 | 确立热更（scp + compose build）与全量（push + pull + build）两套流程 |
| 2026-05 | 时间轴 UI：世界线并入主轴、移除左侧 WorldLinePanel、手机端详情导航 |
