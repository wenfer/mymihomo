# mymihomo

基于 mihomo 的 Docker 镜像构建工具，提供订阅下载、配置修改和定时更新功能。

## 项目结构

```
.
├── cmd/mymihomo/          # Go 源码
│   ├── main.go           # 入口，命令行解析
│   ├── download.go       # 下载并处理配置文件
│   ├── update.go         # 更新运行中的 mihomo 配置
│   └── render.go         # 渲染导航页模板
├── run.sh                # 容器启动脚本
├── Dockerfile            # 多阶段构建
└── .github/workflows/    # CI/CD
```

## 开发命令

```bash
# 构建
go build -o mymihomo ./cmd/mymihomo/

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o mymihomo-linux-amd64 ./cmd/mymihomo/
GOOS=linux GOARCH=arm64 go build -o mymihomo-linux-arm64 ./cmd/mymihomo/

# 本地 Docker 构建测试
docker build -t mymihomo:test .
```

## CLI 用法

```bash
# 下载配置
mymihomo download -o /root/conf/config.yaml

# 更新运行中的配置
mymihomo update -c /root/conf/config.yaml

# 查看版本
mymihomo version
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| CONF_URL | 订阅地址 | 必填 |
| CONF_FILE | 容器内配置文件路径 | /root/conf/config.yaml |
| MIHOMO_HOME | mihomo 工作目录 | /root/.config/mihomo |
| WEB_CONFIG_PORT | Web 配置 API 监听端口 | 18080 |
| WEB_ENV_FILE | Web 配置持久化文件路径 | /root/conf/web.env |
| WEB_API_LOG | Web 配置 API 日志文件 | /root/conf/web_api.log |
| EXTERNAL_BIND | API 绑定地址 | 0.0.0.0 |
| EXTERNAL_PORT | API 端口 | 9090 |
| EXTERNAL_SECRET | API 鉴权密钥 | - |
| CONF_TIMEOUT_SEC | 订阅下载超时秒数 | 15 |
| CONF_RETRY | 订阅下载重试次数 | 2 |
| CRON_LOG | 定时更新日志文件 | /root/conf/cron_history |
| HTTP_PORT | HTTP 代理端口 | - |
| SOCKS_PORT | SOCKS5 代理端口 | - |
| MIXED_PORT | 混合代理端口 (HTTP+SOCKS5) | 7890 |
| TUN_ENABLE | 启用 TUN 模式 | false |
| TUN_STACK | TUN 协议栈 (system/gvisor/mixed) | system |
| TUN_AUTO_ROUTE | 自动设置路由 | true |
| TUN_AUTO_DETECT | 自动检测网卡 | true |
| BASE64_CONVERT | 是否 base64 解码 | false |
| CUSTOM_CONF | 自定义规则文件路径 | /root/conf/custom.yaml |
| CRON_EXPRESSION | 定时更新表达式 | 1 * * * * |

## 依赖

- Go 1.23+
- gopkg.in/yaml.v3 (唯一外部依赖)

## 发布流程

推送 tag 触发 GitHub Actions 自动构建：
```bash
git tag v1.0.0
git push origin v1.0.0
```

会自动构建并发布：
- Docker 多架构镜像 (amd64/arm64)
- 二进制文件 Release

更新当前文档条件
- 本文档有与实际不符或冲
- 新增源码文件
- 重构项目或文件更新定义
