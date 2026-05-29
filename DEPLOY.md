# 公网部署指南（纯 IP）

本文说明如何将 life-sim 部署到云服务器，并通过 **进门密码 + 用户登录** 双重鉴权限制访问。

## 前置条件

- 云服务器（Linux 推荐）已安装 Docker 与 Docker Compose
- 安全组 **仅开放 TCP 80**（后续有域名再开 443）
- **不要** 对公网开放 8080（后端仅 Docker 内网可达）

## 1. 准备环境变量

在项目根目录复制并编辑 `.env`：

```bash
cp .env.example .env
```

生产必填项：

```env
DEEPSEEK_API_KEY=sk-xxx

AUTH_ENABLED=true
JWT_SECRET=至少32位随机字符串
SITE_ACCESS_CODE=小团体共享进门密码
AUTH_BOOTSTRAP_USERS=user1:初始密码,user2:初始密码,user3:初始密码

# 纯 IP 访问时设为 http://你的公网IP（nginx 同源反代时可与访问地址一致）
CORS_ORIGINS=http://123.45.67.89

AUTH_COOKIE_SECURE=false
```

说明：

- `AUTH_BOOTSTRAP_USERS` 仅在对应用户名**不存在**时创建，不会覆盖已有密码
- 本地开发可设 `AUTH_ENABLED=false` 跳过鉴权

## 2. 启动

```bash
docker compose up --build -d
```

访问：`http://你的公网IP`

流程：进门密码 → 登录 → 进入应用

## 3. 安全组建议

| 端口 | 入站 | 说明 |
|------|------|------|
| 80 | 允许 0.0.0.0/0 | 唯一 Web 入口 |
| 8080 | **拒绝** | 后端不对公网 |
| 22 | 仅你的 IP | SSH 管理 |

## 4. 安全提醒（纯 IP + HTTP）

当前阶段无域名、无 HTTPS 时：

- 进门密码与登录密码在网络上 **明文传输**
- 双重鉴权可挡住公网扫描与路人，但 **不等于** 完整传输加密
- 切勿在公共 Wi‑Fi 下使用，或等备案完成后启用 HTTPS

## 5. 后续加域名与 HTTPS

备案完成并绑定域名后：

1. nginx 配置 SSL（Let's Encrypt / certbot）
2. 设置 `AUTH_COOKIE_SECURE=true`
3. 更新 `CORS_ORIGINS=https://你的域名`
4. 安全组开放 443

业务代码与 Cookie 机制无需改动。

## 6. 常用运维

日常 **热更**、**提交推送后的全量部署**、SSH、服务器目录与故障排查见 **[docs/OPS-SERVER.md](./docs/OPS-SERVER.md)**。

```bash
# 查看日志
docker compose logs -f backend

# 重启
docker compose restart

# 代码更新（已 push 到 Git 后）
cd /opt/life-sim && git pull && docker compose up --build -d

# 新增 bootstrap 用户：修改 .env 中 AUTH_BOOTSTRAP_USERS 后重启
# 若用户名已存在，需登录后在应用内改密，或手动改 SQLite
```

## 7. 验证清单

- [ ] 未过进门密码无法访问业务 API
- [ ] 过进门未登录无法访问业务 API
- [ ] 三人不同账号登录看到同一批人物（共享工作区）
- [ ] 公网无法直接访问 `:8080`
- [ ] 连续错误进门/登录触发限流（429）
