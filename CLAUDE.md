# mymihomo

基于 mihomo 的 Docker 镜像构建工具，提供订阅下载、配置修改、导航页渲染、配置 API 服务和定时更新功能。辅助二进制工具使用 Rust 编写。

## 项目结构

```
.
├── src/                  # Rust 源码
│   ├── main.rs           # 入口，CLI 命令行解析与分发
│   ├── download.rs       # 下载并处理配置文件（YAML修改/规则合并）
│   ├── update.rs         # 调用 REST API 更新运行中的 mihomo 配置
│   ├── render.rs         # 渲染导航页模板
│   ├── serve.rs          # Axum Web 配置 API 服务
│   └── envfile.rs        # web.env 持久化配置读写
├── Cargo.toml            # Rust 包管理与依赖
├── Cargo.lock            # 依赖锁定文件
├── run.sh                # 容器启动脚本
├── Dockerfile            # 多阶段构建 (rust:alpine -> metacubex/mihomo)
└── .github/workflows/    # CI/CD (多架构镜像构建与 GitHub Release)
```

## 开发命令

```bash
# 检查与运行测试
cargo check
cargo test

# 构建 Release
cargo build --release

# 交叉编译 (使用 cargo-zigbuild)
cargo zigbuild --release --target x86_64-unknown-linux-musl
cargo zigbuild --release --target aarch64-unknown-linux-musl

# 本地 Docker 构建测试
docker build -t mymihomo:test .
```

## CLI 用法

```bash
# 下载配置
mymihomo download -o /root/conf/config.yaml

# 更新运行中的配置
mymihomo update -c /root/conf/config.yaml

# 渲染导航页
mymihomo render -o /root/.config/mihomo/ui/index.html

# 启动配置 API 服务
mymihomo serve --addr :18080 -c /root/conf/config.yaml -e /root/conf/web.env

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

- Rust 1.75+ (Edition 2021)
- tokio, axum, reqwest, serde, serde_yaml, clap

## 发布流程

推送 tag 触发 GitHub Actions 自动构建：
```bash
git tag v1.0.0
git push origin v1.0.0
```

会自动构建并发布：
- Docker 多架构镜像 (amd64/arm64)
- 二进制文件 Release
