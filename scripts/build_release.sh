#!/usr/bin/env bash
# 编译「智领进销存」为发布包（默认 Linux amd64，静态编译，无需目标机装 Go）
# 用法：
#   bash scripts/build_release.sh                 # 默认 linux/amd64
#   GOARCH=arm64 bash scripts/build_release.sh    # linux/arm64
#   GOOS=darwin GOARCH=arm64 bash scripts/build_release.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"
NAME="zhiling-erp_${GOOS}_${GOARCH}"
OUT="dist"
PKG="${OUT}/${NAME}"

echo ">> 编译 ${GOOS}/${GOARCH} ..."
rm -rf "$PKG"
mkdir -p "$PKG"

CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
  go build -trimpath -ldflags "-s -w" -o "$PKG/pisa_server" ./cmd/server

echo ">> 打包静态资源与配置示例 ..."
cp -r web "$PKG/web"
# 不携带真实密钥，仅示例
cp config.yaml.example "$PKG/config.yaml.example" 2>/dev/null || true

cat > "$PKG/README.txt" <<'EOF'
部署：
  1) 服务器安装并启动 MySQL，创建数据库 pisa
  2) cp config.yaml.example config.yaml  并填写数据库/JWT
  3) ./pisa_server                        # 首次启动自动建表、初始化 admin/123456
工作目录：程序按 ./web 和 ./config.yaml 读取，请在包根目录运行。
EOF

tar -C "$OUT" -czf "${OUT}/${NAME}.tar.gz" "$NAME"
rm -rf "$PKG"

echo ">> 完成：${OUT}/${NAME}.tar.gz"
ls -lh "${OUT}/${NAME}.tar.gz"
