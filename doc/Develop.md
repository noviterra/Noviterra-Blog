# Noviterra 博客平台 · 架构与实现设计说明书

> 文档路径：`doc/Develop.md`
> 文档角色：本项目唯一的**架构基线（Architecture Baseline）**，后续所有目录创建、代码编写、Code Review 均以本文为准。
> 约定：本文只做**设计**，不创建任何代码文件。所有文件与目录均由项目所有者按本文 **[附录 A 文件创建清单](#附录-a文件创建清单必须由项目所有者逐项创建)** 审查后手动创建。

| 项目 | 内容 |
| --- | --- |
| 项目名 | Noviterra |
| 类型 | 大型多用户博客 / 内容平台（服务端微服务架构） |
| 主语言 | Go（服务端）、TypeScript（前端） |
| 核心框架 | go-zero（rest / zrpc / goctl） |
| 存储 | MySQL 8 + GORM v2、Redis 7、对象存储（MinIO/OSS/COS） |
| 注册与配置 | Etcd 3.5（服务发现 + 配置中心） |
| 当前仓库状态 | 空骨架：`doc/`、`front/`、`server/` 三个空目录 + 根 `go.mod`（`module Noviterra`, `go 1.25.5`） |

**版本记录**

| 版本 | 日期 | 修改人 | 说明 |
| --- | --- | --- | --- |
| v1.0 | 初稿 | 架构师 | 首版完整设计：服务拆分、目录分层、共享包、第三方接入、数据与接口规范、实施路线图 |
| **v1.1** | 修订 | 架构师 | ★ **架构修订**：采用 go-zero 官方《目录拆分》的**域自治**结构——每个域固定 `{api,rpc,model}` 三目录；取消集中式 `gateway-api`（改为边缘路由 + 各域 api 自行鉴权）；新增 `portal` 承接跨域聚合；补充"路径第一段唯一归属""api 不碰 model"等硬约束、部署形态（先合并后拆分）、以及全套接口归属与路径清单 |

---

## 目录

- [0. 阅读指南与全局约定](#0-阅读指南与全局约定)
- [1. 项目定位、边界与技术选型](#1-项目定位边界与技术选型)
- [2. 总体架构](#2-总体架构)
- [3. 代码仓库目录结构总览](#3-代码仓库目录结构总览)
- [4. 共享工具包 `server/pkg` 详规](#4-共享工具包-serverpkg-详规)
- [5. 第三方 API 接入规范（重点）](#5-第三方-api-接入规范重点)
- [6. 各服务详细设计](#6-各服务详细设计)
- [7. 数据设计](#7-数据设计)
- [8. 接口规范](#8-接口规范)
- [9. 鉴权与安全](#9-鉴权与安全)
- [10. 可观测性与运维](#10-可观测性与运维)
- [11. 部署与交付](#11-部署与交付)
- [12. 测试策略与质量门禁](#12-测试策略与质量门禁)
- [13. 工程规范](#13-工程规范)
- [14. 实施路线图与里程碑](#14-实施路线图与里程碑)
- [15. 附录](#15-附录)

---

## 0. 阅读指南与全局约定

### 0.1 本文怎么用

1. **先读第 1~3 章**：确认技术选型与目录骨架，创建仓库骨架文件。
2. **再读第 4~5 章**：先把 `server/pkg` 打通（日志、DB、Redis、JWT、错误码、响应、第三方适配层），这是所有服务的公共地基。
3. **然后按第 14 章路线图**逐阶段实现服务，每个服务的细节看第 6 章对应小节。
4. **开发过程中**：接口看第 8 章、数据看第 7 章、安全看第 9 章、规范看第 13 章。
5. **附录 A** 是可直接勾选的建仓清单。

### 0.2 术语与命名约定

| 术语 | 含义 | 本项目约定 |
| --- | --- | --- |
| api 服务 | go-zero `rest` 服务，**域内 HTTP 门面**（面向终端的业务实现） | 后缀 `-api`，如 `article-api` |
| rpc 服务 | go-zero `zrpc` 服务，**给其它子系统的基础数据入口** | 后缀 `-rpc`，如 `article-rpc` |
| 定时/异步服务 | 无对外接口的常驻进程 | 后缀 `-cron` / `-worker` |
| PO | Persistent Object，数据库表映射结构体（GORM Model） | `app/<域>/model/entity` |
| DO | Domain Object，领域对象 | `internal/domain`（重服务才建） |
| DTO | 服务间/内部传输对象 | `internal/dto` |
| VO | View Object，对前端输出对象 | `internal/vo` |
| IDL | 接口描述文件（`.api` / `.proto`） | `server/api`、`server/proto` |

**命名总则**

- Go 包名：全小写、无下划线、尽量单个单词（`logx`、`idgen`、`thirdparty`）。
- 目录名：小写 + 连字符仅用于服务名（`user-rpc` 的目录名实际为 `user`，服务名在配置中声明）。
- 数据库：库名 `noviterra_<service>`，表名小写蛇形，字段小写蛇形。
- Redis Key：`nv:{env}:{service}:{entity}:{id}`，详见 [7.5 缓存设计](#75-缓存设计redis-kv)。
- 错误码：5 位数字分段，详见 [8.3](#83-错误码规范)。
- 时间：数据库统一存 UTC（`datetime`），时区转换只在展示层做。

### 0.3 依赖方向铁律（Review 时必须检查）

```
【域内】以 article 域为例
  app/article/api  ──►  app/article/rpc      （允许且唯一：api 经 gRPC 取本域数据）
  app/article/api  ──✗─► app/article/model   （严禁：api 不得 import 本域数据层，不得持有 *gorm.DB）
  app/article/rpc  ──►  app/article/model    （允许：model 的唯一使用者）

【跨域】
  app/A/api 或 app/A/rpc  ──►  app/B/rpc     （只允许调对方 rpc；严禁调对方 api）
  app/A  ──►  app/B                          （仅限 rpc client / MQ 事件；严禁直连 B 的库表或 Redis Key）
  rpc 调用必须单向，禁止 A ⇄ B 循环依赖

【通用】
  app/*  ──►  pkg/*                          （允许）
  pkg/*  ──►  app/*                          （严禁：工具包不得反向依赖业务）
  app/*  ──►  pkg/thirdparty/*               （允许：严禁在 logic 里裸写 HTTP 请求第三方）
  编排型服务（portal/admin）  ──►  各域 rpc    （允许；但它们不拥有 model，也不得直连任何库）
```

违反以上任一条的 PR 一律打回。CI 建议加一条静态检查：`api/` 目录下 grep 不到 `model` 与 `gorm` 的 import。

### 0.4 分层职责与"十不准"

**先分清"协议入口"与"层"**：`api` / `rpc` 是同一域的两个**协议入口**（不是两层）；真正的层是 `handler → logic → svc → repository → DB/缓存`。两个入口共用同一份 `app/<域>/model`。

单个域内部各层职责：

| 层 | 位置 | 职责 | 不准做什么 |
| --- | --- | --- | --- |
| `handler` | `api/internal/handler`、`rpc/internal/server` | 协议转换：请求 → logic → 响应 | 不准写业务判断、不准访问 DB/Redis |
| `logic`（api） | `api/internal/logic` | **编排**：聚合多域数据、按终端裁剪字段、频控与幂等 | 不准写领域规则（状态机/可见性/归属）、不准 import `model`、不准写 SQL |
| `logic`（rpc） | `rpc/internal/logic` | **领域用例**：规则判定、事务边界、缓存策略 | 不准写 SQL（走 repository）、不准裸调 HTTP |
| `domain` | `rpc/internal/domain` | 领域规则的唯一来源（状态机、算法、权限判定） | 不准依赖 gin/rest/grpc 等协议类型 |
| `svc` | `*/internal/svc` | 依赖容器（api 侧只放 rpc client；rpc 侧放 DB/Redis/Repo） | 不准写业务逻辑 |
| `repository` | `model/repository` | 数据访问（GORM）、缓存读写 | 不准写业务规则、不准调其他域 |
| `middleware` | `*/internal/middleware`、`pkg/middleware` | 鉴权 / trace / 日志 / 恢复 / 限流 | 不准写领域逻辑 |

**十不准（速记）**

1. 不准在 `logic` 层出现 `gorm.DB`、`SELECT`、`redis.Client` 的直接使用（`api` 侧更是**连 `model` 都不准 import**）。
2. 不准跨域查表、跨域复用对方的 Redis Key；跨域取数只能调对方 `rpc`。
3. 不准把业务逻辑写进 `pkg`。
4. 不准在 `pkg` 中读取环境变量之外的全局配置（配置必须由调用方注入）。
5. 不准在循环里发 RPC / 查库 / 调第三方（必须批量）。
6. 不准返回裸 `error` 给前端，必须经 `pkg/errx` 归一化。
7. 不准 `panic` 作为业务控制流。
8. 不准提交带密钥的配置文件（密钥走环境变量 / K8s Secret / Vault）。
9. 不准在生产环境用 `AutoMigrate` 改表。
10. 不准在 handler 层 `if err != nil { return }` 静默吞错，必须记录并转换。

---

## 1. 项目定位、边界与技术选型

### 1.1 业务范围（In Scope）

Noviterra 是一个**多用户内容平台**（写作 + 阅读 + 互动 + 后台运营），核心域：

| 域 | 能力诉求 |
| --- | --- |
| 账号域 | 注册/登录（账号密码、邮箱验证码、第三方 OAuth）、个人主页、关注粉丝、角色权限 |
| 内容域 | 文章（草稿/审核/发布/定时发布/版本历史）、分类、标签、专栏系列、SEO |
| 互动域 | 评论（楼中楼）、点赞、收藏、举报、敏感词审核、消息通知 |
| 媒体域 | 图片/附件/头像上传、分片与秒传、图片处理（压缩/水印/WebP）、对象存储 |
| 检索域 | 全文检索、高亮、筛选、搜索建议、热搜统计 |
| 运营域 | 后台管理（用户/内容/评论/分类/配置/公告/审计日志）、数据看板 |
| 增强域 | 第三方能力：翻译（百度等）、AI 摘要/润色、IP 归属地、短信/邮件、内容安全审核 |
| 平台域 | 网关、限流熔断、链路追踪、指标监控、配置中心、定时任务 |

### 1.2 非功能目标（NFR）

| 维度 | 目标（MVP → 成熟期） |
| --- | --- |
| 可用性 | 99.5% → 99.95%；无状态服务可水平扩容；单点故障不影响读 |
| 性能 | 读接口 P99 < 200ms（Redis 命中 < 50ms）；写接口 P99 < 500ms |
| 吞吐 | 支撑 1w DAU → 50w DAU；文章详情 QPS 500 → 5000（靠多级缓存） |
| 一致性 | 核心写路径强一致（MySQL 事务）；统计计数、搜索索引、缓存最终一致（秒级） |
| 安全 | 全站 HTTPS、JWT 双令牌、RBAC + 数据级越权校验、上传白名单、限流防刷 |
| 可维护 | 单服务 < 1.5w 行；新服务从 goctl 生成到联调 < 1 天 |
| 可观测 | 100% 请求带 `trace_id`；日志可检索；核心指标 + 告警；慢 SQL 全量记录 |

### 1.3 技术选型矩阵

> 版本以落地时最新稳定版为准，**必须在 `go.mod` 中锁定**，禁止使用 `latest`。

| 层次 | 选型 | 备选 | 选型理由 |
| --- | --- | --- | --- |
| 语言 | Go 1.25.x | — | 已有 `go.mod` 声明 `go 1.25.5` |
| 微服务框架 | go-zero v1.7+（`rest` / `zrpc` / `goctl`） | Kratos、Gin+自研 | 用户指定；内置限流、熔断、自适应降载、缓存、链路追踪，代码生成完善 |
| 代码生成 | `goctl`（api / rpc / model ） | — | 与 go-zero 一体；`.api` 即接口文档 |
| ORM | GORM v2（`gorm.io/gorm` + `driver/mysql`） | sqlx（go-zero 默认）、ent | 用户指定；生态成熟，钩子/软删除/读写分离插件齐全 |
| 关系库 | MySQL 8.0（utf8mb4_0900_ai_ci，InnoDB） | PostgreSQL | 生态与运维最熟，JSON 字段够用 |
| 缓存 | Redis 7（`github.com/redis/go-redis/v9`） | — | 用户指定；缓存 + 分布式锁 + 限流 + 计数器 + Stream |
| 服务发现/配置 | Etcd 3.5 | Consul、Nacos | 用户指定；go-zero `zrpc` 原生支持 etcd 注册发现 |
| 对象存储 | MinIO（开发）/ 阿里云 OSS / 腾讯 COS | 本地磁盘 | 通过 `pkg/thirdparty/storage` 接口屏蔽差异 |
| 搜索 | Meilisearch（起步）→ Elasticsearch（规模化） | 纯 MySQL LIKE | 通过 `search-rpc` + `Indexer` 接口屏蔽；起步选轻量的 Meilisearch |
| 异步/队列 | Redis Stream 或 `asynq`（起步）→ Kafka（规模化） | RabbitMQ | 通过 `pkg/mq` 接口屏蔽，避免早期引入 Kafka 运维成本 |
| 配置管理 | 本地 `etc/*.yaml`（dev）+ Etcd 配置中心（test/prod） | Apollo、Nacos | 与 Etcd 复用同一套基础设施 |
| 日志 | `go.uber.org/zap`（经 `logx` 包统一封装） | logrus | 用户指定；结构化、高性能、支持采样 |
| 鉴权 | JWT（access）+ 不透明 refresh token（Redis） | Session | 无状态 + 可撤销兼顾 |
| 校验 | go-zero `httpx.Parse` + `go-playground/validator/v10` | — | 声明式校验，错误信息可归一化 |
| 链路追踪 | OpenTelemetry + Jaeger/Tempo | zipkin | go-zero 内建 `trace` 支持 OTLP |
| 指标 | Prometheus + Grafana | — | go-zero 内建 metrics 输出 |
| 迁移工具 | `goose`（版本化 SQL） | golang-migrate、atlas | 简单可靠，生产禁用 `AutoMigrate` |
| 测试 | `testing` + `testify` + `go-sqlmock` + `testcontainers-go` | — | 单元 + 集成双层 |
| 前端 | 前台 Next.js（SSR/SEO）/ 后台 Vue3 + Element Plus | — | 见 [附录 F 待决策问题](#附录-f待决策问题open-questions) |
| 容器 | Docker + Docker Compose（dev）/ K8s + Kustomize（prod） | 裸机 systemd | 起步 Compose，成熟上 K8s |
| CI/CD | GitHub Actions（或 GitLab CI） | Jenkins | 与代码托管绑定 |

### 1.4 关键设计原则（Architecture Principles）

1. **契约先行**：先写 `.api` / `.proto`，再 `goctl` 生成骨架，最后填 `logic`。
2. **单库不跨服务**：一个服务独占自己的 schema；跨服务数据通过 RPC 或事件同步。
3. **一切外部依赖可替换**：DB、Redis、存储、搜索、第三方 API 都通过接口注入，实现靠配置切换。
4. **缓存是加速器不是数据源**：任何缓存都必须能从 DB/上游重建；写路径先 DB 后失效缓存（Cache-Aside + 延迟双删）。
5. **失败要显式**：第三方失败、降级、限流都必须返回明确错误码并打点，禁止静默返回空数据。
6. **先能跑，再跑好**：MVP 允许评论/搜索/统计合并到相邻服务，但目录与接口边界预留，避免后期返工。
7. **可观测默认开启**：新服务脚手架必须含 trace、metrics、结构化日志、健康检查。

---

## 2. 总体架构

### 2.1 分层架构总览

```
┌───────────────────────────────────────────────────────────────────────────────┐
│  客户端层  前台门户(Next.js SSR)   后台管理(Vue3)   移动端/小程序(预留)  开放API │
└───────────────────────────────┬───────────────────────────────────────────────┘
                                │ HTTPS / JSON（前端只认一个域名，按路径前缀分发）
┌───────────────────────────────▼───────────────────────────────────────────────┐
│  边缘层   Nginx / Ingress                                                      │
│  职责：TLS 终止、路径路由、静态资源、CORS、基础 IP 限流、WAF、灰度权重          │
│  ★ 这一层不是 Go 服务，不写任何业务代码                                        │
└───────────────────────────────┬───────────────────────────────────────────────┘
                                │ HTTP：/api/v1/{domain}/**  →  {domain}-api
                                │       /api/v1/admin/**     →  admin
┌───────────────────────────────▼───────────────────────────────────────────────┐
│  业务服务层（域自治：每个域一个目录，两种协议入口 api / rpc）                    │
│                                                                               │
│   app/user/{api,rpc,model}       app/article/{api,rpc,model}                  │
│   app/comment/{api,rpc,model}    app/file/{api,rpc,model}                     │
│   app/message/{api,rpc,model}    app/search/{api,rpc,model}                   │
│   app/stat/{api,rpc,model}       app/third/{api,rpc,model}                    │
│   app/system/{api,rpc,model}                                                  │
│   ── 只有 api 的编排型服务：app/portal（首页等跨域聚合）、app/admin（后台聚合） │
└───────────────────────────────┬───────────────────────────────────────────────┘
                                │ 域内：api ──gRPC──► 本域 rpc
                                │ 跨域：api / rpc ──gRPC──► 他域 rpc（etcd 服务发现）
┌───────────────────────────────▼───────────────────────────────────────────────┐
│  异步与调度层   app/job（定时任务）   mq worker（事件消费，内嵌于各服务）       │
└───────────────────────────────┬───────────────────────────────────────────────┘
                                │
┌───────────────────────────────▼───────────────────────────────────────────────┐
│  基础设施层  MySQL 8 | Redis 7 | Etcd 3.5 | MinIO/OSS | Meilisearch/ES | MQ    │
│  支撑设施    Prometheus + Grafana | Jaeger | Alertmanager | 配置中心(etcd)      │
└───────────────────────────────────────────────────────────────────────────────┘
```

**为什么是"域自治（api + rpc）+ 边缘路由"，而不是"集中式 BFF 网关"**（这是本项目最重要的架构决策，依据 go-zero 官方《目录拆分》的最佳实践）：

1. **api 是"对终端的业务入口"，rpc 是"给其它子系统的基础数据入口"**——同一个域的两种协议门面，共享同一份 `model`。它们面对不同消费者，因此不是重复实现。
2. **发布解耦**：评论域的接口改动只需重新部署 `comment-api`，不会像集中网关那样"改一个接口、全站重新发布、一个 bug 影响所有域"。
3. **跳数不变**：`Client → article-api → article-rpc` 与 `Client → gateway-api → article-rpc` 同样是 **2 段网络 + 1 次协议转换**，域自治并不更慢；省下的反而是一个横跨所有域的共用进程。
4. **边缘职责用边缘设施承担**：TLS、路由、静态资源、IP 限流交给 Nginx/Ingress（成熟、零维护成本），不要为了"架构完整"再写一个 Go 网关进程。
5. **JWT 校验下沉到各 api 服务**：由 `pkg/middleware/auth.go` 统一实现（同一份代码、同一份配置模板），避免 N 份实现漂移。

> ⚠️ **绝对禁止**：在保留域内 `api` 的同时再串一个业务网关（`Client → gateway → article-api → article-rpc`）。那会造成两次 HTTP 解析、三段网络、两套鉴权，是本项目最容易出现的架构返工点。

### 2.2 服务拆分原则与边界

拆分遵循 6 条原则：

1. **按业务能力拆，不按技术层拆**：`user-rpc` 而不是 `account-rpc` + `profile-rpc`。
2. **数据所有权唯一**：一张表只属于一个服务，只有属主能写。
3. **高内聚**：会同时修改、同时查询的数据放一起（文章 + 标签 + 分类 + 版本）。
4. **有限耦合**：跨服务只保留 ID 引用，展示需要的冗余字段由事件补偿。
5. **变化频率隔离**：第三方（变化最频繁）单独成服务/包；搜索（索引重建重）单独成服务。
6. **规模与团队匹配**：单人/小团队阶段允许"逻辑拆分、物理合并"（见 2.7 与 14 章阶段策略）。

**服务边界的"不做什么"清单**（防止服务膨胀）：

| 服务（域） | 明确不做 |
| --- | --- |
| `*.api`（所有域的 api 入口） | 不直连 DB/Redis、不 import 本域 `model`（只能经 `rpc` 拿数据）、不做领域规则判定 |
| `portal` / `admin`（编排型 api） | 不拥有任何表、不实现业务规则，只做跨域聚合、字段裁剪与 RBAC 编排 |
| `user` | 不发文、不存文件、不发通知（只发事件） |
| `article` | 不做评论树、不做搜索索引、不做文件存储、不管理用户资料 |
| `comment` | 不维护文章计数（发事件让 article 更新） |
| `file` | 不做文图关联的业务规则（只存 file_id 引用） |
| `message` | 不决定"什么时候发"，只负责"怎么发"（模板 + 通道 + 重试） |
| `search` | 不做业务校验，只做索引与检索 |
| `third` | 不做业务规则，只做能力编排、配额、审计、降级 |
| `system` | 不做各域数据的管理端点（后台用户管理调 `user.rpc`）；只管 RBAC、配置、字典、审计 |
| `job` | 不直连任何业务库，只通过各域 rpc / MQ 完成任务 |

### 2.3 服务清单总表

**表 1：业务域**——每个域一个目录，内含 `api/`（HTTP 门面）、`rpc/`（gRPC 门面）、`model/`（共享数据层）。

| # | 域 | 目录 | api 端口 | rpc 端口 | api 面向终端（业务需求实现） | rpc 面向其它子系统（基础数据） | 独占 schema | 优先级 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | user | `app/user/` | 8801 | 9001 | 注册登录、令牌刷新、资料、关注粉丝、后台用户管理 | 用户信息、批量用户摘要、角色、关注关系判定 | `noviterra_user` | **P0** |
| 2 | system | `app/system/` | 8802 | 9002 | 后台 RBAC/字典/配置/公告/审计/任务管理界面接口 | 权限码展开、用户菜单、站点公开配置、操作日志写入 | `noviterra_system` | P1 |
| 3 | article | `app/article/` | 8803 | 9003 | 文章读写、发布/审核/定时、分类标签专栏、点赞收藏、翻译触发 | 文章详情、批量摘要、互动状态判定、计数增减 | `noviterra_content` | **P0** |
| 4 | comment | `app/comment/` | 8804 | 9004 | 评论列表/发表/删除、点赞、举报 | 评论数批量查询、文本审核（供他域复用）、文章删除联动 | `noviterra_comment` | P1 |
| 5 | file | `app/file/` | 8805 | 9005 | 上传（单文件/分片/秒传）、凭证签发、附件管理 | 文件元数据查询、引用登记/解除、私有文件签名 URL | `noviterra_file` | P1 |
| 6 | message | `app/message/` | 8806 | 9006 | 站内信列表、未读数、已读、WebSocket 推送 | 统一发送入口、模板渲染、群发任务 | `noviterra_message` | P2 |
| 7 | search | `app/search/` | 8807 | 9007 | 检索、筛选、高亮、搜索建议、热搜词 | 索引增删改、全量重建、索引统计 | `noviterra_search`（轻） | P2 |
| 8 | stat | `app/stat/` | 8808 | 9008 | 后台看板、趋势图、Top 榜、作者数据 | 埋点上报、计数增减、日聚合触发 | `noviterra_stat` | P2 |
| 9 | third | `app/third/` | 8809 | 9009 | 后台第三方配额与调用日志界面 | 翻译、AI 摘要、内容审核、IP 归属、短信邮件、OAuth 换票 | `noviterra_third` | P2 |

**表 2：编排型与调度型服务**——它们不拥有数据（`portal`/`admin` 无 `model`；`job` 无 `api`/`model`）。

| # | 服务 | 目录 | 端口 | 类型 | 职责 | 优先级 |
| --- | --- | --- | --- | --- | --- | --- |
| 10 | portal | `app/portal/` | 8810 | rest（仅 api） | 首页/聚合页等跨域视图：文章 + 热榜 + 标签 + 公告 + 站点配置一次返回 | P1 |
| 11 | admin | `app/admin/` | 8811 | rest（仅 api） | 后台统一入口：跨域管理聚合、RBAC 接口级鉴权、操作日志采集、导出任务触发 | P1 |
| 12 | job | `app/job/` | 9100 | cron | 定时任务：定时发布、索引重建、统计汇总、清理、失败重投 | P1 |

**端口约定**

| 类型 | 规则 | 示例 |
| --- | --- | --- |
| 域 api（HTTP） | `88xx`，末两位与所属域一致 | `article-api` → 8803 |
| 域 rpc（gRPC） | `90xx`，末两位与所属域一致 | `article-rpc` → 9003 |
| 编排型 api | 8810 起顺延 | portal 8810、admin 8811 |
| 调度服务 | 9100 | job |
| 指标/健康端口 | 业务端口 + 10000，仅内网 | `article-api` → 18803、`article-rpc` → 19003 |

每个进程都暴露 `/metrics`、`/healthz`、`/readyz`、`/version`（metrics 端口）。开发环境中间件端口见 [11.1](#111-本地开发环境)。

**P0/P1/P2 说明**：P0 = 第一个可上线版本必须；P1 = 第二里程碑；P2 = 第三里程碑（`api` 与 `rpc` 可先合并同进程跑，目录结构不变，见 [11.4](#114-部署形态先合并后拆分)）。

### 2.4 关键链路设计

#### 2.4.1 登录链路

```
前端 → Nginx/Ingress → user-api    POST /api/v1/auth/login
  ├─ 【api 层】参数校验 + 图形验证码校验（Redis: nv:{env}:user:captcha:{cid}）
  ├─ 【api 层】登录失败计数器（Redis INCR + 过期，5 次锁 15 分钟）
  ├─ 【api 层】调用本域 rpc：user-rpc.Login
  │    └─ 【rpc 层】
  │         ├─ 查 user_auth（email/phone）→ bcrypt 校验
  │         ├─ 检查用户状态（封禁/未激活）
  │         ├─ 生成 access JWT（2h）+ refresh token（30d，Redis 存储）
  │         ├─ 写 user_login_log、更新 last_login_at/ip
  │         └─ 发事件 user.login（供 stat/message 消费）
  └─ 【api 层】装配响应 { access_token, refresh_token, expires_in, user_info }
```

**分工要点**：`user-api` 只做"HTTP 契约 + 校验 + 频控 + 响应装配"，认证规则（密码校验、令牌签发）全在 `user-rpc`。这样后台批量建号、其它服务内部登录校验等场景可以直接复用 rpc，不必绕过 HTTP。

刷新：`POST /api/v1/auth/refresh` → 校验 refresh（Redis 存在且未撤销）→ **轮转**（旧的立即失效）→ 若检测到已轮转的旧 token 被复用，则撤销该用户全部设备的 refresh（判定为令牌泄漏）。

#### 2.4.2 发布文章链路（含审核 / 定时 / 异步副作用）

```
作者 → Nginx/Ingress → article-api   POST /api/v1/articles   (需登录)
  ├─ 【api 层】JWT 校验 → 注入 x-uid / x-roles / x-trace-id
  ├─ 【api 层】参数校验（标题长度、正文大小、标签数量）→ article-rpc.CreateArticle
  │    └─ 【rpc 层】
  │         ├─ 领域权限校验（作者本人 / 专栏归属）
  │         ├─ 写 article + article_content（大字段拆表）
  │         ├─ 写 article_version v1（版本历史）
  │         ├─ 状态机：draft →（提交）reviewing →（审核）published | rejected
  │         │    publishing_now：published_at = now
  │         │    scheduled   ：status=scheduled + scheduled_at，由 app/job 到点发布
  │         └─ 同一事务写 outbox（事件 article.published）
  └─ 【api 层】返回文章摘要（不回正文）
        │
        └─ outbox 投递（异步） article.published.v1
             ├─ search-rpc 消费者  → 建/更新索引
             ├─ stat-rpc 消费者    → 初始化文章统计行
             ├─ third-rpc 消费者   →（可选）生成 AI 摘要、机器翻译
             └─ message-rpc 消费者 → 通知关注者
```

**关键点**：索引、统计、翻译、通知全部**异步**，主链路只写 MySQL，`P99 < 300ms`。事件发送使用 **本地消息表 + 重试**（在同一事务里写 `outbox` 表，后台协程投递），保证"DB 提交成功但消息丢失"可控。

#### 2.4.3 文章详情读链路（多级缓存）

```
GET /api/v1/articles/:id
  Nginx/Ingress → article-api
    ├─ 【api 层】可选解析 JWT（登录用户才有"是否已赞/已藏"）
    ├─ 【api 层】并行发起（用 mapreduce / errgroup 并发，禁止串行）：
    │    ├─ article-rpc.GetArticle       → 文章基础数据（含作者冗余字段、分类、标签）
    │    ├─ comment-rpc.BatchGetCount    → 评论数
    │    └─ article-rpc.CheckInteraction → 当前用户是否已赞/已藏
    ├─ 【rpc 层】GetArticle 的多级缓存：
    │    L1  进程内缓存（ristretto / freecache，TTL 5s，只缓存热点 TopN）
    │    L2  Redis：nv:{env}:article:detail:{id}（TTL 10min ± 随机 60s 防雪崩）
    │    L3  MySQL：article + article_content + 分类标签 + 作者冗余字段
    │    防穿透：空结果缓存 "NULL" 占位 60s
    │    防击穿：singleflight 合并同 Key 并发回源
    ├─ 【api 层】按"页面视图"裁剪字段（游客不返回草稿字段、私有文章校验可见性）
    └─ 【api 层】异步计浏览（Redis Hash 累加，app/job 每 5 分钟批量落库），不阻塞响应
```

**为什么聚合放在 api 层**：这正是 go-zero 对 `api` 的定义——"**业务需求实现**"。rpc 提供的是与终端无关的基础数据（文章就是文章），而"详情页要哪些字段、游客与登录用户看到什么"是终端需求，归 `api`。这也保证了 rpc 可被其它子系统安全复用，不会因为某个前端改版而受影响。

#### 2.4.4 翻译链路（第三方接入示例）

```
作者/后台 → Nginx → article-api → article-rpc.TranslateArticle(lang=ja)
  ├─ 【api 层】鉴权 + 参数校验
  └─ 【rpc 层】article-rpc
       ├─ 查 article_translation（该文该语种是否已有）→ 命中直接返回
       ├─ 跨域调用 third-rpc.Translate
       │    ├─ Redis 缓存：nv:{env}:third:translate:{provider}:{from}:{to}:{sha1(text)}
       │    ├─ 配额检查（Redis 计数器 + third_api_quota 表）
       │    ├─ 调 pkg/thirdparty/translate/baidu（超时 3s、重试 2 次、熔断、限流）
       │    └─ 失败降级：返回原文 + 标记 status=degraded，写入重试队列
       ├─ 成功：★ 由 article-rpc 落库 article_translation（数据所有权在 content 域）
       └─ 发事件 article.translated
```

**跨域调用方向**：`article-rpc → third-rpc` 是单向的（文章需要翻译能力）；`third-rpc` **绝不反向调用** `article-rpc`，也绝不写 `noviterra_content` 库。这满足 [0.3 依赖方向铁律](#03-依赖方向铁律review-时必须检查)。

### 2.5 通信与协议约定

| 场景 | 协议 | 约定 |
| --- | --- | --- |
| 客户端 → 边缘层 | HTTPS + JSON | Nginx/Ingress 终结 TLS；按前缀路由到对应 api 服务 |
| 客户端 → 域 api | HTTP + JSON | RESTful，路径 `/api/v1/**`；统一响应体；`Authorization: Bearer <access_token>` |
| 域 api → 本域 rpc | gRPC（zrpc） | 通过 etcd 解析（合并部署时可用直连地址绕过 etcd）；超时 3s（读）/ 5s（写）；开启熔断 |
| 域 api → 他域 rpc | gRPC（zrpc） | 跨域取数一律走 rpc，**不得直连他域 DB/Redis** |
| rpc → rpc（同步） | gRPC | 必须单向（禁止循环依赖）；只能调用对方 `client` 包暴露的方法 |
| 服务 → 服务（异步） | MQ / Redis Stream | 事件命名 `{domain}.{action}.v{n}`，如 `article.published.v1` |
| 服务 → 第三方 | HTTPS | 必须经 `pkg/thirdparty` 统一客户端 |
| 服务 → 中间件 | 原生协议 | 连接池参数统一在 `pkg/db`、`pkg/cache` 中固化 |

**上下文透传**（gRPC metadata）：

| Key | 含义 | 注入方 |
| --- | --- | --- |
| `x-trace-id` | 链路 ID（全链路一致） | 域 api（或边缘层生成后透传） |
| `x-request-id` | 单次请求 ID | 域 api |
| `x-uid` | 用户 ID（0=匿名） | 域 api 解析 JWT 后注入 |
| `x-username` | 用户名（日志用） | 域 api |
| `x-roles` | 角色码列表（逗号分隔） | 域 api |
| `x-perm-ver` | 权限版本号（用于踢下线/权限变更） | 域 api |
| `x-client-ip` | 真实客户端 IP（经可信代理白名单计算，取 `X-Forwarded-For` 首个非可信地址） | 域 api |
| `x-internal-token` | 内部调用凭证（非 api 来源的调用） | 服务间调用 |

**安全要求**：`rpc` 服务**不信任**任意来源传入的 `x-uid`（否则任何能访问内网的进程都能伪造身份），只接受以下两类调用：

1. 来自本域或他域 **api 服务**（已验签并注入 metadata）的调用；
2. 持有 `x-internal-token` 的**内部系统调用**（如 `app/job` 触发的任务）。

生产环境所有 `rpc` 端口不对外暴露（K8s NetworkPolicy / 安全组），仅集群内可达。

### 2.6 服务发现与配置中心（Etcd）

**服务注册发现**

- 每个 `zrpc` 服务启动时向 Etcd 注册，Key 规范：`/noviterra/{env}/{service}/{host}:{port}`，TTL 租约 15s。
- 调用方配置示例（`etc/*.yaml`）：

```yaml
UserRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: user.rpc
  NonBlock: true
  Timeout: 3000
```

- 生产要求：`NonBlock: true`（避免启动期强依赖）、`Timeout` 显式设置、开启 `Breaker` 或依赖 go-zero 默认自适应熔断。

**配置中心（分环境策略）**

| 环境 | 配置来源 | 说明 |
| --- | --- | --- |
| dev | `server/app/<svc>/etc/<svc>.yaml` | 提交到仓库，**不含任何真实密钥**，密钥用占位符 |
| test | Etcd 配置中心（`/noviterra/test/<svc>`） | 可热更新：限流阈值、开关、第三方配额 |
| prod | Etcd 配置中心 + 环境变量注入密钥 | 密钥来自 K8s Secret / Vault，**不落配置文件、不进镜像** |

**必须支持热更新的配置项**（无需重启）：功能开关（评论审核开关、第三方降级开关）、限流阈值、缓存 TTL、日志级别、敏感词库版本。**不支持热更新**：DB 连接串、端口、服务名（需滚动发布）。

### 2.7 编排型服务（`portal` / `admin`）的边界

**核心问题**：域自治之后，两类"天然跨域"的需求放在哪里？——① 首页/聚合页（文章 + 热榜 + 标签 + 公告）；② 后台管理（用户 + 内容 + 评论 + 配置 + 看板）。以及：后台要不要再建一套 `admin-rpc`？

**决策**

1. **跨域聚合新建 `app/portal`（只有 `api`，无 `rpc`/`model`）**。绝不把首页聚合塞进 `article-api`——那会让文章域反向依赖标签、公告、统计，破坏域边界。
2. **后台保留单个 `app/admin`（只有 `api`）**，不建 `admin-rpc`，也不按域拆后台 api。后台是运营工具：跨域聚合多、RBAC 集中、接口变更频率低，拆开收益小、配置成本大。
3. **`portal` 与 `admin` 与各域 api 是平级关系，不串联**：仍是 `Client → admin → 各域 rpc`（2 段网络），不存在三跳。

| 方案 | 优点 | 缺点 | 结论 |
| --- | --- | --- | --- |
| A. 建 `admin-rpc` 独立领域服务 | 后台逻辑与前台隔离彻底 | 用户管理/文章管理逻辑与 `user.rpc`/`article.rpc` **大量重复**，双写风险高 | ❌ 不采用 |
| B. `app/admin`（api）+ 复用各域 rpc | 无重复业务逻辑，数据所有权清晰 | admin 需做多服务聚合（可接受） | ✅ **采用** |
| C. 后台完全独立一套服务与库 | 隔离性最强 | 运维与开发成本翻倍，数据同步复杂 | ❌ 不采用 |
| D. 把后台接口塞进各域 api 的 `/admin` 分组 | 域自治最彻底 | RBAC 中间件要在 N 个 api 重复配置，权限映射分散易漂移 | ⚠️ 仅在后台极简时可选 |

**`admin` 的职责边界**

1. **后台专用接口的编排与聚合**（如后台看板 = `stat.rpc` + `user.rpc` + `article.rpc` 组合）。
2. **RBAC 接口级鉴权**：路由 → 权限码映射，在中间件统一校验（映射表来自 `system.rpc` 的权限表，本地缓存 60s）。
3. **操作日志采集**：所有非 GET 请求记录到 `sys_operation_log`（调 `system.rpc` 或落 MQ 异步写）。
4. **后台专用的宽松分页/导出**：导出走异步任务（`app/job`），不在请求内同步生成大文件。
5. **不重复实现业务规则**：例如"封禁用户"= 调 `user.rpc.AdminUpdateStatus`，规则在 `user.rpc` 内。

**后台功能 → 服务归属映射表**

| 后台模块 | 归属 api | 数据来源 |
| --- | --- | --- |
| 用户列表/详情/封禁/角色分配 | `admin` | `user.rpc`（角色数据来自 `system.rpc`） |
| 文章列表/审核/推荐/置顶/删除 | `admin` | `article.rpc` |
| 分类/标签/专栏管理 | `admin` | `article.rpc` |
| 评论审核/删除/举报处理 | `admin` | `comment.rpc` |
| 敏感词库维护 | `admin` | `comment.rpc` |
| 文件/附件管理 | `admin` | `file.rpc` |
| 公告/站内信群发 | `admin` | `message.rpc`（群发 = 异步任务） |
| 角色/权限/菜单/字典/系统配置 | `admin` | `system.rpc` |
| 操作日志/登录日志查询 | `admin` | `system.rpc` / `user.rpc` |
| 数据看板/趋势图 | `admin` | `stat.rpc` |
| 定时任务管理与手动触发 | `admin` | `app/job`（内部 HTTP 管理端口 + `x-internal-token`） |
| 第三方配额与调用日志 | `admin` | `third.rpc` |

**`portal` 的职责边界**

1. 只做**跨域聚合**：首页、聚合页、关注流（`user.rpc` 取关注列表 + `article.rpc` 批量取文章）。
2. 不拥有任何表、不写库；结果可缓存（`nv:{env}:portal:home`，TTL 60s）。
3. 若某个聚合页只涉及单一域，**应下沉到该域 api**，不要都堆到 portal。

---

## 3. 代码仓库目录结构总览

### 3.1 仓库根目录（当前已有骨架 + 建议补齐）

```
Noviterra/                                 # 仓库根（git init 于此处）
├── .gitignore                             # 忽略 bin/ dist/ *.log .env etc/*.local.yaml
├── .gitattributes                         # 统一 LF，*.go text eol=lf
├── .editorconfig                          # 缩进/换行/字符集（Go: tab；TS: 2 空格）
├── .golangci.yml                          # Go 静态检查规则（放在根，供 server/ 使用）
├── .dockerignore
├── LICENSE                                # 开源协议（如 MIT / Apache-2.0）
├── README.md                              # 项目简介、快速开始（本地 5 分钟跑起来）
├── CONTRIBUTING.md                        # 开发规范入口（链接 doc/Develop.md 第 13 章）
├── CHANGELOG.md                           # 版本变更（按 Keep a Changelog）
├── Makefile                               # 根级聚合命令（make server / make web / make infra）
├── go.mod                                 # 当前 module Noviterra（建议改为域名前缀，见 3.8）
├── go.sum
├── .github/
│   └── workflows/
│       ├── ci-server.yml                   # 后端：lint → test → build（paths: server/**）
│       ├── ci-web.yml                      # 前端：lint → build（paths: front/**）
│       └── release.yml                     # 打 tag 时构建镜像并推送
├── doc/                                   # 文档（本文所在）
│   ├── Develop.md                         # ★ 本文件：架构与实现设计
│   ├── api/                               # 接口文档产物（goctl 导出的 markdown / OpenAPI）
│   ├── db/                                # ER 图、表设计说明
│   ├── deploy.md                          # 部署手册（环境、步骤、回滚）
│   ├── ops.md                             # 运维手册（告警处理、常见故障）
│   └── adr/                               # 架构决策记录 0001-xxx.md（每次重大决策留档）
├── front/                                 # 前端（与 server 平级，独立构建）
│   ├── web/                               # 前台门户（Next.js，SEO 优先）
│   ├── admin/                             # 后台管理（Vue3 + Element Plus）
│   ├── shared/                            # 共享：API 类型定义、请求封装、常量（可发布为内部包）
│   ├── package.json                       # pnpm workspace 根
│   └── pnpm-workspace.yaml
└── server/                                # ★ 后端全部代码（本设计主体）
    ├── api/common/                        # 仅存放公共 HTTP 契约（base.api / auth.api）；域契约跟随服务（见 3.5）
    ├── proto/common/                      # 仅存放公共 gRPC 契约（common.proto / event.proto）
    ├── app/<域>/{api,rpc,model}/          # ★ 每个业务域固定三目录：HTTP 门面 / gRPC 门面 / 共享数据层（见 3.2、3.3）
    ├── pkg/                               # 共享工具包（见第 4 章）
    ├── deploy/                            # 部署与基础设施（见 3.7）
    ├── scripts/                           # 开发/运维脚本（见 3.7）
    ├── test/                              # 集成测试与压测（见 3.7）
    └── Makefile                           # 后端专用命令（生成、构建、运行、迁移）
```

> **为什么不把 `pkg` 放在仓库根？** 因为 `pkg` 只服务后端 Go 代码；放 `server/pkg` 可以让"后端"成为一个自洽的可迁移单元（未来若拆仓，整个 `server/` 可直接独立出去）。

### 3.2 `server/app` 服务目录：每个服务固定预留 `api` / `rpc` / `model`

**约定（本项目强制，Code Review 必查）**：`server/app/<domain>/` 下**固定预留三个子目录** `api/`、`rpc/`、`model/`。

```
server/app/
├── user/                        # ── 账号域
│   ├── api/                     # user-api     HTTP 门面（8801）：面向终端的业务实现
│   ├── rpc/                     # user-rpc     gRPC 门面（9001）：给其它子系统的基础数据
│   └── model/                   # ★ api 与 rpc 共享的唯一数据层（PO + repository + cache）
├── system/   {api, rpc, model}  # RBAC / 配置 / 字典 / 审计
├── article/  {api, rpc, model}  # 文章 / 分类标签 / 专栏 / 互动
├── comment/  {api, rpc, model}  # 评论 / 举报 / 敏感词
├── file/     {api, rpc, model}  # 上传 / 分片 / 秒传 / 图片处理
├── message/  {api, rpc, model}  # 站内信 / 邮件 / 短信 / WS 推送
├── search/   {api, rpc, model}  # 索引 / 检索 / 建议 / 热词
├── stat/     {api, rpc, model}  # 埋点 / 统计 / 看板
├── third/    {api, rpc, model}  # 第三方能力编排
├── portal/   {api}              # 编排型：只有 api，无 rpc/model（不拥有数据）
├── admin/    {api}              # 编排型：只有 api
└── job/      {job.go, internal/} # 调度型：无 api/rpc/model（不直连业务库）
```

`article` 域展开后的样子（其余业务域完全同构）：

```
server/app/article/
├── api/        # article-api 进程（HTTP 8803）
├── rpc/        # article-rpc 进程（gRPC 9003）
└── model/      # 被 api 与 rpc 共同可见的唯一数据层
```

**为什么 `model` 与 `api`、`rpc` 平级**（依据 go-zero 官方《目录拆分》给出的完整工程目录示例 `order/{api,model,rpc}`）：

1. **数据访问只有一份** → 物理上不可能出现"两套 repository、两套 SQL、两套缓存 Key"。
2. **域是自洽单元**：契约、代码、数据层、配置、部署清单都在同一目录，将来要拆成独立仓库时可整体搬走。
3. **模型迁移/表结构脚本有唯一归属**，不会散落在多个进程目录里。

**纪律（重要）**：虽然 `api` 在物理上可以 import `model`，但本项目**禁止 `api` import 本域 `model` 或 `*gorm.DB`**。`api` 只能通过本域 `rpc` 或他域 `rpc` 取数据，否则会绕过领域规则直接改库，让状态机、审核、计数等规则形同虚设。

> 备选：若你更倾向"api 绝对碰不到数据层"的物理隔离，可把 `model` 放进 `rpc/`（go-zero 官方 quickstart 的 bookstore 示例即把 model 放在 `rpc/` 下）。本文档采用**平级方案**，理由是它与官方《目录拆分》一致、便于将来扩展，且靠上述纪律来约束，而不是靠目录位置。

**服务全名与目录名**：目录名用短业务名（`user`）；服务全名（`user-api` / `user-rpc`）写在 `etc/*.yaml` 的 `Name` 字段与 etcd Key 中，并作为进程名、镜像名、K8s Deployment 名的后缀，便于运维识别。

### 3.3 单个服务的标准分层（以 `article` 域为例，逐文件说明）

> 这是**每个业务域必须遵守的模板**。业务域 = `api/` + `rpc/` + `model/` 三块；编排型服务（`portal`/`admin`）只取 `api/` 一块。

**一、`api/` —— HTTP 门面（`article-api`，8803）**

```
server/app/article/api/
├── article.go                          # 【生成】入口：conf.MustLoad → svc.NewServiceContext → rest.Start
├── article.api                         # 【手写】HTTP 契约源文件（goctl 生成代码的依据，见 3.5）
├── etc/
│   ├── article-api.yaml                # 【生成】dev 配置：Name/Host/Port/Etcd/Jwt/Redis/各域 RpcClientConf
│   └── article-api.local.yaml          # 【手写|gitignore】个人本地覆盖，不入库
└── internal/
    ├── config/config.go                # 【生成后手改】Config：Rest + Jwt + Redis + ArticleRpc/UserRpc/CommentRpc…
    ├── svc/service_context.go          # 【生成后手改】★ 只放"各域 rpc client"，绝不放 *gorm.DB / repository
    ├── handler/                        # 【生成】协议转换：HTTP 出入参 ↔ logic（不写业务）
    │   ├── createarticlehandler.go
    │   └── …
    ├── logic/                          # 【生成后填充】编排：聚合、裁剪、频控、幂等键
    │   ├── createarticlelogic.go        # 校验 + 调 article-rpc.CreateArticle + 装配响应
    │   ├── getarticledetaillogic.go     # ★ 并发聚合：article-rpc + comment-rpc（本域/跨域）
    │   ├── listarticleslogic.go
    │   └── …（一个 HTTP 接口一个文件）
    ├── types/                          # 【生成】HTTP 出入参结构体
    ├── middleware/                     # 【按需】本域特有中间件（通用件一律放 pkg/middleware）
    └── consts/                         # 本域 HTTP 层常量与错误码引用
```

**二、`rpc/` —— gRPC 门面与领域实现（`article-rpc`，9003）**

```
server/app/article/rpc/
├── article.go                          # 【生成】入口：conf.MustLoad → svc.NewServiceContext → zrpc.Start
├── article.proto                       # 【手写】gRPC 契约源文件（见 3.5）
├── etc/
│   ├── article-rpc.yaml                # 【生成】Name/Host/Port/Etcd/DB/Redis/各域 RpcClientConf
│   └── article-rpc.local.yaml          # 【手写|gitignore】
├── pb/                                 # 【生成】protobuf 产物 —— 禁止手改
├── article/                            # 【生成】对外 client 包（供他域 import）—— 禁止手改
│   └── article.go                      # NewArticle(zrpc.ClientConf) → Article
└── internal/
    ├── config/config.go                # 【生成后手改】Config：ZRpc + DB + Redis + MQ + 他域 RpcClientConf
    ├── svc/service_context.go          # 【生成后手改】★ 这里才注入 *gorm.DB、Redis、各 Repo、他域 rpc client
    ├── server/article_server.go         # 【生成】把 Logic 注册到 zrpc server，此处不改
    ├── logic/                          # 【生成后填充】领域用例：一个 RPC 方法一个文件
    │   ├── createarticlelogic.go        # 领域规则 + 事务 + outbox
    │   ├── publisharticlelogic.go       # 状态机流转
    │   └── …
    ├── middleware/                     # RPC 拦截器：auth（还原 uid）、trace、log、recovery
    ├── dto/                            # 服务内部/跨服务数据契约（不直接暴露 PO）
    ├── vo/                             # 对调用方输出的视图对象
    ├── assembler/                      # 转换器：PO ↔ DTO ↔ VO（禁止在 logic 里手写长赋值）
    ├── domain/                         # ★ 领域规则唯一来源
    │   ├── article_status.go            # 状态机：CanPublish/CanReview/CanDelete
    │   ├── article_visibility.go        # 可见性判定
    │   ├── heat.go                      # 热度算法
    │   └── permission.go                # 谁能改这篇文章
    ├── event/                          # 事件：topics.go / publisher.go / consumer_*.go
    └── consts/                         # consts.go / enums.go / errors.go
```

**三、`model/` —— 唯一数据层（被 `api` 与 `rpc` 共享的包）**

```
server/app/article/model/
├── entity/                             # 【手写】PO（GORM Model）
│   ├── article.go                      #   Article / ArticleContent / ArticleVersion
│   ├── taxonomy.go                     #   Category / Tag / ArticleTag / Series
│   ├── interaction.go                  #   ArticleLike / ArticleCollect / CollectFolder
│   ├── translation.go                  #   ArticleTranslation
│   ├── outbox.go                       #   本地消息表
│   └── table_name.go                   #   统一 TableName()（可选）
├── repository/                         # 【手写】仓储：接口 + GORM 实现
│   ├── article_repo.go                 #   type ArticleRepo interface { … }
│   ├── article_repo_impl.go            #   type articleRepo struct { db *gorm.DB; cache *cache.ArticleCache }
│   ├── category_repo.go
│   ├── tag_repo.go
│   └── outbox_repo.go
├── cache/                              # 【手写】缓存 Key 构造 + 序列化 + 空值占位 + TTL 抖动
│   ├── keys.go                         #   func ArticleDetailKey(id int64) string
│   └── article_cache.go
└── migrate/                            # 【可选】表结构说明；真正的迁移 SQL 统一放 deploy/sql（见 3.7）
```

**四、域内依赖方向（必须遵守）**

```
        ┌──────────────────────┐        ┌──────────────────────┐
        │  api/handler         │        │  rpc/server          │
        │  api/logic  (编排)    │        │  rpc/logic  (领域)    │
        └──────────┬───────────┘        └──────────┬───────────┘
                   │ gRPC（zrpc client）            │ 直接调用
                   ▼                               ▼
        ┌──────────────────────────────────────────────────────┐
        │  model/repository  ← 唯一数据访问出口                 │
        └──────────────────────────┬───────────────────────────┘
                                   ▼
                            MySQL / Redis
   api ──✗──► model        （api 不得 import model：纪律红线）
   api ──✓──► 他域 rpc      （跨域取数一律走 rpc）
   model ──✗──► 任何 rpc    （数据层不反向依赖领域/协议层）
```

**五、生成物 vs 手写物（重要）**

| 目录/文件 | 来源 | 可否手改 |
| --- | --- | --- |
| `rpc/pb/`、`rpc/<svc>/`（client 包） | goctl | ❌ 重新生成会覆盖 |
| `api/internal/handler/`、`api/internal/types/`、`rpc/internal/server/` | goctl | ⚠️ 结构可手改但重新生成会覆盖 |
| `api/internal/logic/`、`rpc/internal/logic/`、`*/internal/svc/`、`*/internal/config/` | goctl 生成空壳 | ✅ 主体手写 |
| `api/*.api`、`rpc/*.proto` | 手写（唯一契约源） | ✅ |
| `model/**`、`*/middleware/`、`event/`、`domain/`、`dto/`、`vo/`、`assembler/`、`consts/` | 全手写 | ✅ |
| `etc/*.yaml` | goctl 生成骨架 | ✅ 手改 |

**六、对应的 goctl 命令**

```bash
# HTTP 壳（在 server/ 目录下）
goctl api go -api app/article/api/article.api -dir app/article/api
# gRPC 壳 + pb + client
goctl rpc protoc app/article/rpc/article.proto \
  --go_out=app/article/rpc/pb --go-grpc_out=app/article/rpc/pb \
  --zrpc_out=app/article/rpc --module=noviterra/server/app/article/rpc
# 契约变更后重新生成（handler/types 与 pb/client 会更新，logic 已存在则跳过不覆盖）
```

> 建议把上述规则写进 `CONTRIBUTING.md`，并在 `.gitattributes` 中对 `**/pb/*.pb.go` 标记 `linguist-generated=true`。

### 3.4 `server/pkg` 目录（概览）

```
server/pkg/
├── logx/            # zap 封装：初始化、字段规范、GORM/gin 桥接、采样、脱敏
├── jwt/             # 令牌签发/解析/轮转/黑名单
├── crypt/           # bcrypt、AES-GCM、RSA、HMAC、随机串、hash
├── db/              # GORM 初始化、连接池、日志适配、读写分离、分页、事务助手、软删除
├── cache/           # Redis 客户端、Key 前缀、分布式锁、限流器、singleflight、序列化
├── errx/            # 错误码定义 + 业务错误构造 + HTTP/gRPC 错误映射
├── response/        # 统一响应体、分页响应、错误响应
├── middleware/      # 通用中间件/拦截器（鉴权、trace、日志、恢复、CORS、操作日志）
├── idgen/           # 雪花 ID、UUID、短码
├── paginate/        # 分页参数解析与游标分页
├── conv/            # 类型转换（含 nil 安全）、切片去重、map 工具
├── validatorx/      # 自定义校验规则（手机号、密码强度、slug、URL）
├── ctxx/            # context 取值（uid/trace/ip）、超时控制
├── constants/       # 全局常量：枚举值、正则、缓存 TTL、限制值
├── mq/              # 生产者/消费者抽象（redisstream / asynq / kafka 实现）
├── storage/         # 【别名导出 thirdparty/storage】文件存储通用能力（见第 5 章）
├── thirdparty/      # ★ 第三方能力适配层（见第 5 章）
├── observability/   # trace、metrics、pprof、健康检查注册
├── schedule/        # 分布式定时调度（Redis/etcd 锁 + cron 表达式）
└── version/         # 版本号、构建信息（ldflags 注入）
```

> `pkg` 内的每个包必须：**无全局可变状态**（配置由构造函数注入）、**有单元测试**、**有包级注释说明用途与使用示例**。

### 3.5 IDL 目录：`.api` 与 `.proto` 怎么放

**原则：业务契约跟随服务，公共契约集中共享。**

```
server/
├── api/
│   └── common/                         # 公共 HTTP 契约（被各域 .api import）
│       ├── base.api                    # 分页、时间、通用响应 type
│       └── auth.api                    # 鉴权相关公共 type
├── proto/
│   └── common/
│       ├── common.proto                # PageReq / PageResp / Empty / IdsReq / TimeRange / Operator
│       └── event.proto                 # 事件信封 EventEnvelope
└── app/
    ├── article/
    │   ├── api/article.api             # ★ 文章域 HTTP 契约（goctl 据此生成 api/ 骨架）
    │   └── rpc/article.proto           # ★ 文章域 gRPC 契约（goctl 据此生成 rpc/ 骨架 + pb + client）
    ├── user/    {api/user.api,      rpc/user.proto}
    ├── comment/ {api/comment.api,   rpc/comment.proto}
    ├── file/    {api/file.api,      rpc/file.proto}
    ├── message/ {api/message.api,   rpc/message.proto}
    ├── search/  {api/search.api,    rpc/search.proto}
    ├── stat/    {api/stat.api,      rpc/stat.proto}
    ├── third/   {api/third.api,     rpc/third.proto}
    ├── system/  {api/system.api,    rpc/system.proto}
    ├── portal/api/portal.api           # 首页/聚合页契约（只有 api）
    └── admin/api/                      # 后台契约（只有 api）
        ├── admin.api                   # 主入口（import 下列模块）
        ├── user.admin.api
        ├── article.admin.api
        ├── comment.admin.api
        ├── system.admin.api
        ├── stat.admin.api
        └── job.admin.api
```

**规则**

1. **域契约跟随服务目录**：`.api` 放 `<域>/api/`，`.proto` 放 `<域>/rpc/`，**不在 `server/api`、`server/proto` 下另存副本**——一个域一个所有权，新增/改名/删除都在同一目录内完成，避免"契约与服务分离、改一处忘一处"。
2. **公共契约集中共享**：`server/api/common/`、`server/proto/common/`，由各域 import；**公共 type 禁止在业务契约里重复定义**。
3. **一个 `.api` / `.proto` 文件一个域**：文件过大（> 500 行）时按子域拆分，再用主文件 import 聚合（例如 `article.api` import `article_category.api`）。
4. **import 路径**：goctl 以主 `.api` 文件所在目录为基准解析 import，跨目录时写相对路径。
5. **proto 版本兼容**：字段只增不删，删除字段必须 `reserved`；枚举 0 值必须是 `UNSPECIFIED`。
6. **命名一致性**：proto 用 `snake_case`，对外 JSON 用 `camelCase`，由 goctl 自动转换，**不要手写映射**。
7. **接口文档产物**：`goctl api doc app/article/api/article.api -dir doc/api` 导出 markdown，随 PR 更新。
8. **后台路径统一前缀**：`admin.api` 内所有路由使用 `prefix: /api/v1/admin`，由边缘层整段路由到 `app/admin`。

### 3.6 `front` 目录（前端，简要约定）

```
front/
├── pnpm-workspace.yaml
├── package.json                    # 根脚本：dev/build/lint
├── web/                            # 前台门户
│   ├── src/app|pages/              # 路由页面（Next.js App Router）
│   ├── src/components/             # 通用组件
│   ├── src/features/{article,user,comment,search}/   # 按域组织
│   ├── src/lib/api/                # 请求客户端（统一拦截器、错误码映射、token 刷新）
│   ├── src/lib/types/              # ★ 由后端 openapi/ts 生成（或从 front/shared 同步）
│   └── public/
├── admin/                          # 后台管理
│   ├── src/views/{user,article,comment,system,stat}/
│   ├── src/router/                 # 路由 + 权限指令（菜单来自 /admin/me/menus）
│   └── src/api/
└── shared/
    ├── src/types/                  # 与后端契约同步的类型定义
    └── src/constants/error-code.ts # ★ 与 pkg/errx 的错误码保持同源（用脚本生成）
```

**跨端一致性机制**：错误码、枚举值、分页结构这三类东西**必须有单一来源**。做法：后端 `pkg/errx` 与各域 `rpc/internal/consts` 用 `go generate` 导出 JSON，前端脚本生成 `front/shared/src/constants/error-code.ts`。禁止前端手抄错误码。

### 3.7 `deploy`、`scripts`、`test` 目录

```
server/deploy/
├── docker/
│   ├── Dockerfile                  # 通用多阶段构建（ARG SERVICE=<域> + ARG MODE=merged|api|rpc）
│   └── Dockerfile.front            # 前端镜像（Node build + Nginx）
├── compose/
│   ├── docker-compose.infra.yml    # MySQL/Redis/Etcd/MinIO/Meili/Jaeger/Prometheus/Grafana
│   ├── docker-compose.app.yml      # 全部后端服务
│   └── .env.example                # 端口、密码占位（真实 .env 不入库）
├── k8s/
│   ├── base/                       # 通用：Deployment/Service/ConfigMap/Probe 模板
│   ├── overlays/{dev,test,prod}/    # kustomize 覆盖：副本数、资源、Secret、Ingress
│   └── hpa/                        # 自动扩缩容（按 CPU + QPS）
├── sql/
│   ├── migrations/                 # goose 版本化迁移：00001_init_user.sql …
│   └── seed/                       # 初始化数据：角色、权限、系统配置、敏感词基础库
├── etcd/                           # 配置中心初始化脚本与各环境配置模板
└── monitor/
    ├── prometheus/                 # prometheus.yml + 抓取规则
    ├── alertmanager/               # 告警规则（错误率、P99、DB 连接池、第三方失败率）
    ├── grafana/                    # 看板 JSON（服务概览、业务大盘）
    └── jaeger/                     # 追踪采样配置

server/scripts/
├── gen.sh / gen.ps1                # 代码生成入口（api/rpc/model）
├── dev-up.sh / dev-down.sh         # 起停本地基础设施
├── migrate.sh                      # 迁移 up/down/status
├── seed.sh                         # 灌初始化数据
├── new-service.sh                  # 新服务脚手架（生成目录 + 配置 + Docker + Makefile 条目）
└── check.sh                        # 提交前检查：fmt + vet + lint + test

server/test/
├── integration/                    # 跨服务集成测试（testcontainers）
├── e2e/                            # 端到端 API 测试（针对各域 api：user-api / article-api / portal / admin）
├── benchmark/                      # 基准测试（go test -bench / k6 脚本）
└── fixture/                        # 测试数据 SQL / JSON
```

### 3.8 模块组织：单 `go.mod` vs `go.work` 多模块

**现状**：根目录 `go.mod`，`module Noviterra`，`go 1.25.5`。

**推荐（阶段 1~3）：保持单 module**，即 `server/` 下所有服务共享根 `go.mod`。

| 方案 | 优点 | 缺点 |
| --- | --- | --- |
| **A. 单 go.mod（推荐起步）** | goctl 生成零障碍；IDE 索引快；依赖版本天然统一；重构跨服务方便 | 任一服务加依赖会影响全部；无法独立发版 |
| B. go.work + 每服务独立 module | 依赖隔离、可独立发布 | goctl 生成需额外 `--module` 参数；版本漂移风险；IDE 与 CI 配置更复杂 |

**演进路径**：当出现以下任一信号时，再迁移到 `go.work`：
- 服务数量 > 10 且需要独立发版节奏；
- 依赖冲突频繁（如 A 服务要 gorm v2.30、B 要 v2.25）；
- 构建时间 > 3 分钟。

迁移方案（届时执行）：`server/pkg` 独立为 `module <host>/noviterra/pkg`，各服务独立 `go.mod`，根 `go.work` 聚合，`pkg` 用 `replace` 或 `go.work` 本地解析。

> **待办（P0，需项目所有者决定）**：`module Noviterra` 建议改为域名限定的小写路径，例如 `module github.com/<your-org>/noviterra`（Go 官方约定：module path 应为小写、可解析）。若暂不公开，也可用 `module noviterra`。改名影响所有 import 路径，**建议在写第一行业务代码前完成**。

### 3.9 目录与文件粒度规范

| 项目 | 规范 |
| --- | --- |
| 单文件行数 | 建议 ≤ 400 行，> 600 行必须拆分 |
| 单 `logic` 文件 | 只实现 1 个 RPC 方法 |
| 单函数 | ≤ 80 行，圈复杂度 ≤ 15（golangci-lint 卡口） |
| 单包导出符号 | ≤ 15 个，超出说明职责不单一 |
| 目录深度 | ≤ 6 层（`app/article/rpc/internal/logic` 已是 5 层） |
| 测试文件 | 与被测文件同目录同包（`_test.go`） |
| mock | 统一放 `test/mock/`（由 `mockery` 或 `go:generate` 生成） |

---

## 4. 共享工具包 `server/pkg` 详规

### 4.0 总则

**放什么进 `pkg`**：与业务无关、被 ≥ 2 个服务使用、可独立测试的能力。
**不放什么**：任何含业务规则的东西（如"文章能发布吗"、"这个用户能评论吗"）；业务共享逻辑放各自服务的 `domain`，需要跨服务复用的**业务契约**放 `proto/common`。

**每个包的强制要求**

1. 构造函数注入依赖，**禁止包级全局变量**（`logx` 的全局 logger 是唯一例外，见 4.1）。
2. 所有对外函数必须有**包注释 + 函数注释 + 使用示例**。
3. 必须提供**接口 + 默认实现**（便于测试 mock）：如 `db.New(cfg) (*gorm.DB, error)`。
4. 错误必须用 `errx` 包装，不返回裸 `errors.New`。
5. 单元测试覆盖 ≥ 70%（CI 卡口见第 12 章）。

**import 路径**（假设 module 最终为 `github.com/your-org/noviterra`）：

```go
import (
    "github.com/your-org/noviterra/server/pkg/logx"
    "github.com/your-org/noviterra/server/pkg/errx"
)
```

> 若暂不改 module 名（保持 `Noviterra`），则路径为 `Noviterra/server/pkg/logx`。本文后续示例统一使用 **`noviterra/server/pkg/...`** 简写，实际以最终 module 名为准。

### 4.1 `pkg/logx` —— zap 日志封装

**为什么需要封装**：直接用 zap 会导致字段名不统一（`uid`/`user_id`/`userId` 混用）、trace_id 丢失、日志脱敏缺失、GORM/go-zero 自带日志与 zap 双轨。

**文件与职责**

```
server/pkg/logx/
├── logx.go          # Init(cfg) / L() / With(fields...) / Sugar 包装
├── config.go        # Config: Level/Format/Output/FilePath/Rotation/Sampling/ServiceName/Env
├── fields.go        # 统一字段名常量与构造函数：Uid()、TraceId()、Cost()、Err()、Module()
├── context.go       # FromContext(ctx) *zap.Logger —— 自动带 trace_id/uid
├── gorm.go          # GORM logger 适配（慢 SQL、错误 SQL 打到 zap，参数脱敏）
├── zerolog.go       # go-zero logx 与 zap 桥接（可选：统一输出到 zap）
├── mask.go          # 脱敏：手机号、邮箱、身份证、token、密码字段
└── rotate.go        # 日志切割（lumberjack）与保留策略
```

**统一字段名（禁止自定义）**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `service` | string | 服务名，Init 时注入 |
| `env` | string | dev/test/prod |
| `trace_id` | string | 链路 ID |
| `request_id` | string | 单请求 ID |
| `uid` | int64 | 用户 ID（0 = 匿名） |
| `module` | string | 业务模块，如 `article.create` |
| `method` | string | RPC/HTTP 方法名 |
| `cost_ms` | int64 | 耗时毫秒 |
| `errcode` | int | 业务错误码 |
| `error` | string | 错误信息（脱敏后） |
| `client_ip` | string | 客户端 IP |
| `rows` | int64 | DB 影响行数 |
| `sql_ms` | int64 | SQL 耗时 |

**使用示例**

```go
// 初始化（在 main 中，svc 之前）
if err := logx.Init(logx.Config{
    ServiceName: "article-rpc",
    Env:         cfg.Env,
    Level:       cfg.Log.Level,        // debug|info|warn|error
    Format:      cfg.Log.Format,       // json(prod) | console(dev)
    Output:      cfg.Log.Output,       // stdout | file
    FilePath:    cfg.Log.FilePath,
    Sampling:    cfg.Log.Sampling,     // 1s 内同类日志超 100 条后按 1/10 采样
}); err != nil {
    panic(err)
}
defer logx.Sync()

// 业务中（推荐带 ctx，自动注入 trace_id/uid）
log := logx.FromContext(ctx).With(logx.Module("article.create"))
log.Info("article created", logx.ArticleId(id), logx.Cost(cost))
log.Error("create article failed", logx.Err(err), logx.Uid(uid))
```

**硬性要求**

- **禁止** `fmt.Println` / `println` / 标准库 `log`（CI 用 `forbidigo` 检查）。
- 生产环境日志格式必须是 JSON，便于 Loki/ES 检索。
- 密码、token、身份证、手机号中段必须脱敏（`logx.Mask`）。
- **禁止打印整个请求体**；只打关键字段。
- 日志级别支持配置中心热更新。
- 慢 SQL 阈值默认 200ms，超阈值 WARN；SQL 执行错误 ERROR 并带 `trace_id`。

### 4.2 `pkg/jwt` —— 令牌与会话

```
server/pkg/jwt/
├── jwt.go          # Manager：Sign / Parse / Verify
├── claims.go       # Claims 结构体（uid/uname/roles/jti/ver/iss/aud/exp/iat/nbf）
├── config.go       # Config: Algorithm(HS256|RS256)/Secret/PrivateKey/Issuer/Audience/AccessTTL/RefreshTTL
├── refresh.go      # 不透明 refresh token 的生成/校验/轮转/复用检测
├── blacklist.go    # jti 黑名单（Redis）：Revoke / IsRevoked
└── errs.go         # ErrExpired / ErrInvalid / ErrRevoked / ErrWrongAlg
```

**Claims 设计**

```go
type Claims struct {
    Uid      int64    `json:"uid"`
    Username string   `json:"uname"`
    Roles    []string `json:"roles"`   // 角色码，接口级粗粒度判定
    PermVer  int64    `json:"ver"`     // 权限版本号：用户权限/角色变更时 +1，旧令牌失效
    DeptID   int64    `json:"dept,omitempty"`
    jwt.RegisteredClaims               // Issuer/Subject/Audience/ExpiresAt/NotBefore/IssuedAt/ID(jti)
}
```

**令牌策略**

| 令牌 | 类型 | TTL | 存储 | 撤销方式 |
| --- | --- | --- | --- | --- |
| access_token | JWT | 2 小时 | 客户端 | jti 黑名单（Redis，TTL = 剩余有效期）；PermVer 不匹配拒绝 |
| refresh_token | 随机 32 字节（base64url） | 30 天 | Redis：`nv:{env}:user:refresh:{uid}:{deviceId}` → `sha256(token)` | 删除 Key 即撤销；轮转 + 复用检测 |

**核心接口**

```go
type Manager interface {
    Sign(claims Claims) (string, error)
    Parse(token string) (*Claims, error)                 // 校验签名/过期/算法白名单
    Revoke(ctx context.Context, claims *Claims) error     // 加入黑名单
    IsRevoked(ctx context.Context, jti string) (bool, error)
}

// 刷新（含轮转与复用检测）
func (m *manager) Refresh(ctx context.Context, uid int64, deviceID, refreshToken string) (newAccess, newRefresh string, err error)
```

**安全要求（必做）**

1. **算法白名单**：`Parse` 必须校验 `token.Method.Alg()` 与配置一致，防 `alg=none` 攻击。
2. **密钥强度**：HS256 密钥 ≥ 32 字节随机；生产建议 RS256，私钥仅签发方持有（网关只持公钥验签）。
3. **禁止把敏感信息放 Claims**（Claims 只 base64 编码，不是加密）。
4. 过期时间必须校验，`nbf` 允许 30s 时钟偏移。
5. 权限变更、改密码、封禁 → `PermVer++` 并清空该用户所有 refresh token。
6. 网关验签后**只透传 uid/roles/ver**，不透传原始 token 给下游。

### 4.3 `pkg/crypt` —— 加解密与哈希

```
server/pkg/crypt/
├── password.go     # HashPassword(bcrypt cost=12) / VerifyPassword / NeedRehash（成本升级）
├── aes.go          # AES-256-GCM 加解密（用于敏感字段、第三方密钥加密存储）
├── rsa.go          # RSA-OAEP 加解密与签名（对接需要 RSA 签名的第三方，如部分支付/短信）
├── hmac.go         # HMAC-SHA256 签名与验签（回调验签、URL 签名）
├── random.go       # 随机字符串/数字（验证码、盐、邀请码）；crypto/rand
└── hash.go         # MD5/SHA256/文件 hash（流式计算，用于秒传）
```

**规则**

- 用户密码只能用 **bcrypt**（或 argon2id），**禁止** MD5/SHA1/SHA256 直接存密码（CI 加 `gosec` 卡口）。
- 每次登录校验成功后，若 `bcrypt.Cost` 低于当前配置，自动重新哈希（平滑升级成本）。
- 用于加解密第三方密钥的主密钥（KEK）来自环境变量/Vault，**不进配置文件**。
- 随机数一律 `crypto/rand`，禁止 `math/rand`（CI 卡口）。

### 4.4 `pkg/db` —— GORM 接入层

**这是 go-zero + GORM 结合的关键包**：go-zero 默认用 `sqlx` + 自研缓存，本项目改用 GORM，必须自己把连接、日志、缓存、事务、分页、读写分离这一层做扎实。

```
server/pkg/db/
├── db.go           # New(cfg) (*gorm.DB, error)：初始化 + 连接池 + 插件 + 日志
├── config.go       # Config：DSN/连接池/日志级别/慢 SQL/读写分离/自动迁移开关
├── logger.go       # zap 适配 GORM logger（Trace：慢 SQL + 错误 SQL + trace_id）
├── plugin_trace.go # GORM 插件：从 ctx 取 trace_id 注入日志，统计指标
├── plugin_metric.go# 指标：SQL 次数、耗时直方图、错误数（Prometheus）
├── transaction.go  # Tx(ctx, fn) 助手：自动 begin/commit/rollback + 嵌套事务（SavePoint）
├── paginate.go     # Scope 分页：Paginate(page,size) / CursorPaginate(id,size)
├── softdelete.go   # 软删除约定与唯一索引冲突的解决方案说明
├── scope.go        # 常用 Scope：InIds、LikeIfNotBlank、TimeRange、OrderBySafe(防注入)
├── migrate.go      # 迁移入口（仅 dev 可用 AutoMigrate；test/prod 用 goose）
└── errors.go       # 错误归一化：1062 唯一键冲突 → errx.ErrDuplicate；1213 死锁 → 可重试
```

**初始化示例**

```go
gdb, err := db.New(db.Config{
    DSN:             cfg.DB.DSN,          // user:pass@tcp(host:port)/db?charset=utf8mb4&parseTime=true&loc=UTC
    MaxIdleConns:    10,
    MaxOpenConns:    100,                 // ≈ 数据库最大连接数的 60% / 服务实例数
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: 10 * time.Minute,
    SlowThreshold:   200 * time.Millisecond,
    LogLevel:        "warn",              // silent|error|warn|info
    PrepareStmt:     true,                // 预编译，降低解析开销
    SkipDefaultTransaction: true,         // 单条写不开启事务，性能更好
    ReadWriteSplit:  cfg.DB.ReadWriteSplit, // 主从分离（dbresolver）
})
```

**GORM 配置基线（写入 `db.New`，不可被业务覆盖）**

```go
&gorm.Config{
    SkipDefaultTransaction: true,
    PrepareStmt:            true,
    Logger:                 logger,                 // pkg/db.logger
    NowFunc:                func() time.Time { return time.Now().UTC() },
    DisableForeignKeyConstraintWhenMigrating: true, // 外键由应用层保证，避免迁移受限
    CreateBatchSize:        200,
}
```

**事务规范**

```go
// ✅ 正确：事务边界在 logic，SQL 细节在 repository（通过 Tx 传递 *gorm.DB）
err := db.Tx(ctx, func(tx *gorm.DB) error {
    if err := s.articleRepo.WithTx(tx).Create(ctx, article); err != nil { return err }
    if err := s.outboxRepo.WithTx(tx).Append(ctx, evt); err != nil { return err }
    return nil
})

// ❌ 禁止：logic 里出现 db.Model(&Article{}).Where(...) 这类直接调用
```

**Repository 模式约定**

```go
// app/article/model/repository/article_repo.go
type ArticleRepo interface {
    Create(ctx context.Context, po *entity.Article) error
    GetByID(ctx context.Context, id int64) (*entity.Article, error)     // 命中缓存
    GetContent(ctx context.Context, id int64) (*entity.ArticleContent, error)
    Update(ctx context.Context, po *entity.Article, fields ...string) error
    List(ctx context.Context, q ArticleQuery) ([]*entity.Article, int64, error)
    UpdateCounters(ctx context.Context, id int64, delta map[string]int64) error
    Delete(ctx context.Context, id int64) error                          // 软删除
    WithTx(tx *gorm.DB) ArticleRepo                                      // 事务传播
}
```

**关键约定**

| 项 | 约定 |
| --- | --- |
| 主键 | `int64` 雪花 ID（`pkg/idgen`），字段名 `id`，`bigint unsigned` |
| 公共字段 | `created_at`、`updated_at`、`deleted_at`（软删除表）；`created_by`、`updated_by` 仅在需要审计的表加 |
| 软删除 | 内容/用户类用 `gorm.DeletedAt`；**统计/日志/流水类不软删**（用 `status` 或分区归档） |
| 软删除 + 唯一索引冲突 | MySQL 中 `deleted_at` 为 NULL 时唯一索引不生效重复约束。方案：唯一索引改为 `(col, deleted_at)` 并把 `deleted_at` 默认值设为 `0`（用 `int64` 而非 `*time.Time`），或删除时改写 `col`（加时间戳后缀）。**项目统一采用前者**（`deleted_at bigint not null default 0`） |
| 布尔 | `tinyint(1)`，Go 侧 `bool`；禁止用 `is_xxx` 存枚举 |
| 金额/比率 | 一律用整数最小单位（分/万分比），不用 float |
| JSON 字段 | 用 `datatypes.JSON` 或自定义 `JSONMap`，读取必须有默认值兜底 |
| 枚举 | DB 存 `tinyint`，Go 侧定义 `type XxxStatus int8` + `String()` + `Valid()` |
| 时间 | 存 UTC，`datetime(3)`；区间查询用 `>= AND <`（半开区间） |
| 分页深翻页 | `page > 1000` 时改用游标分页（`WHERE id < last_id ORDER BY id DESC LIMIT n`） |
| 排序字段 | 只允许白名单字段（`db.OrderBySafe`），禁止把用户输入拼进 `Order` |
| 批量写入 | 用 `CreateInBatches(..., 200)` |
| N+1 | 列表查询必须 `Preload` 或二次批量查询；禁止在循环里查库 |
| 大字段 | 正文等 LONGTEXT 拆表（`article_content`），列表查询绝不 `SELECT *` 带正文 |

**迁移策略**

- **dev**：可 `AutoMigrate` 快速迭代（配置开关 `AutoMigrate: true`）。
- **test/prod**：**禁止 AutoMigrate**。使用 `goose` 管理 `server/deploy/sql/migrations/` 下的版本化 SQL，由 `app/job` 的迁移任务或 CI 在发布前执行。
- 迁移必须**向后兼容**（先加列/加索引，再改代码，最后删旧列），保证滚动发布期间新旧版本共存可用。

### 4.5 `pkg/cache` —— Redis 封装

```
server/pkg/cache/
├── redis.go        # New(cfg) *redis.Client（go-redis v9）+ 连接池 + 超时 + 钩子
├── config.go       # Config：Addr/Password/DB/PoolSize/MinIdleConns/DialTimeout/ReadTimeout/KeyPrefix
├── keys.go         # Key 构造器与规范（含 env 前缀、防冲突）
├── codec.go        # JSON/MsgPack 编解码 + 空值占位（NULL_SENTINEL）
├── lock.go         # 分布式锁：TryLock/Unlock（Lua 保证原子）+ 自动续期
├── limiter.go      # 滑动窗口/令牌桶限流（本地 + Redis 全局双层）
├── counter.go      # 计数器、HyperLogLog UV、Bitmap 签到、ZSet 排行榜
├── singleflight.go # 回源合并（防缓存击穿）
├── stream.go       # Redis Stream 生产/消费（轻量 MQ，见 pkg/mq）
└── bloom.go        # 布隆过滤器（防缓存穿透，可选）
```

**客户端基线**

```go
redis.NewClient(&redis.Options{
    Addr:         cfg.Addr,
    Password:     cfg.Password,
    DB:           cfg.DB,
    PoolSize:     100,           // 每实例连接数
    MinIdleConns: 10,
    DialTimeout:  2 * time.Second,
    ReadTimeout:  1 * time.Second,
    WriteTimeout: 1 * time.Second,
    PoolTimeout:  2 * time.Second,
    MaxRetries:   2,
})
```

**分布式锁规范**

```go
lock := cache.NewLock(rdb, "lock:article:publish:"+id, 10*time.Second)
ok, err := lock.TryLock(ctx)      // SET key value NX PX ttl
if !ok { return errx.ErrLocked }
defer lock.Unlock(ctx)            // Lua: GET == value 才 DEL（防误删他人锁）
```

- 锁必须设 TTL；长任务用 `lock.Renew(ctx)` 续期（看门狗协程）。
- 获取失败**不要自旋重试超过 3 次**，直接返回"操作进行中"。
- **不要用锁替代唯一索引**：防重复提交首选"幂等表 + 唯一索引"，锁只是减少无效竞争。

**缓存三防（必须实现）**

| 问题 | 方案 | 实现位置 |
| --- | --- | --- |
| 穿透（查不存在的 Key 打库） | 空值占位缓存（`"__NULL__"`，TTL 60s）+ 可选布隆过滤器 | `cache/codec.go`、`cache/bloom.go` |
| 击穿（热点 Key 失效瞬间） | `singleflight` 合并回源 + 逻辑过期（热点 Key 不设物理 TTL，后台异步刷新） | `cache/singleflight.go` |
| 雪崩（大批 Key 同时失效） | TTL 加 ±10% 随机抖动；不同业务 TTL 分级；Redis 集群/哨兵 | `cache/keys.go` |

### 4.6 `pkg/errx` + `pkg/response` —— 错误与响应

```
server/pkg/errx/
├── code.go         # 错误码常量（分段，见 8.3）与 Code 类型
├── errx.go         # New(code, msg) / Wrap(err, code) / WithMsg / WithDetail
├── map.go          # Code → HTTP Status / gRPC Code 映射表
├── handler.go      # go-zero httpx.SetErrorHandler 与 zrpc 错误拦截器接入
└── detect.go       # 识别 context.DeadlineExceeded、gorm.ErrRecordNotFound 等并归一化

server/pkg/response/
├── response.go     # Ok(ctx, data) / Fail(ctx, err) / FailWithCode(ctx, code, msg)
├── page.go         # PageResp{List, Total, Page, PageSize} / CursorResp{List, NextCursor, HasMore}
└── types.go        # 通用响应结构体定义（与 front/shared 类型同源）
```

**统一响应体（所有对外接口必须遵守）**

```json
{
  "code": 0,
  "msg": "ok",
  "data": { },
  "trace_id": "3f2b...",
  "server_time": 1735689600000
}
```

- `code = 0` 表示成功；非 0 见 [8.3 错误码规范](#83-错误码规范)。
- HTTP 状态码只表达**传输层**语义（200/400/401/403/404/429/500），业务失败一律 HTTP 200 + 非 0 `code`（RESTful 纯粹主义者可改用语义化状态码，**但全项目必须二选一，本项目选择"HTTP 200 + 业务码"**，理由：前端拦截器逻辑简单、网关统一、避免 CDN/代理改写 4xx/5xx）。
- **例外**：鉴权失败返回 `401`，无权限返回 `403`，限流返回 `429`——便于网关与前端做通用处理，同时 `code` 也给出细分。
- 生产环境**禁止**把 stack trace、SQL、内部地址返回给前端；错误详情只进日志。

**使用示例**

```go
// errorx → 对外错误
if user == nil {
    return nil, errx.New(errx.CodeUserNotFound)                       // 40101 用户不存在
}
if article.UserID != uid {
    return nil, errx.New(errx.CodeNoPermission, errx.WithMsg("无权修改他人文章"))
}
```

### 4.7 `pkg/middleware` —— 通用中间件与拦截器

```
server/pkg/middleware/
├── auth.go          # HTTP：解析 JWT → 注入 x-uid 等 metadata（网关用）
├── authz.go         # HTTP：路由 → 权限码校验（admin 用）
├── trace.go         # 生成/透传 trace_id、request_id
├── accesslog.go     # 访问日志（含耗时、状态、错误码、uid、UA）
├── recovery.go      # panic 恢复 → 500 + 日志 + 指标
├── cors.go          # 跨域（白名单域名，禁止 * 与 credentials 同开）
├── ratelimit.go     # 限流（按 IP / 按 uid / 按接口）：本地令牌桶 + Redis 全局
├── signature.go     # 开放 API 签名校验（HMAC + 时间戳 + nonce 防重放）
├── opLog.go         # 后台操作日志采集（异步落库）
├── rpcmeta.go       # zrpc 拦截器：从 metadata 还原 uid/trace 到 context
├── rpcauth.go       # zrpc 拦截器：校验 x-internal-token（非网关来源）
└── timeout.go       # 请求级超时（默认读 3s / 写 5s，可注解覆盖）
```

**限流矩阵（基线，可在配置中心热调）**

| 维度 | 场景 | 阈值（初版） | 超限响应 |
| --- | --- | --- | --- |
| 全局 IP | 所有接口 | 600 次/分钟 | 429 + `CodeRateLimited` |
| 登录 | `/auth/login` | 10 次/分钟/IP，5 次失败锁 15 分钟 | 429 |
| 注册 | `/auth/register` | 5 次/小时/IP | 429 |
| 验证码 | `/auth/captcha/send` | 1 次/60 秒，10 次/天/邮箱 | 429 |
| 发文 | 创建/更新文章 | 30 次/小时/uid | 429 |
| 评论 | 发表评论 | 20 次/10 分钟/uid | 429 |
| 点赞 | 点赞/收藏 | 120 次/分钟/uid | 429 |
| 上传 | 文件上传 | 100 次/小时/uid，单文件 ≤ 20MB | 429 / 413 |
| 第三方 | 翻译/AI/审核 | 按供应商配额分配（见 5.7） | 503 + 降级标记 |
| 搜索 | 全文检索 | 60 次/分钟/IP | 429 |

### 4.8 `pkg/idgen` —— ID 生成

```
server/pkg/idgen/
├── snowflake.go    # 雪花：workerID 从 etcd 分配（或配置注入），返回 int64
├── uuid.go         # UUID v4 / v7（对外暴露的不可枚举标识）
├── shortcode.go    # 短码（分享链接、邀请码）：Base62 + 冲突重试
└── slug.go         # 文章 slug 生成（标题转拼音/英文 + 去重后缀 -2）
```

**规则**

- 内部主键统一雪花 `int64`；对外可枚举场景（分享链接、附件下载）额外用 `uuid` 或 `slug`。
- workerID 分配：dev 用配置固定值；生产通过 etcd 租约分配（`/noviterra/{env}/idgen/workers/{host}`），避免多实例重复。
- 时钟回拨处理：回拨 ≤ 5ms 自旋等待，> 5ms 直接返回错误并告警（禁止阻塞业务）。

### 4.9 其它共享包

| 包 | 关键 API | 说明 |
| --- | --- | --- |
| `pkg/paginate` | `Parse(page,size) → (offset,limit)`、`Normalize` | page ≥ 1，size ∈ [1,100]，默认 10；超出改游标 |
| `pkg/conv` | `ToInt64/ToStringSafe/ToBool`、`Unique[T]`、`Chunk[T]`、`Filter/Map` | 泛型工具；nil 安全；批量分片（RPC/DB 批查） |
| `pkg/validatorx` | `Register()` 注册自定义规则 | 用户名（4-20 位字母数字下划线）、密码强度、slug、手机号、URL、安全文件名 |
| `pkg/ctxx` | `WithUid/Uid`、`WithTrace/TraceID`、`WithClientIP`、`TimeoutOrDefault` | context 键统一，禁止用裸 string 作 key |
| `pkg/constants` | 枚举（文章状态、评论状态、通知类型）、正则、限制值、缓存 TTL 常量 | **与前端共享的枚举以这里为单一来源** |
| `pkg/mq` | `Producer.Publish(ctx, topic, payload)`、`Consumer.Subscribe(topic, handler)` | 实现：`redisstream`（起步）/`asynq`（重试友好）/`kafka`（规模化）；事件信封含 `event_id`（幂等键）、`trace_id`、`occurred_at`、`version` |
| `pkg/observability` | `InitTrace(cfg)`、`RegisterMetrics()`、`Healthz()` | OTLP 接入 Jaeger/Tempo；Prometheus 指标注册；健康检查聚合（DB/Redis/依赖服务） |
| `pkg/schedule` | `Every(expr, fn)`、`WithLock(key)` | 基于 `robfig/cron/v3` + 分布式锁，供 app/job 与各服务内部定时任务使用 |
| `pkg/version` | `Version/Commit/BuildTime` | 由 `-ldflags "-X ..."` 注入，`/version` 接口与启动日志输出 |

**`pkg/mq` 事件信封（统一定义，禁止各服务自定义格式）**

```go
type Envelope struct {
    EventID    string          `json:"event_id"`     // 雪花/UUID，消费方幂等键
    Topic      string          `json:"topic"`        // article.published.v1
    OccurredAt int64           `json:"occurred_at"`  // 毫秒时间戳
    TraceID    string          `json:"trace_id"`
    Producer   string          `json:"producer"`     // article-rpc
    Version    int             `json:"version"`      // 载荷版本
    Payload    json.RawMessage `json:"payload"`
}
```

**消费端强制要求**：幂等（按 `event_id` 去重，Redis SETNX + TTL 24h 或 `processed_event` 表）、失败重试（指数退避，最多 5 次）、超过次数进死信（`{topic}.dlq`）并告警。

---

## 5. 第三方 API 接入规范（重点）

### 5.1 核心结论：第三方代码分三层放，绝不混写

**用户最关心的问题**："百度翻译这样的第三方 API 工具，写在哪个项目？"

**答案：写在 `server/pkg/thirdparty/` 下（作为共享能力包），但"调用它的业务流程"写在 `server/app/third/`（third-rpc 服务）里，而"在文章上点翻译按钮"这个动作在 `server/app/article/` 的 logic 里发起。**

三层职责，**必须严格分离**：

```
┌────────────────────────────────────────────────────────────────────────────┐
│ 第 1 层：能力接口（Capability Interface）                                   │
│ 位置：server/pkg/thirdparty/translate/translate.go                          │
│ 内容：type Translator interface { Translate(ctx, *Req) (*Resp, error) }     │
│ 特点：定义"能做什么"，不含任何供应商细节；业务层只依赖这个接口               │
├────────────────────────────────────────────────────────────────────────────┤
│ 第 2 层：供应商适配器（Vendor Adapter）                                     │
│ 位置：server/pkg/thirdparty/translate/baidu/baidu.go                        │
│ 内容：HTTP 调用、签名、报文映射、错误码翻译、限流重试熔断                    │
│ 特点：只做"协议适配"，不碰数据库、不做业务判断、不知道"文章"是什么           │
├────────────────────────────────────────────────────────────────────────────┤
│ 第 3 层：业务编排（Business Orchestration）                                 │
│ 位置：server/app/third/rpc/internal/logic/translatelogic.go                │
│ 内容：命中缓存判断、配额扣减、结果落库(article_translation)、审计、降级决策   │
│ 特点：有数据库、有业务规则；通过 svc 注入的 Translator 接口调用第 2 层        │
└────────────────────────────────────────────────────────────────────────────┘
```

**一句话记忆**：`pkg/thirdparty` = **能调用**，`app/third` = **该不该调、调完存哪、超了怎么办**，`app/article` = **谁在什么场景下触发**。

### 5.2 目录结构

```
server/pkg/thirdparty/
├── README.md                # 本目录的接入规范与新增供应商步骤（必读）
├── provider.go              # 通用：Provider 元信息、注册表、工厂函数
├── options.go               # 统一 Option：超时/重试/熔断/限流/缓存/降级
├── transport.go             # 统一 http.Client（连接池、超时、代理、UA、重试传输层）
├── errors.go                # 错误归一化：ErrAuth/ErrQuota/ErrRateLimit/ErrNetwork/ErrContent/ErrTimeout
├── retry.go                 # 指数退避 + 抖动重试（仅幂等请求；区分可重试错误）
├── breaker.go               # 熔断包装（go-zero breaker 或 sony/gobreaker）
├── ratelimit.go             # 本地令牌桶（QPS）+ Redis 全局限额（日/月配额）
├── cache.go                 # 结果缓存装饰器（相同请求命中 Redis）
├── metrics.go               # 指标：调用量、耗时、失败率、配额余量（Prometheus）
├── audit.go                 # 调用审计：请求/响应摘要（脱敏）→ 交给上层落库
├── mask.go                  # 日志脱敏（密钥、手机号、邮箱、身份证）
│
├── translate/               # 【翻译能力】
│   ├── translate.go         # interface Translator + Req/Resp DTO + Lang 枚举
│   ├── factory.go           # New(provider string, cfg Config) (Translator, error)
│   ├── baidu/baidu.go       # 百度翻译（通用文本翻译 API）
│   ├── tencent/tencent.go   # 腾讯云 TMT（TC3-HMAC-SHA256 签名）
│   ├── deepl/deepl.go       # DeepL
│   └── mock/mock.go         # 测试用实现（返回固定/回显结果）
│
├── ai/                      # 【大模型能力】
│   ├── ai.go                # interface LLM: Chat / Summarize / Polish / Embed / Moderate
│   ├── factory.go
│   ├── openai/openai.go     # OpenAI 兼容协议
│   ├── deepseek/deepseek.go # DeepSeek（OpenAI 兼容，可复用 openai 实现 + 换 BaseURL）
│   ├── ollama/ollama.go     # 本地模型（开发环境零成本）
│   └── prompt/              # 提示词模板（摘要、标题生成、标签推荐、润色）
│
├── ocr/                     # 【文字识别】百度 OCR / 腾讯 OCR（图片转文字，用于扫描稿）
├── contentsec/              # 【内容安全】百度内容审核 / 阿里云绿网 / 腾讯天御 + 本地 DFA 词库
├── storage/                 # 【对象存储】见 5.6.1
│   ├── storage.go           # interface Storage: Put/Get/Delete/Stat/SignURL/InitMultipart/UploadPart/Complete
│   ├── local/local.go       # 本地磁盘（dev）
│   ├── minio/minio.go       # MinIO / S3 兼容
│   ├── aliyunoss/aliyunoss.go
│   ├── tencentcos/tencentcos.go
│   └── image/               # 图片处理：压缩、缩放、水印、WebP、EXIF 清理（本地 imaging 或云处理参数）
├── sms/                     # 【短信】阿里云 / 腾讯云 / 云片（模板短信、验证码）
├── email/                   # 【邮件】SMTP / Resend / SendGrid（HTML 模板 + 附件）
├── ipregion/                # 【IP 归属地】ip2region 离线库（推荐，无网络依赖）/ 高德 / 太平洋
├── captcha/                 # 【人机验证】本地图形码 / 腾讯滑块 / hCaptcha / Cloudflare Turnstile
├── oauth/                   # 【第三方登录】github / gitee / wechat / google
├── push/                    # 【推送】WebSocket 站内推送 / 邮件订阅 / 浏览器 Web Push
└── payment/                 # 【可选】赞助/打赏/会员：支付宝 / 微信支付（预留）
```

**规则**：新增一个第三方供应商 = **新增一个目录 + 实现已有接口 + 注册工厂 + 加配置项 + 加测试**，**不得改动业务代码**。若必须改业务代码，说明抽象漏了。

### 5.3 接口定义示例（以翻译为例）

```go
// server/pkg/thirdparty/translate/translate.go
package translate

import "context"

// Lang 语种（统一内部编码，各供应商在适配器内自行映射）
type Lang string

const (
    LangAuto Lang = "auto"
    LangZH   Lang = "zh"
    LangEN   Lang = "en"
    LangJA   Lang = "ja"
    LangKO   Lang = "ko"
    LangFR   Lang = "fr"
    LangDE   Lang = "de"
    LangRU   Lang = "ru"
    LangES   Lang = "es"
)

// Req 翻译请求（供应商无关）
type Req struct {
    Text     string   // 单次文本（适配器内部按供应商限制自动分段）
    Texts    []string // 批量模式（优先使用）
    From     Lang
    To       Lang
    Domain   string   // 领域：通用/IT/医学（可选，百度支持 vertical）
    TraceID  string   // 透传，便于审计
}

// Resp 翻译响应（供应商无关）
type Resp struct {
    Items    []Item // 与请求顺序一一对应
    Provider string // 实际使用的供应商
    Latency  int64  // 毫秒
    Cached   bool   // 是否命中缓存
}

type Item struct {
    Src string
    Dst string
}

// Translator 翻译能力接口（业务层唯一依赖）
type Translator interface {
    // Name 供应商名（用于日志、指标、审计）
    Name() string
    // Translate 执行翻译；上层负责缓存、配额、落库
    Translate(ctx context.Context, req *Req) (*Resp, error)
    // Health 健康检查（可选，用于就绪探针与降级判断）
    Health(ctx context.Context) error
}
```

```go
// server/pkg/thirdparty/translate/baidu/baidu.go（百度翻译适配器，示意）
package baidu

const (
    defaultEndpoint = "https://fanyi-api.baidu.com/api/trans/vip/translate"
    maxTextLen      = 6000 // 百度单次上限，超过需分段
)

// sign = MD5(appid + q + salt + 密钥)
func (c *client) sign(q, salt string) string {
    raw := c.appID + q + salt + c.appKey
    sum := md5.Sum([]byte(raw))
    return hex.EncodeToString(sum[:])
}

func (c *client) doTranslate(ctx context.Context, q string, from, to translate.Lang) (string, error) {
    salt := crypt.RandomDigits(8)
    vals := url.Values{}
    vals.Set("q", q)
    vals.Set("from", string(from))
    vals.Set("to", string(to))
    vals.Set("appid", c.appID)
    vals.Set("salt", salt)
    vals.Set("sign", c.sign(q, salt))

    var out apiResp // { from, to, trans_result:[{src,dst}], error_code, error_msg }
    if err := c.http.PostForm(ctx, c.endpoint, vals, &out); err != nil {
        return "", err   // pkg/thirdparty/errors.go 已归一化为 ErrNetwork/ErrTimeout
    }
    if out.ErrorCode != "" && out.ErrorCode != "52000" {
        return "", mapBaiduErr(out.ErrorCode, out.ErrorMsg) // 54003=QPS超限 54004=余额不足 52001=超时
    }
    // ... 拼接 trans_result
}
```

**适配器必须遵守**

1. 只依赖 `pkg/thirdparty` 的 `transport/retry/breaker/ratelimit/errors`，**禁止**自己 `http.Get`、自己 `time.Sleep` 重试。
2. 供应商错误码必须映射到统一错误（`errors.go` 的 6 类），上层只认统一错误。
3. 单次请求超时默认 **3s**（翻译）/ **10s**（AI 生成）/ **15s**（OCR）；外部总超时由上层 ctx 控制。
4. 分段/分页逻辑（如百度 6000 字符上限）在适配器内实现，对上层透明。
5. 日志只记长度与 hash，**不记原文**（隐私）；错误日志记供应商原始 `error_code`。
6. 任何适配器都要有 `mock` 或 `httptest` 级别的单元测试。

### 5.4 统一客户端基线能力（每个外部调用都必须具备）

| 能力 | 实现位置 | 默认参数 | 说明 |
| --- | --- | --- | --- |
| 连接池 | `transport.go` | `MaxIdleConnsPerHost: 50`、`IdleConnTimeout: 90s` | 复用 TCP，降低握手开销 |
| 超时 | `transport.go` + ctx | 连接 2s / 读 10s / 整体 3~15s 按能力 | 禁无限等待 |
| DNS 缓存 | `transport.go` | 5 分钟 | 避免高频 DNS |
| 重试 | `retry.go` | 最多 2 次，基数 200ms，指数退避 + ±20% 抖动 | **仅对幂等请求**（GET/翻译/查询） |
| 熔断 | `breaker.go` | 连续 20 次失败率 > 50% 开启，30s 半开 | 防止拖垮自身 |
| 限流 | `ratelimit.go` | 本地令牌桶（供应商 QPS 的 80%）+ Redis 全局限额 | 多实例共享配额 |
| 缓存 | `cache.go` | 翻译 7 天、AI 摘要 30 天、IP 归属 30 天 | Key 含内容 hash，避免重复计费 |
| 降级 | 各能力定义 | 见下表 | **必须显式**，禁止静默返回空 |
| 指标 | `metrics.go` | `thirdparty_request_total{provider,capability,result}`、`..._duration_seconds`、`..._quota_remaining` | 告警依据 |
| 审计 | `audit.go` | 记录调用方、能力、供应商、耗时、成败、错误码、计费量 | 落 `third_api_log` |
| 脱敏 | `mask.go` | 手机号/邮箱/密钥 | 日志与审计均脱敏 |

**降级策略表（必须写进代码与文档，禁止口头约定）**

| 能力 | 失败时的行为 | 是否影响主链路 | 用户可见表现 |
| --- | --- | --- | --- |
| 翻译 | 返回原文 + `status=degraded`，进重试队列 | 否 | "翻译暂不可用，请稍后重试" |
| AI 摘要 | 跳过该步骤，文章照常发布，后台可重试 | 否 | 无摘要（不报错） |
| 内容安全审核（发布时） | **fail-closed**：转人工审核队列 | 是（延迟发布） | "已提交审核" |
| 内容安全审核（评论时） | **fail-safe**：本地词库先过，未命中则放行并打标待复核 | 否 | 正常显示（后台可见标记） |
| IP 归属地 | 返回"未知" | 否 | 显示"IP 属地：未知" |
| 短信验证码 | 主通道失败切备用通道；全失败则提示稍后再试 | 是 | "验证码发送失败" |
| 邮件 | 进重试队列，最多 5 次 | 否 | 稍后收到 |
| 对象存储 | 主存储失败切备用（若配置） | 是 | "上传失败，请重试" |
| 第三方登录 | 提示改用账号密码登录 | 否 | 明确提示 |

> **fail-open / fail-closed 的选择必须显式决策并写入 ADR**（`doc/adr/`）。涉及合规（内容安全）必须 fail-closed。

### 5.5 配置与密钥管理

**配置结构（`etc/<svc>.yaml` 与 etcd 配置中心同构）**

```yaml
ThirdParty:
  # 能力 → 供应商选择（支持主备）
  Translate:
    Provider: baidu            # baidu | tencent | deepl | mock
    FallbackProvider: tencent  # 主通道熔断后自动切换
    Timeout: 3000
    Retry: 2
    CacheTTL: 604800           # 7 天（秒）
    QPS: 8                     # 供应商配额
    DailyQuota: 100000         # 每日免费额度
  AI:
    Provider: deepseek
    Model: deepseek-chat
    BaseURL: https://api.deepseek.com/v1
    Timeout: 20000
    MaxTokens: 2048
    Temperature: 0.7
    CacheTTL: 2592000
  ContentSec:
    Provider: baidu
    FailMode: closed           # closed | open —— 必须显式
    LocalWordLibVersion: v1
  Storage:
    Provider: minio
    Endpoint: minio:9000
    Bucket: noviterra
    UseSSL: false
    SignURLExpire: 3600
  SMS:   { Provider: aliyun, SignName: "Noviterra", TemplateCode: "SMS_xxx" }
  Email: { Provider: smtp, Host: smtp.exmail.qq.com, Port: 465, From: "noreply@noviterra.dev" }
  IpRegion: { Provider: ip2region, DataFile: "./data/ip2region.xdb" }
  Captcha:  { Provider: local, Length: 4, Expire: 300 }
  OAuth:
    Github: { ClientID: "", Scopes: ["user:email"] }

# ★ 密钥一律不写在配置文件里
# 使用 ${ENV_NAME} 占位，由 conf 的 UseEnv 机制或启动脚本注入
ThirdPartySecrets:
  TranslateBaiduAppID:  ${NV_BAIDU_TRANSLATE_APPID}
  TranslateBaiduKey:    ${NV_BAIDU_TRANSLATE_KEY}
  AIApiKey:             ${NV_DEEPSEEK_API_KEY}
  StorageAccessKey:     ${NV_MINIO_ACCESS_KEY}
  StorageSecretKey:     ${NV_MINIO_SECRET_KEY}
  SMSAccessKey:         ${NV_ALIYUN_SMS_AK}
  SMTPPassword:         ${NV_SMTP_PASSWORD}
```

**密钥管理规则（安全红线）**

| 环境 | 密钥来源 | 要求 |
| --- | --- | --- |
| dev | 项目根 `.env`（**必须 gitignore**） | 用第三方沙箱/测试额度 |
| test | K8s Secret 或 CI 变量 | 不写入 etcd 明文 |
| prod | Vault / K8s Secret（SealedSecret）+ 定期轮换 | 至少 90 天轮换一次；离职/泄漏立即轮换 |

- 加载顺序：**环境变量 > etcd 配置中心 > 本地 yaml**；同名以高优先级覆盖。
- 启动时做**配置自检**：必填项缺失直接启动失败并打出清晰错误（禁止带着空密钥启动）。
- 日志与错误信息中**密钥必须整体遮蔽**（只留前 4 位）。
- 生产环境禁止 `GET /debug/config` 之类接口。

### 5.6 第三方清单与选型建议

#### 5.6.1 已规划清单

| 能力 | 首选 | 备选 | 免费额度/成本 | 触发场景 | 优先级 |
| --- | --- | --- | --- | --- | --- |
| 文本翻译 | 百度翻译开放平台（通用文本翻译） | 腾讯云 TMT、DeepL | 百度标准版有免费额度，超出按字符计费 | 文章/评论一键翻译、多语言站点 | P2 |
| 大模型（摘要/润色/标签/标题） | DeepSeek（性价比高，OpenAI 兼容） | OpenAI、通义千问、本地 Ollama | 按 token 计费 | 发布后自动摘要、标签推荐、SEO 描述 | P2 |
| 文字识别 OCR | 百度 OCR | 腾讯 OCR | 有免费额度 | 扫描稿/图片文章导入 | P3 |
| 内容安全审核 | 百度内容审核 | 阿里云绿网、腾讯天御 | 按量计费 | 文章发布、评论提交、昵称/头像 | P1 |
| 本地敏感词 | 自建 DFA（`contentsec` 内） | — | 0 | 全量文本初筛（先本地后云） | **P0** |
| 对象存储 | MinIO（自建，dev+小规模） | 阿里云 OSS、腾讯 COS、七牛 | 自建成本可控 | 所有上传 | **P0** |
| 短信 | 阿里云短信 | 腾讯云短信、云片 | 按条计费 | 手机号注册/登录/改密 | P2 |
| 邮件 | 企业邮箱 SMTP | Resend、SendGrid、阿里云邮件推送 | 免费额度 | 邮箱验证、通知、订阅推送 | P1 |
| IP 归属地 | `ip2region` 离线库 | 高德 Web API、太平洋 IP | 离线库 0 成本 | 评论/登录显示属地（合规要求） | P1 |
| 人机验证 | 本地图形验证码 | 腾讯滑块、hCaptcha、Turnstile | Turnstile 免费 | 注册、登录、发评论 | **P0** |
| 第三方登录 | GitHub OAuth | Gitee、微信、Google | 免费 | 快速注册登录 | P2 |
| 地图/天气等 | 暂不接入 | — | — | — | — |
| 支付（如做会员/打赏） | 支付宝当面付 | 微信支付 | 需资质 | 赞助、会员 | P3 |

#### 5.6.2 `storage` 接口设计（最常用的第三方能力）

```go
// server/pkg/thirdparty/storage/storage.go
type Storage interface {
    Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (*Object, error)
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Stat(ctx context.Context, key string) (*Object, error)
    SignURL(ctx context.Context, key string, expire time.Duration, method string) (string, error)
    // 分片上传（大文件）
    InitMultipart(ctx context.Context, key string, contentType string) (uploadID string, err error)
    UploadPart(ctx context.Context, key, uploadID string, index int32, r io.Reader, size int64) (etag string, err error)
    CompleteMultipart(ctx context.Context, key, uploadID string, parts []Part) (*Object, error)
    AbortMultipart(ctx context.Context, key, uploadID string) error
}

type Object struct {
    Key         string
    Size        int64
    ETag        string
    ContentType string
    URL         string    // 公开可访问 URL（私有桶时为空）
    ModTime     time.Time
}
```

**对象 Key 命名规范**：`{biz}/{yyyymm}/{dd}/{uuid}.{ext}`，例如 `article/202501/15/8f3c....webp`、`avatar/202501/15/....png`。
`biz` 取值：`avatar`、`cover`、`article`、`comment`、`attachment`、`temp`。
**禁止**使用用户上传的原始文件名作为 Key（路径穿越、同名覆盖、非法字符）。

### 5.7 配额、计量与成本控制

```
server/app/third/rpc/internal/
├── logic/
│   ├── translatelogic.go       # 翻译编排
│   ├── summarizeLogic.go       # AI 摘要
│   ├── moderateLogic.go        # 内容审核
│   ├── ipregionlogic.go        # IP 归属地
│   └── quota_logic.go          # 配额查询/扣减/告警
└── task/
    └── quotareset_task.go      # 每日 00:05 重置日配额、汇总用量、生成报表
```

**配额机制（双层）**

1. **本地 QPS 令牌桶**（进程内，防止瞬时打满供应商 QPS）。
2. **Redis 全局限额**：`nv:{env}:third:quota:{provider}:{capability}:{yyyymmdd}` → 已用量（`INCRBY`，TTL 到下一天）。
3. **兜底**：`third_api_quota` 表存配置额度（可在后台修改），Redis 丢失时从 DB 重建。
4. **告警**：用量达 80% WARN、95% ERROR 并触发通知；达 100% 直接走降级策略。

**成本控制手段（按优先级）**

1. 结果缓存（翻译、AI 摘要按内容 hash 缓存，命中率通常 > 60%）。
2. 本地优先（敏感词用本地 DFA；IP 归属用离线库）。
3. 异步化 + 批量化（AI 摘要走队列，一次处理多篇；翻译批量提交）。
4. 内容去重（相同正文只计费一次）。
5. 限额开关（配置中心一键关闭非核心第三方能力）。

### 5.8 落库与审计

**表 1：`third_api_log`（调用流水，按月分表或分区）**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | bigint | 雪花 |
| capability | varchar(32) | translate / ai / ocr / contentsec / sms / email / storage |
| provider | varchar(32) | baidu / deepseek / aliyun … |
| caller | varchar(64) | 调用方服务，如 `third-rpc`、`article-rpc` |
| biz_type / biz_id | varchar(32)/bigint | 关联业务（如 article/1001） |
| req_hash | char(64) | 请求内容 SHA256（去重与计费依据） |
| req_size | int | 计费量（字符数 / token 数 / 条数） |
| success | tinyint | 1/0 |
| errcode | int | 统一错误码 |
| provider_code | varchar(32) | 供应商原始错误码（排障用） |
| latency_ms | int | 耗时 |
| cached | tinyint | 是否命中缓存（命中不重复计费） |
| trace_id | varchar(64) | 链路 |
| created_at | datetime(3) | 时间（按此字段分区/月度归档） |

**表 2：`third_api_quota`**：`provider + capability` 维度的日/月额度、已用量、告警阈值、是否启用。

**表 3：`article_translation`**（业务结果，归 `noviterra_content` 库，但由 `third-rpc` 写入——**约定：翻译结果的所有权归 content 域，third-rpc 通过 repository 接口写入自己的库？不允许跨库**）

> ⚠️ **重要修正**：`third-rpc` 不能写 `noviterra_content` 库。正确做法二选一：
> - **方案 A（推荐）**：`third-rpc` 只返回翻译结果，**由 `article-rpc` 落库** `article_translation`。`third-rpc` 的库只放 `third_api_log`、`third_api_quota`。
> - **方案 B**：`article_translation` 表归 `third-rpc`，`article-rpc` 通过 RPC 查询。缺点是 content 域的查询要跨服务。
>
> **决策：采用方案 A**——数据所有权清晰，翻译是"内容的一种属性"，属主是 content 域。

### 5.9 测试与本地开发

| 手段 | 说明 |
| --- | --- |
| Mock 实现 | 每个能力都有 `mock` 包，dev 配置 `provider: mock`，全链路无外部依赖即可跑通 |
| `httptest` | 适配器单测用 `httptest.Server` 模拟供应商响应（含超时、错误码、限流分支） |
| 录制回放 | 复杂响应（AI/OCR）把真实响应存 `test/fixture/thirdparty/*.json`，回放测试解析 |
| 契约测试 | 定期（每周/发版前）用真实密钥跑一次冒烟，验证供应商未改变协议 |
| 故障注入 | 单测中注入超时/500/配额超限，验证重试、熔断、降级路径正确 |
| 真机联调 | `make third-smoke` 脚本，读取 `.env` 中的密钥，逐个能力打一次真实调用并输出结果 |

### 5.10 决策速查：新第三方能力该放哪？

| 情况 | 放哪 | 例子 |
| --- | --- | --- |
| 纯协议适配，无业务语义 | `pkg/thirdparty/<capability>/<vendor>/` | 百度翻译、DeepL |
| 需要 DB、配额、审计、跨服务复用 | `app/third/`（third-rpc） | 翻译编排、AI 摘要任务 |
| 只有一个服务用，且极其简单（如某服务独有的 webhook 验签） | 该服务 `internal/` 内，但 HTTP 客户端仍用 `pkg/thirdparty/transport` | 某平台回调验签 |
| 是基础设施而非"第三方 API" | 独立 `pkg`（如 `pkg/db`、`pkg/cache`） | MySQL、Redis |
| 前端直连的第三方（如地图 JS、CDN） | `front/*/src/lib/` | 前端埋点 SDK |

---

## 6. 各服务详细设计

### 6.0 本章阅读方式

每个服务按统一模板展开：

> **职责 / 不做什么 → 目录结构 → 对外接口清单 → 数据表 → 缓存 → 依赖 → 关键流程与规则 → 错误码 → 扩展点**

接口清单里的路径指**网关对外的 REST 路径**（RPC 方法名在 proto 中定义，通常与路径一一对应）。

**REST 路径前缀约定**

| 前缀 | 归属 | 鉴权 |
| --- | --- | --- |
| `/api/v1/**` | 各域 api（归属见 [6.1 路由表](#61-接入层与前台-http-契约按域归属-api-服务)） | 部分需登录（见接口表"鉴权"列） |
| `/api/v1/admin/**` | `admin` | 必须登录 + 对应权限码 |
| `/ws/**` | `message-api` | 连接时校验 access token |
| `/api/v1/open/**` | 预留开放平台 | HMAC 签名 |
| `/metrics`、`/healthz`、`/readyz`、`/version` | 所有进程（指标端口，业务端口 + 10000） | 仅内网 |

---

### 6.1 接入层与前台 HTTP 契约（按域归属 api 服务）

**边缘层（Nginx / Ingress）职责**

1. TLS 终止、HTTP/2、gzip/br 压缩、静态资源与 CDN 回源。
2. 按路径前缀把请求路由到对应的 api 服务（路由表见下）。
3. 基础防护：IP 维度限流、请求体大小上限、慢连接防护、恶意 UA/爬虫拦截、WAF。
4. 灰度：按权重或 Header 将流量打到新版本 Deployment。
5. **不写任何业务逻辑**（"这个用户能不能看这篇私有文章"这类判断一律在 api/rpc 内完成）。

**路由表**

| 路径前缀 | 归属服务 | 端口 |
| --- | --- | --- |
| `/api/v1/auth/**`、`/api/v1/users/**` | `user-api` | 8801 |
| `/api/v1/articles/**`、`/api/v1/categories/**`、`/api/v1/tags/**`、`/api/v1/series/**`、`/api/v1/collect-folders/**` | `article-api` | 8803 |
| `/api/v1/comments/**` | `comment-api` | 8804 |
| `/api/v1/files/**` | `file-api` | 8805 |
| `/api/v1/messages/**`、`/ws/**` | `message-api` | 8806 |
| `/api/v1/search/**` | `search-api` | 8807 |
| `/api/v1/site/**` | `system-api` | 8802 |
| `/api/v1/stats/**` | `stat-api` | 8808 |
| `/api/v1/home`、`/api/v1/feed` | `portal-api` | 8810 |
| `/api/v1/admin/**` | `admin` | 8811 |
| `/api/v1/open/**`（预留） | 开放平台 | — |

> **路由约束（由域自治直接推导出的硬约束）**：**路径的第一段必须唯一归属一个域**，否则边缘层无法路由。
> - ❌ `GET /api/v1/articles/:id/comments`（`articles` 归 article 域、内容却归 comment 域，一条路由跨两域）
> - ✅ `GET /api/v1/comments?articleId=:id`（comment 域拥有）
> - 同理 ❌ `/api/v1/users/:id/articles` → ✅ `/api/v1/articles?authorId=:id`
> - **结论**：凡跨域子资源，一律"**拍平成顶层资源 + 查询参数**"，不做多级嵌套路径。

**每个域 api 的统一职责**（`pkg/middleware` 提供，各 api 复用同一份实现与配置模板）

1. JWT 校验 → 注入 metadata（`x-uid` / `x-roles` / `x-perm-ver` / `x-trace-id` / `x-client-ip`）。
2. 参数校验（`httpx.Parse` + `pkg/validatorx`）。
3. 限流（本地令牌桶 + Redis 全局限额）、幂等键、请求体大小限制。
4. 调用本域/他域 rpc，完成**聚合与字段裁剪**——这就是 go-zero 所说的"业务需求实现"。
5. 统一响应（`pkg/response`）与错误码映射（`pkg/errx`）。
6. 埋点上报（异步投 MQ，由 `stat-rpc` 消费）。

**api 服务目录**：标准结构见 [3.3 一、`api/`](#33-单个服务的标准分层以-article-域为例逐文件说明)；编排型服务（`portal`/`admin`）同构，但 `svc` 中只注入各域 rpc client。

**前台 HTTP 契约总表**

> **归属规则**：路径第一段即归属域（`auth`/`users` → `user-api`；`articles`/`categories`/`tags`/`series`/`collect-folders` → `article-api`；`comments` → `comment-api`；`files` → `file-api`；`messages`/`ws` → `message-api`；`search` → `search-api`；`site` → `system-api`；`stats` → `stat-api`；`home`/`feed` → `portal-api`）。下表就是各域 `api/api/*.api` 里应当有的契约清单。

| 模块 | 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- | --- |
| 认证 | POST | `/api/v1/auth/register` | 公开 | 邮箱/用户名注册，需图形验证码 |
| | POST | `/api/v1/auth/login` | 公开 | 账号密码登录（限流） |
| | POST | `/api/v1/auth/login/email-code` | 公开 | 邮箱验证码登录 |
| | POST | `/api/v1/auth/refresh` | 公开 | 刷新令牌（轮转） |
| | POST | `/api/v1/auth/logout` | 登录 | 撤销当前设备令牌 |
| | POST | `/api/v1/auth/logout/all` | 登录 | 撤销全部设备 |
| | GET | `/api/v1/auth/captcha` | 公开 | 图形验证码（返回 base64 + cid） |
| | POST | `/api/v1/auth/email-code` | 公开 | 发送邮箱验证码 |
| | POST | `/api/v1/auth/password/forgot` | 公开 | 找回密码 |
| | POST | `/api/v1/auth/password/reset` | 公开 | 重置密码（凭 token） |
| | GET | `/api/v1/auth/oauth/:provider/url` | 公开 | 第三方登录跳转 URL |
| | POST | `/api/v1/auth/oauth/:provider/callback` | 公开 | 第三方登录回调换令牌 |
| 用户 | GET | `/api/v1/users/me` | 登录 | 当前用户信息 + 权限 |
| | PUT | `/api/v1/users/me` | 登录 | 修改资料 |
| | PUT | `/api/v1/users/me/password` | 登录 | 改密码（踢下线） |
| | PUT | `/api/v1/users/me/avatar` | 登录 | 换头像（关联 file-rpc） |
| | GET | `/api/v1/users/:id` | 公开 | 用户主页（按隐私设置裁剪字段） |
| | GET | `/api/v1/articles?authorId=:id` | 公开 | 某人的公开文章。★ 拍平：不用 `/users/:id/articles` |
| | POST | `/api/v1/users/:id/follow` | 登录 | 关注 |
| | DELETE | `/api/v1/users/:id/follow` | 登录 | 取关 |
| | GET | `/api/v1/users/:id/followers` | 公开 | 粉丝列表 |
| | GET | `/api/v1/users/:id/following` | 公开 | 关注列表 |
| 文章 | GET | `/api/v1/articles` | 公开 | 文章列表（分类/标签/作者/关键词/排序） |
| | GET | `/api/v1/articles/:id` | 公开 | 详情（含正文、作者、分类标签、上下篇） |
| | GET | `/api/v1/articles/slug/:slug` | 公开 | 按 slug 访问（SEO 友好） |
| | POST | `/api/v1/articles` | 登录 | 创建（草稿） |
| | PUT | `/api/v1/articles/:id` | 登录 | 更新 |
| | POST | `/api/v1/articles/:id/publish` | 登录 | 提交发布/申请审核 |
| | POST | `/api/v1/articles/:id/schedule` | 登录 | 定时发布 |
| | DELETE | `/api/v1/articles/:id` | 登录 | 删除（软删） |
| | GET | `/api/v1/articles/:id/versions` | 登录 | 版本历史 |
| | POST | `/api/v1/articles/:id/versions/:v/restore` | 登录 | 回滚版本 |
| | GET | `/api/v1/articles/mine` | 登录 | 我的文章（含草稿） |
| | GET | `/api/v1/articles/hot` | 公开 | 热榜（Redis ZSet） |
| | GET | `/api/v1/articles/feed` | 登录 | 关注流（时间线） |
| | POST | `/api/v1/articles/:id/like` | 登录 | 点赞/取消 |
| | POST | `/api/v1/articles/:id/collect` | 登录 | 收藏/取消 |
| | GET | `/api/v1/articles/:id/related` | 公开 | 相关文章（标签相似度） |
| | POST | `/api/v1/articles/:id/translate` | 登录 | 触发翻译（→ third-rpc） |
| | GET | `/api/v1/articles/:id/translation` | 公开 | 获取译文 |
| 分类标签 | GET | `/api/v1/categories` | 公开 | 分类树 |
| | GET | `/api/v1/tags` | 公开 | 标签云（按引用数） |
| | GET | `/api/v1/series` | 公开 | 专栏列表 |
| | GET | `/api/v1/series/:id` | 公开 | 专栏详情 + 文章目录 |
| 评论 | GET | `/api/v1/comments?articleId=:id` | 公开 | 评论列表（按热度/时间、游标分页）。★ 拍平：不用 `/articles/:id/comments` |
| | POST | `/api/v1/comments` | 登录 | 发表评论/回复（body 带 articleId） |
| | DELETE | `/api/v1/comments/:id` | 登录 | 删除自己的评论 |
| | POST | `/api/v1/comments/:id/like` | 登录 | 评论点赞 |
| | POST | `/api/v1/comments/:id/report` | 登录 | 举报 |
| 收藏夹 | GET/POST | `/api/v1/collect-folders` | 登录 | 收藏夹管理 |
| 文件 | POST | `/api/v1/files/upload` | 登录 | 单文件上传（表单） |
| | POST | `/api/v1/files/multipart/init` | 登录 | 初始化分片上传 |
| | POST | `/api/v1/files/multipart/part` | 登录 | 上传分片 |
| | POST | `/api/v1/files/multipart/complete` | 登录 | 完成分片上传 |
| | POST | `/api/v1/files/check` | 登录 | 秒传检查（hash 命中） |
| 消息 | GET | `/api/v1/messages` | 登录 | 站内信列表 |
| | GET | `/api/v1/messages/unread-count` | 登录 | 未读数 |
| | POST | `/api/v1/messages/read` | 登录 | 批量已读 |
| | GET | `/api/v1/messages/subscribe` (WS → `/ws/messages`) | 登录 | WebSocket 实时推送（归 message 域） |
| 搜索 | GET | `/api/v1/search` | 公开 | 全文检索（高亮） |
| | GET | `/api/v1/search/suggest` | 公开 | 搜索建议 |
| | GET | `/api/v1/search/hot` | 公开 | 热搜词 |
| 站点 | GET | `/api/v1/site/config` | 公开 | 站点配置（标题、公告、开关、备案号）→ 归 system 域 |
| | GET | `/api/v1/stats/site` | 公开 | 站点统计（文章数、用户数、访问量）→ 归 stat 域 |
| 聚合 | GET | `/api/v1/home` | 公开 | 首页聚合（轮播 + 推荐文章 + 热榜 + 标签 + 公告） |

> 上面的清单是**接口设计的基线**，实施时可增补，但**新增接口必须先更新对应域的契约文件 `<域>/api/*.api`**（如新增文章接口改 `app/article/api/article.api`），再 `goctl` 生成。

**缓存**：各 api 层自身不缓存业务数据；仅缓存站点配置（60s，进程内）与后台权限映射（60s）。

**关键规则**

1. **api 层的 logic 只做三件事**：协议转换、聚合与字段裁剪、频控与幂等。**不得出现领域规则**（状态机、可见性、归属判定一律在 `rpc/domain` 里）；纯拼装代码超过 30 行要抽到 `assembler`。
2. **api 层禁止 import 本域 `model`、禁止持有 `*gorm.DB`**（见 [3.2 纪律](#32-serverapp-服务目录每个服务固定预留-api--rpc--model)）；跨域取数只能调用对方 `rpc`。
3. 所有 `POST/PUT/DELETE` 必须校验 **CSRF 风险**（Bearer Token 方案下风险低，但仍需校验 `Origin`）。
4. 请求体 ≤ 1MB（上传接口除外，走分片）。
5. 统一注入 `x-trace-id`、`x-client-ip`（从 `X-Forwarded-For` 中取**第一个非可信代理**地址，禁止直接信任整串）。
6. 灰度：边缘层按权重/Header 把流量打到新版本 Deployment，api 与 rpc 各自独立灰度。

---

### 6.2 `app/admin`（后台统一入口，只有 api）

**职责**：见 [2.7](#27-编排型服务portal--admin的边界)。核心是 **RBAC 接口级鉴权 + 操作日志 + 后台数据聚合**。

**目录**

```
server/app/admin/
├── admin.go
├── api/                                     # 后台契约（只有 api，无 rpc/model）
│   ├── admin.api                            # 主入口（prefix: /api/v1/admin）
│   ├── user.admin.api / article.admin.api / comment.admin.api
│   ├── system.admin.api / stat.admin.api / job.admin.api
│   └── third.admin.api
└── internal/
    ├── config/config.go                     # Rest + Jwt + Redis + 各域 RpcClientConf
    ├── svc/service_context.go               # ★ 只注入各域 rpc client（UserRpc/ArticleRpc/SystemRpc…）
    ├── handler/                             # 【生成】
    ├── logic/                               # 编排 + 聚合
    ├── types/                               # 【生成】
    ├── middleware/
    │   ├── authz.go                         # 路由 → 权限码校验（映射表来自 system-rpc，本地缓存 60s）
    │   └── oplog.go                         # 操作日志采集（异步批量落库）
    └── consts/
```

**核心接口清单（后台，前缀 `/api/v1/admin`）**

| 模块 | 方法 | 路径 | 权限码 |
| --- | --- | --- | --- |
| 登录 | POST | `/auth/login` | 公开（后台独立登录，校验后台角色） |
| | GET | `/me` | 登录 |
| | GET | `/me/menus` | 登录 | 返回该用户可见菜单树（前端动态路由） |
| 用户 | GET | `/users` | `user:list` |
| | GET | `/users/:id` | `user:detail` |
| | PUT | `/users/:id/status` | `user:ban` |
| | PUT | `/users/:id/roles` | `user:assign-role` |
| | POST | `/users/:id/reset-password` | `user:reset-pwd` |
| | GET | `/users/:id/login-logs` | `user:log` |
| 文章 | GET | `/articles` | `article:list` |
| | GET | `/articles/:id` | `article:detail` |
| | PUT | `/articles/:id/review` | `article:review` |
| | PUT | `/articles/:id/top` | `article:top` |
| | PUT | `/articles/:id/recommend` | `article:recommend` |
| | DELETE | `/articles/:id` | `article:delete` |
| | PUT | `/articles/batch/status` | `article:review`（批量审核） |
| 分类标签 | GET/POST/PUT/DELETE | `/categories`、`/tags`、`/series` | `category:*` / `tag:*` |
| 评论 | GET | `/comments` | `comment:list` |
| | PUT | `/comments/:id/status` | `comment:audit` |
| | DELETE | `/comments/:id` | `comment:delete` |
| | GET | `/comment-reports` | `comment:report:list` |
| | PUT | `/comment-reports/:id/handle` | `comment:report:handle` |
| 敏感词 | GET/POST/PUT/DELETE | `/sensitive-words` | `sensitive:*` |
| | POST | `/sensitive-words/import` | `sensitive:import` |
| 文件 | GET | `/files` | `file:list` |
| | DELETE | `/files/:id` | `file:delete` |
| | GET | `/files/quota` | `file:quota` |
| 消息公告 | GET/POST/PUT/DELETE | `/announcements` | `announcement:*` |
| | POST | `/messages/broadcast` | `message:broadcast`（异步群发） |
| | GET | `/msg-templates` | `message:template` |
| 系统 | GET/POST/PUT/DELETE | `/roles`、`/permissions`、`/menus` | `system:role` / `system:perm` |
| | GET/PUT | `/configs` | `system:config` |
| | GET/POST/PUT/DELETE | `/dicts`、`/dict-items` | `system:dict` |
| | GET | `/operation-logs` | `system:oplog` |
| | GET | `/jobs`、`POST /jobs/:id/run` | `system:job` |
| 统计 | GET | `/dashboard/overview` | `stat:overview` |
| | GET | `/dashboard/trend` | `stat:trend` |
| | GET | `/dashboard/top-articles` | `stat:top` |
| 第三方 | GET | `/third/logs` | `third:log` |
| | GET/PUT | `/third/quotas` | `third:quota` |

**RBAC 鉴权实现**

```go
// middleware/authz.go 核心逻辑
// 1. 启动时/每 60s 从 system-rpc 拉取 路由→权限码 映射表（path + method → permCode）
// 2. 支持通配：DELETE /api/v1/admin/articles/* → article:delete
// 3. 当前用户权限集合来自 JWT roles → system-rpc 展开（本地缓存 60s，Key 含 permVer）
// 4. 命中即放行；未命中返回 403 + CodeNoPermission
// 5. 映射表未覆盖的路径 → 默认拒绝（default deny），并打 WARN 日志提示补配置
```

**操作日志**：非 GET 请求一律记录，字段见 [7.6](#76-系统表noviterra_system)。实现为**异步**（channel + 批量写，或投递 MQ），失败不影响主流程但必须打点告警。

---

### 6.3 `app/user`（账号与用户中心）

**职责**：账号域由 `user-api`（终端入口）与 `user-rpc`（基础数据）共同构成。

| 入口 | 职责 |
| --- | --- |
| `user-api` | HTTP 契约、图形/邮箱验证码前置校验、登录频控与幂等、资料与关注接口的**字段裁剪与装配**、后台用户管理接口的编排 |
| `user-rpc` | 认证规则（密码校验、令牌签发与轮转、踢下线）、用户资料与设置、关注粉丝关系、用户状态与登录日志、"批量取用户摘要"供他域调用 |

**不做什么**：不发文章、不存文件、不直接发通知（发事件由 message 消费）、不管理角色权限定义（归 `system` 域，本域只持有 `user_role` 关联）。

**目录**

```
server/app/user/
├── api/                                 # ── user-api（HTTP 8801）
│   ├── user.go
│   ├── user.api
│   ├── etc/user-api.yaml
│   └── internal/
│       ├── config/config.go              # Rest + Jwt + Redis + UserRpc/FileRpc…
│       ├── svc/service_context.go        # ★ 只有 rpc client，无 *gorm.DB
│       ├── handler/                      # 【生成】
│       ├── logic/                        # registerlogic.go / loginlogic.go / refreshtokenlogic.go / …
│       ├── types/                        # 【生成】
│       └── middleware/ consts/
├── rpc/                                 # ── user-rpc（gRPC 9001）
│   ├── user.go
│   ├── user.proto
│   ├── etc/user-rpc.yaml
│   ├── pb/  user/                        # 【生成】禁止手改
│   └── internal/
│       ├── config/config.go              # ZRpc + DB + Redis + MQ + 他域 rpc client
│       ├── svc/service_context.go        # ★ 这里注入 DB、Redis、Jwt Manager、各 Repo、事件发布器
│       ├── server/user_server.go         # 【生成】
│       ├── logic/                        # 领域用例（清单见下）
│       ├── domain/                       # 密码策略、令牌规则、关注关系规则、用户状态机
│       ├── event/                        # user.registered / user.login / user.banned / user.followed / user.updated
│       ├── middleware/ dto/ vo/ assembler/ consts/
└── model/                               # ── 共享数据层（本域只有 rpc 使用）
    ├── entity/                           # user、user_auth、user_profile、user_setting、user_stat、user_follow、user_role、user_login_log
    ├── repository/                       # UserRepo / AuthRepo / FollowRepo / LoginLogRepo
    └── cache/                            # 用户资料缓存、关注关系缓存、验证码
```

`rpc/internal/logic/` 清单（一个 RPC 方法一个文件）：
`registerlogic.go`、`loginlogic.go`、`loginbycodelogic.go`、`refreshtokenlogic.go`、`logoutlogic.go`、`getcaptchalogic.go`、`sendemailcodelogic.go`、`resetpasswordlogic.go`、`getuserlogic.go`、`getuserbatchlogic.go`（★ 批量，供他域调用）、`updateprofilelogic.go`、`updateavatarLogic.go`、`updatepasswordlogic.go`、`followlogic.go`、`unfollowlogic.go`、`listfollowerslogic.go`、`listfollowinglogic.go`、`oauthloginlogic.go`、`adminlistuserslogic.go`、`adminupdatestatuslogic.go`、`assignroleslogic.go`、`getuserroleslogic.go`

**RPC 接口（proto 摘要）**

```protobuf
service User {
  // 认证
  rpc Register(RegisterReq) returns (LoginResp);
  rpc Login(LoginReq) returns (LoginResp);
  rpc LoginByEmailCode(LoginByEmailCodeReq) returns (LoginResp);
  rpc RefreshToken(RefreshTokenReq) returns (LoginResp);
  rpc Logout(LogoutReq) returns (Empty);
  rpc LogoutAll(LogoutAllReq) returns (Empty);
  rpc SendEmailCode(SendEmailCodeReq) returns (Empty);
  rpc GetCaptcha(GetCaptchaReq) returns (GetCaptchaResp);
  rpc ResetPassword(ResetPasswordReq) returns (Empty);
  rpc OAuthLogin(OAuthLoginReq) returns (LoginResp);
  // 资料
  rpc GetUser(GetUserReq) returns (UserInfo);
  rpc GetUserBatch(GetUserBatchReq) returns (GetUserBatchResp);   // ★ 必须批量
  rpc GetUserByUsername(GetUserByUsernameReq) returns (UserInfo);
  rpc UpdateProfile(UpdateProfileReq) returns (UserInfo);
  rpc UpdateAvatar(UpdateAvatarReq) returns (Empty);
  rpc UpdatePassword(UpdatePasswordReq) returns (Empty);
  rpc UpdateSetting(UpdateSettingReq) returns (Empty);
  rpc GetSetting(GetSettingReq) returns (UserSetting);
  // 关系
  rpc Follow(FollowReq) returns (Empty);
  rpc Unfollow(UnfollowReq) returns (Empty);
  rpc ListFollowers(ListFollowReq) returns (ListFollowResp);
  rpc ListFollowing(ListFollowReq) returns (ListFollowResp);
  rpc CheckFollow(CheckFollowReq) returns (CheckFollowResp);      // 批量判断关注状态
  // 后台
  rpc AdminListUsers(AdminListUsersReq) returns (AdminListUsersResp);
  rpc AdminGetUser(AdminGetUserReq) returns (AdminUserDetail);
  rpc AdminUpdateStatus(AdminUpdateStatusReq) returns (Empty);
  rpc AdminResetPassword(AdminResetPasswordReq) returns (Empty);
  rpc AdminAssignRoles(AdminAssignRolesReq) returns (Empty);
  rpc AdminListLoginLogs(AdminListLoginLogsReq) returns (AdminListLoginLogsResp);
  rpc GetUserRoles(GetUserRolesReq) returns (GetUserRolesResp);   // admin 鉴权用
  // 内部
  rpc GetUserStat(GetUserStatReq) returns (UserStat);             // 计数由事件驱动更新
}
```

**注册流程（务必按序）**

```
1. 校验图形验证码（Redis GETDEL，一次性）
2. 校验参数（用户名唯一、邮箱格式、密码强度、邮箱验证码有效）
3. 事务：
   a. INSERT user（status=active, email_verified=0/1）
   b. INSERT user_auth（identity_type=email|username, credential=bcrypt(pwd)）
   c. INSERT user_profile / user_setting（默认值）
   d. INSERT user_stat（全 0）
   e. INSERT user_role（默认角色：reader，角色 ID 由 system-rpc 提供或固定内置）
   f. INSERT outbox（事件 user.registered）
4. 事务提交后：投递事件 → message-rpc 发欢迎邮件；stat-rpc 更新注册数
5. 直接返回登录态（access + refresh）
```

**用户名/邮箱唯一性**：靠**数据库唯一索引**兜底（`uk_username`、`uk_email`），应用层查重只为给出友好提示；捕获 `1062` 错误转为 `CodeUsernameExists/CodeEmailExists`。

**登录安全规则**

| 规则 | 值 |
| --- | --- |
| 密码哈希 | bcrypt cost 12（校验成功后按需 rehash） |
| 失败锁定 | 同账号 5 次失败 → 锁 15 分钟（Redis）；同 IP 10 次/分钟 → 429 |
| 验证码 | 连续 3 次失败后强制图形验证码 |
| 异地登录 | 与上次登录 IP 城市不同 → 发提醒邮件（可关） |
| 令牌 | 见 [4.2](#42-pkgjwt--令牌与会话) |
| 密码修改/重置 | 撤销全部 refresh token + `PermVer++` + 通知邮件 |
| 注销账号 | 状态置 `deleted`，保留 30 天可恢复，之后匿名化（保留文章但作者显示"已注销用户"） |

**缓存设计**

| Key | 类型 | TTL | 用途 |
| --- | --- | --- | --- |
| `nv:{env}:user:info:{uid}` | String(JSON) | 10min | 用户资料 |
| `nv:{env}:user:info:{username}` | String | 10min | 用户名 → uid |
| `nv:{env}:user:refresh:{uid}:{device}` | String | 30d | refresh token 哈希 |
| `nv:{env}:user:blacklist:{jti}` | String | 剩余 TTL | 已撤销 access token |
| `nv:{env}:user:captcha:{cid}` | String | 5min | 图形验证码答案 |
| `nv:{env}:user:emailcode:{scene}:{email}` | String | 10min | 邮箱验证码 |
| `nv:{env}:user:login:fail:{account}` | String(计数) | 15min | 登录失败次数 |
| `nv:{env}:user:follow:following:{uid}` | Set | 1h | 关注列表 |
| `nv:{env}:user:follow:followers:{uid}` | Set | 1h | 粉丝列表 |
| `nv:{env}:user:stat:{uid}` | Hash | 1h | 文章数/粉丝数等冗余计数 |

**事件**

| 事件 | 载荷 | 消费者 |
| --- | --- | --- |
| `user.registered.v1` | uid, username, email | message-rpc（欢迎信）、stat-rpc |
| `user.login.v1` | uid, ip, ua, device | stat-rpc（活跃）、system-rpc（登录日志归档） |
| `user.banned.v1` | uid, reason, until | article-rpc（隐藏其内容）、comment-rpc、message-rpc |
| `user.followed.v1` | uid, targetUid | message-rpc（关注通知）、stat-rpc |
| `user.updated.v1` | uid, changedFields | article-rpc / comment-rpc（刷新作者冗余字段） |

**扩展点**：多因素认证（TOTP）、登录设备管理（列表 + 远程下线）、用户等级/积分、用户主页自定义主题。

---

### 6.4 `app/system`（后台系统管理域）

**职责**

1. **RBAC 模型**：角色、权限（菜单/按钮/接口）、用户-角色、角色-权限。
2. **系统配置**：`sys_config`（站点名、备案号、开关、上传限制、审核策略）。
3. **数据字典**：`sys_dict` / `sys_dict_item`（前端下拉与状态映射）。
4. **公告**：`sys_announcement`。
5. **审计**：操作日志、接口调用日志查询。
6. **定时任务元数据**：任务定义、启停、执行历史（实际执行在 `app/job`）。

**不做什么**：不管理用户资料（归 `user` 域）、不管理内容（归 `article` 域）。

**目录**：标准三目录结构（见 [3.2](#32-serverapp-服务目录每个服务固定预留-api--rpc--model)）：

- `app/system/api/`：`system-api`（8802）——后台 RBAC/字典/配置/公告/审计/任务管理界面接口，以及**前台公开的站点配置** `GET /api/v1/site/config`。
- `app/system/rpc/`：`system-rpc`（9002）——权限码展开（供 `admin` 鉴权）、用户菜单、公开配置、操作日志写入。
- `app/system/model/entity/`：`role`、`permission`、`user_role`、`role_permission`、`sys_config`、`sys_dict`、`sys_dict_item`、`sys_announcement`、`sys_operation_log`、`sys_job`、`sys_job_log`。

**RPC 接口（摘要）**

```protobuf
service System {
  // RBAC
  rpc ListRoles(ListRolesReq) returns (ListRolesResp);
  rpc CreateRole / UpdateRole / DeleteRole
  rpc ListPermissions(ListPermissionsReq) returns (ListPermissionsResp);  // 树形
  rpc CreatePermission / UpdatePermission / DeletePermission
  rpc GetRolePermissions(GetRolePermissionsReq) returns (IdsResp);
  rpc AssignRolePermissions(AssignRolePermissionsReq) returns (Empty);
  rpc GetUserPermissions(GetUserPermissionsReq) returns (GetUserPermissionsResp); // ★ admin 鉴权核心
  rpc GetUserMenus(GetUserMenusReq) returns (GetUserMenusResp);                   // 前端动态路由
  // 配置与字典
  rpc GetConfigs(GetConfigsReq) returns (GetConfigsResp);
  rpc SetConfig(SetConfigReq) returns (Empty);
  rpc GetPublicConfigs(GetPublicConfigsReq) returns (GetPublicConfigsResp); // 前台站点配置
  rpc ListDicts / ListDictItems / CreateDict / CreateDictItem / UpdateDictItem / DeleteDictItem
  // 公告与审计
  rpc ListAnnouncements / CreateAnnouncement / UpdateAnnouncement / DeleteAnnouncement
  rpc WriteOperationLog(WriteOperationLogReq) returns (Empty);   // 异步批量
  rpc ListOperationLogs(ListOperationLogsReq) returns (ListOperationLogsResp);
  // 任务元数据
  rpc ListJobs / UpdateJob / ListJobLogs / TriggerJob
}
```

**权限模型（三层）**

| 类型 | 说明 | 示例 |
| --- | --- | --- |
| `menu` | 菜单/页面可见性 | `article:list`（对应路由 `/article/list`） |
| `button` | 页面内操作按钮 | `article:delete` |
| `api` | 接口级（兜底，细粒度） | `article:review` |

**权限判定**：`admin` 收到 `GetUserPermissions(uid)` → 返回 `{permCodes: [], roles: [], permVer: n}`，本地缓存 60s（缓存 Key 带 `permVer`，权限变更时 `PermVer++` 自动失效）。

**内置角色（seed 数据）**

| 角色码 | 名称 | 权限范围 |
| --- | --- | --- |
| `super_admin` | 超级管理员 | 全部（含权限配置） |
| `admin` | 管理员 | 除权限配置外全部 |
| `editor` | 内容编辑 | 文章审核、分类标签、评论审核 |
| `operator` | 运营 | 用户查看、公告、消息群发、统计 |
| `auditor` | 审核员 | 仅文章/评论审核 |
| `reader` | 普通读者 | 前台权限（无后台菜单） |
| `author` | 作者 | 前台发文权限（可能需要实名/申请） |

**系统配置分组**：`site`（站点信息）、`security`（密码策略、登录限制）、`content`（审核开关、敏感词策略、默认封面）、`upload`（大小、类型、存储）、`third`（第三方开关与配额）、`mail`（SMTP 开关）、`feature`（功能开关：评论、搜索、翻译、AI）。

---

### 6.5 `app/article`（内容中心，项目核心）

**职责**

| 入口 | 职责 |
| --- | --- |
| `article-api` | 文章读写 HTTP 契约、**详情页/列表页的聚合与裁剪**（并发取 article + comment 数 + 互动状态）、编辑器草稿自动保存接口、翻译与摘要的触发入口 |
| `article-rpc` | 文章全生命周期、版本历史、分类/标签/专栏、点赞收藏、阅读计数与热度、翻译结果与 AI 摘要的落库、"批量文章摘要"供他域调用 |

`article-rpc` 的领域职责细目：

1. 文章全生命周期：草稿 → 提交审核 → 发布 / 驳回 / 定时发布 / 归档 / 回收站。
2. 版本历史与回滚（每次更新落版本）。
3. 分类（树形）、标签、专栏/系列。
4. 点赞、收藏、收藏夹。
5. 阅读计数（异步聚合）、热度计算。
6. 翻译结果与 AI 摘要的存储（调 `third` 域取结果，**本域落库**）。
7. 相关文章推荐（基于标签/分类相似度，MVP 用 SQL/Redis，规模化后交 `search` 域）。
8. **对外提供批量文章摘要接口**，供 `comment`/`stat`/`search`/`portal` 使用。

**不做什么**：不管评论树、不做索引（发事件给 `search` 域）、不存文件（只存 file_id 引用）、不负责通知、不做"页面要什么字段"的裁剪（那是 `article-api` 的事）。

**目录**（标准三目录，重点看 `rpc/domain` 与 `rpc/event`）

```
server/app/article/
├── api/                              # ── article-api（HTTP 8803）：面向终端的业务实现
│   ├── article.go                    # 【生成】入口
│   ├── article.api                   # 【手写】HTTP 契约（含分类/标签/专栏/收藏夹）
│   ├── etc/article-api.yaml
│   └── internal/
│       ├── svc/service_context.go    # ★ 只有 rpc client：ArticleRpc/CommentRpc/UserRpc/ThirdRpc
│       ├── handler/ logic/ types/    # logic 负责聚合与裁剪（详情页并发聚合是本层核心）
│       └── middleware/ consts/
│
├── rpc/                              # ── article-rpc（gRPC 9003）：领域与数据
│   ├── article.go  article.proto  etc/article-rpc.yaml  pb/  article/
│   └── internal/
│       ├── svc/service_context.go    # ★ DB、Redis、UserRpc、FileRpc、ThirdRpc、事件发布器
│       ├── server/article_server.go  # 【生成】
│       ├── logic/                    # 一个 RPC 方法一个文件（约 40+ 个）
│       ├── domain/                   # ★ 领域规则唯一来源
│       │   ├── article_status.go     #   状态机：CanPublish/CanReview/CanDelete
│       │   ├── article_visibility.go #   可见性判定（public/private/password/fans）
│       │   ├── slug.go               #   slug 生成与唯一性
│       │   ├── heat.go               #   热度算法：view*1 + like*5 + collect*8 + comment*10，按时间衰减
│       │   └── permission.go         #   谁能改这篇文章（作者/管理员/协作作者）
│       ├── event/
│       │   ├── topics.go             #   主题常量
│       │   ├── publisher.go          #   通过 outbox 发事件
│       │   └── consumer_userupdated.go  # 消费 user.updated 刷新作者冗余字段
│       ├── assembler/ middleware/ dto/ vo/ consts/{consts.go,enums.go,errors.go}
│       └── …
│
└── model/                            # ── 共享数据层（本域只有 rpc 使用）
    ├── entity/                       #   见下方表清单
    ├── repository/                   #   ArticleRepo / CategoryRepo / TagRepo / SeriesRepo / OutboxRepo…
    └── cache/                        #   keys.go + 各实体缓存读写
```

**RPC 接口（proto 摘要）**

```protobuf
service Article {
  // 写
  rpc CreateArticle(CreateArticleReq) returns (CreateArticleResp);
  rpc UpdateArticle(UpdateArticleReq) returns (Empty);
  rpc PublishArticle(PublishArticleReq) returns (Empty);
  rpc ScheduleArticle(ScheduleArticleReq) returns (Empty);
  rpc DeleteArticle(DeleteArticleReq) returns (Empty);
  rpc RestoreArticle(RestoreArticleReq) returns (Empty);
  rpc UpdateArticleStatus(UpdateArticleStatusReq) returns (Empty);   // 审核用
  rpc SaveDraft(SaveDraftReq) returns (Empty);
  // 版本
  rpc ListVersions(ListVersionsReq) returns (ListVersionsResp);
  rpc GetVersion(GetVersionReq) returns (ArticleVersion);
  rpc RestoreVersion(RestoreVersionReq) returns (Empty);
  // 读
  rpc GetArticle(GetArticleReq) returns (ArticleDetail);            // id 或 slug
  rpc BatchGetArticleSummary(BatchGetArticleSummaryReq) returns (BatchGetArticleSummaryResp); // ★
  rpc ListArticles(ListArticlesReq) returns (ListArticlesResp);     // 通用列表（多条件）
  rpc ListMyArticles(ListMyArticlesReq) returns (ListArticlesResp);
  rpc ListHotArticles(ListHotArticlesReq) returns (ListArticlesResp);
  rpc ListRelatedArticles(ListRelatedArticlesReq) returns (ListArticlesResp);
  rpc ListFeed(ListFeedReq) returns (ListArticlesResp);             // 关注流
  rpc ListPendingArticles(ListPendingArticlesReq) returns (ListArticlesResp); // 后台审核
  // 互动
  rpc LikeArticle(LikeArticleReq) returns (Empty);                  // 幂等：已赞则取消
  rpc CollectArticle(CollectArticleReq) returns (Empty);
  rpc ListCollectedArticles(ListCollectedArticlesReq) returns (ListArticlesResp);
  rpc ListCollectFolders / CreateCollectFolder / RenameCollectFolder / DeleteCollectFolder
  rpc CheckInteraction(CheckInteractionReq) returns (CheckInteractionResp);  // 批量：是否已赞/已藏
  // 分类标签专栏
  rpc ListCategories(ListCategoriesReq) returns (ListCategoriesResp);        // 树
  rpc CreateCategory / UpdateCategory / DeleteCategory / SortCategories
  rpc ListTags / CreateTag / UpdateTag / DeleteTag / ListHotTags
  rpc ListSeries / GetSeries / CreateSeries / UpdateSeries / DeleteSeries / AddArticlesToSeries / SortSeriesArticles
  // 翻译与摘要
  rpc TranslateArticle(TranslateArticleReq) returns (TranslateArticleResp);  // 内部调 third-rpc
  rpc GetTranslation(GetTranslationReq) returns (Translation);
  rpc GenerateSummary(GenerateSummaryReq) returns (Empty);                   // 异步任务触发
  // 统计（内部）
  rpc IncrView(IncrViewReq) returns (Empty);
  rpc GetArticleStat(GetArticleStatReq) returns (ArticleStat);
  rpc HandleUserBanned(HandleUserBannedReq) returns (Empty);                 // 消费事件
}
```

**文章状态机（必须严格实现，禁止散落 if-else）**

```
                 ┌──────────┐
   create ──────►│  draft   │◄──────── 撤回（编辑可将 published 改回 draft 并留版本）
                 └────┬─────┘
        submit        │        delete
        ┌─────────────┼──────────────┐
        ▼             │              ▼
  ┌───────────┐       │        ┌──────────┐
  │ reviewing │       │        │ deleted  │(软删，30 天后物理清理)
  └─────┬─────┘       │        └──────────┘
   approve│ reject     │
   ┌──────┴──────┐    │
   ▼             ▼    │
┌───────────┐ ┌──────────┐
│ published │ │ rejected │─── 修改后可重新 submit
└─────┬─────┘ └──────────┘
      │ archive / 定时任务
      ▼
┌───────────┐
│ archived  │（归档：不可编辑，仅管理员可恢复）
└───────────┘

定时发布：status=scheduled + scheduled_at → app/job 每分钟扫描到期 → 转 published
（用 DB 行锁 or 乐观锁 version 字段防止重复发布）
```

**状态流转规则表（`domain/article_status.go` 的单一来源）**

| 当前状态 | 允许动作 | 目标状态 | 谁可以 |
| --- | --- | --- | --- |
| draft | submit | reviewing 或 published（免审时直达） | 作者 |
| draft | delete | deleted | 作者/管理员 |
| reviewing | approve | published | 有 `article:review` 权限 |
| reviewing | reject | rejected | 有 `article:review` 权限 |
| reviewing | withdraw | draft | 作者 |
| published | update | published（内容更新，留版本） | 作者 |
| published | unpublish | draft | 作者 |
| published | archive | archived | 管理员 |
| published | delete | deleted | 作者/管理员 |
| rejected | submit | reviewing | 作者 |
| archived | restore | published | 管理员 |
| deleted | restore | draft | 管理员 |

**免审开关**：`sys_config.content.auto_review`（作者信用分足够/管理员/受信任作者可免审直达 published）。

**数据表（`noviterra_content` 库）**

| 表 | 关键字段 | 索引要点 |
| --- | --- | --- |
| `article` | id, user_id, title, slug, summary, cover, category_id, status, visibility, password, is_top, is_recommend, is_original, source_url, allow_comment, view_count, like_count, collect_count, comment_count, word_count, reading_minutes, seo_title, seo_keywords, seo_description, published_at, scheduled_at, review_at, reviewer_id, reject_reason, created_at, updated_at, deleted_at | `idx_user_status(user_id,status)`、`uk_slug(slug,deleted_at)`、`idx_status_published(status,published_at)`、`idx_category(category_id,status)`、`idx_top(top, published_at)`、`idx_scheduled(status,scheduled_at)` |
| `article_content` | article_id(PK), content_md(LONGTEXT), content_html(LONGTEXT), toc(JSON) | 主键即外键；与主表 1:1，避免列表查询拖大字段 |
| `article_version` | id, article_id, version, title, summary, content_md, editor_id, change_note, created_at | `idx_article_version(article_id,version desc)` |
| `article_tag` | article_id, tag_id, created_at | 联合主键 `(article_id,tag_id)`，`idx_tag(tag_id)` |
| `category` | id, parent_id, name, slug, description, cover, icon, sort, status, article_count, path, level | `uk_slug`、`idx_parent(parent_id,sort)`、`idx_path(path)` |
| `tag` | id, name, slug, description, color, article_count, ref_count, is_hot, status | `uk_name`、`idx_hot(is_hot,article_count desc)` |
| `series` | id, user_id, name, slug, cover, description, sort, article_count, status | `uk_slug`、`idx_user(user_id)` |
| `series_article` | series_id, article_id, sort | 联合主键，`idx_series_sort(series_id,sort)` |
| `article_like` | id, article_id, user_id, created_at | `uk_article_user(article_id,user_id)`、`idx_user(user_id,created_at desc)` |
| `article_collect` | id, user_id, article_id, folder_id, created_at | `uk_user_article`、`idx_folder(folder_id,created_at desc)` |
| `collect_folder` | id, user_id, name, is_public, sort, article_count | `uk_user_name(user_id,name)` |
| `article_translation` | id, article_id, lang, title, summary, content_md, provider, status(pending/done/degraded/failed), token_used, created_by, created_at, updated_at | `uk_article_lang(article_id,lang)` |
| `article_stat_daily` | id, article_id, stat_date, pv, uv, like_add, collect_add, comment_add, share_add | `uk_article_date(article_id,stat_date)`（也可归 stat 库，见 7.7 说明） |
| `article_view_log` | id, article_id, user_id, ip, ua, referer, stay_ms, created_at | 按月分表 `article_view_log_202501`；不做软删 |
| `outbox` | id, event_id, topic, payload, status(pending/sent/failed), retry, next_retry_at, created_at | `idx_status_next(status,next_retry_at)` |

**缓存设计**

| Key | 类型 | TTL | 说明 |
| --- | --- | --- | --- |
| `nv:{env}:article:detail:{id}` | String(JSON) | 10min±60s | 详情（不含正文大字段，正文单独 Key） |
| `nv:{env}:article:content:{id}` | String | 30min | 正文 HTML |
| `nv:{env}:article:slug:{slug}` | String | 1h | slug → id |
| `nv:{env}:article:hot:zset` | ZSet | 1h | 热榜（score = 热度） |
| `nv:{env}:article:like:{id}` | Set | — | 点赞用户集合（可选，用于判重） |
| `nv:{env}:article:view:buf` | Hash | 5min | 浏览计数缓冲（field=articleId, value=增量） |
| `nv:{env}:article:list:{queryHash}` | String | 3min | 列表首页缓存（只缓存前 3 页） |
| `nv:{env}:article:related:{id}` | String | 1h | 相关文章 ID 列表 |
| `nv:{env}:article:translation:{id}:{lang}` | String | 7d | 译文 |
| `nv:{env}:article:seq:{categoryId}` | String | — | 分类下文章序号（可选） |

**浏览量处理（防写热点）**

```
请求详情 → pipeline: HINCRBY nv:...:article:view:buf {id} 1   （不阻塞响应）
app/job 每 5 分钟：
  1. 用 Lua 原子读取并删除该 Hash（HGETALL + DEL）
  2. 批量 UPDATE article SET view_count = view_count + ? WHERE id IN (...)
  3. 更新 article_stat_daily、Redis 计数缓存
优势：把 N 次 UPDATE 合并为 1 批次；Redis 故障时降级为直接写库（有开关）
```

**热度算法（`domain/heat.go`）**

```
score = (view*1 + like*5 + collect*8 + comment*10 + share*3)
        / POWER(HOURS_SINCE_PUBLISH + 2, 1.5)     # 时间衰减（牛顿冷却）
每 10 分钟由 app/job 重算 Top 500 写入 ZSet；超过 30 天的文章不参与热榜
```

**关键业务规则**

| 规则 | 说明 |
| --- | --- |
| slug 唯一 | 标题转 slug（英文/拼音 + 短 hash 后缀），冲突时追加 `-2`、`-3`；**历史 slug 保留重定向表**（`article_slug_history`，可选） |
| 文章长度 | 正文 ≤ 100,000 字符；标题 ≤ 200；摘要 ≤ 500 |
| 分类 | 每篇 1 个主分类（可空）；标签 ≤ 8 个，自动创建不存在的标签 |
| 字数统计 | 中文按字符、英文按单词；`reading_minutes = ceil(words/400)` |
| 更新频率限制 | 同一文章 1 分钟内最多改 5 次（Redis 计数） |
| 草稿自动保存 | 前端 30s 一次 `SaveDraft`（只更新 content_md，不产生版本） |
| 版本策略 | `PublishArticle` 与显式 `update with version=true` 才生成版本；普通编辑不生成（防版本爆炸） |
| 可见性 | `private` 仅作者可见；`password` 需校验访问密码（Redis 记录已通过 1h）；`fans` 仅粉丝可见（调 user-rpc 判定） |
| 删除 | 软删；30 天后 app/job 物理删除（含正文、翻译、版本） |
| 点赞幂等 | `INSERT ... ON DUPLICATE KEY` / 唯一索引冲突即视为取消赞（用 Redis 缓存状态 + 异步落库） |
| 计数一致性 | `article.like_count` 由事件/异步任务校准；每日 03:00 用 `COUNT(*)` 对账修正偏差 > 1% 的记录 |
| 作者信息 | 文章表冗余 `author_name`、`author_avatar`，由 `user.updated` 事件刷新（最终一致，列表页不调 user-rpc） |
| 浏览去重 | 同一 uid+文章 5 分钟内只计 1 次 PV；UV 用 Redis HLL 按天统计 |

**事件**

| 事件 | 载荷 | 消费者 |
| --- | --- | --- |
| `article.created.v1` | articleId, uid, status | stat-rpc |
| `article.published.v1` | articleId, uid, title, summary, tags, categoryId, publishedAt | search-rpc（索引）、stat-rpc、message-rpc（通知粉丝）、third-rpc（可选 AI 摘要） |
| `article.updated.v1` | articleId, changedFields | search-rpc（增量更新索引） |
| `article.deleted.v1` | articleId | search-rpc（删索引）、comment-rpc（级联隐藏评论）、file-rpc（释放附件引用） |
| `article.liked.v1` | articleId, uid, targetUid | message-rpc、stat-rpc |
| `article.commented.v1` | 由 comment-rpc 发出，article-rpc 消费 | 更新 comment_count |
| `article.translated.v1` | articleId, lang | 通知作者 |

---

### 6.6 `app/comment`（互动中心）

**职责**

1. 评论、回复（支持两级：根评论 + 回复；二级以上用"回复某人"扁平化，避免无限嵌套）。
2. 评论审核（先审后发 / 先发后审，由配置决定）、敏感词过滤、频率限制。
3. 评论点赞、举报、举报处理。
4. 敏感词库维护（本地 DFA + 可选云审核）。
5. 评论计数（发事件让 article-rpc 更新 `comment_count`）。

**不做什么**：不维护文章数据（只存 article_id）、不通知用户（发事件）。

**目录**（标准三目录）：`app/comment/api/`（`comment-api` 8804：评论列表/发表/删除、点赞、举报接口，路径统一为 `/api/v1/comments**`，用 `articleId` 查询参数关联文章）、`app/comment/rpc/`（`comment-rpc` 9004：评论树、审核、敏感词 DFA、举报处理）、`app/comment/model/`（`comment`、`comment_like`、`comment_report`、`sensitive_word`）。

**评论模型（关键决策）**

```
采用「两层结构 + 楼中楼扁平化」：
  root_id    根评论 ID（根评论时等于自身 id）
  parent_id  直接父评论 ID
  reply_to_user_id  被回复用户（用于 @提醒 展示）

列表查询：按 root_id 聚合，根评论分页，回复按时间正序，每页附带前 3 条回复 + "查看全部 N 条"
优点：查询简单（一次 WHERE root_id IN (...)），不会无限递归；缺点：楼层感弱于无限嵌套
```

**RPC 接口（摘要）**

```protobuf
service Comment {
  rpc CreateComment(CreateCommentReq) returns (CreateCommentResp);
  rpc DeleteComment(DeleteCommentReq) returns (Empty);              // 作者/管理员/评论者
  rpc ListComments(ListCommentsReq) returns (ListCommentsResp);     // 根评论分页
  rpc ListReplies(ListRepliesReq) returns (ListRepliesResp);        // 某根评论下的回复
  rpc ListUserComments(ListUserCommentsReq) returns (ListCommentsResp); // 某用户的评论
  rpc LikeComment(LikeCommentReq) returns (Empty);
  rpc ReportComment(ReportCommentReq) returns (Empty);
  // 后台
  rpc AdminListComments(AdminListCommentsReq) returns (ListCommentsResp);
  rpc AdminUpdateStatus(AdminUpdateStatusReq) returns (Empty);
  rpc AdminListReports(AdminListReportsReq) returns (ListReportsResp);
  rpc AdminHandleReport(AdminHandleReportReq) returns (Empty);
  rpc AdminListSensitiveWords / CreateSensitiveWord / UpdateSensitiveWord / DeleteSensitiveWord / ImportSensitiveWords
  rpc ModerateText(ModerateTextReq) returns (ModerateTextResp);     // ★ 供其它服务复用（昵称/简介）
  // 内部
  rpc BatchGetCommentCount(BatchGetCommentCountReq) returns (BatchGetCommentCountResp);
  rpc DeleteByArticle(DeleteByArticleReq) returns (Empty);          // 文章删除联动
  rpc HandleUserBanned(HandleUserBannedReq) returns (Empty);
}
```

**发表评论校验链（顺序不可乱）**

```
1. 频率限制（20 次/10 分钟/uid；新用户 5 次/10 分钟）
2. 参数校验（长度 1~1000；内容非空白；article_id 合法性）
3. 文章可评论（allow_comment=true，且文章 published，调 article-rpc 缓存 60s）
4. 用户状态（未封禁/未禁言；禁言用户返回明确错误）
5. 敏感词检测：本地 DFA（O(n) 一次扫描）
   ├─ hit=block → 直接拒绝（或转待审）
   └─ hit=review → 标记待审
6. （可选）云内容安全审核（异步或同步，按配置）
7. 事务写入 comment + outbox(comment.created)
8. 增量更新计数（Redis INCR，异步落库）
9. 返回评论体（前端乐观更新）
```

**数据表（`noviterra_comment` 库）**

| 表 | 关键字段 | 索引 |
| --- | --- | --- |
| `comment` | id, article_id, user_id, root_id, parent_id, reply_to_user_id, content, content_html, status(pending/approved/rejected/deleted), is_top, like_count, reply_count, ip, region, ua, audit_by, audit_at, created_at, updated_at, deleted_at | `idx_article_root(article_id,root_id,created_at)`、`idx_root(root_id,created_at)`、`idx_user(user_id,created_at desc)`、`idx_status(status,created_at)` |
| `comment_like` | id, comment_id, user_id, created_at | `uk_comment_user`、`idx_user` |
| `comment_report` | id, comment_id, user_id, reason, detail, status(pending/handled/rejected), handler_id, handle_note, handled_at, created_at | `idx_status(status,created_at)`、`idx_comment` |
| `sensitive_word` | id, word, level(block/review/replace), category, status, created_at | `uk_word`、`idx_level_status(level,status)` |

**缓存**

| Key | 类型 | TTL | 说明 |
| --- | --- | --- | --- |
| `nv:{env}:comment:count:{articleId}` | String | 1h | 评论总数 |
| `nv:{env}:comment:root:{articleId}:{sort}:{cursor}` | String | 60s | 首屏评论（只缓存第一页） |
| `nv:{env}:comment:like:{commentId}` | Set | — | 点赞用户集合 |
| `nv:{env}:comment:rate:{uid}` | String | 10min | 频率计数 |
| `nv:{env}:comment:dfa:ver` | String | — | 敏感词库版本（变更时 bump，各实例热重载） |

**敏感词实现（`rpc/internal/domain/dfa.go`）**

- 构建 **DFA Trie**（Aho-Corasick 或简单 Trie），一次扫描 O(n)。
- 入库/更新后 bump 版本号 → 各实例 30s 轮询版本号，变化则重建 Trie（原子替换 `atomic.Pointer`）。
- 三种级别：`block`（拒绝）、`review`（转待审）、`replace`（替换为 `*`）。
- 变体处理：去除零宽字符、全角转半角、简单拼音/谐音库（可选，别过度）。
- 提供 `ImportSensitiveWords`（批量导入 txt，一行一个词）与导出。

**规则**

| 项 | 规则 |
| --- | --- |
| 嵌套层级 | 最多 2 层（根 + 回复）；回复的回复仍挂同一 root |
| 长度 | 1~1000 字符；富文本仅允许白名单标签（前端渲染时净化） |
| 编辑 | 允许 5 分钟内编辑一次（编辑后标记 `edited_at` 并重新过敏感词） |
| 删除 | 软删，保留结构（显示"该评论已删除"），回复仍可见 |
| 置顶 | 仅文章作者/管理员可置顶 |
| 举报阈值 | 同一评论被 3 个不同用户举报 → 自动转待审 |
| 计数校准 | 每日 03:30 校准 `article.comment_count` 与 `comment.reply_count` |

**XSS 防护（双端）**：后端对 `content` 做 HTML 转义后存 `content_html`（或只存纯文本 + 前端渲染），前端渲染时再用 DOMPurify 净化。**禁止**直接 `v-html` / `dangerouslySetInnerHTML` 未净化内容。

---

### 6.7 `app/file`（文件与媒体中心）

**职责**

1. 文件上传：直传（≤ 5MB）、分片上传（大文件）、秒传（hash 命中）。
2. 存储抽象：本地 / MinIO / 阿里云 OSS / 腾讯 COS，配置切换。
3. 图片处理：压缩、缩略图、WebP、水印、EXIF 清理（含 GPS 隐私）。
4. 文件元数据管理、引用计数、回收站。
5. 私有文件签名 URL（有效期）、防盗链。
6. 上传安全检查：扩展名/MIME 白名单、文件头魔数校验、大小限制、可执行内容检测。

**不做什么**：不管理"文件与文章的关联规则"（`article-rpc` 只存 file_id 或 URL）。

**目录**（标准三目录）：`app/file/api/`（`file-api` 8805：上传接口、凭证签发、后台附件管理）、`app/file/rpc/`（`file-rpc` 9005：存储抽象、分片状态机、图片处理、引用计数）、`app/file/model/`（`file`、`file_chunk`、`file_ref`）。

**上传流程（两条路径）**

```
路径 A：服务端中转（小文件，默认 ≤ 5MB）
  客户端 → Nginx → file-api → file-rpc.Upload
    ├─ 【api】JWT + 大小/类型预校验 + 用户配额
    ├─ 【rpc】魔数校验 → 图片压缩/缩略图/WebP/EXIF 清理
    ├─ 【rpc】上传存储（storage.Put）
    ├─ 【rpc】写 file 表；发事件 file.uploaded
    └─ 返回 { fileId, url, thumbUrl, width, height, size }

路径 B：客户端直传（大文件 / 减轻网关压力，推荐用于 > 5MB）
  1. 客户端算文件 SHA256 → POST /files/check {hash, size, name}
     ├─ 命中（file 表有 hash 且 status=done）→ 秒传，直接返回已有 fileId/URL
     └─ 未命中 → 返回 uploadId（预签名 URL 或分片上传凭证）
  2. 客户端按 5MB 分片上传：
     - 方式 B1：file-rpc 返回 MinIO/OSS 的预签名 URL，客户端直传对象存储（推荐）
     - 方式 B2：客户端把分片传到 file-rpc，由 file-rpc 转发（网关压力大，仅小规模用）
  3. POST /files/multipart/complete {uploadId, parts}
     ├─ file-rpc 调用 storage.CompleteMultipart
     ├─ 校验最终 size 与声明一致、计算 hash 一致（可选）
     ├─ 图片：异步生成缩略图（job 或协程）
     └─ 写 file 表，返回最终 URL
```

**分片上传状态机（Redis + DB 双写）**

- `nv:{env}:file:upload:{uploadId}` (Hash)：`file_hash`、`file_name`、`size`、`total_parts`、`done_parts`(Set)、`storage_key`、`expire_at`。
- `file_chunk` 表：持久化分片记录（Redis 丢失可恢复）。
- **断点续传**：客户端先 `GET /files/multipart/:uploadId/status` 拿到已上传分片索引，只补缺失分片。
- 24 小时未完成 → app/job 调用 `AbortMultipart` 清理（避免产生存储碎片费用）。

**数据表（`noviterra_file` 库）**

| 表 | 关键字段 | 索引 |
| --- | --- | --- |
| `file` | id, user_id, biz_type, biz_id, origin_name, storage(local/minio/oss/cos), bucket, object_key, url, mime, ext, size, hash_md5, hash_sha256, width, height, duration, status(uploading/done/failed/deleted), is_public, ref_count, ip, created_at, updated_at, deleted_at | `uk_sha256_size(hash_sha256,size)`（秒传）、`idx_user(user_id,created_at desc)`、`idx_biz(biz_type,biz_id)`、`idx_status(status,created_at)` |
| `file_chunk` | id, upload_id, file_hash, chunk_index, chunk_hash, size, storage_key, status, created_at | `uk_upload_index(upload_id,chunk_index)`、`idx_expire(created_at)` |
| `file_ref` | id, file_id, biz_type, biz_id, created_at | `uk_file_biz(file_id,biz_type,biz_id)`（引用计数，删除前必须 ref_count=0） |

**上传安全规则**

| 项 | 规则 |
| --- | --- |
| 大小限制 | 头像 ≤ 2MB；图片 ≤ 10MB；普通附件 ≤ 20MB；视频 ≤ 200MB（可配置） |
| 扩展名白名单 | 图片：jpg/jpeg/png/gif/webp/svg（**SVG 必须净化或禁止**）/bmp；文档：pdf/txt/md/zip；音视频：mp4/webm/mp3 |
| MIME 校验 | 必须同时校验 `Content-Type` 声明、扩展名、**文件头魔数**（三者一致才放行） |
| 文件名 | 服务端重命名（`{biz}/{yyyymm}/{dd}/{uuid}.{ext}`），**永不使用用户文件名作为 Key** |
| SVG/XSS | SVG 视为脚本风险：默认禁止上传，或强制下载（`Content-Disposition: attachment`）+ CSP |
| 图片二次处理 | 一律重新编码（`imaging`），剥离 EXIF（含 GPS），顺带消除内嵌脚本 |
| 用户配额 | 普通用户 500MB / 作者 2GB / 管理员不限（Redis 计数 + DB 兜底） |
| 防盗链 | 私有文件用签名 URL（默认 1h）；公开图片校验 `Referer` 白名单 |
| 类型嗅探 | `http.DetectContentType` + 白名单二进制签名比对 |

**缓存**

| Key | 类型 | TTL | 说明 |
| --- | --- | --- | --- |
| `nv:{env}:file:sec:{sha256}` | String(fileId) | 24h | 秒传命中 |
| `nv:{env}:file:upload:{uploadId}` | Hash | 24h | 分片上传状态 |
| `nv:{env}:file:quota:{uid}` | String | 1h | 已用容量 |
| `nv:{env}:file:sign:{fileId}:{uid}` | String | 1h | 私有文件访问凭证 |

**事件**：`file.uploaded.v1`（fileId, uid, bizType, size）、`file.deleted.v1`（供 article 清理引用）。

---

### 6.8 `app/message`（消息与通知中心）

**职责**

1. **站内信**：评论/回复/点赞/关注/@提及/审核结果/系统公告。
2. **多通道投递**：站内信、邮件、短信、WebSocket 实时推送、Web Push。
3. **模板管理**：`msg_template`（变量占位 `{{.Username}}`），支持多语言。
4. **投递可靠性**：异步队列 + 重试 + 死信 + 发送日志。
5. **订阅偏好**：用户可关闭某类通知（在 user_setting 中存偏好，本服务读取判断）。
6. **聚合去重**：同一目标在 5 分钟内多次点赞 → 合并为"XXX 等 3 人赞了你"。

**不做什么**：不决定"业务上该不该通知"（由事件驱动，本服务只做"通道与模板"）。

**目录**

```
server/app/message/
├── api/                                 # ── message-api（HTTP 8806）
│   ├── message.go  message.api  etc/message-api.yaml
│   └── internal/{config,svc,handler,logic,types,middleware,consts}/
│        └── logic/                      # 站内信列表/未读/已读 + ★ WebSocket 入口（/ws/messages）
├── rpc/                                 # ── message-rpc（gRPC 9006）
│   ├── message.go  message.proto  etc/message-rpc.yaml  pb/  message/
│   └── internal/
│       ├── svc/                         # DB、Redis、SmsClient、EmailClient、PushHub
│       ├── logic/                       # 统一发送、模板、群发、重试
│       ├── channel/                     # ★ 通道适配，实现统一 Sender 接口
│       │   ├── sender.go                #   type Sender interface { Send(ctx, *Message) error; Channel() string }
│       │   ├── site.go
│       │   ├── email.go                 #   包装 pkg/thirdparty/email
│       │   ├── sms.go                   #   包装 pkg/thirdparty/sms
│       │   └── websocket.go             #   Hub：连接管理、心跳、按 uid 推送
│       ├── consumer/                    # 事件消费者：comment.created / article.liked / user.followed / article.published …
│       ├── template/                    # 模板渲染（text/template + 多语言）
│       └── middleware/ domain/ dto/ vo/ consts/
└── model/{entity,repository,cache}      # message / msg_template / msg_send_log / msg_broadcast
```

**RPC 接口（摘要）**

```protobuf
service Message {
  rpc ListMessages(ListMessagesReq) returns (ListMessagesResp);
  rpc GetUnreadCount(GetUnreadCountReq) returns (UnreadCountResp);   // 按类型分组
  rpc MarkRead(MarkReadReq) returns (Empty);                         // 批量
  rpc MarkAllRead(MarkAllReadReq) returns (Empty);
  rpc DeleteMessage(DeleteMessageReq) returns (Empty);
  rpc Send(SendReq) returns (SendResp);                              // 统一发送入口（内部）
  rpc SendByTemplate(SendByTemplateReq) returns (SendResp);
  rpc Broadcast(BroadcastReq) returns (BroadcastResp);               // 后台群发（异步任务）
  // 管理
  rpc ListTemplates / CreateTemplate / UpdateTemplate / DeleteTemplate
  rpc ListSendLogs(ListSendLogsReq) returns (ListSendLogsResp);
  rpc RetryFailed(RetryFailedReq) returns (Empty);
}
```

**统一发送流程（所有通道共用）**

```
输入：{ templateCode, channel[], toUserIds[], vars, bizType, bizId, dedupKey }
1. 模板解析 + 变量渲染（多语言按用户 locale 选择）
2. 偏好过滤（用户关闭该类通知 → 跳过该通道，站内信可选强制）
3. 去重（dedupKey = sha1(templateCode + bizType + bizId + toUserId)，Redis SETNX 24h）
4. 聚合（同类型同目标 5 分钟内合并，更新既有消息内容与时间）
5. 落库 message（站内信）→ WebSocket 实时推送（在线用户）
6. 外发通道（邮件/短信）→ 写 msg_send_log(status=pending) → 异步 worker 投递
7. 投递结果回写日志；失败指数退避重试 5 次 → 死信 + 告警
```

**WebSocket 设计**

| 项 | 设计 |
| --- | --- |
| 端点 | `GET /ws/messages?token=<access_token>`（`message-api`，边缘层单独路由 `/ws/**`） |
| 鉴权 | 连接时校验 access token（不能放 Header 时用 query，但**必须**限制 TTL 与来源） |
| 会话 | `Hub{uid → []*Conn}`，支持同一用户多设备 |
| 心跳 | 服务端 30s Ping，客户端 45s 无 Pong 断开 |
| 推送格式 | `{type: "notification", data: {...}, ts: 1735...}` |
| 未读 | 推送同时更新 `nv:{env}:msg:unread:{uid}` 计数，前端角标实时更新 |
| 扩容 | 多实例时用 Redis Pub/Sub 广播（`nv:{env}:msg:push`），各实例只推自己的连接；连接数 > 1w 时考虑独立 WS 网关 |
| 断线补偿 | 重连时前端拉 `/messages?since=<lastId>` 补齐 |

**数据表（`noviterra_message` 库）**

| 表 | 关键字段 | 索引 |
| --- | --- | --- |
| `message` | id, to_user_id, from_user_id, type(system/comment/reply/like/follow/mention/audit/announcement), biz_type, biz_id, title, content, extra(JSON), is_read, read_at, created_at | `idx_to_read(to_user_id,is_read,created_at desc)`、`idx_to_type(to_user_id,type,created_at desc)` |
| `msg_template` | id, code, name, channel(site/email/sms/push), locale, title_tpl, content_tpl, vars(JSON), status, created_at, updated_at | `uk_code_channel_locale` |
| `msg_send_log` | id, template_code, channel, to_user_id, to_addr, title, content, provider, provider_msg_id, status(pending/sent/failed/dead), retry_count, next_retry_at, error, created_at, sent_at | `idx_status_next(status,next_retry_at)`、`idx_to_user(to_user_id,created_at desc)` |
| `msg_broadcast` | id, title, content, target(all/role/user_ids), total, sent, failed, status, created_by, created_at | `idx_status` |

**缓存**：`nv:{env}:msg:unread:{uid}`（Hash：type → count，TTL 1h）、`nv:{env}:msg:dedup:{sha1}`（24h）、`nv:{env}:msg:push`（Pub/Sub 频道）。

**默认模板清单（seed）**：`welcome`、`email_verify`、`password_reset`、`comment_reply`、`comment_mention`、`article_liked`、`article_collected`、`new_follower`、`article_reviewed_pass`、`article_reviewed_reject`、`user_banned`、`announcement`。

---

### 6.9 `app/search`（检索中心）

**职责**

1. 索引维护：消费 `article.published/updated/deleted` 事件，增删改索引。
2. 全文检索：关键词、筛选（分类/标签/作者/时间/热度）、排序（相关度/最新/最热）。
3. 高亮片段、搜索建议（前缀补全）、拼写纠错（可选）。
4. 热搜词统计（Redis ZSet + 定时落库）。
5. 索引重建（全量，供后台"重建索引"按钮与首次初始化）。

**引擎选择**：起步用 **Meilisearch**（轻量、开箱即用、中文分词可用、部署简单）；文章量 > 100w 或需要复杂聚合时迁移到 **Elasticsearch/OpenSearch**。**用接口屏蔽差异**：

```go
// server/app/search/rpc/internal/engine/engine.go
type Engine interface {
    Index(ctx context.Context, doc *Document) error
    BulkIndex(ctx context.Context, docs []*Document) error
    Delete(ctx context.Context, id string) error
    Search(ctx context.Context, q *Query) (*Result, error)
    Suggest(ctx context.Context, prefix string, limit int) ([]string, error)
    RecreateIndex(ctx context.Context, mapping string) error
    Health(ctx context.Context) error
}
// 实现：engine/meilisearch/、engine/elasticsearch/、engine/memory/（测试用）
```

**目录**（标准三目录）：`app/search/api/`（`search-api` 8807：检索、筛选、高亮、建议、热搜）、`app/search/rpc/`（`search-rpc` 9007：索引增删改、全量重建、索引统计，含 `internal/engine/`）、`app/search/model/`（`search_keyword_log`、`search_rebuild_task`；索引本体在 Meilisearch/ES，不在 MySQL）。

**索引文档结构**

```json
{
  "id": "1001",
  "title": "Go 微服务实践",
  "summary": "……",
  "content": "纯文本正文（HTML 已剥离，用于检索）",
  "tags": ["Go", "微服务"],
  "category_id": 3,
  "category_name": "后端",
  "author_id": 7,
  "author_name": "张三",
  "status": "published",
  "visibility": "public",
  "published_at": 1735689600,
  "view_count": 1200,
  "like_count": 88,
  "comment_count": 12,
  "cover": "https://…",
  "lang": "zh"
}
```

**索引规则**

| 项 | 规则 |
| --- | --- |
| 可搜索字段 | title（权重 3）、summary（2）、content（1）、tags（2）、author_name（1.5） |
| 过滤字段 | status、visibility、category_id、tags、author_id、published_at |
| 排序 | 相关度（默认）、`published_at desc`、热度 |
| 仅索引 | `status=published AND visibility=public` 的文章（其它状态主动删除索引） |
| 索引延迟 | 事件驱动，正常 < 3s；失败进重试队列；每小时对账（DB 与索引差异） |
| 全量重建 | `app/job` 每日 04:00 或手动触发；用别名切换（`articles_v1` → `articles_v2`）实现零停机重建 |
| 敏感内容 | 被举报/审核中的文章从索引移除 |

**RPC 接口（摘要）**

```protobuf
service Search {
  rpc SearchArticles(SearchArticlesReq) returns (SearchArticlesResp);   // 支持分页、筛选、排序、高亮
  rpc Suggest(SuggestReq) returns (SuggestResp);
  rpc HotKeywords(HotKeywordsReq) returns (HotKeywordsResp);
  rpc RecordKeyword(RecordKeywordReq) returns (Empty);                 // 记录搜索词
  rpc IndexArticle(IndexArticleReq) returns (Empty);                   // 手动/补偿
  rpc DeleteArticleIndex(DeleteArticleIndexReq) returns (Empty);
  rpc RebuildIndex(RebuildIndexReq) returns (RebuildIndexResp);        // 异步，返回 taskId
  rpc GetRebuildStatus(GetRebuildStatusReq) returns (RebuildStatusResp);
  rpc GetIndexStats(GetIndexStatsReq) returns (IndexStatsResp);        // 后台看板
}
```

**缓存**：`nv:{env}:search:hot:{yyyymmdd}`（ZSet 热词）、`nv:{env}:search:suggest:{prefix}`（String，5min）、`nv:{env}:search:result:{queryHash}`（String，60s，仅首页）。

**中文分词**：Meilisearch 内置；若用 ES 需装 IK 分词器（`analysis-ik`），在 mapping 中指定 `analyzer: ik_max_word` / `search_analyzer: ik_smart`。

---

### 6.10 `app/stat`（统计与埋点）

**职责**

1. 实时计数：PV/UV、点赞/收藏/评论增量、在线用户数。
2. 离线聚合：按日/小时维度的站点、文章、用户、分类统计。
3. 后台看板数据：概览卡片、趋势图、Top 榜。
4. 埋点接收与落库（前端埋点、服务端事件）。
5. 搜索词统计（与 search-rpc 协作，落库明细）。

**不做什么**：不做业务写入；不做报表导出文件（由 app/job 生成）。

**目录**（标准三目录）：`app/stat/api/`（`stat-api` 8808：后台看板/趋势/Top 榜，以及前台 `GET /api/v1/stats/site`）、`app/stat/rpc/`（`stat-rpc` 9008：埋点上报、计数增减、日聚合）、`app/stat/model/`（`stat_site_daily`、`stat_article_daily`、`stat_user_daily`、`stat_category_daily`、`stat_search_keyword`、`stat_visit_log`）。

**数据表（`noviterra_stat` 库）**

| 表 | 关键字段 | 说明 |
| --- | --- | --- |
| `stat_site_daily` | id, stat_date, pv, uv, new_user, active_user, new_article, new_comment, new_like, search_count | 唯一键 `uk_date` |
| `stat_article_daily` | id, article_id, stat_date, pv, uv, like_add, collect_add, comment_add, share_add, avg_stay_ms | `uk_article_date` |
| `stat_user_daily` | id, user_id, stat_date, publish_count, comment_count, view_received, like_received | `uk_user_date` |
| `stat_category_daily` | id, category_id, stat_date, article_add, view_sum | `uk_category_date` |
| `stat_search_keyword` | id, keyword, stat_date, count, result_count, click_count | `uk_keyword_date` |
| `stat_visit_log` | id, uid, session_id, ip, region, ua, device, os, browser, path, referer, stay_ms, created_at | **按天分区**（`PARTITION BY RANGE (TO_DAYS(created_at))`），保留 90 天 |
| `stat_online` | Redis only | 在线：`nv:{env}:stat:online`（ZSet，score=最后活跃时间） |

**指标采集方式**

| 指标 | 采集方式 | 存储 |
| --- | --- | --- |
| 文章 PV/UV | 详情接口异步 HINCR（Redis Hash）+ 5 分钟批量落库 | `article_stat_daily` |
| 站点 PV/UV | 网关埋点事件 → MQ → stat-rpc 消费 | `stat_site_daily` |
| 活跃用户 | 登录/请求去重（Redis Bitmap：`stat:active:{date}`，offset=uid） | Bitmap + 每日落库 |
| 在线用户 | WebSocket 连接 + 心跳写 ZSet，定时清理超时 | Redis |
| 搜索词 | search-rpc 每次搜索 `RecordKeyword`（Redis ZSet 计数） | Redis + 每日落库 |

**RPC 接口（摘要）**

```protobuf
service Stat {
  rpc ReportEvent(ReportEventReq) returns (Empty);              // 前端/服务端埋点统一入口
  rpc IncrCounter(IncrCounterReq) returns (Empty);              // 批量计数
  rpc GetSiteOverview(GetSiteOverviewReq) returns (SiteOverview);      // 后台看板卡片
  rpc GetTrend(GetTrendReq) returns (GetTrendResp);                    // 趋势（date range + dimension）
  rpc GetTopArticles(GetTopArticlesReq) returns (TopArticlesResp);
  rpc GetTopAuthors / GetCategoryDistribution / GetUserGrowth
  rpc GetOnlineCount(GetOnlineCountReq) returns (GetOnlineCountResp);
  rpc GetArticleStat(GetArticleStatReq) returns (ArticleStatResp);     // 单篇明细（作者看）
  rpc AggregateDaily(AggregateDailyReq) returns (Empty);               // app/job 调用
}
```

**聚合策略**：日维度在**次日 00:30** 由 app/job 触发 `AggregateDaily`，从 Redis 计数 + `article_stat_daily` 汇总到 `stat_site_daily`；重跑安全（按 `stat_date` 幂等覆盖，用 `INSERT ... ON DUPLICATE KEY UPDATE`）。

**数据量控制**：`stat_visit_log` 是最大表 → 按天分区 + 保留 90 天 + 超期 `DROP PARTITION`（秒级、无锁）；如需更长周期分析，导出到 ClickHouse（后续演进）。

---

### 6.11 `app/third`（第三方能力编排服务）

**职责**：见 [第 5 章](#5-第三方-api-接入规范重点)。总结为四件事：

1. **编排**：选供应商、主备切换、超时重试熔断、批量拆包。
2. **治理**：配额、限流、缓存、降级、开关（配置中心热更新）。
3. **审计**：`third_api_log` 全量记录（脱敏）、用量报表。
4. **任务化**：长耗时能力（AI 摘要、批量翻译、OCR）转为异步任务。

**分包归属（先看第 5 章再看这里）**：`server/pkg/thirdparty/**` = 能力适配（能调用）；`app/third/rpc` = 该不该调、配额、落库、降级（编排）。

**目录**

```
server/app/third/
├── api/                                    # ── third-api（HTTP 8809）：仅后台界面用
│   ├── third.go  third.api  etc/third-api.yaml
│   └── internal/{config,svc,handler,logic,types,middleware,consts}/
│        └── logic/                         # 配额查询、调用日志列表（薄转发）
├── rpc/                                    # ── third-rpc（gRPC 9009）
│   ├── third.go  third.proto  etc/third-rpc.yaml  pb/  third/
│   └── internal/
│       ├── config/config.go                # 各能力配置 + 密钥（来自环境变量）
│       ├── svc/service_context.go          # 注入 pkg/thirdparty/* 的各 Translator/LLM/Storage/…
│       ├── logic/
│       │   ├── translatelogic.go
│       │   ├── translatesummarylogic.go    # 摘要翻译
│       │   ├── summarizeLogic.go           # AI 摘要
│       │   ├── polishlogic.go              # AI 润色
│       │   ├── generatetagslogic.go        # AI 标签推荐
│       │   ├── moderatelogic.go            # 内容审核
│       │   ├── ipregionlogic.go
│       │   ├── sendemailLogic.go / sendsmslogic.go
│       │   ├── oauthloginlogic.go          # 第三方登录换取用户信息
│       │   ├── getquotalogic.go
│       │   └── listlogslogic.go
│       ├── task/                           # 异步任务（Redis Stream / asynq）
│       │   ├── translate_task.go
│       │   └── summarize_task.go
│       ├── consumer/                       # 消费 article.published 触发摘要/翻译
│       └── domain/ middleware/ dto/ vo/ consts/
└── model/{entity,repository}               # ThirdApiLog、ThirdApiQuota、ThirdTask
```

**RPC 接口（摘要）**

```protobuf
service Third {
  // 同步能力（低延迟）
  rpc Translate(TranslateReq) returns (TranslateResp);
  rpc ModerateText(ModerateTextReq) returns (ModerateTextResp);
  rpc ModerateImage(ModerateImageReq) returns (ModerateTextResp);
  rpc DetectLanguage(DetectLanguageReq) returns (DetectLanguageResp);
  rpc IpRegion(IpRegionReq) returns (IpRegionResp);
  rpc SendEmail(SendEmailReq) returns (SendResp);
  rpc SendSms(SendSmsReq) returns (SendResp);
  rpc OAuthExchange(OAuthExchangeReq) returns (OAuthUserInfo);
  // 异步能力（返回 taskId）
  rpc SubmitTask(SubmitTaskReq) returns (SubmitTaskResp);       // type: summarize|translate_article|ocr
  rpc GetTask(GetTaskReq) returns (TaskInfo);
  rpc CancelTask(CancelTaskReq) returns (Empty);
  // 治理
  rpc GetQuota(GetQuotaReq) returns (QuotaInfo);
  rpc ListLogs(ListLogsReq) returns (ListLogsResp);
  rpc Health(HealthReq) returns (HealthResp);                    // 各供应商健康状态
}
```

**异步任务模型**：`third_task` 表（id, type, biz_type, biz_id, params(JSON), status(pending/running/done/failed/canceled), progress, result(JSON), errcode, retry, created_at, updated_at）+ Redis Stream 队列 + worker 协程池（并发 5，AI 类限 2）。

**降级开关（配置中心）**：`third.translate.enabled`、`third.ai.enabled`、`third.contentsec.fail_mode`、`third.cache.enabled`。运维可在不重启的情况下关闭任一能力。

---

### 6.12 `app/job`（定时任务与后台作业）

**职责**：集中管理所有**跨服务/重任务**，避免每个服务各起一个 cron 造成重复执行与难以观测。

**为什么需要独立进程**：分布式环境下多实例同时跑定时任务会重复执行；统一到一个服务 + 分布式锁，任务清单清晰、日志集中、可手动触发与查看历史。

**任务清单（初版）**

| 任务 | Cron | 说明 | 幂等键 |
| --- | --- | --- | --- |
| `publish_scheduled_articles` | `* * * * *` | 扫描 `status=scheduled AND scheduled_at<=now`，转 published 并发事件 | article_id |
| `flush_view_counts` | `*/5 * * * *` | 把 Redis 浏览计数批量落库 | 批次时间 |
| `recalc_hot_articles` | `*/10 * * * *` | 重算热榜 ZSet | 时间桶 |
| `reconcile_counters` | `0 3 * * *` | 对账文章/评论/用户计数，修正偏差 | stat_date |
| `aggregate_daily_stats` | `30 0 * * *` | 汇总昨日统计 | stat_date |
| `clean_expired_data` | `0 4 * * *` | 清理过期：软删 > 30 天的文章/评论、过期验证码、失败日志 | date |
| `rebuild_search_index` | `0 4 * * 0` | 每周全量重建索引 | week |
| `sync_search_index` | `*/30 * * * *` | 增量对账（DB 与索引差异） | 时间桶 |
| `retry_failed_messages` | `*/2 * * * *` | 重投失败的邮件/短信 | log_id |
| `retry_outbox` | `*/1 * * * *` | 重投未成功投递的 outbox 事件 | event_id |
| `clean_multipart_uploads` | `0 * * * *` | 清理 24h 未完成的分片上传 | upload_id |
| `refresh_ip2region` | `0 5 * * 1` | 更新 IP 库（如有） | week |
| `reset_third_quota` | `5 0 * * *` | 重置第三方日配额并生成用量报表 | date |
| `warm_cache` | `0 6 * * *` | 预热首页/热榜/分类页缓存 | date |
| `archive_visit_log` | `0 2 * * *` | 创建明日分区、删除 90 天前分区 | date |
| `check_health_digest` | `0 9 * * *` | 生成每日健康摘要（错误率、慢接口）推送管理员 | date |

**任务实现骨架**

```go
// 每个任务：注册 + 分布式锁 + 幂等 + 超时 + 日志 + 指标
func init() {
    schedule.Register("flush_view_counts", "*/5 * * * *", schedule.WithLock("job:flush_view:lock", 4*time.Minute), func(ctx context.Context) error {
        ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
        defer cancel()
        return logic.NewFlushViewCountsLogic(ctx, svcCtx).Flush()
    })
}
```

**规则**

1. **每任务必须加分布式锁**（`pkg/cache` 的 Lock），锁 TTL < 任务间隔，避免重入。
2. **必须幂等**：重复执行结果一致（用幂等键去重或 `ON DUPLICATE KEY UPDATE`）。
3. **必须设置超时**，且超时 < 下次触发时间。
4. **禁止在任务里跑无界循环**（必须分批 + 每批检查 ctx）。
5. 支持**手动触发**（内部 HTTP `POST /internal/jobs/{name}/run`，需 `x-internal-token`）。
6. 执行历史写 `sys_job_log`（开始/结束/耗时/结果/错误），后台可查询。
7. 关键任务失败**必须告警**（Prometheus 指标 + Alertmanager）。

**目录**（调度型：**无 `api/`、无 `rpc/`、无 `model/`**，因为它不对终端暴露协议、也不拥有数据）

```
server/app/job/
├── job.go
├── etc/job.yaml
└── internal/
    ├── config/
    ├── svc/service_context.go     # 各域 rpc client（通过 RPC 触发逻辑，绝不直连别人的库！）
    ├── server/http.go             # 内部管理端口：/internal/jobs、/healthz、/metrics
    ├── jobs/                      # 每个任务一个文件
    │   ├── publish_scheduled.go
    │   ├── flush_view_counts.go
    │   └── …
    └── consts/
```

> ⚠️ **铁律**：`app/job` **不直连任何业务库**，只通过各域 rpc / MQ 完成任务。它自己只连 Redis（锁与计数）与 `noviterra_system`（任务元数据与日志）。这样保证数据所有权不被破坏。

---

## 7. 数据设计

### 7.1 数据库总体规划

**一服务一 schema（逻辑隔离），共用一套 MySQL 实例（起步）**：

| Schema | 属主服务 | 说明 |
| --- | --- | --- |
| `noviterra_user` | user-rpc | 用户、认证、关系、登录日志 |
| `noviterra_system` | system-rpc | RBAC、配置、字典、公告、审计、任务元数据 |
| `noviterra_content` | article-rpc | 文章、版本、分类标签、专栏、点赞收藏、翻译、outbox |
| `noviterra_comment` | comment-rpc | 评论、点赞、举报、敏感词 |
| `noviterra_file` | file-rpc | 文件、分片、引用 |
| `noviterra_message` | message-rpc | 站内信、模板、发送日志、群发 |
| `noviterra_third` | third-rpc | 第三方调用日志、配额、异步任务 |
| `noviterra_search` | search-rpc | 搜索统计（热词明细、重建任务）—— 索引本体在 Meilisearch/ES |
| `noviterra_stat` | stat-rpc | 各类统计与埋点明细 |

**演进**：单实例 → 主从（读写分离）→ 按压力拆库（content/stat 独立实例）→ 超大表分区/归档。**禁止**在起步阶段就上分库分表（复杂度不划算）。

**字符集与基础设置**

```sql
-- 统一规则
CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci
ENGINE=InnoDB
ROW_FORMAT=DYNAMIC
-- 时区：MySQL 实例设 UTC；连接串带 parseTime=true&loc=UTC
```

**命名规范**

| 对象 | 规范 | 示例 |
| --- | --- | --- |
| 表 | 小写下划线，单数，语义前缀按域可选 | `article`、`article_version`、`user_login_log` |
| 字段 | 小写下划线 | `published_at`、`view_count` |
| 布尔 | `is_` 前缀，`tinyint(1)` | `is_top`、`is_read` |
| 时间 | `_at` 后缀（UTC，`datetime(3)`） | `created_at`、`deleted_at` |
| 计数 | `_count` 后缀（`int unsigned`） | `like_count` |
| 索引 | `idx_`（普通）、`uk_`（唯一）、`ft_`（全文） | `uk_slug`、`idx_user_status` |
| 外键 | **不使用数据库外键**，靠应用层与索引保证 | — |

**公共字段（所有业务表）**

```sql
`id`         bigint unsigned NOT NULL COMMENT '雪花ID',
`created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
`updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
`deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '0=未删除，否则为删除时间戳(ms)',  -- 需要软删的表
PRIMARY KEY (`id`)
```

> **软删除方案说明**：采用 `deleted_at bigint`（0 表示未删）而非 `datetime NULL`，原因是 MySQL 唯一索引对 NULL 不去重——用 0 可让唯一索引在"未删除"集合上真正生效，同时保留唯一约束。查询统一由 GORM 的 `SoftDelete` 插件/RawCondition 注入 `deleted_at = 0`（在 `pkg/db` 中统一封装，业务代码无感）。

### 7.2 表清单汇总（速查）

| Schema | 表 | 归属服务 | 预估量级 | 归档策略 |
| --- | --- | --- | --- | --- |
| user | `user` | user-rpc | 10w~100w | 不归档 |
| user | `user_auth`、`user_profile`、`user_setting`、`user_stat` | user-rpc | 与 user 同量 | 不归档 |
| user | `user_follow` | user-rpc | 千万级 | 3 年归档 |
| user | `user_role` | user-rpc | 小 | — |
| user | `user_login_log` | user-rpc | 千万级 | 按月归档，保留 6 个月 |
| system | `role`、`permission`、`role_permission`、`sys_config`、`sys_dict`、`sys_dict_item`、`sys_announcement` | system-rpc | 小 | — |
| system | `sys_operation_log` | system-rpc | 千万级 | 按月归档，保留 12 个月 |
| system | `sys_job`、`sys_job_log` | system-rpc | 百万级 | 保留 3 个月 |
| content | `article` | article-rpc | 100w | 不归档（冷文章可下沉归档表） |
| content | `article_content`、`article_version` | article-rpc | 100w / 500w | 版本保留最近 20 个 |
| content | `category`、`tag`、`article_tag`、`series`、`series_article` | article-rpc | 小 / 千万 | — |
| content | `article_like`、`article_collect`、`collect_folder` | article-rpc | 千万 | 2 年归档 |
| content | `article_translation` | article-rpc | 百万 | — |
| content | `article_slug_history` | article-rpc | 小 | — |
| content | `article_view_log` | article-rpc | 亿级 | 按天分区，保留 90 天 |
| content | `outbox` | 各写入服务 | 百万 | 已投递的 7 天后清理 |
| comment | `comment` | comment-rpc | 千万 | 3 年归档 |
| comment | `comment_like`、`comment_report`、`sensitive_word` | comment-rpc | 千万/百万/万 | 2 年归档 |
| file | `file`、`file_chunk`、`file_ref` | file-rpc | 百万/百万/千万 | chunk 24h 清理 |
| message | `message`、`msg_send_log`、`msg_broadcast` | message-rpc | 千万/千万/小 | message 保留 6 个月 |
| message | `msg_template` | message-rpc | 小 | — |
| third | `third_api_log` | third-rpc | 千万 | 按天分区，保留 6 个月 |
| third | `third_api_quota`、`third_task` | third-rpc | 小 / 百万 | task 保留 30 天 |
| search | `search_keyword_log`、`search_rebuild_task` | search-rpc | 百万 | 保留 6 个月 |
| stat | `stat_site_daily`、`stat_article_daily`、`stat_user_daily`、`stat_category_daily`、`stat_search_keyword` | stat-rpc | 万~千万 | 保留 3 年（可下沉数仓） |
| stat | `stat_visit_log` | stat-rpc | 亿级 | 按天分区，保留 90 天 |

**完整的建表 SQL 由 `server/deploy/sql/migrations/` 下的 goose 迁移文件承载**，本文只定义结构原则与关键索引，不重复 DDL。

### 7.3 索引与查询规范

| 规则 | 说明 |
| --- | --- |
| 每表必须有主键 | 雪花 `id`；**不使用联合主键作为唯一主键**（关系表可用联合主键 + 独立 id 视情况） |
| 索引数量 | 单表索引 ≤ 6 个（含主键）；写多读少的表更要克制 |
| 联合索引顺序 | 等值条件在前、范围条件在后；区分度高的列优先 |
| 覆盖索引 | 高频列表查询（如 `(user_id, status, published_at)`）尽量做成覆盖索引，避免回表 |
| 禁止左模糊 | `LIKE '%xx'` 禁用；前缀匹配 `LIKE 'xx%'` 可用（走索引）；全文检索走 search-rpc |
| 禁止函数列 | `WHERE DATE(created_at)='2025-01-01'` 禁用 → 改 `created_at >= ? AND created_at < ?` |
| 隐式转换 | 禁止字符串字段与数字比较（索引失效） |
| 大偏移分页 | `LIMIT 100000,20` 禁用 → 游标分页 |
| 单条 SQL 返回 | ≤ 1000 行；批量查 IN 的元素 ≤ 500（超出分批） |
| 事务时长 | ≤ 500ms；禁止在事务内做 RPC、HTTP、发消息（outbox 除外——它只写本地表） |
| 慢 SQL | > 200ms 记录并告警；每周 Review `performance_schema` Top 20 |
| 死锁 | 统一"按 id 升序"更新多行，降低死锁概率；捕获 1213 自动重试 1 次 |

### 7.4 分区、分表与归档

| 场景 | 策略 |
| --- | --- |
| 日志/埋点类大表（`stat_visit_log`、`article_view_log`、`third_api_log`、`sys_operation_log`） | **RANGE 分区**（按 `TO_DAYS(created_at)` 或 `created_at`），app/job 每日预建分区、删除过期分区 |
| 关系表（`user_follow`、`article_like`） | 数据量过大时按 `user_id` 哈希分 16 表（用中间件或应用层路由），**仅在单表 > 5000w 时才做** |
| 冷数据 | 文章 > 2 年且 PV 极低 → 迁移到 `article_archive`，主表保留摘要 |
| 删除策略 | 分区表优先 `DROP PARTITION`（秒级）；普通表批量删除必须分批（每批 ≤ 1000 行 + 间隔 50ms） |

### 7.5 缓存设计（Redis KV）

**Key 规范**

```
nv:{env}:{service}:{entity}:{id}[:{sub}]
  nv     项目前缀（避免与其它系统共用 Redis 时冲突）
  {env}  dev|test|prod（多环境共用实例时防窜）
  {service} user|article|comment|file|message|search|stat|third|system
  {entity} 业务实体（detail|list|count|lock|quota|captcha…）
  {id}   业务 ID / hash / 分页游标
```

**TTL 分级（统一在 `pkg/constants` 定义常量）**

| 等级 | TTL | 适用 |
| --- | --- | --- |
| L1 极短 | 5~60s | 热点列表首页、权限映射、站点配置 |
| L2 短 | 5~30min | 详情对象、正文、用户资料 |
| L3 长 | 1~24h | 关联 ID 列表、关注集合、秒传记录 |
| L4 超长 | 7~30d | 第三方结果（翻译/摘要）、IP 归属 |
| 会话类 | 按令牌有效期 | refresh token、黑名单 |

**统一要求**

1. **所有 TTL 必须加 ±10% 随机抖动**（`cache.Jitter(ttl)`），防雪崩。
2. **所有缓存 Value 必须带版本号**：写入 `{"v":1,"data":{...}}`，结构变更时通过版本号直接失效旧数据，避免灰度期解析错乱。
3. **缓存 Key 必须集中定义**（`app/<域>/model/cache/keys.go`），禁止散落字符串拼接。
4. **一致性策略**（本项目统一采用）：
   - **写路径**：先更新 DB（事务提交）→ 再删除缓存（del，不是 set）→ 延迟 500ms 再删一次（应对主从延迟导致的旧值回填）。
   - **读路径**：Cache-Aside（读缓存未命中 → 查库 → 回填）。
   - **不允许**在事务内直接操作缓存（回滚后缓存已变脏）。
   - 强一致场景（如余额、配额）**不缓存**，直读 DB。
5. **Redis 不做唯一数据源**：任何 Key 丢失都必须能从 DB 重建。
6. **禁止使用 `KEYS`**（用 `SCAN`）；批量操作用 pipeline 或 Lua。
7. **大 Key 限制**：单个 String ≤ 10KB；Hash/Set/ZSet 元素 ≤ 1000；超出拆 Key。**热 Key**（QPS > 5000）改用本地缓存 + 逻辑过期。

### 7.6 系统表（`noviterra_system`）

| 表 | 关键字段 |
| --- | --- |
| `role` | id, code, name, description, data_scope(self/dept/all), status, sort, created_at, updated_at |
| `permission` | id, parent_id, code, name, type(menu/button/api), path, component, icon, method, api_path, sort, status |
| `role_permission` | role_id, permission_id |
| `sys_config` | id, group_key, config_key, config_value, value_type(string/int/bool/json), is_public, remark, updated_by, updated_at — 唯一键 `uk_group_key(group_key,config_key)` |
| `sys_dict` | id, code, name, remark, status |
| `sys_dict_item` | id, dict_id, label, value, sort, status, extra |
| `sys_announcement` | id, title, content, type, target(all/role/user), status(draft/published/offline), start_at, end_at, is_top, created_by |
| `sys_operation_log` | id, user_id, username, module, action, method, path, query, body(脱敏), ip, region, ua, status, errcode, error, duration_ms, trace_id, created_at |
| `sys_job` | id, name, cron, handler, params(JSON), status(enabled/disabled), remark, last_run_at, next_run_at |
| `sys_job_log` | id, job_name, start_at, end_at, duration_ms, status(success/failed), result, error, trigger(manual/cron), created_at |

### 7.7 数据归属与跨服务查询原则

**三条铁律**

1. **只有属主服务能写自己的 schema**（`app/job` 也不能例外，它通过 RPC 触发）。
2. **跨服务查询只通过 RPC**，不接受"为了列表性能直接 JOIN 别人的表"。
3. **展示所需的跨域字段用冗余 + 事件同步**（如 `article.author_name`），并明确"允许最终一致"。

**冗余字段清单（明确登记，便于排查不一致）**

| 冗余字段 | 所在表 | 真源 | 同步方式 | 容忍延迟 |
| --- | --- | --- | --- | --- |
| `article.author_name` / `author_avatar` | article | user | `user.updated` 事件 | 5 分钟 |
| `article.category_name` | article（可选，列表用） | category（同库） | 同库 JOIN 或读时补 | 实时 |
| `article.tag_names` | article（可选 JSON） | tag（同库） | 同库 JOIN | 实时 |
| `article.like_count/collect_count/view_count/comment_count` | article | 各自明细表 | 计数缓冲 + 事件 + 每日对账 | 5 分钟 |
| `user_stat.article_count/follower_count` | user_stat | article / user_follow | 事件 + 每日对账 | 5 分钟 |
| `category.article_count` | category | article | 事件 + 每日对账 | 5 分钟 |
| `tag.article_count` | tag | article_tag | 事件 + 每日对账 | 5 分钟 |
| `file.ref_count` | file | file_ref | 引用登记/解除时同步 + 每周对账 | 实时 |

> **对账任务**（`reconcile_counters`）是保证冗余字段不失控的关键，必须实现并告警。

### 7.8 实体关系概览（文字版 ER）

```
user 1─n article           (user.id = article.user_id)
user 1─n comment
user 1─n file
user 1─1 user_auth (n 条：多登录方式)
user 1─n user_login_log
user n─n user (user_follow: user_id → follow_user_id)
user n─n role (user_role)
role n─n permission (role_permission)

article 1─1 article_content
article 1─n article_version
article n─n tag (article_tag)
article n─1 category
article n─n series (series_article)
article 1─n comment        (comment.article_id)
article 1─n article_like / article_collect
article 1─n article_translation (按 lang)

comment 1─n comment        (root_id / parent_id 自关联，最多 2 层)
comment 1─n comment_like / comment_report

file 1─n file_ref          (biz_type + biz_id 指向 article/comment/user)
message n─1 user           (to_user_id)
```

---

## 8. 接口规范

### 8.1 RESTful 约定

| 项 | 约定 | 示例 |
| --- | --- | --- |
| 资源命名 | 复数名词、小写、中划线分隔 | `/api/v1/articles`、`/api/v1/collect-folders` |
| 层级 | 最多 3 层 | `/api/v1/articles/{id}/comments` |
| 动作 | 用 HTTP 方法；非 CRUD 动作用子资源 + POST | `POST /articles/{id}/publish`、`POST /comments/{id}/report` |
| 批量操作 | `/batch/xxx` 或 body 传 ids | `PUT /articles/batch/status` |
| 查询参数 | 小驼峰（JSON 一致） | `?page=1&pageSize=10&categoryId=3&sort=latest` |
| 路径参数 | `{id}`（雪花 ID 或 slug） | `/articles/1001`、`/articles/slug/go-micro` |
| 时间格式 | 毫秒时间戳（响应）/ ISO8601 也接受（请求） | `publishedAt: 1735689600000` |
| 空值 | 显式 `null`，禁止用 `""` 代替 | — |
| 字段命名 | JSON 一律 **lowerCamelCase** | `viewCount`、`createdAt` |
| 枚举 | 返回数字 + 语义字符串成对 | `{"status": 1, "statusText": "published"}` |
| 幂等 | PUT/DELETE 必须幂等；POST 用 `Idempotency-Key` 头或业务幂等键 | — |

**标准请求头**

| Header | 必填 | 说明 |
| --- | --- | --- |
| `Authorization` | 视接口 | `Bearer <access_token>` |
| `Content-Type` | POST/PUT | `application/json`（上传用 `multipart/form-data`） |
| `X-Request-Id` | 否 | 客户端生成，服务端回显，用于排查 |
| `X-Client-Version` | 否 | 客户端版本（灰度与兼容判断） |
| `X-Device-Id` | 否 | 设备标识（多设备令牌管理） |
| `Idempotency-Key` | 写接口建议 | 防重复提交 |

**响应头**

| Header | 说明 |
| --- | --- |
| `X-Trace-Id` | 链路 ID，必须返回，便于用户报障时定位 |
| `X-RateLimit-Remaining` / `X-RateLimit-Reset` | 限流提示 |

### 8.2 统一响应与分页

**成功**

```json
{ "code": 0, "msg": "ok", "data": { "id": 1001, "title": "…" }, "trace_id": "a1b2c3", "server_time": 1735689600000 }
```

**失败**

```json
{ "code": 30201, "msg": "文章不存在或已删除", "data": null, "trace_id": "a1b2c3", "server_time": 1735689600000 }
```

**分页响应**

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "list": [ /* … */ ],
    "total": 1234,
    "page": 1,
    "pageSize": 10,
    "hasMore": true
  }
}
```

**游标分页响应**（评论、消息、时间线）

```json
{ "code": 0, "msg": "ok", "data": { "list": [], "nextCursor": "1000", "hasMore": true } }
```

**分页参数规则**：`page ≥ 1`（默认 1）、`pageSize ∈ [1,100]`（默认 10，超出报错而非静默截断）；`page > 1000` 时返回 `CodeDeepPagination`（提示改用游标或缩小范围）。

### 8.3 错误码规范

**分段规则（5 位数字）**

| 段位 | 归属 | 示例 |
| --- | --- | --- |
| `0` | 成功 | 0 |
| `10xxx` | 通用（参数、框架、鉴权框架） | 10001 参数错误 |
| `11xxx` | 网关（限流、熔断、路由、版本） | 11001 请求过于频繁 |
| `20xxx` | 用户服务 | 20101 用户不存在 |
| `21xxx` | 认证与令牌 | 21101 密码错误 |
| `22xxx` | 权限 RBAC | 22101 无权限 |
| `30xxx` | 内容（文章/分类/标签/专栏） | 30201 文章不存在 |
| `31xxx` | 审核 | 31101 审核不通过 |
| `40xxx` | 互动（评论/点赞/收藏/关注） | 40101 评论不存在 |
| `50xxx` | 文件与存储 | 50101 文件过大 |
| `60xxx` | 消息与通知 | 60101 验证码错误 |
| `70xxx` | 搜索 | 70101 搜索服务不可用 |
| `80xxx` | 第三方与外部依赖 | 80101 翻译服务不可用 |
| `90xxx` | 系统与中间件（DB/Redis/MQ/未知） | 90101 服务内部错误 |

**常用错误码（首批固化到 `pkg/errx/code.go`）**

| Code | HTTP | 常量 | 消息 |
| --- | --- | --- | --- |
| 0 | 200 | `CodeOK` | ok |
| 10001 | 200 | `CodeInvalidParam` | 参数错误 |
| 10002 | 200 | `CodeMissingParam` | 缺少必要参数 |
| 10003 | 200 | `CodeInvalidFormat` | 参数格式错误 |
| 10004 | 200 | `CodeNotFound` | 资源不存在 |
| 10005 | 200 | `CodeTooManyRequests` | 操作过于频繁 |
| 10006 | 200 | `CodeDeepPagination` | 分页过深，请使用游标分页 |
| 10007 | 200 | `CodeIdempotentConflict` | 重复提交 |
| 11001 | 429 | `CodeRateLimited` | 请求过于频繁，请稍后再试 |
| 11002 | 503 | `CodeServiceUnavailable` | 服务暂时不可用 |
| 11003 | 200 | `CodeCircuitOpen` | 依赖服务异常，请稍后重试 |
| 20101 | 200 | `CodeUserNotFound` | 用户不存在 |
| 20102 | 200 | `CodeUserExists` | 用户已存在 |
| 20103 | 200 | `CodeUsernameTaken` | 用户名已被占用 |
| 20104 | 200 | `CodeEmailTaken` | 邮箱已被注册 |
| 20105 | 200 | `CodeUserBanned` | 账号已被封禁 |
| 20106 | 200 | `CodeUserMuted` | 账号已被禁言 |
| 20107 | 200 | `CodeCannotFollowSelf` | 不能关注自己 |
| 21101 | 200 | `CodeWrongPassword` | 账号或密码错误 |
| 21102 | 401 | `CodeTokenExpired` | 登录已过期，请重新登录 |
| 21103 | 401 | `CodeTokenInvalid` | 登录状态无效 |
| 21104 | 401 | `CodeTokenRevoked` | 登录状态已失效，请重新登录 |
| 21105 | 200 | `CodeAccountLocked` | 账号已锁定，请 15 分钟后重试 |
| 21106 | 200 | `CodeCaptchaWrong` | 验证码错误 |
| 21107 | 200 | `CodeCaptchaExpired` | 验证码已过期 |
| 22101 | 403 | `CodeNoPermission` | 无权限执行该操作 |
| 22102 | 403 | `CodeDataScopeDenied` | 无权访问该数据 |
| 30101 | 200 | `CodeArticleNotFound` | 文章不存在或已删除 |
| 30102 | 200 | `CodeArticleNotPublished` | 文章未发布 |
| 30103 | 200 | `CodeArticleNoPermission` | 无权操作该文章 |
| 30104 | 200 | `CodeArticleStatusInvalid` | 当前状态不允许该操作 |
| 30105 | 200 | `CodeArticleTooLong` | 文章内容超出长度限制 |
| 30106 | 200 | `CodeSlugTaken` | 链接标识已被占用 |
| 30201 | 200 | `CodeCategoryNotFound` | 分类不存在 |
| 30202 | 200 | `CodeCategoryHasChildren` | 请先删除子分类 |
| 30203 | 200 | `CodeCategoryHasArticles` | 分类下存在文章，无法删除 |
| 30301 | 200 | `CodeTagNotFound` | 标签不存在 |
| 30302 | 200 | `CodeTagLimitExceeded` | 标签数量超出限制 |
| 30401 | 200 | `CodeAlreadyLiked` | 已点赞 |
| 30402 | 200 | `CodeAlreadyCollected` | 已收藏 |
| 31101 | 200 | `CodeReviewRejected` | 内容未通过审核 |
| 31102 | 200 | `CodeContentRisky` | 内容包含违规信息 |
| 40101 | 200 | `CodeCommentNotFound` | 评论不存在或已删除 |
| 40102 | 200 | `CodeCommentClosed` | 该文章已关闭评论 |
| 40103 | 200 | `CodeCommentTooFrequent` | 评论过于频繁 |
| 40104 | 200 | `CodeCommentContentInvalid` | 评论内容不合规 |
| 40105 | 200 | `CodeAlreadyReported` | 您已举报过该内容 |
| 50101 | 200 | `CodeFileTooLarge` | 文件超出大小限制 |
| 50102 | 200 | `CodeFileTypeNotAllowed` | 不支持的文件类型 |
| 50103 | 200 | `CodeFileQuotaExceeded` | 存储空间已满 |
| 50104 | 200 | `CodeUploadIncomplete` | 上传未完成，请重试 |
| 50201 | 200 | `CodeStorageUnavailable` | 存储服务暂时不可用 |
| 60101 | 200 | `CodeEmailCodeWrong` | 邮箱验证码错误 |
| 60102 | 200 | `CodeEmailCodeExpired` | 邮箱验证码已过期 |
| 60103 | 200 | `CodeEmailSendFailed` | 邮件发送失败 |
| 60201 | 200 | `CodeSmsSendFailed` | 短信发送失败 |
| 70101 | 503 | `CodeSearchUnavailable` | 搜索服务暂时不可用 |
| 80101 | 200 | `CodeTranslateUnavailable` | 翻译服务暂时不可用 |
| 80102 | 503 | `CodeThirdQuotaExceeded` | 第三方服务配额已用尽 |
| 80103 | 200 | `CodeThirdDegraded` | 服务已降级 |
| 90101 | 500 | `CodeInternalError` | 服务内部错误 |
| 90102 | 500 | `CodeDBError` | 数据服务异常 |
| 90103 | 500 | `CodeCacheError` | 缓存服务异常 |
| 90104 | 500 | `CodeConfigError` | 配置错误 |

**规则**

1. 错误码**全局唯一且不可复用**；已发布的码不得改语义（改语义必须换新码）。
2. 每个码必须有**中文默认消息**，禁止把 Go 原始 error 直接返回给用户。
3. 生产环境的 `msg` 只给用户友好提示；技术细节（SQL、堆栈、供应商原始错误）只进日志。
4. 前端通过 `code` 做统一处理（如 21102/21103/21104 → 跳登录）。
5. 新增错误码必须同步更新 `pkg/errx/code.go` 与前端 `error-code.ts`（脚本生成）。

### 8.4 gRPC / protobuf 规范

| 项 | 规范 |
| --- | --- |
| 包名 | `noviterra.{svc}.v1` |
| 文件位置 | `server/proto/{svc}/{svc}.proto` |
| 消息命名 | 请求 `XxxReq`，响应 `XxxResp`；实体 `Article`、`UserInfo` |
| 字段命名 | `snake_case`；字段编号只增不删；删字段用 `reserved` |
| 枚举 | 0 值必须是 `XXX_UNSPECIFIED` |
| 时间 | `int64` 毫秒时间戳（跨语言一致），禁止 `google.protobuf.Timestamp` 混用（团队统一用 int64） |
| 空值 | 用 `google.protobuf.StringValue` 等包装类型仅在**必须区分"未设置"**时使用；否则用默认值 + `update_mask` |
| 更新语义 | 局部更新用 `repeated string update_fields`（字段掩码），避免"零值覆盖"陷阱 |
| 大响应 | 分页返回；单条消息 ≤ 1MB |
| 错误 | 业务错误用 `status.Error(codes.Xxx, errx.JSON(err))`，客户端 `errx.ParseFromRPC` 还原 |
| 元数据 | 见 [2.5 上下文透传](#25-通信与协议约定) |
| 超时 | 客户端必须设置（默认 3s）；服务端不设置长于客户端的超时 |

**gRPC → HTTP 状态映射（`pkg/errx/map.go`）**

| gRPC Code | HTTP | 说明 |
| --- | --- | --- |
| OK | 200 | 成功 |
| InvalidArgument | 200（业务码 10001） | 参数错误 |
| Unauthenticated | 401 | 未登录 |
| PermissionDenied | 403 | 无权限 |
| NotFound | 200（业务码 10004） | 不存在 |
| ResourceExhausted | 429 | 限流 |
| DeadlineExceeded / Unavailable | 503 | 超时/不可用 |
| Internal / Unknown | 500 | 内部错误 |

### 8.5 版本与兼容

| 项 | 策略 |
| --- | --- |
| HTTP 版本 | 路径前缀 `/api/v1`；不兼容变更升 `/api/v2`，v1 至少保留 6 个月 |
| protobuf | 只增字段不改语义；服务端必须能处理旧客户端请求 |
| 兼容变更 | 新增可选字段、新增接口、放宽校验 |
| 不兼容变更 | 删除字段、改字段类型/语义、改错误码语义、收紧必填 |
| 灰度 | 新字段先只写不读 → 观察 → 再读；新旧服务并存期间**双写兼容** |
| 废弃 | 在响应头加 `Deprecation: true` + `Sunset` 日期；文档标记；提前 3 个月通知 |

### 8.6 幂等与并发控制

**防重复提交（三层防护）**

| 层 | 手段 | 适用 |
| --- | --- | --- |
| 前端 | 按钮 loading + 提交后禁用 | 所有表单 |
| 网关/服务 | `Idempotency-Key` 头：Redis SETNX 10 分钟，重复请求返回首次结果 | 创建类接口（发文章、评论、上传完成） |
| 数据库 | 唯一索引兜底 | 点赞、收藏、关注、注册 |

**并发控制**

| 场景 | 方案 |
| --- | --- |
| 计数更新（点赞数、浏览量） | Redis 原子 `INCR/HINCRBY` + 异步落库（不用行锁） |
| 库存/配额类 | DB 乐观锁（`version` 字段）或 `UPDATE ... WHERE remain >= n` |
| 文章编辑冲突 | 乐观锁：`UPDATE article SET ..., version = version+1 WHERE id=? AND version=?`，影响行数 0 → `CodeConflict` |
| 定时发布重复 | 乐观锁/行锁 + 状态机校验（`WHERE status='scheduled'`） |
| 跨服务长流程 | 分布式锁（`pkg/cache.Lock`）+ 幂等键 |
| 缓存重建 | singleflight |

---

## 9. 鉴权与安全

### 9.1 鉴权全流程

```
① 注册/登录 → user-rpc 签发 access_token(JWT, 2h) + refresh_token(30d, Redis)
② 请求 → 边缘层(Nginx/Ingress) 按路径转发 → 目标域 api
③ 域 api 中间件（pkg/middleware，各 api 复用同一实现）：
     解析 Bearer → jwt.Parse（验签、验过期、验算法白名单）
       → 查黑名单（Redis exists jti）
       → 校验 PermVer 与 Redis 中 user:permver:{uid} 一致
       → 注入 metadata: x-uid / x-username / x-roles / x-perm-ver / x-trace-id / x-client-ip
④ 域 rpc：拦截器从 metadata 还原到 context，logic 用 ctxx.Uid(ctx) 取用户
⑤ 权限判定：
     - 接口级：admin 中间件按 (method, path) → permCode 校验
     - 数据级：rpc domain 内判定归属（作者/管理员/协作者）
⑥ 刷新：/auth/refresh → 校验 refresh（Redis）→ 轮转（旧 token 立即失效）→ 新令牌对
⑦ 登出：access_token 的 jti 入黑名单；删除该设备 refresh
⑧ 强制下线：删除 refresh + PermVer++（所有旧 access 立即失效）
```

**令牌校验性能**：各 api 对 access_token 的验签是纯 CPU 操作（RS256 稍慢，可用公钥缓存）；黑名单查询每次打 Redis（1 次 GET，可 `MGET` 合并），可加本地缓存 5s（容忍 5s 延迟撤销）。
**代价与缓解**：鉴权逻辑分散在 N 个 api 进程 → 由 `pkg/middleware` 提供**唯一实现** + 统一配置模板（`etc/*-api.yaml` 的 `Jwt` 段落由模板生成）来消除漂移风险；CI 增加"各 api 中间件配置一致性"检查。

### 9.2 RBAC 权限模型

```
用户 ──n:n── 角色 ──n:n── 权限
      user_role       role_permission
权限类型：menu（菜单/路由）| button（页面按钮）| api（接口，兜底）
权限码格式：{domain}:{resource}[:{action}]，如 article:review、user:ban、system:config
```

**鉴权位置分工**

| 层 | 校验内容 |
| --- | --- |
| 各域 api | 只校验"是否登录"并注入身份（前台接口） |
| `admin` | 登录 + 接口权限码（默认拒绝） |
| 各域 rpc | 数据级权限（是否本人/协作者；后台管理员语义下放行） + 敏感操作二次校验 |

**关键设计**：`rpc` **不能假设调用方已鉴权**——被内部服务调用时也要做数据归属校验；只有持有 `x-internal-token` 且声明为 system 级操作（如封禁联动）才跳过。

### 9.3 越权防护（IDOR）

| 风险 | 防护 |
| --- | --- |
| 水平越权（改他人文章） | 每个写操作必须校验 `resource.owner_id == ctx.uid || isAdmin` |
| 垂直越权（普通用户调后台接口） | `/admin/**` 独立网关 + 权限码校验 |
| 遍历 ID | 雪花 ID 不可枚举；关键资源额外校验归属；敏感资源列表加限流 |
| 私有文章泄露 | 详情接口按 `visibility` 与关注关系裁剪；`private` 仅作者；缓存 Key 不含用户维度但**响应按用户裁剪**（或私有内容不进共享缓存） |
| 文件越权下载 | 私有文件必须签名 URL，且签名时校验归属/引用关系 |
| 消息越权 | `message.to_user_id == ctx.uid` 强校验（改 id 也取不到别人的） |
| 批量接口越权 | 批量操作逐个校验归属，不要只校验第一个 |

### 9.4 攻击防护矩阵

| 攻击 | 防护措施 | 落地位置 |
| --- | --- | --- |
| SQL 注入 | 全量参数化（GORM）；禁止字符串拼 SQL；排序字段白名单 | `pkg/db`、code review |
| XSS | 后端转义/白名单净化 + 前端 DOMPurify；CSP 头；Cookie HttpOnly | comment/article 渲染、Nginx |
| CSRF | Bearer Token（非 Cookie）天然免疫；仍校验 `Origin`/`Referer`（写接口）；不用 Cookie 存令牌 | 各域 api + 边缘层 |
| SSRF | 禁止用户提供 URL 让服务端请求；如需（如外链图片抓取）必须**域名白名单 + 禁内网 IP + 超时** | file-rpc / third-rpc |
| 文件上传漏洞 | 扩展名 + MIME + 魔数三重校验；重命名；图片二次编码；SVG 禁用/净化；存储与站点域名分离 | file-rpc |
| 路径穿越 | 对象 Key 不含用户输入；`filepath.Clean` + 前缀校验 | file-rpc |
| 暴力破解 | 登录失败锁定 + 图形验证码 + 限流 + 异常 IP 告警 | user-rpc |
| 短信/邮件轰炸 | 严格限流（60s/次、10 次/天/目标）+ 图形验证码前置 + 黑名单 | user-rpc / third-rpc |
| 刷量（点赞/浏览） | 按 uid+IP 限流；浏览 5 分钟去重；异常增长告警 | article-rpc |
| 爬虫 | UA 黑名单 + IP 频率限制 + 关键数据接口要求登录；提供 RSS/sitemap 作为替代 | 边缘层 + 各域 api |
| 敏感信息泄露 | 错误响应不含内部细节；生产关 `pprof`/`debug` 对外；`.gitignore` 排除密钥；日志脱敏 | 全局 |
| 越权访问中间件 | Redis/MySQL 强密码 + 仅内网监听 + 不在公网暴露 6379/3306/2379 | deploy |
| 依赖漏洞 | CI 跑 `govulncheck` + Dependabot；禁止引入无维护库 | CI |
| 弱密码 | 密码策略：≥ 8 位，含字母 + 数字；常见弱密码黑名单 | user-rpc |
| JWT 攻击 | 算法白名单、密钥强度、禁止敏感信息入 Claims | pkg/jwt |
| 内容合规 | 本地敏感词 + 云审核 + 人工复核队列 + 举报机制 | comment-rpc / article-rpc |

**安全响应头（Nginx / 网关统一加）**

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; img-src 'self' data: https:; script-src 'self'
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

### 9.5 数据合规与隐私

| 项 | 要求 |
| --- | --- |
| 密码 | bcrypt 哈希，永不明文；不可逆；日志永不打印 |
| 手机号/邮箱 | 展示脱敏（`138****8888`、`ab***@mail.com`）；日志脱敏 |
| IP 归属地 | 仅存归属地文本 + 脱敏 IP（可按策略存完整 IP 但限管理员可见），满足合规展示要求 |
| 数据导出 | 仅管理员 + 操作日志留痕；导出走异步任务，生成带水印文件 |
| 账号注销 | 支持注销；30 天冷静期；之后匿名化（保留文章，作者显示"已注销用户"） |
| 未成年人/敏感内容 | 依赖内容审核与举报机制；留存处置记录 |
| 备份 | 每日全量 + binlog 增量；备份加密；每季度恢复演练 |
| 审计留痕 | 后台所有写操作记录 `sys_operation_log`，保留 12 个月 |

---

## 10. 可观测性与运维

### 10.1 日志规范

| 项 | 规范 |
| --- | --- |
| 输出 | dev: console（彩色、易读）；test/prod: JSON 单行 |
| 目标 | stdout（容器标准做法），由采集器（Filebeat/Promtail）收集；**不在容器内写文件** |
| 级别 | `debug`（仅 dev）、`info`（关键业务事件）、`warn`（可恢复异常、降级）、`error`（需人工介入）、`fatal`（启动失败） |
| 必带字段 | `ts`、`level`、`service`、`env`、`trace_id`、`msg` |
| 业务事件 | 必须结构化（如 `logx.ArticleId(id)`），禁止拼字符串 |
| 采样 | 高频 info 日志（如访问日志）按 1/10 采样；error 不采样 |
| 切割 | 若必须落文件，用 lumberjack（100MB × 10 份 × 30 天） |
| 检索 | prod 用 Loki 或 ES，按 `trace_id`/`uid`/`errcode` 可查 |

**日志分级示例**

```go
log.Debug("cache miss", logx.Key(key))                                     // dev 调试
log.Info("article published", logx.ArticleId(id), logx.Uid(uid))            // 业务里程碑
log.Warn("interceptor party degraded", logx.Provider("baidu"), logx.Err(err))     // 降级
log.Error("create article failed", logx.Err(err), logx.Uid(uid))            // 需处理
```

### 10.2 指标（Prometheus）

**命名规范**：`{namespace}_{subsystem}_{name}_{unit}`，namespace = `noviterra`。

| 指标 | 类型 | 标签 | 说明 |
| --- | --- | --- | --- |
| `noviterra_http_requests_total` | Counter | service, method, path, status | 请求量 |
| `noviterra_http_request_duration_seconds` | Histogram | service, method, path | 延迟（P50/P95/P99） |
| `noviterra_rpc_requests_total` | Counter | service, method, result | RPC 调用 |
| `noviterra_rpc_request_duration_seconds` | Histogram | service, method | RPC 延迟 |
| `noviterra_db_query_duration_seconds` | Histogram | service, table, op | SQL 延迟 |
| `noviterra_db_errors_total` | Counter | service, table, errtype | SQL 错误 |
| `noviterra_cache_hits_total` / `_misses_total` | Counter | service, key_prefix | 缓存命中率 |
| `noviterra_cache_op_duration_seconds` | Histogram | service, op | Redis 延迟 |
| `noviterra_mq_published_total` / `_consumed_total` / `_failed_total` | Counter | topic, service | 消息 |
| `noviterra_thirdparty_requests_total` | Counter | provider, capability, result | 第三方调用 |
| `noviterra_thirdparty_duration_seconds` | Histogram | provider, capability | 第三方延迟 |
| `noviterra_thirdparty_quota_remaining` | Gauge | provider, capability, period | 配额余量 |
| `noviterra_business_articles_published_total` | Counter | — | 业务指标 |
| `noviterra_business_users_registered_total` | Counter | source | 业务指标 |
| `noviterra_goroutines` / `_gc_duration` | Gauge | service | 运行时（go-zero 自带部分） |

**核心告警规则（Alertmanager）**

| 告警 | 条件 | 级别 |
| --- | --- | --- |
| 服务不可用 | `up == 0` 持续 1 分钟 | P0 |
| 错误率过高 | 5xx 或 `errcode >= 90000` 占比 > 1%（5 分钟） | P0 |
| P99 延迟过高 | HTTP P99 > 1s 持续 5 分钟 | P1 |
| RPC 失败率高 | 失败率 > 5%（5 分钟） | P1 |
| 熔断打开 | breaker open 事件 | P1 |
| DB 连接池耗尽 | 使用率 > 90% 持续 2 分钟 | P0 |
| 慢 SQL 激增 | 慢查询 > 10/分钟 | P1 |
| Redis 命中率下降 | 命中率 < 70% 持续 10 分钟 | P2 |
| 消息积压 | 队列长度 > 1w 或最老消息 > 10 分钟 | P1 |
| 第三方失败率 | > 20%（5 分钟） | P1 |
| 第三方配额 | 用量 > 80% / 95% | P2 / P1 |
| 磁盘空间 | > 85% | P1 |
| 证书过期 | < 14 天 | P1 |
| 定时任务失败 | 连续 2 次失败 | P1 |

### 10.3 链路追踪

- 采用 **OpenTelemetry**（`go.opentelemetry.io/otel`）+ OTLP exporter → Jaeger/Tempo。
- go-zero 已内置 trace 支持，通过拦截器串联：HTTP → zrpc → GORM → Redis → 外部 HTTP。
- **采样策略**：dev 100%；prod 头部采样 10% + **错误请求全采** + **慢请求（> 1s）全采**。
- **必须埋点的 span**：HTTP 请求、RPC 调用、DB 查询（含表名，不含 SQL 参数）、Redis 操作、第三方调用、MQ 生产/消费。
- `trace_id` 全链路透传：响应头返回给前端 → 前端报错时带上 → 用户报障可精确定位。

### 10.4 健康检查与优雅停机

| 端点 | 用途 | 内容 |
| --- | --- | --- |
| `/healthz`（liveness） | 进程存活 | 恒返回 200（除非死锁） |
| `/readyz`（readiness） | 是否可接流量 | 检查 DB ping、Redis ping、关键依赖（超时 500ms，超时判定为未就绪） |
| `/metrics` | Prometheus 抓取 | 仅内网 |
| `/version` | 版本信息 | version/commit/buildTime |
| `/debug/pprof` | 性能分析 | **仅 dev/test**，prod 关闭或加内网 + 鉴权 |

**优雅停机顺序**（`main` 中注册 `signal.Notify` → `SIGTERM`）：

```
1. 停止接收新请求（从 etcd 注销服务 / readiness 置 false）
2. 等待在途请求完成（最多 15s，超时强制）
3. 停止 MQ 消费者（等在途消息处理完）
4. 停止定时任务（等当前任务结束或超时）
5. 关闭 DB 连接池、Redis 连接（flush 未落库的计数缓冲！）
6. logx.Sync() → 退出
```

> **注意**：优雅停机必须 flush 缓冲（浏览计数、批量埋点、WebSocket 推送队列），否则丢数据。K8s 中 `terminationGracePeriodSeconds` 要 > 停机耗时（建议 30s），并配置 `preStop` sleep 5s。

---

## 11. 部署与交付

### 11.1 本地开发环境

`server/deploy/compose/docker-compose.infra.yml` 提供全部中间件：

| 组件 | 镜像 | 端口 | 用途 |
| --- | --- | --- | --- |
| MySQL | `mysql:8.0` | 3306 | 主库 |
| Redis | `redis:7-alpine` | 6379 | 缓存/锁/计数 |
| Etcd | `quay.io/coreos/etcd:v3.5` | 2379/2380 | 服务发现 + 配置中心 |
| MinIO | `minio/minio` | 9000/9001 | 对象存储 |
| Meilisearch | `getmeili/meilisearch` | 7700 | 全文检索 |
| Jaeger | `jaegertracing/all-in-one` | 16686/4317 | 链路追踪 |
| Prometheus | `prom/prometheus` | 9090 | 指标 |
| Grafana | `grafana/grafana` | 3000 | 看板 |
| MailHog | `mailhog/mailhog` | 1025/8025 | 本地收邮件（不吃真实配额） |
| Redis Commander / phpMyAdmin（可选） | — | 8081/8080 | 可视化调试 |

**本地启动流程（写入 README）**

```bash
# 1. 起中间件（含边缘层 Nginx）
docker compose -f server/deploy/compose/docker-compose.infra.yml up -d
# 2. 初始化数据库（建 schema + 迁移 + seed）
make -C server migrate-up && make -C server seed
# 3. 起服务（按依赖顺序；开发期用合并部署，每域一个进程）
make -C server run SVC=user        # 同时开 user-api(8801) + user-rpc(9001)
make -C server run SVC=article     # 同时开 article-api(8803) + article-rpc(9003)
make -C server run SVC=admin
# 4. 验证（经边缘层，端口 80）
curl http://127.0.0.1/healthz
curl http://127.0.0.1/api/v1/stats/site
```

**约定**：所有第三方在 dev 环境使用 `mock` 或本地替代（MailHog 代替 SMTP、Meilisearch 代替 ES、MinIO 代替 OSS、mock 实现代替付费 API），**保证新人 10 分钟能跑通全链路且零外部成本**。

### 11.2 配置分层

```
优先级（高 → 低）：
1. 环境变量（NV_*）              ← 密钥、环境相关
2. Etcd 配置中心（test/prod）    ← 可热更新项
3. etc/<svc>.yaml                ← 结构与非敏感默认值
4. etc/<svc>.local.yaml          ← 个人本地覆盖（gitignore）
5. 代码内默认值                   ← 兜底（必须合理）
```

**配置项命名**：YAML 用大驼峰分组（go-zero 风格：`UserRpc`、`DB`、`Redis`），环境变量统一 `NV_` 前缀 + 大写下划线（`NV_DB_DSN`）。

**启动自检**：`pkg/config` 提供 `Validate()`，检查必填项、端口范围、DSN 可解析、密钥非空；失败直接退出并打印明确缺失项。

### 11.3 镜像与构建

```dockerfile
# server/deploy/docker/Dockerfile（通用多阶段）
# 用法：SERVICE=<域> MODE=merged|api|rpc
#   merged → 编译 app/<域>/<域>.go（合并部署入口，同时开 HTTP + gRPC）
#   api    → 编译 app/<域>/api/<域>.go
#   rpc    → 编译 app/<域>/rpc/<域>.go
FROM golang:1.25-alpine AS builder
ARG SERVICE
ARG MODE=merged
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# 依据 MODE 选择入口文件（用 shell 拼路径，避免 Dockerfile 条件语法）
RUN set -e; \
    case "$MODE" in \
      merged) ENTRY="./server/app/${SERVICE}"; ;; \
      api)    ENTRY="./server/app/${SERVICE}/api"; ;; \
      rpc)    ENTRY="./server/app/${SERVICE}/rpc"; ;; \
      *) echo "bad MODE"; exit 1; ;; \
    esac; \
    CGO_ENABLED=0 GOOS=linux go build -trimpath \
      -ldflags "-s -w -X noviterra/server/pkg/version.Version=${VERSION} \
                -X noviterra/server/pkg/version.Commit=${COMMIT} \
                -X noviterra/server/pkg/version.BuildTime=${BUILD_TIME}" \
      -o /out/app "$ENTRY"

FROM alpine:3.20
ARG SERVICE
ARG MODE=merged
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=builder /out/app /app/app
# 配置目录随部署形态不同：merged 用 <域>/etc，独立部署用 <域>/{api,rpc}/etc
COPY --from=builder /src/server/app/${SERVICE} /app/conf-src
EXPOSE 8080
ENTRYPOINT ["/app/app", "-f", "/app/etc/config.yaml"]
```

> 各部署单元的 `etc` 目录在 K8s 中由 ConfigMap 挂载，镜像内只保留不含密钥的骨架配置。

| 规范 | 要求 |
| --- | --- |
| 基础镜像 | `alpine` 或 `gcr.io/distroless/static`；**禁止 latest 标签** |
| 镜像标签 | `{service}:{git-sha}` + `{service}:{semver}`；禁止只用 `latest` |
| 大小 | 单服务 < 50MB（静态编译 + strip） |
| 运行用户 | 非 root（uid 10001） |
| 时区 | 容器 TZ=UTC（业务层做本地化） |
| 配置 | 不打进镜像的敏感配置 → ConfigMap/Secret 挂载 |
| 健康检查 | `HEALTHCHECK` 或 K8s probe |

### 11.4 部署形态：先合并、后拆分

> 这是"域自治目录 + 可控进程数"之间的关键折中。**目录结构从第一天就按 `{api,rpc,model}` 建好，但进程可以晚点再拆。**

**背景**：若每个域都把 `api` 与 `rpc` 部署成独立进程，`portal` + `admin` + `job` + 9 个域 × 2 = **21 个进程**。在开发机与早期生产环境上，这是不必要的负担。

**两种形态对比**

| 形态 | 进程数 | 适用 | 代价 |
| --- | --- | --- | --- |
| **A. 合并部署**（每域 1 进程同时开 HTTP + gRPC） | 12 个（9 域 + portal + admin + job） | 阶段 2~6（开发、内测、早期生产） | api 与 rpc 无法独立扩缩容；一个进程内 HTTP 故障可能影响 gRPC；`main` 需手写（goctl 不生成双 server 的入口） |
| **B. 独立部署**（api / rpc 各自成进程） | 21 个 | 阶段 7 之后，或某个域出现明显的读放大需要独立扩容 | 进程数、配置、探针翻倍 |

**合并部署的实现要点**

```go
// server/app/article/article.go —— 手写入口：同一进程同时启动 gRPC 与 HTTP
func main() {
    var c config.Config
    conf.MustLoad(*configFile, &c)
    ctx := svc.NewServiceContext(c)

    // ① 先起 gRPC（供本域 api 与他域 rpc 调用）
    rpcSrv := zrpc.MustNewServer(c.RpcServerConf, func(s *grpc.Server) {
        article.RegisterArticleServer(s, server.NewArticleServer(ctx))
    })
    defer rpcSrv.Stop()
    go rpcSrv.Start()

    // ② 再起 HTTP（对外）
    httpSrv := rest.MustNewServer(c.RestConf)
    defer httpSrv.Stop()
    handler.RegisterHandlers(httpSrv, ctx)
    httpSrv.Start()
}
```

**合并态下 api 如何调用本域 rpc**：仍走 zrpc client，但配置**直连**而非 etcd 发现，避免回环走 etcd：

```yaml
# app/article/api/etc/article-api.yaml（合并部署）
ArticleRpc:
  Endpoints:
    - 127.0.0.1:9003        # 同进程内的 gRPC 端口
  NonBlock: true
  Timeout: 3000
```

**拆分步骤（零目录改动）**

1. 把 `app/article/article.go` 拆成 `api/article.go` 与 `rpc/article.go` 两个入口（就是 goctl 原本生成的样子）。
2. `api/etc/*.yaml` 的 `ArticleRpc` 从 `Endpoints` 改为 `Etcd`，指向 `article.rpc`。
3. 加一个 Deployment/Service（或 Compose 服务），镜像用同一个 Dockerfile + 不同的 `SERVICE` 参数。
4. `pkg/svc`、`logic`、`model` **一行都不用改**。

**建议时间点**：阶段 6 结束前用合并形态；阶段 7/8 做压测时，对"读放大最明显"的域（通常是 `article` 与 `search`）先拆，其余保持合并。

**K8s 注意**：合并态下一个 Pod 暴露两个端口，需要两个 Service 指向同一组 Pod：
- `article-api`（ClusterIP，80 → 8803）对外；
- `article-rpc`（ClusterIP，9003）仅内网，配合 NetworkPolicy 限制来源。
拆分后改为各自独立的 Deployment + Service。

### 11.5 Kubernetes 部署

```
server/deploy/k8s/
├── base/
│   ├── kustomization.yaml
│   ├── deployment.yaml.tpl      # 通用模板：副本、资源、探针、亲和性
│   ├── service.yaml.tpl
│   └── ingress.yaml
└── overlays/
    ├── dev/    (1 副本, 资源小, 镜像 tag=dev)
    ├── test/   (2 副本)
    └── prod/   (3+ 副本, HPA, PDB, 反亲和)
```

**关键配置项**

| 项 | 建议值 | 说明 |
| --- | --- | --- |
| `resources.requests` | cpu 200m / mem 256Mi | 按实测调整 |
| `resources.limits` | cpu 1000m / mem 512Mi | Go 服务注意 GOMEMLIMIT |
| `GOMAXPROCS` | 用 `uber-go/automaxprocs` 自适应 | 避免 CPU 限额下的调度抖动 |
| `GOMEMLIMIT` | limits × 0.8 | 降低 OOM 风险 |
| `livenessProbe` | `/healthz` 10s 间隔，3 次失败 | 不查依赖 |
| `readinessProbe` | `/readyz` 5s 间隔 | 查依赖 |
| `startupProbe` | `/healthz`，最长 60s | 避免启动慢被杀 |
| `terminationGracePeriodSeconds` | 30 | 配合优雅停机 |
| `preStop` | `sleep 5` | 等待 Endpoint 摘除生效 |
| `HPA` | CPU 70% 或 QPS 自定义指标，2~10 副本 | 网关与 article 服务必须配 |
| `PodDisruptionBudget` | minAvailable: 1 | 避免滚动更新全挂 |
| `NetworkPolicy` | 只允许网关 → 领域服务；领域服务 → 中间件 | 缩小攻击面 |
| 发布策略 | RollingUpdate（maxSurge 1 / maxUnavailable 0） | 零停机；重要版本用金丝雀 |

### 11.6 CI/CD 流水线

**CI（每个 PR 必跑，`server/**` 路径触发）**

```
1. gofmt / goimports 检查
2. go vet ./...
3. golangci-lint run（配置见 .golangci.yml，含 gosec、forbidigo、errcheck、gocritic）
4. go test ./... -race -coverprofile（覆盖率门槛：pkg ≥ 70%，app ≥ 50%）
5. govulncheck ./...（依赖漏洞）
6. go build ./server/...（编译验证）
7. proto/api 一致性检查（proto 改动时重新生成并 diff，不一致则失败）
8. 数据库迁移 dry-run（对空库执行全部迁移）
```

**CD（打 tag 或合并 main）**

```
1. 构建镜像（矩阵：所有服务，并行）
2. 打标签并推送到镜像仓库
3. 部署到 test 环境 → 跑冒烟测试（e2e 脚本）
4. 人工审批
5. 部署到 prod（按依赖顺序：user/system → article/comment/file → search/message/stat/third → portal/admin/job；边缘层路由最后切）
6. 部署后健康检查 + 关键接口拨测
7. 失败自动回滚（保留上一个镜像 tag）
```

**分支与发布**

| 分支 | 用途 |
| --- | --- |
| `main` | 稳定，随时可发布 |
| `develop`（可选） | 集成分支 |
| `feature/{issue}-{desc}` | 功能开发 |
| `fix/{issue}-{desc}` | 修复 |
| `release/v1.2.0`（可选） | 发布准备 |
| tag `v1.2.0` | 生产发布点，触发 CD |

### 11.7 Makefile 与脚本

**根 `Makefile`（聚合）**

```make
.PHONY: help infra-up infra-down server-build web-dev migrate seed lint test
help:            ## 显示所有命令
infra-up:        ## 起本地中间件
infra-down:      ## 停本地中间件
server-build:    ## 构建全部后端服务
web-dev:         ## 起前端开发服务器
migrate:         ## 执行数据库迁移
seed:            ## 灌入初始化数据
lint:            ## 前后端静态检查
test:            ## 运行全部测试
```

**`server/Makefile`（后端）**

```make
DOMAINS := user system article comment file message search stat third
ORCH    := portal admin
SCHED   := job
MODE    ?= merged          # merged | api | rpc —— 见 11.4

gen-api:         ## 生成 HTTP 骨架： goctl api go -api app/article/api/article.api -dir app/article/api
gen-rpc:         ## 生成 gRPC 骨架：goctl rpc protoc app/article/rpc/article.proto --zrpc_out=app/article/rpc
gen-model:       ## 生成 GORM 实体骨架（自定义模板，仅作参考）
gen-doc:         ## 导出接口文档到 doc/api
build:           ## 编译全部部署单元到 bin/（按 MODE 决定入口）
run:             ## 运行一个域： make run SVC=article MODE=merged
migrate-up/down/status
seed:
test / test-cover / bench
lint / fmt / check（提交前总检查）
docker-build:    ## 构建全部镜像（SERVICE=<域> MODE=<形态>）
smoke:           ## 冒烟测试（经边缘层打各域 api）
```

> **重要**：`make check` 是提交前必须通过的关卡（`fmt + vet + lint + test`），并在 `scripts/` 提供 git pre-commit hook 安装脚本（`make hooks`）。

---

## 12. 测试策略与质量门禁

### 12.1 测试分层

| 层级 | 范围 | 工具 | 覆盖率目标 | 运行时机 |
| --- | --- | --- | --- | --- |
| 单元测试 | 函数/方法、domain 规则、assembler、校验 | `testing` + `testify` | pkg ≥ 70%，app ≥ 50% | 每次提交 |
| 仓储测试 | repository 的 SQL 正确性 | `go-sqlmock` 或 `testcontainers` + 真实 MySQL | 关键仓储覆盖 | 每次提交 |
| 集成测试 | 服务内 logic + DB + Redis | `testcontainers-go` 起真实依赖 | 核心用例 | PR |
| 契约测试 | proto/API 兼容性 | 生成产物 diff、接口冒烟 | — | PR |
| 端到端测试 | 经边缘路由走各域 api 的完整链路 | Go + `httpexpect` / Postman/newman | 主流程 100% | 部署 test 后 |
| 压测 | 关键接口性能 | `k6` / `wrk` / `go test -bench` | P99 达标 | 版本发布前 |
| 混沌/故障测试 | 依赖宕机、超时、慢查询 | 手工或 Chaos Mesh | 降级生效 | 季度 |

### 12.2 必须覆盖的测试用例（最小集）

1. **认证**：注册（重复用户名/邮箱）、登录（密码错、锁定、验证码）、刷新（轮转、旧 token 复用检测）、登出、权限变更后旧令牌失效。
2. **文章**：状态机全路径（含非法流转拒绝）、slug 冲突、定时发布、版本回滚、可见性（private/password/fans）、越权编辑被拒。
3. **评论**：敏感词三级行为、频率限制、楼中楼结构、删除后回复保留、XSS 内容转义。
4. **文件**：类型/大小/MIME 校验、秒传命中、分片断点续传、越权下载私有文件被拒。
5. **第三方**：超时/500/配额超限下的重试、熔断、降级路径（用 mock 注入）。
6. **缓存**：缓存命中/未命中/空值占位/更新后失效。
7. **幂等**：重复提交同一 `Idempotency-Key` 只生效一次。
8. **并发**：并发点赞计数一致、并发发布同一文章只成功一次。

### 12.3 质量门禁（PR 合并条件）

| 门禁 | 要求 |
| --- | --- |
| 编译 | 必须通过 |
| 静态检查 | `golangci-lint` 0 error（warning 需说明） |
| 测试 | 全部通过，新增代码有测试 |
| 覆盖率 | 不得低于上一版本（CI 记录基线） |
| 接口变更 | 改了 `.api`/`.proto` 必须同步生成产物并更新 `doc/api` |
| 迁移变更 | 必须提供 `up` 与 `down`，且经过一次真实执行验证 |
| 依赖变更 | 新增依赖需在 PR 描述说明理由 |
| 安全 | `govulncheck` 无高危 |
| 文档 | 涉及架构决策的必须加 ADR；接口变更更新文档 |
| Review | 至少 1 人通过（核心模块 2 人） |

---

## 13. 工程规范

### 13.1 代码风格

| 项 | 规范 |
| --- | --- |
| 格式化 | `gofmt` + `goimports`（分组：标准库 / 第三方 / 本项目） |
| 命名 | 包名小写单词；导出符号大驼峰；缩写保持一致（`ID` 而非 `Id` 在导出名中，但 JSON tag 用 `id`/`userId`） |
| 错误 | 用 `%w` 包装保留链；错误信息小写、无标点结尾、不含敏感信息 |
| 注释 | 导出符号必须有注释，以符号名开头；复杂逻辑必须写"为什么"而非"是什么" |
| 常量 | 禁止裸魔法数字（用 `pkg/constants` 或服务内 `consts`） |
| 接口 | 定义在使用方（consumer side），不要为了"以后可能换实现"而提前抽象 |
| 泛型 | 工具函数可用；业务代码慎用（可读性优先） |
| 并发 | `goroutine` 必须有退出路径（ctx）；禁止裸 `go func()` 无 recover |
| 时间 | 一律 `time.Now().UTC()`；时长用 `time.Duration` |
| 指针 | 大结构体用指针；小结构体值传递；切片/Map 注意 nil 与空切片区别 |

### 13.2 Git 工作流

**Commit 规范（Conventional Commits）**

```
<type>(<scope>): <subject>

[body]

[footer]
```

| type | 含义 |
| --- | --- |
| `feat` | 新功能 |
| `fix` | 修复 |
| `docs` | 文档 |
| `style` | 格式（不影响逻辑） |
| `refactor` | 重构 |
| `perf` | 性能优化 |
| `test` | 测试 |
| `build` | 构建/依赖 |
| `ci` | CI 配置 |
| `chore` | 杂项 |
| `revert` | 回滚 |

`scope` 用服务名或包名：`feat(article): 支持定时发布`、`fix(user): 修复刷新令牌复用检测失效`。

**分支命名**：`feature/123-article-schedule`、`fix/456-token-refresh`。

**PR 要求**：标题同 commit 规范；描述含"改了什么/为什么/怎么验证/影响范围/回滚方式"；关联 issue；截图或 curl 输出作为验证证据；**单 PR 变更 ≤ 800 行**（超过请拆分）。

### 13.3 Code Review 检查清单（Reviewer 必查）

- [ ] 是否违反 [0.3 依赖方向铁律](#03-依赖方向铁律review-时必须检查)与 [0.4 十不准](#04-分层职责与十不准)？
- [ ] logic 层是否出现 SQL / Redis / 裸 HTTP？
- [ ] 是否跨服务直连了别人的库表或 Redis Key？
- [ ] 新增/修改的接口是否先改了 `.api`/`.proto`？
- [ ] 错误是否经 `errx` 归一化？是否有吞错？
- [ ] 是否有越权风险（写操作是否校验归属）？
- [ ] 是否有 N+1 查询、循环内 RPC、大偏移分页、`SELECT *` 拖大字段？
- [ ] 缓存是否设置 TTL + 抖动？写后是否失效？是否加了防穿透/击穿？
- [ ] 是否有并发问题（计数、幂等、重复提交）？
- [ ] 日志是否脱敏？是否包含 trace_id？
- [ ] 是否加了必要的测试？
- [ ] 是否有配置项/密钥硬编码？
- [ ] 是否影响兼容性（接口、proto、DB 迁移）？
- [ ] 是否有明确的超时与降级？

### 13.4 文档规范

| 文档 | 位置 | 更新时机 |
| --- | --- | --- |
| 架构设计（本文） | `doc/Develop.md` | 架构变更时同步修改并升版本 |
| 架构决策记录 ADR | `doc/adr/NNNN-title.md` | 每个重大技术决策（选型、协议、引入新中间件） |
| 接口文档 | `doc/api/`（goctl 导出） | 接口变更时重新生成 |
| 数据库设计 | `doc/db/`（表说明 + ER 图） | 迁移变更时 |
| 部署手册 | `doc/deploy.md` | 环境/流程变更时 |
| 运维手册 | `doc/ops.md`（含告警处理、故障预案） | 每次故障复盘后 |
| 故障复盘 | `doc/postmortem/` | 每次 P0/P1 故障 |

**ADR 模板**

```markdown
# NNNN. 决策标题
- 状态：提议 / 已接受 / 已废弃
- 日期：YYYY-MM-DD
- 决策者：
## 背景与问题
## 备选方案（至少 2 个，含优缺点）
## 决策
## 理由
## 影响与代价
## 后续跟进
```

---

## 14. 实施路线图与里程碑

> **总原则**：每个阶段结束时都必须"可运行、可验证"，禁止长期处于"半成品不可跑"状态。
> 时间估算按 **1 名全职后端 + 1 名前端（可兼职）** 计算，仅供参考。

### 阶段 0：仓库与工程地基（1~2 周）

| 项 | 内容 |
| --- | --- |
| 目标 | 仓库结构成型，本地基础设施一键起，CI 能跑通空项目 |
| 交付物 | 根目录工程文件；`server/` 目录骨架；`docker-compose.infra.yml`；`Makefile`；`.golangci.yml`；CI 流水线；README 快速开始 |
| 验收 | 新人按 README 在 10 分钟内起好中间件；`make server-build` 通过；CI 绿灯；`doc/Develop.md` 已评审通过 |
| 依赖 | 无 |
| 风险 | module 名未定（见 [附录 F](#附录-f待决策问题open-questions)）——**必须在本阶段解决** |

### 阶段 1：公共包 `pkg`（2 周）

| 项 | 内容 |
| --- | --- |
| 目标 | 所有服务的地基可用，且每个包有测试与示例 |
| 交付物 | `logx`、`errx`、`response`、`db`、`cache`、`jwt`、`crypt`、`idgen`、`paginate`、`conv`、`ctxx`、`validatorx`、`constants`、`middleware`、`observability`、`version` |
| 验收 | 每个包有单测（覆盖 ≥ 70%）；一个"样板服务"（可临时）能连通 DB + Redis + 日志 + 错误码 + 健康检查；`errx` 错误码表评审通过 |
| 依赖 | 阶段 0 |
| 风险 | 过早抽象（YAGNI）——只实现当前确定要用的能力 |

### 阶段 2：账号域 MVP（2~3 周）

| 项 | 内容 |
| --- | --- |
| 目标 | 注册、登录、令牌、个人资料、关注粉丝可用 |
| 交付物 | `app/user/{api,rpc,model}` 三目录与全量接口；`user.api` / `user.proto` 契约；`noviterra_user` 迁移与 seed；邮件验证码（MailHog 本地） |
| 验收 | Postman/newman 跑通：注册 → 登录 → 刷新 → 改资料 → 换头像 → 关注 → 登出；登录失败锁定、令牌复用检测有测试 |
| 依赖 | 阶段 1 |
| 风险 | 邮件/短信通道未通 → dev 用 MailHog + mock，不阻塞；**本阶段必须把 3.2 的三目录约定与 api 不碰 model 的纪律固化到脚手架** |

### 阶段 3：内容域 MVP（3 周）

| 项 | 内容 |
| --- | --- |
| 目标 | 写作 → 发布 → 阅读 → 列表/详情/热榜全链路可用 |
| 交付物 | `app/article/{api,rpc,model}`：`article-rpc`（文章/版本/分类/标签/专栏/点赞收藏）+ `article-api`（详情页并发聚合、编辑器接口）；`noviterra_content` 迁移；outbox 事件机制；缓存与多级读优化 |
| 验收 | 文章状态机测试全绿；详情接口 P99 < 100ms（缓存命中）；列表分页正确；浏览量异步落库且不丢；**`article-api` 中 grep 不到 `model` 与 `gorm` 的 import** |
| 依赖 | 阶段 2（作者信息、JWT） |
| 风险 | 范围膨胀（评论、搜索容易混进来）→ 严格按本阶段清单交付 |

### 阶段 4：后台管理域（2~3 周）

| 项 | 内容 |
| --- | --- |
| 目标 | 运营可独立完成用户/内容/评论/配置管理 |
| 交付物 | `app/system/{api,rpc,model}`（RBAC/配置/字典/公告/审计/任务元数据）；`app/admin`（跨域聚合 + RBAC 中间件 + 操作日志）；seed 内置角色与权限；内容审核流 |
| 验收 | 6 个内置角色权限边界正确；越权访问后台接口返回 403；操作日志完整落库；前台接口未被后台改动影响 |
| 依赖 | 阶段 3 |
| 风险 | 权限设计返工 → 权限码在阶段 4 开始时一次性评审定稿 |

### 阶段 5：互动与媒体（2~3 周）

| 项 | 内容 |
| --- | --- |
| 目标 | 评论体系与文件上传可用 |
| 交付物 | `app/comment/{api,rpc,model}`（评论/点赞/举报/敏感词）；`app/file/{api,rpc,model}`（上传/分片/秒传/图片处理/MinIO） |
| 验收 | 楼中楼结构正确；敏感词三级别生效；上传类型/大小/越权校验有测试；私有文件签名 URL 可用 |
| 依赖 | 阶段 3、4（审核与举报处理在后台） |
| 风险 | 大文件直传与对象存储联调 → 提前在阶段 1 就把 MinIO 跑起来 |

### 阶段 6：检索、消息、统计、调度（3 周）

| 项 | 内容 |
| --- | --- |
| 目标 | 搜索、站内信/邮件通知、后台看板、定时任务全部可用 |
| 交付物 | `app/search/{api,rpc,model}`（Meilisearch 索引与检索）；`app/message/{api,rpc,model}`（站内信/邮件/WS）；`app/stat/{api,rpc,model}`（埋点与聚合）；`app/portal`（首页跨域聚合）；`app/job`（全部定时任务） |
| 验收 | 发布文章 3 秒内可被搜到；评论后 5 秒内收到站内信（在线即时）；看板数据与 DB 对账一致；定时任务幂等且带分布式锁 |
| 依赖 | 阶段 3、5 |
| 风险 | 消息通道与 WS 扩容 → 单实例起步，预留 Redis Pub/Sub 方案 |

### 阶段 7：第三方能力（2 周）

| 项 | 内容 |
| --- | --- |
| 目标 | 翻译、AI 摘要、内容安全、IP 归属地接入，且可降级、可计量 |
| 交付物 | `pkg/thirdparty/*` 各适配器 + mock；`app/third/{api,rpc,model}`（编排/配额/审计/异步任务）；`article_translation` 由 `article-rpc` 落库；后台配额与日志页 |
| 验收 | 关闭开关后主链路不受影响；配额超限触发降级与告警；缓存命中率 > 60%（模拟数据）；每个适配器有单测与冒烟脚本 |
| 依赖 | 阶段 6（异步任务、统计） |
| 风险 | 供应商协议变更/额度 → 必须实现主备切换与 mock 兜底 |

### 阶段 8：上线与稳定性（2 周）

| 项 | 内容 |
| --- | --- |
| 目标 | 生产可用、可观测、可回滚 |
| 交付物 | K8s 部署清单（base + overlays）；Prometheus/Grafana/Alertmanager 配置；备份与恢复脚本；压测报告；运维手册；故障预案 |
| 验收 | 压测达标（详情接口 500 QPS / P99 < 200ms）；告警能真实触发；模拟依赖宕机时降级正确；备份可成功恢复；灰度发布与回滚演练通过 |
| 依赖 | 阶段 0~7 |
| 风险 | 性能不达标 → 提前在阶段 3 就做基准测试，别等到最后 |

### 里程碑总览

| 里程碑 | 包含阶段 | 可对外演示的能力 |
| --- | --- | --- |
| M1 地基可用 | 0 + 1 | 无（内部） |
| M2 能用（内测） | 2 + 3 | 注册登录、写文章、读文章 |
| M3 能运营（公测） | 4 + 5 | 后台管理、评论互动、图片上传 |
| M4 完整（正式发布） | 6 + 7 | 搜索、通知、看板、翻译/AI |
| M5 稳定（生产加固） | 8 | 高可用、监控告警、备份恢复 |

---

## 15. 附录

### 附录 A：文件创建清单（必须由项目所有者逐项创建）

> **使用方式**：按阶段顺序创建；每个条目都是**具体文件/目录路径**。全部相对仓库根 `D:\GoProject\IT\Noviterra\`。
> **纪律**：先创建"阶段 0"全部条目并跑通，再进入阶段 1；每完成一项打 `[x]`。
> **说明**：`goctl` 生成的文件标注为【生成】，不要手写；其余为手写。

#### A.0 阶段 0：仓库与工程地基

- [ ] `.gitignore`（忽略 `bin/`、`dist/`、`.env`、`etc/*.local.yaml`、`*.log`、`node_modules/`、`.idea/`）
- [ ] `.gitattributes`（`* text=auto eol=lf`；`*.pb.go linguist-generated=true`）
- [ ] `.editorconfig`
- [ ] `.golangci.yml`
- [ ] `.dockerignore`
- [ ] `LICENSE`
- [ ] `README.md`
- [ ] `CONTRIBUTING.md`
- [ ] `CHANGELOG.md`
- [ ] `Makefile`（根级聚合）
- [ ] `.github/workflows/ci-server.yml`
- [ ] `.github/workflows/ci-web.yml`
- [ ] `.github/workflows/release.yml`
- [ ] `doc/deploy.md`
- [ ] `doc/ops.md`
- [ ] `doc/adr/0001-project-module-path.md`（记录 module 命名决策）
- [ ] `doc/adr/0002-microservice-split.md`（记录服务拆分决策）
- [ ] `doc/adr/0003-domain-autonomous-layout.md`（★ 记录"每域 `{api,rpc,model}` + 边缘路由、不设业务网关"的决策）
- [ ] `doc/db/README.md`（ER 图与表说明入口）
- [ ] `server/Makefile`
- [ ] `server/scripts/gen.sh` / `gen.ps1`
- [ ] `server/scripts/dev-up.sh` / `dev-down.sh`
- [ ] `server/scripts/migrate.sh`
- [ ] `server/scripts/seed.sh`
- [ ] `server/scripts/check.sh`
- [ ] `server/deploy/compose/docker-compose.infra.yml`
- [ ] `server/deploy/compose/.env.example`
- [ ] `server/deploy/docker/Dockerfile`
- [ ] `server/deploy/nginx/nginx.conf` 与 `server/deploy/nginx/conf.d/noviterra.conf`（★ 边缘层路由表：路径前缀 → 各域 api，见 6.1）
- [ ] `server/deploy/monitor/prometheus/prometheus.yml`
- [ ] `server/deploy/monitor/grafana/`（数据源与看板 JSON）
- [ ] `server/deploy/sql/migrations/`（目录，goose 迁移文件放这里）
- [ ] `server/deploy/sql/seed/`（目录）
- [ ] `front/pnpm-workspace.yaml`、`front/package.json`
- [ ] `front/web/`、`front/admin/`、`front/shared/`（脚手架由前端工具创建）

#### A.1 阶段 1：`server/pkg` 公共包

- [ ] `server/pkg/logx/{logx.go,config.go,fields.go,context.go,gorm.go,mask.go,rotate.go}`
- [ ] `server/pkg/errx/{code.go,errx.go,map.go,handler.go,detect.go}`
- [ ] `server/pkg/response/{response.go,page.go,types.go}`
- [ ] `server/pkg/db/{db.go,config.go,logger.go,plugin_trace.go,transaction.go,paginate.go,scope.go,errors.go}`
- [ ] `server/pkg/cache/{redis.go,config.go,keys.go,codec.go,lock.go,limiter.go,counter.go,singleflight.go}`
- [ ] `server/pkg/jwt/{jwt.go,claims.go,config.go,refresh.go,blacklist.go,errs.go}`
- [ ] `server/pkg/crypt/{password.go,aes.go,hmac.go,random.go,hash.go}`
- [ ] `server/pkg/idgen/{snowflake.go,uuid.go,shortcode.go,slug.go}`
- [ ] `server/pkg/paginate/paginate.go`
- [ ] `server/pkg/conv/{convert.go,slice.go,chunk.go}`
- [ ] `server/pkg/ctxx/context.go`
- [ ] `server/pkg/validatorx/validator.go`
- [ ] `server/pkg/constants/{enums.go,ttl.go,limits.go,regex.go}`
- [ ] `server/pkg/middleware/{trace.go,accesslog.go,recovery.go,cors.go,ratelimit.go,auth.go,rpcmeta.go}`
- [ ] `server/pkg/observability/{trace.go,metrics.go,health.go}`
- [ ] `server/pkg/mq/{producer.go,consumer.go,envelope.go}`
- [ ] `server/pkg/schedule/{schedule.go,lock.go}`
- [ ] `server/pkg/version/version.go`
- [ ] 各包对应 `*_test.go`
- [ ] `server/proto/common/{common.proto,event.proto}`

#### A.2 阶段 2：账号域（`app/user/{api,rpc,model}`）

- [ ] `server/api/common/{base.api,auth.api}`（公共 HTTP type）
- [ ] `server/app/user/api/user.api`（★ HTTP 契约，手写）
- [ ] `server/app/user/rpc/user.proto`（★ gRPC 契约，手写）
- [ ] `server/app/user/api/{user.go, etc/user-api.yaml, internal/{config,svc,handler,logic,types,middleware,consts}}`【goctl 生成后手改】
- [ ] `server/app/user/rpc/{user.go, etc/user-rpc.yaml, pb/, user/, internal/{config,svc,server,logic,middleware,dto,vo,assembler,consts}}`【goctl 生成后手改】
- [ ] `server/app/user/model/{entity,repository,cache}/**`（entity：user、user_auth、user_profile、user_setting、user_stat、user_follow、user_role、user_login_log）
- [ ] `server/app/user/rpc/internal/domain/**`（密码策略、令牌规则、关注关系规则、用户状态机）
- [ ] `server/app/user/rpc/internal/event/**`（user.registered / user.login / user.banned / user.followed / user.updated）
- [ ] `server/deploy/sql/migrations/00001_init_user.sql`
- [ ] `server/deploy/sql/migrations/00002_init_system_base.sql`（角色表基础）
- [ ] `server/deploy/sql/seed/0001_roles_permissions.sql`
- [ ] `server/test/integration/auth_test.go`

#### A.3 阶段 3：内容域（`app/article/{api,rpc,model}`）

- [ ] `server/app/article/api/article.api`（含文章/分类/标签/专栏/收藏夹）
- [ ] `server/app/article/rpc/article.proto`
- [ ] `server/app/article/api/**`【生成后手改】（重点：`internal/logic/getarticledetaillogic.go` 的并发聚合）
- [ ] `server/app/article/rpc/**`【生成后手改】（重点：`internal/domain/`、`internal/event/`）
- [ ] `server/app/article/model/entity/*.go`（article、article_content、article_version、category、tag、article_tag、series、series_article、article_like、article_collect、collect_folder、article_slug_history、outbox）
- [ ] `server/app/article/model/{repository,cache}/**`
- [ ] `server/deploy/sql/migrations/00003_init_content.sql`
- [ ] `server/test/integration/article_test.go`

#### A.4 阶段 4：后台管理域

- [ ] `server/app/system/api/system.api` + `server/app/system/rpc/system.proto`
- [ ] `server/app/system/{api,rpc,model}/**`（model/entity：role、permission、user_role、role_permission、sys_config、sys_dict、sys_dict_item、sys_announcement、sys_operation_log、sys_job、sys_job_log）
- [ ] `server/app/admin/admin.go`
- [ ] `server/app/admin/api/{admin.api,user.admin.api,article.admin.api,comment.admin.api,system.admin.api,stat.admin.api,job.admin.api,third.admin.api}`
- [ ] `server/app/admin/internal/{config,svc,handler,logic,types,middleware/{authz.go,oplog.go},consts}/**`
- [ ] `server/deploy/sql/migrations/00004_init_system.sql`
- [ ] `server/deploy/sql/seed/0002_system_config.sql`（站点配置默认值）
- [ ] `server/deploy/sql/seed/0003_dict.sql`

#### A.5 阶段 5：互动与媒体

- [ ] `server/app/comment/api/comment.api` + `server/app/comment/rpc/comment.proto`
- [ ] `server/app/comment/{api,rpc,model}/**`（重点：`rpc/internal/domain/dfa.go` 敏感词自动机）
- [ ] `server/app/file/api/file.api` + `server/app/file/rpc/file.proto`
- [ ] `server/app/file/{api,rpc,model}/**`（重点：`rpc/internal/storage/` 对 `pkg/thirdparty/storage` 的调用适配）
- [ ] `server/deploy/sql/migrations/00005_init_comment.sql`
- [ ] `server/deploy/sql/migrations/00006_init_file.sql`
- [ ] `server/deploy/sql/seed/0004_sensitive_words.sql`

#### A.6 阶段 6：检索、消息、统计、门户、调度

- [ ] `server/app/search/{api/search.api, rpc/search.proto}` + `{api,rpc,model}/**`（重点：`rpc/internal/engine/{engine.go,meilisearch/,memory/}`）
- [ ] `server/app/message/{api/message.api, rpc/message.proto}` + `{api,rpc,model}/**`（重点：`rpc/internal/channel/`、`rpc/internal/consumer/`、WS 入口在 api 层）
- [ ] `server/app/stat/{api/stat.api, rpc/stat.proto}` + `{api,rpc,model}/**`
- [ ] `server/app/portal/{portal.go, api/portal.api, internal/{config,svc,handler,logic,types,middleware,consts}}`（★ 只有 api，无 rpc/model）
- [ ] `server/app/job/{job.go, etc/job.yaml, internal/{config,svc,server,jobs,consts}}`（★ 无 api/rpc/model）
- [ ] `server/deploy/sql/migrations/00007_init_search.sql`、`00008_init_message.sql`、`00009_init_stat.sql`

#### A.7 阶段 7：第三方能力

- [ ] `server/pkg/thirdparty/{README.md,provider.go,options.go,transport.go,errors.go,retry.go,breaker.go,ratelimit.go,cache.go,metrics.go,audit.go,mask.go}`
- [ ] `server/pkg/thirdparty/translate/{translate.go,factory.go,mock/mock.go}`
- [ ] `server/pkg/thirdparty/translate/baidu/baidu.go`
- [ ] `server/pkg/thirdparty/ai/{ai.go,factory.go,prompt/}`
- [ ] `server/pkg/thirdparty/ai/deepseek/deepseek.go`
- [ ] `server/pkg/thirdparty/contentsec/{contentsec.go,local/dfa.go,baidu/baidu.go}`
- [ ] `server/pkg/thirdparty/storage/{storage.go,local/,minio/,image/}`
- [ ] `server/pkg/thirdparty/email/{email.go,smtp/}`
- [ ] `server/pkg/thirdparty/sms/{sms.go,aliyun/}`
- [ ] `server/pkg/thirdparty/ipregion/{ipregion.go,ip2region/}`
- [ ] `server/pkg/thirdparty/captcha/{captcha.go,local/}`
- [ ] `server/pkg/thirdparty/oauth/{oauth.go,github/}`
- [ ] `server/app/third/{api/third.api, rpc/third.proto}`
- [ ] `server/app/third/{api,rpc,model}/**`（重点：`rpc/internal/task/`、`rpc/internal/consumer/`）
- [ ] `server/deploy/sql/migrations/00010_init_third.sql`
- [ ] `server/deploy/sql/migrations/00011_article_translation.sql`
- [ ] `server/deploy/compose/docker-compose.app.yml`

#### A.8 阶段 8：上线加固

- [ ] `server/deploy/k8s/base/{kustomization.yaml,deployment.yaml,service.yaml,ingress.yaml,configmap.yaml}`
- [ ] `server/deploy/k8s/overlays/{dev,test,prod}/kustomization.yaml`
- [ ] `server/deploy/k8s/hpa/*.yaml`
- [ ] `server/deploy/monitor/alertmanager/alert-rules.yml`
- [ ] `server/deploy/monitor/grafana/dashboards/{service-overview.json,business-overview.json}`
- [ ] `server/deploy/monitor/jaeger/config.yaml`
- [ ] `server/test/e2e/*_test.go`
- [ ] `server/test/benchmark/*.js`（k6 脚本）
- [ ] `doc/postmortem/TEMPLATE.md`

---

### 附录 B：常用 goctl 命令速查

> 仅为命令模板，**不要直接复制执行**前先确认参数与当前目录。

```bash
# 1) 生成 RPC 骨架（首次，在 server/ 目录下执行；注意输出到 app/<域>/rpc）
goctl rpc protoc app/user/rpc/user.proto \
  --go_out=app/user/rpc/pb --go-grpc_out=app/user/rpc/pb \
  --zrpc_out=app/user/rpc --module=Noviterra/server/app/user/rpc

# 2) proto 变更后重新生成（pb 与 client 会更新；已存在的 logic 不会被覆盖）
goctl rpc protoc app/article/rpc/article.proto \
  --go_out=app/article/rpc/pb --go-grpc_out=app/article/rpc/pb \
  --zrpc_out=app/article/rpc --module=Noviterra/server/app/article/rpc

# 3) 生成 HTTP（api）骨架（输出到 app/<域>/api）
goctl api go -api app/article/api/article.api -dir app/article/api --style=goZero

# 4) api 变更后重新生成 handler/types（logic 不覆盖）
goctl api go -api app/article/api/article.api -dir app/article/api --style=goZero

# 5) 导出接口文档（markdown）
goctl api doc app/article/api/article.api -dir ../doc/api

# 6) 契约文件较大时按子域拆分，再用主文件 import 聚合
#    例：app/article/api/article.api 里 import "article_category.api"

# 7) 格式化 + 校验 api 文件
goctl api format --dir app/article/api
goctl api validate --api app/article/api/article.api

# 8) 生成 Dockerfile / K8s 参考（按需，项目使用自维护模板）
goctl docker -go app/user/rpc/user.go
goctl kube deploy -name article-rpc -namespace noviterra -image registry/noviterra/article-rpc:tag -o deploy/k8s
```

**注意**：`goctl` 默认生成 `sqlx` model，本项目用 GORM，**不要**用 `goctl model` 生成到生产目录；`app/<域>/model/` 一律手写（可用自定义模板 `--home` 定制）。

---

### 附录 C：命名与常量速查表

**服务与端口**

| 服务 | 端口 | etcd Key | 目录 | 数据库 |
| --- | --- | --- | --- | --- |
| user-api | 8801 | `user.api` | `app/user/api/` | — |
| user-rpc | 9001 | `user.rpc` | `app/user/rpc/` | noviterra_user |
| system-api | 8802 | `system.api` | `app/system/api/` | — |
| system-rpc | 9002 | `system.rpc` | `app/system/rpc/` | noviterra_system |
| article-api | 8803 | `article.api` | `app/article/api/` | — |
| article-rpc | 9003 | `article.rpc` | `app/article/rpc/` | noviterra_content |
| comment-api | 8804 | `comment.api` | `app/comment/api/` | — |
| comment-rpc | 9004 | `comment.rpc` | `app/comment/rpc/` | noviterra_comment |
| file-api | 8805 | `file.api` | `app/file/api/` | — |
| file-rpc | 9005 | `file.rpc` | `app/file/rpc/` | noviterra_file |
| message-api | 8806 | `message.api` | `app/message/api/` | — |
| message-rpc | 9006 | `message.rpc` | `app/message/rpc/` | noviterra_message |
| search-api | 8807 | `search.api` | `app/search/api/` | — |
| search-rpc | 9007 | `search.rpc` | `app/search/rpc/` | noviterra_search |
| stat-api | 8808 | `stat.api` | `app/stat/api/` | — |
| stat-rpc | 9008 | `stat.rpc` | `app/stat/rpc/` | noviterra_stat |
| third-api | 8809 | `third.api` | `app/third/api/` | — |
| third-rpc | 9009 | `third.rpc` | `app/third/rpc/` | noviterra_third |
| portal-api | 8810 | `portal.api` | `app/portal/api/` | —（不拥有数据） |
| admin | 8811 | `admin.api` | `app/admin/api/` | —（不拥有数据） |
| job | 9100 | `job.cron` | `app/job/` | —（只用 system 库的任务表） |

> 指标/健康端口 = 业务端口 + 10000（如 `article-api` → 18803、`article-rpc` → 19003）。

**事件主题**

| 主题 | 生产者 | 消费者 |
| --- | --- | --- |
| `user.registered.v1` / `user.login.v1` / `user.banned.v1` / `user.followed.v1` / `user.updated.v1` | user-rpc | message、stat、article、comment |
| `article.created.v1` / `article.published.v1` / `article.updated.v1` / `article.deleted.v1` / `article.liked.v1` | article-rpc | search、stat、message、third |
| `comment.created.v1` / `comment.deleted.v1` / `comment.reported.v1` | comment-rpc | article、message、stat |
| `file.uploaded.v1` / `file.deleted.v1` | file-rpc | article、stat |
| `message.sent.v1` / `message.failed.v1` | message-rpc | stat、system（审计） |
| `third.quota.exceeded.v1` / `third.degraded.v1` | third-rpc | system（告警）、stat |

**缓存 Key 前缀**

| 前缀 | 归属 |
| --- | --- |
| `nv:{env}:user:` | user-rpc |
| `nv:{env}:system:` | system-rpc |
| `nv:{env}:article:` | article-rpc |
| `nv:{env}:comment:` | comment-rpc |
| `nv:{env}:file:` | file-rpc |
| `nv:{env}:msg:` | message-rpc |
| `nv:{env}:search:` | search-rpc |
| `nv:{env}:stat:` | stat-rpc |
| `nv:{env}:third:` | third-rpc |
| `nv:{env}:lock:` | 全局分布式锁 |

---

### 附录 D：术语表

| 术语 | 说明 |
| --- | --- |
| BFF | Backend For Frontend，面向前端的聚合层；本项目的 `app/portal`（首页聚合）与 `app/admin`（后台聚合）即此角色，**但不再是唯一的 HTTP 入口** |
| outbox | 本地消息表模式，事务内写事件表，后台投递，保证"数据变了事件一定发" |
| Cache-Aside | 旁路缓存：读时回填，写时删除 |
| singleflight | 合并同一 Key 的并发回源请求，只打一次 DB |
| 熔断 | 依赖连续失败后快速失败一段时间，避免雪崩 |
| 降级 | 依赖不可用时提供有损但可用的替代结果 |
| 幂等 | 同一请求执行多次结果一致 |
| IDOR | Insecure Direct Object Reference，越权访问他人资源 |
| DFA | 确定有限自动机，本项目用于敏感词快速匹配 |
| ADR | Architecture Decision Record，架构决策记录 |
| NFR | Non-Functional Requirements，非功能需求 |

---

### 附录 E：风险清单与应对

| # | 风险 | 影响 | 概率 | 应对 |
| --- | --- | --- | --- | --- |
| 1 | 服务拆得过细，单人维护成本高 | 进度延期 | 高 | 严格按 P0/P1/P2 推进；P2 服务（search/stat/third）可先以包形式内嵌，不单独起进程 |
| 2 | 微服务带来的分布式复杂度（一致性、排障） | Bug 增多 | 高 | outbox + 对账任务 + 全链路 trace；冗余字段登记表 |
| 3 | 缓存与 DB 不一致 | 数据错误 | 中 | 统一"先库后删缓存 + 延迟双删"；每日对账 |
| 4 | 第三方额度/协议变更 | 功能不可用 | 中 | 主备切换 + mock 兜底 + 配额告警 + 契约冒烟 |
| 5 | 早期过度设计（Kafka、ES、K8s 全上） | 拖慢进度 | 中 | 起步用 Redis Stream/Meilisearch/Compose，接口预留可切换 |
| 6 | 前端与后端契约漂移 | 联调返工 | 中 | `.api`/`.proto` 为单一来源；类型与错误码脚本生成 |
| 7 | 内容合规风险 | 法律风险 | 低~中 | 本地词库 + 云审核 + 人工复核 + 举报 + 留痕 |
| 8 | 大表拖慢（view_log / operation_log） | 查询变慢 | 高 | 分区 + 归档 + 预建/删分区任务（阶段 6 必须落地） |
| 9 | 数据丢失（无备份演练） | 灾难 | 低 | 每日全备 + binlog，季度恢复演练，备份加密 |
| 10 | 密钥泄漏 | 严重 | 低 | 环境变量/Vault + 轮换 + `govulncheck` + 提交前扫描（gitleaks） |
| 11 | 单人项目中途放弃复杂度 | 项目烂尾 | 中 | 按里程碑交付可演示版本，每阶段都能"跑起来" |
| 12 | 性能不达标（详情页） | 体验差 | 中 | 多级缓存 + 异步计数 + 阶段 3 就做基准测试 |

---

### 附录 F：待决策问题（Open Questions）

> 这些问题**需要项目所有者确认**，确认后应写入 `doc/adr/` 并更新本文。

| # | 问题 | 建议方案 | 影响 | 最晚决策时间 |
| --- | --- | --- | --- | --- |
| 1 | Go module 路径是否改名（当前 `Noviterra`） | 改为 `github.com/<org>/noviterra`（小写、域名限定） | 所有 import 路径 | **阶段 0 结束前（P0）** |
| 2 | 是否多站点/多租户（一个实例承载多个独立博客） | 起步单站点，表结构预留 `site_id` 字段 | 数据库与路由设计 | 阶段 3 前 |
| 3 | 前端技术栈 | 前台 Next.js（SSR，利于 SEO）+ 后台 Vue3 + Element Plus | 目录与构建 | 阶段 0 |
| 4 | 搜索引擎 | 起步 Meilisearch，规模化换 ES | `app/search/rpc` 实现 | 阶段 6 前 |
| 5 | 消息队列 | 起步 Redis Stream / asynq，规模化换 Kafka | `pkg/mq` 实现 | 阶段 3 前（outbox 投递方式） |
| 6 | 部署形态（合并 vs 独立进程） | 阶段 2~6 每域合并成 1 进程；阶段 7 起按需拆分（见 [11.4](#114-部署形态先合并后拆分)） | 进程数、K8s 清单 | 阶段 2 前 |
| 7 | 生产运行载体 | 起步 Docker Compose，正式 K8s | 交付物 | 阶段 8 前 |
| 8 | 边缘层选型（Nginx vs Ingress vs APISIX） | 起步 Nginx/Ingress 手写路由表；需要 WAF/插件生态时再评估 APISIX | 路由与鉴权边界 | 阶段 8 前 |
| 9 | 邮件/短信服务商 | 邮件用自有企业邮箱 SMTP，短信用阿里云 | thirdparty 实现 | 阶段 2 前（邮箱验证码） |
| 10 | `comment` 是否需要独立域 | 已按独立域设计；若人手不足，可先只建 `rpc` 并把 `api` 并入 `article-api`（但会破坏"第一段路径唯一归属"的路由约束，需同步调整路径） | 拆分粒度 | 阶段 5 前 |
| 11 | 是否需要开放 API（第三方接入发文） | 预留 `/open/**` 与签名中间件，暂不实现 | 边缘路由与签名 | 阶段 8 前 |
| 12 | 是否做付费/会员/打赏 | 预留 `payment` 占位，暂不实现 | 表与模块 | 不紧急 |
| 13 | 文章正文字段上限与是否支持富文本 | Markdown 优先，富文本只存 HTML 白名单 | 表结构与编辑器 | 阶段 3 前 |
| 14 | 是否需要多语言界面（i18n） | 前台支持 zh-CN/en，后台仅 zh-CN | 消息模板与前端 | 阶段 6 前 |
| 15 | 对象存储选型（自建 MinIO vs 云 OSS） | dev 自建 MinIO，生产按成本决定 | `app/file` 配置 | 阶段 5 前 |
| 16 | 是否接入 GitOps（ArgoCD） | 阶段 8 视团队规模决定 | 发布流程 | 阶段 8 |

---

### 文档结束语

本文是 Noviterra 的架构基线。**任何偏离本文的实现都必须在 `doc/adr/` 留档，并同步更新本文对应章节。**

实施时请牢记三条最容易被违反、代价也最高的纪律：

1. **数据所有权唯一**——谁的表谁写，跨服务只走 RPC / 事件。
2. **契约先行**——先 `.api` / `.proto`，再生成，再填 logic。
3. **每一步都可运行**——不允许出现"等所有服务写完才能跑"的状态。









