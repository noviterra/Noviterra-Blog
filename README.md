# Noviterra

Noviterra 是一个全栈个人博客平台，采用微服务架构，包含 Go 后端服务和 Nuxt 3 前端应用。

> 状态：开发中

---

## 项目概览

Noviterra 提供完整的博客功能：用户注册与认证、文章发布与管理、评论互动（点赞 / 收藏 / 嵌套评论）、公告发布、友链申请等，同时配备一个编辑感设计风格的前台与轻量后台管理界面。

---

## 架构

```
┌─────────────────────────────────────────────────┐
│                    Client                        │
│           (Nuxt 3 · Vue 3 · TypeScript)          │
└─────────────┬───────────────────────────────────┘
              │ REST API (HTTP)
┌─────────────▼───────────────────────────────────┐
│  API Gateway Layer (go-zero REST)                │
│  ┌──────────┬──────────┬──────────┬──────────┐  │
│  │ userapi  │articleapi│commentapi│identityapi│  │
│  │  :8003   │  :8004   │  :8005   │  :8001   │  │
│  └────┬─────┴────┬─────┴────┬─────┴────┬─────┘  │
└───────┼──────────┼──────────┼──────────┼────────┘
        │ gRPC     │ gRPC     │ gRPC     │ gRPC
┌───────▼──────────▼──────────▼──────────▼────────┐
│  RPC Service Layer (go-zero gRPC)                │
│  ┌──────────┬──────────┬──────────┬──────────┐  │
│  │ identity │  userrpc │articlerpc│commentrpc│  │
│  │  :9001   │  :9003   │  :9004   │  :9005   │  │
│  └──────────┴──────────┴──────────┴──────────┘  │
│  ┌──────────┬──────────┐                        │
│  │ noticerpc│ siterpc  │                        │
│  │  :9002   │  :9006   │                        │
│  └──────────┴──────────┘                        │
└───────┬──────────────┬──────────────────────────┘
        │              │
   ┌────▼────┐   ┌─────▼─────┐
   │  MySQL  │   │   Redis   │
   └─────────┘   └───────────┘
        │
   ┌────▼────┐
   │  Etcd   │  (服务注册与发现)
   └─────────┘
```

### 服务说明

| 服务 | REST 端口 | gRPC 端口 | 职责 |
|------|-----------|-----------|------|
| **Identity** | 8001 | 9001 | 注册、登录（账号密码 / 邮箱验证码）、Token 生成与刷新、忘记密码 |
| **Notice** | — | 9002 | 公告发布与展示、邮箱验证码发送、友链申请进度邮件通知 |
| **User** | 8003 | 9003 | 个人信息查询与更新、密码修改、管理员用户管理（拉黑 / 解封 / 批量查询） |
| **Article** | 8004 | 9004 | 文章 CRUD、列表查询、搜索、归档、用户文章列表、点赞 / 收藏文章列表 |
| **Comment** | 8005 | 9005 | 文章点赞 / 收藏、发表评论、获取顶层评论与嵌套子评论树 |
| **Site** | — | 9006 | 友链申请 |

---

## 技术栈

### 后端

| 类别 | 选型 | 说明 |
|------|------|------|
| 语言 | Go 1.25 | — |
| 框架 | go-zero v1.10 | REST + gRPC 双协议微服务框架 |
| 数据库 | MySQL | 通过 GORM v1.31 进行 ORM 操作与自动迁移 |
| 缓存 | Redis v9 | 点赞 / 收藏计数、文章浏览量、Lua 脚本原子操作 |
| 服务发现 | Etcd | gRPC 服务注册与发现 |
| 认证 | JWT (golang-jwt/v4) | Access Token (2h) + Refresh Token (7d) 双 Token 策略 |
| ID 生成 | Snowflake (自实现) | 分布式唯一 ID |
| 校验 | go-playground/validator | 请求参数校验 |
| 邮件 | gopkg.in/gomail.v2 | SMTP 邮件发送 |
| 日志 | uber-go/zap + lumberjack | 结构化日志 + 日志轮转 |
| 密码 | golang.org/x/crypto | bcrypt 加密 |

### 前端

| 类别 | 选型 | 说明 |
|------|------|------|
| 框架 | Nuxt 3 (Vue 3 · SSR) | 文件路由、组合式 API |
| 语言 | TypeScript | 全量类型 |
| 样式 | 原生 CSS (Cascade Layers) | 设计令牌、明暗双主题，无 UI 组件库 |
| 图标 | Phosphor Icons | 统一经 BaseIcon 封装 |
| 字体 | Geist Variable (自托管) | 中文回退系统字体 |

---

## 项目结构

```
Noviterra/
├── server/                          # 后端 Go 代码
│   ├── app/                         # 应用服务
│   │   ├── model/                   # 共享 GORM 数据模型
│   │   │   ├── usersModel.go        #   - 用户账户、用户信息
│   │   │   ├── articlesModel.go     #   - 文章信息
│   │   │   ├── commentModel.go      #   - 评论、统计、用户互动
│   │   │   ├── identityModel.go     #   - Refresh Token
│   │   │   ├── noticeModel.go       #   - 公告
│   │   │   └── siteModel.go         #   - 友链
│   │   ├── article/                 # 文章服务
│   │   │   ├── articleapi/          #   REST API 层
│   │   │   └── articlerpc/          #   gRPC RPC 层
│   │   ├── comment/                 # 评论互动服务
│   │   │   ├── commentapi/
│   │   │   └── commentrpc/
│   │   ├── identity/                # 认证服务
│   │   │   ├── identityapi/
│   │   │   └── identityrpc/
│   │   ├── notice/                  # 通知服务
│   │   ├── site/                    # 站点服务
│   │   ├── user/                    # 用户服务
│   │   │   ├── userapi/
│   │   │   └── userrpc/
│   │   └── internal/                # 内部共享组件
│   │       ├── interceptor/         #   gRPC 服务令牌拦截器
│   │       └── middleware/          #   HTTP 中间件（JWT 鉴权、浏览记录、可选鉴权）
│   └── utils/                       # 工具包
│       ├── bcryptinfo/              # 密码加密
│       ├── bloomfilter/             # 布隆过滤器
│       ├── cache/                   # Redis 缓存操作 + Lua 脚本
│       ├── common/                  # 通用工具（类型转换、字符串处理、去重）
│       ├── ctxdata/                 # Context 数据存取
│       ├── email/                   # 邮件模板与发送
│       ├── errcode/                 # 统一错误码
│       ├── jwtauth/                 # JWT 生成与解析
│       ├── repo/                    # 数据库仓储层 (GORM)
│       ├── response/                # 响应体构造（VO 转换）
│       ├── snowflake/               # Snowflake ID 生成器
│       ├── validate/                # 请求参数校验
│       ├── woker/                   # 后台 Worker（点赞/收藏同步）
│       └── zaplogger/               # 日志初始化
├── front/                           # 前端 Nuxt 3 应用（独立仓库）
├── doc/                             # 文档与 API 定义
│   ├── api/                         # go-zero .api 定义文件
│   ├── proto/                       # Protobuf 定义文件
│   └── front/                       # 前端说明文档
├── static/articlecover/             # 文章封面图
├── tests/                           # 测试 HTML 页面
├── bin/                             # 编译产物
├── go.mod / go.sum                  # Go 模块依赖
├── start.bat                        # Windows 本地一键构建 & 启动脚本
└── .env                             # 环境变量
```

---

## 快速开始

### 环境要求

- Go 1.25+
- MySQL 8.0+
- Redis 7.0+
- Etcd 3.5+
- Node.js 18+（前端）

### 1. 启动基础设施

确保 MySQL、Redis、Etcd 已启动并可通过默认端口访问：

- MySQL: `127.0.0.1:3306`（数据库 `noviterra`，自动建表迁移）
- Redis: `127.0.0.1:6379`
- Etcd: `127.0.0.1:2379`

### 2. 启动后端服务

**方式一：一键启动 (Windows)**

```bat
start.bat
```

该脚本会依次关闭旧进程、编译所有服务并启动：

- `userapi.exe` / `userrpc.exe`
- `articleapi.exe` / `articlerpc.exe`
- `identityapi.exe` / `identityrpc.exe`

**方式二：手动编译与运行**

```bash
# 编译
go build -o bin/userapi.exe    ./server/app/user/userapi
go build -o bin/userrpc.exe    ./server/app/user/userrpc
go build -o bin/articleapi.exe ./server/app/article/articleapi
go build -o bin/articlerpc.exe ./server/app/article/articlerpc
go build -o bin/identityapi.exe ./server/app/identity/identityapi
go build -o bin/identityrpc.exe ./server/app/identity/identityrpc

# 启动
./bin/userapi.exe
./bin/userrpc.exe
# ... 依次启动其他服务
```

### 3. 启动前端

```bash
cd front
npm install
npm run dev
```

前端开发服务器默认运行在 `http://localhost:3000`。

---

## 核心设计

### 认证流程

- 用户注册 / 登录后获得 AccessToken（2 小时有效）和 RefreshToken（7 天有效）
- 每次请求在 `Authorization: Bearer <accessToken>` 中携带
- AccessToken 过期后前端自动使用 RefreshToken 换取新 Token 对
- 支持可选鉴权中间件：未登录用户降级为游客，登录用户通过 Token 识别身份

### API → RPC 通信

- API 层（REST）通过 gRPC 调用对应的 RPC 层服务
- gRPC 请求通过 service-token 元数据进行服务间认证
- 用户身份信息通过 user-id 元数据传递

### 点赞 / 收藏计数

- 高频互动操作（点赞 / 收藏 / 浏览）先写入 Redis，通过 Lua 脚本保证原子性
- 后台 LikeSyncWorker 定期（每 5 秒）将 Redis 增量数据同步到 MySQL

### 数据库

- 使用 GORM AutoMigrate 自动建表与迁移
- 按服务拆分 GORM 仓储：ArticleRepository、CommentRepository、IdentityRepository 等
- 软删除：文章 / 用户支持逻辑删除

---

## API 概览

### 认证 (Identity)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /register | 用户注册 |
| POST | /login/id | 账号密码登录 |
| POST | /login/email | 邮箱验证码登录 |
| POST | /forgetpassword | 忘记密码 |
| POST | /server/verificationcode | 发送邮箱验证码 |
| POST | /server/refreshaccesstoken | 刷新 Token |

### 用户 (User)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /info | 获取个人信息 |
| POST | /updateinfo | 更新个人资料 |
| POST | /changepassword/id | 旧密码修改密码 |
| POST | /changepassword/email | 邮箱验证修改密码 |
| GET | /batchinfo | 管理员批量查询用户 |
| POST | /blackusers | 管理员拉黑用户 |
| POST | /unblackusers | 管理员解封用户 |

### 文章 (Article)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /publisharticle | 发布文章 |
| DELETE | /article/:id | 删除文章 |
| PUT | /article/:id | 更新文章 |
| PUT | /articlecover/:id | 更新文章封面 |
| GET | /articles | 文章列表（分页） |
| GET | /article/:id | 文章详情 |
| GET | /articles/search | 搜索文章 |
| GET | /archive | 文章归档 |
| GET | /user/:id/articles | 用户发布的文章列表 |
| GET | /user/:id/likecollect | 用户点赞 / 收藏的文章列表 |

### 评论互动 (Comment)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /likearticle | 点赞 / 取消点赞 |
| POST | /collectarticle | 收藏 / 取消收藏 |
| POST | /article/comment | 发表评论 |
| GET | /article/:id/comment | 获取文章顶层评论 |
| GET | /comment/:id/tree | 获取评论子评论树 |

### 站点与通知 (Site & Notice)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /applyfriendlink | 申请友链 |
| POST | /publishnotice | 发布公告（管理员） |
| GET | /notice | 获取公告列表 |

---

## 设计风格

前端采用编辑感博客设计，以低饱和翡翠绿为强调色，冷调中性色，1px 边框、克制阴影、明暗双主题。详见 [doc/front/README.md](doc/front/README.md)。

---

## License

MIT
