#!/usr/bin/env sh

CONF_FILE="/root/conf/config.yaml"

echo "修改导航页的连接"
sed -i "s|\${EXTERNAL_PORT}|${EXTERNAL_PORT}|g" /root/files/clash-dashboard/index.html
sed -i "s|\${EXTERNAL_SECRET}|${EXTERNAL_SECRET}|g" /root/files/clash-dashboard/index.html
echo "导航页生成完毕"

dl-clash-conf $CONF_FILE
# 启动定时下载配置文件
if [ ! -z "$CRON_EXPRESSION" ]; then
  # 设置环境变量
  CRON_EXPRESSION="${CRON_EXPRESSION:-'1 * * * *'}"
  SCRIPT="update-clash-conf $CONF_FILE  >> /root/conf/cron_history 2>&1 &"
  # 清除现有的定时任务
  crontab -r || true
  crond -f &
  # 添加新的定时任务
  echo "$CRON_EXPRESSION $SCRIPT" | crontab -
fi

# 启动代理
/mihomo -d /root/files/ -f $CONF_FILE

