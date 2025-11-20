#!/usr/bin/env sh
# 下载配置文件

CONF_FILE=$1
CONF_PATH=$(dirname "$CONF_FILE")
mkdir -p "${CONF_PATH}"

echo "从${CONF_URL} 下载配置文件并写入到${CONF_FILE}"
curl -L --compressed -A "Wget/1.21.3" "$CONF_URL"  -o "$CONF_FILE"  --retry 3

# 若文件下载失败, 则返回并报错
if [ $? -eq 0 ]; then
  if [ "$BASE64_CONVERT" = "true" ]; then
      echo "配置文件转换中..."
      mv "${CONF_FILE}" "${CONF_FILE}".ori && base64 -d "${CONF_FILE}".ori > "${CONF_FILE}"
  fi
else
  echo "配置文件下载失败"
  exit $?
fi


echo "修改配置文件中的ui界面指向路径"
yq eval '.external-controller = env(EXTERNAL_BIND) + ":" + env(EXTERNAL_PORT)' -i "$CONF_FILE"
yq eval '.external-ui = "/root/clash-dashboard"'  -i "$CONF_FILE"

# 鉴权信息
if [ ! -z "$EXTERNAL_SECRET" ];
then
  #modify_config "secret" "$EXTERNAL_SECRET"
  yq eval '.secret = strenv(EXTERNAL_SECRET)' -i "$CONF_FILE"
fi

# 必须开启局域网连接, 否则外部无法连接
#modify_config "allow-lan" "true"
yq eval '.allow-lan = true' -i "$CONF_FILE"
yq eval '.mixed-port = env(SOCKET_PORT)' -i "$CONF_FILE"
yq eval '.port = env(SOCKET_PORT)'  -i "$CONF_FILE"

yq eval 'del(.port)' -i "$CONF_FILE"
yq eval 'del(.socks-port)' -i "$CONF_FILE"

if [ -f "${CUSTOM_CONF}" ]; then
  echo  "合并自定义规则文件...${CUSTOM_CONF} ${CONF_FILE}"
  yq eval-all '
    select(fileIndex == 0) as $first |
    select(fileIndex == 1) as $second |
    ($first *+? $second) |
    .rules = ($first.rules + $second.rules)
  ' "$CONF_FILE"  "$CUSTOM_CONF" -i "$CONF_FILE"
fi

