# 伊宁县委宣传部部务工作平台

面向县级党委宣传部的内部部务工作平台，开源免费，可自部署。

- 🗂️ 收文管理（呈批单/传阅登记卡/热敏标签打印）、考勤点到、请假管理、公车报备、值守排班、公共资料、周月年报、通讯录
- ⚙️ Go + Vue3 + SQLite 单二进制部署，零运维
- 🔒 角色权限（管理员/科室/通讯员），收文操作限办公室部门
- 📄 各模块台账一键导出 Excel，支持 A4 打印与 80×50mm 热敏标签
- 🌐 完全离线可部署，Nginx / Caddy 均可

## 功能模块

| 模块 | 说明 |
|---|---|
| 收文管理 | 上级来文登记、呈批单/传阅登记卡/热敏标签生成打印、文件退回管理 |
| 考勤点到 | 管理员晨会手工点到、月度/年度统计与打印 |
| 请假管理 | 年假及各类假期登记、统计与 Excel 导出 |
| 公车管理 | 车辆基础信息（车架号/保险等）、用车报备、派车单打印 |
| 公共资料 | 共享资料发布与阅读 |
| 通讯录 | 按部门/姓名查询 |
| 值守排班 | 日历视图当天值守（至21:00收文）、可标记县委大院排班 |
| 周/月/年报 | 上报、审阅、统计 |
| 系统管理 | 用户、角色、部门管理 |

> 各模块台账（用车报备/请假/考勤/排班/收文）均支持 Excel 导出。

## 技术栈

- 后端：**Go 1.25**（标准库 HTTP）+ SQLite（纯 Go，无 CGO）
- 前端：**Vue 3 + Element Plus + Pinia + Vue Router + Vite**
- Excel 导出：excelize/v2
- 部署：主程序 + 静态资源目录，反向代理（Nginx / Caddy 均可），systemd 托管

## 目录结构

```
├── backend/                 # Go 后端
│   ├── cmd/server/          # 入口
│   ├── internal/            # 后端业务代码（auth/config/database/handlers/...）
│   └── static/              # 前端静态资源（本仓库已内置，开箱即用）
├── frontend/                # Vue 前端源码
│   └── src/
│       ├── views/           # 各业务页面
│       ├── router/          # 前端路由
│       ├── store/           # Pinia 状态
│       └── utils/           # axios 封装 + 导出下载
├── config.json.example      # 配置模板
└── docs/                    # 部署与架构文档
```

## 快速开始

### 后端

```bash
cd backend
cp ../config.json.example config.json   # 修改密钥与管理密码
go run ./cmd/server -config config.json
# 默认监听 :8080，首次运行自动建库
```

> `backend/static/` 已内置前端构建产物，**开箱即用**，无需再构建前端。

### 修改前端后重新构建

```bash
cd frontend
npm install
npm run build
# 将 dist 复制回 backend/static：
rm -rf ../backend/static && mkdir -p ../backend/static && cp -r dist/* ../backend/static/
```

### 交叉编译（Linux amd64）

```bash
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ynxcb-server ./cmd/server
```

## 部署

系统由两部分组成，**都要放到服务器上**：
1. **主程序** `ynxcb-server`（编译出的可执行文件）
2. **前端页面** `backend/static/`（网页文件）

> 本仓库的 `backend/static/` **已经帮你准备好了前端页面**，直接使用即可，无需自己构建。

**部署步骤（简版）**：

```bash
# 1. 把整个 backend 目录（含 static/）上传到服务器
# 2. 编译主程序（或在别处编译好后上传）
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ynxcb-server ./cmd/server

# 3. 运行
./ynxcb-server -config config.json
```

**什么时候需要重新构建前端？**
只有当你想**修改前端界面**（换背景图、改文案、改界面布局）时，才需要按上面的「修改前端后重新构建」用 `npm run build` 重新生成 `backend/static/`。不想改前端就完全不用管。

完整部署见 [docs/DEPLOY.md](docs/DEPLOY.md)（Nginx）或 [docs/DEPLOY-CADDY.md](docs/DEPLOY-CADDY.md)（Caddy）。

## 权限模型

- **管理员（admin）**：全部功能 + 用户管理
- **科室工作人员（staff）**：业务办理
- **乡镇/通讯员（reporter）**：投稿、接收通知、上报材料

> 收文的新增/编辑/删除仅限"办公室"部门用户，其他部门只读。

## 默认账号

首次部署后：`admin` / `admin123`（**登录后立即修改密码**，可在个人中心修改）。

## License

[Apache License 2.0](LICENSE)

## 声明

- 项目中的党徽图片为官方标识，仅供授权单位内部使用，部署者须遵守相关规定
- 本项目用于单位内部管理，部署时请遵循当地网络安全与保密要求
- 登录页与主页面底部预置了"ICP备案号占位 / 公网安备号占位"，公网部署者请替换为自己的真实备案号（在 `frontend/src/views/Login.vue` 与 `Layout.vue` 中修改后重新构建）
