# 构建阶段
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w -X main.version=${VERSION}" \
    -o /mymihomo ./cmd/mymihomo/

# 运行阶段
FROM metacubex/mihomo:latest

ADD https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.metadb /root/.config/mihomo/geoip.metadb
ADD https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.dat /root/.config/mihomo/geosite.dat
ADD https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.dat /root/.config/mihomo/geoip.dat

ADD https://github.com/eorendel/clash-dashboard/archive/refs/heads/main.zip  /tmp/dashboard.zip
ADD https://github.com/haishanh/yacd/releases/download/v0.3.8/yacd.tar.xz  /tmp/yacd.tar.xz
ADD https://github.com/Zephyruso/zashboard/releases/download/v2.7.0/dist.zip /tmp/zashboard.zip

COPY --from=builder /mymihomo /bin/mymihomo
COPY ./run.sh /bin/run

# 常用配置（其他配置见 README.md）
ENV MIXED_PORT="7890"
ENV CRON_EXPRESSION="1 * * * *"
ENV CONF_FILE=/root/conf/config.yaml
ENV MIHOMO_HOME=/root/.config/mihomo
ENV WEB_CONFIG_PORT=18080
ENV WEB_ENV_FILE=/root/conf/web.env
ENV WEB_API_LOG=/root/conf/web_api.log
ENV CUSTOM_CONF=/root/.config/mihomo/custom.yaml

RUN chmod +x /bin/run \
    && chmod +x /bin/mymihomo \
    && mkdir -p /root/conf \
    && mkdir -p /root/.config/mihomo/ui \
    && cd /tmp && unzip dashboard.zip \
    && mv clash-dashboard-main /root/.config/mihomo/ui/dashboard \
    && tar -Jxf yacd.tar.xz \
    && mv public /root/.config/mihomo/ui/yacd \
    && mkdir -p /tmp/zashboard \
    && unzip -q zashboard.zip -d /tmp/zashboard \
    && mkdir -p /root/.config/mihomo/ui/zashboard \
    && if [ -d /tmp/zashboard/dist ]; then cp -r /tmp/zashboard/dist/. /root/.config/mihomo/ui/zashboard/; else cp -r /tmp/zashboard/. /root/.config/mihomo/ui/zashboard/; fi \
    && rm -rf /tmp/zashboard \
    && rm -f dashboard.zip yacd.tar.xz zashboard.zip

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -q -T 3 -O - "http://127.0.0.1:${EXTERNAL_PORT:-9090}/version" >/dev/null || exit 1

ENTRYPOINT ["/bin/run"]
