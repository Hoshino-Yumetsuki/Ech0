# 视频 brief · Ech0 品牌宣传片

- **状态**：正式片（`plan.demo = false`）。
- **用户决定**（2026-09-28）：
  - 范围：只针对当前最新产品（v5.7.0），不区分功能是哪个版本新增的；用途是产品宣传营销，不是更新日志。
  - 平台面：Web 桌面端（公开站点 + 管理后台）。
  - 风格：沿用产品设计（`repo`）。
  - 规格：横版 16:9 · 英文。第一版 60 秒；第二轮加长并演示管理后台（80 秒）；第三轮加入 Echo 详情页的读者互动（点赞、分享、评论，92.5 秒）；第四轮在写作章节演示附件上传（用户指定的本地视频 data/files/videos），现为 97.5 秒。
  - 主题：全程 light（第二轮用户要求，去掉了 Keep 章节的暗色切换）。
  - 方向：用户未指定，采用推荐的 A「Echo 回声」（见 DIRECTION.md）。
- **plan.scope**：`{"versions":"current product (v5.7.0)","platforms":["web-desktop"],"purpose":"brand promo, not release notes"}`
- **未覆盖的平台面**：TUI / CLI、移动端与 PWA、外部 MCP 客户端的实际调用、Webhook、S3 存储配置。管理后台中只演示了仪表盘、评论管理、文件管理、MCP、系统日志、数据导出。片中的 docker 命令是 README 原文字卡，不是 CLI 界面演示。
- **链接 / CTA**：用户未作限制；片尾只写产品官网 `ech0.app`（README 顶部链接）。演示数据里的网址都用 `example.com` / `.example`。
- **仓库**：本仓库根目录（`../../../..`；制作时只读，未改动产品代码）。
- **品牌资源**：`web/public/Ech0.svg`（片尾按其几何重绘为 SVG 分件动画）、主题 tokens、`--font-family-display`（Iowan Old Style）。
- **声音**：代码原创配乐（`tools/score.py`），音效用 skill 内置 WAV（本机没有录音素材库，见 evidence/audio-selection.json）。
- **证据**：evidence/feature-evidence.md、style-audit.md、component-usage.json、audio-selection.json、copy-review.md。
