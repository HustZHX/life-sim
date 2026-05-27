# 开发约定

本文档记录 life-sim 项目的开发与协作约定。

## 服务重启

代码变更后需重启对应服务，改动才会生效：

| 变更范围 | 需重启 | 默认端口 |
|----------|--------|----------|
| `backend/`（Go 代码、prompts、配置） | 后端 | `:8080` |
| `frontend/`（Vue / TS / 样式） | 前端 | `:5173` |

Windows 下常用命令：

```powershell
# 后端（读取项目根目录 .env）
cd backend
go run .
# 或双击 run-backend.bat

# 前端
cd frontend
npm run dev
```

若端口被占用，先结束占用 `:8080` / `:5173` 的旧进程再启动。

## 环境变量

- 复制 `.env.example` 为 `.env`，填入 `DEEPSEEK_API_KEY`
- **勿将 `.env` 提交到 Git**（已在 `.gitignore` 中排除）
- 本地 SQLite 数据库位于 `backend/data/`，同样不提交

## 数据与兼容性

- 旧时间轴节点可能缺少 `entities`（实体标注）或 `scene`（场景元数据），需重新生成或重算后才有
- 后端重启会将进行中的异步任务标记为失败，需在前端重新提交生成

## GitHub

本地已初始化 Git 仓库。首次推送远程：

```powershell
gh auth login
gh repo create life-sim --public --source=. --remote=origin --push
```
