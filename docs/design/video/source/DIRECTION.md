# 影片方向 · Ech0 品牌宣传片（English, 97.5s, 16:9）

## 1. 参考拆解
用户未提供参考视频。

## 2. 产品气质
- **做什么 / 给谁 / 感觉**：自托管的个人微博客时间线。给想在自己域名上写短内容、发照片和链接、又不想依附平台的个人。感觉安静、克制、温暖、有主人感（README："A self-hosted personal microblog where your timeline can be shared, discussed, and fully owned."；作者在时间线上写过"克制比堆功能更难"）。
- **设计语言**（来源 `web/src/themes/tokens/*.scss`）：暖纸色画布 `#f4f1ec`，白色卡片表面，深褐正文 `#33251d`，次级文字 `#5b4f46`，强调色焦橙 `#b84200`；展示字体 `Iowan Old Style`（`--font-family-display`），界面正文用系统无衬线；圆角 0.375–0.875rem，阴影很轻。另有 dark 与 sunny 两套主题（`semantic.dark.scss` / `semantic.sunny.scss`）。
- **母题候选**：
  - 标志几何（`web/public/Ech0.svg`）：橙色圆点、短横、斜线、底横四个笔画，拼成一张安静的"脸"。
  - 核心界面：时间线里每条 Echo 都以一颗**焦橙日期圆点**开头，由一根竖线串起来，与标志里的橙点同源。
  - 产品名的字面意象：Echo，回声。
  - 头部标题的打字机光标（`HomeHeader.vue`：`home-header__cursor`）。

## 3. 三个方向
| 方向 | 底色与光 | 字体声音 | 母题来源 | 镜头语言 | 节奏 | 音乐 |
|---|---|---|---|---|---|---|
| A · Echo 回声 | 产品原生纸白，无发光 | Iowan Old Style 衬线大字，编辑感 | 名字意象 + 时间线橙点 | 真实界面微距，沿时间线竖向推进，匹配剪辑 | 安静叙事 | 钢琴 + 轻脉冲 |
| B · Own it | dark 主题黑舞台，橙点发光 | 粗重无衬线冲击 | 标志四笔画拆装 | 硬切卡拍，3D 舞台 | 预告片式紧凑 | 电子脉冲 |
| C · Talk to your timeline | 暖白 + 等宽技术字 | 等宽技术感 | Copilot 活动时间线 | 一镜到底完成一件任务 | 从容 | 氛围铺底 |

## 4. 选择与理由

> 第二轮修改（用户反馈）：加长到 80 秒，新增管理后台章节（mock 数据），全程保持 light 主题。
> 第三轮修改（用户反馈）：发布后点时间进入 Echo 详情页，演示点赞、分享、评论这些读者互动；片长 92.5 秒。
> 第四轮修改（用户反馈）：写作章节演示附件上传——把本地实例里的一段视频拖进编辑器上传，发布后的 Echo 带着视频；片长 97.5 秒。

- **选 A**。用户要的是面向全量产品的营销片（不是更新日志），平台面只拍 Web 桌面端，风格沿用产品设计。A 让产品名、配色和核心界面落在同一个母题里：一条想法写下 → 发布 → 被读者、订阅和 Copilot "回响" → 最终仍然属于你。B 的暗色和冲击感与产品的克制气质相悖；C 只讲 AI，覆盖面太窄。
- **本片专属手法**：
  1. **一颗橙点贯穿全片**：因为时间线上每条 Echo 都以焦橙日期圆点开头、标志里也有同一颗橙点，所以每个镜头都从一颗橙点开始或落到一颗橙点上。开场的点变成时间线上的点，片尾的点变成标志的"眼睛"。
  2. **回声环**：因为产品名是 Echo，所以发布、被阅读、被引用时，从日期圆点荡出 1–3 圈细线同心环（1.5px，焦橙，透明度递减）。在 Reach 镜头里，每圈环碰到一个去处（RSS、评论、Ech0 Hub、Copilot），就点亮那个去处。
  3. **沿时间线竖向推进**：因为时间线是一根竖线串起的流，所以镜头之间主要做竖向推移和"推近接续"，不用横向滑动，也不用花哨转场。
  4. **产品自己的打字光标**：因为头部标题本身带打字机光标，所以开场句和结尾句都用这个光标的样式（焦橙细竖条）逐字打出，不另造动效。
  5. **标志由四个笔画拼成**：因为 Ech0.svg 是圆点、短横、斜线、底横组成的一张脸，所以片尾从前面镜头里带出这四个笔画，拼回标志。
- **与本工作区既往影片的区别**：本工作区此前没有用本 skill 做过影片，无重复风险。

## 5. 画面规范
- **画幅**：1920×1080，30fps；底色 `#f4f1ec`（`--color-bg-canvas`）；安全区左右 120px，上下 96px。
- **字号阶梯**：英文主标题 Iowan Old Style 104px / 行高 1.02 / 字距 -0.015em / `#33251d`；英文副标题（说明）系统无衬线 34px / 行高 1.35 / `#5b4f46`，单行不超过 44 字符；标签用等宽 22px 大写，字距 0.12em，颜色 `#8e847d`。本片为英文单语（用户选择），不出中文。
- **产品上镜**：时间线和编辑器整体放大 1.7–1.9 倍，正文 16px → 27–30px（≥ 22px）；Copilot 对话放大 1.36 倍；管理后台放大 1.5 倍（Keep 里导出页 1.24 倍）。全片保持 light 主题（用户要求，2026-09-28 第二轮修改）。
- **动效语法**：入场 0.5–0.7s `expo.out`；镜头推移用 `power2.inOut`，持续 0.9–1.4s；文字入场交替使用逐字模糊（标题）、打字光标（开场 / 结尾）和整行上移（说明），不让所有元素用同一种入场。禁止：发光、粒子、镜头光晕、旋转、弹跳缓动，同屏只允许一组回声环。
- **强调色使用**：焦橙只用于圆点、回声环、光标和一个关键词；其余全部是墨色和纸色。

## 6. 镜头表（96 BPM，1 拍 = 0.625s，1 小节 = 2.5s；镜头边界都落在半小节网格上）
| # | 时间 | 镜头 | 主角 | 画面与动作 | 文案 | 声音 |
|---|---|---|---|---|---|---|
| 1 | 0.0–5.0 | open | 橙点 | 纸白空画面；第 1 拍（0.625s）亮起一颗橙点，产品光标在右侧逐字打出一句；第 5 拍荡出第一圈回声环 | **Say it once.** | 钢琴单音对点；pop、打字、sweep |
| 2 | 5.0–8.75 | echo | 回声环 | 三圈环依次扩到画外，钢琴动机带延迟回声；7.0s 起橙点滑到左侧，长出 rail（本片自己的时间线） | **Let it echo.** | 动机 + 三次延迟回声；resolve、whoosh |
| 3 | 8.75–22.5 | write | 真实编辑器 | 首页 Publish 标签页（真实 TheEditor）放大 1.85 倍；10.0 点附件 → 10.625 选 Video → 光标带着文件卡片从画面右缘进来，12.5 在真实上传区触发 drop → 进度条 12.5–14.4 → 15.0 点返回，视频预览挂在输入框下方（镜头拉到 1.6 倍）；之后真实打字；打开标签选择器选 #notes；点 "+" 打开发布选项，第 35 拍（21.875s）点 "Publish as public" | **Write it down.** / Markdown, photos, links and tags, all from one editor. | 钢琴 ostinato 入；打字、click、pop |
| 4 | 22.5–28.75 | land | 真实时间线 | 真实发布流程跳回时间线，新 Echo（Just now）在第一条；它的日期橙点荡出两圈回声环 | **It lands on your timeline.** / The new post sits on top of your own timeline, on your own domain. | 贝斯在下拍进入；success |
| 5 | 28.75–41.25 | detail | 真实 Echo 详情页 | 29.375 真实点新 Echo 的时间标签进入 /echo/e-new（EchoView）；30.625 点分享 → 真实分享面板（Copy as Markdown、带链接的复制）；31.875 复制链接；32.5 页面以未登录读者视角重新打开（提示条 "A reader opens the shared link"）；33.125 点赞 0 → 1；33.75 展开评论表单，填昵称 Jonas、邮箱和回复，实时预览；37.5 提交，出现 "Comment published" 和 #2 Jonas；40.625 真实点 back 回到时间线；镜头逐帧跟随焦点：页头 → 点赞/分享行（分享面板完整入画）→ 评论表单 → 新评论停留约 1 秒 → 回到页头 | **Pass it on.** / Every Echo has its own page. Readers can like it, share it and reply. | 更轻、更私人的律动（无底鼓）；click、pop、sweep、打字、success |
| 6 | 41.25–48.75 | stream | 真实时间线 | 真实主列滚动：照片、Markdown 笔记、第二张照片、链接卡片 | **One quiet stream.** / Notes, photos and link cards, in the order you lived them. | 钢琴旋律 |
| 7 | 48.75–56.25 | reach | 真实 Status 页 | 镜头拉远，光标点侧栏 Status；rail 橙点荡出三圈环，依次点亮 Connect（53.03s）、RSS（53.43s）、Comments（53.72s）；再推近到可读尺度 | **Echoes that reach people.** / Readers follow by RSS, reply in comments, and connect their own Ech0. | click + 三次轻 tick |
| 8 | 56.25–66.25 | copilot | 真实 Copilot | 真实 TheChatBox：打出问题 → 发送 → 两次检索、覆盖 6 条 Echo → 流式回答 + 引用来源 | **Ask your own timeline.** / Ech0 Copilot answers from your posts, and shows which ones. | 琶音脉冲；打字、whoosh、pop |
| 9 | 66.25–83.75 | panel | 真实管理后台 | 字卡正常进场，后台以全景出现在右侧；68.75s 字卡收成页眉（缩到 0.5、移到顶部，橙点随行），后台推近到全宽 1.5 倍，顶部欢迎区用渐变遮罩收掉。之后全是真实点击：71.25 Comments → Comment Management 审核表；73.75 Storage → File Manager → 展开 Local Storage → 打开 images/；76.25 Extensions → MCP（真实 manifest：29 个工具，endpoint 为 https://mira.example/mcp），镜头下移扫过工具列表；80 Logs，实时日志逐行推入 | **Run it your way.** / A calm admin panel: stats, comments, files, MCP and live logs. | 稳定明亮的律动；每次切页 click，子标签 click-alt，树节点 pop，日志 toggle |
| 10 | 83.75–92.5 | own | 数据导出 + 部署命令 | 字卡回到常规位置；真实点 Data → Export，后台左侧导航和顶部被遮罩收掉，只留 Snapshot / Capsule 导出卡；左下浅色卡片逐字打出 README 的 docker run 命令，87.5s 标出 `/app/data` | **Yours to keep.** / Export a snapshot or a portable capsule any time. It all runs in one container. | 温暖渐强（Bb → C）；click、打字、click-alt |
| 11 | 92.5–97.5 | close | 标志 | 橙点从 rail 飞到标志的"眼睛"；短横、斜线、底横三笔依次飞入，底板淡入；产品光标打出 "Ech0"，接一句话和一行信息 | **Ech0** / A timeline you own. / Open source · Self-hosted · ech0.app | 90s 落 Fmaj9 终止和弦；三次 pop、打字 |
