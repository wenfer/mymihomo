FROM metacubex/mihomo:latest


ADD https://cdn.jsdelivr.net/gh/Dreamacro/maxmind-geoip@release/Country.mmdb /root/files/Country.mmdb
ADD https://github.com/eorendel/clash-dashboard/archive/refs/heads/main.zip  /tmp/dashboard.zip
ADD https://github.com/haishanh/yacd/releases/download/v0.3.8/yacd.tar.xz  /tmp/yacd.tar.xz
COPY ./run.sh /bin/run
COPY ./dl-clash-conf.sh /bin/dl-clash-conf
COPY ./index.html /root/clash-dashboard/index.html
COPY ./update-clash-conf.sh /bin/update-clash-conf

# 配置文件地址
ENV CONF_URL=http://test.com
# RESTful API 地址, 可为空
ENV EXTERNAL_BIND="0.0.0.0"
ENV EXTERNAL_PORT="9090"
# RESTful API 鉴权
ENV EXTERNAL_SECRET="123456"
ENV CRON_EXPRESSION="1 * * * *"
ENV SOCKET_PORT=7890
ENV BASE64_CONVERT=false
ENV CUSTOM_CONF=/root/conf/custom.yaml

RUN chmod +x /bin/run \
    && chmod +x /bin/update-clash-conf \
    && chmod +x /bin/dl-clash-conf \
    && unzip /tmp/dashboard.zip -d /tmp \
    && mv /tmp/clash-dashboard-main  /root/clash-dashboard/dashboard \
    && tar -Jxf /tmp/yacd.tar.xz -C /tmp \
    && mv /tmp/public /root/clash-dashboard/yacd \
    && apk add yq curl

ENTRYPOINT ["run"]
