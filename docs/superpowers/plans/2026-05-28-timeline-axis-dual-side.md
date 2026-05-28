# Timeline Axis 双侧布局 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 桌面端时间轴实现“世界线事件在轴线左侧、人生节点在轴线右侧”的真实双侧布局；同时在桌面/手机端调整世界线卡片样式尺寸，使其与人生节点明显区分但不抢主视觉。

**Architecture:** 替换 `TimelineAxis.vue` 中对 Element Plus `el-timeline`（轴线固定左侧）的依赖，改为自绘“中轴线 + 双侧卡片”的 DOM/CSS。保持现有数据排序（`axisItems`）与交互（世界线展开/收起、人生节点展开/收起、节点定位 `id=node-<id>`）。

**Tech Stack:** Vue 3 + Element Plus（继续使用 `el-card`/`el-button`）+ scoped CSS。

---

## File Structure（变更范围）

- Modify: `frontend/src/components/TimelineAxis.vue`
  - 用自绘轴替代 `el-timeline`
  - 新增桌面端三列网格：左栏（世界线）/ 中轴（竖线+圆点+时间戳）/ 右栏（人生节点）
  - 响应式：手机端堆叠为单列（世界线在上、人生在下），轴线缩到左侧或切换为“左侧细线+圆点”
  - 世界线卡片视觉：更小、更淡（背景/边框/字号/padding）

- (Optional) Modify: `frontend/src/views/TimelineView.vue`
  - 若需要更明显区分：在“世界线简介”折叠块里添加轻量提示（不影响本次核心）

## Task 1：实现自绘双侧布局（桌面端略偏右）

**Files:**
- Modify: `frontend/src/components/TimelineAxis.vue`

- [ ] **Step 1: 读取并确认现有 `axisItems` 输出与 key**
  - 保持排序：按 year 升序；同年世界线在前；人生节点按 sequence

- [ ] **Step 2: 用新 DOM 替换 `el-timeline`**
  - 目标结构（伪代码）：

```vue
<div class="axis">
  <div class="axis-row" v-for="item in axisItems">
    <div class="left">
      <WorldCard v-if="item.kind==='world'" />
    </div>
    <div class="mid">
      <div class="dot" />
      <div class="time">{{ ... }}</div>
    </div>
    <div class="right">
      <NodeCard v-if="item.kind==='node'" :id="node-${id}" />
    </div>
  </div>
</div>
```

- [ ] **Step 3: CSS 实现“略偏右”的轴线位置**
  - 桌面端：`grid-template-columns: minmax(180px, 0.9fr) 56px minmax(360px, 1.4fr)`
  - 中轴列负责绘制竖线（可用伪元素 `::before`）
  - dot 居中；timestamp 放在 dot 右侧（或上方），避免占左右栏宽度

- [ ] **Step 4: 手机端堆叠**
  - `@media (max-width: 960px)`：
    - `grid-template-columns: 24px 1fr`
    - 轴线退到最左（中轴列变成左侧细线列）
    - 世界线卡片与人生节点卡片在同一列上下堆叠（同一条 row 内：world 在上、node 在下；或 world row/node row 各自一行）

- [ ] **Step 5: 保持交互与定位不回归**
  - 人生节点 `:id="node-${item.node.id}"` 不变
  - `@click` 仍 `emit('select', node)`
  - 世界线卡片 `toggleWorld(id)` 保持

- [ ] **Step 6: 本地构建验证**
  - Run: `cd frontend && npm run build`
  - Expected: build 成功；桌面端左右分布明显；手机端堆叠且不溢出

## Task 2：世界线卡片“更小、更淡”的样式区分（桌面/手机统一）

**Files:**
- Modify: `frontend/src/components/TimelineAxis.vue`

- [ ] **Step 1: 调整世界线卡片尺寸与层级**
  - 标题字号略小（例如 `0.88rem`）
  - padding 更紧（减少 card body 内边距）
  - 背景更淡（保持白底但加淡色边/淡底）

- [ ] **Step 2: 视觉区分但不抢主视觉**
  - 世界线：细色条 + muted 文本；展开时提示色（warning）但不高亮到超过 node
  - 人生节点：维持现有 active/expanded 高亮

- [ ] **Step 3: 本地构建验证**
  - Run: `cd frontend && npm run build`
  - Expected: 通过；世界线与人生节点一眼可区分

## Task 3：热更新到服务器

**Files:**
- Deploy: `frontend/src/components/TimelineAxis.vue`（必要时连同 `TimelineView.vue` 一起）

- [ ] **Step 1: 用 scp 覆盖服务器文件**
  - Copy to: `root@47.108.84.163:/opt/life-sim/frontend/src/components/TimelineAxis.vue`

- [ ] **Step 2: 服务器重建并重启**
  - Run (server): `cd /opt/life-sim && docker compose up --build -d`

- [ ] **Step 3: 健康检查**
  - Run (server): `curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1/health`
  - Expected: `200`

