# 伊宁县委宣传部部务工作平台 - 部署文档

服务器配置：Debian 13 (2C / 2G) · HTTPS 公网访问

## 一、目录结构

```
/opt/ynxcb/
├── ynxcb-server          # Go 编译后的二进制（含前端静态资源）
├── config.json           # 配置文件
├── data/
│   ├── ynxcb.db          # SQLite 数据库
│   └── uploads/          # 上传附件
└── backup.sh             # 备份脚本
```

## 二、部署步骤

### 1. 上传并解压发布包

```bash
# 在服务器上执行
mkdir -p /opt/ynxcb
# 将 ynxcb-server（二进制）、config.json、backup.sh 上传到 /opt/ynxcb/
cd /opt/ynxcb
chmod +x ynxcb-server backup.sh
```

### 2. 创建运行用户（安全加固）

```bash
sudo useradd -r -s /usr/sbin/nologin ynxcb
sudo chown -R ynxcb:ynxcb /opt/ynxcb
```

### 3. 修改配置文件

编辑 `/opt/ynxcb/config.json`：
- `jwt.secret`：改为 64 位以上随机字符串（可用 `openssl rand -base64 48` 生成）
- `admin.password`：设置初始管理员密码（首次登录后务必修改）
- 如使用国内服务器无法访问外网，上传附件大小限制 `upload.max_mb` 可按需调整

### 4. 配置 systemd 服务

```bash
sudo cp /opt/ynxcb/ynxcb.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable ynxcb
sudo systemctl start ynxcb
sudo systemctl status ynxcb
```

### 5. 反向代理 + HTTPS（Nginx）

将 `../deploy/nginx/ynxcb.conf` 上传到 `/etc/nginx/sites-available/`，
替换 `your-domain.com` 为你的域名，`/path/to/fullchain.pem` 和 `/path/to/privkey.pem`
替换为你的证书路径（阿里云/腾讯云 SSL 或 Let's Encrypt）。

```bash
sudo cp ynxcb.conf /etc/nginx/sites-available/
sudo ln -s /etc/nginx/sites-available/ynxcb.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 6. 防火墙

```bash
sudo ufw allow 443/tcp
sudo ufw allow 80/tcp   # 用于证书续期
sudo ufw enable
```

后端 8080 端口无需对外开放（Nginx 内网反代）。

## 三、数据备份

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

备份文件保存在 `/opt/ynxcb-backup/`，默认保留 14 天。
**重要**：请将备份目录定期同步到另一台机器或对象存储（如 rsync 到异机、上传 OSS）。

### 恢复备份
```bash
tar -xzf /opt/ynxcb-backup/ynxcb_20260804_020000.tar.gz -C /tmp/restore
sudo systemctl stop ynxcb
sudo cp /tmp/restore/ynxcb_backup_*.db /opt/ynxcb/data/ynxcb.db
sudo chown ynxcb:ynxcb /opt/ynxcb/data/ynxcb.db
sudo systemctl start ynxcb
```

## 四、日常维护

### 查看日志
```bash
sudo journalctl -u ynxcb -f
```

### 升级版本
```bash
sudo systemctl stop ynxcb
# 上传新二进制覆盖 /opt/ynxcb/ynxcb-server
sudo chmod +x /opt/ynxcb/ynxcb-server
sudo systemctl start ynxcb
```

### 磁盘检查
```bash
df -h /opt/ynxcb
du -sh /opt/ynxcb/data/uploads
```

## 五、默认账号

首次部署后用以下账号登录（**登录后立即修改密码**）：
- 管理员：`admin` / config.json 中 `admin.password` 设置的密码

## 六、注意事项

1. 平台使用 SQLite（WAL 模式），已优化为适合单机并发访问，勿在 NFS 等网络盘上运行数据库
2. 上传附件默认限制 50MB，按需调整
3. 服务器为 2C2G，已通过 systemd 内存限制（1G）防止 OOM，若并发量大可调优

## 七、无互联网（内网/隔离网）部署

平台已完全本地化，**运行时不依赖任何在线资源**（无 CDN、无在线字体、无远程脚本、无外部 API）。

- 在联网构建机上完成 `go build` 和 `npm run build`，将编译好的二进制 + `static/` 目录上传到内网服务器即可
- 服务器全程无需联网，只开 80/443 端口
- 详细排查说明与验证清单见 [OFFLINE-DEPLOYMENT.md](OFFLINE-DEPLOYMENT.md)
