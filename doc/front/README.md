# Noviterra 前端

基于 Nuxt 3、Vue 3 与 TypeScript 的 Noviterra 前端。界面按访问角色和操作范围划分：访客使用公开站点，登录用户使用账号空间，管理员与超级管理员使用管理后台。

## 视觉与交互

界面采用编辑型博客方向：大图首屏、米白纸面、森林墨色与单一青色强调，结构接近 Linear 的克制层级，阅读区接近 Apple Notes 的安静纸面。视觉参数固定为 `DESIGN_VARIANCE: 6`、`MOTION_INTENSITY: 4`、`VISUAL_DENSITY: 3`，优先保证内容密度、可读性与低干扰动效。

- 首页以全幅站点照片作为情绪入口，正文区用圆角纸面承接文章流；没有数据时不渲染空公告或空标签模块。
- 文章广场保留列表与卡片两种视图；卡片在封面左上角显示日期章，悬停时显示阅读入口。
- 阅读页保持单栏正文、右侧目录与低对比度元信息，代码块、提示块、表格和图片仍由统一 `.prose` 样式接管。
- 友链页使用深绿青色本站信息卡、GitHub 与邮箱卡片；登录后才显示申请表单。
- 所有页面共用同一套语义令牌，浅色与深色主题只切换令牌，不在组件内散落颜色值。

## 背景与主题

- 右上角图片按钮提供 `林间`、`纸面`、`雾蓝` 三种背景；默认使用 `林间`，即项目根目录 `static/station/default.jpg`。
- 背景选择保存在 `localStorage` 的 `noviterra.background`，主题保存在 `noviterra.theme`。
- 首屏内联脚本在绘制前恢复主题与背景，避免刷新闪烁；插件在 `app:mounted` 后同步 Vue 状态，保证 SSR 首次渲染与客户端一致。
- 背景状态写入 `html[data-bg]`，主题写入 `html[data-theme]`；样式全部由 `assets/css/backgrounds.css` 与 `assets/css/tokens.css` 驱动。
- `/station/default.jpg` 由 `server/routes/station/[...path].get.ts` 读取，默认目录为项目根目录的 `static/station`，可用 `NOVITERRA_STATION_DIR` 覆盖。

## 启动

环境要求：Node.js 20+，后端服务已启动。

```powershell
cd front
npm install
npm run dev
```

默认地址：`http://localhost:3000`。开发代理把 `/novblog/*` 转发到本机微服务：

| 服务 | 端口 |
| --- | --- |
| identity | 8001 |
| notice | 8002 |
| user | 8003 |
| article | 8004 |
| comment | 8005 |
| site | 8006 |
| agent | 8007 |

如果后端不在本机，可在启动前设置：

```powershell
$env:NUXT_PUBLIC_API_BASE="http://后端地址:端口"
npm run dev
```

生产构建：

```powershell
npm run build
npm run preview
```

如果开发服务器正在运行，Nuxt 会阻止第二次构建。可在当前终端临时绕过锁：

```powershell
$env:NUXT_IGNORE_LOCK="1"
npm run build
```

构建会重新生成 `.nuxt` 并触发开发服务器重启，这是 Nuxt 的正常行为。

## 角色与页面

### 公开站点

| 路由 | 功能 | 数据来源 |
| --- | --- | --- |
| `/` | 首页、最新文章、公告、标签与站点入口 | article、notice |
| `/articles` | 文章广场、分页、列表/卡片视图、标签筛选 | article |
| `/articles/:id` | 文章正文、目录、封面、点赞、收藏、评论 | article、comment |
| `/archive` | 按年份归档；用近期文章补全可跳转的文章 ID | article |
| `/guestbook` | 站点留言与回复 | comment |
| `/friends` | 本站信息、联系方式；登录后提交友链申请 | site |
| `/assistant` | 站点助手对话，本地保存最近消息 | agent |
| `/about` | 站点、服务与技术栈说明 | 本地配置 |
| `/search` | 服务端文章搜索 | article |
| `/login` | 账号密码或邮箱验证码登录 | identity |
| `/register` | 注册账号 | identity、notice |
| `/forgot-password` | 邮箱验证码重设密码 | identity、notice |

### 账号空间

| 路由 | 功能 | 数据来源 |
| --- | --- | --- |
| `/me/profile` | 昵称、头像、年龄、性别、简介 | user |
| `/me/password` | 旧密码或邮箱验证码修改密码 | user、notice |
| `/me/library` | 当前用户点赞与收藏的文章 | article |
| `/me/articles` | 当前用户发布的文章；管理员可进入写作页 | article |

### 管理后台

| 路由 | 功能 | 数据来源 |
| --- | --- | --- |
| `/admin` | 工作台、近期内容、服务状态入口 | article、user、notice |
| `/admin/articles` | 文章列表、筛选、预览、编辑、删除 | article |
| `/admin/articles/new` | 富文本写作与封面生成/上传 | article |
| `/admin/articles/:id` | 编辑文章、替换封面、删除 | article |
| `/admin/users` | 用户分页、封禁与解封 | user |
| `/admin/notices` | 公告列表与发布 | notice |
| `/admin/assistant` | 助手模型配置与接口测试 | agent |

`/admin/**` 与 `/me/**` 关闭 SSR，并在客户端通过 `auth`、`admin` 中间件校验会话。普通用户访问后台会回到首页；未登录用户会跳到登录页并携带 `redirect`。

## 目录结构

```text
front/
├─ pages/                 # 文件路由，按公开站点、账号空间、后台分组
├─ layouts/               # default、auth、account、admin
├─ middleware/            # auth、admin、guest
├─ components/
│  ├─ site/               # 页头、用户菜单、账号导航、页脚
│  ├─ article/            # 正文、卡片、行、目录、评论
│  ├─ editor/             # RichEditor、ArticleEditor、CoverPicker
│  ├─ admin/              # 后台侧栏与顶栏
│  └─ ui/                 # 按钮、图标、弹窗、封面、空状态、Toast
├─ composables/           # API、鉴权、会话、外观、各服务数据访问
├─ utils/                 # 类型、常量、格式化、封面、富文本渲染
├─ assets/css/            # 设计令牌与模块化样式
├─ server/routes/static/  # 文章封面文件读取
├─ server/routes/station/ # 站点背景文件读取
└─ public/                # favicon 与默认头像
```

### CSS 模块

| 文件 | 职责 |
| --- | --- |
| `tokens.css` | 浅色/深色语义令牌、字体、圆角、阴影与动效曲线 |
| `base.css` | 全局重置、滚动条、页面氛围层与应用外壳 |
| `backgrounds.css` | 三种背景、首屏叠色与深色背景变体 |
| `layout.css` | 页头、页脚、页面宽度与通用页面标题 |
| `components.css` | 按钮、表单、下拉、标签、Toast、主题切换等通用组件 |
| `editorial.css` | 首页、文章广场、阅读页、归档、留言、关于、友链的编辑型布局 |
| `prose.css` | 文章正文、代码块、引用、表格与图片排版 |
| `editor.css` | 富文本编辑器工具栏、斜杠菜单与写作界面 |
| `pages.css` | 登录、注册、账号页等页面补充样式 |
| `admin.css` | 后台侧栏、数据面板与表格样式 |

## API 约定

- 所有请求收口到 `composables/useApi.ts`，统一处理 JSON、错误信息、Bearer Token、401 单飞刷新与 8 秒超时。
- 所有服务地址集中在 `utils/constants.ts` 的 `ENDPOINTS`。
- 登录 Token 保存在 `nv_access`、`nv_refresh` Cookie；用户资料缓存键为 `noviterra.user`。
- 后端统一响应可能是 `response` 或 `resp` 包装，`useApi` 同时兼容 `status`、`errcode`、`code`。
- 管理端分页接口不返回总数；是否还有下一页通过“本页返回条数是否等于 page size”推断。

### 服务映射

| Composable | 服务 | 主要能力 |
| --- | --- | --- |
| `useAuth` | identity / notice | 注册、登录、验证码、忘记密码、资料与密码修改 |
| `useSession` | user | 当前用户、角色判断、本地缓存恢复 |
| `useArticles` | article | 列表、详情、搜索、归档、个人文章、点赞收藏列表、发布、编辑、封面、删除 |
| `useComments` | comment | 顶层评论、回复树、评论、点赞、收藏 |
| `useUsers` | user | 管理员用户分页、封禁、解封 |
| `useNotices` | notice | 公告列表与发布 |
| `useFriends` | site | 友链申请 |
| `useAgent` | agent | 助手配置与对话 |

## 文章封面

- 发布文章和更新封面都必须使用 `multipart/form-data`。
- 发布字段：`article_id`、`title`、`content`、`information`、`tags`（可重复）与 `cover` 文件。
- 支持 JPG、JPEG、PNG、WebP，最大 5 MB。
- 未选择图片时，前端生成 1200×675 的 JPEG 文字封面。
- 后端保存为 `cover{articleID}time{unix}.{ext}`，前端通过 `/static/articlecover/{filename}` 读取。
- Nuxt 服务端路由默认读取项目根目录的 `static/articlecover`；可用 `NOVITERRA_COVER_DIR` 覆盖。

## 富文本编辑器

`components/editor/RichEditor.vue` 输出经过白名单清洗的 HTML，正文在阅读页通过统一 `.prose` 样式渲染。

支持：

- 多级标题、正文段落、有序/无序列表、引用、分割线
- 粗体、斜体、下划线、删除线、行内代码、高亮
- 带语言标记的代码块与自动语法高亮、复制按钮
- 提示块、表格、图片与图注、链接
- 斜杠命令菜单
- Markdown 快捷输入：`#`、`##`、`###`、`>`、`-` / `*` / `+`、`1.`、`---`、```` ``` ````
- `Tab` / `Shift+Tab` 缩进与反缩进
- 富文本粘贴清洗，禁止脚本、事件属性与危险协议

`components/editor/ArticleEditor.vue` 负责文章元数据、草稿与发布：

- 本地草稿自动保存，重新进入时提示恢复或还原
- `Ctrl/Cmd + S` 保存
- 标题最多 20 字，摘要最多 200 字，标签最多 5 个且每个最多 10 字，正文最多 50000 字符
- 这些限制与后端 RPC 的实际运行校验保持一致

## 后端契约注意点

前端已适配当前后端的真实行为，维护时不要只依据 `.api` 类型注解：

- 文章列表、用户分页和点赞收藏列表是 GET 查询参数，结构体必须保留 `form` tag。
- 文章详情路径参数为 `path:"article"`。
- 发布文章与更新封面是 multipart；更新封面处理器需先 `ParseMultipartForm`。
- 文章列表与搜索响应必须映射 `article_id` 与 `created_at`，否则前端无法跳转或显示日期。
- 文章更新接口的 `tags` 是字符串数组；后端 repository 使用 map 更新时必须显式写入 JSON 文本，不能直接传 `[]string` 给 MySQL JSON 列。
- 评论顶层路由当前带拼写错误：`/novblog/comment/commet/top`；前端保留正确路径 fallback。
- 站点留言使用后端约定的根地址 `3489673482`；顶层评论发送 `parent_id: 0`，API 校验应为 `gte=0`。`/station/*` 是前端本地静态路由，不经过微服务代理。
- 角色值：`0` 普通用户，`1` 管理员，`2` 超级管理员。后台入口要求 `Role >= 1`；公告发布要求 `Role >= 2`。

## 已知限制

- `agent/chat` 与 `agent/deploy` 目前是后端脚手架，前端如实显示接口返回，不伪造模型结果。
- 归档接口不返回文章 ID；前端仅能用近期文章列表按标题补全部分跳转，旧文章可能只能显示标题。
- 用户批量接口不返回总数与封禁状态；后台同时提供封禁、解封操作。
- 后端没有“当前文章是否已点赞/收藏”的查询接口；文章页按钮状态只在当前页面会话内维护。
- 后端没有友链列表接口；`/friends` 展示本站信息与申请表单，不伪造已收录友链列表。

## 验证命令

```powershell
cd front
npm run build
```

后端契约改动可按需做编译检查：

```powershell
go test ./server/app/article/articleapi/... ./server/app/user/userapi/...
```

该命令只编译与运行包测试，不启动微服务。

