# pocker

Docker 风格的 [skopeo](https://github.com/containers/skopeo) 包装 CLI，支持 **Linux** 与 **Windows 原生**（非 WSL / WSL2）。

## 依赖

- 已安装并可在 `PATH` 中找到的 `skopeo`（Windows 为 `skopeo.exe`）
- `pull` 需要本机 Docker Daemon（Docker Desktop / Engine）

可通过配置指定二进制路径：

```bash
pocker config skopeo_path /usr/bin/skopeo
# Windows 示例
pocker config skopeo_path C:\path\to\skopeo.exe
```

## 安装

```bash
# 需要 Go 1.22+
export GOPROXY=https://goproxy.cn,direct
go install github.com/pocker/pocker@latest   # 或在本仓库: go build -o pocker .
```

交叉编译：

```bash
make build          # 当前平台
make build-all      # linux/amd64 + windows/amd64
```

## 代理配置（持久化）

配置写入本机文件，之后所有 `pull` / `save` 都会带上代理环境变量（若进程环境中已有同名变量，则以环境变量为准）。

| 平台 | 路径 |
|------|------|
| Linux | `~/.config/pocker/config.json`（或 `$XDG_CONFIG_HOME/pocker/`） |
| Windows | `%APPDATA%\pocker\config.json` |

支持 `http://`、`https://`、`socks5://`、`socks5h://`：

```bash
# HTTP 代理
pocker config http_proxy http://127.0.0.1:7890
pocker config https_proxy http://127.0.0.1:7890

# SOCKS5（推荐写 all_proxy；socks5h = 远程 DNS）
pocker config all_proxy socks5://127.0.0.1:1080
# 或
pocker config http_proxy socks5://127.0.0.1:1080   # 会同时导出 ALL_PROXY

pocker config no_proxy localhost,127.0.0.1

pocker config                 # 列出
pocker config http_proxy      # 读取
pocker config --unset all_proxy
```

也可用 `POCKER_CONFIG_DIR` 覆盖配置目录。

## 命令对照

本地镜像名与 `pull` / `save` 参数一致（**保留完整 tag**，含 `ghcr.io/` 前缀，便于 compose 区分 Hub / GHCR）。

### pull — 远程 → 本地 Docker Daemon

```bash
pocker pull ubuntu:latest
# ≡ skopeo copy docker://docker.io/library/ubuntu:latest docker-daemon:ubuntu:latest

pocker pull ghcr.io/username/image:tag
# ≡ skopeo copy docker://ghcr.io/username/image:tag docker-daemon:ghcr.io/username/image:tag

# 额外 tag（原有 tag 保留；pull 前会检查追加 tag 是否已在本地占用）
pocker pull ghcr.io/username/image:tag --tag username/images:tag
# 拉取后本地至少有:
#   ghcr.io/username/image:tag
#   username/images:tag
# 可重复: -t a:1 -t b:2
# 若 username/images:tag 本地已存在 → 直接报错，不开始 pull
```

### save — 远程 → 本地 docker-archive

可指定输出路径（位置参数或 `-o`），以及额外 tag（`--tag` / `-t`）：

```bash
pocker save nginx:alpine
# ≡ docker-archive:nginx-alpine.tar.gz:nginx:alpine

pocker save ghcr.io/stefanprodan/podinfo:latest
# ≡ docker-archive:ghcr.io-stefanprodan-podinfo-latest.tar.gz:ghcr.io/stefanprodan/podinfo:latest

pocker save nginx:alpine /tmp/nginx.tar.gz
pocker save nginx:alpine -o /tmp/nginx.tar.gz

pocker save ghcr.io/username/image:tag -o out.tar.gz --tag username/images:tag
# archive 内同时带有两个 tag
```

未指定路径时，默认文件名由完整引用中的 `/`、`:` 换成 `-` 得到。

## 开发

```bash
go test ./...
go build -o pocker .
./pocker version
```
