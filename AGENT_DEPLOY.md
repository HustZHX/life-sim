# Life-Sim 阿里云 Ubuntu 部署指令（Agent 执行手册）

> **目标**：在阿里云 Ubuntu ECS 上，用 Docker Compose 部署 life-sim，公网仅暴露 80 端口，启用「进门密码 + 用户登录」双重鉴权。
>
> **执行者**：服务器上的 Agent。请按顺序执行，每步验证通过后再继续。

---

## 0. 部署前向用户确认的信息

执行前必须拿到以下信息（缺一项则暂停并询问用户）：

| 变量 | 说明 | 示例 |
|------|------|------|
| `PUBLIC_IP` | 服务器公网 IP | `123.45.67.89` |
| `DEEPSEEK_API_KEY` | DeepSeek API 密钥 | `sk-...` |
| `SITE_ACCESS_CODE` | 小团体共享进门密码 | 用户自定 |
| `AUTH_BOOTSTRAP_USERS` | 2–3 个初始账号 | `user1:pass1,user2:pass2` |
| `PROJECT_DIR` | 代码目录（默认 `/opt/life-sim`） | `/opt/life-sim` |
| 代码来源 | Git 仓库 URL **或** 代码已在服务器某路径 | — |

可选：若用户已有 JWT 密钥则使用；否则 Agent 自行生成。

---

## 1. 环境检查

```bash
# 系统
uname -a
lsb_release -a 2>/dev/null || cat /etc/os-release

# 公网 IP（与用户提供的 PUBLIC_IP 对照）
curl -s --max-time 5 ifconfig.me || curl -s --max-time 5 ip.sb

# Docker
docker --version
docker compose version
```

**若 Docker 未安装**，执行：

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg

sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo systemctl enable docker
sudo systemctl start docker
```

**若 80 端口被占用**：

```bash
sudo ss -tlnp | grep ':80 '
```

- 若有其他 nginx/apache 占用，先 `sudo systemctl stop nginx apache2 2>/dev/null` 或告知用户改 docker-compose 端口映射。

**防火墙（若 ufw 启用）**：

```bash
sudo ufw status
sudo ufw allow 22
sudo ufw allow 80
```

---

## 2. 获取代码

设 `PROJECT_DIR=/opt/life-sim`（或用户指定路径）。

### 方式 A：Git 克隆

```bash
sudo mkdir -p /opt
cd /opt
sudo git clone <REPO_URL> life-sim
cd /opt/life-sim
```

### 方式 B：代码已在服务器

```bash
cd /opt/life-sim   # 改为实际路径
ls -la
# 应能看到 docker-compose.yml、backend/、frontend/
```

---

## 3. 编写 `.env`

```bash
cd /opt/life-sim
cp .env.example .env
```

生成 JWT 密钥：

```bash
JWT_SECRET=$(openssl rand -hex 32)
echo "JWT_SECRET=$JWT_SECRET"
```

用以下模板写入 `.env`（**替换占位符，勿留 example 值**）：

```env
DEEPSEEK_API_KEY=<用户提供的 DEEPSEEK_API_KEY>
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL_FAST=deepseek-v4-flash
DEEPSEEK_MODEL_HEAVY=deepseek-v4-pro

AUTH_ENABLED=true
JWT_SECRET=<生成的随机串>
SITE_ACCESS_CODE=<用户提供的进门密码>
AUTH_BOOTSTRAP_USERS=<user1:pass1,user2:pass2,user3:pass3>
AUTH_ACCESS_TTL=2h
AUTH_REFRESH_TTL=720h
AUTH_GATE_TTL=24h
AUTH_RATE_LIMIT=5
AUTH_COOKIE_SECURE=false

CORS_ORIGINS=http://<PUBLIC_IP>
```

写入示例（Agent 用 heredoc）：

```bash
cat > /opt/life-sim/.env << 'EOF'
DEEPSEEK_API_KEY=替换
AUTH_ENABLED=true
JWT_SECRET=替换
SITE_ACCESS_CODE=替换
AUTH_BOOTSTRAP_USERS=替换
CORS_ORIGINS=http://替换公网IP
AUTH_COOKIE_SECURE=false
EOF
chmod 600 /opt/life-sim/.env
```

**禁止**：将 `.env` 提交到 Git；在日志中打印完整 API Key。

---

## 4. 阿里云安全组提醒（Agent 需告知用户）

请确认阿里云 ECS **安全组入方向**已配置：

| 端口 | 策略 |
|------|------|
| 80 | 允许 0.0.0.0/0 |
| 22 | 建议仅管理员 IP |
| 8080 | **不要**对公网开放 |

Agent 无法改阿里云控制台，若外网访问失败，提示用户检查安全组。

---

## 5. 构建并启动

```bash
cd /opt/life-sim
docker compose down 2>/dev/null || true
docker compose up --build -d
```

等待构建完成（首次约 3–10 分钟），检查：

```bash
docker compose ps
```

期望：`backend`、`frontend` 均为 `running`。

查看后端日志：

```bash
docker compose logs backend --tail 50
```

期望包含：

- `life-sim 服务启动，监听 :8080`
- `鉴权已启用`（AUTH_ENABLED=true 时）
- 可选：`已 bootstrap N 个用户`

---

## 6. 本机验证

```bash
# 健康检查
curl -s http://127.0.0.1/health

# 鉴权状态（应 enabled:true）
curl -s http://127.0.0.1/api/v1/auth/status

# 未带 Cookie 访问受保护 API 应 401
curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1/api/v1/history
# 期望: 401

# 8080 不应被 host 直接监听（backend 仅 expose，不 publish）
curl -s --max-time 2 http://127.0.0.1:8080/health && echo "WARN:8080 exposed" || echo "OK:8080 not on host"
```

### 完整鉴权链路测试（可选）

```bash
GATE_CODE="<SITE_ACCESS_CODE>"

# 1. 进门
curl -s -c /tmp/lf_cookies.txt -X POST http://127.0.0.1/api/v1/auth/gate \
  -H "Content-Type: application/json" \
  -d "{\"code\":\"$GATE_CODE\"}"

# 2. 登录（替换用户名密码）
curl -s -b /tmp/lf_cookies.txt -c /tmp/lf_cookies.txt -X POST http://127.0.0.1/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"pass1"}'

# 3. 访问业务 API
curl -s -b /tmp/lf_cookies.txt http://127.0.0.1/api/v1/auth/me
```

期望：`/auth/me` 返回用户名；`/history` 返回 code 0。

---

## 7. 交付给用户

部署成功后，向用户报告：

```
访问地址：http://<PUBLIC_IP>
进门密码：<SITE_ACCESS_CODE>
初始账号：<AUTH_BOOTSTRAP_USERS 中的用户名，密码告知用户自行保管>

流程：打开网址 → 输入进门密码 → 登录 → 使用应用
```

并提醒：

1. 当前为 **HTTP 纯 IP**，密码明文传输，仅供小范围自用
2. 备案有域名后，参考 DEPLOY.md 启用 HTTPS，设置 `AUTH_COOKIE_SECURE=true`
3. 数据持久化在 Docker 卷 `lifesim-data`，`docker compose down` 不会删数据
4. 更新：`git pull && docker compose up --build -d`（详见 [docs/OPS-SERVER.md](./docs/OPS-SERVER.md) 全量流程；紧急可先 scp 热更）

---

## 8. 常用运维命令

```bash
cd /opt/life-sim

# 日志
docker compose logs -f backend
docker compose logs -f frontend

# 重启
docker compose restart

# 停止
docker compose down

# 重新构建
docker compose up --build -d
```

---

## 9. 故障排查

| 现象 | 排查 |
|------|------|
| 外网无法访问 | 阿里云安全组是否放行 80；`docker compose ps`；`curl localhost` |
| 502 / 空白页 | `docker compose logs frontend backend`；backend 是否 crash |
| 配置加载失败 | `.env` 是否缺 `JWT_SECRET` / `SITE_ACCESS_CODE`（AUTH_ENABLED=true 时必填） |
| 进门/登录 429 | 限流触发，等 1 分钟重试 |
| AI 调用失败 | 检查 `DEEPSEEK_API_KEY`；服务器能否 `curl https://api.deepseek.com` |
| 80 端口冲突 | `ss -tlnp \| grep :80`；停掉占用进程或改 compose 端口 |

---

## 10. Agent 禁止事项

- **不要** 在 docker-compose 中将 backend `8080` 映射到 `0.0.0.0:8080`
- **不要** 将 `.env` 提交 Git 或粘贴到公开聊天
- **不要** 修改 plan 文件
- **不要** 在生产环境设置 `AUTH_ENABLED=false`
- **不要** 在没有用户确认的情况下覆盖已有 `lifesim-data` 卷

---

## 11. 完成检查清单

- [ ] Docker 与 compose 可用
- [ ] `.env` 已配置且 `chmod 600`
- [ ] `docker compose ps` 两容器 running
- [ ] `curl http://127.0.0.1/health` 正常
- [ ] 未鉴权访问 `/api/v1/history` 返回 401
- [ ] 8080 未在宿主机公网暴露
- [ ] 已向用户提供访问 URL、进门密码、账号说明
