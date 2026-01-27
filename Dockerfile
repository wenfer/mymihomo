# 运行阶段
FROM metacubex/mihomo:latest

ARG TARGETARCH

ADD https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.metadb /root/.config/mihomo/geoip.metadb
ADD https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.dat /root/.config/mihomo/geosite.dat
ADD https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.dat /root/.config/mihomo/geoip.dat

ADD https://github.com/eorendel/clash-dashboard/archive/refs/heads/main.zip  /tmp/dashboard.zip
ADD https://github.com/haishanh/yacd/releases/download/v0.3.8/yacd.tar.xz  /tmp/yacd.tar.xz

COPY dist/myclash-linux-${TARGETARCH} /bin/myclash
COPY ./run.sh /bin/run

# 常用配置（其他配置见 README.md）
ENV MIXED_PORT="7890"
ENV CRON_EXPRESSION="1 * * * *"
ENV CUSTOM_CONF=/root/.config/mihomo/custom.yaml

RUN chmod +x /bin/run \
    && chmod +x /bin/myclash \
    && mkdir -p /root/.config/mihomo/ui \
    && cd /tmp && unzip dashboard.zip \
    && mv clash-dashboard-main /root/.config/mihomo/ui/dashboard \
    && tar -Jxf yacd.tar.xz \
    && mv public /root/.config/mihomo/ui/yacd \
    && rm -f dashboard.zip yacd.tar.xz

ENTRYPOINT ["run"]
