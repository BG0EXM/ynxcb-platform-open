# 伊宁县委宣传部部务工作平台

面向县级党委宣传部的内部部务工作平台，包含收文管理（呈批单/传阅登记卡/标签打印）、考勤点到、请假管理、公车报备、公共资料、通讯录、值守排班、周月年报收集与用户权限。

## 技术栈

- 后端：Go 1.25（标准库 HTTP）+ SQLite（modernc.org/sqlite，纯 Go，无 CGO）
- 前端：Vue 3 + Element Plus + Pinia + Vue Router + Vite
- 部署：单二进制 + 静态资源，Nginx 反代 HTTPS，systemd 托管

## 项目结构

```
ynxcb-platform/
├── backend/                 # Go 后端
│   ├── cmd/server/          # 入口
│   ├── internal/
│   │   ├── auth/            # JWT 认证
│   │   ├── config/          # 配置加载
│   │   ├── database/        # SQLite 建表与初始化
│   │   ├── handlers/        # 业务接口
│   │   ├── middleware/      # 认证/权限中间件
│   │   ├── models/          # 数据模型
│   │   └── router/          # 路由
│   └── static/              # 前端构建产物（SPA）
├── frontend/                # Vue 前端源码
│   └── src/
│       ├── views/           # 各业务页面
│       ├── router/          # 前端路由
│       ├── store/           # Pinia 状态
│       └── utils/           # axios 封装
└── deploy/                  # 部署文件
    ├── backup.sh            # 备份脚本
    ├── release.sh           # 发布包生成脚本
    ├── config.json.example  # 配置模板
    ├── nginx/ynxcb.conf     # Nginx HTTPS 配置
    └── systemd/ynxcb.service
```

> 全部文档集中在 `docs/` 目录：[README](README.md)（本项目）· [DEPLOY](DEPLOY.md)（部署·Nginx）· [DEPLOY-CADDY](DEPLOY-CADDY.md)（部署·Caddy）· [DEPLOY-WORDPRESS-COEXIST](DEPLOY-WORDPRESS-COEXIST.md)（WordPress 共存）· [OFFLINE-DEPLOYMENT](OFFLINE-DEPLOYMENT.md)（离线部署）· [ARCHITECTURE](ARCHITECTURE.md)（架构梳理）

## 本地开发

### 后端
```bash
cd backend
go run ./cmd/server -config config.json
# 默认监听 :8080，首次运行自动生成 config.json 和数据库
```

### 前端
```bash
cd frontend
npm install
npm run dev   # 开发服务器 :5173，代理 /api 到 :8080
npm run build # 构建到 dist/
```

构建后需将 `frontend/dist/*` 复制到 `backend/static/`。

### 交叉编译（Linux amd64）
在 Windows/Linux 上均可：
```bash
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ynxcb-server ./cmd/server
```

## 功能模块

| 模块 | 说明 |
|------|------|
| 收文管理 | 上级来文登记、呈批单/传阅登记卡/热敏标签生成与打印、文件退回管理 |
| 考勤点到 | 管理员晨会手工点到、月度/年度统计与打印 |
| 请假管理 | 年假及各类假期登记、统计与 Excel 导出 |
| 公车管理 | 车辆基础信息、用车报备、派车单打印、Excel 导出 |
| 公共资料 | 共享资料发布与阅读 |
| 通讯录 | 按部门/姓名查询 |
| 值守排班 | 日历视图当天值守（至21:00收文）、可标记县委大院排班 |
| 周/月/年报 | 上报、审阅、统计 |
| 系统管理 | 用户、角色、部门管理 |

> 各模块台账（用车报备/请假/考勤/排班/收文）均支持 Excel 导出。

## 角色与权限

- 管理员（admin）：全部功能 + 用户管理，**报表审阅**
- 科室工作人员（staff）：业务办理
- 乡镇/通讯员（reporter）：投稿、接收通知、上报材料

> 收文的新增/编辑/删除仅限"办公室"部门用户，其他部门只读；报表审阅仅管理员，普通用户只能提交/查看自己的报表。

## 部署

见 [DEPLOY.md](DEPLOY.md)。核心要点：
- 单二进制部署到 `/opt/ynxcb`
- systemd 托管自动重启
- Nginx 反代 HTTPS
- crontab 每日备份
