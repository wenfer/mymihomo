#!/usr/bin/env sh

set -eu

log() {
  printf '[mymihomo] %s\n' "$1"
}

fatal() {
  printf '[mymihomo] %s\n' "$1" >&2
  exit 1
}

WEB_ENV_FILE="${WEB_ENV_FILE:-/root/conf/web.env}"

# 1. 优先加载持久化的环境变量覆盖 (web.env)
if [ -f "$WEB_ENV_FILE" ]; then
  while IFS='=' read -r key val || [ -n "$key" ]; do
    key=$(echo "$key" | tr -d '[:space:]')
    case "$key" in
      \#*|"") continue ;;
      *)
        val="${val#\"}"
        val="${val%\"}"
        export "$key=$val"
        ;;
    esac
  done < "$WEB_ENV_FILE"
  log "已从持久化文件加载配置: $WEB_ENV_FILE"
fi

CONF_FILE="${CONF_FILE:-/root/conf/config.yaml}"
MIHOMO_HOME="${MIHOMO_HOME:-/root/.config/mihomo}"
CRON_LOG="${CRON_LOG:-/root/conf/cron_history}"
WEB_CONFIG_PORT="${WEB_CONFIG_PORT:-18080}"
WEB_API_LOG="${WEB_API_LOG:-/root/conf/web_api.log}"

# 2. 设置系统时区
if [ -n "${TZ:-}" ] && [ -f "/usr/share/zoneinfo/${TZ}" ]; then
  ln -snf "/usr/share/zoneinfo/${TZ}" /etc/localtime
  echo "${TZ}" > /etc/timezone
  log "系统时区已设置为: ${TZ}"
fi

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

mkdir -p "$(dirname "$CONF_FILE")" "$MIHOMO_HOME" "$(dirname "$CRON_LOG")" "$(dirname "$WEB_API_LOG")"

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
  fi
fi

mymihomo serve -addr ":${WEB_CONFIG_PORT}" -c "$CONF_FILE" -e "$WEB_ENV_FILE" >> "$WEB_API_LOG" 2>&1 &
CONFIG_API_PID="$!"
sleep 1
if ! kill -0 "$CONFIG_API_PID" >/dev/null 2>&1; then
  fatal "配置 API 启动失败，请检查日志: ${WEB_API_LOG}"
fi
log "配置 API 已启动: ${WEB_CONFIG_PORT}，持久化文件: ${WEB_ENV_FILE}"

# 同步定时更新任务并确保 crond 运行 (引导模式下会自动安全跳过)
if ! mymihomo sync-cron; then
  log "警告: 初始化同步定时任务未完成，将在后续配置生效时更新"
fi

exec /mihomo -d "$MIHOMO_HOME" -f "$CONF_FILE"
