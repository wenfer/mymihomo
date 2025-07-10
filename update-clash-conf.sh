#!/usr/bin/env sh
# 更新配置文件
ConfFile=$1

# 不需要更新
if [ -z "$EXTERNAL_BIND" || -z "$EXTERNAL_PORT" ];
then
  exit 0
fi
# 下载文件
dl-clash-conf $ConfFile
# 若文件下载失败, 则返回并报错
if [ $? -ne 0 ];
then
  exit $?
fi
# 文件下载成功, 进行更新
if [ ! -z "$EXTERNAL_SECRET" ];
then
  curl -H "Authorization: Bearer $EXTERNAL_SECRET" -X PUT -d "{\"path\": \"$ConfFile\"}" 127.0.0.1:$EXTERNAL_PORT/configs?force=true
else
  curl -X PUT -d "{\"path\": \"$ConfFile\"}" 127.0.0.1:$EXTERNAL_PORT/configs?force=true
fi

