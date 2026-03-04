#!/usr/bin/env sh

set -eu

CONF_FILE="${CONF_FILE:-/root/conf/config.yaml}"
MIHOMO_HOME="${MIHOMO_HOME:-/root/.config/mihomo}"
CRON_LOG="${CRON_LOG:-/root/conf/cron_history}"
WEB_CONFIG_PORT="${WEB_CONFIG_PORT:-18080}"
WEB_ENV_FILE="${WEB_ENV_FILE:-/root/conf/web.env}"
WEB_API_LOG="${WEB_API_LOG:-/root/conf/web_api.log}"

log() {
  printf '[mymihomo] %s\n' "$1"
}

fatal() {
  printf '[mymihomo] %s\n' "$1" >&2
  exit 1
}

write_bootstrap_config() {
  bind_addr="${EXTERNAL_BIND:-0.0.0.0}"
  ext_port="${EXTERNAL_PORT:-9090}"
  mixed_port="${MIXED_PORT:-7890}"

  cat > "$CONF_FILE" <<CFG
mixed-port: ${mixed_port}
allow-lan: true
mode: Rule
log-level: info
external-controller: ${bind_addr}:${ext_port}
external-ui: ${MIHOMO_HOME}/ui
proxies: []
proxy-groups: []
rules:
  - MATCH,DIRECT
CFG
}

for required_cmd in mymihomo crond crontab; do
  command -v "$required_cmd" >/dev/null 2>&1 || fatal "缺少必要命令: $required_cmd"
done
[ -x /mihomo ] || fatal "缺少必要命令: /mihomo"

mkdir -p "$(dirname "$CONF_FILE")" "$MIHOMO_HOME" "$(dirname "$CRON_LOG")"

BOOTSTRAP_MODE="false"
TMP_CONF="${CONF_FILE}.download"
if mymihomo download -o "$TMP_CONF" && /mihomo -d "$MIHOMO_HOME" -f "$TMP_CONF" -t; then
  mv -f "$TMP_CONF" "$CONF_FILE"
else
  rm -f "$TMP_CONF" >/dev/null 2>&1 || true
  if [ -s "$CONF_FILE" ] && /mihomo -d "$MIHOMO_HOME" -f "$CONF_FILE" -t; then
    log "订阅拉取或校验失败，继续使用已有配置。"
  else
    log "未获取到有效订阅，进入引导模式。"
    write_bootstrap_config
    BOOTSTRAP_MODE="true"
  fi
fi

mymihomo serve -addr ":${WEB_CONFIG_PORT}" -c "$CONF_FILE" -e "$WEB_ENV_FILE" >> "$WEB_API_LOG" 2>&1 &
CONFIG_API_PID="$!"
sleep 1
if ! kill -0 "$CONFIG_API_PID" >/dev/null 2>&1; then
  fatal "配置 API 启动失败，请检查日志: ${WEB_API_LOG}"
fi
log "配置 API 已启动: ${WEB_CONFIG_PORT}，持久化文件: ${WEB_ENV_FILE}"

if [ "$BOOTSTRAP_MODE" = "false" ] && [ -n "${CRON_EXPRESSION:-}" ]; then
  crontab -r >/dev/null 2>&1 || true
  echo "$CRON_EXPRESSION mymihomo update -c $CONF_FILE >> $CRON_LOG 2>&1" | crontab -
  crond
  log "已启用定时更新: ${CRON_EXPRESSION}"
else
  log "当前处于引导模式或未设置 CRON_EXPRESSION，跳过定时更新"
fi

exec /mihomo -d "$MIHOMO_HOME" -f "$CONF_FILE"
