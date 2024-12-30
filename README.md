# Cert

一个简单的证书申请&续签工具。

## 部署

- Docker-cli

```bash
docker run -d \
  --name cert \
  --restart always \
  -v ./cert/config:/data/cert/config \
  -v ./cert/log:/data/cert/log \
  -v ./cert/output:/data/cert/output \
  wjqserver/cert:latest
```

- Docker-compose

```yaml
version: '3'
services:
  cert:
    image: 'wjqserver/cert:latest'
    restart: always
    volumes:
      - './cert/config:/data/cert/config'
      - './cert/log:/data/cert/log'
      - './cert/output:/data/cert/output'
```

## 配置

对./cert/config/config.toml进行配置

```toml
[log]
logfilepath = "/data/cert/log/cert.log"  # 日志文件路径
maxlogsize = 5 # 日志文件最大大小，单位MB

[account]
email = "demo@example.com" # 申请证书的邮箱
token = "" # cloudflare API token

[path]
cert = "/data/cert/output/cert.crt" # 证书存储路径
key = "/data/cert/output/cert.key" # 私钥存储路径
caCert = "/data/cert/output/ca.crt" # CA证书存储路径
json = "/data/cert/output/cert.json" # 证书信息存储路径

[domain]
name = "*.example.com" # 需要申请证书的域名

```

## 运行

经过上述步骤,程序会自动申请证书并存储到./cert/output目录下。程序会自动根据`cert.json`文件中的`notBefore`和`notAfter`字段自动续期证书。

目前版本,容器内存占用稳定在`1,4MiB`