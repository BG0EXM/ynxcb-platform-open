# 伊宁县委宣传部部务工作平台 - 部署文档（Caddy）

> 适用：Debian · Caddy Web 服务器（自动 HTTPS）· 可与 WordPress 共存

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
└── data/                 # 运行时自动生成
```

## 二、两种部署方式

### 方式 A：本地编译好后上传（推荐，最简单）

**第 1 步：在电脑上编译**

```bash
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ynxcb-server ./cmd/server
```

（Windows 装 Go：https://go.dev/dl ；Linux：`sudo apt install golang`；不想装 Go 可用别人编译好的文件）

**第 2 步：上传两个东西到服务器同一目录**

```bash
sudo mkdir -p /opt/ynxcb
# 上传 ynxcb-server 和 static/（用仓库 backend/static/）到 /opt/ynxcb/
cd /opt/ynxcb
sudo chmod +x ynxcb-server
```

**第 3 步：创建运行用户 + 配置**

```bash
sudo useradd -r -s /usr/sbin/nologin ynxcb
sudo chown -R ynxcb:ynxcb /opt/ynxcb
sudo cp config.json.example config.json
sudo nano config.json   # 改 jwt.secret 和 admin.password
```

**第 4 步：systemd 自启**

```bash
sudo cp deploy/systemd/ynxcb.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now ynxcb
sudo systemctl status ynxcb
```

### 方式 B：服务器上直接编译（懂行的用）

```bash
# 传源码到服务器 → cd backend → 编译（同上命令）→ static 同目录 → 后续步骤同方式 A
```

---

## 三、Caddy 反代（自动 HTTPS）

在 Caddyfile（`/etc/caddy/Caddyfile`）末尾追加，替换 `bw.yourdomain.com` 为你的域名：

```caddyfile
bw.yourdomain.com {
    encode zstd gzip
    reverse_proxy 127.0.0.1:8080 {
        header_up X-Real-IP {remote_host}
    }
    request_body {
        max_size 50MB
    }
}
```

> 域名需先解析到服务器 IP。Caddy 自动申请/续期 Let's Encrypt 证书，无需手动配置。

```bash
sudo caddy reload
```

## 四、防火墙

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

## 五、备份

```bash
sudo -u ynxcb /opt/ynxcb/backup.sh
# crontab: 0 2 * * * /opt/ynxcb/backup.sh >> /var/log/ynxcb-backup.log 2>&1
```

备份在 `/opt/ynxcb-backup/`，保留 14 天。建议定期同步到异机或对象存储。

## 六、默认账号

首次部署后：`admin` / `admin123`。

> **安全机制**：使用默认密码登录后，系统会**强制要求修改密码**，未修改前无法使用其他功能。

## 七、常见问题

**页面 404 或空白？**
- 99% 是 `static/` 文件夹没放对或没传
- 确认服务器上 `ynxcb-server` 和 `static/` 在**同一目录**

**525 / SSL 握手失败？**
- 确认域名 DNS 已解析到服务器公网 IP
- 关闭 CDN/边缘加速（如阿里云 ESA）直连源站，否则证书验证可能失败
- `sudo journalctl -u caddy` 查看证书申请日志

**平台服务检查：**
```bash
sudo systemctl status ynxcb
curl -I http://127.0.0.1:8080
```
