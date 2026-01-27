#!/usr/bin/env sh

CONF_FILE="/root/conf/config.yaml"

mymihomo download -o $CONF_FILE
if [ $? -ne 0 ]; then
  echo "启动失败: 配置文件下载失败"
  exit 1
fi

# 启动定时下载配置文件
if [ ! -z "$CRON_EXPRESSION" ]; then
  CRON_EXPRESSION="${CRON_EXPRESSION:-'1 * * * *'}"
  SCRIPT="mymihomo update -c $CONF_FILE >> /root/conf/cron_history 2>&1 &"
  crontab -r || true
  crond -f &
  echo "$CRON_EXPRESSION $SCRIPT" | crontab -
fi

# 启动代理
/mihomo -d /root/.config/mihomo/ -f $CONF_FILE
