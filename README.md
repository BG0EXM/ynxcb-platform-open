# 伊宁县委宣传部部务工作平台

> 版本：V1.2

面向县级党委宣传部的内部部务工作平台，开源免费，可自部署。

- 🗂️ 收文管理（呈批单/传阅登记卡/热敏标签打印）、考勤点到、请假管理、公车报备、值守排班、公共资料、周月年报、通讯录
- ⚙️ Go + Vue3 + SQLite 单二进制部署，零运维
- 🔒 角色权限（管理员/科室/通讯员），收文操作限办公室部门，报表审阅限管理员
- 🛡️ 登录失败 5 次锁定 10 分钟，防暴力破解
- 🔑 首次登录强制修改默认密码，保障账号安全
- 📄 各模块台账一键导出 Excel，支持 A4 打印与 80×50mm 热敏标签
- 📊 列表接口分页，数据量大时依然流畅
- ⚡ 前端按需加载，首屏体积减小约 75%
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

## 更新日志

### V1.2

- 🛡️ **登录限流**：同一账号连续 5 次登录失败后锁定 10 分钟，防暴力破解
- 🔒 **报表审阅权限**：报表审阅仅限管理员，普通用户只能提交/查看自己的报表
- 📊 **列表分页**：收文、通讯录、请假、用车报备、报表、考勤列表均支持分页，接口返回 total/page 供前端翻页
- ⚡ **前端按需加载**：Element Plus 按需引入，主包体积从约 1.2MB 降至约 290KB，首屏加载更快
- 🔄 **顶栏待办刷新改事件驱动**：收文登记/删除后实时刷新待办数，不再 60 秒轮询
- 🧹 清理无效代码（CreateUser 冗余空判断）

### V1.1

- 🔑 **首次登录强制改密**：使用默认密码登录后，强制要求修改密码，未修改前无法使用其他功能
- 🐛 **修复用户/通讯录部门显示空白**：部门列字段名修正

## 技术栈

- 后端：**Go 1.25**（标准库 HTTP）+ SQLite（纯 Go，无 CGO）
- 前端：**Vue 3 + Element Plus + Pinia + Vue Router + Vite**
- Excel 导出：excelize/v2
- 部署：主程序 + 静态资源目录，反向代理（Nginx / Caddy 均可），systemd 托管

## 目录结构

> 下面描述的是**源码仓库**的结构。部署到服务器时，只需取出「可执行文件 + static 文件夹」两样（见下方「部署」章节）。

```
├── backend/                 # Go 后端
│   ├── cmd/server/          # 入口
│   ├── internal/            # 后端业务代码（auth/config/database/handlers/...）
│   └── static/              # 前端静态资源（部署时取出来，与主程序放一起）
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
2. **static 文件夹**（前端网页文件）

部署后服务器目录结构如下：

```
/opt/ynxcb/
├── ynxcb-server        # 可执行文件
├── static/             # 前端页面文件夹（与可执行文件同级）
├── config.json         # 配置文件
└── data/               # 运行时自动生成（数据库/上传文件）
```

**部署有两种方式，任选其一：**

---

### 方式一：本地编译好后上传（推荐，最简单，服务器不用装环境）

> 适合小白，你只需要在自己电脑上做两步，然后把两个东西传到服务器。

**第 1 步：在自己电脑上编译出可执行文件**

在电脑上打开终端（命令行），进入项目 backend 目录：

```bash
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ynxcb-server ./cmd/server
```

> 前提：电脑上装了 Go 语言环境。没装的话，看文末「常见问题」里的安装方法。
> 如果你连 Go 都不想装，也可以找已编译好的 `ynxcb-server` 文件直接用。

**第 2 步：把两个东西传到服务器**

需要上传到服务器的就这两样（放进同一个目录，比如 `/opt/ynxcb/`）：

| 文件/文件夹 | 说明 |
|---|---|
| `ynxcb-server` | 编译出的可执行文件 |
| `static/` | 前端页面文件夹（用仓库 `backend/static/` 里的） |

**第 3 步：在服务器上运行**

```bash
cd /opt/ynxcb
./ynxcb-server -config config.json
```

> 用 systemd 托管可开机自启，见 `deploy/systemd/ynxcb.service`。

---

### 方式二：在服务器上直接编译（懂行的用，服务器要装 Go）

> 适合在服务器上有 Go 环境、且把源码传到服务器的人。

```bash
# 1. 把整个项目源码上传到服务器（含 backend、frontend 等）
# 2. 进入 backend 目录编译
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ynxcb-server ./cmd/server

# 3. 把 static 文件夹放到与 ynxcb-server 同一目录，然后运行
./ynxcb-server -config config.json
```

---

> **关键点**：后端在**运行目录**下找 `static/` 文件夹，所以无论用哪种方式，`ynxcb-server` 和 `static/` 必须放在**同一目录**。仓库里的 `backend/static/` 已帮你准备好前端页面，直接把它作为服务器上的 `static/` 文件夹用即可。

**什么时候需要重新构建前端？**
只有当你想**修改前端界面**（换背景图、改文案、改界面布局）时，才需要用 `npm run build` 重新生成 `static/` 文件夹。不想改前端就完全不用管。

完整部署见 [docs/DEPLOY.md](docs/DEPLOY.md)（Nginx）或 [docs/DEPLOY-CADDY.md](docs/DEPLOY-CADDY.md)（Caddy）。

## 权限模型

- **管理员（admin）**：全部功能 + 用户管理
- **科室工作人员（staff）**：业务办理
- **乡镇/通讯员（reporter）**：投稿、接收通知、上报材料

> 收文的新增/编辑/删除仅限"办公室"部门用户，其他部门只读。

## 默认账号

首次部署后：`admin` / `admin123`。

> **安全机制**：使用默认密码 `admin123` 或初始密码 `123456` 登录后，系统会**强制要求修改密码**，未修改前无法使用其他功能。请登录后立即设置新密码。

## 常见问题

**Q1：电脑没装 Go，怎么编译？**
- Windows：去 https://go.dev/dl 下载安装包，装完重启终端
- Linux：`sudo apt install golang`
- 如果实在不想装 Go，可以让已编译好的人给你一份 `ynxcb-server` 文件直接上传

**Q2：`CGO_ENABLED=0 GOOS=linux GOARCH=amd64` 是什么？**
- 表示"编译出 Linux 64 位系统能运行的程序"（不管你在 Windows 还是 Mac 上编译）
- 照抄这行命令即可，不用理解每个参数

**Q3：上传到哪里？**
- 可执行文件和 static 文件夹放**同一个目录**，建议 `/opt/ynxcb/`

**Q4：static 文件夹哪来的？**
- 直接用仓库里的 `backend/static/` 文件夹，整文件夹复制到服务器

**Q5：页面打开是 404 或空白？**
- 99% 是 static 文件夹没放对位置或没传上去
- 确认服务器上 `ynxcb-server` 和 `static/` 在**同一目录**

## License

[Apache License 2.0](LICENSE)

## 声明

- 项目中的党徽图片为官方标识，仅供授权单位内部使用，部署者须遵守相关规定
- 本项目用于单位内部管理，部署时请遵循当地网络安全与保密要求
- 登录页与主页面底部预置了"ICP备案号占位 / 公网安备号占位"，公网部署者请替换为自己的真实备案号（在 `frontend/src/views/Login.vue` 与 `Layout.vue` 中修改后重新构建）
