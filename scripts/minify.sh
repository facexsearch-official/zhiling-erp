#!/usr/bin/env bash
# 压缩前端脚本：web/js/index-app.js -> web/js/index-app.min.js
# 依赖：node + terser（npx 自动获取）
# 用法： bash scripts/minify.sh
#
# 注意：不启用 --mangle toplevel，以保留全局函数名（HTML 里的 onclick 直接调用）。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

IN="web/js/index-app.js"
OUT="web/js/index-app.min.js"
[ -f "$IN" ] || { echo "缺少 $IN"; exit 1; }

echo ">> terser 压缩 $IN ..."
npx --yes terser "$IN" -c passes=2 -m -o "$OUT"

echo ">> 结果："
ls -l "$IN" "$OUT"
printf "   gzip: %s -> %s bytes\n" "$(gzip -c "$IN" | wc -c | tr -d ' ')" "$(gzip -c "$OUT" | wc -c | tr -d ' ')"
node --check "$OUT" && echo ">> 语法 OK"
echo ">> 记得把 index.html 的引用版本号 ?v= 更新一下。"
