# 伊宁县委宣传部部务工作平台 — 架构梳理文档

> 更新日期：2026-08-05
> 状态：现行架构快照，供技术演进参考（尚未重构）

## 一、总体架构

单体全栈应用，Go 单二进制同时提供 API 和前端静态资源，SQLite 单文件数据库，Nginx 反代 HTTPS。

```
浏览器 (Vue3 SPA + Element Plus)
   │  HTTPS
   ▼
Nginx (443, 反代 + 静态缓存)
   │  127.0.0.1:8080
   ▼
Go 单二进制 (ynxcb-server, ~10MB)
   ├─ Router (http.ServeMux, Go1.22 模式路由)
   ├─ Middleware: Auth(JWT) → RequireRole
   ├─ Handlers: 8 大业务域
   └─ SQLite (WAL) + 文件存储 (data/uploads/)
```

## 二、技术栈

| 层 | 技术 | 版本/说明 |
|---|---|---|
| 前端 | Vue 3 (Composition API) | 3.5.x |
| UI | Element Plus | 2.9.x + zh-cn |
| 前端路由 | Vue Router | 4.5.x history 模式 |
| 状态 | Pinia | 2.x（仅 auth） |
| HTTP | Axios + 拦截器 | token 注入 / 401 跳转 |
| 后端 | Go net/http | 1.25，标准库，未用 Web 框架 |
| 数据库 | SQLite (modernc.org/sqlite) | 纯 Go 无 CGO，WAL，MaxOpenConns=1 |
| Excel 导出 | excelize/v2 | 纯 Go 生成 xlsx |
| 认证 | JWT HS256 | golang-jwt/v5，24h 无刷新 |
| 密码 | bcrypt | golang.org/x/crypto |
| 品牌资源 | 官方党徽 PNG (1024x1024) | src/assets/danghui.png + public/favicon.png |
| 部署 | Nginx + systemd + crontab | 单机单进程，备份脚本 |

## 三、后端代码结构 (backend/internal)

```
cmd/server/main.go      入口：配置→JWT→数据库→启动
├── config/             JSON 配置（server/database/jwt/upload/admin）
├── database/           建表 + 迁移 + 种子数据 + 密码哈希
├── models/             数据模型
├── auth/               JWT 生成/解析
├── middleware/         Auth / RequireRole / JSON 响应辅助
├── handlers/           业务逻辑（SQL 直写，无 service 层）
│   ├── auth.go         登录/资料/改密/用户管理/角色/部门
│   ├── incoming.go     收文登记/传阅（含呈批单/传阅卡数据源）
│   ├── attendance.go   考勤点到 + 请假管理 + 月度/年度统计
│   ├── vehicle.go      公车信息 + 用车报备
│   ├── misc.go         通讯录 + 值守排班 + 周月年报
│   ├── study.go        公共资料
│   ├── export.go       Excel 导出（5 模块台账）
│   └── uploads.go      文件上传下载 + 首页统计
└── router/             全部路由集中注册（单文件）
```

## 四、数据库（SQLite 15 张表）

- 系统：users / roles / departments
- 业务：attendances / leave_records
-       vehicles / vehicle_applies
-       contacts / duty_schedules / reports
-       incoming_docs + circulation_records / attachments
-       study_materials

> 业务规则要点：
> - 值守排班：当天值守至 21:00 收文，一天可安排一至两人（UNIQUE duty_date, user_id），`is_dawangyuan` 标记当天是否同时有县委大院排班
> - 收文管理：incoming_docs 记录上级来文，含文件编号/需退回/是否已退/退回日期；circulation_records 记录传阅人；可打印呈批单、传阅登记卡、80×50mm 热敏标签
> - 考勤：管理员晨会手工点到（出勤/请假/出差/未到），请假细分年假/病假/事假等，支持月度/年度统计与打印

## 五、前端结构 (frontend/src)

```
router/index.js      路由表（meta.admin 权限标记）
store/auth.js       登录态 + 角色判断
utils/request.js    Axios 封装 + exportFile 导出下载
views/              18 页面
  Layout / Login / Dashboard
  IncomingDocs(+3打印: 呈批单/传阅卡/标签)
  Attendance(+1统计打印) / Leave / Vehicles(+1派车单打印)
  Study / Contacts / Duty / Reports / Users / Profile
```

## 六、已知架构问题

1. **Handler 职责混杂**：misc.go 为"杂项"包（通讯录+排班+报表）；uploads.go 含文件+统计
2. **无 service/repository 层**：SQL 直写在 handler，难测试、难复用、改动易遗漏
3. **路由权限硬编码**：router.go 单文件 139 行，模块增多会失控；无统一响应封装
4. **SQLite 单写并发限制**：MaxOpenConns=1，当前可接受，量级上来是瓶颈
5. **前端权限依赖 localStorage**：role_code 明文存于前端，安全依赖后端校验（后端已校验）
6. **无自动化测试 / 无 CI / 无日志分级**

## 七、结论与后续方向（暂不实施）

- 保持 SQLite（适合 2C2G 单机，零运维）
- 若重构，优先：按业务域拆分 handler → 抽 service/repository → 统一响应结构 → 补 go test
- 备选：数据库换 PostgreSQL（多写并发）、加 Redis 缓存（高频接口）
