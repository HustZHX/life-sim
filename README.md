# Life-Sim — AI 人生模拟

基于 DeepSeek V4 的历史名人 / 随机人物人生时间轴生成与编辑回溯系统。

## 功能

- **名人模式**：输入姓名 → AI 消歧 → 确认身份 → 生成档案 → 生成约 3 年一节的人生节点
- **随机模式**：AI 随机生成虚构人物，可选导入名人性格与时代背景
- **编辑级联**：修改某一节点后，后续节点由 AI 重新推演，并展示变更 diff
- **版本回溯**：查看历史版本、对比差异、一键回滚

## 技术栈

- 后端：Go 1.24 + Gin + SQLite
- 前端：Vue 3 + TypeScript + Vite + Element Plus + Pinia
- AI：DeepSeek V4（`deepseek-v4-flash` / `deepseek-v4-pro`）

## 快速开始

### 1. 配置环境变量

复制 `.env.example` 为 `.env` 并填入 `DEEPSEEK_API_KEY`：

```bash
cp .env.example .env
```

### 2. 启动后端

```powershell
cd backend
$env:DEEPSEEK_API_KEY="sk-xxx"
go run .
```

默认监听 `http://localhost:8080`

### 3. 启动前端

```powershell
cd frontend
npm install
npm run dev
```

访问 `http://localhost:5173`

## 开发约定

开发与协作约定（服务重启、环境变量、数据兼容性等）见 [DEVELOPMENT.md](./DEVELOPMENT.md)。

## API 概览

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/characters` | 创建角色 `{mode: "famous"\|"random"}` |
| POST | `/api/v1/characters/:id/resolve` | 名人消歧 |
| POST | `/api/v1/characters/:id/confirm` | 确认候选 |
| POST | `/api/v1/characters/:id/profile/generate` | 生成档案 |
| GET | `/api/v1/models` | 可选模型列表（`flash` / `pro`） |
| POST | `/api/v1/characters/:id/timeline/generate` | 异步生成时间轴（body: `{model}`） |
| GET | `/api/v1/jobs/:jobId` | 查询任务进度 |
| PATCH | `/api/v1/characters/:id/nodes/:nodeId` | 编辑节点并级联重算 |

## Docker

```bash
docker compose up --build
```

公网部署（双重鉴权、纯 IP 访问）见 [DEPLOY.md](./DEPLOY.md)。已上线后的 **热更 / 全量发布 / 服务器运维** 见 [docs/OPS-SERVER.md](./docs/OPS-SERVER.md)。

## 鉴权（可选）

设置 `AUTH_ENABLED=true` 后启用「进门密码 + 用户登录」。本地开发默认 `AUTH_ENABLED=false`，详见 `.env.example`。

## 免责声明

AI 生成内容为演绎性质，不构成历史学术考证。名人资料可能存在幻觉，请以权威史料为准。
