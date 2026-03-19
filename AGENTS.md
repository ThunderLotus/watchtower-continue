# Watchtower 项目指南

## 项目概述

Watchtower 是一个用于自动化 Docker 容器基础镜像更新的进程。它能够监控 Docker 容器，当检测到新的镜像发布时，自动拉取新镜像、优雅地关闭现有容器，并使用与初始部署时相同的选项重启容器。

**⚠️ 重要提示**: 该项目已不再维护，详细信息请参见 https://github.com/containrrr/watchtower/discussions/2135

### 主要技术栈

- **编程语言**: Go 1.20
- **依赖管理**: Go Modules
- **CLI 框架**: Cobra
- **日志**: Logrus
- **测试**: Ginkgo + Gomega
- **监控**: Prometheus
- **通知**: 支持多种通知方式（Email、Slack、Gotify、Microsoft Teams 等）

### 项目架构

```
watchtower/
├── main.go                    # 程序入口
├── cmd/                       # 命令行接口
│   ├── root.go               # 根命令，包含主要逻辑
│   └── notify-upgrade.go     # 通知升级命令
├── internal/                  # 内部包（不对外暴露）
│   ├── actions/              # 核心操作（检查、更新）
│   ├── flags/                # 命令行标志处理
│   ├── meta/                 # 元数据（版本信息）
│   └── util/                 # 工具函数
├── pkg/                       # 公共包
│   ├── api/                  # HTTP API（更新、指标）
│   ├── container/            # Docker 容器客户端
│   ├── filters/              # 容器过滤器
│   ├── lifecycle/            # 生命周期钩子
│   ├── metrics/              # Prometheus 指标
│   ├── notifications/        # 通知系统
│   ├── registry/             # 镜像仓库（认证、摘要、清单）
│   ├── session/              # 会话管理
│   ├── sorter/               # 排序器
│   └── types/                # 类型定义
├── dockerfiles/               # Dockerfile 配置
├── docs/                      # 项目文档
├── scripts/                   # 构建和测试脚本
└── grafana/                   # Grafana 仪表板配置
```

## 构建和运行

### 前置要求

- Go 1.20 或更高版本
- Docker（用于运行和测试）

### 本地构建

```bash
# 编译项目
go build -o watchtower

# 运行测试
go test ./... -v

# 运行程序
./watchtower
```

### 构建 Docker 镜像

```bash
# 使用本地文件构建镜像
docker build . -f dockerfiles/Dockerfile.dev-self-contained -t containrrr/watchtower

# 使用 GitHub 仓库构建镜像
docker build . -f dockerfiles/Dockerfile.self-contained -t containrrr/watchtower
```

### 运行 Watchtower

```bash
# 基本运行
docker run --detach \
    --name watchtower \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    containrrr/watchtower

# 带配置运行
docker run --detach \
    --name watchtower \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    containrrr/watchtower \
    --interval 86400 \
    --cleanup
```

### 开发环境

使用 docker-compose 启动完整的开发环境（包括 Prometheus 和 Grafana）：

```bash
docker-compose up
```

这将启动：
- Watchtower 容器
- Prometheus（端口 9090）
- Grafana（端口 3000）
- 测试容器（parent 和 child）

## 主要功能

### 1. 容器更新监控
- 定期检查容器镜像更新
- 支持通过标签过滤容器
- 支持通过名称过滤容器
- 支持监控范围（scope）隔离

### 2. HTTP API
- **更新 API**: 手动触发容器更新
- **指标 API**: 获取 Prometheus 指标
- 支持令牌认证
- 端口：8080

### 3. 通知系统
支持多种通知渠道：
- Email
- Slack
- Gotify
- Microsoft Teams
- Shoutrrr（支持多种通知服务）
- 自定义通知

### 4. 生命周期钩子
- 预停止钩子（Pre-stop）
- 后启动钩子（Post-start）
- 支持通过环境变量配置

### 5. 滚动重启
- 支持依赖容器的滚动重启
- 使用标签定义容器依赖关系

## 开发规范

### 代码风格
- 遵循 Go 标准代码风格
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量

### 测试规范
- 使用 Ginkgo 作为测试框架
- 使用 Gomega 作为断言库
- 所有新功能必须包含测试
- 运行 `go test ./... -v` 执行所有测试

### 提交规范
- 保持提交信息清晰简洁
- 一个提交解决一个问题
- 提交前确保所有测试通过

### 目录结构
- `internal/` 包含不对外暴露的内部实现
- `pkg/` 包含可复用的公共代码
- 每个包都有对应的测试文件

## 重要文件说明

### 核心文件
- **main.go**: 程序入口点
- **cmd/root.go**: 命令行根命令和主执行逻辑
- **cmd/notify-upgrade.go**: 通知升级命令
- **build.sh**: 构建脚本

### 配置文件
- **go.mod**: Go 模块依赖定义
- **docker-compose.yml**: 开发环境配置
- **prometheus/prometheus.yml**: Prometheus 配置
- **grafana/**: Grafana 仪表板和数据源配置

### 文档文件
- **README.md**: 项目说明和快速开始指南
- **CONTRIBUTING.md**: 贡献指南
- **docs/**: 详细文档目录
  - **arguments.md**: 命令行参数说明
  - **usage-overview.md**: 使用概述
  - **notifications.md**: 通知配置
  - **metrics.md**: 指标说明

## 常用命令

### 测试
```bash
# 运行所有测试
go test ./... -v

# 运行特定包的测试
go test ./pkg/container -v

# 运行测试并生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 调试
```bash
# 以调试模式运行
./watchtower --debug

# 以单次运行模式执行
./watchtower --run-once

# 启用详细日志
./watchtower --log-level trace
```

### Docker 相关
```bash
# 构建镜像
docker build . -f dockerfiles/Dockerfile.dev-self-contained -t my-watchtower

# 运行容器
docker run --detach \
    --name watchtower \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    my-watchtower

# 查看日志
docker logs -f watchtower
```

## 依赖说明

### 主要依赖
- `github.com/containrrr/shoutrrr`: 通知服务
- `github.com/docker/docker`: Docker SDK
- `github.com/spf13/cobra`: CLI 框架
- `github.com/spf13/viper`: 配置管理
- `github.com/robfig/cron`: 定时任务
- `github.com/prometheus/client_golang`: Prometheus 指标
- `github.com/onsi/ginkgo`: 测试框架
- `github.com/onsi/gomega`: 断言库

### 管理依赖
```bash
# 添加新依赖
go get github.com/username/package

# 更新依赖
go get -u ./...

# 整理依赖
go mod tidy
```

## 注意事项

1. **项目状态**: ⚠️ 该项目已不再维护，建议考虑替代方案
2. **生产环境**: 不建议在生产环境使用，建议使用 Kubernetes 等容器编排工具
3. **权限**: Watchtower 需要访问 Docker socket，需要相应的权限
4. **资源占用**: 定期检查会消耗一定的资源，建议合理设置检查间隔
5. **容器依赖**: 使用滚动重启时，需要正确配置容器依赖关系

## 故障排查

### 常见问题

1. **无法访问 Docker API**
   - 确保 Docker socket 已正确挂载
   - 检查权限设置

2. **容器更新失败**
   - 检查网络连接
   - 验证镜像仓库访问权限
   - 查看日志了解详细错误信息

3. **通知不工作**
   - 验证通知配置
   - 检查网络连接
   - 确认通知服务可用

4. **性能问题**
   - 调整检查间隔
   - 使用过滤器减少监控容器数量
   - 考虑使用多个 Watchtower 实例分担负载

## 相关资源

- **项目主页**: https://github.com/containrrr/watchtower
- **完整文档**: https://containrrr.dev/watchtower
- **Docker Hub**: https://hub.docker.com/r/containrrr/watchtower
- **Discussions**: https://github.com/containrrr/watchtower/discussions

## 许可证

Apache-2.0 License