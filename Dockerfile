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
    -o /myclash ./cmd/myclash/

# 运行阶段
FROM metacubex/mihomo:latest

ADD https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.metadb /root/files/geoip.metadb
ADD https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.dat /root/files/geosite.dat
ADD https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/country.mmdb /root/files/Country.mmdb
ADD https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.dat /root/files/geoip.dat

ADD https://github.com/eorendel/clash-dashboard/archive/refs/heads/main.zip  /tmp/dashboard.zip
ADD https://github.com/haishanh/yacd/releases/download/v0.3.8/yacd.tar.xz  /tmp/yacd.tar.xz

COPY --from=builder /myclash /bin/myclash
COPY ./run.sh /bin/run

# 配置文件地址
ENV CONF_URL=http://test.com
# RESTful API 地址
ENV EXTERNAL_BIND="0.0.0.0"
ENV EXTERNAL_PORT="9090"
# RESTful API 鉴权
ENV EXTERNAL_SECRET="123456"
ENV CRON_EXPRESSION="1 * * * *"
# 代理端口配置
ENV HTTP_PORT=""
ENV SOCKS_PORT=""
ENV MIXED_PORT="7890"
# TUN 模式配置
ENV TUN_ENABLE=false
ENV TUN_STACK=system
ENV TUN_AUTO_ROUTE=true
ENV TUN_AUTO_DETECT=true
# 其他配置
ENV BASE64_CONVERT=false
ENV CUSTOM_CONF=/root/conf/custom.yaml

RUN chmod +x /bin/run \
    && chmod +x /bin/myclash \
    && unzip /tmp/dashboard.zip -d /tmp \
    && mv /tmp/clash-dashboard-main  /root/files/clash-dashboard/dashboard \
    && tar -Jxf /tmp/yacd.tar.xz -C /tmp \
    && mv /tmp/public /root/files/clash-dashboard/yacd \
    && rm -f /tmp/dashboard.zip /tmp/yacd.tar.xz

ENTRYPOINT ["run"]
