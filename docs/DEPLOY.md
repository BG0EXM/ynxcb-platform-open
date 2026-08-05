# 伊宁县委宣传部部务工作平台 - 部署文档（Nginx）

服务器配置：Debian (2C / 2G) · HTTPS 公网访问

## 〇、系统由两部分组成

部署到服务器需要**两样东西**：

| 名称 | 说明 |
|---|---|
| `ynxcb-server` | 主程序（可执行文件） |
| `static/` 文件夹 | 前端网页页面 |

> **`ynxcb-server` 和 `static/` 必须放在同一个目录**（后端在运行目录下找 static）。
> 本仓库的 `backend/static/` 已帮你准备好前端页面，直接用即可。

## 一、部署后的目录结构

```
/opt/ynxcb/
├── ynxcb-server          # 主程序（可执行文件）
├── static/               # 前端页面文件夹（与主程序同级）
├── config.json           # 配置文件
├── backup.sh             # 备份脚本
└── data/                 # 运行时自动生成（数据库/上传文件）
```

## 二、两种部署方式

### 方式 A：本地编译好后上传（推荐，最简单，服务器不用装环境）

> 适合新手：在你自己电脑上编译一次，然后把两个东西传到服务器。

**第 1 步：在电脑上编译**

打开终端，进入项目 `backend` 目录：

```bash
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ynxcb-server ./cmd/server
```

- Windows 装 Go：https://go.dev/dl
- Linux 装 Go：`sudo apt install golang`
- 实在不想装 Go，可以让已编译好的人给你一份 `ynxcb-server` 文件

**第 2 步：上传到服务器**

把这两样放进同一目录（如 `/opt/ynxcb/`）：
- `ynxcb-server`（编译出的文件）
- `static/`（用仓库 `backend/static/` 里的，整个文件夹复制）

```bash
# 在服务器上
sudo mkdir -p /opt/ynxcb
# 用 SFTP/SCP 上传 ynxcb-server 和 static/ 到 /opt/ynxcb/
cd /opt/ynxcb
sudo chmod +x ynxcb-server
```

**第 3 步：创建运行用户（安全加固，可跳过）**

```bash
sudo useradd -r -s /usr/sbin/nologin ynxcb
sudo chown -R ynxcb:ynxcb /opt/ynxcb
```

**第 4 步：配置 config.json**

```bash
sudo cp /opt/ynxcb/config.json.example /opt/ynxcb/config.json
sudo nano /opt/ynxcb/config.json
```

修改：
- `jwt.secret`：改为随机长字符串（可用 `openssl rand -base64 48` 生成）
- `admin.username` / `admin.password`：设置管理员账号。**只要 `admin.password` 非空，程序每次启动都会用 config 中的用户名和密码覆盖管理员账号**，改完重启即生效（默认 `admin` / `admin123`）

> `data/` 目录无需手动创建，systemd 服务会在启动前自动创建并设置权限。

### 方式 B：在服务器上直接编译（懂行的用，服务器要装 Go）

```bash
# 1. 把整个项目源码传到服务器
# 2. 编译
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ynxcb-server ./cmd/server
# 3. 确保 static/ 和 ynxcb-server 同目录
# 4. 后续步骤同方式 A 的第 3、4 步
```

---

## 三、配置 systemd 服务（开机自启）

```bash
sudo cp /opt/ynxcb/deploy/systemd/ynxcb.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now ynxcb
sudo systemctl status ynxcb
```

> 如果目录不是 `/opt/ynxcb`，需修改 `ynxcb.service` 里的路径。

## 四、反向代理 + HTTPS（Nginx）

1. 把 `deploy/nginx/ynxcb.conf` 复制到服务器，修改域名和证书路径
2. 启用：

```bash
sudo cp /opt/ynxcb/deploy/nginx/ynxcb.conf /etc/nginx/sites-available/
sudo ln -s /etc/nginx/sites-available/ynxcb.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

> 证书可用阿里云/腾讯云 SSL，或 Let's Encrypt（`certbot`）。

## 五、防火墙

```bash
sudo ufw allow 443/tcp
sudo ufw allow 80/tcp   # 用于证书续期
sudo ufw enable
```

> 后端 8080 端口无需对外开放（Nginx 内网反代）。

## 六、数据备份

### 手动备份
```bash
sudo -u ynxcb /opt/ynxcb/backup.sh
```

### 定时备份（每天 2:00）
```bash
sudo crontab -e
# 加入：
0 2 * * * /opt/ynxcb/backup.sh >> /var/log/ynxcb-backup.log 2>&1
```

备份文件在 `/opt/ynxcb-backup/`，保留 14 天。**建议定期同步到另一台机器或对象存储。**

### 恢复备份
```bash
tar -xzf /opt/ynxcb-backup/ynxcb_xxx.tar.gz -C /tmp/restore
sudo systemctl stop ynxcb
sudo cp /tmp/restore/ynxcb_backup_*.db /opt/ynxcb/data/ynxcb.db
sudo chown ynxcb:ynxcb /opt/ynxcb/data/ynxcb.db
sudo systemctl start ynxcb
```

## 七、日常维护

```bash
# 查看日志
sudo journalctl -u ynxcb -f

# 升级版本（只替换主程序，static 不用动）
sudo systemctl stop ynxcb
# 上传新 ynxcb-server 覆盖 /opt/ynxcb/ynxcb-server
sudo chmod +x /opt/ynxcb/ynxcb-server
sudo systemctl start ynxcb

# 磁盘检查
df -h /opt/ynxcb
du -sh /opt/ynxcb/data/uploads
```

## 八、默认账号

首次部署后登录（**登录后立即修改密码**）：
- 管理员：`admin` / `admin123`

> **安全机制**：使用默认密码 `admin123` 或初始密码 `123456` 登录后，系统会**强制要求修改密码**，未修改前无法使用其他功能。

## 九、常见问题

**页面 404 或空白？**
- 99% 是 `static/` 文件夹没放对或没传
- 确认服务器上 `ynxcb-server` 和 `static/` 在**同一目录**

**升级后页面没变？**
- 浏览器强刷（Ctrl+F5）

## 十、注意事项

1. 平台使用 SQLite（WAL 模式），勿在 NFS 等网络盘上运行数据库
2. 上传附件默认限制 50MB，按需调整
3. 2C2G 服务器已通过 systemd 内存限制（1G）防止 OOM

## 十一、无互联网（内网/隔离网）部署

平台运行时不依赖任何在线资源。在联网机器上完成编译后，把 `ynxcb-server` + `static/` 传到内网即可。详见 [OFFLINE-DEPLOYMENT.md](OFFLINE-DEPLOYMENT.md)。
