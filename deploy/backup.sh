#!/bin/bash
# 伊宁县委宣传部部务工作平台 - 数据备份脚本
# 用法: 手动执行 ./backup.sh  或加入 crontab 定时执行
#
# 建议 crontab 配置（每天凌晨2点备份，保留14天）:
#   0 2 * * * /opt/ynxcb/backup.sh >> /var/log/ynxcb-backup.log 2>&1

set -e

APP_DIR="/opt/ynxcb"
DATA_DIR="$APP_DIR/data"
BACKUP_DIR="/opt/ynxcb-backup"
RETENTION_DAYS=14

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/ynxcb_$DATE.tar.gz"

mkdir -p "$BACKUP_DIR"

# 使用 sqlite 在线备份（比直接复制更安全，WAL 模式下也一致）
if command -v sqlite3 >/dev/null 2>&1; then
    DB_BAK="/tmp/ynxcb_backup_$DATE.db"
    sqlite3 "$DATA_DIR/ynxcb.db" ".backup '$DB_BAK'"
    tar -czf "$BACKUP_FILE" -C /tmp "ynxcb_backup_$DATE.db" -C "$APP_DIR/data" "uploads" 2>/dev/null || \
    tar -czf "$BACKUP_FILE" -C /tmp "ynxcb_backup_$DATE.db"
    rm -f "$DB_BAK"
else
    # 无 sqlite3 时直接复制（先 checkpoint WAL）
    tar -czf "$BACKUP_FILE" -C "$DATA_DIR" "ynxcb.db" "uploads" 2>/dev/null || \
    tar -czf "$BACKUP_FILE" -C "$DATA_DIR" "ynxcb.db"
fi

# 清理过期备份
find "$BACKUP_DIR" -name "ynxcb_*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "[$(date '+%Y-%m-%d %H:%M:%S')] 备份完成: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"
