# My Gin Server

一个基于Gin框架的Go Web服务器。

## 快速开始

### 本地运行

1. 安装依赖：
```bash
go mod download
```

2. 运行服务器：
```bash
go run main.go
```

服务器将在 http://localhost:8081 启动。

### Docker运行

1. 构建镜像：
```bash
docker build -t myginserver .
```

2. 运行容器：
```bash
docker run -p 8081:8081 myginserver
```

## 配置

服务器使用 `config.yaml` 文件进行配置。确保在运行前正确配置数据库连接等参数。

## 端口

- 主服务端口：8081
- Prometheus监控端口：18081（如果启用）

## 技术栈

- Gin Web框架
- MySQL数据库
- JWT认证
- Prometheus监控（可选）
- 结构化日志