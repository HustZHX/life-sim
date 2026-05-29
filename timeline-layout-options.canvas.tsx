import React from "react";
import {
  Card,
  CardBody,
  CardHeader,
  Code,
  Divider,
  Grid,
  H1,
  H2,
  H3,
  Pill,
  Row,
  Spacer,
  Stack,
  Text,
  useHostTheme,
} from "cursor/canvas";

type ApproachId = "A" | "B" | "C";

function MiniAxis({
  title,
  subtitle,
  approach,
}: {
  title: string;
  subtitle: string;
  approach: ApproachId;
}) {
  const theme = useHostTheme();
  const t = theme.tokens;

  const stroke = t.borderSubtle;
  const bg = t.surface;
  const bgMuted = t.surfaceMuted;
  const textMuted = t.textMuted;
  const text = t.text;
  const accent = t.accent;

  const axisStroke = t.border;
  const axisDot = t.textMuted;

  const worldBg = t.surfaceMuted;
  const nodeBg = t.surface;

  const worldBorder = t.borderSubtle;
  const nodeBorder = t.border;

  const w = 84;
  const h = 54;

  const cardBase: React.CSSProperties = {
    border: `1px solid ${stroke}`,
    borderRadius: 10,
    overflow: "hidden",
    background: bg,
  };

  const header: React.CSSProperties = {
    padding: "10px 12px",
    background: bgMuted,
    borderBottom: `1px solid ${stroke}`,
  };

  const body: React.CSSProperties = { padding: 12 };

  const canvas: React.CSSProperties = {
    position: "relative",
    height: 148,
    background: bg,
    border: `1px dashed ${stroke}`,
    borderRadius: 10,
  };

  const axisX = approach === "A" ? 18 : 50;

  const axis: React.CSSProperties = {
    position: "absolute",
    left: `${axisX}%`,
    top: 10,
    bottom: 10,
    width: 2,
    background: axisStroke,
    borderRadius: 999,
  };

  const dot = (y: number, tone: "world" | "node") => ({
    position: "absolute" as const,
    left: `calc(${axisX}% - 4px)`,
    top: y,
    width: 10,
    height: 10,
    borderRadius: 999,
    background: tone === "world" ? axisDot : accent,
    border: `2px solid ${bg}`,
  });

  const box = (x: number, y: number, tone: "world" | "node") => ({
    position: "absolute" as const,
    left: x,
    top: y,
    width: w,
    height: h,
    borderRadius: 10,
    background: tone === "world" ? worldBg : nodeBg,
    border: `1px solid ${tone === "world" ? worldBorder : nodeBorder}`,
    display: "flex",
    flexDirection: "column" as const,
    justifyContent: "center",
    padding: "0 10px",
    color: text,
  });

  const label: React.CSSProperties = { fontSize: 11, color: textMuted, marginTop: 2 };
  const titleStyle: React.CSSProperties = { fontSize: 12, fontWeight: 650 as const, lineHeight: 1.2 };

  // A: 仍用 el-timeline，左侧“世界线”只能在内容区左列（轴线仍在最左）
  // B: 自绘中轴线（推荐），世界线/人生真正左右分布
  // C: 两条轴（左右各一条）或上下双轨：改动大但更灵活
  const leftX = approach === "A" ? 34 : 6;
  const rightX = approach === "A" ? 34 + w + 10 : 56;

  const worldX = approach === "C" ? 6 : leftX;
  const nodeX = approach === "C" ? 56 : rightX;

  const axisNote =
    approach === "A"
      ? "轴线固定在左侧"
      : approach === "B"
        ? "中轴线可居中"
        : "可双轴/双轨";

  return (
    <div style={cardBase}>
      <div style={header}>
        <Row align="center" justify="space-between">
          <Stack gap={2}>
            <Text weight="semibold">{title}</Text>
            <Text muted style={{ fontSize: 12 }}>
              {subtitle}
            </Text>
          </Stack>
          <Pill tone={approach === "B" ? "accent" : "neutral"}>{`方案 ${approach}`}</Pill>
        </Row>
      </div>
      <div style={body}>
        <div style={canvas}>
          <div style={axis} />
          <div style={dot(22, "world")} />
          <div style={dot(84, "node")} />
          <div style={dot(116, "world")} />

          <div style={box(worldX, 14, "world")}>
            <div style={titleStyle}>世界线事件</div>
            <div style={label}>默认只标题</div>
          </div>

          <div style={box(nodeX, 76, "node")}>
            <div style={titleStyle}>人生节点</div>
            <div style={label}>可展开内容</div>
          </div>

          <div style={box(worldX, 108, "world")}>
            <div style={titleStyle}>世界线事件</div>
            <div style={label}>点开看详情</div>
          </div>

          <div
            style={{
              position: "absolute",
              right: 10,
              bottom: 10,
              fontSize: 11,
              color: textMuted,
            }}
          >
            {axisNote}
          </div>
        </div>
      </div>
    </div>
  );
}

export default function Canvas() {
  const theme = useHostTheme();
  const t = theme.tokens;

  return (
    <div
      style={{
        padding: 18,
        color: t.text,
        background: t.background,
        fontFamily: t.fontFamily,
      }}
    >
      <H1>时间轴左右分布（桌面/手机）改造方案</H1>
      <Text muted>
        目标：桌面端“世界线在轴线左侧、人生在右侧”；手机端可上下堆叠；同时世界线卡片在两端都更小、更淡，和人生节点明显区分。
      </Text>

      <Spacer size="m" />

      <Grid columns={3} gap="m">
        <MiniAxis
          approach="A"
          title="A：继续用 Element Plus 时间线"
          subtitle="改动最小，但轴线永远在最左"
        />
        <MiniAxis
          approach="B"
          title="B：自绘中轴线 + 双侧卡片（推荐）"
          subtitle="真正左右分布；兼容性最好"
        />
        <MiniAxis
          approach="C"
          title="C：双轴/双轨（高级）"
          subtitle="信息密度更高，但改造成本最大"
        />
      </Grid>

      <Spacer size="m" />
      <Divider />
      <Spacer size="m" />

      <Grid columns={2} gap="m">
        <Card>
          <CardHeader title="样式区分建议（两端一致）" />
          <CardBody>
            <Stack gap="s">
              <Row align="center" justify="space-between">
                <Text weight="semibold">世界线事件卡片（建议）</Text>
                <Pill tone="neutral">更小、更淡</Pill>
              </Row>
              <Text muted>
                - 标题字号略小、行高更紧
                {"\n"}- 背景更淡（surfaceMuted），左边色条更窄
                {"\n"}- 默认只标题；展开后才显示详情段落
              </Text>
              <Spacer size="s" />
              <Row align="center" justify="space-between">
                <Text weight="semibold">人生节点卡片（保持）</Text>
                <Pill tone="accent">主视觉</Pill>
              </Row>
              <Text muted>
                - 维持现有尺寸/层级
                {"\n"}- 展开后显示经历/内心/性格等
              </Text>
            </Stack>
          </CardBody>
        </Card>

        <Card>
          <CardHeader title="我建议选 B 的原因" />
          <CardBody>
            <Stack gap="s">
              <Text>
                现在你看到“都在一侧”，本质是 <Code>el-timeline</Code> 的轴线结构固定：圆点/竖线在最左，内容只能在右侧区域布局。
              </Text>
              <Text muted>
                选 B 后我们完全掌控 DOM：中轴线居中，世界线与人生真正分列；手机端用一个媒体查询直接堆叠即可。
              </Text>
            </Stack>
          </CardBody>
        </Card>
      </Grid>
    </div>
  );
}

