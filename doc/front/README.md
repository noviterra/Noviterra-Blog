# Noviterra 前端说明文档

Noviterra 前端是一套基于 **Nuxt 3（单框架）** 的模块化应用，对接项目内 go-zero 用户微服务，覆盖主页、文章、友链、关于、认证、个人设置与后台管理。本文档说明依赖安装、一键启动、各地址（路由 / 接口）对应内容、模块划分，以及「请求到数据刷新」的对应关系，便于维护与新增模块。

> 权限边界：本工程只包含前端代码与本文档。后端（`server/`）只读，前端不修改、不新增任何后端文件。

---

## 1. 技术栈

| 类别 | 选型 | 说明 |
| --- | --- | --- |
| 框架 | Nuxt 3（Vue 3，SSR） | 单框架，文件路由、组合式 API |
| 语言 | TypeScript | 全量类型，接口类型见 `utils/apiTypes.ts` |
| 样式 | 原生 CSS + 设计令牌 | cascade layers，不使用 Tailwind / UI 组件库 |
| 字体 | Geist Variable / Geist Mono Variable（自托管） | 中文回退系统字体，`font-display: swap` |
| 图标 | Phosphor（`@phosphor-icons/vue`） | 统一经 `BaseIcon` 调用，不手写 SVG |
| 状态 | 组合式函数 + 本地存储 | Token 持久化，无额外状态库 |

设计基调：编辑感博客 + 轻后台。低饱和翡翠绿单色强调、冷调中性色、1px 边框、克制阴影、明暗双主题；动效仅滚动揭示与路由过渡，不使用视差、跑马灯与滚动监听。

---

## 2. 目录结构（模块划分）

前端代码全部位于项目根下 `front/`：

```text
front/
├─ app.vue                 # 根组件：布局 + 页面 + Toast
├─ app.config.ts           # 站点配置：名称、导航、博主、社交、页脚
├─ nuxt.config.ts          # Nuxt 配置：主题首帧脚本、全局样式、apiBase、过渡
├─ assets/css/             # 设计系统样式
│  ├─ tokens.css           # 设计令牌（明/暗主题：颜色、字号、间距、圆角、阴影）
│  ├─ base.css             # 基础重置与排版
│  ├─ layout.css           # 页头、页脚、容器、栅格
│  ├─ components.css       # 按钮、表单、徽章、卡片、弹窗、表格等
│  ├─ markdown.css         # .prose 正文排版
│  ├─ chrome.css           # 后台框架样式
│  ├─ pages.css            # 各页面专属样式
│  └─ main.css             # 入口，依次引入以上文件
├─ utils/
│  ├─ apiTypes.ts          # 后端请求/响应/用户类型
│  ├─ constants.ts         # ENDPOINTS、枚举文案、分页约束、存储键
│  └─ content/             # 本地内容适配层（未来接口替换接缝）
│     ├─ posts.ts          # 文章本地数据
│     └─ friends.ts        # 友链本地数据
├─ composables/            # 组合式函数（自动导入）
│  ├─ useApi.ts            # HTTP 收口：信封归一化、Token 注入、过期刷新
│  ├─ useAuth.ts           # 鉴权收口：登录/注册/改密/资料、会话
│  ├─ useTokenStore.ts     # access/refresh Token 存储
│  ├─ useToast.ts          # 全局轻提示
│  ├─ useTheme.ts          # 明暗主题
│  ├─ useCountdown.ts      # 验证码冷却
│  ├─ usePagination.ts     # 分页状态
│  ├─ useFormat.ts         # 日期、性别、角色、阅读时长格式化
│  ├─ usePosts.ts          # 文章数据访问（内容接口唯一接缝）
│  └─ useFriends.ts        # 友链数据访问
├─ plugins/
│  ├─ auth.client.ts       # 客户端启动：主题同步 + 会话恢复
│  └─ directives.ts        # v-reveal 滚动揭示指令（IntersectionObserver）
├─ middleware/
│  ├─ auth.ts              # 需登录
│  └─ admin.ts             # 需管理员
├─ layouts/
│  ├─ default.vue          # 站点默认布局（页头 + 页脚）
│  ├─ bare.vue             # 认证页极简布局
│  └─ admin.vue            # 后台布局（侧栏 + 移动抽屉）
├─ components/
│  ├─ common/               # 通用元件
│  ├─ layout/               # 页头、页脚、主题切换
│  ├─ home/                 # 主页模块（列表行、置顶卡）
│  ├─ article/              # 文章模块（正文渲染、目录）
│  ├─ friends/              # 友链模块（友链卡片）
│  └─ admin/                # 后台模块（用户表格）
├─ pages/                  # 文件路由（见第 4 节）
├─ public/favicon.svg
└─ error.vue               # 全局错误页（404 等）
```

模块约定：每个业务模块自带 `components/<模块>/` 与对应 `pages/`，跨模块通用件放 `components/common/`。新增模块时新增一个组件目录与一个页面目录即可，不影响既有模块。

---

## 3. 安装与启动

### 3.1 环境要求

- Node.js 18 及以上（开发环境使用 Node 22）
- npm（也可用 pnpm，命令等价）

### 3.2 安装依赖

```bash
cd front
npm install
```

`postinstall` 会自动执行 `nuxt prepare` 生成类型。

### 3.3 一键启动（开发）

```bash
npm run dev
```

启动后前端默认运行在 **http://localhost:3000**。

后端用户服务默认运行在 **http://localhost:8888**，接口前缀 `/novblog/user`，并已开启 CORS（`*`），前端可直接跨域调用。请先启动后端服务再进行登录、注册等操作。

### 3.4 生产构建与预览

```bash
npm run build      # 构建
npm run preview    # 本地预览构建产物
npm run generate   # 生成静态站点（内容接口稳定后可选）
```

---

## 4. 前端地址（路由）对应内容

| 路由 | 页面 | 布局 / 权限 | 说明 |
| --- | --- | --- | --- |
| `/` | 主页 | default | 置顶精选、最新文章、主题入口 |
| `/posts` | 文章列表 | default | 支持 `?topic=主题` 与搜索筛选 |
| `/posts/:slug` | 文章详情 | default | 正文、目录、上一篇/下一篇；不存在返回 404 |
| `/friends` | 友链 | default | 按标签筛选的友链卡片 |
| `/about` | 关于 | default | 博主简介、技术栈、时间线、联系方式 |
| `/auth/login` | 登录 | bare | 支持账号密码 / 邮箱验证码；`?redirect=目标`、`?prefill=ID` |
| `/auth/register` | 注册 | bare | 邮箱验证码注册 |
| `/auth/forgot` | 找回密码 | bare | 邮箱验证后重置密码 |
| `/settings` | 个人设置 | default，需登录 | 更新资料、修改密码 |
| `/admin` | 控制台 | admin，需管理员 | 概览统计、最近用户、快捷入口 |
| `/admin/users` | 用户管理 | admin，需管理员 | 分页用户列表、拉黑 / 解封 |
| `/admin/posts` | 文章管理 | admin，需管理员 | 内容接口占位，结构已预留 |

---

## 5. 后端接口对接

### 5.1 基础地址

- 默认 `apiBase`：`http://localhost:8888/novblog/user`
- 配置位置：`nuxt.config.ts` 的 `runtimeConfig.public.apiBase`
- 可用环境变量覆盖：

```bash
# Windows PowerShell
$env:NUXT_PUBLIC_API_BASE="http://你的后端地址/novblog/user"; npm run dev
# Linux / macOS
NUXT_PUBLIC_API_BASE="http://你的后端地址/novblog/user" npm run dev
```

### 5.2 响应约定

- 业务响应 HTTP 状态恒为 `200`，成败看业务信封 `code`：`0` 成功、`1` 失败。
- 信封字段历史上存在 `response` / `Resp` 两种命名，`useApi` 已统一兼容。
- 鉴权中间件失败（无 Token、格式错、解析失败、账号封禁）返回非 2xx，`useApi` 据此触发 Token 刷新或清理会话。
- 鉴权请求统一携带请求头：`Authorization: Bearer <accessToken>`。

### 5.3 接口清单

base = `apiBase`。

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| POST | `/register` | 否 | 注册：`id`(6-10)、`password`(6-20)、`email`、`verification_code`(6)、`verif_password` |
| POST | `/login/id` | 否 | 账号密码登录，返回 access / refresh Token |
| POST | `/login/email` | 否 | 邮箱 + 验证码登录 |
| POST | `/server/verificationcode` | 否 | 发送验证码，`type`：`register` / `login` / `change_password` / `forget_password` |
| POST | `/forgetpassword` | 否 | 找回密码：`id`、`email`、`verification_code`、`new_password`、`verif_password` |
| POST | `/server/refreshaccesstoken` | 否 | 用 refreshToken 换取新的 access / refresh Token |
| GET | `/info` | 是 | 当前用户信息 |
| POST | `/updateinfo` | 是 | 更新资料：`username`(2-20)、`avatar`、`age`(0-150)、`gender`、`self_introduction`(≤200) |
| POST | `/changepassword/id` | 是 | 旧密码改密：`old_password`、`new_password` |
| POST | `/changepassword/email` | 是 | 邮箱改密：`email`、`verification_code`、`new_password` |
| GET | `/batchinfo?size=&page=` | 是（管理员） | 分页用户，`size` 5-15 默认 10，`page` 默认 1，不返回总数 |
| POST | `/blackusers` | 是（管理员） | 按 `id` 拉黑用户 |
| POST | `/unblackusers` | 是（管理员） | 按 `id` 解封用户 |

用户字段：`id`、`Role`（0 普通 / 1 超管）、`username`、`email`、`avatar`、`age`、`gender`（0 未知 / 1 男 / 2 女）、`self_introduction`。

> 说明：`/batchinfo` 不返回用户封禁状态（Status）与总数。因此列表同时提供拉黑与解封操作；是否还有下一页通过「本页返回条数是否等于 size」推断。

---

## 6. 请求到数据刷新的对应关系

所有 HTTP 请求收口到 `useApi`，所有鉴权动作收口到 `useAuth`，保证刷新逻辑单一、可维护：

| 触发动作 | 请求 | 成功后的数据刷新 |
| --- | --- | --- |
| 登录（ID / 邮箱） | `/login/id` 或 `/login/email` | 保存 Token，拉取 `/info`，再跳转 `redirect` 或首页 |
| 注册 | `/register` | 跳转登录页并预填 ID |
| 找回密码 | `/forgetpassword` | 跳转登录页 |
| 更新资料 | `/updateinfo` | 重新拉取 `/info`（单一刷新点，不手动拼接本地状态） |
| 修改密码 | `/changepassword/id` 或 `/changepassword/email` | 提示成功并清空密码表单 |
| 拉黑 / 解封 | `/blackusers` 或 `/unblackusers` | 重新拉取当前页用户列表 |
| 翻页 / 改每页条数 | `/batchinfo` | 用新分页参数重新拉取列表 |
| accessToken 过期 | 自动 `/server/refreshaccesstoken` | 单飞刷新（并发 401 共享一次），随后重放原请求 |
| refreshToken 失效 / 账号被封 | 鉴权中间件返回非 2xx | 清理会话，跳转登录页 |

Token 策略：access 有效期 2 小时、refresh 有效期 7 天；刷新接口同时轮转 access 与 refresh。

---

## 7. 内容适配层（本地源到真实接口）

后端当前仅实现用户能力，**没有文章 / 友链等内容接口**。为不阻塞前端，内容类页面由本地内容适配层承载：

- 数据位于 `utils/content/posts.ts`、`utils/content/friends.ts`。
- 页面与组件只面对 `usePosts()` / `useFriends()` 返回的稳定数据结构，与未来 API 保持一致。
- **切换到真实接口时，只需修改 `composables/usePosts.ts` / `useFriends.ts` 内部实现**，页面与组件无需改动。这是内容接口的唯一接缝。
- 文章正文当前为作者自有可信 HTML，经 `ArticleContent.vue` 的 `.prose` 渲染；未来若返回 Markdown，仅在该组件内接入解析。

后台控制台中文章、评论、访问等暂无后端的模块统一标注「待接入」，不展示虚构数字。

---

## 8. 部署与同源代理建议

- 后端已开启 CORS，生产环境前端可直接跨域调用，也可通过反向代理（Nginx）或 Nitro 路由规则将 `/novblog` 代理到后端，实现同源访问。
- 使用同源代理时，将 `NUXT_PUBLIC_API_BASE` 指向同源路径（例如 `/novblog/user`），由反向代理转发到后端服务。
- 主题在首帧前由内联脚本设定（localStorage 键 `noviterra.theme`），不会出现主题闪烁。

---

## 9. 新增模块指南

1. 在 `components/<新模块>/` 下放置模块专属组件。
2. 在 `pages/` 下按文件路由新增页面，需要权限时在 `definePageMeta` 指定 `middleware`。
3. 数据访问优先复用 `useApi`；内容类数据在对应 `useXxx` 组合式函数中收口，保持页面与数据来源解耦。
4. 复用 `components/common/` 的表单、按钮、弹窗、分页、空状态与骨架，保持视觉与交互一致。
5. 涉及写操作时，遵循第 6 节约定：成功后通过重新拉取数据来刷新界面，避免手动维护多份本地状态。
