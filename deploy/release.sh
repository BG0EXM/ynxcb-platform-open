#!/bin/bash
# 生成 Linux amd64 发布包
# 用法: ./release.sh   或  bash release.sh

set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RELEASE_DIR="$ROOT/release"
VERSION=$(date +%Y%m%d)

echo "==> 构建后端 (linux/amd64)"
cd "$ROOT/backend"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "$RELEASE_DIR/ynxcb-server" ./cmd/server

echo "==> 构建前端"
cd "$ROOT/frontend"
npm run build
rm -rf "$ROOT/backend/static"
mkdir -p "$ROOT/backend/static"
cp -r dist/* "$ROOT/backend/static/"

echo "==> 重新构建后端（含静态资源）"
cd "$ROOT/backend"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "$RELEASE_DIR/ynxcb-server" ./cmd/server

echo "==> 复制部署文件"
cp "$ROOT/deploy/config.json.example" "$RELEASE_DIR/config.json"
cp "$ROOT/deploy/backup.sh" "$RELEASE_DIR/"
cp "$ROOT/deploy/systemd/ynxcb.service" "$RELEASE_DIR/"
cp "$ROOT/deploy/nginx/ynxcb.conf" "$RELEASE_DIR/"
cp "$ROOT/deploy/DEPLOY.md" "$RELEASE_DIR/README.md"
chmod +x "$RELEASE_DIR/ynxcb-server" "$RELEASE_DIR/backup.sh"

echo "==> 打包"
cd "$ROOT"
tar -czf "ynxcb-release-$VERSION.tar.gz" -C "$RELEASE_DIR" .

echo "完成: $ROOT/ynxcb-release-$VERSION.tar.gz"
echo "上传到服务器后按 README.md 部署"
