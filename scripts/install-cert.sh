#!/usr/bin/env bash
# 签发/安装 facexsearch.com 证书，并重启 nginx（定时任务请自行配置）
# 用法：
#   export Ali_Key="AccessKeyId"; export Ali_Secret="AccessKeySecret"
#   bash install-cert.sh            # 证书不存在则签发，存在则复用
#   bash install-cert.sh --force    # 强制重新签发（域名/SAN 变更时用）
set -euo pipefail

# ── 配置 ──
ACME="/root/.acme.sh/acme.sh"
MAIN_DOMAIN="facexsearch.com"
DOMAINS=(-d "$MAIN_DOMAIN" -d "*.${MAIN_DOMAIN}")
KEYLENGTH="ec-256"
CERT_DIR="/root/.acme.sh/${MAIN_DOMAIN}_ecc"
KEY_FILE="${CERT_DIR}/${MAIN_DOMAIN}.key"
FULLCHAIN_FILE="${CERT_DIR}/fullchain.cer"
RELOAD_CMD="systemctl restart nginx"
DNS_PLUGIN="dns_ali"

FORCE=""
[ "${1:-}" = "--force" ] && FORCE="--force"

# ── 前置检查 ──
[ -x "$ACME" ] || { echo "错误：找不到 $ACME，请先安装 acme.sh"; exit 1; }

# ── 1) 签发（DNS-01，自动改 DNS 记录） ──
if [ -n "$FORCE" ] || [ ! -f "$FULLCHAIN_FILE" ]; then
  if [ "$DNS_PLUGIN" = "dns_ali" ]; then
    : "${Ali_Key:?请先 export Ali_Key=... }"
    : "${Ali_Secret:?请先 export Ali_Secret=... }"
  fi
  echo ">> 签发证书 ${DOMAINS[*]} ..."
  "$ACME" --issue --dns "$DNS_PLUGIN" --keylength "$KEYLENGTH" $FORCE "${DOMAINS[@]}"
else
  echo ">> 已存在证书，复用：$FULLCHAIN_FILE"
fi

# ── 2) 安装证书（续期时执行下方的 reloadcmd） ──
echo ">> 安装证书 ..."
"$ACME" --install-cert -d "$MAIN_DOMAIN" --ecc \
  --key-file       "$KEY_FILE" \
  --fullchain-file "$FULLCHAIN_FILE" \
  --reloadcmd      "$RELOAD_CMD"

# ── 3) 校验并重启 nginx ──
echo ">> 校验并重启 nginx ..."
nginx -t
$RELOAD_CMD

# ── 4) 验证 ──
echo ">> 证书 SAN："
openssl x509 -in "$FULLCHAIN_FILE" -noout -subject -ext subjectAltName || true
echo ">> 完成。"
