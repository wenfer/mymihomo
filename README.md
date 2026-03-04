# mymihomo

基于 [mihomo](https://github.com/MetaCubeX/mihomo) 的 Docker 镜像，提供订阅自动下载、配置覆写和定时更新功能。

## 功能特性

- 自动下载并处理订阅配置
- 定时任务自动更新订阅
- 集成 Web UI 界面 (yacd / clash-dashboard / zashboard)
- 集成 Web 配置中心（可在页面内保存并应用配置）
- 支持 HTTP / SOCKS5 / Mixed 代理模式
- 支持 TUN 模式透明代理
- 支持自定义规则覆写
- 支持 Base64 订阅解码
- 多架构支持 (amd64 / arm64)

## 快速开始

```yaml
# docker-compose.yml
services:
  mymihomo:
    image: ghcr.io/wenfer/mymihomo:latest
    container_name: mymihomo
    restart: unless-stopped
    ports:
      - 9090:9090  # Web UI
      - 7890:7890  # 代理端口
      - 18080:18080 # 配置 API（导航页配置表单依赖）
    volumes:
      - ./conf:/root/conf
    environment:
      - TZ=Asia/Shanghai
      - CONF_URL=https://your-subscription-url  # 替换为你的订阅地址
      - EXTERNAL_SECRET=your-password           # Web UI 密码
```

```bash
docker-compose up -d
```

访问 `http://ip:9090` 进入 Web UI。

如果未设置 `CONF_URL`，容器会进入引导模式并正常启动 UI。  
此时可在首页“运行配置”表单填写订阅并点击“保存并应用”完成初始化。

## 配置示例

### 基础配置

```yaml
services:
  mymihomo:
    image: ghcr.io/wenfer/mymihomo:latest
    container_name: mymihomo
    restart: unless-stopped
    ports:
      - 9090:9090   # Web UI
      - 7890:7890   # Mixed 代理 (HTTP + SOCKS5)
      - 18080:18080 # 配置 API
    volumes:
      - ./conf:/root/conf
    environment:
      - TZ=Asia/Shanghai
      - CONF_URL=https://your-subscription-url
      - EXTERNAL_SECRET=123456
      - CRON_EXPRESSION=0 2 * * *  # 每天凌晨2点更新订阅
```

### 分离端口配置

```yaml
services:
  mymihomo:
    image: ghcr.io/wenfer/mymihomo:latest
    container_name: mymihomo
    restart: unless-stopped
    ports:
      - 9090:9090   # Web UI
      - 7890:7890   # HTTP 代理
      - 7891:7891   # SOCKS5 代理
      - 18080:18080 # 配置 API
    volumes:
      - ./conf:/root/conf
    environment:
      - TZ=Asia/Shanghai
      - CONF_URL=https://your-subscription-url
      - EXTERNAL_SECRET=123456
      - HTTP_PORT=7890
      - SOCKS_PORT=7891
      - MIXED_PORT=        # 留空禁用 mixed 端口
```

### TUN 模式 (透明代理)

```yaml
services:
  mymihomo:
    image: ghcr.io/wenfer/mymihomo:latest
    container_name: mymihomo
    restart: unless-stopped
    privileged: true        # TUN 模式必须
    network_mode: host      # 推荐使用 host 网络
    volumes:
      - ./conf:/root/conf
    environment:
      - TZ=Asia/Shanghai
      - CONF_URL=https://your-subscription-url
      - EXTERNAL_SECRET=123456
      - TUN_ENABLE=true
      - TUN_STACK=system    # system / gvisor / mixed
```

### 自定义规则覆写

```yaml
services:
  mymihomo:
    image: ghcr.io/wenfer/mymihomo:latest
    container_name: mymihomo
    restart: unless-stopped
    ports:
      - 9090:9090
      - 7890:7890
      - 18080:18080 # 配置 API
    volumes:
      - ./conf:/root/conf
    environment:
      - TZ=Asia/Shanghai
      - CONF_URL=https://your-subscription-url
      - EXTERNAL_SECRET=123456
      - CUSTOM_CONF=/root/conf/custom.yaml
```

自定义规则文件 `./conf/custom.yaml`:

```yaml
rules:
  - DOMAIN-SUFFIX,openai.com,Proxy
  - DOMAIN-SUFFIX,github.com,Proxy
  - DOMAIN-KEYWORD,google,Proxy
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `CONF_URL` | 订阅地址 | **必填** |
| `CONF_FILE` | 容器内配置文件路径 | /root/conf/config.yaml |
| `MIHOMO_HOME` | mihomo 工作目录 | /root/.config/mihomo |
| `WEB_CONFIG_PORT` | Web 配置 API 监听端口 | 18080 |
| `WEB_ENV_FILE` | Web 配置持久化文件路径 | /root/conf/web.env |
| `WEB_API_LOG` | Web 配置 API 日志文件 | /root/conf/web_api.log |
| `EXTERNAL_BIND` | API 绑定地址 | 0.0.0.0 |
| `EXTERNAL_PORT` | Web UI 端口 | 9090 |
| `EXTERNAL_SECRET` | Web UI 密码 | 123456 |
| `CRON_EXPRESSION` | 定时更新表达式（留空禁用） | 1 * * * * |
| `CRON_LOG` | 定时更新日志文件 | /root/conf/cron_history |
| `CONF_TIMEOUT_SEC` | 订阅下载超时秒数 | 15 |
| `CONF_RETRY` | 订阅下载重试次数 | 2 |
| `HTTP_PORT` | HTTP 代理端口 | - |
| `SOCKS_PORT` | SOCKS5 代理端口 | - |
| `MIXED_PORT` | 混合代理端口 | 7890 |
| `TUN_ENABLE` | 启用 TUN 模式 | false |
| `TUN_STACK` | TUN 协议栈 | system |
| `TUN_AUTO_ROUTE` | 自动设置路由 | true |
| `TUN_AUTO_DETECT` | 自动检测网卡 | true |
| `BASE64_CONVERT` | Base64 解码订阅 | false |
| `CUSTOM_CONF` | 自定义规则文件 | /root/conf/custom.yaml |

## 更新日志

### 2026/3/4
- 优化容器运行环境：新增启动前配置校验、健康检查与可配置运行路径
- 修复定时更新在默认 `EXTERNAL_PORT` 配置下不生效的问题
- 增强订阅下载稳定性（超时与重试控制）
- 导航页升级为配置中心，支持页面内保存并应用配置
- 新增 zashboard 前端（v2.7.0）

### 2025/1/27
- 使用 Go 重构配置处理工具
- 新增 HTTP / SOCKS5 / Mixed 端口独立配置
- 新增 TUN 模式支持
- 移除 yq 依赖，减小镜像体积

### 2025/7/10
- 新增 yacd 界面和界面切换导航页
- 新增 Base64 规则转换
- 新增自定义覆写规则功能

### 2025/5/24
- 新增 arm64 镜像支持
