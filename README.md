
## 示例

- 定时任务自动更新订阅
- 集成ui界面 yacd
- 自动重写config.yaml配置用于局域网代理


```yaml
services:
  myclash:
    image: ghcr.io/wenfer/myclash:latest
    container_name: myclash
    restart: unless-stopped
    ports:
      - 9090:9090 # ui端口
      - 7890:7890 # 代理端口
    volumes:
      - ./conf:/root/conf
    environment:
      - TZ=Asia/Shanghai
      - CRON_EXPRESSION=0 2 * * * # 每天凌晨2点 定时任务更新订阅
      - CONF_URL=https://test.com # 务必替换成你的订阅地址
      - EXTERNAL_BIND=0.0.0.0 # 绑定网卡ip，可以不写
      - EXTERNAL_PORT=9090  # ui端口
      - EXTERNAL_SECRET=123456  # ui密码
      # - BASE64_CONVERT=true   如果需要base64转换，放开这个配置
      # - CUSTOM_CONF=/root/conf/custom.yaml  # 如果需要自定义扩展规则，添加这个自定义文件，环境变量是默认的，需要修改路径再放开配置

```


# 2025/7/10更新
- 新增了yacd界面和界面切换导航页
- 新增了base64的规则转换，因为我没有响应的规则，所以我没测试，有问题可以issues
- 新增一个自定义覆写规则的功能，目前只支持覆写rules规则，示例：
/root/conf/custom.yaml
```yaml
rules:
  - DOMAIN-SUFFIX,wsj.com,Proxy
```


# 2025/5/24更新
新增了arm64的镜像，需要换成ghcr.io的仓库，dockerhub拉不到

